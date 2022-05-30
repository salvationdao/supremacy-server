package xsyn_rpcclient

import (
	"server"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type DefaultWarMachinesResp struct {
	WarMachines []*server.WarMachineMetadata `json:"warMachines"`
}
type WarMachineQueuePositionReq struct {
	ApiKey                  string
	WarMachineQueuePosition []*WarMachineQueueStat `json:"warMachineQueuePosition"`
}
type WarMachineQueueStat struct {
	Hash           string          `json:"hash"`
	Position       *int            `json:"position,omitempty"`
	ContractReward decimal.Decimal `json:"contractReward,omitempty"`
}

type WarMachineQueuePositionResp struct{}

type FactionAllReq struct{}

type FactionAllResp struct {
	Factions []*server.Faction `json:"factions"`
}
type SpendSupsReq struct {
	ApiKey               string                      `json:"apiKey"`
	Amount               string                      `json:"amount"`
	FromUserID           uuid.UUID                   `json:"fromUserID"`
	ToUserID             uuid.UUID                   `json:"toUserID"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
	Group                string                      `json:"group,omitempty"`
	SubGroup             string                      `json:"subGroup"`
	Description          string                      `json:"description"`

	NotSafe bool `json:"notSafe"`
}

type RefundTransactionReq struct {
	ApiKey        string
	TransactionID string `json:"transaction_id"`
}

type RefundTransactionResp struct {
	TransactionID string `json:"transaction_id"`
}
type TransactionGroup string

type SpendSupsResp struct {
	TransactionID string `json:"transaction_id"`
}

type ReleaseTransactionsReq struct {
	ApiKey string
	TxIDs  []string `json:"txIDs"`
}
type ReleaseTransactionsResp struct{}
type TickerTickReq struct {
	ApiKey  string
	UserMap map[int][]server.UserID `json:"userMap"`
}
type TickerTickResp struct{}

type GetSpoilOfWarReq struct{}
type GetSpoilOfWarResp struct {
	Amount string
}
type UserSupsMultiplierSendReq struct {
	ApiKey                  string
	UserSupsMultiplierSends []*server.UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
}

type UserSupsMultiplierSendResp struct{}
type TransferBattleFundToSupPoolReq struct{}
type TransferBattleFundToSupPoolResp struct{}
type TopSupsContributorReq struct {
	ApiKey    string    `json:"apiKey"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type TopSupsContributorResp struct {
	TopSupsContributors       []*server.User    `json:"topSupsContributors"`
	TopSupsContributeFactions []*server.Faction `json:"topSupsContributeFactions"`
}
type GetAllReq struct{}
type GetAll struct {
	AssetPayload    types.JSON
	MetadataPayload types.JSON
	StorePayload    types.JSON
	UserPayload     types.JSON
	FactionPayload  types.JSON
}

type AssetUnlockToServiceResp struct {
}

type AssetUnlockToServiceReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

type AssetLockToServiceResp struct {
}

type AssetLockToServiceReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}
