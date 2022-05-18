package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type MechAnimation struct {
	*CollectionDetails
	ID             string      `json:"id"`
	BlueprintID    string      `json:"blueprint_id"`
	Label          string      `json:"label"`
	MechModel      string      `json:"mech_model"`
	EquippedOn     null.String `json:"equipped_on,omitempty"`
	IntroAnimation null.Bool   `json:"intro_animation,omitempty"`
	OutroAnimation null.Bool   `json:"outro_animation,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

func (b *MechAnimation) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintMechAnimation struct {
	ID             string    `json:"id"`
	Collection     string    `json:"collection"`
	Label          string    `json:"label"`
	MechModel      string    `json:"mech_model"`
	Tier           string    `json:"tier,omitempty"`
	IntroAnimation null.Bool `json:"intro_animation,omitempty"`
	OutroAnimation null.Bool `json:"outro_animation,omitempty"`
	CreatedAt      time.Time `json:"created_at"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID decimal.NullDecimal `json:"limited_release_token_id,omitempty"`
}

func (b *BlueprintMechAnimation) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintMechAnimationFromBoiler(animation *boiler.BlueprintMechAnimation) *BlueprintMechAnimation {
	return &BlueprintMechAnimation{
		ID:             animation.ID,
		Collection:     animation.Collection,
		Label:          animation.Label,
		MechModel:      animation.MechModel,
		Tier:           animation.Tier,
		IntroAnimation: animation.IntroAnimation,
		OutroAnimation: animation.OutroAnimation,
		CreatedAt:      animation.CreatedAt,
	}
}

func MechAnimationFromBoiler(animation *boiler.MechAnimation, collection *boiler.CollectionItem) *MechAnimation {
	return &MechAnimation{
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
		ID:             animation.ID,
		BlueprintID:    animation.BlueprintID,
		Label:          animation.Label,
		MechModel:      animation.MechModel,
		EquippedOn:     animation.EquippedOn,
		IntroAnimation: animation.IntroAnimation,
		OutroAnimation: animation.OutroAnimation,
		CreatedAt:      animation.CreatedAt,
	}
}
