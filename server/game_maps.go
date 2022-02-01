package server

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
	TokenID         uint64    `json:"tokenID"`
	OwnedByID       UserID    `json:"ownedByID"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ExternalUrl     string    `json:"externalUrl"`
	Image           string    `json:"image"`
	MaxHitPoint     int       `json:"maxHitPoint"`
	RemainHitPoint  int       `json:"remainHitPoint"`
	MaxShield       int       `json:"maxShield"`
	RemainShield    int       `json:"remainShield"`
	Speed           int       `json:"speed"`
	Durability      int       `json:"durability"`
	PowerGrid       int       `json:"powerGrid"`
	CPU             int       `json:"cpu"`
	WeaponHardpoint int       `json:"weaponHardpoint"`
	TurretHardpoint int       `json:"turretHardpoint"`
	UtilitySlots    int       `json:"utilitySlots"`
	FactionID       FactionID `json:"factionID"`
	Faction         *Faction  `json:"faction"`
	Position        *Vector3  `json:"position"`
	Rotation        int       `json:"rotation"`
}

type Vector3 struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}
