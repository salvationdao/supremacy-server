package battle

import (
	"github.com/sasha-s/go-deadlock"
	"go.uber.org/atomic"
)

type PlayerAbilitiesSystem struct {
	_battle *Battle

	end    chan bool
	closed *atomic.Bool
	deadlock.RWMutex
}

func NewPlayerAbilitiesSystem() *PlayerAbilitiesSystem {

	pas := &PlayerAbilitiesSystem{
		closed: atomic.NewBool(false),
	}

	return pas
}
