package passport

import (
	"server"
	"server/comms"
)

func (pp *Passport) WarMachineQueuePositionBroadcast(wmp []*comms.WarMachineQueueStat) {
	if len(wmp) == 0 {
		return
	}
	err := pp.Comms.Call("S.SupremacyWarMachineQueuePositionHandler", comms.WarMachineQueuePositionReq{WarMachineQueuePosition: wmp}, &comms.WarMachineQueuePositionResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyWarMachineQueuePositionHandler").Msg("rpc error")
	}
}

// AssetRepairStat redeem faction contract reward
func (pp *Passport) AssetRepairStat(arr *server.AssetRepairRecord) {
	err := pp.Comms.Call("S.SupremacyAssetRepairStatUpdateHandler", comms.AssetRepairStatReq{AssetRepairRecord: arr}, &comms.AssetRepairStatResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyAssetRepairStatUpdateHandler").Msg("rpc error")
	}
}

func (pp *Passport) SupremacyQueueUpdate(arr *server.SupremacyQueueUpdateReq) {
	err := pp.Comms.Call("S.SupremacyQueueUpdateHandler", arr, &comms.AssetRepairStatResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyQueueUpdateHandler").Msg("rpc error")
	}
}
