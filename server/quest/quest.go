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
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

type System struct {
	playerQuestChan chan *playerQuestCheck
}

type playerQuestCheck struct {
	questKey  string
	playerID  string
	checkFunc func(playerID string, quest *boiler.Quest) bool
}

type PlayerQuestProgression struct {
	QuestID string `json:"quest_id"`
	Current int    `json:"current"`
	Goal    int    `json:"goal"`
}

func New() (*System, error) {
	q := &System{
		playerQuestChan: make(chan *playerQuestCheck, 50),
	}

	err := regenExpiredQuest()
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
			err := regenExpiredQuest()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to regen new quest")
			}
		case pqc := <-q.playerQuestChan:
			l := gamelog.L.With().Str("quest key", pqc.questKey).Str("player id", pqc.playerID).Logger()
			// get all the ability related quests
			pqs, err := boiler.Quests(
				boiler.QuestWhere.Key.EQ(pqc.questKey),
				boiler.QuestWhere.ExpiresAt.GT(time.Now()), // impact by quest regen
			).All(gamedb.StdConn)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get quest")
				return
			}

			for _, pq := range pqs {
				playerQuest, err := pq.PlayersQuests(
					boiler.PlayersQuestWhere.PlayerID.EQ(pqc.playerID),
				).One(gamedb.StdConn)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					l.Error().Err(err).Msg("Failed to check player quest")
					return
				}

				// skip, if player has already done the quest
				if playerQuest != nil {
					continue
				}

				if pqc.checkFunc(pqc.playerID, pq) {
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

// regenExpiredQuest check expired quest and regenerate new quest for it
func regenExpiredQuest() error {
	now := time.Now()
	l := gamelog.L.With().Str("func name", "regenExpiredQuest()").Logger()

	// get expired rounds
	rounds, err := boiler.Rounds(
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
		newRound := &boiler.Round{
			Name:        r.Name,
			StartedAt:   now,
			EndAt:       now.AddDate(0, 0, r.LastForDays),
			LastForDays: r.LastForDays,
			Repeatable:  r.Repeatable,
			RoundNumber: r.RoundNumber + 1, // increment round number by one
		}

		err = func() error {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				l.Error().Err(err).Msg("Failed to start db transaction")
				return terror.Error(err, "Failed to start db transaction")
			}

			defer tx.Rollback()

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
			if r.R != nil {
				for _, q := range r.R.Quests {
					// regenerate new quest
					newQuest := &boiler.Quest{
						RoundID:       newRound.ID,
						Name:          q.Name,
						Key:           q.Key,
						Description:   q.Description,
						RequestAmount: q.RequestAmount,
						ExpiresAt:     newRound.EndAt,
						CreatedAt:     newRound.StartedAt,
					}

					err = newQuest.Insert(tx, boil.Infer())
					if err != nil {
						return terror.Error(err, "Failed to generate new quest")
					}
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

	return nil
}

func playerQuestGrant(playerID string, questID string) error {
	err := db.PlayerQuestUpsert(playerID, questID)
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

// AbilityKillQuestCheck gain players ability kill quest if they are eligible.
func (q *System) AbilityKillQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyAbilityKill).Str("player id", playerID).Logger()

	q.playerQuestChan <- &playerQuestCheck{
		questKey: boiler.QuestKeyAbilityKill,
		playerID: playerID,
		checkFunc: func(playerID string, pq *boiler.Quest) bool {
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

			// broadcast changes
			ws.PublishMessage(
				fmt.Sprintf("/user/%s/quest_progression", playerID),
				server.HubKeyPlayerQuestProgressions,
				[]*PlayerQuestProgression{{pq.ID, totalKill, pq.RequestAmount}},
			)

			// return if not eligible
			if totalKill < pq.RequestAmount {
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
		checkFunc: func(playerID string, pq *boiler.Quest) bool {
			// check player eligible to claim
			mechKillCount, err := db.PlayerMechKillCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player mech kill count")
				return false
			}

			// broadcast changes
			ws.PublishMessage(
				fmt.Sprintf("/user/%s/quest_progression", playerID),
				server.HubKeyPlayerQuestProgressions,
				[]*PlayerQuestProgression{{pq.ID, mechKillCount, pq.RequestAmount}},
			)

			if mechKillCount < pq.RequestAmount {
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
		checkFunc: func(playerID string, pq *boiler.Quest) bool {
			// check player eligible to claim
			battleCount, err := db.PlayerTotalBattleMechCommanderUsed(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Str("quest id", pq.ID).Msg("Failed to count total battles.")
				return false
			}

			// broadcast changes
			ws.PublishMessage(
				fmt.Sprintf("/user/%s/quest_progression", playerID),
				server.HubKeyPlayerQuestProgressions,
				[]*PlayerQuestProgression{{pq.ID, battleCount, pq.RequestAmount}},
			)

			if battleCount < pq.RequestAmount {
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
		checkFunc: func(playerID string, pq *boiler.Quest) bool {
			// check player eligible to claim
			blockCount, err := db.PlayerRepairForOthersCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			// broadcast changes
			ws.PublishMessage(
				fmt.Sprintf("/user/%s/quest_progression", playerID),
				server.HubKeyPlayerQuestProgressions,
				[]*PlayerQuestProgression{{pq.ID, blockCount, pq.RequestAmount}},
			)

			if blockCount < pq.RequestAmount {
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
		checkFunc: func(playerID string, pq *boiler.Quest) bool {
			// check player eligible to claim
			chatCount, err := db.PlayerChatSendCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			// broadcast changes
			ws.PublishMessage(
				fmt.Sprintf("/user/%s/quest_progression", playerID),
				server.HubKeyPlayerQuestProgressions,
				[]*PlayerQuestProgression{{pq.ID, chatCount, pq.RequestAmount}},
			)

			if chatCount < pq.RequestAmount {
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
		checkFunc: func(playerID string, pq *boiler.Quest) bool {
			// check player eligible to claim
			mechCount, err := db.PlayerMechJoinBattleCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			// broadcast changes
			ws.PublishMessage(
				fmt.Sprintf("/user/%s/quest_progression", playerID),
				server.HubKeyPlayerQuestProgressions,
				[]*PlayerQuestProgression{{pq.ID, mechCount, pq.RequestAmount}},
			)

			if mechCount < pq.RequestAmount {
				return false
			}

			return true
		},
	}
}
