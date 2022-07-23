package syndicate

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type ElectionSystem struct {
	syndicateID string
	factionID   string
	isClosed    atomic.Bool
	sync.RWMutex
}

func newElectionSystem(s *Syndicate) (*ElectionSystem, error) {
	es := &ElectionSystem{
		syndicateID: s.ID,
		factionID:   s.FactionID,
	}

	es.isClosed.Store(false)

	go es.start()

	return es, nil
}

type ElectionResult struct {
	CandidateID string `json:"candidate_id" db:"voted_for_candidate_id"`
	TotalVotes  int64  `json:"total_votes" db:"total_votes"`
}

func (es *ElectionSystem) start() {
	for {
		// check every 10 seconds
		time.Sleep(10 * time.Second)

		now := time.Now()
		se, err := boiler.SyndicateElections(
			boiler.SyndicateElectionWhere.SyndicateID.EQ(es.syndicateID),
			boiler.SyndicateElectionWhere.FinalisedAt.IsNull(),
			qm.Load(
				boiler.SyndicateElectionRels.SyndicateElectionCandidates,
				boiler.SyndicateElectionCandidateWhere.ResignedAt.IsNull(),
			),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Str("syndicate id", es.syndicateID).Msg("Failed to load unfinished syndicate election.")
			continue
		}

		// handle syndicate termination
		if es.isClosed.Load() {
			func() {
				// lock the system
				es.Lock()
				defer es.Unlock()

				if se != nil {
					// terminate ongoing syndicate election
					se.FinalisedAt = null.TimeFrom(time.Now())
					se.Result = null.StringFrom(boiler.SyndicateElectionResultTERMINATED)
					_, err = se.Update(gamedb.StdConn,
						boil.Whitelist(
							boiler.SyndicateElectionColumns.FinalisedAt,
							boiler.SyndicateElectionColumns.Result,
						),
					)
					// remove ongoing election in the frontend
					ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, nil)
				}
			}()
			return
		}

		// if no election found or candidate registration date is not ended
		if se == nil || se.CandidateRegisterCloseAt.After(now) {
			continue
		}

		// count candidate
		candidateCount := 0
		if se.R != nil && se.R.SyndicateElectionCandidates != nil {
			candidateCount = len(se.R.SyndicateElectionCandidates)
		}

		// if all candidates are resigned
		if candidateCount == 0 {
			se.FinalisedAt = null.TimeFrom(time.Now())
			se.Result = null.StringFrom(boiler.SyndicateElectionResultNO_CANDIDATE)
			_, err = se.Update(gamedb.StdConn,
				boil.Whitelist(
					boiler.SyndicateElectionColumns.FinalisedAt,
					boiler.SyndicateElectionColumns.Result,
				),
			)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to finalised election")
				return
			}
			// remove ongoing election in the frontend
			ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, nil)
		}

		// if not ended and more than one candidate
		if se.EndAt.After(now) && candidateCount > 1 {
			continue
		}

		// parse election result
		func(es *ElectionSystem, se *boiler.SyndicateElection) {
			es.Lock()
			defer es.Unlock()

			// instant pass, if only one candidate
			if candidateCount == 1 {
				es.setElectionWinner(se, se.R.SyndicateElectionCandidates[0].CandidateID)
				return
			}

			// if more than one candidate
			q := `
				SELECT 
				    sev.voted_for_candidate_id, 
				    COUNT(sev.voted_by_id) as total_votes 
				FROM syndicate_election_votes sev
				WHERE sev.syndicate_election_id = $1 AND 
				      EXISTS(
				          	SELECT 1 FROM syndicate_election_candidates sec 
				        	WHERE sec.candidate_id = sev.voted_for_candidate_id AND 
				        	      sec.resigned_at ISNULL 
				      )
				GROUP BY sev.voted_for_candidate_id
				ORDER BY COUNT(sev.voted_by_id) DESC
			`
			rows, err := gamedb.StdConn.Query(q, se.ID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("syndicate id", se.SyndicateID).Msg("Failed to parse syndicate election result.")
				return
			}

			var result []*ElectionResult
			for rows.Next() {
				cv := &ElectionResult{}
				err := rows.Scan(&cv.CandidateID, &cv.TotalVotes)
				if err != nil {
					gamelog.L.Error().
						Str("db func", "GetPlayerContributions").Err(err).Msg("Failed to scan syndicate candidate votes.")
					return
				}
				result = append(result, cv)
			}

			// if there is no one vote goes to any candidate
			if len(result) == 0 {
				se.FinalisedAt = null.TimeFrom(time.Now())
				se.Result = null.StringFrom(boiler.SyndicateElectionResultNO_VOTE)
				_, err = se.Update(gamedb.StdConn,
					boil.Whitelist(
						boiler.SyndicateElectionColumns.FinalisedAt,
						boiler.SyndicateElectionColumns.Result,
					),
				)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to finalised election")
					return
				}

				// remove ongoing election in the frontend
				ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, nil)

				// TODO: broadcast election result
				return
			}

			// if there is a winner
			if len(result) == 1 || result[0].TotalVotes > result[1].TotalVotes {
				es.setElectionWinner(se, result[0].CandidateID)
				return
			}

			// if tie happened...
			es.handleTieElection(se, result)

		}(es, se)

	}
}

