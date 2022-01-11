package server

type Faction struct {
	ID       FactionID `json:"id"`
	Label    string    `json:"label"`
	ImageUrl string    `json:"imageUrl"`
	Colour   string    `json:"colour"`
}

type FactionAbilityType string

const (
	FactionAbilityTypeAirStrike FactionAbilityType = "AIRSTRIKE"
	FactionAbilityTypeNuke      FactionAbilityType = "NUKE"
	FactionAbilityTypeHealing   FactionAbilityType = "HEALING"
)

type FactionAbility struct {
	ID                     FactionAbilityID   `json:"id" db:"id"`
	FactionID              FactionID          `json:"factionID" db:"faction_id"`
	Label                  string             `json:"label" db:"label"`
	Type                   FactionAbilityType `json:"type" db:"type"`
	Colour                 string             `json:"colour" db:"colour"`
	SupremacyTokenCost     int                `json:"supremacyTokenCost" db:"supremacy_token_cost"`
	ImageUrl               string             `json:"imageUrl" db:"image_url"`
	CooldownDurationSecond int                `json:"cooldownDurationSecond" db:"cooldown_duration_second"`
}
