package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
)

type BattleControllerWS struct {
	API *API
}

func NewBattleController(api *API) *BattleControllerWS {
	bc := &BattleControllerWS{
		API: api,
	}

	api.Command(server.HubKeyBattleMechHistoryList, bc.BattleMechHistoryListHandler)
	api.Command(server.HubKeyBattleMechStats, bc.BattleMechStatsHandler)

	return bc
}

type BattleMechHistoryRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Filter   *db.ListFilterRequest `json:"filter"`
		Sort     *db.ListSortRequest   `json:"sort"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type BattleDetailed struct {
	*boiler.Battle `json:"battle"`
	GameMap        *boiler.GameMap `json:"game_map"`
}

type BattleMechDetailed struct {
	*boiler.BattleMech
	Battle *BattleDetailed `json:"battle"`
}

type BattleMechHistoryResponse struct {
	Total         int64                    `json:"total"`
	BattleHistory []*db.BattleMechDetailed `json:"battle_history"`
}

func (bc *BattleControllerWS) BattleMechHistoryListHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleMechHistoryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, battleMechs, err := db.BattleMechsListPaginated(req.Payload.Filter, req.Payload.Sort, offset, req.Payload.PageSize)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "SaleAbilitiesList").Err(err).Interface("arguments", req.Payload).Msg("unable to get list of sale abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(BattleMechHistoryResponse{
		total,
		battleMechs,
	})
	return nil
}

type BattleMechStatsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

type BattleMechExtraStats struct {
	WinRate            float32 `json:"win_rate"`
	SurvivalRate       float32 `json:"survival_rate"`
	KillPercentile     uint8   `json:"kill_percentile"`
	SurvivalPercentile uint8   `json:"survival_percentile"`
}

type BattleMechStatsResponse struct {
	*boiler.MechStat
	ExtraStats BattleMechExtraStats `json:"extra_stats"`
}

func (bc *BattleControllerWS) BattleMechStatsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleMechStatsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	ms, err := boiler.MechStats(boiler.MechStatWhere.MechID.EQ(req.Payload.MechID)).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		reply(nil)
		return nil
	}
	if err != nil {
		return err
	}

	var total int
	var maxKills int
	var minKills int
	var maxSurvives int
	var minSurvives int
	err = gamedb.StdConn.QueryRow(`
	SELECT
		count(mech_id),
		max(total_kills),
		min(total_kills),
		max(total_wins),
		min(total_wins)
	FROM
		mech_stats
`).Scan(&total, &maxKills, &minKills, &maxSurvives, &minSurvives)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "QueryRow").Err(err).Msg("unable to get max, min value of total_kills")
		return terror.Error(err, "Unable to retrieve ")
	}

	var killPercentile uint8
	killPercentile = 0
	if maxKills-minKills > 0 {
		killPercentile = 100 - uint8(float64(ms.TotalKills-minKills)*100/float64(maxKills-minKills))
	}

	var survivePercentile uint8
	survivePercentile = 0
	if maxSurvives-minSurvives > 0 {
		survivePercentile = 100 - uint8(float64(ms.TotalWins-minSurvives)*100/float64(maxSurvives-minSurvives))
	}

	reply(BattleMechStatsResponse{
		MechStat: ms,
		ExtraStats: BattleMechExtraStats{
			WinRate:            float32(ms.TotalWins) / float32(ms.TotalLosses+ms.TotalWins),
			SurvivalRate:       float32(ms.BattlesSurvived) / float32(ms.TotalDeaths+ms.BattlesSurvived),
			KillPercentile:     killPercentile,
			SurvivalPercentile: survivePercentile,
		},
	})

	return nil
}