func (es *ElectionSystem) setElectionWinner(se *boiler.SyndicateElection, candidateID string) {
	// load syndicate
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, se.SyndicateID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", se.SyndicateID).Msg("Failed to load syndicate detail.")
		return
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return
	}

	defer tx.Rollback()

	// finalise election
	se.WinnerID = null.StringFrom(candidateID)
	se.FinalisedAt = null.TimeFrom(time.Now())
	se.Result = null.StringFrom(boiler.SyndicateElectionResultWINNER_APPEAR)
	_, err = se.Update(tx,
		boil.Whitelist(
			boiler.SyndicateElectionColumns.WinnerID,
			boiler.SyndicateElectionColumns.FinalisedAt,
			boiler.SyndicateElectionColumns.Result,
		),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", se.SyndicateID).Msg("Failed to update syndicate election result")
		return
	}

	// update syndicate
	var updatedCols []string
	switch se.Type {
	case boiler.SyndicateElectionTypeCEO:
		syndicate.CeoPlayerID = null.StringFrom(candidateID)
		updatedCols = append(updatedCols, boiler.SyndicateColumns.CeoPlayerID)
	case boiler.SyndicateElectionTypeADMIN:
		syndicate.AdminID = null.StringFrom(candidateID)
		updatedCols = append(updatedCols, boiler.SyndicateColumns.AdminID)
	}

	_, err = syndicate.Update(tx, boil.Whitelist(updatedCols...))
	if err != nil {
		gamelog.L.Error().Err(err).Strs("updated column", updatedCols).Str("syndicate id", syndicate.ID).Msg("Failed to update syndicate leader")
		return
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return
	}

	// remove ongoing election in the frontend
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, nil)

	// TODO: broadcast election result
}

func (es *ElectionSystem) handleTieElection(se *boiler.SyndicateElection, result []*ElectionResult) {
	now := time.Now()

	// remove ongoing election in the frontend
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, nil)

	// terminate, if it is a second round
	if se.ParentElectionID.Valid {
		se.FinalisedAt = null.TimeFrom(now)
		se.Result = null.StringFrom(boiler.SyndicateElectionResultTIE_SECOND_TIME)
		_, err := se.Update(gamedb.StdConn,
			boil.Whitelist(
				boiler.SyndicateElectionColumns.FinalisedAt,
				boiler.SyndicateElectionColumns.Result,
			),
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to finalised election")
			return
		}

		// broadcast election result
		ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, se)
		return
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return
	}

	defer tx.Rollback()

	se.FinalisedAt = null.TimeFrom(now)
	se.Result = null.StringFrom(boiler.SyndicateElectionResultTIE)
	_, err = se.Update(tx,
		boil.Whitelist(
			boiler.SyndicateElectionColumns.FinalisedAt,
			boiler.SyndicateElectionColumns.Result,
		),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to finalised election")
		return
	}

	highestVotes := result[0].TotalVotes
	var newCandidates []string
	for _, r := range result {
		if r.TotalVotes < highestVotes {
			break
		}
		newCandidates = append(newCandidates, r.CandidateID)
	}

	newElection := boiler.SyndicateElection{
		SyndicateID:              se.SyndicateID,
		ParentElectionID:         null.StringFrom(se.ID),
		Type:                     se.Type,
		StartedAt:                now,
		CandidateRegisterCloseAt: now,                  // disable candidate registration
		EndAt:                    now.AddDate(0, 0, 1), // election only last for one day
	}

	err = newElection.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start new election.")
		return
	}

	// insert candidates
	for _, r := range result {
		ec := &boiler.SyndicateElectionCandidate{
			SyndicateID:         newElection.SyndicateID,
			SyndicateElectionID: newElection.ID,
			CandidateID:         r.CandidateID,
		}

		err = ec.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("election candidate", ec).Msg("Failed to insert election candidate.")
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return
	}

	// broadcast new election
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, newElection)

	// TODO: email all the syndicate members
}

