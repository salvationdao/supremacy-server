package battle

import (
	"context"
	"encoding/json"
	"server"
	"server/db"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (opts *Opts) Command(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		span := tracer.StartSpan("ws.Command", tracer.ResourceName(string(key)))
		defer span.Finish()
		return fn(ctx, wsc, payload, reply)
	})
}

func (opts *Opts) SecureUserCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		span := tracer.StartSpan("ws.SecureUserCommand", tracer.ResourceName(string(key)))
		span.SetTag("user", wsc.Identifier())
		defer span.Finish()
		if wsc.Identifier() == "" {
			return terror.Error(terror.ErrForbidden)
		}

		return fn(ctx, wsc, payload, reply)
	})
}

type FactionCommandFunc func(ctx context.Context, hub *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error

func (opts *Opts) SecureUserFactionCommand(key hub.HubCommandKey, fn FactionCommandFunc) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		span := tracer.StartSpan("ws.SecureUserFactionCommand", tracer.ResourceName(string(key)))
		span.SetTag("user", wsc.Identifier())
		defer span.Finish()

		userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
		if userID.IsNil() {
			return terror.Error(terror.ErrForbidden)
		}

		// get user faction
		factionID, err := db.PlayerFactionIDGet(ctx, opts.Conn, userID)
		if err != nil {
			return terror.Error(err)
		}

		if factionID == nil || factionID.IsNil() {
			return terror.Error(terror.ErrForbidden)
		}

		return fn(ctx, wsc, payload, *factionID, reply)
	})
}

// HubSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions (returns sessionID and arguments)
type HubSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error)

// SubscribeCommand registers a subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) SubscribeCommand(key hub.HubCommandKey, fn ...HubSubscribeCommandFunc) {
	opts.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// SecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) SecureUserSubscribeCommand(key hub.HubCommandKey, fn ...HubSubscribeCommandFunc) {
	opts.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return wsc.Identifier() == ""
	})
}

// SecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) SecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn ...HubSubscribeCommandFunc) {
	opts.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
		if userID.IsNil() {
			return true
		}

		// get user faction
		factionID, err := db.PlayerFactionIDGet(context.Background(), opts.Conn, userID)
		if err != nil {
			return true
		}

		return factionID == nil || factionID.IsNil()
	})
}

// SubscribeCommandWithAuthCheck registers a subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) SubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn []HubSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	var err error
	busKey := messagebus.BusKey("")
	transactionID := ""

	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		transactionID, busKey, err = fn[0](ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		opts.MessageBus.Sub(busKey, wsc, transactionID)

		return nil
	})

	// Unsubscribe
	unsubscribeKey := hub.HubCommandKey(key + ":UNSUBSCRIBE")
	opts.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		req := &hub.HubCommandRequest{}
		err := json.Unmarshal(payload, req)
		if err != nil {
			return terror.Error(err, "Invalid request received")
		}

		// remove subscription if buskey not empty from message bus
		if busKey != "" {
			opts.MessageBus.Unsub(busKey, wsc, req.TransactionID)
		}

		return nil
	})
}

/***************************
* Net Message Subscription *
***************************/

// HubNetSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions
type HubNetSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte) (messagebus.NetBusKey, error)

// NetSubscribeCommand registers a net message subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) NetSubscribeCommand(key hub.HubCommandKey, fn ...HubNetSubscribeCommandFunc) {
	opts.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// NetSecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) NetSecureUserSubscribeCommand(key hub.HubCommandKey, fn ...HubNetSubscribeCommandFunc) {
	opts.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return wsc.Identifier() == ""
	})
}

// NetSecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) NetSecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn ...HubNetSubscribeCommandFunc) {
	opts.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
		if userID.IsNil() {
			return true
		}

		// get user faction
		factionID, err := db.PlayerFactionIDGet(context.Background(), opts.Conn, userID)
		if err != nil {
			return true
		}

		return factionID == nil || factionID.IsNil()
	})
}

// NetSubscribeCommandWithAuthCheck registers a net message subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) NetSubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn []HubNetSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	var err error
	var busKey messagebus.NetBusKey
	busKey = ""

	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		busKey, err = fn[0](ctx, wsc, payload)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		opts.NetMessageBus.Sub(busKey, wsc)

		return nil
	})

	// Unsubscribe
	unsubscribeKey := hub.HubCommandKey(key + ":UNSUBSCRIBE")
	opts.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		if busKey != "" {
			// remove subscription if buskey not empty from message bus
			opts.NetMessageBus.Unsub(busKey, wsc)
		}

		return nil
	})
}
