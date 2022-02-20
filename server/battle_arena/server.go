package battle_arena

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"server"
	"server/passport"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"github.com/ninja-software/terror/v2"

	"github.com/antonholmquist/jason"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type BattleCommand string

type ReplyFunc func(interface{})

type BattleCommandFunc func(ctx context.Context, payload []byte, reply ReplyFunc) error

type Request struct {
	BattleCommand BattleCommand `json:"battleCommand"`
	Payload       []byte        `json:"payload"`
}

type GameMessage struct {
	BattleCommand BattleCommand `json:"battleCommand"`
	Payload       interface{}   `json:"payload"`
	context       context.Context
	cancel        context.CancelFunc
}

type BattleArena struct {
	server   *http.Server
	Log      *zerolog.Logger
	Conn     *pgxpool.Pool
	passport *passport.Passport
	addr     string
	commands map[BattleCommand]BattleCommandFunc
	Events   BattleArenaEvents
	send     chan *GameMessage
	ctx      context.Context
	close    context.CancelFunc
	battle   *server.Battle

	// battle queue channels
	BattleQueueMap map[server.FactionID]chan func(*WarMachineQueuingList)
}

// NewBattleArenaClient creates a new battle arena client
func NewBattleArenaClient(ctx context.Context, logger *zerolog.Logger, conn *pgxpool.Pool, passport *passport.Passport, addr string) *BattleArena {
	ctx, cancel := context.WithCancel(ctx)

	ba := &BattleArena{
		Log:      logger,
		addr:     addr,
		Conn:     conn,
		commands: make(map[BattleCommand]BattleCommandFunc),
		send:     make(chan *GameMessage),
		passport: passport,
		Events:   BattleArenaEvents{map[Event][]EventHandler{}, sync.RWMutex{}},
		ctx:      ctx,
		close:    cancel,
		battle:   &server.Battle{},

		// channel for battle queue
		BattleQueueMap: make(map[server.FactionID]chan func(*WarMachineQueuingList)),
	}
	// add the commands here

	// battle state
	ba.Command(BattleReadyCommand, ba.BattleReadyHandler)
	ba.Command(BattleStartCommand, ba.BattleStartHandler)
	ba.Command(BattleEndCommand, ba.BattleEndHandler)

	// war machines
	ba.Command(WarMachineDestroyedCommand, ba.WarMachineDestroyedHandler)

	go ba.SetupAfterConnections()
	return ba
}

