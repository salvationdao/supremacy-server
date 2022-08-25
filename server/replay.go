package server

import (
	"github.com/volatiletech/null/v8"
	"server/db/boiler"
)

type BattleReplay struct {
	ID               string      `json:"id"`
	StreamID         null.String `json:"stream_id,omitempty"`
	ArenaID          string      `json:"arena_id"`
	BattleID         string      `json:"battle_id"`
	IsCompleteBattle bool        `json:"is_complete_battle"`
	RecordingStatus  string      `json:"recording_status"`
	StartedAt        null.Time   `json:"started_at,omitempty"`
	StoppedAt        null.Time   `json:"stopped_at,omitempty"`
}

func BattleReplayFromBoiler(replay *boiler.BattleReplay) *BattleReplay {
	return &BattleReplay{
		ID:               replay.ID,
		StreamID:         replay.StreamID,
		ArenaID:          replay.ArenaID,
		BattleID:         replay.BattleID,
		IsCompleteBattle: replay.IsCompleteBattle,
		RecordingStatus:  replay.RecordingStatus,
		StartedAt:        replay.StartedAt,
		StoppedAt:        replay.StoppedAt,
	}
}
