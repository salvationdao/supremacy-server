package api

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

// AddSyndicate add new syndicate to the system
func (ss *SyndicateSystem) AddSyndicate(syndicate *boiler.Syndicate) error {
	ss.Lock()
	defer ss.Unlock()

	s, err := newSyndicate(syndicate)
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

// AddMotion add new motion to the syndicate system
func (ss *SyndicateSystem) AddMotion(user *boiler.Player, bsm *boiler.SyndicateMotion) error {
	// get syndicate
	s, err := ss.GetSyndicate(user.SyndicateID.String)
	if err != nil {
		return err
	}

	if s.Type == boiler.SyndicateTypeCORPORATION && !user.DirectorOfSyndicateID.Valid {
		return terror.Error(fmt.Errorf("only director can issue motion"), "Only director can issue motion")
	}

	// check motion is valid and generate a clean motion
	newMotion, err := s.motionSystem.motionValidCheck(user, bsm)
	if err != nil {
		return err
	}

	// check already exists
	err = s.motionSystem.duplicatedMotionCheck(newMotion)
	if err != nil {
		return err
	}

	// add motion to the motion system
	err = s.motionSystem.addMotion(newMotion)
	if err != nil {
		return err
	}

	return nil
}

// VoteMotion get the motion from the syndicate and fire vote logic
func (ss *SyndicateSystem) VoteMotion(user *boiler.Player, motionID string, isAgreed bool) error {
	// get syndicate
	s, err := ss.GetSyndicate(user.SyndicateID.String)
	if err != nil {
		return err
	}

	if s.Type == boiler.SyndicateTypeCORPORATION && !user.DirectorOfSyndicateID.Valid {
		return terror.Error(fmt.Errorf("only director can vote"), "Only director can vote")
	}

	// fire motion vote
	om, err := s.motionSystem.getOngoingMotion(motionID)
	if err != nil {
		return err
	}

	if om.SyndicateID != user.SyndicateID.String {
		return terror.Error(fmt.Errorf("not a member of the syndicate"), "Player is not a member of the syndicate")
	}

	return om.vote(user, isAgreed)
}

type Syndicate struct {
	*boiler.Syndicate
	sync.RWMutex // for update syndicate

	motionSystem *SyndicateMotionSystem
}

func (s *Syndicate) store(syndicate *boiler.Syndicate) {
	s.Lock()
	defer s.Unlock()

	s.Syndicate = syndicate
}

func (s *Syndicate) load() boiler.Syndicate {
	s.RLock()
	defer s.RUnlock()

	return *s.Syndicate
}

func newSyndicate(syndicate *boiler.Syndicate) (*Syndicate, error) {
	s := &Syndicate{
		Syndicate: syndicate,
	}
	motionSystem, err := s.newSyndicateMotionSystem()
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate motion system.")
		return nil, terror.Error(err, "Failed to spin up syndicate motion system.")
	}

	s.motionSystem = motionSystem

	return s, nil
}

type SyndicateMotionSystem struct {
	syndicate      *Syndicate
	ongoingMotions map[string]*SyndicateMotion
	sync.RWMutex

	isClosed atomic.Bool
}

