package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
)

// TRANSACTION:CHECK_LIST

type CheckTransactionsResponse struct {
	Payload struct {
		Transactions []*server.Transaction `json:"transactions"`
	} `json:"payload"`
}

func (pp *Passport) CheckTransactions(ctx context.Context, transactions []server.TransactionReference) ([]*server.Transaction, error) {
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
			Key: "TRANSACTION:CHECK_LIST",
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
	resp := &CheckTransactionsResponse{}
	err = json.Unmarshal(msg, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	return resp.Payload.Transactions, nil
}
