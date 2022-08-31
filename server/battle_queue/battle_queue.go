package battle_queue

import (
	"server/db"
	"server/gamelog"
	"time"

	"github.com/sasha-s/go-deadlock"
	"go.uber.org/atomic"
)

type Deploy struct {
	MechIDs []string // mech ids
}

type BattleQueueManager struct {
	lastBattleMechs []string // [mech_id]

	// KVs
	QueueTickerIntervalSeconds int

	// on mech deploy
	Deploy chan *Deploy

	closed *atomic.Bool
	deadlock.RWMutex
}

func NewBattleQueueSystem() *BattleQueueManager {
	qs := &BattleQueueManager{
		lastBattleMechs:            []string{},
		QueueTickerIntervalSeconds: db.GetIntWithDefault(db.KeyQueueTickerIntervalSeconds, 5),
	}

	go qs.BattleQueueUpdater()

	return qs
}

func (qs *BattleQueueManager) BattleQueueUpdater() {
	queueTicker := time.NewTicker(time.Duration(qs.QueueTickerIntervalSeconds) * time.Second)

	defer func() {
		queueTicker.Stop()
		qs.closed.Store(true)
	}()

	for {
		select {
		case <-queueTicker.C:
			gamelog.L.Debug().Msg("moving entries from queue backlog to battle queue")
			// Every 5 seconds (set by KV), move exactly 3 entries (per faction) from battle_queue_backlog to the
			// battle_queue table, only if they satisfy the following criteria:
			// - No two mechs can belong to the same owner, unless the faction's battle_queue is empty
			// - Mechs did not participate in the previous battle (ignoring server resets)

			// case deploy := <-qs.Deploy:
		}
	}
}