// newSyndicateMotionSystem generate a new syndicate motion system
func (s *Syndicate) newSyndicateMotionSystem() (*SyndicateMotionSystem, error) {
	ms, err := boiler.SyndicateMotions(
		boiler.SyndicateMotionWhere.SyndicateID.EQ(s.ID),
		boiler.SyndicateMotionWhere.EndedAt.GT(time.Now()),
		boiler.SyndicateMotionWhere.ActualEndedAt.IsNull(),
		qm.Load(boiler.SyndicateMotionRels.MotionSyndicateMotionVotes),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", s.ID).Err(err).Msg("Failed to load syndicate motions.")
		return nil, terror.Error(err, "Failed to load syndicate motions")
	}

	sms := &SyndicateMotionSystem{
		syndicate:      s,
		ongoingMotions: map[string]*SyndicateMotion{},
		isClosed:       atomic.Bool{},
	}
	sms.isClosed.Store(false)

	for _, m := range ms {
		err = sms.addMotion(m)
		if err != nil {
			// the duplicate check failed, so log error instead of return the system
			gamelog.L.Error().Err(err).Msg("Failed to add motion")
			return nil, err
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
func (sms *SyndicateMotionSystem) getOngoingMotion(id string) (*SyndicateMotion, error) {
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

// addMotion check motion duplicated content and append new motion to motion system
func (sms *SyndicateMotionSystem) addMotion(bsm *boiler.SyndicateMotion) error {
	// add motion to system
	sms.Lock()
	defer sms.Unlock()

	// return if syndicate is closed
	if sms.isClosed.Load() {
		return terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	// insert syndicate motion if it does not have id
	if bsm.ID == "" {
		err := bsm.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Failed to insert syndicate motion")
		}
	}

	m := &SyndicateMotion{
		syndicate:       sms.syndicate,
		SyndicateMotion: bsm,
		isClosed:        atomic.Bool{},
		forceClosed:     atomic.Bool{},
	}

	// function for removing motion from motion system map
	m.isClosed.Store(false)
	m.forceClosed.Store(false)
	m.onClose = func() {
		sms.Lock()
		defer sms.Unlock()

		delete(sms.ongoingMotions, m.ID)
	}

	// TODO: check total vote member, instant pass if vote member less than 3

	// add motion to the map
	sms.ongoingMotions[m.ID] = m

	// spin up motion
	go m.start()

	return nil
}

// motionValidCheck check incoming motion has valid input, and generate a new boiler motion for voting
func (sms *SyndicateMotionSystem) motionValidCheck(user *boiler.Player, bsm *boiler.SyndicateMotion) (*boiler.SyndicateMotion, error) {
	// record old syndicate info to the motion
	s := sms.syndicate.load()

	if bsm.Reason == "" {
		return nil, terror.Error(fmt.Errorf("missing motion reason"), "Missing motion reason")
	}
	motion := &boiler.SyndicateMotion{
		SyndicateID: user.SyndicateID.String,
		IssuedByID:  user.ID,
		Type:        bsm.Type,
		Reason:      bsm.Reason,
	}
	// start issue motion
	switch motion.Type {
	case boiler.SyndicateMotionTypeCHANGE_GENERAL_DETAIL:
		if !bsm.NewSyndicateName.Valid && !bsm.NewSymbolID.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		// change symbol, name
		if bsm.NewSyndicateName.Valid {
			motion.NewSyndicateName = bsm.NewSyndicateName
			motion.OldSyndicateName = null.StringFrom(s.Name)
		}
		if bsm.NewSymbolID.Valid {
			motion.NewSymbolID = bsm.NewSymbolID
			motion.OldSymbolID = null.StringFrom(s.SymbolID)
		}

		motion.NewNamingConvention = bsm.NewNamingConvention
		motion.OldNamingConvention = s.NamingConvention

	case boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE:
		if !motion.NewJoinFee.Valid && !motion.NewExitFee.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		joinFeeAfterChanged := s.JoinFee
		// change
		if bsm.NewJoinFee.Valid {
			if bsm.NewJoinFee.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("join fee cannot less than zero"), "Join fee cannot less than zero.")
			}
			motion.NewJoinFee = bsm.NewJoinFee
			motion.OldJoinFee = decimal.NewNullDecimal(s.JoinFee)

			joinFeeAfterChanged = bsm.NewJoinFee.Decimal
		}

		exitFeeAfterChange := s.ExitFee
		if bsm.NewExitFee.Valid {
			// exit fee cannot less than zero
			if bsm.NewExitFee.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("exit fee cannot less than zero"), "Exit fee cannot less than zero.")
			}
			motion.NewExitFee = bsm.NewExitFee
			motion.OldExitFee = decimal.NewNullDecimal(s.ExitFee)

			exitFeeAfterChange = bsm.NewExitFee.Decimal
		}

		if exitFeeAfterChange.GreaterThan(joinFeeAfterChanged) {
			return nil, terror.Error(fmt.Errorf("exit fee cannot be higher than join fee"), "Exit fee cannot be higher than join fee.")
		}

	case boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
		if !motion.NewDeployingMemberCutPercentage.Valid &&
			!motion.NewMemberAssistCutPercentage.Valid &&
			!motion.NewMechOwnerCutPercentage.Valid &&
			!motion.NewSyndicateCutPercentage.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		deployMemberCutAfterChange := s.DeployingMemberCutPercentage
		if bsm.NewDeployingMemberCutPercentage.Valid {
			// exit fee cannot less than zero
			if bsm.NewDeployingMemberCutPercentage.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("mech deploying member cut cannot less than zero"), "Mech deploying member cut cannot less than zero.")
			}
			motion.NewDeployingMemberCutPercentage = bsm.NewDeployingMemberCutPercentage
			motion.OldDeployingMemberCutPercentage = decimal.NewNullDecimal(s.DeployingMemberCutPercentage)
			deployMemberCutAfterChange = bsm.NewDeployingMemberCutPercentage.Decimal
		}

		memberAssistCutAfterChange := s.MemberAssistCutPercentage
		if bsm.NewMemberAssistCutPercentage.Valid {
			// exit fee cannot less than zero
			if bsm.NewMemberAssistCutPercentage.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("member assistance cut cannot less than zero"), "Member assistance cut cannot less than zero.")
			}

			motion.NewMemberAssistCutPercentage = bsm.NewMemberAssistCutPercentage
			motion.OldMemberAssistCutPercentage = decimal.NewNullDecimal(s.MemberAssistCutPercentage)
			memberAssistCutAfterChange = bsm.NewMemberAssistCutPercentage.Decimal
		}

		mechOwnerCutAfterChange := s.MechOwnerCutPercentage
		if bsm.NewMechOwnerCutPercentage.Valid {
			// exit fee cannot less than zero
			if bsm.NewMechOwnerCutPercentage.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("mech owner cut cannot less than zero"), "Mech owner cut cannot less than zero.")
			}

			motion.NewMechOwnerCutPercentage = bsm.NewMechOwnerCutPercentage
			motion.OldMechOwnerCutPercentage = decimal.NewNullDecimal(s.MechOwnerCutPercentage)
			mechOwnerCutAfterChange = bsm.NewMechOwnerCutPercentage.Decimal
		}

		syndicateCutAfterChange := s.SyndicateCutPercentage
		if bsm.NewSyndicateCutPercentage.Valid {
			// exit fee cannot less than zero
			if bsm.NewSyndicateCutPercentage.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("syndicate cut cannot less than zero"), "Syndicate cut cannot less than zero.")
			}

			motion.NewSyndicateCutPercentage = bsm.NewSyndicateCutPercentage
			motion.OldSyndicateCutPercentage = decimal.NewNullDecimal(s.SyndicateCutPercentage)

			syndicateCutAfterChange = bsm.NewSyndicateCutPercentage.Decimal
		}

		if decimal.NewFromInt(100).LessThan(deployMemberCutAfterChange.Add(memberAssistCutAfterChange).Add(mechOwnerCutAfterChange).Add(syndicateCutAfterChange)) {
			return nil, terror.Error(fmt.Errorf("percentage over 100"), "Total percentage cannot be over 100")
		}

	case boiler.SyndicateMotionTypeADD_RULE:
		if !bsm.NewRuleContent.Valid {
			return nil, terror.Error(fmt.Errorf("rule content is not provided"), "Rule content is not provided.")
		}
		motion.NewRuleContent = bsm.NewRuleContent
		motion.NewRuleNumber = bsm.NewRuleNumber

	case boiler.SyndicateMotionTypeREMOVE_RULE:
		if !bsm.RuleID.Valid {
			return nil, terror.Error(fmt.Errorf("missing rule id"), "Missing rule id")
		}
		_, err := boiler.FindSyndicateRule(gamedb.StdConn, bsm.RuleID.String)
		if err != nil {
			return nil, terror.Error(err, "Syndicate rule does not exist")
		}
		motion.RuleID = bsm.RuleID

	case boiler.SyndicateMotionTypeCHANGE_RULE:
		if !bsm.RuleID.Valid {
			return nil, terror.Error(fmt.Errorf("missing rule id"), "Missing rule id")
		}
		_, err := boiler.FindSyndicateRule(gamedb.StdConn, bsm.RuleID.String)
		if err != nil {
			return nil, terror.Error(err, "Syndicate rule does not exist")
		}

		if !bsm.NewRuleNumber.Valid && !bsm.NewRuleContent.Valid {
			return nil, terror.Error(fmt.Errorf("missing rule change"), "Changing content is not provided.")
		}

		motion.NewRuleNumber = bsm.NewRuleNumber
		motion.NewRuleContent = bsm.NewRuleContent

	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
		if sms.syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return nil, terror.Error(fmt.Errorf("only corporation syndicate can appoint director"), "Only corporation syndicate can appoint director.")
		}
		if !bsm.DirectorID.Valid {
			return nil, terror.Error(fmt.Errorf("missing player id"), "Missing player id")
		}
		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.DirectorID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.DirectorID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if !player.SyndicateID.Valid || player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("not syndicate memeber"), "Player is not a member of the syndicate")
		}

		if player.DirectorOfSyndicateID.Valid {
			return nil, terror.Error(fmt.Errorf("already a director"), "Player is already a director of the syndicate")
		}

		motion.DirectorID = bsm.DirectorID

	case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
		if sms.syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return nil, terror.Error(fmt.Errorf("only corporation syndicate can appoint director"), "Only corporation syndicate can appoint director.")
		}
		if !bsm.DirectorID.Valid {
			return nil, terror.Error(fmt.Errorf("missing player id"), "Missing player id")
		}

		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.DirectorID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.DirectorID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("not syndicate memeber"), "Player is not a member of the syndicate")
		}

		if !player.DirectorOfSyndicateID.Valid {
			return nil, terror.Error(fmt.Errorf("not a director"), "Player is not a director of the syndicate")
		}
		motion.DirectorID = bsm.DirectorID

	case boiler.SyndicateMotionTypeREMOVE_FOUNDER:
		if sms.syndicate.HonoraryFounder {
			return nil, terror.Error(fmt.Errorf("syndicate founder has no power"), "The syndicate founder has no power")
		}
	default:
		gamelog.L.Debug().Str("motion type", bsm.Type).Msg("Invalid motion type")
		return nil, terror.Error(fmt.Errorf("invalid motion type"), "Invalid motion type")
	}

	return motion, nil
}

