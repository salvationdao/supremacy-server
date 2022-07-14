package syndicate

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"
)

type MotionSystem struct {
	syndicate *Syndicate

	ongoingMotions map[string]*Motion
	sync.RWMutex

	isClosed atomic.Bool
}

// newMotionSystem generate a new syndicate motion system
func (s *Syndicate) newMotionSystem() (*MotionSystem, error) {
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

	sms := &MotionSystem{
		syndicate:      s,
		ongoingMotions: map[string]*Motion{},
		isClosed:       atomic.Bool{},
	}
	sms.isClosed.Store(false)

	for _, m := range ms {
		err = sms.addMotion(m, false)
		if err != nil {
			// the duplicate check failed, so log error instead of return the system
			gamelog.L.Error().Err(err).Msg("Failed to add motion")
			return nil, err
		}
	}

	return sms, nil
}

func (sms *MotionSystem) terminated() {
	sms.Lock()
	defer sms.Unlock()
	sms.isClosed.Store(true)

	// force close all the motion
	for _, om := range sms.ongoingMotions {
		om.forceClosed.Store(true) // close motion without calculate result
		om.isClosed.Store(true)    // terminate all the processing motion
	}
}

func (sms *MotionSystem) getOngoingMotionList() ([]*boiler.SyndicateMotion, error) {
	sms.RLock()
	defer sms.RUnlock()
	// return if syndicate is closed
	if sms.isClosed.Load() {
		return nil, terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	motions := []*boiler.SyndicateMotion{}
	for _, om := range sms.ongoingMotions {
		if om.isClosed.Load() {
			continue
		}
		motions = append(motions, om.SyndicateMotion)
	}

	return motions, nil
}

// GetOngoingMotion return ongoing motion by id
func (sms *MotionSystem) getOngoingMotion(id string) (*Motion, error) {
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
func (sms *MotionSystem) addMotion(bsm *boiler.SyndicateMotion, isNewMotion bool) error {
	// add motion to system
	sms.Lock()
	defer sms.Unlock()

	// return if syndicate is closed
	if sms.isClosed.Load() {
		return terror.Error(fmt.Errorf("syndicate is closed"), "Syndicate is closed")
	}

	// check whether the motion is new
	if isNewMotion {
		// check duplicate ongoing motion
		err := sms.duplicatedMotionCheck(bsm)
		if err != nil {
			return err
		}

		// insert, if motion is not in the database
		if bsm.ID == "" {
			err = bsm.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return terror.Error(err, "Failed to insert syndicate motion")
			}
		}
	}

	// syndicate motion
	m := &Motion{
		SyndicateMotion: bsm,
		syndicate:       sms.syndicate,
		isClosed:        atomic.Bool{},
		forceClosed:     atomic.Bool{},
		onClose: func() {
			sms.Lock()
			defer sms.Unlock()

			// clean up motion from the list
			if _, ok := sms.ongoingMotions[bsm.ID]; ok {
				delete(sms.ongoingMotions, bsm.ID)

				// broadcast ongoing motion list
				ws.PublishMessage("", "", sms.ongoingMotions)
			}
		},
	}
	m.isClosed.Store(false)
	m.forceClosed.Store(false)

	totalAvailableVoter, err := sms.syndicate.getTotalAvailableMotionVoter()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to total syndicate users")
		return err
	}

	// instantly pass, if total available voting player is no more than 3
	if totalAvailableVoter <= 3 {
		m.parseResult()
		return nil
	}

	// add motion to the map
	sms.ongoingMotions[m.ID] = m

	// spin up motion
	go m.start()

	// broadcast ongoing motions
	ws.PublishMessage("", "", sms.ongoingMotions)

	return nil
}

