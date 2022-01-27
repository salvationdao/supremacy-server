package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"server"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
)

func (pp *Passport) UpgradeUserConnection(ctx context.Context, sessionID hub.SessionID, txID string) {
	pp.send <- &Request{
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
}

// SendHoldSupsMessage tells the passport to hold sups
func (pp *Passport) SendHoldSupsMessage(ctx context.Context, userID server.UserID, supsChange server.BigInt, txID string, reason string) (server.TransactionReference, error) {
	supTransactionReference := uuid.Must(uuid.NewV4())
	supTxRefString := server.TransactionReference(fmt.Sprintf("%s|%s", reason, supTransactionReference.String()))
	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "SUPREMACY:HOLD_SUPS",
			Payload: struct {
				Amount               server.BigInt               `json:"amount"`
				FromUserID           server.UserID               `json:"userID"`
				TransactionReference server.TransactionReference `json:"transactionReference"`
			}{
				FromUserID:           userID,
				Amount:               supsChange,
				TransactionReference: supTxRefString,
			},
			TransactionID: txID,
			context:       ctx,
		}}

	msg := <-replyChannel
	resp := struct {
		Payload struct {
			IsSuccess bool `json:"isSuccess"`
		}
	}{}

	err := json.Unmarshal(msg, &resp)
	if err != nil {
		return supTxRefString, terror.Error(err)
	}

	// this doesn't mean the sup transfer was successful, it means the server was successful at getting our message
	if !resp.Payload.IsSuccess {
		return supTxRefString, terror.Error(fmt.Errorf("passport success resp is false"))
	}

	return supTxRefString, nil
}

// SendTickerMessage sends the client map and multipliers to the passport to handle giving out sups
func (pp *Passport) SendTickerMessage(ctx context.Context, userMap map[int][]server.UserID) (string, error) {
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:TICKER_TICK",
			Payload: struct {
				UserMap map[int][]server.UserID `json:"userMap"`
			}{
				UserMap: userMap,
			},
			context: ctx,
		}}

	return "", nil
}
