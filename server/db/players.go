package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	l := gamelog.L.With().Str("dbFunc", "GetPlayer").Str("playerID", playerID).Logger()

	player, err := boiler.FindPlayer(gamedb.StdConn, playerID)
	if err != nil {
		l.Error().Err(err).Msg("unable to find player")
		return nil, err
	}

	l = l.With().Interface("GetPlayerFeaturesByID", playerID).Logger()
	features, err := GetPlayerFeaturesByID(player.ID)
	if err != nil {
		l.Error().Err(err).Msg("unable to find player's features")
		return nil, err
	}

	l = l.With().Interface("PlayerFromBoiler", playerID).Logger()
	serverPlayer := server.PlayerFromBoiler(player, features)
	if err != nil {
		l.Error().Err(err).Msg("unable to create server player struct")
		return nil, err
	}
	return serverPlayer, nil
}

func GetPublicPlayerByID(playerID string) (*server.PublicPlayer, error) {
	l := gamelog.L.With().Str("dbFunc", "GetPlayer").Str("playerID", playerID).Logger()

	player, err := boiler.FindPlayer(gamedb.StdConn, playerID)
	if err != nil {
		l.Error().Err(err).Msg("unable to find player")
		return nil, err
	}

	pp := &server.PublicPlayer{
		ID:        player.ID,
		Username:  player.Username,
		Gid:       player.Gid,
		FactionID: player.FactionID,
		AboutMe:   player.AboutMe,
		Rank:      player.Rank,
		CreatedAt: player.CreatedAt,
	}

	return pp, nil
}

func PlayerQuestStatGet(playerID string) ([]*server.QuestStat, error) {
	result := []*server.QuestStat{}

	err := boiler.NewQuery(
		qm.Select(
			fmt.Sprintf("%s AS id", qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.ID)),
			fmt.Sprintf("%s AS round_name", qm.Rels(boiler.TableNames.Rounds, boiler.RoundColumns.Name)),
			fmt.Sprintf("%s AS name", qm.Rels(boiler.TableNames.BlueprintQuests, boiler.BlueprintQuestColumns.Name)),
			fmt.Sprintf("%s AS key", qm.Rels(boiler.TableNames.BlueprintQuests, boiler.BlueprintQuestColumns.Key)),
			fmt.Sprintf("%s AS description", qm.Rels(boiler.TableNames.BlueprintQuests, boiler.BlueprintQuestColumns.Description)),
			fmt.Sprintf(
				`COALESCE(
    		    		(SELECT TRUE FROM %s WHERE %s = %s AND %s = '%s'),
    		    		FALSE
    				) AS obtained
				`,
				boiler.TableNames.PlayersObtainedQuests,
				qm.Rels(boiler.TableNames.PlayersObtainedQuests, boiler.PlayersObtainedQuestColumns.ObtainedQuestID),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.ID),
				qm.Rels(boiler.TableNames.PlayersObtainedQuests, boiler.PlayersObtainedQuestColumns.PlayerID),
				playerID,
			),
		),
		qm.From(
			fmt.Sprintf(
				"(SELECT * FROM %[1]s WHERE %[2]s ISNULL AND %[3]s ISNULL) %[1]s",
				boiler.TableNames.Quests,
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.ExpiresAt),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.DeletedAt),
			),
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.BlueprintQuests,
				qm.Rels(boiler.TableNames.BlueprintQuests, boiler.BlueprintQuestColumns.ID),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.BlueprintID),
			),
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.Rounds,
				qm.Rels(boiler.TableNames.Rounds, boiler.RoundColumns.ID),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.RoundID),
			),
		),
	).Bind(context.Background(), gamedb.StdConn, &result)
	if err != nil {
		gamelog.L.Error().Err(err).Str("player id", playerID).Msg("Failed to get player quests.")
		return nil, terror.Error(err, "Failed to get player quests.")
	}

	fmt.Println(result)

	return result, nil
}

