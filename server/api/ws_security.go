package api

import (
	"context"
	"encoding/json"
	"server"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

func (api *API) Command(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		return fn(ctx, wsc, payload, reply)
	})
}

func (api *API) SecureUserCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if wsc.Identifier() == "" {
			return terror.Error(terror.ErrForbidden)
		}

		return fn(ctx, wsc, payload, reply)
	})
}

func (api *API) SecureUserFactionCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if wsc.Identifier() == "" {
			return terror.Error(terror.ErrForbidden)
		}

		hcd, err := api.getClientDetailFromChannel(wsc)
		if err != nil {
			return terror.Error(err)
		}

		if hcd.FactionID.IsNil() {
			return terror.Error(terror.ErrForbidden)
		}

		return fn(ctx, wsc, payload, reply)
	})
}

// SecureUserCommandWithPerm registers a command to the hub that will only run if the websocket has authenticated and the user has the specified permission
func (api *API) SecureUserCommandWithPerm(key hub.HubCommandKey, fn hub.HubCommandFunc, perm server.Perm) {
	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		// TODO: Add middleware
		//if wsc.User == nil || !wsc.User.HasPermission(perm) {
		//	return hub.ErrSecureError
		//}
		// check 2fa
		//err := hub.check2FA(ctx, wsc)
		//if err != nil {
		//	return ErrSecureError
		//}
		return fn(ctx, wsc, payload, reply)
	})
}

// HubSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions (returns sessionID and arguments)
type HubSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error)

// SubscribeCommand registers a subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) SubscribeCommand(key hub.HubCommandKey, fn ...HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// SecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) SecureUserSubscribeCommand(key hub.HubCommandKey, fn ...HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		if wsc.Identifier() == "" {
			return true
		}
		return false
	})
}

// SecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) SecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn ...HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		if wsc.Identifier() == "" {
			return true
		}

		hcd, err := api.getClientDetailFromChannel(wsc)
		if err != nil {
			return true
		}

		if hcd.FactionID.IsNil() {
			return true
		}
		return false
	})
}

// SubscribeCommandWithAuthCheck registers a subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) SubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn []HubSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	var err error
	busKey := messagebus.BusKey("")
	transactionID := ""

	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		transactionID, busKey, err = fn[0](ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		api.MessageBus.Sub(busKey, wsc, transactionID)

		return nil
	})

	// Unsubscribe
	unsubscribeKey := hub.HubCommandKey(key + ":UNSUBSCRIBE")
	api.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
			api.MessageBus.Unsub(busKey, wsc, req.TransactionID)
		}

		return nil
	})
}
