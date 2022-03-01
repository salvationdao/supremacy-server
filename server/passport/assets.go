package passport

import (
	"server"
)

type WarMachineQueuePositionReq struct {
	UserWarMachineQueuePosition []*WarMachineQueuePosition `json:"userWarMachineQueuePosition"`
}

type WarMachineQueuePosition struct {
	Hash     string `json:"hash"`
	Position *int   `json:"position,omitempty"`
}

type WarMachineQueuePositionResp struct{}

func (pp *Passport) WarMachineQueuePositionBroadcast(wmp []*WarMachineQueuePosition) {
	if len(wmp) == 0 {
		return
	}
	err := pp.Comms.Call("C.SupremacyWarMachineQueuePositionHandler", WarMachineQueuePositionReq{wmp}, &WarMachineQueuePositionResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyWarMachineQueuePositionHandler").Msg("rpc error")
	}
}

type AssetRepairStatReq struct {
	AssetRepairRecord *server.AssetRepairRecord `json:"assetRepairRecord"`
}

type AssetRepairStatResp struct{}

// AssetRepairStat redeem faction contract reward
func (pp *Passport) AssetRepairStat(arr *server.AssetRepairRecord) {
	err := pp.Comms.Call("C.SupremacyAssetRepairStatUpdateHandler", AssetRepairStatReq{arr}, &AssetRepairStatResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyAssetRepairStatUpdateHandler").Msg("rpc error")
	}
}
