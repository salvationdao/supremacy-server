package server

import (
	"database/sql/driver"
	"fmt"
	"github.com/sasha-s/go-deadlock"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/gofrs/uuid"
)

// HubClientID aliases uuid.UUID.
// Doing this prevents situations where you use HubClientID where it doesn't belong.
type HubClientID uuid.UUID

// IsNil returns true for a nil uuid.
func (id HubClientID) IsNil() bool {
	return id == HubClientID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id HubClientID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id HubClientID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *HubClientID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = HubClientID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id HubClientID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *HubClientID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = HubClientID(uid)
	// Retrun error
	return err
}

// BlobID aliases uuid.UUID.
// Doing this prevents situations where you use BlobID where it doesn't belong.
type BlobID uuid.UUID

// IsNil returns true for a nil uuid.
func (id BlobID) IsNil() bool {
	return id == BlobID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id BlobID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id BlobID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *BlobID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = BlobID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id BlobID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *BlobID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = BlobID(uid)
	// Retrun error
	return err
}

// UserID aliases uuid.UUID.
// Doing this prevents situations where you use UserID where it doesn't belong.
type UserID uuid.UUID

// IsNil returns true for a nil uuid.
func (id UserID) IsNil() bool {
	return id == UserID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id UserID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id UserID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *UserID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = UserID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id UserID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *UserID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = UserID(uid)
	// Retrun error
	return err
}

// IssueTokenID aliases uuid.UUID.
// Doing this prevents situations where you use IssueTokenID where it doesn't belong.
type IssueTokenID uuid.UUID

// IsNil returns true for a nil uuid.
func (id IssueTokenID) IsNil() bool {
	return id == IssueTokenID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id IssueTokenID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id IssueTokenID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *IssueTokenID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = IssueTokenID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id IssueTokenID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *IssueTokenID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = IssueTokenID(uid)
	// Retrun error
	return err
}

// RoleID aliases uuid.UUID.
// Doing this prevents situations where you use RoleID where it doesn't belong.
type RoleID uuid.UUID

// IsNil returns true for a nil uuid.
func (id RoleID) IsNil() bool {
	return id == RoleID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id RoleID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id RoleID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *RoleID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = RoleID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id RoleID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *RoleID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = RoleID(uid)
	// Retrun error
	return err
}

//func (f FactionID) IsValid() bool {
//	switch f {
//	case BostonCyberneticsFactionID, RedMountainFactionID, ZaibatsuFactionID, FactionID(uuid.Nil):
//		return true
//	}
//	return false
//}

// FactionID aliases uuid.UUID.
// Doing this prevents situations where you use FactionID where it doesn't belong.
type FactionID uuid.UUID

// IsNil returns true for a nil uuid.
func (id FactionID) IsNil() bool {
	return id == FactionID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id FactionID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id FactionID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *FactionID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = FactionID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id FactionID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *FactionID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = FactionID(uid)
	// Retrun error
	return err
}

type PlayerAbilityKills struct {
	ID        string `json:"id" db:"id"`
	FactionID string `json:"faction_id" db:"faction_id"`
	KillCount int64  `json:"kill_count" db:"kill_count"`
}

// GameAbilityID aliases uuid.UUID.
// Doing this prevents situations where you use GameAbilityID where it doesn't belong.
type GameAbilityID uuid.UUID

// IsNil returns true for a nil uuid.
func (id GameAbilityID) IsNil() bool {
	return id == GameAbilityID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id GameAbilityID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id GameAbilityID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *GameAbilityID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = GameAbilityID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id GameAbilityID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *GameAbilityID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = GameAbilityID(uid)
	// Retrun error
	return err
}

type GameMapID uuid.UUID

// IsNil returns true for a nil uuid.
func (id GameMapID) IsNil() bool {
	return id == GameMapID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id GameMapID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id GameMapID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *GameMapID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = GameMapID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id GameMapID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *GameMapID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = GameMapID(uid)
	// Retrun error
	return err
}

type WarMachineDestroyedEventID uuid.UUID

