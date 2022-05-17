package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
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
			jsonErr, wErr := json.Marshal(errObj)
			if wErr != nil {
				DatadogTracer.HttpFinishSpan(r.Context(), code, wErr)
				terror.Echo(wErr)
				http.Error(w, `{"message":"JSON failed, please contact IT.","error_code":"00001"}`, code)
				return
			}
			DatadogTracer.HttpFinishSpan(r.Context(), code, err)
			http.Error(w, string(jsonErr), code)
			return
		}
		DatadogTracer.HttpFinishSpan(r.Context(), code, nil)
	}
	return fn
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

func WithCookie(api *API, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		token := ""
		cookie, err := r.Cookie("xsyn-token")
		if err != nil {
			fmt.Fprintf(w, "cookie not found: %v", err)
			return
		}
		if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
			fmt.Fprintf(w, "decryption error: %v", err)
			return
		}
		_, err = api.TokenLogin(token)
		if err != nil {
			fmt.Fprintf(w, "authentication error: %v", err)
			return
		}
		next(w, r)
	}
	return fn
}

// check passport http request secret
func WithPassportSecret(secret string, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fmt.Println("11111111111111111111111111111111111111111111111111111111111111111111")
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Passport-Authorization") != secret {
			gamelog.L.Warn().Str("header secret", r.Header.Get("Passport-Authorization")).Str("webhook secret", secret).Msg("authentication failed")
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}
		next(w, r)
	}
	return fn
}
