package syndicate

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"sync"
	"time"
)

type RecruitSystem struct {
	syndicate *Syndicate
	isLocked  atomic.Bool

	applicationMap map[string]*Application

	sync.RWMutex
}

func newRecruitSystem(s *Syndicate) *RecruitSystem {
	rs := &RecruitSystem{
		syndicate:      s,
		isLocked:       atomic.Bool{},
		applicationMap: make(map[string]*Application),
	}
	rs.isLocked.Store(false)

	return rs
}

func (rs *RecruitSystem) getApplication(id string) (*Application, error) {
	rs.RLock()
	defer rs.RUnlock()

	a, ok := rs.applicationMap[id]
	if !ok {
		return nil, terror.Error(fmt.Errorf("application not found"), "Application does not exist.")
	}

	if a.isClosed.Load() {
		return nil, terror.Error(fmt.Errorf("application is finalised"), "Application is finalised")
	}

	return a, nil
}

func (rs *RecruitSystem) receiveApplication(application *boiler.SyndicateJoinApplication) error {
	// charge applicant upfront
	rs.Lock()
	defer rs.Unlock()

	if rs.isLocked.Load() {
		return terror.Error(fmt.Errorf("recurit system is closed"), "Recruit system is closed.")
	}

	txid, err := rs.syndicate.system.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.FromStringOrNil(application.ApplicantID),
		ToUserID:             uuid.FromStringOrNil(server.SupremacyGameUserID),
		Amount:               application.PaidAmount.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("submit_syndicate_join_application|%s|%d", application.SyndicateID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupSyndicate),
		Description:          "submit syndicate join application.",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("player_id", application.ApplicantID).Str("amount", application.PaidAmount.String()).Err(err).Msg("Failed to submit syndicate join application fee.")
		return terror.Error(err, "Failed to pay the fee of syndicate join application.")
	}

	application.TXID = null.StringFrom(txid)

	// update transaction id
	go func(application *boiler.SyndicateJoinApplication) {
		_, err = application.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateJoinApplicationColumns.TXID))
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("tx id", txid).
				Str("application id", application.ID).
				Str("amount", application.PaidAmount.String()).
				Msg("Failed to update transaction id of the syndicate join application.")
		}
	}(application)

	a := &Application{
		SyndicateJoinApplication: application,
		recruitSystem:            rs,
		isClosed:                 atomic.Bool{},
		onClose: func() {
			rs.Lock()
			defer rs.Unlock()

			delete(rs.applicationMap, application.ID)
		},
	}

	a.isClosed.Store(false)

	go a.start()

	rs.applicationMap[application.ID] = a

	return nil
}

func (rs *RecruitSystem) VoteApplication(applicationID string, userID string, isAgreed bool) error {

	a, err := rs.getApplication(applicationID)
	if err != nil {
		return err
	}

	err = a.vote(userID, isAgreed)
	if err != nil {
		return err
	}

	return nil
}

type Application struct {
	*boiler.SyndicateJoinApplication
	recruitSystem *RecruitSystem
	isClosed      atomic.Bool
	sync.Mutex
	onClose func()
}

func (a *Application) start() {
	defer func() {
		// NOTE: this is the ONLY place where an application is closed and removed from the system!
		a.onClose()
	}()

	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		if a.ExpireAt.After(time.Now()) || !a.isClosed.Load() {
			continue
		}

		// parse result
		a.parseResult()

		// do action
		a.Action()

		return
	}
}

