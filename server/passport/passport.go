package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
)

type Command string

type ReplyFunc func(interface{})

type CommandFunc func(ctx context.Context, payload []byte, reply ReplyFunc) error

type Request struct {
	*Message
	ReplyChannel chan []byte
	ErrChan      chan error
}

type Message struct {
	Key           Command     `json:"key"`
	Payload       interface{} `json:"payload"`
	TransactionID string      `json:"transactionID"`
	context       context.Context
}

type Passport struct {
	Conn *Connection

	Log      *zerolog.Logger
	addr     string
	commands map[Command]CommandFunc
	Events   Events
	send     chan *Request

	clientID     string
	clientSecret string

	ctx   context.Context
	close context.CancelFunc
}

func NewPassport(ctx context.Context, logger *zerolog.Logger, addr, clientID, clientSecret string) *Passport {
	ctx, cancel := context.WithCancel(ctx)
	newPP := &Passport{

		Log:      logger,
		addr:     addr,
		commands: make(map[Command]CommandFunc),
		send:     make(chan *Request),
		Events:   Events{map[Event][]EventHandler{}, sync.RWMutex{}},

		clientID:     clientID,
		clientSecret: clientSecret,

		ctx:   ctx,
		close: cancel,
	}

	return newPP
}

func (pp *Passport) Close() {
	pp.Log.Info().Msg("Stopping passport connection")
	//pp.Conn.closeChan <- fmt.Errorf("close called")
	if pp.Conn.ws != nil {
		err := pp.Conn.ws.Close(websocket.StatusNormalClosure, "close called")
		if err != nil {
			pp.Log.Warn().Err(err).Msg("")
		}
	}
}

type Connection struct {
	Connected bool
	ws        *websocket.Conn
	//*sync.Mutex
	lock sync.Mutex
	cond *sync.Cond
}

func (pp *Passport) Connect(ctx context.Context) error {

	// this holds the callbacks for some requests
	callbackChannels := make(map[string]chan []byte)
	pp.Conn = &Connection{
		Connected: false,
		lock:      sync.Mutex{},
	}
	pp.Conn.cond = sync.NewCond(&pp.Conn.lock)

	// set up the connection in a goroutine that loops checking for connection and if false attempts to reconnect
	go func() {
		for {
			pp.Conn.lock.Lock()
			for pp.Conn.Connected { // while connected, wait
				pp.Conn.cond.Wait() // wait for connect to be disconnected, we wait in a loop because we cannot be sure we're disconnected on signal
			}

			pp.Log.Info().Msgf("Attempting to connect to passport on %v", pp.addr)
			var err error
			pp.Conn.ws, _, err = websocket.Dial(ctx, pp.addr, &websocket.DialOptions{
				//HTTPClient:nil,
				//HTTPHeader:nil,
				//Subprotocols:nil,
				//CompressionMode:0,
				//CompressionThreshold:0,
			})
			if err != nil {
				pp.Log.Warn().Err(err).Msg("Failed to connect to passport server, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				pp.Conn.lock.Unlock()
				continue
			}

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
				pp.Log.Warn().Err(err).Msg("Failed to marshall auth request to passport server, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				pp.Conn.lock.Unlock()
				continue
			}

			err = pp.Conn.ws.Write(ctx, websocket.MessageText, authJson)
			if err != nil {
				pp.Log.Warn().Err(err).Msg("Failed to auth with passport server, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				pp.Conn.lock.Unlock()
				continue
			}
			pp.Log.Info().Msg("Connection and auth to passport server successful")
			pp.Conn.Connected = true
			pp.Conn.cond.Broadcast()
			pp.Conn.lock.Unlock()
		}

	}()

	//// TODO: setup ping pong
	//go func() {
	//	for {
	//		if pp.ws != nil && pp.WSConnected {
	//			err := pp.ws.Ping(pp.ctx)
	//			if err != nil {
	//				pp.Log.Warn().Err(err).Msg("failed to ping passport")
	//				pp.WSConnected = false
	//			}
	//		}
	//	}
	//}()

	// send messages
	go pp.sendPump(callbackChannels)

	// listening for message
	for {
		select {
		case <-ctx.Done():
			pp.Close()
			return fmt.Errorf("context canceled")
		default:
			pp.Conn.lock.Lock()
			for !pp.Conn.Connected {
				pp.Conn.cond.Wait() // while not connected, wait, we wait in a loop because we cannot be sure the condition is true when signaled the release
			}
			pp.Conn.lock.Unlock()

			_, r, err := pp.Conn.ws.Reader(ctx)
			if err != nil {
				pp.Conn.Connected = false
				pp.Conn.cond.Broadcast()
				pp.Log.Warn().Err(err).Msg("issue reading from passport connection")
				continue
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

			transactionID, err := v.GetString("transactionID")
			if err != nil {
				continue
			}

			// if we have a transactionID call the channel in the callback map
			if transactionID != "" {
				cb, ok := callbackChannels[transactionID]
				if ok {
					cb <- payload
				} else {
					pp.Log.Warn().Msgf("missing callback for transactionID %s", transactionID)
				}
			}

			cmdKey, err := v.GetString("key")
			if err != nil {
				pp.Log.Err(err).Msgf(`error getting key from payload`)
				continue
			}

			if cmdKey == "" {
				pp.Log.Err(fmt.Errorf("missing key value")).Msgf("missing key/command value")
				continue
			}

			// send received message to the hub to handle
			pp.Events.Trigger(context.Background(), Event(fmt.Sprintf("PASSPORT:%s", cmdKey)), payload)
		}
	}
}

func (pp *Passport) sendPump(callbackChannels map[string]chan []byte) {
	for {
		msg := <-pp.send
		if !pp.Conn.Connected {
			if msg.ErrChan != nil {
				msg.ErrChan <- terror.Error(fmt.Errorf("no passport connection"))
			}
			continue // if no connection then just drop the send message
		}

		// if we got given a tx id and a reply channel
		if msg.ReplyChannel != nil && msg.TransactionID != "" {
			callbackChannels[msg.TransactionID] = msg.ReplyChannel
		}

		err := writeTimeout(&Message{
			Key:           msg.Key,
			Payload:       msg.Payload,
			TransactionID: msg.TransactionID,
			context:       msg.context,
		}, time.Second*5, pp.Conn.ws)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				if msg.ErrChan != nil {
					msg.ErrChan <- terror.Error(fmt.Errorf("no passport connection"))
				}
				pp.Conn.Connected = false
				pp.Conn.cond.Broadcast()
				continue
			}
			if msg.ErrChan != nil {
				msg.ErrChan <- terror.Error(err)
			}
			pp.Log.Warn().Err(err).Msg("failed to send message to passport")
		}
	}
}

// writeTimeout enforces a timeout on websocket writes
func writeTimeout(msg *Message, timeout time.Duration, c *websocket.Conn) error {
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
