package comms

import (
	"server"
	"time"

	"github.com/ninja-syndicate/hub"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type AssetRepairStatReq struct {
	AssetRepairRecord *server.AssetRepairRecord `json:"assetRepairRecord"`
}

type AssetRepairStatResp struct{}
type DefaultWarMachinesReq struct {
	FactionID server.FactionID `json:"factionID"`
}

type DefaultWarMachinesResp struct {
	WarMachines []*server.WarMachineMetadata `json:"warMachines"`
}
type WarMachineQueuePositionReq struct {
	WarMachineQueuePosition []*WarMachineQueueStat `json:"warMachineQueuePosition"`
}
type WarMachineQueueStat struct {
	Hash           string          `json:"hash"`
	Position       *int            `json:"position,omitempty"`
	ContractReward decimal.Decimal `json:"contractReward,omitempty"`
}

type WarMachineQueuePositionResp struct{}

type UserConnectionUpgradeReq struct {
	SessionID hub.SessionID `json:"sessionID"`
}

type UserConnectionUpgradeResp struct{}
type FactionAllReq struct{}

type FactionAllResp struct {
	Factions []*server.Faction `json:"factions"`
}
type SpendSupsReq struct {
	Amount               string                      `json:"amount"`
	FromUserID           server.UserID               `json:"fromUserID"`
	ToUserID             *server.UserID              `json:"toUserID,omitempty"`
	TransactionReference server.TransactionReference `json:"transactionReference"`
	GroupID              TransactionGroup            `json:"groupID,omitempty"`
}

type TransactionGroup string
type SpendSupsResp struct {
	TXID string `json:"txid"`
}
type ReleaseTransactionsReq struct {
	TxIDs []string `json:"txIDs"`
}
type ReleaseTransactionsResp struct{}
type TickerTickReq struct {
	UserMap map[int][]server.UserID `json:"userMap"`
}
type TickerTickResp struct{}

type GetSpoilOfWarReq struct{}
type GetSpoilOfWarResp struct {
	Amount string
}
type UserSupsMultiplierSendReq struct {
	UserSupsMultiplierSends []*server.UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
}

type UserSupsMultiplierSendResp struct{}
type TransferBattleFundToSupPoolReq struct{}
type TransferBattleFundToSupPoolResp struct{}
type TopSupsContributorReq struct {
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
