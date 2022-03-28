package api

import (
	"context"
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

type BanVotePhase string

const (
	BanVotePhaseVoting BanVotePhase = "VOTING"
	BanVotePhaseHold   BanVotePhase = "HOLD"
)

type BanVoteStage struct {
	Phase   BanVotePhase
	EndTime time.Time
}

type BanVoteInstance struct {
	*boiler.BanVote
	BanType *boiler.BanType `json:"ban_type"`
}

type BanVoteTracker struct {
	FactionID string

	// ban vote tracker
	BanVoteID string
	StartedAt time.Time
	EndedAt   time.Time
	Stage     *BanVoteStage

	// receive vote from player
	VoteChan           chan *BanVote
	AgreedPlayerIDs    map[string]bool
	DisagreedPlayerIDs map[string]bool

	// mutex lock for issue vote
	deadlock.Mutex

	// message bus
	MessageBus *messagebus.MessageBus

	// broadcast result
	broadcastResult chan *BanVoteResult
}

type BanVote struct {
	BanVoteID string
	playerID  string
	IsAgreed  bool
}

func (api *API) BanVoteTrackerSetup() {
	// get factions
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup faction ban vote tracker")
		return
	}

	for _, f := range factions {
		// initialise
		bv := &BanVoteTracker{
			FactionID:          f.ID,
			MessageBus:         api.MessageBus,
			AgreedPlayerIDs:    make(map[string]bool),
			DisagreedPlayerIDs: make(map[string]bool),
			broadcastResult:    make(chan *BanVoteResult),
		}

		// store ban vote instance of each faction
		api.FactionBanVote[f.ID] = bv

		// run debounce broadcast ban vote result
		go bv.debounceBroadcastResult()

		// start ban vote tracker
		go bv.Run()
	}
}

