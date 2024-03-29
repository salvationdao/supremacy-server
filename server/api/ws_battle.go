package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/go-chi/chi/v5"
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
	api.Command(HubKeyPlayerBattleMechHistoryList, bc.PlayerBattleMechHistoryListHandler)
	api.Command(HubKeyBattleMechStats, bc.BattleMechStatsHandler)

	// commands from battle
	api.SecureUserFactionCommand(battle.HubKeyPlayerAbilityUse, api.ArenaManager.PlayerAbilityUse)
	api.SecureUserFactionCommand(battle.HubKeyPlayerSupportAbilityUse, api.ArenaManager.PlayerSupportAbilityUse)

	// mech move command related
	api.SecureUserFactionCommand(battle.HubKeyMechMoveCommandCancel, api.ArenaManager.MechMoveCommandCancelHandler)
	return bc
}

const HubKeyGameMapList = "GAME:MAP:LIST"

func (api *API) GameMapListSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	gameMap, err := boiler.GameMaps(
		boiler.GameMapWhere.DisabledAt.IsNull(),
		boiler.GameMapWhere.BackgroundURL.NEQ(""),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("func", "GameMapListSubscribeHandler").Msg("Failed to load game maps.")
		return terror.Error(err, "Failed to load game maps.")
	}

	if gameMap == nil {
		reply([]*boiler.GameMap{})
		return nil
	}

	reply(gameMap)

	return nil
}

type BattleMechHistoryRequest struct {
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

type BattleDetailed struct {
	*boiler.Battle `json:"battle"`
	GameMap        *boiler.GameMap `json:"game_map"`
	BattleReplayID *string         `json:"battle_replay,omitempty"`
	ArenaGID       int             `json:"arena_gid"`
}

type BattleMechDetailed struct {
	*boiler.BattleMech
	Battle *BattleDetailed `json:"battle"`
	Mech   *boiler.Mech    `json:"mech"`
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
		battleMechDetail := BattleMechDetailed{
			BattleMech: o,
		}
		if o.R != nil && o.R.Battle != nil {
			battleMechDetail.Battle = &BattleDetailed{
				Battle:  o.R.Battle,
				GameMap: o.R.Battle.R.GameMap,
			}
			replay, err := boiler.BattleReplays(
				boiler.BattleReplayWhere.BattleID.EQ(o.R.Battle.ID),
				boiler.BattleReplayWhere.ArenaID.EQ(o.R.Battle.ArenaID),
				boiler.BattleReplayWhere.IsCompleteBattle.EQ(true),
				boiler.BattleReplayWhere.RecordingStatus.EQ(boiler.RecordingStatusSTOPPED),
				boiler.BattleReplayWhere.StreamID.IsNotNull(),
				qm.Load(boiler.BattleReplayRels.Arena),
			).One(gamedb.StdConn)
			if err != nil && err != sql.ErrNoRows {
				gamelog.L.Error().Err(err).Msg("Failed to get battle replay")
			}
			if replay != nil {
				battleMechDetail.Battle.BattleReplayID = &replay.ID
				if replay.R != nil && replay.R.Arena != nil {
					battleMechDetail.Battle.ArenaGID = replay.R.Arena.Gid
				}
			}
		}

		output = append(output, battleMechDetail)
	}

	reply(BattleMechHistoryResponse{
		len(output),
		output,
	})
	return nil
}

const HubKeyPlayerBattleMechHistoryList = "PLAYER:BATTLE:MECH:HISTORY:LIST"

type PlayerBattleMechHistoryRequest struct {
	Payload struct {
		PlayerID string `json:"player_id"`
	} `json:"payload"`
}

