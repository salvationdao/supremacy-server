package api

import (
	"context"
	"server"

	"github.com/ninja-syndicate/ws"
)

func MustLogin(ctx context.Context) bool {
	// get user from xsyn service
	_, err := server.RetrieveUser(ctx)
	if err != nil {
		return false
	}

	return true
}

func MustHaveFaction(ctx context.Context) bool {
	// get user from xsyn service
	u, err := server.RetrieveUser(ctx)
	if err != nil {
		return false
	}

	return u.FactionID.Valid
}

func (api *API) Command(key string, fn ws.CommandFunc) {
	api.Commander.Command(key, fn)
}

func (api *API) SecureUserCommand(key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecure(fn))
}

func (api *API) SecureUserFactionCommand(key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(string(key), server.MustSecureFaction(fn))
}

func (api *API) SecureUserFeatureCheckCommand(featureType string, key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecureWithFeature(featureType, fn))
}

func (api *API) SecureUserFactionFeatureCheckCommand(featureType string, key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(string(key), server.MustSecureFactionWithFeature(featureType, fn))
}
