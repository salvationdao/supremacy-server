package gameserver

import (
	"database/sql/driver"

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

// FactionActionID aliases uuid.UUID.
// Doing this prevents situations where you use FactionActionID where it doesn't belong.
type FactionActionID uuid.UUID

// IsNil returns true for a nil uuid.
func (id FactionActionID) IsNil() bool {
	return id == FactionActionID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id FactionActionID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id FactionActionID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *FactionActionID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = FactionActionID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id FactionActionID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *FactionActionID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = FactionActionID(uid)
	// Retrun error
	return err
}

type WarMachineID uuid.UUID

// IsNil returns true for a nil uuid.
func (id WarMachineID) IsNil() bool {
	return id == WarMachineID{}
}

// String aliases UUID.String which returns a canonical RFC-4122 string representation of the UUID.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.String.
func (id WarMachineID) String() string {
	return uuid.UUID(id).String()
}

// MarshalText aliases UUID.MarshalText which implements the encoding.TextMarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.MarshalText.
func (id WarMachineID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText aliases UUID.UnmarshalText which implements the encoding.TextUnmarshaler interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.UnmarshalText.
func (id *WarMachineID) UnmarshalText(text []byte) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.UnmarshalText(text)
	// Convert back to original type
	*id = WarMachineID(uid)
	// Retrun error
	return err
}

// Value aliases UUID.Value which implements the driver.Valuer interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Value.
func (id WarMachineID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements the sql.Scanner interface.
// For more details see https://pkg.go.dev/github.com/gofrs/uuid#UUID.Scan.
func (id *WarMachineID) Scan(src interface{}) error {
	// Convert to uuid.UUID
	uid := uuid.UUID(*id)
	// Unmarshal as uuid.UUID
	err := uid.Scan(src)
	// Convert back to original type
	*id = WarMachineID(uid)
	// Retrun error
	return err
}
