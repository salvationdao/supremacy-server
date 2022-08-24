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
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"

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
	api.Command(HubKeyNextBattleDetails, bc.NextBattleDetails)

	// commands from battle

	// faction queue
	api.SecureUserFactionCommand(battle.WSQueueJoin, api.ArenaManager.QueueJoinHandler)
	api.SecureUserFactionCommand(battle.WSMechArenaStatusUpdate, api.ArenaManager.AssetUpdateRequest)

	api.SecureUserFactionCommand(battle.HubKeyPlayerAbilityUse, api.ArenaManager.PlayerAbilityUse)

	// mech move command related
	api.SecureUserFactionCommand(battle.HubKeyMechMoveCommandCancel, api.ArenaManager.MechMoveCommandCancelHandler)
	// battle ability related (bribing)
	api.SecureUserFactionCommand(battle.HubKeyAbilityLocationSelect, api.ArenaManager.AbilityLocationSelect)

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
		boiler.BattleMechWhere.OwnerID.EQ(req.Payload.PlayerID),
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
		output = append(output, BattleMechDetailed{
			BattleMech: o,
			Battle: &BattleDetailed{
				Battle:  o.R.Battle,
				GameMap: o.R.Battle.R.GameMap,
			},
			Mech: mech,
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

func (api *API) QueueStatusSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	queueLength, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("factionID", user.FactionID.String).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	reply(battle.QueueStatusResponse{
		QueueLength: queueLength, // return the current queue length
		QueueCost:   db.GetDecimalWithDefault(db.KeyBattleQueueFee, decimal.New(250, 18)),
	})
	return nil
}

func (api *API) PlayerAssetMechQueueSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	mechID := chi.RouteContext(ctx).URLParam("mech_id")

	queueDetails, err := db.MechArenaStatus(user.ID, mechID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Invalid request received.")
	}

	reply(queueDetails)
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

const HubKeyNextBattleDetails = "BATTLE:NEXT:DETAILS"

type BattleMap struct {
	Name          string `json:"name,omitempty"`
	BackgroundURL string `json:"background_url,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
}
type NextBattle struct {
	Map        *BattleMap `json:"map,omitempty"`
	BCMechIDs  []string   `json:"bc_mech_ids,omitempty"`
	ZHIMechIDs []string   `json:"zhi_mech_ids,omitempty"`
	RMMechIDs  []string   `json:"rm_mech_ids,omitempty"`
}

func (bc *BattleControllerWS) NextBattleDetails(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {

	// get queue
	queue, err := db.LoadBattleQueue(context.Background(), 3)

	if err != nil {
		return err
	}

	rm, err := boiler.Factions(boiler.FactionWhere.Label.EQ("Red Mountain Offworld Mining Corporation")).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "failed getting faction (RM)")
	}

	zhi, err := boiler.Factions(boiler.FactionWhere.Label.EQ("Zaibatsu Heavy Industries")).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "failed getting faction (ZHI)")
	}

	boc, err := boiler.Factions(boiler.FactionWhere.Label.EQ("Boston Cybernetics")).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "failed getting faction (BOC)")
	}

	rmMechIDs := []string{}
	zhiMechIDs := []string{}
	bcMechIDs := []string{}

	for _, q := range queue {
		if q.FactionID == rm.ID {
			rmMechIDs = append(rmMechIDs, q.MechID)
		}

		if q.FactionID == zhi.ID {
			zhiMechIDs = append(rmMechIDs, q.MechID)
		}

		if q.FactionID == boc.ID {
			bcMechIDs = append(rmMechIDs, q.MechID)
		}
	}

	resp := NextBattle{
		BCMechIDs:  bcMechIDs,
		ZHIMechIDs: zhiMechIDs,
		RMMechIDs:  rmMechIDs,
	}
	reply(resp)

	return nil
}
