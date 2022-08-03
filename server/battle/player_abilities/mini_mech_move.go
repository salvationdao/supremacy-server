package player_abilities

import (
	"fmt"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
)

type MiniMechMoveCommand struct {
	BattleID       string
	CellX          int
	CellY          int
	TriggeredByID  string
	FactionID      string
	MechHash       string
	CooldownExpiry time.Time
	CancelledAt    null.Time
	ReachedAt      null.Time
	CreatedAt      time.Time

	deadlock.RWMutex
}

type ReadFunc func(*MiniMechMoveCommand)

func (mm *MiniMechMoveCommand) Read(fn ReadFunc) {
	mm.RLock()
	defer mm.RUnlock()

	fn(mm)
}

func (mm *MiniMechMoveCommand) Cancel() error {
	mm.Lock()
	defer mm.Unlock()

	if mm.CancelledAt.Valid {
		return fmt.Errorf("Mech move command is already cancelled.")
	}

	if mm.ReachedAt.Valid {
		return fmt.Errorf("Mech has already reached the commanded spot")
	}

	now := time.Now()
	mm.CancelledAt = null.TimeFrom(now)

	return nil
}

func (mm *MiniMechMoveCommand) Complete() {
	mm.Lock()
	defer mm.Unlock()

	now := time.Now()
	mm.ReachedAt = null.TimeFrom(now)
}
