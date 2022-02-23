package passport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jpillora/backoff"

	"github.com/gofrs/uuid"

	"github.com/antonholmquist/jason"
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
	Key           Command          `json:"key"`
	Payload       interface{}      `json:"payload"`
	TransactionID string           `json:"transactionID"`
	Callback      func(msg []byte) `json:"-"`
}

type Passport struct {
	//Conn *Connection
	ws        *websocket.Conn
	Connected bool
	Log       *zerolog.Logger
	addr      string
	commands  map[Command]CommandFunc
	Events    Events
	send      chan *Message

	callbacks   sync.Map
	clientToken string
	txRWMutex   sync.RWMutex

	ctx   context.Context
	close context.CancelFunc
}

func NewPassport(ctx context.Context, logger *zerolog.Logger, addr, clientToken string) *Passport {
	ctx, cancel := context.WithCancel(ctx)
	newPP := &Passport{

		Log:      logger,
		addr:     addr,
		commands: make(map[Command]CommandFunc),
		send:     make(chan *Message),
		Events:   Events{map[Event][]EventHandler{}, sync.RWMutex{}},

		callbacks:   sync.Map{},
		clientToken: clientToken,
		txRWMutex:   sync.RWMutex{},

		ctx:   ctx,
		close: cancel,
	}

	return newPP
}

// type callbackChannel struct {
// 	ReplyChannel chan []byte
// 	errChan      chan error
// }

type responseError struct {
	Key           string `json:"key"`
	TransactionID string `json:"transactionID"`
	Message       string `json:"message"`
}

func (pp *Passport) Connect(ctx context.Context) error {
	// this holds the callbacks for some requests

	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2,
	}

reconnectLoop:
	for {
		authed := false
		pp.Connected = false

		pp.Log.Info().Msgf("Attempting to connect to passport on %v", pp.addr)
		var err error
		pp.ws, _, err = websocket.Dial(context.Background(), pp.addr, &websocket.DialOptions{
			//HTTPClient:nil,
			//HTTPHeader:nil,
			//Subprotocols:nil,
			//CompressionMode:0,
			//CompressionThreshold:0,
		})
		if err != nil {
			pp.Log.Warn().Err(err).Msgf("Failed to connect to passport server, retrying in %v seconds...", b.Duration().Seconds())
			//cancel()
			time.Sleep(b.Duration())
			//pp..lock.Unlock()
			continue
		}
		pp.ws.SetReadLimit(104857600) // set to 100mbs

		authTxID := uuid.Must(uuid.NewV4())

		msg := &Message{
			Key:           "AUTH:SERVERCLIENT",
			TransactionID: authTxID.String(),
			Payload: struct {
				ClientToken string `json:"clientToken"`
			}{
				ClientToken: pp.clientToken,
			},
		}

		jsn, err := json.Marshal(msg)
		if err != nil {
			pp.Log.Warn().Err(err).Msg("failed to write json to send to passport")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

		func() {
			defer cancel()
			err = pp.ws.Write(ctx, websocket.MessageText, jsn)
		}()
		if err != nil {
			pp.Log.Warn().Err(err).Msgf("Failed to connect to passport server, retrying in %v seconds...", b.Duration().Seconds())
			time.Sleep(b.Duration())
			pp.ws.Close(websocket.StatusNormalClosure, "close called")
			continue
		}

		authTimeout := make(chan error)

		// if not authed after 5 seconds, cancel context and try reconnecting
		go func() {
			time.Sleep(5 * time.Second)
			if !authed {
				authTimeout <- fmt.Errorf("failed to auth after 5 seconds")
				return
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
		go pp.sendPump()

		// listening for message
		for {
			select {
			case err := <-authTimeout:
				if err != nil {
					pp.Log.Warn().Err(err).Msg("issue reading from passport connection")
					time.Sleep(b.Duration())
					pp.ws.Close(websocket.StatusNormalClosure, "close called")
					continue reconnectLoop
				}
			default:
				_, r, err := pp.ws.Reader(context.Background())
				if err != nil {
					fmt.Println("reader")
					pp.Log.Warn().Err(err).Msg("issue reading from passport connection")
					pp.ws.Close(websocket.StatusNormalClosure, "close called")
					time.Sleep(b.Duration())
					continue reconnectLoop
				}

				var buf bytes.Buffer

				_, err = io.Copy(&buf, r)
				if err != nil {
					pp.Log.Warn().Err(err).Msgf("Failed to connect to passport server, retrying in %v seconds...", b.Duration().Seconds())
					pp.ws.Close(websocket.StatusNormalClosure, "close called")
					time.Sleep(b.Duration())
					continue reconnectLoop
				}

				payload := buf.Bytes()

				v, err := jason.NewObjectFromBytes(payload)
				if err != nil {
					pp.Log.Err(err).Msgf(`error making object from bytes`)
					continue
				}

				transactionID, _ := v.GetString("transactionID")

				if transactionID == authTxID.String() {
					cmdKey, err := v.GetString("key")
					if err != nil {
						pp.Log.Warn().Err(err).Msgf("Failed to auth to passport server, retrying in %v seconds...", b.Duration().Seconds())
						pp.ws.Close(websocket.StatusNormalClosure, "close called")
						time.Sleep(b.Duration())
						continue reconnectLoop
					}
					if cmdKey == "HUB:ERROR" {
						// parse whether it is an error
						errMsg := &responseError{}
						err := json.Unmarshal(payload, errMsg)
						if err != nil {
							pp.Log.Warn().Err(err).Msgf("Failed to auth to passport server, retrying in %v seconds...", b.Duration().Seconds())
							pp.ws.Close(websocket.StatusNormalClosure, "close called")
							time.Sleep(b.Duration())
							continue reconnectLoop
						}

						pp.Log.Warn().Err(fmt.Errorf(errMsg.Message)).Msgf("Failed to auth to passport server, retrying in %v seconds...", b.Duration().Seconds())
						pp.ws.Close(websocket.StatusNormalClosure, "close called")
						time.Sleep(b.Duration())
						continue reconnectLoop
					}

					b.Reset()
					authed = true
					pp.Connected = true
					pp.Log.Info().Msgf("Successfully to connect and authed to passport on %v", pp.addr)
					continue
				} else if transactionID != "" {
					cb, ok := pp.callbacks.LoadAndDelete(transactionID)
					if ok {
						callback := cb.(func(msg []byte))
						go callback(payload)
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
				pp.Events.Trigger(Event(fmt.Sprintf("PASSPORT:%s", cmdKey)), payload)
			}
		}
	}
}

func (pp *Passport) sendPump() {
	for {
		msg := <-pp.send
		if msg.TransactionID != "" {
			if msg.Callback != nil {
				pp.callbacks.Store(msg.TransactionID, msg.Callback)
			}
		}

		jsn, err := json.Marshal(&Message{
			Key:           msg.Key,
			Payload:       msg.Payload,
			TransactionID: msg.TransactionID,
		})
		if err != nil {
			pp.Log.Warn().Err(err).Msg("failed to write json to send to passport")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
		func() {
			defer cancel()
			err = pp.ws.Write(ctx, websocket.MessageText, jsn)
		}()
		if err != nil {
			pp.Log.Warn().Err(err).Msg("failed to send to passport")
			continue
		}
	}
}
