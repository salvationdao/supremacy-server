package api

import (
	"context"
	"server"
	"server/db/boiler"
	"server/gamedb"

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

		// get user faction
		player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
		if err != nil {
			return terror.Error(terror.ErrForbidden)
		}

		if !player.FactionID.Valid {
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
func (api *API) SubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// SecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) SecureUserSubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return wsc.Identifier() == ""
	})
}

// SecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) SecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		if wsc.Identifier() == "" {
			return true
		}

		// get user faction
		player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
		if err != nil {
			return true
		}
		return !player.FactionID.Valid
	})
}

// SubscribeCommandWithAuthCheck registers a subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) SubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn HubSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		transactionID, busKey, err := fn(ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		api.MessageBus.Sub(busKey, wsc, transactionID)

		return nil
	})
	unsubscribeKey := hub.HubCommandKey(key + ":UNSUBSCRIBE")
	api.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		transactionID, busKey, err := fn(ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		api.MessageBus.Unsub(busKey, wsc, transactionID)

		return nil
	})

}

/***************************
* Net Message Subscription *
***************************/

// HubNetSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions
type HubNetSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte) (messagebus.BusKey, error)

// NetSubscribeCommand registers a net message subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) NetSubscribeCommand(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc) {
	api.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// NetSecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) NetSecureUserSubscribeCommand(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc) {
	api.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return wsc.Identifier() == ""
	})
}

// NetSecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) NetSecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc) {
	api.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		if wsc.Identifier() == "" {
			return true
		}

		// get user faction
		player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
		if err != nil {
			return true
		}
		return !player.FactionID.Valid
	})
}

// NetSubscribeCommandWithAuthCheck registers a net message subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) NetSubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	var err error
	var busKey messagebus.BusKey
	busKey = ""

	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		busKey, err = fn(ctx, wsc, payload)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		api.MessageBus.SubClient(busKey, wsc)

		return nil
	})

	// Unsubscribe
	unsubscribeKey := hub.HubCommandKey(key + ":UNSUBSCRIBE")
	api.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		if busKey != "" {
			// remove subscription if buskey not empty from message bus
			api.MessageBus.UnsubClient(busKey, wsc)
		}

		return nil
	})
}
