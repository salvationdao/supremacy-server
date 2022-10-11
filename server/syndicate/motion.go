package syndicate

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/sasha-s/go-deadlock"
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
	deadlock.RWMutex

	isClosed atomic.Bool
}

// newMotionSystem generate a new syndicate motion system
func newMotionSystem(s *Syndicate) (*MotionSystem, error) {
	ms, err := boiler.SyndicateMotions(
		boiler.SyndicateMotionWhere.SyndicateID.EQ(s.ID),
		boiler.SyndicateMotionWhere.EndAt.GT(time.Now()),
		boiler.SyndicateMotionWhere.FinalisedAt.IsNull(),
		qm.Load(boiler.SyndicateMotionRels.NewLogo),
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
		err = sms.addMotion(m, m.R.NewLogo, false)
		if err != nil {
			// the duplicate check failed, so log error instead of return the system
			gamelog.L.Error().Err(err).Msg("Failed to add motion")
			return nil, err
		}
	}

	return sms, nil
}

func (sms *MotionSystem) terminate() {
	sms.Lock()
	defer sms.Unlock()
	sms.isClosed.Store(true)

	// force close all the motion
	for _, om := range sms.ongoingMotions {
		om.forceClosed.Store(boiler.SyndicateMotionResultTERMINATED) // close motion without calculate result
	}
}

// forceCloseTypes close a specific type of motion
func (sms *MotionSystem) forceCloseTypes(reason string, types ...string) {
	sms.Lock()
	defer sms.Unlock()

	for _, t := range types {
		for _, om := range sms.ongoingMotions {
			if om.Type == t {
				om.forceClosed.Store(boiler.SyndicateMotionResultTERMINATED)
			}
		}
	}
}

func (sms *MotionSystem) finaliseMotion(playerPosition string, motionID string, isAccepted bool) error {
	sms.Lock()
	defer sms.Unlock()

	if sms.isClosed.Load() {
		return terror.Error(fmt.Errorf("motion system is closed"), "Motion system is already closed.")
	}

	sm, ok := sms.ongoingMotions[motionID]
	if !ok {
		return terror.Error(fmt.Errorf("motion not exist"), "Motion does not exist")
	}

	if sm.isClosed.Load() {
		return terror.Error(fmt.Errorf("motion is closed"), "Motion is already closed.")
	}

	if sm.forceClosed.Load() != "" {
		return terror.Error(fmt.Errorf("motion is finalised"), "Motion is already finalised.")
	}

	// finalise motion
	decision := fmt.Sprintf("%s_ACCEPT", playerPosition)
	if !isAccepted {
		decision = fmt.Sprintf("%s_REJECT", playerPosition)
	}
	sm.forceClosed.Store(decision)

	return nil
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
func (sms *MotionSystem) addMotion(bsm *boiler.SyndicateMotion, logo *boiler.Blob, isNewMotion bool) error {
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

		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
			return terror.Error(err, "Failed to add motion to the system.")
		}

		defer func(tx *sql.Tx) {
			_ = tx.Rollback()
		}(tx)

		if logo != nil && logo.ID != "" {
			err = logo.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to insert logo into database")
				return terror.Error(err, "Failed to add motion to the system")
			}
		}

		// insert, if motion is not in the database
		if bsm.ID == "" {
			err = bsm.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to insert motion into db")
				return terror.Error(err, "Failed to add motion to the system")
			}
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
			return terror.Error(err, "Failed to add motion to the system")
		}
	}

	// syndicate motion
	m := &Motion{
		SyndicateMotion: bsm,
		syndicate:       sms.syndicate,
		isClosed:        atomic.Bool{},
		forceClosed:     atomic.String{},
		onClose: func() {
			sms.Lock()
			defer sms.Unlock()

			// clean up motion from the list
			if _, ok := sms.ongoingMotions[bsm.ID]; ok {
				delete(sms.ongoingMotions, bsm.ID)

				// broadcast ongoing motion list
				ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_motions", sms.syndicate.FactionID, bsm.SyndicateID), server.HubKeySyndicateOngoingMotionSubscribe, sms.ongoingMotions)
			}
		},
	}
	m.isClosed.Store(false)
	m.forceClosed.Store("")

	totalAvailableVoter, err := db.GetSyndicateTotalAvailableMotionVoters(sms.syndicate.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to total syndicate users")
		return err
	}

	// instantly pass, if total available voting player is no more than 3
	if totalAvailableVoter <= 3 {
		m.action()
		return nil
	}

	// add motion to the map
	sms.ongoingMotions[m.ID] = m

	// spin up motion
	go m.start()

	// broadcast ongoing motions
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/ongoing_motions", sms.syndicate.FactionID, bsm.SyndicateID), server.HubKeySyndicateOngoingMotionSubscribe, sms.ongoingMotions)

	return nil
}

