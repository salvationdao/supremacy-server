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
	Player       *server.Player `json:"player"`
	MechSurvives int            `db:"mech_survive_count" json:"mech_survive_count"`
}

func GetPlayerMechSurvives(roundID null.String) ([]*PlayerMechSurvives, error) {
	args := []interface{}{}
	whereClause := ""
	if roundID.Valid {
		r, err := boiler.FindRound(gamedb.StdConn, roundID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", roundID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = "WHERE created_at BETWEEN $1 AND $2"
		args = append(args, r.StartedAt, r.Endat)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), bw.mech_survive_count
		FROM (
			SELECT owner_id, COUNT(mech_id) AS mech_survive_count 
        	FROM battle_wins 
			%s
        	GROUP BY owner_id 
        	ORDER BY COUNT(mech_id) DESC 
        	LIMIT 100
		) bw
		INNER JOIN (
			SELECT id, username, faction_id, gid, rank FROM players 
		) p ON p.id = bw.owner_id
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
		ms := &PlayerMechSurvives{}

		err := rows.Scan(&ms.Player, &ms.MechSurvives)

		if err != nil {
			gamelog.L.Error().
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player mech survives into struct")
			return nil, err
		}

		resp = append(resp, ms)
	}

	return resp, nil
}

type PlayerMechsOwned struct {
	Player     *server.Player  `json:"player"`
	MechsOwned decimal.Decimal `db:"mechs_owned" json:"mechs_owned"`
}

func GetPlayerMechsOwned() ([]*PlayerMechsOwned, error) {
	q := `
		SELECT TO_JSON(p.*), ci.mechs_owned
		FROM (
		    SELECT owner_id, COUNT(id) AS mechs_owned 
		    FROM collection_items
			WHERE item_type = 'mech'
			GROUP BY owner_id
			ORDER BY COUNT(id) DESC
			LIMIT 100
		) ci
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = ci.owner_id
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
		pmo := &PlayerMechsOwned{}

		err := rows.Scan(&pmo.Player, &pmo.MechsOwned)

		if err != nil {
			gamelog.L.Error().
				Str("db func", "GetPlayerContributions").Err(err).Msg("unable to scan player mechs owned into struct")
			return nil, err
		}

		resp = append(resp, pmo)
	}

	return resp, nil
}

func MostFrequentAbilityExecutors(battleID uuid.UUID) ([]*boiler.Player, error) {
	players := []*boiler.Player{}
	q := `
		SELECT p.id, p.username, p.faction_id
		FROM (
		    SELECT player_id, count(id) as ability_trigger_count
		    FROM battle_ability_triggers
		    WHERE battle_id = $1
		    GROUP BY player_id
		    ORDER BY COUNT(id) DESC
		    LIMIT 2
		) bat
		INNER JOIN (
		    SELECT id, username, faction_id  FROM players
		) p ON p.id = bat.player_id
		ORDER BY bat.ability_trigger_count DESC;
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
		err := rows.Scan(&pl.ID, &pl.Username, &pl.FactionID)
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

func TopBattleViewers(roundID null.String) ([]*PlayerBattlesSpectated, error) {
	args := []interface{}{}
	whereClause := ""
	if roundID.Valid {
		r, err := boiler.FindRound(gamedb.StdConn, roundID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", roundID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = `
			INNER JOIN battles b ON b.id = bv.battle_id 
			WHERE b.started_at BETWEEN $1 AND $2
		`
		args = append(args, r.StartedAt, r.Endat)
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
    		limit 100
		)bv
		INNER JOIN (
			select id, username, faction_id, gid, rank from players
		) p ON p.id = bv.player_id
		ORDER BY bv.battle_count DESC;
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

func TopMechKillPlayers(roundID null.String) ([]*PlayerMechKills, error) {
	args := []interface{}{}
	whereClause := ""
	if roundID.Valid {
		r, err := boiler.FindRound(gamedb.StdConn, roundID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", roundID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = "AND bh.created_at BETWEEN $1 AND $2"
		args = append(args, r.StartedAt, r.Endat)
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
		    LIMIT 100
		) mkc
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = mkc.owner_id
		ORDER BY mkc.mech_kill_count DESC;
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

func TopAbilityKillPlayers(roundID null.String) ([]*PlayerAbilityKills, error) {
	args := []interface{}{}
	whereClause := ""
	if roundID.Valid {
		r, err := boiler.FindRound(gamedb.StdConn, roundID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", roundID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = "WHERE pkg.created_at BETWEEN $1 AND $2"
		args = append(args, r.StartedAt, r.Endat)
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
			LIMIT 100
		) pak
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = pak.player_id
		ORDER BY pak.ability_kill_count DESC;
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

func TopAbilityTriggerPlayers(roundID null.String) ([]*PlayerAbilityTriggers, error) {
	args := []interface{}{}
	whereClause := ""
	if roundID.Valid {
		r, err := boiler.FindRound(gamedb.StdConn, roundID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", roundID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}
		whereClause = "WHERE triggered_at BETWEEN $1 AND $2"
		args = append(args, r.StartedAt, r.Endat)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), bat.ability_trigger_count AS total_ability_triggered
		FROM (
		    SELECT player_id, COUNT(id) as ability_trigger_count
		    FROM battle_ability_triggers
		    %s
		    GROUP BY player_id
		    ORDER BY COUNT(id) DESC
		    LIMIT 100
		) bat
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = bat.player_id
		ORDER BY bat.ability_trigger_count DESC;
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

type PlayerRepairBlocks struct {
	Player             *server.Player `json:"player"`
	TotalBlockRepaired int            `db:"total_block_repaired" json:"total_block_repaired"`
}

func TopRepairBlockPlayers(roundID null.String) ([]*PlayerRepairBlocks, error) {
	args := []interface{}{}
	whereClause := ""
	if roundID.Valid {
		r, err := boiler.FindRound(gamedb.StdConn, roundID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", roundID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}
		whereClause = "AND finished_at BETWEEN $1 AND $2"
		args = append(args, r.StartedAt, r.Endat)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), ra.total_block_repaired
		FROM (
		    SELECT player_id, COUNT(id) as total_block_repaired
		    FROM repair_agents
		    WHERE finished_reason = 'SUCCEEDED' AND finished_at NOTNULL %s
		    GROUP BY player_id
		    ORDER BY COUNT(id) DESC
		    LIMIT 100
		) ra
		INNER JOIN (
		    SELECT id, username, faction_id, gid, rank FROM players
		) p ON p.id = ra.player_id
		ORDER BY ra.total_block_repaired DESC;
	`, whereClause)

	rows, err := gamedb.StdConn.Query(q, args...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player ability trigger leaderboard.")
		return nil, terror.Error(err, "Failed to get player ability trigger leaderboard.")
	}

	resp := []*PlayerRepairBlocks{}
	for rows.Next() {
		pbs := &PlayerRepairBlocks{}
		err = rows.Scan(&pbs.Player, &pbs.TotalBlockRepaired)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player ability trigger count from db.")
			return nil, terror.Error(err, "Failed to load player ability trigger count.")
		}

		resp = append(resp, pbs)
	}

	return resp, nil
}
