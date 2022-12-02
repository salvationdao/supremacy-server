package player_abilities

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
)

type MiniMechMoveCommand struct {
	BattleID       string
	CellX          decimal.Decimal
	CellY          decimal.Decimal
	TriggeredByID  string
	FactionID      string
	MechHash       string
	CooldownExpiry time.Time
	CancelledAt    null.Time
	ReachedAt      null.Time
	CreatedAt      time.Time
	IsMoving       bool

	deadlock.RWMutex
}

type ReadMiniMechMoveCommandFunc func(*MiniMechMoveCommand)

func (mm *MiniMechMoveCommand) Read(fn ReadMiniMechMoveCommandFunc) {
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
	mm.IsMoving = false

	return nil
}

func (mm *MiniMechMoveCommand) Complete() {
	mm.Lock()
	defer mm.Unlock()

	now := time.Now()
	mm.ReachedAt = null.TimeFrom(now)
	mm.IsMoving = false
}