// validateIncomingMotion check incoming motion has valid input, and generate a new boiler motion for voting
func (sms *MotionSystem) validateIncomingMotion(user *boiler.Player, bsm *boiler.SyndicateMotion) (*boiler.SyndicateMotion, error) {
	// validate motion
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
		if !bsm.NewSyndicateName.Valid && !bsm.NewSymbol.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		// change symbol, name
		if bsm.NewSyndicateName.Valid {
			// verify syndicate name
			newName, err := sms.syndicate.system.SyndicateNameVerification(bsm.NewSyndicateName.String)
			if err != nil {
				return nil, err
			}
			motion.NewSyndicateName = null.StringFrom(newName)
		}
		if bsm.NewSymbol.Valid {
			// verify symbol
			newSymbol, err := sms.syndicate.system.SyndicateNameVerification(bsm.NewSymbol.String)
			if err != nil {
				return nil, err
			}
			motion.NewSymbol = null.StringFrom(newSymbol)
		}
	case boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE:
		if !bsm.NewJoinFee.Valid && !bsm.NewExitFee.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		s, err := boiler.FindSyndicate(gamedb.StdConn, sms.syndicate.ID)
		if err != nil {
			return nil, terror.Error(err, "Failed to load syndicate detail")
		}

		joinFeeAfterChanged := s.JoinFee
		// change
		if bsm.NewJoinFee.Valid {
			if bsm.NewJoinFee.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("join fee cannot less than zero"), "Join fee cannot less than zero.")
			}
			motion.NewJoinFee = bsm.NewJoinFee

			joinFeeAfterChanged = bsm.NewJoinFee.Decimal
		}

		exitFeeAfterChange := s.ExitFee
		if bsm.NewExitFee.Valid {
			// exit fee cannot less than zero
			if bsm.NewExitFee.Decimal.LessThan(decimal.Zero) {
				return nil, terror.Error(fmt.Errorf("exit fee cannot less than zero"), "Exit fee cannot less than zero.")
			}
			motion.NewExitFee = bsm.NewExitFee

			exitFeeAfterChange = bsm.NewExitFee.Decimal
		}

		if exitFeeAfterChange.GreaterThan(joinFeeAfterChanged) {
			return nil, terror.Error(fmt.Errorf("exit fee cannot be higher than join fee"), "Exit fee cannot be higher than join fee.")
		}

	case boiler.SyndicateMotionTypeCHANGE_MONTHLY_DUES:
		if !bsm.NewMonthlyDues.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		motion.NewMonthlyDues = bsm.NewMonthlyDues

	case boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
		if !bsm.NewDeployingMemberCutPercentage.Valid &&
			!bsm.NewMemberAssistCutPercentage.Valid &&
			!bsm.NewMechOwnerCutPercentage.Valid &&
			!bsm.NewSyndicateCutPercentage.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		s, err := boiler.FindSyndicate(gamedb.StdConn, sms.syndicate.ID)
		if err != nil {
			return nil, terror.Error(err, "Failed to load syndicate detail")
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
			return nil, terror.Error(fmt.Errorf("not corporation syndicate"), "Only corporation syndicate can appoint director.")
		}

		isDirector, err := sms.syndicate.isDirector(user.ID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("no a director of syndicate"), "Only the directors of the syndicate can issue appoint director motion.")
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

		isDirector, err = sms.syndicate.isDirector(player.ID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("not a director"), fmt.Sprintf("%s is already a director of the syndicate", player.Username.String))
		}

		motion.DirectorID = bsm.DirectorID

	case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
		if sms.syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return nil, terror.Error(fmt.Errorf("not corporation syndicate"), "Only corporation syndicate can appoint director.")
		}

		isDirector, err := sms.syndicate.isDirector(user.ID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("no a director of syndicate"), "Only the directors of the syndicate can issue remove director motion.")
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

		isDirector, err = sms.syndicate.isDirector(player.ID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("not a director"), fmt.Sprintf("%s is not a director of the syndicate", player.Username.String))
		}

		motion.DirectorID = bsm.DirectorID

	default:
		gamelog.L.Debug().Str("motion type", bsm.Type).Msg("Invalid motion type")
		return nil, terror.Error(fmt.Errorf("invalid motion type"), "Invalid motion type")
	}

	return motion, nil
}

func (sms *MotionSystem) duplicatedMotionCheck(bsm *boiler.SyndicateMotion) error {
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
			if bsm.NewSymbol.Valid && om.NewSymbol.Valid {
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
		case boiler.SyndicateMotionTypeCHANGE_MONTHLY_DUES:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for member monthly dues.")
		case boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing entry fee.")
		case boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is an ongoing motion for changing battle win cut.")
		}
	}

	return nil
}

type Motion struct {
	*boiler.SyndicateMotion
	syndicate *Syndicate
	sync.Mutex

	isClosed    atomic.Bool
	forceClosed atomic.Bool
	onClose     func()
}

