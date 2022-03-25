package api

import (
	"database/sql"
	"errors"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BanVoteTracker struct {
	factionID string

	// active player tracker
	deadlock.RWMutex
	ActivePlayers map[string]bool

	// ban vote tracker
	deadlock.Mutex
	BanVoteID        string
	AgreedPlayers    map[string]bool
	DisagreedPlayers map[string]bool
	StartedAt        time.Time
	EndedAt          time.Time
}

func (api *API) BanVoteTrackerSetup() {
	// get factions
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup faction ban vote tracker")
		return
	}

	for _, f := range factions {
		api.factionBanVoteTracker[f.ID] = &BanVoteTracker{
			factionID:     f.ID,
			ActivePlayers: make(map[string]bool),

			AgreedPlayers:    make(map[string]bool),
			DisagreedPlayers: make(map[string]bool),
		}

		// get the latest ban vote from db
		banVote, err := boiler.BanVotes(
			boiler.BanVoteWhere.FactionID.EQ(f.ID),
			boiler.BanVoteWhere.Status.EQ(BanVoteStatusPending),
			qm.OrderBy(boiler.BanVoteColumns.CreatedAt),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("Unable to get ban vote from db")
			return
		}

		if banVote != nil {
			api.factionBanVoteTracker[f.ID].BanVoteID = banVote.ID
			api.factionBanVoteTracker[f.ID].StartedAt = time.Now()

			// if the ban vote is unfinalised
			if banVote.StartedAt.Valid {
				api.factionBanVoteTracker[f.ID].StartedAt = banVote.StartedAt.Time
			}

			api.factionBanVoteTracker[f.ID].EndedAt = time.Now().Add(30 * time.Second)

			// set timer for finalising the ban vote

			// broadcast to faction user
		}

	}
}

func (bvt *BanVoteTracker) Load() {

}

func (bvt *BanVoteTracker) Vote(isAgreed bool) {

}

func (bvt *BanVoteTracker) Finalise() {

}
