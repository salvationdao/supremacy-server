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
	"server/rpcclient"
	"server/telegram"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"nhooyr.io/websocket"
)

type Arena struct {
	conn                     db.Conn
	socket                   *websocket.Conn
	timeout                  time.Duration
	messageBus               *messagebus.MessageBus
	_currentBattle           *Battle
	syndicates               map[string]boiler.Faction
	AIPlayers                map[string]db.PlayerWithFaction
	RPCClient                *rpcclient.PassportXrpcClient
	gameClientLock           sync.Mutex
	sms                      server.SMS
	gameClientMinimumBuildNo uint64
	telegram                 server.Telegram

	sync.RWMutex
}

func (arena *Arena) BattleSeconds() decimal.Decimal {
	btl := arena.currentBattle()
	if btl == nil {
		lastBattle, err := boiler.Battles(qm.OrderBy("battle_number DESC"), qm.Limit(1)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("not able to load previous battle")
			return decimal.Zero
		}
		if lastBattle != nil {
			if lastBattle.EndedBattleSeconds.Valid {
				return lastBattle.EndedBattleSeconds.Decimal.Copy()
			} else {
				return lastBattle.StartedBattleSeconds.Decimal.Copy()
			}
		}

		return decimal.Zero
	}
	return btl.battleSeconds()
}

func (arena *Arena) currentBattle() *Battle {
	arena.RLock()
	defer arena.RUnlock()
	return arena._currentBattle
}

func (arena *Arena) storeCurrentBattle(btl *Battle) {
	arena.Lock()
	defer arena.Unlock()
	arena._currentBattle = btl
}

