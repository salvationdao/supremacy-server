package battle

import (
	"context"
	"encoding/binary"
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
	"strconv"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/hub/ext/messagebus"

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
	AbilityRightRatioTick
	VotePriceTick
	VotePriceForecastTick
	AbilityTargetPriceTick
	ViewerLiveCountTick
	SpoilOfWarTick
)

// BATTLESPAWNCOUNT defines how many mechs to spawn
// this should be refactored to a number in the database
// config table may be necessary, suggest key/value
const BATTLESPAWNCOUNT int = 3

func (mt MessageType) String() string {
	return [...]string{"JSON", "Tick", "Live Vote Tick", "Ability Right Ratio Tick",
		"Vote Price Tick", "Vote Price Forecast Tick", "Ability Target Price Tick", "Viewer Live Count Tick", "Spoils of War Tick"}[mt]
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
	opts.SecureUserFactionCommand(HubKeFactionUniqueAbilityContribute, arena.FactionUniqueAbilityContribute)
	opts.Command(HubKeyGameSettingsUpdated, arena.SendSettings)

	// subscribe functions
	opts.SubscribeCommand(HubKeyGameNotification, arena.GameNotificationSubscribeHandler)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribeStageUpdateSubscribe, arena.GabsBribeStageSubscribe)
	opts.SecureUserFactionSubscribeCommand(HubKeGabsBribingWinnerSubscribe, arena.GabsBribingWinnerSubscribe)

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

func (arena *Arena) FactionUniqueAbilityContribute(ctx context.Context, wsc *hub.Client, payload []byte, factionID server.FactionID, reply hub.ReplyFunc) error {
	if arena.currentBattle == nil {
		return nil
	}
	btl := arena.currentBattle
	err := btl.abilities.AbilityContribute(wsc, payload, factionID)
	if err != nil {
		return terror.Error(err)
	}
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

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
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

const HubKeGabsBribingWinnerSubscribe hub.HubCommandKey = "BRIBE:WINNER:SUBSCRIBE"

// GabsBribingWinnerSubscribe subscribe on winner notification
func (arena *Arena) GabsBribingWinnerSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID))

	return req.TransactionID, busKey, nil
}

func (arena *Arena) SendSettings(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	if arena.currentBattle == nil {
		return nil
	}
	btl := arena.currentBattle
	reply(btl.updatePayload())
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
		Stage: &BattleState{
			Stage: BattleStagStart,
		},
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

	_, err = db.Battle(btl.ID, uuid.UUID(gameMap.ID), bmd)
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID.String()).Err(err).Msg("unable to insert battle into database")
		//TODO: something more dramatic
	}

	return btl
}

const HubKeyWarMachineLocationUpdated hub.HubCommandKey = "WAR:MACHINE:LOCATION:UPDATED"

func (btl *Battle) start(payload *BattleStartPayload) {
	for _, wm := range payload.WarMachines {
		for i, wm2 := range btl.WarMachines {
			if wm.Hash == wm2.Hash {
				btl.WarMachines[i].ParticipantID = wm.ParticipantID
				continue
			}
		}
	}

	// set up the abilities for current battle
	btl.abilities = NewAbilitiesSystem(btl)

	btl.BroadcastUpdate()
}

func (btl *Battle) end(payload *BattleEndPayload) {
	ids := make([]uuid.UUID, len(btl.WarMachines))
	err := db.ClearQueue(ids...)
	if err != nil {
		gamelog.L.Error().Interface("ids", ids).Err(err).Msg("db.ClearQueue() returned error")
		return
	}

	btl.Stage.Lock()
	btl.Stage.Stage = BattleStageEnd
	btl.Stage.Unlock()

	mws := make([]*db.MechWithOwner, len(payload.WinningWarMachines))

	for i, wmwin := range payload.WinningWarMachines {
		var wm *WarMachine
		for _, w := range btl.WarMachines {
			if w.Hash == wmwin.Hash {
				wm = w
				break
			}
		}
		if wm == nil {
			gamelog.L.Error().Str("Battle ID", btl.ID.String()).Msg("unable to match war machine to battle with hash")
			return
		}
		mechId, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID.String()).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert mech id to uuid")
			return
		}
		ownedById, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID.String()).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert owned id to uuid")
			return
		}
		factionId, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID.String()).
				Str("faction ID", wm.FactionID).
				Err(err).
				Msg("unable to convert faction id to uuid")
			return
		}
		mws[i] = &db.MechWithOwner{
			OwnerID:   ownedById,
			MechID:    mechId,
			FactionID: factionId,
		}
	}
	err = db.WinBattle(btl.ID, payload.WinCondition, mws...)
	if err != nil {
		gamelog.L.Error().
			Str("Battle ID", btl.ID.String()).
			Err(err).
			Msg("unable to store mech wins")
		return
	}
}