func (es *ElectionSystem) heldElection() error {
	es.Lock()
	defer es.Unlock()

	if es.isClosed.Load() {
		return terror.Error(fmt.Errorf("election system is closed"), "Syndicate election system is closed.")
	}

	now := time.Now()
	// get current syndicate
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, es.syndicateID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load syndicate detail.")
		return terror.Error(err, "Failed to load syndicate detail.")
	}

	// check election exists
	se, err := boiler.SyndicateElections(
		boiler.SyndicateElectionWhere.SyndicateID.EQ(syndicate.ID),
		boiler.SyndicateElectionWhere.EndAt.LTE(now),
		boiler.SyndicateElectionWhere.FinalisedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate id", es.syndicateID).Msg("Failed to load unfinished syndicate election.")
		return terror.Error(err, "Failed to held a syndicate election.")
	}
	if se != nil {
		return terror.Error(fmt.Errorf("incomplete election"), "There is an ongoing election.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return terror.Error(err, "Failed held a syndicate election.")
	}

	defer tx.Rollback()

	// create an election base on syndicate type
	se = &boiler.SyndicateElection{
		SyndicateID:              es.syndicateID,
		StartedAt:                now,
		CandidateRegisterCloseAt: now.AddDate(0, 0, 1),
		EndAt:                    now.AddDate(0, 0, 3),
		Type:                     boiler.SyndicateElectionTypeADMIN,
	}

	if syndicate.Type == boiler.SyndicateTypeCORPORATION {
		se.Type = boiler.SyndicateElectionTypeCEO
	}

	err = se.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("election", se).Msg("Failed to insert new syndicate election.")
		return terror.Error(err, "Failed to held a syndicate election.")
	}

	switch se.Type {
	case boiler.SyndicateElectionTypeCEO:
		// insert ceo player as the first election candidate
		if syndicate.CeoPlayerID.Valid {
			ec := boiler.SyndicateElectionCandidate{
				SyndicateID:         se.SyndicateID,
				SyndicateElectionID: se.ID,
				CandidateID:         syndicate.CeoPlayerID.String,
			}

			err = ec.Insert(tx, boil.Infer())
			if err != nil {
				return terror.Error(err, "Failed to set ceo player as election candidate.")
			}
		}
	case boiler.SyndicateElectionTypeADMIN:
		// insert admin player as the first election candidate
		if syndicate.AdminID.Valid {
			ec := boiler.SyndicateElectionCandidate{
				SyndicateID:         se.SyndicateID,
				SyndicateElectionID: se.ID,
				CandidateID:         syndicate.AdminID.String,
			}

			err = ec.Insert(tx, boil.Infer())
			if err != nil {
				return terror.Error(err, "Failed to set admin player as election candidate.")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to held a syndicate election.")
	}

	// broadcast to all the members
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_election", es.factionID, es.syndicateID), server.HubKeySyndicateOngoingElectionSubscribe, se)

	// TODO: email all the syndicate members

	return nil
}

func (es *ElectionSystem) registerCandidate(candidateID string) error {
	es.Lock()
	defer es.Unlock()

	if es.isClosed.Load() {
		return terror.Error(fmt.Errorf("election system is closed"), "Syndicate election system is closed.")
	}

	// load syndicate election
	se, err := boiler.SyndicateElections(
		boiler.SyndicateElectionWhere.SyndicateID.EQ(es.syndicateID),
		boiler.SyndicateElectionWhere.FinalisedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate id", es.syndicateID).Msg("Failed to load syndicate election.")
		return terror.Error(err, "Failed to load syndicate election.")
	}

	if se == nil {
		return terror.Error(fmt.Errorf("no election"), "There is no ongoing election.")
	}

	if se.CandidateRegisterCloseAt.Before(time.Now()) {
		return terror.Error(fmt.Errorf("registration window closed"), "Candidate registration period is over.")
	}

	// check player is already registered
	sec, err := boiler.SyndicateElectionCandidates(
		boiler.SyndicateElectionCandidateWhere.SyndicateElectionID.EQ(se.ID),
		boiler.SyndicateElectionCandidateWhere.CandidateID.EQ(candidateID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to load syndicate election candidate")
		return terror.Error(err, "Failed to check candidate list.")
	}

	if sec != nil {
		return terror.Error(fmt.Errorf("already registered"), "You have already registered as an election candidate.")
	}

	sec = &boiler.SyndicateElectionCandidate{
		SyndicateElectionID: se.ID,
		CandidateID:         candidateID,
		SyndicateID:         se.SyndicateID,
	}

	err = sec.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("candidate", sec).Msg("Failed to insert syndicate election candidate.")
		return terror.Error(err, "Failed to register syndicate election candidate.")
	}

	return nil
}

func (es *ElectionSystem) candidateResign(userID string) error {
	es.Lock()
	defer es.Unlock()

	if es.isClosed.Load() {
		return terror.Error(fmt.Errorf("election system is closed"), "Syndicate election system is closed.")
	}

	// load syndicate election
	se, err := boiler.SyndicateElections(
		boiler.SyndicateElectionWhere.SyndicateID.EQ(es.syndicateID),
		boiler.SyndicateElectionWhere.FinalisedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate id", es.syndicateID).Msg("Failed to load syndicate election.")
		return terror.Error(err, "Failed to load syndicate election.")
	}

	if se == nil {
		return terror.Error(fmt.Errorf("no election"), "There is no ongoing election.")
	}

	// check player is already registered
	sec, err := boiler.SyndicateElectionCandidates(
		boiler.SyndicateElectionCandidateWhere.SyndicateElectionID.EQ(se.ID),
		boiler.SyndicateElectionCandidateWhere.CandidateID.EQ(userID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to load syndicate election candidate")
		return terror.Error(err, "Failed to check candidate list.")
	}

	if sec == nil {
		return terror.Error(fmt.Errorf("not candidate"), "You are not a candidate of the election.")
	}

	if sec.ResignedAt.Valid {
		return terror.Error(fmt.Errorf("already resigned"), "You have already resigned.")
	}

	sec.ResignedAt = null.TimeFrom(time.Now())
	_, err = sec.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateElectionCandidateColumns.ResignedAt))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("candidate", sec).Msg("Failed to update syndicate election candidate.")
		return terror.Error(err, "Failed to resign election.")
	}

	return nil
}