func PlayerMechKillCount(playerID string, afterTime time.Time) (int, error) {
	q := fmt.Sprintf(`
		SELECT COUNT(bh.id) 
		FROM (
		    SELECT * FROM %[1]s WHERE %[2]s = 'killed' AND %[3]s NOTNULL AND %[4]s >= $2
		) bh
		INNER JOIN %[5]s bm ON bh.%[6]s = bm.%[7]s AND bm.%[8]s = bh.%[3]s AND bm.%[9]s = $1;
	`,
		boiler.TableNames.BattleHistory,             // 1
		boiler.BattleHistoryColumns.EventType,       // 2
		boiler.BattleHistoryColumns.WarMachineTwoID, // 3
		boiler.BattleHistoryColumns.CreatedAt,       // 4
		boiler.TableNames.BattleMechs,               // 5
		boiler.BattleHistoryColumns.BattleID,        // 6
		boiler.BattleMechColumns.BattleID,           // 7
		boiler.BattleMechColumns.MechID,             // 8
		boiler.BattleMechColumns.OwnerID,            // 9
	)

	mechKillCount := 0
	err := gamedb.StdConn.QueryRow(q, playerID, afterTime).Scan(&mechKillCount)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to scan player mech kill count")
		return 0, terror.Error(err, "Failed to get player mech kill count")
	}

	return mechKillCount, nil
}

func PlayerTotalBattleMechCommanderUsed(playerID string, startFromTime time.Time) (int, error) {
	q := fmt.Sprintf(
		"SELECT COUNT(DISTINCT %s) FROM %s WHERE %s = $1 AND %s >= $2;",
		boiler.MechMoveCommandLogColumns.BattleID,
		boiler.TableNames.MechMoveCommandLogs,
		boiler.MechMoveCommandLogColumns.TriggeredByID,
		boiler.MechMoveCommandLogColumns.CreatedAt,
	)

	battleCount := 0
	err := gamedb.StdConn.QueryRow(q, playerID, startFromTime).Scan(&battleCount)
	if err != nil {
		return 0, terror.Error(err, "Failed to get battle count.")
	}

	return battleCount, nil
}

func PlayerRepairForOthersCount(playerID string, startFromTime time.Time) (int, error) {
	q := fmt.Sprintf(`
		SELECT COUNT(ra.id) 
		FROM (
		    SELECT * FROM %[1]s WHERE %[2]s = 'SUCCEEDED' AND %[3]s = $1 AND %[4]s >= $2
		) ra
		INNER JOIN %[5]s ro ON ra.%[6]s = ro.%[7]s AND ro.%[8]s NOTNULL AND ro.%[8]s != ra.%[3]s;
	`,
		boiler.TableNames.RepairAgents,           // 1
		boiler.RepairAgentColumns.FinishedReason, // 2
		boiler.RepairAgentColumns.PlayerID,       // 3
		boiler.RepairAgentColumns.FinishedAt,     // 4

		boiler.TableNames.RepairOffers,          // 5
		boiler.RepairAgentColumns.RepairOfferID, // 6
		boiler.RepairOfferColumns.ID,            // 7
		boiler.RepairOfferColumns.OfferedByID,   // 8
	)

	blockCount := 0
	err := gamedb.StdConn.QueryRow(q, playerID, startFromTime).Scan(&blockCount)
	if err != nil {
		return 0, terror.Error(err, "Failed to get battle count.")
	}

	return blockCount, nil
}

func PlayerMechJoinBattleCount(playerID string, startFromTime time.Time) (int, error) {
	q := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s = $1 AND %s >= $2;",
		boiler.TableNames.BattleMechs,
		boiler.BattleMechColumns.OwnerID,
		boiler.BattleMechColumns.CreatedAt,
	)

	mechCount := 0
	err := gamedb.StdConn.QueryRow(q, playerID, startFromTime).Scan(&mechCount)
	if err != nil {
		return 0, terror.Error(err, "Failed to get battle count.")
	}

	return mechCount, nil
}

func PlayerChatSendCount(playerID string, startFromTime time.Time) (int, error) {
	q := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s = $1 AND %s >= $2;",
		boiler.TableNames.ChatHistory,
		boiler.ChatHistoryColumns.PlayerID,
		boiler.ChatHistoryColumns.CreatedAt,
	)

	chatCount := 0
	err := gamedb.StdConn.QueryRow(q, playerID, startFromTime).Scan(&chatCount)
	if err != nil {
		return 0, terror.Error(err, "Failed to get battle count.")
	}

	return chatCount, nil
}
