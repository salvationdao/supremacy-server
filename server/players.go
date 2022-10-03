package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"

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
	SyndicateID      null.String     `json:"syndicate_id"`
	AcceptsMarketing null.Bool       `json:"accepts_marketing"`

	Stat      *boiler.PlayerStat `json:"stat"`
	Syndicate *boiler.Syndicate  `json:"syndicate"`

	Features []*Feature `json:"features"`
}

type PublicPlayer struct {
	ID        string      `json:"id"`
	Username  null.String `json:"username"`
	Gid       int         `json:"gid"`
	FactionID null.String `json:"faction_id"`
	AboutMe   null.String `json:"about_me"`
	Rank      string      `json:"rank"`
	CreatedAt time.Time   `json:"created_at"`
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
		SyndicateID:      player.SyndicateID,
		Features:         serverFeatures,
		AcceptsMarketing: player.AcceptsMarketing,
	}

	if player.R != nil {
		serverPlayer.Stat = player.R.IDPlayerStat
		serverPlayer.Syndicate = player.R.Syndicate
	}

	return serverPlayer
}

func PublicPlayerFromBoiler(p *boiler.Player) *PublicPlayer {
	return &PublicPlayer{
		ID:        p.ID,
		Username:  p.Username,
		FactionID: p.FactionID,
		Gid:       p.Gid,
		AboutMe:   p.AboutMe,
		Rank:      p.Rank,
		CreatedAt: p.CreatedAt,
	}
}

// Brief trim off confidential data from player
func (p *Player) Brief() *Player {
	return &Player{
		ID:        p.ID,
		FactionID: p.FactionID,
		Username:  p.Username,
		Gid:       p.Gid,
		Rank:      p.Rank,
		Stat:      p.Stat,
		Syndicate: p.Syndicate,
	}
}

type QuestStat struct {
	ID            string    `db:"id" json:"id"`
	RoundName     string    `db:"round_name" json:"round_name"`
	StartedAt     time.Time `db:"started_at" json:"started_at"`
	EndAt         time.Time `db:"end_at" json:"end_at"`
	Key           string    `db:"key" json:"key"`
	Name          string    `db:"name" json:"name"`
	Description   string    `db:"description" json:"description"`
	RequestAmount int       `db:"request_amount" json:"request_amount"`
	Obtained      bool      `db:"obtained" json:"obtained"`
}

type PlayerQueueStatus struct {
	TotalQueued int `json:"total_queued"`
	QueueLimit  int `json:"queue_limit"`
}
