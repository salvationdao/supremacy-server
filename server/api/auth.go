package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strings"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func AuthRouter(api *API) chi.Router {
	r := chi.NewRouter()
	r.Get("/xsyn", api.XSYNAuth)
	r.Get("/check", WithError(api.AuthCheckHandler))
	r.Get("/logout", WithError(api.LogoutHandler))

	return r
}

func (api *API) XSYNAuth(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `token missing`, http.StatusBadRequest)
		return
	}

	resp, err := api.Passport.OneTimeTokenLogin(token, r.UserAgent(), "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// player upsert
	player, err := boiler.FindPlayer(gamedb.StdConn, resp.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	// insert player if not exists
	if player == nil {
		player := boiler.Player{
			ID:            resp.ID,
			Username:      null.StringFrom(resp.Username),
			PublicAddress: resp.PublicAddress,
			FactionID:     resp.FactionID,
		}

		err = player.Insert(tx, boil.Infer())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// insert user state
		userStat := boiler.PlayerStat{
			ID:                    resp.ID,
			ViewBattleCount:       0,
			AbilityKillCount:      0,
			TotalAbilityTriggered: 0,
			MechKillCount:         0,
		}

		err = userStat.Insert(tx, boil.Infer())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// update user profile
		player.Username = null.StringFrom(resp.Username)
		player.PublicAddress = resp.PublicAddress
		player.FactionID = resp.FactionID

		_, err = player.Update(tx, boil.Whitelist(
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.PublicAddress,
			boiler.PlayerColumns.FactionID,
		))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	err = api.WriteCookie(w, r, resp.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, api.Config.AuthCallbackURL+"?token=true", http.StatusSeeOther)
}

func (api *API) AuthCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		// check whether token is attached
		gamelog.L.Debug().Msg("Cookie not found")

		token := r.URL.Query().Get("token")
		if token == "" {
			return http.StatusBadRequest, terror.Warn(fmt.Errorf("no cookie and token are provided"), "Player are not signed in.")
		}

		// check user from token
		player, err := api.TokenLogin(token)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
		}

		// write cookie
		err = api.WriteCookie(w, r, token)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to write cookie")
		}

		return helpers.EncodeJSON(w, player)
	}

	var token string
	if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to decrypt token")
	}

	// check user from token
	player, err := api.TokenLogin(token)

	if err != nil {
		if errors.Is(err, errors.New("session is expired")) {
			err := api.DeleteCookie(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusBadRequest, terror.Error(err, "Session is expired")
		}
		return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
	}

	return helpers.EncodeJSON(w, player)
}

func (api *API) LogoutHandler(w http.ResponseWriter, r *http.Request) (int, error) {

	_, err := r.Cookie("xsyn-token")
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Player is not login")
	}

	err = api.DeleteCookie(w, r)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to delete cookie")
	}

	return http.StatusOK, nil
}

func (api *API) WriteCookie(w http.ResponseWriter, r *http.Request, token string) error {
	b64, err := api.Cookie.EncryptToBase64(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Encryption error: %v", err), http.StatusBadRequest)
		return err
	}

	cookie := &http.Cookie{
		Name:     "xsyn-token",
		Value:    b64,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		Secure:   api.IsCookieSecure,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Domain:   domain(r.Host),
	}
	http.SetCookie(w, cookie)
	return nil
}

func (api *API) DeleteCookie(w http.ResponseWriter, r *http.Request) error {
	cookie := &http.Cookie{
		Name:     "xsyn-token",
		Value:    "",
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
		Secure:   api.IsCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Domain:   domain(r.Host),
	}
	http.SetCookie(w, cookie)
	return nil
}

func domain(host string) string {
	parts := strings.Split(host, ".")
	//this is rigid as fuck
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}
