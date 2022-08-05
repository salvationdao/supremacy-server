package db

import (
	"database/sql"
	"github.com/shopspring/decimal"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
)

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
		resp.QueuePosition = bqp.QueuePosition
		resp.CanDeploy = false
		return resp, nil
	}

	// check damaged
	mrc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mech id", mechID).Msg("Failed to load mech rapair stat")
		return nil, terror.Error(err, "Failed to load mech stat")
	}

	if mrc != nil && !mrc.CompletedAt.Valid {
		resp.Status = server.MechArenaStatusDamaged
		canDeployRatio := GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))
		if decimal.NewFromInt(int64(mrc.BlocksRepaired)).Div(decimal.NewFromInt(int64(mrc.BlocksRequiredRepair))).LessThan(canDeployRatio) {
			resp.CanDeploy = false
			return resp, nil
		}
	}

	return resp, nil
}

// MechQueuePosition return a list of mech queue position of the player (exclude in battle)
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
