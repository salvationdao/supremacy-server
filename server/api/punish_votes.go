package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/ninja-software/tickle"

	"github.com/gofrs/uuid"

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

func (api *API) PunishVoteTrackerSetup() error {
	// get factions
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup faction punish vote tracker")
		return terror.Error(err, "Failed to setup faction punish vote tracker")
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

	// create a tickle to update vote price every 24 hours
	playerPunishVoteCostUpdater := tickle.New("Player Punish Vote Cost updater", 24*60*60, func() (int, error) {
		err = db.UpdatePunishVoteCost()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to update player punish vote cost and report cost")
			return http.StatusInternalServerError, terror.Error(err)
		}
		return http.StatusOK, nil
	})
	playerPunishVoteCostUpdater.Log = gamelog.L

	err = playerPunishVoteCostUpdater.SetIntervalAt(24*time.Hour, 1, 0)
	if err != nil {
		return terror.Error(err, "Failed to setup player punish vote cost updater")
	}

	return nil
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
		uss, err := boiler.PlayerStats(
			boiler.PlayerStatWhere.ID.IN(dbSearchList),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get player stat from db")
		}

		secondCheckList := []string{}
		for _, player := range uss {
			// add player list to second check list
			if player.AbilityKillCount < 100 {
				secondCheckList = append(secondCheckList, player.ID)
				continue
			}
			// player is eligible to vote if they have more than 100 kills in lifetime
			result[player.ID] = true
		}

		if len(secondCheckList) > 0 {
			// check last 7 days kills count
			paks, err := boiler.PlayerKillLogs(
				boiler.PlayerKillLogWhere.PlayerID.IN(secondCheckList),
				boiler.PlayerKillLogWhere.CreatedAt.GT(time.Now().AddDate(0, 0, -7)),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msg("Failed to get player kill count from db")
			}

			if paks != nil && len(paks) > 0 {
				for _, playerID := range secondCheckList {
					killCount := 0
					for _, pak := range paks {
						if playerID != pak.PlayerID {
							continue
						}
						if !pak.IsTeamKill {
							killCount++
							continue
						}
						killCount--
					}

					// player is eligible to vote if they have more than 5 kills in last 7 days
					if killCount >= 5 {
						result[playerID] = true
					}
				}
			}
		}

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

	// vote passed, if the amount of the agreed players pass 50%
	if len(pvt.CurrentPunishVote.AgreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 {
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
		if len(pvt.CurrentPunishVote.AgreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 {
			err := pvt.VotePassed()
			if err != nil {
				gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msgf("Failed to process failed vote due to %s", err.Error())
				return terror.Error(err, "Failed to process the result")
			}
			return nil
		}
	} else {
		pvt.CurrentPunishVote.DisagreedPlayerIDs[playerID] = true
		// check result
		if len(pvt.CurrentPunishVote.DisagreedPlayerIDs) > len(pvt.CurrentPunishVote.PlayerPool)/2 {
			err := pvt.VoteFailed()
			if err != nil {
				gamelog.L.Error().Str("punish vote id", pvt.CurrentPunishVote.ID).Err(err).Msgf("Failed to process failed vote due to %s", err.Error())
				return terror.Error(err, "Failed to process the result")
			}
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

	punishDuration := time.Now().Add(time.Duration(punishOption.PunishDurationHours) * time.Hour)

	if pvt.api.Config.Address == "staging" || pvt.api.Config.Address == "development" {
		punishDuration = time.Now().Add(time.Duration(5) * time.Minute)
	}

	// punish user
	bp := &boiler.PunishedPlayer{
		PlayerID:            punishVote.ReportedPlayerID,
		PunishOptionID:      punishOption.ID,
		PunishUntil:         punishDuration,
		RelatedPunishVoteID: null.StringFrom(punishVote.ID),
	}
	err = bp.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("punish player", bp).
			Err(err).Msg("Failed to insert player into punish list")
		return terror.Error(err, "Failed to insert player into punish list")
	}

	// broadcast success punish notification on chat
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

	// broadcast failed punish result notification on chat
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

	// broadcast undefined to clean up the form in the frontend
	pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, pvt.FactionID)), nil)

	// construct punish vote message
	chatMessage := &ChatMessage{
		Type:   ChatMessageTypePunishVote,
		SentAt: time.Now(),
		Data: MessagePunishVote{
			IssuedByUser: boiler.Player{
				ID:        punishVote.IssuedByID,
				Username:  null.StringFrom(punishVote.IssuedByUsername),
				FactionID: null.StringFrom(punishVote.FactionID),
				Gid:       punishVote.IssuedByGid,
			},
			ReportedUser: boiler.Player{
				ID:        punishVote.ReportedPlayerID,
				Username:  null.StringFrom(punishVote.ReportedPlayerUsername),
				FactionID: null.StringFrom(punishVote.FactionID),
				Gid:       punishVote.ReportedPlayerGid,
			},
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

	if isPassed {
		// get current player's punishment
		punishments, err := boiler.PunishedPlayers(
			boiler.PunishedPlayerWhere.PlayerID.EQ(punishVote.ReportedPlayerID),
			boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
			qm.Load(boiler.PunishedPlayerRels.PunishOption),
			qm.Load(boiler.PunishedPlayerRels.RelatedPunishVote),
		).All(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to get player's punishment from db")
			return
		}

		if punishments == nil || len(punishments) == 0 {
			return
		}

		playerPunishments := []*PlayerPunishment{}
		for _, punishment := range punishments {
			playerPunishments = append(playerPunishments, &PlayerPunishment{
				PunishedPlayer:    punishment,
				RelatedPunishVote: punishment.R.RelatedPunishVote,
				PunishOption:      punishment.R.PunishOption,
			})
		}

		punishedPlayerID := uuid.FromStringOrNil(punishVote.ReportedPlayerID)

		// send to the player
		pvt.api.BattleArena.SendToOnlinePlayer(punishedPlayerID, HubKeyPlayerPunishmentList, playerPunishments)
	}
}
