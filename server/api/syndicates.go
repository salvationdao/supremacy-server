package api

import (
	"github.com/ninja-software/terror/v2"
	"server/db/boiler"
	"server/gamedb"
	"sync"
)

type SyndicateSystem struct {
	syndicateMap map[string]*Syndicate
	sync.RWMutex
}

func NewSyndicateSystem() (*SyndicateSystem, error) {
	syndicates, err := boiler.Syndicates().All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get syndicates from db")
	}

	ss := &SyndicateSystem{
		syndicateMap: make(map[string]*Syndicate),
	}
	for _, s := range syndicates {
		ss.AddSyndicate(s)
	}

	return ss, nil
}

func (ss *SyndicateSystem) AddSyndicate(s *boiler.Syndicate) {
	ss.Lock()
	defer ss.Unlock()

	ss.syndicateMap[s.ID] = NewSyndicate(s)
}

func (ss *SyndicateSystem) RemoveSyndicate(id string) {
	ss.Lock()
	defer ss.Unlock()

	// TODO: close all the channel in the syndicate

	// delete syndicate from the map
	delete(ss.syndicateMap, id)
}

type Syndicate struct {
	*boiler.Syndicate

	issueMotionChan chan *SyndicateIssueMotionRequest
}

func NewSyndicate(s *boiler.Syndicate) *Syndicate {
	syndicate := &Syndicate{
		s,
		make(chan *SyndicateIssueMotionRequest),
	}

	// setup syndicate

	return syndicate
}