func (arena *Arena) currentBattleNumber() int {
	arena.RLock()
	defer arena.RUnlock()
	if arena._currentBattle == nil {
		return -1
	}
	return arena._currentBattle.BattleNumber
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

func (arena *Arena) SendToOnlinePlayer(playerID uuid.UUID, key hub.HubCommandKey, payload interface{}) {
	arena.RLock()
	defer arena.RUnlock()
	if arena._currentBattle == nil {
		return
	}

	arena._currentBattle.users.Send(key, payload, playerID)
}

type Opts struct {
	Conn                     db.Conn
	Addr                     string
	Timeout                  time.Duration
	Hub                      *hub.Hub
	MessageBus               *messagebus.MessageBus
	RPCClient                *rpcclient.PassportXrpcClient
	SMS                      server.SMS
	GameClientMinimumBuildNo uint64
	Telegram                 *telegram.Telegram
}

type MessageType byte

// NetMessageTypes
const (
	JSON MessageType = iota
	Tick
	LiveVotingTick
	ViewerLiveCountTick
	SpoilOfWarTick
	GameAbilityProgressTick
	BattleAbilityProgressTick
)

// BATTLESPAWNCOUNT defines how many mechs to spawn
// this should be refactored to a number in the database
// config table may be necessary, suggest key/value
const BATTLESPAWNCOUNT int = 3

func (mt MessageType) String() string {
	return [...]string{"JSON", "Tick", "Live Vote Tick", "Viewer Live Count Tick", "Spoils of War Tick", "game ability progress tick", "battle ability progress tick", "unknown", "unknown wtf"}[mt]
}

var VoteBucket = leakybucket.NewCollector(8, 8, true)

func NewArena(opts *Opts) *Arena {
	l, err := net.Listen("tcp", opts.Addr)

	if err != nil {
		gamelog.L.Fatal().Str("Addr", opts.Addr).Err(err).Msg("unable to bind Arena to Battle Server address")
	}

	arena := &Arena{
		conn: opts.Conn,
	}

	arena.timeout = opts.Timeout
	arena.messageBus = opts.MessageBus
	arena.RPCClient = opts.RPCClient
	arena.sms = opts.SMS
	arena.gameClientMinimumBuildNo = opts.GameClientMinimumBuildNo
	arena.telegram = opts.Telegram

	arena.AIPlayers, err = db.DefaultFactionPlayers()
	if err != nil {
		gamelog.L.Fatal().Err(err).Msg("no faction users found")
	}

	if arena.timeout == 0 {
		arena.timeout = 15 * time.Hour * 24
	}

	server := &http.Server{
		Handler:      arena,
		ReadTimeout:  arena.timeout,
		WriteTimeout: arena.timeout,
	}

	// faction queue
	opts.SecureUserFactionCommand(WSQueueJoin, arena.QueueJoinHandler)
	opts.SecureUserFactionCommand(WSQueueLeave, arena.QueueLeaveHandler)
	opts.SecureUserFactionCommand(WSAssetQueueStatus, arena.AssetQueueStatusHandler)
	opts.SecureUserFactionCommand(WSAssetQueueStatusList, arena.AssetQueueStatusListHandler)
	opts.SecureUserFactionSubscribeCommand(WSQueueStatusSubscribe, arena.QueueStatusSubscribeHandler)
	opts.SecureUserFactionSubscribeCommand(WSQueueUpdatedSubscribe, arena.QueueUpdatedSubscribeHandler)
	opts.SecureUserFactionSubscribeCommand(WSAssetQueueStatusSubscribe, arena.AssetQueueStatusSubscribeHandler)

	opts.SecureUserCommand(HubKeyGameUserOnline, arena.UserOnline)
	opts.SecureUserCommand(HubKeyPlayerRankGet, arena.PlayerRankGet)
	opts.SubscribeCommand(HubKeyWarMachineDestroyedUpdated, arena.WarMachineDestroyedUpdatedSubscribeHandler)

	// subscribe functions
	opts.SubscribeCommand(HubKeyGameSettingsUpdated, arena.SendSettings)

	opts.SubscribeCommand(HubKeyGameNotification, arena.GameNotificationSubscribeHandler)
	opts.SecureUserCommand(HubKeyMultiplierUpdate, arena.HubKeyMultiplierUpdate)

	opts.SecureUserSubscribeCommand(HubKeyUserStatSubscribe, arena.UserStatUpdatedSubscribeHandler)

	// battle ability related (bribing)
	opts.SecureUserFactionCommand(HubKeyBattleAbilityBribe, arena.BattleAbilityBribe)
	opts.SecureUserFactionCommand(HubKeyAbilityLocationSelect, arena.AbilityLocationSelect)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribeStageUpdateSubscribe, arena.GabsBribeStageSubscribe)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribingWinnerSubscribe, arena.GabsBribingWinnerSubscribe)
	opts.SecureUserFactionSubscribeCommand(HubKeyBattleAbilityUpdated, arena.BattleAbilityUpdateSubscribeHandler)

	// faction unique ability related (sup contribution)
	opts.SecureUserFactionCommand(HubKeFactionUniqueAbilityContribute, arena.FactionUniqueAbilityContribute)
	opts.SecureUserFactionSubscribeCommand(HubKeyFactionUniqueAbilitiesUpdated, arena.FactionAbilitiesUpdateSubscribeHandler)
	opts.SecureUserFactionSubscribeCommand(HubKeyWarMachineAbilitiesUpdated, arena.WarMachineAbilitiesUpdateSubscribeHandler)

	// net message subscribe
	opts.NetSecureUserFactionSubscribeCommand(HubKeyBattleAbilityProgressBarUpdated, arena.FactionProgressBarUpdateSubscribeHandler)
	opts.NetSecureUserFactionSubscribeCommand(HubKeyAbilityPriceUpdated, arena.FactionAbilityPriceUpdateSubscribeHandler)
	opts.NetSubscribeCommand(HubKeyWarMachineLocationUpdated, arena.WarMachineLocationUpdateSubscribeHandler)
	opts.NetSecureUserFactionSubscribeCommand(HubKeyLiveVoteCountUpdated, arena.LiveVoteCountUpdateSubscribeHandler)
	opts.NetSecureUserSubscribeCommand(HubKeySpoilOfWarUpdated, arena.SpoilOfWarUpdateSubscribeHandler)

	// start player rank updater
	arena.PlayerRankUpdater()

	go func() {
		err = server.Serve(l)

		if err != nil {
			gamelog.L.Fatal().Str("Addr", opts.Addr).Err(err).Msg("unable to start Battle Arena server")
		}
	}()

	return arena
}

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
		gamelog.L.Fatal().Interface("payload", payload).Err(err).Msg("unable to marshal data for battle arena")
	}

	gamelog.L.Debug().Str("message data", string(b)).Msg("sending packet to game client")

	arena.socket.Write(ctx, websocket.MessageBinary, b)
}

