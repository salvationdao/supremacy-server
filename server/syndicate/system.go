package syndicate

import (
	"database/sql"
	"fmt"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/profanities"
	"server/xsyn_rpcclient"
	"strings"
	"sync"
)

type System struct {
	profanityManager *profanities.ProfanityManager
	Passport         *xsyn_rpcclient.XsynXrpcClient
	syndicateMap     map[string]*Syndicate
	sync.RWMutex
}

func (ss *System) getSyndicate(id string) (*Syndicate, error) {
	ss.RLock()
	defer ss.RUnlock()

	s, ok := ss.syndicateMap[id]
	if !ok {
		return nil, terror.Error(fmt.Errorf("syndicate not exist"), "Syndicate does not exit")
	}

	if s.isLiquidated.Load() {
		return nil, terror.Error(fmt.Errorf("syndicate is liquidated"), "The syndicate is liquidated.")
	}

	return s, nil
}

func (ss *System) removeSyndicate(id string) {
	ss.Lock()
	defer ss.Unlock()

	if _, ok := ss.syndicateMap[id]; ok {
		delete(ss.syndicateMap, id)
	}
}

func (ss *System) addSyndicate(s *Syndicate) {
	ss.Lock()
	defer ss.Unlock()

	ss.syndicateMap[s.ID] = s
}

// NewSystem create a new syndicate system
func NewSystem(Passport *xsyn_rpcclient.XsynXrpcClient, profanityManager *profanities.ProfanityManager) (*System, error) {
	syndicates, err := boiler.Syndicates().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load syndicated from db")
		return nil, terror.Error(err, "Failed to get syndicates from db")
	}

	ss := &System{
		profanityManager: profanityManager,
		Passport:         Passport,
		syndicateMap:     make(map[string]*Syndicate),
	}

	for _, syndicate := range syndicates {
		s, err := newSyndicate(ss, syndicate)
		if err != nil {
			gamelog.L.Error().Str("syndicate id", syndicate.ID).Err(err).Msg("Failed to spin up syndicate")
			return nil, terror.Error(err, "Failed to spin up syndicate")
		}

		ss.addSyndicate(s)
	}

	return ss, nil
}

// CreateSyndicate create new syndicate in the system
func (ss *System) CreateSyndicate(syndicateID string) error {
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, syndicateID)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", syndicateID).Err(err).Msg("Failed to get syndicate from db")
		return terror.Error(err, "Failed to load syndicate detail")
	}

	s, err := newSyndicate(ss, syndicate)
	if err != nil {
		gamelog.L.Error().Str("syndicate id", syndicateID).Err(err).Msg("Failed to spin up syndicate")
		return terror.Error(err, "Failed to spin up syndicate")
	}

	ss.addSyndicate(s)

	return nil
}

// AddMotion add new motion to the syndicate system
func (ss *System) AddMotion(user *boiler.Player, bsm *boiler.SyndicateMotion, blob *boiler.Blob) error {
	// get syndicate
	s, err := ss.getSyndicate(user.SyndicateID.String)
	if err != nil {
		return err
	}

	// check motion is valid and generate a clean motion
	newMotion, err := s.motionSystem.validateIncomingMotion(user.ID, bsm, blob)
	if err != nil {
		return err
	}

	// add motion to the motion system
	err = s.motionSystem.addMotion(newMotion, blob, true)
	if err != nil {
		return err
	}

	return nil
}

// VoteMotion get the motion from the syndicate and trigger vote
func (ss *System) VoteMotion(user *boiler.Player, motionID string, isAgreed bool) error {
	// get syndicate
	s, err := ss.getSyndicate(user.SyndicateID.String)
	if err != nil {
		return err
	}

	// fire motion vote
	om, err := s.motionSystem.getOngoingMotion(motionID)
	if err != nil {
		return err
	}

	return om.vote(user, isAgreed)
}

// ForceCloseMotionsByType close specific type of motions
func (ss *System) ForceCloseMotionsByType(syndicateID string, reason string, motionType ...string) error {
	if len(motionType) == 0 {
		return terror.Error(fmt.Errorf("no type provided"), "Did not specify which type of motion to close.")
	}

	s, err := ss.getSyndicate(syndicateID)
	if err != nil {
		return err
	}

	s.motionSystem.forceCloseTypes(reason, motionType...)

	return nil
}

