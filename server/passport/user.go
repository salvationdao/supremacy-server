package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"server"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
)

type User struct {
	ID          server.UserID   `json:"id"`
	Faction     *server.Faction `json:"faction"`
	Sups        server.BigInt   `json:"sups"`
	PassportURL string          `json:"passportURL"`
}

type TwitchAuthResponse struct {
	Payload TwitchAuthPayload `json:"payload"`
}

type TwitchAuthPayload struct {
	User server.User `json:"user"`
}

func (pp *Passport) TwitchAuth(ctx context.Context, token string, txID string) (*server.User, error) {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "AUTH:TWITCH",
			Payload: struct {
				Token string `json:"token"`
			}{
				Token: token,
			},
			TransactionId: txID,
			context:       ctx,
			cancel:        cancel,
		}}

	msg := <-replyChannel
	resp := &TwitchAuthResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &resp.Payload.User, nil
}

type UserGetByIDResponse struct {
	User server.User `json:"payload"`
}

func (pp *Passport) UserGetByID(ctx context.Context, userID server.UserID, txID string) (*server.User, error) {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "USER:GET",
			Payload: struct {
				ID server.UserID `json:"id"`
			}{
				ID: userID,
			},
			TransactionId: txID,
			context:       ctx,
			cancel:        cancel,
		}}

	msg := <-replyChannel
	resp := &UserGetByIDResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &resp.User, nil

}

type UserGetByUsernameResponse struct {
	User server.User `json:"payload"`
}

func (pp *Passport) UserGetByUsername(ctx context.Context, username string, txID string) (*server.User, error) {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "USER:GET",
			Payload: struct {
				Username string `json:"username"`
			}{
				Username: username,
			},
			TransactionId: txID,
			context:       ctx,
			cancel:        cancel,
		}}

	msg := <-replyChannel

	resp := &UserGetByUsernameResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return nil, terror.Error(err)
	}
	return &resp.User, nil
}

// UserFactionUpdate update the faction of the given user
func (pp *Passport) UserFactionUpdate(ctx context.Context, userID server.UserID, factionID server.FactionID, txID string) error {
	ctx, cancel := context.WithCancel(ctx)

	pp.send <- &Request{
		Message: &Message{
			Key: "USER:FACTION:UPDATE",
			Payload: struct {
				UserID    server.UserID    `json:"userID"`
				FactionID server.FactionID `json:"factionID"`
			}{
				UserID:    userID,
				FactionID: factionID,
			},
			TransactionId: txID,
			context:       ctx,
			cancel:        cancel,
		}}

	return nil
}

// SendTakeSupsMessage tells the passport to transfer sups
// THIS DOES NOT CONFIRM IF THE TRANSACTION WAS SUCCESSFUL
// TO CONFIRM SUCCESS NEED TO CALL ENDPOINT FOR THE transactionReference (TODO: THIS)
func (pp *Passport) SendTakeSupsMessage(ctx context.Context, userID server.UserID, supsChange server.BigInt, txID string, reason string) (server.TransactionReference, error) {
	ctx, cancel := context.WithCancel(ctx)

	supTransactionReference := uuid.Must(uuid.NewV4())
	supTxRefString := server.TransactionReference(fmt.Sprintf("%s|%s", reason, supTransactionReference.String()))
	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "SUPREMACY:TAKE_SUPS",
			Payload: struct {
				Amount               server.BigInt               `json:"amount"`
				FromUserID           server.UserID               `json:"userId"`
				TransactionReference server.TransactionReference `json:"transactionReference"`
			}{
				FromUserID:           userID,
				Amount:               supsChange,
				TransactionReference: supTxRefString,
			},
			TransactionId: txID,
			context:       ctx,
			cancel:        cancel,
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
	ctx, cancel := context.WithCancel(ctx)

	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:TICKER_TICK",
			Payload: struct {
				UserMap map[int][]server.UserID `json:"userMap"`
			}{
				UserMap: userMap,
			},
			context: ctx,
			cancel:  cancel,
		}}

	return "", nil
}
