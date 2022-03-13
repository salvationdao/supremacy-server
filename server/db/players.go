package db

import (
	"context"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// PlayerRegister new user who may or may not be enlisted
func PlayerRegister(ID uuid.UUID, Username string, FactionID uuid.UUID, PublicAddress common.Address) (*boiler.Player, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, fmt.Errorf("start tx: %w", err)
	}
	defer func() {
		tx.Rollback()
	}()
	exists, err := boiler.PlayerExists(tx, ID.String())
	if err != nil {
		return nil, err
	}
	var player *boiler.Player
	if exists {
		player, err = boiler.FindPlayer(tx, ID.String())
		if err != nil {
			return nil, err
		}
		player.PublicAddress = null.NewString(PublicAddress.Hex(), true)
		player.Username = null.NewString(Username, true)
		player.FactionID = null.NewString(FactionID.String(), !FactionID.IsNil())

		_, err = player.Update(tx, boil.Infer())
		if err != nil {
			return nil, err
		}
	} else {
		player = &boiler.Player{
			ID:            ID.String(),
			PublicAddress: null.NewString(PublicAddress.Hex(), true),
			Username:      null.NewString(Username, true),
			FactionID:     null.NewString(FactionID.String(), !FactionID.IsNil()),
		}
		err = player.Insert(tx, boil.Infer())
		if err != nil {
			return nil, err
		}
	}
	tx.Commit()
	return player, nil
}

func UserStatsAll(ctx context.Context, conn Conn) ([]*server.UserStat, error) {
	userStats := []*server.UserStat{}
	q := `
		SELECT 
			us.id,
			COALESCE(us.view_battle_count,0) AS view_battle_count,
			COALESCE(us.total_ability_triggered,0) AS total_ability_triggered,
			COALESCE(us.kill_count,0) AS kill_count
		FROM user_stats us`

	err := pgxscan.Select(ctx, conn, &userStats, q)
	if err != nil {
		return nil, err
	}
	return userStats, nil

}

func UserStatsGet(ctx context.Context, conn Conn, userID server.UserID) (*server.UserStat, error) {
	userStat := &server.UserStat{}
	q := `
		SELECT 
			us.id,
			COALESCE(us.view_battle_count,0) AS view_battle_count,
			COALESCE(us.total_ability_triggered,0) AS total_ability_triggered,
			COALESCE(us.kill_count,0) AS kill_count
		FROM user_stats us
		WHERE us.id = $1`

	err := pgxscan.Select(ctx, conn, userStat, q, userID)
	if err != nil {
		return nil, err
	}
	return userStat, nil
}

func UserStatsRecalculate(ctx context.Context, conn Conn, CurrentBattleID string) error {
	q := `
	INSERT INTO 
    user_stats (id, view_battle_count, kill_count, total_ability_triggered)
	SELECT 
	    p1.id, COALESCE(p2.view_battle_count,0), COALESCE(p4.kill_count,0), COALESCE(p3.total_ability_triggered,0)
	FROM(
	        SELECT p.id
	        FROM players p
	    ) p1
	        LEFT JOIN LATERAL (
	    	SELECT COUNT(*) AS view_battle_count FROM battles_viewers buv
	    	WHERE buv.player_id = p1.id AND buv.battle_id = $1
	    	GROUP BY buv.player_id 
	    ) p2 ON true 
	    	LEFT JOIN lateral(
	    	SELECT COUNT(bat.id ) AS total_ability_triggered FROM battle_ability_triggers bat 
	    	WHERE bat.player_id = p1.id AND bat.battle_id = $1
	    	GROUP by bat.player_id 
	    )p3 ON true
	    	LEFT JOIN lateral(
	    	SELECT COUNT(bh.id) AS kill_count FROM battle_history bh 
	    	INNER JOIN battle_mechs bm ON bm.mech_id = bh.war_machine_one_id AND bm.owner_id = p1.id
			WHERE bh.battle_id = $1
	    	GROUP BY bm.owner_id 
	    )p4 ON true
	ON CONFLICT DO UPDATE SET 
		view_battle_count = view_battle_count + EXCLUDED.view_battle_count
		total_ability_triggered = total_ability_triggered + EXCLUDED.total_ability_triggered
		kill_count = kill_count + EXCLUDED.kill_count
	`
	_, err := conn.Exec(ctx, q, CurrentBattleID)
	if err != nil {
		return err
	}

	return nil
}
