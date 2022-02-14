package db

import (
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
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

// FactionStatGet return the stat by the given faction id
func FactionStatGet(ctx context.Context, conn Conn, factionStat *server.FactionStat) error {
	q := `
		SELECT * FROM faction_stats
		WHERE id = $1;
	`
	err := pgxscan.Get(ctx, conn, factionStat, q, factionStat.ID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// FactionStatAll return the stat of all factions
func FactionStatAll(ctx context.Context, conn Conn) ([]*server.FactionStat, error) {
	result := []*server.FactionStat{}
	q := `
		SELECT * FROM faction_stats;
	`
	err := pgxscan.Select(ctx, conn, &result, q)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}
