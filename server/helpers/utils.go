package helpers

import (
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
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
			gamelog.L.Warn().Err(err).Msgf("Failed to connect to passport server. Please make sure passport is running.") /*  */
			errorCallback(err)
		}
	}()
	start <- true
}

// Unpacks 8 booleans from a single byte
func UnpackBooleansFromByte(packedByte byte) []bool {
	booleans := make([]bool, 8)
	for i := 0; i < 8; i++ {
		booleans[i] = (packedByte & (1 << i)) != 0
	}
	return booleans
}

// Packs up to 8 booleans into a single byte/
func PackBooleansIntoByte(booleans []bool) byte {
	var packedByte byte
	for i, b := range booleans {
		if b {
			packedByte |= 1 << i
		}
	}
	return packedByte
}

// Converts byte array to int32
func BytesToInt(bytes []byte) int32 {
	_ = bytes[3] // bounds check hint to compiler
	return int32(bytes[3]) | int32(bytes[2])<<8 | int32(bytes[1])<<16 | int32(bytes[0])<<24
}

func UUIDArray2StrArray(uids []uuid.UUID) []string {
	result := []string{}
	for _, uid := range uids {
		result = append(result, uid.String())
	}

	return result
}
