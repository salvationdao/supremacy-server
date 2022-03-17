package passport

import (
	"context"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/shopspring/decimal"
)

type TickerTickReq struct {
	UserMap map[int][]server.UserID `json:"userMap"`
}
type TickerTickResp struct{}

// SendTickerMessage sends the client map and multipliers to the passport to handle giving out sups
func (pp *Passport) SendTickerMessage(userMap map[int][]server.UserID) {
	err := pp.RPCClient.Call("S.TickerTickHandler", TickerTickReq{userMap}, &TickerTickResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "TickerTickHandler").Msg("rpc error")
	}
}

type GetSpoilOfWarReq struct{}
type GetSpoilOfWarResp struct {
	Amount string
}

// GetSpoilOfWarAmount get current sup pool amount
func (pp *Passport) GetSpoilOfWarAmount() string {
	result := &GetSpoilOfWarResp{}
	err := pp.RPCClient.Call("S.SupremacyGetSpoilOfWarHandler", GetSpoilOfWarReq{}, result)
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyGetSpoilOfWarHandler").Msg("rpc error")
	}
	return result.Amount
}

type UserSupsMultiplierSendReq struct {
	UserSupsMultiplierSends []*server.UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
}

type UserSupsMultiplierSendResp struct{}

// UserSupsMultiplierSend send user sups multipliers
func (pp *Passport) UserSupsMultiplierSend(ctx context.Context, userSupsMultiplierSends []*server.UserSupsMultiplierSend) {
	err := pp.RPCClient.Call("S.UserSupsMultiplierSendHandler", UserSupsMultiplierSendReq{userSupsMultiplierSends}, &UserSupsMultiplierSendResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "UserSupsMultiplierSendHandler").Msg("rpc error")
	}
}

type UsersGetReq struct {
	UserIDs []server.UserID `json:"userIDs"`
}

type UsersGetResp struct {
	Users []*server.User `json:"users"`
}

// UserGet get user by id
func (pp *Passport) UsersGet(userIDs []server.UserID, callback func(users []*server.User)) {
	resp := &UsersGetResp{}
	err := pp.RPCClient.Call("S.SupremacyUsersGetHandler", UsersGetReq{userIDs}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyUsersGetHandler").Msg("rpc error")
		return
	}
	callback(resp.Users)
}

type UserStatSendReq struct {
	UserStatSends []*UserStatSend `json:"userStatSends"`
}

type UserStatSend struct {
	ToUserSessionID *hub.SessionID   `json:"toUserSessionID,omitempty"`
	Stat            *server.UserStat `json:"stat"`
}

type UserStatSendResp struct{}

// UserStatSend send user sups multipliers
func (pp *Passport) UserStatSend(ctx context.Context, userStatSends []*UserStatSend) {
	if len(userStatSends) == 0 {
		return
	}
	err := pp.RPCClient.Call("S.SupremacyUserStatSendHandler", UserStatSendReq{userStatSends}, &UserStatSend{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyUserStatSendHandler").Msg("rpc error")
	}
}

type UserBalanceGetReq struct {
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}

// UserBalanceGet return the sups balance from the given user id
func (pp *Passport) UserBalanceGet(userID uuid.UUID) decimal.Decimal {
	resp := &UserBalanceGetResp{}
	err := pp.RPCClient.Call("S.SupremacyUserBalanceGetHandler", UserBalanceGetReq{userID}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyUserBalanceGetHandler").Msg("rpc error")
		return decimal.Zero
	}

	return resp.Balance
}
