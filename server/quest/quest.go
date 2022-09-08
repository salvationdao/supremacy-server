package quest

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type System struct {
	playerQuestChan chan *playerQuestCheck
}

type playerQuestCheck struct {
	questKey  string
	playerID  string
	checkFunc func(playerID string, quest *boiler.Quest, blueprintQuest *boiler.BlueprintQuest) bool
}

const QuestEventNameProvingGround = "Proving Grounds"
const QuestEventNameDaily = "Daily Challenge"

func New() (*System, error) {
	q := &System{
		playerQuestChan: make(chan *playerQuestCheck, 50),
	}

	if !server.IsProductionEnv() {
		// insert proving ground quest event
		r, err := boiler.QuestEvents(
			boiler.QuestEventWhere.Type.EQ(boiler.QuestEventTypeProvingGrounds),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, terror.Error(err, "Failed to load staging quest")
		}

		if r == nil {
			now := time.Now()
			r = &boiler.QuestEvent{
				Type:               boiler.QuestEventTypeProvingGrounds,
				Name:               QuestEventNameProvingGround,
				StartedAt:          now,
				EndAt:              now.AddDate(0, 0, 10), // default value
				DurationType:       boiler.QuestEventDurationTypeCustom,
				CustomDurationDays: null.IntFrom(10),
				Repeatable:         true,
				QuestEventNumber:   1,
			}
			err = r.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return nil, terror.Error(err, "Failed to insert staging quests.")
			}
		}
	}

	// check test quests exists
	r, err := boiler.QuestEvents(
		boiler.QuestEventWhere.Name.EQ(QuestEventNameDaily),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to load staging quest")
	}

	if r == nil {
		now := time.Now()
		r = &boiler.QuestEvent{
			Type:             boiler.QuestEventTypeDailyQuest,
			Name:             QuestEventNameDaily,
			StartedAt:        now,
			EndAt:            now.AddDate(0, 0, 1), // default value
			DurationType:     boiler.QuestEventDurationTypeDaily,
			Repeatable:       true,
			QuestEventNumber: 1,
		}
		err = r.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return nil, terror.Error(err, "Failed to insert staging quests.")
		}
	}

	err = syncQuests()
	if err != nil {
		return nil, err
	}

	go q.Run()

	return q, nil
}

func (q *System) Run() {
	regenerateTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-regenerateTicker.C:
			err := syncQuests()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to regen new quest")
			}
		case pqc := <-q.playerQuestChan:
			l := gamelog.L.With().Str("quest key", pqc.questKey).Str("player id", pqc.playerID).Logger()
			// get all the ability related quests
			pqs, err := boiler.Quests(
				boiler.QuestWhere.ExpiredAt.IsNull(),
				qm.InnerJoin(
					fmt.Sprintf(
						"%s ON %s = %s AND %s = ?",
						boiler.TableNames.BlueprintQuests,
						qm.Rels(boiler.TableNames.BlueprintQuests, boiler.BlueprintQuestColumns.ID),
						qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.BlueprintID),
						qm.Rels(boiler.TableNames.BlueprintQuests, boiler.BlueprintQuestColumns.Key),
					),
					pqc.questKey,
				),
				qm.Load(
					boiler.QuestRels.ObtainedQuestPlayersObtainedQuests,
					boiler.PlayersObtainedQuestWhere.PlayerID.EQ(pqc.playerID),
				),
				qm.Load(
					boiler.QuestRels.Blueprint,
				),
			).All(gamedb.StdConn)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get quest")
				return
			}

			for _, pq := range pqs {

				// skip, if player has already done the quest
				if pq.R != nil && pq.R.ObtainedQuestPlayersObtainedQuests != nil && len(pq.R.ObtainedQuestPlayersObtainedQuests) > 0 {
					continue
				}

				if pqc.checkFunc(pqc.playerID, pq, pq.R.Blueprint) {
					err = playerQuestGrant(pqc.playerID, pq.ID)
					if err != nil {
						l.Error().Err(err).Str("quest id", pq.ID).Msg("Failed to grant player quest.")
					}
				}
			}
		}
	}
}

