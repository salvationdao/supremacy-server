package api

import (
	"github.com/ninja-syndicate/ws"
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

func (api *API) SecureUserFeatureCheckCommand(featureType string, key string, fn server.SecureCommandFunc) {
	api.SecureUserCommander.Command(string(key), server.MustSecureWithFeature(featureType, fn))
}

func (api *API) SecureUserFactionFeatureCheckCommand(featureType string, key string, fn server.SecureFactionCommandFunc) {
	api.SecureFactionCommander.Command(string(key), server.MustSecureFactionWithFeature(featureType, fn))
}
