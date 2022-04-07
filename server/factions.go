package server

import (
	"fmt"
	"server/db/boiler"

	"github.com/ninja-software/terror/v2"

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

var RedMountainPlayerID = "305da475-53dc-4973-8d78-a30d390d3de5"
var BostonCyberneticsPlayerID = "15f29ee9-e834-4f76-aff8-31e39faabe2d"
var ZaibatsuPlayerID = "1a657a32-778e-4612-8cc1-14e360665f2b"

var FactionUsers = map[string]string{
	RedMountainFactionID.String():       RedMountainPlayerID,
	BostonCyberneticsFactionID.String(): BostonCyberneticsPlayerID,
	ZaibatsuFactionID.String():          ZaibatsuPlayerID,
}

type Faction struct {
	ID               FactionID     `json:"id" db:"id"`
	Label            string        `json:"label" db:"label"`
	Theme            *FactionTheme `json:"theme" db:"theme"`
	LogoBlobID       BlobID        `json:"logo_blob_id,omitempty"`
	BackgroundBlobID BlobID        `json:"background_blob_id,omitempty"`
	VotePrice        string        `json:"vote_price" db:"vote_price"`
	ContractReward   string        `json:"contract_reward" db:"contract_reward"`
	Description      string        `json:"description" db:"description"`
}

func (f *Faction) ToBoilerFaction() *boiler.Faction {
	newFaction := &boiler.Faction{
		ID:             f.ID.String(),
		VotePrice:      f.VotePrice,
		ContractReward: f.ContractReward,
		Label:          f.Label,
		//GuildID: , ?
		//DeletedAt:,
		//UpdatedAt:,
		//CreatedAt:,
		PrimaryColor:    f.Theme.Primary,
		SecondaryColor:  f.Theme.Secondary,
		BackgroundColor: f.Theme.Background,
	}
	return newFaction
}

func (f *Faction) SetFromBoilerFaction(bf *boiler.Faction) error {
	//f.LogoBlobID = bf. ?
	//f.BackgroundBlobID = bf. ?

	uuidFromString, err := uuid.FromString(bf.ID)
	if err != nil {
		return terror.Error(err, fmt.Sprintf("unable to parse %s to uuid", bf.ID))
	}
	f.ID = FactionID(uuidFromString)
	f.Label = bf.Label
	f.Theme = &FactionTheme{
		bf.PrimaryColor, bf.SecondaryColor, bf.BackgroundColor,
	}
	f.VotePrice = bf.VotePrice
	f.ContractReward = bf.ContractReward
	return nil
}

type FactionBrief struct {
	Label      string        `json:"label"`
	LogoBlobID BlobID        `json:"logo_blob_id,omitempty"`
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
	*boiler.FactionStat
	MvpPlayerUsername string `json:"mvp_username"`
	MemberCount       int64  `json:"member_count"`
}

type BattleAbility struct {
	ID                     string `json:"id" db:"id"`
	Label                  string `json:"label" db:"label"`
	Description            string `json:"description" db:"description"`
	CooldownDurationSecond int    `json:"cooldown_duration_second" db:"cooldown_duration_second"`
	Colour                 string `json:"colour"`
	TextColour             string `json:"text_colour"`
	ImageUrl               string `json:"image_url"`
}

func (ga *BattleAbility) Brief() *AbilityBrief {
	return &AbilityBrief{
		Label:    ga.Label,
		ImageUrl: ga.ImageUrl,
		Colour:   ga.Colour,
	}
}

type PunishmentOption string

const (
	PunishmentOptionRestrictLocationSelect   = "restrict_location_select"
	PunishmentOptionRestrictChat             = "restrict_chat"
	PunishmentOptionRestrictSupsContribution = "restrict_sups_contribution"
)