func (bv *BanVoteTracker) Run() {
	mainTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-mainTicker.C:
			switch bv.Stage.Phase {
			case BanVotePhaseVoting:
				// skip, if voting still going on
				if bv.EndedAt.After(time.Now()) {
					continue
				}

				now := time.Now()

				// switch stage to hold (block incoming vote from players)
				bv.Stage.Phase = BanVotePhaseHold
				bv.Stage.EndTime = now.AddDate(1, 0, 0)

				// process the banning action
				banVote, err := boiler.FindBanVote(gamedb.StdConn, bv.BanVoteID)
				if err != nil {
					gamelog.L.Error().Str("ban vote id", bv.BanVoteID).Err(err).Msg("Failed to get ban vote from db")
					return
				}

				// get all the agreed/disagreed count from db
				playerBanVotes, err := boiler.PlayersBanVotes(
					boiler.PlayersBanVoteWhere.BanVoteID.EQ(bv.BanVoteID),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("ban vote id", bv.BanVoteID).Err(err).Msg("Failed to get player ban vote from db")
					return
				}

				// ban is failed, finalize current ban vote
				if len(playerBanVotes) == 0 {
					banVote.EndedAt = null.TimeFrom(time.Now())
					banVote.Status = BanVoteStatusFailed

					_, err = banVote.Update(gamedb.StdConn, boil.Whitelist(boiler.BanVoteColumns.EndedAt, boiler.BanVoteColumns.Status))
					if err != nil {
						gamelog.L.Error().
							Str("ban vote id", bv.BanVoteID).
							Str("finalise status", banVote.Status).
							Str("ban vote end time", banVote.EndedAt.Time.String()).
							Err(err).Msg("Failed to finalise current ban vote")
						return
					}

					// increase reported player fee
					reportedPlayer, err := boiler.FindPlayer(gamedb.StdConn, banVote.ReportedPlayerID)
					if err != nil {
						gamelog.L.Error().Str("player id", banVote.ReportedPlayerID).Err(err).Msg("Failed to get reported player from db")
						return
					}

					reportedPlayer.ReportedCost = reportedPlayer.ReportedCost.Mul(decimal.NewFromInt(2))

					_, err = reportedPlayer.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.ReportedCost))
					if err != nil {
						gamelog.L.Error().Str("player id", banVote.ReportedPlayerID).Err(err).Msg("Failed to update report cost of the player")
						return
					}

					// TODO: broadcast failed ban result notification  on chat
					continue
				}

				// calculate result
				agreedCount := 0
				disagreedCount := 0
				for _, pbv := range playerBanVotes {
					if pbv.IsAgreed {
						agreedCount += 1
						continue
					}
					disagreedCount += 1
				}

				// ban success
				if agreedCount >= disagreedCount {
					// pass ban
					banVote.EndedAt = null.TimeFrom(time.Now())
					banVote.Status = BanVoteStatusPassed
					_, err = banVote.Update(gamedb.StdConn, boil.Whitelist(boiler.BanVoteColumns.EndedAt, boiler.BanVoteColumns.Status))
					if err != nil {
						gamelog.L.Error().
							Str("ban vote id", bv.BanVoteID).
							Str("finalise status", banVote.Status).
							Str("ban vote end time", banVote.EndedAt.Time.String()).
							Err(err).Msg("Failed to finalise current ban vote")
						return
					}

					// get ban type
					banType, err := boiler.FindBanType(gamedb.StdConn, banVote.BanTypeID)
					if err != nil {
						gamelog.L.Error().
							Str("ban type id", banVote.BanTypeID).
							Err(err).Msg("Failed to get ban type from db")
						return
					}

					// ban user
					bp := &boiler.BannedPlayer{
						PlayerID:         banVote.ReportedPlayerID,
						BanTypeID:        banType.ID,
						BanUntil:         time.Now().Add(time.Duration(banType.BanForHours) * time.Hour),
						RelatedBanVoteID: null.StringFrom(banVote.ID),
					}
					err = bp.Insert(gamedb.StdConn, boil.Infer())
					if err != nil {
						gamelog.L.Error().
							Interface("ban player", bp).
							Err(err).Msg("Failed to insert player into ban list")
						return
					}

					// TODO: broadcast success ban notification on chat

				} else {
					// failed ban
					banVote.EndedAt = null.TimeFrom(time.Now())
					banVote.Status = BanVoteStatusFailed

					_, err = banVote.Update(gamedb.StdConn, boil.Whitelist(boiler.BanVoteColumns.EndedAt, boiler.BanVoteColumns.Status))
					if err != nil {
						gamelog.L.Error().
							Str("ban vote id", bv.BanVoteID).
							Str("finalise status", banVote.Status).
							Str("ban vote end time", banVote.EndedAt.Time.String()).
							Err(err).Msg("Failed to finalise current ban vote")
						return
					}

					// increase reported player fee
					reportedPlayer, err := boiler.FindPlayer(gamedb.StdConn, banVote.ReportedPlayerID)
					if err != nil {
						gamelog.L.Error().Str("player id", banVote.ReportedPlayerID).Err(err).Msg("Failed to get reported player from db")
						return
					}

					reportedPlayer.ReportedCost = reportedPlayer.ReportedCost.Mul(decimal.NewFromInt(2))

					_, err = reportedPlayer.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.ReportedCost))
					if err != nil {
						gamelog.L.Error().Str("player id", banVote.ReportedPlayerID).Err(err).Msg("Failed to update report cost of the player")
						return
					}

					// TODO: broadcast failed ban notification on chat
				}

			case BanVotePhaseHold:
				// check whether there is another ban issue in db
				banVote, err := boiler.BanVotes(
					boiler.BanVoteWhere.FactionID.EQ(bv.FactionID),
					boiler.BanVoteWhere.Status.EQ(BanVoteStatusPending),
					qm.OrderBy(boiler.BanVoteColumns.CreatedAt),
					qm.Load(boiler.BanVoteRels.BanType),
				).One(gamedb.StdConn)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					gamelog.L.Error().Str("faction id", bv.FactionID).Err(err).Msg("Failed to load new ban vote from db")
					return
				}

				// skip, if there is no ban vote
				if banVote == nil {
					continue
				}

				// clean up map, before setup
				for key := range bv.AgreedPlayerIDs {
					delete(bv.AgreedPlayerIDs, key)
				}
				for key := range bv.DisagreedPlayerIDs {
					delete(bv.DisagreedPlayerIDs, key)
				}

				now := time.Now()
				endTime := now.Add(20 * time.Second)

				// update current ban vote, start/end time
				banVote.StartedAt = null.TimeFrom(now)
				banVote.EndedAt = null.TimeFrom(endTime)
				_, err = banVote.Update(gamedb.StdConn, boil.Whitelist(boiler.BanVoteColumns.StartedAt, boiler.BanVoteColumns.EndedAt))
				if err != nil {
					gamelog.L.Error().Str("ban vote id", banVote.ID).Err(err).Msg("Failed to update the start time of the ban vote")
					return
				}

				// otherwise, set up the detail of ban vote
				bv.BanVoteID = banVote.ID
				bv.StartedAt = time.Now()

				// change stage
				bv.Stage.Phase = BanVotePhaseVoting
				bv.Stage.EndTime = endTime

				// broadcast initial result
				bv.broadcastResult <- &BanVoteResult{
					BanVoteID:             bv.BanVoteID,
					AgreedPlayerNumber:    0,
					DisagreedPlayerNumber: 0,
				}

				// broadcast new vote to online faction users
				bv.MessageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBanVoteSubscribe, bv.FactionID)), &BanVoteInstance{
					BanVote: banVote,
					BanType: banVote.R.BanType,
				})
			}

		case playerVote := <-bv.VoteChan:
			// check voting phase and targeted vote is available
			if bv.Stage.Phase != BanVotePhaseVoting || bv.Stage.EndTime.Before(time.Now()) || bv.BanVoteID != playerVote.BanVoteID {
				continue
			}

			// check player has voted
			pbv, err := boiler.PlayersBanVotes(
				boiler.PlayersBanVoteWhere.BanVoteID.EQ(bv.BanVoteID),
				boiler.PlayersBanVoteWhere.PlayerID.EQ(playerVote.playerID),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Str("ban_vote_id", bv.BanVoteID).Str("player_id", playerVote.playerID).Err(err).Msg("Failed to get player ban vote from db")
				continue
			}

			// skip, if player has already voted
			if pbv != nil {
				continue
			}

			// store player vote result into database
			pbv = &boiler.PlayersBanVote{
				BanVoteID: bv.BanVoteID,
				PlayerID:  playerVote.playerID,
				IsAgreed:  playerVote.IsAgreed,
			}
			err = pbv.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("ban_vote_id", bv.BanVoteID).Str("player_id", playerVote.playerID).Err(err).Msg("Failed to insert player vote result into db")
				continue
			}

			// update result
			if pbv.IsAgreed {
				bv.AgreedPlayerIDs[playerVote.playerID] = true
			} else {
				bv.DisagreedPlayerIDs[playerVote.playerID] = true
			}

			// broadcast ban vote result
			bv.broadcastResult <- &BanVoteResult{
				BanVoteID:             bv.BanVoteID,
				AgreedPlayerNumber:    len(bv.AgreedPlayerIDs),
				DisagreedPlayerNumber: len(bv.DisagreedPlayerIDs),
			}
		}
	}
}

func (bv *BanVoteTracker) debounceBroadcastResult() {
	var result *BanVoteResult

	interval := 500 * time.Millisecond
	timer := time.NewTimer(interval)

	for {
		select {
		case result = <-bv.broadcastResult:
			timer.Reset(interval)
		case <-timer.C:
			if result != nil {
				bv.MessageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBanVoteResultSubscribe, bv.FactionID)), result)
			}
		}
	}
}