// validateIncomingMotion check incoming motion has valid input, and generate a new boiler motion for voting
func (sms *MotionSystem) validateIncomingMotion(userID string, bsm *boiler.SyndicateMotion, logo *boiler.Blob) (*boiler.SyndicateMotion, error) {
	// validate motion
	if bsm.Reason == "" {
		return nil, terror.Error(fmt.Errorf("missing motion reason"), "Missing motion reason")
	}
	motion := &boiler.SyndicateMotion{
		SyndicateID: sms.syndicate.ID,
		IssuedByID:  userID,
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
		// change logo
		if logo != nil && logo.File.Valid && len(logo.File.Bytes) > 0 {
			avatarID := uuid.Must(uuid.NewV4())
			// set logo id, if image file is provided
			logo.ID = avatarID.String()
			bsm.NewLogoID = null.StringFrom(avatarID.String())
		}
	case boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE:
		if !bsm.NewJoinFee.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		// change
		if bsm.NewJoinFee.Decimal.LessThan(decimal.Zero) {
			return nil, terror.Error(fmt.Errorf("join fee cannot less than zero"), "Join fee cannot less than zero.")
		}
		motion.NewJoinFee = bsm.NewJoinFee

	case boiler.SyndicateMotionTypeCHANGE_MONTHLY_DUES:
		if !bsm.NewMonthlyDues.Valid {
			return nil, terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}

		if bsm.NewMonthlyDues.Decimal.LessThan(decimal.Zero) {
			return nil, terror.Error(fmt.Errorf("monthly dues cannot less than zero"), "Member monthly dues cannot less than zero.")
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

		// sum of the cut cannot be more than 100%
		if decimal.NewFromInt(100).LessThan(deployMemberCutAfterChange.Add(memberAssistCutAfterChange).Add(mechOwnerCutAfterChange)) {
			return nil, terror.Error(fmt.Errorf("percentage not 100"), "Total percentage should be 100%.")
		}

	case boiler.SyndicateMotionTypeADD_RULE:
		if !bsm.NewRuleContent.Valid {
			return nil, terror.Error(fmt.Errorf("rule content is not provided"), "Rule content is not provided.")
		}
		motion.NewRuleContent = bsm.NewRuleContent

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

	case boiler.SyndicateMotionTypeREMOVE_MEMBER:
		if !bsm.MemberID.Valid {
			return nil, terror.Error(fmt.Errorf("member id not provided"), "Member id is not provided")
		}

		if userID == bsm.MemberID.String {
			return nil, terror.Error(fmt.Errorf("cannot remove yourself"), "Cannot issue remove motion against yourself.")
		}

		// get issuer
		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.MemberID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.MemberID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if !player.SyndicateID.Valid || player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("player is not in your syndicate"), "Player is not a member of the syndicate")
		}

		isCommittee, err := boiler.SyndicateCommittees(
			boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sms.syndicate.ID),
			boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
		).Exists(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("syndicate id", sms.syndicate.ID).Str("player id", player.ID).Err(err).Msg("Failed to query syndicate committee.")
			return nil, terror.Error(err, "Failed to issue motion")
		}

		if isCommittee {
			return nil, terror.Error(fmt.Errorf("committee cannot be removed"), "Committee cannot be removed from syndicate")
		}

		if sms.syndicate.Type == boiler.SyndicateTypeCORPORATION {
			isDirector, err := boiler.SyndicateDirectors(
				boiler.SyndicateDirectorWhere.SyndicateID.EQ(sms.syndicate.ID),
				boiler.SyndicateDirectorWhere.PlayerID.EQ(player.ID),
			).Exists(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("syndicate id", sms.syndicate.ID).Str("player id", player.ID).Err(err).Msg("Failed to query syndicate director.")
				return nil, terror.Error(err, "Failed to issue motion")
			}

			if isDirector {
				return nil, terror.Error(fmt.Errorf("only board can remove directors"), "Directors can only be removed by the board")
			}
		}

		motion.MemberID = bsm.MemberID

	case boiler.SyndicateMotionTypeAPPOINT_COMMITTEE:
		if !bsm.MemberID.Valid {
			return nil, terror.Error(fmt.Errorf("member id not provided"), "Member id is not provided")
		}

		// get issuer
		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.MemberID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.MemberID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if !player.SyndicateID.Valid || player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("player is not in your syndicate"), "Player is not a member of the syndicate")
		}

		// check already a committee
		isCommittee, err := boiler.SyndicateCommittees(
			boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sms.syndicate.ID),
			boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
		).Exists(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("syndicate id", sms.syndicate.ID).Str("player id", player.ID).Err(err).Msg("Failed to query syndicate committee.")
			return nil, terror.Error(err, "Failed to issue motion")
		}

		if isCommittee {
			return nil, terror.Error(fmt.Errorf("already a committee"), "The player is already a committee of the syndicate")
		}

		// check already a director
		if sms.syndicate.Type == boiler.SyndicateTypeCORPORATION {
			isDirector, err := boiler.SyndicateDirectors(
				boiler.SyndicateDirectorWhere.SyndicateID.EQ(sms.syndicate.ID),
				boiler.SyndicateDirectorWhere.PlayerID.EQ(player.ID),
			).Exists(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("syndicate id", sms.syndicate.ID).Str("player id", player.ID).Err(err).Msg("Failed to query syndicate director.")
				return nil, terror.Error(err, "Failed to issue motion")
			}

			if isDirector {
				return nil, terror.Error(fmt.Errorf("already a director"), "The player is already a director of the syndicate")
			}
		}

		motion.MemberID = bsm.MemberID

	case boiler.SyndicateMotionTypeREMOVE_COMMITTEE:
		if !bsm.MemberID.Valid {
			return nil, terror.Error(fmt.Errorf("member id not provided"), "Member id is not provided")
		}

		// get issuer
		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.MemberID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.MemberID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if !player.SyndicateID.Valid || player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("player is not in your syndicate"), "Player is not a member of the syndicate")
		}

		// check is a committee
		isCommittee, err := boiler.SyndicateCommittees(
			boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sms.syndicate.ID),
			boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
		).Exists(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("syndicate id", sms.syndicate.ID).Str("player id", player.ID).Err(err).Msg("Failed to query syndicate committee.")
			return nil, terror.Error(err, "Failed to issue motion")
		}

		if !isCommittee {
			return nil, terror.Error(fmt.Errorf("not a committee"), "The player is not a committee of the syndicate")
		}

		motion.MemberID = bsm.MemberID

	case boiler.SyndicateMotionTypeDEPOSE_ADMIN:
		// check member is the admin of the syndicate
		syndicate, err := boiler.FindSyndicate(gamedb.StdConn, bsm.SyndicateID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load syndicate detail")
			return nil, terror.Error(err, "Failed to load syndicate detail")
		}

		if syndicate.Type != boiler.SyndicateTypeDECENTRALISED {
			return nil, terror.Error(fmt.Errorf("decentralised syndicate only"), "Only decentralised syndicate can issue motion to depose admin.")
		}

		if !syndicate.AdminID.Valid {
			return nil, terror.Error(fmt.Errorf("no admin player at the moment"), "Your syndicate does not have an admin.")
		}

	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
		// check syndicate type
		syndicate, err := boiler.FindSyndicate(gamedb.StdConn, bsm.SyndicateID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load syndicate detail")
			return nil, terror.Error(err, "Failed to load syndicate detail")
		}

		if syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return nil, terror.Error(fmt.Errorf("not corporation syndicate"), "Only corporation syndicate can appoint director.")
		}

		isDirector, err := db.IsSyndicateDirector(sms.syndicate.ID, userID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("no a director of syndicate"), "Only the directors of the syndicate can issue motion to appoint director.")
		}

		if !bsm.MemberID.Valid {
			return nil, terror.Error(fmt.Errorf("missing player id"), "Missing player id")
		}
		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.MemberID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.MemberID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if !player.SyndicateID.Valid || player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("not syndicate memeber"), "Player is not a member of the syndicate")
		}

		isDirector, err = db.IsSyndicateDirector(sms.syndicate.ID, player.ID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("not a director"), fmt.Sprintf("%s is already a director of the syndicate", player.Username.String))
		}

		motion.MemberID = bsm.MemberID

	case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
		// check syndicate type
		syndicate, err := boiler.FindSyndicate(gamedb.StdConn, bsm.SyndicateID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load syndicate detail")
			return nil, terror.Error(err, "Failed to load syndicate detail")
		}

		if syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return nil, terror.Error(fmt.Errorf("not corporation syndicate"), "Only corporation syndicate can appoint director.")
		}

		isDirector, err := db.IsSyndicateDirector(sms.syndicate.ID, userID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("not a director of syndicate"), "Only the directors of the syndicate can issue motion to remove director.")
		}

		if !bsm.MemberID.Valid {
			return nil, terror.Error(fmt.Errorf("missing player id"), "Missing player id")
		}

		if syndicate.CeoPlayerID.Valid && syndicate.CeoPlayerID.String == bsm.MemberID.String {
			return nil, terror.Error(fmt.Errorf("ceo can only be depose"), "Syndicate CEO can only be deposed instead.")
		}

		if syndicate.AdminID.Valid && syndicate.AdminID.String == bsm.MemberID.String {
			return nil, terror.Error(fmt.Errorf("ceo can only be depose"), "Syndicate admin cannot be removed from the board.")
		}

		player, err := boiler.FindPlayer(gamedb.StdConn, bsm.MemberID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", bsm.MemberID.String).Err(err).Msg("Failed to get data from db")
			return nil, terror.Error(err, "Player not found")
		}

		if player.SyndicateID.String != sms.syndicate.ID {
			return nil, terror.Error(fmt.Errorf("not syndicate memeber"), "Player is not a member of the syndicate")
		}

		isDirector, err = db.IsSyndicateDirector(sms.syndicate.ID, player.ID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("not a director"), fmt.Sprintf("%s is not a director of the syndicate", player.Username.String))
		}

		motion.MemberID = bsm.MemberID

	case boiler.SyndicateMotionTypeDEPOSE_CEO:
		// check syndicate type
		syndicate, err := boiler.FindSyndicate(gamedb.StdConn, bsm.SyndicateID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load syndicate detail")
			return nil, terror.Error(err, "Failed to load syndicate detail")
		}

		if syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return nil, terror.Error(fmt.Errorf("not corporation syndicate"), "Only corporation syndicate can appoint director.")
		}

		isDirector, err := db.IsSyndicateDirector(sms.syndicate.ID, userID)
		if err != nil {
			return nil, err
		}

		if !isDirector {
			return nil, terror.Error(fmt.Errorf("no a director of syndicate"), "Only the directors of the syndicate can issue motion to appoint director.")
		}

		if !syndicate.CeoPlayerID.Valid {
			return nil, terror.Error(fmt.Errorf("no a ceo player"), "Your syndicate does not have a ceo.")
		}

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

	if bsm.FinalisedAt.Valid || bsm.Result.Valid {
		return terror.Error(fmt.Errorf("motion is already ended"), "Motion is already ended.")
	}

	for _, om := range sms.ongoingMotions {
		// skip, if different motion type
		if om.isClosed.Load() {
			continue
		}

		// skip, if motion have different type
		if bsm.Type != om.Type {
			// excluding remove / change rule pair
			if (bsm.Type != boiler.SyndicateMotionTypeREMOVE_RULE || om.Type != boiler.SyndicateMotionTypeCHANGE_RULE) &&
				(bsm.Type != boiler.SyndicateMotionTypeCHANGE_RULE || om.Type != boiler.SyndicateMotionTypeREMOVE_RULE) {
				continue
			}
		}

		// check change content is duplicated
		switch om.Type {
		case boiler.SyndicateMotionTypeCHANGE_GENERAL_DETAIL:
			if bsm.NewSyndicateName.Valid && om.NewSyndicateName.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for changing syndicate name.")
			}
			if bsm.NewSymbol.Valid && om.NewSymbol.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for changing syndicate symbol.")
			}
			if bsm.NewLogoID.Valid && om.NewLogoID.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for changing syndicate logo.")
			}
		case boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for changing entry fee.")
		case boiler.SyndicateMotionTypeCHANGE_MONTHLY_DUES:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for member monthly dues.")
		case boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for changing battle win cut.")
		case boiler.SyndicateMotionTypeADD_RULE:
			if bsm.NewRuleContent.Valid && om.NewRuleContent.Valid {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for adding the same rule.")
			}
		case boiler.SyndicateMotionTypeREMOVE_RULE, boiler.SyndicateMotionTypeCHANGE_RULE:
			if bsm.RuleID.String == om.RuleID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for changing the same rule.")
			}
		case boiler.SyndicateMotionTypeREMOVE_MEMBER:
			if bsm.MemberID.String == om.MemberID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for removing the same member.")
			}
		case boiler.SyndicateMotionTypeAPPOINT_COMMITTEE:
			if bsm.MemberID.String == om.MemberID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for appointing the same member.")
			}
		case boiler.SyndicateMotionTypeREMOVE_COMMITTEE:
			if bsm.MemberID.String == om.MemberID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for removing the same committee.")
			}
		case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
			if bsm.MemberID.String == om.MemberID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for appointing the same player.")
			}
		case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
			if bsm.MemberID.String == om.MemberID.String {
				return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for removing the same director.")
			}
		case boiler.SyndicateMotionTypeDEPOSE_ADMIN:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for deposing admin.")
		case boiler.SyndicateMotionTypeDEPOSE_CEO:
			return terror.Error(fmt.Errorf("duplicate motion content"), "There is already an ongoing motion for deposing ceo.")
		}
	}

	return nil
}

