package xsyn_rpcclient

import (
	"server"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type SpendSupsReq struct {
	ApiKey               string                      `json:"apiKey"`
	Amount               string                      `json:"amount"`
	FromUserID           uuid.UUID                   `json:"fromUserID"`
	ToUserID             uuid.UUID                   `json:"toUserID"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
	Group                string                      `json:"group,omitempty"`
	SubGroup             string                      `json:"subGroup"`
	Description          string                      `json:"description"`
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

type TransferBattleFundToSupPoolReq struct{}
type TransferBattleFundToSupPoolResp struct{}
type TopSupsContributorReq struct {
	ApiKey    string    `json:"apiKey"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
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

type SyndicateCreateReq struct {
	ApiKey      string `json:"api_key"`
	SyndicateID string `json:"syndicate_id"`
	FoundedByID string `json:"founded_by_id"`
	Name        string `json:"name"`
}
type SyndicateCreateResp struct{}

type SyndicateNameCreateReq struct {
	ApiKey      string `json:"api_key"`
	SyndicateID string `json:"syndicate_id"`
	Name        string `json:"name"`
}
type SyndicateNameChangeResp struct{}

type SyndicateLiquidateReq struct {
	ApiKey        string   `json:"api_key"`
	SyndicateID   string   `json:"syndicate_id"`
	RemainUserIDs []string `json:"remain_user_ids"`
}
type SyndicateLiquidateResp struct{}

type GetCurrentSupPriceReq struct{}

type GetCurrentSupPriceResp struct {
	PriceUSD decimal.Decimal `json:"price_usd"`
}

type GetExchangeRatesReq struct{}

type GetExchangeRatesResp struct {
	SUPtoUSD decimal.Decimal `json:"sup_to_usd"`
	ETHtoUSD decimal.Decimal `json:"eth_to_usd"`
	BNBtoUSD decimal.Decimal `json:"bnb_to_usd"`
}
