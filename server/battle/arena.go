package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
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
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"nhooyr.io/websocket"
)

type NewBattleChan struct {
	BattleNumber int
}
type Arena struct {
	server                   *http.Server
	opts                     *Opts
	socket                   *websocket.Conn
	connected                *atomic.Bool
	timeout                  time.Duration
	_currentBattle           *Battle
	syndicates               map[string]boiler.Faction
	AIPlayers                map[string]db.PlayerWithFaction
	RPCClient                *xsyn_rpcclient.XsynXrpcClient
	gameClientLock           sync.Mutex
	sms                      server.SMS
	gameClientMinimumBuildNo uint64
	telegram                 server.Telegram
	SystemBanManager         *SystemBanManager
	NewBattleChan            chan *NewBattleChan
	sync.RWMutex
}

func (arena *Arena) IsClientConnected() error {
	connected := arena.connected.Load()
	if !connected {
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

func (arena *Arena) WarMachineDestroyedDetail(mechID string) *WMDestroyedRecord {
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
}

// return a copy of current battle user list
func (arena *Arena) currentBattleUsersCopy() []*BattleUser {
	arena.RLock()
	defer arena.RUnlock()
	if arena._currentBattle == nil {
		return nil
	}

	// copy current user map to list
	battleUsers := []*BattleUser{}
	arena._currentBattle.users.RLock()
	for _, bu := range arena._currentBattle.users.m {
		battleUsers = append(battleUsers, bu)
	}
	arena._currentBattle.users.RUnlock()

	return battleUsers
}

type Opts struct {
	Addr                     string
	Timeout                  time.Duration
	RPCClient                *xsyn_rpcclient.XsynXrpcClient
	SMS                      server.SMS
	GameClientMinimumBuildNo uint64
	Telegram                 *telegram.Telegram
}

type MessageType byte

// NetMessageTypes
const (
	JSON MessageType = iota
	Tick
)

// BATTLESPAWNCOUNT defines how many mechs to spawn
// this should be refactored to a number in the data base
// config table may be necessary, suggest key/value
const BATTLESPAWNCOUNT int = 3

func (mt MessageType) String() string {
	return [...]string{"JSON", "Tick", "Live Vote Tick", "Viewer Live Count Tick", "Spoils of War Tick", "game ability progress tick", "battle ability progress tick", "unknown", "unknown wtf"}[mt]
}

var VoteBucket = leakybucket.NewCollector(8, 8, true)

func NewArena(opts *Opts) *Arena {
	arena := &Arena{
		connected:                atomic.NewBool(false),
		timeout:                  opts.Timeout,
		RPCClient:                opts.RPCClient,
		sms:                      opts.SMS,
		gameClientMinimumBuildNo: opts.GameClientMinimumBuildNo,
		telegram:                 opts.Telegram,
		opts:                     opts,
		SystemBanManager:         NewSystemBanManager(),
		NewBattleChan:            make(chan *NewBattleChan, 10),
	}

	var err error
	arena.AIPlayers, err = db.DefaultFactionPlayers()
	if err != nil {
		gamelog.L.Fatal().Str("log_name", "battle arena").Err(err).Msg("no faction users found")
	}

	if arena.timeout == 0 {
		arena.timeout = 15 * time.Hour * 24
	}

	arena.server = &http.Server{
		Handler:      arena,
		ReadTimeout:  arena.timeout,
		WriteTimeout: arena.timeout,
	}

	// start player rank updater
	arena.PlayerRankUpdater()

	return arena
}

func (arena *Arena) Serve() {
	l, err := net.Listen("tcp", arena.opts.Addr)
	if err != nil {
		gamelog.L.Fatal().Str("log_name", "battle arena").Str("Addr", arena.opts.Addr).Err(err).Msg("unable to bind Arena to Battle Server address")
	}
	go func() {
		gamelog.L.Info().Msgf("Starting Battle Arena Server on: %v", arena.opts.Addr)

		err := arena.server.Serve(l)
		if err != nil {
			gamelog.L.Fatal().Str("log_name", "battle arena").Str("Addr", arena.opts.Addr).Err(err).Msg("unable to start Battle Arena server")
		}
	}()
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

func (btl *Battle) QueueDefaultMechs() error {
	defMechs, err := db.DefaultMechs()
	if err != nil {
		return err
	}

	for _, mech := range defMechs {
		mech.Name = helpers.GenerateStupidName()
		mechToUpdate := boiler.Mech{
			ID:   mech.ID,
			Name: mech.Name,
		}
		_, _ = mechToUpdate.Update(gamedb.StdConn, boil.Whitelist(boiler.MechColumns.Name))

		// insert default mech into battle
		ownerID, err := uuid.FromString(mech.OwnerID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
			return err
		}

		existMech, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mech.ID)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mech.ID).Err(err).Msg("check mech exists in queue")
			return terror.Error(err, "Failed to check whether mech is in the battle queue")
		}

		if existMech != nil {
			continue
		}

		result, err := db.QueueLength(uuid.FromStringOrNil(mech.FactionID.String))
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("factionID", mech.FactionID).Err(err).Msg("unable to retrieve queue length")
			return err
		}

		queueStatus := CalcNextQueueStatus(result)

		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("unable to begin tx")
			return fmt.Errorf(terror.Echo(err))
		}

		defer tx.Rollback()

		bc := &boiler.BattleContract{
			MechID:         mech.ID,
			FactionID:      mech.FactionID.String,
			PlayerID:       ownerID.String(),
			ContractReward: queueStatus.ContractReward,
			Fee:            queueStatus.QueueCost,
		}
		err = bc.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Interface("mech", mech).
				Str("contractReward", queueStatus.ContractReward.String()).
				Str("queueFee", queueStatus.QueueCost.String()).
				Err(err).Msg("unable to create battle contract")
			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		bq := &boiler.BattleQueue{
			MechID:           mech.ID,
			QueuedAt:         time.Now(),
			FactionID:        mech.FactionID.String,
			OwnerID:          ownerID.String(),
			BattleContractID: null.StringFrom(bc.ID),
		}

		err = bq.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Interface("mech", mech).
				Err(err).Msg("unable to insert mech into queue")
			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Interface("mech", mech).
				Err(err).Msg("unable to commit mech insertion into queue")
			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

	}

	return nil
}

