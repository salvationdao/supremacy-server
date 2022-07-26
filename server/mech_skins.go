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
	MechModelID           string      `json:"mech_model_id"`
	MechModelName         string      `json:"mech_model_name"`
	EquippedOn            null.String `json:"equipped_on,omitempty"`
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
	ID               string      `json:"id"`
	Collection       string      `json:"collection"`
	Label            string      `json:"label"`
	Tier             string      `json:"tier,omitempty"`
	// TODO: vinnie add
	//ImageURL         null.String `json:"image_url,omitempty"`
	//AnimationURL     null.String `json:"animation_url,omitempty"`
	//CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	//LargeImageURL    null.String `json:"large_image_url,omitempty"`
	//AvatarURL        null.String `json:"avatar_url,omitempty"`
	//BackgroundColor  null.String `json:"background_color,omitempty"`
	//YoutubeURL       null.String `json:"youtube_url,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`

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
		Label:            mechSkin.Label,
		Tier:             mechSkin.Tier,
		CreatedAt:        mechSkin.CreatedAt,
	}
}

func MechSkinFromBoiler(skin *boiler.MechSkin, collection *boiler.CollectionItem) *MechSkin {
	mskin := &MechSkin{
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
			AssetHidden:      collection.AssetHidden,
		},
		ID:               skin.ID,
		BlueprintID:      skin.BlueprintID,
		GenesisTokenID:   skin.GenesisTokenID,
		Label:            skin.Label,
		MechModelID:      skin.MechModel,
		EquippedOn:       skin.EquippedOn,
		CreatedAt:        skin.CreatedAt,
	}

	if skin.R != nil && skin.R.MechSkinMechModel != nil {
		mskin.MechModelName = skin.R.MechSkinMechModel.Label
	}

	return mskin
}
