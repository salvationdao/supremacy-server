package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	mech, err := boiler.FindMech(tx, mechID.String())
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

func Template(templateID uuid.UUID) (*server.TemplateContainer, error) {
	// TODO: Fix this
	//template, err := boiler.FindTemplate(gamedb.StdConn, templateID.String())
	//if err != nil {
	//	return nil, err
	//}
	//chassis, err := template.BlueprintChassis().One(gamedb.StdConn)
	//if err != nil {
	//	return nil, err
	//}
	//
	//blueprintWeaponJoins, err := boiler.BlueprintChassisBlueprintWeapons(
	//	qm.Where("blueprint_chassis_id = ? AND mount_location = 'ARM'", chassis.ID),
	//).All(gamedb.StdConn)
	//if err != nil {
	//	return nil, err
	//}
	//
	//blueprintWeapons := map[int]*boiler.BlueprintWeapon{}
	//for _, join := range blueprintWeaponJoins {
	//	weapon, err := boiler.FindBlueprintWeapon(gamedb.StdConn, join.BlueprintWeaponID)
	//	if err != nil {
	//		return nil, err
	//	}
	//	blueprintWeapons[join.SlotNumber] = weapon
	//}
	//
	//blueprintTurretJoins, err := boiler.BlueprintChassisBlueprintWeapons(
	//	qm.Where("blueprint_chassis_id = ? AND mount_location = 'TURRET'", chassis.ID),
	//).All(gamedb.StdConn)
	//if err != nil {
	//	return nil, err
	//}
	//
	//blueprintTurrets := map[int]*boiler.BlueprintWeapon{}
	//for _, join := range blueprintTurretJoins {
	//	blueprintTurret, err := boiler.FindBlueprintWeapon(gamedb.StdConn, join.BlueprintWeaponID)
	//	if err != nil {
	//		return nil, err
	//	}
	//	blueprintTurrets[join.SlotNumber] = blueprintTurret
	//}
	//
	//blueprintModuleJoins, err := boiler.BlueprintChassisBlueprintModules(
	//	boiler.BlueprintChassisBlueprintModuleWhere.BlueprintChassisID.EQ(chassis.ID),
	//).All(gamedb.StdConn)
	//if err != nil {
	//	return nil, err
	//}
	//
	//blueprintModules := map[int]*boiler.BlueprintModule{}
	//for _, join := range blueprintModuleJoins {
	//	blueprintModule, err := boiler.FindBlueprintModule(gamedb.StdConn, join.BlueprintModuleID)
	//	if err != nil {
	//		return nil, err
	//	}
	//	blueprintModules[join.SlotNumber] = blueprintModule
	//}
	//
	//result := &server.TemplateContainer{
	//	Template:         template,
	//	BlueprintChassis: chassis,
	//	BlueprintWeapons: blueprintWeapons,
	//	BlueprintTurrets: blueprintTurrets,
	//	BlueprintModules: blueprintModules,
	//}

	//return result, nil
	return nil, nil
}

func TemplatePurchasedCount(templateID uuid.UUID) (int, error) {
	// TODO: Fix this

	//count, err := boiler.Mechs(boiler.MechWhere.TemplateID.EQ(templateID.String())).Count(gamedb.StdConn)
	//if err != nil {
	//	return 0, err
	//}
	//return int(count), nil
	return 0, nil
}

func DefaultMechs() ([]*server.Mech, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	idq := `SELECT id FROM mechs WHERE is_default=true`

	result, err := gamedb.Conn.Query(ctx, idq)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	ids := []uuid.UUID{}
	for result.Next() {
		id := ""
		err = result.Scan(&id)
		if err != nil {
			return nil, err
		}
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uid)
	}

	return Mechs(ids...)
}

var ErrNotAllMechsReturned = fmt.Errorf("not all mechs returned")

