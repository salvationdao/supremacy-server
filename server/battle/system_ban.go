package battle

import (
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type SystemBanManager struct {
	teamKillCourtroom map[string]*TeamKillDefendant

	sync.RWMutex
}

func NewSystemBanManager() *SystemBanManager {
	sbm := &SystemBanManager{
		teamKillCourtroom: make(map[string]*TeamKillDefendant),
	}

	return sbm
}

func (sbm *SystemBanManager) HasOngoingTeamKillCases(playerID string) bool {
	sbm.Lock()
	defer sbm.Unlock()

	// not in the courtroom
	tkj, ok := sbm.teamKillCourtroom[playerID]
	if !ok {
		return false
	}

	// does not have any instance
	if tkj.GetCase() == "" {
		return false
	}

	return true
}

func (sbm *SystemBanManager) SendToTeamKillCourtroom(playerID string, relativeOfferingID string) {
	sbm.Lock()
	defer sbm.Unlock()

	tkj, ok := sbm.teamKillCourtroom[playerID]
	if !ok {
		tkj = newTeamKillDefendant(playerID, sbm)
		sbm.teamKillCourtroom[playerID] = tkj
	}

	tkj.addCase(relativeOfferingID)
}

type TeamKillDefendant struct {
	playerID             string
	relatedOfferingIDMap map[string]bool
	sync.Mutex

	judgingCountdownSeconds int
	systemBanManager        *SystemBanManager
}

func newTeamKillDefendant(playerID string, sbm *SystemBanManager) *TeamKillDefendant {
	tkj := &TeamKillDefendant{
		playerID:                playerID,
		relatedOfferingIDMap:    make(map[string]bool),
		judgingCountdownSeconds: db.GetIntWithDefault(db.KeyJudgingCountdownSeconds, 3),
		systemBanManager:        sbm,
	}

	go tkj.startJudgingCycle()

	return tkj
}

func (tkj *TeamKillDefendant) addCase(relatedOfferingID string) {
	tkj.Lock()
	defer tkj.Unlock()

	ok := tkj.relatedOfferingIDMap[relatedOfferingID]
	if !ok {
		tkj.relatedOfferingIDMap[relatedOfferingID] = true
	}
}

func (tkj *TeamKillDefendant) removeCase(relatedOfferingID string) {
	tkj.Lock()
	defer tkj.Unlock()

	delete(tkj.relatedOfferingIDMap, relatedOfferingID)
}

func (tkj *TeamKillDefendant) GetCase() string {
	tkj.Lock()
	defer tkj.Unlock()

	// return the first offering id of the map
	for offeringID := range tkj.relatedOfferingIDMap {
		return offeringID
	}

	return ""
}

func (tkj *TeamKillDefendant) startJudgingCycle() {
	noInstanceDurationSeconds := 0

	for {
		time.Sleep(1 * time.Second)
		if offeringID := tkj.GetCase(); offeringID != "" {
			tkj.judging(offeringID)
			tkj.removeCase(offeringID)
			noInstanceDurationSeconds = 0
			continue
		}

		noInstanceDurationSeconds += 1

		// clean up judgment cycle if there is no instance for 30 seconds
		if noInstanceDurationSeconds > 30 {
			// pause system ban manager
			tkj.systemBanManager.Lock()

			// final check instance
			if tkj.GetCase() != "" {
				// reset countdown and restart
				noInstanceDurationSeconds = 0
				tkj.systemBanManager.Unlock()
				continue
			}

			// remove instance and return the function
			delete(tkj.systemBanManager.teamKillCourtroom, tkj.playerID)
			tkj.systemBanManager.Unlock()
			return
		}
	}
}

func (tkj *TeamKillDefendant) judging(relativeOfferingID string) {
	tkj.Lock()
	defer tkj.Unlock()

	// just make sure the ability log is completed
	time.Sleep(time.Duration(tkj.judgingCountdownSeconds) * time.Second)

	// start the judgment
	bat, err := boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.AbilityOfferingID.EQ(relativeOfferingID),
		boiler.BattleAbilityTriggerWhere.PlayerID.IsNotNull(),
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

	// skip, if not team kill
	if teamKillCount <= 0 {
		return
	}

	// ban player

	// check how many system location ban the player has involved
	systemTeamKillDefaultReason := db.GetStrWithDefault(db.KeySystemBanTeamKillDefaultReason, "Team kill activity is detected (SYSTEM)")

	pbs, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BanFrom.EQ(boiler.BanFromTypeSYSTEM),
		boiler.PlayerBanWhere.BannedPlayerID.EQ(bat.PlayerID.String),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
		boiler.PlayerBanWhere.Reason.EQ(systemTeamKillDefaultReason),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle id", bat.BattleID).Msg("Failed to load previous system team kill ban")
		return
	}

	// add a new ban
	systemTeamKillBanBaseDurationHours := db.GetIntWithDefault(db.KeySystemBanTeamKillBanBaseDurationHours, 1)
	banDurationMultiplier := db.GetIntWithDefault(db.KeySystemBanTeamKillBanDurationMultiplier, 4)
	systemTeamKillPermanentBanBottomLine := db.GetIntWithDefault(db.KeySystemBanTeamKillPermanentBanBottomLineHours, 24*7)

	banDurationHours := systemTeamKillBanBaseDurationHours
	for range pbs {
		banDurationHours = banDurationHours * banDurationMultiplier
	}

	banUntil := time.Now().Add(time.Duration(banDurationHours) * time.Hour)
	// set end time to 100 years, if permanent ban bottom line is reached
	if banDurationHours > systemTeamKillPermanentBanBottomLine {
		banUntil = time.Now().AddDate(100, 0, 0)
	}

	playerBan := boiler.PlayerBan{
		BanFrom:           boiler.BanFromTypeSYSTEM,
		BannedByID:        server.SupremacyBattleUserID,
		BannedPlayerID:    bat.PlayerID.String,
		Reason:            systemTeamKillDefaultReason,
		EndAt:             banUntil,
		BanLocationSelect: true,
	}

	err = playerBan.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("player ban", playerBan).Msg("Failed to insert system team kill ban into db")
		return
	}
}