func (sms *SyndicateMotionSystem) duplicatedMotionCheck(bsm *boiler.SyndicateMotion) error {
	sms.RLock()
	defer sms.RUnlock()

	// return if syndicate is closed
	if sms.isClosed.Load() {
		return terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	if bsm.ActualEndedAt.Valid || bsm.Result.Valid {
		return terror.Error(fmt.Errorf("motion is already endded"), "Motion is already ended.")
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
			if bsm.NewSyndicateName.Valid && om.NewSyndicateName.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate name.")
			}
			if bsm.NewSymbolID.Valid && om.NewSymbolID.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate symbol.")
			}
			if bsm.NewNamingConvention.Valid && om.NewNamingConvention.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing syndicate naming convention.")
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
		case boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing entry fee.")
		case boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing battle win cut.")
		case boiler.SyndicateMotionTypeREMOVE_FOUNDER:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for removing the founder.")
		}
	}

	return nil
}

type SyndicateMotion struct {
	*boiler.SyndicateMotion
	syndicate *Syndicate
	sync.Mutex

	isClosed    atomic.Bool
	forceClosed atomic.Bool
	passCheck   func() bool
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

		// calculate result
		sm.parseResult()

		return
	}
}

// Vote check motion is closed or not before firing the function logic
func (sm *SyndicateMotion) vote(user *boiler.Player, isAgreed bool) error {
	sm.Lock()
	defer sm.Unlock()

	// only fire the function when motion is still open
	if sm.isClosed.Load() || sm.forceClosed.Load() {
		return terror.Error(fmt.Errorf("motion is closed"), "Motion is closed")
	}

	// check already exists
	mv, err := boiler.SyndicateMotionVotes(
		boiler.SyndicateMotionVoteWhere.MotionID.EQ(sm.ID),
		boiler.SyndicateMotionVoteWhere.VoteByID.EQ(user.ID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("motion id", sm.ID).Str("user id", user.ID).Msg("Failed to vote the motion")
		return terror.Error(err, "Failed to vote the motion")
	}

	if mv != nil {
		return terror.Error(fmt.Errorf("player already voted"), "You have already voted.")
	}

	// log vote to db
	mv = &boiler.SyndicateMotionVote{
		MotionID: sm.ID,
		VoteByID: user.ID,
		IsAgreed: isAgreed,
	}
	err = mv.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("motion vote", mv).Err(err).Msg("Failed to insert motion vote")
		return terror.Error(err, "Failed to vote the motion")
	}

	query := boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(sm.SyndicateID))
	if sm.syndicate.Type == boiler.SyndicateTypeCORPORATION {
		query = boiler.PlayerWhere.DirectorOfSyndicateID.EQ(null.StringFrom(sm.SyndicateID))
	}
	totalVoteCount, err := boiler.Players(query).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Interface("query", query).Err(err).Msg("Failed to total syndicate users")
		return nil
	}
	currentVoteCount, err := boiler.SyndicateMotionVotes(
		boiler.SyndicateMotionVoteWhere.MotionID.EQ(sm.ID),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate motion id", sm.ID).Err(err).Msg("Failed to current voted users")
		return nil
	}

	// close the vote, if only three people can vote or all the player have voted
	if totalVoteCount <= 3 || totalVoteCount == currentVoteCount {
		sm.isClosed.Store(true)
		return nil
	}

	return nil
}

