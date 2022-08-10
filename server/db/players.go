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

	q := `
		select
    		q.id,
    		q.name,
    		q.key,
    		q.description,
    		COALESCE(
    		    (SELECT true FROM players_quests pq WHERE pq.quest_id = q.id AND pq.player_id = $1),
    		    false
    		) as obtained
    	from quests q where q.deleted_at isnull;
	`
	rows, err := gamedb.StdConn.Query(q, playerID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("query", q).Msg("Failed to get player quests.")
		return nil, terror.Error(err, "Failed to get player quests.")
	}

	for rows.Next() {
		pq := &server.QuestStat{
			Quest: &boiler.Quest{},
		}
		err = rows.Scan(&pq.ID, &pq.Name, &pq.Key, &pq.Description, &pq.Obtained)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player quests.")
			return nil, terror.Error(err, "Failed to parse player quest.")
		}

		result = append(result, pq)
	}

	return result, nil
}

func PlayerQuestUpsert(playerID string, questID string) error {
	q := `
		INSERT INTO 
		    players_quests (player_id, quest_id)
		VALUES 
			($1, $2)
		ON CONFLICT 
		    (player_id, quest_id)
		DO NOTHING 
	`

	_, err := gamedb.StdConn.Exec(q, playerID, questID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("player id", playerID).Str("quest id", questID).Msg("Failed to upsert player quest")
		return terror.Error(err, "Failed to upsert player quest.")
	}

	return nil
}

func PlayerMechKillCount(playerID string, afterTime time.Time) (int, error) {
	q := `
		SELECT count(bm.owner_id) FROM battle_history bh
		INNER JOIN battle_mechs bm ON bh.battle_id = bm.battle_id AND bm.mech_id = bh.war_machine_two_id AND bm.owner_id = $1
		where event_type = 'killed' AND bh.war_machine_two_id notnull AND bh.created_at >= $2;
	`

	mechKillCount := 0
	err := gamedb.StdConn.QueryRow(q, playerID, afterTime).Scan(&mechKillCount)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to scan player mech kill count")
		return 0, terror.Error(err, "Failed to get player mech kill count")
	}

	return mechKillCount, nil
}
