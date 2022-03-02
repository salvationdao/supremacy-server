package server

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const GameClientTileSize = 2000

// To get the location in game its
//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels

type GameMap struct {
	ID            GameMapID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	ImageUrl      string    `json:"imageUrl" db:"image_url"`
	MaxSpawns     int       `json:"maxSpawns" db:"max_spawns"`
	Width         int       `json:"width" db:"width"`
	Height        int       `json:"height" db:"height"`
	CellsX        int       `json:"cellsX" db:"cells_x"`
	CellsY        int       `json:"cellsY" db:"cells_y"`
	TopPixels     int       `json:"top" db:"top_pixels"`
	LeftPixels    int       `json:"left" db:"left_pixels"`
	Scale         float64   `json:"scale" db:"scale"`
	DisabledCells []int     `json:"disabledCells" db:"disabled_cells"`
}

type WarMachineMetadata struct {
	Hash               string             `json:"hash"`
	ParticipantID      byte               `json:"participantID"`
	OwnedByID          UserID             `json:"ownedByID"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	ExternalUrl        string             `json:"externalUrl"`
	Image              string             `json:"image"`
	Model              string             `json:"model"`
	Skin               string             `json:"skin"`
	MaxHealth          int                `json:"maxHealth"`
	Health             int                `json:"health"`
	MaxShield          int                `json:"maxShield"`
	Shield             int                `json:"shield"`
	ShieldRechargeRate float64            `json:"shieldRechargeRate"`
	Speed              int                `json:"speed"`
	Durability         int                `json:"durability"`
	PowerGrid          int                `json:"powerGrid"`
	CPU                int                `json:"cpu"`
	WeaponHardpoint    int                `json:"weaponHardpoint"`
	TurretHardpoint    int                `json:"turretHardpoint"`
	UtilitySlots       int                `json:"utilitySlots"`
	FactionID          FactionID          `json:"factionID"`
	Faction            *Faction           `json:"faction"`
	WeaponNames        []string           `json:"weaponNames"`
	Position           *Vector3           `json:"position"`
	Rotation           int                `json:"rotation"`
	Abilities          []*AbilityMetadata `json:"abilities"`

	ContractReward decimal.Decimal `json:"contractReward"`
}

type WarMachineBrief struct {
	ImageUrl string        `json:"imageUrl"`
	Name     string        `json:"name"`
	Faction  *FactionBrief `json:"faction"`
}

func (wm *WarMachineMetadata) Brief() *WarMachineBrief {
	wmb := &WarMachineBrief{
		ImageUrl: wm.Image,
		Name:     wm.Name,
	}

	if wm.Faction != nil {
		wmb.Faction = wm.Faction.Brief()
	}

	return wmb
}

type AbilityMetadata struct {
	ID                GameAbilityID `json:"id" db:"id"` // used for zaibatsu faction ability
	Identity          uuid.UUID     `json:"identity"`   // used to track ability price update
	Colour            string        `json:"colour"`     // used for game ability colour
	TextColour        string        `json:"textColour"` // used for game ability text colour
	Hash              string        `json:"hash"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	ExternalUrl       string        `json:"externalUrl"`
	Image             string        `json:"image"`
	SupsCost          string        `json:"supsCost"`
	GameClientID      int           `json:"gameClientID"`
	RequiredSlot      string        `json:"requiredSlot"`
	RequiredPowerGrid int           `json:"requiredPowerGrid"`
	RequiredCPU       int           `json:"requiredCPU"`
}

type Vector3 struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type GameAbility struct {
	ID                  GameAbilityID    `json:"id" db:"id"`
	Identity            uuid.UUID        `json:"identity"` // used for tracking ability price
	GameClientAbilityID byte             `json:"gameClientAbilityID" db:"game_client_ability_id"`
	BattleAbilityID     *BattleAbilityID `json:"battleAbilityID,omitempty" db:"battle_ability_id,omitempty"`
	Colour              string           `json:"colour" db:"colour"`
	TextColour          string           `json:"textColour" db:"text_colour"`
	Description         string           `json:"description" db:"description"`
	ImageUrl            string           `json:"imageUrl" db:"image_url"`
	FactionID           FactionID        `json:"factionID" db:"faction_id"`
	Label               string           `json:"label" db:"label"`
	SupsCost            string           `json:"supsCost" db:"sups_cost"`
	CurrentSups         string           `json:"currentSups"`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	AbilityHash    string
	WarMachineHash string
	ParticipantID  *byte

	// Category title for frontend to group the abilities together
	Title string `json:"title"`
}

type AbilityBrief struct {
	Label    string `json:"label"`
	ImageUrl string `json:"imageUrl"`
	Colour   string `json:"colour"`
}

func (ga *GameAbility) Brief() *AbilityBrief {
	return &AbilityBrief{
		Label:    ga.Label,
		ImageUrl: ga.ImageUrl,
		Colour:   ga.Colour,
	}
}
