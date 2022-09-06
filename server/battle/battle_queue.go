package battle

import (
	"fmt"
	"math"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/ninja-syndicate/ws"
	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
)

type StartBattlesFn func()

type Deploy struct {
	FactionID    string
	StartBattles StartBattlesFn
}

type BattleQueueManager struct {
	arenaManager        *ArenaManager
	blacklistedOwnerIDs []string // [mech owner IDs]

	// KVs
	QueueTickerIntervalSeconds int

	closed *atomic.Bool
	deadlock.RWMutex
}

func NewBattleQueueSystem(rpc *xsyn_rpcclient.XsynXrpcClient) (*BattleQueueManager, error) {
	zaiQueueCount, err := db.GetNumberOfMechsInQueueFromFactionID(server.ZaibatsuFactionID)
	if err != nil {
		return nil, err
	}
	rmQueueCount, err := db.GetNumberOfMechsInQueueFromFactionID(server.RedMountainFactionID)
	if err != nil {
		return nil, err
	}
	bcQueueCount, err := db.GetNumberOfMechsInQueueFromFactionID(server.BostonCyberneticsFactionID)
	if err != nil {
		return nil, err
	}

	cullCount := int(math.Min(math.Min(float64(zaiQueueCount), float64(rmQueueCount)), float64(bcQueueCount)))
	if (cullCount % db.FACTION_MECH_LIMIT) != 0 {
		cullCount = cullCount - (cullCount % db.FACTION_MECH_LIMIT)
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	zaiQueues, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.FactionID.EQ(server.ZaibatsuFactionID),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.BattleQueueColumns.InsertedAt)),
		qm.Limit(int(zaiQueueCount)-cullCount),
		qm.Load(boiler.BattleQueueRels.Fee),
	).All(tx)
	if err != nil {
		return nil, err
	}
	rmQueues, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.FactionID.EQ(server.RedMountainFactionID),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.BattleQueueColumns.InsertedAt)),
		qm.Limit(int(rmQueueCount)-cullCount),
		qm.Load(boiler.BattleQueueRels.Fee),
	).All(tx)
	if err != nil {
		return nil, err
	}
	bcQueues, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.FactionID.EQ(server.BostonCyberneticsFactionID),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.BattleQueueColumns.InsertedAt)),
		qm.Limit(int(bcQueueCount)-cullCount),
		qm.Load(boiler.BattleQueueRels.Fee),
	).All(tx)
	if err != nil {
		return nil, err
	}

	deletables := boiler.BattleQueueSlice{}
	deletables = append(deletables, zaiQueues...)
	deletables = append(deletables, rmQueues...)
	deletables = append(deletables, bcQueues...)

	refundables := boiler.BattleQueueFeeSlice{}
	for _, d := range deletables {
		refundables = append(refundables, d.R.Fee)
	}

	n, err := deletables.DeleteAll(tx)
	if err != nil {
		return nil, err
	}
	if int(n) != len(deletables) {
		return nil, fmt.Errorf("could not refund invalid battle queue entries")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	for _, r := range refundables {
		// Refund deleted battle queue entry fees
		_, err = rpc.RefundSupsMessage(r.PaidTXID.String)
		if err != nil {
			gamelog.L.Error().Str("txID", r.PaidTXID.String).Err(err).Msg("failed to refund queue fee")
			continue
		}
	}

	qs := &BattleQueueManager{
		blacklistedOwnerIDs:        []string{},
		QueueTickerIntervalSeconds: db.GetIntWithDefault(db.KeyQueueTickerIntervalSeconds, 1),
	}

	go qs.BattleQueueUpdater()

	return qs, nil
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
			l.Debug().Msg("moving entries from queue backlog to battle queue")
			// 1. load whole queue backlog
			// 2. if there are missing mechs, fill in the missing mech from backlog to queue
			// 3. otherwise, insert sets (3 mech per fection from each faction)
			// 4. check idle arena, and run arena.initNextBattle()

			blacklisted, err := db.GetPreviousBattleOwnerIDs()
			if err != nil {
				l.Warn().Err(err).Msg("Failed to fetch blacklisted owner ids")
				continue
			}

			// Get mechs from backlog
			_zaiPendingMechs, err := db.GetPendingMechsFromFactionID(server.ZaibatsuFactionID, blacklisted, db.FACTION_MECH_LIMIT)
			if err != nil {
				l.Warn().Err(err).Msg("Failed to fetch pending backlogged mechs")
				continue
			}
			if len(_zaiPendingMechs) < db.FACTION_MECH_LIMIT {
				continue
			}
			_rmPendingMechs, err := db.GetPendingMechsFromFactionID(server.RedMountainFactionID, blacklisted, db.FACTION_MECH_LIMIT)
			if err != nil {
				l.Warn().Err(err).Msg("Failed to fetch pending backlogged mechs")
				continue
			}
			if len(_rmPendingMechs) < db.FACTION_MECH_LIMIT {
				continue
			}
			_bcPendingMechs, err := db.GetPendingMechsFromFactionID(server.BostonCyberneticsFactionID, blacklisted, db.FACTION_MECH_LIMIT)
			if err != nil {
				l.Warn().Err(err).Msg("Failed to fetch pending backlogged mechs")
				continue
			}
			if len(_bcPendingMechs) < db.FACTION_MECH_LIMIT {
				continue
			}

			func(zaiPendingMechs, rmPendingMechs, bcPendingMechs boiler.BattleQueueBacklogSlice) {
				tx, err := gamedb.StdConn.Begin()
				if err != nil {
					l.Warn().Err(err).Msg("failed to create db transaction")
					return
				}
				defer tx.Rollback()

				pendingMechs := []*boiler.BattleQueueBacklog{}
				pendingMechs = append(pendingMechs, zaiPendingMechs...)
				pendingMechs = append(pendingMechs, rmPendingMechs...)
				pendingMechs = append(pendingMechs, bcPendingMechs...)
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
						l.Warn().Err(err).Msg("failed to insert into battle queue")
						return
					}

					pm.Delete(tx)
					if err != nil {
						l.Warn().Err(err).Msg("failed to remove from battle queue backlog")
						return
					}
				}
				err = tx.Commit()
				if err != nil {
					l.Warn().Err(err).Msg("failed to commit db transaction")
					return
				}

				if len(pendingMechs) > 0 {
					eta, err := db.GetBattleETASecondsFromMechID(pendingMechs[0].MechID, pendingMechs[0].FactionID)
					if err != nil {
						l.Warn().Err(err).Msg("failed to get battle eta for mech")
					}

					for _, pm := range pendingMechs {
						ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", pm.FactionID, pm.MechID), WSPlayerAssetMechQueueSubscribe, &server.MechArenaInfo{
							Status:           server.MechArenaStatusQueue,
							CanDeploy:        false,
							BattleETASeconds: null.Int64From(eta),
						})
					}
				}

				// Call BeginBattle on idle arenas
				for _, a := range qs.arenaManager.IdleArenas() {
					a.BeginBattle()
				}
			}(_zaiPendingMechs, _rmPendingMechs, _bcPendingMechs)

			go func() {
				CalcNextQueueStatus(server.ZaibatsuFactionID)
				CalcNextQueueStatus(server.RedMountainFactionID)
				CalcNextQueueStatus(server.BostonCyberneticsFactionID)
			}()
		}
	}
}
