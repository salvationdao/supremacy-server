package supermigrate

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func isDefaultHash(in string) bool {
	defaultHashes := []string{"ZXga92AmGD", "dbYaD4a0Zj", "l7epj2pPL4", "kN7aVgAenK", "wdBAN1aeo5", "018pkXaRWM", "B8x3qdAy6K", "D16aRep0Zo", "4Q1p8dpqwX"}
	for _, defaultHash := range defaultHashes {
		if in == defaultHash {
			return true
		}
	}
	return false
}

func ProcessMech(tx *sql.Tx, data *AssetPayload, metadata *MetadataPayload) (bool, bool, error) {
	att := GetAttributes(metadata.Attributes)

	if isDefaultHash(data.MetadataHash) {
		return true, false, nil
	}
	exists, err := boiler.PlayerExists(tx, data.UserID)
	if err != nil {
		return false, false, nil
	}
	if !exists {
		newPlayer := &boiler.Player{ID: data.UserID}
		err = newPlayer.Insert(tx, boil.Infer())
		if err != nil {
			return false, false, nil
		}
	}
	mechExists, err := boiler.Mechs(boiler.MechWhere.Hash.EQ(data.MetadataHash)).Exists(tx)
	if err != nil {
		return false, false, fmt.Errorf("check mech exist %s: %w", data.MetadataHash, err)
	}
	if mechExists {
		// Update instead of processing all the damn pieces
		existingMech, err := boiler.Mechs(boiler.MechWhere.Hash.EQ(data.MetadataHash)).One(tx)
		if err != nil {
			return false, false, fmt.Errorf("get existing mech: %w", err)
		}

		existingMech.OwnerID = data.UserID
		existingMech.Name = att.Name

		_, err = existingMech.Update(tx, boil.Whitelist(boiler.MechColumns.OwnerID, boiler.MechColumns.Name))
		if err != nil {
			return false, false, fmt.Errorf("update mech: %w", err)
		}
		return false, true, nil
	}

	label, _ := TemplateLabelSlug(att.Brand, att.Model, att.SubModel)
	if label == "Zaibatsu Tenshi Mk1 Gundam" {
		label = "Zaibatsu Tenshi Mk1 Gumdan"
	}
	templateExists, err := boiler.Templates(boiler.TemplateWhere.Label.EQ(label)).Exists(tx)
	if err != nil {
		return false, false, fmt.Errorf("check template exist: %w", err)
	}
	if !templateExists {
		return false, false, fmt.Errorf("matching template does not exist: %s", label)
	}
	label, _ = TemplateLabelSlug(att.Brand, att.Model, att.SubModel)
	if label == "Zaibatsu Tenshi Mk1 Gundam" {
		label = "Zaibatsu Tenshi Mk1 Gumdan"
	}
	template, err := boiler.Templates(boiler.TemplateWhere.Label.EQ(label)).One(tx)
	if err != nil {
		return false, false, fmt.Errorf("check mech exist %s: %w", label, err)
	}
	err = template.L.LoadBlueprintChassis(tx, true, template, nil)
	if err != nil {
		return false, false, fmt.Errorf("load blueprint chassis: %w", err)
	}

	brandExists, err := boiler.Brands(qm.Where("label = ?", BrandMap[att.Brand])).Exists(tx)
	if err != nil {
		return false, false, fmt.Errorf("check brand exist: %w", err)
	}
	if !brandExists {
		return false, false, fmt.Errorf("brand does not exist: %s", att.Brand)
	}
	brand, err := boiler.Brands(qm.Where("label = ?", BrandMap[att.Brand])).One(tx)
	if err != nil {
		return false, false, fmt.Errorf("get brand: %w", err)
	}
	chassis, err := ProcessChassis(brand, metadata.Attributes)
	if err != nil {
		return false, false, fmt.Errorf("process chassis: %w", err)
	}

	weapon1, err := ProcessWeapon("ARM", 1, brand, metadata.Attributes)
	if err != nil {
		return false, false, fmt.Errorf("process weapon: %w", err)
	}

	weapon2, err := ProcessWeapon("ARM", 2, brand, metadata.Attributes)
	if err != nil {
		return false, false, fmt.Errorf("process weapon: %w", err)
	}

	turret1, err := ProcessWeapon("TURRET", 1, brand, metadata.Attributes)
	if err != nil {
		return false, false, fmt.Errorf("process weapon: %w", err)
	}

	turret2, err := ProcessWeapon("TURRET", 2, brand, metadata.Attributes)
	if err != nil {
		return false, false, fmt.Errorf("process weapon: %w", err)
	}

	module, err := ProcessModule(brand, metadata.Attributes)
	if err != nil {
		return false, false, fmt.Errorf("process module: %w", err)
	}

	err = chassis.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert chassis: %w", err)
	}
	externalTokenID, err := strconv.Atoi(data.ExternalTokenID)
	if err != nil {
		return false, false, fmt.Errorf("parse external token ID: %w", err)
	}
	label, slug := MechLabelSlug(att.Brand, att.Model, att.SubModel)
	newMech := &boiler.Mech{
		ID:               uuid.Must(uuid.NewV4()).String(),
		ImageURL:         metadata.Image,
		AnimationURL:     metadata.AnimationURL,
		AvatarURL:        template.AvatarURL,
		CardAnimationURL: template.CardAnimationURL,
		ExternalTokenID:  externalTokenID,
		Tier:             strings.ToUpper(strings.ReplaceAll(att.Rarity, " ", "_")),
		OwnerID:          data.UserID,
		TemplateID:       template.ID,
		ChassisID:        chassis.ID,
		Hash:             data.MetadataHash,
		Name:             att.Name,
		Label:            label,
		Slug:             slug,
		AssetType:        att.AssetType,
	}

	err = newMech.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert mech: %w", err)
	}

	err = weapon1.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert weapon 1: %w", err)
	}
	join1 := &boiler.ChassisWeapon{
		WeaponID:      weapon1.ID,
		ChassisID:     chassis.ID,
		MountLocation: weapon1.WeaponType,
		SlotNumber:    0,
	}
	err = join1.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert weapon 1 join : %w", err)
	}
	err = weapon2.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert weapon 2: %w", err)
	}
	join2 := &boiler.ChassisWeapon{
		WeaponID:      weapon2.ID,
		ChassisID:     chassis.ID,
		MountLocation: weapon2.WeaponType,
		SlotNumber:    1,
	}
	err = join2.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert weapon 2 join : %w", err)
	}

	if turret1 != nil {
		err = turret1.Insert(tx, boil.Infer())
		if err != nil {
			return false, false, fmt.Errorf("insert turret 1: %w", err)
		}
		join := &boiler.ChassisWeapon{
			WeaponID:      turret1.ID,
			ChassisID:     chassis.ID,
			MountLocation: turret1.WeaponType,
			SlotNumber:    0,
		}
		err = join.Insert(tx, boil.Infer())
		if err != nil {
			return false, false, fmt.Errorf("insert turret 1 join: %w", err)
		}
	}
	if turret2 != nil {
		err = turret2.Insert(tx, boil.Infer())
		if err != nil {
			return false, false, fmt.Errorf("insert turret 2: %w", err)
		}
		join := &boiler.ChassisWeapon{
			WeaponID:      turret2.ID,
			ChassisID:     chassis.ID,
			MountLocation: turret2.WeaponType,
			SlotNumber:    1,
		}
		err = join.Insert(tx, boil.Infer())
		if err != nil {
			return false, false, fmt.Errorf("insert turret 2 join: %w", err)
		}
	}
	err = module.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert module: %w", err)
	}
	moduleJoin := &boiler.ChassisModule{
		ChassisID:  chassis.ID,
		ModuleID:   module.ID,
		SlotNumber: 0,
	}
	err = moduleJoin.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert module join: %w", err)
	}

	return false, false, nil
}
func ProcessChassis(brand *boiler.Brand, attributes []Attributes) (*boiler.Chassis, error) {
	att := GetAttributes(attributes)
	label, slug := ChassisLabelSlug(att.Brand, att.Model, att.SubModel)
	result := &boiler.Chassis{
		ID:                 uuid.Must(uuid.NewV4()).String(),
		BrandID:            brand.ID,
		ShieldRechargeRate: att.ShieldRechargeRate,
		MaxShield:          att.MaxShieldHitPoints,
		Label:              label,
		Slug:               slug,
		HealthRemaining:    att.MaxStructureHitPoints,
		WeaponHardpoints:   att.WeaponHardpoints,
		TurretHardpoints:   att.TurretHardpoints,
		UtilitySlots:       att.UtilitySlots,
		Speed:              att.Speed,
		Skin:               att.SubModel,
		Model:              att.Model,
		MaxHitpoints:       att.MaxStructureHitPoints,
	}
	return result, nil
}
func ProcessModule(brand *boiler.Brand, attributes []Attributes) (*boiler.Module, error) {
	att := GetAttributes(attributes)
	label, slug := ModuleLabelSlug(att.Brand, att.UtilityOne)
	result := &boiler.Module{
		ID:               uuid.Must(uuid.NewV4()).String(),
		Label:            label,
		Slug:             slug,
		HitpointModifier: 100,
		ShieldModifier:   100,
	}
	return result, nil
}

func ProcessWeapon(weaponType string, index int, brand *boiler.Brand, attributes []Attributes) (*boiler.Weapon, error) {
	att := GetAttributes(attributes)
	label := ""
	weapslug := ""
	if weaponType == "TURRET" {
		if att.TurretHardpoints == 0 {
			return nil, nil
		}
		if index == 1 {
			label, weapslug = WeaponLabelSlug(brand.Label, att.TurretOne)
		}
		if index == 2 {
			label, weapslug = WeaponLabelSlug(brand.Label, att.TurretTwo)
		}
	}

	if weaponType == "ARM" {
		if att.WeaponHardpoints == 0 {
			return nil, nil
		}
		if index == 1 {
			label, weapslug = WeaponLabelSlug(brand.Label, att.WeaponOne)
		}
		if index == 2 {
			label, weapslug = WeaponLabelSlug(brand.Label, att.WeaponTwo)
		}
	}

	if label == "" || weapslug == "" {
		return nil, errors.New("could not find label, weapon or type")
	}
	result := &boiler.Weapon{
		ID:         uuid.Must(uuid.NewV4()).String(),
		Label:      label,
		Slug:       weapslug,
		Damage:     -1,
		WeaponType: weaponType,
	}
	return result, nil
}