func (arena *Arena) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
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

	arena.socket = c
	arena.connected.Store(true)

	defer func() {
		if c != nil {
			arena.connected.Store(false)
			gamelog.L.Error().Str("log_name", "battle arena").Err(fmt.Errorf("game client has disconnected")).Msg("lost connection to game client")
			c.Close(websocket.StatusInternalError, "game client has disconnected")

			btl := arena.CurrentBattle()
			if btl != nil && btl.spoils != nil {
				btl.spoils.End()
			}
		}
	}()

	arena.Start()
}

type BribeGabRequest struct {
	Payload struct {
		AbilityOfferingID string          `json:"ability_offering_id"`
		Percentage        decimal.Decimal `json:"percentage"` // "0.1", "0.5%", "1%"
	} `json:"payload"`
}

const HubKeyBattleAbilityBribe = "BATTLE:ABILITY:BRIBE"

func (arena *Arena) BattleAbilityBribe(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	b := VoteBucket.Add(user.ID, 1)
	if b == 0 {
		return nil
	}

	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("bribe", user.ID).Msg("current battle is nil")
		return nil
	}

	req := &BribeGabRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("json", string(payload)).Msg("json unmarshal failed")
		return terror.Error(err, "Invalid request received")
	}

	// check percentage amount is valid
	if _, ok := MinVotePercentageCost[req.Payload.Percentage.String()]; !ok {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("payload", req).
			Str("userID", user.ID).
			Str("percentage", req.Payload.Percentage.String()).
			Msg("invalid vote percentage amount received")
		return terror.Error(err, "Invalid vote percentage amount received")
	}

	// check user is banned on limit sups contribution
	isBanned, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanSupsContribute.EQ(true),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err, "Failed to check player")
	}

	// if limited sups contribute, return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to contribute sups"), "You are banned to contribute sups")
	}

	cannotTrigger, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err, "Failed to check player")
	}

	userID := uuid.FromStringOrNil(user.ID)
	if userID.IsNil() {
		gamelog.L.Error().Str("log_name", "battle arena").Str("user id is nil", user.ID).Msg("cant make users")
		return terror.Error(terror.ErrForbidden)
	}

	arena.CurrentBattle().abilities().BribeGabs(factionID, userID, req.Payload.AbilityOfferingID, req.Payload.Percentage, cannotTrigger, reply)

	return nil
}