// syncQuests check expired quest and regenerate new quest for it
func syncQuests() error {
	now := time.Now()
	l := gamelog.L.With().Str("func name", "regenExpiredQuest()").Logger()

	hasChanged := false

	// handle quests sync for all the available quest

	// get current available quest
	rounds, err := boiler.QuestEvents(
		boiler.QuestEventWhere.StartedAt.LTE(now),
		boiler.QuestEventWhere.EndAt.GT(now),
		boiler.QuestEventWhere.NextQuestEventID.IsNull(),
		qm.Load(boiler.QuestEventRels.Quests),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to query available quests")
		return terror.Error(err, "Failed to load current available quest")
	}

	if rounds != nil && len(rounds) > 0 {
		roundTypes := []string{}
		for _, r := range rounds {
			roundTypes = append(roundTypes, r.Type)
		}

		bqs, err := boiler.BlueprintQuests(
			boiler.BlueprintQuestWhere.QuestEventType.IN(roundTypes),
		).All(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to load blueprint quest")
		}

		addedQuests := []*boiler.Quest{}
		removedQuests := []*boiler.Quest{}
		for _, r := range rounds {
			roundType := r.Type
			roundQuests := []*boiler.Quest{}
			if r.R != nil && r.R.Quests != nil {
				roundQuests = r.R.Quests
			}

			// get added quests
			for _, bq := range bqs {
				// skip, if different round type
				if bq.QuestEventType != roundType {
					continue
				}

				// add new quest, if the quest not exists
				if slices.IndexFunc(roundQuests, func(rq *boiler.Quest) bool { return rq.BlueprintID == bq.ID }) == -1 {
					addedQuests = append(addedQuests, &boiler.Quest{
						QuestEventID: r.ID,
						BlueprintID:  bq.ID,
						CreatedAt:    r.StartedAt,
					})

					// track change flag
					hasChanged = true
				}
			}

			// get removed quests
			for _, rq := range roundQuests {
				index := slices.IndexFunc(bqs, func(bq *boiler.BlueprintQuest) bool {
					return bq.QuestEventType == roundType && bq.ID == rq.BlueprintID
				})

				// remove, if quest no longer exists
				if index == -1 {
					removedQuests = append(removedQuests, rq)

					// track change flag
					hasChanged = true
				}
			}
		}

		if hasChanged {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				l.Error().Err(err).Msg("Failed to start db transaction")
				return terror.Error(err, "Failed to start db transaction")
			}

			defer tx.Rollback()

			for _, aq := range addedQuests {
				err = aq.Insert(tx, boil.Infer())
				if err != nil {
					l.Error().Err(err).Interface("quest", aq).Msg("Failed to insert new quest.")
					return terror.Error(err, "Failed to insert new quest.")
				}
			}

			for _, rq := range removedQuests {
				rq.DeletedAt = null.TimeFrom(now)
				_, err := rq.Update(tx, boil.Whitelist(boiler.QuestColumns.DeletedAt))
				if err != nil {
					l.Error().Err(err).Interface("quest", rq).Msg("Failed to remove quest.")
					return terror.Error(err, "Failed to remove quest.")
				}
			}

			err = tx.Commit()
			if err != nil {
				l.Error().Err(err).Msg("Failed to commit db transaction")
				return terror.Error(err, "Failed to commit db transaction.")
			}
		}
	}

	// handle expired rounds
	rounds, err = boiler.QuestEvents(
		boiler.QuestEventWhere.Repeatable.EQ(true),
		boiler.QuestEventWhere.EndAt.LTE(now),
		boiler.QuestEventWhere.NextQuestEventID.IsNull(),
		qm.Load(boiler.QuestEventRels.Quests),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to query expired quests")
		return terror.Error(err, "Failed to load expired quests")
	}

	for _, r := range rounds {
		// track change flag
		hasChanged = true

		err = func() error {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				l.Error().Err(err).Msg("Failed to start db transaction")
				return terror.Error(err, "Failed to start db transaction")
			}

			defer tx.Rollback()

			// update old quests' expiry
			if r.R != nil && r.R.Quests != nil {
				_, err = r.R.Quests.UpdateAll(tx, boiler.M{boiler.QuestColumns.ExpiredAt: null.TimeFrom(now)})
				if err != nil {
					l.Error().Err(err).Interface("quests", r.R.Quests).Msg("Failed to update quests expiry.")
					return terror.Error(err, "Failed to update quests expiry.")
				}
			}

			// regen new quests
			newRound := &boiler.QuestEvent{
				Type:               r.Type,
				Name:               r.Name,
				StartedAt:          now,
				EndAt:              now.AddDate(0, 0, 1), // default value
				DurationType:       r.DurationType,
				CustomDurationDays: r.CustomDurationDays,
				Repeatable:         r.Repeatable,
				QuestEventNumber:   r.QuestEventNumber + 1, // increment round number by one
			}

			switch newRound.DurationType {
			case boiler.QuestEventDurationTypeDaily:
				newRound.EndAt = now.AddDate(0, 0, 1)

			case boiler.QuestEventDurationTypeWeekly:
				newRound.EndAt = now.AddDate(0, 0, 7)

			case boiler.QuestEventDurationTypeMonthly:
				newRound.EndAt = now.AddDate(0, 1, 0)

			case boiler.QuestEventDurationTypeCustom:
				if newRound.CustomDurationDays.Valid {
					newRound.EndAt = now.AddDate(0, 0, newRound.CustomDurationDays.Int)
				} else {
					l.Warn().Interface("round", newRound).Msg("Missing custom duration days field.")
				}
			}

			err = newRound.Insert(tx, boil.Infer())
			if err != nil {
				l.Error().Err(err).Interface("round", newRound).Msg("Failed to insert new quest")
				return terror.Error(err, "Failed to insert new quest")
			}

			r.NextQuestEventID = null.StringFrom(newRound.ID)
			_, err = r.Update(tx, boil.Whitelist(boiler.QuestEventColumns.NextQuestEventID))
			if err != nil {
				l.Error().Err(err).Interface("involved round", r).Msg("Failed to update next quest id column")
				return terror.Error(err, "Failed to update expired quest")
			}

			// regenerate new quest
			bqs, err := boiler.BlueprintQuests(
				boiler.BlueprintQuestWhere.QuestEventType.EQ(r.Type),
			).All(tx)
			if err != nil {
				l.Error().Err(err).Str("round type", r.Type).Msg("Failed to get blueprint quests.")
				return terror.Error(err, "Failed to get blueprint quests.")
			}

			for _, bq := range bqs {
				q := boiler.Quest{
					QuestEventID: newRound.ID,
					BlueprintID:  bq.ID,
				}
				err = q.Insert(tx, boil.Infer())
				if err != nil {
					l.Error().Err(err).Interface("quest", q).Msg("Failed to insert new quest.")
					return terror.Error(err, "Failed to insert new quest.")
				}
			}

			err = tx.Commit()
			if err != nil {
				l.Error().Err(err).Msg("Failed to commit db transaction")
				return terror.Error(err, "Failed to commit db transaction.")
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	// broadcast changes to all the connected players
	if hasChanged {
		wg := sync.WaitGroup{}
		for _, playerID := range ws.TrackedIdents() {
			wg.Add(1)
			go func(playerID string) {
				defer wg.Done()

				// broadcast player quest stat
				playerQuestStat, err := db.PlayerQuestStatGet(playerID)
				if err != nil {
					l.Error().Err(err).Str("player id", playerID).Msg("Failed to load player quest status")
					return
				}
				ws.PublishMessage(fmt.Sprintf("/secure/user/%s/quest_stat", playerID), server.HubKeyPlayerQuestStats, playerQuestStat)

				// broadcast progressions
				progressions, err := db.PlayerQuestProgressions(playerID)
				if err != nil {
					l.Error().Err(err).Str("player id", playerID).Msg("Failed to load player progressions")
					return
				}
				ws.PublishMessage(fmt.Sprintf("/secure/user/%s/quest_progression", playerID), server.HubKeyPlayerQuestProgressions, progressions)

			}(playerID)
		}
		wg.Wait()
	}

	return nil
}

func playerQuestGrant(playerID string, questID string) error {
	poq := boiler.PlayersObtainedQuest{
		PlayerID:        playerID,
		ObtainedQuestID: questID,
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed complete quest")
	}
	defer tx.Rollback()

	err = poq.Upsert(tx, false, nil, boil.Columns{}, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to upsert player")
	}

	// reward mini mech for quest completion
	miniMechBlueprint, err := boiler.BlueprintPlayerAbilities(
		boiler.BlueprintPlayerAbilityWhere.GameClientAbilityID.EQ(18), // TODO: when player abilities are in static data, use ID instead
	).One(tx)
	if err != nil {
		return terror.Error(err, "Failed to get mini mech reward player")
	}
	// Update player ability count
	pa, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.BlueprintID.EQ(miniMechBlueprint.ID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(playerID),
	).One(tx)
	if errors.Is(err, sql.ErrNoRows) {
		pa = &boiler.PlayerAbility{
			OwnerID:         playerID,
			BlueprintID:     miniMechBlueprint.ID,
			LastPurchasedAt: time.Now(),
		}

		err = pa.Insert(tx, boil.Infer())
		if err != nil {
			return terror.Error(err, "Failed to get mini mech reward player")
		}
	} else if err != nil {
		return terror.Error(err, "Failed to get mini mech reward player")
	}

	pa.Count = pa.Count + 1

	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to get mini mech reward player")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed complete quest")
	}

	// Tell client to update their player abilities list
	pas, err := db.PlayerAbilitiesList(playerID)
	if err != nil {
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/player_abilities", playerID), server.HubKeyPlayerAbilitiesList, pas)

	playerQuestStat, err := db.PlayerQuestStatGet(playerID)
	if err != nil {
		return terror.Error(err, "Failed to get player quest stat")
	}

	// broadcast player quest stat
	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/quest_stat", playerID), server.HubKeyPlayerQuestStats, playerQuestStat)

	return nil
}

