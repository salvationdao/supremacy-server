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
	}

	// add the commands here

	// battle state
	ba.Command(BattleStartCommand, ba.BattleStartHandler)
	ba.Command(BattleEndCommand, ba.BattleEndHandler)

	// war machines
	ba.Command(WarMachineDestroyedCommand, ba.WarMachineDestroyedHandler)

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

func (ba *BattleArena) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"gameserver-v1"},
	})
	if err != nil {
		ba.Log.Printf("%v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	if c.Subprotocol() != "gameserver-v1" {
		ba.Log.Printf("client must speak the gameserver-v1 subprotocol")

		c.Close(websocket.StatusPolicyViolation, "client must speak the gameserver-v1 subprotocol")
		return
	}

	// send message
	go func() {
		err := ba.sendPump(c)
		if err != nil {
			ba.Log.Err(err).Msgf("error sending message to server")
			return
		}
	}()

	// listening for message
	for {
		ctx := context.Background()
		_, r, err := c.Reader(ctx)
		if err != nil {
			ba.Log.Err(err).Msgf(err.Error())
		}

		payload, err := ioutil.ReadAll(r)
		if err != nil {
			ba.Log.Err(err).Msgf(`error reading out buffer`)
			continue
		}

		v, err := jason.NewObjectFromBytes(payload)
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

		ba.runGameCommand(ctx, c, BattleCommand(cmdKey), payload)

		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			ba.Log.Err(err).Msg(err.Error())
			return
		}
	}
}

func (ba *BattleArena) sendPump(c *websocket.Conn) error {
	for {
		select {
		case msg := <-ba.send:
			err := writeTimeout(msg, time.Second*5, c)
			if err != nil {
				ba.close()
				return terror.Error(err)
			}
		case <-ba.ctx.Done():
			return terror.Error(ba.ctx.Err())
		}
	}
}

// writeTimeout enforces a timeout on websocket writes
func writeTimeout(msg *GameMessage, timeout time.Duration, c *websocket.Conn) error {
	ctx, cancel := context.WithTimeout(msg.context, timeout)
	defer func() {
		cancel()
		msg.cancel()
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