func Mech(mechID uuid.UUID) (*server.Mech, error) {
	// TODO: Fix this
	//mc := &server.Mech{}
	//query := `SELECT
	//   ` + strings.Join([]string{
	//	`mechs.` + boiler.MechColumns.ID,
	//	`mechs.` + boiler.MechColumns.OwnerID,
	//	`mechs.` + boiler.MechColumns.TemplateID,
	//	`mechs.` + boiler.MechColumns.ChassisID,
	//	`mechs.` + boiler.MechColumns.ExternalTokenID,
	//	`mechs.` + boiler.MechColumns.Tier,
	//	`mechs.` + boiler.MechColumns.IsDefault,
	//	`mechs.` + boiler.MechColumns.ImageURL,
	//	`mechs.` + boiler.MechColumns.AnimationURL,
	//	`mechs.` + boiler.MechColumns.CardAnimationURL,
	//	`mechs.` + boiler.MechColumns.AvatarURL,
	//	`mechs.` + boiler.MechColumns.Hash,
	//	`mechs.` + boiler.MechColumns.Name,
	//	`mechs.` + boiler.MechColumns.Label,
	//	`mechs.` + boiler.MechColumns.Slug,
	//	`mechs.` + boiler.MechColumns.AssetType,
	//	`mechs.` + boiler.MechColumns.DeletedAt,
	//	`mechs.` + boiler.MechColumns.UpdatedAt,
	//	`mechs.` + boiler.MechColumns.CreatedAt,
	//	`mechs.` + boiler.MechColumns.LargeImageURL,
	//	`mechs.` + boiler.MechColumns.CollectionSlug,
	//}, ",") + `,
	//   (SELECT to_json(chassis.*) FROM chassis WHERE id=mechs.` + boiler.MechColumns.ChassisID + `) as chassis,
	//   to_json(
	//       (SELECT jsonb_object_agg(cw.` + boiler.ChassisWeaponColumns.SlotNumber + `, wpn.* ORDER BY cw.` + boiler.ChassisWeaponColumns.SlotNumber + ` ASC)
	//        FROM chassis_weapons cw
	//        INNER JOIN weapons wpn ON wpn.id = cw.` + boiler.ChassisWeaponColumns.WeaponID + `
	//        WHERE cw.chassis_id=` + `mechs.` + boiler.MechColumns.ChassisID + ` AND cw.` + boiler.ChassisWeaponColumns.MountLocation + ` = 'ARM')
	//    ) as weapons,
	//   to_json(
	//       (SELECT jsonb_object_agg(cwt.` + boiler.ChassisWeaponColumns.SlotNumber + `, wpn.* ORDER BY cwt.` + boiler.ChassisWeaponColumns.SlotNumber + ` ASC)
	//        FROM chassis_weapons cwt
	//        INNER JOIN weapons wpn ON wpn.id = cwt.` + boiler.ChassisWeaponColumns.WeaponID + `
	//        WHERE cwt.` + boiler.MechColumns.ChassisID + `=` + `mechs.` + boiler.MechColumns.ChassisID + ` AND cwt.` + boiler.ChassisWeaponColumns.MountLocation + ` = 'TURRET'
	//        )
	//    ) as turrets,
	//    to_json(
	//        (SELECT jsonb_object_agg(mods.` + boiler.ChassisModuleColumns.SlotNumber + `, mds.* ORDER BY mods.` + boiler.ChassisModuleColumns.SlotNumber + ` ASC)
	//            FROM chassis_modules mods
	//            INNER JOIN modules mds ON mds.` + boiler.ModuleColumns.ID + ` = mods.` + boiler.ChassisModuleColumns.ModuleID + `
	//         WHERE mods.` + boiler.ChassisModuleColumns.ChassisID + `=` + `mechs.` + boiler.MechColumns.ChassisID + `)
	//    ) as modules,
	//   to_json(ply.*) as player,
	//   to_json(fct.*) as faction
	//	from mechs
	//	INNER JOIN players ply ON ply.id = mechs.` + boiler.MechColumns.OwnerID + `
	//	INNER JOIN factions fct ON fct.id = ply.` + boiler.PlayerColumns.FactionID + `
	//    WHERE mechs.id = $1
	//    GROUP BY mechs.id, ply.id, fct.id`
	//
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	//defer cancel()
	//
	//result, err := gamedb.Conn.Query(ctx, query, mechID.String())
	//if err != nil {
	//	return nil, err
	//}
	//defer result.Close()
	//
	//for result.Next() {
	//	err = result.Scan(
	//		&mc.ID,
	//		&mc.OwnerID,
	//		&mc.TemplateID,
	//		&mc.ChassisID,
	//		&mc.ExternalTokenID,
	//		&mc.Tier,
	//		&mc.IsDefault,
	//		&mc.ImageURL,
	//		&mc.AnimationURL,
	//		&mc.CardAnimationURL,
	//		&mc.AvatarURL,
	//		&mc.Hash,
	//		&mc.Name,
	//		&mc.Label,
	//		&mc.Slug,
	//		&mc.AssetType,
	//		&mc.DeletedAt,
	//		&mc.UpdatedAt,
	//		&mc.CreatedAt,
	//		&mc.LargeImageURL,
	//		&mc.CollectionSlug,
	//		&mc.Chassis,
	//		&mc.Weapons,
	//		&mc.Turrets,
	//		&mc.Modules,
	//		&mc.Player,
	//		&mc.Faction)
	//	if mc.Faction != nil {
	//		mc.FactionID = mc.Faction.ID
	//	}
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//result.Close()
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//return mc, err
	return nil, nil
}

