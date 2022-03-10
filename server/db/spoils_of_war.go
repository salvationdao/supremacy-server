package db

import (
	"server/gamedb"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"golang.org/x/net/context"
)

type Multipliers struct {
	PlayerID        uuid.UUID
	TotalMultiplier int
}

func PlayerMultipliers(battle_number int) ([]*Multipliers, error) {
	result := []*Multipliers{}
	q := `
SELECT p.id AS player_id, SUM(value) AS multiplier_sum FROM user_multipliers um 
INNER JOIN players p ON p.id = um.player_id
INNER JOIN multipliers m ON m.id = um.multiplier
WHERE um.from_battle_number <= $1
AND um.until_battle_number >= $1
GROUP BY p.id;
`

	err := pgxscan.Get(context.Background(), gamedb.Conn, &result, q, battle_number)
	if err != nil {
		return nil, err
	}
	return result, nil
}
