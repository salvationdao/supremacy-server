package server

import (
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type PowerCore struct {
	*CollectionDetails
	ID           string          `json:"id"`
	Label        string          `json:"label"`
	Size         string          `json:"size"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	EquippedOn   null.String     `json:"equipped_on,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type BlueprintPowerCore struct {
	ID           string          `json:"id"`
	Collection   string          `json:"collection"`
	Label        string          `json:"label"`
	Size         string          `json:"size"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	Tier         string          `json:"tier,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID decimal.NullDecimal `json:"limited_release_token_id,omitempty"`
}

func BlueprintPowerCoreFromBoiler(core *boiler.BlueprintPowerCore) *BlueprintPowerCore {
	return &BlueprintPowerCore{
		ID:           core.ID,
		Collection:   core.Collection,
		Label:        core.Label,
		Size:         core.Size,
		Capacity:     core.Capacity,
		MaxDrawRate:  core.MaxDrawRate,
		RechargeRate: core.RechargeRate,
		Armour:       core.Armour,
		MaxHitpoints: core.MaxHitpoints,
		Tier:         core.Tier,
		CreatedAt:    core.CreatedAt,
	}
}

func PowerCoreFromBoiler(skin *boiler.PowerCore, collection *boiler.CollectionItem) *PowerCore {
	return &PowerCore{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: collection.CollectionSlug,
			Hash:           collection.Hash,
			TokenID:        collection.TokenID,
			ItemType:       collection.ItemType,
			ItemID:         collection.ItemID,
			Tier:           collection.Tier,
			OwnerID:        collection.OwnerID,
			OnChainStatus:  collection.OnChainStatus,
		},
		ID:           skin.ID,
		Label:        skin.Label,
		Size:         skin.Size,
		Capacity:     skin.Capacity,
		MaxDrawRate:  skin.MaxDrawRate,
		RechargeRate: skin.RechargeRate,
		Armour:       skin.Armour,
		MaxHitpoints: skin.MaxHitpoints,
		EquippedOn:   skin.EquippedOn,
		CreatedAt:    skin.CreatedAt,
	}
}
