package api

import (
	"database/sql"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type SyndicateSystem struct {
	syndicateMap map[string]*Syndicate
	sync.RWMutex
}

func NewSyndicateSystem() (*SyndicateSystem, error) {
	syndicates, err := boiler.Syndicates().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load syndicated from db")
		return nil, terror.Error(err, "Failed to get syndicates from db")
	}

	ss := &SyndicateSystem{
		syndicateMap: make(map[string]*Syndicate),
	}

	for _, s := range syndicates {
		err = ss.AddSyndicate(s)
		if err != nil {
			gamelog.L.Error().Interface("syndicate", s).Err(err).Msg("Failed to add syndicate")
			return nil, terror.Error(err, "Failed to add syndicate")
		}
	}

	return ss, nil
}

// VoteMotion get the motion from the syndicate and fire vote logic
func (ss *SyndicateSystem) VoteMotion(syndicateID string, motionID string, logic func(motion *SyndicateMotion) error) error {
	// get syndicate
	s, err := ss.GetSyndicate(syndicateID)
	if err != nil {
		return err
	}

	// fire motion vote
	om, err := s.motionSystem.GetOngoingMotion(motionID)
	if err != nil {
		return err
	}

	return om.Vote(logic)
}

// AddSyndicate add new syndicate to the system
func (ss *SyndicateSystem) AddSyndicate(syndicate *boiler.Syndicate) error {
	ss.Lock()
	defer ss.Unlock()

	s, err := NewSyndicate(syndicate)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate")
		return terror.Error(err, "Failed to spin up syndicate")
	}

	ss.syndicateMap[s.ID] = s

	return nil
}

// RemoveSyndicate remove syndicate from the system
func (ss *SyndicateSystem) RemoveSyndicate(id string) error {
	ss.Lock()
	defer ss.Unlock()

	if s, ok := ss.syndicateMap[id]; ok {
		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to begin db transaction")
			return terror.Error(err, "Failed to remove syndicate")
		}

		defer func(tx *sql.Tx) {
			_ = tx.Rollback()
		}(tx)

		// evacuate all the members
		_, err = boiler.Players(
			boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(s.ID)),
		).UpdateAll(tx, boiler.M{boiler.PlayerColumns.SyndicateID: null.StringFromPtr(nil)})
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to players' syndicate id")
			return terror.Error(err, "Failed to remove syndicate")
		}

		_, err = boiler.Players(
			boiler.PlayerWhere.DirectorOfSyndicateID.EQ(null.StringFrom(s.ID)),
		).UpdateAll(tx, boiler.M{boiler.PlayerColumns.DirectorOfSyndicateID: null.StringFromPtr(nil)})
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to players' director syndicate id")
			return terror.Error(err, "Failed to remove syndicate")
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
			return terror.Error(err, "Failed to remove syndicate")
		}

		// stop all the motion in the syndicate
		s.motionSystem.terminated()
	}

	// delete syndicate from the map
	delete(ss.syndicateMap, id)

	return nil
}

func (ss *SyndicateSystem) GetSyndicate(id string) (*Syndicate, error) {
	ss.RLock()
	defer ss.RUnlock()

	s, ok := ss.syndicateMap[id]
	if !ok {
		return nil, terror.Error(fmt.Errorf("syndicate not exist"), "Syndicate does not exit")
	}

	return s, nil
}

type Syndicate struct {
	*boiler.Syndicate
	motionSystem *SyndicateMotionSystem
}

func NewSyndicate(syndicate *boiler.Syndicate) (*Syndicate, error) {
	motionSystem, err := NewSyndicateMotionSystem(syndicate)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate motion system.")
		return nil, terror.Error(err, "Failed to spin up syndicate motion system.")
	}

	s := &Syndicate{
		syndicate,
		motionSystem,
	}

	return s, nil
}

type SyndicateMotionSystem struct {
	syndicate      *boiler.Syndicate
	ongoingMotions map[string]*SyndicateMotion
	sync.RWMutex

	isClosed atomic.Bool
}

// NewSyndicateMotionSystem generate a new syndicate motion system
func NewSyndicateMotionSystem(syndicate *boiler.Syndicate) (*SyndicateMotionSystem, error) {
	ms, err := boiler.SyndicateMotions(
		boiler.SyndicateMotionWhere.SyndicateID.EQ(syndicate.ID),
		boiler.SyndicateMotionWhere.EndedAt.GT(time.Now()),
		qm.Load(boiler.SyndicateMotionRels.MotionSyndicateMotionVotes),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to load syndicate motions.")
		return nil, terror.Error(err, "Failed to load syndicate motions")
	}

	sms := &SyndicateMotionSystem{
		syndicate:      syndicate,
		ongoingMotions: map[string]*SyndicateMotion{},
		isClosed:       atomic.Bool{},
	}
	sms.isClosed.Store(false)

	for _, m := range ms {
		err = sms.AddMotion(m)
		if err != nil {
			// the duplicate check failed, so log error instead of return the system
			gamelog.L.Error().Err(err).Msg("Failed to add motion")
		}
	}

	return sms, nil
}

func (sms *SyndicateMotionSystem) terminated() {
	sms.Lock()
	defer sms.Unlock()
	sms.isClosed.Store(true)

	// force close all the motion
	for _, om := range sms.ongoingMotions {
		om.forceClosed.Store(true) // close motion without calculate result
		om.isClosed.Store(true)    // terminate all the processing motion
	}
}

