package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/ninja-software/terror/v2"
)

type User struct {
	ID          server.UserID   `json:"id"`
	Faction     *server.Faction `json:"faction"`
	Sups        int64           `json:"sups"`
	PassportURL string          `json:"passportURL"`
}

func (pp *Passport) TwitchAuth(ctx context.Context, twitchToken string, txID string) (*server.User, error) {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "TWITCH:AUTH",
			Payload: struct {
				TwitchToken string `json:"twitchToken"`
			}{
				TwitchToken: twitchToken,
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

	select {
	case msg := <-replyChannel:
		resp := &UserGetByUsernameResponse{}
		err := json.Unmarshal(msg, resp)
		if err != nil {
			return nil, terror.Error(err)
		}
		return &resp.User, nil
	}
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

// UserSupsUpdate update user sups
func (pp *Passport) UserSupsUpdate(ctx context.Context, userID server.UserID, supsChange int64, txID string) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "USER:SUPS:UPDATE",
			Payload: struct {
				UserID     server.UserID `json:"userID"`
				SupsChange int64         `json:"supsChange"`
			}{
				UserID:     userID,
				SupsChange: supsChange,
			},
			TransactionId: txID,
			context:       ctx,
			cancel:        cancel,
		}}

	msg := <-replyChannel
	resp := struct {
		isSuccess bool
	}{
		isSuccess: true,
	}
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		return false, terror.Error(err)
	}
	return resp.isSuccess, nil
}
