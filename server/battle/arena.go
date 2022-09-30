package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/quest"
	"server/replay"
	"server/system_messages"
	"server/telegram"
	"server/xsyn_rpcclient"
	"strconv"
	"strings"
	"time"

	"github.com/sasha-s/go-deadlock"

	"golang.org/x/exp/slices"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"nhooyr.io/websocket"
)

type NewBattleChan struct {
	ID           string
	BattleNumber int
}

type ArenaManager struct {
	server                   *http.Server
	Addr                     string
	timeout                  time.Duration
	RPCClient                *xsyn_rpcclient.XsynXrpcClient
	sms                      server.SMS
	telegram                 server.Telegram
	gameClientMinimumBuildNo uint64
	SystemBanManager         *SystemBanManager
	NewBattleChan            chan *NewBattleChan
	SystemMessagingManager   *system_messages.SystemMessagingManager
	RepairFuncMx             deadlock.Mutex
	BattleQueueFuncMx        deadlock.Mutex
	QuestManager             *quest.System

	arenas           map[string]*Arena
	deadlock.RWMutex // lock for arena

	ChallengeFundUpdateChan chan bool
	LobbyFuncMx             *deadlock.Mutex
}

type Opts struct {
	Addr                     string
	Timeout                  time.Duration
	RPCClient                *xsyn_rpcclient.XsynXrpcClient
	SMS                      server.SMS
	GameClientMinimumBuildNo uint64
	Telegram                 *telegram.Telegram
	SystemMessagingManager   *system_messages.SystemMessagingManager
	QuestManager             *quest.System
}

func NewArenaManager(opts *Opts) (*ArenaManager, error) {
	am := &ArenaManager{
		Addr:                     opts.Addr,
		timeout:                  opts.Timeout,
		RPCClient:                opts.RPCClient,
		sms:                      opts.SMS,
		telegram:                 opts.Telegram,
		gameClientMinimumBuildNo: opts.GameClientMinimumBuildNo,
		SystemBanManager:         NewSystemBanManager(),
		NewBattleChan:            make(chan *NewBattleChan),
		SystemMessagingManager:   opts.SystemMessagingManager,
		QuestManager:             opts.QuestManager,
		arenas:                   make(map[string]*Arena),

		ChallengeFundUpdateChan: make(chan bool),
		LobbyFuncMx:             &deadlock.Mutex{},
	}

	am.server = &http.Server{
		Handler:      am,
		ReadTimeout:  am.timeout,
		WriteTimeout: am.timeout,
	}

	// delete all the unfinished AI driven battles
	_, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.EndedAt.IsNull(),
		boiler.BattleLobbyWhere.IsAiDrivenMatch.EQ(true),
	).UpdateAll(gamedb.StdConn, boiler.M{
		boiler.BattleLobbyColumns.DeletedAt: null.TimeFrom(time.Now()),
	})
	if err != nil {
		return nil, terror.Error(err, "Failed to delete unfinished AI battles.")
	}

	// start player rank updater
	am.PlayerRankUpdater()

	// check default battle lobbies
	err = am.SetDefaultPublicBattleLobbies()
	if err != nil {
		return nil, err
	}

	// start repair offer cleaner
	go am.RepairOfferCleaner()

	return am, nil
}

func (am *ArenaManager) GetArenaFromContext(ctx context.Context) (*Arena, error) {
	arenaID := chi.RouteContext(ctx).URLParam("arena_id")
	if arenaID == "" {
		return nil, terror.Error(fmt.Errorf("missing arena id"), "Missing arena id")
	}

	arena, err := am.GetArena(arenaID)
	if err != nil {
		return nil, err
	}

	return arena, nil
}

func (am *ArenaManager) GetArena(arenaID string) (*Arena, error) {
	am.RLock()
	defer am.RUnlock()
	arena, ok := am.arenas[arenaID]
	if !ok || arena.Stage.Load() == ArenaStageHijacked {
		return nil, terror.Error(fmt.Errorf("arena not exits"), "The battle arena does not exist.")
	}
	if !arena.connected.Load() {
		return nil, terror.Error(fmt.Errorf("arena not available"), "The battle arena is not available")
	}

	return arena, nil
}

// KickIdleArenas reactivate idle arena
func (am *ArenaManager) KickIdleArenas() {
	am.RLock()
	defer am.RUnlock()
	for _, a := range am.arenas {
		if a.connected.Load() && a.Stage.Load() == ArenaStageIdle {
			go a.BeginBattle()
		}
	}
}

type ArenaBrief struct {
	ID    string `json:"id"`
	Gid   int    `json:"gid"`
	Name  string `json:"name"`
	Stage string `json:"stage"`
}

func (am *ArenaManager) AvailableBattleArenas() []*ArenaBrief {
	am.RLock()
	defer am.RUnlock()

	resp := []*ArenaBrief{}
	for _, arena := range am.arenas {
		if arena.Stage.Load() != ArenaStageHijacked && arena.connected.Load() {
			resp = append(resp, &ArenaBrief{
				ID:    arena.ID,
				Gid:   arena.Gid,
				Name:  arena.Name,
				Stage: arena.Stage.Load(),
			})
		}
	}
	return resp
}

func (am *ArenaManager) GetCurrentBattleLobbyIDs() []string {
	var battleLobbyIDs []string
	for _, a := range am.arenas {
		lobbyID := a.currentLobbyID.Load()
		if lobbyID == "" || a.Stage.Load() == ArenaStageHijacked || !a.connected.Load() {
			continue
		}

		battleLobbyIDs = append(battleLobbyIDs, lobbyID)
	}

	return battleLobbyIDs
}


func (am *ArenaManager) Serve() {
	l, err := net.Listen("tcp", am.Addr)
	if err != nil {
		gamelog.L.Fatal().Str("Addr", am.Addr).Err(err).Msg("unable to bind Arena to Battle Server address")
	}
	go func() {
		gamelog.L.Info().Msgf("Starting Battle Arena Server on: %v", am.Addr)

		err := am.server.Serve(l)
		if err != nil {
			gamelog.L.Fatal().Str("Addr", am.Addr).Err(err).Msg("unable to start Battle Arena server")
		}
	}()
}

func (am *ArenaManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	arenaID := r.URL.Query().Get("id")
	if arenaID == "" {
		gamelog.L.Error().Msg("Arena id is missing.")
		return
	}

	// check arena exists
	battleArena, err := boiler.FindBattleArena(gamedb.StdConn, arenaID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("arena id", arenaID).Msg("Failed to load battle arena from db")
		return
	}

	if battleArena == nil {
		gamelog.L.Error().Err(err).Str("arena id", arenaID).Msg("Battle arena does not exist in db")
		return
	}

	gamelog.L.Info().Str("arena id", arenaID).Msg("New arena is connected.")

	wsConn, err := websocket.Accept(w, r, nil)
	if err != nil {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ipaddr, _, _ := net.SplitHostPort(r.RemoteAddr)
			userIP := net.ParseIP(ipaddr)
			if userIP == nil {
				ip = ipaddr
			} else {
				ip = userIP.String()
			}
		}
		gamelog.L.Warn().Str("request_ip", ip).Err(err).Msg("unable to start Battle Arena server")
		return
	}

	// create new arena
	arena, err := am.NewArena(battleArena, wsConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to add arena onto arena manager")
		return
	}

	// broadcast a new arena list to frontend
	ws.PublishMessage("/public/arena_list", server.HubKeyBattleArenaListSubscribe, am.AvailableBattleArenas())

	// handle arena close
	defer func() {
		am.Lock()
		defer am.Unlock()

		// remove arena from the map, if it is still connected
		if arena.connected.Load() {
			arena.connected.Store(false)
			// delete arena from the map
			delete(am.arenas, arena.ID)

			// tell frontend the arena is closed
			ws.PublishMessage(fmt.Sprintf("/public/arena/%s/closed", arena.ID), server.HubKeyBattleArenaClosedSubscribe, true)

			// broadcast a new arena list to frontend
			arenaList := []*boiler.BattleArena{}
			for _, a := range am.arenas {
				if a.connected.Load() {
					arenaList = append(arenaList, a.BattleArena)
				}
			}

			// broadcast a new arena list to frontend
			ws.PublishMessage("/public/arena_list", server.HubKeyBattleArenaListSubscribe, arenaList)
		}

		// clean up ws, if connection still exists
		if wsConn != nil {
			// clean up
			gamelog.L.Error().Err(fmt.Errorf("game client has disconnected")).Msg("lost connection to game client")

			err = wsConn.Close(websocket.StatusInternalError, "game client has disconnected")
			if err != nil {
				gamelog.L.Error().Str("arena id", arena.ID).Err(err).Msg("Failed to close ws connection")
			}
		}

		// clean up current battle, if exists
		if btl := arena.CurrentBattle(); btl != nil {
			btl.endAbilities()

			if btl.replaySession.ReplaySession != nil {
				battle := *btl.Battle
				go func(battle boiler.Battle, replayID string) {
					err = replay.RecordReplayRequest(&battle, replayID, replay.StopRecording)
					if err != nil {
						gamelog.L.Error().Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Err(err).Msg("failed to stop recording during game client disconnection")
					}
				}(battle, btl.replaySession.ReplaySession.ID)

				eventByte, err := json.Marshal(btl.replaySession.Events)
				if err != nil {
					gamelog.L.Error().Err(err).Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Interface("Events", btl.replaySession.Events).Msg("Failed to marshal json into battle replay")
				} else {
					btl.replaySession.ReplaySession.BattleEvents = null.JSONFrom(eventByte)
				}
				btl.replaySession.ReplaySession.StoppedAt = null.TimeFrom(time.Now())
				btl.replaySession.ReplaySession.RecordingStatus = boiler.RecordingStatusSTOPPED
				_, err = btl.replaySession.ReplaySession.Update(
					gamedb.StdConn,
					boil.Whitelist(
						boiler.BattleReplayColumns.StoppedAt,
						boiler.BattleReplayColumns.RecordingStatus,
					),
				)
				if err != nil {
					gamelog.L.Error().Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Err(err).Msg("Failed to update replay session")
				}
			}
		}

		// kick idle arenas
		for _, a := range am.arenas {
			if a.connected.Load() && a.Stage.Load() == ArenaStageIdle {
				go a.BeginBattle()
			}
		}
	}()

	arena.Start()
}

