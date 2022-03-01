package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/ninja-software/terror/v2"
)

type ErrorMessage string

const (
	Unauthorised          ErrorMessage = "Unauthorised - Please log in or contact System Administrator"
	Forbidden             ErrorMessage = "Forbidden - You do not have permissions for this, please contact System Administrator"
	InternalErrorTryAgain ErrorMessage = "Internal Error - Please try again in a few minutes or Contact Support"
	InputError            ErrorMessage = "Input Error - Please try again"
)

// ErrorObject is used by the front end react-fetching-library
type ErrorObject struct {
	Message string `json:"message"`
}

func WithError(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contents, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewReader(contents))
		code, err := next(w, r)
		if err != nil {
			terror.Echo(err)
			errObj := &ErrorObject{Message: err.Error()}
			jsonErr, err := json.Marshal(errObj)
			if err != nil {
				terror.Echo(err)
				http.Error(w, `{"message":"JSON failed, please contact IT.","error_code":"00001"}`, code)
				return
			}
			http.Error(w, string(jsonErr), code)
			return
		}
	}
	return fn
}

// generateErrorCode generates codes for error messages
func generateErrorCode(codeLen int) string {
	shortCodeArrayChars := []string{"a", "b", "c", "d", "e", "f", "g", "h", "j", "k", "m", "n", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	shortCodeArrayNumbers := []string{"2", "3", "4", "5", "6", "7", "8", "9"}
	shortCode := ""

	for i := 0; i < codeLen; i++ {
		if i%2 == 0 {
			shortCode += shortCodeArrayChars[randInt(0, len(shortCodeArrayChars))]
			continue
		}
		shortCode += shortCodeArrayNumbers[randInt(0, len(shortCodeArrayNumbers))]
	}
	return shortCode
}

// randInt gives rand int between given ints
func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func WithToken(apiToken string, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Authorization") != apiToken {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}
		next(w, r)
	}
	return fn
}

// check passport http request secret
func WithPassportSecret(secret string, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Passport-Authorization") != secret {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}
		next(w, r)
	}
	return fn
}
