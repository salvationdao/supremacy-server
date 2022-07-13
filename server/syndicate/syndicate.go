package syndicate

import (
	"database/sql"
	"errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
)

type Syndicate struct {
	system *System
	*boiler.Syndicate
	sync.RWMutex // for update syndicate

	motionSystem *MotionSystem
}

func (s *Syndicate) liquidate() error {
	s.Lock()
	defer s.Unlock()

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction")
		return terror.Error(err, "Failed to remove syndicate")
	}

	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// delete all directors from tables
	_, err = boiler.SyndicateDirectors(boiler.SyndicateDirectorWhere.SyndicateID.EQ(s.ID)).DeleteAll(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to players' director syndicate id")
		return terror.Error(err, "Failed to remove syndicate")
	}

	// delete all committees from table
	_, err = boiler.SyndicateCommittees(boiler.SyndicateCommitteeWhere.SyndicateID.EQ(s.ID)).DeleteAll(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to players' director syndicate id")
		return terror.Error(err, "Failed to remove syndicate")
	}

	// set all member player to null
	_, err = boiler.Players(
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(s.ID)),
	).UpdateAll(tx, boiler.M{boiler.PlayerColumns.SyndicateID: null.StringFromPtr(nil)})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to players' syndicate id")
		return terror.Error(err, "Failed to remove syndicate")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return terror.Error(err, "Failed to remove syndicate")
	}

	// stop all the motion in the syndicate
	s.motionSystem.terminated()

	return nil
}

func (s *Syndicate) isDirector(userID string) (bool, error) {
	// check availability
	exist, err := boiler.SyndicateDirectors(
		boiler.SyndicateDirectorWhere.SyndicateID.EQ(s.ID),
		boiler.SyndicateDirectorWhere.PlayerID.EQ(userID),
	).Exists(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Str("player id", userID).Msg("Failed to check syndicate director list from db.")
		return false, terror.Error(err, "Failed to submit new motion to syndicate")
	}

	return exist, nil
}
