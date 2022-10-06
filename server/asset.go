package server

import (
	"encoding/json"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type CollectionItem struct {
	CollectionSlug      string      `json:"collection_slug"`
	Hash                string      `json:"hash"`
	TokenID             int64       `json:"token_id"`
	ItemType            string      `json:"item_type"`
	ItemID              string      `json:"item_id"`
	Tier                string      `json:"tier"`
	OwnerID             string      `json:"owner_id"`
	MarketLocked        bool        `json:"market_locked"`
	XsynLocked          bool        `json:"xsyn_locked"`
	LockedToMarketplace bool        `json:"locked_to_marketplace"`
	AssetHidden         null.String `json:"asset_hidden"`
}

type EquippedOnDetails struct {
	ID    string `json:"id"`
	Hash  string `json:"hash"`
	Label string `json:"label"`
	Name  string `json:"name"`
}

// AssetKeycard is a keycard asset struct
type AssetKeycard struct {
	ID                 string            `json:"id" boil:"player_keycards.id"`
	PlayerID           string            `json:"player_id" boil:"player_keycards.player_id"`
	BlueprintKeycardID string            `json:"blueprint_keycard_id" boil:"player_keycards.blueprint_keycard_id"`
	Count              int64             `json:"count" boil:"player_keycards.count"`
	CreatedAt          time.Time         `json:"created_at" boil:"player_keycards.created_at"`
	MarketListedCount  int64             `json:"market_listed_count" boil:"market_listed_count"`
	ItemSaleIDs        types.StringArray `json:"item_sale_ids" boil:"item_sale_ids"`

	Blueprints AssetKeycardBlueprint `json:"blueprints" boil:"blueprint_keycards,bind"`
}

type AssetKeycardBlueprint struct {
	ID             string      `boil:"blueprint_keycards.id" json:"id"`
	Label          string      `boil:"blueprint_keycards.label" json:"label"`
	Description    string      `boil:"blueprint_keycards.description" json:"description"`
	Collection     string      `boil:"blueprint_keycards.collection" json:"collection"`
	KeycardTokenID int         `boil:"blueprint_keycards.keycard_token_id" json:"keycard_token_id"`
	ImageURL       string      `boil:"blueprint_keycards.image_url" json:"image_url"`
	AnimationURL   null.String `boil:"blueprint_keycards.animation_url" json:"animation_url"`
	KeycardGroup   string      `boil:"blueprint_keycards.keycard_group" json:"keycard_group"`
	Syndicate      null.String `boil:"blueprint_keycards.syndicate" json:"syndicate"`
	CreatedAt      time.Time   `boil:"blueprint_keycards.created_at" json:"created_at"`
}

func (b AssetKeycardBlueprint) MarshalJSON() ([]byte, error) {
	if b.ID == "" {
		return null.NullBytes, nil
	}
	type localAssetKeycardBlueprint AssetKeycardBlueprint
	return json.Marshal(localAssetKeycardBlueprint(b))
}

type RepairOffer struct {
	*boiler.RepairOffer
	BlocksRequiredRepair int             `db:"blocks_required_repair" json:"blocks_required_repair"`
	BlocksRepaired       int             `db:"blocks_repaired" json:"blocks_repaired"`
	SupsWorthPerBlock    decimal.Decimal `db:"sups_worth_per_block" json:"sups_worth_per_block"`
	WorkingAgentCount    int             `db:"working_agent_count" json:"working_agent_count"`
	JobOwner             *PublicPlayer   `json:"job_owner"`
}
