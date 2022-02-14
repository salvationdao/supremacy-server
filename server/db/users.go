package db

import (
	"context"
	"fmt"
	"server"

	"github.com/ninja-software/terror/v2"
)

func UserBattleViewRecord(ctx context.Context, conn Conn, userIDs []server.UserID) error {
	if len(userIDs) == 0 {
		return nil
	}

	q := `
		INSERT INTO 
			users (id)
		VALUES

	`

	for i, userID := range userIDs {
		q += fmt.Sprintf(`('%s')`, userID)
		if i < len(userID)-1 {
			q += ","
		}
	}

	q += `
		ON CONFLICT 
			(id)
		DO UPDATE SET
			view_battle_count += 1;

	`

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
