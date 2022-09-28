package xsyn_rpcclient

import (
	"database/sql"
	"server"
	"server/gamelog"
	"strings"
	"time"

	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

type UserGetReq struct {
	ApiKey string        `json:"apiKey"`
	UserID server.UserID `json:"userID"`
}

type UserGetResp struct {
	User *server.User `json:"user"`
}

type UserReq struct {
	ID uuid.UUID
}

type UserResp struct {
	ID            string
	Username      string
	FactionID     null.String
	PublicAddress null.String
}

// UserGet get user by id
func (pp *XsynXrpcClient) UserGet(userID server.UserID) (*UserResp, error) {
	resp := &UserResp{}
	err := pp.XrpcClient.Call("S.UserGetHandler", UserGetReq{pp.ApiKey, userID}, resp)

	if err != nil {
		gamelog.L.Err(err).Str("method", "UserGetHandler").Msg("rpc error")
		return nil, terror.Error(err, "Failed to get user from passport server")
	}
	return resp, nil
}

type TokenResp struct {
	*UserResp
	Token     string
	ExpiredAt time.Time
}

func (pp *XsynXrpcClient) OneTimeTokenLogin(tokenBase64, device, action string) (*TokenResp, error) {
	resp := &TokenResp{}
	err := pp.XrpcClient.Call("S.OneTimeTokenLogin", OneTimeTokenReq{pp.ApiKey, tokenBase64, device, action}, resp)

	if err != nil {
		gamelog.L.Err(err).Str("method", "OneTimeTokenLogin").Msg("rpc error")
		return nil, terror.Error(err, "Failed to get user from passport server")
	}
	return resp, nil
}

type OneTimeTokenReq struct {
	ApiKey      string
	TokenBase64 string
	Device      string
	Action      string
}

type TokenReq struct {
	ApiKey      string
	TokenBase64 string
}

func (pp *XsynXrpcClient) TokenLogin(tokenBase64 string) (*UserResp, error) {
	resp := &UserResp{}
	err := pp.XrpcClient.Call("S.TokenLogin", TokenReq{pp.ApiKey, tokenBase64}, resp)

	if err != nil && !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		gamelog.L.Error().Err(err).Str("method", "TokenLogin").Msg("rpc error")
		return nil, terror.Error(err, "Failed to get user from passport server")
	}
	return resp, nil
}

type UserBalanceGetReq struct {
	ApiKey string    `json:"apiKey"`
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}

// UserBalanceGet return the sups balance from the given user id
func (pp *XsynXrpcClient) UserBalanceGet(userID uuid.UUID) decimal.Decimal {
	resp := &UserBalanceGetResp{}
	err := pp.XrpcClient.Call("S.UserBalanceGetHandler", UserBalanceGetReq{pp.ApiKey, userID}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "UserBalanceGetHandler").Msg("rpc error")
		return decimal.Zero
	}

	return resp.Balance
}

type UsernameUpdateReq struct {
	UserID      string `json:"user_id"`
	NewUsername string `json:"new_username"`
	ApiKey      string
}

type UsernameUpdateResp struct {
	Username string
}

// UserUpdateUsername updates username
func (pp *XsynXrpcClient) UserUpdateUsername(userID string, newUsername string) error {
	resp := &UsernameUpdateResp{}
	err := pp.XrpcClient.Call("S.UserUpdateUsername", UsernameUpdateReq{
		ApiKey:      pp.ApiKey,
		UserID:      userID,
		NewUsername: newUsername},
		resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "UserUpdateUsername").Msg("rpc error")
		return err
	}
	return nil
}

type UserFactionEnlistReq struct {
	ApiKey    string
	UserID    string `json:"userID"`
	FactionID string `json:"factionID"`
}

type UserFactionEnlistResp struct{}

// UserFactionEnlist update user faction
func (pp *XsynXrpcClient) UserFactionEnlist(userID string, factionID string) error {
	resp := &UserFactionEnlistResp{}
	err := pp.XrpcClient.Call("S.UserFactionEnlistHandler", UserFactionEnlistReq{pp.ApiKey, userID, factionID}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "UserFactionEnlistHandler").Msg("rpc error")
		return err
	}

	return nil
}

type GenOneTimeTokenReq struct {
	ApiKey string
	UserID string
}

type GenOneTimeTokenResp struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}

// GenOneTimeToken generates a token to create a QR code to log a player into the companion app
func (pp *XsynXrpcClient) GenOneTimeToken(userID string) (*GenOneTimeTokenResp, error) {
	l := gamelog.L.With().Str("func", "GenOneTimeToken").Str("userID", userID).Logger()

	resp := &GenOneTimeTokenResp{}
	err := pp.XrpcClient.Call("S.GenOneTimeToken", GenOneTimeTokenReq{pp.ApiKey, userID}, resp)
	if err != nil {
		l.Error().Err(err).Msg("rpc error")
		return nil, terror.Error(err, "Failed to get user from passport server")
	}
	return resp, nil
}
