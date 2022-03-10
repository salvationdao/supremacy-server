package passport

import (
	"server/rpcclient"

	"github.com/rs/zerolog"
)

type Passport struct {
	Log         *zerolog.Logger
	RPCClient   *rpcclient.XrpcClient
	clientToken string
}

func NewPassport(logger *zerolog.Logger, addr, clientToken string, comms *rpcclient.XrpcClient) *Passport {
	newPP := &Passport{
		Log:         logger,
		clientToken: clientToken,
		RPCClient:   comms,
	}

	return newPP
}
