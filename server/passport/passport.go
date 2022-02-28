package passport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"server/comms"
	"server/gamelog"
	"server/helpers"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sasha-s/go-deadlock"

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
	ws          *websocket.Conn
	Connected   bool
	Log         *zerolog.Logger
	addr        string
	commands    map[Command]CommandFunc
	Events      Events
	send        chan *Message
	Comms       *comms.C
	callbacks   deadlock.Map
	clientToken string
}

func NewPassport(logger *zerolog.Logger, addr, clientToken string, comms *comms.C) *Passport {
	newPP := &Passport{

		Log:      logger,
		addr:     addr,
		commands: make(map[Command]CommandFunc),
		send:     make(chan *Message, 5),
		Events:   Events{map[Event][]EventHandler{}, deadlock.RWMutex{}},

		callbacks:   deadlock.Map{},
		clientToken: clientToken,
		Comms:       comms,
	}

	return newPP
}

// type callbackChannel struct {
// 	ReplyChannel chan []byte
// 	errChan      chan error
// }

var connectionAttempts = 0

func (pp *Passport) Connect() error {
	// this holds the callbacks for some requests

	pp.Log.Info().Msgf("Attempting to connect to passport on %v", pp.addr)
	var err error
	pp.ws, _, err = websocket.Dial(context.Background(), pp.addr, &websocket.DialOptions{})
	if err != nil {
		pp.Log.Warn().Err(err).Msgf("Failed to connect to passport server. Please make sure passport is running.")
		connectionAttempts++
		if connectionAttempts > 5 {
			pp.Log.Warn().Err(err).Msgf("Failed to connect to passport server. Attempted 5 times.")
			time.Sleep(5 * time.Second)
		}
		return pp.Connect()
	}
	pp.ws.SetReadLimit(1048576000) // set to 100mbs

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
		time.Sleep(5 * time.Second)
	}

	err = pp.ws.Write(context.Background(), websocket.MessageText, jsn)
	if err != nil {
		pp.Log.Warn().Err(err).Msgf("Failed to connect to passport server")
		pp.ws.Close(websocket.StatusNormalClosure, "close called")
		time.Sleep(5 * time.Second)
	}

	// send messages
	go pp.sendPump()

	wg := deadlock.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// listening for message
		for {

			_, r, err := pp.ws.Reader(context.Background())
			if err != nil {
				pp.Log.Warn().Err(err).Msg("issue reading from passport connection. reconnecting in 10 seconds...")
				pp.ws.Close(websocket.StatusNormalClosure, "close called")
				time.Sleep(10 * time.Second)
				_ = pp.Connect()
			}
			if r == nil {
				pp.Log.Warn().Err(err).Msg("ws reader is nil, continue")
				continue
			}
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, r)
			if err != nil {
				pp.Log.Warn().Err(err).Msg("io copy failed, passport connect")
				continue
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
					pp.Log.Warn().Err(err).Msgf("Failed to read key from message from passport")
					continue
				}
				if cmdKey == "HUB:ERROR" {
					pp.Log.Warn().Err(err).Msgf("Failed to auth to passport server, retrying in 5 seconds...")
					pp.ws.Close(websocket.StatusNormalClosure, "close called")
					time.Sleep(5 * time.Second)
					_ = pp.Connect()
				}
				connectionAttempts = 0
				pp.Log.Info().Msgf("Successfully connected to passport on %v", pp.addr)
				wg.Done()
				continue
			} else if transactionID != "" {
				cb, ok := pp.callbacks.LoadAndDelete(transactionID)
				if ok {
					callback := cb.(func(msg []byte))
					helpers.Gotimeout(func() { callback(payload) }, 10*time.Second, func(err error) {
						gamelog.GameLog.Warn().Err(err).Msg("callback from passport message has timed out (10 seconds).")
					})
				}
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
			pp.Events.Trigger(Event(fmt.Sprintf("PASSPORT:%s", cmdKey)), payload)
		}
	}()
	wg.Wait()
	return nil
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
			if pp.ws != nil {
				err = pp.ws.Write(ctx, websocket.MessageText, jsn)
				if err != nil {
					pp.Log.Warn().Err(err).Msg("ws write error")
				}
			}
		}()
		if err != nil {
			pp.Log.Warn().Err(err).Msg("failed to send to passport")
			continue
		}
	}
}
