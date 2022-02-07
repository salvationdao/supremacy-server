package server

import "github.com/gofrs/uuid"

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

var RedMountainFactionID = FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060")))
var BostonCyberneticsFactionID = FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2")))
var ZaibatsuFactionID = FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d")))

type Faction struct {
	ID            FactionID     `json:"id" db:"id"`
	Label         string        `json:"label" db:"label"`
	Theme         *FactionTheme `json:"theme" db:"theme"`
	LogoUrl       string        `json:"logoUrl,omitempty"`
	BackgroundUrl string        `json:"backgroundUrl,omitempty"`
}

type BattleAbility struct {
	ID                     BattleAbilityID `json:"id" db:"id"`
	Label                  string          `json:"label" db:"label"`
	Colour                 string          `json:"colour" db:"colour"`
	ImageUrl               string          `json:"imageUrl" db:"image_url"`
	CooldownDurationSecond int             `json:"cooldownDurationSecond" db:"cooldown_duration_second"`
}

type FactionAbilityType string

const (
	FactionAbilityTypeAirStrike FactionAbilityType = "AIRSTRIKE"
	FactionAbilityTypeNuke      FactionAbilityType = "NUKE"
	FactionAbilityTypeHealing   FactionAbilityType = "HEALING"
)

type FactionAbility struct {
	ID                  FactionAbilityID `json:"id" db:"id"`
	GameClientAbilityID byte             `json:"gameClientAbilityID" db:"game_client_ability_id"`
	BattleAbilityID     BattleAbilityID  `json:"collectionID" db:"battle_ability_id"`
	FactionID           FactionID        `json:"factionID" db:"faction_id"`
	Label               string           `json:"label" db:"label"`
	USDCentCost         int              `json:"usdCentCost" db:"usd_cent_cost"`
	SupsCost            BigInt           `json:"supsCost"`
}