type Motion struct {
	*boiler.SyndicateMotion
	syndicate *Syndicate
	sync.Mutex

	isClosed    atomic.Bool
	forceClosed atomic.String
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
		if sm.EndAt.After(time.Now()) && !sm.isClosed.Load() && sm.forceClosed.Load() == "" {
			continue
		}

		// calculate result
		sm.parseResult()

		// execute result
		sm.action()

		return
	}
}

// Vote check motion is closed or not before firing the function logic
func (sm *Motion) vote(user *boiler.Player, isAgreed bool) error {
	// check vote right
	switch sm.Type {
	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR, boiler.SyndicateMotionTypeREMOVE_DIRECTOR, boiler.SyndicateMotionTypeDEPOSE_CEO:
		isDirector, err := db.IsSyndicateDirector(sm.SyndicateID, user.ID)
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
	if sm.isClosed.Load() || sm.forceClosed.Load() != "" {
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

	totalVoters, err := boiler.Players(
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(sm.SyndicateID)),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to total syndicate users")
		return nil
	}
	votes, err := boiler.SyndicateMotionVotes(
		boiler.SyndicateMotionVoteWhere.MotionID.EQ(sm.ID),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate motion id", sm.ID).Err(err).Msg("Failed to current voted users")
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
		sm.isClosed.Store(true)
		return nil
	}

	return nil
}

func (sm *Motion) parseResult() {
	sm.Lock()
	defer sm.Unlock()
	sm.isClosed.Store(true)

	// only update result, if force closed
	if sm.forceClosed.Load() != "" {
		// parse force close reason
		switch sm.forceClosed.Load() {
		case boiler.SyndicateMotionResultTERMINATED:
			sm.broadcastEndResult(boiler.SyndicateMotionResultTERMINATED, "Terminated by system.")
		case "CEO_ACCEPT":
			sm.broadcastEndResult(boiler.SyndicateMotionResultLEADER_ACCEPTED, "Accepted by syndicate ceo.")
		case "CEO_REJECT":
			sm.broadcastEndResult(boiler.SyndicateMotionResultLEADER_REJECTED, "Rejected by syndicate ceo.")
		case "ADMIN_ACCEPT":
			sm.broadcastEndResult(boiler.SyndicateMotionResultLEADER_ACCEPTED, "Accepted by syndicate admin.")
		case "ADMIN_REJECT":
			sm.broadcastEndResult(boiler.SyndicateMotionResultLEADER_REJECTED, "Rejected by syndicate admin.")
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
	rate := decimal.NewFromInt(int64(agreedCount * 100)).Div(totalVote)
	if (rate.LessThanOrEqual(decimal.NewFromInt(50))) || (sm.Type == boiler.SyndicateMotionTypeDEPOSE_CEO && rate.LessThanOrEqual(decimal.NewFromInt(80))) {
		// broadcast end result
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, "Not enough votes")
		return
	}

	sm.broadcastEndResult(boiler.SyndicateMotionResultPASSED, "Majority of members agreed.")
}

func (sm *Motion) action() {
	if !sm.Result.Valid {
		return
	}

	// do not trigger action if motion failed
	switch sm.Result.String {
	case
		boiler.SyndicateMotionResultTERMINATED,
		boiler.SyndicateMotionResultFAILED,
		boiler.SyndicateMotionResultLEADER_REJECTED:
		return
	}

	// if motion type is any following types, fire action straight away
	switch sm.Type {
	case boiler.SyndicateMotionTypeDEPOSE_CEO:
		sm.deposeCEO()
	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
		sm.appointDirector()
	case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
		sm.removeDirector()
	}

	// check syndicate type
	if sm.syndicate.Type == boiler.SyndicateTypeCORPORATION &&
		sm.Result.String != boiler.SyndicateMotionResultLEADER_ACCEPTED {
		// send to pending table for syndicate ceo or admin to approved

		spm := boiler.SyndicatePendingMotion{
			SyndicateID: sm.SyndicateID,
			MotionID:    sm.ID,
		}

		err := spm.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to insert syndicate pending motion.")
		}

		return
	}

	// process action
	switch sm.Type {
	case boiler.SyndicateMotionTypeCHANGE_GENERAL_DETAIL,
		boiler.SyndicateMotionTypeCHANGE_ENTRY_FEE,
		boiler.SyndicateMotionTypeCHANGE_MONTHLY_DUES,
		boiler.SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT:
		sm.updateSyndicate()
	case boiler.SyndicateMotionTypeADD_RULE:
		sm.addRule()
	case boiler.SyndicateMotionTypeREMOVE_RULE:
		sm.removeRule()
	case boiler.SyndicateMotionTypeCHANGE_RULE:
		sm.changeRule()
	case boiler.SyndicateMotionTypeREMOVE_MEMBER:
		sm.removeMember()
	case boiler.SyndicateMotionTypeAPPOINT_COMMITTEE:
		sm.appointCommittee()
	case boiler.SyndicateMotionTypeREMOVE_COMMITTEE:
		sm.removeCommittee()
	case boiler.SyndicateMotionTypeDEPOSE_ADMIN:
		sm.deposeAdmin()
	}
}

func (sm *Motion) broadcastEndResult(result string, note string) {
	sm.Result = null.StringFrom(result)
	sm.Note = null.StringFrom(note)
	sm.FinalisedAt = null.TimeFrom(time.Now())

	// update motion result and actual end time
	_, err := sm.Update(
		gamedb.StdConn,
		boil.Whitelist(
			boiler.SyndicateMotionColumns.Result,
			boiler.SyndicateMotionColumns.Note,
			boiler.SyndicateMotionColumns.FinalisedAt,
		),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update motion result")
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

	updatedMotionCols := []string{}
	updatedSyndicateCols := []string{}
	if sm.NewSyndicateName.Valid {
		sm.OldSyndicateName = null.StringFrom(syndicate.Name)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldSyndicateName)

		syndicate.Name = sm.NewSyndicateName.String
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.Name)
	}
	if sm.NewSymbol.Valid {
		sm.OldSymbol = null.StringFrom(syndicate.Symbol)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldSymbol)

		syndicate.Symbol = sm.NewSymbol.String
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.Symbol)
	}
	if sm.NewLogoID.Valid {
		sm.OldLogoID = syndicate.LogoID
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldLogoID)

		syndicate.LogoID = sm.NewLogoID
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.LogoID)
	}
	if sm.NewJoinFee.Valid {
		sm.OldJoinFee = decimal.NewNullDecimal(syndicate.JoinFee)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldJoinFee)

		syndicate.JoinFee = sm.NewJoinFee.Decimal
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.JoinFee)
	}
	if sm.NewMonthlyDues.Valid {
		sm.OldMonthlyDues = decimal.NewNullDecimal(syndicate.MemberMonthlyDues)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldMonthlyDues)

		syndicate.MemberMonthlyDues = sm.NewMonthlyDues.Decimal
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.MemberMonthlyDues)
	}
	if sm.NewDeployingMemberCutPercentage.Valid {
		sm.OldDeployingMemberCutPercentage = decimal.NewNullDecimal(syndicate.DeployingMemberCutPercentage)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldDeployingMemberCutPercentage)

		syndicate.DeployingMemberCutPercentage = sm.NewDeployingMemberCutPercentage.Decimal
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.DeployingMemberCutPercentage)
	}
	if sm.NewMemberAssistCutPercentage.Valid {
		sm.OldMemberAssistCutPercentage = decimal.NewNullDecimal(syndicate.MemberAssistCutPercentage)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldMemberAssistCutPercentage)

		syndicate.MemberAssistCutPercentage = sm.NewMemberAssistCutPercentage.Decimal
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.MemberAssistCutPercentage)
	}
	if sm.NewMechOwnerCutPercentage.Valid {
		sm.OldMechOwnerCutPercentage = decimal.NewNullDecimal(syndicate.MechOwnerCutPercentage)
		updatedMotionCols = append(updatedMotionCols, boiler.SyndicateMotionColumns.OldMechOwnerCutPercentage)

		syndicate.MechOwnerCutPercentage = sm.NewMechOwnerCutPercentage.Decimal
		updatedSyndicateCols = append(updatedSyndicateCols, boiler.SyndicateColumns.MechOwnerCutPercentage)
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction")
		return
	}

	defer tx.Rollback()

	// record all the old value in motion
	_, err = sm.Update(tx, boil.Whitelist(updatedMotionCols...))
	if err != nil {
		gamelog.L.Error().Interface("updated motion", sm).Strs("updated column", updatedMotionCols).Err(err).Msg("Failed to update motion columns")
		return
	}

	// update syndicate detail
	_, err = syndicate.Update(tx, boil.Whitelist(updatedSyndicateCols...))
	if err != nil {
		gamelog.L.Error().Interface("updated syndicate", syndicate).Strs("updated column", updatedSyndicateCols).Err(err).Msg("Failed to update syndicate from motion")
		return
	}

	// change syndicate name in passport, if it is updated
	for _, updatedCol := range updatedSyndicateCols {
		if updatedCol != boiler.SyndicateColumns.Name {
			continue
		}

		err = sm.syndicate.system.Passport.SyndicateNameChangeHandler(syndicate.ID, syndicate.Name)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to update syndicate name")
			return
		}
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

