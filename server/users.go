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
	Sups        BigInt `json:"sups"`
	// for dev env only
	TwitchID string `json:"twitchID" db:"twitch_id"`
}

// IssueToken contains token information used for login and verifying accounts
type IssueToken struct {
	ID     IssueTokenID `json:"id" db:"id"`
	UserID UserID       `json:"userID" db:"user_id"`
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
	BattleID  BattleID `json:"battleID" db:"battle_id"`
	UserID    UserID   `json:"userID" db:"user_id"`
	VoteCount int64    `json:"voteCount" db:"vote_count"`
}

type UserStat struct {
	ID                    UserID `json:"id" db:"id"`
	ViewBattleCount       int64  `json:"viewBattleCount" db:"view_battle_count"`
	TotalVoteCount        int64  `json:"totalVoteCount" db:"total_vote_count"`
	TotalAbilityTriggered int64  `json:"totalAbilityTriggered" db:"total_ability_triggered"`
	KillCount             int64  `json:"killCount" db:"kill_count"`
}

type UserBrief struct {
	ID       UserID        `json:"id"`
	Username string        `json:"username"`
	AvatarID *BlobID       `json:"avatarID,omitempty"`
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
