package db

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/net/context"
)

type Multipliers struct {
	PlayerID        uuid.UUID       `json:"player_id" db:"player_id"`
	TotalMultiplier decimal.Decimal `json:"total_multiplier" db:"multiplier_sum"`
}

// read multiplier from db
type PlayerMultiplier struct {
	PlayerID         uuid.UUID       `db:"player_id"`
	MultiplierValue  decimal.Decimal `db:"multiplier_value"`
	IsMultiplicative bool            `db:"is_multiplicative"`
}

func PlayerMultipliers(battleNumber int) ([]*Multipliers, error) {
	result := []*Multipliers{}

	dbResult := []*PlayerMultiplier{}
	q := `
		SELECT p.id AS player_id, um.value AS multiplier_value, m.is_multiplicative FROM user_multipliers um 
		INNER JOIN players p ON p.id = um.player_id
		INNER JOIN multipliers m ON m.id = um.multiplier_id
		WHERE um.from_battle_number <= $1
		AND um.until_battle_number > $1;
		`

	err := pgxscan.Select(context.Background(), gamedb.Conn, &dbResult, q, battleNumber)
	if err != nil {
		return nil, terror.Error(err)
	}

	// create temporary map the store player multipliers
	type multiplierStat struct {
		value               decimal.Decimal
		multiplicativeValue decimal.Decimal
	}

	playerMultiplierStatMap := make(map[uuid.UUID]*multiplierStat)
	for _, dr := range dbResult {
		if _, ok := playerMultiplierStatMap[dr.PlayerID]; !ok {
			playerMultiplierStatMap[dr.PlayerID] = &multiplierStat{
				value:               decimal.Zero,
				multiplicativeValue: decimal.Zero,
			}
		}

		// add the value, if the multipliers is not multiplicative
		if !dr.IsMultiplicative {
			playerMultiplierStatMap[dr.PlayerID].value = playerMultiplierStatMap[dr.PlayerID].value.Add(dr.MultiplierValue)
			continue
		}

		// increase multiplicative value
		playerMultiplierStatMap[dr.PlayerID].multiplicativeValue = playerMultiplierStatMap[dr.PlayerID].multiplicativeValue.Add(dr.MultiplierValue)
	}

	// calculate actual total multiplier amount
	for playerID, multiplierStat := range playerMultiplierStatMap {
		// set multiplicative to 1 if the value is zero
		if multiplierStat.multiplicativeValue.Equal(decimal.Zero) {
			multiplierStat.multiplicativeValue = decimal.NewFromInt(1)
		}

		result = append(result, &Multipliers{
			PlayerID:        playerID,
			TotalMultiplier: multiplierStat.value.Mul(multiplierStat.multiplicativeValue),
		})
	}
	return result, nil
}

func ExtendCitizenMulti(um *boiler.UserMultiplier) error {

	q := `
	update
		user_multipliers
	set
		until_battle_number = $4
	where 
		player_id = $1 and multiplier_id = $2 and from_battle_number = $3;
	`

	_, err := gamedb.StdConn.Exec(q, um.PlayerID, um.MultiplierID, um.FromBattleNumber, um.UntilBattleNumber)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func CitizenPlayerIDs(untilBattleNumber int) ([]uuid.UUID, error) {
	userIDs := []uuid.UUID{}

	q := `
	select um.player_id  from user_multipliers um 
	inner join multipliers m on m.id = um.multiplier_id and m."key" = 'citizen'
	where um.from_battle_number <= $1 and um.until_battle_number > $1
	`

	err := pgxscan.Select(context.Background(), gamedb.Conn, &userIDs, q, untilBattleNumber)
	if err != nil {
		return userIDs, terror.Error(err)
	}

	return userIDs, nil
}

func TotalSpoils() (decimal.Decimal, error) {
	result := struct{ Sum decimal.Decimal }{}
	q := `SELECT COALESCE(sum(amount), 0) as sum FROM battle_contributions WHERE processed_at IS NULL`
	err := pgxscan.Get(context.Background(), gamedb.Conn, &result, q)
	if err != nil {
		return decimal.Zero, terror.Error(err)
	}
	return result.Sum, nil
}

func Spoils(battleID string) ([]*boiler.BattleContribution, decimal.Decimal, error) {
	contributions, err := boiler.BattleContributions(qm.Where("battle_id = ?", battleID)).All(gamedb.StdConn)
	if err != nil {
		return nil, decimal.Zero, err
	}

	accumulator := decimal.Zero
	for _, contrib := range contributions {
		accumulator = accumulator.Add(contrib.Amount)
	}
	return contributions, accumulator, nil
}

func MarkAllContributionsProcessed() error {
	q := `UPDATE battle_contributions SET processed_at = NOW()`
	_, err := gamedb.Conn.Exec(context.Background(), q)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func MarkContributionProcessed(id uuid.UUID) error {
	tx, err := boiler.FindBattleContribution(gamedb.StdConn, id.String())
	if err != nil {
		return err
	}
	tx.ProcessedAt = null.TimeFrom(time.Now())
	_, err = tx.Update(gamedb.StdConn, boil.Whitelist(boiler.PendingTransactionColumns.ProcessedAt))
	if err != nil {
		return err
	}
	return nil
}

func UnprocessedPendingTransactions() ([]*boiler.PendingTransaction, error) {
	txes, err := boiler.PendingTransactions(
		boiler.PendingTransactionWhere.ProcessedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	return txes, nil
}
func MarkPendingTransactionProcessed(id uuid.UUID) error {
	tx, err := boiler.FindPendingTransaction(gamedb.StdConn, id.String())
	if err != nil {
		return err
	}
	tx.ProcessedAt = null.TimeFrom(time.Now())
	_, err = tx.Update(gamedb.StdConn, boil.Whitelist(boiler.PendingTransactionColumns.ProcessedAt))
	if err != nil {
		return err
	}
	return nil
}

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
	q := `
	 SELECT 
	  (sow.amount - sow.amount_sent) as amount
	 FROM 
	  spoils_of_war sow
	 ORDER BY 
	  sow.created_at DESC
	 LIMIT
	  2
	`

	rows, err := gamedb.StdConn.Query(q)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "LastTwoSpoilOfWarAmount").Err(err).Msg("unable to query spoils of war 2")
		return []decimal.Decimal{}, terror.Error(err)
	}
	defer rows.Close()

	result := make([]decimal.Decimal, 2)
	i := 0
	for rows.Next() {
		var amnt decimal.Decimal
		err := rows.Scan(&amnt)
		if err != nil {
			gamelog.L.Error().
				Str("db func", "LastTwoSpoilOfWarAmount").Err(err).Msg("unable to scan spoils of war 2")
			return nil, err
		}
		result[i] = amnt
		i++
	}

	return result, nil
}