func (a *Application) vote(userID string, isAgreed bool) error {
	a.Lock()
	defer a.Unlock()

	if a.isClosed.Load() {
		return terror.Error(fmt.Errorf("application is finalised"), "The application is finalised.")
	}

	// check player has voted already
	isVoted, err := boiler.ApplicationVoteExists(gamedb.StdConn, a.ID, userID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to check application vote in db.")
		return terror.Error(err, "Failed to vote on the application.")
	}

	if isVoted {
		return terror.Error(fmt.Errorf("already voted"), "You have already voted.")
	}

	av := &boiler.ApplicationVote{
		ApplicationID: a.ID,
		VotedByID:     userID,
		IsAgreed:      isAgreed,
	}

	err = av.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("application vote", av).Msg("Failed to insert application vote.")
		return terror.Error(fmt.Errorf("failed to vote"), "Failed to vote.")
	}

	totalVoters, err := boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(a.SyndicateID),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", a.SyndicateID).Msg("Failed to check total syndicate committees.")
	}

	// total vote count
	agreedCount, err := a.ApplicationApplicationVotes(
		boiler.ApplicationVoteWhere.IsAgreed.EQ(true),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("application id", a.ID).Msg("Failed to check vote count")
		return nil
	}

	// close the vote, if more than half of committees agreed
	if agreedCount > totalVoters/2 {
		a.isClosed.Store(true)
		return nil
	}

	disagreedCount, err := a.ApplicationApplicationVotes(
		boiler.ApplicationVoteWhere.IsAgreed.EQ(false),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("application id", a.ID).Msg("Failed to check vote count")
		return nil
	}

	// close the vote, if more than half of committees disagreed
	if disagreedCount > totalVoters/2 || totalVoters == disagreedCount+agreedCount {
		a.isClosed.Store(true)
		return nil
	}

	return nil
}

func (a *Application) parseResult() {
	a.Lock()
	defer a.Unlock()

	a.isClosed.Store(true)

	// TODO: admin or ceo interrupt

	// get application vote
	votes, err := a.ApplicationApplicationVotes().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("application id", a.ID).Err(err).Msg("Failed to load committee votes")
		return
	}

	agreedCount := 0
	disagreedCount := 0

	for _, v := range votes {
		if v.IsAgreed {
			agreedCount += 1
			continue
		}
		disagreedCount += 1
	}

	if agreedCount <= disagreedCount {
		a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultREJECTED)
		a.FinalisedAt = null.TimeFrom(time.Now())
		a.Note = null.StringFrom("Not enough committees agreed on the application.")
		return
	}

	// when passed
	a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultACCEPTED)
	a.FinalisedAt = null.TimeFrom(time.Now())
	a.Note = null.StringFrom("Majority of the committees agreed.")
}

func (a *Application) Action() {
	// load applicant
	applicant, err := a.Applicant().One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load applicant.")
		return
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
		return
	}

	defer tx.Rollback()

	// insert application
	_, err = a.Update(tx, boil.Whitelist(
		boiler.SyndicateJoinApplicationColumns.Result,
		boiler.SyndicateJoinApplicationColumns.FinalisedAt,
		boiler.SyndicateJoinApplicationColumns.Note,
	))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert application to db.")
		return
	}

	if a.Result.String == boiler.SyndicateJoinApplicationResultREJECTED {
		// refund application fee to applicant
		refundID, err := a.recruitSystem.syndicate.system.Passport.RefundSupsMessage(a.TXID.String)
		if err != nil {
			gamelog.L.Error().Str("player_id", a.ApplicantID).Str("amount", a.PaidAmount.String()).Err(err).Msg("Failed to submit syndicate join application fee.")
			return
		}

		a.RefundTXID = null.StringFrom(refundID)
		_, err = a.Update(tx, boil.Whitelist(
			boiler.SyndicateJoinApplicationColumns.Result,
			boiler.SyndicateJoinApplicationColumns.FinalisedAt,
			boiler.SyndicateJoinApplicationColumns.Note,
		))
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to insert application to db.")
		}
	} else {
		applicant.SyndicateID = null.StringFrom(a.SyndicateID)
		_, err := applicant.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
		if err != nil {
			gamelog.L.Error().Str("player id", applicant.ID).Str("syndicate id", a.SyndicateID).Err(err).Msg("Failed to update player syndicate id.")
			return
		}
		err = a.recruitSystem.syndicate.accountSystem.receiveFund(
			server.SupremacyGameUserID,
			a.PaidAmount,
			server.TransactionReference(fmt.Sprintf("syndicate_member_join_fee|%s|%s|%d", a.ApplicantID, a.SyndicateID, time.Now().UnixNano())),
			fmt.Sprintf("Player '%s' #%d join fee.", applicant.Username.String, applicant.Gid),
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to transfer syndicate join fee.")
			return
		}

		ws.PublishMessage(fmt.Sprintf("/user/%s", applicant.ID), server.HubKeyUserSubscribe, applicant)
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/join_applicant/%s", applicant.FactionID.String, a.SyndicateID, a.ID), server.HubKeySyndicateJoinApplicationUpdate, a)

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return
	}
}