func Mechs(mechIDs ...uuid.UUID) ([]*server.Mech, error) {
	// TODO: Fix this
	//
	//if len(mechIDs) == 0 {
	//	return nil, errors.New("no mech ids provided")
	//}
	//mcs := make([]*server.XsynMechContainer, len(mechIDs))
	//
	//mechids := make([]interface{}, len(mechIDs))
	//var paramrefs string
	//for i, id := range mechIDs {
	//	paramrefs += `$` + strconv.Itoa(i+1) + `,`
	//	mechids[i] = id.String()
	//}
	//paramrefs = paramrefs[:len(paramrefs)-1]
	//
	//query := `SELECT
	//   ` + strings.Join([]string{
	//	`mechs.` + boiler.MechColumns.ID,
	//	`mechs.` + boiler.MechColumns.OwnerID,
	//	`mechs.` + boiler.MechColumns.TemplateID,
	//	`mechs.` + boiler.MechColumns.ChassisID,
	//	`mechs.` + boiler.MechColumns.ExternalTokenID,
	//	`mechs.` + boiler.MechColumns.Tier,
	//	`mechs.` + boiler.MechColumns.IsDefault,
	//	`mechs.` + boiler.MechColumns.ImageURL,
	//	`mechs.` + boiler.MechColumns.AnimationURL,
	//	`mechs.` + boiler.MechColumns.CardAnimationURL,
	//	`mechs.` + boiler.MechColumns.AvatarURL,
	//	`mechs.` + boiler.MechColumns.Hash,
	//	`mechs.` + boiler.MechColumns.Name,
	//	`mechs.` + boiler.MechColumns.Label,
	//	`mechs.` + boiler.MechColumns.Slug,
	//	`mechs.` + boiler.MechColumns.AssetType,
	//	`mechs.` + boiler.MechColumns.DeletedAt,
	//	`mechs.` + boiler.MechColumns.UpdatedAt,
	//	`mechs.` + boiler.MechColumns.CreatedAt,
	//	`mechs.` + boiler.MechColumns.LargeImageURL,
	//	`mechs.` + boiler.MechColumns.CollectionSlug,
	//}, ",") + `,
	//   (SELECT to_json(chassis.*) FROM chassis WHERE id=mechs.` + boiler.MechColumns.ChassisID + `) as chassis,
	//   to_json(
	//       (SELECT jsonb_object_agg(cw.` + boiler.ChassisWeaponColumns.SlotNumber + `, wpn.* ORDER BY cw.` + boiler.ChassisWeaponColumns.SlotNumber + ` ASC)
	//        FROM chassis_weapons cw
	//        INNER JOIN weapons wpn ON wpn.id = cw.` + boiler.ChassisWeaponColumns.WeaponID + `
	//        WHERE cw.chassis_id=` + `mechs.` + boiler.MechColumns.ChassisID + ` AND cw.` + boiler.ChassisWeaponColumns.MountLocation + ` = 'ARM')
	//    ) as weapons,
	//   to_json(
	//       (SELECT jsonb_object_agg(cwt.` + boiler.ChassisWeaponColumns.SlotNumber + `, wpn.* ORDER BY cwt.` + boiler.ChassisWeaponColumns.SlotNumber + ` ASC)
	//        FROM chassis_weapons cwt
	//        INNER JOIN weapons wpn ON wpn.id = cwt.` + boiler.ChassisWeaponColumns.WeaponID + `
	//        WHERE cwt.` + boiler.MechColumns.ChassisID + `=` + `mechs.` + boiler.MechColumns.ChassisID + ` AND cwt.` + boiler.ChassisWeaponColumns.MountLocation + ` = 'TURRET'
	//        )
	//    ) as turrets,
	//    to_json(
	//        (SELECT jsonb_object_agg(mods.` + boiler.ChassisModuleColumns.SlotNumber + `, mds.* ORDER BY mods.` + boiler.ChassisModuleColumns.SlotNumber + ` ASC)
	//            FROM chassis_modules mods
	//            INNER JOIN modules mds ON mds.` + boiler.ModuleColumns.ID + ` = mods.` + boiler.ChassisModuleColumns.ModuleID + `
	//         WHERE mods.` + boiler.ChassisModuleColumns.ChassisID + `=` + `mechs.` + boiler.MechColumns.ChassisID + `)
	//    ) as modules,
	//   to_json(ply.*) as player,
	//   to_json(fct.*) as faction
	//	from mechs
	//	INNER JOIN players ply ON ply.id = mechs.` + boiler.MechColumns.OwnerID + `
	//	INNER JOIN factions fct ON fct.id = ply.` + boiler.PlayerColumns.FactionID + `
	//	WHERE mechs.id IN (` + paramrefs + `)
	//	GROUP BY mechs.id, ply.id, fct.id
	// 	ORDER BY fct.id;`
	//
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	//defer cancel()
	//
	//result, err := gamedb.Conn.Query(ctx, query, mechids...)
	//if err != nil {
	//	return nil, err
	//}
	//defer result.Close()
	//
	//i := 0
	//for result.Next() {
	//	mc := &server.XsynMechContainer{}
	//	err = result.Scan(
	//		&mc.ID,
	//		&mc.OwnerID,
	//		&mc.TemplateID,
	//		&mc.ChassisID,
	//		&mc.ExternalTokenID,
	//		&mc.Tier,
	//		&mc.IsDefault,
	//		&mc.ImageURL,
	//		&mc.AnimationURL,
	//		&mc.CardAnimationURL,
	//		&mc.AvatarURL,
	//		&mc.Hash,
	//		&mc.Name,
	//		&mc.Label,
	//		&mc.Slug,
	//		&mc.AssetType,
	//		&mc.DeletedAt,
	//		&mc.UpdatedAt,
	//		&mc.CreatedAt,
	//		&mc.LargeImageURL,
	//		&mc.CollectionSlug,
	//		&mc.Chassis,
	//		&mc.Weapons,
	//		&mc.Turrets,
	//		&mc.Modules,
	//		&mc.Player,
	//		&mc.Faction)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if mc.Faction != nil {
	//		mc.FactionID = mc.Faction.ID
	//	}
	//	mcs[i] = mc
	//	i++
	//}
	//
	//if i < len(mechIDs) {
	//	mcs = mcs[:len(mcs)-i]
	//	return mcs, ErrNotAllMechsReturned
	//}
	//
	//return mcs, err
	return nil, nil
}

