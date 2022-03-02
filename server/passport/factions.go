package passport

import (
	"server"

	"github.com/ninja-syndicate/hub"
)

type FactionAllReq struct{}

type FactionAllResp struct {
	Factions []*server.Faction `json:"factions"`
}

// FactionAll get all the factions from passport server
func (pp *Passport) FactionAll() ([]*server.Faction, error) {
	resp := &FactionAllResp{}
	err := pp.Comms.Call("C.SupremacyFactionAllHandler", FactionAllReq{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Factions, nil
}

//****************************************
//  STAT
//****************************************

type FactionStatSendReq struct {
	FactionStatSends []*FactionStatSend `json:"factionStatSends"`
}

type FactionStatSend struct {
	FactionStat     *server.FactionStat `json:"factionStat"`
	ToUserID        *server.UserID      `json:"toUserID,omitempty"`
	ToUserSessionID *hub.SessionID      `json:"toUserSessionID,omitempty"`
}

type FactionStatSendResp struct{}

// FactionStatsSend send faction stat to passport serer
func (pp *Passport) FactionStatsSend(factionStatSends []*FactionStatSend) {
	err := pp.Comms.Call("C.SupremacyFactionStatSendHandler", FactionStatSendReq{factionStatSends}, &FactionStatSendResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyFactionStatSendHandler").Msg("rpc error")
	}
}

//****************************************
//  CONTRACT REWARD
//****************************************

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

//****************************************
//  QUEUE COST
//****************************************

type FactionQueuePriceUpdateReq struct {
	FactionID     server.FactionID `json:"factionID"`
	QueuingLength int              `json:"queuingLength"`
}

type FactionQueuePriceUpdateResp struct {
}

func (pp *Passport) FactionQueueCostUpdate(fcr *FactionQueuePriceUpdateReq) {
	err := pp.Comms.Call("C.SupremacyFactionQueuingCostHandler", fcr, &FactionQueuePriceUpdateResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyFactionQueuingCostHandler").Msg("rpc error")
	}
}
