package passport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jpillora/backoff"
	"github.com/ninja-software/terror/v2"
	"io"
	"sync"
	"time"

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
	Key           Command     `json:"key"`
	Payload       interface{} `json:"payload"`
	TransactionID string      `json:"transactionID"`
	context       context.Context
}

type Passport struct {
	//Conn *Connection
	ws        *websocket.Conn
	Connected bool
	Log       *zerolog.Logger
	addr      string
	commands  map[Command]CommandFunc
	Events    Events
	send      chan *Request

	clientToken string

	ctx   context.Context
	close context.CancelFunc
}

func NewPassport(ctx context.Context, logger *zerolog.Logger, addr, clientToken string) *Passport {
	ctx, cancel := context.WithCancel(ctx)
	newPP := &Passport{

		Log:      logger,
		addr:     addr,
		commands: make(map[Command]CommandFunc),
		send:     make(chan *Request),
		Events:   Events{map[Event][]EventHandler{}, sync.RWMutex{}},

		clientToken: clientToken,

		ctx:   ctx,
		close: cancel,
	}

	return newPP
}

type callbackChannel struct {
	ReplyChannel chan []byte
	errChan      chan error
}

type responseError struct {
	Key           string `json:"key"`
	TransactionID string `json:"transactionID"`
	Message       string `json:"message"`
}

func (pp *Passport) Connect(ctx context.Context) error {
	// this holds the callbacks for some requests
	callbackChannels := make(map[string]*callbackChannel)

	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2,
	}

reconnectLoop:
	for {
		connectCtx, cancel := context.WithCancel(ctx)
		authed := false
		pp.Connected = false

		pp.Log.Info().Msgf("Attempting to connect to passport on %v", pp.addr)
		var err error
		pp.ws, _, err = websocket.Dial(ctx, pp.addr, &websocket.DialOptions{
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

		err = writeTimeout(connectCtx, &Message{
			Key:           "AUTH:SERVERCLIENT",
			TransactionID: authTxID.String(),
			Payload: struct {
				ClientToken string `json:"clientToken"`
			}{
				ClientToken: pp.clientToken,
			},
			context: ctx,
		}, time.Second*5, pp.ws)

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
		go pp.sendPump(connectCtx, cancel, callbackChannels)

		// listening for message
		for {
			select {
			case <-ctx.Done():
				cancel()
				pp.ws.Close(websocket.StatusNormalClosure, "close called")
				return ctx.Err()
			case <-connectCtx.Done():
				pp.ws.Close(websocket.StatusNormalClosure, "close called")
				continue reconnectLoop
			case err := <-authTimeout:
				if err != nil {
					pp.Log.Warn().Err(err).Msg("issue reading from passport connection")
					time.Sleep(b.Duration())
					pp.ws.Close(websocket.StatusNormalClosure, "close called")
					continue reconnectLoop
				}
			default:
				_, r, err := pp.ws.Reader(ctx)
				if err != nil {
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

				transactionID, err := v.GetString("transactionID")
				if err != nil {
					continue
				}

				// if we have a transactionID call the channel in the callback map
				if transactionID != "" {
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
					}

					cb, ok := callbackChannels[transactionID]
					if !ok {
						pp.Log.Warn().Msgf("missing callback for transactionID %s", transactionID)
						continue
					}

					// parse whether it is an error
					errMsg := &responseError{}
					err := json.Unmarshal(payload, errMsg)
					if err != nil {
						cb.errChan <- terror.Error(err, "Invalid json response")
						continue
					}

					if errMsg.Key == "HUB:ERROR" {
						cb.errChan <- terror.Error(fmt.Errorf(errMsg.Message), errMsg.Message)
						continue
					}

					cb.ReplyChannel <- payload
					continue
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
				pp.Events.Trigger(ctx, Event(fmt.Sprintf("PASSPORT:%s", cmdKey)), payload)
			}
		}
	}
}

func (pp *Passport) sendPump(ctx context.Context, cancelFunc context.CancelFunc, callbackChannels map[string]*callbackChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg := <-pp.send

			if msg.TransactionID != "" {
				if msg.ReplyChannel == nil {
					msg.ErrChan <- terror.Error(fmt.Errorf("missing reply channel"))
					continue
				}

				if msg.ErrChan == nil {
					msg.ErrChan <- terror.Error(fmt.Errorf("missing err channel"))
					continue
				}

				callbackChannels[msg.TransactionID] = &callbackChannel{
					ReplyChannel: msg.ReplyChannel,
					errChan:      msg.ErrChan,
				}
			}

			err := writeTimeout(ctx, &Message{
				Key:           msg.Key,
				Payload:       msg.Payload,
				TransactionID: msg.TransactionID,
				context:       msg.context,
			}, time.Second*5, pp.ws)
			if err != nil {
				if msg.ErrChan != nil {
					msg.ErrChan <- terror.Error(err)
				}
				pp.Log.Warn().Err(err).Msg("failed to send message to passport")
				cancelFunc()
			}
		}
	}
}

// writeTimeout enforces a timeout on websocket writes
func writeTimeout(serverCtx context.Context, msg *Message, timeout time.Duration, c *websocket.Conn) error {
	if c == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(msg.context, timeout)
	defer func() {
		cancel()
	}()

	go func() {
		select {
		case <-serverCtx.Done():
			cancel()
			return
		case <-ctx.Done():
			return
		}

	}()

	jsn, err := json.Marshal(msg)
	if err != nil {
		return terror.Error(err)
	}

	return c.Write(ctx, websocket.MessageText, jsn)
}
