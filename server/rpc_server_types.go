package server

import (
	"server/db/boiler"
)

type MechContainer struct {
	boiler.Mech
	Chassis   boiler.Chassis            `db:"chassis" json:"chassis"`
	Weapons   map[string]*boiler.Weapon `db:"weapons" json:"weapons"`
	Turrets   map[string]*boiler.Weapon `db:"turrets" json:"turrets"`
	Modules   map[string]*boiler.Module `db:"modules" json:"modules"`
	FactionID string                    `db:"faction_id" json:"faction_id"`
	Faction   *boiler.Faction           `db:"faction" json:"faction"`
	Player    *boiler.Player            `db:"player" json:"player"`
}

type TemplateContainer struct {
	Template         *boiler.Template
	BlueprintChassis *boiler.BlueprintChassis
	BlueprintWeapons map[int]*boiler.BlueprintWeapon
	BlueprintTurrets map[int]*boiler.BlueprintWeapon
	BlueprintModules map[int]*boiler.BlueprintModule
}
