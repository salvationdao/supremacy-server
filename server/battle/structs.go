package battle

import (
	"encoding/json"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
)

type BattleStage string

const (
	BattleStagStart = "START"
	BattleStageEnd  = "END"
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

func (u *usersMap) ForEach(fn func(user *BattleUser) bool) {
	u.RLock()
	for _, user := range u.m {
		if !fn(user) {
			break
		}
	}
	u.RUnlock()
	return
}

func (u *usersMap) Send(key hub.HubCommandKey, payload interface{}, ids ...uuid.UUID) error {
	u.RLock()
	if len(ids) == 0 {
		for _, user := range u.m {
			user.Send(key, payload)
		}
	} else {
		for _, id := range ids {
			if user, ok := u.m[id]; ok {
				user.Send(key, payload)
			} else {
				gamelog.L.Warn().Str("user_id", id.String()).Msg("tried to send user a msg but not in online map")
			}
		}
	}
	u.RUnlock()
	return nil
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

type Battle struct {
	arena       *Arena
	stage       string
	battle      *boiler.Battle
	ID          uuid.UUID     `json:"battleID" db:"id"`
	MapName     string        `json:"mapName"`
	WarMachines []*WarMachine `json:"warMachines"`
	SpawnedAI   []*WarMachine `json:"SpawnedAI"`
	lastTick    *[]byte
	gameMap     *server.GameMap
	abilities   *AbilitiesSystem
	users       usersMap
	factions    map[uuid.UUID]*boiler.Faction
}

type Started struct {
	BattleID           string        `json:"battleID"`
	WarMachines        []*WarMachine `json:"warMachines"`
	WarMachineLocation []byte        `json:"warMachineLocation"`
}

type BattleUser struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	FactionColour string    `json:"faction_colour"`
	FactionID     string    `json:"faction_id"`
	FactionLogoID string    `json:"faction_logo_id"`
	wsClient      map[*hub.Client]bool
	deadlock.RWMutex
}

var FactionLogos = map[string]string{
	"98bf7bb3-1a7c-4f21-8843-458d62884060": "471354c5-d910-4408-852a-6b44b497680f",
	"7c6dde21-b067-46cf-9e56-155c88a520e2": "e1973047-f120-4c36-ba5d-2d1c5100a22f",
	"880db344-e405-428d-84e5-6ebebab1fe6d": "fd3b1345-48e3-43ba-96bb-f0848dc70012",
}

func (bu *BattleUser) AvatarID() string {
	return FactionLogos[bu.FactionID]
}

func (bu *BattleUser) Send(key hub.HubCommandKey, payload interface{}) error {
	if bu.wsClient == nil || len(bu.wsClient) == 0 {
		return fmt.Errorf("user does not have a websocket client")
	}

	b, err := json.Marshal(&BroadcastPayload{
		Key:     key,
		Payload: payload,
	})

	if err != nil {
		return err
	}

	for wsc, _ := range bu.wsClient {
		go wsc.Send(b)
	}
	return nil
}

type Multiplier struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type BattleEndDetail struct {
	BattleID                     string        `json:"battle_id"`
	BattleIdentifier             int           `json:"battle_identifier"`
	StartedAt                    time.Time     `json:"started_at"`
	EndedAt                      time.Time     `json:"ended_at"`
	WinningCondition             string        `json:"winning_condition"`
	WinningFaction               *Faction      `json:"winning_faction"`
	WinningWarMachines           []*WarMachine `json:"winning_war_machines"`
	TopSupsContributors          []*BattleUser `json:"top_sups_contributors"`
	TopSupsContributeFactions    []*Faction    `json:"top_sups_contribute_factions"`
	MostFrequentAbilityExecutors []*BattleUser `json:"most_frequent_ability_executors"`
	*MultiplierUpdate
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
	Stat               *Stat           `json:"stat"`
	ImageAvatar        string          `json:"imageAvatar"`
	Position           *server.Vector3 `json:"position"`
	Rotation           int             `json:"rotation"`
	OwnedByID          string          `json:"ownedByID"`
	Name               string          `json:"name"`
	Description        *string         `json:"description"`
	ExternalUrl        string          `json:"externalUrl"`
	Image              string          `json:"image"`
	Skin               string          `json:"skin"`
	ShieldRechargeRate float64         `json:"shieldRechargeRate"`
	Speed              int             `json:"speed"`
	Durability         int             `json:"durability"`
	PowerGrid          int             `json:"powerGrid"`
	CPU                int             `json:"cpu"`
	WeaponHardpoint    int             `json:"weaponHardpoint"`
	TurretHardpoint    int             `json:"turretHardpoint"`
	UtilitySlots       int             `json:"utilitySlots"`
	Faction            *Faction        `json:"faction"`
	WeaponNames        []string        `json:"weaponNames"`
	Abilities          []*GameAbility  `json:"abilities"`
	Tier               string          `json:"tier"`
}

type GameAbility struct {
	ID                  server.GameAbilityID `json:"id" db:"id"`
	GameClientAbilityID byte                 `json:"game_client_ability_id" db:"game_client_ability_id"`
	BattleAbilityID     *uuid.UUID           `json:"battle_ability_id,omitempty" db:"battle_ability_id,omitempty"`
	Colour              string               `json:"colour" db:"colour"`
	TextColour          string               `json:"text_colour" db:"text_colour"`
	Description         string               `json:"description" db:"description"`
	ImageUrl            string               `json:"image_url" db:"image_url"`
	FactionID           uuid.UUID            `json:"faction_id" db:"faction_id"`
	Label               string               `json:"label" db:"label"`
	SupsCost            decimal.Decimal      `json:"sups_cost"`
	CurrentSups         decimal.Decimal      `json:"current_sups"`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	WarMachineHash string
	ParticipantID  *byte

	// Category title for frontend to group the abilities together
	Title string `json:"title"`

	// price locker
}

type Ability struct {
	ID                uuid.UUID `json:"id" db:"id"`  // used for zaibatsu faction ability
	Identity          uuid.UUID `json:"identity"`    // used to track ability price update
	Colour            string    `json:"colour"`      // used for game ability colour
	TextColour        string    `json:"text_colour"` // used for game ability text colour
	Hash              string    `json:"hash"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	ExternalUrl       string    `json:"external_url"`
	Image             string    `json:"image"`
	SupsCost          string    `json:"sups_cost"`
	GameClientID      int       `json:"game_client_id"`
	RequiredSlot      string    `json:"required_slot"`
	RequiredPowerGrid int       `json:"required_power_grid"`
	RequiredCPU       int       `json:"required_cpu"`
}

type GameAbilityPrice struct {
	GameAbility    *GameAbility
	isReached      bool
	MaxTargetPrice decimal.Decimal
	TargetPrice    decimal.Decimal
	CurrentSups    decimal.Decimal

	TxRefs []string
}

type MultiplierUpdate struct {
	TotalMultipliers string        `json:"total_multipliers"`
	UserMultipliers  []*Multiplier `json:"multipliers"`
}

var fakeMultipliers = []*Multiplier{
	&Multiplier{
		Key:         "citizen",
		Value:       "1x",
		Description: "When a player is within the top 80% of voting average.",
	},
	&Multiplier{
		Key:         "contributor",
		Value:       "5x",
		Description: "When a player is within the top 50% of voting average.",
	},
	&Multiplier{
		Key:         "super contributor",
		Value:       "10x",
		Description: "When a player is within the top 75% of voting average.",
	},
	&Multiplier{
		Key:         "a fool and his money",
		Description: "For a player who has put the most individual SUPS in to vote but still lost.",
		Value:       "5x",
	},
	&Multiplier{
		Key:         "air support",
		Description: "For a player who triggered an airstrike.",
		Value:       "5x",
	},
	&Multiplier{
		Key:         "now i am become death",
		Description: "For a player who triggered a nuke.",
		Value:       "5x",
	},
	&Multiplier{
		Key:         "destroyer of worlds",
		Description: "For a player who has triggered the previous three nukes.",
		Value:       "10x",
	},
}
