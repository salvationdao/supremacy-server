package battle_queue

import (
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"
)

type StartBattlesFn func()

type Deploy struct {
	FactionID    string
	StartBattles StartBattlesFn
}

type BattleQueueManager struct {
	blacklistedOwnerIDs   []string // [mech owner IDs]
	canNextBattleCommence bool     // the next battle will not start if this is false

	// KVs
	QueueTickerIntervalSeconds int

	// on mech(s) deploy
	Deploy chan *Deploy

	rmMissingCount  int // 0
	bcMissingCount  int // 0
	zaiMissingCount int // 0

	closed *atomic.Bool
	deadlock.RWMutex
}

func NewBattleQueueSystem() *BattleQueueManager {
	qs := &BattleQueueManager{
		blacklistedOwnerIDs:        []string{},
		QueueTickerIntervalSeconds: db.GetIntWithDefault(db.KeyQueueTickerIntervalSeconds, 10),
	}

	go qs.BattleQueueUpdater()

	return qs
}

func (qs *BattleQueueManager) CurrentMechOwnerBlacklist() []string {
	qs.RLock()
	defer qs.RUnlock()

	return qs.blacklistedOwnerIDs
}

func (qs *BattleQueueManager) SetMechOwnerBlacklist(newBlacklist []string) {
	qs.Lock()
	defer qs.Unlock()

	qs.blacklistedOwnerIDs = newBlacklist
}

func (qs *BattleQueueManager) MovePendingMechs(factionID string, limit int) {
	l := gamelog.L.With().Str("func", "MovePendingMechs").Logger()

	qs.RLock()
	// Get mechs from backlog
	pendingMechs, err := db.GetPendingMechsFromFactionID(factionID, qs.blacklistedOwnerIDs, limit)
	if err != nil {
		l.Error().Err(err).Msg("Failed to fetch pending backlogged mechs")
		return
	}
	qs.RUnlock()

	if len(pendingMechs) < limit {
		return
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		l.Error().Err(err).Msg("failed to create db transaction")
		return
	}
	defer tx.Rollback()

	// Remove the mechs from the backlog and place them in the battle queue
	for _, pm := range pendingMechs {
		if pm == nil {
			continue
		}

		bq := boiler.BattleQueue{
			MechID:             pm.MechID,
			FactionID:          pm.FactionID,
			OwnerID:            pm.OwnerID,
			FeeID:              pm.FeeID,
			QueueFeeTXID:       pm.QueueFeeTXID,
			QueueFeeTXIDRefund: pm.QueueFeeTXIDRefund,
			UpdatedAt:          pm.UpdatedAt,
			QueuedAt:           pm.QueuedAt,
		}
		err := bq.Insert(tx, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("failed to insert into battle queue")
			return
		}

		pm.Delete(tx)
		if err != nil {
			l.Error().Err(err).Msg("failed to remove from battle queue backlog")
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		l.Error().Err(err).Msg("failed to commit db transaction")
		return
	}

	qs.MovePendingMechs(factionID, db.FACTION_MECH_LIMIT)
}

func (qs *BattleQueueManager) BattleQueueUpdater() {
	l := gamelog.L.With().Str("func", "BattleQueueUpdater").Logger()

	queueTicker := time.NewTicker(time.Duration(qs.QueueTickerIntervalSeconds) * time.Second)

	defer func() {
		queueTicker.Stop()
		qs.closed.Store(true)
	}()

	for {
		select {
		case <-queueTicker.C:
			gamelog.L.Debug().Msg("moving entries from queue backlog to battle queue")
			// Every 10 seconds (set by KV), check whether the battle queue has enough entries to start a
			// battle. If not, then pause the battle start

		case deploy := <-qs.Deploy:
			// todo:
			// 1. load whole queue backlog
			// 2. if there are missing mechs, fill in the missing mech from backlog to queue
			// 3. otherwise, insert sets (3 mech per fection from each faction)
			// 4. check idle arena, and run arena.initNextBattle()

			// On mech deploy, move entries from the battle_queue_backlog to the
			// battle_queue table, only if they satisfy the following criteria:
			// - No two mechs can belong to the same owner, unless the faction's battle_queue is empty
			// - Mechs did not participate in the previous battle (ignoring server resets)
			inQueueCount, err := db.GetNumberOfMechsInQueueFromFactionID(deploy.FactionID)
			if err != nil {
				l.Error().Err(err).Msg("Failed to fetch pending backlogged mechs")
				continue
			}

			// Determine how many faction mechs we want to transfer from the queue backlog
			limit := db.FACTION_MECH_LIMIT
			if (inQueueCount % db.FACTION_MECH_LIMIT) != 0 {
				limit = int(db.FACTION_MECH_LIMIT - (inQueueCount % db.FACTION_MECH_LIMIT))
			}

			qs.MovePendingMechs(deploy.FactionID, limit)

			deploy.StartBattles()
		}
	}
}
