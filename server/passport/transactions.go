package passport

import (
	"context"
	"encoding/json"
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

	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	txID, err := uuid.NewV4()
	if err != nil {
		cancel()
		return nil, terror.Error(err)
	}
	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "SUPREMACY:COMMIT_TRANSACTIONS",
			Payload: struct {
				TransactionReferences []server.TransactionReference `json:"transactionReferences"`
			}{
				TransactionReferences: transactions,
			},
			TransactionId: txID.String(),
			context:       ctx,
			cancel:        cancel,
		}}

	msg := <-replyChannel
	resp := &CommitTransactionsResponse{}
	err = json.Unmarshal(msg, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	return resp.Transactions, nil
}