type LocationSelectRequest struct {
	Payload struct {
		StartCoords server.CellLocation  `json:"start_coords"`
		EndCoords   *server.CellLocation `json:"end_coords,omitempty"`
	} `json:"payload"`
}

const HubKeyAbilityLocationSelect = "ABILITY:LOCATION:SELECT"

var locationSelectBucket = leakybucket.NewCollector(1, 1, true)

func (arena *Arena) AbilityLocationSelect(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if locationSelectBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many Requests")
	}

	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Msg("no current battle")
		return nil
	}

	req := &LocationSelectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil || userID.IsNil() {
		gamelog.L.Warn().Err(err).Msgf("can't create uuid from wsc identifier %s", user.ID)
		return terror.Error(terror.ErrForbidden)
	}

	if arena.CurrentBattle().abilities() == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Msg("abilities is nil even with current battle not being nil")
		return terror.Error(terror.ErrForbidden)
	}

	err = arena.CurrentBattle().abilities().LocationSelect(userID, req.Payload.StartCoords, req.Payload.EndCoords)
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

func (arena *Arena) MinimapUpdatesSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return terror.Error(terror.ErrForbidden, "There is no battle currently to use this ability on.")
	}

	reply(nil)

	return nil
}

// PublicBattleAbilityUpdateSubscribeHandler return battle ability for non login player
func (arena *Arena) PublicBattleAbilityUpdateSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// get a random faction id
	if arena.CurrentBattle() != nil {
		btl := arena.CurrentBattle()
		if btl.abilities() != nil {
			ga, _ := btl.abilities().FactionBattleAbilityGet(server.RedMountainFactionID)
			if ga != nil {
				reply(GameAbility{
					ID:                     ga.ID,
					GameClientAbilityID:    byte(ga.GameClientAbilityID),
					ImageUrl:               ga.ImageUrl,
					Description:            ga.Description,
					FactionID:              ga.FactionID,
					Label:                  ga.Label,
					SupsCost:               ga.SupsCost,
					CurrentSups:            ga.CurrentSups,
					Colour:                 ga.Colour,
					TextColour:             ga.TextColour,
					CooldownDurationSecond: ga.CooldownDurationSecond,
					OfferingID:             uuid.Nil, // remove offering id to disable bribing
				})
			}
		}
	}
	return nil
}

const HubKeyBattleAbilityUpdated = "BATTLE:ABILITY:UPDATED"

func (arena *Arena) BattleAbilityUpdateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// return data if, current battle is not null
	if arena.CurrentBattle() != nil {
		btl := arena.CurrentBattle()
		if btl.abilities() != nil {
			ability, _ := btl.abilities().FactionBattleAbilityGet(factionID)
			reply(ability)
		}
	}

	return nil
}

type GameAbilityContributeRequest struct {
	Payload struct {
		AbilityIdentity   string          `json:"ability_identity"`
		AbilityOfferingID string          `json:"ability_offering_id"`
		Percentage        decimal.Decimal `json:"percentage"` // "0.1", "0.5%", "1%"
	} `json:"payload"`
}

const HubKeFactionUniqueAbilityContribute = "FACTION:UNIQUE:ABILITY:CONTRIBUTE"

