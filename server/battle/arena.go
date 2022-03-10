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
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
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

	ppClient *passport.Passport
}

type Opts struct {
	Conn          db.Conn
	Addr          string
	Timeout       time.Duration
	Hub           *hub.Hub
	MessageBus    *messagebus.MessageBus
	NetMessageBus *messagebus.NetBus
	PPClient      *passport.Passport
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

	opts.SecureUserFactionCommand(WSJoinQueue, arena.Join)
	opts.Command(HubKeyGameUserOnline, arena.UserOnline)

	// subscribe functions
	opts.SubscribeCommand(HubKeyGameSettingsUpdated, arena.SendSettings)

	opts.SubscribeCommand(HubKeyGameNotification, arena.GameNotificationSubscribeHandler)
	opts.SubscribeCommand(HubKeyMultiplierUpdate, arena.HubKeyMultiplierUpdate)

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
		Amount int64 `json:"amount"` // 1, 25, 100
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

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle.abilities.BribeGabs(factionID, userID, decimal.New(req.Payload.Amount, 18))

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
		return nil
	}

	req := &LocationSelectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	err = arena.currentBattle.abilities.LocationSelect(userID, req.Payload.XIndex, req.Payload.YIndex)
	if err != nil {
		return terror.Error(err)
	}

	return nil
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
		reply(btl.abilities.FactionBattleAbilityGet(factionID))
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBattleAbilityUpdated, factionID.String())), nil
}

type GameAbilityContributeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityIdentity string `json:"ability_identity"`
		Amount          int64  `json:"amount"` // 1, 25, 100
	} `json:"payload"`
}

const HubKeFactionUniqueAbilityContribute hub.HubCommandKey = "FACTION:UNIQUE:ABILITY:CONTRIBUTE"

func (arena *Arena) FactionUniqueAbilityContribute(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	if arena.currentBattle == nil {
		return nil
	}

	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	arena.currentBattle.abilities.AbilityContribute(factionID, userID, req.Payload.AbilityIdentity, decimal.New(req.Payload.Amount, 18))

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
		reply(btl.abilities.FactionUniqueAbilitiesGet(factionID))
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
		ga := btl.abilities.WarMachineAbilitiesGet(factionID, req.Payload.Hash)
		if ga != nil {
			reply(ga)
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineAbilitiesUpdated, req.Payload.Hash))
	return req.TransactionID, busKey, nil
}

func (arena *Arena) UserOnline(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	if arena.currentBattle == nil {
		return nil
	}

	var user *BattleUser
	err := json.Unmarshal(payload, user)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to unmarshal online user")
		return err
	}

	arena.currentBattle.userOnline(user, wsc)
	return nil
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
		reply(btl.abilities.BribeStageGet())
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
	btl := &Battle{
		arena:   arena,
		ID:      uuid.Must(uuid.NewV4()),
		MapName: gameMap.Name,
		gameMap: gameMap,
		stage:   BattleStagStart,
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
					Str("Battle ID", btl.ID.String()).
					Str("Faction ID", factionID.String()).
					Err(err).Msg("unable to retrieve faction from database")

			}
			factions[factionID] = faction
		}
	}

	btl.factions = factions

	btl.Battle, err = db.Battle(btl.ID, uuid.UUID(gameMap.ID), bmd)
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID.String()).Err(err).Msg("unable to insert battle into database")
		//TODO: something more dramatic
	}

	return btl
}
