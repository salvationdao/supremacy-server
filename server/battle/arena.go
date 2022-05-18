package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-syndicate/ws"
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

	"github.com/volatiletech/null/v8"

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
	SecureUserCommander      *ws.Commander
	SecureFactionCommander   *ws.Commander
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

func (arena *Arena) currentBattleWarMachine(participantID int) *WarMachine {
	arena.RLock()
	defer arena.RUnlock()

	if arena._currentBattle == nil {
		return nil
	}

	for _, wm := range arena._currentBattle.WarMachines {
		if int(wm.ParticipantID) == participantID {
			return wm
		}
	}

	return nil
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
		connected:                atomic.NewBool(false),
		timeout:                  opts.Timeout,
		messageBus:               opts.MessageBus,
		RPCClient:                opts.RPCClient,
		sms:                      opts.SMS,
		gameClientMinimumBuildNo: opts.GameClientMinimumBuildNo,
		telegram:                 opts.Telegram,
	}

	arena.SecureUserCommander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})
	arena.SecureFactionCommander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})

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
	arena.SecureUserFactionCommand(WSQueueJoin, arena.QueueJoinHandler)
	arena.SecureUserFactionCommand(WSQueueLeave, arena.QueueLeaveHandler)
	arena.SecureUserFactionCommand(WSAssetQueueStatus, arena.AssetQueueStatusHandler)
	arena.SecureUserFactionCommand(WSAssetQueueStatusList, arena.AssetQueueStatusListHandler)

	arena.SecureUserFactionCommand(HubKeyAssetMany, arena.AssetManyHandler)

	// TODO: handle insurance and repair
	//arena.SecureUserFactionCommand(HubKeyAssetRepairPayFee, arena.AssetRepairPayFeeHandler)
	//arena.SecureUserFactionCommand(HubKeyAssetRepairStatus, arena.AssetRepairStatusHandler)

	// TODO: handle player ability use
	//arena.SecureUserCommand(HubKeyPlayerAbilityUse, arena.PlayerAbilityUse)

	// battle ability related (bribing)
	arena.SecureUserFactionCommand(HubKeyBattleAbilityBribe, arena.BattleAbilityBribe)
	arena.SecureUserFactionCommand(HubKeyAbilityLocationSelect, arena.AbilityLocationSelect)

	// faction unique ability related (sup contribution)
	arena.SecureUserFactionCommand(HubKeFactionUniqueAbilityContribute, arena.FactionUniqueAbilityContribute)

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

type AuthMiddleware func(required bool, userIDMustMatch bool) func(next http.Handler) http.Handler
type AuthFactionMiddleware func(factionIDMustMatch bool) func(next http.Handler) http.Handler

const HubKeyBribingWinnerSubscribe = "BRIBE:WINNER:SUBSCRIBE"

