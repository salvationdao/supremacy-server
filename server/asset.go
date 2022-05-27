package server

import (
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

// Asset is a generic Asset struct, used for xsyn
type Asset struct {
	ID             string     `json:"id"`
	CollectionSlug string     `json:"collection_id"`
	TokenID        int64      `json:"external_token_id"`
	Tier           string     `json:"tier"`
	Hash           string     `json:"hash"`
	OwnerID        string     `json:"owner_id"`
	Data           types.JSON `json:"data"`
	OnChainStatus  string     `json:"on_chain_status"`
}

// AssetKeycard is a keycard asset struct
type AssetKeycard struct {
	ID                 string    `json:"id" boil:"id"`
	PlayerID           string    `json:"player_id" boil:"player_id"`
	BlueprintKeycardID string    `json:"blueprint_keycard_id" boil:"blueprint_keycard_id"`
	Count              int       `json:"count" boil:"count"`
	CreatedAt          time.Time `json:"created_at" boil:"created_at"`

	Blueprints struct {
		ID             string      `boil:"id" json:"id"`
		Label          string      `boil:"label" json:"label"`
		Description    string      `boil:"description" json:"description"`
		Collection     string      `boil:"collection" json:"collection"`
		KeycardTokenID int         `boil:"keycard_token_id" json:"keycard_token_id"`
		ImageURL       string      `boil:"image_url" json:"image_url"`
		AnimationURL   null.String `boil:"animation_url" json:"animation_url"`
		KeycardGroup   string      `boil:"keycard_group" json:"keycard_group"`
		Syndicate      null.String `boil:"syndicate" json:"syndicate"`
		CreatedAt      time.Time   `boil:"created_at" json:"created_at"`
	} `json:"blueprints" boil:"BlueprintKeycard,blueprints"`
}
