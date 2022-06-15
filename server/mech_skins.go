package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

type MechSkin struct {
	*CollectionItem
	ID                    string      `json:"id"`
	BlueprintID           string      `json:"blueprint_id"`
	GenesisTokenID        null.Int64  `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64  `json:"limited_release_token_id,omitempty"`
	Label                 string      `json:"label"`
	MechModel             string      `json:"mech_model"`
	EquippedOn            null.String `json:"equipped_on,omitempty"`
	ImageURL              null.String `json:"image_url,omitempty"`
	AnimationURL          null.String `json:"animation_url,omitempty"`
	CardAnimationURL      null.String `json:"card_animation_url,omitempty"`
	AvatarURL             null.String `json:"avatar_url,omitempty"`
	LargeImageURL         null.String `json:"large_image_url,omitempty"`
	CreatedAt             time.Time   `json:"created_at"`
}

func (b *MechSkin) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintMechSkin struct {
	ID               string      `json:"id"`
	Collection       string      `json:"collection"`
	MechModel        string      `json:"mech_model"`
	Label            string      `json:"label"`
	Tier             string      `json:"tier,omitempty"`
	ImageURL         null.String `json:"image_url,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	BackgroundColor  null.String `json:"background_color,omitempty"`
	YoutubeURL       null.String `json:"youtube_url,omitempty"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

func (b *BlueprintMechSkin) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintMechSkinFromBoiler(mechSkin *boiler.BlueprintMechSkin) *BlueprintMechSkin {
	return &BlueprintMechSkin{
		ID:               mechSkin.ID,
		Collection:       mechSkin.Collection,
		MechModel:        mechSkin.MechModel,
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

func MechSkinFromBoiler(skin *boiler.MechSkin, collection *boiler.CollectionItem) *MechSkin {
	return &MechSkin{
		CollectionItem: &CollectionItem{
			CollectionSlug:   collection.CollectionSlug,
			Hash:             collection.Hash,
			TokenID:          collection.TokenID,
			ItemType:         collection.ItemType,
			ItemID:           collection.ItemID,
			Tier:             collection.Tier,
			OwnerID:          collection.OwnerID,
			MarketLocked:     collection.MarketLocked,
			XsynLocked:       collection.XsynLocked,
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:               skin.ID,
		BlueprintID:      skin.BlueprintID,
		GenesisTokenID:   skin.GenesisTokenID,
		Label:            skin.Label,
		MechModel:        skin.MechModel,
		EquippedOn:       skin.EquippedOn,
		ImageURL:         skin.ImageURL,
		AnimationURL:     skin.AnimationURL,
		CardAnimationURL: skin.CardAnimationURL,
		AvatarURL:        skin.AvatarURL,
		LargeImageURL:    skin.LargeImageURL,
		CreatedAt:        skin.CreatedAt,
	}
}
