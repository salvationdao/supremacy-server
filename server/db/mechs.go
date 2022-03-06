package db

import (
	"encoding/hex"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/speps/go-hashids/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func MechsByOwnerID(ownerID uuid.UUID) ([]*server.Mech, error) {
	mechs, err := boiler.Mechs(boiler.MechWhere.OwnerID.EQ(ownerID.String())).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*server.Mech{}
	for _, mech := range mechs {
		record, err := Mech(uuid.Must(uuid.FromString(mech.ID)))
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}
func MechSetName(mechID uuid.UUID, name string) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	mech, err := boiler.FindMech(gamedb.StdConn, mechID.String())
	if err != nil {
		return err
	}
	mech.Name = name
	_, err = mech.Update(tx, boil.Whitelist(boiler.MechColumns.Name))
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
func MechSetOwner(mechID uuid.UUID, ownerID uuid.UUID) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	mech, err := boiler.FindMech(gamedb.StdConn, mechID.String())
	if err != nil {
		return err
	}
	mech.OwnerID = ownerID.String()
	_, err = mech.Update(tx, boil.Whitelist(boiler.MechColumns.OwnerID))
	if err != nil {
		return err
	}
	tx.Commit()
	return nil

}
func Template(templateID uuid.UUID) (*server.Template, error) {
	template, err := boiler.FindTemplate(gamedb.StdConn, templateID.String())
	if err != nil {
		return nil, err
	}
	chassis, err := template.BlueprintChassis().One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	blueprintWeaponJoins, err := boiler.BlueprintChassisBlueprintWeapons(
		qm.Where("chassis_id = ? AND mount_location = 'ARM'", chassis.ID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	blueprintWeapons := map[int]*boiler.BlueprintWeapon{}
	for _, join := range blueprintWeaponJoins {
		weapon, err := boiler.FindBlueprintWeapon(gamedb.StdConn, join.BlueprintWeaponID)
		if err != nil {
			return nil, err
		}
		blueprintWeapons[join.SlotNumber] = weapon
	}

	blueprintTurretJoins, err := boiler.BlueprintChassisBlueprintWeapons(
		qm.Where("chassis_id = ? AND mount_location = 'TURRET'", chassis.ID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	blueprintTurrets := map[int]*boiler.BlueprintWeapon{}
	for _, join := range blueprintTurretJoins {
		blueprintTurret, err := boiler.FindBlueprintWeapon(gamedb.StdConn, join.BlueprintWeaponID)
		if err != nil {
			return nil, err
		}
		blueprintTurrets[join.SlotNumber] = blueprintTurret
	}

	blueprintModuleJoins, err := boiler.BlueprintChassisBlueprintModules(
		boiler.ChassisModuleWhere.ChassisID.EQ(chassis.ID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	blueprintModules := map[int]*boiler.BlueprintModule{}
	for _, join := range blueprintModuleJoins {
		blueprintModule, err := boiler.FindBlueprintModule(gamedb.StdConn, join.BlueprintModuleID)
		if err != nil {
			return nil, err
		}
		blueprintModules[join.SlotNumber] = blueprintModule
	}

	result := &server.Template{
		Template:         template,
		BlueprintChassis: chassis,
		BlueprintWeapons: blueprintWeapons,
		BlueprintTurrets: blueprintTurrets,
		BlueprintModules: blueprintModules,
	}

	return result, nil
}
func TemplatePurchasedCount(templateID uuid.UUID) (int, error) {
	count, err := boiler.Mechs(boiler.MechWhere.TemplateID.EQ(templateID.String())).Count(gamedb.StdConn)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
func TemplatesByFactionID(factionID uuid.UUID) ([]*server.Template, error) {
	templates, err := boiler.Templates().All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*server.Template{}
	for _, tpl := range templates {
		template, err := Template(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return nil, err
		}
		result = append(result, template)
	}

	return result, nil
}

func Mech(mechID uuid.UUID) (*server.Mech, error) {
	mech, err := boiler.FindMech(gamedb.StdConn, mechID.String())
	if err != nil {
		return nil, err
	}
	chassis, err := mech.Chassis().One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	weaponJoins, err := boiler.ChassisWeapons(
		qm.Where("chassis_id = ? AND mount_location = 'ARM'", chassis.ID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	weapons := map[int]*boiler.Weapon{}
	for _, join := range weaponJoins {
		weapon, err := boiler.FindWeapon(gamedb.StdConn, join.WeaponID)
		if err != nil {
			return nil, err
		}
		weapons[join.SlotNumber] = weapon
	}

	turretJoins, err := boiler.ChassisWeapons(
		qm.Where("chassis_id = ? AND mount_location = 'TURRET'", chassis.ID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	turrets := map[int]*boiler.Weapon{}
	for _, join := range turretJoins {
		turret, err := boiler.FindWeapon(gamedb.StdConn, join.WeaponID)
		if err != nil {
			return nil, err
		}
		turrets[join.SlotNumber] = turret
	}

	moduleJoins, err := boiler.ChassisModules(
		boiler.ChassisModuleWhere.ChassisID.EQ(chassis.ID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	modules := map[int]*boiler.Module{}
	for _, join := range moduleJoins {
		module, err := boiler.FindModule(gamedb.StdConn, join.ModuleID)
		if err != nil {
			return nil, err
		}
		modules[join.SlotNumber] = module
	}

	result := &server.Mech{
		Mech:    mech,
		Chassis: chassis,
		Weapons: weapons,
		Turrets: turrets,
		Modules: modules,
	}

	return result, nil
}

// MechRegister copies everything out of a template into a new mech
func MechRegister(templateID uuid.UUID, ownerID uuid.UUID) (uuid.UUID, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()
	exists, err := boiler.PlayerExists(tx, ownerID.String())
	if err != nil {
		return uuid.Nil, err
	}
	if !exists {
		newPlayer := &boiler.Player{ID: ownerID.String()}
		err = newPlayer.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, err
		}
	}
	template, err := boiler.FindTemplate(tx, templateID.String())
	if err != nil {
		return uuid.Nil, err
	}

	blueprintChassis, err := template.BlueprintChassis().One(tx)
	if err != nil {
		return uuid.Nil, err
	}
	chassis := &boiler.Chassis{
		BrandID:            blueprintChassis.BrandID,
		Label:              blueprintChassis.Label,
		Model:              blueprintChassis.Model,
		Skin:               blueprintChassis.Skin,
		Slug:               blueprintChassis.Slug,
		ShieldRechargeRate: blueprintChassis.ShieldRechargeRate,
		HealthRemaining:    blueprintChassis.MaxHitpoints,
		WeaponHardpoints:   blueprintChassis.WeaponHardpoints,
		TurretHardpoints:   blueprintChassis.TurretHardpoints,
		UtilitySlots:       blueprintChassis.UtilitySlots,
		Speed:              blueprintChassis.Speed,
		MaxHitpoints:       blueprintChassis.MaxHitpoints,
		MaxShield:          blueprintChassis.MaxShield,
	}
	err = chassis.Insert(tx, boil.Infer())
	if err != nil {
		return uuid.Nil, err
	}

	weaponJoins, err := boiler.BlueprintChassisBlueprintWeapons(boiler.BlueprintChassisBlueprintWeaponWhere.BlueprintChassisID.EQ(template.BlueprintChassisID)).All(tx)
	if err != nil {
		return uuid.Nil, err
	}
	for _, join := range weaponJoins {
		blueprintWeapon, err := boiler.FindBlueprintWeapon(tx, join.BlueprintWeaponID)
		if err != nil {
			return uuid.Nil, err
		}
		newWeapon := &boiler.Weapon{
			BrandID:    blueprintWeapon.BrandID,
			Label:      blueprintWeapon.Label,
			Slug:       blueprintWeapon.Slug,
			Damage:     blueprintWeapon.Damage,
			WeaponType: blueprintWeapon.WeaponType,
		}
		err = newWeapon.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, err
		}
		newJoin := &boiler.ChassisWeapon{
			ChassisID:     chassis.ID,
			WeaponID:      newWeapon.ID,
			SlotNumber:    join.SlotNumber,
			MountLocation: join.MountLocation,
		}
		err = newJoin.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, err
		}
	}
	moduleJoins, err := boiler.BlueprintChassisBlueprintModules(boiler.BlueprintChassisBlueprintModuleWhere.BlueprintChassisID.EQ(template.BlueprintChassisID)).All(tx)
	if err != nil {
		return uuid.Nil, err
	}
	for _, join := range moduleJoins {
		blueprintModule, err := boiler.FindBlueprintModule(tx, join.BlueprintModuleID)
		if err != nil {
			return uuid.Nil, err
		}
		newModule := &boiler.Module{
			BrandID:          blueprintModule.BrandID,
			Slug:             blueprintModule.Slug,
			Label:            blueprintModule.Label,
			HitpointModifier: blueprintModule.HitpointModifier,
			ShieldModifier:   blueprintModule.ShieldModifier,
		}
		err = newModule.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, err
		}
		newJoin := &boiler.ChassisModule{
			ChassisID:  chassis.ID,
			ModuleID:   newModule.ID,
			SlotNumber: join.SlotNumber,
		}
		err = newJoin.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, err
		}
	}

	newMechID, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, err
	}
	newChassisID, err := uuid.FromString(chassis.ID)
	if err != nil {
		return uuid.Nil, err
	}
	mechHash, err := GenerateHashID(newMechID, newChassisID)
	if err != nil {
		return uuid.Nil, err
	}
	newMech := &boiler.Mech{
		ID:           newMechID.String(),
		OwnerID:      ownerID.String(),
		TemplateID:   templateID.String(),
		ChassisID:    chassis.ID,
		Tier:         template.Tier,
		IsDefault:    template.IsDefault,
		ImageURL:     template.ImageURL,
		AnimationURL: template.AnimationURL,
		Hash:         mechHash,
		Name:         "",
		Label:        template.Label,
		Slug:         template.Slug,
	}
	err = newMech.Insert(tx, boil.Infer())
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.FromString(newMech.ID)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
func GenerateHashID(mechID uuid.UUID, chassisID uuid.UUID) (string, error) {
	hd := hashids.NewData()
	hd.Salt = mechID.String()
	hd.MinLength = 10
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", terror.Error(err)
	}

	e, err := h.EncodeHex(hex.EncodeToString(chassisID.Bytes()))
	if err != nil {
		return "", terror.Error(err)
	}
	_, err = h.DecodeWithError(e)
	if err != nil {
		return "", terror.Error(err)
	}

	return e, nil
}