func (sm *Motion) start() {
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
func (sm *Motion) vote(user *boiler.Player, isAgreed bool) error {
	// check vote right
	switch sm.Type {
	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR, boiler.SyndicateMotionTypeREMOVE_DIRECTOR, boiler.SyndicateMotionTypeDEPOSE_CEO, boiler.SyndicateMotionTypeCEO_ELECTION:
		isDirector, err := sm.syndicate.isDirector(user.ID)
		if err != nil {
			return terror.Error(err, "Failed to get check syndicate director")
		}

		if !isDirector {
			return terror.Error(fmt.Errorf("no a director"), "This motion is for the board of directors only.")
		}
	}

	sm.Lock()
	defer sm.Unlock()

	// only fire the function when motion is still open
	if sm.isClosed.Load() || sm.forceClosed.Load() {
		return terror.Error(fmt.Errorf("motion is closed"), "Motion is closed")
	}

	// check already voted
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

	totalVoteCount, err := boiler.Players(
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(sm.SyndicateID)),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to total syndicate users")
		return nil
	}
	currentVoteCount, err := boiler.SyndicateMotionVotes(
		boiler.SyndicateMotionVoteWhere.MotionID.EQ(sm.ID),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate motion id", sm.ID).Err(err).Msg("Failed to current voted users")
		return nil
	}

	// close the vote, if only all the player have voted
	if totalVoteCount == currentVoteCount {
		sm.isClosed.Store(true)
		return nil
	}

	return nil
}

func (sm *Motion) parseResult() {
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
		if (rate.LessThanOrEqual(decimal.NewFromInt(50))) || (sm.Type == boiler.SyndicateMotionTypeDEPOSE_CEO && rate.LessThanOrEqual(decimal.NewFromInt(80))) {
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
		boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT,
		boiler.SyndicateMotionTypeCHANGE_MONTHLY_DUES:
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
	case boiler.SyndicateMotionTypeDEPOSE_CEO:
		sm.deposeCEO()
	}
}

// broadcastUpdatedSyndicate broadcast the latest detail of the syndicate
func (sm *Motion) broadcastUpdatedSyndicate() {
	s, err := boiler.Syndicates(
		boiler.SyndicateWhere.ID.EQ(sm.SyndicateID),
		qm.Load(boiler.SyndicateRels.Admin),
		qm.Load(boiler.SyndicateRels.CeoPlayer),
		qm.Load(boiler.SyndicateRels.FoundedBy),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load updated syndicate")
		return
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s", sm.syndicate.FactionID, sm.SyndicateID), server.HubKeySyndicateGeneralDetailSubscribe, server.SyndicateBoilerToServer(s))
}

// broadcastUpdateRules broadcast the latest rule list
func (sm *Motion) broadcastUpdateRules() {
	rules, err := boiler.SyndicateRules(
		boiler.SyndicateRuleWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateRuleWhere.DeletedAt.IsNull(),
		qm.OrderBy(boiler.SyndicateRuleColumns.Number),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load new rule list")
		return
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/rules", sm.syndicate.FactionID, sm.SyndicateID), server.HubKeySyndicateRulesSubscribe, rules)
}

// updateGeneralDetail update syndicate's join fee, exit fee and battle win cut percentage
func (sm *Motion) updateSyndicate() {
	// load syndicate
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, sm.SyndicateID)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to get syndicate from id")
		return
	}

	updateMotionCols := []string{}
	updateSyndicateCols := []string{}
	if sm.NewSyndicateName.Valid {
		sm.OldSyndicateName = null.StringFrom(syndicate.Name)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldSyndicateName)

		syndicate.Name = sm.NewSyndicateName.String
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.Name)
	}
	if sm.NewSymbol.Valid {
		sm.OldSymbol = null.StringFrom(syndicate.Symbol)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldSymbol)

		syndicate.Symbol = sm.NewSymbol.String
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.Symbol)
	}
	if sm.NewJoinFee.Valid {
		sm.OldJoinFee = decimal.NewNullDecimal(syndicate.JoinFee)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldJoinFee)

		syndicate.JoinFee = sm.NewJoinFee.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.JoinFee)
	}
	if sm.NewExitFee.Valid {
		sm.OldExitFee = decimal.NewNullDecimal(syndicate.ExitFee)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldExitFee)

		syndicate.ExitFee = sm.NewExitFee.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.ExitFee)
	}
	if sm.NewDeployingMemberCutPercentage.Valid {
		sm.OldDeployingMemberCutPercentage = decimal.NewNullDecimal(syndicate.DeployingMemberCutPercentage)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldDeployingMemberCutPercentage)

		syndicate.DeployingMemberCutPercentage = sm.NewDeployingMemberCutPercentage.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.DeployingMemberCutPercentage)
	}
	if sm.NewMemberAssistCutPercentage.Valid {
		sm.OldMemberAssistCutPercentage = decimal.NewNullDecimal(syndicate.MemberAssistCutPercentage)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldMemberAssistCutPercentage)

		syndicate.MemberAssistCutPercentage = sm.NewMemberAssistCutPercentage.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.MemberAssistCutPercentage)
	}
	if sm.NewMechOwnerCutPercentage.Valid {
		sm.OldMechOwnerCutPercentage = decimal.NewNullDecimal(syndicate.MechOwnerCutPercentage)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldMechOwnerCutPercentage)

		syndicate.MechOwnerCutPercentage = sm.NewMechOwnerCutPercentage.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.MechOwnerCutPercentage)
	}
	if sm.NewSyndicateCutPercentage.Valid {
		sm.OldSyndicateCutPercentage = decimal.NewNullDecimal(syndicate.SyndicateCutPercentage)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldSyndicateCutPercentage)

		syndicate.SyndicateCutPercentage = sm.NewSyndicateCutPercentage.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.SyndicateCutPercentage)
	}
	if sm.NewMonthlyDues.Valid {
		sm.OldMonthlyDues = decimal.NewNullDecimal(syndicate.MonthlyDues)
		updateMotionCols = append(updateMotionCols, boiler.SyndicateMotionColumns.OldMonthlyDues)

		syndicate.MonthlyDues = sm.NewMonthlyDues.Decimal
		updateSyndicateCols = append(updateSyndicateCols, boiler.SyndicateColumns.MonthlyDues)
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction")
		return
	}

	defer tx.Rollback()

	_, err = sm.Update(tx, boil.Whitelist(updateMotionCols...))
	if err != nil {
		gamelog.L.Error().Interface("updated motion", sm).Strs("updated column", updateMotionCols).Err(err).Msg("Failed to update motion columns")
		return
	}

	_, err = syndicate.Update(tx, boil.Whitelist(updateSyndicateCols...))
	if err != nil {
		gamelog.L.Error().Interface("updated syndicate", syndicate).Strs("updated column", updateSyndicateCols).Err(err).Msg("Failed to update syndicate from motion")
		return
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return
	}

	sm.broadcastUpdatedSyndicate()
}