func (sm *Motion) removeMember() {
	// get target player
	player, err := boiler.FindPlayer(gamedb.StdConn, sm.MemberID.String)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get member from db")
		return
	}

	// check the player is still a member of the syndicate
	if !player.SyndicateID.Valid || player.SyndicateID.String != sm.SyndicateID {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is no longer a member of our syndicate.", player.Username.String, player.Gid))
		return
	}

	// check the player is a committee
	isCommittee, err := boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to check syndicate committee list")
	}

	if isCommittee {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is a committee of our syndicate now.", player.Username.String, player.Gid))
		return
	}

	// check the player is a director
	isDirector, err := boiler.SyndicateDirectors(
		boiler.SyndicateDirectorWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateDirectorWhere.PlayerID.EQ(player.ID),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to check syndicate director list")
	}

	if isDirector {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is a director of our syndicate now.", player.Username.String, player.Gid))
		return
	}

	// vote pass
	player.SyndicateID = null.StringFromPtr(nil)
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
	if err != nil {
		gamelog.L.Error().Str("player id", player.ID).Err(err).Msg("Failed to update player syndicate id in db")
	}

	err = player.L.LoadRole(gamedb.StdConn, true, player, nil)
	if err != nil {
		gamelog.L.Error().Str("player id", player.ID).Err(err).Msg("Failed to load role_id")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s", player.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(player))
}