func (am *ArenaManager) NewArena(battleArena *boiler.BattleArena, wsConn *websocket.Conn) (*Arena, error) {
	am.Lock()
	defer am.Unlock()
	var err error

	// if previous arena is not closed properly.
	if a, ok := am.arenas[battleArena.ID]; ok {
		// set connected flag of the prev arena to false
		a.connected.Store(false)

		// change arena stage to hijacked
		a.Stage.Store(ArenaStageHijacked)

		// stop recording from previous arena
		if btl := a.CurrentBattle(); btl != nil && btl.replaySession.ReplaySession != nil {
			err = replay.RecordReplayRequest(btl.Battle, btl.replaySession.ReplaySession.ID, replay.StopRecording)
			if err != nil {
				gamelog.L.Error().Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Err(err).Msg("failed to stop recording during game client disconnection")
			}

			var eventByte []byte
			eventByte, err = json.Marshal(btl.replaySession.Events)
			if err != nil {
				gamelog.L.Error().Err(err).Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Interface("Events", btl.replaySession.Events).Msg("Failed to marshal json into battle replay")
			} else {
				btl.replaySession.ReplaySession.BattleEvents = null.JSONFrom(eventByte)
			}
			btl.replaySession.ReplaySession.StoppedAt = null.TimeFrom(time.Now())
			btl.replaySession.ReplaySession.RecordingStatus = boiler.RecordingStatusSTOPPED
			_, err = btl.replaySession.ReplaySession.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Err(err).Msg("Failed to update replay session")
			}
		}
	}

	arena := &Arena{
		BattleArena:            battleArena,
		currentLobbyID:         atomic.NewString(""),
		Name:                   helpers.GenerateStupidArenaName(),
		Stage:                  atomic.NewString(ArenaStageIdle),
		socket:                 wsConn,
		connected:              atomic.NewBool(true),
		gameClientJsonDataChan: make(chan []byte, 3),
		MechCommandCheckMap: &MechCommandCheckMap{
			m: make(map[string]chan bool),
		},

		// objects inherited from arena manager
		Manager: am,
	}

	arena.AIPlayers, err = db.DefaultFactionPlayers()
	if err != nil {
		gamelog.L.Fatal().Err(err).Msg("no faction users found")
	}

	if arena.timeout == 0 {
		arena.timeout = 15 * time.Hour * 24
	}

	// speed up mech stat broadcast by separate json data and binary data
	go arena.GameClientJsonDataParser()

	am.arenas[arena.ID] = arena

	return arena, nil
}

const (
	ArenaStageHijacked   string = "HIJACKED"
	ArenaStageIdle       string = "IDLE"
	ArenaStageProcessing string = "PROCESSING"
)

type Arena struct {
	*boiler.BattleArena
	Manager        *ArenaManager
	currentLobbyID *atomic.String // only used for assigning lobby

	Name             string
	Stage            *atomic.String // hijacked, idle, running
	socket           *websocket.Conn
	connected        *atomic.Bool
	timeout          time.Duration
	_currentBattle   *Battle
	LastBattleResult *BattleEndDetail
	AIPlayers        map[string]db.PlayerWithFaction

	gameClientJsonDataChan chan []byte

	MechCommandCheckMap *MechCommandCheckMap
	deadlock.RWMutex

	beginBattleMux deadlock.Mutex
}

type MechCommandCheckMap struct {
	m map[string]chan bool
	deadlock.RWMutex
}

func (mc *MechCommandCheckMap) Register(key string, ch chan bool) {
	mc.Lock()
	defer mc.Unlock()

	mc.m[key] = ch
}
func (mc *MechCommandCheckMap) Remove(key string) {
	mc.Lock()
	defer mc.Unlock()
	if _, ok := mc.m[key]; ok {
		delete(mc.m, key)
	}
}
func (mc *MechCommandCheckMap) Send(key string, isValid bool) {
	mc.RLock()
	defer mc.RUnlock()
	if ch, ok := mc.m[key]; ok {
		ch <- isValid
	}
}

func (am *ArenaManager) IsClientConnected() error {
	count := 0
	for _, arena := range am.arenas {
		if arena.connected.Load() {
			count += 1
		}
	}
	if count == 0 {
		return fmt.Errorf("no gameclient connected")
	}
	return nil
}

func (arena *Arena) CurrentBattle() *Battle {
	arena.RLock()
	defer arena.RUnlock()
	return arena._currentBattle
}

func (arena *Arena) storeCurrentBattle(btl *Battle) {
	arena.Lock()
	defer arena.Unlock()
	arena._currentBattle = btl
}

func (arena *Arena) currentBattleState() int32 {
	arena.RLock()
	defer arena.RUnlock()
	if arena._currentBattle == nil {
		return BattleStageEnd
	}

	arena._currentBattle.RLock()
	stage := arena._currentBattle.stage.Load()
	arena._currentBattle.RUnlock()

	return stage
}

func (arena *Arena) currentBattleNumber() int {
	arena.RLock()
	defer arena.RUnlock()
	if arena._currentBattle == nil {
		return -1
	}
	return arena._currentBattle.BattleNumber
}

func (arena *Arena) currentBattleWarMachineIDs(factionIDs ...string) []string {
	arena.RLock()
	defer arena.RUnlock()

	ids := []string{}

	if arena._currentBattle == nil {
		return ids
	}

	if factionIDs != nil && len(factionIDs) > 0 {
		// only return war machines' id from the faction
		for _, wm := range arena._currentBattle.WarMachines {
			if wm.FactionID == factionIDs[0] {
				ids = append(ids, wm.ID)
			}
		}
	} else {
		// return all the war machines' id
		ids = arena._currentBattle.warMachineIDs
	}

	return ids
}

func (arena *Arena) CurrentBattleWarMachineOrAIByHash(hash string) *WarMachine {
	arena.RLock()
	defer arena.RUnlock()

	for _, wm := range arena._currentBattle.WarMachines {
		if wm.Hash == hash {
			return wm
		}
	}

	for _, wm := range arena._currentBattle.SpawnedAI {
		if wm.Hash == hash {
			return wm
		}
	}

	return nil
}

func (arena *Arena) CurrentBattleWarMachineByHash(hash string) *WarMachine {
	arena.RLock()
	defer arena.RUnlock()

	for _, wm := range arena._currentBattle.WarMachines {
		if wm.Hash == hash {
			return wm
		}
	}

	return nil
}

func (arena *Arena) CurrentBattleWarMachineByID(id string) *WarMachine {
	arena.RLock()
	defer arena.RUnlock()

	for _, wm := range arena._currentBattle.WarMachines {
		if wm.ID == id {
			return wm
		}
	}

	return nil
}

func (arena *Arena) CurrentBattleWarMachine(participantID int) *WarMachine {
	arena.RLock()
	defer arena.RUnlock()

	if arena._currentBattle == nil {
		return nil
	}

	for _, wm := range arena._currentBattle.WarMachines {
		if checkWarMachineByParticipantID(wm, participantID) {
			return wm
		}
	}

	for _, wm := range arena._currentBattle.SpawnedAI {
		if checkWarMachineByParticipantID(wm, participantID) {
			return wm
		}
	}

	return nil
}

