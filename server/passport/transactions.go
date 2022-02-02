package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"server"

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

	replyChannel := make(chan []byte)
	errChan := make(chan error)

	txID, err := uuid.NewV4()
	if err != nil {
		return nil, terror.Error(err)
	}
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
			TransactionID: txID.String(),
			context:       ctx,
		}}

	for {
		select {
		case msg := <-replyChannel:
			resp := &CommitTransactionsResponse{}
			err = json.Unmarshal(msg, resp)
			if err != nil {
				return nil, terror.Error(err)
			}
			return resp.Transactions, nil
		case err := <-errChan:
			return nil, terror.Error(err)
		}
	}
}

// SendHoldSupsMessage tells the passport to hold sups
func (pp *Passport) SendHoldSupsMessage(ctx context.Context, userID server.UserID, supsChange server.BigInt, txID string, reason string) (server.TransactionReference, error) {
	supTransactionReference := uuid.Must(uuid.NewV4())
	supTxRefString := server.TransactionReference(fmt.Sprintf("%s|%s", reason, supTransactionReference.String()))
	replyChannel := make(chan []byte)
	errChan := make(chan error)

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
			TransactionID: txID,
			context:       ctx,
		}}
	for {
		select {
		case <-replyChannel:
			return supTxRefString, nil
		case err := <-errChan:
			return supTxRefString, terror.Error(err)
		}
	}
}

type DistributeBattleRewardRequest struct {
	WinnerFactionID               server.FactionID `json:"winnerFactionID"`
	WinningFactionViewerIDs       []server.UserID  `json:"winningFactionViewerIDs"`
	WinningWarMachineOwnerIDs     []server.UserID  `json:"winningWarMachineOwnerIDs"`
	ExecuteKillWarMachineOwnerIDs []server.UserID  `json:"executeKillWarMachineOwnerIDs"`
}

// DistributeBattleReward tells the passport to distribute battle reward
func (pp *Passport) DistributeBattleReward(ctx context.Context, battleReward *DistributeBattleRewardRequest, txID string) error {
	replyChannel := make(chan []byte)
	errChan := make(chan error)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key:           "SUPREMACY:DISTRIBUTE_BATTLE_REWARD",
			Payload:       battleReward,
			TransactionID: txID,
			context:       ctx,
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
