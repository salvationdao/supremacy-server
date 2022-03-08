package server

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/null/v8"
)

var XsynTreasuryUserID = UserID(uuid.Must(uuid.FromString("ebf30ca0-875b-4e84-9a78-0b3fa36a1f87")))

// User is a single user on the platform
type User struct {
	ID                  UserID      `json:"id" db:"id"`
	Faction             *Faction    `json:"faction"`
	FactionID           FactionID   `json:"faction_id"`
	FirstName           string      `json:"first_name" db:"first_name"`
	LastName            string      `json:"last_name" db:"last_name"`
	Email               null.String `json:"email" db:"email"`
	Username            string      `json:"username" db:"username"`
	Verified            bool        `json:"verified" db:"verified"`
	OldPasswordRequired bool        `json:"old_password_required" db:"old_password_required"`
	RoleID              RoleID      `json:"role_id" db:"role_id"`
	Role                Role        `json:"role" db:"role"`
	AvatarID            *BlobID     `json:"avatar_id" db:"avatar_id"`
	HasRecoveryCode     bool        `json:"has_recovery_code" db:"has_recovery_code"`
	Pass2FA             bool        `json:"pass_2_fa"`
	CreatedAt           time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time  `json:"deleted_at" db:"deleted_at"`

	PublicAddress null.String `json:"public_address,omitempty" db:"public_address"`

	PassportURL string `json:"passport_url"`
	Sups        BigInt `json:"sups"`
	// for dev env only
	TwitchID null.String `json:"twitch_id" db:"twitch_id"`
}

// IssueToken contains token information used for login and verifying accounts
type IssueToken struct {
	ID     IssueTokenID `json:"id" db:"id"`
	UserID UserID       `json:"user_id" db:"user_id"`
}

func (i IssueToken) Whitelisted() bool {
	return !i.ID.IsNil()
}

func (i IssueToken) TokenID() uuid.UUID {
	return uuid.UUID(i.ID)
}

// UserOnlineStatusChange is the payload sent to when a user online status changes
type UserOnlineStatusChange struct {
	ID     UserID `json:"id" db:"id"`
	Online bool   `json:"online"`
}

type BattleUserVote struct {
	BattleID  BattleID `json:"battle_id" db:"battle_id"`
	UserID    UserID   `json:"user_id" db:"user_id"`
	VoteCount int64    `json:"vote_count" db:"vote_count"`
}

type UserStat struct {
	ID                    UserID `json:"id" db:"id"`
	ViewBattleCount       int64  `json:"view_battle_count" db:"view_battle_count"`
	TotalVoteCount        int64  `json:"total_vote_count" db:"total_vote_count"`
	TotalAbilityTriggered int64  `json:"total_ability_triggered" db:"total_ability_triggered"`
	KillCount             int64  `json:"kill_count" db:"kill_count"`
}

type UserBrief struct {
	ID       UserID        `json:"id"`
	Username string        `json:"username"`
	AvatarID *BlobID       `json:"avatar_id,omitempty"`
	Faction  *FactionBrief `json:"faction"`
}

func (u *User) Brief() *UserBrief {
	ub := &UserBrief{
		ID:       u.ID,
		Username: u.Username,
		AvatarID: u.AvatarID,
	}

	if u.Faction != nil {
		ub.Faction = u.Faction.Brief()
	}

	return ub
}

type UserSupsMultiplierSend struct {
	ToUserID        UserID            `json:"to_user_id"`
	ToUserSessionID *hub.SessionID    `json:"to_user_session_id,omitempty"`
	SupsMultipliers []*SupsMultiplier `json:"sups_multiplier"`
}

type SupsMultiplier struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	ExpiredAt time.Time `json:"expired_at"`
}