func NextExternalTokenID(tx *sql.Tx, isDefault bool, collectionSlug null.String) (int, error) {
	// TODO: Fix this
	//count, err := boiler.Mechs(
	//	boiler.MechWhere.IsDefault.EQ(isDefault),
	//	boiler.MechWhere.CollectionSlug.EQ(collectionSlug),
	//).Count(tx)
	//if err != nil {
	//	return 0, err
	//}
	//if count == 0 {
	//	return 0, nil
	//}
	//
	//highestMechID, err := boiler.Mechs(
	//	boiler.MechWhere.IsDefault.EQ(isDefault),
	//	boiler.MechWhere.CollectionSlug.EQ(collectionSlug),
	//	qm.OrderBy("external_token_id DESC"),
	//).One(tx)
	//if err != nil {
	//	return 0, err
	//}
	//
	//return highestMechID.ExternalTokenID + 1, nil
	return 0, nil
}

// MechRegister copies everything out of a template into a new mech
func MechRegister(templateID uuid.UUID, ownerID uuid.UUID) (uuid.UUID, error) {
	// TODO: Fix this
	//tx, err := gamedb.StdConn.Begin()
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("start tx: %w", err)
	//}
	//defer tx.Rollback()
	//exists, err := boiler.PlayerExists(tx, ownerID.String())
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("check player exists: %w", err)
	//}
	//if !exists {
	//	newPlayer := &boiler.Player{ID: ownerID.String()}
	//	err = newPlayer.Insert(tx, boil.Infer())
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf("insert new player: %w", err)
	//	}
	//}
	//template, err := boiler.FindTemplate(tx, templateID.String())
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("find template: %w", err)
	//}
	//
	//blueprintChassis, err := template.BlueprintChassis().One(tx)
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("get blueprint chassis: %w", err)
	//}
	//chassis := &boiler.Chassis{
	//	BrandID:            blueprintChassis.BrandID,
	//	Label:              blueprintChassis.Label,
	//	Model:              blueprintChassis.Model,
	//	Skin:               blueprintChassis.Skin,
	//	Slug:               blueprintChassis.Slug,
	//	ShieldRechargeRate: blueprintChassis.ShieldRechargeRate,
	//	HealthRemaining:    blueprintChassis.MaxHitpoints,
	//	WeaponHardpoints:   blueprintChassis.WeaponHardpoints,
	//	TurretHardpoints:   blueprintChassis.TurretHardpoints,
	//	UtilitySlots:       blueprintChassis.UtilitySlots,
	//	Speed:              blueprintChassis.Speed,
	//	MaxHitpoints:       blueprintChassis.MaxHitpoints,
	//	MaxShield:          blueprintChassis.MaxShield,
	//}
	//err = chassis.Insert(tx, boil.Infer())
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf(": %w", err)
	//}
	//
	//weaponJoins, err := boiler.BlueprintChassisBlueprintWeapons(boiler.BlueprintChassisBlueprintWeaponWhere.BlueprintChassisID.EQ(template.BlueprintChassisID)).All(tx)
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("get blueprint weapon joins: %w", err)
	//}
	//for _, join := range weaponJoins {
	//	blueprintWeapon, err := boiler.FindBlueprintWeapon(tx, join.BlueprintWeaponID)
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf("get blueprint weapon: %w", err)
	//	}
	//	newWeapon := &boiler.Weapon{
	//		BrandID:    blueprintWeapon.BrandID,
	//		Label:      blueprintWeapon.Label,
	//		Slug:       blueprintWeapon.Slug,
	//		Damage:     blueprintWeapon.Damage,
	//		WeaponType: blueprintWeapon.WeaponType,
	//	}
	//	err = newWeapon.Insert(tx, boil.Infer())
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf(": %w", err)
	//	}
	//	newJoin := &boiler.ChassisWeapon{
	//		ChassisID:     chassis.ID,
	//		WeaponID:      newWeapon.ID,
	//		SlotNumber:    join.SlotNumber,
	//		MountLocation: join.MountLocation,
	//	}
	//	err = newJoin.Insert(tx, boil.Infer())
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf("insert blueprint weapon join: %w", err)
	//	}
	//}
	//moduleJoins, err := boiler.BlueprintChassisBlueprintModules(boiler.BlueprintChassisBlueprintModuleWhere.BlueprintChassisID.EQ(template.BlueprintChassisID)).All(tx)
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("get blueprint module joins: %w", err)
	//}
	//for _, join := range moduleJoins {
	//	blueprintModule, err := boiler.FindBlueprintModule(tx, join.BlueprintModuleID)
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf("get blueprint module: %w", err)
	//	}
	//	newModule := &boiler.Module{
	//		BrandID:          blueprintModule.BrandID,
	//		Slug:             blueprintModule.Slug,
	//		Label:            blueprintModule.Label,
	//		HitpointModifier: blueprintModule.HitpointModifier,
	//		ShieldModifier:   blueprintModule.ShieldModifier,
	//	}
	//	err = newModule.Insert(tx, boil.Infer())
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf("insert blueprint module: %w", err)
	//	}
	//	newJoin := &boiler.ChassisModule{
	//		ChassisID:  chassis.ID,
	//		ModuleID:   newModule.ID,
	//		SlotNumber: join.SlotNumber,
	//	}
	//	err = newJoin.Insert(tx, boil.Infer())
	//	if err != nil {
	//		return uuid.Nil, fmt.Errorf("insert blueprint module join: %w", err)
	//	}
	//}
	//
	//newMechID, err := uuid.NewV4()
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("create mech id: %w", err)
	//}
	//shortID, err := shortid.Generate()
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("create short id: %w", err)
	//}
	//nextID, err := NextExternalTokenID(tx, template.IsDefault, template.CollectionSlug)
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("get next external token id: %w", err)
	//}
	//newMech := &boiler.Mech{
	//	ID:              newMechID.String(),
	//	OwnerID:         ownerID.String(),
	//	TemplateID:      templateID.String(),
	//	ChassisID:       chassis.ID,
	//	Tier:            template.Tier,
	//	IsDefault:       template.IsDefault,
	//	Hash:            shortID,
	//	Name:            "",
	//	ExternalTokenID: nextID,
	//	Label:           template.Label,
	//	Slug:            template.Slug,
	//	AssetType:       template.AssetType,
	//	CollectionSlug:  template.CollectionSlug,
	//
	//	AvatarURL:        template.AvatarURL,
	//	LargeImageURL:    template.LargeImageURL,
	//	ImageURL:         template.ImageURL,
	//	AnimationURL:     template.AnimationURL,
	//	CardAnimationURL: template.CardAnimationURL,
	//}
	//err = newMech.Insert(tx, boil.Infer())
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("insert mech: %w", err)
	//}
	//tx.Commit()
	//return newMechID, nil
	return uuid.Nil, nil
}