func (arena *Arena) FactionUniqueAbilityContribute(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	b := VoteBucket.Add(user.ID, 1)
	if b == 0 {
		return nil
	}

	if arena == nil || arena.CurrentBattle() == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Bool("arena", arena == nil).
			Str("factionID", factionID).
			Bool("current_battle", arena.CurrentBattle() == nil).
			Str("userID", user.ID).Msg("unable to find player from user id")
		return nil
	}

	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("payload", req).
			Str("userID", user.ID).Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	// check percentage amount is valid
	if _, ok := MinVotePercentageCost[req.Payload.Percentage.String()]; !ok {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("payload", req).
			Str("userID", user.ID).
			Str("percentage", req.Payload.Percentage.String()).
			Msg("invalid vote percentage amount received")
		return terror.Error(err, "Invalid vote percentage amount received")
	}

	// check user is banned on limit sups contribution
	isBanned, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanSupsContribute.EQ(true),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err, "Failed to check player")
	}

	// if limited sups contribute, return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to contribute sups"), "You are banned to contribute sups")
	}

	userID := uuid.FromStringOrNil(user.ID)
	if userID.IsNil() {
		gamelog.L.Error().Str("log_name", "battle arena").Str("percentage", req.Payload.Percentage.String()).
			Str("userID", user.ID).Msg("unable to contribute forbidden")
		return terror.Error(terror.ErrForbidden)
	}

	cannotTrigger, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err, "Failed to check player")
	}

	arena.CurrentBattle().abilities().AbilityContribute(factionID, userID, req.Payload.AbilityIdentity, req.Payload.AbilityOfferingID, req.Payload.Percentage, cannotTrigger, reply)

	return nil
}

const HubKeyFactionUniqueAbilitiesUpdated = "FACTION:UNIQUE:ABILITIES:UPDATED"

func (arena *Arena) FactionAbilitiesUpdateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// return data if, current battle is not null
	btl := arena.CurrentBattle()
	if btl != nil {
		if btl.abilities() != nil {
			reply(btl.abilities().FactionUniqueAbilitiesGet(uuid.FromStringOrNil(factionID)))
		}
	}

	return nil
}

const HubKeyWarMachineAbilitiesUpdated = "WAR:MACHINE:ABILITIES:UPDATED"

// WarMachineAbilitiesUpdateSubscribeHandler subscribe on war machine abilities
func (arena *Arena) WarMachineAbilitiesUpdateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	slotNumber := cctx.URLParam("slotNumber")
	if slotNumber == "" {
		return fmt.Errorf("slot number is required")
	}

	participantID, err := strconv.Atoi(slotNumber)
	if err != nil {
		return fmt.Errorf("invalid participant id")
	}

	wm := arena.CurrentBattleWarMachine(participantID)

	if wm == nil {
		return nil
	}
	if wm.FactionID != factionID {
		gamelog.L.Warn().Str("war_machine_faction_id", wm.FactionID).Str("user_faction_id", factionID).Msg("War machine faction id does not match")
		return nil
	}

	gameAbilities := []GameAbility{}
	for _, ga := range wm.Abilities {
		ga.RLock()
		gameAbilities = append(gameAbilities, *ga)
		ga.RUnlock()
	}

	reply(gameAbilities)

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

//func (arena *Arena) WarMachineStatUpdatedSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
//	cctx := chi.RouteContext(ctx)
//	slotNumber := cctx.URLParam("slotNumber")
//	if slotNumber == "" {
//		return fmt.Errorf("slot number is required")
//	}
//
//	participantID, err := strconv.Atoi(slotNumber)
//	if err != nil {
//		return fmt.Errorf("invalid participant id")
//	}
//
//	wm := arena.CurrentBattleWarMachine(participantID)
//
//	if wm != nil {
//		wm.RLock()
//		defer wm.RUnlock()
//		reply(WarMachineStat{
//			ParticipantID: participantID,
//			Position:      wm.Position,
//			Rotation:      wm.Rotation,
//			Health:        wm.Health,
//			Shield:        wm.Shield,
//			IsHidden:      false,
//		})
//	}
//
//	return nil
//}

const HubKeyBribeStageUpdateSubscribe = "BRIBE:STAGE:UPDATED:SUBSCRIBE"

// BribeStageSubscribe subscribe on bribing stage change
func (arena *Arena) BribeStageSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// return data if, current battle is not null
	if arena.CurrentBattle() != nil {
		btl := arena.CurrentBattle()
		if btl.abilities() != nil {
			reply(btl.abilities().BribeStageGet())
		}
	}

	return nil
}

const HubKeyBattleAbilityProgressBarUpdated = "BATTLE:ABILITY:PROGRESS:BAR:UPDATED"

const HubKeyAbilityPriceUpdated = "ABILITY:PRICE:UPDATED"

