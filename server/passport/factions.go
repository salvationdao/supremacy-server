package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
)

type FactionAllResponse struct {
	Factions []*server.Faction `json:"payload"`
}

// FactionAll get all the factions from passport server
func (pp *Passport) FactionAll(ctx context.Context, txID string) ([]*server.Faction, error) {
	replyChannel := make(chan []byte, 1)
	errChan := make(chan error, 1)

	pp.send <- &Request{
		ErrChan:      errChan,
		ReplyChannel: replyChannel,
		Message: &Message{
			Key:           "FACTION:ALL",
			TransactionID: txID,
			context:       ctx,
		}}

	for {
		select {
		case msg := <-replyChannel:
			resp := &FactionAllResponse{}
			err := json.Unmarshal(msg, resp)
			if err != nil {
				return nil, terror.Error(err)
			}
			return resp.Factions, nil
		case err := <-errChan:
			return nil, terror.Error(err)
		}
	}
}

type FactionStatSend struct {
	FactionStat     *server.FactionStat `json:"factionStat"`
	ToUserID        *server.UserID      `json:"toUserID,omitempty"`
	ToUserSessionID *hub.SessionID      `json:"toUserSessionID,omitempty"`
}

// FactionStatsSend send faction stat to passport serer
func (pp *Passport) FactionStatsSend(ctx context.Context, factionStatSends []*FactionStatSend) error {
	replyChannel := make(chan []byte, 1)
	errChan := make(chan error, 1)

	pp.send <- &Request{
		ErrChan:      errChan,
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "SUPREMACY:FACTION_STAT_SEND",
			Payload: struct {
				FactionStatSends []*FactionStatSend `json:"factionStatSends"`
			}{
				FactionStatSends: factionStatSends,
			},
			TransactionID: uuid.Must(uuid.NewV4()).String(),
			context:       ctx,
		}}

	for {
		select {
		case <-replyChannel:
			return nil
		case err := <-errChan:
			return terror.Error(err)
		}
	}
}
