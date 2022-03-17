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
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	ImageUrl      string    `json:"image_url" db:"image_url"`
	MaxSpawns     int       `json:"max_spawns" db:"max_spawns"`
	Width         int       `json:"width" db:"width"`
	Height        int       `json:"height" db:"height"`
	CellsX        int       `json:"cells_x" db:"cells_x"`
	CellsY        int       `json:"cells_y" db:"cells_y"`
	TopPixels     int       `json:"top" db:"top_pixels"`
	LeftPixels    int       `json:"left" db:"left_pixels"`
	Scale         float64   `json:"scale" db:"scale"`
	DisabledCells []int64   `json:"disabled_cells" db:"disabled_cells"`
}

type WarMachineMetadata struct {
	Hash               string             `json:"hash"`
	ParticipantID      byte               `json:"participant_id"`
	OwnedByID          UserID             `json:"owned_by_id"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	ExternalUrl        string             `json:"external_url"`
	Image              string             `json:"image"`
	Model              string             `json:"model"`
	Skin               string             `json:"skin"`
	MaxHealth          int                `json:"max_health"`
	Health             int                `json:"health"`
	MaxShield          int                `json:"max_shield"`
	Shield             int                `json:"shield"`
	ShieldRechargeRate float64            `json:"shield_recharge_rate"`
	Speed              int                `json:"speed"`
	Durability         int                `json:"durability"`
	PowerGrid          int                `json:"power_grid"`
	CPU                int                `json:"cpu"`
	WeaponHardpoint    int                `json:"weapon_hardpoint"`
	TurretHardpoint    int                `json:"turret_hardpoint"`
	UtilitySlots       int                `json:"utility_slots"`
	FactionID          FactionID          `json:"faction_id"`
	Faction            *Faction           `json:"faction"`
	WeaponNames        []string           `json:"weapon_names"`
	Position           *Vector3           `json:"position"`
	Rotation           int                `json:"rotation"`
	Abilities          []*AbilityMetadata `json:"abilities"`
	ImageAvatar        string             `json:"image_avatar"`

	ContractReward decimal.Decimal `json:"contract_reward"`
	Fee            decimal.Decimal `json:"fee"`
}

type WarMachineBrief struct {
	ImageUrl    string        `json:"image_url"`
	ImageAvatar string        `json:"image_avatar"`
	Name        string        `json:"name"`
	Faction     *FactionBrief `json:"faction"`
}

func (wm *WarMachineMetadata) Brief() *WarMachineBrief {
	wmb := &WarMachineBrief{
		ImageUrl:    wm.Image,
		ImageAvatar: wm.ImageAvatar,
		Name:        wm.Name,
	}

	if wm.Faction != nil {
		wmb.Faction = wm.Faction.Brief()
	}

	return wmb
}

type AbilityMetadata struct {
	ID                GameAbilityID `json:"id" db:"id"`  // used for zaibatsu faction ability
	Identity          uuid.UUID     `json:"identity"`    // used to track ability price update
	Colour            string        `json:"colour"`      // used for game ability colour
	TextColour        string        `json:"text_colour"` // used for game ability text colour
	Hash              string        `json:"hash"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	ExternalUrl       string        `json:"external_url"`
	Image             string        `json:"image"`
	SupsCost          string        `json:"sups_cost"`
	GameClientID      int           `json:"game_client_id"`
	RequiredSlot      string        `json:"required_slot"`
	RequiredPowerGrid int           `json:"required_power_grid"`
	RequiredCPU       int           `json:"required_cpu"`
}

type Vector3 struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type GameAbility struct {
	ID                  uuid.UUID        `json:"id" db:"id"`
	Identity            uuid.UUID        `json:"identity"` // used for tracking ability price
	GameClientAbilityID byte             `json:"game_client_ability_id" db:"game_client_ability_id"`
	BattleAbilityID     *BattleAbilityID `json:"battle_ability_id,omitempty" db:"battle_ability_id,omitempty"`
	Colour              string           `json:"colour" db:"colour"`
	TextColour          string           `json:"text_colour" db:"text_colour"`
	Description         string           `json:"description" db:"description"`
	ImageUrl            string           `json:"image_url" db:"image_url"`
	FactionID           uuid.UUID        `json:"faction_id" db:"faction_id"`
	Label               string           `json:"label" db:"label"`
	SupsCost            string           `json:"sups_cost" db:"sups_cost"`
	CurrentSups         string           `json:"current_sups"`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	AbilityHash    string
	WarMachineHash string
	ParticipantID  *byte

	// Category title for frontend to group the abilities together
	Title string `json:"title"`
}

type AbilityBrief struct {
	Label    string `json:"label"`
	ImageUrl string `json:"image_url"`
	Colour   string `json:"colour"`
}

func (ga *GameAbility) Brief() *AbilityBrief {
	return &AbilityBrief{
		Label:    ga.Label,
		ImageUrl: ga.ImageUrl,
		Colour:   ga.Colour,
	}
}
