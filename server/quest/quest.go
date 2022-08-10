package quest

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

func New() *System {
	q := &System{
		playerQuestChan: make(chan *playerQuestCheck, 50),
	}

	go q.Run()

	return q
}

func (q *System) Run() {
	regenerateTicker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-regenerateTicker.C:
			regenerateQuest()
		case pqc := <-q.playerQuestChan:
			l := gamelog.L.With().Str("quest key", pqc.questKey).Str("player id", pqc.playerID).Logger()
			// get all the ability related quests
			pqs, err := boiler.Quests(
				boiler.QuestWhere.Key.EQ(pqc.questKey),
				boiler.QuestWhere.EndedAt.IsNull(), // impact by quest regen
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

// regenerateQuest check expired quest and regenerate new quest for it
func regenerateQuest() {
	pqs, err := boiler.Quests(
		boiler.QuestWhere.NextQuestID.IsNull(),
		boiler.QuestWhere.EndedAt.IsNotNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load ended quest.")
		return
	}

	for _, pq := range pqs {
		newQuest := &boiler.Quest{
			Name:          pq.Name,
			Key:           pq.Key,
			Description:   pq.Description,
			RequestAmount: pq.RequestAmount,
		}

		func() {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
				return
			}

			defer tx.Rollback()

			err = newQuest.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("quest", newQuest).Msg("Failed to insert new quest")
				return
			}

			pq.NextQuestID = null.StringFrom(newQuest.ID)
			_, err = pq.Update(tx, boil.Whitelist(boiler.QuestColumns.NextQuestID))
			if err != nil {
				gamelog.L.Error().Err(err).Interface("involved quest", pq).Msg("Failed to update next quest id column")
				return
			}

			err = tx.Commit()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
				return
			}
		}()
	}

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
			mechCount, err := db.PlayerChatSendCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get total repair block")
				return false
			}

			if mechCount < pq.RequestAmount {
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

			if mechCount < pq.RequestAmount {
				return false
			}

			return true
		},
	}
}
