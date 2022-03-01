package passport

import (
	"server"
)

type WarMachineQueuePositionReq struct {
	UserWarMachineQueuePosition []*UserWarMachineQueuePosition `json:"userWarMachineQueuePosition"`
}

type UserWarMachineQueuePosition struct {
	UserID                   server.UserID              `json:"userID"`
	WarMachineQueuePositions []*WarMachineQueuePosition `json:"warMachineQueuePositions"`
}

type WarMachineQueuePosition struct {
	WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	Position           int                        `json:"position"`
}

type WarMachineQueuePositionResp struct{}

func (pp *Passport) WarMachineQueuePositionBroadcast(uwm []*UserWarMachineQueuePosition) {
	if len(uwm) == 0 {
		return
	}
	err := pp.Comms.Call("C.SupremacyWarMachineQueuePositionHandler", WarMachineQueuePositionReq{uwm}, &WarMachineQueuePositionResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyWarMachineQueuePositionHandler").Msg("rpc error")
	}
}

type RedeemFactionContractRewardReq struct {
	UserID               server.UserID               `json:"userID"`
	FactionID            server.FactionID            `json:"factionID"`
	Amount               string                      `json:"amount"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
}

type RedeemFactionContractRewardResp struct{}

// AssetContractRewardRedeem redeem faction contract reward
func (pp *Passport) AssetContractRewardRedeem(userID server.UserID, factionID server.FactionID, amount string, txRef server.TransactionReference) {
	err := pp.Comms.Call("C.SupremacyRedeemFactionContractRewardHandler", RedeemFactionContractRewardReq{userID, factionID, amount, txRef}, &RedeemFactionContractRewardResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyRedeemFactionContractRewardHandler").Msg("rpc error")
	}
}

type AssetRepairStatReq struct {
	AssetRepairRecord *server.AssetRepairRecord `json:"assetRepairRecord"`
}

type AssetRepairStatResp struct{}

// AssetContractRewardRedeem redeem faction contract reward
func (pp *Passport) AssetRepairStat(arr *server.AssetRepairRecord) {
	err := pp.Comms.Call("C.SupremacyAssetRepairStatUpdateHandler", AssetRepairStatReq{arr}, &AssetRepairStatResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyAssetRepairStatUpdateHandler").Msg("rpc error")
	}
}
