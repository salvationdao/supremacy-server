package server

import (
	"time"

	"github.com/gofrs/uuid"
)

// User is a single user on the platform
type User struct {
	ID                  UserID     `json:"id" db:"id"`
	Faction             *Faction   `json:"faction"`
	FactionID           FactionID  `json:"factionID"`
	FirstName           string     `json:"firstName" db:"first_name"`
	LastName            string     `json:"lastName" db:"last_name"`
	Email               string     `json:"email" db:"email"`
	Username            string     `json:"username" db:"username"`
	Verified            bool       `json:"verified" db:"verified"`
	OldPasswordRequired bool       `json:"oldPasswordRequired" db:"old_password_required"`
	RoleID              RoleID     `json:"roleID" db:"role_id"`
	Role                Role       `json:"role" db:"role"`
	AvatarID            *BlobID    `json:"avatarID" db:"avatar_id"`
	HasRecoveryCode     bool       `json:"hasRecoveryCode" db:"has_recovery_code"`
	Pass2FA             bool       `json:"pass2FA"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt           *time.Time `json:"deletedAt" db:"deleted_at"`

	PassportURL string `json:"passportURL"`
	Sups        int64  `json:"sups"`
	// for dev env only
	TwitchID string `json:"twitchID" db:"twitch_id"`
}

// IssueToken contains token information used for login and verifying accounts
type IssueToken struct {
	ID     IssueTokenID `json:"id" db:"id"`
	UserID UserID       `json:"userId" db:"user_id"`
}

func (i IssueToken) Whitelisted() bool {
	if !i.ID.IsNil() {
		return true
	}
	return false
}

func (i IssueToken) TokenID() uuid.UUID {
	return uuid.UUID(i.ID)
}

// UserOnlineStatusChange is the payload sent to when a user online status changes
type UserOnlineStatusChange struct {
	ID     UserID `json:"id" db:"id"`
	Online bool   `json:"online"`
}