func (arena *Arena) currentDisableCells() []int64 {
	arena.RLock()
	defer arena.RUnlock()

	if arena._currentBattle == nil {
		return nil
	}

	arena._currentBattle.RLock()
	defer arena._currentBattle.RUnlock()
	return arena._currentBattle.gameMap.DisabledCells
}

func checkWarMachineByParticipantID(wm *WarMachine, participantID int) bool {
	wm.RLock()
	defer wm.RUnlock()
	if int(wm.ParticipantID) == participantID {
		return true
	}
	return false
}

func (am *ArenaManager) WarMachineDestroyedDetail(mechID string) *WMDestroyedRecord {
	am.RLock()
	defer am.RUnlock()

	for _, arena := range am.arenas {
		dr := func(arena *Arena) *WMDestroyedRecord {
			arena.RLock()
			defer arena.RUnlock()

			if arena._currentBattle == nil {
				return nil
			}

			record, ok := arena._currentBattle.destroyedWarMachineMap[mechID]
			if !ok {
				return nil
			}

			return record
		}(arena)

		if dr != nil {
			return dr
		}
	}

	return nil
}

type MessageType byte

// NetMessageTypes
const (
	JSON MessageType = iota
	Tick
)

func (mt MessageType) String() string {
	return [...]string{"JSON", "Tick", "Live Vote Tick", "Viewer Live Count Tick", "Spoils of War Tick", "game ability progress tick", "battle ability progress tick", "unknown", "unknown wtf"}[mt]
}

type AuthMiddleware func(required bool, userIDMustMatch bool) func(next http.Handler) http.Handler
type AuthFactionMiddleware func(factionIDMustMatch bool) func(next http.Handler) http.Handler

const HubKeyBribingWinnerSubscribe = "BRIBE:WINNER:SUBSCRIBE"

const BATTLEINIT = "BATTLE:INIT"

// Start begins the battle arena, blocks on listen
func (arena *Arena) Start() {
	arena.start()
}

func (arena *Arena) Message(cmd string, payload interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := json.Marshal(struct {
		Command string      `json:"battle_command"`
		Payload interface{} `json:"payload"`
	}{Payload: payload, Command: cmd})
	if err != nil {
		gamelog.L.Fatal().Str("log_name", "battle arena").Interface("payload", payload).Err(err).Msg("unable to marshal data for battle arena")
	}
	err = arena.socket.Write(ctx, websocket.MessageBinary, b)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("payload", payload).Err(err).Msg("failed to write websocket message to game client")
		return
	}
	gamelog.L.Info().RawJSON("message data", b).Msg("game client message sent")
}

type LocationSelectRequest struct {
	Payload struct {
		ArenaID string `json:"arena_id"`

		StartCoords server.CellLocation  `json:"start_coords"`
		EndCoords   *server.CellLocation `json:"end_coords,omitempty"`
	} `json:"payload"`
}

const HubKeyAbilityLocationSelect = "ABILITY:LOCATION:SELECT"

var locationSelectBucket = leakybucket.NewCollector(1, 1, true)

func (am *ArenaManager) AbilityLocationSelect(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if locationSelectBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many Requests")
	}

	req := &LocationSelectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	arena, err := am.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	btl := arena.CurrentBattle()
	// skip, if current not battle
	if btl == nil {
		gamelog.L.Warn().Msg("no current battle")
		return nil
	}

	as := btl.AbilitySystem()

	if !AbilitySystemIsAvailable(as) {
		gamelog.L.Error().Str("log_name", "battle arena").Msg("AbilitySystem is nil")
		return terror.Error(fmt.Errorf("ability system is closed"), "Ability system is closed")
	}

	err = as.LocationSelect(user.ID, factionID, req.Payload.StartCoords, req.Payload.EndCoords)
	if err != nil {
		gamelog.L.Warn().Err(err).Msgf("Unable to select location")
		return terror.Error(err, "Unable to select location")
	}

	reply(true)
	return nil
}

type MinimapEvent struct {
	ID            string              `json:"id"`
	GameAbilityID int                 `json:"game_ability_id"`
	Duration      int                 `json:"duration"`
	Radius        int                 `json:"radius"`
	Coords        server.CellLocation `json:"coords"`
}

const HubKeyMinimapUpdatesSubscribe = "MINIMAP:UPDATES:SUBSCRIBE"

func (am *ArenaManager) MinimapUpdatesSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return terror.Error(terror.ErrForbidden, "There is no battle currently to use this ability on.")
	}

	minimapUpdates := []MinimapEvent{}
	if btl := arena.CurrentBattle(); btl != nil {
		if pam := arena.CurrentBattle().playerAbilityManager(); pam != nil {
			for id, b := range pam.Blackouts() {
				minimapUpdates = append(minimapUpdates, MinimapEvent{
					ID:            id,
					GameAbilityID: BlackoutGameAbilityID,
					Duration:      BlackoutDurationSeconds,
					Radius:        int(BlackoutRadius),
					Coords:        b.CellCoords,
				})
			}
		}
	}
	reply(minimapUpdates)
	return nil
}

const HubKeyMinimapEventsSubscribe = "MINIMAP:EVENTS:SUBSCRIBE"

func (am *ArenaManager) MinimapEventsSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	// if current battle still running
	btl := arena.CurrentBattle()
	if btl == nil {
		return nil
	}

	// send landmine, pickup locations and the hive map state
	hasMessages, mapEventsPacked := btl.MapEventList.Pack()
	if hasMessages {
		reply(mapEventsPacked)
	}

	return nil
}

const HubKeyBattleAbilityUpdated = "BATTLE:ABILITY:UPDATED"

// PublicBattleAbilityUpdateSubscribeHandler return battle ability for non login player
func (am *ArenaManager) PublicBattleAbilityUpdateSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	// get a random faction id
	if arena.CurrentBattle() != nil {
		as := arena.CurrentBattle().AbilitySystem()
		if AbilitySystemIsAvailable(as) {
			ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
			ga, err := boiler.GameAbilities(
				boiler.GameAbilityWhere.BattleAbilityID.EQ(null.StringFrom(ba.ID)),
			).One(gamedb.StdConn)
			if err != nil {
				return terror.Error(err, "Failed to get battle ability")
			}
			reply(GameAbility{
				ID:                     ga.ID,
				GameClientAbilityID:    byte(ga.GameClientAbilityID),
				ImageUrl:               ga.ImageURL,
				Description:            ba.Description,
				FactionID:              ga.FactionID,
				Label:                  ga.Label,
				Colour:                 ga.Colour,
				TextColour:             ga.TextColour,
				CooldownDurationSecond: ba.CooldownDurationSecond,
			})
		}
	}
	return nil
}

const HubKeyBattleAbilityOptInCheck = "BATTLE:ABILITY:OPT:IN:CHECK"

// BattleAbilityOptInSubscribeHandler return battle ability for non login player
func (am *ArenaManager) BattleAbilityOptInSubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	// get a random faction id
	if arena.CurrentBattle() != nil {
		as := arena.CurrentBattle().AbilitySystem()
		if AbilitySystemIsAvailable(as) {
			offeringID := as.BattleAbilityPool.BattleAbility.LoadOfferingID()

			isOptedIn, err := boiler.BattleAbilityOptInLogs(
				boiler.BattleAbilityOptInLogWhere.BattleAbilityOfferingID.EQ(offeringID),
				boiler.BattleAbilityOptInLogWhere.PlayerID.EQ(user.ID),
			).Exists(gamedb.StdConn)
			if err != nil {
				return terror.Error(err, "Failed to check opt in stat")
			}

			reply(isOptedIn)
		}
	}
	return nil
}

const HubKeyWarMachineAbilitiesUpdated = "WAR:MACHINE:ABILITIES:UPDATED"

type MechGameAbility struct {
	boiler.GameAbility
	CoolDownSeconds int `json:"cool_down_seconds"`
}

// WarMachineAbilitiesUpdateSubscribeHandler subscribe on war machine AbilitySystem
func (am *ArenaManager) WarMachineAbilitiesUpdateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	if arena.currentBattleState() != BattleStageStart {
		return nil
	}

	slotNumber := chi.RouteContext(ctx).URLParam("slotNumber")
	if slotNumber == "" {
		return fmt.Errorf("slot number is required")
	}

	participantID, err := strconv.Atoi(slotNumber)
	if err != nil {
		return fmt.Errorf("invalid participant id")
	}

	wm := arena.CurrentBattleWarMachine(participantID)
	if wm == nil {
		return fmt.Errorf("failed to load war machine")
	}

	if wm.OwnedByID != user.ID {
		reply([]*boiler.GameAbility{})
		return nil
	}

	// load game ability
	gas, err := boiler.GameAbilities(
		boiler.GameAbilityWhere.FactionID.EQ(wm.FactionID),
		boiler.GameAbilityWhere.Level.EQ(boiler.AbilityLevelPLAYER),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("faction id", wm.FactionID).Str("ability level", boiler.AbilityLevelPLAYER).Err(err).Msg("Failed to get game AbilitySystem from db")
		return terror.Error(err, "Failed to load game AbilitySystem.")
	}

	reply(gas)

	return nil
}

