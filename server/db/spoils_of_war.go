package db

import (
	"database/sql"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/friendsofgo/errors"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

func TopSupsContributeFactions(battleID uuid.UUID) ([]*boiler.Faction, error) {
	q := `
		SELECT f.id, f.vote_price, f.contract_reward, f.label, f.guild_id, f.created_at, f.primary_color, f.secondary_color, f.background_color
		FROM battle_contributions bc 
		INNER JOIN players p ON p.id = bc.player_id
		INNER JOIN factions f ON f.id = p.faction_id
		WHERE battle_id = $1 
		GROUP BY f.id
		ORDER BY SUM(amount) DESC LIMIT 2;
	`

	rows, err := gamedb.StdConn.Query(q, battleID)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "TopSupsContributeFactions").Err(err).Msg("unable to query factions")
		return nil, err
	}

	defer rows.Close()

	factions := make([]*boiler.Faction, 0)

	for rows.Next() {
		f := &boiler.Faction{}

		err := rows.Scan(&f.ID, &f.VotePrice, &f.ContractReward, &f.Label, &f.GuildID, &f.CreatedAt, &f.PrimaryColor, &f.SecondaryColor, &f.BackgroundColor)
		if err != nil {

			gamelog.L.Error().
				Str("db func", "TopSupsContributeFactions").Err(err).Msg("unable to scan faction into struct")
			return nil, err
		}

		factions = append(factions, f)
	}

	return factions, nil
}

func TopSupsContributors(battleID uuid.UUID) ([]*boiler.Player, error) {
	players := []*boiler.Player{}
	q := `
	 SELECT p.id, p.faction_id, p.username, p.public_address, p.is_ai, p.created_at
	 FROM battle_contributions bc
	 INNER JOIN players p ON p.id = bc.player_id
	 INNER JOIN factions f ON f.id = p.faction_id
	 WHERE battle_id = $1
	 GROUP BY p.id ORDER BY SUM(amount) DESC LIMIT 2;
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

func LastTwoSpoilOfWarAmount() ([]decimal.Decimal, error) {
	result := []decimal.Decimal{}

	// get the latest sow
	latestSOW, err := boiler.SpoilsOfWars(
		qm.Select(
			boiler.SpoilsOfWarColumns.ID,
			boiler.SpoilsOfWarColumns.Amount,
			boiler.SpoilsOfWarColumns.AmountSent,
		),
		qm.OrderBy(boiler.SpoilsOfWarColumns.BattleNumber+" DESC"),
		qm.Limit(1),
	).One(gamedb.StdConn)
	if err != nil {
		return []decimal.Decimal{}, terror.Error(err, "Failed to get the latest spoil of war")
	}

	// append the latest spoil of war to the result
	result = append(result, latestSOW.Amount.Sub(latestSOW.AmountSent))

	// unfinished spoil of wor
	unfinishedSOWs, err := boiler.SpoilsOfWars(
		qm.Select(
			boiler.SpoilsOfWarColumns.ID,
			boiler.SpoilsOfWarColumns.Amount,
			boiler.SpoilsOfWarColumns.AmountSent,
			boiler.SpoilsOfWarColumns.LeftoverAmount,
		),
		boiler.SpoilsOfWarWhere.CreatedAt.GT(time.Now().AddDate(0, 0, -1)),
		boiler.SpoilsOfWarWhere.ID.NEQ(latestSOW.ID),
		boiler.SpoilsOfWarWhere.LeftoversTransactionID.IsNull(),
		qm.And("amount > amount_sent"),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return []decimal.Decimal{}, terror.Error(err, "Failed to get the unfinished spoil of war")
	}

	totalRemain := decimal.Zero
	for _, unfinishedSOW := range unfinishedSOWs {
		totalRemain = totalRemain.Add(unfinishedSOW.Amount.Sub(unfinishedSOW.AmountSent).Sub(unfinishedSOW.LeftoverAmount))
	}

	// append total remain spoil of war amount
	result = append(result, totalRemain)

	return result, nil
}
