package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"

	"github.com/gofrs/uuid"
)

type Faction struct {
	ID              string `json:"id" db:"id"`
	Label           string `json:"label" db:"label"`
	VotePrice       string `json:"vote_price" db:"vote_price"`
	ContractReward  string `json:"contract_reward" db:"contract_reward"`
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	BackgroundColor string `json:"background_color"`
}

func (b *Faction) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

const SupremacyBattleUserID = "87c60803-b051-4abb-aa60-487104946bd7"
const SupremacySystemModeratorUserID = "7bba7172-932a-4293-9765-ebd0ae98f0ea"
const SupremacySystemAdminUserID = "7bea1ab5-cc2e-46bb-95d4-e8082e141f1f"

var RedMountainFactionID = "98bf7bb3-1a7c-4f21-8843-458d62884060"
var BostonCyberneticsFactionID = "7c6dde21-b067-46cf-9e56-155c88a520e2"
var ZaibatsuFactionID = "880db344-e405-428d-84e5-6ebebab1fe6d"

var RedMountainPlayerID = "305da475-53dc-4973-8d78-a30d390d3de5"
var BostonCyberneticsPlayerID = "15f29ee9-e834-4f76-aff8-31e39faabe2d"
var ZaibatsuPlayerID = "1a657a32-778e-4612-8cc1-14e360665f2b"

var FactionUsers = map[string]string{
	RedMountainFactionID:       RedMountainPlayerID,
	BostonCyberneticsFactionID: BostonCyberneticsPlayerID,
	ZaibatsuFactionID:          ZaibatsuPlayerID,
}

func (f *Faction) SetFromBoilerFaction(bf *boiler.Faction) error {
	//f.LogoBlobID = bf. ?
	//f.BackgroundBlobID = bf. ?
	f.ID = bf.ID
	f.Label = bf.Label
	f.PrimaryColor = bf.R.FactionPalette.Primary
	f.SecondaryColor = bf.R.FactionPalette.Text
	f.BackgroundColor = bf.R.FactionPalette.Background
	f.VotePrice = bf.VotePrice
	f.ContractReward = bf.ContractReward
	return nil
}

type FactionBrief struct {
	Label      string `json:"label"`
	LogoBlobID BlobID `json:"logo_blob_id,omitempty"`
	//Theme      *FactionTheme `json:"theme"`
}

func (f *Faction) Brief() *FactionBrief {
	return &FactionBrief{
		Label: f.Label,
		//Theme: f.Theme,
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

type MechQueuePosition struct {
	MechID   uuid.UUID `json:"mechID"`
	Position int       `json:"position"`
}
