package db

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/shopspring/decimal"
)

type PlayerContribution struct {
	PlayerID   string          `db:"player_id"`
	Amount     decimal.Decimal `db:"amount"`
	DidTrigger bool            `db:"did_trigger"`
}

func GetPlayerContributions(abilityOfferingID string) ([]*boiler.BattleContribution, error) {
	q := `SELECT * from (
	 SELECT bc.player_id, sum(bc.amount) AS amount from battle_contributions bc 
	 WHERE bc.ability_offering_id = $1 
	 GROUP BY bc.player_id
	 ) bc1 LEFT JOIN LATERAL (
 	 SELECT bc.did_trigger from battle_contributions bc 
 	 WHERE bc.player_id = bc1.player_id and bc.ability_offering_id = $1
 	 ORDER BY bc.did_trigger DESC
 	 LIMIT 1
	 ) bc2 ON TRUE;
	`

	rows, err := gamedb.StdConn.Query(q, abilityOfferingID)
	if err != nil {
		gamelog.L.Error().Str("ability_offering_id", abilityOfferingID).Err(err).Msg("Failed to get players contribution")
		return nil, err
	}

	battleContribution := make([]*boiler.BattleContribution, 0)

	defer rows.Close()
	for rows.Next() {
		bc := &boiler.BattleContribution{}

		err := rows.Scan(&bc.PlayerID, &bc.Amount, &bc.DidTrigger)
		if err != nil {

			gamelog.L.Error().
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player contribtions into struct")
			return nil, err
		}

		battleContribution = append(battleContribution, bc)
	}

	return battleContribution, nil
}
