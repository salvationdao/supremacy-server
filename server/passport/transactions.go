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
