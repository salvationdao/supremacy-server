package db

import (
	"context"
	"fmt"
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

func UserBattleViewUpsert(ctx context.Context, conn Conn, userIDs []server.UserID) error {
	if len(userIDs) == 0 {
		return nil
	}

	q := `
		INSERT INTO 
			users (id, view_battle_count)
		VALUES

	`

	for i, userID := range userIDs {
		q += fmt.Sprintf(`('%s', 1)`, userID)
		if i < len(userIDs)-1 {
			q += ","
		}
	}

	q += `
		ON CONFLICT 
			(id)
		DO UPDATE SET
			view_battle_count = users.view_battle_count + 1;
	`

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func UserBattleVoteCountInsert(ctx context.Context, conn Conn, battleID server.BattleID, userVoteCounts []*server.BattleUserVote) error {
	q := `
		INSERT INTO 
			battles_user_votes (battle_id, user_id, vote_count)
		VALUES

	`

	for i, buv := range userVoteCounts {
		q += fmt.Sprintf("('%s','%s',%d)", battleID, buv.UserID, buv.VoteCount)
		if i < len(userVoteCounts)-1 {
			q += ","
			continue
		}
	}

	q += " ON CONFLICT (battle_id, user_id) DO NOTHING"

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func UsersMostFrequentTriggerAbility(ctx context.Context, conn Conn, battleID server.BattleID) ([]*server.User, error) {
	users := []*server.User{}

	q := `
	SELECT 
		bega.triggered_by_user_id as id
	FROM 
		battle_events_game_ability bega 
	INNER JOIN 
		battle_events be ON be.id = bega.event_id AND 
							be.battle_id = $1
	WHERE 
		bega.is_triggered = true AND 
		bega.triggered_by_user_id NOTNULL
	GROUP BY 
		bega.triggered_by_user_id
	ORDER BY 
		COUNT(bega.id) DESC 
	LIMIT 5
	`

	err := pgxscan.Select(ctx, conn, &users, q, battleID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return users, nil
}

func UserStatMaterialisedViewRefresh(ctx context.Context, conn Conn) error {
	q := `
		REFRESH MATERIALIZED VIEW user_stats;
	`
	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func UserStatMany(ctx context.Context, conn Conn, userIDs []server.UserID) ([]*server.UserStat, error) {
	users := []*server.UserStat{}

	q := `
		SELECT * FROM user_stats us
		WHERE us.id IN (
	`

	for i, userID := range userIDs {
		q += fmt.Sprintf("'%s'", userID)
		if i < len(userIDs)-1 {
			q += ","
			continue
		}
		q += ")"
	}

	err := pgxscan.Select(ctx, conn, &users, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return users, nil
}

func UserStatGet(ctx context.Context, conn Conn, userID server.UserID) (*server.UserStat, error) {
	user := &server.UserStat{}

	q := `
		SELECT * FROM user_stats us
		WHERE us.id = $1
	`

	err := pgxscan.Get(ctx, conn, user, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return user, nil
}