// GetOngoingMotion return ongoing motion by id
func (sms *SyndicateMotionSystem) GetOngoingMotion(id string) (*SyndicateMotion, error) {
	sms.RLock()
	defer sms.RUnlock()
	// return if syndicate is closed
	if sms.isClosed.Load() {
		return nil, terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	om, ok := sms.ongoingMotions[id]
	if !ok {
		return nil, terror.Error(fmt.Errorf("motion does not exist"), "Motion does not exist")
	}

	if om.isClosed.Load() {
		return nil, terror.Error(fmt.Errorf("motion is closed"), "Motion is closed")
	}

	return om, nil
}

// AddMotion check motion duplicated content and append new motion to motion system
func (sms *SyndicateMotionSystem) AddMotion(bsm *boiler.SyndicateMotion) error {
	// check already exists
	err := sms.DuplicatedMotionCheck(bsm)
	if err != nil {
		return err
	}

	// add motion to system
	sms.Lock()
	defer sms.Unlock()

	// return if syndicate is closed
	if sms.isClosed.Load() {
		return terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	m := &SyndicateMotion{
		syndicate:       sms.syndicate,
		SyndicateMotion: bsm,
		votes:           []*boiler.SyndicateMotionVote{},
		isClosed:        atomic.Bool{},
		forceClosed:     atomic.Bool{},
	}

	if bsm.R != nil {
		for _, smv := range bsm.R.MotionSyndicateMotionVotes {
			m.votes = append(m.votes, smv)
		}
	}

	// function for removing motion from motion system map
	m.isClosed.Store(false)
	m.forceClosed.Store(false)
	m.onClose = func() {
		sms.Lock()
		defer sms.Unlock()

		delete(sms.ongoingMotions, m.ID)
	}

	// add motion to the map
	sms.ongoingMotions[m.ID] = m

	// spin up motion
	go m.start()

	return nil
}

func (sms *SyndicateMotionSystem) DuplicatedMotionCheck(bsm *boiler.SyndicateMotion) error {
	sms.RLock()
	defer sms.RUnlock()

	// return if syndicate is closed
	if sms.isClosed.Load() {
		return terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	for _, om := range sms.ongoingMotions {
		// skip, if different motion type
		if om.isClosed.Load() {
			continue
		}

		// if motion have different type, excluding remove rule and change rule
		if bsm.Type != om.Type && bsm.Type != boiler.SyndicateMotionTypeREMOVE_RULE && bsm.Type != boiler.SyndicateMotionTypeCHANGE_RULE {
			continue
		}

		// check change content is duplicated
		switch om.Type {
		case boiler.SyndicateMotionTypeCHANGE_GENERAL_DETAIL:
			if bsm.NewName.Valid && om.NewName.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate name.")
			}
			if bsm.NewSymbolID.Valid && om.NewSymbolID.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate symbol.")
			}
			if bsm.NewNamingConvention.Valid && om.NewNamingConvention.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate naming convention.")
			}
		case boiler.SyndicateMotionTypeCHANGE_PAYMENT_SETTING:
			if bsm.NewJoinFee.Valid && om.NewJoinFee.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate join fee.")
			}
			if bsm.NewExitFee.Valid && om.NewExitFee.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate exit fee.")
			}
			if bsm.NewDeployingUserPercentage.Valid && om.NewDeployingUserPercentage.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate deploying user cut percentage.")
			}
			if bsm.NewAbilityKillPercentage.Valid && om.NewAbilityKillPercentage.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate ability kill cut percentage.")
			}
			if bsm.NewMechOwnerPercentage.Valid && om.NewMechOwnerPercentage.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate mech owner cut percentage.")
			}
			if bsm.NewSyndicateCutPercentage.Valid && om.NewSyndicateCutPercentage.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate cut percentage.")
			}
		case boiler.SyndicateMotionTypeADD_RULE:
			if bsm.NewRuleContent.Valid && om.NewRuleContent.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for adding the same rule.")
			}
		case boiler.SyndicateMotionTypeREMOVE_RULE, boiler.SyndicateMotionTypeCHANGE_RULE:
			if bsm.RuleID.String == om.RuleID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing the same rule.")
			}
		case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
			if bsm.DirectorID.String == om.DirectorID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for appointing the same player.")
			}
		case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
			if bsm.DirectorID.String == om.DirectorID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for removing the same director.")
			}
		case boiler.SyndicateMotionTypeREMOVE_FOUNDER:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for removing the founder.")
		}
	}

	return nil
}

type SyndicateMotion struct {
	*boiler.SyndicateMotion
	syndicate *boiler.Syndicate
	votes     []*boiler.SyndicateMotionVote
	sync.Mutex

	isClosed    atomic.Bool
	forceClosed atomic.Bool
	onClose     func()
}

func (sm *SyndicateMotion) start() {
	defer func() {
		// NOTE: this is the ONLY place where a motion close and be removed from its motion system!
		sm.onClose()
	}()

	for {
		time.Sleep(1 * time.Second)
		// if motion is not ended and not closed
		if sm.EndedAt.After(time.Now()) && !sm.isClosed.Load() && !sm.forceClosed.Load() {
			continue
		}

		// directly clean up the motion, if force close
		if sm.forceClosed.Load() {
			return
		}

		// calculate result
		sm.calcResult()

		return
	}
}

// Vote check motion is closed or not before firing the function logic
func (sm *SyndicateMotion) Vote(logic func(motion *SyndicateMotion) error) error {
	sm.Lock()
	defer sm.Unlock()

	if sm.isClosed.Load() || sm.forceClosed.Load() {
		return terror.Error(fmt.Errorf("motion is closed"), "Motion is closed")
	}

	// only fire the function when motion is still open
	return logic(sm)
}

func (sm *SyndicateMotion) calcResult() {
	sm.Lock()
	defer sm.Unlock()

	// skip if force closed
	if sm.forceClosed.Load() {
		return
	}

	sm.isClosed.Store(true)

	// TODO: start calculate result
}
