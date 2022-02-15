package server

import (
	"github.com/gofrs/uuid"
)

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

var RedMountainFactionID = FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060")))
var BostonCyberneticsFactionID = FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2")))
var ZaibatsuFactionID = FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d")))

type Faction struct {
	ID               FactionID     `json:"id" db:"id"`
	Label            string        `json:"label" db:"label"`
	Theme            *FactionTheme `json:"theme" db:"theme"`
	LogoBlobID       BlobID        `json:"logoBlobID,omitempty"`
	BackgroundBlobID BlobID        `json:"backgroundBlobID,omitempty"`
	VotePrice        string        `json:"votePrice" db:"vote_price"`
}

type FactionStat struct {
	ID         FactionID `json:"id" db:"id"`
	WinCount   *int64    `json:"winCount,omitempty" db:"win_count,omitempty"`
	LossCount  *int64    `json:"lossCount,omitempty" db:"loss_count,omitempty"`
	KillCount  *int64    `json:"killCount,omitempty" db:"kill_count,omitempty"`
	DeathCount *int64    `json:"deathCount,omitempty" db:"death_count,omitempty"`
}

type BattleAbility struct {
	ID                     BattleAbilityID `json:"id" db:"id"`
	Label                  string          `json:"label" db:"label"`
	CooldownDurationSecond int             `json:"cooldownDurationSecond" db:"cooldown_duration_second"`
	Colour                 string          `json:"colour"`
	ImageUrl               string          `json:"imageUrl"`
}

type GameAbility struct {
	ID                  GameAbilityID    `json:"id" db:"id"`
	GameClientAbilityID byte             `json:"gameClientAbilityID" db:"game_client_ability_id"`
	BattleAbilityID     *BattleAbilityID `json:"battleAbilityID,omitempty" db:"battle_ability_id,omitempty"`
	Colour              string           `json:"colour" db:"colour"`
	ImageUrl            string           `json:"imageUrl" db:"image_url"`
	FactionID           FactionID        `json:"factionID" db:"faction_id"`
	Label               string           `json:"label" db:"label"`
	SupsCost            string           `json:"supsCost" db:"sups_cost"`
	CurrentSups         string           `json:"currentSups"`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	AbilityTokenID    uint64
	WarMachineTokenID uint64
	ParticipantID     *byte
	WarMachineName    string
	WarMachineImage   string

	// Category title for frontend to group the abilities together
	Title string `json:"title"`
}
