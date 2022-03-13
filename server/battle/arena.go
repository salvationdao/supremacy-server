package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/passport"
	"server/rpcclient"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"nhooyr.io/websocket"
)

type Arena struct {
	conn          db.Conn
	socket        *websocket.Conn
	timeout       time.Duration
	messageBus    *messagebus.MessageBus
	netMessageBus *messagebus.NetBus
	currentBattle *Battle
	syndicates    map[string]boiler.Faction
	AIPlayers     map[string]db.PlayerWithFaction
	RPCClient     *rpcclient.XrpcClient
	ppClient      *passport.Passport
}

type Opts struct {
	Conn          db.Conn
	Addr          string
	Timeout       time.Duration
	Hub           *hub.Hub
	MessageBus    *messagebus.MessageBus
	NetMessageBus *messagebus.NetBus
	PPClient      *passport.Passport
	RPCClient     *rpcclient.XrpcClient
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
	return [...]string{"JSON", "Tick", "Live Vote Tick", "Viewer Live Count Tick", "Spoils of War Tick", "game ability progress tick", "battle ability progress tick"}[mt]
}

const WSJoinQueue hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:JOIN")
const WSLeaveQueue hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:LEAVE")
const WSQueueStatus hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:STATUS")
const WSWarMachineQueueStatus hub.HubCommandKey = hub.HubCommandKey("WAR:MACHINE:QUEUE:STATUS")

func NewArena(opts *Opts) *Arena {
	l, err := net.Listen("tcp", opts.Addr)

	if err != nil {
		gamelog.L.Fatal().Str("Addr", opts.Addr).Err(err).Msg("unable to bind Arena to Battle Server address")
	}

	arena := &Arena{
		conn: opts.Conn,
	}

	arena.timeout = opts.Timeout
	arena.netMessageBus = opts.NetMessageBus
	arena.messageBus = opts.MessageBus
	arena.ppClient = opts.PPClient
	arena.RPCClient = opts.RPCClient

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

	// queue
	opts.SecureUserFactionCommand(WSJoinQueue, arena.Join)
	opts.SecureUserFactionCommand(WSLeaveQueue, arena.Leave)
	opts.SecureUserFactionSubscribeCommand(WSQueueStatus, arena.QueueStatus)
	opts.SecureUserFactionSubscribeCommand(WSWarMachineQueueStatus, arena.WarMachineQueueStatus)

	opts.SecureUserCommand(HubKeyGameUserOnline, arena.UserOnline)
	opts.SubscribeCommand(HubKeyWarMachineDestroyedUpdated, arena.WarMachineDestroyedUpdatedSubscribeHandler)

	// subscribe functions
	opts.SubscribeCommand(HubKeyGameSettingsUpdated, arena.SendSettings)

	opts.SubscribeCommand(HubKeyGameNotification, arena.GameNotificationSubscribeHandler)
	opts.SecureUserSubscribeCommand(HubKeyMultiplierUpdate, arena.HubKeyMultiplierUpdate)
	opts.SecureUserSubscribeCommand(HubKeyViewerLiveCountUpdated, arena.ViewerLiveCountUpdateSubscribeHandler)

	opts.SecureUserSubscribeCommand(HubKeyUserStatSubscribe, arena.UserStatUpdatedSubscribeHandler)

	// battle ability related (bribing)
	opts.SecureUserFactionCommand(HubKeyBattleAbilityBribe, arena.BattleAbilityBribe)
	opts.SecureUserFactionCommand(HubKeyAbilityLocationSelect, arena.AbilityLocationSelect)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribeStageUpdateSubscribe, arena.GabsBribeStageSubscribe)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribingWinnerSubscribe, arena.GabsBribingWinnerSubscribe)
	opts.SecureUserFactionSubscribeCommand(HubKeyBattleAbilityUpdated, arena.BattleAbilityUpdateSubscribeHandler)

	opts.SecureUserSubscribeCommand(HubKeyMultiplierMapSubscribe, arena.MultiplierMapSubScribeHandler)

	// faction unique ability related (sup contribution)
	opts.SecureUserFactionCommand(HubKeFactionUniqueAbilityContribute, arena.FactionUniqueAbilityContribute)
	opts.SecureUserFactionSubscribeCommand(HubKeyFactionUniqueAbilitiesUpdated, arena.FactionAbilitiesUpdateSubscribeHandler)
	opts.SecureUserFactionSubscribeCommand(HubKeyWarMachineAbilitiesUpdated, arena.WarMachineAbilitiesUpdateSubscribeHandler)

	// net message subscribe
	opts.NetSecureUserFactionSubscribeCommand(HubKeyBattleAbilityProgressBarUpdated, arena.FactionProgressBarUpdateSubscribeHandler)
	opts.NetSecureUserFactionSubscribeCommand(HubKeyAbilityPriceUpdated, arena.FactionAbilityPriceUpdateSubscribeHandler)
	opts.NetSecureUserFactionSubscribeCommand(HubKeyWarMachineLocationUpdated, arena.WarMachineLocationUpdateSubscribeHandler)
	opts.NetSecureUserFactionSubscribeCommand(HubKeyLiveVoteCountUpdated, arena.LiveVoteCountUpdateSubscribeHandler)
	opts.NetSecureUserSubscribeCommand(HubKeySpoilOfWarUpdated, arena.SpoilOfWarUpdateSubscribeHandler)

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

	defer c.Close(websocket.StatusInternalError, "game client has disconnected")

	arena.Start()
}

