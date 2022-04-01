package api

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/ninja-software/terror/v2"

	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PunishVotePhase string

const (
	PunishVotePhaseVoting PunishVotePhase = "VOTING"
	PunishVotePhaseHold   PunishVotePhase = "HOLD"
)

type PunishVoteStage struct {
	Phase   PunishVotePhase
	EndTime time.Time
}

type PunishVoteTracker struct {
	FactionID string
	// punish vote tracker
	Stage *PunishVoteStage

	// message bus
	MessageBus *messagebus.MessageBus
	// broadcast result
	broadcastResult chan *PunishVoteResult

	// mutex lock for issue vote
	CurrentPunishVote *PunishVoteInstance
	deadlock.RWMutex

	api *API
}

type PunishVoteInstance struct {
	ID                 string
	PlayerPool         map[string]bool
	AgreedPlayerIDs    map[string]bool
	DisagreedPlayerIDs map[string]bool
	StartedAt          time.Time
	EndedAt            time.Time
}

type PunishVote struct {
	PunishVoteID string
	playerID     string
	IsAgreed     bool
}

func (api *API) PunishVoteTrackerSetup() {
	// get factions
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup faction punish vote tracker")
		return
	}

	for _, f := range factions {
		// initialise
		pvt := &PunishVoteTracker{
			FactionID:       f.ID,
			MessageBus:      api.MessageBus,
			broadcastResult: make(chan *PunishVoteResult),
			Stage:           &PunishVoteStage{PunishVotePhaseHold, time.Now().AddDate(1, 0, 0)},
			api:             api,
		}

		// start punish vote tracker
		go pvt.Run()

		// store punish vote instance of each faction
		api.FactionPunishVote[f.ID] = pvt
	}
}

func (pvt *PunishVoteTracker) Run() {
	mainTicker := time.NewTicker(1 * time.Second)

	// run debounce broadcast punish vote result
	go pvt.debounceBroadcastResult()

	for {
		select {
		case <-mainTicker.C:
			switch pvt.Stage.Phase {
			case PunishVotePhaseVoting:
				pvt.VotingPhaseProcess()
			case PunishVotePhaseHold:
				pvt.HoldingPhaseProcess()
			}
		}
	}
}

// CurrentEligiblePlayers return a map of current active players that has positive ability kills
// NOTE: Ensure the function is fired, when the current punish vote is exist
//       Otherwise, it will panic!!!
func (pvt *PunishVoteTracker) CurrentEligiblePlayers() map[string]bool {
	fap, ok := pvt.api.FactionActivePlayers[pvt.FactionID]
	if !ok {
		return nil
	}

	fap.Lock()
	defer fap.Unlock()

	result := make(map[string]bool)
	dbSearchList := []string{}

	for playerID := range fap.Map {
		dbSearchList = append(dbSearchList, playerID)
	}

	// get active player with positive ability kill count
	if len(dbSearchList) > 0 {
		us, err := boiler.UserStats(
			qm.Select(
				boiler.UserStatColumns.ID,
				boiler.UserStatColumns.KillCount,
			),
			boiler.UserStatWhere.KillCount.GT(0),
			boiler.UserStatWhere.ID.IN(dbSearchList),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get player kill count from db")
		}
		for _, player := range us {
			result[player.ID] = true
		}
	}

	spew.Dump(result)

	// fill the list with voted players
	for pid := range pvt.CurrentPunishVote.AgreedPlayerIDs {
		result[pid] = true
	}
	for pid := range pvt.CurrentPunishVote.DisagreedPlayerIDs {
		result[pid] = true
	}

	return result
}

// VotingPhaseProcess process when it is in voting phase
func (pvt *PunishVoteTracker) VotingPhaseProcess() {
	pvt.Lock()
	defer pvt.Unlock()

	// skip, if the vote ended early or vote still going on
	if pvt.CurrentPunishVote == nil || pvt.Stage.Phase != PunishVotePhaseVoting || pvt.Stage.EndTime.After(time.Now()) {
		return
	}

	// get latest eligible user list
	playerPool := pvt.CurrentEligiblePlayers()

	// vote passed, if the amount of the agreed players pass 50%
	if len(pvt.CurrentPunishVote.AgreedPlayerIDs) > len(playerPool)/2 {
		err := pvt.VotePassed()
		if err != nil {
			gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msgf("Failed to process passed vote due to %s", err.Error())
			return
		}
	}

	// Otherwise, vote is failed
	err := pvt.VoteFailed()
	if err != nil {
		gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msgf("Failed to process failed vote due to %s", err.Error())
		return
	}
}

