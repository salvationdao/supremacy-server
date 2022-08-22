package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"net"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/quest"
	"server/system_messages"
	"server/telegram"
	"server/xsyn_rpcclient"
	"strconv"
	"sync"
	"time"

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
	RepairOfferFuncChan      chan func()
	QuestManager             *quest.System

	arenas       map[string]*Arena
	sync.RWMutex // lock for arena
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

func NewArenaManager(opts *Opts) *ArenaManager {
	am := &ArenaManager{
		Addr:                     opts.Addr,
		timeout:                  opts.Timeout,
		RPCClient:                opts.RPCClient,
		sms:                      opts.SMS,
		telegram:                 opts.Telegram,
		gameClientMinimumBuildNo: opts.GameClientMinimumBuildNo,
		SystemBanManager:         NewSystemBanManager(),
		NewBattleChan:            make(chan *NewBattleChan, 10),
		SystemMessagingManager:   opts.SystemMessagingManager,
		RepairOfferFuncChan:      make(chan func()),
		QuestManager:             opts.QuestManager,
		arenas:                   make(map[string]*Arena),
	}

	am.server = &http.Server{
		Handler:      am,
		ReadTimeout:  am.timeout,
		WriteTimeout: am.timeout,
	}

	// start player rank updater
	am.PlayerRankUpdater()
	go am.RepairOfferCleaner()

	return am
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

func (am *ArenaManager) Range(fn func(arena *Arena)) {
	am.RLock()
	defer am.RUnlock()
	for _, arena := range am.arenas {
		fn(arena)
	}
}

func (am *ArenaManager) DeleteArena(arenaID string) {
	am.Lock()
	defer am.Unlock()
	delete(am.arenas, arenaID)
}

func (am *ArenaManager) GetArena(arenaID string) (*Arena, error) {
	am.RLock()
	defer am.RUnlock()
	arena, ok := am.arenas[arenaID]
	if !ok {
		return nil, terror.Error(fmt.Errorf("arena not exits"), "The battle arena does not exist.")
	}
	if !arena.connected.Load() {
		return nil, terror.Error(fmt.Errorf("arena not available"), "The battle arena is not available")
	}

	return arena, nil
}

func (am *ArenaManager) AvailableBattleArenas() []*boiler.BattleArena {
	am.RLock()
	defer am.RUnlock()

	resp := []*boiler.BattleArena{}
	for _, arena := range am.arenas {
		if arena.connected.Load() {
			resp = append(resp, arena.BattleArena)
		}
	}
	return resp
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
	arena, err := am.NewArena(wsConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to add arena onto arena manager")
		return
	}

	// broadcast a new arena list to frontend
	ws.PublishMessage("/public/arena_list", server.HubKeyBattleArenaListSubscribe, am.AvailableBattleArenas())

	defer func() {
		if wsConn != nil {
			arena.connected.Store(false)

			// tell frontend the arena is closed
			ws.PublishMessage(fmt.Sprintf("/public/arena/%s/closed", arena.ID), server.HubKeyBattleArenaClosedSubscribe, true)

			gamelog.L.Error().Err(fmt.Errorf("game client has disconnected")).Msg("lost connection to game client")
			err = wsConn.Close(websocket.StatusInternalError, "game client has disconnected")
			if err != nil {
				gamelog.L.Error().Str("arena id", arena.ID).Err(err).Msg("Failed to close ws connection")
			}

			btl := arena.CurrentBattle()
			if btl != nil {
				btl.endAbilities()
			}

			// delete arena from the map
			am.DeleteArena(arena.ID)

			// broadcast a new arena list to frontend
			ws.PublishMessage("/public/arena_list", server.HubKeyBattleArenaListSubscribe, am.AvailableBattleArenas())
		}
	}()

	arena.Start()
}

func (am *ArenaManager) NewArena(wsConn *websocket.Conn) (*Arena, error) {
	am.Lock()
	defer am.Unlock()

	var ba *boiler.BattleArena
	var err error
	existingArenaID := []string{}
	storyArenaExist := false
	for key, a := range am.arenas {
		// clean up any disconnected arena from the map
		if !a.connected.Load() {
			delete(am.arenas, key)
			continue
		}

		existingArenaID = append(existingArenaID, a.ID)
		if a.Type == boiler.ArenaTypeEnumSTORY {
			storyArenaExist = true
		}
	}

	// if story mode not exist, assign story mode arena
	if !storyArenaExist {
		ba, err = boiler.BattleArenas(
			boiler.BattleArenaWhere.Type.EQ(boiler.ArenaTypeEnumSTORY),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to get story mode battle arena from db")
			return nil, terror.Error(err, "Failed to get story mode battle arena from db")
		}
	} else {
		// assign an expedition arena instead
		ba, err = boiler.BattleArenas(
			boiler.BattleArenaWhere.Type.EQ(boiler.ArenaTypeEnumEXPEDITION),
			boiler.BattleArenaWhere.ID.NIN(existingArenaID),
			qm.OrderBy(boiler.BattleArenaColumns.Gid),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("Failed to get expedition mode battle arena from db")
			return nil, terror.Error(err, "Failed to get expedition mode battle arena from db")
		}

		// insert a new expedition battle arena if not exist
		if ba == nil {
			gamelog.L.Debug().Msg("inserting a new arena to db")
			ba = &boiler.BattleArena{
				Type: boiler.ArenaTypeEnumEXPEDITION,
			}

			err = ba.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to insert new expedition battle arena into db")
				return nil, terror.Error(err, "Failed to insert new expedition battle arena into db")
			}
		}
	}

	arena := &Arena{
		BattleArena:            ba,
		socket:                 wsConn,
		connected:              atomic.NewBool(true),
		gameClientJsonDataChan: make(chan []byte, 3),

		// objects inherited from arena manager
		RPCClient:                am.RPCClient,
		sms:                      am.sms,
		gameClientMinimumBuildNo: am.gameClientMinimumBuildNo,
		telegram:                 am.telegram,
		timeout:                  am.timeout,
		SystemBanManager:         am.SystemBanManager,
		SystemMessagingManager:   am.SystemMessagingManager,
		NewBattleChan:            am.NewBattleChan,
		QuestManager:             am.QuestManager,
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

type Arena struct {
	*boiler.BattleArena
	socket                   *websocket.Conn
	connected                *atomic.Bool
	timeout                  time.Duration
	_currentBattle           *Battle
	AIPlayers                map[string]db.PlayerWithFaction
	RPCClient                *xsyn_rpcclient.XsynXrpcClient
	sms                      server.SMS
	gameClientMinimumBuildNo uint64
	telegram                 server.Telegram
	SystemBanManager         *SystemBanManager
	SystemMessagingManager   *system_messages.SystemMessagingManager
	NewBattleChan            chan *NewBattleChan

	QuestManager *quest.System

	gameClientJsonDataChan chan []byte
	sync.RWMutex
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

func (arena *Arena) currentBattleWarMachineIDs(factionIDs ...string) []uuid.UUID {
	arena.RLock()
	defer arena.RUnlock()

	ids := []uuid.UUID{}

	if arena._currentBattle == nil {
		return ids
	}

	if factionIDs != nil && len(factionIDs) > 0 {
		// only return war machines' id from the faction
		for _, wm := range arena._currentBattle.WarMachines {
			if wm.FactionID == factionIDs[0] {
				ids = append(ids, uuid.FromStringOrNil(wm.ID))
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
		Command string      `json:"battleCommand"`
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
	gamelog.L.Info().Str("message data", string(b)).Msg("game client message sent")
}

type QueueDefaultMechReq struct {
	factionID string    // faction of the mech
	queuedAt  time.Time // the time of queue
	amount    int       // amount of mechs should queue
}

func (btl *Battle) GenerateDefaultQueueRequest(bqs []*boiler.BattleQueue) map[string]*QueueDefaultMechReq {
	reqMap := make(map[string]*QueueDefaultMechReq)
	reqMap[server.RedMountainFactionID] = &QueueDefaultMechReq{
		factionID: server.RedMountainFactionID,
		queuedAt:  time.Now(),
		amount:    3,
	}
	reqMap[server.BostonCyberneticsFactionID] = &QueueDefaultMechReq{
		factionID: server.BostonCyberneticsFactionID,
		queuedAt:  time.Now(),
		amount:    3,
	}
	reqMap[server.ZaibatsuFactionID] = &QueueDefaultMechReq{
		factionID: server.ZaibatsuFactionID,
		queuedAt:  time.Now(),
		amount:    3,
	}

	for _, bq := range bqs {
		req, ok := reqMap[bq.FactionID]
		if ok {
			req.amount -= 1
			if bq.QueuedAt.Before(req.queuedAt) {
				req.queuedAt = bq.QueuedAt.Add(-1 * time.Second)
			}
		}
	}

	if server.IsProductionEnv() {
		// get maximum queue number
		maxNum := 0
		for _, req := range reqMap {
			if req.amount > maxNum {
				maxNum = req.amount
			}
		}

		// set max amount default mech to all the factions
		for _, req := range reqMap {
			req.amount = maxNum
		}
	}

	return reqMap
}

func (btl *Battle) QueueDefaultMechs(queueReqMap map[string]*QueueDefaultMechReq) error {
	defMechs, err := db.DefaultMechs()
	if err != nil {
		return err
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("unable to begin tx")
		return fmt.Errorf(terror.Echo(err))
	}

	defer tx.Rollback()

	for _, mech := range defMechs {
		qr, ok := queueReqMap[mech.FactionID.String]
		if !ok || qr.amount == 0 {
			continue
		}

		mech.Name = helpers.GenerateStupidName()
		mechToUpdate := boiler.Mech{
			ID:   mech.ID,
			Name: mech.Name,
		}
		_, _ = mechToUpdate.Update(tx, boil.Whitelist(boiler.MechColumns.Name))

		existMech, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mech.ID)).One(tx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mech.ID).Err(err).Msg("check mech exists in queue")
			return terror.Error(err, "Failed to check whether mech is in the battle queue")
		}

		if existMech != nil {
			continue
		}

		bq := &boiler.BattleQueue{
			MechID:    mech.ID,
			QueuedAt:  qr.queuedAt,
			FactionID: qr.factionID,
			OwnerID:   mech.OwnerID,
		}

		err = bq.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Interface("mech", mech).
				Err(err).Msg("unable to insert mech into queue")
			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		qr.amount -= 1

		// pay queue fee from treasury when it is not in production
		if !server.IsProductionEnv() {
			amount := db.GetDecimalWithDefault(db.KeyBattleQueueFee, decimal.New(250, 18))

			bqf := &boiler.BattleQueueFee{
				MechID:   mech.ID,
				PaidByID: mech.OwnerID,
				Amount:   amount,
			}

			err = bqf.Insert(tx, boil.Infer())
			if err != nil {
				return terror.Error(err, "Failed to insert battle queue fee.")
			}

			paidTxID, err := btl.arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.UUID(server.XsynTreasuryUserID), // paid from treasury fund
				ToUserID:             uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
				Amount:               amount.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("battle_queue_fee|%s|%d", mech.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          "queue mech to join the battle arena.",
			})
			if err != nil {
				gamelog.L.Error().
					Str("faction_user_id", mech.OwnerID).
					Str("mech id", mech.ID).
					Str("amount", amount.StringFixed(0)).
					Err(err).Msg("Failed to pay sups on queuing default mech.")
				return terror.Error(err, "Failed to pay sups on queuing mech.")
			}

			refundFunc := func() {
				_, err = btl.arena.RPCClient.RefundSupsMessage(paidTxID)
				if err != nil {
					gamelog.L.Error().
						Str("player_id", server.XsynTreasuryUserID.String()).
						Str("mech id", mech.ID).
						Str("amount", amount.StringFixed(0)).
						Err(err).Msg("Failed to refund sups on queuing default mech.")
				}
			}

			bq.QueueFeeTXID = null.StringFrom(paidTxID)
			bq.FeeID = null.StringFrom(bqf.ID)
			_, err = bq.Update(tx, boil.Whitelist(boiler.BattleQueueColumns.QueueFeeTXID, boiler.BattleQueueColumns.FeeID))
			if err != nil {
				refundFunc() // refund player
				gamelog.L.Error().Err(err).Msg("Failed to record queue fee tx id")
				return terror.Error(err, "Failed to update battle queue")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Err(err).Msg("unable to commit mech insertion into queue")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	return nil
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

	reply(nil)

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
				Description:            ga.Description,
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

	if arena.currentBattleState() != BattleStageStart {
		return nil
	}

	bn := arena.currentBattleNumber()
	if bn == -1 {
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
		lastTrigger, err := boiler.MechAbilityTriggerLogs(
			boiler.MechAbilityTriggerLogWhere.MechID.EQ(wm.ID),
			boiler.MechAbilityTriggerLogWhere.GameAbilityID.EQ(mechAbilityID),
			boiler.MechAbilityTriggerLogWhere.BattleNumber.EQ(bn),
			boiler.MechAbilityTriggerLogWhere.DeletedAt.IsNull(),
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
		lastTrigger, err := boiler.MechAbilityTriggerLogs(
			boiler.MechAbilityTriggerLogWhere.MechID.EQ(wm.ID),
			boiler.MechAbilityTriggerLogWhere.GameAbilityID.EQ(mechAbilityID),
			boiler.MechAbilityTriggerLogWhere.CreatedAt.GT(time.Now().Add(time.Duration(-coolDownSeconds)*time.Second)),
			boiler.MechAbilityTriggerLogWhere.DeletedAt.IsNull(),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("mech id", wm.ID).Str("game ability id", mechAbilityID).Err(err).Msg("Failed to get mech ability trigger from db")
			return terror.Error(err, "Failed to load game ability")
		}

		if lastTrigger != nil {
			reply(coolDownSeconds - int(time.Now().Sub(lastTrigger.CreatedAt).Seconds()))
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
		reply(WarMachineStat{
			ParticipantID: participantID,
			Position:      wm.Position,
			Rotation:      wm.Rotation,
			Health:        wm.Health,
			Shield:        wm.Shield,
			IsHidden:      wm.IsHidden,
		})
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
	BattleCommand string          `json:"battleCommand"`
	Payload       json.RawMessage `json:"payload"`
}

type BattleStartPayload struct {
	WarMachines []struct {
		Hash          string `json:"hash"`
		ParticipantID byte   `json:"participantID"`
	} `json:"warMachines"`
	BattleID      string `json:"battleID"`
	ClientBuildNo string `json:"clientBuildNo"`
}

type MapDetailsPayload struct {
	Details     server.GameMap      `json:"details"`
	BattleZones []server.BattleZone `json:"battleZones"`
	BattleID    string              `json:"battleID"`
}

type BattleEndPayload struct {
	WinningWarMachines []struct {
		Hash   string `json:"hash"`
		Health int    `json:"health"`
	} `json:"winningWarMachines"`
	BattleID     string `json:"battleID"`
	WinCondition string `json:"winCondition"`
}

type AbilityMoveCommandCompletePayload struct {
	BattleID       string `json:"battleID"`
	WarMachineHash string `json:"warMachineHash"`
}

type ZoneChangePayload struct {
	BattleID  string `json:"battleID"`
	ZoneIndex int    `json:"zoneIndex"`
	WarnTime  int    `json:"warnTime"`
}

type ZoneChangeEvent struct {
	Location   server.GameLocation `json:"location"`
	Radius     int                 `json:"radius"`
	ShrinkTime int                 `json:"shrinkTime"`
	WarnTime   int                 `json:"warnTime"`
}

type BattleWMDestroyedPayload struct {
	DestroyedWarMachineEvent struct {
		DestroyedWarMachineHash string    `json:"destroyedWarMachineHash"`
		KillByWarMachineHash    string    `json:"killByWarMachineHash"`
		RelatedEventIDString    string    `json:"relatedEventIDString"`
		RelatedEventID          uuid.UUID `json:"RelatedEventID"`
		DamageHistory           []struct {
			Amount         int    `json:"amount"`
			InstigatorHash string `json:"instigatorHash"`
			SourceHash     string `json:"sourceHash"`
			SourceName     string `json:"sourceName"`
		} `json:"damageHistory"`
		KilledBy      string `json:"killedBy"`
		ParticipantID int    `json:"participantID"`
	} `json:"destroyedWarMachineEvent"`
	BattleID string `json:"battleID"`
}

type AISpawnedRequest struct {
	BattleID       string          `json:"battleID"`
	SpawnedAIEvent *SpawnedAIEvent `json:"spawnedAIEvent"`
}

type AIType string

const (
	Reinforcement AIType = "Reinforcement"
	MiniMech      AIType = "Mini Mech"
	RobotDog      AIType = "Robot Dog"
)

type SpawnedAIEvent struct {
	ParticipantID byte            `json:"participantID"`
	Hash          string          `json:"hash"`
	UserID        string          `json:"userID"`
	Name          string          `json:"name"`
	Model         string          `json:"model"`
	Skin          string          `json:"skin"`
	MaxHealth     uint32          `json:"maxHealth"`
	Health        uint32          `json:"health"`
	MaxShield     uint32          `json:"maxShield"`
	Shield        uint32          `json:"shield"`
	FactionID     string          `json:"factionID"`
	Position      *server.Vector3 `json:"position"`
	Rotation      int             `json:"rotation"`
	Type          AIType          `json:"type"`
}

type BattleWMPickupPayload struct {
	WarMachineHash string `json:"warMachineHash"`
	EventID        string `json:"eventID"`
	BattleID       string `json:"battleID"`
}

func (arena *Arena) start() {
	ctx := context.Background()
	arena.beginBattle()

	for {
		_, payload, err := arena.socket.Read(ctx)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("empty game client disconnected")
			break
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

		L := gamelog.L.With().Str("game_client_data", string(data)).Int("message_type", int(JSON)).Str("battleCommand", msg.BattleCommand).Logger()
		L.Info().Msg("game client message received")

		switch msg.BattleCommand {
		case "BATTLE:MAP_DETAILS":
			var dataPayload *MapDetailsPayload
			if err = json.Unmarshal(msg.Payload, &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message payload")
				continue
			}

			// update map detail
			btl.storeGameMap(dataPayload.Details, dataPayload.BattleZones)

			err = btl.setBattleQueue()
			if err != nil {
				L.Error().Err(err).Msg("battle start load out has failed")
				return
			}

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

			if gameClientBuildNo < arena.gameClientMinimumBuildNo {
				L.Panic().Str("current_game_client_build", dataPayload.ClientBuildNo).Uint64("minimum_game_client_build", arena.gameClientMinimumBuildNo).Msg("unsupported game client build number")
			}

			err = btl.preIntro(dataPayload)
			if err != nil {
				L.Error().Msg("battle start load out has failed")
				return
			}
			arena.NewBattleChan <- &NewBattleChan{BattleNumber: btl.BattleNumber}
		case "BATTLE:OUTRO_FINISHED":
			arena.beginBattle()
		case "BATTLE:INTRO_FINISHED":
			btl.start()
		case "BATTLE:WAR_MACHINE_DESTROYED":
			// do not process, if battle already ended
			if btl.stage.Load() == BattleStageEnd {
				continue
			}

			var dataPayload BattleWMDestroyedPayload
			if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
				L.Warn().Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
				continue
			}
			btl.Destroyed(&dataPayload)
		case "BATTLE:WAR_MACHINE_PICKUP":
			// NOTE: repair ability is moved to mech ability, this endpoint maybe used for other pickup ability
		case "BATTLE:END":
			var dataPayload *BattleEndPayload
			if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
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
		default:
			L.Warn().Err(err).Msg("Battle Arena WS: no command response")
		}
		L.Debug().Msg("game client message handled")
	}
}

func (arena *Arena) beginBattle() {
	gamelog.L.Trace().Str("func", "beginBattle").Msg("start")
	defer gamelog.L.Trace().Str("func", "beginBattle").Msg("end")

	// delete all the unfinished mech command
	_, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
		boiler.MechMoveCommandLogWhere.DeletedAt.IsNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechMoveCommandLogColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to clean up unfinished mech move command")
	}

	// delete all the mech ability trigger logs
	_, err = boiler.MechAbilityTriggerLogs(
		boiler.MechAbilityTriggerLogWhere.DeletedAt.IsNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechAbilityTriggerLogColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to clean up mech ability trigger logs.")
	}

	gm, err := db.GameMapGetRandom(false)
	if err != nil {
		gamelog.L.Err(err).Msg("unable to get random map")
		return
	}

	gameMap := &server.GameMap{
		ID:   uuid.Must(uuid.FromString(gm.ID)),
		Name: gm.Name,
	}

	var battleID string
	var battle *boiler.Battle
	inserted := false

	// query last battle
	lastBattle, err := boiler.Battles(
		boiler.BattleWhere.ArenaID.EQ(arena.ID),
		qm.OrderBy("battle_number DESC"), qm.Limit(1),
		qm.Load(
			boiler.BattleRels.GameMap,
			qm.Select(boiler.GameMapColumns.Name),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("not able to load previous battle")
	}

	// if last battle is ended or does not exist, create a new battle
	if lastBattle == nil || lastBattle.EndedAt.Valid {

		battleID = uuid.Must(uuid.NewV4()).String()
		battle = &boiler.Battle{
			ID:        battleID,
			GameMapID: gameMap.ID.String(),
			StartedAt: time.Now(),
			ArenaID:   arena.ID,
		}
	} else {
		// if there is an unfinished battle
		battle = lastBattle
		battleID = lastBattle.ID

		gamelog.L.Info().Msg("Running unfinished battle map")
		gameMap.ID = uuid.Must(uuid.FromString(lastBattle.GameMapID))
		gameMap.Name = lastBattle.R.GameMap.Name

		inserted = true
	}

	btl := &Battle{
		arena:                  arena,
		MapName:                gameMap.Name,
		gameMap:                gameMap,
		BattleID:               battleID,
		Battle:                 battle,
		inserted:               inserted,
		stage:                  atomic.NewInt32(BattleStageStart),
		destroyedWarMachineMap: make(map[string]*WMDestroyedRecord),
	}
	gamelog.L.Debug().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up incognito manager")
	btl.storePlayerAbilityManager(NewPlayerAbilityManager())

	err = btl.Load()
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to load out mechs")
	}

	// order the mechs by faction id

	arena.storeCurrentBattle(btl)
	arena.Message(BATTLEINIT, btl)

	go arena.NotifyUpcomingWarMachines()
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
	if payload.BattleID != btl.BattleID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.BattleID, payload.BattleID))
	}

	if payload.SpawnedAIEvent == nil {
		return terror.Error(fmt.Errorf("missing Spawned AI event"))
	}

	// get spawned AI
	spawnedAI := &WarMachine{
		ParticipantID: payload.SpawnedAIEvent.ParticipantID,
		Hash:          payload.SpawnedAIEvent.Hash,
		OwnedByID:     payload.SpawnedAIEvent.UserID,
		Name:          payload.SpawnedAIEvent.Name,
		Model:         payload.SpawnedAIEvent.Model,
		Skin:          payload.SpawnedAIEvent.Skin,
		MaxHealth:     payload.SpawnedAIEvent.MaxHealth,
		Health:        payload.SpawnedAIEvent.MaxHealth,
		MaxShield:     payload.SpawnedAIEvent.MaxShield,
		Shield:        payload.SpawnedAIEvent.MaxShield,
		FactionID:     payload.SpawnedAIEvent.FactionID,
		Position:      payload.SpawnedAIEvent.Position,
		Rotation:      payload.SpawnedAIEvent.Rotation,
		Image:         "https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-mini-mech.png",
		ImageAvatar:   "https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-mini-mech.png",
		AIType:        &payload.SpawnedAIEvent.Type,
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

	if payload.BattleID != btl.BattleID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.BattleID, payload.BattleID))
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
			MechMoveCommandLog:    mmc,
			RemainCooldownSeconds: MechMoveCooldownSeconds - int(time.Now().Sub(mmc.CreatedAt).Seconds()),
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
				RemainCooldownSeconds: int(mmmc.CooldownExpiry.Sub(time.Now()).Seconds()),
				IsMiniMech:            true,
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
	if payload.BattleID != btl.BattleID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.BattleID, payload.BattleID))
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