func (sm *Motion) appointCommittee() {
	// get target player
	player, err := boiler.FindPlayer(gamedb.StdConn, sm.MemberID.String)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get member from db")
		return
	}

	// check the player is still a member of the syndicate
	if !player.SyndicateID.Valid || player.SyndicateID.String != sm.SyndicateID {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is no longer a member of our syndicate.", player.Username.String, player.Gid))
		return
	}

	// check the player is a committee
	isCommittee, err := boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to check syndicate committee list")
	}

	if isCommittee {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is already a committee of our syndicate now.", player.Username.String, player.Gid))
		return
	}

	// check the player is a director
	isDirector, err := boiler.SyndicateDirectors(
		boiler.SyndicateDirectorWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateDirectorWhere.PlayerID.EQ(player.ID),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to check syndicate director list")
	}

	if isDirector {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is already a director of our syndicate now.", player.Username.String, player.Gid))
		return
	}

	// add committee to syndicate
	sc := boiler.SyndicateCommittee{
		SyndicateID: sm.SyndicateID,
		PlayerID:    player.ID,
	}

	err = sc.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to insert syndicate committee")
		return
	}

	// load new director list
	scs, err := db.GetSyndicateCommittees(sm.SyndicateID)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to get syndicate directors")
		return
	}

	// broadcast syndicate director list
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/committees", sm.syndicate.FactionID, sm.SyndicateID), server.HubKeySyndicateCommitteesSubscribe, scs)
}

