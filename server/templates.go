package server

import "time"

type TemplateContainer struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`

	BlueprintMech          []*BlueprintMech          `json:"blueprint_mech,omitempty"`
	BlueprintWeapon        []*BlueprintWeapon        `json:"blueprint_weapon,omitempty"`
	BlueprintUtility       []*BlueprintUtility       `json:"blueprint_utility,omitempty"`
	BlueprintMechSkin      []*BlueprintMechSkin      `json:"blueprint_mech_skin,omitempty"`
	BlueprintMechAnimation []*BlueprintMechAnimation `json:"blueprint_mech_animation,omitempty"`
	BlueprintEnergyCore    []*BlueprintEnergyCore    `json:"blueprint_energy_core,omitempty"`
	// TODO: AMMO //BlueprintAmmo []*
}
