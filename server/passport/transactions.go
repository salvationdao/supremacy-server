package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
)

type CommitTransactionsResponse struct {
	Transactions []*server.Transaction `json:"payload"`
}

func (pp *Passport) CommitTransactions(ctx context.Context, transactions []server.TransactionReference) ([]*server.Transaction, error) {
	if len(transactions) == 0 {
		return nil, nil
	}

	fmt.Println("")
	fmt.Println("")
	fmt.Println("STARTING TX COMMIT")
	fmt.Println("")
	fmt.Println("")

	replyChannel := make(chan []byte)
	errChan := make(chan error)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:COMMIT_TRANSACTIONS",
			Payload: struct {
				TransactionReferences []server.TransactionReference `json:"transactionReferences"`
			}{
				TransactionReferences: transactions,
			},
			TransactionID: uuid.Must(uuid.NewV4()).String(),
		}}

	for {
		select {
		case msg := <-replyChannel:
			pp.Log.Info().Msg("")
			pp.Log.Info().Msg("")
			pp.Log.Info().Msg("ENDING TX COMMIT")
			pp.Log.Info().Msg("")
			pp.Log.Info().Msg("")
			resp := &CommitTransactionsResponse{}
			err := json.Unmarshal(msg, resp)
			if err != nil {
				return nil, terror.Error(err)
			}
			return resp.Transactions, nil
		case err := <-errChan:
			pp.Log.Info().Msg("")
			pp.Log.Info().Msg("")
			pp.Log.Info().Msg("ENDING TX COMMIT ERR")
			pp.Log.Err(err).Msg("")
			pp.Log.Info().Msg("")
			return nil, terror.Error(err)
		}
	}
}

// SendHoldSupsMessage tells the passport to hold sups
func (pp *Passport) SendHoldSupsMessage(ctx context.Context, userID server.UserID, supsChange server.BigInt, reason string) (server.TransactionReference, error) {
	supTransactionReference := uuid.Must(uuid.NewV4())
	supTxRefString := server.TransactionReference(fmt.Sprintf("%s|%s", reason, supTransactionReference.String()))
	replyChannel := make(chan []byte)
	errChan := make(chan error)

	id := fmt.Sprintf("sups hold - %s", uuid.Must(uuid.NewV4()).String())
	pp.Log.Info().Str("TransactionID", id).Msg("STARTING SUP HOLD")

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:HOLD_SUPS",
			Payload: struct {
				FromUserID           server.UserID               `json:"userID"`
				Amount               server.BigInt               `json:"amount"`
				TransactionReference server.TransactionReference `json:"transactionReference"`
				IsBattleVote         bool                        `json:"isBattleVote"`
			}{
				FromUserID:           userID,
				Amount:               supsChange,
				TransactionReference: supTxRefString,
				IsBattleVote:         true,
			},
			TransactionID: id,
		}}
	for {
		select {
		case <-replyChannel:
			pp.Log.Info().Msg("ENDING SUP HOLD")
			return supTxRefString, nil
		case err := <-errChan:
			pp.Log.Info().Msg("ENDING SUP HOLD ERR")
			return supTxRefString, terror.Error(err)
		}
	}
}

// ReleaseTransactions tells the passport to transfer fund to sup pool
func (pp *Passport) ReleaseTransactions(ctx context.Context, transactions []server.TransactionReference) {
	if len(transactions) == 0 {
		return
	}

	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:RELEASE_TRANSACTIONS",
			Payload: struct {
				TransactionReferences []server.TransactionReference `json:"transactionReferences"`
			}{
				TransactionReferences: transactions,
			},
		},
	}
}

// TransferBattleFundToSupsPool tells the passport to transfer fund to sup pool
func (pp *Passport) TransferBattleFundToSupsPool(ctx context.Context) error {
	replyChannel := make(chan []byte)
	errChan := make(chan error)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key:           "SUPREMACY:TRANSFER_BATTLE_FUND_TO_SUP_POOL",
			TransactionID: uuid.Must(uuid.NewV4()).String(),
		}}

	for {
		select {
		case <-replyChannel:
			return nil
		case err := <-errChan:
			return terror.Error(err)
		}
	}
}

type SupremacyTopSupsContributorResponse struct {
	Payload SupremacyTopSupsContributor `json:"payload"`
}

type SupremacyTopSupsContributor struct {
	TopSupsContributors       []*server.User    `json:"topSupsContributors"`
	TopSupsContributeFactions []*server.Faction `json:"topSupsContributeFactions"`
}

// TopSupsContributorsGet tells the passport to return the top three most sups contributors with in the time frame
func (pp *Passport) TopSupsContributorsGet(ctx context.Context, startTime, endTime time.Time) ([]*server.User, []*server.Faction, error) {
	replyChannel := make(chan []byte)
	errChan := make(chan error)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:TOP_SUPS_CONTRIBUTORS",
			Payload: struct {
				StartTime time.Time `json:"startTime"`
				EndTime   time.Time `json:"endTime"`
			}{
				StartTime: startTime,
				EndTime:   endTime,
			},
			TransactionID: uuid.Must(uuid.NewV4()).String(),
		}}

	for {
		select {
		case msg := <-replyChannel:
			resp := &SupremacyTopSupsContributorResponse{}
			err := json.Unmarshal(msg, resp)
			if err != nil {
				return nil, nil, terror.Error(err)
			}
			return resp.Payload.TopSupsContributors, resp.Payload.TopSupsContributeFactions, nil
		case err := <-errChan:
			return nil, nil, terror.Error(err)
		}
	}
}
