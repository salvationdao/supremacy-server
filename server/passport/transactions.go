package passport

import (
	"server"
	"time"
)

type HoldSupsMessageResponse struct {
	Transaction string `json:"payload"`
}

type SpendSupsReq struct {
	Amount               string                      `json:"amount"`
	FromUserID           server.UserID               `json:"fromUserID"`
	ToUserID             *server.UserID              `json:"toUserID,omitempty"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
	Group                string                      `json:"group,omitempty"`
	SubGroup             string                      `json:"subGroup"`    //TODO: send battle id
	Description          string                      `json:"description"` //TODO: send descritpion

	NotSafe bool `json:"notSafe"`
}
type SpendSupsResp struct {
	TXID string `json:"txid"`
}

// SpendSupMessage tells the passport to hold sups
func (pp *Passport) SpendSupMessage(req SpendSupsReq, callback func(txID string), errorCallback func(err error)) {
	resp := &SpendSupsResp{}
	pp.Comms.GoCall("S.SupremacySpendSupsHandler", req, resp, func(err error) {
		if err != nil {
			pp.Log.Err(err).Str("method", "SupremacySpendSupsHandler").Msg("rpc error")
			errorCallback(err)
			return
		}
		callback(resp.TXID)
	})
}

type ReleaseTransactionsReq struct {
	TxIDs []string `json:"txIDs"`
}
type ReleaseTransactionsResp struct{}

// ReleaseTransactions tells the passport to transfer fund to sup pool
func (pp *Passport) ReleaseTransactions(transactions []string) {
	if len(transactions) == 0 {
		return
	}
	err := pp.Comms.Call("S.ReleaseTransactionsHandler", ReleaseTransactionsReq{transactions}, &ReleaseTransactionsResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "ReleaseTransactionsHandler").Msg("rpc error")
	}
}

type TransferBattleFundToSupPoolReq struct{}
type TransferBattleFundToSupPoolResp struct{}

func (pp *Passport) TransferBattleFundToSupsPool() {
	err := pp.Comms.Call("S.TransferBattleFundToSupPoolHandler", TransferBattleFundToSupPoolReq{}, &TransferBattleFundToSupPoolResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "TransferBattleFundToSupPoolHandler").Msg("rpc error")
	}
}

type TopSupsContributorReq struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type TopSupsContributorResp struct {
	TopSupsContributors       []*server.User    `json:"topSupsContributors"`
	TopSupsContributeFactions []*server.Faction `json:"topSupsContributeFactions"`
}

// ReleaseTransactions tells the passport to transfer fund to sup pool
func (pp *Passport) TopSupsContributorsGet(startTime, endTime time.Time, callback func(result *TopSupsContributorResp)) {
	resp := &TopSupsContributorResp{}
	err := pp.Comms.Call("S.TopSupsContributorHandler", TopSupsContributorReq{
		StartTime: startTime,
		EndTime:   endTime,
	}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "TopSupsContributorHandler").Msg("rpc error")
		return
	}

	callback(resp)
}
