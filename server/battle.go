package server

import (
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
)

type Battle struct {
	ID                 BattleID               `json:"battleID" db:"id"`
	Identifier         int64                  `json:"identifier"`
	GameMapID          GameMapID              `json:"gameMapID" db:"game_map_id"`
	StartedAt          time.Time              `json:"startedAt" db:"started_at"`
	EndedAt            *time.Time             `json:"endedAt" db:"ended_at"`
	WinningCondition   *string                `json:"winningCondition" db:"winning_condition"`
	WarMachines        []*WarMachineMetadata  `json:"warMachines"`
	WinningWarMachines []*WarMachineMetadata  `json:"winningWarMachines"`
	GameMap            *GameMap               `json:"map"`
	FactionMap         map[FactionID]*Faction `json:"factionMap"`
	BattleHistory      [][]byte               `json:"battleHistory"`

	// used for destroyed notification subscription
	WarMachineDestroyedRecordMap map[byte]*WarMachineDestroyedRecord

	// State
	State BattleState
}

type WarMachineDestroyedRecord struct {
	DestroyedWarMachine *WarMachineMetadata `json:"destroyedWarMachine"`
	KilledByWarMachine  *WarMachineMetadata `json:"killedByWarMachine,omitempty"`
	KilledBy            string              `json:"killedBy"`
	DamageRecords       []*DamageRecord     `json:"damageRecords"`
}

type DamageRecord struct {
	Amount             int                 `json:"amount"` // The total amount of damage taken from this source
	CausedByWarMachine *WarMachineMetadata `json:"causedByWarMachine,omitempty"`
	SourceName         string              `json:"sourceName,omitempty"` // The name of the weapon / damage causer (in-case of now TokenID)
}

type FactionWarMachineQueue struct {
	FactionID  FactionID `json:"factionID"`
	QueueTotal int       `json:"queueTotal"`
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

type BattleQueueMetadata struct {
	WarMachineMetadata *WarMachineMetadata `json:"warMachineMetadata" db:"war_machine_metadata"`
	QueuedAt           time.Time           `json:"queuedAt" db:"queued_at"`
	DeletedAt          *time.Time          `json:"deletedAt,omitempty" db:"deleted_at,omitempty"`
	ContractReward     string              `json:"contractReward" db:"contract_reward"`
	IsInsured          bool                `json:"isInsured" db:"is_insured"`
}

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

type RepairMode string

const (
	RepairModeFast     = "FAST"
	RepairModeStandard = "STANDARD"
)

type AssetRepairRecord struct {
	Hash              string     `json:"hash" db:"hash"`
	ExpectCompletedAt time.Time  `json:"expectCompletedAt" db:"expect_completed_at"`
	RepairMode        RepairMode `json:"repairMode" db:"repair_mode"`
	IsPaidToComplete  bool       `json:"isPaidToComplete" db:"is_paid_to_complete"`
	CompletedAt       *time.Time `json:"completedAt,omitempty" db:"completed_at,omitempty"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
}
