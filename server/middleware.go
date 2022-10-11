package server

import (
	"context"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
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

func MustSecureAdmin(fn SecureCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		err = user.L.LoadRole(gamedb.StdConn, true, user, nil)
		if err != nil {
			return terror.Error(err, "Failed to update player's marketing preferences.")
		}

		if user.R != nil && user.R.Role != nil {
			if user.R.Role.RoleType == boiler.RoleNamePLAYER {
				return fmt.Errorf("user has no admin privillege")
			}
		} else {
			return fmt.Errorf("failed to get users role")
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
	userID, ok := ctx.Value("auth_user_id").(string)

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

func MustSecureWithFeature(featureName string, fn SecureCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		for _, pf := range user.R.PlayersFeatures {

			if pf.FeatureName == featureName {
				return fn(ctx, user, key, payload, reply)
			}
		}

		return terror.Error(fmt.Errorf("player: %s does not have necessary feature", user.ID), "You do not have the necessary feature to perform this action, try again or contact support.")
	}
}

func MustSecureFactionWithFeature(featureName string, fn SecureFactionCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		for _, pf := range user.R.PlayersFeatures {
			if pf.FeatureName == featureName {
				return fn(ctx, user, user.FactionID.String, key, payload, reply)
			}
		}
		return terror.Error(fmt.Errorf("player: %s does not have necessary feature", user.ID), "You do not have the necessary feature to perform this action, try again or contact support.")

	}
}

// Tracer is a ws middleware used to implement datadog for WS Handlers.
func Tracer(fn ws.CommandFunc, environment string) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		requestUri, _ := ctx.Value("Origin").(string)
		if environment != "development" {
			span, augmentedCtx := tracer.StartSpanFromContext(
				ctx,
				"ws_handler",
				tracer.ResourceName(key),
				tracer.Tag("ws_key", key),
				tracer.Tag("env", environment),
				tracer.Tag("origin", requestUri),
			)
			defer span.Finish()
			ctx = augmentedCtx
		}
		return fn(ctx, key, payload, reply)
	}
}

// SecureUserTracer is a ws middleware used to implement datadog for WS Handlers.
func SecureUserTracer(fn SecureCommandFunc, environment string) SecureCommandFunc {
	return func(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
		requestUri, _ := ctx.Value("Origin").(string)
		if environment != "development" {
			span, augmentedCtx := tracer.StartSpanFromContext(
				ctx,
				"ws_handler",
				tracer.ResourceName(key),
				tracer.Tag("ws_key", key),
				tracer.Tag("env", environment),
				tracer.Tag("origin", requestUri),
			)
			defer span.Finish()
			ctx = augmentedCtx
		}
		return fn(ctx, user, key, payload, reply)
	}
}

// SecureFactionTracer is a ws middleware used to implement datadog for WS Handlers (factions).
func SecureFactionTracer(fn SecureFactionCommandFunc, environment string) SecureFactionCommandFunc {
	return func(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
		requestUri, _ := ctx.Value("Origin").(string)
		if environment != "development" {
			span, augmentedCtx := tracer.StartSpanFromContext(
				ctx,
				"ws_handler",
				tracer.ResourceName(key),
				tracer.Tag("ws_key", key),
				tracer.Tag("env", environment),
				tracer.Tag("origin", requestUri),
			)
			defer span.Finish()
			ctx = augmentedCtx
		}
		return fn(ctx, user, user.FactionID.String, key, payload, reply)
	}
}

func AddOriginToCtx() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "Origin", r.Header.Get("Origin"))))
			return
		}
		return http.HandlerFunc(fn)
	}
}

func RestDatadogTrace(environment string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if environment != "development" {
				span, augmentedCtx := tracer.StartSpanFromContext(
					r.Context(),
					"http_handler",
					tracer.ResourceName(fmt.Sprintf("%s %s", r.Method, r.URL.Path)),
					tracer.Tag("http.method", r.Method),
					tracer.Tag("http.url", r.URL.Path),
					tracer.Tag("origin", r.Header.Get("Origin")),
				)
				defer span.Finish()
				r = r.WithContext(augmentedCtx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
