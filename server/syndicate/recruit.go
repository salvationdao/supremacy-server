package syndicate

import (
	"fmt"
	"github.com/gofrs/uuid"
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
	"server/xsyn_rpcclient"
	"sync"
	"time"
)

type RecruitSystem struct {
	syndicate *Syndicate
	isClosed  atomic.Bool

	applicationMap map[string]*Application

	deadlock.RWMutex
}

func newRecruitSystem(s *Syndicate) (*RecruitSystem, error) {
	rs := &RecruitSystem{
		syndicate:      s,
		applicationMap: make(map[string]*Application),
	}
	rs.isClosed.Store(false)

	// load incomplete applications
	as, err := boiler.SyndicateJoinApplications(
		boiler.SyndicateJoinApplicationWhere.SyndicateID.EQ(s.ID),
		boiler.SyndicateJoinApplicationWhere.FinalisedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to load syndicate join application")
	}

	// load applications into system
	for _, application := range as {
		a := &Application{
			SyndicateJoinApplication: application,
			recruitSystem:            rs,
			onClose: func() {
				rs.Lock()
				defer rs.Unlock()

				delete(rs.applicationMap, application.ID)
			},
		}

		a.isClosed.Store(false)
		a.forceClosed.Store("")

		go a.start()

		rs.applicationMap[application.ID] = a
	}

	return rs, nil
}

func (rs *RecruitSystem) terminate() {
	rs.Lock()
	defer rs.Unlock()

	rs.isClosed.Store(true)

	for _, a := range rs.applicationMap {
		a.forceClosed.Store(boiler.SyndicateJoinApplicationResultTERMINATED)
	}

}

func (rs *RecruitSystem) getApplication(id string) (*Application, error) {
	rs.RLock()
	defer rs.RUnlock()

	if rs.isClosed.Load() {
		return nil, terror.Error(fmt.Errorf("recuit system is closed"), "Syndicate recruit system is closed.")
	}

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

	if rs.isClosed.Load() {
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
		onClose: func() {
			rs.Lock()
			defer rs.Unlock()

			delete(rs.applicationMap, application.ID)
		},
	}

	a.isClosed.Store(false)
	a.forceClosed.Store("")

	go a.start()

	rs.applicationMap[application.ID] = a

	return nil
}

func (rs *RecruitSystem) voteApplication(applicationID string, userID string, isAgreed bool) error {
	// NOTE: do not lock rs, otherwise deadlock will occur.

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
	forceClosed   atomic.String
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
		if a.ExpireAt.After(time.Now()) || !a.isClosed.Load() || a.forceClosed.Load() == "" {
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

	if a.recruitSystem.isClosed.Load() {
		return terror.Error(fmt.Errorf("recuit system is closed"), "Syndicate recruit system is closed.")
	}

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
	votes, err := a.ApplicationApplicationVotes().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("application id", a.ID).Msg("Failed to check vote count")
		return nil
	}

	agreedCount := int64(0)
	disagreedCount := int64(0)
	for _, v := range votes {
		if v.IsAgreed {
			agreedCount += 1
			continue
		}
		disagreedCount += 1
	}
	currentVoteCount := agreedCount + disagreedCount

	// close the vote
	if agreedCount > totalVoters/2 || disagreedCount > totalVoters/2 || currentVoteCount == totalVoters {
		a.isClosed.Store(true)
		return nil
	}

	return nil
}

func (rs *RecruitSystem) finaliseApplication(playerPosition string, applicationID string, isAccepted bool) error {
	rs.Lock()
	defer rs.Unlock()

	if rs.isClosed.Load() {
		return terror.Error(fmt.Errorf("recruit system is closed"), "Recruit system is already closed.")
	}

	a, ok := rs.applicationMap[applicationID]
	if !ok {
		return terror.Error(fmt.Errorf("application not exist"), "Application does not exist")
	}

	if a.isClosed.Load() {
		return terror.Error(fmt.Errorf("application is closed"), "application is already closed.")
	}

	if a.forceClosed.Load() != "" {
		return terror.Error(fmt.Errorf("application is finalised"), "application is already finalised.")
	}

	// finalise application
	decision := fmt.Sprintf("%s_ACCEPT", playerPosition)
	if !isAccepted {
		decision = fmt.Sprintf("%s_REJECT", playerPosition)
	}
	a.forceClosed.Store(decision)

	return nil
}

func (a *Application) parseResult() {
	a.Lock()
	defer a.Unlock()

	a.isClosed.Store(true)

	now := time.Now()

	// admin and ceo interruption
	if a.forceClosed.Load() != "" {
		// parse force close reason
		switch a.forceClosed.Load() {
		case boiler.SyndicateJoinApplicationResultTERMINATED:
			a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultTERMINATED)
			a.Note = null.StringFrom("Terminated by system.")

		case "CEO_ACCEPT":
			a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultACCEPTED)
			a.Note = null.StringFrom("Accepted by syndicate ceo.")

		case "CEO_REJECT":
			a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultREJECTED)
			a.Note = null.StringFrom("Rejected by syndicate ceo.")

		case "ADMIN_ACCEPT":
			a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultACCEPTED)
			a.Note = null.StringFrom("Accepted by syndicate admin.")

		case "ADMIN_REJECT":
			a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultREJECTED)
			a.Note = null.StringFrom("Rejected by syndicate admin.")
		}
		a.FinalisedAt = null.TimeFrom(now)
		return
	}

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
		a.FinalisedAt = null.TimeFrom(now)
		a.Note = null.StringFrom("Not enough committees agreed on the application.")
		return
	}

	// when passed
	a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultACCEPTED)
	a.FinalisedAt = null.TimeFrom(now)
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

	// update application
	_, err = a.Update(tx, boil.Whitelist(
		boiler.SyndicateJoinApplicationColumns.Result,
		boiler.SyndicateJoinApplicationColumns.FinalisedAt,
		boiler.SyndicateJoinApplicationColumns.Note,
	))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert application to db.")
		return
	}

	switch a.Result.String {
	case boiler.SyndicateJoinApplicationResultREJECTED,
		boiler.SyndicateJoinApplicationResultTERMINATED:
		// refund application fee to applicant
		refundID, err := a.recruitSystem.syndicate.system.Passport.RefundSupsMessage(a.TXID.String)
		if err != nil {
			gamelog.L.Error().Str("player_id", a.ApplicantID).Str("amount", a.PaidAmount.String()).Err(err).Msg("Failed to submit syndicate join application fee.")
			return
		}

		a.RefundTXID = null.StringFrom(refundID)
		_, err = a.Update(tx, boil.Whitelist(
			boiler.SyndicateJoinApplicationColumns.RefundTXID,
		))
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to insert application to db.")
		}

	case boiler.SyndicateJoinApplicationResultACCEPTED:
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
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return
	}

	// TODO: send message to applicant

	err = applicant.L.LoadRole(gamedb.StdConn, true, applicant, nil)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load role")
		return
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s", applicant.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(applicant))

	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/join_applicant/%s", applicant.FactionID.String, a.SyndicateID, a.ID), server.HubKeySyndicateJoinApplicationUpdate, a)

}