func broadcastProgression(playerID string, questID string, currentProgress int, goal int) {
	if currentProgress > goal {
		currentProgress = goal
	}

	// broadcast changes
	ws.PublishMessage(
		fmt.Sprintf("/secure/user/%s/quest_progression", playerID),
		server.HubKeyPlayerQuestProgressions,
		[]*db.PlayerQuestProgression{{questID, currentProgress, goal}},
	)
}

// AbilityKillQuestCheck gain players ability kill quest if they are eligible.
func (q *System) AbilityKillQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyAbilityKill).Str("player id", playerID).Logger()

	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyAbilityKill,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest, bq *boiler.BlueprintQuest) bool {
			// check player ability kill match the amount
			playerKillLogs, err := boiler.PlayerKillLogs(
				boiler.PlayerKillLogWhere.PlayerID.EQ(playerID),
				boiler.PlayerKillLogWhere.CreatedAt.GT(pq.CreatedAt), // involve the logs after the quest issue time
			).All(gamedb.StdConn)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player kill logs")
				return false
			}

			totalKill := 0
			for _, pkl := range playerKillLogs {
				if pkl.IsTeamKill {
					totalKill -= 1
					continue
				}
				totalKill += 1
			}

			if totalKill < 0 {
				totalKill = 0
			}

			broadcastProgression(playerID, pq.ID, totalKill, bq.RequestAmount)

			// return if not eligible
			if totalKill < bq.RequestAmount {
				return false
			}

			return true
		},
	}
}