const HubKeyWarMachineAbilitySubscribe = "WAR:MACHINE:ABILITY:SUBSCRIBE"

func (am *ArenaManager) WarMachineAbilitySubscribe(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	btl := arena.CurrentBattle()
	if btl == nil || btl.stage.Load() != BattleStageStart {
		return nil
	}

	cctx := chi.RouteContext(ctx)
	slotNumber := cctx.URLParam("slotNumber")
	if slotNumber == "" {
		return fmt.Errorf("slot number is required")
	}
	mechAbilityID := cctx.URLParam("mech_ability_id")
	if mechAbilityID == "" {
		return fmt.Errorf("mech ability is required")
	}

	participantID, err := strconv.Atoi(slotNumber)
	if err != nil {
		return fmt.Errorf("invalid participant id")
	}

	wm := arena.CurrentBattleWarMachine(participantID)
	if wm == nil {
		return fmt.Errorf("failed to load mech detail")
	}

	if wm.OwnedByID != user.ID {
		return terror.Error(fmt.Errorf("does not own the mech"), "You do not own the mech.")
	}

	ga, err := boiler.FindGameAbility(gamedb.StdConn, mechAbilityID)
	if err != nil {
		return terror.Error(err, "Failed to load game ability.")
	}

	coolDownSeconds := db.GetIntWithDefault(db.KeyMechAbilityCoolDownSeconds, 30)

	// validate the ability can be triggered
	switch ga.Label {
	case "REPAIR":
		// get ability from db
		lastTrigger, err := boiler.BattleAbilityTriggers(
			boiler.BattleAbilityTriggerWhere.OnMechID.EQ(null.StringFrom(wm.ID)),
			boiler.BattleAbilityTriggerWhere.GameAbilityID.EQ(mechAbilityID),
			boiler.BattleAbilityTriggerWhere.BattleID.EQ(btl.ID),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Failed to get last ability trigger")
		}

		if lastTrigger != nil {
			reply(86400)
			return nil
		}

	default:
		// get ability from db
		lastTrigger, err := boiler.BattleAbilityTriggers(
			boiler.BattleAbilityTriggerWhere.OnMechID.EQ(null.StringFrom(wm.ID)),
			boiler.BattleAbilityTriggerWhere.GameAbilityID.EQ(mechAbilityID),
			boiler.BattleAbilityTriggerWhere.TriggeredAt.GT(time.Now().Add(time.Duration(-coolDownSeconds)*time.Second)),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("mech id", wm.ID).Str("game ability id", mechAbilityID).Err(err).Msg("Failed to get mech ability trigger from db")
			return terror.Error(err, "Failed to load game ability")
		}

		if lastTrigger != nil {
			reply(coolDownSeconds - int(time.Now().Sub(lastTrigger.TriggeredAt).Seconds()))
			return nil
		}

	}
	reply(0)

	return nil
}

type WarMachineStat struct {
	ParticipantID int             `json:"participant_id"`
	Position      *server.Vector3 `json:"position"`
	Rotation      int             `json:"rotation"`
	Health        uint32          `json:"health"`
	Shield        uint32          `json:"shield"`
	IsHidden      bool            `json:"is_hidden"`
}

const HubKeyWarMachineStatUpdated = "WAR:MACHINE:STAT:UPDATED"

// WarMachineStatSubscribe subscribe on bribing stage change
func (am *ArenaManager) WarMachineStatSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	slotNumber := chi.RouteContext(ctx).URLParam("slotNumber")
	if slotNumber == "" {
		return fmt.Errorf("slot number is required")
	}

	participantID, err := strconv.Atoi(slotNumber)
	if err != nil || participantID == 0 {
		return fmt.Errorf("invlid participant id")
	}

	// return data if, current battle is not null
	wm := arena.CurrentBattleWarMachine(participantID)
	if wm != nil {
		wStat := &WarMachineStat{
			ParticipantID: participantID,
			Health:        wm.Health,
			Position:      wm.Position,
			Rotation:      wm.Rotation,
			IsHidden:      wm.IsHidden,
			Shield:        wm.Shield,
		}

		// Hidden/Incognito
		if wStat.Position != nil {
			hideMech := arena.CurrentBattle().playerAbilityManager().IsWarMachineHidden(wm.Hash)
			hideMech = hideMech || arena.CurrentBattle().playerAbilityManager().IsWarMachineInBlackout(server.GameLocation{
				X: wStat.Position.X,
				Y: wStat.Position.Y,
			})
			if hideMech {
				wStat.IsHidden = true
				wStat.Position = &server.Vector3{
					X: -1,
					Y: -1,
					Z: -1,
				}
			}
		}

		reply(wStat)
	}
	return nil
}

const HubKeyBribeStageUpdateSubscribe = "BRIBE:STAGE:UPDATED:SUBSCRIBE"

// BribeStageSubscribe subscribe on bribing stage change
func (am *ArenaManager) BribeStageSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}
	// return data if, current battle is not null
	if arena.CurrentBattle() != nil {
		btl := arena.CurrentBattle()
		if AbilitySystemIsAvailable(btl.AbilitySystem()) {
			reply(btl.AbilitySystem().BribeStageGet())
		}
	}
	return nil
}

type GameAbilityPriceResponse struct {
	ID          string `json:"id"`
	OfferingID  string `json:"offering_id"`
	SupsCost    string `json:"sups_cost"`
	CurrentSups string `json:"current_sups"`
	ShouldReset bool   `json:"should_reset"`
}

func (am *ArenaManager) SendSettings(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	// response game setting, if current battle exists
	if arena.CurrentBattle() != nil {
		reply(GameSettingsPayload(arena.CurrentBattle()))
	}

	return nil
}

type BattleMsg struct {
	BattleCommand string          `json:"battle_command"`
	Payload       json.RawMessage `json:"payload"`
}

type BattleStartPayload struct {
	WarMachines []struct {
		Hash          string `json:"hash"`
		ParticipantID byte   `json:"participant_id"`
	} `json:"war_machines"`
	BattleID      string `json:"battle_id"`
	ClientBuildNo string `json:"client_build_no"`
	MapName       string `json:"map_name"` // The name of the map actually loaded
}

type MapDetailsPayload struct {
	Details     server.GameMap      `json:"details"`
	BattleZones []server.BattleZone `json:"battle_zones"`
	BattleID    string              `json:"battle_id"`
}

type BattleEndPayload struct {
	WinningWarMachines []struct {
		Hash   string `json:"hash"`
		Health int    `json:"health"`
	} `json:"winning_war_machines"`
	BattleID     string `json:"battle_id"`
	WinCondition string `json:"win_condition"`
}

type AbilityMoveCommandCompletePayload struct {
	BattleID       string `json:"battle_id"`
	WarMachineHash string `json:"war_machine_hash"`
}

type ZoneChangePayload struct {
	BattleID  string `json:"battle_id"`
	ZoneIndex int    `json:"zone_index"`
	WarnTime  int    `json:"warn_time"`
}

type ZoneChangeEvent struct {
	Location   server.GameLocation `json:"location"`
	Radius     int                 `json:"radius"`
	ShrinkTime int                 `json:"shrink_time"`
	WarnTime   int                 `json:"warn_time"`
}

type MechMoveCommandResponsePayload struct {
	BattleID       string `json:"battle_id"`
	WarMachineHash string `json:"war_machine_hash"`
	EventID        string `json:"event_id"`
	IsValid        bool   `json:"is_valid"`
}

type AbilityCompletePayload struct {
	BattleID string `json:"battle_id"`
	EventID  string `json:"event_id"`
}

type BattleWMDestroyedPayload struct {
	BattleID                string `json:"battle_id"`
	DestroyedWarMachineHash string `json:"destroyed_war_machine_hash"`
	KilledByWarMachineHash  string `json:"killed_by_war_machine_hash"`
	RelatedEventIDString    string `json:"related_event_id_string"`
	DamageHistory           []struct {
		Amount         int    `json:"amount"`
		InstigatorHash string `json:"instigator_hash"`
		SourceHash     string `json:"source_hash"`
		SourceName     string `json:"source_name"`
	} `json:"damage_history"`
	KilledBy      string `json:"killed_by"`
	ParticipantID int    `json:"participant_id"`
}

