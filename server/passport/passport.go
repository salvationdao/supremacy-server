package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/antonholmquist/jason"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"io/ioutil"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

type Command string

type ReplyFunc func(interface{})

type CommandFunc func(ctx context.Context, payload []byte, reply ReplyFunc) error

type Request struct {
	Key     Command `json:"key"`
	Payload []byte  `json:"payload"`
	context context.Context
	cancel  context.CancelFunc
}

type Message struct {
	Key     Command     `json:"key"`
	Payload interface{} `json:"payload"`
	context context.Context
	cancel  context.CancelFunc
}

type Passport struct {
	Log      *zerolog.Logger
	addr     string
	commands map[Command]CommandFunc
	Events   Events
	send     chan *Message

	clientID     string
	clientSecret string

	ctx   context.Context
	close context.CancelFunc
}

func NewPassport(ctx context.Context, logger *zerolog.Logger, addr, clientID, clientSecret string) (*Passport, error) {
	ctx, cancel := context.WithCancel(ctx)
	newPP := &Passport{
		Log:      logger,
		addr:     addr,
		commands: make(map[Command]CommandFunc),
		send:     make(chan *Message),
		Events:   Events{map[Event][]EventHandler{}, sync.RWMutex{}},

		clientID:     clientID,
		clientSecret: clientSecret,

		ctx:   ctx,
		close: cancel,
	}

	//err := newPP.initWebsocket(clientID, clientSecret)
	//if err != nil {
	//	return nil, terror.Error(err)
	//}

	return newPP, nil
}

func (pp *Passport) Connect(ctx context.Context) error {
	pp.Log.Info().Msgf("Connecting to passport on %v", pp.addr)

	ws, _, err := websocket.Dial(ctx, pp.addr, &websocket.DialOptions{
		//HTTPClient:nil,
		//HTTPHeader:nil,
		//Subprotocols:nil,
		//CompressionMode:0,
		//CompressionThreshold:0,
	})
	if err != nil {
		return terror.Error(err)
	}
	defer ws.Close(websocket.StatusInternalError, "websocket closed")

	// send message
	go func() {
		err := pp.sendPump(ws)
		if err != nil {
			pp.Log.Err(err).Msgf("error sending message to passport")
			return
		}
	}()

	// auth
	authJson, err := json.Marshal(&Message{
		Key: "AUTH:SERVERCLIENT",
		Payload: struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}{
			ClientID:     pp.clientID,
			ClientSecret: pp.clientSecret,
		},
	})

	if err != nil {
		return terror.Error(err)
	}

	err = ws.Write(ctx, websocket.MessageText, authJson)
	if err != nil {
		return terror.Error(err)
	}

	// listening for message
	for {
		ctx := context.Background()
		_, r, err := ws.Reader(ctx)
		if err != nil {
			return terror.Error(err)
		}

		payload, err := ioutil.ReadAll(r)
		if err != nil {
			return terror.Error(err)
		}

		v, err := jason.NewObjectFromBytes(payload)
		if err != nil {
			pp.Log.Err(err).Msgf(`error making object from bytes`)
			continue
		}

		cmdKey, err := v.GetString("key")
		if err != nil {
			pp.Log.Err(err).Msgf(`missing json key "key"`)
			continue
		}

		if cmdKey == "" {
			pp.Log.Err(fmt.Errorf("missing key value")).Msgf("missing key/command value")
			continue
		}

		// send received message to the hub to handle
		pp.Events.Trigger(context.Background(), Event(fmt.Sprintf("PASSPORT:%s", cmdKey)), payload)

		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return terror.Error(err)
		}
		if err != nil {
			pp.Log.Err(err).Msg(err.Error())
			return terror.Error(err)
		}
	}
}

func (pp *Passport) sendPump(c *websocket.Conn) error {
	for {
		select {
		case msg := <-pp.send:
			err := writeTimeout(msg, time.Second*5, c)
			if err != nil {
				pp.close()
				return terror.Error(err)
			}
		case <-pp.ctx.Done():
			return terror.Error(pp.ctx.Err())
		}
	}
}

// writeTimeout enforces a timeout on websocket writes
func writeTimeout(msg *Message, timeout time.Duration, c *websocket.Conn) error {
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
