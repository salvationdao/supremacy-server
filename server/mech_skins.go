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
	*Images
	SkinSwatch            *Images
	ID                    string      `json:"id"`
	BlueprintID           string      `json:"blueprint_id"`
	GenesisTokenID        null.Int64  `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64  `json:"limited_release_token_id,omitempty"`
	Label                 string      `json:"label"`
	Level                 int         `json:"level"`
	EquippedOn            null.String `json:"equipped_on,omitempty"`
	LockedToMech          bool        `json:"locked_to_mech"`
	CreatedAt             time.Time   `json:"created_at"`

	EquippedOnDetails *EquippedOnDetails
}

func (b *MechSkin) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintMechSkin struct {
	ID           string    `json:"id"`
	Collection   string    `json:"collection"`
	Label        string    `json:"label"`
	Tier         string    `json:"tier,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	DefaultLevel int       `json:"default_level"`

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
		ID:           mechSkin.ID,
		Collection:   mechSkin.Collection,
		Label:        mechSkin.Label,
		Tier:         mechSkin.Tier,
		DefaultLevel: mechSkin.DefaultLevel,
		CreatedAt:    mechSkin.CreatedAt,
	}
}

func MechSkinFromBoiler(skin *boiler.MechSkin, collection *boiler.CollectionItem, skinDetails *boiler.MechModelSkinCompatibility, blueprintMechSkinDetails *boiler.BlueprintMechSkin) *MechSkin {
	mskin := &MechSkin{
		CollectionItem: &CollectionItem{
			CollectionSlug: collection.CollectionSlug,
			Hash:           collection.Hash,
			TokenID:        collection.TokenID,
			ItemType:       collection.ItemType,
			ItemID:         collection.ItemID,
			Tier:           collection.Tier,
			OwnerID:        collection.OwnerID,
			MarketLocked:   collection.MarketLocked,
			XsynLocked:     collection.XsynLocked,
			AssetHidden:    collection.AssetHidden,
		},
		Images: &Images{
			ImageURL:         skinDetails.ImageURL,
			CardAnimationURL: skinDetails.CardAnimationURL,
			AvatarURL:        skinDetails.AvatarURL,
			LargeImageURL:    skinDetails.LargeImageURL,
			BackgroundColor:  skinDetails.BackgroundColor,
			AnimationURL:     skinDetails.AnimationURL,
			YoutubeURL:       skinDetails.YoutubeURL,
		},
		SkinSwatch: &Images{
			ImageURL:         blueprintMechSkinDetails.ImageURL,
			CardAnimationURL: blueprintMechSkinDetails.CardAnimationURL,
			AvatarURL:        blueprintMechSkinDetails.AvatarURL,
			LargeImageURL:    blueprintMechSkinDetails.LargeImageURL,
			BackgroundColor:  blueprintMechSkinDetails.BackgroundColor,
			AnimationURL:     blueprintMechSkinDetails.AnimationURL,
			YoutubeURL:       blueprintMechSkinDetails.YoutubeURL,
		},
		Label:          skin.R.Blueprint.Label,
		ID:             skin.ID,
		BlueprintID:    skin.BlueprintID,
		GenesisTokenID: skin.GenesisTokenID,
		EquippedOn:     skin.EquippedOn,
		CreatedAt:      skin.CreatedAt,
	}

	return mskin
}
