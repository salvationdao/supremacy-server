package rpctypes

import (
	"fmt"
	"server/gamelog"
	"time"

	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
)

/*
	THIS FILE SHOULD CONTAIN ZERO BOILER STRUCTS
	These are the objects things using this rpc server expect and a migration change shouldn't break external services!

	We should have convert functions on our objects that convert them to our api objects, for example
	apiMech := server.Mech.ToApiMechV1()
*/

type AssetReq struct {
	AssetHash string `json:"asset_hash"`
}

type AssetResp struct {
	Asset *XsynAsset `json:"asset"`
}

type TemplateRegisterReq struct {
	TemplateID uuid.UUID
	OwnerID    uuid.UUID
}
type TemplateRegisterResp struct {
	Assets []*XsynAsset
}


type Attributes []*Attribute

func (a Attributes) AreValid() error {
	errCount := 0
	for _, val := range a {
		if val.DisplayType != "" {
			_, intOk := val.Value.(int)
			_, floatOK := val.Value.(float32)
			_, float64OK := val.Value.(float64)
			if !intOk && !floatOK && !float64OK {
				gamelog.L.Error().Err(fmt.Errorf("invalid attribute value %v for display type %s", val.Value, val.DisplayType)).Msg("invalid value in metadata")
				errCount++
			}
		}
	}
	if errCount > 0 {
		return fmt.Errorf("attributes are invalid")
	}

	return nil
}

type Attribute struct {
	DisplayType DisplayType `json:"display_type,omitempty"`
	TraitType   string      `json:"trait_type"`
	AssetHash   string      `json:"asset_hash,omitempty"`
	Value       interface{} `json:"value"` // string or number only
}

type DisplayType string

const (
	BoostNumber     DisplayType = "boost_number"
	BoostPercentage DisplayType = "boost_percentage"
	Number          DisplayType = "number"
	Date            DisplayType = "date"
)

type XsynAsset struct {
	ID               string      `json:"id,omitempty"`
	CollectionSlug   string      `json:"collection_slug,omitempty"`
	TokenID          int64       `json:"token_id,omitempty"`
	Tier             string      `json:"tier,omitempty"`
	Hash             string      `json:"hash,omitempty"`
	OwnerID          string      `json:"owner_id,omitempty"`
	Data             types.JSON  `json:"data,omitempty"`
	Attributes       Attributes  `json:"attributes,omitempty"`
	Name             string      `json:"name,omitempty"`
	AssetType        null.String `json:"asset_type,omitempty"`
	ImageURL         null.String `json:"image_url,omitempty"`
	ExternalURL      null.String `json:"external_url,omitempty"`
	Description      null.String `json:"description,omitempty"`
	BackgroundColor  null.String `json:"background_color,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	YoutubeURL       null.String `json:"youtube_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	UnlockedAt       time.Time   `json:"unlocked_at,omitempty"`
	MintedAt         null.Time   `json:"minted_at,omitempty"`
	OnChainStatus    string      `json:"on_chain_status,omitempty"`
	Service          string                 `json:"xsyn_locked"`
}
