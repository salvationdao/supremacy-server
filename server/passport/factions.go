package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/ninja-software/terror/v2"
)

type FactionAllResponse struct {
	Factions []*server.Faction `json:"payload"`
}

// FactionAll get all the factions from passport server
func (pp *Passport) FactionAll(ctx context.Context, txID string) ([]*server.Faction, error) {
	replyChannel := make(chan []byte)
	errChan := make(chan error)

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
		case err := <-errChan:
			if err != nil {
				return nil, terror.Error(err)
			}
		}
	}
}
