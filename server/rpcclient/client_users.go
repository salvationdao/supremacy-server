package rpcclient

import (
	"server"
	"server/gamelog"

	"github.com/ethereum/go-ethereum/common"
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
	ID            uuid.UUID
	Username      string
	FactionID     null.String
	PublicAddress common.Address
}

// UserGet get user by id
func (pp *PassportXrpcClient) UserGet(userID server.UserID) (*server.User, error) {
	resp := &UserGetResp{}
	err := pp.XrpcClient.Call("S.UserGetHandler", UserGetReq{pp.ApiKey, userID}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "UserGetHandler").Msg("rpc error")
		return nil, terror.Error(err, "Failed to get user from passport server")
	}
	return resp.User, nil
}

type UserBalanceGetReq struct {
	ApiKey string    `json:"apiKey"`
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}

// UserBalanceGet return the sups balance from the given user id
func (pp *PassportXrpcClient) UserBalanceGet(userID uuid.UUID) decimal.Decimal {
	resp := &UserBalanceGetResp{}
	err := pp.XrpcClient.Call("S.UserBalanceGetHandler", UserBalanceGetReq{pp.ApiKey, userID}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "UserBalanceGetHandler").Msg("rpc error")
		return decimal.Zero
	}

	return resp.Balance
}
