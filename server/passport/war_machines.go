package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

// GetDefaultWarMachines gets the default war machines for a given faction
func (pp *Passport) GetDefaultWarMachines(ctx context.Context, factionID server.FactionID, amount int) ([]*server.WarMachineNFT, error) {
	replyChannel := make(chan []byte, 1)
	errChan := make(chan error, 1)

	txID, err := uuid.NewV4()
	if err != nil {
		return nil, terror.Error(err)
	}

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:GET_DEFAULT_WAR_MACHINES",
			Payload: struct {
				FactionID server.FactionID `json:"factionID"`
				Amount    int              `json:"amount"`
			}{
				FactionID: factionID,
				Amount:    amount,
			},
			TransactionID: txID.String(),
			context:       ctx,
		}}

	for {
		select {
		case msg := <-replyChannel:
			resp := struct {
				WarMachines []*server.WarMachineNFT `json:"payload"`
			}{}
			err := json.Unmarshal(msg, &resp)
			if err != nil {
				return nil, terror.Error(err)
			}
			return resp.WarMachines, nil
		case err := <-errChan:
			return nil, terror.Error(err)
		}
	}
}

// FactionWarMachineContractRewardUpdate gets the default war machines for a given faction
func (pp *Passport) FactionWarMachineContractRewardUpdate(ctx context.Context, fwm []*server.FactionWarMachineQueue) {
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:WAR_MACHINE_QUEUE_CONTRACT_UPDATE",
			Payload: struct {
				FactionWarMachineQueues []*server.FactionWarMachineQueue `json:"factionWarMachineQueues"`
			}{
				FactionWarMachineQueues: fwm,
			},
			context: ctx,
		},
	}
}