type AISpawnedRequest struct {
	BattleID      string          `json:"battle_id"`
	ParticipantID byte            `json:"participant_id"`
	Hash          string          `json:"hash"`
	UserID        string          `json:"user_id"`
	Name          string          `json:"name"`
	Model         string          `json:"model"`
	Skin          string          `json:"skin"`
	MaxHealth     uint32          `json:"health_max"`
	Health        uint32          `json:"health"`
	MaxShield     uint32          `json:"shield_max"`
	Shield        uint32          `json:"shield"`
	FactionID     string          `json:"faction_id"`
	Position      *server.Vector3 `json:"position"`
	Rotation      int             `json:"rotation"`
	Type          AIType          `json:"type"`
}

type AIType string

const (
	Reinforcement AIType = "Reinforcement"
	MiniMech      AIType = "Mini Mech"
	RobotDog      AIType = "Robot Dog"
)

type BattleWMPickupPayload struct {
	WarMachineHash string `json:"war_machine_hash"`
	EventID        string `json:"event_id"`
	BattleID       string `json:"battle_id"`
}

type WarMachineStatusPayload struct {
	WarMachineHash string `json:"war_machine_hash"`
	EventID        string `json:"event_id"`
	BattleID       string `json:"battle_id"`
	Status         struct {
		IsHacked  bool `json:"is_hacked"`
		IsStunned bool `json:"is_stunned"`
	} `json:"war_machine_status"`
}

func (arena *Arena) start() {
	ctx := context.Background()
	arena.BeginBattle()

	for {
		_, payload, err := arena.socket.Read(ctx)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("empty game client disconnected")
			break
		}

		// if connected flag is false, skip all the process
		// NOTE: this only happen when there are multiple game clients spin up at the same time
		if !arena.connected.Load() {
			btl := arena.CurrentBattle()
			if btl != nil {
				btl.endAbilities()
				arena.storeCurrentBattle(nil)
			}
			// TODO: send shutdown command to the game client to prevent it from reconnecting again.
			continue
		}

		if len(payload) == 0 {
			gamelog.L.Warn().Bytes("payload", payload).Err(err).Msg("empty game client payload")
			continue
		}

		mt := MessageType(payload[0])
		if err != nil {
			gamelog.L.Warn().Int("message_type", int(mt)).Bytes("payload", payload).Err(err).Msg("websocket to game client failed")
			return
		}

		data := payload[1:]
		switch mt {
		case JSON:
			msg := &BattleMsg{}
			err := json.Unmarshal(data, msg)
			if err != nil {
				gamelog.L.Warn().Str("msg", string(data)).Err(err).Msg("unable to unmarshal battle message")
				continue
			}

			// set battle stage to end when receive "battle:end" message
			btl := arena.CurrentBattle()
			if btl != nil && msg.BattleCommand == "BATTLE:END" {
				btl.stage.Store(BattleStageEnd)
			}
			// handle message through channel
			arena.gameClientJsonDataChan <- data
		case Tick:
			if btl := arena.CurrentBattle(); btl != nil && btl.stage.Load() == BattleStageStart {
				btl.Tick(payload)
			}

		default:
			gamelog.L.Warn().Str("message_type", string(mt)).Str("msg", string(payload)).Err(err).Msg("Battle Arena WS: no message response")
		}
	}
}

func (arena *Arena) GameClientJsonDataParser() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic on GameClientJsonDataParser!", r)
		}
	}()
	for {
		data := <-arena.gameClientJsonDataChan
		msg := &BattleMsg{}
		err := json.Unmarshal(data, msg)
		if err != nil {
			gamelog.L.Warn().Str("msg", string(data)).Err(err).Msg("unable to unmarshal battle message")
			continue
		}

		btl := arena.CurrentBattle()
		if btl == nil && msg.BattleCommand != "BATTLE:OUTRO_FINISHED" {
			gamelog.L.Warn().Msg("current battle has already been cleaned up")
			continue
		}

		L := gamelog.L.With().RawJSON("game_client_data", data).Int("message_type", int(JSON)).Str("battleCommand", msg.BattleCommand).Logger()
		L.Info().Msg("game client message received")

		command := strings.TrimSpace(msg.BattleCommand) // temp fix for issue on gameclient
		switch command {
		case "BATTLE:MAP_DETAILS":
			var dataPayload *MapDetailsPayload
			if err = json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message payload")
				continue
			}

			// update map detail
			btl.storeGameMap(dataPayload.Details, dataPayload.BattleZones)

		case "BATTLE:START":
			var dataPayload *BattleStartPayload
			if err = json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message payload")
				continue
			}

			gameClientBuildNo, err := strconv.ParseUint(dataPayload.ClientBuildNo, 10, 64)
			if err != nil {
				L.Panic().Str("game_client_build_no", dataPayload.ClientBuildNo).Msg("invalid game client build number received")
			}

			if gameClientBuildNo < arena.Manager.gameClientMinimumBuildNo {
				L.Panic().Str("current_game_client_build", dataPayload.ClientBuildNo).Uint64("minimum_game_client_build", arena.Manager.gameClientMinimumBuildNo).Msg("unsupported game client build number")
			}

			err = btl.preIntro(dataPayload)
			if err != nil {
				L.Error().Msg("battle start load out has failed")
				return
			}

			arena.Manager.NewBattleChan <- &NewBattleChan{btl.ID, btl.BattleNumber}
		case "BATTLE:OUTRO_FINISHED":
			if btl.replaySession.ReplaySession != nil {
				err = replay.RecordReplayRequest(btl.Battle, btl.replaySession.ReplaySession.ID, replay.StopRecording)
				if err != nil {
					if err != replay.ErrDontLogRecordingStatus {
						gamelog.L.Error().Err(err).Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Msg("Failed to start recording")
					}
				}

				eventByte, err := json.Marshal(btl.replaySession.Events)
				if err != nil {
					gamelog.L.Error().Err(err).Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Interface("Events", btl.replaySession.Events).Msg("Failed to marshal json into battle replay")
				} else {
					btl.replaySession.ReplaySession.BattleEvents = null.JSONFrom(eventByte)
				}

				btl.replaySession.ReplaySession.StoppedAt = null.TimeFrom(time.Now())
				btl.replaySession.ReplaySession.RecordingStatus = boiler.RecordingStatusSTOPPED
				btl.replaySession.ReplaySession.IsCompleteBattle = true
				_, err = btl.replaySession.ReplaySession.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Err(err).Msg("Failed to update recording status to RECORDING while starting battle")
				}
			}

			// set idle is true
			arena.Stage.Store(ArenaStageIdle)

			// begin battle
			arena.BeginBattle()

		case "BATTLE:INTRO_FINISHED":
			if btl.replaySession.ReplaySession != nil {
				btl.replaySession.ReplaySession.IntroEndedAt = null.TimeFrom(time.Now())
			}
			btl.start()
		case "BATTLE:WAR_MACHINE_DESTROYED":
			// do not process, if battle already ended
			if btl.stage.Load() == BattleStageEnd {
				continue
			}

			var dataPayload BattleWMDestroyedPayload
			if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
				continue
			}
			btl.Destroyed(&dataPayload)
		case "BATTLE:END":
			var dataPayload *BattleEndPayload
			if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
				continue
			}
			btl.end(dataPayload)
		case "BATTLE:AI_SPAWNED":
			var dataPayload *AISpawnedRequest
			if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message payload")
				continue
			}
			err = btl.AISpawned(dataPayload)
			if err != nil {
				L.Error().Err(err).Msg("failed to spawn ai")
			}
		case "BATTLE:ABILITY_MOVE_COMMAND_COMPLETE":
			var dataPayload *AbilityMoveCommandCompletePayload
			if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal ability move command complete payload")
				continue
			}
			err = btl.CompleteWarMachineMoveCommand(dataPayload)
			if err != nil {
				L.Error().Err(err).Msg("failed update war machine move command")
			}
		case "BATTLE:ZONE_CHANGE":
			var dataPayload *ZoneChangePayload
			if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle zone change payload")
				continue
			}

			err = btl.ZoneChange(dataPayload)
			if err != nil {
				L.Error().Err(err).Msg("failed to zone change")
			}
		case "BATTLE:WAR_MACHINE_PICKUP":
			// do not process, if battle already ended
			if btl.stage.Load() == BattleStageEnd {
				continue
			}

			var dataPayload BattleWMPickupPayload
			if err = json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message repair pick up payload")
				continue
			}

			eventID := dataPayload.EventID

			// skip if ability not exists, or it is not a picked-up ability
			if da := btl.MiniMapAbilityDisplayList.Get(eventID); da == nil {
				continue
			}

			// remove repair from pending list, and broadcast
			ws.PublishMessage(
				fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", btl.ArenaID),
				server.HubKeyMiniMapAbilityDisplayList,
				btl.MiniMapAbilityDisplayList.Remove(dataPayload.EventID),
			)

		case "BATTLE:WAR_MACHINE_STATUS":
			// do not process, if battle already ended
			if btl.stage.Load() == BattleStageEnd {
				continue
			}

			var dataPayload *WarMachineStatusPayload
			if err = json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle zone change payload")
				continue
			}

			wm := arena.CurrentBattleWarMachineByHash(dataPayload.WarMachineHash)
			if wm == nil {
				continue
			}

			wm.Status.IsStunned = dataPayload.Status.IsStunned
			wm.Status.IsHacked = dataPayload.Status.IsHacked

			// EMP
			bpas, err := boiler.BlueprintPlayerAbilities(
				boiler.BlueprintPlayerAbilityWhere.GameClientAbilityID.IN([]int{12, 13}),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load player abilities")
				continue
			}

			for _, bpa := range bpas {
				offeringID := fmt.Sprintf("%s:%s", wm.ID, bpa.MechDisplayEffectType)
				mma := &MiniMapAbilityContent{
					OfferingID:               offeringID,
					LocationSelectType:       bpa.LocationSelectType,
					MiniMapDisplayEffectType: bpa.MiniMapDisplayEffectType,
					MechDisplayEffectType:    bpa.MechDisplayEffectType,
					Colour:                   bpa.Colour,
					ImageUrl:                 bpa.ImageURL,
					MechID:                   wm.ID,
				}
				switch bpa.GameClientAbilityID {
				case 12: // EMP
					if wm.Status.IsStunned {
						// add ability onto pending list, and broadcast
						ws.PublishMessage(
							fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", btl.ArenaID),
							server.HubKeyMiniMapAbilityDisplayList,
							btl.MiniMapAbilityDisplayList.Add(offeringID, mma),
						)
						continue
					}
					ws.PublishMessage(
						fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", btl.ArenaID),
						server.HubKeyMiniMapAbilityDisplayList,
						btl.MiniMapAbilityDisplayList.Remove(offeringID),
					)
				case 13: // HACKER DRONE
					if wm.Status.IsHacked {
						// add ability onto pending list, and broadcast
						ws.PublishMessage(
							fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", btl.ArenaID),
							server.HubKeyMiniMapAbilityDisplayList,
							btl.MiniMapAbilityDisplayList.Add(offeringID, mma),
						)
						continue
					}
					ws.PublishMessage(
						fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", btl.ArenaID),
						server.HubKeyMiniMapAbilityDisplayList,
						btl.MiniMapAbilityDisplayList.Remove(offeringID),
					)
				}
			}
		case "BATTLE:ABILITY_MOVE_COMMAND_RESPONSE":
			// do not process, if battle already ended
			if btl.stage.Load() == BattleStageEnd {
				continue
			}

			var dataPayload *MechMoveCommandResponsePayload
			if err = json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle zone change payload")
				continue
			}

			// send response to mech command check
			arena.MechCommandCheckMap.Send(dataPayload.EventID, dataPayload.IsValid)

		case "BATTLE:ABILITY_COMPLETE":
			// do not process, if battle already ended
			if btl.stage.Load() == BattleStageEnd {
				continue
			}

			var dataPayload *AbilityCompletePayload
			if err = json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle zone change payload")
				continue
			}

			eventID := dataPayload.EventID

			// skip if ability not exists, or it is a picked-up ability
			if da := btl.MiniMapAbilityDisplayList.Get(eventID); da == nil {
				continue
			}

			// remove ability from pending list, and broadcast
			ws.PublishMessage(
				fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", btl.ArenaID),
				server.HubKeyMiniMapAbilityDisplayList,
				btl.MiniMapAbilityDisplayList.Remove(eventID),
			)
		default:
			L.Warn().Err(err).Msg("Battle Arena WS: no command response")
		}
		L.Debug().Msg("game client message handled")
	}
}


