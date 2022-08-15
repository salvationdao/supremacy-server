package quest

import (
	"fmt"
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

func New() (*System, error) {
	q := &System{
		playerQuestChan: make(chan *playerQuestCheck, 50),
	}

	err := syncQuests()
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
				boiler.QuestWhere.ExpiresAt.IsNull(),
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
						return
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
	rounds, err := boiler.Rounds(
		boiler.RoundWhere.StartedAt.LTE(now),
		boiler.RoundWhere.EndAt.GT(now),
		boiler.RoundWhere.NextRoundID.IsNull(),
		qm.Load(boiler.RoundRels.Quests),
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
			boiler.BlueprintQuestWhere.RoundType.IN(roundTypes),
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
				if bq.RoundType != roundType {
					continue
				}

				// add new quest, if the quest not exists
				if slices.IndexFunc(roundQuests, func(rq *boiler.Quest) bool { return rq.BlueprintID == bq.ID }) == -1 {
					addedQuests = append(addedQuests, &boiler.Quest{
						RoundID:     r.ID,
						BlueprintID: bq.ID,
						CreatedAt:   r.StartedAt,
					})

					// track change flag
					hasChanged = true
				}
			}

			// get removed quests
			for _, rq := range roundQuests {
				index := slices.IndexFunc(bqs, func(bq *boiler.BlueprintQuest) bool {
					return bq.RoundType == roundType && bq.ID == rq.BlueprintID
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
	rounds, err = boiler.Rounds(
		boiler.RoundWhere.Repeatable.EQ(true),
		boiler.RoundWhere.EndAt.LTE(now),
		boiler.RoundWhere.NextRoundID.IsNull(),
		qm.Load(boiler.RoundRels.Quests),
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
				_, err = r.R.Quests.UpdateAll(tx, boiler.M{boiler.QuestColumns.ExpiresAt: null.TimeFrom(now)})
				if err != nil {
					l.Error().Err(err).Interface("quests", r.R.Quests).Msg("Failed to update quests expiry.")
					return terror.Error(err, "Failed to update quests expiry.")
				}
			}

			// regen new quests
			newRound := &boiler.Round{
				Type:        r.Type,
				Name:        r.Name,
				StartedAt:   now,
				EndAt:       now.AddDate(0, 0, r.LastForDays),
				LastForDays: r.LastForDays,
				Repeatable:  r.Repeatable,
				RoundNumber: r.RoundNumber + 1, // increment round number by one
			}

			err = newRound.Insert(tx, boil.Infer())
			if err != nil {
				l.Error().Err(err).Interface("round", newRound).Msg("Failed to insert new quest")
				return terror.Error(err, "Failed to insert new quest")
			}

			r.NextRoundID = null.StringFrom(newRound.ID)
			_, err = r.Update(tx, boil.Whitelist(boiler.RoundColumns.NextRoundID))
			if err != nil {
				l.Error().Err(err).Interface("involved round", r).Msg("Failed to update next quest id column")
				return terror.Error(err, "Failed to update expired quest")
			}

			// regenerate new quest
			bqs, err := boiler.BlueprintQuests(
				boiler.BlueprintQuestWhere.RoundType.EQ(r.Type),
			).All(tx)
			if err != nil {
				l.Error().Err(err).Str("round type", r.Type).Msg("Failed to get blueprint quests.")
				return terror.Error(err, "Failed to get blueprint quests.")
			}

			for _, bq := range bqs {
				q := boiler.Quest{
					RoundID:     newRound.ID,
					BlueprintID: bq.ID,
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
				ws.PublishMessage(fmt.Sprintf("/user/%s/quest_stat", playerID), server.HubKeyPlayerQuestStats, playerQuestStat)

				// broadcast progressions
				progressions, err := db.PlayerQuestProgressions(playerID)
				if err != nil {
					l.Error().Err(err).Str("player id", playerID).Msg("Failed to load player progressions")
					return
				}
				ws.PublishMessage(fmt.Sprintf("/user/%s/quest_progression", playerID), server.HubKeyPlayerQuestProgressions, progressions)

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

	err := poq.Upsert(gamedb.StdConn, false, nil, boil.Columns{}, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to upsert player")
	}

	playerQuestStat, err := db.PlayerQuestStatGet(playerID)
	if err != nil {
		return terror.Error(err, "Failed to get player quest stat")
	}

	// broadcast player quest stat
	ws.PublishMessage(fmt.Sprintf("/user/%s/quest_stat", playerID), server.HubKeyPlayerQuestStats, playerQuestStat)

	return nil
}

func broadcastProgression(playerID string, questID string, currentProgress int, goal int) {
	if currentProgress > goal {
		currentProgress = goal
	}

	// broadcast changes
	ws.PublishMessage(
		fmt.Sprintf("/user/%s/quest_progression", playerID),
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