// Serve starts the battle arena server
func (ba *BattleArena) Serve(ctx context.Context) error {
	l, err := net.Listen("tcp", ba.addr)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Starting BattleArena Server on %v", l.Addr())
	ba.server = &http.Server{
		Handler:      ba,
		ReadTimeout:  writeWait,
		WriteTimeout: writeWait,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- ba.server.Serve(l)
	}()

	select {
	case err := <-errChan:
		ba.Log.Err(err).Msgf("Shutting down battle arena server.")
	case <-ctx.Done():
		ba.Log.Info().Msgf("Shutting down battle arena server.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return ba.server.Shutdown(ctx)
}

type NetMessageType byte

// NetMessageTypes
const (
	NetMessageTypeJSON NetMessageType = iota
	NetMessageTypeTick
	NetMessageTypeLiveVotingTick
	NetMessageTypeAbilityRightRatioTick
	NetMessageTypeVotePriceTick
	NetMessageTypeVotePriceForecastTick
	NetMessageTypeAbilityTargetPriceTick
	NetMessageTypeViewerLiveCountTick
	NetMessageTypeSpoilOfWarTick
)

func (ba *BattleArena) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		//Subprotocols: []string{"gameserver-v1"},
	})
	if err != nil {
		ba.Log.Err(err).Msg("")
		cancel()
		return
	}

	ba.Log.Info().Msg("game client connected")

	defer c.Close(websocket.StatusInternalError, "game client disconnected")

	// if c.Subprotocol() != "gameserver-v1" {
	// 	ba.Log.Printf("client must speak the gameserver-v1 subprotocol")

	// 	c.Close(websocket.StatusPolicyViolation, "client must speak the gameserver-v1 subprotocol")
	// 	return
	// }

	// send message
	go ba.sendPump(ctx, cancel, c)

	go func() {
		time.Sleep(2 * time.Second)
		_ = ba.InitNextBattle()
	}()

	// listening for message
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, r, err := c.Reader(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(err) == websocket.StatusGoingAway {
					return
				}
				ba.Log.Err(err).Msgf("battle arena connection reader error")
				cancel()
				return
			}

			payload, err := ioutil.ReadAll(r)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(err) == websocket.StatusGoingAway {
					return
				}
				ba.Log.Err(err).Msgf(`error reading out buffer`)
				cancel()
				return
			}

			msgType := NetMessageType(payload[0])

			switch msgType {
			case NetMessageTypeJSON:
				v, err := jason.NewObjectFromBytes(payload[1:])
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
						websocket.CloseStatus(err) == websocket.StatusGoingAway {
						return
					}
					ba.Log.Err(err).Msgf(`error making object from bytes`)
					cancel()
					return
				}
				cmdKey, err := v.GetString("battleCommand")
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
						websocket.CloseStatus(err) == websocket.StatusGoingAway {
						return
					}
					ba.Log.Err(err).Msgf(`missing json key "key"`)
					continue
				}
				if cmdKey == "" {
					if errors.Is(err, context.Canceled) {
						return
					}
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
						websocket.CloseStatus(err) == websocket.StatusGoingAway {
						return
					}
					ba.Log.Err(fmt.Errorf("missing key value")).Msgf("missing key/command value")
					continue
				}
				ba.runGameCommand(ctx, c, BattleCommand(cmdKey), payload[1:])
			case NetMessageTypeTick:
				ba.WarMachinesTick(ctx, payload)
			default:
				ba.Log.Err(fmt.Errorf("unknown message type")).Msg("")
			}
		}
	}

}

func (ba *BattleArena) sendPump(ctx context.Context, cancel context.CancelFunc, c *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ba.send:
			err := writeTimeout(msg, writeWait, c)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(err) == websocket.StatusGoingAway {
					return
				}
				ba.Log.Err(err).Msg("error sending message to game client")
			}
		case <-ticker.C:
			err := pingTimeout(ctx, writeWait, c)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(err) == websocket.StatusGoingAway {
					return
				}
				ba.Log.Err(err).Msgf("error with pinging gameclient")
				cancel()
			}
		}
	}
}

// PING COMMENTED OUT BY VINNIE, not getting ping responses and breaks connection

//pingTimeout enforces a timeout on websocket writes
func pingTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Ping(ctx)
}

// writeTimeout enforces a timeout on websocket writes
func writeTimeout(msg *GameMessage, timeout time.Duration, c *websocket.Conn) error {
	ctx, cancel := context.WithTimeout(msg.context, timeout)
	defer cancel()

	jsn, err := json.Marshal(msg)
	if err != nil {
		return terror.Error(err)
	}

	return c.Write(ctx, websocket.MessageText, jsn)
}

