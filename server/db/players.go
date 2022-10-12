package db

import (
	"context"
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sort"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// PlayerRegister new user who may or may not be enlisted
func PlayerRegister(ID uuid.UUID, Username string, FactionID uuid.UUID, PublicAddress common.Address, AcceptsMarketing null.Bool) (*boiler.Player, error) {
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
		player.AcceptsMarketing = AcceptsMarketing

		_, err = player.Update(tx, boil.Infer())
		if err != nil {
			return nil, err
		}
	} else {
		player = &boiler.Player{
			ID:               ID.String(),
			PublicAddress:    null.NewString(hexPublicAddress, hexPublicAddress != ""),
			Username:         null.NewString(Username, true),
			FactionID:        null.NewString(FactionID.String(), !FactionID.IsNil()),
			AcceptsMarketing: AcceptsMarketing,
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

	player, err := boiler.Players(
		boiler.PlayerWhere.ID.EQ(playerID),
		qm.Load(boiler.PlayerRels.Role),
	).One(gamedb.StdConn)
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
			fmt.Sprintf("%s AS round_name", qm.Rels(boiler.TableNames.QuestEvents, boiler.QuestEventColumns.Name)),
			fmt.Sprintf("%s AS started_at", qm.Rels(boiler.TableNames.QuestEvents, boiler.QuestEventColumns.StartedAt)),
			fmt.Sprintf("%s AS end_at", qm.Rels(boiler.TableNames.QuestEvents, boiler.QuestEventColumns.EndAt)),
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
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.ExpiredAt),
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
				boiler.TableNames.QuestEvents,
				qm.Rels(boiler.TableNames.QuestEvents, boiler.QuestEventColumns.ID),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.QuestEventID),
			),
		),
	).Bind(context.Background(), gamedb.StdConn, &result)
	if err != nil {
		gamelog.L.Error().Err(err).Str("player id", playerID).Msg("Failed to get player quests.")
		return nil, terror.Error(err, "Failed to get player quests.")
	}

	// sort list
	sort.Slice(result, func(i, j int) bool {
		return result[i].EndAt.Sub(result[i].StartedAt) < result[j].EndAt.Sub(result[j].StartedAt)
	})

	checkedRoundName := make(map[string]bool)
	resp := []*server.QuestStat{}
	for _, r := range result {
		roundName := r.RoundName

		// check round name is already done
		if _, ok := checkedRoundName[roundName]; ok {
			continue
		}

		// append quest to response
		for _, q := range result {
			if q.RoundName == roundName {
				resp = append(resp, q)
			}
		}

		// record round name
		checkedRoundName[roundName] = true
	}

	return resp, nil
}

type PlayerQuestProgression struct {
	QuestID string `json:"quest_id"`
	Current int    `json:"current"`
	Goal    int    `json:"goal"`
}

