package server

import "server/db/boiler"

type Mech struct {
	Mech    *boiler.Mech
	Chassis *boiler.Chassis
	Weapons map[int]*boiler.Weapon
	Turrets map[int]*boiler.Weapon
	Modules map[int]*boiler.Module
}
type Template struct {
	Template         *boiler.Template
	BlueprintChassis *boiler.BlueprintChassis
	BlueprintWeapons map[int]*boiler.BlueprintWeapon
	BlueprintTurrets map[int]*boiler.BlueprintWeapon
	BlueprintModules map[int]*boiler.BlueprintModule
}
