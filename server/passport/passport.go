package passport

import (
	"server/comms"

	"github.com/rs/zerolog"
)

type Passport struct {
	Log         *zerolog.Logger
	Comms       *comms.C
	clientToken string
}

func NewPassport(logger *zerolog.Logger, addr, clientToken string, comms *comms.C) *Passport {
	newPP := &Passport{

		Log:         logger,
		clientToken: clientToken,
		Comms:       comms,
	}

	return newPP
}
