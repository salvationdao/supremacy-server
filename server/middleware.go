package server

import (
	"context"
	"fmt"
	"github.com/ninja-syndicate/ws"
	"server/db/boiler"
	"server/gamedb"
)

type SecureCommandFunc func(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error
type SecureFactionCommandFunc func(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error
type SecureSecretFactionCommandFunc func(ctx context.Context, secretKey string, secret string, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error

func MustSecure(fn SecureCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		return fn(ctx, user, key, payload, reply)
	}
}

func MustSecureFaction(fn SecureFactionCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		// get faction id
		if !user.FactionID.Valid {
			return fmt.Errorf("faction id is required")
		}

		return fn(ctx, user, user.FactionID.String, key, payload, reply)
	}
}

func RetrieveUser(ctx context.Context) (*boiler.Player, error) {
	userID, ok := ctx.Value("user_id").(string)

	if !ok || userID == "" {
		return nil, fmt.Errorf("can not retrieve user id")
	}

	user, err := boiler.FindPlayer(gamedb.StdConn, userID)
	if err != nil {
		return nil, fmt.Errorf("not authorized to access this endpoint")
	}

	return user, nil
}