func (arena *Arena) SetMessageBus(mb *messagebus.MessageBus, nb *messagebus.NetBus) {
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
	// skip, if current not battle
	if arena.currentBattle == nil {
		return nil
	}

	req := &BribeGabRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	d, err := decimal.NewFromString(req.Payload.Amount)
	if err != nil {
		return terror.Error(err, "Failed to parse string to decimal.deciaml")
	}
	amount := d.Mul(decimal.New(1, 18))

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle.abilities.BribeGabs(factionID, userID, amount)

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
	if arena.currentBattle == nil {
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

	if arena.currentBattle.abilities == nil {
		gamelog.L.Error().Msg("abilities is nil even with current battle not being nil")
		return terror.Error(terror.ErrForbidden)
	}

	err = arena.currentBattle.abilities.LocationSelect(userID, req.Payload.XIndex, req.Payload.YIndex)
	if err != nil {
		gamelog.L.Warn().Err(err).Msgf("can't create uuid from wsc identifier %s", wsc.Identifier())
		return terror.Error(err)
	}

	return nil
}

type MultiplierMapResponse struct {
	Multipliers      []*db.Multipliers `json:"multipliers"`
	CitizenPlayerIDs []uuid.UUID       `json:"citizen_player_ids"`
}

const HubKeyMultiplierMapSubscribe hub.HubCommandKey = "MULTIPLIER:MAP:SUBSCRIBE"

func (arena *Arena) MultiplierMapSubScribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if arena.currentBattle != nil {
		multipliers, err := db.PlayerMultipliers(arena.currentBattle.BattleNumber)
		if err != nil {
			return "", "", terror.Error(err, "unable to retrieve multipliers")
		}

		for _, m := range multipliers {
			m.TotalMultiplier = m.TotalMultiplier.Shift(-1)
		}

		// get the citizen list
		citizenPlayerIDs, err := db.CitizenPlayerIDs(arena.currentBattle.BattleNumber)
		if err != nil {
			return "", "", terror.Error(err)
		}

		reply(&MultiplierMapResponse{
			Multipliers:      multipliers,
			CitizenPlayerIDs: citizenPlayerIDs,
		})
	}

	return req.TransactionID, messagebus.BusKey(HubKeyMultiplierMapSubscribe), nil
}

const HubKeyBattleAbilityUpdated hub.HubCommandKey = "BATTLE:ABILITY:UPDATED"

