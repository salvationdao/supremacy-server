package battle

import (
	"database/sql"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"sync"
)

type RepairSystem struct {
	passport *xsyn_rpcclient.XsynXrpcClient
	offerMap map[string]*RepairOffer
	sync.RWMutex
}

func New(pp *xsyn_rpcclient.XsynXrpcClient) *RepairSystem {
	rs := &RepairSystem{
		passport: pp,
		offerMap: make(map[string]*RepairOffer),
	}

	return rs
}

type RepairOffer struct {
	*boiler.RepairOffer
	isClosed                bool // protected by channel
	repairAgentRegisterChan chan *boiler.Player
	repairAgentCancelChan   chan *boiler.Player
	repairAgentProgressChan chan *boiler.Player
	instantRepairChan       chan bool
	onClose                 func()
}

func (rs *RepairSystem) OfferRepairJob(ro *boiler.RepairOffer) {
	offer := &RepairOffer{
		RepairOffer:             ro,
		isClosed:                false,
		repairAgentRegisterChan: make(chan *boiler.Player),
		repairAgentCancelChan:   make(chan *boiler.Player),
		repairAgentProgressChan: make(chan *boiler.Player),
		instantRepairChan:       make(chan bool),
		onClose: func() {
			rs.Lock()
			defer rs.Unlock()

			delete(rs.offerMap, ro.ID)
		},
	}

	rs.offerMap[ro.ID] = offer

	go offer.start()

}

func (ro *RepairOffer) start() {
	// this is the ONLY place to close a repair offer
	defer ro.onClose()

	for {
		if ro.isClosed {
			return
		}

		select {
		case p := <-ro.repairAgentRegisterChan:
			// check repair agent exists
			ra, err := boiler.FindRepairAgent(gamedb.StdConn, ro.ID, p.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Err(err).Str("repair offer id", ro.ID).Str("repair agent id", p.ID).Msg("Failed to load repair agent")
				continue
			}

			// if already registered
			if ra != nil {
				continue
			}

			ra = &boiler.RepairAgent{
				RepairOfferID: ro.ID,
				AgentID:       p.ID,
			}
			err = ra.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Str("repair offer id", ro.ID).Str("repair agent id", p.ID).Msg("Failed to insert repair agent")
				continue
			}

		case p := <-ro.repairAgentProgressChan:
			ra, err := boiler.FindRepairAgent(gamedb.StdConn, ro.ID, p.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Err(err).Str("repair offer id", ro.ID).Str("repair agent id", p.ID).Msg("Failed to load repair agent")
				continue
			}

			// repair agent is not registered
			if ra == nil {
				continue
			}

			ra.Progress += 1
			if ra.Progress == 100 {
				ra.Status = boiler.RepairAgentStatusSUCCESS
			}
			_, err = ra.Update(gamedb.StdConn, boil.Whitelist(
				boiler.RepairAgentColumns.Progress,
				boiler.RepairAgentColumns.Status,
			))
			if err != nil {
				gamelog.L.Error().Err(err).Str("repair offer id", ro.ID).Str("repair agent id", p.ID).Msg("Failed to update repair progress")
				continue
			}

			// close all other repair agents
			if ra.Status == boiler.RepairAgentStatusSUCCESS {
				ro.isClosed = true

				// update all the incomplete repair agent
				go func() {
					_, err = boiler.RepairAgents(
						boiler.RepairAgentWhere.RepairOfferID.EQ()
							boiler.RepairAgentWhere.Status.EQ("WIP"),
					).UpdateAll(gamedb.StdConn, boiler.M{})
					if err != nil {

						return
					}
				}()
			}

		case <-ro.instantRepairChan:
			ro.isClosed = true
		}
	}
}
