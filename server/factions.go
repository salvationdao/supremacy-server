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

type FactionBrief struct {
	Label      string        `json:"label"`
	LogoBlobID BlobID        `json:"logoBlobID,omitempty"`
	Theme      *FactionTheme `json:"theme"`
}

func (f *Faction) Brief() *FactionBrief {
	return &FactionBrief{
		Label:      f.Label,
		LogoBlobID: f.LogoBlobID,
		Theme:      f.Theme,
	}
}

type FactionStat struct {
	ID         FactionID `json:"id" db:"id"`
	WinCount   *int64    `json:"winCount" db:"win_count,omitempty"`
	LossCount  *int64    `json:"lossCount" db:"loss_count,omitempty"`
	KillCount  *int64    `json:"killCount" db:"kill_count,omitempty"`
	DeathCount *int64    `json:"deathCount" db:"death_count,omitempty"`
}

type BattleAbility struct {
	ID                     BattleAbilityID `json:"id" db:"id"`
	Label                  string          `json:"label" db:"label"`
	CooldownDurationSecond int             `json:"cooldownDurationSecond" db:"cooldown_duration_second"`
	Colour                 string          `json:"colour"`
	ImageUrl               string          `json:"imageUrl"`
}

func (ga *BattleAbility) Brief() *AbilityBrief {
	return &AbilityBrief{
		Label:    ga.Label,
		ImageUrl: ga.ImageUrl,
		Colour:   ga.Colour,
	}
}
