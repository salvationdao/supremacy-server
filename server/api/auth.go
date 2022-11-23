package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strings"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ethereum/go-ethereum/common"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func AuthRouter(api *API) chi.Router {
	r := chi.NewRouter()
	r.Post("/xsyn", WithError((api.XSYNAuth)))
	r.Post("/check", WithError(api.AuthCheckHandler))
	r.Get("/logout", WithError(api.LogoutHandler))
	r.Get("/bot_check", WithError(api.AuthBotCheckHandler))

	r.Post("/companion_app_token_login", WithError(api.AuthAppTokenLoginHandler))
	r.Post("/qr_code_login", WithError(api.AuthQRCodeLoginHandler))

	return r
}

func (api *API) XSYNAuth(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &struct {
		IssueToken  string       `json:"issue_token"`
		Fingerprint *Fingerprint `json:"fingerprint"`
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if req.IssueToken == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing issue token"), "Missing issue token.")
	}

	player, err := api.TokenLogin(req.IssueToken)
	if err != nil {
		gamelog.L.Warn().Msg("No token found")
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no token are provided"), "User are not signed in.")
	}

	err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, req.Fingerprint, player.AcceptsMarketing)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
	}

	err = api.WriteCookie(w, r, req.IssueToken)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to write cookie")
	}
	return helpers.EncodeJSON(w, player)
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
			return http.StatusBadRequest, terror.Error(fmt.Errorf("no cookie and token are provided"), "Player are not signed in.")
		}

		// check user from token
		player, err := api.TokenLogin(token)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
		}

		err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, req.Fingerprint, player.AcceptsMarketing)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
		}

		// write cookie
		err = api.WriteCookie(w, r, token)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to write cookie")
		}

		user, err := boiler.Players(
			boiler.PlayerWhere.ID.EQ(player.ID),
			qm.Load(boiler.PlayerRels.Role),
		).One(gamedb.StdConn)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to find player.")
		}

		if user.R != nil && user.R.Role != nil {
			player.RoleType = user.R.Role.RoleType
		}

		return helpers.EncodeJSON(w, player)
	}

	var token string
	err = api.Cookie.DecryptBase64(cookie.Value, &token)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to decrypt token")
	}

	// check user from token
	player, err := api.TokenLogin(token)
	if err != nil {
		if errors.Is(err, errors.New("session is expired")) {
			api.DeleteCookie(w, r)
			return http.StatusBadRequest, terror.Error(err, "Session is expired")
		}
		return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
	}

	err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, req.Fingerprint, player.AcceptsMarketing)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
	}

	user, err := boiler.Players(
		boiler.PlayerWhere.ID.EQ(player.ID),
		qm.Load(boiler.PlayerRels.Role),
	).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find player.")
	}

	if user.R != nil && user.R.Role != nil {
		player.RoleType = user.R.Role.RoleType
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
			api.DeleteCookie(w, r)
			return http.StatusBadRequest, terror.Error(err, "Session is expired")
		}
		return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
	}

	err = api.UpsertPlayer(player.ID, player.Username, player.PublicAddress, player.FactionID, nil, null.Bool{})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update player.")
	}

	return helpers.EncodeJSON(w, player)
}

// AuthAppTokenLoginHandler logs a player into the companion app
func (api *API) AuthAppTokenLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	// Get token
	token := r.Header.Get("token")
	if token == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no token provided"), "Player is not signed in.")
	}

	// Check user from token
	player, err := api.TokenLogin(token)
	if err != nil {
		if errors.Is(err, errors.New("Session is expired")) {
			return http.StatusBadRequest, terror.Error(err, "Session is expired.")
		}
		return http.StatusBadRequest, terror.Error(err, "Failed to authenticate player.")
	}

	return helpers.EncodeJSON(w, player)
}

