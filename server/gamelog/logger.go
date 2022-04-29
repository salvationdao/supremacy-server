package gamelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/go-chi/chi/middleware"
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
	// Create log event to attach details to
	event := L.WithLevel(zerolog.FatalLevel).CallerSkipFrame(1).Interface("panic", r)

	// Get stack details
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

	// Execute log
	event.RawJSON("stack", jStack).Msg(msg)
}

// DatadogLog implements the datadog logger interface
type DatadogLog struct {
	L *zerolog.Logger
}

func (dl DatadogLog) Log(msg string) {
	dl.L.Info().CallerSkipFrame(1).Msg(msg)
}

func ChiLogger(lvl zerolog.Level) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return middleware.RequestLogger(logFormatter{lvl: lvl})(next)
	}
}

type logFormatter struct {
	lvl zerolog.Level
}

func (l logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	le := logEntry{
		requestID:    middleware.GetReqID(r.Context()),
		userAgent:    r.Header.Get("user-agent"),
		method:       r.Method,
		from:         r.RemoteAddr,
		request_path: fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI),
		protocol:     r.Proto,
		lvl:          l.lvl,
	}

	return le
}

type logEntry struct {
	requestID    string
	userAgent    string
	method       string
	from         string
	request_path string
	protocol     string
	lvl          zerolog.Level
}

func (l logEntry) Write(status int, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	if strings.HasPrefix(l.request_path, "/health_check") {
		return
	}

	e := L.WithLevel(l.lvl)
	e.Str("user_agent", l.userAgent).
		Str("request_id", l.requestID).
		Str("method", l.method).
		Str("from", l.from).
		Str("request_path", l.request_path).
		Int("status", status).
		Int("bytes", bytes).
		Dur("duration", elapsed).
		Send()
}

func (l logEntry) Panic(v interface{}, stack []byte) {
	fmt.Println("panic", v)
}
