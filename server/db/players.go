package db

import (
	"context"
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err, "Failed to commit db transaction")
	}
	return player, nil
}

func GetUserLanguage(playerID string) string {
	q := `SELECT mode() within group (order by lang) from chat_history WHERE player_id = $1 LIMIT 10;`
	row := gamedb.StdConn.QueryRow(q, playerID)
	lang := "English"
	switch err := row.Scan(&lang); err {
	case sql.ErrNoRows:
		return "English"
	case nil:
		return lang
	default:
		return "English"
	}
}
func UserStatsGet(playerID string) (*server.UserStat, error) {
	us, err := boiler.FindUserStat(gamedb.StdConn, playerID)
	if err != nil {
		return nil, err
	}

	userStat := &server.UserStat{
		UserStat:           us,
		LastSevenDaysKills: 0,
	}

	// get last seven days kills
	abilityKills, err := boiler.PlayerKillLogs(
		boiler.PlayerKillLogWhere.PlayerID.EQ(playerID),
		boiler.PlayerKillLogWhere.CreatedAt.GT(time.Now().AddDate(0, 0, -7)),
	).All(gamedb.StdConn)
	if err != nil {
		return userStat, nil
	}

	for _, abilityKill := range abilityKills {
		if abilityKill.IsTeamKill {
			userStat.LastSevenDaysKills -= 1
			continue
		}
		userStat.LastSevenDaysKills += 1
	}

	return userStat, nil
}

func UserStatAddAbilityKill(playerID string) (*boiler.UserStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, terror.Error(err)
	}

	userStat.AbilityKillCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.UserStatColumns.AbilityKillCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user kill count")
		return nil, terror.Error(err)
	}

	return userStat, nil
}

func UserStatSubtractAbilityKill(playerID string) (*boiler.UserStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, terror.Error(err)
	}

	userStat.AbilityKillCount -= 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.UserStatColumns.AbilityKillCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user kill count")
		return nil, terror.Error(err)
	}

	return userStat, nil
}

func UserStatAddMechKill(playerID string) (*boiler.UserStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, terror.Error(err)
	}

	userStat.MechKillCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.UserStatColumns.MechKillCount))
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

func PlayerFactionContributionList(battleID string, factionID string, abilityOfferingID string) ([]uuid.UUID, error) {
	playerList := []uuid.UUID{}
	q := `
		select bc.player_id from battle_contributions bc 
			where bc.battle_id = $1 and bc.faction_id = $2 and bc.ability_offering_id = $3
			group by player_id
		order by sum(amount) desc 
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	result, err := gamedb.Conn.Query(ctx, q, battleID, factionID, abilityOfferingID)
	if err != nil {
		gamelog.L.Error().Str("battle_id", battleID).Str("faction_id", factionID).Err(err).Msg("failed to get player list from db")
		return []uuid.UUID{}, err
	}

	defer result.Close()

	for result.Next() {
		var idStr string
		err = result.Scan(
			&idStr,
		)
		if err != nil {
			gamelog.L.Error().Str("battle_id", battleID).Str("faction_id", factionID).Err(err).Msg("failed to scan from result ")
			return []uuid.UUID{}, err
		}

		playerID, err := uuid.FromString(idStr)
		if err != nil {
			gamelog.L.Error().Str("battle_id", battleID).Str("faction_id", factionID).Err(err).Msg("failed to convert from result")
			return []uuid.UUID{}, err
		}

		playerList = append(playerList, playerID)
	}

	return playerList, nil
}

// GetPositivePlayerAbilityKillByFactionID return player ability kill by given faction id
func GetPositivePlayerAbilityKillByFactionID(factionID server.FactionID) ([]*server.PlayerAbilityKills, error) {
	abilityKills, err := boiler.PlayerKillLogs(
		boiler.PlayerKillLogWhere.FactionID.EQ(factionID.String()),
		boiler.PlayerKillLogWhere.CreatedAt.GT(time.Now().AddDate(0, 0, -7)),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get player ability kills from db")
	}

	// build ability kill map
	playerAbilityKillMap := make(map[string]int64)
	for _, abilityKill := range abilityKills {
		pak, ok := playerAbilityKillMap[abilityKill.PlayerID]
		if !ok {
			pak = 0
		}

		if !abilityKill.IsTeamKill {
			pak++
		} else {
			pak--
		}
		playerAbilityKillMap[abilityKill.PlayerID] = pak
	}

	playerAbilityKills := []*server.PlayerAbilityKills{}
	for playerID, killCount := range playerAbilityKillMap {
		if killCount <= 0 {
			continue
		}

		playerAbilityKills = append(playerAbilityKills, &server.PlayerAbilityKills{playerID, factionID.String(), killCount})
	}

	return playerAbilityKills, nil
}

func UpdatePunishVoteCost() error {
	// update punish vote cost
	q := `
		UPDATE
			players
		SET
			issue_punish_fee = issue_punish_fee / 2
		WHERE
			issue_punish_fee > 10
	`

	_, err := gamedb.StdConn.Exec(q)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update players' punish vote cost")
		return terror.Error(err, "Failed to update players' punish vote cost")
	}

	// update report cost
	q = `
		UPDATE
			players
		SET 
			reported_cost = reported_cost / 2
		WHERE
			reported_cost > 10
	`
	_, err = gamedb.StdConn.Exec(q)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update players' report cost")
		return terror.Error(err, "Failed to update players' report cost")
	}

	return nil
}
