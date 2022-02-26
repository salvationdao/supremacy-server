package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"server/gamelog"
	"time"

	"github.com/ninja-software/terror/v2"
)

// EncodeJSON will encode json to response writer and return status ok.
// Warning, this is to be used with `return` or error tracing may be inaccurate or even missing
func EncodeJSON(w http.ResponseWriter, result interface{}) (int, error) {
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		// create custom terror.TError struct because otherwise trace will fault EncodeJSON function instead
		// of where it actually faulted
		// this is because `return EncodeJSON()` has been used without use of terror.Error()
		pc, file, line, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()

		terr := &terror.TError{
			Level:    terror.ErrLevelError,
			File:     file,
			FuncName: funcName,
			Line:     line,
			Message:  err.Error(),
			Err:      err,
			ErrKind:  terror.ErrKindSystem,
			Meta:     map[string]string{},
		}

		return http.StatusInternalServerError, terr
	}
	return http.StatusOK, nil
}

func Gotimeout(cb func(), timeout time.Duration, errorCallback func(error)) {
	start := make(chan bool, 1)

	go func() {
		select {
		case <-start:
			cb()
		case <-time.After(timeout):
			err := errors.New("callback has timed out")
			gamelog.GameLog.Warn().Err(err).Msgf("Failed to connect to passport server. Please make sure passport is running.") /*  */
			errorCallback(err)
		}
	}()
	start <- true
	close(start)
}
