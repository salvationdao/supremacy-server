package server

const GameClientTileSize = 2000

// To get the location in game its
//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels

type GameMap struct {
	ID            GameMapID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	ImageUrl      string    `json:"imageUrl" db:"image_url"`
	Width         int       `json:"width" db:"width"`
	Height        int       `json:"height" db:"height"`
	CellsX        int       `json:"cellsX" db:"cells_x"`
	CellsY        int       `json:"cellsY" db:"cells_y"`
	TopPixels     int       `json:"top" db:"top_pixels"`
	LeftPixels    int       `json:"left" db:"left_pixels"`
	Scale         float64   `json:"scale" db:"scale"`
	DisabledCells []int     `json:"disabledCells" db:"disabled_cells"`
}

type WarMachineNFT struct {
	TokenID         uint64        `json:"tokenID"`
	ParticipantID   byte          `json:"participantID"`
	OwnedByID       UserID        `json:"ownedByID"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	ExternalUrl     string        `json:"externalUrl"`
	Image           string        `json:"image"`
	MaxHealth       int           `json:"maxHealth"`
	Health          int           `json:"health"`
	MaxShield       int           `json:"maxShield"`
	Shield          int           `json:"shield"`
	Speed           int           `json:"speed"`
	Durability      int           `json:"durability"`
	PowerGrid       int           `json:"powerGrid"`
	CPU             int           `json:"cpu"`
	WeaponHardpoint int           `json:"weaponHardpoint"`
	TurretHardpoint int           `json:"turretHardpoint"`
	UtilitySlots    int           `json:"utilitySlots"`
	FactionID       FactionID     `json:"factionID"`
	Faction         *Faction      `json:"faction"`
	WeaponNames     []string      `json:"weaponNames"`
	Position        *Vector3      `json:"position"`
	Rotation        int           `json:"rotation"`
	Abilities       []*AbilityNFT `json:"abilities"`
}

type AbilityNFT struct {
	TokenID           uint64 `json:"tokenID"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ExternalUrl       string `json:"externalUrl"`
	Image             string `json:"image"`
	SupsCost          string `json:"supsCost"`
	GameClientID      int    `json:"gameClientID"`
	RequiredSlot      string `json:"requiredSlot"`
	RequiredPowerGrid int    `json:"requiredPowerGrid"`
	RequiredCPU       int    `json:"requiredCPU"`
}

type Vector3 struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}
