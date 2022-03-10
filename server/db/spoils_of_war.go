package db

import (
	"server/db/boiler"
	"server/gamedb"
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
	PlayerID        uuid.UUID
	TotalMultiplier decimal.Decimal
}

func PlayerMultipliers(battle_number int) ([]*Multipliers, error) {
	result := []*Multipliers{}
	q := `
SELECT p.id AS player_id, SUM(um.value) AS multiplier_sum FROM user_multipliers um 
INNER JOIN players p ON p.id = um.player_id
INNER JOIN multipliers m ON m.id = um.multiplier_id
WHERE um.from_battle_number <= $1
AND um.until_battle_number >= $1
GROUP BY p.id;
`

	err := pgxscan.Select(context.Background(), gamedb.Conn, &result, q, battle_number)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
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
