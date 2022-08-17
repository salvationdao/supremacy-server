package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"net"
	"net/http"
	"server"
	"server/db"
	"server/gamelog"
	"strings"

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

func WithCookie(api *API, next func(user *server.Player, w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		token := ""
		cookie, err := r.Cookie("xsyn-token")
		if err != nil {
			return http.StatusForbidden, fmt.Errorf("cookie not found")
		}
		if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
			return http.StatusForbidden, fmt.Errorf("invalid cookie")
		}
		user, err := api.TokenLogin(token)
		if err != nil {
			return http.StatusForbidden, fmt.Errorf("authentication failed")
		}
		return next(user, w, r)
	}
	return fn
}

// WithPassportSecret check passport http request secret
func WithPassportSecret(secret string, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
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


func (api *API) AuthWS(required bool, userIDMustMatch bool, onlyAuthPaths ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string
			var ok bool

			cookie, err := r.Cookie("xsyn-token")
			if err != nil {
				token = r.URL.Query().Get("token")
				if token == "" {
					token, ok = r.Context().Value("token").(string)
					if !ok || token == "" {
						if required {
							gamelog.L.Debug().Err(err).Msg("missing token and cookie")
							http.Error(w, "Unauthorized", http.StatusUnauthorized)
							return
						}
						next.ServeHTTP(w, r)
						return
					}
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					if required {
						gamelog.L.Debug().Err(err).Msg("decrypting cookie error")
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			if !required {
				path := r.URL.Path

				exists := false
				for _, p := range onlyAuthPaths {
					if strings.Contains(path, p) {
						exists = true
						break
					}
				}

				if !exists {
					next.ServeHTTP(w, r)
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				gamelog.L.Debug().Err(err).Msg("authentication error")
				return
			}

			// get ip
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ipaddr, _, _ := net.SplitHostPort(r.RemoteAddr)
				userIP := net.ParseIP(ipaddr)
				if userIP == nil {
					ip = ipaddr
				} else {
					ip = userIP.String()
				}
			}

			// upsert player ip logs
			err = db.PlayerIPUpsert(user.ID, ip)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to log player ip")
				return
			}

			if userIDMustMatch {
				userID := chi.URLParam(r, "user_id")
				if userID == "" || userID != user.ID {
					gamelog.L.Debug().Err(fmt.Errorf("user id check failed")).
						Str("userID", userID).
						Str("user.ID", user.ID).
						Str("r.URL.Path", r.URL.Path).
						Msg("user id check failed")
					return
				}
			}

			ctx := context.WithValue(r.Context(), "user_id", user.ID)
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		return http.HandlerFunc(fn)
	}
}
