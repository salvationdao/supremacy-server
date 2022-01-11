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

type WarMachine struct {
	ID              WarMachineID `json:"id" db:"id"`
	BaseHealthPoint int          `json:"baseHealthPoint" db:"base_health_point"`
	BaseShieldPoint int          `json:"baseShieldPoint" db:"base_shield_point"`
	Name            string       `json:"name" db:"name"`
	TurretHardpint  *int         `json:"turretHardpoint" db:"turret_hardpoint"`
	FactionID       *FactionID   `json:"factionID,omitempty"`
	Faction         *Faction     `json:"faction,omitempty"`
	Position        *Vector3     `json:"position"`
	Rotation        int          `json:"rotation"`
}

type Vector3 struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}
