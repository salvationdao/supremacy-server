package passport

import (
	"context"
	"encoding/json"
	"server"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
)

func (pp *Passport) UpgradeUserConnection(ctx context.Context, sessionID hub.SessionID, txID string) error {
	replyChannel := make(chan []byte)
	errChan := make(chan error)
	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:USER_CONNECTION_UPGRADE",
			Payload: struct {
				SessionID hub.SessionID `json:"sessionID"`
			}{
				SessionID: sessionID,
			},
			TransactionID: txID,
			context:       ctx,
		},
	}
	for {
		select {
		case <-replyChannel:
			return nil
		case err := <-errChan:
			return terror.Error(err)
		}
	}
}

// SendTickerMessage sends the client map and multipliers to the passport to handle giving out sups
func (pp *Passport) SendTickerMessage(ctx context.Context, userMap map[int][]server.UserID) {
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:TICKER_TICK",
			Payload: struct {
				UserMap map[int][]server.UserID `json:"userMap"`
			}{
				UserMap: userMap,
			},
			context: ctx,
		},
	}
}

type SpoilOfWarAmountRequest struct {
	Amount string `json:"payload"`
}

// GetSpoilOfWarAmount get current sup pool amount
func (pp *Passport) GetSpoilOfWarAmount(ctx context.Context) (string, error) {
	txID, err := uuid.NewV4()
	if err != nil {
		return "", terror.Error(err)
	}
	replyChannel := make(chan []byte)
	errChan := make(chan error)
	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key:           "SUPREMACY:SUPS_POOL_AMOUNT",
			context:       ctx,
			TransactionID: txID.String(),
		},
	}

	for {
		select {
		case msg := <-replyChannel:
			resp := &SpoilOfWarAmountRequest{}
			err = json.Unmarshal(msg, resp)
			if err != nil {
				return "", terror.Error(err)
			}
			return resp.Amount, nil
		case err := <-errChan:
			return "", terror.Error(err)
		}
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
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:USER_SUPS_MULTIPLIER_SEND",
			Payload: struct {
				UserSupsMultiplierSends []*UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
			}{
				UserSupsMultiplierSends: userSupsMultiplierSends,
			},
			context: ctx,
		},
	}
}
