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
	ID                 string    `json:"id" boil:"player_keycards.id"`
	PlayerID           string    `json:"player_id" boil:"player_keycards.player_id"`
	BlueprintKeycardID string    `json:"blueprint_keycard_id" boil:"player_keycards.blueprint_keycard_id"`
	Count              int       `json:"count" boil:"player_keycards.count"`
	CreatedAt          time.Time `json:"created_at" boil:"player_keycards.created_at"`

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