// MechIDFromHash retrieve a mech ID from a hash
func MechIDFromHash(hash string) (uuid.UUID, error) {
	q := `SELECT id FROM mechs WHERE hash = $1`
	var id string
	err := gamedb.Conn.QueryRow(context.Background(), q, hash).
		Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	uid, err := uuid.FromString(id)

	if err != nil {
		return uuid.Nil, err
	}

	return uid, err
}

// MechIDFromHash retrieve a slice mech IDs from hash variatic
func MechIDsFromHash(hashes ...string) ([]uuid.UUID, error) {
	var paramrefs string
	idintf := []interface{}{}
	for i, hash := range hashes {
		if hash != "" {
			paramrefs += `$` + strconv.Itoa(i+1) + `,`
			idintf = append(idintf, hash)
		}
	}
	paramrefs = paramrefs[:len(paramrefs)-1]
	q := `SELECT id, hash FROM mechs WHERE mechs.hash IN (` + paramrefs + `)`

	result, err := gamedb.Conn.Query(context.Background(), q, idintf...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	ids := make([]uuid.UUID, len(hashes))
	i := 0
	for result.Next() {
		var idStr string
		var hash string
		err = result.Scan(&idStr, &hash)
		if err != nil {
			return nil, err
		}

		uid, err := uuid.FromString(idStr)
		if err != nil {
			gamelog.L.Error().Str("mechID", idStr).Str("db func", "MechIDsFromHash").Err(err).Msg("unable to convert id to uuid")
		}

		// set id in correct order
		for index, h := range hashes {
			if h == hash {
				ids[index] = uid
				i++
			}
		}
	}

	if i == 0 {
		return nil, errors.New("no ids were scanned from result")
	}

	return ids, err
}

type BattleQueuePosition struct {
	MechID           uuid.UUID `db:"mech_id"`
	QueuePosition    int64     `db:"queue_position"`
	BattleContractID string    `db:"battle_contract_id"`
}

// MechQueuePosition return a list of mech queue position of the player (exclude in battle)
func MechQueuePosition(factionID string, ownerID string) ([]*BattleQueuePosition, error) {
	q := `
		SELECT
			x.mech_id,
			x.queue_position,
		    x.battle_contract_id
		FROM
			(
				SELECT
					bq.id,
					bq.mech_id,
				    bq.owner_id,
				    bq.battle_contract_id,
					row_number () over (ORDER BY bq.queued_at) AS queue_position
				FROM
					battle_queue bq
				WHERE 
					bq.faction_id = $1 AND bq.battle_id isnull
			) x
		WHERE
			x.owner_id = $2
		ORDER BY
			x.queue_position
	`

	result, err := gamedb.StdConn.Query(q, factionID, ownerID)
	if err != nil {
		return nil, terror.Error(err)
	}

	mqp := []*BattleQueuePosition{}
	for result.Next() {
		qp := &BattleQueuePosition{}
		err = result.Scan(&qp.MechID, &qp.QueuePosition, &qp.BattleContractID)
		if err != nil {
			return nil, terror.Error(err)
		}

		mqp = append(mqp, qp)
	}

	return mqp, nil
}