// runGameCommand tried to run a game command
func (ba *BattleArena) runGameCommand(ctx context.Context, c *websocket.Conn, cmd BattleCommand, payload []byte) {
	fn, ok := ba.commands[cmd]
	if !ok {
		err := fmt.Errorf("no command found for %s", cmd)
		ba.Log.Err(err).Msg("Missing Command")
		err = wsjson.Write(ctx, c, err.Error())
		if err != nil {
			panic(err)
		}

		return
	}

	reply := func(payload interface{}) {
		err := wsjson.Write(ctx, c, payload)
		if err != nil {
			ba.Log.Error().Err(err).Msg("send error")
		}
	}

	err := func() error {
		// //fnSpan := span.StartChild("cmd")
		// defer func() {
		// 	//fnSpan.Finish()
		// 	//ba.Log.Trace().Msgf("%s:%s all took %s", request.BattleCommand, "cmd", time.Since(fnSpan.StartTime))
		// }()
		return fn(ctx, payload, reply)
	}()

	if err != nil {
		// Log the error
		log_helpers.TerrorEcho(ctx, err, ba.Log)

		resp := struct {
			Command BattleCommand `json:"battleCommand"`
			//TransactionID string        `json:"transactionID"`
			Success bool        `json:"success"`
			Payload interface{} `json:"payload"`
		}{
			Command: cmd,
			//TransactionID: request.TransactionID,
			Success: false,
			Payload: err.Error(),
		}

		var bErr *terror.TError
		if errors.As(err, &bErr) {
			resp.Payload = bErr.Message
		}

		err := wsjson.Write(ctx, c, resp)
		if err != nil {
			panic(err)
		}
	}
}

// Command registers a command to the game server
func (ba *BattleArena) Command(command BattleCommand, fn BattleCommandFunc) {
	if _, ok := ba.commands[command]; ok {
		ba.Log.Panic().Msgf("command has already been registered to hub: %s", command)
	}
	ba.commands[command] = fn
	ba.Log.Trace().Msg(string(command))
}

func (ba *BattleArena) SetupAfterConnections() {
	var factions []*server.Faction
	var err error
	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2,
	}

	// get factions from passport, retrying every 10 seconds until we ge them.
	for len(factions) <= 0 {
		if !ba.passport.Connected {
			time.Sleep(b.Duration())
			continue
		}

		factions, err = ba.passport.FactionAll(ba.ctx)
		if err != nil {
			ba.Log.Err(err).Msg("unable to get factions")
		}

		if len(factions) > 0 {
			break
		}
		time.Sleep(b.Duration())
	}

	ba.battle.WarMachineDestroyedRecordMap = make(map[byte]*server.WarMachineDestroyedRecord)

	ba.battle.FactionMap = make(map[server.FactionID]*server.Faction)
	for _, faction := range factions {
		ba.battle.FactionMap[faction.ID] = faction
	}

	// get all the faction list from passport server
	for _, faction := range factions {
		// start battle queue
		ba.BattleQueueMap[faction.ID] = make(chan func(*WarMachineQueuingList))
		go ba.startBattleQueue(faction.ID)
	}

	battleContractRewardUpdaterLogger := log_helpers.NamedLogger(ba.Log, "Contract Reward Updater").Level(zerolog.Disabled)
	battleContractRewardUpdater := tickle.New("Contract Reward Updater", 10, func() (int, error) {
		rQueueNumberChan := make(chan int)
		ba.BattleQueueMap[server.RedMountainFactionID] <- func(wmql *WarMachineQueuingList) {
			rQueueNumberChan <- len(wmql.WarMachines)
		}
		bQueueNumberChan := make(chan int)
		ba.BattleQueueMap[server.BostonCyberneticsFactionID] <- func(wmql *WarMachineQueuingList) {
			bQueueNumberChan <- len(wmql.WarMachines)
		}
		zQueueNumberChan := make(chan int)
		ba.BattleQueueMap[server.ZaibatsuFactionID] <- func(wmql *WarMachineQueuingList) {
			zQueueNumberChan <- len(wmql.WarMachines)
		}

		ba.passport.FactionWarMachineContractRewardUpdate(
			ba.ctx,
			[]*server.FactionWarMachineQueue{
				{
					FactionID:  server.RedMountainFactionID,
					QueueTotal: <-rQueueNumberChan,
				},
				{
					FactionID:  server.BostonCyberneticsFactionID,
					QueueTotal: <-bQueueNumberChan,
				},
				{
					FactionID:  server.ZaibatsuFactionID,
					QueueTotal: <-zQueueNumberChan,
				},
			},
		)

		return http.StatusOK, nil
	})
	battleContractRewardUpdater.Log = &battleContractRewardUpdaterLogger
	battleContractRewardUpdater.Start()
}
