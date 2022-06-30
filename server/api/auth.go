package api

import (
	"database/sql"
	"encoding/json"
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
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func AuthRouter(api *API) chi.Router {
	r := chi.NewRouter()
	r.Get("/xsyn", api.XSYNAuth)
	r.Post("/check", WithError(api.AuthCheckHandler))
	r.Get("/logout", WithError(api.LogoutHandler))
	r.Get("/bot_check", WithError(api.AuthBotCheckHandler))

	return r
}

func (api *API) XSYNAuth(w http.ResponseWriter, r *http.Request) {
	isHangar := r.URL.Query().Get("isHangar") != ""

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

	err = api.UpsertPlayer(resp.ID, null.StringFrom(resp.Username), resp.PublicAddress, resp.FactionID, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.WriteCookie(w, r, resp.Token)
	if err != nil {

		return
	}

	callbackUrl := api.Config.AuthCallbackURL
	if isHangar {
		callbackUrl = api.Config.AuthHangarCallbackURL
	}

	http.Redirect(w, r, callbackUrl+"?token=true", http.StatusSeeOther)
}

func (api *API) AuthCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &struct {
		Fingerprint *Fingerprint `json:"fingerprint"`
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

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

		err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, req.Fingerprint)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
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
	
	err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, req.Fingerprint)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
	}

	return helpers.EncodeJSON(w, player)
}

func (api *API) AuthBotCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no token are provided"), "Player are not signed in.")
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

	err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, nil)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
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

type Fingerprint struct {
	VisitorID  string  `json:"visitor_id"`
	OSCPU      string  `json:"os_cpu"`
	Platform   string  `json:"platform"`
	Timezone   string  `json:"timezone"`
	Confidence float32 `json:"confidence"`
	UserAgent  string  `json:"user_agent"`
}

func FingerprintUpsert(fingerprint Fingerprint, playerID string) error {
	// Attempt to find fingerprint or create one
	fingerprintExists, err := boiler.Fingerprints(boiler.FingerprintWhere.VisitorID.EQ(fingerprint.VisitorID)).Exists(gamedb.StdConn)
	if err != nil {
		return err
	}

	if !fingerprintExists {
		fp := boiler.Fingerprint{
			VisitorID:  fingerprint.VisitorID,
			OsCPU:      null.StringFrom(fingerprint.OSCPU),
			Platform:   null.StringFrom(fingerprint.Platform),
			Timezone:   null.StringFrom(fingerprint.Timezone),
			Confidence: decimal.NewNullDecimal(decimal.NewFromFloat32(fingerprint.Confidence)),
			UserAgent:  null.StringFrom(fingerprint.UserAgent),
		}
		err = fp.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}

	f, err := boiler.Fingerprints(boiler.FingerprintWhere.VisitorID.EQ(fingerprint.VisitorID)).One(gamedb.StdConn)
	if err != nil {
		return err
	}

	// Link fingerprint to user
	playerFingerprintExists, err := boiler.PlayerFingerprints(boiler.PlayerFingerprintWhere.PlayerID.EQ(playerID), boiler.PlayerFingerprintWhere.FingerprintID.EQ(f.ID)).Exists(gamedb.StdConn)
	if err != nil {
		return err
	}
	if !playerFingerprintExists {
		// User fingerprint does not exist; create one
		newPlayerFingerprint := boiler.PlayerFingerprint{
			PlayerID:      playerID,
			FingerprintID: f.ID,
		}
		err = newPlayerFingerprint.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

func (api *API) UpsertPlayer(playerID string, username null.String, publicAddress null.String, factionID null.String, fingerprint *Fingerprint) error {
	player, err := boiler.FindPlayer(gamedb.StdConn, playerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to retrieve player.")
	}

	playStat, err := boiler.FindPlayerStat(gamedb.StdConn, playerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to retrieve player stat.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to update player.")
	}

	defer tx.Rollback()

	// insert player if not exists
	if player == nil {
		player = &boiler.Player{
			ID:            playerID,
			Username:      username,
			PublicAddress: publicAddress,
			FactionID:     factionID,
		}

		err = player.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Str("player id", playerID).Msg("player insert")
			return terror.Error(err, "Failed to insert player.")
		}
	} else {
		// update user profile
		player.Username = username
		player.PublicAddress = publicAddress
		player.FactionID = factionID

		_, err = player.Update(tx, boil.Whitelist(
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.PublicAddress,
			boiler.PlayerColumns.FactionID,
		))
		if err != nil {
			gamelog.L.Error().Str("player id", playerID).Err(err).Msg("player update")
			return terror.Error(err, "Failed to update player detail.")
		}
	}

	// fingerprint
	if fingerprint != nil {
		err = FingerprintUpsert(*fingerprint, playerID)
		if err != nil {
			gamelog.L.Error().Str("player id", playerID).Err(err).Msg("player finger print upsert")
			return terror.Error(err, "browser identification fail.")
		}
	}

	if playStat == nil {
		// check player stat
		playStat = &boiler.PlayerStat{
			ID:                    playerID,
			ViewBattleCount:       0,
			AbilityKillCount:      0,
			TotalAbilityTriggered: 0,
			MechKillCount:         0,
		}

		err = playStat.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("player id", playerID).Err(err).Msg("player stat insert")
			return terror.Error(err, "Failed to insert player stat.")
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to update player")
		return terror.Error(err, "Failed to update player")
	}

	return nil
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

	cookie = &http.Cookie{
		Name:     "xsyn-token",
		Value:    "",
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
		Secure:   api.IsCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)

	return nil
}

func domain(host string) string {
	parts := strings.Split(host, ".")
	//this is rigid as fuck
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}