// HoldingPhaseProcess process when vote is in hold phase
func (pvt *PunishVoteTracker) HoldingPhaseProcess() {
	pvt.Lock()
	defer pvt.Unlock()

	if pvt.Stage.Phase != PunishVotePhaseHold {
		return
	}

	// reset stage time
	pvt.Stage.EndTime = time.Now().AddDate(1, 0, 0)

	// get next punish issue from db
	punishVote, err := boiler.PunishVotes(
		boiler.PunishVoteWhere.FactionID.EQ(pvt.FactionID),
		boiler.PunishVoteWhere.Status.EQ(string(PunishVoteStatusPending)),
		qm.OrderBy(boiler.PunishVoteColumns.CreatedAt),
		qm.Load(boiler.PunishVoteRels.PunishOption),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("faction id", pvt.FactionID).Err(err).Msg("Failed to load new punish vote from db")
		return
	}

	// skip, if there is no punish vote
	if punishVote == nil {
		return
	}

	now := time.Now()
	endTime := now.Add(30 * time.Second)

	// update current punish vote, start/end time
	punishVote.StartedAt = null.TimeFrom(now)
	punishVote.EndedAt = null.TimeFrom(endTime)
	_, err = punishVote.Update(gamedb.StdConn, boil.Whitelist(boiler.PunishVoteColumns.StartedAt, boiler.PunishVoteColumns.EndedAt))
	if err != nil {
		gamelog.L.Error().Str("punish vote id", punishVote.ID).Err(err).Msg("Failed to update the start time of the punish vote")
		return
	}

	// initialise a new punish vote
	pvt.CurrentPunishVote = &PunishVoteInstance{
		ID:                 punishVote.ID,
		StartedAt:          punishVote.StartedAt.Time,
		EndedAt:            punishVote.EndedAt.Time,
		AgreedPlayerIDs:    make(map[string]bool),
		DisagreedPlayerIDs: make(map[string]bool),
	}

	// initialise current eligible players
	pvt.CurrentPunishVote.PlayerPool = pvt.CurrentEligiblePlayers()

	// change stage
	pvt.Stage.Phase = PunishVotePhaseVoting
	pvt.Stage.EndTime = endTime

	// broadcast new vote to online faction users
	pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, pvt.FactionID)), &PunishVoteResponse{
		PunishVote:   punishVote,
		PunishOption: punishVote.R.PunishOption,
	})
}

