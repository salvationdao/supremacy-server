package battle

import (
	"sync"

	"go.uber.org/atomic"
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
