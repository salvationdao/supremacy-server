package server

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

type Player struct {
	ID               string          `json:"id"`
	FactionID        null.String     `json:"faction_id,omitempty"`
	Username         null.String     `json:"username,omitempty"`
	PublicAddress    null.String     `json:"public_address,omitempty"`
	IsAi             bool            `json:"is_ai"`
	DeletedAt        null.Time       `json:"deleted_at,omitempty"`
	UpdatedAt        time.Time       `json:"updated_at"`
	CreatedAt        time.Time       `json:"created_at"`
	MobileNumber     null.String     `json:"mobile_number,omitempty"`
	IssuePunishFee   decimal.Decimal `json:"issue_punish_fee"`
	ReportedCost     decimal.Decimal `json:"reported_cost"`
	Gid              int             `json:"gid"`
	Rank             string          `json:"rank"`
	SentMessageCount int             `json:"sent_message_count"`
	SynidcateID      null
	Stat             *boiler.PlayerStat `json:"stat"`

	Features []*Feature `json:"features"`
}

func (p *Player) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, p)
}

func PlayerFromBoiler(player *boiler.Player, features ...boiler.FeatureSlice) *Player {
	var serverFeatures []*Feature
	if len(features) > 0 {
		serverFeatures = FeaturesFromBoiler(features[0])
	}

	serverPlayer := &Player{
		ID:               player.ID,
		FactionID:        player.FactionID,
		Username:         player.Username,
		PublicAddress:    player.PublicAddress,
		IsAi:             player.IsAi,
		DeletedAt:        player.DeletedAt,
		UpdatedAt:        player.UpdatedAt,
		CreatedAt:        player.CreatedAt,
		MobileNumber:     player.MobileNumber,
		IssuePunishFee:   player.IssuePunishFee,
		ReportedCost:     player.ReportedCost,
		Gid:              player.Gid,
		Rank:             player.Rank,
		SentMessageCount: player.SentMessageCount,
		Features:         serverFeatures,
	}

	if player.R != nil {
		serverPlayer.Stat = player.R.IDPlayerStat
	}

	return serverPlayer
}

func (p *Player) Brief() *Player {
	return &Player{
		ID:        p.ID,
		FactionID: p.FactionID,
		Username:  p.Username,
		Gid:       p.Gid,
		Rank:      p.Rank,
		Stat:      p.Stat,
	}
}