func (sm *SyndicateMotion) parseResult() {
	sm.Lock()
	defer sm.Unlock()
	sm.isClosed.Store(true)

	sm.ActualEndedAt = null.TimeFrom(time.Now())

	// only update result, if force closed
	if sm.forceClosed.Load() {
		sm.Result = null.StringFrom(boiler.SyndicateMotionResultFORCE_CLOSED)
		// update motion result and actual end time
		_, err := sm.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateMotionColumns.Result, boiler.SyndicateMotionColumns.ActualEndedAt))
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to update motion result")
		}
		return
	}

	votes, err := boiler.SyndicateMotionVotes(
		boiler.SyndicateMotionVoteWhere.MotionID.EQ(sm.ID),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("motion id", sm.ID).Err(err).Msg("Failed to get all the votes from syndicate motion")
		return
	}

	agreedCount := 0
	disagreedCount := 0

	for _, v := range votes {
		if v.IsAgreed {
			agreedCount += 1
		} else {
			disagreedCount += 1
		}
	}

	// calculate pass rate
	totalVote := decimal.NewFromInt(int64(agreedCount + disagreedCount))

	// check whether vote is failed, if more than three members voted
	if totalVote.GreaterThan(decimal.NewFromInt(3)) {
		rate := decimal.NewFromInt(int64(agreedCount * 100)).Div(totalVote)
		if (rate.LessThanOrEqual(decimal.NewFromInt(50))) || (sm.Type == boiler.SyndicateMotionTypeREMOVE_FOUNDER && rate.LessThanOrEqual(decimal.NewFromInt(80))) {
			sm.Result = null.StringFrom(boiler.SyndicateMotionResultFAILED)
			// update motion result and actual end time
			_, err := sm.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateMotionColumns.Result, boiler.SyndicateMotionColumns.ActualEndedAt))
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to update motion result")
			}

			return
		}
	}

	sm.Result = null.StringFrom(boiler.SyndicateMotionResultPASSED)
	// update motion result and actual end time
	_, err = sm.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateMotionColumns.Result, boiler.SyndicateMotionColumns.ActualEndedAt))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update motion result")
	}

	// process action
	switch sm.Type {
	case boiler.SyndicateMotionTypeCHANGE_GENERAL_DETAIL,
		boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE,
		boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
		sm.updateSyndicate()
	case boiler.SyndicateMotionTypeADD_RULE:
		sm.addRule()
	case boiler.SyndicateMotionTypeREMOVE_RULE:
		sm.removeRule()
	case boiler.SyndicateMotionTypeCHANGE_RULE:
		sm.changeRule()
	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
		sm.appointDirector()
	case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
		sm.removeDirector()
	case boiler.SyndicateMotionTypeREMOVE_FOUNDER:
		sm.removeFounder()
	}
}

