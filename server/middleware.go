package server

import (
	"context"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
)

type SecureCommandFunc func(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error
type SecureFactionCommandFunc func(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error

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

	user, err := boiler.Players(
		boiler.PlayerWhere.ID.EQ(userID),
		qm.Load(boiler.PlayerRels.PlayersFeatures),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, fmt.Errorf("not authorized to access this endpoint")
	}

	return user, nil
}

func MustSecureWithFeature(featureType string, fn SecureCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		feature, err := boiler.Features(
			boiler.FeatureWhere.Type.EQ(featureType),
		).One(gamedb.StdConn)
		if err != nil {
			return err
		}

		for _, pf := range user.R.PlayersFeatures {

			if pf.FeatureID == feature.ID {
				return fn(ctx, user, key, payload, reply)
			}
		}

		return terror.Error(fmt.Errorf("player: %s does not have necessary feature", user.ID), "You do not have the necessary feature to perform this action, try again or contact support.")
	}
}

func MustSecureFactionWithFeature(featureType string, fn SecureFactionCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		feature, err := boiler.Features(
			boiler.FeatureWhere.Type.EQ(featureType),
		).One(gamedb.StdConn)
		if err != nil {
			return err
		}

		for _, pf := range user.R.PlayersFeatures {
			if pf.FeatureID == feature.ID {
				return fn(ctx, user, user.FactionID.String, key, payload, reply)
			}
		}
		return terror.Error(fmt.Errorf("player: %s does not have necessary feature", user.ID), "You do not have the necessary feature to perform this action, try again or contact support.")

	}
}