type BroadcastPayload struct {
	Key     hub.HubCommandKey `json:"key"`
	Payload interface{}       `json:"payload"`
}

type GameSettingsResponse struct {
	GameMap            *server.GameMap `json:"game_map"`
	WarMachines        []*WarMachine   `json:"war_machines"`
	SpawnedAI          []*WarMachine   `json:"spawned_ai"`
	WarMachineLocation []byte          `json:"war_machine_location"`
}

func (btl *Battle) updatePayload() *GameSettingsResponse {
	var lt []byte
	if btl.lastTick != nil {
		lt = *btl.lastTick
	}
	return &GameSettingsResponse{
		GameMap:            btl.gameMap,
		WarMachines:        btl.WarMachines,
		SpawnedAI:          btl.SpawnedAI,
		WarMachineLocation: lt,
	}
}

const HubKeyGameSettingsUpdated = hub.HubCommandKey("GAME:SETTINGS:UPDATED")

func (btl *Battle) BroadcastUpdate() {
	btl.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameSettingsUpdated), btl.updatePayload())
}

func (btl *Battle) Tick(payload []byte) {
	// Save to history
	// btl.BattleHistory = append(btl.BattleHistory, payload)

	broadcast := false
	// broadcast
	if btl.lastTick == nil {
		broadcast = true
	}
	btl.lastTick = &payload

	btl.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeyWarMachineLocationUpdated), payload)

	// Update game settings (so new players get the latest position, health and shield of all warmachines)
	count := payload[1]
	var c byte
	offset := 2
	for c = 0; c < count; c++ {
		participantID := payload[offset]
		offset++

		// Get Warmachine Index
		warMachineIndex := -1
		for i, wmn := range btl.WarMachines {
			if wmn.ParticipantID == participantID {
				warMachineIndex = i
				break
			}
		}

		// Get Sync byte (tells us which data was updated for this warmachine)
		syncByte := payload[offset]
		offset++

		// Position + Yaw
		if syncByte >= 100 {
			x := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			y := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			rotation := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4

			if warMachineIndex != -1 {
				if btl.WarMachines[warMachineIndex].Position == nil {
					btl.WarMachines[warMachineIndex].Position = &server.Vector3{}
				}
				btl.WarMachines[warMachineIndex].Position.X = x
				btl.WarMachines[warMachineIndex].Position.X = y
				btl.WarMachines[warMachineIndex].Rotation = rotation
			}
		}
		// Health
		if syncByte == 1 || syncByte == 11 || syncByte == 101 || syncByte == 111 {
			health := binary.BigEndian.Uint32(payload[offset : offset+4])
			offset += 4
			if warMachineIndex != -1 {
				btl.WarMachines[warMachineIndex].Health = health
			}
		}
		// Shield
		if syncByte == 10 || syncByte == 11 || syncByte == 110 || syncByte == 111 {
			shield := binary.BigEndian.Uint32(payload[offset : offset+4])
			offset += 4
			if warMachineIndex != -1 {
				btl.WarMachines[warMachineIndex].Shield = shield
			}
		}
	}
	if broadcast {
		btl.BroadcastUpdate()
	}
}

func (arena *Arena) reset() {
	gamelog.L.Warn().Msg("arena state resetting")
}

type JoinPaylod struct {
	AssetHash   string `json:"asset_hash"`
	NeedInsured bool   `json:"need_insured"`
}

func (arena *Arena) Join(ctx context.Context, wsc *hub.Client, payload []byte, factionID server.FactionID, reply hub.ReplyFunc) error {
	span := tracer.StartSpan("ws.Command", tracer.ResourceName(string(WSJoinQueue)))
	defer span.Finish()

	msg := &JoinPaylod{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	mechId, err := db.MechIDFromHash(msg.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	mech, err := db.Mech(mechId)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechId.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechId.String()).Err(err).Msg("mech's owner player has no faction")
		return err
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return err
	}

	pos, err := db.JoinQueue(&db.BattleMechData{
		MechID:    mechId,
		OwnerID:   ownerID,
		FactionID: uuid.UUID(factionID),
	})

	if err != nil {
		gamelog.L.Error().Interface("factionID", mech.FactionID).Err(err).Msg("unable to insert mech into queue")
		return err
	}

	reply(pos)

	return err
}

