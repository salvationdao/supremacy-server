package gamelog

import (
	"os"

	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

var L *zerolog.Logger

func New(environment, level string) {
	log := log_helpers.LoggerInitZero(environment, level)

	if environment == "production" || environment == "staging" {
		logPtr := zerolog.New(os.Stdout)
		logPtr = logPtr.With().Logger()
		log = &logPtr
	}
	log.Info().Msg("zerolog initialised")
	if L != nil {
		panic("GameLog already initialised")
	}
	L = log
	*L = L.With().Caller().Logger()
}
