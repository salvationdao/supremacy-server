package passport

import (
	"context"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
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

type UserGetReq struct {
	UserID server.UserID `json:"userID"`
}

type UserGetResp struct {
	User *server.User `json:"user"`
}

// UserGet get user by id
func (pp *Passport) UserGet(userID server.UserID) (*server.User, error) {
	resp := &UserGetResp{}
	err := pp.RPCClient.Call("S.SupremacyUserGetHandler", UserGetReq{userID}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyUserGetHandler").Msg("rpc error")
		return nil, terror.Error(err, "Failed to get user from passport server")
	}
	return resp.User, nil
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