func (btl *Battle) DefaultMechs() error {
	defMechs, err := db.DefaultMechs()
	if err != nil {
		return err
	}

	btl.WarMachines = btl.MechsToWarMachines(defMechs)
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
	}

	arena.socket = c

	defer func() {
		if c != nil {
			c.Close(websocket.StatusInternalError, "game client has disconnected")
		}
	}()

	arena.Start()
}

func (arena *Arena) SetMessageBus(mb *messagebus.MessageBus) {
	arena.messageBus = mb
}

type BribeGabRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount string `json:"amount"` // "0.1", "1", "10"
	} `json:"payload"`
}

const HubKeyBattleAbilityBribe hub.HubCommandKey = "BATTLE:ABILITY:BRIBE"

func (arena *Arena) BattleAbilityBribe(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	b := VoteBucket.Add(wsc.Identifier(), 1)
	if b == 0 {
		return nil
	}

	// skip, if current not battle
	if arena.currentBattle() == nil {
		gamelog.L.Warn().Str("bribe", wsc.Identifier()).Msg("current battle is nil")
		return nil
	}

	req := &BribeGabRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Str("json", string(payload)).Msg("json unmarshal failed")
		return terror.Error(err, "Invalid request received")
	}

	// check user is banned on limit sups contribution
	isBanned, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		boiler.PunishedPlayerWhere.PlayerID.EQ(wsc.Identifier()),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s and %s = ?",
				boiler.TableNames.PunishOptions,
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.ID),
				qm.Rels(boiler.TableNames.PunishedPlayers, boiler.PunishedPlayerColumns.PunishOptionID),
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.Key),
			),
			server.PunishmentOptionRestrictSupsContribution,
		),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err)
	}

	// if limited sups contribute, return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to contribute sups"), "You are banned to contribute sups")
	}

	d, err := decimal.NewFromString(req.Payload.Amount)
	if err != nil {
		gamelog.L.Error().Str("amount", req.Payload.Amount).Msg("cant make moneys")
		return terror.Error(err, "Failed to parse string to decimal.deciaml")
	}
	amount := d.Mul(decimal.New(1, 18))

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		gamelog.L.Error().Str("user id is nil", wsc.Identifier()).Msg("cant make users")

		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle().abilities().BribeGabs(factionID, userID, amount)

	return nil
}

type LocationSelectRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		XIndex int `json:"x"`
		YIndex int `json:"y"`
	} `json:"payload"`
}

const HubKeyAbilityLocationSelect hub.HubCommandKey = "ABILITY:LOCATION:SELECT"

func (arena *Arena) AbilityLocationSelect(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	// skip, if current not battle
	if arena.currentBattle() == nil {
		gamelog.L.Warn().Msg("no current battle")
		return nil
	}

	req := &LocationSelectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(wsc.Identifier())
	if err != nil || userID.IsNil() {
		gamelog.L.Warn().Err(err).Msgf("can't create uuid from wsc identifier %s", wsc.Identifier())
		return terror.Error(terror.ErrForbidden)
	}

	if arena.currentBattle().abilities == nil {
		gamelog.L.Error().Msg("abilities is nil even with current battle not being nil")
		return terror.Error(terror.ErrForbidden)
	}

	err = arena.currentBattle().abilities().LocationSelect(userID, req.Payload.XIndex, req.Payload.YIndex)
	if err != nil {
		gamelog.L.Warn().Err(err).Msgf("can't create uuid from wsc identifier %s", wsc.Identifier())
		return terror.Error(err)
	}

	return nil
}