type GameAbilityPriceResponse struct {
	ID          string `json:"id"`
	OfferingID  string `json:"offering_id"`
	SupsCost    string `json:"sups_cost"`
	CurrentSups string `json:"current_sups"`
	ShouldReset bool   `json:"should_reset"`
}

const HubKeySpoilOfWarUpdated = "SPOIL:OF:WAR:UPDATED"

func (arena *Arena) SpoilOfWarUpdateSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	sows, err := db.LastTwoSpoilOfWarAmount()
	if err != nil || len(sows) == 0 {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get last two spoil of war amount")
		return nil
	}

	spoilOfWars := []string{}
	for _, sow := range sows {
		spoilOfWars = append(spoilOfWars, sow.String())
	}

	reply(spoilOfWars)

	return nil
}

func (arena *Arena) SendSettings(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
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
	Details  server.GameMap `json:"details"`
	BattleID string         `json:"battleID"`
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
		KilledBy string `json:"killedBy"`
	} `json:"destroyedWarMachineEvent"`
	BattleID string `json:"battleID"`
}

type AISpawnedRequest struct {
	BattleID       string          `json:"battleID"`
	SpawnedAIEvent *SpawnedAIEvent `json:"spawnedAIEvent"`
}

type SpawnedAIEvent struct {
	ParticipantID byte            `json:"participantID"`
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
}

type BattleWMPickupPayload struct {
	WarMachineHash string `json:"warMachineHash"`
	EventID        string `json:"eventID"`
	BattleID       string `json:"battleID"`
}

func (arena *Arena) start() {
	//defer func() {
	//	if r := recover(); r != nil {
	//		gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic on battle arena!", r)
	//	}
	//}()

	ctx := context.Background()
	arena.beginBattle()

	for {
		_, payload, err := arena.socket.Read(ctx)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("empty game client disconnected")
			break
		}
		btl := arena.CurrentBattle()
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
				gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message")
				continue
			}

			gamelog.L.Info().Str("game_client_data", string(data)).Int("message_type", int(mt)).Msg("game client message received")

			switch msg.BattleCommand {
			case "BATTLE:MAP_DETAILS":
				var dataPayload *MapDetailsPayload
				if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message payload")
					continue
				}

				// update map detail
				btl.storeGameMap(dataPayload.Details)

			case "BATTLE:START":
				var dataPayload *BattleStartPayload
				if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message payload")
					continue
				}

				gameClientBuildNo, err := strconv.ParseUint(dataPayload.ClientBuildNo, 10, 64)
				if err != nil {
					gamelog.L.Panic().Str("game_client_build_no", dataPayload.ClientBuildNo).Msg("invalid game client build number received")
				}

				if gameClientBuildNo < arena.gameClientMinimumBuildNo {
					gamelog.L.Panic().Str("current_game_client_build", dataPayload.ClientBuildNo).Uint64("minimum_game_client_build", arena.gameClientMinimumBuildNo).Msg("unsupported game client build number")
				}

				err = btl.preIntro(dataPayload)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("battle start load out has failed")
					return
				}
				battleInfo := &NewBattleChan{BattleNumber: btl.BattleNumber}
				arena.NewBattleChan <- battleInfo

			case "BATTLE:OUTRO_FINISHED":
				gamelog.L.Info().Msg("Battle outro is finished, starting a new battle")
				arena.beginBattle()

			case "BATTLE:INTRO_FINISHED":
				btl.start()

			case "BATTLE:WAR_MACHINE_DESTROYED":
				var dataPayload BattleWMDestroyedPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
					continue
				}
				btl.Destroyed(&dataPayload)

			case "BATTLE:WAR_MACHINE_PICKUP":
				// NOTE: repair ability is moved to mech ability, this endpoint maybe used for other pickup ability

			case "BATTLE:END":
				var dataPayload *BattleEndPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
					continue
				}
				btl.end(dataPayload)

			case "BATTLE:AI_SPAWNED":
				var dataPayload *AISpawnedRequest
				if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message payload")
					continue
				}
				err = btl.AISpawned(dataPayload)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err)
				}

			case "BATTLE:ABILITY_MOVE_COMMAND_COMPLETE":
				var dataPayload *AbilityMoveCommandCompletePayload
				if err := json.Unmarshal(msg.Payload, &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal ability move command complete payload")
					continue
				}
				err = btl.UpdateWarMachineMoveCommand(dataPayload)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err)
				}

			default:
				gamelog.L.Warn().Str("battleCommand", msg.BattleCommand).Err(err).Msg("Battle Arena WS: no command response")
			}
		case Tick:
			btl.Tick(payload)
		default:
			gamelog.L.Warn().Str("MessageType", string(mt)).Err(err).Msg("Battle Arena WS: no message response")
		}
	}
}