func (bc *BattleControllerWS) PlayerBattleMechHistoryListHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerBattleMechHistoryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	battleMechs, err := boiler.BattleMechs(
		boiler.BattleMechWhere.PilotedByID.EQ(req.Payload.PlayerID),
		qm.OrderBy("created_at desc"),
		qm.Limit(10),
		qm.Load(boiler.BattleMechRels.Mech),
		qm.Load(qm.Rels(boiler.BattleMechRels.Battle, boiler.BattleRels.GameMap)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("BattleMechWhere", req.Payload.PlayerID).
			Str("db func", "BattleMechs").Err(err).Msg("unable to get battle mech history")
		return terror.Error(err, "Unable to retrieve battle history, try again or contact support.")
	}

	output := []BattleMechDetailed{}
	for _, o := range battleMechs {

		var mech *boiler.Mech
		if o.R != nil && o.R.Mech != nil {
			mech = o.R.Mech
		}

		battleMechDetail := BattleMechDetailed{
			BattleMech: o,
			Mech:       mech,
		}

		if o.R != nil && o.R.Battle != nil {
			battleMechDetail.Battle = &BattleDetailed{
				Battle:  o.R.Battle,
				GameMap: o.R.Battle.R.GameMap,
			}
			replay, err := boiler.BattleReplays(
				boiler.BattleReplayWhere.BattleID.EQ(o.R.Battle.ID),
				boiler.BattleReplayWhere.ArenaID.EQ(o.R.Battle.ArenaID),
				boiler.BattleReplayWhere.IsCompleteBattle.EQ(true),
				boiler.BattleReplayWhere.RecordingStatus.EQ(boiler.RecordingStatusSTOPPED),
				boiler.BattleReplayWhere.StreamID.IsNotNull(),
				qm.Load(boiler.BattleReplayRels.Arena),
			).One(gamedb.StdConn)
			if err != nil && err != sql.ErrNoRows {
				gamelog.L.Error().Err(err).Msg("Failed to get battle replay")
			}
			if replay != nil {
				battleMechDetail.Battle.BattleReplayID = &replay.ID
				if replay.R != nil && replay.R.Arena != nil {
					battleMechDetail.Battle.ArenaGID = replay.R.Arena.Gid
				}
			}
		}

		output = append(output, battleMechDetail)
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
	err = gamedb.StdConn.QueryRow(fmt.Sprintf(`
				SELECT
					COUNT(%[1]s),
					MAX(%[2]s),
					MIN(%[2]s),
					MAX(%[3]s),
					MIN(%[3]s)
				FROM %[4]s
			`,
		boiler.MechStatColumns.MechID,
		boiler.MechStatColumns.TotalKills,
		boiler.MechStatColumns.TotalWins,
		boiler.TableNames.MechStats,
	)).Scan(&total, &maxKills, &minKills, &maxSurvives, &minSurvives)
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

func (api *API) PlayerAssetMechQueueSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	mechStatus, err := db.GetMechQueueStatus(chi.RouteContext(ctx).URLParam("mech_id"))
	if err != nil {
		return err
	}

	reply(mechStatus)
	return nil
}

func (api *API) ArenaListSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	reply(api.ArenaManager.AvailableBattleArenas())
	return nil
}

func (api *API) ArenaClosedSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	_, err := api.ArenaManager.GetArenaFromContext(ctx)
	if err != nil {
		// send arena is closed
		reply(true)
		return nil
	}

	// send arena isn't close
	reply(false)
	return nil
}

func (api *API) BattleEndDetail(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := api.ArenaManager.GetArenaFromContext(ctx)
	if err != nil {
		reply(nil)
		return nil
	}

	reply(arena.LastBattleResult)
	return nil
}

func (api *API) MiniMapAbilityDisplayList(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := api.ArenaManager.GetArenaFromContext(ctx)
	if err != nil {
		reply(nil)
		return nil
	}

	// if current battle still running
	btl := arena.CurrentBattle()
	if btl == nil {
		return nil
	}

	if btl != nil {
		reply(btl.MiniMapAbilityDisplayList.List())
	}
	return nil
}

func (api *API) ChallengeFundSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	reply(api.ChallengeFund)
	return nil
}

func (api *API) BattleState(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := api.ArenaManager.GetArenaFromContext(ctx)
	if err != nil {
		reply(battle.EndState)
		return nil
	}

	reply(arena.CurrentBattleState())
	return nil
}

func (api *API) FactionStakedMechs(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "FactionStakedMechs").Logger()
	sms, err := boiler.StakedMechs(
		boiler.StakedMechWhere.FactionID.EQ(factionID),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load staked mechs.")
		return terror.Error(err, "Failed to load staked mechs.")
	}

	mechIDs := []string{}
	for _, sm := range sms {
		mechIDs = append(mechIDs, sm.MechID)
	}

	if len(mechIDs) > 0 {
		mbs, err := db.LobbyMechsBrief("", mechIDs...)
		if err != nil {
			return err
		}

		reply(mbs)
	}

	return nil
}