// addRule add a new rule to db, and shift existing rules
func (sm *Motion) addRule() {
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

func (sm *Motion) removeRule() {
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

func (sm *Motion) changeRule() {
	updatedMotionCols := []string{}

	rule, err := boiler.FindSyndicateRule(gamedb.StdConn, sm.RuleID.String)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to load rules")
		return
	}

	if sm.NewRuleContent.Valid {
		sm.OldRuleContent = null.StringFrom(rule.Content)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldRuleContent)

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
		sm.OldRuleNumber = null.IntFrom(rule.Number)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldRuleNumber)

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

		rule.Number = sm.NewRuleNumber.Int
	}

	_, err = sm.Update(tx, boil.Whitelist(updatedMotionCols...))
	if err != nil {
		gamelog.L.Error().Interface("motion", sm).Err(err).Msg("Failed to update motion")
		return
	}

	_, err = rule.Update(tx, boil.Whitelist(boiler.SyndicateRuleColumns.Content, boiler.SyndicateRuleColumns.Number))
	if err != nil {
		gamelog.L.Error().Interface("rule", rule).Err(err).Msg("Failed to update rules")
		return
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return
	}
	sm.broadcastUpdateRules()
}

func (sm *Motion) appointDirector() {
	sd := boiler.SyndicateDirector{
		SyndicateID: sm.SyndicateID,
		PlayerID:    sm.DirectorID.String,
	}

	err := sd.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.DirectorID.String).Err(err).Msg("Failed to appoint player to director")
		return
	}

	// load new director list
	sds, err := db.GetSyndicateDirectors(sm.SyndicateID)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to get syndicate directors")
		return
	}

	// broadcast syndicate director list
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/directors", sm.syndicate.FactionID, sm.SyndicateID), server.HubKeySyndicateDirectorsSubscribe, sds)
}

func (sm *Motion) removeDirector() {
	_, err := boiler.SyndicateDirectors(
		boiler.SyndicateDirectorWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateDirectorWhere.PlayerID.EQ(sm.DirectorID.String),
	).DeleteAll(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.DirectorID.String).Err(err).Msg("Failed to remove player from syndicate director list")
		return
	}

	// load new director list
	sds, err := db.GetSyndicateDirectors(sm.SyndicateID)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to get syndicate directors")
		return
	}

	// broadcast syndicate director list
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/directors", sm.syndicate.FactionID, sm.SyndicateID), server.HubKeySyndicateDirectorsSubscribe, sds)
}

func (sm *Motion) deposeCEO() {
	_, err := boiler.Syndicates(
		boiler.SyndicateWhere.ID.EQ(sm.SyndicateID),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.SyndicateColumns.CeoPlayerID: null.StringFromPtr(nil)})
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", sm.SyndicateID).Msg("Failed to depose ceo of the syndicate.")
		return
	}

	sm.broadcastUpdatedSyndicate()
}
