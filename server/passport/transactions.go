package passport

import (
	"fmt"
	"server"
	"time"

	"github.com/gofrs/uuid"
)

type HoldSupsMessageResponse struct {
	Transaction string `json:"payload"`
}

type SpendSupsReq struct {
	FromUserID           server.UserID               `json:"userID"`
	Amount               string                      `json:"amount"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
	GroupID              string
}
type SpendSupsResp struct {
	TXID string `json:"txid"`
}

// SpendSupMessage tells the passport to hold sups
func (pp *Passport) SpendSupMessage(userID server.UserID, supsChange server.BigInt, battleID server.BattleID, reason string, callback func(txID string)) {
	supTransactionReference := uuid.Must(uuid.NewV4())
	supTxRefString := server.TransactionReference(fmt.Sprintf("%s|%s", reason, supTransactionReference.String()))
	resp := &SpendSupsResp{}
	pp.Comms.GoCall("C.SupremacySpendSupsHandler", SpendSupsReq{FromUserID: userID, Amount: supsChange.String(), TransactionReference: supTxRefString, GroupID: battleID.String()}, resp, func(err error) {
		if err != nil {
			pp.Log.Err(err).Str("method", "SupremacySpendSupsHandler").Msg("rpc error")
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
	err := pp.Comms.Call("C.ReleaseTransactionsHandler", ReleaseTransactionsReq{transactions}, &ReleaseTransactionsResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "ReleaseTransactionsHandler").Msg("rpc error")
	}
}

type TransferBattleFundToSupPoolReq struct{}
type TransferBattleFundToSupPoolResp struct{}

// ReleaseTransactions tells the passport to transfer fund to sup pool
func (pp *Passport) TransferBattleFundToSupsPool() {
	err := pp.Comms.Call("C.TransferBattleFundToSupPoolHandler", TransferBattleFundToSupPoolReq{}, &TransferBattleFundToSupPoolResp{})
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
	err := pp.Comms.Call("C.TopSupsContributorHandler", TopSupsContributorReq{}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "TopSupsContributorHandler").Msg("rpc error")
	}
	callback(resp)
}
