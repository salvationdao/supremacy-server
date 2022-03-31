package api

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

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

type PunishVoteInstance struct {
	*boiler.PunishVote
	PunishOption *boiler.PunishOption `json:"punish_option"`
}

type PunishVoteTracker struct {
	FactionID string

	// punish vote tracker
	PunishVoteID string
	StartedAt    time.Time
	EndedAt      time.Time
	Stage        *PunishVoteStage

	// receive vote from player
	VoteChan           chan *PunishVote
	AgreedPlayerIDs    map[string]bool
	DisagreedPlayerIDs map[string]bool

	// mutex lock for issue vote
	deadlock.Mutex

	// message bus
	MessageBus *messagebus.MessageBus

	// broadcast result
	broadcastResult chan *PunishVoteResult
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
			FactionID:          f.ID,
			MessageBus:         api.MessageBus,
			AgreedPlayerIDs:    make(map[string]bool),
			DisagreedPlayerIDs: make(map[string]bool),
			broadcastResult:    make(chan *PunishVoteResult),
			Stage:              &PunishVoteStage{PunishVotePhaseHold, time.Now().AddDate(1, 0, 0)},
			VoteChan:           make(chan *PunishVote),
		}

		// store punish vote instance of each faction
		api.FactionPunishVote[f.ID] = pvt

		// start punish vote tracker
		go pvt.Run()
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
				now := time.Now()

				// skip, if voting still going on
				if pvt.Stage.EndTime.After(now) {
					continue
				}

				// switch stage to hold (block incoming vote from players)
				pvt.Stage.Phase = PunishVotePhaseHold
				pvt.Stage.EndTime = now.AddDate(1, 0, 0)

				// process the punishning action
				punishVote, err := boiler.FindPunishVote(gamedb.StdConn, pvt.PunishVoteID)
				if err != nil {
					gamelog.L.Error().Str("punish vote id", pvt.PunishVoteID).Err(err).Msg("Failed to get punish vote from db")
					return
				}

				// get all the agreed/disagreed count from db
				playerPunishVotes, err := boiler.PlayersPunishVotes(
					boiler.PlayersPunishVoteWhere.PunishVoteID.EQ(pvt.PunishVoteID),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("punish vote id", pvt.PunishVoteID).Err(err).Msg("Failed to get player punish vote from db")
					return
				}

				// punish is failed, finalize current punish vote
				if len(playerPunishVotes) == 0 {
					punishVote.EndedAt = null.TimeFrom(time.Now())
					punishVote.Status = string(PunishVoteStatusFailed)

					_, err = punishVote.Update(gamedb.StdConn, boil.Whitelist(boiler.PunishVoteColumns.EndedAt, boiler.PunishVoteColumns.Status))
					if err != nil {
						gamelog.L.Error().
							Str("punish vote id", pvt.PunishVoteID).
							Str("finalise status", punishVote.Status).
							Str("punish vote end time", punishVote.EndedAt.Time.String()).
							Err(err).Msg("Failed to finalise current punish vote")
						return
					}

					// increase reported player fee
					reportedPlayer, err := boiler.FindPlayer(gamedb.StdConn, punishVote.ReportedPlayerID)
					if err != nil {
						gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to get reported player from db")
						return
					}

					reportedPlayer.ReportedCost = reportedPlayer.ReportedCost.Mul(decimal.NewFromInt(2))

					_, err = reportedPlayer.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.ReportedCost))
					if err != nil {
						gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to update report cost of the player")
						return
					}

					// TODO: broadcast failed punish result notification  on chat
					continue
				}

				// calculate result
				agreedCount := 0
				disagreedCount := 0
				for _, pbv := range playerPunishVotes {
					if pbv.IsAgreed {
						agreedCount += 1
						continue
					}
					disagreedCount += 1
				}

				// punish success
				if agreedCount >= disagreedCount {
					// pass punish
					punishVote.EndedAt = null.TimeFrom(time.Now())
					punishVote.Status = string(PunishVoteStatusPassed)
					_, err = punishVote.Update(gamedb.StdConn, boil.Whitelist(boiler.PunishVoteColumns.EndedAt, boiler.PunishVoteColumns.Status))
					if err != nil {
						gamelog.L.Error().
							Str("punish vote id", pvt.PunishVoteID).
							Str("finalise status", punishVote.Status).
							Str("punish vote end time", punishVote.EndedAt.Time.String()).
							Err(err).Msg("Failed to finalise current punish vote")
						return
					}

					// get punish type
					punishOption, err := boiler.FindPunishOption(gamedb.StdConn, punishVote.PunishOptionID)
					if err != nil {
						gamelog.L.Error().
							Str("punish type id", punishVote.PunishOptionID).
							Err(err).Msg("Failed to get punish type from db")
						return
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
						return
					}

					// TODO: broadcast success punish notification on chat

				} else {
					// failed punish
					punishVote.EndedAt = null.TimeFrom(time.Now())
					punishVote.Status = string(PunishVoteStatusFailed)

					_, err = punishVote.Update(gamedb.StdConn, boil.Whitelist(boiler.PunishVoteColumns.EndedAt, boiler.PunishVoteColumns.Status))
					if err != nil {
						gamelog.L.Error().
							Str("punish vote id", pvt.PunishVoteID).
							Str("finalise status", punishVote.Status).
							Str("punish vote end time", punishVote.EndedAt.Time.String()).
							Err(err).Msg("Failed to finalise current punish vote")
						return
					}

					// increase reported player fee
					reportedPlayer, err := boiler.FindPlayer(gamedb.StdConn, punishVote.ReportedPlayerID)
					if err != nil {
						gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to get reported player from db")
						return
					}

					reportedPlayer.ReportedCost = reportedPlayer.ReportedCost.Mul(decimal.NewFromInt(2))

					_, err = reportedPlayer.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.ReportedCost))
					if err != nil {
						gamelog.L.Error().Str("player id", punishVote.ReportedPlayerID).Err(err).Msg("Failed to update report cost of the player")
						return
					}

					// TODO: broadcast failed punish notification on chat
				}

			case PunishVotePhaseHold:
				// check whether there is another punish issue in db
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
					continue
				}

				// clean up map, before setup
				for key := range pvt.AgreedPlayerIDs {
					delete(pvt.AgreedPlayerIDs, key)
				}
				for key := range pvt.DisagreedPlayerIDs {
					delete(pvt.DisagreedPlayerIDs, key)
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

				// otherwise, set up the detail of punish vote
				pvt.PunishVoteID = punishVote.ID
				pvt.StartedAt = now

				// change stage
				pvt.Stage.Phase = PunishVotePhaseVoting
				pvt.Stage.EndTime = endTime

				// broadcast initial result
				pvt.broadcastResult <- &PunishVoteResult{
					PunishVoteID:          pvt.PunishVoteID,
					AgreedPlayerNumber:    0,
					DisagreedPlayerNumber: 0,
				}

				// broadcast new vote to online faction users
				pvt.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, pvt.FactionID)), &PunishVoteInstance{
					PunishVote:   punishVote,
					PunishOption: punishVote.R.PunishOption,
				})
			}

		case playerVote := <-pvt.VoteChan:
			// check voting phase and targeted vote is available
			if pvt.Stage.Phase != PunishVotePhaseVoting || pvt.Stage.EndTime.Before(time.Now()) || pvt.PunishVoteID != playerVote.PunishVoteID {
				continue
			}

			// check player has voted
			pbv, err := boiler.PlayersPunishVotes(
				boiler.PlayersPunishVoteWhere.PunishVoteID.EQ(pvt.PunishVoteID),
				boiler.PlayersPunishVoteWhere.PlayerID.EQ(playerVote.playerID),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Str("punish_vote_id", pvt.PunishVoteID).Str("player_id", playerVote.playerID).Err(err).Msg("Failed to get player punish vote from db")
				continue
			}

			// skip, if player has already voted
			if pbv != nil {
				continue
			}

			// store player vote result into database
			pbv = &boiler.PlayersPunishVote{
				PunishVoteID: pvt.PunishVoteID,
				PlayerID:     playerVote.playerID,
				IsAgreed:     playerVote.IsAgreed,
			}
			err = pbv.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("punish_vote_id", pvt.PunishVoteID).Str("player_id", playerVote.playerID).Err(err).Msg("Failed to insert player vote result into db")
				continue
			}

			// update result
			if pbv.IsAgreed {
				pvt.AgreedPlayerIDs[playerVote.playerID] = true
			} else {
				pvt.DisagreedPlayerIDs[playerVote.playerID] = true
			}

			// broadcast punish vote result
			pvt.broadcastResult <- &PunishVoteResult{
				PunishVoteID:          pvt.PunishVoteID,
				AgreedPlayerNumber:    len(pvt.AgreedPlayerIDs),
				DisagreedPlayerNumber: len(pvt.DisagreedPlayerIDs),
			}
		}
	}
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
