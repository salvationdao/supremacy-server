package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BattleControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewBattleController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *BattleControllerWS {
	bc := &BattleControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "twitch_hub"),
		API:  api,
	}

	api.Command(HubKeyBattleMechHistoryList, bc.BattleMechHistoryListHandler)
	api.Command(HubKeyBattleMechStats, bc.BattleMechStatsHandler)

	return bc
}

type BattleMechHistoryRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

type BattleDetailed struct {
	*boiler.Battle
	GameMap *boiler.GameMap `json:"game_map"`
}

type BattleMechDetailed struct {
	*boiler.BattleMech
	Battle *BattleDetailed `json:"battle"`
}

type BattleMechHistoryResponse struct {
	Total         int                  `json:"total"`
	BattleHistory []BattleMechDetailed `json:"battle_history"`
}

const HubKeyBattleMechHistoryList = hub.HubCommandKey("BATTLE:MECH:HISTORY:LIST")

func (bc *BattleControllerWS) BattleMechHistoryListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &BattleMechHistoryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	battleMechs, err := boiler.BattleMechs(boiler.BattleMechWhere.MechID.EQ(req.Payload.MechID), qm.OrderBy("created_at desc"), qm.Limit(10), qm.Load(qm.Rels(boiler.BattleMechRels.Battle, boiler.BattleRels.GameMap))).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "BattleMechs").Err(err).Msg("unable to get battle mech history")
		return terror.Error(err)
	}

	output := []BattleMechDetailed{}
	for _, o := range battleMechs {
		output = append(output, BattleMechDetailed{
			BattleMech: o,
			Battle: &BattleDetailed{
				Battle:  o.R.Battle,
				GameMap: o.R.Battle.R.GameMap,
			},
		})
	}

	reply(BattleMechHistoryResponse{
		len(output),
		output,
	})

	return nil
}

type BattleMechStatsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"mech_id"`
}

type BattleMechExtraStats struct {
	WinRate            float32 `json:"win_rate"`
	KillPercentile     uint8   `json:"kill_percentile"`
	SurvivalPercentile uint8   `json:"survival_percentile"`
}

type BattleMechStatsResponse struct {
	*boiler.MechStat
	ExtraStats BattleMechExtraStats `json:"extra_stats"`
}

const HubKeyBattleMechStats = hub.HubCommandKey("BATTLE:MECH:STATS")

func (bc *BattleControllerWS) BattleMechStatsHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &BattleMechHistoryRequest{}
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
		return terror.Error(err)
	}

	var total int
	var maxKills int
	var minKills int
	var maxSurvives int
	var minSurvives int
	err = gamedb.Conn.QueryRow(context.Background(), `
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
		return terror.Error(err)
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
			WinRate:            float32(ms.TotalWins) / float32(ms.TotalDeaths+ms.TotalWins),
			KillPercentile:     killPercentile,
			SurvivalPercentile: survivePercentile,
		},
	})

	return nil
}
