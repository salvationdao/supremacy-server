package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (api *API) XSYNAuth(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `token missing`, http.StatusBadRequest)
		return
	}

	resp, err := api.Passport.OneTimeTokenLogin(token, r.UserAgent(), "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

		_, err = player.Update(tx, boil.Whitelist(boiler.PlayerColumns.ID))
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

	err = api.WriteCookie(w, resp.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, api.Config.AuthCallbackURL+"?token=true", http.StatusSeeOther)
}

func (api *API) WriteCookie(w http.ResponseWriter, token string) error {
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
	}
	http.SetCookie(w, cookie)
	return nil
}