func PlayerQuestProgressions(playerID string) ([]*PlayerQuestProgression, error) {
	l := gamelog.L.With().Str("player id", playerID).Str("func name", "PlayerQuestProgressions").Logger()
	// get all the available quests
	quests, err := boiler.Quests(
		boiler.QuestWhere.ExpiredAt.IsNull(),
		qm.Load(
			boiler.QuestRels.ObtainedQuestPlayersObtainedQuests,
			boiler.PlayersObtainedQuestWhere.PlayerID.EQ(playerID),
		),
		qm.Load(boiler.QuestRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to query available quest")
		return nil, terror.Error(err, "Failed to get available quests")
	}

	result := []*PlayerQuestProgression{}

	// cache data to speed up the process
	abilityKillCount := null.IntFromPtr(nil)
	mechKillCount := null.IntFromPtr(nil)
	mechJoinBattleCount := null.IntFromPtr(nil)
	chatSentCount := null.IntFromPtr(nil)
	repairOtherCount := null.IntFromPtr(nil)
	mechCommandBattleCount := null.IntFromPtr(nil)

	// loop through quest
	for _, q := range quests {
		pqp := &PlayerQuestProgression{
			QuestID: q.ID,
			Current: 0,
			Goal:    q.R.Blueprint.RequestAmount,
		}

		// if player already obtained the quest
		if q.R != nil && q.R.ObtainedQuestPlayersObtainedQuests != nil && len(q.R.ObtainedQuestPlayersObtainedQuests) > 0 {
			// set current score to goal, and append to result
			pqp.Current = pqp.Goal
			result = append(result, pqp)
			continue
		}

		// log quest id
		l = l.With().Str("quest id", q.ID).Logger()

		// otherwise, load current progression
		switch q.R.Blueprint.Key {
		case boiler.QuestKeyAbilityKill:
			// fill data, if already query once
			if abilityKillCount.Valid {
				pqp.Current = abilityKillCount.Int
				result = append(result, pqp)
				continue
			}

			playerKillLogs, err := boiler.PlayerKillLogs(
				boiler.PlayerKillLogWhere.PlayerID.EQ(playerID),
				boiler.PlayerKillLogWhere.CreatedAt.GT(q.CreatedAt), // involve the logs after the quest issue time
			).All(gamedb.StdConn)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player kill logs")
				return nil, err
			}

			for _, pkl := range playerKillLogs {
				if pkl.IsTeamKill {
					pqp.Current -= 1
					continue
				}
				pqp.Current += 1
			}

			// cap at zero
			if pqp.Current < 0 {
				pqp.Current = 0
			}

			// cap current score with the quest goal
			if pqp.Current > pqp.Goal {
				pqp.Current = pqp.Goal
			}

			abilityKillCount = null.IntFrom(pqp.Current)

		case boiler.QuestKeyMechKill:
			if mechKillCount.Valid {
				pqp.Current = mechKillCount.Int
				result = append(result, pqp)
				continue
			}

			pqp.Current, err = PlayerMechKillCount(playerID, q.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player mech kill count")
				return nil, err
			}

			// cap current score with the quest goal
			if pqp.Current > pqp.Goal {
				pqp.Current = pqp.Goal
			}

			mechKillCount = null.IntFrom(pqp.Current)

		case boiler.QuestKeyMechJoinBattle:
			if mechJoinBattleCount.Valid {
				pqp.Current = mechJoinBattleCount.Int
				result = append(result, pqp)
				continue
			}

			pqp.Current, err = PlayerMechJoinBattleCount(playerID, q.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return nil, err
			}

			// cap current score with the quest goal
			if pqp.Current > pqp.Goal {
				pqp.Current = pqp.Goal
			}

			mechJoinBattleCount = null.IntFrom(pqp.Current)

		case boiler.QuestKeyChatSent:
			if chatSentCount.Valid {
				pqp.Current = chatSentCount.Int
				result = append(result, pqp)
				continue
			}

			pqp.Current, err = PlayerChatSendCount(playerID, q.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return nil, err
			}

			// cap current score with the quest goal
			if pqp.Current > pqp.Goal {
				pqp.Current = pqp.Goal
			}

			chatSentCount = null.IntFrom(pqp.Current)

		case boiler.QuestKeyRepairForOther:
			if repairOtherCount.Valid {
				pqp.Current = repairOtherCount.Int
				result = append(result, pqp)
				continue
			}

			pqp.Current, err = PlayerRepairForOthersCount(playerID, q.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return nil, err
			}

			// cap current score with the quest goal
			if pqp.Current > pqp.Goal {
				pqp.Current = pqp.Goal
			}

			repairOtherCount = null.IntFrom(pqp.Current)

		case boiler.QuestKeyTotalBattleUsedMechCommander:
			if mechCommandBattleCount.Valid {
				pqp.Current = mechCommandBattleCount.Int
				result = append(result, pqp)
				continue
			}

			pqp.Current, err = PlayerTotalBattleMechCommanderUsed(playerID, q.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to count total battles.")
				return nil, err
			}

			// cap current score with the quest goal
			if pqp.Current > pqp.Goal {
				pqp.Current = pqp.Goal
			}

			mechCommandBattleCount = null.IntFrom(pqp.Current)
		}

		result = append(result, pqp)
	}

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