const HubKeyPlayerRankGet hub.HubCommandKey = "PLAYER:RANK:GET"

func (arena *Arena) PlayerRankGet(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	player, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Rank,
		),
		boiler.PlayerWhere.ID.EQ(wsc.Identifier()),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("player id", wsc.Identifier()).Err(err).Msg("Failed to get player rank from db")
		return terror.Error(err, "Failed to get player rank from db")
	}

	reply(player.Rank)

	return nil
}

const HubKeyBattleAbilityUpdated hub.HubCommandKey = "BATTLE:ABILITY:UPDATED"

func (arena *Arena) BattleAbilityUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get faction id
	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find player from user id")
		return "", "", terror.Error(terror.ErrForbidden)
	}

	if needProcess {
		// return data if, current battle is not null
		if arena.currentBattle() != nil {
			btl := arena.currentBattle()
			if btl.abilities() != nil {
				ability, _ := btl.abilities().FactionBattleAbilityGet(factionID.String())
				reply(ability)
			}
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBattleAbilityUpdated, factionID.String())), nil
}

type GameAbilityContributeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityIdentity string `json:"ability_identity"`
		Amount          string `json:"amount"` // "0.1", "1", ""
	} `json:"payload"`
}

const HubKeFactionUniqueAbilityContribute hub.HubCommandKey = "FACTION:UNIQUE:ABILITY:CONTRIBUTE"

func (arena *Arena) FactionUniqueAbilityContribute(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	b := VoteBucket.Add(wsc.Identifier(), 1)
	if b == 0 {
		return nil
	}

	if arena == nil || arena.currentBattle() == nil || factionID.IsNil() {
		gamelog.L.Error().Bool("arena", arena == nil).
			Bool("factionID", factionID.IsNil()).
			Bool("current_battle", arena.currentBattle() == nil).
			Str("userID", wsc.Identifier()).Msg("unable to find player from user id")
		return nil
	}

	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Interface("payload", req).
			Str("userID", wsc.Identifier()).Msg("invalid request receieved")
		return terror.Error(err, "Invalid request received")
	}

	// check user is banned on limit sups contribution
	isBanned, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		boiler.PunishedPlayerWhere.PlayerID.EQ(wsc.Identifier()),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s and %s = ?",
				boiler.TableNames.PunishOptions,
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.ID),
				qm.Rels(boiler.TableNames.PunishedPlayers, boiler.PunishedPlayerColumns.PunishOptionID),
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.Key),
			),
			server.PunishmentOptionRestrictSupsContribution,
		),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err)
	}

	// if limited sups contribute, return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to contribute sups"), "You are banned to contribute sups")
	}

	d, err := decimal.NewFromString(req.Payload.Amount)
	if err != nil {
		gamelog.L.Error().Str("amount", req.Payload.Amount).
			Str("userID", wsc.Identifier()).Msg("Failed to parse string to decimal.deciaml")
		return terror.Error(err, "Failed to parse string to decimal.deciaml")
	}
	amount := d.Mul(decimal.New(1, 18))

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		gamelog.L.Error().Str("amount", req.Payload.Amount).
			Str("userID", wsc.Identifier()).Msg("unable to contribute forbidden")
		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle().abilities().AbilityContribute(factionID, userID, req.Payload.AbilityIdentity, amount)

	return nil
}

const HubKeyFactionUniqueAbilitiesUpdated hub.HubCommandKey = "FACTION:UNIQUE:ABILITIES:UPDATED"

func (arena *Arena) FactionAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get faction id
	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find player from user id")
		return "", "", terror.Error(err)
	}

	// skip, if user is non faction
	if factionID.IsNil() {
		return "", "", nil
	}

	if needProcess {
		// return data if, current battle is not null
		btl := arena.currentBattle()
		if btl != nil {
			if btl.abilities() != nil {
				reply(btl.abilities().FactionUniqueAbilitiesGet(factionID))
			}
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionUniqueAbilitiesUpdated, factionID.String()))
	return req.TransactionID, busKey, nil
}