func (btl *Battle) Destroyed(dp *BattleWMDestroyedPayload) {
	// check destroyed war machine exist
	if btl.ID.String() != dp.BattleID {
		gamelog.L.Warn().Str("battle.ID", btl.ID.String()).Str("gameclient.ID", dp.BattleID).Msg("battle state does not match game client state")
		btl.arena.reset()
		return
	}

	var destroyedWarMachine *WarMachine
	dHash := dp.DestroyedWarMachineEvent.DestroyedWarMachineHash
	for i, wm := range btl.WarMachines {
		if wm.Hash == dHash {
			// set health to 0
			btl.WarMachines[i].Health = 0
			destroyedWarMachine = wm
			break
		}
	}
	if destroyedWarMachine == nil {
		gamelog.L.Warn().Str("hash", dHash).Msg("can't match destroyed mech with battle state")
		return
	}

	var killByWarMachine *WarMachine
	if dp.DestroyedWarMachineEvent.KillByWarMachineHash != "" {
		for _, wm := range btl.WarMachines {
			if wm.Hash == dp.DestroyedWarMachineEvent.KillByWarMachineHash {
				killByWarMachine = wm
			}
		}
		if destroyedWarMachine == nil {
			gamelog.L.Warn().Str("killed_by_hash", dp.DestroyedWarMachineEvent.KillByWarMachineHash).Msg("can't match killer mech with battle state")
			return
		}
	}

	gamelog.L.Info().Msgf("battle Update: %s - War Machine Destroyed: %s", btl.ID, dHash)

	// save to database
	//tx, err := ba.Conn.Begin(ctx)
	//if err != nil {
	//	return terror.Error(err)
	//}
	//
	//defer func(tx pgx.Tx, ctx context.Context) {
	//	err := tx.Rollback(ctx)
	//	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
	//		ba.Log.Err(err).Msg("error rolling back")
	//	}
	//}(tx, ctx)

	var warMachineID uuid.UUID
	var killByWarMachineID uuid.UUID
	ids, err := db.MechIDsFromHash(destroyedWarMachine.Hash, dp.DestroyedWarMachineEvent.KillByWarMachineHash)

	if err != nil || len(ids) == 0 {
		gamelog.L.Warn().
			Str("hashes", fmt.Sprintf("%s, %s", destroyedWarMachine.Hash, dp.DestroyedWarMachineEvent.KillByWarMachineHash)).
			Str("battle_id", btl.ID.String()).
			Err(err).
			Msg("can't retrieve mech ids")

	} else {
		warMachineID = ids[0]
		if len(ids) > 1 {
			killByWarMachineID = ids[1]
		}

		//TODO: implement related id
		if dp.DestroyedWarMachineEvent.RelatedEventIDString != "" {
			relatedEventuuid, err := uuid.FromString(dp.DestroyedWarMachineEvent.RelatedEventIDString)
			if err != nil {
				gamelog.L.Warn().
					Str("relatedEventuuid", dp.DestroyedWarMachineEvent.RelatedEventIDString).
					Str("battle_id", btl.ID.String()).
					Msg("can't create uuid from non-empty related event idf")
			}
			dp.DestroyedWarMachineEvent.RelatedEventID = relatedEventuuid
		}

		evt := &db.BattleEvent{
			BattleID:  btl.ID,
			WM1:       warMachineID,
			WM2:       killByWarMachineID,
			EventType: db.Btlevnt_Killed,
			CreatedAt: time.Now(),
			RelatedID: dp.DestroyedWarMachineEvent.RelatedEventIDString,
		}

		_, err = db.StoreBattleEvent(btl.ID, dp.DestroyedWarMachineEvent.RelatedEventID, warMachineID, killByWarMachineID, db.Btlevnt_Killed, time.Now())
		if err != nil {
			gamelog.L.Warn().
				Interface("event_data", evt).
				Str("battle_id", btl.ID.String()).
				Msg("unable to store mech event data")
		}
	}

	//err = db.WarMachineDestroyedEventCreate(ctx, tx, dp.BattleID, dp.DestroyedWarMachineEvent)
	//if err != nil {
	//	return terror.Error(err)
	//}

	//err = db.StoreBattleEvent(&db.BattleEvent{})

	// TODO: Add kill assists
	//if len(assistedWarMachineIDs) > 0 {
	//	err = db.WarMachineDestroyedEventAssistedWarMachineSet(ctx, tx, dp.DestroyedWarMachineEvent.ID, assistedWarMachineIDs)
	//	if err != nil {
	//		return terror.Error(err)
	//	}
	//}

	//err = tx.Commit(ctx)
	//if err != nil {
	//	return terror.Error(err)
	//}

	_, err = db.UpdateBattleMech(btl.ID, warMachineID, false, true, killByWarMachineID)
	if err != nil {
		gamelog.L.Error().
			Str("battle_id", btl.ID.String()).
			Interface("mech_id", warMachineID).
			Bool("killed", true).
			Msg("can't update battle mech")
	}

	// prepare destroyed record
	destroyedRecord := &WMDestroyedRecord{
		DestroyedWarMachine: destroyedWarMachine,
		KilledByWarMachine:  killByWarMachine,
		KilledBy:            dp.DestroyedWarMachineEvent.KilledBy,
		DamageRecords:       []*DamageRecord{},
	}

	// calc total damage and merge the duplicated damage source
	totalDamage := 0
	newDamageHistory := []*DamageHistory{}
	for _, damage := range dp.DestroyedWarMachineEvent.DamageHistory {
		totalDamage += damage.Amount
		// check instigator token id exist in the list
		if damage.InstigatorHash != "" {
			exists := false
			for _, hist := range newDamageHistory {
				if hist.InstigatorHash == damage.InstigatorHash {
					hist.Amount += damage.Amount
					exists = true
					break
				}
			}
			if !exists {
				newDamageHistory = append(newDamageHistory, &DamageHistory{
					Amount:         damage.Amount,
					InstigatorHash: damage.InstigatorHash,
					SourceName:     damage.SourceName,
					SourceHash:     damage.SourceHash,
				})
			}
			continue
		}
		// check source name
		exists := false
		for _, hist := range newDamageHistory {
			if hist.SourceName == damage.SourceName {
				hist.Amount += damage.Amount
				exists = true
				break
			}
		}
		if !exists {
			newDamageHistory = append(newDamageHistory, &DamageHistory{
				Amount:         damage.Amount,
				InstigatorHash: damage.InstigatorHash,
				SourceName:     damage.SourceName,
				SourceHash:     damage.SourceHash,
			})
		}
	}

	// get total damage amount for calculating percentage
	for _, damage := range newDamageHistory {
		damageRecord := &DamageRecord{
			SourceName: damage.SourceName,
			Amount:     (damage.Amount * 1000000 / totalDamage) / 100,
		}
		if damage.InstigatorHash != "" {
			for _, wm := range btl.WarMachines {
				if wm.Hash == damage.InstigatorHash {
					damageRecord.CausedByWarMachineHash = wm.Hash
				}
			}
		}
		destroyedRecord.DamageRecords = append(destroyedRecord.DamageRecords, damageRecord)
	}

	//// cache record in battle, for future subscription
	//btl.WarMachineDestroyedRecordMap[destroyedWarMachine.ParticipantID] = destroyedRecord

	// send event to hub clients
	//ba.Events.Trigger(ctx, EventWarMachineDestroyed, &EventData{
	//	WarMachineDestroyedRecord: destroyedRecord,
	//})

	wmd := struct {
		DestroyedWarMachine *WarMachineBrief `json:"destroyedWarMachine"`
		KilledByWarMachine  *WarMachineBrief `json:"killedByWarMachineID,omitempty"`
		KilledBy            string           `json:"killedBy"`
	}{
		DestroyedWarMachine: &WarMachineBrief{
			ImageUrl:    destroyedWarMachine.Image,
			ImageAvatar: destroyedWarMachine.Image, // TODO: should be imageavatar
			Name:        destroyedWarMachine.Name,
			Hash:        destroyedWarMachine.Hash,
			Faction: &FactionBrief{
				ID:    destroyedWarMachine.FactionID,
				Label: destroyedWarMachine.Faction.Label,
				Theme: destroyedWarMachine.Faction.Theme,
			},
		},
	}

	if killByWarMachine != nil {
		wmd.KilledByWarMachine = &WarMachineBrief{
			ImageUrl:    killByWarMachine.Image,
			ImageAvatar: killByWarMachine.Image, // TODO: should be imageavatar
			Name:        killByWarMachine.Name,
			Hash:        killByWarMachine.Hash,
			Faction: &FactionBrief{
				ID:    killByWarMachine.FactionID,
				Label: killByWarMachine.Faction.Label,
				Theme: killByWarMachine.Faction.Theme,
			},
		}
	}

	btl.arena.messageBus.Send(context.Background(),
		messagebus.BusKey(
			fmt.Sprintf(
				"%s:%x",
				hub.HubCommandKey("WAR:MACHINE:DESTROYED:UPDATED"),
				destroyedWarMachine.ParticipantID,
			),
		),
		wmd,
	)
}