func (arena *Arena) BattleAbilityUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

	// return data if, current battle is not null
	if arena.currentBattle != nil {
		btl := arena.currentBattle
		if btl.abilities != nil {
			_, abil := btl.abilities.FactionBattleAbilityGet(factionID)
			reply(abil)
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
	if arena == nil || arena.currentBattle == nil || factionID.IsNil() {
		gamelog.L.Error().Bool("arena", arena == nil).
			Bool("factionID", factionID.IsNil()).
			Bool("current_battle", arena.currentBattle == nil).
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

	arena.currentBattle.abilities.AbilityContribute(factionID, userID, req.Payload.AbilityIdentity, amount)

	return nil
}

const HubKeyFactionUniqueAbilitiesUpdated hub.HubCommandKey = "FACTION:UNIQUE:ABILITIES:UPDATED"

func (arena *Arena) FactionAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

	// skip, if user is non faction or Zaibatsu faction
	if factionID.IsNil() || factionID.String() == server.ZaibatsuFactionID.String() {
		return "", "", nil
	}

	// return data if, current battle is not null
	if arena.currentBattle != nil {
		btl := arena.currentBattle
		if btl.abilities != nil {
			reply(btl.abilities.FactionUniqueAbilitiesGet(factionID))
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
func (arena *Arena) WarMachineAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

	// skip, if user is non faction or not Zaibatsu faction
	if factionID.IsNil() || factionID.String() != server.ZaibatsuFactionID.String() {
		return "", "", nil
	}

	// NOTE: current only return faction unique ability
	// get war machine ability
	if arena.currentBattle != nil {
		btl := arena.currentBattle
		if btl.abilities != nil {
			ga := btl.abilities.WarMachineAbilitiesGet(factionID, req.Payload.Hash)
			if ga != nil {
				reply(ga)
			}
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineAbilitiesUpdated, req.Payload.Hash))
	return req.TransactionID, busKey, nil
}

func (arena *Arena) UserOnline(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	if arena.currentBattle == nil {
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
	if err != nil || user == nil || user.R.Faction == nil {
		gamelog.L.Error().Err(err).Msg("invalid input data")
		return terror.Error(terror.ErrInvalidInput)
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

	arena.currentBattle.userOnline(battleUser, wsc)
	return nil
}

type WarMachineDestroyedUpdatedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ParticipantID byte `json:"participantID"`
	} `json:"payload"`
}

const HubKeyWarMachineDestroyedUpdated = hub.HubCommandKey("WAR:MACHINE:DESTROYED:UPDATED")

func (arena *Arena) WarMachineDestroyedUpdatedSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &WarMachineDestroyedUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if arena.currentBattle != nil {
		if wmd, ok := arena.currentBattle.destroyedWarMachineMap[req.Payload.ParticipantID]; ok {
			reply(wmd)
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%x", HubKeyWarMachineDestroyedUpdated, req.Payload.ParticipantID)), nil
}

const HubKeGabsBribeStageUpdateSubscribe hub.HubCommandKey = "BRIBE:STAGE:UPDATED:SUBSCRIBE"

// GabsBribeStageSubscribe subscribe on bribing stage change
func (arena *Arena) GabsBribeStageSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	// return data if, current battle is not null
	if arena.currentBattle != nil {
		btl := arena.currentBattle
		if btl.abilities != nil {
			reply(btl.abilities.BribeStageGet())
		}
	}

	return req.TransactionID, messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), nil
}

const HubKeyBattleAbilityProgressBarUpdated hub.HubCommandKey = "BATTLE:ABILITY:PROGRESS:BAR:UPDATED"

func (arena *Arena) FactionProgressBarUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "FactionProgressBarUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")

	return messagebus.NetBusKey(HubKeyBattleAbilityProgressBarUpdated), nil
}

const HubKeyAbilityPriceUpdated hub.HubCommandKey = "ABILITY:PRICE:UPDATED"

type AbilityPriceUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityIdentity string `json:"ability_identity"`
	} `json:"payload"`
}

func (arena *Arena) FactionAbilityPriceUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	req := &AbilityPriceUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", terror.Error(err, "Invalid request received")
	}

	return messagebus.NetBusKey(fmt.Sprintf("%s,%s", HubKeyAbilityPriceUpdated, req.Payload.AbilityIdentity)), nil
}

func (arena *Arena) LiveVoteCountUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	return messagebus.NetBusKey(HubKeyLiveVoteCountUpdated), nil
}

func (arena *Arena) WarMachineLocationUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	return messagebus.NetBusKey(HubKeyWarMachineLocationUpdated), nil
}

const HubKeySpoilOfWarUpdated hub.HubCommandKey = "SPOIL:OF:WAR:UPDATED"

func (arena *Arena) SpoilOfWarUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "SpoilOfWarUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	return messagebus.NetBusKey(HubKeySpoilOfWarUpdated), nil
}