const HubKeyWarMachineAbilitiesUpdated hub.HubCommandKey = "WAR:MACHINE:ABILITIES:UPDATED"

type WarMachineAbilitiesUpdatedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Hash string `json:"hash"`
	} `json:"payload"`
}

// TODO: refactor this to become a fetch instead of a subscription

// WarMachineAbilitiesUpdateSubscribeHandler subscribe on war machine abilities
func (arena *Arena) WarMachineAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "WarMachineAbilitiesUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	req := &WarMachineAbilitiesUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get faction id

	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find player from user id")
		return "", "", terror.Error(err)
	}

	// skip, if user is non faction
	if factionID.IsNil() {
		return "", "", nil
	}

	if needProcess {
		// NOTE: current only return faction unique ability
		// get war machine ability
		if arena.currentBattle() != nil {
			btl := arena.currentBattle()
			for _, wm := range btl.WarMachines {
				if wm.Hash == req.Payload.Hash {
					reply(wm.Abilities)
					break
				}
			}
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineAbilitiesUpdated, req.Payload.Hash))
	return req.TransactionID, busKey, nil
}

func (arena *Arena) UserOnline(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	if arena.currentBattle() == nil {
		return nil
	}
	uID, err := uuid.FromString(wsc.Identifier())
	if uID.IsNil() || err != nil {
		gamelog.L.Error().Str("uuid", wsc.Identifier()).Err(err).Msg("invalid input data")
		return fmt.Errorf("unable to construct user uuid")
	}
	userID := server.UserID(uID)

	user, err := boiler.Players(
		boiler.PlayerWhere.ID.EQ(userID.String()),
		qm.Load(boiler.PlayerRels.Faction),
	).One(gamedb.StdConn)
	if err != nil || user == nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("player does not have a faction")
		return terror.Error(terror.ErrInvalidInput)
	}

	// TODO: handle faction swap from non-faction to faction
	if !user.FactionID.Valid {
		return nil
	}

	var color = "#000000"
	if user.R.Faction != nil {
		color = user.R.Faction.PrimaryColor
	}

	battleUser := &BattleUser{
		ID:            uuid.FromStringOrNil(userID.String()),
		Username:      user.Username.String,
		FactionID:     user.FactionID.String,
		FactionColour: color,
		FactionLogoID: FactionLogos[user.FactionID.String],
		wsClient:      map[*hub.Client]bool{},
	}

	arena.currentBattle().userOnline(battleUser, wsc)
	return nil
}

type WarMachineDestroyedUpdatedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ParticipantID byte `json:"participantID"`
	} `json:"payload"`
}

const HubKeyWarMachineDestroyedUpdated = hub.HubCommandKey("WAR:MACHINE:DESTROYED:UPDATED")

func (arena *Arena) WarMachineDestroyedUpdatedSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &WarMachineDestroyedUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if needProcess {
		if arena.currentBattle() != nil {
			if wmd, ok := arena.currentBattle().destroyedWarMachineMap[req.Payload.ParticipantID]; ok {
				reply(wmd)
			}
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%x", HubKeyWarMachineDestroyedUpdated, req.Payload.ParticipantID)), nil
}

const HubKeGabsBribeStageUpdateSubscribe hub.HubCommandKey = "BRIBE:STAGE:UPDATED:SUBSCRIBE"

// GabsBribeStageSubscribe subscribe on bribing stage change
func (arena *Arena) GabsBribeStageSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	if needProcess {
		// return data if, current battle is not null
		if arena.currentBattle() != nil {
			btl := arena.currentBattle()
			if btl.abilities() != nil {
				reply(btl.abilities().BribeStageGet())
			}
		}
	}

	return req.TransactionID, messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), nil
}

const HubKeyBattleAbilityProgressBarUpdated hub.HubCommandKey = "BATTLE:ABILITY:PROGRESS:BAR:UPDATED"