func (es *ElectionSystem) vote(voterID string, candidateID string) error {
	es.Lock()
	defer es.Unlock()

	if es.isClosed.Load() {
		return terror.Error(fmt.Errorf("election system is closed"), "Syndicate election system is closed.")
	}

	// load syndicate election
	se, err := boiler.SyndicateElections(
		boiler.SyndicateElectionWhere.SyndicateID.EQ(es.syndicateID),
		boiler.SyndicateElectionWhere.FinalisedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate id", es.syndicateID).Msg("Failed to load syndicate election.")
		return terror.Error(err, "Failed to load syndicate election.")
	}

	if se == nil {
		return terror.Error(fmt.Errorf("no election"), "There is no ongoing election.")
	}

	// check candidate exists
	sec, err := boiler.SyndicateElectionCandidates(
		boiler.SyndicateElectionCandidateWhere.SyndicateElectionID.EQ(se.ID),
		boiler.SyndicateElectionCandidateWhere.CandidateID.EQ(candidateID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("election id", se.ID).Str("candidate id", candidateID).Msg("Failed to load syndicate election candidate")
		return terror.Error(err, "Failed to check candidate list.")
	}

	if sec == nil {
		return terror.Error(fmt.Errorf("no candidate"), "The candidate does not exist.")
	}

	// check player is already voted
	sev, err := boiler.SyndicateElectionVotes(
		boiler.SyndicateElectionVoteWhere.SyndicateElectionID.EQ(se.ID),
		boiler.SyndicateElectionVoteWhere.VotedByID.EQ(voterID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("election id", se.ID).Str("voter id", voterID).Msg("Failed to load syndicate election vote")
		return terror.Error(err, "Failed to check election vote.")
	}

	if sev != nil {
		return terror.Error(fmt.Errorf("already voted"), "You have already voted.")
	}

	// insert vote
	sev = &boiler.SyndicateElectionVote{
		SyndicateElectionID: se.ID,
		VotedByID:           voterID,
		VotedForCandidateID: candidateID,
	}
	err = sev.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("syndicate election vote", sev).Msg("Failed to insert syndicate election vote")
		return terror.Error(err, "Failed to vote for the candidate.")
	}

	return nil
}
