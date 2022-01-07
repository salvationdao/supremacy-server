package gameserver

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

type WarMachineDestroyed struct {
	DestroyedWarMachineID WarMachineID    `json:"destroyedWarMachineId"`
	KillerWarMachineID    *WarMachineID   `json:"killerWarMachineId"` // make this a pointer, possibly killed via environment or event
	AssistedWarMachineIDs []*WarMachineID `json:"assistedWarMachineIds"`
	KilledBy              string          `json:"killedBy"` // this will hold weapon name or event name?
}

type FactionAbility struct {
	FactionID  FactionID     `json:"factionId"`
	Action     FactionAction `json:"action"`
	Successful bool
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
