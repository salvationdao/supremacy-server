package db

import (
	"fmt"
	"server"
	"server/gamelog"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
)

func MechMetadata(ctx context.Context, conn Conn, hash string) (server.UserID, server.FactionID, string, error) {
	gamelog.GameLog.Debug().Str("fn", "MechMetadata").Msg("db func")
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
	gamelog.GameLog.Debug().Str("fn", "IsInsured").Str("hash", hash).Msg("db func")
	var result bool
	q := `SELECT is_insured FROM battle_war_machine_queues WHERE war_machine_hash = $1`
	err := pgxscan.Get(ctx, tx, &result, q, hash)
	if err != nil {
		return false, fmt.Errorf("check is insured: %w", err)
	}
	return result, nil
}
func IsDefaultWarMachine(ctx context.Context, tx Conn, hash string) bool {
	gamelog.GameLog.Debug().Str("fn", "IsDefaultWarMachine").Str("hash", hash).Msg("db func")
	var count int
	q := `SELECT count(war_machine_hash) FROM battle_war_machine_queues WHERE war_machine_hash = $1`
	err := pgxscan.Get(ctx, tx, &count, q)
	if err != nil {
		gamelog.GameLog.Err(err).Msg("check if default war machine")
		return false
	}
	return count <= 0
}

func ContractRewardGet(ctx context.Context, tx Conn, hash string) (decimal.Decimal, error) {
	gamelog.GameLog.Debug().Str("fn", "ContractRewardGet").Str("hash", hash).Msg("db func")
	resultStr := struct {
		ContractReward string
	}{}
	q := `SELECT contract_reward FROM issued_contract_rewards WHERE war_machine_hash = $1`
	err := pgxscan.Get(ctx, tx, &resultStr, q, hash)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get contract reward: %w", err)
	}
	result, err := decimal.NewFromString(resultStr.ContractReward)
	if err != nil {
		return decimal.Zero, fmt.Errorf("process contract reward string to decimal: %w", err)
	}
	return result, nil
}
func ContractRewardInsert(ctx context.Context, tx Conn, battleID server.BattleID, reward decimal.Decimal, hash string) error {
	gamelog.GameLog.Debug().Str("fn", "ContractRewardInsert").Str("reward", reward.String()).Str("hash", hash).Str("battle_id", battleID.String()).Msg("db func")
	if reward.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("reward must be greater than 0")
	}
	q := `INSERT INTO issued_contract_rewards (battle_id, reward, war_machine_hash, is_paid) VALUES ($1, $2, $3, NULL);`
	_, err := tx.Exec(ctx, q, battleID, reward, hash)
	if err != nil {
		return fmt.Errorf("insert into issued_contract_rewards: %w", err)
	}
	return nil
}

func ContractRewardMarkIsPaid(ctx context.Context, tx Conn, battleID server.BattleID, hash string) error {
	gamelog.GameLog.Debug().Str("fn", "ContractRewardMarkIsPaid").Str("battle_id", battleID.String()).Msg("db func")
	q := `UPDATE issued_contract_rewards SET is_paid = NOW();`
	_, err := tx.Exec(ctx, q, battleID, hash)
	if err != nil {
		return fmt.Errorf("insert into issued_contract_rewards: %w", err)
	}
	return nil
}
