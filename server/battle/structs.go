package battle

import "github.com/gofrs/uuid"

type Battle struct {
	arena       *Arena
	ID          uuid.UUID     `json:"battleID" db:"id"`
	MapName     string        `json:"mapName"`
	WarMachines []*WarMachine `json:"warMachines"`
}

type Started struct {
	BattleID           string        `json:"battleID"`
	WarMachines        []*WarMachine `json:"warMachines"`
	WarMachineLocation []byte        `json:"warMachineLocation"`
}