func (q *System) MechKillQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyMechKill).Str("player id", playerID).Logger()

	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyMechKill,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest, bq *boiler.BlueprintQuest) bool {
			// check player eligible to claim
			mechKillCount, err := db.PlayerMechKillCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player mech kill count")
				return false
			}

			broadcastProgression(playerID, pq.ID, mechKillCount, bq.RequestAmount)

			if mechKillCount < bq.RequestAmount {
				return false
			}

			return true
		},
	}
}

func (q *System) MechCommanderQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyTotalBattleUsedMechCommander).Str("player id", playerID).Logger()

	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyTotalBattleUsedMechCommander,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest, bq *boiler.BlueprintQuest) bool {
			// check player eligible to claim
			battleCount, err := db.PlayerTotalBattleMechCommanderUsed(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Str("quest id", pq.ID).Msg("Failed to count total battles.")
				return false
			}

			broadcastProgression(playerID, pq.ID, battleCount, bq.RequestAmount)

			if battleCount < bq.RequestAmount {
				return false
			}

			return true
		},
	}
}

func (q *System) RepairQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyRepairForOther).Str("player id", playerID).Logger()
	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyRepairForOther,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest, bq *boiler.BlueprintQuest) bool {
			// check player eligible to claim
			blockCount, err := db.PlayerRepairForOthersCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			broadcastProgression(playerID, pq.ID, blockCount, bq.RequestAmount)

			if blockCount < bq.RequestAmount {
				return false
			}

			return true
		},
	}
}

func (q *System) ChatMessageQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyChatSent).Str("player id", playerID).Logger()
	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyChatSent,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest, bq *boiler.BlueprintQuest) bool {
			// check player eligible to claim
			chatCount, err := db.PlayerChatSendCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			broadcastProgression(playerID, pq.ID, chatCount, bq.RequestAmount)

			if chatCount < bq.RequestAmount {
				return false
			}

			return true
		},
	}
}

func (q *System) MechJoinBattleQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyMechJoinBattle).Str("player id", playerID).Logger()
	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyMechJoinBattle,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest, bq *boiler.BlueprintQuest) bool {
			// check player eligible to claim
			mechCount, err := db.PlayerMechJoinBattleCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			broadcastProgression(playerID, pq.ID, mechCount, bq.RequestAmount)

			if mechCount < bq.RequestAmount {
				return false
			}

			return true
		},
	}
}
