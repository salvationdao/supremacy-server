package server

import (
	"github.com/volatiletech/null/v8"
	"server/db/boiler"
)

type BattleReplay struct {
	Battle           *boiler.Battle  `json:"battle"`
	GameMap          *boiler.GameMap `json:"game_map"`
	ID               string          `json:"id"`
	StreamID         null.String     `json:"stream_id,omitempty"`
	ArenaID          string          `json:"arena_id"`
	BattleID         string          `json:"battle_id"`
	IsCompleteBattle bool            `json:"is_complete_battle"`
	RecordingStatus  string          `json:"recording_status"`
	StartedAt        null.Time       `json:"started_at,omitempty"`
	StoppedAt        null.Time       `json:"stopped_at,omitempty"`
	Events           null.JSON       `json:"events"`
}

func BattleReplayFromBoilerWithEvent(replay *boiler.BattleReplay) *BattleReplay {
	battleReplay := &BattleReplay{
		ID:               replay.ID,
		StreamID:         replay.StreamID,
		ArenaID:          replay.ArenaID,
		BattleID:         replay.BattleID,
		IsCompleteBattle: replay.IsCompleteBattle,
		RecordingStatus:  replay.RecordingStatus,
		StartedAt:        replay.StartedAt,
		StoppedAt:        replay.StoppedAt,
		Events:           replay.BattleEvents,
	}

	if replay.R != nil && replay.R.Battle != nil {
		battleReplay.Battle = replay.R.Battle

		if replay.R.Battle.R != nil && replay.R.Battle.R.GameMap != nil {
			battleReplay.GameMap = replay.R.Battle.R.GameMap
		}
	}

	return battleReplay
}

func BattleReplaySliceFromBoilerWithEvent(replays []*boiler.BattleReplay) []*BattleReplay {
	var slice []*BattleReplay
	for _, replay := range replays {
		slice = append(slice, BattleReplayFromBoilerWithEvent(replay))
	}
	return slice
}

func BattleReplayFromBoilerNoEvent(replay *boiler.BattleReplay) *BattleReplay {
	battleReplay := &BattleReplay{
		ID:               replay.ID,
		StreamID:         replay.StreamID,
		ArenaID:          replay.ArenaID,
		BattleID:         replay.BattleID,
		IsCompleteBattle: replay.IsCompleteBattle,
		RecordingStatus:  replay.RecordingStatus,
		StartedAt:        replay.StartedAt,
		StoppedAt:        replay.StoppedAt,
	}

	if replay.R != nil && replay.R.Battle != nil {
		battleReplay.Battle = replay.R.Battle

		if replay.R.Battle.R != nil && replay.R.Battle.R.GameMap != nil {
			battleReplay.GameMap = replay.R.Battle.R.GameMap
		}
	}

	return battleReplay
}

func BattleReplaySliceFromBoilerNoEvent(replays []*boiler.BattleReplay) []*BattleReplay {
	var slice []*BattleReplay
	for _, replay := range replays {
		slice = append(slice, BattleReplayFromBoilerWithEvent(replay))
	}
	return slice
}
