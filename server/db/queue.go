package db

import (
	"server/gamedb"
)

// MechQueuePosition return a list of mech queue position of the player (exclude in battle)
func MechQueuePosition(mechID, factionID string) (*BattleQueuePosition, error) {
	q := `
		SELECT
			bq.mech_id,
			coalesce(_bq.queue_position, 0) AS queue_position,
			bq.battle_contract_id
		FROM battle_queue bq
		LEFT OUTER JOIN (SELECT
							 _bq.mech_id,
							 _bq.battle_contract_id,
							 row_number () over (ORDER BY _bq.queued_at) AS queue_position
						 FROM
							 battle_queue _bq
						 WHERE
								 _bq.faction_id = $1 AND _bq.battle_id isnull) _bq ON _bq.mech_id = bq.mech_id
		WHERE bq.mech_id = $2
	`
	qp := &BattleQueuePosition{}
	err := gamedb.StdConn.QueryRow(q, factionID, mechID).Scan(&qp.MechID, &qp.QueuePosition, &qp.BattleContractID)
	if err != nil {
		return nil, err
	}

	return qp, nil
}

func FactionQueue(factionID string) ([]*BattleQueuePosition, error) {
	q := `
		SELECT
			bq.mech_id,
			coalesce(_bq.queue_position, 0) AS queue_position,
			bq.battle_contract_id
		FROM battle_queue bq
		LEFT OUTER JOIN (SELECT
							 _bq.mech_id,
							 _bq.battle_contract_id,
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
		err = qResult.Scan(&qp.MechID, &qp.QueuePosition, &qp.BattleContractID)
		if err != nil {
			return nil, err
		}

		mqp = append(mqp, qp)
	}

	return mqp, nil
}