const HubKeGabsBribingWinnerSubscribe hub.HubCommandKey = "BRIBE:WINNER:SUBSCRIBE"

// GabsBribingWinnerSubscribe subscribe on winner notification
func (arena *Arena) GabsBribingWinnerSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

func (arena *Arena) SendSettings(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to unmarshal json payload for send settings subscribe")
	}
	if arena.currentBattle == nil {
		return "", "", fmt.Errorf("battle is not currently running")
	}

	btl := arena.currentBattle
	reply(btl.updatePayload())

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
	BattleID string `json:"battleID"`
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

func (arena *Arena) init() {
	btl := arena.Battle()
	arena.Message(BATTLEINIT, btl)
	arena.currentBattle = btl
}

//listen listens for new commands and blocks indefinitely
func (arena *Arena) start() {
	ctx := context.Background()
	arena.init()

	for {
		_, payload, err := arena.socket.Read(ctx)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("empty game client disconnected")
			break
		}
		btl := arena.currentBattle
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
				btl.start(dataPayload)
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
				time.Sleep(time.Second * 20)
				arena.init()
			default:
				gamelog.L.Warn().Str("battleCommand", msg.BattleCommand).Err(err).Msg("Battle Arena WS: no command response")
			}
		case Tick:
			btl.Tick(payload)
		default:
			gamelog.L.Warn().Str("MessageType", MessageType(mt).String()).Err(err).Msg("Battle Arena WS: no message response")
		}
	}
}

func (arena *Arena) Battle() *Battle {
	gameMap, err := db.GameMapGetRandom(context.Background(), arena.conn)
	if err != nil {
		gamelog.L.Err(err).Msg("unable to get random map")
		return nil
	}
	id := uuid.Must(uuid.NewV4())

	btl := &Battle{
		arena:   arena,
		MapName: gameMap.Name,
		gameMap: gameMap,
		Battle: &boiler.Battle{
			ID:        id.String(),
			GameMapID: gameMap.ID.String(),
			StartedAt: time.Now(),
		},
		stage: BattleStagStart,
		users: usersMap{
			m: make(map[uuid.UUID]*BattleUser),
		},
		destroyedWarMachineMap: make(map[byte]*WMDestroyedRecord),
	}

	err = btl.Battle.Insert(gamedb.StdConn, boil.Infer())
	btl.BattleID = btl.ID

	if err != nil {
		gamelog.L.Panic().Interface("battle", btl).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to insert Battle into database")
		return nil
	}

	err = btl.Load()
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to load out mechs")
	}

	bmd := make([]*db.BattleMechData, len(btl.WarMachines))

	factions := map[uuid.UUID]*boiler.Faction{}

	for i, wm := range btl.WarMachines {
		mechID, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().Str("ownerID", wm.ID).Err(err).Msg("unable to convert owner id from string")
			return nil
		}

		ownerID, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().Str("ownerID", wm.OwnedByID).Err(err).Msg("unable to convert owner id from string")
			return nil
		}

		factionID, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().Str("factionID", wm.FactionID).Err(err).Msg("unable to convert faction id from string")
			return nil
		}

		bmd[i] = &db.BattleMechData{
			MechID:    mechID,
			OwnerID:   ownerID,
			FactionID: factionID,
		}

		_, ok := factions[factionID]
		if !ok {
			faction, err := boiler.FindFaction(gamedb.StdConn, factionID.String())
			if err != nil {
				gamelog.L.Error().
					Str("Battle ID", btl.ID).
					Str("Faction ID", factionID.String()).
					Err(err).Msg("unable to retrieve faction from database")

			}
			factions[factionID] = faction
		}
	}

	btl.factions = factions

	err = db.BattleMechs(btl.Battle, bmd)
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID).Err(err).Msg("unable to insert battle into database")
		//TODO: something more dramatic
	}

	return btl
}

const HubKeyUserStatSubscribe hub.HubCommandKey = "USER:STAT:SUBSCRIBE"

func (uc *Arena) UserStatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {

	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(client.Identifier())
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}
	us, err := db.UserStatGet(ctx, uc.conn, server.UserID(userID))
	if err != nil {
		return "", "", terror.Error(err, "failed to get user")
	}

	if us != nil {
		reply(us)
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, client.Identifier())), nil
}
