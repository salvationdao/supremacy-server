package passport

import (
	"context"
	"fmt"
	"server"
	"time"

	"github.com/gofrs/uuid"
)

// type CommitTransactionsResponse struct {
// 	Transactions []*server.Transaction `json:"payload"`
// }

// func (pp *Passport) CommitTransactions(ctx context.Context, transactions []server.TransactionReference) ([]*server.Transaction, error) {
// 	if len(transactions) == 0 {
// 		return nil, nil
// 	}

// 	replyChannel := make(chan []byte)
// 	errChan := make(chan error)

// 	pp.send <- &Request{
// 		ReplyChannel: replyChannel,
// 		ErrChan:      errChan,
// 		Message: &Message{
// 			Key: "SUPREMACY:COMMIT_TRANSACTIONS",
// 			Payload: struct {
// 				TransactionReferences []server.TransactionReference `json:"transactionReferences"`
// 			}{
// 				TransactionReferences: transactions,
// 			},
// 			TransactionID: uuid.Must(uuid.NewV4()).String(),
// 		}}

// 	for {
// 		select {
// 		case msg := <-replyChannel:
// 			resp := &CommitTransactionsResponse{}
// 			err := json.Unmarshal(msg, resp)
// 			if err != nil {
// 				return nil, terror.Error(err)
// 			}
// 			return resp.Transactions, nil
// 		case err := <-errChan:
// 			return nil, terror.Error(err)
// 		}
// 	}
// }

type HoldSupsMessageResponse struct {
	Transaction string `json:"payload"`
}

// SpendSupMessage tells the passport to hold sups
func (pp *Passport) SpendSupMessage(userID server.UserID, supsChange server.BigInt, battleID server.BattleID, reason string, callback func(msg []byte)) {
	supTransactionReference := uuid.Must(uuid.NewV4())
	supTxRefString := server.TransactionReference(fmt.Sprintf("%s|%s", reason, supTransactionReference.String()))
	id := fmt.Sprintf("sups hold - %s", uuid.Must(uuid.NewV4()).String())

	pp.send <- &Message{
		Key:      "SUPREMACY:HOLD_SUPS",
		Callback: callback,
		Payload: struct {
			FromUserID           server.UserID               `json:"userID"`
			Amount               server.BigInt               `json:"amount"`
			TransactionReference server.TransactionReference `json:"transactionReference"`
			GroupID              string                      `json:"groupID"`
		}{
			FromUserID:           userID,
			Amount:               supsChange,
			TransactionReference: supTxRefString,
			GroupID:              battleID.String(),
		},
		TransactionID: id,
	}
}

// ReleaseTransactions tells the passport to transfer fund to sup pool
func (pp *Passport) ReleaseTransactions(ctx context.Context, transactions []string) {
	if len(transactions) == 0 {
		return
	}

	pp.send <- &Message{
		Key: "SUPREMACY:RELEASE_TRANSACTIONS",
		Payload: struct {
			Transactions []string `json:"txIDs"`
		}{
			Transactions: transactions,
		},
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
