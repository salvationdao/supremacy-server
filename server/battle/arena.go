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
	"server/rpcclient"
	"server/telegram"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/volatiletech/sqlboiler/v4/boil"

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
	connected                *atomic.Bool
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

func (arena *Arena) IsClientConnected() error {
	connected := arena.connected.Load()
	if !connected {
		return fmt.Errorf("no gameclient connected")
	}
	return nil
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

func (arena *Arena) currentBattleWarMachineIDs() []uuid.UUID {
	arena.RLock()
	defer arena.RUnlock()

	if arena._currentBattle == nil {
		return []uuid.UUID{}
	}

	return arena._currentBattle.warMachineIDs
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
		conn:                     opts.Conn,
		connected:                atomic.NewBool(false),
		timeout:                  opts.Timeout,
		messageBus:               opts.MessageBus,
		RPCClient:                opts.RPCClient,
		sms:                      opts.SMS,
		gameClientMinimumBuildNo: opts.GameClientMinimumBuildNo,
		telegram:                 opts.Telegram,
	}

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

	opts.SecureUserFactionCommand(HubKeyAssetMany, arena.AssetManyHandler)

	// TODO: handle insurance and repair
	//opts.SecureUserFactionCommand(HubKeyAssetRepairPayFee, arena.AssetRepairPayFeeHandler)
	//opts.SecureUserFactionCommand(HubKeyAssetRepairStatus, arena.AssetRepairStatusHandler)

	opts.SecureUserCommand(HubKeyGameUserOnline, arena.UserOnline)
	opts.SecureUserCommand(HubKeyPlayerRankGet, arena.PlayerRankGet)
	opts.SubscribeCommand(HubKeyWarMachineDestroyedUpdated, arena.WarMachineDestroyedUpdatedSubscribeHandler)

	// subscribe functions
	opts.SubscribeCommand(HubKeyGameSettingsUpdated, arena.SendSettings)

	opts.SubscribeCommand(HubKeyGameNotification, arena.GameNotificationSubscribeHandler)
	opts.SecureUserSubscribeCommand(HubKeyMultiplierSubscribe, arena.HubKeyMultiplierUpdate)

	opts.SecureUserSubscribeCommand(HubKeyUserStatSubscribe, arena.UserStatUpdatedSubscribeHandler)

	// battle ability related (bribing)
	opts.SecureUserFactionCommand(HubKeyBattleAbilityBribe, arena.BattleAbilityBribe)
	opts.SecureUserFactionCommand(HubKeyAbilityLocationSelect, arena.AbilityLocationSelect)
	opts.SubscribeCommand(HubKeGabsBribeStageUpdateSubscribe, arena.GabsBribeStageSubscribe)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribingWinnerSubscribe, arena.GabsBribingWinnerSubscribe)
	opts.SubscribeCommand(HubKeyBattleAbilityUpdated, arena.BattleAbilityUpdateSubscribeHandler)

	// faction unique ability related (sup contribution)
	opts.SecureUserFactionCommand(HubKeFactionUniqueAbilityContribute, arena.FactionUniqueAbilityContribute)
	opts.SecureUserFactionSubscribeCommand(HubKeyFactionUniqueAbilitiesUpdated, arena.FactionAbilitiesUpdateSubscribeHandler)
	opts.SecureUserFactionSubscribeCommand(HubKeyWarMachineAbilitiesUpdated, arena.WarMachineAbilitiesUpdateSubscribeHandler)

	// player ability related
	opts.SecureUserFactionCommand(HubKeyPlayerAbilityUse, arena.PlayerAbilityUse)

	// net message subscribe
	opts.NetSubscribeCommand(HubKeyBattleAbilityProgressBarUpdated, arena.BattleAbilityProgressBarUpdateSubscribeHandler)
	opts.NetSecureUserFactionSubscribeCommand(HubKeyAbilityPriceUpdated, arena.FactionAbilityPriceUpdateSubscribeHandler)
	opts.NetSubscribeCommand(HubKeyWarMachineLocationUpdated, arena.WarMachineLocationUpdateSubscribeHandler)
	opts.NetSubscribeCommand(HubKeyLiveVoteCountUpdated, arena.LiveVoteCountUpdateSubscribeHandler)
	opts.NetSubscribeCommand(HubKeySpoilOfWarUpdated, arena.SpoilOfWarUpdateSubscribeHandler)

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

