package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
)

// FactionCreate create a new faction
func FactionCreate(ctx context.Context, conn Conn, faction *server.Faction) error {
	q := `
		INSERT INTO
			factions (id, vote_price)
		VALUES
			($1, $2)
		RETURNING
			id, vote_price
	`

	err := pgxscan.Get(ctx, conn, faction, q, faction.ID, faction.VotePrice)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionVotePriceGet create a new faction
func FactionVotePriceGet(ctx context.Context, conn Conn, faction *server.Faction) error {
	q := `
		SELECT vote_price FROM factions
		WHERE id = $1
	`

	err := pgxscan.Get(ctx, conn, faction, q, faction.ID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionVotePriceUpdate create a new faction
func FactionVotePriceUpdate(ctx context.Context, conn Conn, faction *server.Faction) error {
	q := `
		UPDATE
			factions
		SET
			vote_price = $2
		WHERE
			id = $1
	`

	_, err := conn.Exec(ctx, q, faction.ID, faction.VotePrice)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionContractRewardUpdate create a new faction
func FactionContractRewardUpdate(ctx context.Context, conn Conn, factionID server.FactionID, contractReward string) error {
	q := `
		UPDATE
			factions
		SET
			contract_reward = $2
		WHERE
			id = $1
	`

	_, err := conn.Exec(ctx, q, factionID, contractReward)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionStatMaterialisedViewRefresh
func FactionStatMaterialisedViewRefresh(ctx context.Context, conn Conn) error {
	q := `
		REFRESH MATERIALIZED VIEW faction_stats;
	`
	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// // FactionStatGet return the stat by the given faction id
// func FactionStatGet(ctx context.Context, conn Conn, factionStat *server.FactionStat) error {
// 	q := `
// 		SELECT * FROM faction_stats
// 		WHERE id = $1;
// 	`
// 	err := pgxscan.Get(ctx, conn, factionStat, q, factionStat.ID)
// 	if err != nil {
// 		return terror.Error(err)
// 	}
// 	return nil
// }

// // FactionStatAll return the stat of all factions
// func FactionStatAll(ctx context.Context, conn Conn) ([]*server.FactionStat, error) {
// 	result := []*server.FactionStat{}
// 	q := `
// 		SELECT * FROM faction_stats;
// 	`
// 	err := pgxscan.Select(ctx, conn, &result, q)
// 	if err != nil {
// 		return nil, terror.Error(err)
// 	}
// 	return result, nil
// }

func FactionAll(ctx context.Context, conn Conn) (boiler.FactionSlice, error) {
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	return factions, nil
}

func FactionGet(factionID string) (*boiler.Faction, error) {
	faction, err := boiler.Factions(boiler.FactionWhere.ID.EQ(factionID)).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	return faction, nil
}

func FactionAddContribute(factionID string, amount decimal.Decimal) error {
	// NOTE: faction contribution only show integer in frontend, so just store the actual sups amount
	storeAmount := amount.Div(decimal.New(1, 18)).IntPart()

	q := `
		UPDATE
			faction_stats
		SET
			sups_contribute = sups_contribute + $2
		WHERE
			id = $1
	`
	_, err := gamedb.StdConn.Exec(q, factionID, storeAmount)
	if err != nil {
		gamelog.L.Error().Str("faction_id", factionID).Str("amount", amount.String()).Err(err).Msg("Failed to update faction contribution")
		return terror.Error(err)
	}

	return nil
}

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
		return terror.Error(err)
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
		return terror.Error(err)
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
		return terror.Error(err)
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
		return terror.Error(err)
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
		return terror.Error(err)
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
		return terror.Error(err)
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