func (sm *Motion) removeCommittee() {
	// get target player
	player, err := boiler.FindPlayer(gamedb.StdConn, sm.MemberID.String)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get member from db")
		return
	}

	// check the player is still a member of the syndicate
	if !player.SyndicateID.Valid || player.SyndicateID.String != sm.SyndicateID {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is no longer a member of our syndicate.", player.Username.String, player.Gid))
		return
	}

	// check the player is a committee
	isCommittee, err := boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to check syndicate committee list")
	}

	if !isCommittee {
		sm.broadcastEndResult(boiler.SyndicateMotionResultFAILED, fmt.Sprintf("Player %s #%d is no longer a committee of our syndicate.", player.Username.String, player.Gid))
		return
	}

	// add committee to syndicate
	_, err = boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateCommitteeWhere.PlayerID.EQ(player.ID),
	).DeleteAll(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", player.ID).Err(err).Msg("Failed to delete syndicate committee")
		return
	}

	// load new committee list
	scs, err := db.GetSyndicateCommittees(sm.SyndicateID)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Err(err).Msg("Failed to get syndicate directors")
		return
	}

	// broadcast syndicate committee list
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/committees", sm.syndicate.FactionID, sm.SyndicateID), server.HubKeySyndicateCommitteesSubscribe, scs)
}