// broadcastUpdatedSyndicate broadcast the latest detail of the syndicate
func (sm *SyndicateMotion) broadcastUpdatedSyndicate() {
	s, err := boiler.FindSyndicate(gamedb.StdConn, sm.SyndicateID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load updated syndicate")
		return
	}

	sm.syndicate.store(s)

	ws.PublishMessage("", "", s)
}

// broadcastUpdateRules broadcast the latest rule list
func (sm *SyndicateMotion) broadcastUpdateRules() {
	rules, err := boiler.SyndicateRules(
		boiler.SyndicateRuleWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateRuleWhere.DeletedAt.IsNull(),
		qm.OrderBy(boiler.SyndicateRuleColumns.Number),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load new rule list")
		return
	}

	ws.PublishMessage("", "", rules)
}

// updateGeneralDetail update syndicate's join fee, exit fee and battle win cut percentage
func (sm *SyndicateMotion) updateSyndicate() {
	syndicate := &boiler.Syndicate{
		ID: sm.syndicate.ID,
	}
	updateCols := []string{}
	if sm.NewSyndicateName.Valid {
		syndicate.Name = sm.NewSyndicateName.String
		updateCols = append(updateCols, boiler.SyndicateColumns.Name)
	}
	if sm.NewSymbolID.Valid {
		syndicate.SymbolID = sm.NewSymbolID.String
		updateCols = append(updateCols, boiler.SyndicateColumns.SymbolID)
	}
	if sm.NewNamingConvention.Valid {
		syndicate.NamingConvention = sm.NewNamingConvention
		updateCols = append(updateCols, boiler.SyndicateColumns.NamingConvention)
	}
	if sm.NewJoinFee.Valid {
		syndicate.JoinFee = sm.NewJoinFee.Decimal
		updateCols = append(updateCols, boiler.SyndicateColumns.JoinFee)
	}
	if sm.NewExitFee.Valid {
		syndicate.ExitFee = sm.NewExitFee.Decimal
		updateCols = append(updateCols, boiler.SyndicateColumns.ExitFee)
	}
	if sm.NewDeployingMemberCutPercentage.Valid {
		syndicate.DeployingMemberCutPercentage = sm.NewDeployingMemberCutPercentage.Decimal
		updateCols = append(updateCols, boiler.SyndicateColumns.DeployingMemberCutPercentage)
	}
	if sm.NewMemberAssistCutPercentage.Valid {
		syndicate.MemberAssistCutPercentage = sm.NewMemberAssistCutPercentage.Decimal
		updateCols = append(updateCols, boiler.SyndicateColumns.MemberAssistCutPercentage)
	}
	if sm.NewMechOwnerCutPercentage.Valid {
		syndicate.MechOwnerCutPercentage = sm.NewMechOwnerCutPercentage.Decimal
		updateCols = append(updateCols, boiler.SyndicateColumns.MechOwnerCutPercentage)
	}
	if sm.NewSyndicateCutPercentage.Valid {
		syndicate.SyndicateCutPercentage = sm.NewSyndicateCutPercentage.Decimal
		updateCols = append(updateCols, boiler.SyndicateColumns.SyndicateCutPercentage)
	}

	_, err := syndicate.Update(gamedb.StdConn, boil.Whitelist(updateCols...))
	if err != nil {
		gamelog.L.Error().Interface("motion", sm.SyndicateMotion).Interface("updated syndicate", syndicate).Strs("updated column", updateCols).Err(err).Msg("Failed to update syndicate from motion")
		return
	}

	sm.broadcastUpdatedSyndicate()
}

