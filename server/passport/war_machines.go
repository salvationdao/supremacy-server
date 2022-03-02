package passport

import (
	"context"
	"fmt"
	"server"
)

type DefaultWarMachinesReq struct {
	FactionID server.FactionID `json:"factionID"`
}

type DefaultWarMachinesResp struct {
	WarMachines []*server.WarMachineMetadata `json:"warMachines"`
}

// GetDefaultWarMachines gets the default war machines for a given faction
func (pp *Passport) GetDefaultWarMachines(ctx context.Context, factionID server.FactionID) ([]*server.WarMachineMetadata, error) {
	resp := &DefaultWarMachinesResp{}
	err := pp.Comms.Call("C.SupremacyDefaultWarMachinesHandler", DefaultWarMachinesReq{factionID}, resp)
	if err != nil {
		return nil, fmt.Errorf("GetDefaultWarMachines: %w", err)
	}
	return resp.WarMachines, nil
}
