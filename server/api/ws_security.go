package api

import (
	"context"
	"server"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-syndicate/ws"
)

func (api *API) Command(key string, fn ws.CommandFunc) {
	api.Commander.Command(key, server.Tracer(fn, api.Config.Environment))
}

func (api *API) SecureUserCommand(key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecure(server.SecureUserTracer(fn, api.Config.Environment)))
}

func (api *API) SecureAdminCommand(key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecureAdmin(server.SecureUserTracer(fn, api.Config.Environment)))
}

func (api *API) SecureUserFactionCommand(key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(string(key), server.MustSecureFaction(server.SecureFactionTracer(fn, api.Config.Environment)))
}

func MustHaveFaction(ctx context.Context) bool {
	// get user from xsyn service
	u, err := server.RetrieveUser(ctx)
	if err != nil {
		return false
	}

	return u.FactionID.Valid
}

func MustMatchUserID(ctx context.Context) bool {
	// get auth user id from context
	authUserID, ok := ctx.Value("auth_user_id").(string)
	if !ok || authUserID == "" {
		return false
	}
	// check user id matched the user id on url
	userID := chi.RouteContext(ctx).URLParam("user_id")

	return userID != "" && userID == authUserID
}

func MustMatchSyndicate(ctx context.Context) bool {
	// NOTE: syndicate is ONLY available on development at the moment
	if !server.IsDevelopmentEnv() {
		return false
	}

	cctx := chi.RouteContext(ctx)
	syndicateID := cctx.URLParam("syndicate_id")

	// check syndicate id not empty
	if syndicateID == "" {
		return false
	}

	// get user from xsyn service
	user, err := server.RetrieveUser(ctx)
	if err != nil {
		return false
	}

	return user.SyndicateID.Valid && user.SyndicateID.String == syndicateID
}

func MustHaveUrlParam(paramKey string) func(ctx context.Context) bool {
	return func(ctx context.Context) bool {
		return chi.RouteContext(ctx).URLParam(paramKey) != ""
	}
}

func (api *API) SecureUserFeatureCheckCommand(featureType string, key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(key, server.MustSecureWithFeature(featureType, server.SecureUserTracer(fn, api.Config.Environment)))
}

func (api *API) SecureUserFactionFeatureCheckCommand(featureType string, key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(key, server.MustSecureFactionWithFeature(featureType, server.SecureFactionTracer(fn, api.Config.Environment)))
}