// GetOngoingMotions get the motions from the syndicate
func (ss *System) GetOngoingMotions(user *boiler.Player) ([]*boiler.SyndicateMotion, error) {
	// get syndicate
	s, err := ss.getSyndicate(user.SyndicateID.String)
	if err != nil {
		return nil, err
	}

	// fire motion vote
	oms, err := s.motionSystem.getOngoingMotionList()
	if err != nil {
		return nil, err
	}

	return oms, nil
}

// LiquidateSyndicate remove syndicate from the system
func (ss *System) LiquidateSyndicate(tx *sql.Tx, id string) error {
	s, err := ss.getSyndicate(id)
	if err != nil {
		return err
	}

	err = s.liquidate(tx)
	if err != nil {
		return err
	}

	err = ss.Passport.SyndicateLiquidateHandler(s.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to liquidate syndicate in xsyn services.")
		return terror.Error(err, "Failed to liquidate syndicate.")
	}

	ss.removeSyndicate(id)

	return nil
}

func (ss *System) SyndicateNameVerification(inputName string) (string, error) {
	syndicateName := strings.TrimSpace(inputName)

	if len(syndicateName) == 0 {
		return "", terror.Error(fmt.Errorf("empty syndicate name"), "The name of syndicate is empty")
	}

	// check profanity
	if ss.profanityManager.Detector.IsProfane(syndicateName) {
		return "", terror.Error(fmt.Errorf("profanity detected"), "The syndicate name contains profanity")
	}

	if len(syndicateName) > 64 {
		return "", terror.Error(fmt.Errorf("too many characters"), "The syndicate name should not be longer than 64 characters")
	}

	// name similarity check
	syndicates, err := boiler.Syndicates(
		qm.Select(boiler.SyndicateColumns.ID, boiler.SyndicateColumns.Name),
		qm.Where(
			fmt.Sprintf("SIMILARITY(%s,$1) > 0.4", qm.Rels(boiler.TableNames.Syndicates, boiler.SyndicateColumns.Name)),
			syndicateName,
		),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate name", syndicateName).Msg("Failed to get syndicate by name from db")
		return "", terror.Error(err, "Failed to verify syndicate name")
	}

	// https://github.com/adrg/strutil
	for _, s := range syndicates {
		if strutil.Similarity(syndicateName, s.Name, metrics.NewHamming()) >= 0.9 {
			return "", terror.Error(fmt.Errorf("similar syndicate name"), fmt.Sprintf("'%s' has already been taken by other syndicate", syndicateName))
		}
	}

	return syndicateName, nil
}

func (ss *System) SyndicateSymbolVerification(inputSymbol string) (string, error) {
	// get rid of all the spaces
	symbol := strings.ToUpper(strings.ReplaceAll(inputSymbol, " ", ""))

	if len(symbol) < 4 || len(symbol) > 5 {
		return "", terror.Error(fmt.Errorf("must be 4 or 5 character"), "The length of symbol must be at least four and no more than five excluding spaces.")
	}

	// check profanity
	if ss.profanityManager.Detector.IsProfane(symbol) {
		return "", terror.Error(fmt.Errorf("profanity detected"), "The syndicate symbol contains profanity")
	}

	// name similarity check
	syndicates, err := boiler.Syndicates(
		qm.Select(boiler.SyndicateColumns.ID, boiler.SyndicateColumns.Symbol),
		qm.Where(
			fmt.Sprintf("SIMILARITY(%s,$1) > 0.4", qm.Rels(boiler.TableNames.Syndicates, boiler.SyndicateColumns.Symbol)),
			symbol,
		),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate name", symbol).Msg("Failed to get syndicate by name from db")
		return "", terror.Error(err, "Failed to verify syndicate name")
	}

	// https://github.com/adrg/strutil
	for _, s := range syndicates {
		if strutil.Similarity(symbol, s.Symbol, metrics.NewHamming()) >= 0.9 {
			return "", terror.Error(fmt.Errorf("similar syndicate name"), fmt.Sprintf("'%s' has already been taken by other syndicate", symbol))
		}
	}

	return symbol, nil
}
