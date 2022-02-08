package server

import (
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
)

type BattleEventType string

const (
	BattleEventTypeWarMachineDestroyed BattleEventType = "WAR_MACHINE_DESTROYED"
	BattleEventTypeFactionAbility      BattleEventType = "FACTION_ABILITY"
	BattleEventTypeStateChange         BattleEventType = "STATE"
)

type BattleEvent struct {
	ID        EventID         `json:"id" db:"id"`
	BattleID  BattleID        `json:"battle_id" db:"battle_id"`
	EventType BattleEventType `json:"event_type" db:"event_type"`
	Event     interface{}     `json:"event" db:"event"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type DamageHistory struct {
	Amount            int    `json:"amount"`            // The total amount of damage taken from this source
	InstigatorTokenID uint64 `json:"instigatorTokenID"` // The TokenID of the WarMachine that caused the damage (0 if none, ie: an Airstrike)
	SourceTokenID     uint64 `json:"sourceTokenID"`     // The TokenID of the weapon
	SourceName        string `json:"sourceName"`        // The name of the weapon / damage causer (in-case of now TokenID)
}

type WarMachineDestroyedEvent struct {
	ID                    WarMachineDestroyedEventID `json:"id" db:"id"`
	EventID               EventID                    `json:"eventID" db:"event_id"`
	DestroyedWarMachineID uint64                     `json:"destroyedWarMachineID" db:"destroyed_war_machine_id"`
	KillByWarMachineID    *uint64                    `json:"killByWarMachineID,omitempty" db:"kill_by_war_machine_id,omitempty"`
	RelatedEventID        *EventID                   `json:"relatedEventID,omitempty" db:"related_event_id,omitempty"`
	RelatedEventIDString  string                     `json:"relatedEventIDString,omitempty"` // The related EventID in string form (received from game client)
	DamageHistory         []DamageHistory            `json:"damageHistory"`                  // Compiled History of all the damage this WarMachine took and from who/what
	KilledBy              string                     `json:"killedBy"`                       // Name of who/what killed the WarMachine (in-case of no EventID or TokenID)
}

type FactionAbilityEvent struct {
	ID                  FactionAbilityEventID `json:"id" db:"id"`
	EventID             EventID               `json:"eventID" db:"event_id"`
	FactionAbilityID    FactionAbilityID      `json:"factionAbilityID" db:"faction_ability_id"`
	GameClientAbilityID byte                  `json:"gameClientAbilityID" db:"game_client_ability_id"`
	IsTriggered         bool                  `json:"isTriggered" db:"is_triggered"`
	TriggeredByUserID   *string               `json:"TriggeredByUserID,omitempty" db:"triggered_by_user_id,omitempty"`
	TriggeredByUsername *string               `json:"triggeredByUsername"`
	TriggeredOnCellX    *int                  `json:"triggeredOnCellX,omitempty" db:"triggered_on_cell_x,omitempty"`
	TriggeredOnCellY    *int                  `json:"triggeredOnCellY,omitempty" db:"triggered_on_cell_y,omitempty"`
	GameLocation        struct {
		X int `json:"X"`
		Y int `json:"Y"`
	} `json:"gameLocation"`
}

type BattleEventState string

const (
	BattleEventBattleStart = "START"
	BattleEventBattleEnd   = "END"
)

type BattleEventStateChange struct {
	ID      FactionAbilityEventID `json:"id" db:"id"`
	EventID EventID               `json:"eventID" db:"event_id"`
	State   BattleEventState      `json:"state" db:"state"`
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
