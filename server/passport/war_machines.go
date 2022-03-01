package passport

import (
	"context"
	"server"
)

type DefaultWarMachinesReq struct {
	FactionID server.FactionID `json:"factionID"`
}

type DefaultWarMachinesResp struct {
	WarMachines []*server.WarMachineMetadata `json:"warMachines"`
}

// GetDefaultWarMachines gets the default war machines for a given faction
func (pp *Passport) GetDefaultWarMachines(ctx context.Context, factionID server.FactionID, callback func(warMachines []*server.WarMachineMetadata)) {
	resp := &DefaultWarMachinesResp{}
	err := pp.Comms.Call("C.SupremacyDefaultWarMachinesHandler", DefaultWarMachinesReq{factionID}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyDefaultWarMachinesHandler").Msg("rpc error")
		return
	}
	callback(resp.WarMachines)
}
