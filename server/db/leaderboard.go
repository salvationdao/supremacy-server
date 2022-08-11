package db

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/shopspring/decimal"
)

type PlayerMechSurvives struct {
	Player       *boiler.Player `json:"player"`
	MechSurvives int            `db:"mech_survive_count" json:"mech_survive_count"`
}

func GetPlayerMechSurvives(startTime null.Time, endTime null.Time) ([]*PlayerMechSurvives, error) {
	args := []interface{}{}
	whereClause := ""
	if startTime.Valid && endTime.Valid {
		whereClause = "WHERE bw.created_at BETWEEN $1 AND $2"
		args = append(args, startTime.Time, endTime.Time)
	}

	q := fmt.Sprintf(`
        WITH bw AS (
        	SELECT owner_id, COUNT(mech_id) AS mech_survive_count 
        	FROM battle_wins bw 
			%s
        	GROUP BY owner_id 
        	ORDER BY COUNT(mech_id) DESC 
        	LIMIT 10
        )
        SELECT p.id, p.username, p.faction_id, p.gid, p.rank, bw.mech_survive_count FROM players p
        INNER JOIN bw on p.id = bw.owner_id
        ORDER BY bw.mech_survive_count DESC;
    `, whereClause)
	rows, err := gamedb.StdConn.Query(q, args...)
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
		WITH ci AS (
			SELECT owner_id, COUNT(id) AS mechs_owned 
			FROM collection_items ci 
			WHERE "item_type" = 'mech' 
			GROUP BY owner_id 
			ORDER BY COUNT(id) 
			DESC LIMIT 10
		)
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

func MostFrequentAbilityExecutors(battleID uuid.UUID) ([]*boiler.Player, error) {
	players := []*boiler.Player{}
	q := `
	 SELECT p.id, p.faction_id, p.username, p.public_address, p.is_ai, p.created_at
	 FROM battle_contributions bc
	 INNER JOIN players p ON p.id = bc.player_id
	 INNER JOIN factions f ON f.id = p.faction_id
	 WHERE battle_id = $1 AND did_trigger = TRUE
	 GROUP BY p.id ORDER BY COUNT(bc.id) DESC LIMIT 2;
	`

	rows, err := gamedb.StdConn.Query(q, battleID)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "TopSupsContributors").Err(err).Msg("unable to query factions")
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		pl := &boiler.Player{}
		err := rows.Scan(&pl.ID, &pl.FactionID, &pl.Username, &pl.PublicAddress, &pl.IsAi, &pl.CreatedAt)
		if err != nil {
			gamelog.L.Error().
				Str("db func", "TopSupsContributors").Err(err).Msg("unable to scan player into struct")
			return nil, err
		}
		players = append(players, pl)
	}

	return players, err
}

type PlayerBattlesSpectated struct {
	Player          *server.Player `json:"player"`
	ViewBattleCount int            `db:"view_battle_count" json:"view_battle_count"`
}

func TopBattleViewers(startTime, endTime null.Time) ([]*PlayerBattlesSpectated, error) {
	args := []interface{}{}
	whereClause := ""
	if startTime.Valid && endTime.Valid {
		whereClause = `
			INNER JOIN battles b ON b.id = bv.battle_id 
			WHERE b.started_at BETWEEN $1 AND $2
		`
		args = append(args, startTime.Time, endTime.Time)
	}

	q := fmt.Sprintf(`
		select
	    	to_json(p.*),
 	   		bv.battle_count as view_battle_count
		from (
			select player_id, count(battle_id) as battle_count
			from battle_viewers bv
			%s
    		group by bv.player_id
    		order by count(bv.battle_id) DESC
    		limit 10
		)bv
		INNER JOIN (
			select id, username, faction_id, gid, rank from players
		) p ON p.id = bv.player_id;
	`, whereClause)

	rows, err := gamedb.StdConn.Query(q, args...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player battles spectated.")
		return nil, terror.Error(err, "Failed to get leaderboard player battles spectated.")
	}

	resp := []*PlayerBattlesSpectated{}
	for rows.Next() {
		pbs := &PlayerBattlesSpectated{}
		err = rows.Scan(&pbs.Player, &pbs.ViewBattleCount)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player battle spectated from db.")
			return nil, terror.Error(err, "Failed to load player battles spectated.")
		}

		resp = append(resp, pbs)
	}

	return resp, nil
}

type PlayerMechKills struct {
	Player        *server.Player `json:"player"`
	MechKillCount int            `db:"mech_kill_count" json:"mech_kill_count"`
}