// assignBattleLobby assign the next
// skipLobbyCheck ONLY happen on the battle end
func (arena *Arena) assignBattleLobby() {
	L := gamelog.L.With().Str("func", "assignBattleLobby").Str("arena id", arena.ID).Logger()
	battleLobbyID := arena.currentLobbyID.Load()

	//  check current lobby is valid
	if battleLobbyID != "" {
		bl, err := boiler.BattleLobbies(
			boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
			boiler.BattleLobbyWhere.ReadyAt.IsNotNull(),
			boiler.BattleLobbyWhere.EndedAt.IsNull(),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			L.Error().Err(err).Msg("Failed to load battle lobby")
		}

		// if assigned lobby is valid
		if bl != nil {
			arena.Stage.Store(ArenaStageProcessing)
			return
		}
	}

	// otherwise, clean up battle lobby id and assign a new one
	arena.currentLobbyID.Store("")

	// collect active battle lobbies' id
	battleLobbyIDs := arena.Manager.GetCurrentBattleLobbyIDs()

	L = L.With().Strs("battleLobbyIDs", battleLobbyIDs).Logger()

	// get the next valid battle lobby
	bl, err := db.GetNextBattleLobby(battleLobbyIDs)
	if err != nil {
		L.Error().Err(err).Msg("failed to get .")
	}

	L = L.With().Interface("bl", bl).Logger()

	// if no available lobby
	if bl == nil {
		L.Debug().Str("battle arena id", arena.ID).Msg("no lobby is available")
		bl, err = GenerateAIDrivenBattle()
		if err != nil {
			L.Error().Err(err).Msg("Failed to generate AI driven match.")
			return
		}
	}

	// assign battle lobby
	arena.currentLobbyID.Store(bl.ID)
	arena.Stage.Store(ArenaStageProcessing)

	//broadcast next lobby
	go func(){
		battleLobbyIDs := arena.Manager.GetCurrentBattleLobbyIDs()
		bl, err := db.GetNextBattleLobby(battleLobbyIDs)
		if err != nil {
			return
		}

		if bl == nil {
			return 
		}

		resp, err := server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl})
		if err != nil {
			return
		}

		if len(resp) != 1 {
			return
		}

		ws.PublishMessage("/public/upcoming_battle", server.HubKeyNextBattleDetails, resp[0])
	}()
}

