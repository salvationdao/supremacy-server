package db

import (
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
)

func GetBattleMechsFromFactionID(factionID string) (boiler.BattleQueueSlice, error) {
	battleMechs, err := GetNextBattleMech(factionID, []*boiler.BattleQueue{}, false)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to get battle mechs")
		return nil, err
	}
	spew.Dump(battleMechs)
	return battleMechs, nil
}

const FACTION_MECH_LIMIT = 3

// GetNextBattleMech returns the next 3 (or less) mechs in queue that belong to the specified faction.
// By default, it excludes mechs with the same owner ID (i.e. no two mechs with the same owner ID will be returned).
// However, if 3 queued faction mechs with unique owner IDs does not currently exist, GetNextBattleMech may return
// mechs with the same owner ID.
func GetNextBattleMech(factionID string, battleMechs boiler.BattleQueueSlice, disableOwnerCheck bool) (boiler.BattleQueueSlice, error) {
	if len(battleMechs) == FACTION_MECH_LIMIT {
		return battleMechs, nil
	}

	notInIDs := []string{}
	notInOwnerIDs := []string{}
	for _, bm := range battleMechs {
		notInIDs = append(notInIDs, bm.ID)
		if !disableOwnerCheck {
			notInOwnerIDs = append(notInOwnerIDs, bm.OwnerID)
		}
	}

	nextBattleMech, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.FactionID.EQ(factionID),
		boiler.BattleQueueWhere.ID.NIN(notInIDs),
		boiler.BattleQueueWhere.OwnerID.NIN(notInOwnerIDs),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.BattleQueueColumns.QueuedAt)),
		qm.Limit(1),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		if disableOwnerCheck {
			return battleMechs, nil
		} else {
			return GetNextBattleMech(factionID, battleMechs, true)
		}
	} else if err != nil {
		return nil, err
	}
	battleMechs = append(battleMechs, nextBattleMech)

	return GetNextBattleMech(factionID, battleMechs, false)
}

func GetMinimumQueueWaitTimeSecondsFromFactionID(factionID string) (int64, error) {
	averageBattleLengthSecs, err := GetAverageBattleLengthSeconds()
	if err != nil {
		return -1, err
	}

	var qp struct {
		NextQueuePosition int64 `json:"next_queue_position"`
	}
	err = boiler.NewQuery(
		qm.Select(fmt.Sprintf("count(DISTINCT %s) as next_queue_position",
			qm.Rels(boiler.TableNames.BattleQueue, boiler.BattleQueueColumns.OwnerID),
		)),
		qm.From(boiler.TableNames.BattleQueue),
		qm.Where(fmt.Sprintf("%s = ?",
			qm.Rels(boiler.TableNames.BattleQueue, boiler.BattleQueueColumns.FactionID)),
			factionID),
	).Bind(nil, gamedb.StdConn, &qp)
	if err != nil {
		return -1, err
	}

	return ((qp.NextQueuePosition + 1) / FACTION_MECH_LIMIT) * averageBattleLengthSecs, nil
}

func GetAverageBattleLengthSeconds() (int64, error) {
	var bl struct {
		AveLengthSeconds int64 `boil:"ave_length_seconds"`
	}
	err := boiler.NewQuery(
		qm.SQL(fmt.Sprintf(`
		SELECT coalesce(avg(battle_length.length), 0)::numeric::integer as ave_length_seconds
		FROM (
			SELECT extract(EPOCH FROM ended_at - started_at) AS length
			FROM %s
			WHERE %s IS NOT NULL
			ORDER BY %s DESC
			LIMIT 100
		) battle_length;
	`, boiler.TableNames.Battles, boiler.BattleColumns.EndedAt, boiler.BattleColumns.StartedAt)),
	).Bind(nil, gamedb.StdConn, &bl)
	if err != nil {
		return -1, err
	}

	return bl.AveLengthSeconds, nil
}

// MechArenaStatus return mech arena status from given collection item
func MechArenaStatus(userID string, mechID string, factionID string) (*server.MechArenaInfo, error) {
	resp := &server.MechArenaInfo{
		Status:    server.MechArenaStatusIdle,
		CanDeploy: true,
	}

	// check ownership of the mech
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(userID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to find mech from db")
	}

	// check market
	now := time.Now()
	is, err := collectionItem.ItemSales(
		boiler.ItemSaleWhere.EndAt.GT(now),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to check mech in marketplace")
	}

	if is != nil {
		resp.Status = server.MechArenaStatusMarket
		resp.CanDeploy = false
		return resp, nil
	}

	// check in battle
	bq, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.MechID.EQ(collectionItem.ItemID),
		boiler.BattleQueueWhere.BattleID.IsNotNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to get war machine battle state from db")
	}

	// if mech is in battle
	if bq != nil {
		resp.Status = server.MechArenaStatusBattle
		resp.CanDeploy = false
		return resp, nil
	}

	// check mech is in queue
	bqp, err := MechQueuePosition(collectionItem.ItemID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to check mech position")
	}

	if bqp != nil {
		resp.Status = server.MechArenaStatusQueue
		resp.CanDeploy = false
		return resp, nil
	}

	// check damaged
	mrc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mech id", mechID).Msg("Failed to load mech rapair stat")
		return nil, terror.Error(err, "Failed to load mech stat")
	}

	if mrc != nil {
		resp.Status = server.MechArenaStatusDamaged
		canDeployRatio := GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))
		totalBlocks := TotalRepairBlocks(mrc.MechID)
		if decimal.NewFromInt(int64(mrc.BlocksRequiredRepair - mrc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).GreaterThan(canDeployRatio) {
			resp.CanDeploy = false
			return resp, nil
		}
	}

	return resp, nil
}

// MechQueuePosition return the queue position of the specified mech (exclude in battle)
func MechQueuePosition(mechID, factionID string) (*BattleQueuePosition, error) {
	q := `
		SELECT
			bq.mech_id,
			coalesce(_bq.queue_position, 0) AS queue_position
		FROM battle_queue bq
		LEFT OUTER JOIN (SELECT
							 _bq.mech_id,
							 row_number () over (ORDER BY _bq.queued_at) AS queue_position
						 FROM
							 battle_queue _bq
						 WHERE
								 _bq.faction_id = $1 AND _bq.battle_id isnull) _bq ON _bq.mech_id = bq.mech_id
		WHERE bq.mech_id = $2
	`
	qp := &BattleQueuePosition{}
	err := gamedb.StdConn.QueryRow(q, factionID, mechID).Scan(&qp.MechID, &qp.QueuePosition)
	if err != nil {
		return nil, err
	}

	return qp, nil
}

func FactionQueue(factionID string) ([]*BattleQueuePosition, error) {
	q := `
		SELECT
			bq.mech_id,
			coalesce(_bq.queue_position, 0) AS queue_position
		FROM battle_queue bq
		LEFT OUTER JOIN (SELECT
							 _bq.mech_id,
							 row_number () over (ORDER BY _bq.queued_at) AS queue_position
						 FROM
							 battle_queue _bq
						 WHERE
								 _bq.faction_id = $1 AND _bq.battle_id isnull) _bq ON _bq.mech_id = bq.mech_id
		WHERE bq.faction_id = $1
	`

	qResult, err := gamedb.StdConn.Query(q, factionID)
	if err != nil {
		return nil, err
	}

	var mqp []*BattleQueuePosition
	for qResult.Next() {
		qp := &BattleQueuePosition{}
		err = qResult.Scan(&qp.MechID, &qp.QueuePosition)
		if err != nil {
			return nil, err
		}

		mqp = append(mqp, qp)
	}

	return mqp, nil
}