func TopMechKillPlayers(startTime null.Time, endTime null.Time) ([]*PlayerMechKills, error) {
	args := []interface{}{}
	whereClause := ""
	if startTime.Valid && endTime.Valid {
		whereClause = `
			AND bh.created_at BETWEEN $1 AND $2
		`
		args = append(args, startTime.Time, endTime.Time)
	}

	q := fmt.Sprintf(`
		SELECT to_json(p.*), mkc.mech_kill_count
		FROM (
		    SELECT bm.owner_id, count(bm.mech_id) AS mech_kill_count
		    FROM battle_history bh
		    INNER JOIN battle_mechs bm ON bh.battle_id = bm.battle_id AND bm.mech_id = bh.war_machine_two_id
		    WHERE bh.event_type = 'killed' %s
		    GROUP BY bm.owner_id
		    ORDER BY count(bm.mech_id) DESC
		    LIMIT 10
		) mkc
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = mkc.owner_id;
	`, whereClause)

	rows, err := gamedb.StdConn.Query(q, args...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player mech kill leaderboard.")
		return nil, terror.Error(err, "Failed to get player mech kill leaderboard.")
	}

	resp := []*PlayerMechKills{}
	for rows.Next() {
		pbs := &PlayerMechKills{}
		err = rows.Scan(&pbs.Player, &pbs.MechKillCount)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player mech kill count from db.")
			return nil, terror.Error(err, "Failed to load player mech kills count.")
		}

		resp = append(resp, pbs)
	}

	return resp, nil
}

type PlayerAbilityKills struct {
	Player           *server.Player `json:"player"`
	AbilityKillCount int            `db:"ability_kill_count" json:"ability_kill_count"`
}

func TopAbilityKillPlayers(startTime null.Time, endTime null.Time) ([]*PlayerAbilityKills, error) {
	args := []interface{}{}
	whereClause := ""
	if startTime.Valid && endTime.Valid {
		whereClause = `
			WHERE pkg.created_at BETWEEN $1 AND $2
		`
		args = append(args, startTime.Time, endTime.Time)
	}

	q := fmt.Sprintf(`
		SELECT to_json(p.*), pak.ability_kill_count
		FROM (
			SELECT pk.player_id, sum(pk.ability_kill_count) as ability_kill_count
			FROM (
				SELECT
					pkg.player_id,
					(
						CASE
							WHEN pkg.is_team_kill = FALSE THEN 
								count(pkg.id)
							ELSE 
								-count(pkg.id)
						END
					) AS ability_kill_count
				FROM player_kill_log pkg
				%s
				GROUP BY pkg.player_id, is_team_kill
			) pk
			GROUP BY pk.player_id
			ORDER BY sum(pk.ability_kill_count) DESC
			LIMIT 10
		) pak
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = pak.player_id
	`, whereClause)

	rows, err := gamedb.StdConn.Query(q, args...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player ability kill leaderboard.")
		return nil, terror.Error(err, "Failed to player ability kill leaderboard.")
	}

	resp := []*PlayerAbilityKills{}
	for rows.Next() {
		pbs := &PlayerAbilityKills{}
		err = rows.Scan(&pbs.Player, &pbs.AbilityKillCount)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player ability kills from db.")
			return nil, terror.Error(err, "Failed to load player ability kill count.")
		}

		resp = append(resp, pbs)
	}

	return resp, nil
}

type PlayerAbilityTriggers struct {
	Player                *server.Player `json:"player"`
	TotalAbilityTriggered int            `db:"total_ability_triggered" json:"total_ability_triggered"`
}

func TopAbilityTriggerPlayers(startTime null.Time, endTime null.Time) ([]*PlayerAbilityTriggers, error) {
	args := []interface{}{}
	whereClause := ""
	if startTime.Valid && endTime.Valid {
		whereClause = `
			WHERE triggered_at BETWEEN $1 AND $2
		`
		args = append(args, startTime.Time, endTime.Time)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), bat.ability_trigger_count
		FROM (
		    SELECT player_id, COUNT(id) as ability_trigger_count
		    FROM battle_ability_triggers
		    %s
		    GROUP BY player_id
		    ORDER BY COUNT(id) DESC
		    LIMIT 10
		) bat
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = bat.player_id
	`, whereClause)

	rows, err := gamedb.StdConn.Query(q, args...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player ability trigger leaderboard.")
		return nil, terror.Error(err, "Failed to get player ability trigger leaderboard.")
	}

	resp := []*PlayerAbilityTriggers{}
	for rows.Next() {
		pbs := &PlayerAbilityTriggers{}
		err = rows.Scan(&pbs.Player, &pbs.TotalAbilityTriggered)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player ability trigger count from db.")
			return nil, terror.Error(err, "Failed to load player ability trigger count.")
		}

		resp = append(resp, pbs)
	}

	return resp, nil
}