func (btl *Battle) QueueDefaultMechs() error {
	defMechs, err := db.DefaultMechs()
	if err != nil {
		return err
	}

	var req QueueJoinRequest
	ctx := context.Background()
	var reply hub.ReplyFunc = func(_ interface{}) {}
	for _, mech := range defMechs {
		mech.Name = helpers.GenerateStupidName()
		_, _ = mech.Update(gamedb.StdConn, boil.Whitelist(boiler.MechColumns.Label))
		req = QueueJoinRequest{
			HubCommandRequest: nil,
			Payload: struct {
				AssetHash                   string `json:"asset_hash"`
				NeedInsured                 bool   `json:"need_insured"`
				EnablePushNotifications     bool   `json:"enable_push_notifications,omitempty"`
				MobileNumber                string `json:"mobile_number,omitempty"`
				EnableTelegramNotifications bool   `json:"enable_telegram_notifications"`
			}{
				AssetHash:                   mech.Hash,
				NeedInsured:                 false,
				EnableTelegramNotifications: false,
				MobileNumber:                "",
				EnablePushNotifications:     false,
			},
		}

		b, _ := json.Marshal(req)

		btl.arena.QueueJoinHandler(ctx, nil, b, uuid.FromStringOrNil(mech.FactionID), reply)
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
		AbilityOfferingID string          `json:"ability_offering_id"`
		Percentage        decimal.Decimal `json:"percentage"` // "0.1", "0.5%", "1%"
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

	// check percentage amount is valid
	if _, ok := MinVotePercentageCost[req.Payload.Percentage.String()]; !ok {
		gamelog.L.Error().Interface("payload", req).
			Str("userID", wsc.Identifier()).
			Str("percentage", req.Payload.Percentage.String()).
			Msg("invalid vote percentage amount received")
		return terror.Error(err, "Invalid vote percentage amount received")
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

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		gamelog.L.Error().Str("user id is nil", wsc.Identifier()).Msg("cant make users")
		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle().abilities().BribeGabs(factionID, userID, req.Payload.AbilityOfferingID, req.Payload.Percentage, reply)

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

type PlayerAbilityUseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		BlueprintAbilityID string                `json:"blueprint_ability_id"`
		LocationSelectType db.LocationSelectType `json:"location_select_type"`
		StartCoords        *CellLocation         `json:"start_coords"` // used for LINE_SELECT and LOCATION_SELECT abilities
		EndCoords          *CellLocation         `json:"end_coords"`   // used only for LINE_SELECT abilities
		MechHash           *string               `json:"mech_hash"`    // used only for MECH_SELECT abilities
	} `json:"payload"`
}

const HubKeyPlayerAbilityUse hub.HubCommandKey = "PLAYER:ABILITY:USE"

func (arena *Arena) PlayerAbilityUse(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	// skip, if current not battle
	if arena.currentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return nil
	}

	req := &PlayerAbilityUseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(wsc.Identifier())
	if err != nil || userID.IsNil() {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Msgf("can't create uuid from wsc identifier %s", wsc.Identifier())
		return terror.Error(terror.ErrForbidden, "You do not have permission to activate this ability.")
	}

	player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(userID.String()), qm.Load(boiler.PlayerRels.Faction)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("userID", userID.String()).Msg("could not find player from given user ID")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	pa, err := boiler.PlayerAbilities(boiler.PlayerAbilityWhere.BlueprintID.EQ(req.Payload.BlueprintAbilityID), boiler.PlayerAbilityWhere.OwnerID.EQ(player.ID), qm.OrderBy(fmt.Sprintf("%s asc", boiler.PlayerAbilityColumns.PurchasedAt))).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("abilityID", req.Payload.BlueprintAbilityID).Msg("failed to get player ability")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	if pa.OwnerID != player.ID {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Str("ability ownerID", pa.OwnerID).Str("abilityID", req.Payload.BlueprintAbilityID).Msgf("player %s tried to execute an ability that wasn't theirs", player.ID)
		return terror.Error(terror.ErrForbidden, "You do not have permission to activate this ability.")
	}

	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the PlayerAbilityUse!", r)
		}
	}()

	currentBattle := arena.currentBattle()
	// check battle end
	if currentBattle.stage.Load() == BattleStageEnd {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("battle stage has en ended")
		return nil
	}

	var event *server.GameAbilityEvent
	switch req.Payload.LocationSelectType {
	case db.LineSelect:
		if req.Payload.StartCoords == nil || req.Payload.EndCoords == nil {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("no start/end coords was provided for executing ability of type %s", db.LineSelect)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when execting this ability.")
		}
		if req.Payload.StartCoords.X < 0 || req.Payload.StartCoords.Y < 0 || req.Payload.EndCoords.X < 0 || req.Payload.EndCoords.Y < 0 {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("invalid start/end coords were provided for executing %s ability", db.LineSelect)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(pa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			GameLocation:        currentBattle.getGameWorldCoordinatesFromCellXY(*req.Payload.StartCoords),
			GameLocationEnd:     currentBattle.getGameWorldCoordinatesFromCellXY(*req.Payload.EndCoords),
		}

		break
	case db.MechSelect:
		if req.Payload.MechHash == nil || *req.Payload.MechHash == "" {
			gamelog.L.Error().Interface("request payload", req.Payload).Err(err).Msgf("no mech hash was provided for executing ability of type %s", db.MechSelect)
			return terror.Error(terror.ErrInvalidInput, "Mech hash must be provided to execute this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(pa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			WarMachineHash:      req.Payload.MechHash,
		}

		break
	case db.LocationSelect:
		if req.Payload.StartCoords == nil {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("no start coords was provided for executing ability of type %s", db.LocationSelect)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when execting this ability.")
		}
		if req.Payload.StartCoords.X < 0 || req.Payload.StartCoords.Y < 0 {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("invalid start coords were provided for executing %s ability", db.LocationSelect)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(pa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			GameLocation:        currentBattle.getGameWorldCoordinatesFromCellXY(*req.Payload.StartCoords),
		}
		break
	case db.Global:
		break
	default:
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Interface("request payload", req.Payload).Msg("no location select type was provided when activating a player ability")
		return terror.Error(terror.ErrInvalidInput, "Something went wrong while activating this ability. Please try again, or contact support if this issue persists.")
	}

	if event == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Interface("request payload", req.Payload).Msg("game ability event is nil for some reason")
		return terror.Error(terror.ErrInvalidInput, "Something went wrong while activating this ability. Please try again, or contact support if this issue persists.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	defer tx.Rollback()

	spew.Dump(event)
	// Create consumed_abilities entry
	ca := boiler.ConsumedAbility{
		BattleID:            currentBattle.BattleID,
		ConsumedBy:          player.ID,
		BlueprintID:         pa.BlueprintID,
		GameClientAbilityID: pa.GameClientAbilityID,
		Label:               pa.Label,
		Colour:              pa.Colour,
		ImageURL:            pa.ImageURL,
		Description:         pa.Description,
		TextColour:          pa.TextColour,
		LocationSelectType:  pa.LocationSelectType,
		ConsumedAt:          time.Now(),
	}
	err = ca.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("consumedAbility", ca).Msg("failed to created consumed ability entry")
		return err
	}

	// Delete player_abilities entry
	_, err = pa.Delete(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to delete player ability")
		return err
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue executing player ability, please try again or contact support.")
	}
	reply(true)

	currentBattle.arena.Message("BATTLE:ABILITY", event)
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeyPlayerAbilitiesListUpdated, userID)), true)

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

	factionID := uuid.Nil

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if !userID.IsNil() {
		// update player's faction id
		factionID, err = GetPlayerFactionID(userID)
		if err != nil || factionID.IsNil() {
			gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find player from user id")
			return "", "", terror.Error(terror.ErrForbidden)
		}
	}

	// HACK: update player's faction id to red mountain, enable player to view the vote panel
	// for those who first join the battle arena and want to understand how things work
	if factionID.IsNil() {
		factionID = uuid.UUID(server.RedMountainFactionID)
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
		AbilityIdentity   string          `json:"ability_identity"`
		AbilityOfferingID string          `json:"ability_offering_id"`
		Percentage        decimal.Decimal `json:"percentage"` // "0.1", "0.5%", "1%"
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
			Str("userID", wsc.Identifier()).Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	// check percentage amount is valid
	if _, ok := MinVotePercentageCost[req.Payload.Percentage.String()]; !ok {
		gamelog.L.Error().Interface("payload", req).
			Str("userID", wsc.Identifier()).
			Str("percentage", req.Payload.Percentage.String()).
			Msg("invalid vote percentage amount received")
		return terror.Error(err, "Invalid vote percentage amount received")
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

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		gamelog.L.Error().Str("percentage", req.Payload.Percentage.String()).
			Str("userID", wsc.Identifier()).Msg("unable to contribute forbidden")
		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle().abilities().AbilityContribute(factionID, userID, req.Payload.AbilityIdentity, req.Payload.AbilityOfferingID, req.Payload.Percentage, reply)

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

func (arena *Arena) BattleAbilityProgressBarUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "BattleAbilityProgressBarUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")

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

type BattleWMPickupPayload struct {
	WarMachineHash string `json:"warMachineHash"`
	EventID        string `json:"eventID"`
	BattleID       string `json:"battleID"`
}

func (arena *Arena) start() {
	defer func() {
		if r := recover(); r != nil {

			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic on battle arena!", r)
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
			case "BATTLE:MAP_DETAILS":
				var dataPayload *MapDetailsPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message payload")
					continue
				}

				// update map detail
				btl.storeGameMap(dataPayload.Details)

			case "BATTLE:START":
				var dataPayload *BattleStartPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
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
			case "BATTLE:WAR_MACHINE_PICKUP":
				var dataPayload BattleWMPickupPayload
				if err := json.Unmarshal([]byte(msg.Payload), &dataPayload); err != nil {
					gamelog.L.Warn().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal battle message warmachine pickup payload")
					continue
				}
				btl.Pickup(&dataPayload)
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
	lastBattle, err := boiler.Battles(qm.OrderBy("battle_number DESC"), qm.Limit(1)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("not able to load previous battle")
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

		inserted = true
	}

	btl := &Battle{
		arena:    arena,
		MapName:  gameMap.Name,
		gameMap:  gameMap,
		BattleID: battleID,
		Battle:   battle,
		inserted: inserted,
		stage:    atomic.NewInt32(BattleStagStart),
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
