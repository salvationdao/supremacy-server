package db

import (
	"database/sql"
	"fmt"
	"math"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/friendsofgo/errors"
)

const FACTION_MECH_LIMIT = 3

func GetPreviousBattleOwnerIDs() ([]string, error) {
	var oids []*struct {
		OwnerID string `json:"owner_id"`
	}
	err := boiler.NewQuery(
		qm.SQL(fmt.Sprintf(`
		SELECT
			owner_id
		FROM
			%s
		ORDER BY %s desc
		LIMIT %d
		`,
			boiler.TableNames.BattleQueue,
			boiler.BattleQueueColumns.InsertedAt,
			FACTION_MECH_LIMIT*3),
		),
	).Bind(nil, gamedb.StdConn, &oids)
	if errors.Is(err, sql.ErrNoRows) {
		return []string{}, nil
	}
	if err != nil {
		return []string{}, err
	}

	ownerIDs := []string{}
	for _, o := range oids {
		ownerIDs = append(ownerIDs, o.OwnerID)
	}

	return ownerIDs, nil
}

func GetNumberOfMechsInQueueFromFactionID(factionID string) (int64, error) {
	count, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.FactionID.EQ(factionID),
		boiler.BattleQueueWhere.BattleID.IsNull(),
	).Count(gamedb.StdConn)
	if err != nil {
		return -1, err
	}

	return count, nil
}

// GetPendingMechsFromFactionID returns the next 3 (or less) mechs in backlog that belong to the specified faction.
// By default, it excludes mechs with the same owner ID (i.e. no two mechs with the same owner ID will be returned).
// However, if 3 backlogged faction mechs with unique owner IDs does not currently exist, GetPendingMechsFromFactionID may return
// mechs with the same owner ID.
func GetPendingMechsFromFactionID(factionID string, excludeOwnerIDs []string, limit int) (boiler.BattleQueueBacklogSlice, error) {
	pendingMechs, err := boiler.BattleQueueBacklogs(
		qm.Select(fmt.Sprintf("DISTINCT ON (%s) %s.*",
			qm.Rels(boiler.TableNames.BattleQueueBacklog, boiler.BattleQueueBacklogColumns.OwnerID),
			boiler.TableNames.BattleQueueBacklog,
		)),
		boiler.BattleQueueBacklogWhere.FactionID.EQ(factionID),
		boiler.BattleQueueBacklogWhere.OwnerID.NIN(excludeOwnerIDs),
		qm.OrderBy(fmt.Sprintf("%s, %s asc", boiler.BattleQueueBacklogColumns.OwnerID, boiler.BattleQueueBacklogColumns.QueuedAt)),
		qm.Limit(limit),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return pendingMechs, nil
}

func GetBattleETASecondsFromMechID(mechID string, factionID string) (int64, error) {
	averageBattleLengthSecs, err := GetAverageBattleLengthSeconds()
	if err != nil {
		return -1, err
	}

	queuePosition, err := MechQueuePosition(mechID, factionID)
	if err != nil {
		return -1, err
	}
	if queuePosition.QueuePosition == 0 {
		return 0, err
	}

	return int64(math.Ceil(float64(queuePosition.QueuePosition)/float64(FACTION_MECH_LIMIT))) * averageBattleLengthSecs, nil
}

func GetMinimumQueueWaitTimeSecondsFromFactionID(factionID string) (int64, error) {
	averageBattleLengthSecs, err := GetAverageBattleLengthSeconds()
	if err != nil {
		return -1, err
	}

	var qps []*struct {
		QueuePosition int64 `json:"queue_position"`
	}
	err = boiler.NewQuery(
		qm.SQL(fmt.Sprintf(`
		SELECT
			row_number() OVER (ORDER BY %s) AS queue_position
		FROM
			%s
		WHERE
			%s = $1
			AND %s IS NULL
		`,
			boiler.BattleQueueColumns.QueuedAt,
			boiler.TableNames.BattleQueue,
			boiler.BattleQueueColumns.FactionID,
			boiler.BattleQueueColumns.BattleID),
			factionID, // $1
		),
	).Bind(nil, gamedb.StdConn, &qps)
	if errors.Is(err, sql.ErrNoRows) {
		return int64(GetIntWithDefault(KeyQueueTickerIntervalSeconds, 10)), nil
	}
	if err != nil {
		return -1, err
	}

	return ((int64(len(qps)) + 1) / FACTION_MECH_LIMIT) * averageBattleLengthSecs, nil
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
	l := gamelog.L.With().Str("func", "GetCollectionItemStatus").Interface("collectionItem", collectionItem).Logger()

	// Check in marketplace
	now := time.Now()
	inMarketplace, err := collectionItem.ItemSales(
		boiler.ItemSaleWhere.EndAt.GT(now),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
	).Exists(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to check in marketplace")
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
		l.Error().Err(err).Msg("failed to check in battle")
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
		l.Error().Err(err).Msg("failed to check in queue backlog")
		return nil, err
	}

	owner, err := collectionItem.Owner().One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("failed to get owner of collection item")
		return nil, err
	}

	if pendingQueue {
		eta, err := GetMinimumQueueWaitTimeSecondsFromFactionID(owner.FactionID.String)
		if err != nil {
			l.Error().Err(err).Msg("failed to get faction queue eta")
			return nil, err
		}

		return &server.MechArenaInfo{
			Status:           server.MechArenaStatusPendingQueue,
			CanDeploy:        false,
			BattleETASeconds: null.Int64From(eta),
		}, nil
	}

	if owner != nil && owner.FactionID.Valid {
		// Check in battle queue
		exists, err := boiler.BattleQueueExists(gamedb.StdConn, mechID)
		if err != nil {
			l.Error().Err(err).Msg("failed to check in queue")
			return nil, err
		}

		if exists {
			eta, err := GetBattleETASecondsFromMechID(mechID, owner.FactionID.String)
			if err != nil {
				l.Error().Err(err).Msg("failed to get battle eta for mech")
				return nil, err
			}
			return &server.MechArenaInfo{
				Status:           server.MechArenaStatusQueue,
				CanDeploy:        false,
				BattleETASeconds: null.Int64From(eta),
			}, nil
		}
	}

	// Check if damaged
	rc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("failed to check if damaged")
		return nil, err
	}

	if rc != nil {
		canDeployRatio := GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))
		totalBlocks := TotalRepairBlocks(rc.MechID)
		if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).GreaterThan(canDeployRatio) {
			// If less than 50% repaired
			return &server.MechArenaInfo{
				Status:    server.MechArenaStatusDamaged,
				CanDeploy: false,
			}, nil
		}
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusDamaged,
			CanDeploy: true,
		}, nil
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
