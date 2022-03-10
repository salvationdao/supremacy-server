package passport

import (
	"server"
	"server/comms"
	"server/rpcclient"
)

func (pp *Passport) WarMachineQueuePositionBroadcast(wmp []*rpcclient.WarMachineQueueStat) {
	if len(wmp) == 0 {
		return
	}
	err := pp.RPCClient.Call("S.SupremacyWarMachineQueuePositionHandler", rpcclient.WarMachineQueuePositionReq{WarMachineQueuePosition: wmp}, &comms.WarMachineQueuePositionResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyWarMachineQueuePositionHandler").Msg("rpc error")
	}
}

// AssetRepairStat redeem faction contract reward
func (pp *Passport) AssetRepairStat(arr *server.AssetRepairRecord) {
	err := pp.RPCClient.Call("S.SupremacyAssetRepairStatUpdateHandler", rpcclient.AssetRepairStatReq{AssetRepairRecord: arr}, &comms.AssetRepairStatResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyAssetRepairStatUpdateHandler").Msg("rpc error")
	}
}

func (pp *Passport) SupremacyQueueUpdate(arr *server.SupremacyQueueUpdateReq) {
	err := pp.RPCClient.Call("S.SupremacyQueueUpdateHandler", arr, &rpcclient.AssetRepairStatResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyQueueUpdateHandler").Msg("rpc error")
	}
}
