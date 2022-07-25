package db

import (
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

	hexPublicAddress := ""
	if PublicAddress != common.HexToAddress("") {
		hexPublicAddress = PublicAddress.Hex()
	}

	if exists {
		player, err = boiler.FindPlayer(tx, ID.String())
		if err != nil {
			return nil, err
		}

		player.PublicAddress = null.NewString(hexPublicAddress, hexPublicAddress != "")
		player.Username = null.NewString(Username, true)
		player.FactionID = null.NewString(FactionID.String(), !FactionID.IsNil())

		_, err = player.Update(tx, boil.Infer())
		if err != nil {
			return nil, err
		}
	} else {
		player = &boiler.Player{
			ID:            ID.String(),
			PublicAddress: null.NewString(hexPublicAddress, hexPublicAddress != ""),
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
	q := `SELECT MODE() WITHIN GROUP (ORDER BY lang) FROM chat_history WHERE player_id = $1 LIMIT 10;`
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
	us, err := boiler.FindPlayerStat(gamedb.StdConn, playerID)
	if err != nil {
		return nil, err
	}

	userStat := &server.UserStat{
		PlayerStat:         us,
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

func UserStatAddAbilityKill(playerID string) (*boiler.PlayerStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, err
	}

	userStat.AbilityKillCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerStatColumns.AbilityKillCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user kill count")
		return nil, err
	}

	return userStat, nil
}

func UserStatSubtractAbilityKill(playerID string) (*boiler.PlayerStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, err
	}

	userStat.AbilityKillCount -= 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerStatColumns.AbilityKillCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user kill count")
		return nil, err
	}

	return userStat, nil
}

func UserStatAddMechKill(playerID string) (*boiler.PlayerStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, err
	}

	userStat.MechKillCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerStatColumns.MechKillCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user kill count")
		return nil, err
	}

	return userStat, nil
}

func UserStatAddTotalAbilityTriggered(playerID string) (*boiler.PlayerStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, err
	}

	userStat.TotalAbilityTriggered += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerStatColumns.TotalAbilityTriggered))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user total ability triggered")
		return nil, err
	}

	return userStat, nil
}

func UserStatAddViewBattleCount(playerID string) (*boiler.PlayerStat, error) {
	userStat, err := UserStatQuery(playerID)
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to query user stat")
		return nil, err
	}

	userStat.ViewBattleCount += 1

	_, err = userStat.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerStatColumns.ViewBattleCount))
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to update user view battle count")
		return nil, err
	}

	return userStat, nil
}

func UserStatQuery(playerID string) (*boiler.PlayerStat, error) {
	userStat, err := boiler.FindPlayerStat(gamedb.StdConn, playerID)
	if err != nil {
		gamelog.L.Warn().Str("player_id", playerID).Err(err).Msg("Failed to get user stat, creating a new user stat")

		userStat, err = UserStatCreate(playerID)
		if err != nil {
			gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to insert user stat")
			return nil, err
		}
	}

	return userStat, nil
}

func UserStatCreate(playerID string) (*boiler.PlayerStat, error) {
	userStat := &boiler.PlayerStat{
		ID: playerID,
	}

	err := userStat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("player_id", playerID).Err(err).Msg("Failed to insert user stat")
		return nil, err
	}

	return userStat, nil
}

func PlayerFactionContributionList(battleID string, factionID string, abilityOfferingID string) ([]uuid.UUID, error) {
	playerList := []uuid.UUID{}
	q := `
		SELECT bc.player_id FROM battle_contributions bc 
			WHERE bc.battle_id = $1 AND bc.faction_id = $2 AND bc.ability_offering_id = $3
			GROUP BY player_id
		ORDER BY SUM(amount) DESC 
	`

	result, err := gamedb.StdConn.Query(q, battleID, factionID, abilityOfferingID)
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
func GetPositivePlayerAbilityKillByFactionID(factionID string) ([]*server.PlayerAbilityKills, error) {
	abilityKills, err := boiler.PlayerKillLogs(
		boiler.PlayerKillLogWhere.FactionID.EQ(factionID),
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

		playerAbilityKills = append(playerAbilityKills, &server.PlayerAbilityKills{playerID, factionID, killCount})
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

func PlayerIPUpsert(playerID string, ip string) error {
	if playerID == "" {
		return terror.Error(fmt.Errorf("missing player id"), "Missing player id")
	}

	if ip == "" {
		return terror.Error(fmt.Errorf("missing ip"), "Missing ip")
	}

	q := `
		INSERT INTO 
			player_ips (player_id, ip, first_seen_at, last_seen_at)
		VALUES 
			($1, $2, NOW(), NOW())
		ON CONFLICT 
			(player_id, ip) 
		DO UPDATE SET
			last_seen_at = NOW()
	`

	_, err := gamedb.StdConn.Exec(q, playerID, ip)
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Str("ip", ip).Err(err).Msg("Failed to upsert player ip")
		return terror.Error(err, "Failed to store player ip")
	}

	return nil
}

func GetPlayer(playerID string) (*server.Player, error) {
	player, err := boiler.FindPlayer(gamedb.StdConn, playerID)
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to get player")
		return nil, terror.Error(err, "Failed to get player")
	}
	features, err := GetPlayerFeaturesByID(player.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to find features")
		return nil, err
	}

	serverPlayer := server.PlayerFromBoiler(player, features)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player by ID")
		return nil, err
	}
	return serverPlayer, nil
}
