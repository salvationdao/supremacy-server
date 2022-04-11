package gamelog

import (
	"bytes"
	"encoding/json"
	"os"
	"runtime/debug"

	"github.com/DataDog/gostackparse"
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

// LogPanicRecovery is intended to be used inside of a recover block.
// 
// `r interface{}` is the empty interface returned from `recover()`
// 
// Usage
//   if r := recover(); r != nil {
//      gamelog.LogPanicRecovery("Message", r)
//      // other recovery code
//   }
func LogPanicRecovery(msg string, r interface{}) {
	event := L.WithLevel(zerolog.PanicLevel).Interface("panic", r)
	s := debug.Stack()
	stack, errs := gostackparse.Parse(bytes.NewReader(s))
	if len(errs) != 0 {
		event = event.Errs("stack_parsing_errors", errs)
	}
	jStack, err := json.Marshal(stack)
	if err != nil {
		event.AnErr("stack_marshal_error", err).Msg(msg)
		return
	}
	event.RawJSON("stack", jStack).Msg(msg)
}
