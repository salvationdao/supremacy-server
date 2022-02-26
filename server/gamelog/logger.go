package gamelog

import (
	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

var GameLog *zerolog.Logger

func New(environment, level string) *zerolog.Logger {
	GameLog := log_helpers.LoggerInitZero(environment, level)
	GameLog.Info().Msg("zerolog initialised")
	return GameLog
}
