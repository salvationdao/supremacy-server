package battle

import (
	"server"
	"server/db/boiler"

	"github.com/gofrs/uuid"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
)

type BattleStage string

const (
	BattleStagStart BattleStage = "START"
	BattleStageEnd  BattleStage = "END"
)

type BattleState struct {
	deadlock.RWMutex
	Stage BattleStage
}

type Battle struct {
	arena       *Arena
	Stage       *BattleState
	ID          uuid.UUID     `json:"battleID" db:"id"`
	MapName     string        `json:"mapName"`
	WarMachines []*WarMachine `json:"warMachines"`
	SpawnedAI   []*WarMachine `json:"SpawnedAI"`
	lastTick    *[]byte
	gameMap     *server.GameMap
	abilities   *AbilitiesSystem
	factions    map[uuid.UUID]*boiler.Faction
}

type Started struct {
	BattleID           string        `json:"battleID"`
	WarMachines        []*WarMachine `json:"warMachines"`
	WarMachineLocation []byte        `json:"warMachineLocation"`
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
	CurrentSups         decimal.Decimal      `json:"current_sups`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	WarMachineHash string
	ParticipantID  *byte

	// Category title for frontend to group the abilities together
	Title string `json:"title"`

	// price locker
	deadlock.RWMutex
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
