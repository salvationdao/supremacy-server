package syndicate

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/profanities"
	"sync"
)

type System struct {
	profanityManager *profanities.ProfanityManager
	syndicateMap     map[string]*Syndicate
	sync.RWMutex
}

func NewSystem(profanityManager *profanities.ProfanityManager) (*System, error) {
	syndicates, err := boiler.Syndicates().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load syndicated from db")
		return nil, terror.Error(err, "Failed to get syndicates from db")
	}

	ss := &System{
		profanityManager: profanityManager,
		syndicateMap:     make(map[string]*Syndicate),
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
func (ss *System) AddSyndicate(syndicate *boiler.Syndicate) error {
	ss.Lock()
	defer ss.Unlock()

	s, err := ss.newSyndicate(syndicate)
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate")
		return terror.Error(err, "Failed to spin up syndicate")
	}

	ss.syndicateMap[s.ID] = s

	return nil
}

// RemoveSyndicate remove syndicate from the system
func (ss *System) RemoveSyndicate(id string) error {
	ss.Lock()
	defer ss.Unlock()

	s, ok := ss.syndicateMap[id]
	if !ok {
		return terror.Error(fmt.Errorf("syndicate not found"), "syndicate not found")
	}

	err := s.liquidate()
	if err != nil {
		return err
	}

	// delete syndicate from the map
	delete(ss.syndicateMap, id)

	return nil
}

func (ss *System) GetSyndicate(id string) (*Syndicate, error) {
	ss.RLock()
	defer ss.RUnlock()

	s, ok := ss.syndicateMap[id]
	if !ok {
		return nil, terror.Error(fmt.Errorf("syndicate not exist"), "Syndicate does not exit")
	}

	return s, nil
}

// AddMotion add new motion to the syndicate system
func (ss *System) AddMotion(user *boiler.Player, bsm *boiler.SyndicateMotion) error {
	// get syndicate
	s, err := ss.GetSyndicate(user.SyndicateID.String)
	if err != nil {
		return err
	}

	// check motion is valid and generate a clean motion
	newMotion, err := s.motionSystem.motionValidCheck(user, bsm)
	if err != nil {
		return err
	}

	// add motion to the motion system
	err = s.motionSystem.addMotion(newMotion, true)
	if err != nil {
		return err
	}

	return nil
}

// VoteMotion get the motion from the syndicate and trigger vote
func (ss *System) VoteMotion(user *boiler.Player, motionID string, isAgreed bool) error {
	// get syndicate
	s, err := ss.GetSyndicate(user.SyndicateID.String)
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

// GetOngoingMotions get the motions from the syndicate
func (ss *System) GetOngoingMotions(user *boiler.Player) ([]*boiler.SyndicateMotion, error) {
	// get syndicate
	s, err := ss.GetSyndicate(user.SyndicateID.String)
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

func (ss *System) newSyndicate(syndicate *boiler.Syndicate) (*Syndicate, error) {
	s := &Syndicate{
		system:    ss,
		Syndicate: syndicate,
	}
	motionSystem, err := s.newMotionSystem()
	if err != nil {
		gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to spin up syndicate motion system.")
		return nil, terror.Error(err, "Failed to spin up syndicate motion system.")
	}

	s.motionSystem = motionSystem

	return s, nil
}
