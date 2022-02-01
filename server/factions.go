package server

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

type Faction struct {
	ID            FactionID     `json:"id" db:"id"`
	Label         string        `json:"label" db:"label"`
	Theme         *FactionTheme `json:"theme" db:"theme"`
	LogoUrl       string        `json:"logoUrl,omitempty"`
	BackgroundUrl string        `json:"backgroundUrl,omitempty"`
}

type FactionAbilityType string

const (
	FactionAbilityTypeAirStrike FactionAbilityType = "AIRSTRIKE"
	FactionAbilityTypeNuke      FactionAbilityType = "NUKE"
	FactionAbilityTypeHealing   FactionAbilityType = "HEALING"
)

type FactionAbility struct {
	ID                     FactionAbilityID   `json:"id" db:"id"`
	GameClientAbilityID    byte               `json:"gameClientAbilityID" db:"game_client_ability_id"`
	FactionID              FactionID          `json:"factionID" db:"faction_id"`
	Label                  string             `json:"label" db:"label"`
	Type                   FactionAbilityType `json:"type" db:"type"`
	Colour                 string             `json:"colour" db:"colour"`
	SupsCost               int                `json:"supsCost" db:"sups_cost"`
	ImageUrl               string             `json:"imageUrl" db:"image_url"`
	CooldownDurationSecond int                `json:"cooldownDurationSecond" db:"cooldown_duration_second"`
}