func (arena *Arena) FactionProgressBarUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "FactionProgressBarUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")

	return messagebus.BusKey(HubKeyBattleAbilityProgressBarUpdated), nil
}

const HubKeyAbilityPriceUpdated hub.HubCommandKey = "ABILITY:PRICE:UPDATED"

type AbilityPriceUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityIdentity string `json:"ability_identity"`
	} `json:"payload"`
}

func (arena *Arena) FactionAbilityPriceUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	req := &AbilityPriceUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", terror.Error(err, "Invalid request received")
	}

	return messagebus.BusKey(fmt.Sprintf("%s,%s", HubKeyAbilityPriceUpdated, req.Payload.AbilityIdentity)), nil
}

func (arena *Arena) LiveVoteCountUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	return messagebus.BusKey(HubKeyLiveVoteCountUpdated), nil
}

func (arena *Arena) WarMachineLocationUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	return messagebus.BusKey(HubKeyWarMachineLocationUpdated), nil
}

const HubKeySpoilOfWarUpdated hub.HubCommandKey = "SPOIL:OF:WAR:UPDATED"

func (arena *Arena) SpoilOfWarUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "SpoilOfWarUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	return messagebus.BusKey(HubKeySpoilOfWarUpdated), nil
}

const HubKeGabsBribingWinnerSubscribe hub.HubCommandKey = "BRIBE:WINNER:SUBSCRIBE"

// GabsBribingWinnerSubscribe subscribe on winner notification
func (arena *Arena) GabsBribingWinnerSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID))

	return req.TransactionID, busKey, nil
}

func (arena *Arena) SendSettings(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to unmarshal json payload for send settings subscribe")
	}

	if needProcess {
		// response game setting, if current battle exists
		if arena.currentBattle() != nil {
			reply(UpdatePayload(arena.currentBattle()))
		}
	}

	return req.TransactionID, messagebus.BusKey(HubKeyGameSettingsUpdated), nil
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

