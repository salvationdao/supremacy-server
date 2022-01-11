package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
)

type User struct {
	ID            server.UserID   `json:"id"`
	Faction       *server.Faction `json:"faction"`
	ConnectPoint  int64           `json:"connectPoint"`
	SupremacyCoin int64           `json:"supremacyCoin"`
	PassportURL   string          `json:"passportURL"`
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

	select {
	case msg := <-replyChannel:
		resp := &UserGetByIDResponse{}
		err := json.Unmarshal(msg, resp)
		if err != nil {
			return nil, terror.Error(err)
		}

		return &resp.User, nil
	}
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

var defaultNamespaceUUID = uuid.Must(uuid.FromString("8f2d7180-bbe3-47b0-96ef-ee3e64697387"))

func (pp *Passport) FakeUserLoginWithFaction(twitchUserID string) *server.User {
	// we will auth with passport, and we'll get a passport user and convert it to a server.user
	return &server.User{
		ID:            server.UserID(uuid.NewV3(defaultNamespaceUUID, twitchUserID)),
		Faction:       FakeFactions[0],
		FactionID:     FakeFactions[0].ID,
		ConnectPoint:  12345,
		SupremacyCoin: 12345,
		PassportURL:   "", // url we get from passport which has a token in the url to take them to their passport page
	}
}

func (pp *Passport) FakeUserLoginWithoutFaction(twitchUserID string) *server.User {
	// we will auth with passport, and we'll get a passport user and convert it to a server.user
	return &server.User{
		ID:            server.UserID(uuid.NewV3(defaultNamespaceUUID, twitchUserID)),
		ConnectPoint:  12345,
		SupremacyCoin: 12345,
		PassportURL:   "https://dev.supremacygame.io/api/temp-random-faction", // url we get from passport which has a token in the url to take them to their passport page
	}
}
