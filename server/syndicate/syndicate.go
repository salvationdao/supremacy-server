package syndicate

import (
	"database/sql"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

type Syndicate struct {
	system *System
	*boiler.Syndicate
	deadlock.RWMutex // for update syndicate

	isLiquidated atomic.Bool

	motionSystem   *MotionSystem
	accountSystem  *AccountSystem
	recruitSystem  *RecruitSystem
	electionSystem *ElectionSystem
}

func newSyndicate(ss *System, syndicate *boiler.Syndicate) (*Syndicate, error) {
	var err error
	s := &Syndicate{
		system:    ss,
		Syndicate: syndicate,
	}
	s.motionSystem, err = newMotionSystem(s)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate motion system.")
		return nil, err
	}

	s.recruitSystem, err = newRecruitSystem(s)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate recruit system.")
		return nil, err
	}

	s.electionSystem, err = newElectionSystem(s)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate election system.")
		return nil, err
	}

	s.accountSystem = newAccountSystem(s)

	return s, nil
}

func (s *Syndicate) liquidate(tx *sql.Tx) error {
	// lock the syndicate
	s.Lock()
	defer s.Unlock()

	s.isLiquidated.Store(true)

	// stop all the ongoing motion in the syndicate
	s.motionSystem.terminate()

	// stop all the ongoing join application
	s.recruitSystem.terminate()

	// stop election system
	s.electionSystem.isClosed.Store(true)

	// liquidate fund
	err := s.accountSystem.liquidate()
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

		err = p.L.LoadRole(gamedb.StdConn, true, p, nil)
		if err != nil {
			gamelog.L.Error().Str("player id", p.ID).Err(err).Msg("Failed to load role_id")
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s", p.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(p))
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
