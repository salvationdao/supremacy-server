package battle

import (
	"fmt"
	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type SystemBanMessageData struct {
	PlayerBan    *boiler.PlayerBan `json:"player_ban"`
	SystemPlayer *boiler.Player    `json:"system_player"`
	BannedPlayer *boiler.Player    `json:"banned_player"`
	FactionID    null.String       `json:"faction_id"`
	BanDuration  string            `json:"ban_duration"`
}

type SystemBanManager struct {
	SystemBanMassageChan chan *SystemBanMessageData

	teamKillCourtroom map[string]*TeamKillDefendant
	deadlock.RWMutex
}

func NewSystemBanManager() *SystemBanManager {
	sbm := &SystemBanManager{
		teamKillCourtroom:    make(map[string]*TeamKillDefendant),
		SystemBanMassageChan: make(chan *SystemBanMessageData),
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
	if tkj.GetNextCase() == "" {
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
	playerID           string
	relatedOfferingIDs []string
	sync.Mutex

	judgingCountdownSeconds int
	systemBanManager        *SystemBanManager
}

func newTeamKillDefendant(playerID string, sbm *SystemBanManager) *TeamKillDefendant {
	tkj := &TeamKillDefendant{
		playerID:                playerID,
		relatedOfferingIDs:      []string{},
		judgingCountdownSeconds: db.GetIntWithDefault(db.KeyJudgingCountdownSeconds, 3),
		systemBanManager:        sbm,
	}

	go tkj.startJudgingCycle()

	return tkj
}

func (tkj *TeamKillDefendant) addCase(relatedOfferingID string) {
	tkj.Lock()
	defer tkj.Unlock()

	if slices.IndexFunc(tkj.relatedOfferingIDs, func(offeringID string) bool { return offeringID == relatedOfferingID }) == -1 {
		tkj.relatedOfferingIDs = append(tkj.relatedOfferingIDs, relatedOfferingID)
	}
}

func (tkj *TeamKillDefendant) removeCase(relatedOfferingID string) {
	tkj.Lock()
	defer tkj.Unlock()

	index := slices.Index(tkj.relatedOfferingIDs, relatedOfferingID)
	if index == -1 {
		return
	}

	tkj.relatedOfferingIDs = slices.Delete(tkj.relatedOfferingIDs, index, index+1)
}

func (tkj *TeamKillDefendant) GetNextCase() string {
	tkj.Lock()
	defer tkj.Unlock()

	if len(tkj.relatedOfferingIDs) == 0 {
		return ""
	}

	return tkj.relatedOfferingIDs[0]
}

func (tkj *TeamKillDefendant) startJudgingCycle() {
	noInstanceDurationSeconds := 0

	for {
		time.Sleep(1 * time.Second)
		if offeringID := tkj.GetNextCase(); offeringID != "" {
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
			if tkj.GetNextCase() != "" {
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

func (tkj *TeamKillDefendant) judging(relatedOfferingID string) {
	tkj.Lock()
	defer tkj.Unlock()

	// just make sure the ability log is completed
	time.Sleep(time.Duration(tkj.judgingCountdownSeconds) * time.Second)

	// start the judgment
	bat, err := boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.AbilityOfferingID.EQ(relatedOfferingID),
		boiler.BattleAbilityTriggerWhere.PlayerID.EQ(null.StringFrom(tkj.playerID)),
		qm.Load(boiler.BattleAbilityTriggerRels.GameAbility),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("ability offering id", relatedOfferingID).Msg("Failed to get battle ability trigger from db")
		return
	}

	bhs, err := boiler.BattleHistories(
		boiler.BattleHistoryWhere.BattleAbilityOfferingID.EQ(null.StringFrom(relatedOfferingID)),
		boiler.BattleHistoryWhere.EventType.EQ(boiler.BattleEventKilled),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("related id", relatedOfferingID).Msg("Failed to get battle history from db")
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
		_, err = boiler.PlayerKillLogs(
			boiler.PlayerKillLogWhere.AbilityOfferingID.EQ(null.StringFrom(relatedOfferingID)),
		).UpdateAll(gamedb.StdConn, boiler.M{boiler.PlayerKillLogColumns.IsVerified: true})
		if err != nil {
			gamelog.L.Error().Err(err).Str("ability offering id", relatedOfferingID).Msg("Failed to set player kill verify flag to true.")
		}
		return
	}

	offeringIDs := []interface{}{relatedOfferingID}

	// if maximum team kill tolerant
	if bat.R != nil && bat.R.GameAbility != nil && bat.R.GameAbility.MaximumTeamKillTolerantCount > 0 {
		ga := bat.R.GameAbility
		q := fmt.Sprintf(
			`
			SELECT DISTINCT (%[2]s) , %[7]s
			FROM %[1]s 
			WHERE %[3]s = $1 AND %[4]s = $2 AND %[5]s = TRUE AND %[6]s = FALSE
			ORDER BY %[7]s;
		`,
			boiler.TableNames.PlayerKillLog,               // 1
			boiler.PlayerKillLogColumns.AbilityOfferingID, // 2
			boiler.PlayerKillLogColumns.PlayerID,          // 3
			boiler.PlayerKillLogColumns.GameAbilityID,     // 4
			boiler.PlayerKillLogColumns.IsTeamKill,        // 5
			boiler.PlayerKillLogColumns.IsVerified,        // 6
			boiler.PlayerKillLogColumns.CreatedAt,         // 7
		)

		rows, err := gamedb.StdConn.Query(q, tkj.playerID, ga.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("related game ability id", bat.R.GameAbility.ID).Msg("Failed to load related player ability kill logs.")
			return
		}

		// re-declare offering id list
		offeringIDs = []interface{}{}
		for rows.Next() {
			offeringID := ""
			createdAt := time.Now() // just for sql
			err = rows.Scan(&offeringID, &createdAt)
			if err != nil {
				gamelog.L.Error().Err(err).Str("related game ability id", bat.R.GameAbility.ID).Msg("Failed to scan offering id.")
				return
			}

			offeringIDs = append(offeringIDs, offeringID)
		}

		// skip, if offingID count hasn't met maximum tolerant count
		if len(offeringIDs) < ga.MaximumTeamKillTolerantCount {
			return
		}

		// trim to maximum tolerant count
		offeringIDs = offeringIDs[:ga.MaximumTeamKillTolerantCount]
	}

	// ban player

	// check how many system location ban the player has involved
	systemTeamKillDefaultReason := db.GetStrWithDefault(db.KeySystemBanTeamKillDefaultReason, "Team kill activity is detected")

	pbs, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BanFrom.EQ(boiler.BanFromTypeSYSTEM),
		boiler.PlayerBanWhere.BannedPlayerID.EQ(bat.PlayerID.String),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
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

	_, err = boiler.PlayerKillLogs(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				qm.Rels(boiler.TableNames.PlayerKillLog, boiler.PlayerKillLogColumns.AbilityOfferingID),
			),
			offeringIDs...,
		),
	).UpdateAll(gamedb.StdConn,
		boiler.M{
			boiler.PlayerKillLogColumns.IsVerified:       true,
			boiler.PlayerKillLogColumns.RelatedPlayBanID: null.StringFrom(playerBan.ID),
		})
	if err != nil {
		gamelog.L.Error().Err(err).Str("ability offering id", relatedOfferingID).Msg("Failed to set player kill verify flag to true.")
	}

	// send player ban to chat
	pb, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.ID.EQ(playerBan.ID),
		qm.Load(
			boiler.PlayerBanRels.BannedBy,
			qm.Select(
				boiler.PlayerColumns.ID,
				boiler.PlayerColumns.Username,
				boiler.PlayerColumns.FactionID,
				boiler.PlayerColumns.Gid,
			),
		),
		qm.Load(
			boiler.PlayerBanRels.BannedPlayer,
			qm.Select(
				boiler.PlayerColumns.ID,
				boiler.PlayerColumns.Username,
				boiler.PlayerColumns.FactionID,
				boiler.PlayerColumns.Gid,
			),
		),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player ban detail")
		return
	}

	banDuration := "1 hr"
	if banDurationHours > 1 {
		banDuration = fmt.Sprintf("%d hrs", banDurationHours)
	}

	tkj.systemBanManager.SystemBanMassageChan <- &SystemBanMessageData{pb, pb.R.BannedBy, pb.R.BannedPlayer, pb.R.BannedPlayer.FactionID, banDuration}
}
