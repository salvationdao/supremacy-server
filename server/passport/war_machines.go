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
	}
	callback(resp.WarMachines)
}

type FactionContractRewardUpdateReq struct {
	FactionContractRewards []*FactionContractReward `json:"factionContractRewards"`
}

type FactionContractReward struct {
	FactionID      server.FactionID `json:"factionID"`
	ContractReward string           `json:"contractReward"`
}

type FactionContractRewardUpdateResp struct {
}

// FactionContractRewardUpdate gets the default war machines for a given faction
func (pp *Passport) FactionContractRewardUpdate(fcr []*FactionContractReward) {
	err := pp.Comms.Call("C.SupremacyFactionContractRewardUpdateHandler", FactionContractRewardUpdateReq{fcr}, &FactionContractRewardUpdateResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyFactionContractRewardUpdateHandler").Msg("rpc error")
	}
}
