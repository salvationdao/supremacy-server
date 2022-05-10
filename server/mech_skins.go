package server

import (
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type MechSkin struct {
	*CollectionDetails
	ID               string              `json:"id"`
	BlueprintID      string              `json:"blueprint_id"`
	CollectionItemID string              `json:"collection_item_id"`
	GenesisTokenID   decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	Label            string              `json:"label"`
	OwnerID          string              `json:"owner_id"`
	ChassisModel     string              `json:"chassis_model"`
	EquippedOn       null.String         `json:"equipped_on,omitempty"`
	Tier             null.String         `json:"tier,omitempty"`
	ImageURL         null.String         `json:"image_url,omitempty"`
	AnimationURL     null.String         `json:"animation_url,omitempty"`
	CardAnimationURL null.String         `json:"card_animation_url,omitempty"`
	AvatarURL        null.String         `json:"avatar_url,omitempty"`
	LargeImageURL    null.String         `json:"large_image_url,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
}

type BlueprintMechSkin struct {
	ID               string      `json:"id"`
	Collection       string      `json:"collection"`
	ChassisModel     string      `json:"chassis_model"`
	Label            string      `json:"label"`
	Tier             null.String `json:"tier,omitempty"`
	ImageURL         null.String `json:"image_url,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

func BlueprintMechSkinFromBoiler(mechSkin *boiler.BlueprintMechSkin) *BlueprintMechSkin {
	return &BlueprintMechSkin{
		ID:               mechSkin.ID,
		Collection:       mechSkin.Collection,
		ChassisModel:     mechSkin.ChassisModel,
		Label:            mechSkin.Label,
		Tier:             mechSkin.Tier,
		ImageURL:         mechSkin.ImageURL,
		AnimationURL:     mechSkin.AnimationURL,
		CardAnimationURL: mechSkin.CardAnimationURL,
		LargeImageURL:    mechSkin.LargeImageURL,
		AvatarURL:        mechSkin.AvatarURL,
		CreatedAt:        mechSkin.CreatedAt,
	}
}
