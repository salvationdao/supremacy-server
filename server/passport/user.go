package passport

import (
	"context"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
)

func (pp *Passport) UpgradeUserConnection(sessionID hub.SessionID) {
	pp.send <- &Message{
		Key: "SUPREMACY:USER_CONNECTION_UPGRADE",
		Payload: struct {
			SessionID hub.SessionID `json:"sessionID"`
		}{
			SessionID: sessionID,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
}

type TickerTickReq struct {
	UserMap map[int][]server.UserID `json:"userMap"`
}
type TickerTickResp struct{}

// SendTickerMessage sends the client map and multipliers to the passport to handle giving out sups
func (pp *Passport) SendTickerMessage(userMap map[int][]server.UserID) {
	err := pp.Comms.Call("C.TickerTickHandler", TickerTickReq{userMap}, &TickerTickResp{})
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
	err := pp.Comms.Call("C.SupremacyGetSpoilOfWarHandler", GetSpoilOfWarReq{}, result)
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
	err := pp.Comms.Call("C.UserSupsMultiplierSendHandler", UserSupsMultiplierSendReq{userSupsMultiplierSends}, &UserSupsMultiplierSendResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "UserSupsMultiplierSendHandler").Msg("rpc error")
	}
}

type UserStatSend struct {
	ToUserSessionID *hub.SessionID   `json:"toUserSessionID,omitempty"`
	Stat            *server.UserStat `json:"stat"`
}

// UserStatSend send user sups multipliers
func (pp *Passport) UserStatSend(ctx context.Context, userStatSends []*UserStatSend) {
	pp.send <- &Message{
		Key: "SUPREMACY:USER_STAT_SEND",
		Payload: struct {
			UserStatSends []*UserStatSend `json:"userStatSends"`
		}{
			UserStatSends: userStatSends,
		},
	}
}

type GetUsers struct {
	Users []*server.User `json:"payload"`
}

// UserGet get user by id
func (pp *Passport) UsersGet(userIDs []server.UserID, callback func(msg []byte)) {
	pp.send <- &Message{
		Key:      "SUPREMACY:GET_USERS",
		Callback: callback,
		Payload: struct {
			UserIDs []server.UserID `json:"userIDs"`
		}{
			UserIDs: userIDs,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
}
