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
)

// GetPendingMechsFromFactionID returns the next 3 (or less) mechs in backlog that belong to the specified faction.
// By default, it excludes mechs with the same owner ID (i.e. no two mechs with the same owner ID will be returned).
// However, if 3 backlogged faction mechs with unique owner IDs does not currently exist, GetPendingMechsFromFactionID may return
// mechs with the same owner ID.
func GetPendingMechsFromFactionID(factionID string, excludeMechIDs []string) (boiler.BattleQueueBacklogSlice, error) {
	battleMechs, err := GetNextBacklogMech(factionID, []*boiler.BattleQueueBacklog{}, excludeMechIDs, []string{}, false)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to get battle mechs")
		return nil, err
	}
	spew.Dump(battleMechs)
	return battleMechs, nil
}

const FACTION_MECH_LIMIT = 3

// GetNextBacklogMech is a recursive function called by GetBattleMechsFromFactionID
func GetNextBacklogMech(factionID string, battleMechs boiler.BattleQueueBacklogSlice, excludeMechIDs, excludeOwnerIDs []string, disableOwnerCheck bool) (boiler.BattleQueueBacklogSlice, error) {
	if len(battleMechs) == FACTION_MECH_LIMIT {
		return battleMechs, nil
	}

	for _, bm := range battleMechs {
		excludeMechIDs = append(excludeMechIDs, bm.MechID)
		if !disableOwnerCheck {
			excludeOwnerIDs = append(excludeOwnerIDs, bm.OwnerID)
		}
	}

	nextBattleMech, err := boiler.BattleQueueBacklogs(
		boiler.BattleQueueBacklogWhere.FactionID.EQ(factionID),
		boiler.BattleQueueBacklogWhere.MechID.NIN(excludeMechIDs),
		boiler.BattleQueueBacklogWhere.OwnerID.NIN(excludeOwnerIDs),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.BattleQueueBacklogColumns.QueuedAt)),
		qm.Limit(1),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		if disableOwnerCheck {
			return battleMechs, nil
		} else {
			return GetNextBacklogMech(factionID, battleMechs, excludeMechIDs, excludeOwnerIDs, true)
		}
	} else if err != nil {
		return nil, err
	}
	battleMechs = append(battleMechs, nextBattleMech)

	return GetNextBacklogMech(factionID, battleMechs, excludeMechIDs, excludeOwnerIDs, false)
}

// todo: rework this to factor in R queue positions
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

func GetCollectionItemStatus(collectionItem boiler.CollectionItem) (*server.MechArenaInfo, error) {
	// Check in marketplace
	now := time.Now()
	inMarketplace, err := collectionItem.ItemSales(
		boiler.ItemSaleWhere.EndAt.GT(now),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
	).Exists(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	if inMarketplace {
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusMarket,
			CanDeploy: false,
		}, nil
	}

	mechID := collectionItem.ItemID

	// Check in battle
	inBattle, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.MechID.EQ(mechID),
		boiler.BattleQueueWhere.BattleID.IsNotNull(),
	).Exists(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	if inBattle {
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusBattle,
			CanDeploy: false,
		}, nil
	}

	// Check in battle queue backlog
	pendingQueue, err := boiler.BattleQueueBacklogExists(gamedb.StdConn, mechID)
	if err != nil {
		return nil, err
	}

	if pendingQueue {
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusPendingQueue,
			CanDeploy: false,
		}, nil
	}

	owner, err := collectionItem.Owner().One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if owner != nil && owner.FactionID.Valid {
		// Check in battle queue
		queuePosition, err := MechQueuePosition(mechID, owner.FactionID.String)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		if queuePosition != nil {
			return &server.MechArenaInfo{
				Status:    server.MechArenaStatusQueue,
				CanDeploy: false,
			}, nil
		}
	}

	// Check if damaged
	rc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if rc != nil {
		canDeployRatio := GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))
		totalBlocks := TotalRepairBlocks(rc.MechID)
		if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).GreaterThan(canDeployRatio) {
			return &server.MechArenaInfo{
				Status:    server.MechArenaStatusDamaged,
				CanDeploy: false,
			}, nil
		}
	}

	return &server.MechArenaInfo{
		Status:    server.MechArenaStatusIdle,
		CanDeploy: true,
	}, nil
}

// MechQueuePosition return the faction queue position of the specified mech.
// If the mech is in battle, MechQueuePosition returns 0.
func MechQueuePosition(mechID string, factionID string) (*BattleQueuePosition, error) {
	q := `
	SELECT
		bq.mech_id,
		coalesce(_bq.queue_position, 0) AS queue_position
	FROM
		battle_queue bq
		LEFT OUTER JOIN (
		SELECT
			_bq.mech_id,
			row_number() OVER (ORDER BY _bq.queued_at) AS queue_position
		FROM
			battle_queue _bq
		WHERE
			_bq.faction_id = $1
			AND _bq.battle_id ISNULL) _bq ON _bq.mech_id = bq.mech_id
	WHERE
		bq.mech_id = $2
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
