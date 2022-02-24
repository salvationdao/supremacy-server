package passport

import (
	"context"
	"server"
	"time"

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

// SendTickerMessage sends the client map and multipliers to the passport to handle giving out sups
func (pp *Passport) SendTickerMessage(userMap map[int][]server.UserID) {
	pp.send <- &Message{
		Key: "SUPREMACY:TICKER_TICK",
		Payload: struct {
			UserMap map[int][]server.UserID `json:"userMap"`
		}{
			UserMap: userMap,
		},
	}
}

type SpoilOfWarAmountRequest struct {
	Amount string `json:"payload"`
}

// GetSpoilOfWarAmount get current sup pool amount
func (pp *Passport) GetSpoilOfWarAmount(callback func(msg []byte)) {
	pp.send <- &Message{
		Key:           "SUPREMACY:SUPS_POOL_AMOUNT",
		TransactionID: uuid.Must(uuid.NewV4()).String(),
		Callback:      callback,
	}
}

type UserSupsMultiplierSend struct {
	ToUserID        server.UserID     `json:"toUserID"`
	ToUserSessionID *hub.SessionID    `json:"toUserSessionID,omitempty"`
	SupsMultipliers []*SupsMultiplier `json:"supsMultiplier"`
}

type SupsMultiplier struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	ExpiredAt time.Time `json:"expiredAt"`
}

// UserSupsMultiplierSend send user sups multipliers
func (pp *Passport) UserSupsMultiplierSend(ctx context.Context, userSupsMultiplierSends []*UserSupsMultiplierSend) {
	pp.send <- &Message{
		Key: "SUPREMACY:USER_SUPS_MULTIPLIER_SEND",
		Payload: struct {
			UserSupsMultiplierSends []*UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
		}{
			UserSupsMultiplierSends: userSupsMultiplierSends,
		},
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