/****************************
* Board of director exclusive
 ***************************/

func (sm *Motion) appointDirector() {
	sd := boiler.SyndicateDirector{
		SyndicateID: sm.SyndicateID,
		PlayerID:    sm.MemberID.String,
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction")
		return
	}

	defer tx.Rollback()

	// insert syndicate director role
	err = sd.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", sm.MemberID.String).Err(err).Msg("Failed to appoint player to director")
		return
	}

	// remove committee role
	_, err = boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(sm.SyndicateID),
		boiler.SyndicateCommitteeWhere.PlayerID.EQ(sm.MemberID.String),
	).DeleteAll(tx)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Str("player id", sm.MemberID.String).Err(err).Msg("Failed to remove player's committee role")
		return
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
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
		boiler.SyndicateDirectorWhere.PlayerID.EQ(sm.MemberID.String),
	).DeleteAll(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", sm.SyndicateID).Interface("player id", sm.MemberID.String).Err(err).Msg("Failed to remove player from syndicate director list")
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

func (sm *Motion) deposeAdmin() {
	s := &boiler.Syndicate{
		ID:      sm.SyndicateID,
		AdminID: null.StringFromPtr(nil),
	}

	// remove both ceo and admin player
	_, err := s.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateColumns.CeoPlayerID, boiler.SyndicateColumns.AdminID))
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", sm.SyndicateID).Msg("Failed to depose ceo of the syndicate.")
		return
	}

	sm.broadcastUpdatedSyndicate()

	// TODO: spin up an admin election
}

func (sm *Motion) deposeCEO() {
	s := &boiler.Syndicate{
		ID:          sm.SyndicateID,
		CeoPlayerID: null.StringFromPtr(nil),
		AdminID:     null.StringFromPtr(nil),
	}

	// remove both ceo and admin player
	_, err := s.Update(gamedb.StdConn, boil.Whitelist(boiler.SyndicateColumns.CeoPlayerID, boiler.SyndicateColumns.AdminID))
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", sm.SyndicateID).Msg("Failed to depose ceo of the syndicate.")
		return
	}

	sm.broadcastUpdatedSyndicate()

	// TODO: spin up a ceo election
}
