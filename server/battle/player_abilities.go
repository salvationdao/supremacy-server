package battle

import (
	"go.uber.org/atomic"
	"sync"
)

type PlayerAbilitiesSystem struct {
	_battle *Battle

	end    chan bool
	closed *atomic.Bool
	sync.RWMutex
}

func NewPlayerAbilitiesSystem() *PlayerAbilitiesSystem {

	pas := &PlayerAbilitiesSystem{
		closed: atomic.NewBool(false),
	}

	return pas
}
