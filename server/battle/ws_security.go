package battle

import (
	"context"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

func (opts *Opts) Command(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		return fn(ctx, wsc, payload, reply)
	})
}

func (opts *Opts) SecureUserCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if wsc.Identifier() == "" {
			return terror.Error(terror.ErrForbidden)
		}

		return fn(ctx, wsc, payload, reply)
	})
}

type FactionCommandFunc func(ctx context.Context, hub *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error

func GetPlayerFactionID(userID uuid.UUID) (uuid.UUID, error) {
	player, err := boiler.FindPlayer(gamedb.StdConn, userID.String())
	if err != nil {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find player from user id")
		return uuid.Nil, terror.Error(err)
	}

	if !player.FactionID.Valid {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("faction id is invalid")

		return uuid.Nil, terror.Error(terror.ErrForbidden)
	}

	fuuid, err := uuid.FromString(player.FactionID.String)
	if err != nil || fuuid.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to convert player faction id into uuid")
		return uuid.Nil, terror.Error(err)
	}
	return fuuid, nil
}

func (opts *Opts) SecureUserFactionCommand(key hub.HubCommandKey, fn FactionCommandFunc) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		userID := uuid.FromStringOrNil(wsc.Identifier())
		if userID.IsNil() {
			return terror.Error(terror.ErrForbidden)
		}

		factionID, err := GetPlayerFactionID(userID)
		if err != nil || factionID.IsNil() {
			gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find player from user id")
			return terror.Error(err)
		}

		return fn(ctx, wsc, payload, factionID, reply)
	})
}

// HubSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions (returns sessionID and arguments)
type HubSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error)

// SubscribeCommand registers a subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) SubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	opts.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// SecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) SecureUserSubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	opts.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return wsc.Identifier() == ""
	})
}

// SecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) SecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	opts.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		userID := uuid.FromStringOrNil(wsc.Identifier())
		if userID.IsNil() {
			return true
		}

		factionID, _ := GetPlayerFactionID(userID)
		return factionID.IsNil()
	})
}

// SubscribeCommandWithAuthCheck registers a subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) SubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn HubSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		transactionID, busKey, err := fn(ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		if opts.MessageBus == nil {
			gamelog.L.Error().Msg("messagebus is nil")
			return fmt.Errorf("messagebus is nil")
		}

		opts.MessageBus.Sub(busKey, wsc, transactionID)
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
func (opts *Opts) NetSubscribeCommand(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc) {
	opts.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return false
	})
}

// NetSecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) NetSecureUserSubscribeCommand(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc) {
	opts.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		return wsc.Identifier() == ""
	})
}

// NetSecureUserFactionSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (opts *Opts) NetSecureUserFactionSubscribeCommand(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc) {
	opts.NetSubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		userID := uuid.FromStringOrNil(wsc.Identifier())
		if userID.IsNil() {
			return true
		}

		// get user faction
		factionID, _ := GetPlayerFactionID(userID)
		return factionID.IsNil()
	})
}

// NetSubscribeCommandWithAuthCheck registers a net message subscription command to the hub
//
// If fn is not provided, will use default
func (opts *Opts) NetSubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn HubNetSubscribeCommandFunc, authCheck func(wsc *hub.Client) bool) {
	opts.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if authCheck(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		busKey, err := fn(ctx, wsc, payload)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		opts.MessageBus.SubClient(busKey, wsc)

		return nil
	})
}
