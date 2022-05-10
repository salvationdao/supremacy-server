package server

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type EnergyCore struct {
	*CollectionDetails
	ID               string          `json:"id"`
	CollectionItemID string          `json:"collection_item_id"`
	OwnerID          string          `json:"owner_id"`
	Label            string          `json:"label"`
	Size             string          `json:"size"`
	Capacity         decimal.Decimal `json:"capacity"`
	MaxDrawRate      decimal.Decimal `json:"max_draw_rate"`
	RechargeRate     decimal.Decimal `json:"recharge_rate"`
	Armour           decimal.Decimal `json:"armour"`
	MaxHitpoints     decimal.Decimal `json:"max_hitpoints"`
	Tier             null.String     `json:"tier,omitempty"`
	EquippedOn       null.String     `json:"equipped_on,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

type BlueprintEnergyCore struct {
	ID           string          `json:"id"`
	Collection   string          `json:"collection"`
	Label        string          `json:"label"`
	Size         string          `json:"size"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	Tier         null.String     `json:"tier,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}