// AuthQRCodeLoginHandler is used to log a player into the companion app
// - uses the one time token displayed the QR code
func (api *API) AuthQRCodeLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	l := gamelog.L.With().Str("func", "AuthQRCodeLoginHandler").Logger()
	// Get token
	token := r.Header.Get("token")
	if token == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no token provided"), "Failed to authenticate device.")
	}

	// Get device info
	deviceID := r.Header.Get("id")
	if deviceID == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no device ID provided"), "Failed to authenticate device.")
	}
	deviceName := r.Header.Get("name")
	if deviceName == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no device name provided"), "Failed to authenticate device.")
	}

	// Get user token from passport
	tokenResp, err := api.Passport.OneTimeTokenLogin(token, r.UserAgent(), "login")
	if err != nil {
		l.Error().Err(err).Msg("Failed to get token from passport.")
		return http.StatusBadRequest, terror.Error(err, "Unable to retrieve player, try again or contact support.")
	}

	// Get user with token
	user, err := api.TokenLogin(tokenResp.Token)
	if err != nil {
		if errors.Is(err, errors.New("Session is expired")) {
			return http.StatusBadRequest, terror.Error(err, "Session is expired")
		}
		return http.StatusBadRequest, terror.Error(err, "Unable to authenticate player, try again or contact support")
	}
	l = l.With().Str("user id", user.ID).Logger()

	// Check if device exists (and is not deleted)
	exists, err := boiler.Devices(
		boiler.DeviceWhere.DeviceID.EQ(deviceID),
		boiler.DeviceWhere.PlayerID.EQ(user.ID),
		boiler.DeviceWhere.DeletedAt.IsNull(),
	).Exists(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to get player devices")
		return http.StatusBadRequest, terror.Error(err, "Failed to authenticate device.")
	}

	// Add device to table
	if !exists {
		d := boiler.Device{
			DeviceID: deviceID,
			PlayerID: user.ID,
			Name:     deviceName,
		}
		err = d.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("Failed to insert user device.")
			return http.StatusInternalServerError, terror.Error(err, "Failed to connect device, try again or contact support.")
		}
	}

	// Write auth token to header - this is saved on the device
	w.Header().Set("xsyn-token", tokenResp.Token)

	return helpers.EncodeJSON(w, user)
}

func (api *API) LogoutHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Player is not logged in")
	}

	var token string
	err = api.Cookie.DecryptBase64(cookie.Value, &token)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to decrypt token")
	}

	resp, err := api.Passport.TokenLogout(token)
	if err != nil || !resp.LogoutSuccess {
		gamelog.L.Warn().Msg("No token found")
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no token are provided"), "User was not signed in.")
	}

	api.DeleteCookie(w, r)

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

func (api *API) UpsertPlayer(playerID string, username null.String, publicAddress null.String, factionID null.String, fingerprint *Fingerprint, acceptsMarketing null.Bool) error {
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

	if publicAddress.Valid {
		publicAddress = null.StringFrom(common.HexToAddress(publicAddress.String).Hex())
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

		if acceptsMarketing.Valid {
			player.AcceptsMarketing = acceptsMarketing
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
		player.AcceptsMarketing = acceptsMarketing

		if acceptsMarketing.Valid {
			player.AcceptsMarketing = acceptsMarketing
		}

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

	if api.Config.Environment == "development" {
		features, err := db.GetAllFeatures()
		if err != nil {
			return terror.Error(err, "Failed get features.")
		}

		playerFeatures, err := db.GetPlayerFeaturesByID(player.ID)
		if err != nil {
			return terror.Error(err, "Failed get features for user.")
		}

		if len(playerFeatures) != len(features) {
			for _, feature := range features {
				err := db.AddFeatureToPlayerIDs(feature.Name, []string{playerID})
				if err != nil {
					return terror.Error(err, "Failed get add feature to user.")
				}
			}
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
		Expires:  time.Now().AddDate(0, 0, 1), // sync with token expiry
		Secure:   api.IsCookieSecure,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Domain:   domain(r.Host),
	}
	http.SetCookie(w, cookie)
	return nil
}

func (api *API) DeleteCookie(w http.ResponseWriter, r *http.Request) {
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
}

func domain(host string) string {
	parts := strings.Split(host, ".")
	//this is rigid as fuck
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

// TokenLogin gets a user from the token
func (api *API) TokenLogin(tokenBase64 string, ignoreErr ...bool) (*server.Player, error) {
	ignoreError := len(ignoreErr) > 0 && ignoreErr[0] == true

	userResp, err := api.Passport.TokenLogin(tokenBase64)
	if err != nil {
		if !ignoreError {
			if err.Error() != "session is expired" && err.Error() != "sql: no rows in result set" {
				gamelog.L.Error().Err(err).Msg("Failed to login with token")
			}
			gamelog.L.Debug().Err(err).Msg("Failed to login with token")
		}
		return nil, err
	}

	err = api.UpsertPlayer(userResp.ID, null.StringFrom(userResp.Username), userResp.PublicAddress, userResp.FactionID, nil, userResp.AcceptsMarketing)
	if err != nil {
		if !ignoreError {
			gamelog.L.Error().Err(err).Msg("Failed to update player detail")
		}
		return nil, err
	}

	serverPlayer, err := db.GetPlayer(userResp.ID)
	if err != nil {
		if !ignoreError {
			gamelog.L.Error().Err(err).Msg("Failed to get player by ID")
		}
		return nil, err
	}
	serverPlayer.AccountID = userResp.AccountID // this we don't store on supremacy server

	return serverPlayer, nil
}
