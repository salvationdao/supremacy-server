package passport

import (
	"server/rpcclient"
)

func (pp *Passport) WarMachineQueuePositionBroadcast(wmp []*rpcclient.WarMachineQueueStat) {
	if len(wmp) == 0 {
		return
	}
	err := pp.RPCClient.Call("S.SupremacyWarMachineQueuePositionHandler", rpcclient.WarMachineQueuePositionReq{WarMachineQueuePosition: wmp}, &rpcclient.WarMachineQueuePositionResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyWarMachineQueuePositionHandler").Msg("rpc error")
	}
}
