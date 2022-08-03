package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
)

var XsynTreasuryUserID = UserID(uuid.Must(uuid.FromString("ebf30ca0-875b-4e84-9a78-0b3fa36a1f87")))

const RepairCenterUserID = "a988b1e3-5556-4cad-83bd-d61c2b149cb7"

// User is a single user on the platform
type User struct {
	ID                  UserID      `json:"id" db:"id"`
	Faction             *Faction    `json:"faction"`
	FactionID           FactionID   `json:"faction_id,omitempty"`
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
	MobileNumber        null.String `json:"mobile_number"`

	PublicAddress null.String `json:"public_address,omitempty" db:"public_address"`

	PassportURL string `json:"passport_url"`
	Sups        BigInt `json:"sups"`
	Gid         int    `json:"gid" db:"gid"`

	SyndicateID null.String `json:"syndicate_id,omitempty"`

	// for dev env only
	TwitchID null.String `json:"twitch_id" db:"twitch_id"`
}

func (b *User) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type PassportUser struct {
	ID                               UserID      `json:"id" db:"id"`
	FirstName                        string      `json:"first_name" db:"first_name"`
	LastName                         string      `json:"last_name" db:"last_name"`
	Email                            null.String `json:"email" db:"email"`
	FacebookID                       null.String `json:"facebook_id" db:"facebook_id"`
	GoogleID                         null.String `json:"google_id" db:"google_id"`
	TwitchID                         null.String `json:"twitch_id" db:"twitch_id"`
	TwitterID                        null.String `json:"twitter_id" db:"twitter_id"`
	DiscordID                        null.String `json:"discord_id" db:"discord_id"`
	FactionID                        *FactionID  `json:"faction_id" db:"faction_id"`
	MobileNumber                     null.String `json:"mobile_number" db:"mobile_number"`
	Faction                          *Faction    `json:"faction"`
	Username                         string      `json:"username" db:"username"`
	Verified                         bool        `json:"verified" db:"verified"`
	OldPasswordRequired              bool        `json:"old_password_required" db:"old_password_required"`
	RoleID                           RoleID      `json:"role_id" db:"role_id"`
	Role                             Role        `json:"role" db:"role"`
	AvatarID                         *BlobID     `json:"avatar_id" db:"avatar_id"`
	Sups                             BigInt
	Online                           bool         `json:"online"`
	TwoFactorAuthenticationActivated bool         `json:"two_factor_authentication_activated" db:"two_factor_authentication_activated"`
	TwoFactorAuthenticationSecret    string       `json:"two_factor_authentication_secret" db:"two_factor_authentication_secret"`
	TwoFactorAuthenticationIsSet     bool         `json:"two_factor_authentication_is_set" db:"two_factor_authentication_is_set"`
	HasRecoveryCode                  bool         `json:"has_recovery_code" db:"has_recovery_code"`
	Pass2FA                          bool         `json:"pass_2_fa"`
	Nonce                            null.String  `json:"-" db:"nonce"`
	PublicAddress                    null.String  `json:"public_address,omitempty" db:"public_address"`
	CreatedAt                        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt                        time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt                        *time.Time   `json:"deleted_at" db:"deleted_at"`
	Metadata                         UserMetadata `json:"metadata" db:"metadata"`
}

type UserMetadata struct {
	BoughtStarterWarmachines int  `json:"bought_starter_warmachines"`
	BoughtLootboxes          int  `json:"bought_lootboxes"`
	WatchedVideo             bool `json:"watched_video"`
}

type Organisation struct {
	ID        OrganisationID `json:"id" db:"id"`
	Slug      string         `json:"slug" db:"slug"`
	Name      string         `json:"name" db:"name"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at" db:"deleted_at"`
}

type OrganisationID uuid.UUID

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

type UserBrief struct {
	ID       uuid.UUID       `json:"id"`
	Username string          `json:"username"`
	Gid      int             `json:"gid"`
	Faction  *boiler.Faction `json:"faction"`
}

const SupremacyGameUserID = "4fae8fdf-584f-46bb-9cb9-bb32ae20177e"
const SupremacyChallengeFundUserID = "5bca9b58-a71c-4134-85d4-50106a8966dc"

var (
	SupremacyZaibatsuUserID          = uuid.Must(uuid.FromString("1a657a32-778e-4612-8cc1-14e360665f2b"))
	SupremacyRedMountainUserID       = uuid.Must(uuid.FromString("305da475-53dc-4973-8d78-a30d390d3de5"))
	SupremacyBostonCyberneticsUserID = uuid.Must(uuid.FromString("15f29ee9-e834-4f76-aff8-31e39faabe2d"))
)

type UserStat struct {
	*boiler.PlayerStat
	LastSevenDaysKills int `json:"last_seven_days_kills"`
}
