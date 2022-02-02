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

	"github.com/antonholmquist/jason"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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
	// TODO: handle ctx with listen? ListenConfig.Listen.(ctx, network, addr)
	l, err := net.Listen("tcp", ba.addr)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Starting BattleArena Server on %v", l.Addr())
	ba.server = &http.Server{
		Handler:      ba,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	go func() {
		<-ctx.Done()
		ba.Close()
	}()

	return ba.server.Serve(l)
}

func (ba *BattleArena) Close() {
	if ba.server != nil {
		ba.Log.Info().Msg("closing battle-arena server")
		err := ba.server.Close()
		if err != nil {
			ba.Log.Warn().Err(err).Msgf("")
		}
	}
}

type NetMessageType byte

// NetMessageTypes
const (
	NetMessageTypeJSON NetMessageType = iota
	NetMessageTypeTick
	NetMessageTypeLiveVotingTick
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

	ba.Log.Info().Msg("game client conneted")

	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// if c.Subprotocol() != "gameserver-v1" {
	// 	ba.Log.Printf("client must speak the gameserver-v1 subprotocol")

	// 	c.Close(websocket.StatusPolicyViolation, "client must speak the gameserver-v1 subprotocol")
	// 	return
	// }

	// send message
	go ba.sendPump(ctx, c)

	// Init first battle
	//ba.Events.Trigger(r.Context(), EventGameInit, &EventData{BattleArena: ba.battle})

	go func() {
		time.Sleep(2 * time.Second)
		_ = ba.InitNextBattle()
	}()
	// listening for message
	for {
		select {
		case <-ctx.Done():
			err := c.Close(websocket.StatusGoingAway, "context done")
			if err != nil {
				ba.Log.Err(err).Msg("")
			}
			cancel()
			return
		default:
			_, r, err := c.Reader(ctx)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					ba.Log.Warn().Msg("game client connection lost")
					cancel()
					return
				}
				ba.Log.Err(err).Msgf(err.Error())
			}

			payload, err := ioutil.ReadAll(r)
			if err != nil {
				ba.Log.Err(err).Msgf(`error reading out buffer`)
				continue
			}

			msgType := NetMessageType(payload[0])

			switch msgType {
			case NetMessageTypeJSON:
				v, err := jason.NewObjectFromBytes(payload[1:])
				if err != nil {
					ba.Log.Err(err).Msgf(`error making object from bytes`)
					continue
				}
				cmdKey, err := v.GetString("battleCommand")
				if err != nil {
					ba.Log.Err(err).Msgf(`missing json key "key"`)
					continue
				}
				if cmdKey == "" {
					ba.Log.Err(fmt.Errorf("missing key value")).Msgf("missing key/command value")
					continue
				}
				ba.runGameCommand(ctx, c, BattleCommand(cmdKey), payload[1:])
			case NetMessageTypeTick:
				ba.WarMachinesTick(payload)
			default:
				ba.Log.Err(fmt.Errorf("unknown message type")).Msg("")
			}
		}
	}

}

func (ba *BattleArena) sendPump(ctx context.Context, c *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ba.send:
			err := writeTimeout(msg, time.Second*5, c)
			if err != nil {
				ba.Log.Err(err).Msg("error sending message to game client")
			}
		}
	}
}

// writeTimeout enforces a timeout on websocket writes
func writeTimeout(msg *GameMessage, timeout time.Duration, c *websocket.Conn) error {
	ctx, cancel := context.WithTimeout(msg.context, timeout)
	defer func() {
		cancel()
	}()

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

	// TODO: FIX THIS, it seems to be super delayed in getting the factions

	// get factions from passport, retrying every 10 seconds until we ge them.
	for len(factions) <= 0 {
		// since the passport spins up concurrently the passport connection may not be setup right away, so we check every second for the connection
		for ba.passport == nil || ba.passport.Conn == nil || !ba.passport.Conn.Connected {
			time.Sleep(2 * time.Second)
		}

		factions, err = ba.passport.FactionAll(ba.ctx, "faction all - gameserver")
		if err != nil {
			ba.Log.Err(err).Msg("unable to get factions")
		}
		time.Sleep(2 * time.Second)
	}

	ba.battle.FactionMap = make(map[server.FactionID]*server.Faction)
	for _, faction := range factions {
		ba.battle.FactionMap[faction.ID] = faction
	}

	// get all the faction list from passport server
	for _, faction := range ba.battle.FactionMap {
		// start battle queue
		ba.BattleQueueMap[faction.ID] = make(chan func(*WarMachineQueuingList))
		go ba.startBattleQueue(faction.ID)
	}
}