func (btl *Battle) Load() error {
	q := []*boiler.BattleQueue{}
	err := db.LoadBattleQueue(context.Background(), &q)
	if err != nil {
		gamelog.L.Warn().Str("battle_id", btl.ID.String()).Err(err).Msg("unable to load out queue")
		return err
	}

	if len(q) < 9 {
		gamelog.L.Warn().Msg("not enough mechs to field a battle. replacing with default battle.")

		err = btl.DefaultMechs()
		if err != nil {
			gamelog.L.Warn().Str("battle_id", btl.ID.String()).Err(err).Msg("unable to load default mechs")
			return err
		}
		return nil
	}

	ids := make([]uuid.UUID, len(q))
	for i, bq := range q {
		ids[i], err = uuid.FromString(bq.MechID)
		if err != nil {
			gamelog.L.Warn().Str("mech_id", bq.MechID).Msg("failed to convert mech id string to uuid")
			return err
		}
	}

	mechs, err := db.Mechs(ids...)
	if err != nil {
		gamelog.L.Warn().Interface("mechs_ids", ids).Str("battle_id", btl.ID.String()).Err(err).Msg("failed to retrieve mechs from mech ids")
		return err
	}
	btl.WarMachines = btl.MechsToWarMachines(mechs)

	return nil
}

