package db

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

type PlayerMechSurvives struct {
	Player       *boiler.Player `json:"player"`
	MechSurvives int            `db:"mech_survive_count" json:"mech_survive_count"`
}

func GetPlayerMechSurvives() ([]*PlayerMechSurvives, error) {
	q := `
        WITH bw AS (SELECT owner_id, COUNT(mech_id) AS mech_survive_count FROM battle_wins bw GROUP BY owner_id ORDER BY COUNT(mech_id) DESC LIMIT 10)
        SELECT p.id, p.username, p.faction_id, p.gid, p.rank, bw.mech_survive_count FROM players p
        INNER JOIN bw on p.id = bw.owner_id
        ORDER BY bw.mech_survive_count DESC;
    `
	rows, err := gamedb.StdConn.Query(q)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player mech survives.")
		return nil, err
	}

	defer rows.Close()

	resp := []*PlayerMechSurvives{}
	for rows.Next() {
		mechSurvive := &PlayerMechSurvives{
			Player: &boiler.Player{},
		}

		err := rows.Scan(&mechSurvive.Player.ID, &mechSurvive.Player.Username, &mechSurvive.Player.FactionID, &mechSurvive.Player.Gid, &mechSurvive.Player.Rank, &mechSurvive.MechSurvives)

		if err != nil {
			gamelog.L.Error().
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player contribtions into struct")
			return nil, err
		}

		resp = append(resp, mechSurvive)
	}

	return resp, nil
}

type PlayerBattleContributions struct {
	Player             *boiler.Player `json:"player"`
	TotalContributions string         `db:"total_contributions" json:"total_contributions"`
}

func GetPlayerBattleContributions() ([]*PlayerBattleContributions, error) {
	q := `
        WITH bc AS (SELECT player_id, SUM(amount) AS total_contributions from battle_contributions bc where bc.refund_transaction_id isnull group by player_id limit 10)
        SELECT p.id, p.username, p.faction_id, p.gid, p.rank, bc.total_contributions FROM players p
        INNER JOIN bc on p.id = bc.player_id
        ORDER BY bc.total_contributions DESC;
    `
	rows, err := gamedb.StdConn.Query(q)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player battle contributions.")
		return nil, err
	}

	defer rows.Close()

	resp := []*PlayerBattleContributions{}
	for rows.Next() {
		battleContributions := &PlayerBattleContributions{
			Player: &boiler.Player{},
		}

		err := rows.Scan(&battleContributions.Player.ID, &battleContributions.Player.Username, &battleContributions.Player.FactionID, &battleContributions.Player.Gid, &battleContributions.Player.Rank, &battleContributions.TotalContributions)

		if err != nil {
			gamelog.L.Error().
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player contribtions into struct")
			return nil, err
		}

		resp = append(resp, battleContributions)
	}

	return resp, nil
}