func (pvt *PunishVoteTracker) Vote(punishVoteID string, playerID string, isAgreed bool) error {
	pvt.Lock()
	defer pvt.Unlock()
	// check voting phase and targeted vote is available
	if pvt.Stage.Phase != PunishVotePhaseVoting || pvt.Stage.EndTime.Before(time.Now()) {
		return terror.Error(terror.ErrInvalidInput, "invalid voting phase")
	}

	if pvt.CurrentPunishVote == nil || pvt.CurrentPunishVote.ID != punishVoteID {
		return terror.Error(terror.ErrInvalidInput, "Punish vote id is mismatched")
	}

	// check player has voted
	if _, ok := pvt.CurrentPunishVote.AgreedPlayerIDs[playerID]; ok {
		return terror.Error(terror.ErrForbidden, "Player has already voted")
	}
	if _, ok := pvt.CurrentPunishVote.DisagreedPlayerIDs[playerID]; ok {
		return terror.Error(terror.ErrForbidden, "Player has already voted")
	}

	// store player's vote result into database
	pbv := &boiler.PlayersPunishVote{
		PunishVoteID: pvt.CurrentPunishVote.ID,
		PlayerID:     playerID,
		IsAgreed:     isAgreed,
	}
	err := pbv.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("punish_vote_id", pvt.CurrentPunishVote.ID).Str("player_id", playerID).Err(err).Msg("Failed to insert player vote result into db")
		return terror.Error(err, "Failed to insert player")
	}

	// update result
	if isAgreed {
		pvt.CurrentPunishVote.AgreedPlayerIDs[playerID] = true
		// check result
		if len(pvt.CurrentPunishVote.AgreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 && pvt.VerifyResult(isAgreed) {
			err := pvt.VotePassed()
			if err != nil {
				gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msgf("Failed to process failed vote due to %s", err.Error())
				return terror.Error(err, "Failed to process the result")
			}

			// broadcast undefined to clean up the form in the frontend
			pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, pvt.FactionID)), nil)
			return nil
		}
	} else {
		pvt.CurrentPunishVote.DisagreedPlayerIDs[playerID] = true
		// check result
		if len(pvt.CurrentPunishVote.DisagreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 && pvt.VerifyResult(isAgreed) {
			err := pvt.VoteFailed()
			if err != nil {
				gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msgf("Failed to process failed vote due to %s", err.Error())
				return terror.Error(err, "Failed to process the result")
			}

			// broadcast undefined to clean up the form in the frontend
			pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, pvt.FactionID)), nil)
			return nil
		}
	}

	// broadcast punish vote result
	pvt.broadcastResult <- &PunishVoteResult{
		PunishVoteID:          pvt.CurrentPunishVote.ID,
		TotalPlayerNumber:     len(pvt.CurrentPunishVote.PlayerPool),
		AgreedPlayerNumber:    len(pvt.CurrentPunishVote.AgreedPlayerIDs),
		DisagreedPlayerNumber: len(pvt.CurrentPunishVote.DisagreedPlayerIDs),
	}

	return nil
}

// VerifyResult return true if the latest result is correct
func (pvt *PunishVoteTracker) VerifyResult(checkAgree bool) bool {
	// get the latest player pools
	pvt.CurrentPunishVote.PlayerPool = pvt.CurrentEligiblePlayers()

	// check on agreed result
	if checkAgree {
		if len(pvt.CurrentPunishVote.AgreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 {
			return true
		}
		return false
	}

	// otherwise, check on disagreed result
	if len(pvt.CurrentPunishVote.DisagreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 {
		return true
	}
	return false
}

// VotePassed punish player when the vote is passed
func (pvt *PunishVoteTracker) VotePassed() error {
	now := time.Now()

	// switch stage to hold
	pvt.Stage.Phase = PunishVotePhaseHold
	pvt.Stage.EndTime = now.AddDate(1, 0, 0)

	// process the punishing action
	punishVote, err := boiler.FindPunishVote(gamedb.StdConn, pvt.CurrentPunishVote.ID)
	if err != nil {
		gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get punish vote from db")
		return terror.Error(err, "Failed to get punish vote from db")
	}

	punishVote.EndedAt = null.TimeFrom(now)
	punishVote.Status = string(PunishVoteStatusPassed)
	_, err = punishVote.Update(gamedb.StdConn, boil.Whitelist(boiler.PunishVoteColumns.EndedAt, boiler.PunishVoteColumns.Status))
	if err != nil {
		gamelog.L.Error().
			Str("punish vote id", pvt.CurrentPunishVote.ID).
			Str("finalise status", punishVote.Status).
			Str("punish vote end time", punishVote.EndedAt.Time.String()).
			Err(err).Msg("Failed to finalise current punish vote")
		return terror.Error(err, "Failed to finalise current punish vote")
	}

	// get punish type
	punishOption, err := boiler.FindPunishOption(gamedb.StdConn, punishVote.PunishOptionID)
	if err != nil {
		gamelog.L.Error().
			Str("punish type id", punishVote.PunishOptionID).
			Err(err).Msg("Failed to get punish type from db")
		return terror.Error(err, "Failed to get punish type from db")
	}

	// punish user
	bp := &boiler.PunishedPlayer{
		PlayerID:            punishVote.ReportedPlayerID,
		PunishOptionID:      punishOption.ID,
		PunishUntil:         time.Now().Add(time.Duration(punishOption.PunishDurationHours) * time.Hour),
		RelatedPunishVoteID: null.StringFrom(punishVote.ID),
	}
	err = bp.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("punish player", bp).
			Err(err).Msg("Failed to insert player into punish list")
		return terror.Error(err, "Failed to insert player into punish list")
	}

	// TODO: broadcast success punish notification on chat
	pvt.BroadcastPunishVoteResult(true)
	return nil
}

// VoteFailed process when vote failed
func (pvt *PunishVoteTracker) VoteFailed() error {
	now := time.Now()

	// switch stage to hold
	pvt.Stage.Phase = PunishVotePhaseHold
	pvt.Stage.EndTime = now.AddDate(1, 0, 0)

	// process the punishing action
	punishVote, err := boiler.FindPunishVote(gamedb.StdConn, pvt.CurrentPunishVote.ID)
	if err != nil {
		gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get punish vote from db")
		return terror.Error(err, "Failed to get punish vote from db")
	}

	punishVote.EndedAt = null.TimeFrom(now)
	punishVote.Status = string(PunishVoteStatusFailed)

	_, err = punishVote.Update(gamedb.StdConn, boil.Whitelist(boiler.PunishVoteColumns.EndedAt, boiler.PunishVoteColumns.Status))
	if err != nil {
		gamelog.L.Error().
			Str("punish vote id", pvt.CurrentPunishVote.ID).
			Str("finalise status", punishVote.Status).
			Str("punish vote end time", punishVote.EndedAt.Time.String()).
			Err(err).Msg("Failed to finalise current punish vote")
		return terror.Error(err, "Failed to finalise current punish vote")
	}

	// increase reported player fee
	reportedPlayer, err := boiler.FindPlayer(gamedb.StdConn, punishVote.ReportedPlayerID)
	if err != nil {
		gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to get reported player from db")
		return terror.Error(err, "Failed to get reported player from db")
	}

	reportedPlayer.ReportedCost = reportedPlayer.ReportedCost.Mul(decimal.NewFromInt(2))

	_, err = reportedPlayer.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.ReportedCost))
	if err != nil {
		gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to update report cost of the player")
		return terror.Error(err, "Failed to update report cost of the player")
	}

	// TODO: broadcast failed punish result notification on chat
	pvt.BroadcastPunishVoteResult(false)
	return nil
}