// addRule add a new rule to db, and shift existing rules
func (sm *SyndicateMotion) addRule() {
	// get rules
	rules, err := boiler.SyndicateRules(
		boiler.SyndicateRuleWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateRuleWhere.DeletedAt.IsNull(),
		qm.OrderBy(boiler.SyndicateRuleColumns.Number),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to load rules")
		return
	}

	r := boiler.SyndicateRule{
		SyndicateID: sm.SyndicateID,
		Number:      len(rules) + 1,
		Content:     sm.NewRuleContent.String,
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return
	}

	defer tx.Rollback()

	// update rule number
	if sm.NewRuleNumber.Valid {
		r.Number = sm.NewRuleNumber.Int

		q := `
			UPDATE 
				syndicate_rules
			SET
				number = number + 1
			WHERE
				syndicate_id = $1 AND number >= $2 AND deleted_at ISNULL;
		`

		_, err = tx.Exec(q, sm.SyndicateID, sm.NewRuleNumber.Int)
		if err != nil {
			gamelog.L.Error().
				Str("query", q).
				Str("syndicate id", sm.SyndicateID).
				Int("rule number", sm.NewRuleNumber.Int).
				Err(err).Msg("Failed to update rules' number")
			return
		}
	}

	err = r.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert new rule")
		return
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return
	}

	sm.broadcastUpdateRules()
}

func (sm *SyndicateMotion) removeRule() {
	// get rule
	rule, err := boiler.FindSyndicateRule(gamedb.StdConn, sm.RuleID.String)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to load rules")
		return
	}

	rule.DeletedAt = null.TimeFrom(time.Now())

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return
	}

	defer tx.Rollback()

	// update rule number
	q := `
		UPDATE 
			syndicate_rules
		SET
			number = number - 1
		WHERE
			syndicate_id = $1 AND number >= $2 AND deleted_at ISNULL;
	`

	_, err = tx.Exec(q, sm.SyndicateID, rule.Number)
	if err != nil {
		gamelog.L.Error().
			Str("query", q).
			Str("syndicate id", sm.SyndicateID).
			Int("rule number", sm.NewRuleNumber.Int).
			Err(err).Msg("Failed to update rules' number")
		return
	}

	_, err = rule.Update(tx, boil.Whitelist(boiler.SyndicateRuleColumns.DeletedAt))
	if err != nil {
		gamelog.L.Error().Str("rule id", rule.ID).Err(err).Msg("Failed to remove rule")
		return
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return
	}

	sm.broadcastUpdateRules()
}

