package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-syndicate/ws"
	"os"
	"server"
)

func (api *API) Command(key string, fn ws.CommandFunc) {
	api.Commander.Command(key, fn)
}

func (api *API) SecureUserCommand(key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecure(fn))
}

func (api *API) SecureUserFactionCommand(key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(string(key), server.MustSecureFaction(fn))
}

func MustHaveFaction(ctx context.Context) bool {
	// get user from xsyn service
	u, err := server.RetrieveUser(ctx)
	if err != nil {
		return false
	}

	return u.FactionID.Valid
}

func MustLogin(ctx context.Context) bool {
	// get user from xsyn service
	_, err := server.RetrieveUser(ctx)
	if err != nil {
		return false
	}

	return true
}

func MustMatchSyndicate(ctx context.Context) bool {
	// NOTE: syndicate is temporary disabled on production
	if os.Getenv("GAMESERVER_ENVIRONMENT") == "production" {
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

	// check syndicate id match
	if !user.SyndicateID.Valid || user.SyndicateID.String != syndicateID {
		return false
	}

	return true
}

func (api *API) SecureUserFeatureCheckCommand(featureType string, key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecureWithFeature(featureType, fn))
}

func (api *API) SecureUserFactionFeatureCheckCommand(featureType string, key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(string(key), server.MustSecureFactionWithFeature(featureType, fn))
}