func (btl *Battle) MechsToWarMachines(mechs []*server.MechContainer) []*WarMachine {
	warmachines := make([]*WarMachine, len(mechs))
	for i, mech := range mechs {
		label := mech.Faction.Label
		if label == "" {
			gamelog.L.Warn().Interface("faction_id", mech.Faction.ID).Str("battle_id", btl.ID.String()).Msg("mech faction is an empty label")
		}
		if len(label) > 10 {
			words := strings.Split(label, " ")
			label = ""
			for _, word := range words {
				label = label + string([]rune(word)[0])
			}
		}

		weaponNames := make([]string, len(mech.Weapons))
		for k, wpn := range mech.Weapons {
			i, err := strconv.Atoi(k)
			if err != nil {
				gamelog.L.Warn().Str("key", k).Interface("weapon", wpn).Str("battle_id", btl.ID.String()).Msg("mech weapon's key is not an int")
			}
			weaponNames[i] = wpn.Label
		}

		warmachines[i] = &WarMachine{
			ID:            mech.ID,
			Name:          mech.Name,
			Hash:          mech.Hash,
			ParticipantID: 0,
			FactionID:     btl.arena.AIPlayers[mech.OwnerID].FactionID.String,
			MaxHealth:     uint32(mech.Chassis.MaxHitpoints),
			Health:        uint32(mech.Chassis.MaxHitpoints),
			MaxShield:     uint32(mech.Chassis.MaxShield),
			Shield:        uint32(mech.Chassis.MaxShield),
			Stat:          nil,
			OwnedByID:     mech.OwnerID,
			ImageAvatar:   mech.AvatarURL,
			Faction: &Faction{
				ID:    mech.Faction.ID,
				Label: label,
				Theme: &FactionTheme{
					Primary:    mech.Faction.PrimaryColor,
					Secondary:  mech.Faction.SecondaryColor,
					Background: mech.Faction.BackgroundColor,
				},
			},
			Speed:              mech.Chassis.Speed,
			Skin:               mech.Chassis.Skin,
			ShieldRechargeRate: float64(mech.Chassis.ShieldRechargeRate),
			Durability:         mech.Chassis.MaxHitpoints,
			WeaponHardpoint:    mech.Chassis.WeaponHardpoints,
			TurretHardpoint:    mech.Chassis.TurretHardpoints,
			UtilitySlots:       mech.Chassis.UtilitySlots,
			Description:        nil,
			ExternalUrl:        "",
			Image:              mech.ImageURL,
			PowerGrid:          1,
			CPU:                1,
			WeaponNames:        weaponNames,
			Tier:               mech.Tier,
		}
	}
	return warmachines
}
