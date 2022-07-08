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
        WITH bw as (SELECT owner_id, COUNT(mech_id) as mech_survive_count FROM battle_wins bw GROUP BY owner_id ORDER BY COUNT(mech_id) DESC LIMIT 10)
        SELECT p.id, p.username, p.faction_id, p.gid, p.rank, bw.mech_survive_count FROM players p
        INNER JOIN bw on p.id = bw.owner_id;
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
