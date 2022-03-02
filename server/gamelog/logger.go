package gamelog

import (
	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

var GameLog *zerolog.Logger

func New(environment, level string) {
	gameLog := log_helpers.LoggerInitZero(environment, level)
	gameLog.Info().Msg("zerolog initialised")
	if GameLog != nil {
		panic("GameLog already initialised")
	}
	GameLog = gameLog
}
