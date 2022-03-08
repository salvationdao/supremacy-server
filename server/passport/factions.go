package passport

import (
	"fmt"
	"server"

	"github.com/ninja-syndicate/hub"
	"github.com/shopspring/decimal"
)

type FactionAllReq struct{}

type FactionAllResp struct {
	Factions []*server.Faction `json:"factions"`
}

// FactionAll get all the factions from passport server
func (pp *Passport) FactionAll() ([]*server.Faction, error) {
	resp := &FactionAllResp{}
	err := pp.Comms.Call("S.SupremacyFactionAllHandler", FactionAllReq{}, resp)
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
	err := pp.Comms.Call("S.SupremacyFactionStatSendHandler", FactionStatSendReq{factionStatSends}, &FactionStatSendResp{})
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
	BattleID             string                      `json:"battleID"` //TODO: SEND BATTLE ID
	Amount               string                      `json:"amount"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
}

type RedeemFactionContractRewardResp struct{}

// AssetContractRewardRedeem redeem faction contract reward
func (pp *Passport) AssetContractRewardRedeem(userID server.UserID, factionID server.FactionID, amount decimal.Decimal, txRef server.TransactionReference, battleID string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("AssetContractRewardRedeem: amount must be greater than zero")
	}
	err := pp.Comms.Call(
		"S.SupremacyRedeemFactionContractRewardHandler",
		RedeemFactionContractRewardReq{
			UserID:               userID,
			FactionID:            factionID,
			Amount:               amount.String(),
			TransactionReference: txRef,
			BattleID:             battleID,
		},
		&RedeemFactionContractRewardResp{})
	if err != nil {
		return fmt.Errorf("SupremacyRedeemFactionContractRewardHandler rpc call: %w", err)
	}
	return nil
}

/*
type FactionContractRewardUpdateReq struct {
	FactionContractRewards []*FactionContractReward `json:"factionContractRewards"`
}

type FactionContractReward struct {
	FactionID      server.FactionID `json:"factionID"`
	ContractReward string           `json:"contractReward"`
}

type FactionContractRewardUpdateResp struct {
}
*/
// FactionContractRewardUpdate gets the default war machines for a given faction
//func (pp *Passport) FactionContractRewardUpdate(fcr []*FactionContractReward) {
//	err := pp.Comms.Call("S.SupremacyFactionContractRewardUpdateHandler", FactionContractRewardUpdateReq{fcr}, &FactionContractRewardUpdateResp{})
//	if err != nil {
//		pp.Log.Err(err).Str("method", "SupremacyFactionContractRewardUpdateHandler").Msg("rpc error")
//	}
//}

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
	err := pp.Comms.Call("S.SupremacyFactionQueuingCostHandler", fcr, &FactionQueuePriceUpdateResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyFactionQueuingCostHandler").Msg("rpc error")
	}
}
