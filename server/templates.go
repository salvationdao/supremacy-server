package server

import (
	"time"
)

type TemplateContainer struct {
	ID                          string    `json:"id"`
	Label                       string    `json:"label"`
	UpdatedAt                   time.Time `json:"updated_at"`
	CreatedAt                   time.Time `json:"created_at"`
	IsGenesis                   bool      `json:"is_genesis"`
	IsLimitedRelease            bool      `json:"is_limited_release"`
	ContainsCompleteMechExactly bool      `json:"contains_complete_mech_exactly"`

	BlueprintMech          []*BlueprintMech          `json:"blueprint_mech,omitempty"`
	BlueprintWeapon        []*BlueprintWeapon        `json:"blueprint_weapon,omitempty"`
	BlueprintWeaponSkin    []*BlueprintWeaponSkin    `json:"blueprint_weapon_skin,omitempty"`
	BlueprintUtility       []*BlueprintUtility       `json:"blueprint_utility,omitempty"`
	BlueprintMechSkin      []*BlueprintMechSkin      `json:"blueprint_mech_skin,omitempty"`
	BlueprintMechAnimation []*BlueprintMechAnimation `json:"blueprint_mech_animation,omitempty"`
	BlueprintPowerCore     []*BlueprintPowerCore     `json:"blueprint_power_core,omitempty"`
	// TODO: AMMO //BlueprintAmmo []*
}