// IsNil returns true for a nil uuid.
func (id WarMachineDestroyedEventID) IsNil() bool {
	return id == WarMachineDestroyedEventID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id WarMachineDestroyedEventID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id WarMachineDestroyedEventID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *WarMachineDestroyedEventID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = WarMachineDestroyedEventID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id WarMachineDestroyedEventID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *WarMachineDestroyedEventID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = WarMachineDestroyedEventID(uid)
	// Retrun error
	return err
}

type GameAbilityEventID uuid.UUID

// IsNil returns true for a nil uuid.
func (id GameAbilityEventID) IsNil() bool {
	return id == GameAbilityEventID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id GameAbilityEventID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id GameAbilityEventID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *GameAbilityEventID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = GameAbilityEventID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id GameAbilityEventID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *GameAbilityEventID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = GameAbilityEventID(uid)
	// Retrun error
	return err
}

type BattleAbilityID uuid.UUID

// IsNil returns true for a nil uuid.
func (id BattleAbilityID) IsNil() bool {
	return id == BattleAbilityID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id BattleAbilityID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id BattleAbilityID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *BattleAbilityID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = BattleAbilityID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id BattleAbilityID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *BattleAbilityID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = BattleAbilityID(uid)
	// Retrun error
	return err
}

// BigInt aliases big.Int
// We do this, so we can create .scan methods.
type BigInt struct {
	big.Int
}

func (b BigInt) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *BigInt) UnmarshalJSON(text []byte) error {
	if text == nil {
		*b = BigInt{}
		return nil
	}
	// remove extra ""s
	str := string(text)
	str = strings.Replace(str, `"`, ``, -1)
	_, ok := b.Int.SetString(str, 10)
	if !ok {
		return fmt.Errorf("invalid number %s", string(text))
	}
	return nil
}

type TAtomBool struct{ flag int32 }

func (b *TAtomBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), int32(i))
}

func (b *TAtomBool) Get() bool {
	return atomic.LoadInt32(&(b.flag)) != 0
}

type CellLocation struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type GameLocation struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type GameAbilityEvent struct {
	EventID             uuid.UUID     `json:"eventID" db:"event_id"`
	GameAbilityID       *uuid.UUID    `json:"gameAbilityID,omitempty" db:"game_ability_id,omitempty"`
	AbilityHash         *string       `json:"abilityHash,omitempty" db:"ability_hash,omitempty"`
	GameClientAbilityID byte          `json:"gameClientAbilityID" db:"game_client_ability_id"`
	ParticipantID       *byte         `json:"participantID,omitempty" db:"participant_id"`
	WarMachineHash      *string       `json:"warMachineHash,omitempty"`
	FactionID           *string       `json:"factionID,omitempty"`
	IsTriggered         bool          `json:"isTriggered" db:"is_triggered"`
	TriggeredByUserID   *uuid.UUID    `json:"TriggeredByUserID,omitempty" db:"triggered_by_user_id,omitempty"`
	TriggeredByUsername *string       `json:"triggeredByUsername"`
	TriggeredOnCellX    *int          `json:"triggeredOnCellX,omitempty" db:"triggered_on_cell_x,omitempty"`
	TriggeredOnCellY    *int          `json:"triggeredOnCellY,omitempty" db:"triggered_on_cell_y,omitempty"`
	GameAbility         *GameAbility  `json:"gameAbility,omitempty"`
	GameLocation        *GameLocation `json:"gameLocation"`
	GameLocationEnd     *GameLocation `json:"gameLocationEnd"`
}

type BattleZone struct {
	Location   GameLocation `json:"location"`
	Radius     int          `json:"radius"`
	Time       int          `json:"time"`
	ShrinkTime int          `json:"shrinkTime"`
}

var env string
var lock = deadlock.RWMutex{}

func SetEnv(environment string) {
	lock.Lock()
	defer lock.Unlock()
	env = environment
}

func Env() string {
	lock.RLock()
	defer lock.RUnlock()
	return env
}

func IsProductionEnv() bool {
	lock.RLock()
	defer lock.RUnlock()
	return env == "production"
}

type OnChainStatus string

const (
	OnChainStatusMintable   OnChainStatus = "MINTABLE"
	OnChainStatusStakable   OnChainStatus = "STAKABLE"
	OnChainStatusUnstakable OnChainStatus = "UNSTAKABLE"
)