func (arena *Arena) beginBattle() {
	// delete all the unfinished mech command
	_, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
		boiler.MechMoveCommandLogWhere.DeletedAt.IsNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechMoveCommandLogColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to clean up unfinished mech move command")
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
	if lastBattle == nil || errors.Is(err, sql.ErrNoRows) || lastBattle.EndedAt.Valid {

		battleID = uuid.Must(uuid.NewV4()).String()
		battle = &boiler.Battle{
			ID:        battleID,
			GameMapID: gameMap.ID.String(),
			StartedAt: time.Now(),
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
		arena:    arena,
		MapName:  gameMap.Name,
		gameMap:  gameMap,
		BattleID: battleID,
		Battle:   battle,
		inserted: inserted,
		stage:    atomic.NewInt32(BattleStageStart),
		users: usersMap{
			m: make(map[uuid.UUID]*BattleUser),
		},
		destroyedWarMachineMap: make(map[string]*WMDestroyedRecord),
		viewerCountInputChan:   make(chan *ViewerLiveCount),
	}
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up incognito manager")
	btl.storePlayerAbilityManager(NewPlayerAbilityManager())

	err = btl.Load()
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to load out mechs")
	}

	// order the mechs by facton id

	// set user online debounce
	go btl.debounceSendingViewerCount(func(result ViewerLiveCount, btl *Battle) {
		ws.PublishMessage("/public/live_data", HubKeyViewerLiveCountUpdated, result)
	})

	arena.storeCurrentBattle(btl)
	arena.Message(BATTLEINIT, btl)

	go arena.NotifyUpcomingWarMachines()
}

const HubKeyUserStatSubscribe = "USER:STAT:SUBSCRIBE"

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
	}

	gamelog.L.Info().Msgf("Battle Update: %s - AI Spawned: %d", payload.BattleID, spawnedAI.ParticipantID)

	// cache record in battle, for future subscription
	btl.spawnedAIMux.Lock()
	btl.SpawnedAI = append(btl.SpawnedAI, spawnedAI)
	btl.spawnedAIMux.Unlock()

	return nil
}

func (btl *Battle) UpdateWarMachineMoveCommand(payload *AbilityMoveCommandCompletePayload) error {
	if payload.BattleID != btl.BattleID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", btl.BattleID, payload.BattleID))
	}

	// check battle state
	if btl.arena.currentBattleState() == BattleStageEnd {
		return terror.Error(fmt.Errorf("current battle is ended"))
	}

	// get mech
	wm := btl.arena.CurrentBattleWarMachineByHash(payload.WarMachineHash)
	if wm == nil {
		return terror.Error(fmt.Errorf("war machine not exists"))
	}

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

	_, err = mmc.Update(gamedb.StdConn, boil.Whitelist(boiler.MechMoveCommandLogColumns.ReachedAt))
	if err != nil {
		return terror.Error(err, "Failed to update mech move command")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_command/%s", wm.FactionID, wm.Hash), HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
		MechMoveCommandLog:    mmc,
		RemainCooldownSeconds: MechMoveCooldownSeconds - int(time.Now().Sub(mmc.CreatedAt).Seconds()),
	})

	err = btl.arena.BroadcastFactionMechCommands(wm.FactionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to broadcast faction mech commands")
	}

	btl.arena.BroadcastMechCommandNotification(&MechCommandNotification{
		MechID:       wm.ID,
		MechLabel:    wm.Name,
		MechImageUrl: wm.ImageAvatar,
		FactionID:    wm.FactionID,
		Action:       MechCommandActionComplete,
	})

	return nil
}