func (sm *SyndicateMotion) changeRule() {
	rule, err := boiler.FindSyndicateRule(gamedb.StdConn, sm.RuleID.String)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to load rules")
		return
	}

	if sm.NewRuleContent.Valid {
		rule.Content = sm.NewRuleContent.String
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return
	}

	defer tx.Rollback()

	// implement number shifting
	if sm.NewRuleNumber.Valid && sm.NewRuleNumber.Int != rule.Number {
		if sm.NewRuleNumber.Int > rule.Number {
			q := `
				UPDATE 
					syndicate_rules
				SET
					number = number - 1
				WHERE
					syndicate_id = $1 AND number > $2 AND number <= $3 AND deleted_at ISNULL;
			`
			_, err = tx.Exec(q, sm.SyndicateID, rule.Number, sm.NewRuleNumber.Int)
			if err != nil {
				gamelog.L.Error().
					Str("query", q).
					Str("syndicate id", sm.SyndicateID).
					Int("$2", rule.Number).
					Int("$3", sm.NewRuleNumber.Int).
					Err(err).Msg("Failed to shift rules' number")
				return
			}
		} else {
			q := `
				UPDATE 
					syndicate_rules
				SET
					number = number + 1
				WHERE
					syndicate_id = $1 AND number >= $2 AND number < $3 AND deleted_at ISNULL;
			`
			_, err = tx.Exec(q, sm.SyndicateID, sm.NewRuleNumber.Int, rule.Number)
			if err != nil {
				gamelog.L.Error().
					Str("query", q).
					Str("syndicate id", sm.SyndicateID).
					Int("$2", sm.NewRuleNumber.Int).
					Int("$3", rule.Number).
					Err(err).Msg("Failed to shift rules' number")
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
			return
		}

		rule.Number = sm.NewRuleNumber.Int
	}

	_, err = rule.Update(tx, boil.Whitelist(boiler.SyndicateRuleColumns.Content, boiler.SyndicateRuleColumns.Number))
	if err != nil {
		gamelog.L.Error().Interface("rule", rule).Err(err).Msg("Failed to update rules")
		return
	}
	sm.broadcastUpdateRules()
}

func (sm *SyndicateMotion) appointDirector() {
	player, err := boiler.Players(
		boiler.PlayerWhere.ID.EQ(sm.DirectorID.String),
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(sm.SyndicateID)),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.DirectorID.String).Err(err).Msg("Failed to get player")
		return
	}

	player.DirectorOfSyndicateID = null.StringFrom(sm.SyndicateID)
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.DirectorOfSyndicateID))
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.DirectorID.String).Err(err).Msg("Failed to appoint player to director")
		return
	}

	// broadcast data
	ws.PublishMessage(fmt.Sprintf("/user/%s", sm.DirectorID.String), HubKeyUserSubscribe, player)

	// broadcast syndicate director list
}

func (sm *SyndicateMotion) removeDirector() {
	player, err := boiler.Players(
		boiler.PlayerWhere.ID.EQ(sm.DirectorID.String),
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(sm.SyndicateID)),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.DirectorID.String).Err(err).Msg("Failed to get player")
		return
	}

	player.DirectorOfSyndicateID = null.StringFromPtr(nil)
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.DirectorOfSyndicateID))
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.DirectorID.String).Err(err).Msg("Failed to remove player from syndicate director list")
		return
	}

	// broadcast data
	ws.PublishMessage(fmt.Sprintf("/user/%s", sm.DirectorID.String), HubKeyUserSubscribe, player)

	// broadcast syndicate director list
}
func (sm *SyndicateMotion) removeFounder() {
	syndicate := boiler.Syndicate{
		ID:              sm.SyndicateID,
		HonoraryFounder: true,
	}

	_, err := syndicate.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateColumns.HonoraryFounder))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update syndicate honorary founder")
		return
	}

	sm.broadcastUpdatedSyndicate()
}
