package passport

import (
	"context"
	"server"

	"github.com/gofrs/uuid"
)

// GetDefaultWarMachines gets the default war machines for a given faction
func (pp *Passport) GetDefaultWarMachines(ctx context.Context, factionID server.FactionID, amount int, callback func(msg []byte)) {
	pp.send <- &Message{
		Key: "SUPREMACY:GET_DEFAULT_WAR_MACHINES",
		Payload: struct {
			FactionID server.FactionID `json:"factionID"`
			Amount    int              `json:"amount"`
		}{
			FactionID: factionID,
			Amount:    amount,
		},
		Callback:      callback,
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
}

// FactionWarMachineContractRewardUpdate gets the default war machines for a given faction
func (pp *Passport) FactionWarMachineContractRewardUpdate(fwm []*server.FactionWarMachineQueue) {
	pp.send <- &Message{
		Key: "SUPREMACY:WAR_MACHINE_QUEUE_CONTRACT_UPDATE",
		Payload: struct {
			FactionWarMachineQueues []*server.FactionWarMachineQueue `json:"factionWarMachineQueues"`
		}{
			FactionWarMachineQueues: fwm,
		},
	}
}
