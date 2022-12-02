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

func GetPlayerMechSurvives(questEventID null.String) ([]*PlayerMechSurvives, error) {
	args := []interface{}{}
	whereClause := ""
	if questEventID.Valid {
		r, err := boiler.FindQuestEvent(gamedb.StdConn, questEventID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", questEventID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = fmt.Sprintf("WHERE %s BETWEEN $1 AND $2", boiler.BattleWinColumns.CreatedAt)
		args = append(args, r.StartedAt, r.EndAt)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), bw.mech_survive_count
		FROM (
			SELECT %[1]s, COUNT(%[2]s) AS mech_survive_count 
        	FROM %[3]s
			%[4]s
        	GROUP BY %[1]s 
        	ORDER BY COUNT(%[2]s) DESC 
        	LIMIT 100
		) bw
		INNER JOIN (
			SELECT %[5]s, %[6]s, %[7]s, %[8]s, %[9]s FROM %[10]s
		) p ON p.%[5]s = bw.%[1]s
        ORDER BY bw.mech_survive_count DESC;
    `,
		boiler.BattleWinColumns.OwnerID, // 1
		boiler.BattleWinColumns.MechID,  // 2
		boiler.TableNames.BattleWins,    // 3
		whereClause,                     // 4

		boiler.PlayerColumns.ID,        // 5
		boiler.PlayerColumns.Username,  // 6
		boiler.PlayerColumns.FactionID, // 7
		boiler.PlayerColumns.Gid,       // 8
		boiler.PlayerColumns.Rank,      // 9
		boiler.TableNames.Players,      // 10
	)
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
	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), ci.mechs_owned
		FROM (
		    SELECT %[1]s, COUNT(%[2]s) AS mechs_owned 
		    FROM %[3]s
			WHERE %[4]s = 'mech'
			GROUP BY %[1]s
			ORDER BY COUNT(%[2]s) DESC
			LIMIT 100
		) ci
		INNER JOIN (
		    SELECT %[5]s, %[6]s, %[7]s, %[8]s, %[9]s FROM %[10]s
		) p ON p.%[5]s = ci.%[1]s
		ORDER BY ci.mechs_owned DESC;
    `,
		boiler.CollectionItemColumns.OwnerID,  // 1
		boiler.CollectionItemColumns.ID,       // 2
		boiler.TableNames.CollectionItems,     // 3
		boiler.CollectionItemColumns.ItemType, // 4

		boiler.PlayerColumns.ID,        // 5
		boiler.PlayerColumns.Username,  // 6
		boiler.PlayerColumns.FactionID, // 7
		boiler.PlayerColumns.Gid,       // 8
		boiler.PlayerColumns.Rank,      // 9
		boiler.TableNames.Players,      // 10
	)
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
	q := fmt.Sprintf(`
		SELECT p.*
		FROM (
		    SELECT %[1]s, COUNT(%[2]s) AS ability_trigger_count
		    FROM %[3]s
		    WHERE %[4]s = $1
		    GROUP BY %[1]s
		    ORDER BY COUNT(%[2]s) DESC
		    LIMIT 2
		) bat
		INNER JOIN (
		    SELECT %[5]s, %[6]s, %[7]s FROM %[8]s
		) p ON p.%[5]s = bat.%[1]s
		ORDER BY bat.ability_trigger_count DESC;
	`,
		boiler.BattleAbilityTriggerColumns.PlayerID, // 1
		boiler.BattleAbilityTriggerColumns.ID,       // 2
		boiler.TableNames.BattleAbilityTriggers,     // 3
		boiler.BattleAbilityTriggerColumns.BattleID, // 4

		boiler.PlayerColumns.ID,        // 5
		boiler.PlayerColumns.Username,  // 6
		boiler.PlayerColumns.FactionID, // 7
		boiler.TableNames.Players,      // 8
	)

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

func TopBattleViewers(questEventID null.String) ([]*PlayerBattlesSpectated, error) {
	args := []interface{}{}
	whereClause := ""
	if questEventID.Valid {
		r, err := boiler.FindQuestEvent(gamedb.StdConn, questEventID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", questEventID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = fmt.Sprintf(`
			INNER JOIN %[1]s b ON b.%[2]s = bv.battle_id 
			WHERE b.%[3]s BETWEEN $1 AND $2
		`,
			boiler.TableNames.Battles,
			boiler.BattleColumns.ID,
			boiler.BattleColumns.StartedAt,
		)
		args = append(args, r.StartedAt, r.EndAt)
	}

	q := fmt.Sprintf(`
		select
	    	to_json(p.*),
 	   		bv.battle_count as view_battle_count
		from (
			select player_id, count(battle_id) as battle_count
			from %s bv
			%s
    		group by bv.player_id
    		order by count(bv.battle_id) DESC
    		limit 100
		)bv
		INNER JOIN (
			select %s, %s, %s, %s, %s from %s
		) p ON p.id = bv.player_id
		ORDER BY bv.battle_count DESC;
	`,
		boiler.TableNames.BattleViewers,
		whereClause,
		boiler.PlayerColumns.ID,
		boiler.PlayerColumns.Username,
		boiler.PlayerColumns.FactionID,
		boiler.PlayerColumns.Gid,
		boiler.PlayerColumns.Rank,
		boiler.TableNames.Players,
	)

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

func TopMechKillPlayers(questEventID null.String) ([]*PlayerMechKills, error) {
	args := []interface{}{}
	whereClause := ""
	if questEventID.Valid {
		r, err := boiler.FindQuestEvent(gamedb.StdConn, questEventID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", questEventID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = fmt.Sprintf("AND bh.%s BETWEEN $1 AND $2", boiler.BattleHistoryColumns.CreatedAt)
		args = append(args, r.StartedAt, r.EndAt)
	}

	q := fmt.Sprintf(`
		SELECT to_json(p.*), mkc.mech_kill_count
		FROM (
		    SELECT bm.%[1]s, count(bm.%[2]s) AS mech_kill_count
		    FROM %[3]s bh
		    INNER JOIN %[4]s bm ON bh.%[5]s = bm.%[6]s AND bm.%[2]s = bh.%[7]s
		    WHERE bh.%[8]s = 'killed' %[9]s
		    GROUP BY bm.%[1]s
		    ORDER BY count(bm.%[2]s) DESC
		    LIMIT 100
		) mkc
		INNER JOIN (
		    select %[10]s, %[11]s, %[12]s, %[13]s, %[14]s from %[15]s
		) p ON p.%[10]s = mkc.%[1]s
		ORDER BY mkc.mech_kill_count DESC;
	`,
		boiler.BattleMechColumns.PilotedByID,        // 1
		boiler.BattleMechColumns.MechID,             // 2
		boiler.TableNames.BattleHistory,             // 3
		boiler.TableNames.BattleMechs,               // 4
		boiler.BattleHistoryColumns.BattleID,        // 5
		boiler.BattleMechColumns.BattleID,           // 6
		boiler.BattleHistoryColumns.WarMachineTwoID, // 7
		boiler.BattleHistoryColumns.EventType,       // 8
		whereClause,                                 // 9

		boiler.PlayerColumns.ID,        // 10
		boiler.PlayerColumns.Username,  // 11
		boiler.PlayerColumns.FactionID, // 12
		boiler.PlayerColumns.Gid,       // 13
		boiler.PlayerColumns.Rank,      // 14
		boiler.TableNames.Players,      // 15
	)

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

func TopAbilityKillPlayers(questEventID null.String) ([]*PlayerAbilityKills, error) {
	args := []interface{}{}
	whereClause := ""
	if questEventID.Valid {
		r, err := boiler.FindQuestEvent(gamedb.StdConn, questEventID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", questEventID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}

		whereClause = fmt.Sprintf("WHERE pkg.%s BETWEEN $1 AND $2", boiler.PlayerKillLogColumns.CreatedAt)
		args = append(args, r.StartedAt, r.EndAt)
	}

	q := fmt.Sprintf(`
		SELECT to_json(p.*), pak.ability_kill_count
		FROM (
			SELECT pk.%[1]s, sum(pk.ability_kill_count) as ability_kill_count
			FROM (
				SELECT
					pkg.%[1]s,
					(
						CASE
							WHEN pkg.is_team_kill = FALSE THEN 
								count(pkg.%[2]s)
							ELSE 
								-count(pkg.%[2]s)
						END
					) AS ability_kill_count
				FROM %[3]s pkg
				%[5]s
				GROUP BY pkg.%[1]s, pkg.%[4]s
			) pk
			GROUP BY pk.%[1]s
			ORDER BY sum(pk.ability_kill_count) DESC
			LIMIT 100
		) pak
		INNER JOIN (
		    select %[6]s, %[7]s, %[8]s, %[9]s, %[10]s from %[11]s
		) p ON p.%[6]s = pak.%[1]s
		ORDER BY pak.ability_kill_count DESC;
	`,
		boiler.PlayerKillLogColumns.PlayerID,   // 1
		boiler.PlayerKillLogColumns.ID,         // 2
		boiler.TableNames.PlayerKillLog,        // 3
		boiler.PlayerKillLogColumns.IsTeamKill, // 4
		whereClause,                            // 5

		boiler.PlayerColumns.ID,        // 6
		boiler.PlayerColumns.Username,  // 7
		boiler.PlayerColumns.FactionID, // 8
		boiler.PlayerColumns.Gid,       // 9
		boiler.PlayerColumns.Rank,      // 10
		boiler.TableNames.Players,      // 11
	)

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

func TopAbilityTriggerPlayers(questEventID null.String) ([]*PlayerAbilityTriggers, error) {
	args := []interface{}{}
	whereClause := ""
	if questEventID.Valid {
		r, err := boiler.FindQuestEvent(gamedb.StdConn, questEventID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", questEventID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}
		whereClause = fmt.Sprintf("WHERE %s BETWEEN $1 AND $2", boiler.BattleAbilityTriggerColumns.TriggeredAt)
		args = append(args, r.StartedAt, r.EndAt)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), bat.ability_trigger_count AS total_ability_triggered
		FROM (
		    SELECT %[1]s, COUNT(%[2]s) as ability_trigger_count
		    FROM %[3]s
		    %[4]s
		    GROUP BY %[1]s
		    ORDER BY COUNT(%[2]s) DESC
		    LIMIT 100
		) bat
		INNER JOIN (
			SELECT %[5]s, %[6]s, %[7]s, %[8]s, %[9]s FROM %[10]s
		) p ON p.%[5]s = bat.%[1]s
		ORDER BY bat.ability_trigger_count DESC;
	`,
		boiler.BattleAbilityTriggerColumns.PlayerID, // 1
		boiler.BattleAbilityTriggerColumns.ID,       // 2
		boiler.TableNames.BattleAbilityTriggers,     // 3
		whereClause,                                 // 4

		boiler.PlayerColumns.ID,        // 5
		boiler.PlayerColumns.Username,  // 6
		boiler.PlayerColumns.FactionID, // 7
		boiler.PlayerColumns.Gid,       // 8
		boiler.PlayerColumns.Rank,      // 9
		boiler.TableNames.Players,      // 10
	)

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

func TopRepairBlockPlayers(questEventID null.String) ([]*PlayerRepairBlocks, error) {
	args := []interface{}{}
	whereClause := ""
	if questEventID.Valid {
		r, err := boiler.FindQuestEvent(gamedb.StdConn, questEventID.String)
		if err != nil {
			gamelog.L.Error().Str("round id", questEventID.String).Err(err).Msg("Failed to get round.")
			return nil, terror.Error(err, "Failed to query round.")
		}
		whereClause = fmt.Sprintf("AND %s BETWEEN $1 AND $2", boiler.RepairAgentColumns.FinishedAt)
		args = append(args, r.StartedAt, r.EndAt)
	}

	q := fmt.Sprintf(`
		SELECT TO_JSON(p.*), ra.total_block_repaired
		FROM (
		    SELECT %[1]s, COUNT(%[2]s) as total_block_repaired
		    FROM %[3]s
		    WHERE %[4]s = 'SUCCEEDED' AND %[5]s NOTNULL %[6]s
		    GROUP BY %[1]s
		    ORDER BY COUNT(%[2]s) DESC
		    LIMIT 100
		) ra
		INNER JOIN (
		    SELECT %[7]s, %[8]s, %[9]s, %[10]s, %[11]s FROM %[12]s
		) p ON p.%[7]s = ra.%[1]s
		ORDER BY ra.total_block_repaired DESC;
	`,
		boiler.RepairAgentColumns.PlayerID,       // 1
		boiler.RepairAgentColumns.ID,             // 2
		boiler.TableNames.RepairAgents,           // 3
		boiler.RepairAgentColumns.FinishedReason, // 4
		boiler.RepairAgentColumns.FinishedAt,     // 5
		whereClause,                              // 6

		boiler.PlayerColumns.ID,        // 7
		boiler.PlayerColumns.Username,  // 8
		boiler.PlayerColumns.FactionID, // 9
		boiler.PlayerColumns.Gid,       // 10
		boiler.PlayerColumns.Rank,      // 11
		boiler.TableNames.Players,      // 12
	)

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
