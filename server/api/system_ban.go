package api

import (
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type SystemBanManager struct {
	teamKillJudgmentSeat     map[string]bool
	judgmentCountdownSeconds int

	sync.RWMutex
}

func NewSystemBanManager() *SystemBanManager {
	sbm := &SystemBanManager{
		teamKillJudgmentSeat:     make(map[string]bool),
		judgmentCountdownSeconds: db.GetIntWithDefault(db.KeyJudgmentCountdownSeconds, 3),
	}

	return sbm
}

func (sbm *SystemBanManager) EnterSaboJudgmentSeat(relativeOfferingID string) {
	sbm.Lock()
	defer sbm.Unlock()

	// skip,  if player is already on the judgment seat
	if ok := sbm.teamKillJudgmentSeat[relativeOfferingID]; ok {
		return
	}

	// add player to judgment seat
	sbm.teamKillJudgmentSeat[relativeOfferingID] = true

	go sbm.TeamKillJudgment(relativeOfferingID)

}

func (sbm *SystemBanManager) TeamKillJudgment(relativeOfferingID string) {
	defer func() {
		// remove player from team kill judgment seat
		sbm.Lock()
		delete(sbm.teamKillJudgmentSeat, relativeOfferingID)
		sbm.Unlock()
	}()

	// just make sure the ability log is completed
	time.Sleep(time.Duration(sbm.judgmentCountdownSeconds) * time.Second)

	// start the judgment
	bat, err := boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.AbilityOfferingID.EQ(relativeOfferingID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("ability offering id", relativeOfferingID).Msg("Failed to get battle ability trigger from db")
		return
	}

	bhs, err := boiler.BattleHistories(
		boiler.BattleHistoryWhere.RelatedID.EQ(null.StringFrom(relativeOfferingID)),
		boiler.BattleHistoryWhere.EventType.EQ(boiler.BattleEventKilled),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("related id", relativeOfferingID).Msg("Failed to get battle history from db")
		return
	}

	teamKillCount := 0
	for _, bh := range bhs {
		bm, err := boiler.BattleMechs(
			qm.Select(boiler.BattleMechColumns.FactionID),
			boiler.BattleMechWhere.BattleID.EQ(bh.BattleID),
			boiler.BattleMechWhere.MechID.EQ(bh.WarMachineOneID),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("battle id", bh.BattleID).Str("war machine id", bh.WarMachineOneID).Msg("Failed to get battle mech from db")
			return
		}

		// check if it is team kill
		if bat.FactionID == bm.FactionID {
			teamKillCount += 1
		} else {
			teamKillCount -= 1
		}
	}

	// if team kill
	if teamKillCount > 0 {
		// check

	}
}
