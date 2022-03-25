package api

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

// AuthControllerWS holds handlers for checking server status
type AuthControllerWS struct {
	Conn   *pgxpool.Pool
	Log    *zerolog.Logger
	API    *API
	Config *server.Config
}

// NewAuthController creates the check hub
func NewAuthController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, config *server.Config) *AuthControllerWS {
	authHub := &AuthControllerWS{
		Conn:   conn,
		Log:    log_helpers.NamedLogger(log, "auth_hub"),
		API:    api,
		Config: config,
	}

	api.Command(HubKeyAuthSessionIDGet, authHub.GetHubSessionID)
	api.Command(HubKeyAuthJWTCheck, authHub.RingCheckJWTAuth)

	return authHub
}

const HubKeyAuthSessionIDGet = hub.HubCommandKey("AUTH:SESSION:ID:GET")

// GetHubSessionID return hub client's session id for ring check authentication
func (ac *AuthControllerWS) GetHubSessionID(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	ac.API.RingCheckAuthMap.Record(string(wsc.SessionID), wsc)
	reply(wsc.SessionID)
	return nil
}

type RingCheckJWTAuthRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Token string `json:"token"`
	} `json:"payload"`
}

const HubKeyAuthJWTCheck = hub.HubCommandKey("AUTH:JWT:CHECK")

// RingCheckJWTAuth auths a user using JWT provided by Passport Server
func (ac *AuthControllerWS) RingCheckJWTAuth(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RingCheckJWTAuthRequest{}
	err := json.Unmarshal(payload, &req)
	if err != nil {
		return terror.Error(err, "Unable to unmarshal json struct")
	}

	tokenStr, err := base64.StdEncoding.DecodeString(req.Payload.Token)
	if err != nil {
		return terror.Error(err, "Failed to convert string to byte array")
	}

	token, err := readJWT(tokenStr, ac.Config.EncryptTokens, []byte(ac.Config.EncryptTokensKey), ac.Config.JwtKey)
	if err != nil {
		gamelog.L.Err(err).Str("reading-jwt", req.Payload.Token).Msg("Failed to read JWT token")
		return terror.Error(err, "Failed to read JWT token please try again")
	}

	// get user id from token
	player := &boiler.Player{}
	playerID, ok := token.Get("user-id")
	if !ok {
		return terror.Error(fmt.Errorf("unable to get playerid from token"), "Unable to get playerID from token")
	}

	player.ID, ok = playerID.(string)
	if !ok {
		return terror.Error(fmt.Errorf("unable to form UUID from token"), "Unable to get playerID from token")
	}

	userID := server.UserID(uuid.FromStringOrNil(player.ID))

	// get user from passport server
	user, err := ac.API.Passport.UserGet(userID)
	if err != nil {
		return terror.Error(err, "Failed to get user from passport server")
	}

	player.PublicAddress = user.PublicAddress
	player.Username = null.StringFrom(user.Username)

	if !user.FactionID.IsNil() {
		player.FactionID = null.StringFrom(user.FactionID.String())
	}

	// store user into player table
	err = db.UpsertPlayer(player)
	if err != nil {
		gamelog.L.Error().Interface("player", player).Err(err).Msg("Failed to upsert player")
		return terror.Error(err, "Failed to add user to database. Please try again")
	}

	if player.FactionID.Valid {
		user.FactionID = server.FactionID(uuid.FromStringOrNil(player.FactionID.String))
		faction, err := boiler.FindFaction(gamedb.StdConn, player.FactionID.String)
		if err != nil {
			return terror.Error(err, "Issues finding faction, try again or contact support.")
		}

		err = user.Faction.SetFromBoilerFaction(faction)
		if err != nil {
			return terror.Error(err, "Issues finding faction, try again or contact support.")
		}
	}
	b, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyUserRingCheck,
		Payload: user,
	})
	if err != nil {
		return terror.Error(err, "Failed to marshal JSON")
	}

	wsc.SetIdentifier(player.ID)

	go wsc.Send(b)

	return nil
}

// ReadJWT grabs the user from the token
func readJWT(tokenB []byte, decryptToken bool, decryptKey, jwtKey []byte) (jwt.Token, error) {
	if !decryptToken {
		token, err := jwt.Parse(tokenB, jwt.WithVerify(jwa.HS256, jwtKey))
		if err != nil {
			return nil, terror.Error(err, "Token verification failed")
		}
		if token.Expiration().Before(time.Now()) {
			return token, terror.Error(fmt.Errorf("token expired"), "Token expired please try login in again")
		}

		return token, nil
	}

	decrpytedToken, err := decrypt(decryptKey, tokenB)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to decrypt token")
		return nil, terror.Error(err, "Error decrypting JWT token")
	}

	return readJWT(decrpytedToken, false, nil, jwtKey)
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to decrypt token")
		return nil, terror.Error(err, "Failed to decrypt token")
	}
	if len(text) < aes.BlockSize {
		return nil, terror.Error(fmt.Errorf("ciphertext too short"), "Ciphertext too short, please try login in again")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to decrypt token")
		return nil, terror.Error(err, "Failed to decrypt token")
	}
	return data, nil
}
