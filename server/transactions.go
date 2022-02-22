package server

import (
	"time"
)

type TransactionReference string

type TransactionStatus string

const (
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID                   string               `json:"id"`
	ToID                 UserID               `json:"toID"`
	FromID               UserID               `json:"fromID"`
	Amount               BigInt               `json:"amount"`
	Status               TransactionStatus    `json:"status"`
	TransactionReference TransactionReference `json:"transactionReference"`
	Reason               string               `json:"reason"`
	Description          string               `json:"description"`
	CreatedAt            time.Time            `json:"created_at"`
}
