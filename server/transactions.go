package server

import (
	"time"

	"github.com/gofrs/uuid"
)

type TransactionReference string

type TransactionStatus string

const (
	TransactionPending TransactionStatus = "pending"
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID                   uuid.UUID            `json:"id"`
	FromID               UserID               `json:"fromId"`
	ToID                 UserID               `json:"toId"`
	Amount               BigInt               `json:"amount"`
	Status               TransactionStatus    `json:"status"`
	TransactionReference TransactionReference `json:"transactionReference"`
	CreatedAt            time.Time            `json:"created_at"`
}
