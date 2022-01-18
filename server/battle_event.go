package server

import (
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
)

type BattleEventType string

const (
	BattleEventWarMachineDestroyed BattleEventType = "WAR_MACHINE_DESTROYED"
	BattleEventFactionAbility      BattleEventType = "FACTION_ABILITY"
)

type BattleEvent struct {
	ID        EventID         `json:"id" db:"id"`
	BattleID  BattleID        `json:"battle_id" db:"battle_id"`
	EventType BattleEventType `json:"event_type" db:"event_type"`
	Event     interface{}     `json:"event" db:"event"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type WarMachineDestroyedEvent struct {
	ID                     WarMachineDestroyedEventID `json:"id" db:"id"`
	EventID                EventID                    `json:"eventID" db:"event_id"`
	DestroyedWarMachineID  uint64                     `json:"destroyedWarMachineId" db:"destroyed_war_machine_id"`
	KillByWarMachineID     *uint64                    `json:"killByWarMachineID,omitempty" db:"kill_by_war_machine_id,omitempty"`
	KillByFactionAbilityID *uint64                    `json:"killByFactionAbilityID,omitempty" db:"kill_by_faction_ability_id,omitempty"`
	AssistedWarMachineIDs  []uint64                   `json:"assistedWarMachineIds"`
	KilledBy               string                     `json:"killedBy"` // this will hold weapon name or event name?
}

type FactionAbilityEvent struct {
	ID               FactionAbilityEventID `json:"id" db:"id"`
	EventID          EventID               `json:"eventID" db:"event_id"`
	FactionAbilityID FactionAbilityID      `json:"factionAbilityID" db:"faction_ability_id"`
	IsTriggered      bool                  `json:"isTriggered" db:"is_triggered"`
	TriggeredByUser  *string               `json:"triggeredByUser,omitempty" db:"triggered_by_user,omitempty"`
	TriggeredOnCellX *int                  `json:"triggeredOnCellX,omitempty" db:"triggered_on_cell_x,omitempty"`
	TriggeredOnCellY *int                  `json:"triggeredOnCellY,omitempty" db:"triggered_on_cell_y,omitempty"`
}

type EventID uuid.UUID

// IsNil returns true for a nil uuid.
func (id EventID) IsNil() bool {
	return id == EventID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id EventID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id EventID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *EventID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = EventID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id EventID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *EventID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = EventID(uid)
	// Retrun error
	return err
}
