package db

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

	// insert into battle id

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
		SELECT 	us.id,
				COALESCE(us.view_battle_count,0) AS view_battle_count,
				COALESCE(us.total_vote_count,0) AS total_vote_count,
				COALESCE(us.total_ability_triggered,0) AS total_ability_triggered,
				COALESCE(us.kill_count,0) AS kill_count
		FROM user_stats us
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
		SELECT 
			us.id,
			COALESCE(us.view_battle_count,0) AS view_battle_count,
			COALESCE(us.total_vote_count,0) AS total_vote_count,
			COALESCE(us.total_ability_triggered,0) AS total_ability_triggered,
			COALESCE(us.kill_count,0) AS kill_count
		FROM user_stats us
		WHERE us.id = $1
	`

	err := pgxscan.Get(ctx, conn, user, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return user, nil
}

type dbMultiplier struct {
	Key           string `json:"key"`
	Value         int    `json:"value"`
	RemainSeconds int    `json:"remainSeconds"`
}

type CitizenTag string

const (
	CitizenTagSuperContributor    CitizenTag = "Super Contributor"    // top 10%
	CitizenTagContributor         CitizenTag = "Contributor"          // top 25%
	CitizenTagSupporter           CitizenTag = "Supporter"            // top 50%
	CitizenTagCitizen             CitizenTag = "Citizen"              // top 80%
	CitizenTagUnproductiveCitizen CitizenTag = "Unproductive Citizen" // other 20%
)

func (e CitizenTag) IsCitizen() bool {
	switch e {
	case CitizenTagSuperContributor,
		CitizenTagContributor,
		CitizenTagSupporter,
		CitizenTagCitizen,
		CitizenTagUnproductiveCitizen:
		return true
	}

	return false
}

// UserMultiplierStore store users' sups multipliers
func UserMultiplierStore(ctx context.Context, conn Conn, usm []*server.UserSupsMultiplierSend) error {
	if len(usm) == 0 {
		return nil
	}

	now := time.Now()
	var args []interface{}

	q := `
		INSERT INTO 
			users (id, sups_multipliers)
		VALUES
			
	`
	for i, us := range usm {
		// reformat the sups multipliers before store
		dbMultipliers := []*dbMultiplier{}
		for _, sm := range us.SupsMultipliers {
			if CitizenTag(sm.Key).IsCitizen() {
				continue
			}

			remainSecond := sm.ExpiredAt.Sub(now).Seconds()
			if remainSecond <= 0 {
				continue
			}
			dbMultipliers = append(dbMultipliers, &dbMultiplier{
				Key:           sm.Key,
				Value:         sm.Value,
				RemainSeconds: int(remainSecond),
			})
		}

		b, err := json.Marshal(dbMultipliers)
		if err != nil {
			return terror.Error(err)
		}

		args = append(args, b)

		q += fmt.Sprintf("('%s', $%d)", us.ToUserID, len(args))

		if i < len(usm)-1 {
			q += ","
			continue
		}
	}

	q += `
		ON CONFLICT (id) DO UPDATE SET sups_multipliers = EXCLUDED.sups_multipliers;
	`

	_, err := conn.Exec(ctx, q, args...)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// UserMultiplierGet read user
func UserMultiplierGet(ctx context.Context, conn Conn, userID server.UserID) ([]*server.SupsMultiplier, error) {
	dms := []*dbMultiplier{}

	q := `
		SELECT sups_multipliers FROM users WHERE id = $1
	`

	err := pgxscan.Get(ctx, conn, &dms, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}

	// reformat the sups multipliers
	result := []*server.SupsMultiplier{}
	now := time.Now()
	for _, dm := range dms {
		result = append(result, &server.SupsMultiplier{
			Key:       dm.Key,
			Value:     dm.Value,
			ExpiredAt: now.Add(time.Duration(dm.RemainSeconds) * time.Second),
		})
	}

	return result, nil
}

// ************************************
// Player
// ************************************

func UpsertPlayer(p *boiler.Player) error {
	err := p.Upsert(
		gamedb.StdConn,
		false,
		[]string{
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.PublicAddress,
			boiler.PlayerColumns.FactionID,
		},
		boil.None(),
		boil.Infer(),
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// PlayerFactionIDGet read user
func PlayerFactionIDGet(ctx context.Context, conn Conn, userID server.UserID) (*uuid.UUID, error) {
	var factionID uuid.UUID

	q := `
		SELECT faction_id FROM players WHERE id = $1
	`

	err := pgxscan.Get(ctx, conn, &factionID, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &factionID, nil
}
