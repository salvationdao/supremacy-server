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
	playerQuestChan chan func()
}

func New() *System {
	q := &System{
		playerQuestChan: make(chan func(), 30),
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
		case fn := <-q.playerQuestChan:
			fn()
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

	q.playerQuestChan <- func() {
		// get all the ability kill quest
		pqs, err := boiler.Quests(
			boiler.QuestWhere.Key.EQ(boiler.QuestKeyAbilityKill),
			boiler.QuestWhere.EndedAt.IsNull(), // impact by quest regen
		).All(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("Failed to get quest")
			return
		}

		for _, pq := range pqs {
			playerQuest, err := pq.PlayersQuests(
				boiler.PlayersQuestWhere.PlayerID.EQ(playerID),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				l.Error().Err(err).Msg("Failed to check player quest")
				return
			}

			// skip, if player has already done the quest
			if playerQuest != nil {
				continue
			}

			// check player ability kill match the amount
			playerKillLogs, err := boiler.PlayerKillLogs(
				boiler.PlayerKillLogWhere.PlayerID.EQ(playerID),
				boiler.PlayerKillLogWhere.CreatedAt.GT(pq.CreatedAt), // involve the logs after the quest issue time
			).All(gamedb.StdConn)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player kill logs")
				return
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
				return
			}

			err = playerQuestGrant(playerID, pq.ID)
			if err != nil {
				l.Error().Err(err).Str("quest id", pq.ID).Msg("Failed to grant player quest.")
				return
			}
		}
	}
}

func (q *System) MechKillQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyMechKill).Str("player id", playerID).Logger()

	q.playerQuestChan <- func() {
		// get all the mech kill quest
		pqs, err := boiler.Quests(
			boiler.QuestWhere.Key.EQ(boiler.QuestKeyMechKill),
			boiler.QuestWhere.EndedAt.IsNull(), // impact by quest regen
		).All(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("Failed to get quest")
			return
		}

		for _, pq := range pqs {
			playerQuest, err := pq.PlayersQuests(
				boiler.PlayersQuestWhere.PlayerID.EQ(playerID),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				l.Error().Err(err).Msg("Failed to check player quest")
				return
			}

			// skip, if player has already done the quest
			if playerQuest != nil {
				continue
			}

			// check player eligible to claim
			mechKillCount, err := db.PlayerMechKillCount(playerID, pq.CreatedAt)
			if err != nil {
				l.Error().Err(err).Msg("Failed to get player mech kill count")
				return
			}

			if mechKillCount < pq.RequestAmount {
				continue
			}

			err = playerQuestGrant(playerID, pq.ID)
			if err != nil {
				l.Error().Err(err).Str("quest id", pq.ID).Msg("Failed to grant player quest.")
				return
			}
		}
	}
}

func (q *System) MechCommanderQuestCheck(playerID string) {
	l := gamelog.L.With().Str("quest key", boiler.QuestKeyTotalBattleUsedMechCommander).Str("player id", playerID).Logger()

	q.playerQuestChan <- func() {
		// get all the mech kill quest
		pqs, err := boiler.Quests(
			boiler.QuestWhere.Key.EQ(boiler.QuestKeyTotalBattleUsedMechCommander),
			boiler.QuestWhere.EndedAt.IsNull(), // impact by quest regen
		).All(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("Failed to get quest")
			return
		}

		for _, pq := range pqs {
			playerQuest, err := pq.PlayersQuests(
				boiler.PlayersQuestWhere.PlayerID.EQ(playerID),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				l.Error().Err(err).Msg("Failed to check player quest")
				return
			}

			// skip, if player has already done the quest
			if playerQuest != nil {
				continue
			}

			// check player eligible to claim

			err = playerQuestGrant(playerID, pq.ID)
			if err != nil {
				l.Error().Err(err).Str("quest id", pq.ID).Msg("Failed to grant player quest.")
				return
			}
		}
	}
}
