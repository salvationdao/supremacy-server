package battle

import (
	"server"
	"server/db/boiler"
	"server/multipliers"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
)

type usersMap struct {
	deadlock.RWMutex
	m map[uuid.UUID]*BattleUser
}

func (u *usersMap) Add(bu *BattleUser) {
	u.Lock()
	u.m[bu.ID] = bu
	u.Unlock()
}

func (u *usersMap) Range(fn func(user *BattleUser) bool) {
	u.RLock()
	for _, user := range u.m {
		if !fn(user) {
			break
		}
	}
	u.RUnlock()
}

func (u *usersMap) OnlineUserIDs() []string {
	userIDs := []string{}
	u.RLock()
	for uid := range u.m {
		userIDs = append(userIDs, uid.String())
	}
	u.RUnlock()

	return userIDs
}

func (u *usersMap) User(id uuid.UUID) (*BattleUser, bool) {
	u.RLock()
	b, ok := u.m[id]
	u.RUnlock()
	return b, ok
}

func (u *usersMap) Delete(id uuid.UUID) {
	u.Lock()
	delete(u.m, id)
	u.Unlock()
}

func (um *usersMap) UsersByFactionID(factionID string) []BattleUser {
	um.RLock()
	users := []BattleUser{}
	for _, bu := range um.m {
		if bu.FactionID == factionID {
			users = append(users, *bu)
		}
	}
	um.RUnlock()
	return users
}

type Started struct {
	BattleID           string        `json:"battleID"`
	WarMachines        []*WarMachine `json:"warMachines"`
	WarMachineLocation []byte        `json:"warMachineLocation"`
}

type BattleUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FactionID string    `json:"faction_id"`
	deadlock.RWMutex
}

var FactionNames = map[string]string{
	"98bf7bb3-1a7c-4f21-8843-458d62884060": "RedMountain",
	"7c6dde21-b067-46cf-9e56-155c88a520e2": "Boston",
	"880db344-e405-428d-84e5-6ebebab1fe6d": "Zaibatsu",
}

var FactionLogos = map[string]string{
	"98bf7bb3-1a7c-4f21-8843-458d62884060": "red_mountain_logo",
	"7c6dde21-b067-46cf-9e56-155c88a520e2": "boston_cybernetics_logo",
	"880db344-e405-428d-84e5-6ebebab1fe6d": "zaibatsu_logo",
}

func (bu *BattleUser) AvatarID() string {
	return FactionLogos[bu.FactionID]
}

type BattleEndDetail struct {
	BattleID                     string            `json:"battle_id"`
	BattleIdentifier             int               `json:"battle_identifier"`
	StartedAt                    time.Time         `json:"started_at"`
	EndedAt                      time.Time         `json:"ended_at"`
	WinningCondition             string            `json:"winning_condition"`
	WinningFaction               *boiler.Faction   `json:"winning_faction"`
	WinningWarMachines           []*WarMachine     `json:"winning_war_machines"`
	TopSupsContributors          []*BattleUser     `json:"top_sups_contributors"`
	TopSupsContributeFactions    []*boiler.Faction `json:"top_sups_contribute_factions"`
	MostFrequentAbilityExecutors []*BattleUser     `json:"most_frequent_ability_executors"`
	*MultiplierUpdate            `json:"battle_multipliers"`
}

type MultiplierUpdate struct {
	Battles []*MultiplierUpdateBattles `json:"battles"`
}

type MultiplierUpdateBattles struct {
	BattleNumber     int                             `json:"battle_number"`
	TotalMultipliers string                          `json:"total_multipliers"`
	UserMultipliers  []*multipliers.PlayerMultiplier `json:"multipliers"`
}

type WarMachine struct {
	ID                 string          `json:"id"`
	Hash               string          `json:"hash"`
	ParticipantID      byte            `json:"participantID"`
	FactionID          string          `json:"factionID"`
	MaxHealth          uint32          `json:"maxHealth"`
	Health             uint32          `json:"health"`
	MaxShield          uint32          `json:"maxShield"`
	Shield             uint32          `json:"shield"`
	Energy             uint32          `json:"energy"`
	Stat               *Stat           `json:"stat"`
	ImageAvatar        string          `json:"imageAvatar"`
	Position           *server.Vector3 `json:"position"`
	Rotation           int             `json:"rotation"`
	OwnedByID          string          `json:"ownedByID"`
	Name               string          `json:"name"`
	Description        *string         `json:"description"`
	ExternalUrl        string          `json:"externalUrl"`
	Image              string          `json:"image"`
	Model              string          `json:"model"`
	Skin               string          `json:"skin"`
	ShieldRechargeRate float64         `json:"shieldRechargeRate"`
	Speed              int             `json:"speed"`
	Durability         int             `json:"durability"`
	PowerGrid          int             `json:"powerGrid"`
	CPU                int             `json:"cpu"`
	WeaponHardpoint    int             `json:"weaponHardpoint"`
	TurretHardpoint    int             `json:"turretHardpoint"`
	UtilitySlots       int             `json:"utilitySlots"`
	Faction            *boiler.Faction `json:"faction"`
	WeaponNames        []string        `json:"weaponNames"`
	Abilities          []GameAbility   `json:"abilities"`
	Tier               string          `json:"tier"`
}

type GameAbility struct {
	ID                  string          `json:"id" db:"id"`
	GameClientAbilityID byte            `json:"game_client_ability_id" db:"game_client_ability_id"`
	BattleAbilityID     *string         `json:"battle_ability_id,omitempty" db:"battle_ability_id,omitempty"`
	Colour              string          `json:"colour" db:"colour"`
	TextColour          string          `json:"text_colour" db:"text_colour"`
	Description         string          `json:"description" db:"description"`
	ImageUrl            string          `json:"image_url" db:"image_url"`
	FactionID           string          `json:"faction_id" db:"faction_id"`
	Label               string          `json:"label" db:"label"`
	Level               string          `json:"level" db:"level"`
	SupsCost            decimal.Decimal `json:"sups_cost"`
	CurrentSups         decimal.Decimal `json:"current_sups"`

	// used to track ability price update
	Identity string `json:"identity"`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	WarMachineHash string
	ParticipantID  *byte

	// Category title for frontend to group the abilities together
	Title string `json:"title"`

	CooldownDurationSecond int `json:"cooldown_duration_second"`

	OfferingID uuid.UUID `json:"ability_offering_id"` // for tracking ability trigger
}

type GameAbilityPrice struct {
	GameAbility    *GameAbility
	isReached      bool
	MaxTargetPrice decimal.Decimal
	TargetPrice    decimal.Decimal
	CurrentSups    decimal.Decimal

	TxRefs []string
}