func (arena *Arena) Route(authUserWS AuthMiddleware, authFactionWS AuthFactionMiddleware) func(s *ws.Server) {
	return func(s *ws.Server) {

		s.WS("/*", HubKeyGameSettingsUpdated, arena.SendSettings)
		s.WS("/notification", HubKeyGameNotification, nil)
		s.WS("/bribe_stage", HubKeyBribeStageUpdateSubscribe, arena.BribeStageSubscribe)
		s.WS("/live_data", "", nil)

		// handle mech stat and destroyed
		s.Mount("/mech/{slotNumber}", ws.NewServer(func(s *ws.Server) {
			s.Use(func(next http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					slotNumber := chi.URLParam(r, "slotNumber")
					if slotNumber == "" {
						http.Error(w, "no slot number", http.StatusBadRequest)
						return
					}
					ctx := context.WithValue(r.Context(), "slotNumber", slotNumber)
					*r = *r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
				return http.HandlerFunc(fn)
			})
			s.WS("/*", HubKeyWarMachineStatUpdated, arena.WarMachineStatUpdatedSubscribe)
		}))

		s.Mount("/faction/{faction_id}", ws.NewServer(func(s *ws.Server) {
			s.Use(authFactionWS(true))
			s.WS("/*", "", nil)
			s.Mount("/faction_commander", arena.SecureFactionCommander)
			s.WS("/queue", WSQueueStatusSubscribe, server.MustSecureFaction(arena.QueueStatusSubscribeHandler))

			s.Mount("/ability", ws.NewServer(func(s *ws.Server) {
				s.WS("/*", HubKeyBattleAbilityUpdated, server.MustSecureFaction(arena.BattleAbilityUpdateSubscribeHandler))
				s.WS("/faction", HubKeyFactionUniqueAbilitiesUpdated, server.MustSecureFaction(arena.FactionAbilitiesUpdateSubscribeHandler))
				s.Mount("/mech/{slotNumber}", ws.NewServer(func(s *ws.Server) {
					s.Use(func(next http.Handler) http.Handler {
						fn := func(w http.ResponseWriter, r *http.Request) {
							slotNumber := chi.URLParam(r, "slotNumber")
							if slotNumber == "" {
								http.Error(w, "no slot number", http.StatusBadRequest)
								return
							}
							ctx := context.WithValue(r.Context(), "slotNumber", slotNumber)
							*r = *r.WithContext(ctx)
							next.ServeHTTP(w, r)
							return
						}
						return http.HandlerFunc(fn)
					})
					s.WS("/*", HubKeyWarMachineAbilitiesUpdated, server.MustSecureFaction(arena.WarMachineAbilitiesUpdateSubscribeHandler))
				}))
			}))
		}))

		s.Mount("/user/{user_id}", ws.NewServer(func(s *ws.Server) {
			s.Use(authUserWS(true, true))
			s.Mount("/user_commander", arena.SecureUserCommander)
		}))
	}
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

	for _, mech := range defMechs {
		// rename default mech
		mech.Name = helpers.GenerateStupidName()
		_, _ = mech.Update(gamedb.StdConn, boil.Whitelist(boiler.MechColumns.Label))

		// insert default mech into battle
		ownerID, err := uuid.FromString(mech.OwnerID)
		if err != nil {
			gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
			return err
		}

		existMech, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mech.ID)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("mech_id", mech.ID).Err(err).Msg("check mech exists in queue")
			return terror.Error(err, "Failed to check whether mech is in the battle queue")
		}

		if existMech != nil {
			continue
		}

		result, err := db.QueueLength(uuid.FromStringOrNil(mech.FactionID))
		if err != nil {
			gamelog.L.Error().Interface("factionID", mech.FactionID).Err(err).Msg("unable to retrieve queue length")
			return err
		}

		queueStatus := CalcNextQueueStatus(result)

		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to begin tx")
			return fmt.Errorf(terror.Echo(err))
		}

		defer tx.Rollback()

		bc := &boiler.BattleContract{
			MechID:         mech.ID,
			FactionID:      mech.FactionID,
			PlayerID:       ownerID.String(),
			ContractReward: queueStatus.ContractReward,
			Fee:            queueStatus.QueueCost,
		}
		err = bc.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Interface("mech", mech).
				Str("contractReward", queueStatus.ContractReward.String()).
				Str("queueFee", queueStatus.QueueCost.String()).
				Err(err).Msg("unable to create battle contract")
			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		bq := &boiler.BattleQueue{
			MechID:           mech.ID,
			QueuedAt:         time.Now(),
			FactionID:        mech.FactionID,
			OwnerID:          ownerID.String(),
			BattleContractID: null.StringFrom(bc.ID),
		}

		err = bq.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Interface("mech", mech).
				Err(err).Msg("unable to insert mech into queue")
			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().
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
			gamelog.L.Error().Err(fmt.Errorf("game client has disconnected")).Msg("lost connection to game client")
			c.Close(websocket.StatusInternalError, "game client has disconnected")

			btl := arena.CurrentBattle()
			if btl != nil && btl.spoils != nil {
				btl.spoils.End()
			}
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
		gamelog.L.Error().Str("json", string(payload)).Msg("json unmarshal failed")
		return terror.Error(err, "Invalid request received")
	}

	// check percentage amount is valid
	if _, ok := MinVotePercentageCost[req.Payload.Percentage.String()]; !ok {
		gamelog.L.Error().Interface("payload", req).
			Str("userID", user.ID).
			Str("percentage", req.Payload.Percentage.String()).
			Msg("invalid vote percentage amount received")
		return terror.Error(err, "Invalid vote percentage amount received")
	}

	// check user is banned on limit sups contribution
	isBanned, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		boiler.PunishedPlayerWhere.PlayerID.EQ(user.ID),
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
		return terror.Error(err, "Failed to check player")
	}

	// if limited sups contribute, return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to contribute sups"), "You are banned to contribute sups")
	}

	userID := uuid.FromStringOrNil(user.ID)
	if userID.IsNil() {
		gamelog.L.Error().Str("user id is nil", user.ID).Msg("cant make users")
		return terror.Error(terror.ErrForbidden)
	}

	arena.CurrentBattle().abilities().BribeGabs(factionID, userID, req.Payload.AbilityOfferingID, req.Payload.Percentage, reply)

	return nil
}

type LocationSelectRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		XIndex int `json:"x"`
		YIndex int `json:"y"`
	} `json:"payload"`
}

const HubKeyAbilityLocationSelect = "ABILITY:LOCATION:SELECT"

func (arena *Arena) AbilityLocationSelect(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
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

	if arena.CurrentBattle().abilities == nil {
		gamelog.L.Error().Msg("abilities is nil even with current battle not being nil")
		return terror.Error(terror.ErrForbidden)
	}

	err = arena.CurrentBattle().abilities().LocationSelect(userID, req.Payload.XIndex, req.Payload.YIndex)
	if err != nil {
		gamelog.L.Warn().Err(err).Msgf("Unable to select location")
		return terror.Error(err, "Unable to select location")
	}

	return nil
}

type PlayerAbilityUseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityID          string                `json:"ability_id"` // player ability id
		LocationSelectType db.LocationSelectType `json:"location_select_type"`
		XIndex             int                   `json:"x"`
		YIndex             int                   `json:"y"`
		MechID             string                `json:"mech_id"`
	} `json:"payload"`
}

const HubKeyPlayerAbilityUse = "PLAYER:ABILITY:USE"

func (arena *Arena) PlayerAbilityUse(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return nil
	}

	req := &PlayerAbilityUseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	userUUID := uuid.FromStringOrNil(user.ID)

	player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(user.ID), qm.Load(boiler.PlayerRels.Faction)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("userID", user.ID).Msg("could not find player from given user ID")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	pa, err := boiler.FindPlayerAbility(gamedb.StdConn, req.Payload.AbilityID)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("abilityID", req.Payload.AbilityID).Msg("failed to get player ability")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	if pa.OwnerID != player.ID {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Str("ability ownerID", pa.OwnerID).Str("abilityID", req.Payload.AbilityID).Msgf("player %s tried to execute an ability that wasn't theirs", player.ID)
		return terror.Error(terror.ErrForbidden, "You do not have permission to activate this ability.")
	}

	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the PlayerAbilityUse!", r)
		}
	}()

	currentBattle := arena.CurrentBattle()
	// check battle end
	if currentBattle.stage.Load() == BattleStageEnd {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("battle stage has en ended")
		return nil
	}

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: byte(pa.GameClientAbilityID),
		TriggeredOnCellX:    &req.Payload.XIndex,
		TriggeredOnCellY:    &req.Payload.YIndex,
		TriggeredByUserID:   &userUUID,
		TriggeredByUsername: &player.Username.String,
		EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
		FactionID:           &player.FactionID.String,
	}
	currentBattle.calcTriggeredLocation(event)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	defer tx.Rollback()

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
		LocationSelectType:  null.StringFrom(pa.LocationSelectType),
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

	faction := player.R.Faction
	arena.CurrentBattle().arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
		Type: LocationSelectTypeTrigger,
		X:    &req.Payload.XIndex,
		Y:    &req.Payload.YIndex,
		Ability: &AbilityBrief{
			Label:    pa.Label,
			ImageUrl: pa.ImageURL,
			Colour:   pa.Colour,
		},
		CurrentUser: &UserBrief{
			ID:        userUUID,
			Username:  player.Username.String,
			FactionID: player.FactionID.String,
			Gid:       player.Gid,
			Faction:   faction,
		},
	})

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
	*hub.HubCommandRequest
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
		gamelog.L.Error().Bool("arena", arena == nil).
			Str("factionID", factionID).
			Bool("current_battle", arena.CurrentBattle() == nil).
			Str("userID", user.ID).Msg("unable to find player from user id")
		return nil
	}

	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Interface("payload", req).
			Str("userID", user.ID).Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	// check percentage amount is valid
	if _, ok := MinVotePercentageCost[req.Payload.Percentage.String()]; !ok {
		gamelog.L.Error().Interface("payload", req).
			Str("userID", user.ID).
			Str("percentage", req.Payload.Percentage.String()).
			Msg("invalid vote percentage amount received")
		return terror.Error(err, "Invalid vote percentage amount received")
	}

	// check user is banned on limit sups contribution
	isBanned, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		boiler.PunishedPlayerWhere.PlayerID.EQ(user.ID),
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
		return terror.Error(err, "Failed to check player")
	}

	// if limited sups contribute, return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to contribute sups"), "You are banned to contribute sups")
	}

	userID := uuid.FromStringOrNil(user.ID)
	if userID.IsNil() {
		gamelog.L.Error().Str("percentage", req.Payload.Percentage.String()).
			Str("userID", user.ID).Msg("unable to contribute forbidden")
		return terror.Error(terror.ErrForbidden)
	}

	arena.CurrentBattle().abilities().AbilityContribute(factionID, userID, req.Payload.AbilityIdentity, req.Payload.AbilityOfferingID, req.Payload.Percentage, reply)

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
	slotNumber, ok := ctx.Value("slotNumber").(string)
	if !ok || slotNumber == "" {
		return fmt.Errorf("slot number is required")
	}

	participantID, err := strconv.Atoi(slotNumber)
	if err != nil {
		return fmt.Errorf("invalid participant id")
	}

	wm := arena.currentBattleWarMachine(participantID)

	if wm == nil {
		return fmt.Errorf("war machine not found")
	}
	if wm.FactionID != factionID {
		return fmt.Errorf("war machine faction id does not match")
	}

	reply(wm.Abilities)

	return nil
}

