package db

import (
	"server/gamedb"
	"server/gamelog"
)

func FactionAddAbilityKillCount(factionID string) error {
	q := `
		UPDATE
			faction_stats
		SET
		kill_count = kill_count + 1
		WHERE
			id = $1
	`
	_, err := gamedb.StdConn.Exec(q, factionID)
	if err != nil {
		gamelog.L.Error().Str("faction_id", factionID).Err(err).Msg("Failed to update faction kill count")
		return err
	}

	return nil
}

func FactionSubtractAbilityKillCount(factionID string) error {
	q := `
		UPDATE
			faction_stats
		SET
		kill_count = kill_count - 1
		WHERE
			id = $1
	`
	_, err := gamedb.StdConn.Exec(q, factionID)
	if err != nil {
		gamelog.L.Error().Str("faction_id", factionID).Err(err).Msg("Failed to update faction kill count")
		return err
	}

	return nil
}

func FactionAddMechKillCount(factionID string) error {
	q := `
		UPDATE
			faction_stats
		SET
		mech_kill_count = mech_kill_count + 1
		WHERE
			id = $1
	`
	_, err := gamedb.StdConn.Exec(q, factionID)
	if err != nil {
		gamelog.L.Error().Str("faction_id", factionID).Err(err).Msg("Failed to update faction kill count")
		return err
	}

	return nil
}

func FactionAddDeathCount(factionID string) error {
	q := `
		UPDATE
			faction_stats
		SET
			death_count = death_count + 1
		WHERE
			id = $1
	`
	_, err := gamedb.StdConn.Exec(q, factionID)
	if err != nil {
		gamelog.L.Error().Str("faction_id", factionID).Err(err).Msg("Failed to update faction death count")
		return err
	}

	return nil
}

func FactionAddWinLossCount(winningFactionID string) error {
	q := `
		UPDATE
			faction_stats
		SET
			win_count = win_count + 1
		WHERE
			id = $1
	`
	_, err := gamedb.StdConn.Exec(q, winningFactionID)
	if err != nil {
		gamelog.L.Error().Str("winning_faction_id", winningFactionID).Err(err).Msg("Failed to update faction win count")
		return err
	}

	q = `
	UPDATE
		faction_stats
	SET
		loss_count = loss_count + 1
	WHERE
		id != $1
	`
	_, err = gamedb.StdConn.Exec(q, winningFactionID)
	if err != nil {
		gamelog.L.Error().Str("winning_faction_id", winningFactionID).Err(err).Msg("Failed to update faction loss count")
		return err
	}

	return nil
}

func FactionStatMVPUpdate(factionID string) error {
	q := `
		update 
			faction_stats fs2 
		set
			mvp_player_id = (
				select bc.player_id from battle_contributions bc 
					where bc.faction_id = fs2.id 
					group by player_id
					order by sum(amount) desc 
				limit 1
			)
		where 
			fs2.id = $1;
	`

	_, err := gamedb.StdConn.Exec(q, factionID)
	if err != nil {
		gamelog.L.Error().Str("faction_id", factionID).Err(err).Msg("Failed to update faction mvp player")
		return err
	}
	return nil
}
