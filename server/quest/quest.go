package quest

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"time"
)

type Quest struct {
	abilityKillChan chan string
}

func New() *Quest {
	q := &Quest{}

	go q.Run()

	return q
}

func (q *Quest) Run() {
	checkTicker := time.NewTicker(1 * time.Minute)
	regenerateTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-checkTicker.C:
			checkQuest()
		case <-regenerateTicker.C:
			regenerateQuest()
		}
	}
}

// regenerateQuest check expired quest and regenerate new quest for it
func regenerateQuest() {
}

// checkQuest ensure players get there quest recorded
func checkQuest() {

}

func abilityKillQuest(playerID string) error {
	// check player already obtain the quest
	pq, err := boiler.PlayersQuests(
		boiler.PlayersQuestWhere.PlayerID.EQ(playerID),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s ISNULL AND %s = %s",
				boiler.TableNames.Quests,
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.ID),
				qm.Rels(boiler.TableNames.PlayersQuests, boiler.PlayersQuestColumns.QuestID),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.DeletedAt),
				qm.Rels(boiler.TableNames.Quests, boiler.QuestColumns.Key),
				boiler.QuestKeyAbilityKill,
			),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to check player quest")
	}

	return nil
}
