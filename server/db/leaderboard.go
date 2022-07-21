package db

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/shopspring/decimal"
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
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player mech survives into struct")
			return nil, err
		}

		resp = append(resp, mechSurvive)
	}

	return resp, nil
}

type PlayerMechsOwned struct {
	Player     *boiler.Player  `json:"player"`
	MechsOwned decimal.Decimal `db:"mechs_owned" json:"mechs_owned"`
}

func GetPlayerMechsOwned() ([]*PlayerMechsOwned, error) {
	q := `
        WITH ci AS (SELECT owner_id, COUNT(id) AS mechs_owned FROM collection_items ci WHERE "item_type" = 'mech' GROUP BY owner_id LIMIT 10)
        SELECT p.id, p.username, p.faction_id, p.gid, p.rank, ci.mechs_owned FROM players p
        INNER JOIN ci on p.id = ci.owner_id
        ORDER BY ci.mechs_owned DESC;
    `
	rows, err := gamedb.StdConn.Query(q)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player mechs owned.")
		return nil, err
	}

	defer rows.Close()

	resp := []*PlayerMechsOwned{}
	for rows.Next() {
		battleContributions := &PlayerMechsOwned{
			Player: &boiler.Player{},
		}

		err := rows.Scan(&battleContributions.Player.ID, &battleContributions.Player.Username, &battleContributions.Player.FactionID, &battleContributions.Player.Gid, &battleContributions.Player.Rank, &battleContributions.MechsOwned)

		if err != nil {
			gamelog.L.Error().
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player mechs owned into struct")
			return nil, err
		}

		resp = append(resp, battleContributions)
	}

	return resp, nil
}
