package server

import (
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
)

type Battle struct {
	ID                 BattleID               `json:"battle_id" db:"id"`
	Identifier         int64                  `json:"identifier"`
	GameMapID          GameMapID              `json:"game_map_id" db:"game_map_id"`
	StartedAt          time.Time              `json:"started_at" db:"started_at"`
	EndedAt            *time.Time             `json:"ended_at" db:"ended_at"`
	WinningCondition   *string                `json:"winning_condition" db:"winning_condition"`
	WarMachines        []*WarMachineMetadata  `json:"war_machines"`
	WinningWarMachines []*WarMachineMetadata  `json:"winning_war_machines"`
	SpawnedAI          []*WarMachineMetadata  `json:"spawned_ai"`
	GameMap            *GameMap               `json:"map"`
	FactionMap         map[FactionID]*Faction `json:"faction_map"`
	BattleHistory      [][]byte               `json:"battle_history"`

	// used for destroyed notification subscription
	WarMachineDestroyedRecordMap map[byte]*WarMachineDestroyedRecord

	// State
	State BattleState
}

type WarMachineDestroyedRecord struct {
	DestroyedWarMachine *WarMachineMetadata `json:"destroyed_war_machine"`
	KilledByWarMachine  *WarMachineMetadata `json:"killed_by_war_machine,omitempty"`
	KilledBy            string              `json:"killed_by"`
	DamageRecords       []*DamageRecord     `json:"damage_records"`
}

type DamageRecord struct {
	Amount             int                 `json:"amount"` // The total amount of dam\age taken from this source
	CausedByWarMachine *WarMachineMetadata `json:"caused_by_war_machine,omitempty"`
	SourceName         string              `json:"source_name,omitempty"` // The name of the weapon / damage causer (in-case of now TokenID)
}

type FactionWarMachineQueue struct {
	FactionID  FactionID `json:"faction_id"`
	QueueTotal int       `json:"queue_total"`
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
	WarNachuneHash     string              `json:"war_machine_hash" db:"war_machine_hash"`
	FactionID          FactionID           `json:"faction_id" db:"faction_id"`
	WarMachineMetadata *WarMachineMetadata `json:"war_machine_metadata" db:"war_machine_metadata"`
	QueuedAt           time.Time           `json:"queued_at" db:"queued_at"`
	DeletedAt          *time.Time          `json:"deleted_at,omitempty" db:"deleted_at,omitempty"`
	ContractReward     string              `json:"contract_reward" db:"contract_reward"`
	IsInsured          bool                `json:"is_insured" db:"is_insured"`
	Fee                string              `json:"fee" db:"fee"`
	CreatedAt          time.Time           `json:"created_at" db:"created_at"`
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
	RepairModeFast     RepairMode = "FAST"
	RepairModeStandard RepairMode = "STANDARD"
)

type AssetRepairRecord struct {
	Hash              string     `json:"hash" db:"hash"`
	ExpectCompletedAt time.Time  `json:"expect_completed_at" db:"expect_completed_at"`
	RepairMode        RepairMode `json:"repair_mode" db:"repair_mode"`
	IsPaidToComplete  bool       `json:"is_paid_to_complete" db:"is_paid_to_complete"`
	CompletedAt       *time.Time `json:"completed_at,omitempty" db:"completed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type SupremacyQueueUpdateReq struct {
	Hash           string  `json:"hash"`
	Position       *int    `json:"position"`
	ContractReward *string `json:"contract_reward"`
}
