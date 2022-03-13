package db

import (
	"context"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
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
			us.view_battle_count,
			us.total_ability_triggered,
			us.kill_count
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
			us.view_battle_count,
			us.total_ability_triggered,
			us.kill_count
		FROM user_stats us
		WHERE us.id = $1`

	err := pgxscan.Select(ctx, conn, userStat, q, userID)
	if err != nil {
		return nil, err
	}
	return userStat, nil
}

func UserStatAddKill(playerID string) (*boiler.UserStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, terror.Error(err)
	}

	userStat.KillCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.UserStatColumns.KillCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user kill count")
		return nil, terror.Error(err)
	}

	return userStat, nil
}

func UserStatAddTotalAbilityTriggered(playerID string) (*boiler.UserStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, terror.Error(err)
	}

	userStat.TotalAbilityTriggered += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.UserStatColumns.TotalAbilityTriggered))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user total ability triggered")
		return nil, terror.Error(err)
	}

	return userStat, nil
}

func UserStatAddViewBattleCount(playerID string) (*boiler.UserStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, terror.Error(err)
	}

	userStat.ViewBattleCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.UserStatColumns.ViewBattleCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user view battle count")
		return nil, terror.Error(err)
	}

	return userStat, nil
}

func UserStatQuery(playerID string) (*boiler.UserStat, error) {
	userStat, err := boiler.FindUserStat(gamedb.StdConn, playerID)
	if err != nil {
		gamelog.L.Warn().Str("player_id", playerID).Err(err).Msg("Failed to get user stat, creating a new user stat")

		userStat, err = UserStatCreate(playerID)
		if err != nil {
			gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to insert user stat")
			return nil, terror.Error(err)
		}
	}

	return userStat, nil
}

func UserStatCreate(playerID string) (*boiler.UserStat, error) {
	userStat := &boiler.UserStat{
		ID: playerID,
	}

	err := userStat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to insert user stat")
		return nil, terror.Error(err)
	}

	return userStat, nil
}
