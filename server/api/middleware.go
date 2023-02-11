package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"server"
	"server/db"
	"server/gamelog"

	"github.com/go-chi/chi/v5"
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
			gamelog.L.Warn().Str("err", terror.Echo(err, false)).Msg("handler error")
			errObj := &ErrorObject{Message: err.Error()}
			jsonErr, wErr := json.Marshal(errObj)
			if wErr != nil {
				DatadogTracer.HttpFinishSpan(r.Context(), code, wErr)
				gamelog.L.Warn().Str("err", terror.Echo(err, false)).Msg("failed to marshal error middleware")
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

func (api *API) AuthWS(userIDMustMatch bool) func(next http.Handler) http.Handler {
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
						gamelog.L.Debug().Err(err).Msg("missing token and cookie")
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					gamelog.L.Debug().Err(err).Msg("decrypting cookie error")
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				gamelog.L.Debug().Err(err).Msg("authentication error")
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

			ctx := context.WithValue(r.Context(), "auth_user_id", user.ID)
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		return http.HandlerFunc(fn)
	}
}

func (api *API) AuthUserFactionWS(factionIDMustMatch bool) func(next http.Handler) http.Handler {
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
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					gamelog.L.Error().Err(err).Msg("decrypting cookie error")
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				fmt.Fprintf(w, "authentication error: %v", err)
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
				fmt.Fprintf(w, "invalid ip address")
				return
			}

			if !user.FactionID.Valid {
				fmt.Fprintf(w, "authentication error: user has not enlisted in one of the factions")
				return
			}

			if factionIDMustMatch {
				factionID := chi.URLParam(r, "faction_id")
				if factionID == "" || factionID != user.FactionID.String {
					fmt.Fprintf(w, "faction id check failed... url faction id: %s, user faction id: %s, url:%s", factionID, user.FactionID.String, r.URL.Path)
					return
				}
			}

			ctxWithUserID := context.WithValue(r.Context(), "auth_user_id", user.ID)
			ctx := context.WithValue(ctxWithUserID, "faction_id", user.FactionID.String)
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
