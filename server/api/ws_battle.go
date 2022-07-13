package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"server/battle"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BattleControllerWS struct {
	API *API
}

func NewBattleController(api *API) *BattleControllerWS {
	bc := &BattleControllerWS{
		API: api,
	}

	api.Command(HubKeyBattleMechHistoryList, bc.BattleMechHistoryListHandler)
	api.Command(HubKeyBattleMechStats, bc.BattleMechStatsHandler)

	// commands from battle

	// faction queue
	api.SecureUserFactionCommand(battle.WSQueueJoin, api.BattleArena.QueueJoinHandler)
	api.SecureUserFactionCommand(battle.WSQueueLeave, api.BattleArena.QueueLeaveHandler)
	api.SecureUserFactionCommand(battle.WSMechArenaStatusUpdate, api.BattleArena.AssetUpdateRequest)

	// TODO: handle insurance and repair
	//api.SecureUserFactionCommand(battle.HubKeyAssetRepairPayFee, api.BattleArena.AssetRepairPayFeeHandler)
	//api.SecureUserFactionCommand(battle.HubKeyAssetRepairStatus, api.BattleArena.AssetRepairStatusHandler)


	api.SecureUserFactionCommand(battle.HubKeyPlayerAbilityUse, api.BattleArena.PlayerAbilityUse)

	// mech move command related
	api.SecureUserFactionCommand(battle.HubKeyMechMoveCommandCancel, api.BattleArena.MechMoveCommandCancelHandler)
	// battle ability related (bribing)
	api.SecureUserFactionCommand(battle.HubKeyBattleAbilityBribe, api.BattleArena.BattleAbilityBribe)
	api.SecureUserFactionCommand(battle.HubKeyAbilityLocationSelect, api.BattleArena.AbilityLocationSelect)

	// faction unique ability related (sup contribution)
	api.SecureUserFactionCommand(battle.HubKeFactionUniqueAbilityContribute, api.BattleArena.FactionUniqueAbilityContribute)

	return bc
}

type BattleMechHistoryRequest struct {
	Payload struct {
		MechID string `json:"mech_id"`
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
	Total         int                  `json:"total"`
	BattleHistory []BattleMechDetailed `json:"battle_history"`
}

const HubKeyBattleMechHistoryList = "BATTLE:MECH:HISTORY:LIST"

func (bc *BattleControllerWS) BattleMechHistoryListHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleMechHistoryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	battleMechs, err := boiler.BattleMechs(boiler.BattleMechWhere.MechID.EQ(req.Payload.MechID), qm.OrderBy("created_at desc"), qm.Limit(10), qm.Load(qm.Rels(boiler.BattleMechRels.Battle, boiler.BattleRels.GameMap))).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("BattleMechWhere", req.Payload.MechID).
			Str("db func", "BattleMechs").Err(err).Msg("unable to get battle mech history")
		return terror.Error(err, "Unable to retrieve battle history, try again or contact support.")
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
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"mech_id"`
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

const HubKeyBattleMechStats = "BATTLE:MECH:STATS"

func (bc *BattleControllerWS) BattleMechStatsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
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
