package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
)

type Chassis struct {
	boiler.Chassis
}

func (c *Chassis) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(b, c)
}

func (c *Faction) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(b, c)
}

type WeaponMap map[string]*boiler.Weapon

func (c *WeaponMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

type ModuleMap map[string]*boiler.Module

func (c *ModuleMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

func (c *Player) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(b, c)
}

type MechContainer struct {
	boiler.Mech
	Chassis   *Chassis  `db:"chassis" json:"chassis"`
	Weapons   WeaponMap `db:"weapons" json:"weapons"`
	Turrets   WeaponMap `db:"turrets" json:"turrets"`
	Modules   ModuleMap `db:"modules" json:"modules"`
	FactionID string    `db:"faction_id" json:"faction_id"`
	Faction   *Faction  `db:"faction" json:"faction"`
	Player    *Player   `db:"player" json:"player"`
}

type TemplateContainer struct {
	Template         *boiler.Template
	BlueprintChassis *boiler.BlueprintChassis
	BlueprintWeapons map[int]*boiler.BlueprintWeapon
	BlueprintTurrets map[int]*boiler.BlueprintWeapon
	BlueprintModules map[int]*boiler.BlueprintModule
}
