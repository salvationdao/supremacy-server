package api

import (
	"server"
	"sync"
	"time"
)

type UserMultiplier struct {
	sync.RWMutex
	CurrentMap map[server.UserID]*Multiplier
	CheckMap   map[server.UserID]*Multiplier
}

type Multiplier struct {
	sync.RWMutex
	Map map[string]*MultiplierAction
}

type MultiplierAction struct {
	MultiplierValue int
	Expiry          time.Time
}

func NewUserMultiplierMap() *UserMultiplier {
	return &UserMultiplier{
		sync.RWMutex{},
		make(map[server.UserID]*Multiplier),
		make(map[server.UserID]*Multiplier),
	}
}

func (um *UserMultiplier) Online(userID server.UserID) {
	um.Lock()
	defer um.Unlock()
	mm, ok := um.CurrentMap[userID]
	if !ok {
		mm = &Multiplier{
			sync.RWMutex{},
			make(map[string]*MultiplierAction),
		}
	}

	mm.RWMutex.Lock()
	defer mm.RWMutex.Unlock()
	m, ok := mm.Map[string(ClientOnline)]
	if !ok {
		m = &MultiplierAction{}
	}

	m.MultiplierValue = 100
	m.Expiry = time.Now().AddDate(1, 0, 0)

	mm.Map[string(ClientOnline)] = m
	um.CurrentMap[userID] = mm
}

func (um *UserMultiplier) Offline(userID server.UserID) {
	um.Lock()
	delete(um.CurrentMap, userID)
	um.Unlock()
}
