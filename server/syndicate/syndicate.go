package syndicate

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type Syndicate struct {
	system *System
	*boiler.Syndicate
	sync.RWMutex // for update syndicate

	isLiquidated atomic.Bool

	motionSystem     *MotionSystem
	accountantSystem *AccountSystem
}

func newSyndicate(ss *System, syndicate *boiler.Syndicate) (*Syndicate, error) {
	s := &Syndicate{
		system:    ss,
		Syndicate: syndicate,
	}
	motionSystem, err := newMotionSystem(s)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate motion system.")
		return nil, terror.Error(err, "Failed to spin up syndicate motion system.")
	}

	s.motionSystem = motionSystem
	s.accountantSystem = NewAccountSystem(s)

	return s, nil
}

func (s *Syndicate) liquidate(tx *sql.Tx) error {
	// lock the syndicate
	s.Lock()
	defer s.Unlock()

	s.isLiquidated.Store(true)

	// stop all the ongoing motion in the syndicate
	s.motionSystem.terminate()

	// liquidate fund
	err := s.accountantSystem.liquidate()
	if err != nil {
		return err
	}

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

	// get all the players
	ps, err := boiler.Players(
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(s.ID)),
	).All(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to players' syndicate id")
		return terror.Error(err, "Failed to remove syndicate")
	}

	_, err = ps.UpdateAll(tx, boiler.M{boiler.PlayerColumns.SyndicateID: null.StringFromPtr(nil)})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to remove player syndicate id.")
		return terror.Error(err, "Failed to remove syndicate")
	}

	for _, p := range ps {
		p.SyndicateID = null.StringFromPtr(nil)
		ws.PublishMessage(fmt.Sprintf("/user/%s", p.ID), server.HubKeyUserSubscribe, p)
	}

	// archive syndicate
	s.DeletedAt = null.TimeFrom(time.Now())
	_, err = s.Update(tx, boil.Whitelist(boiler.SyndicateColumns.DeletedAt))
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to archive syndicate.")
		return terror.Error(err, "Failed to liquidate syndicate.")
	}

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

// getTotalAvailableMotionVoter return total of available motion voter base on syndicate type
func (s *Syndicate) getTotalAvailableMotionVoter() (int64, error) {
	var total int64
	var err error

	switch s.Type {
	case boiler.SyndicateTypeCORPORATION:
		total, err = s.SyndicateDirectors().Count(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to get syndicate director number")
			return 0, terror.Error(err, "Failed to get syndicate directors number")
		}

		return total, nil
	case boiler.SyndicateTypeDECENTRALISED:
		total, err = s.Players().Count(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to get syndicate members number")
			return 0, terror.Error(err, "Failed to get syndicate members number")
		}

		return total, nil

	default:
		gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Str("syndicate type", s.Type).Msg("Failed to get total available motion voters")
		return 0, terror.Error(fmt.Errorf("invalid syndicate type"), "Invalid syndicate type")
	}
}
