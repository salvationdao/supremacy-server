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
	ToID                 UserID               `json:"credit"`
	FromID               UserID               `json:"debit"`
	Amount               BigInt               `json:"amount"`
	Status               TransactionStatus    `json:"status"`
	TransactionReference TransactionReference `json:"transaction_reference"`
	Reason               string               `json:"reason"`
	Description          string               `json:"description"`
	CreatedAt            time.Time            `json:"created_at"`
}

type TransactionGroup string

const (
	TransactionGroupStore           TransactionGroup = "STORE"
	TransactionGroupDeposit         TransactionGroup = "DEPOSIT"
	TransactionGroupWithdrawal      TransactionGroup = "WITHDRAWAL"
	TransactionGroupBattle          TransactionGroup = "BATTLE"
	TransactionGroupSupremacy       TransactionGroup = "SUPREMACY"
	TransactionGroupAssetManagement TransactionGroup = "ASSET MANAGEMENT"
)