func (pvt *PunishVoteTracker) debounceBroadcastResult() {
	var result *PunishVoteResult

	interval := 500 * time.Millisecond
	timer := time.NewTimer(interval)

	for {
		select {
		case result = <-pvt.broadcastResult:
			timer.Reset(interval)
		case <-timer.C:
			if result != nil {
				pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteResultSubscribe, pvt.FactionID)), result)
			}
		}
	}
}

func (pvt *PunishVoteTracker) BroadcastPunishVoteResult(isPassed bool) {
	// get punish vote
	punishVote, err := boiler.FindPunishVote(gamedb.StdConn, pvt.CurrentPunishVote.ID)
	if err != nil {
		gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get current punish vote from db")
		return
	}

	punishOption, err := punishVote.PunishOption().One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get punish option from punish vote")
		return
	}

	// construct punish vote message
	chatMessage := &ChatMessage{
		Type:   ChatMessageTypePunishVote,
		SentAt: time.Now(),
		Data: MessagePunishVote{
			IssuedByPlayerID:        punishVote.IssuedByID,
			IssuedByPlayerUsername:  punishVote.IssuedByUsername,
			IssuedByPlayerFactionID: punishVote.FactionID,
			IssuedByPlayerGid:       punishVote.IssuedByGid,

			ReportedPlayerID:        punishVote.ReportedPlayerID,
			ReportedPlayerUsername:  punishVote.ReportedPlayerUsername,
			ReportedPlayerGid:       punishVote.ReportedPlayerGid,
			ReportedPlayerFactionID: punishVote.FactionID,

			// vote result
			IsPassed:              isPassed,
			TotalPlayerNumber:     len(pvt.CurrentPunishVote.PlayerPool),
			AgreedPlayerNumber:    len(pvt.CurrentPunishVote.AgreedPlayerIDs),
			DisagreedPlayerNumber: len(pvt.CurrentPunishVote.DisagreedPlayerIDs),
			PunishOption:          *punishOption,
			PunishReason:          punishVote.Reason,
		},
	}

	// store message to the chat
	pvt.api.AddFactionChatMessage(pvt.FactionID, chatMessage)

	// broadcast
	pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, pvt.FactionID)), chatMessage)
}
