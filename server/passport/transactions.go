package passport

import (
	"context"
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
func (pp *Passport) ReleaseTransactions(ctx context.Context, transactions []string) {
	if len(transactions) == 0 {
		return
	}
	err := pp.Comms.Call("C.ReleaseTransactionsHandler", ReleaseTransactionsReq{transactions}, &ReleaseTransactionsResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "ReleaseTransactionsHandler").Msg("rpc error")
	}
}

// TransferBattleFundToSupsPool tells the passport to transfer fund to sup pool
func (pp *Passport) TransferBattleFundToSupsPool(ctx context.Context) error {
	pp.send <- &Message{
		Key:           "SUPREMACY:TRANSFER_BATTLE_FUND_TO_SUP_POOL",
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}

type SupremacyTopSupsContributorResponse struct {
	Payload SupremacyTopSupsContributor `json:"payload"`
}

type SupremacyTopSupsContributor struct {
	TopSupsContributors       []*server.User    `json:"topSupsContributors"`
	TopSupsContributeFactions []*server.Faction `json:"topSupsContributeFactions"`
}

// TopSupsContributorsGet tells the passport to return the top three most sups contributors with in the time frame
func (pp *Passport) TopSupsContributorsGet(ctx context.Context, startTime, endTime time.Time, callback func(msg []byte)) error {
	pp.send <- &Message{
		Key: "SUPREMACY:TOP_SUPS_CONTRIBUTORS",
		Payload: struct {
			StartTime time.Time `json:"startTime"`
			EndTime   time.Time `json:"endTime"`
		}{
			StartTime: startTime,
			EndTime:   endTime,
		},
		Callback:      callback,
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}