type BattleEndPayload struct {
	WinningWarMachines []struct {
		Hash   string `json:"hash"`
		Health int    `json:"health"`
	} `json:"winningWarMachines"`
	BattleID     string `json:"battleID"`
	WinCondition string `json:"winCondition"`
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

func (arena *Arena) start() {
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("Panic! Panic! Panic! Panic on battle arena!")
		}
	}()

	ctx := context.Background()
	arena.beginBattle()

	for {
		_, payload, err := arena.socket.Read(ctx)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("empty game client disconnected")
			break
		}
		btl := arena.currentBattle()
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

			gamelog.L.Info().Str("game_client_data", string(data)).Int("message_type", int(mt)).Msg("game client message")

			switch msg.BattleCommand {
			case "BATTLE:START":
				var dataPayload *BattleStartPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message payload")
					continue
				}

				// todocheck
				//gameClientBuildNo, err := strconv.ParseUint(dataPayload.ClientBuildNo, 10, 64)
				//if err != nil {
				//	gamelog.L.Panic().Str("game_client_build_no", dataPayload.ClientBuildNo).Msg("invalid game client build number received")
				//}
				//
				//if gameClientBuildNo < arena.gameClientMinimumBuildNo {
				//	gamelog.L.Panic().Str("current_game_client_build", dataPayload.ClientBuildNo).Uint64("minimum_game_client_build", arena.gameClientMinimumBuildNo).Msg("unsupported game client build number")
				//}

				err = btl.preIntro(dataPayload)
				if err != nil {
					gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("battle start load out has failed")
					return
				}
			case "BATTLE:INTRO_FINISHED":
				btl.start()
			case "BATTLE:WAR_MACHINE_DESTROYED":
				var dataPayload BattleWMDestroyedPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
					continue
				}
				btl.Destroyed(&dataPayload)
			case "BATTLE:END":
				var dataPayload *BattleEndPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message warmachine destroyed payload")
					continue
				}
				btl.end(dataPayload)
				//TODO: this needs to be triggered by a message from the game client
				time.Sleep(time.Second * 30)
				arena.beginBattle()
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
	gm, err := db.GameMapGetRandom(context.Background(), arena.conn)
	if err != nil {
		gamelog.L.Err(err).Msg("unable to get random map")
		return
	}

	gameMap := &server.GameMap{
		ID:            uuid.Must(uuid.FromString(gm.ID)),
		Name:          gm.Name,
		ImageUrl:      gm.ImageURL,
		MaxSpawns:     gm.MaxSpawns,
		Width:         gm.Width,
		Height:        gm.Height,
		CellsX:        gm.CellsX,
		CellsY:        gm.CellsY,
		TopPixels:     gm.TopPixels,
		LeftPixels:    gm.LeftPixels,
		Scale:         gm.Scale,
		DisabledCells: gm.DisabledCells,
	}

	var battleID string
	var battle *boiler.Battle
	inserted := false

	// query last battle
	lastBattle, err := boiler.Battles(qm.OrderBy("battle_number DESC"), qm.Limit(1)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("not able to load previous battle")
	}

	// if last battle is ended or does not exist, create a new battle
	if lastBattle == nil || errors.Is(err, sql.ErrNoRows) || lastBattle.EndedAt.Valid {

		battleID = uuid.Must(uuid.NewV4()).String()
		battle = &boiler.Battle{
			ID:                   battleID,
			GameMapID:            gameMap.ID.String(),
			StartedAt:            time.Now(),
			StartedBattleSeconds: decimal.NullDecimal{decimal.Zero, true},
		}
		// set started battle second to the last battle is ended properly
		if lastBattle != nil && lastBattle.EndedBattleSeconds.Valid {
			battle.StartedBattleSeconds.Decimal = lastBattle.EndedBattleSeconds.Decimal
		}

	} else {

		// if there is an unfinished battle
		battle = lastBattle
		battleID = lastBattle.ID

		inserted = true

		// if last battle does not have started at
		if !battle.StartedBattleSeconds.Valid {
			battle.StartedBattleSeconds = decimal.NullDecimal{decimal.Zero, true}
		}
	}

	var bs decimal.Decimal
	if inserted {
		bs = battle.StartedBattleSeconds.Decimal
	} else {
		if lastBattle != nil && lastBattle.EndedBattleSeconds.Valid {
			bs = lastBattle.EndedBattleSeconds.Decimal
			battle.StartedBattleSeconds = lastBattle.EndedBattleSeconds
		} else {
			bs = decimal.NewFromInt(0)
		}
	}

	btl := &Battle{
		arena:          arena,
		MapName:        gameMap.Name,
		gameMap:        gameMap,
		BattleID:       battleID,
		Battle:         battle,
		inserted:       inserted,
		_battleSeconds: bs,
		stage:          atomic.NewInt32(BattleStagStart),
		users: usersMap{
			m: make(map[uuid.UUID]*BattleUser),
		},
		destroyedWarMachineMap: make(map[byte]*WMDestroyedRecord),
		viewerCountInputChan:   make(chan *ViewerLiveCount),
	}

	err = btl.Load()
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to load out mechs")
	}

	// order the mechs by facton id

	// set user online debounce
	go btl.debounceSendingViewerCount(func(result ViewerLiveCount, btl *Battle) {
		btl.users.Send(HubKeyViewerLiveCountUpdated, result)
	})

	arena.storeCurrentBattle(btl)
	arena.Message(BATTLEINIT, btl)

	go arena.NotifyUpcomingWarMachines()
}

const HubKeyUserStatSubscribe hub.HubCommandKey = "USER:STAT:SUBSCRIBE"

func (arena *Arena) UserStatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(client.Identifier())
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if needProcess {
		us, err := db.UserStatsGet(userID.String())
		if err != nil {
			return "", "", terror.Error(err, "failed to get user stats")
		}
		if us != nil {
			reply(us)
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, client.Identifier())), nil
}
