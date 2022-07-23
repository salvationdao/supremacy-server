package syndicate

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type ElectionSystem struct {
	syndicate       *Syndicate
	ongoingElection *boiler.SyndicateElection
	sync.RWMutex
}

func newElectionSystem(s *Syndicate) (*ElectionSystem, error) {
	es := &ElectionSystem{
		syndicate: s,
	}

	// load any unfinished election
	e, err := boiler.SyndicateElections(
		boiler.SyndicateElectionWhere.SyndicateID.EQ(s.ID),
		boiler.SyndicateElectionWhere.EndAt.GT(time.Now()),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to load syndicate election")
	}

	if e != nil {
		es.ongoingElection = e
	}

	return es, nil
}

func (es *ElectionSystem) heldElection() error {
	es.Lock()
	defer es.Unlock()

	// check election exists
	if es.ongoingElection != nil {
		return terror.Error(fmt.Errorf("incomplete election"), "There is an ongoing election.")
	}

	now := time.Now()
	// create an election base on syndicate type
	se := &boiler.SyndicateElection{
		SyndicateID:              es.syndicate.ID,
		StartedAt:                now,
		CandidateRegisterCloseAt: now.AddDate(0, 0, 1),
		EndAt:                    now.AddDate(0, 0, 3),
	}

	err := se.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("election", se).Msg("Failed to insert new syndicate election.")
		return terror.Error(err, "Failed to start a syndicate election.")
	}

	// set ongoing election
	es.ongoingElection = se

	// broadcast to all the members
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.syndicate.FactionID, es.syndicate.ID), server.HubKeySyndicateOngoingElectionSubscribe, se)

	// TODO: email all the syndicate members

	return nil
}

func (es *ElectionSystem) registerCandidate() error {
	es.Lock()
	defer es.Unlock()

	return nil
}

func (es *ElectionSystem) vote() error {
	es.Lock()
	defer es.Unlock()

	return nil
}