type WarMachineStat struct {
	Position *server.Vector3 `json:"position"`
	Rotation int             `json:"rotation"`
	Health   uint32          `json:"health"`
	Shield   uint32          `json:"shield"`
}

const HubKeyWarMachineStatUpdated = "WAR:MACHINE:STAT:UPDATED"

func (arena *Arena) WarMachineStatUpdatedSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	slotNumber, ok := ctx.Value("slotNumber").(string)
	if !ok || slotNumber == "" {
		return fmt.Errorf("slot number is required")
	}

	participantID, err := strconv.Atoi(slotNumber)
	if err != nil {
		return fmt.Errorf("invalid participant id")
	}

	wm := arena.currentBattleWarMachine(participantID)

	if wm == nil {
		return fmt.Errorf("war machine not found")
	}

	reply(WarMachineStat{
		Position: wm.Position,
		Rotation: wm.Rotation,
		Health:   wm.Health,
		Shield:   wm.Shield,
	})

	return nil
}

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

func (arena *Arena) WarMachineLocationUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, needProcess bool) (messagebus.BusKey, error) {
	return messagebus.BusKey(HubKeyWarMachineLocationUpdated), nil
}

const HubKeySpoilOfWarUpdated = "SPOIL:OF:WAR:UPDATED"

func (arena *Arena) SpoilOfWarUpdateSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	sows, err := db.LastTwoSpoilOfWarAmount()
	if err != nil || len(sows) == 0 {
		gamelog.L.Error().Err(err).Msg("Failed to get last two spoil of war amount")
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
		reply(UpdatePayload(arena.CurrentBattle()))
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
		destroyedWarMachineMap: make(map[string]*WMDestroyedRecord),
		viewerCountInputChan:   make(chan *ViewerLiveCount),
	}

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
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

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
