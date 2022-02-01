package server

import (
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
)

type Battle struct {
	ID                BattleID               `json:"battleID" db:"id"`
	GameMapID         GameMapID              `json:"gameMapID" db:"game_map_id"`
	StartedAt         time.Time              `json:"startedAt" db:"started_at"`
	EndedAt           *time.Time             `json:"endedAt" db:"ended_at"`
	WinningCondition  *string                `json:"winningCondition" db:"winning_condition"`
	WarMachines       []*WarMachineNFT       `json:"warMachines"`
	WinningWarMachine *uint64                `json:"winningWarMachine"`
	GameMap           *GameMap               `json:"map"`
	FactionMap        map[FactionID]*Faction `json:"factionMap"`
}

type BattleState string

const (
	StateLobby      BattleState = "LOBBY"
	StateMatchStart BattleState = "START"
	StateMatchEnd   BattleState = "END"
)

type BattleWinCondition string

const (
	WinConditionLastAlive BattleWinCondition = "LAST_ALIVE"
	WinConditionOther     BattleWinCondition = "OTHER"
)

type BattleID uuid.UUID

// IsNil returns true for a nil uuid.
func (id BattleID) IsNil() bool {
	return id == BattleID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id BattleID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id BattleID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *BattleID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = BattleID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id BattleID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *BattleID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = BattleID(uid)
	// Retrun error
	return err
}
