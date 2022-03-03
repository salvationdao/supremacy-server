package db

import (
	"errors"
	"fmt"
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
)

func MechMetadata(ctx context.Context, conn Conn, hash string) (server.UserID, server.FactionID, string, error) {
	result := struct {
		OwnedByID server.UserID
		FactionID server.FactionID
		Name      string
	}{}
	q := `
	SELECT 
		war_machine_metadata -> 'ownedByID' AS ownedByID,
		war_machine_metadata -> 'factionID' AS factionID,
		war_machine_metadata -> 'name' AS name 
	FROM battle_war_machine_queues WHERE war_machine_hash = $1;
`
	err := pgxscan.Get(ctx, conn, &result, q, hash)
	return result.OwnedByID, result.FactionID, result.Name, err
}
func IsInsured(ctx context.Context, tx Conn, hash string) (bool, error) {
	var result bool
	q := `SELECT is_insured FROM battle_war_machine_queues WHERE war_machine_hash = $1`
	err := pgxscan.Get(ctx, tx, &result, q, hash)
	if err != nil {
		return false, fmt.Errorf("check is insured: %w", err)
	}
	return result, nil
}
func IsDefaultWarMachine(ctx context.Context, tx Conn, hash string) bool {
	_, err := ContractRewardGet(ctx, tx, hash)
	return errors.Is(err, pgx.ErrNoRows)
}

func ContractRewardGet(ctx context.Context, tx Conn, hash string) (decimal.Decimal, error) {
	var result decimal.Decimal
	q := `SELECT contract_reward FROM battle_war_machine_queues WHERE war_machine_hash = $1`
	err := pgxscan.Get(ctx, tx, &result, q, hash)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get contract reward: %w", err)
	}
	return result, nil
}
func ContractRewardInsert(ctx context.Context, tx Conn, battleID server.BattleID, reward decimal.Decimal, hash string) error {
	q := `INSERT INTO issued_contract_rewards (battle_id, reward, war_machine_hash) VALUES ($1, $2, $3);`
	_, err := tx.Exec(ctx, q, battleID, reward, hash)
	if err != nil {
		return err
	}
	return nil
}