func (arena *Arena) BeginBattle() {
	arena.beginBattleMux.Lock()
	defer arena.beginBattleMux.Unlock()

	// skip, if arena is not idle
	if arena.Stage.Load() != ArenaStageIdle {
		return
	}

	// delete all the unfinished mech command
	_, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.ArenaID.EQ(arena.ID),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
		boiler.MechMoveCommandLogWhere.DeletedAt.IsNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechMoveCommandLogColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to clean up unfinished mech move command")
	}

	arena.Manager.Lock()
	arena.assignBattleLobby()
	arena.Manager.Unlock()

	// return, if the stage of the arena is still idle
	if arena.Stage.Load() == ArenaStageIdle || arena.currentLobbyID.Load() == "" {
		return
	}

	gamelog.L.Trace().Str("func", "beginBattle").Msg("start")
	defer gamelog.L.Trace().Str("func", "beginBattle").Msg("end")

	battleLobby, err := boiler.FindBattleLobby(gamedb.StdConn, arena.currentLobbyID.Load())
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle lobby id", arena.currentLobbyID.Load()).Msg("Failed to load battle lobby")
		arena.Stage.Store(ArenaStageIdle)
		return
	}

	// if the previous battle does not end properly
	if battleLobby.AssignedToBattleID.Valid {
		// clean up unfinished battle
		cleanUpBattleRecord(battleLobby.AssignedToBattleID.String)
	}

	// create new battle
	var gameMap *boiler.GameMap

	// generate game map for the lobby, if there isn't one
	if !battleLobby.GameMapID.Valid {
		var gms []*boiler.GameMap
		gms, err = boiler.GameMaps().All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to load game maps.")
			arena.Stage.Store(ArenaStageIdle)
			return
		}

		if gms == nil {
			gamelog.L.Warn().Str("log_name", "battle arena").Msg("No available game maps in db.")
			arena.Stage.Store(ArenaStageIdle)
			return
		}

		// randomly assign game map to battle lobby
		rand.Seed(time.Now().UnixNano())
		gameMap = gms[rand.Intn(len(gms))]

		battleLobby.GameMapID = null.StringFrom(gameMap.ID)
		_, err = battleLobby.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleLobbyColumns.GameMapID))
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("game map", gameMap).Interface("battle lobby", battleLobby).Err(err).Msg("Failed to assign game map to battle lobby.")
			arena.Stage.Store(ArenaStageIdle)
			return
		}
	} else {
		// load game map from last battle
		gameMap, err = battleLobby.GameMap().One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("battle lobby", battleLobby).Msg("Failed to load game map from battle lobby")
			arena.Stage.Store(ArenaStageIdle)
			return
		}
	}

	// insert new battle
	battle := &boiler.Battle{
		ID:        uuid.Must(uuid.NewV4()).String(),
		GameMapID: gameMap.ID,
		StartedAt: time.Now(),
		ArenaID:   arena.ID,
	}
	err = battle.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("battle", battle).Msg("unable to insert Battle into database")
		arena.Stage.Store(ArenaStageIdle)
		return
	}

	// insert battle lobby
	battleLobby.AssignedToBattleID = null.StringFrom(battle.ID)
	battleLobby.AssignedToArenaID = null.StringFrom(arena.ID)
	_, err = battleLobby.Update(gamedb.StdConn,
		boil.Whitelist(
			boiler.BattleLobbyColumns.AssignedToBattleID,
			boiler.BattleLobbyColumns.AssignedToArenaID,
		),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("battle lobby", battleLobby).Msg("unable to update battle lobby")
		arena.Stage.Store(ArenaStageIdle)
		return
	}

	// update assigned battle id
	_, err = battleLobby.BattleLobbiesMechs().UpdateAll(gamedb.StdConn, boiler.M{boiler.BattleLobbiesMechColumns.AssignedToBattleID: null.StringFrom(battle.ID)})
	if err != nil {
		gamelog.L.Error().Interface("battle lobby", battleLobby).Str("db func", "Battle").Err(err).Msg("unable to update battle id of battle lobby mech")
		arena.Stage.Store(ArenaStageIdle)
		return
	}

	// broadcast battle lobby change
	go BroadcastBattleLobbyUpdate(battleLobby.ID)

	btl := &Battle{
		arena:   arena,
		Battle:  battle,
		MapName: gameMap.Name,
		lobby:   battleLobby,
		gameMap: &server.GameMap{
			ID:            uuid.FromStringOrNil(gameMap.ID),
			Name:          gameMap.Name,
			BackgroundUrl: gameMap.BackgroundURL,
		},
		stage:                  atomic.NewInt32(BattleStageStart),
		destroyedWarMachineMap: make(map[string]*WMDestroyedRecord),
		MiniMapAbilityDisplayList: &MiniMapAbilityDisplayList{
			m: make(map[string]*MiniMapAbilityContent),
		},
		MapEventList: NewMapEventList(gameMap.Name),
		replaySession: &RecordingSession{
			ReplaySession: &boiler.BattleReplay{
				ArenaID:         arena.ID,
				BattleID:        battle.ID,
				RecordingStatus: boiler.RecordingStatusIDLE,
			},
			Events: []*RecordingEvents{},
		},
	}

	// load war machines
	err = btl.Load(battleLobby)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle mechs")
		return
	}

	playerMechMap := make(map[string][]string)
	for _, wm := range btl.WarMachines {
		pm, ok := playerMechMap[wm.OwnedByID]
		if !ok {
			pm = []string{}
		}
		pm = append(pm, wm.ID)
		playerMechMap[wm.OwnedByID] = pm

		// check mech join battle quest for each mech owner
		arena.Manager.QuestManager.MechJoinBattleQuestCheck(wm.OwnedByID)
	}

	// broadcast mech status change
	for playerID, mechIDs := range playerMechMap {
		go BroadcastMechQueueStatus(playerID, mechIDs...)
	}

	al, err := db.AbilityLabelList()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load ability labels")
	}
	if al != nil && len(al) > 0 {
		// Indexes correspond to the game_client_ability_id in the db
		// NOTE: game_client_ability_id start from 0
		btl.abilityDetails = make([]*AbilityDetail, al[0].GameClientAbilityID+1)

		for _, a := range al {
			switch a.GameClientAbilityID {
			case 1: // NUKE
				btl.abilityDetails[a.GameClientAbilityID] = &AbilityDetail{
					Radius: 2000,
				}
			case 12: // EMP
				btl.abilityDetails[a.GameClientAbilityID] = &AbilityDetail{
					Radius: 10000,
				}
			case 16: // BLACKOUT
				btl.abilityDetails[a.GameClientAbilityID] = &AbilityDetail{
					Radius: 20000,
				}
			}
		}
	}

	// start battle record
	func() {
		// insert
		err = btl.replaySession.ReplaySession.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to insert new battle replay")
			return
		}

		// url request
		err = replay.RecordReplayRequest(btl.Battle, btl.replaySession.ReplaySession.ID, replay.StartRecording)
		if err != nil {
			if err != replay.ErrDontLogRecordingStatus {
				gamelog.L.Error().Err(err).Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Msg("Failed to start recording")
				return
			}
			return
		}

		// update start time
		btl.replaySession.ReplaySession.StartedAt = null.TimeFrom(time.Now())
		btl.replaySession.ReplaySession.RecordingStatus = boiler.RecordingStatusRECORDING
		_, err = btl.replaySession.ReplaySession.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("battle_id", btl.ID).Str("replay_id", btl.replaySession.ReplaySession.ID).Err(err).Msg("Failed to update recording status to RECORDING while starting battle")
			return
		}
	}()

	gamelog.L.Debug().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up incognito manager")
	btl.storePlayerAbilityManager(NewPlayerAbilityManager())

	arena.storeCurrentBattle(btl)

	arena.Message(BATTLEINIT, &struct {
		BattleID     string                  `json:"battle_id"`
		MapName      string                  `json:"map_name"`
		BattleNumber int                     `json:"battle_number"`
		WarMachines  []*WarMachineGameClient `json:"war_machines"`
	}{
		BattleID:     btl.ID,
		MapName:      btl.MapName,
		WarMachines:  WarMachinesToClient(btl.WarMachines),
		BattleNumber: battle.BattleNumber,
	})

	// broadcast system message for mechs entering the battle
	go btl.BattleStartSystemMessage()

	go arena.NotifyUpcomingWarMachines()
}

type SystemMessageBattleStart struct {
	PlayerID  string               `json:"player_id"`
	FactionID string               `json:"faction_id"`
	Mechs     []*SystemMessageMech `json:"mechs"`
}

type SystemMessageMech struct {
	MechID        string `json:"mech_id"`
	FactionID     string `json:"faction_id"`
	Name          string `json:"name"`
	ImageUrl      string `json:"image_url"`
	Tier          string `json:"tier"`
	TotalBlocks   int    `json:"total_blocks"`
	DamagedBlocks int    `json:"damaged_blocks"`
}

func (btl *Battle) BattleStartSystemMessage() {
	l := gamelog.L.With().Str("func", "BattleStartSystemMessage").Logger()

	playerMechs := make(map[string][]*WarMachine)
	for _, wm := range btl.WarMachines {
		pm, ok := playerMechs[wm.OwnedByID]
		if !ok {
			pm = []*WarMachine{}
		}
		playerMechs[wm.OwnedByID] = append(pm, wm)
	}

	for playerID, mechs := range playerMechs {
		go func(playerID string, mechs []*WarMachine) {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				l.Error().Err(err).Msg("unable to begin tx")
				return
			}
			defer tx.Rollback()

			bsm := &SystemMessageBattleStart{
				PlayerID: playerID,
				Mechs:    []*SystemMessageMech{},
			}

			for _, mech := range mechs {
				smm := &SystemMessageMech{
					MechID:        mech.ID,
					FactionID:     mech.Faction.ID,
					Name:          mech.Label,
					ImageUrl:      mech.ImageAvatar,
					Tier:          mech.Tier,
					TotalBlocks:   db.TotalRepairBlocks(mech.ID),
					DamagedBlocks: mech.damagedBlockCount,
				}

				if mech.Name != "" {
					smm.Name = mech.Name
				}

				bsm.Mechs = append(bsm.Mechs, smm)
			}

			b, err := json.Marshal(bsm)
			if err != nil {
				l.Error().Err(err).Interface("battle start data", bsm).Msg("failed to marshal battle start system message data")
				return
			}

			message := fmt.Sprintf("Your mech enters battle #%d", btl.BattleNumber)
			if len(mechs) > 1 {
				message = fmt.Sprintf("Your mechs enter battle #%d", btl.BattleNumber)
			}

			msg := &boiler.SystemMessage{
				PlayerID: playerID,
				SenderID: server.SupremacyBattleUserID,
				DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeMechBattleBegin)),
				Title:    "Battle Begin",
				Message:  message,
				Data:     null.JSONFrom(b),
			}
			err = msg.Insert(tx, boil.Infer())
			if err != nil {
				l.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new system message into db")
				return
			}

			err = tx.Commit()
			if err != nil {
				l.Error().Err(err).Msg("failed to commit transaction")
				return
			}

			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", playerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
		}(playerID, mechs)
	}
}

func (arena *Arena) UserStatUpdatedSubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	us, err := db.UserStatsGet(userID.String())
	if err != nil {
		return terror.Error(err, "failed to get user stats")
	}

	if us != nil {
		reply(us)
	}

	return nil
}

func (btl *Battle) IsMechOfType(participantID int, aiType AIType) bool {
	btl.spawnedAIMux.RLock()
	defer btl.spawnedAIMux.RUnlock()

	for _, s := range btl.SpawnedAI {
		if int(s.ParticipantID) != participantID {
			continue
		}
		return *s.AIType == aiType
	}
	return false
}

