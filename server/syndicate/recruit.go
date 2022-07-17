package syndicate

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
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

	sync.Mutex
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

type Application struct {
	*boiler.SyndicateJoinApplication
	isClosed atomic.Bool
	sync.Mutex
	onClose func()
}

func (a *Application) start() {
	defer func() {
		// NOTE: this is the ONLY place where a application is closed and removed from the system!
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

		return
	}
}

func (a *Application) parseResult() {
	a.Lock()
	defer a.Unlock()

	a.isClosed.Store(true)

	// get application vote
	votes, err := a.ApplicationApplicationVotes().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("application id", a.ID).Err(err).Msg("Failed to load committee votes")
		return
	}

	agreedCount := 0
	disagreedCount := 0

	if len(votes) == 0 {
		a.Result = null.StringFrom(boiler.SyndicateJoinApplicationResultREJECTED)
		a.FinalisedAt = null.TimeFrom(time.Now())
		a.Note = null.StringFrom("No committee voted before application is expired.")
		return
	}

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