func (btl *Battle) AISpawned(payload *AISpawnedRequest) error {
	// check battle id
	if payload.BattleID != btl.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.ID, payload.BattleID))
	}

	// get spawned AI
	spawnedAI := &WarMachine{
		ParticipantID: payload.ParticipantID,
		Hash:          payload.Hash,
		OwnedByID:     payload.UserID,
		Name:          payload.Name,
		Model:         payload.Model,
		Skin:          payload.Skin,
		MaxHealth:     payload.MaxHealth,
		Health:        payload.MaxHealth,
		MaxShield:     payload.MaxShield,
		Shield:        payload.MaxShield,
		FactionID:     payload.FactionID,
		Position:      payload.Position,
		Rotation:      payload.Rotation,
		Image:         "https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-mini-mech.png",
		ImageAvatar:   "https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-mini-mech.png",
		AIType:        &payload.Type,
		Status:        &Status{},
	}

	gamelog.L.Info().Msgf("Battle Update: %s - AI Spawned: %d", payload.BattleID, spawnedAI.ParticipantID)

	btl.spawnedAIMux.Lock()
	defer btl.spawnedAIMux.Unlock()

	// cache record in battle, for future subscription
	btl.SpawnedAI = append(btl.SpawnedAI, spawnedAI)

	// Broadcast spawn event
	ws.PublishMessage(fmt.Sprintf("/public/arena/%s/minimap", btl.ArenaID), HubKeyBattleAISpawned, btl.SpawnedAI)

	return nil
}

func (btl *Battle) CompleteWarMachineMoveCommand(payload *AbilityMoveCommandCompletePayload) error {
	gamelog.L.Trace().Str("func", "UpdateWarMachineMoveCommand").Msg("start")
	defer gamelog.L.Trace().Str("func", "UpdateWarMachineMoveCommand").Msg("end")

	if payload.BattleID != btl.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.ID, payload.BattleID))
	}

	// check battle state
	if btl.arena.currentBattleState() == BattleStageEnd {
		return terror.Error(fmt.Errorf("current battle is ended"))
	}

	// get mech
	wm := btl.arena.CurrentBattleWarMachineOrAIByHash(payload.WarMachineHash)
	if wm == nil {
		return terror.Error(fmt.Errorf("war machine not exists"))
	}

	isMiniMech := wm.AIType != nil && *wm.AIType == MiniMech
	if !isMiniMech {
		// get the last move command of the mech
		mmc, err := boiler.MechMoveCommandLogs(
			boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
			boiler.MechMoveCommandLogWhere.BattleID.EQ(btl.ID),
			qm.OrderBy(boiler.MechMoveCommandLogColumns.CreatedAt+" DESC"),
		).One(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to get mech move command from db.")
		}

		// update completed_at
		mmc.ReachedAt = null.TimeFrom(time.Now())
		mmc.IsMoving = false
		_, err = mmc.Update(gamedb.StdConn, boil.Whitelist(boiler.MechMoveCommandLogColumns.ReachedAt))
		if err != nil {
			return terror.Error(err, "Failed to update mech move command")
		}

		ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_command/%s", wm.FactionID, btl.ArenaID, wm.Hash), server.HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
			MechMoveCommandLog: mmc,
		})
	} else {
		mmmc, err := btl.arena._currentBattle.playerAbilityManager().CompleteMiniMechMove(wm.Hash)
		if err == nil && mmmc != nil {
			ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_command/%s", wm.FactionID, btl.ArenaID, wm.Hash), server.HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
				MechMoveCommandLog: &boiler.MechMoveCommandLog{
					ID:            fmt.Sprintf("%s_%s", mmmc.BattleID, mmmc.MechHash),
					BattleID:      mmmc.BattleID,
					MechID:        mmmc.MechHash,
					TriggeredByID: mmmc.TriggeredByID,
					CellX:         mmmc.CellX,
					CellY:         mmmc.CellY,
					CancelledAt:   mmmc.CancelledAt,
					ReachedAt:     mmmc.ReachedAt,
					CreatedAt:     mmmc.CreatedAt,
					IsMoving:      mmmc.IsMoving,
				},
				IsMiniMech: true,
			})
		}
	}

	err := btl.arena.BroadcastFactionMechCommands(wm.FactionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to broadcast faction mech commands")
	}
	return nil
}

func (btl *Battle) ZoneChange(payload *ZoneChangePayload) error {
	// check battle id
	if payload.BattleID != btl.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.ID, payload.BattleID))
	}

	if btl.battleZones == nil {
		return terror.Error(fmt.Errorf("recieved battle zone change when battleZones is empty"))
	}

	if payload.ZoneIndex <= -1 || payload.ZoneIndex >= len(btl.battleZones) {
		return terror.Error(fmt.Errorf("invalid zone index"))
	}

	// Update current battle zone
	btl.currentBattleZoneIndex = payload.ZoneIndex

	// Send notification to frontend
	currentZone := btl.battleZones[payload.ZoneIndex]
	event := ZoneChangeEvent{
		Location:   currentZone.Location,
		Radius:     currentZone.Radius,
		ShrinkTime: currentZone.ShrinkTime,
		WarnTime:   payload.WarnTime,
	}
	btl.arena.BroadcastGameNotificationBattleZoneChange(&event)

	return nil
}

func ReversePlayerAbilities(battleID string, battleNumber int) {
	cas, err := boiler.ConsumedAbilities(
		boiler.ConsumedAbilityWhere.BattleID.EQ(battleID),
		qm.Load(boiler.ConsumedAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to query consumed abilities")
		return
	}

	if cas != nil {
		type refundBlueprintAbility struct {
			Amount        int                            `json:"amount"`
			PlayerAbility *boiler.BlueprintPlayerAbility `json:"player_ability"`
		}
		playerAbilityRefundMap := make(map[string][]*refundBlueprintAbility)
		for _, ca := range cas {
			err = db.PlayerAbilityAssign(ca.ConsumedBy, ca.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("player id", ca.ConsumedBy).Str("ability blueprint id", ca.BlueprintID).Msg("Failed to refund ability to the player")
				continue
			}

			pas, ok := playerAbilityRefundMap[ca.ConsumedBy]
			if !ok {
				pas = []*refundBlueprintAbility{}
			}

			index := slices.IndexFunc(pas, func(rba *refundBlueprintAbility) bool {
				return rba.PlayerAbility.ID == ca.BlueprintID
			})

			if index != -1 {
				// increase amount if already exist
				pas[index].Amount += 1
			} else {
				// otherwise append new ability to the list
				pas = append(pas, &refundBlueprintAbility{
					Amount:        1,
					PlayerAbility: ca.R.Blueprint,
				})
			}

			// assign the list bach to the map
			playerAbilityRefundMap[ca.ConsumedBy] = pas
		}

		_, err = cas.DeleteAll(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to delete consumed ability record.")
		}

		// send system message
		for playerID, par := range playerAbilityRefundMap {
			// get all the blueprint ability ids
			bpIDs := []string{}
			for _, pa := range par {
				bpIDs = append(bpIDs, pa.PlayerAbility.ID)
			}

			// reverse ability cool down
			_, err = boiler.PlayerAbilities(
				boiler.PlayerAbilityWhere.BlueprintID.IN(bpIDs),
				boiler.PlayerAbilityWhere.OwnerID.EQ(playerID),
			).UpdateAll(gamedb.StdConn, boiler.M{boiler.PlayerAbilityColumns.CooldownExpiresOn: time.Now()})
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Strs("blueprint player ability IDs", bpIDs).Msg("failed to update player ability cool down expiry.")
			}

			// broadcast current player list
			pas, err := db.PlayerAbilitiesList(playerID)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("boiler func", "PlayerAbilities").Str("ownerID", playerID).Err(err).Msg("unable to get player abilities")
			}
			if pas != nil {
				ws.PublishMessage(fmt.Sprintf("/secure/user/%s/player_abilities", playerID), server.HubKeyPlayerAbilitiesList, pas)
			}

			// send battle reward system message
			b, err := json.Marshal(par)
			if err != nil {
				gamelog.L.Error().Interface("player abilities", par).Err(err).Msg("Failed to marshal player refund data into json.")
				break
			}
			sysMsg := boiler.SystemMessage{
				PlayerID: playerID,
				SenderID: server.SupremacyBattleUserID,
				DataType: null.StringFrom(string(system_messages.SystemMessageDataTypePlayerAbilityRefunded)),
				Title:    "Player Abilities Refunded",
				Message:  fmt.Sprintf("Due to battle #%d being restarted, the consumed player abilities have been returned.", battleNumber),
				Data:     null.JSONFrom(b),
			}
			err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
				break
			}
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", playerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
		}
	}
}
