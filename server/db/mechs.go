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

	"github.com/gofrs/uuid"
	"github.com/teris-io/shortid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func MechsByOwnerID(ownerID uuid.UUID) ([]*server.MechContainer, error) {
	mechs, err := boiler.Mechs(boiler.MechWhere.OwnerID.EQ(ownerID.String())).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*server.MechContainer{}
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
	template, err := boiler.FindTemplate(gamedb.StdConn, templateID.String())
	if err != nil {
		return nil, err
	}
	chassis, err := template.BlueprintChassis().One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	blueprintWeaponJoins, err := boiler.BlueprintChassisBlueprintWeapons(
		qm.Where("blueprint_chassis_id = ? AND mount_location = 'ARM'", chassis.ID),
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
		qm.Where("blueprint_chassis_id = ? AND mount_location = 'TURRET'", chassis.ID),
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
		boiler.BlueprintChassisBlueprintModuleWhere.BlueprintChassisID.EQ(chassis.ID),
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

	result := &server.TemplateContainer{
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

func TemplatesByFactionID(factionID uuid.UUID) ([]*server.TemplateContainer, error) {
	templates, err := boiler.Templates().All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*server.TemplateContainer{}
	for _, tpl := range templates {
		template, err := Template(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return nil, err
		}
		result = append(result, template)
	}

	return result, nil
}

func DefaultMechs() ([]*server.MechContainer, error) {
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

func Mechs(mechIDs ...uuid.UUID) ([]*server.MechContainer, error) {
	if len(mechIDs) == 0 {
		return nil, errors.New("no mech ids provided")
	}
	mcs := make([]*server.MechContainer, len(mechIDs))

	mechids := make([]interface{}, len(mechIDs))
	var paramrefs string
	for i, id := range mechIDs {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
		mechids[i] = id.String()
	}
	paramrefs = paramrefs[:len(paramrefs)-1]

	query := `SELECT
       mechs.*,
       (SELECT to_json(chassis.*) FROM chassis WHERE id=mechs.chassis_id) as chassis,
       to_json(
           (SELECT jsonb_object_agg(cw.slot_number, wpn.* ORDER BY cw.slot_number ASC)
            FROM chassis_weapons cw
            INNER JOIN weapons wpn ON wpn.id = cw.weapon_id
            WHERE cw.chassis_id=mechs.chassis_id AND cw.mount_location = 'ARM')
        ) as weapons,
       to_json(
           (SELECT jsonb_object_agg(cwt.slot_number, wpn.* ORDER BY cwt.slot_number ASC)
            FROM chassis_weapons cwt
            INNER JOIN weapons wpn ON wpn.id = cwt.weapon_id
            WHERE cwt.chassis_id=mechs.chassis_id AND cwt.mount_location = 'TURRET'
            )
        ) as turrets,
        to_json(
            (SELECT jsonb_object_agg(mods.slot_number, mds.* ORDER BY mods.slot_number ASC)
                FROM chassis_modules mods
                INNER JOIN modules mds ON mds.id = mods.module_id
             WHERE mods.chassis_id=mechs.chassis_id)
        ) as modules,
       to_json(ply.*) as player,
       to_json(fct.*) as faction
		from mechs
		INNER JOIN players ply ON ply.id = mechs.owner_id
		INNER JOIN factions fct ON fct.id = ply.faction_id
		WHERE mechs.id IN (` + paramrefs + `)
		GROUP BY mechs.id, ply.id, fct.id
	 	ORDER BY fct.id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	result, err := gamedb.Conn.Query(ctx, query, mechids...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	i := 0
	for result.Next() {
		mc := &server.MechContainer{}
		err = result.Scan(
			&mc.ID,
			&mc.OwnerID,
			&mc.TemplateID,
			&mc.ChassisID,
			&mc.ExternalTokenID,
			&mc.Tier,
			&mc.IsDefault,
			&mc.ImageURL,
			&mc.AnimationURL,
			&mc.CardAnimationURL,
			&mc.AvatarURL,
			&mc.Hash,
			&mc.Name,
			&mc.Label,
			&mc.Slug,
			&mc.AssetType,
			&mc.DeletedAt,
			&mc.UpdatedAt,
			&mc.CreatedAt,
			&mc.LargeImageURL,
			&mc.CollectionSlug,
			&mc.Chassis,
			&mc.Weapons,
			&mc.Turrets,
			&mc.Modules,
			&mc.Player,
			&mc.Faction)
		if err != nil {
			return nil, err
		}
		mcs[i] = mc
		i++
	}

	return mcs, err
}

func Mech(mechID uuid.UUID) (*server.MechContainer, error) {
	mc := &server.MechContainer{}
	query := `SELECT
       mechs.*,
       (SELECT to_json(chassis.*) FROM chassis WHERE id=mechs.chassis_id) as chassis,
       to_json(
           (SELECT jsonb_object_agg(cw.slot_number, wpn.* ORDER BY cw.slot_number ASC)
            FROM chassis_weapons cw
            INNER JOIN weapons wpn ON wpn.id = cw.weapon_id
            WHERE cw.chassis_id=mechs.chassis_id AND cw.mount_location = 'ARM')
        ) as weapons,
       to_json(
           (SELECT jsonb_object_agg(cwt.slot_number, wpn.* ORDER BY cwt.slot_number ASC)
            FROM chassis_weapons cwt
            INNER JOIN weapons wpn ON wpn.id = cwt.weapon_id
            WHERE cwt.chassis_id=mechs.chassis_id AND cwt.mount_location = 'TURRET'
            )
        ) as turrets,
        to_json(
            (SELECT jsonb_object_agg(mods.slot_number, mds.* ORDER BY mods.slot_number ASC)
                FROM chassis_modules mods
                INNER JOIN modules mds ON mds.id = mods.module_id
             WHERE mods.chassis_id=mechs.chassis_id)
        ) as modules,
       to_json(ply.*) as player,
       to_json(fct.*) as faction
		from mechs
		INNER JOIN players ply ON ply.id = mechs.owner_id
		LEFT JOIN factions fct ON fct.id = ply.faction_id
		WHERE mechs.id = $1
		GROUP BY mechs.id, ply.id, fct.id`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	result, err := gamedb.Conn.Query(ctx, query, mechID.String())
	if err != nil {
		return nil, err
	}
	defer result.Close()

	for result.Next() {
		err = result.Scan(
			&mc.ID,
			&mc.OwnerID,
			&mc.TemplateID,
			&mc.ChassisID,
			&mc.ExternalTokenID,
			&mc.Tier,
			&mc.IsDefault,
			&mc.ImageURL,
			&mc.AnimationURL,
			&mc.CardAnimationURL,
			&mc.AvatarURL,
			&mc.Hash,
			&mc.Name,
			&mc.Label,
			&mc.Slug,
			&mc.AssetType,
			&mc.DeletedAt,
			&mc.UpdatedAt,
			&mc.CreatedAt,
			&mc.LargeImageURL,
			&mc.CollectionSlug,
			&mc.Chassis,
			&mc.Weapons,
			&mc.Turrets,
			&mc.Modules,
			&mc.Player,
			&mc.Faction)
		if err != nil {
			return nil, err
		}
	}
	result.Close()

	if err != nil {
		return nil, err
	}

	return mc, err
}

func NextExternalTokenID(tx *sql.Tx, isDefault bool, collectionSlug null.String) (int, error) {
	count, err := boiler.Mechs(
		boiler.MechWhere.IsDefault.EQ(isDefault),
		boiler.MechWhere.CollectionSlug.EQ(collectionSlug),
	).Count(tx)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}

	highestMechID, err := boiler.Mechs(
		boiler.MechWhere.IsDefault.EQ(isDefault),
		boiler.MechWhere.CollectionSlug.EQ(collectionSlug),
		qm.OrderBy("external_token_id DESC"),
	).One(tx)
	if err != nil {
		return 0, err
	}

	return highestMechID.ExternalTokenID + 1, nil
}

// MechRegister copies everything out of a template into a new mech
func MechRegister(templateID uuid.UUID, ownerID uuid.UUID) (uuid.UUID, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return uuid.Nil, fmt.Errorf("start tx: %w", err)
	}
	defer tx.Rollback()
	exists, err := boiler.PlayerExists(tx, ownerID.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("check player exists: %w", err)
	}
	if !exists {
		newPlayer := &boiler.Player{ID: ownerID.String()}
		err = newPlayer.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert new player: %w", err)
		}
	}
	template, err := boiler.FindTemplate(tx, templateID.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("find template: %w", err)
	}

	blueprintChassis, err := template.BlueprintChassis().One(tx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get blueprint chassis: %w", err)
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
		return uuid.Nil, fmt.Errorf(": %w", err)
	}

	weaponJoins, err := boiler.BlueprintChassisBlueprintWeapons(boiler.BlueprintChassisBlueprintWeaponWhere.BlueprintChassisID.EQ(template.BlueprintChassisID)).All(tx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get blueprint weapon joins: %w", err)
	}
	for _, join := range weaponJoins {
		blueprintWeapon, err := boiler.FindBlueprintWeapon(tx, join.BlueprintWeaponID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("get blueprint weapon: %w", err)
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
			return uuid.Nil, fmt.Errorf(": %w", err)
		}
		newJoin := &boiler.ChassisWeapon{
			ChassisID:     chassis.ID,
			WeaponID:      newWeapon.ID,
			SlotNumber:    join.SlotNumber,
			MountLocation: join.MountLocation,
		}
		err = newJoin.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert blueprint weapon join: %w", err)
		}
	}
	moduleJoins, err := boiler.BlueprintChassisBlueprintModules(boiler.BlueprintChassisBlueprintModuleWhere.BlueprintChassisID.EQ(template.BlueprintChassisID)).All(tx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get blueprint module joins: %w", err)
	}
	for _, join := range moduleJoins {
		blueprintModule, err := boiler.FindBlueprintModule(tx, join.BlueprintModuleID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("get blueprint module: %w", err)
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
			return uuid.Nil, fmt.Errorf("insert blueprint module: %w", err)
		}
		newJoin := &boiler.ChassisModule{
			ChassisID:  chassis.ID,
			ModuleID:   newModule.ID,
			SlotNumber: join.SlotNumber,
		}
		err = newJoin.Insert(tx, boil.Infer())
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert blueprint module join: %w", err)
		}
	}

	newMechID, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, fmt.Errorf("create mech id: %w", err)
	}
	shortID, err := shortid.Generate()
	if err != nil {
		return uuid.Nil, fmt.Errorf("create short id: %w", err)
	}
	nextID, err := NextExternalTokenID(tx, template.IsDefault, template.CollectionSlug)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get next external token id: %w", err)
	}
	newMech := &boiler.Mech{
		ID:              newMechID.String(),
		OwnerID:         ownerID.String(),
		TemplateID:      templateID.String(),
		ChassisID:       chassis.ID,
		Tier:            template.Tier,
		IsDefault:       template.IsDefault,
		Hash:            shortID,
		Name:            "",
		ExternalTokenID: nextID,
		Label:           template.Label,
		Slug:            template.Slug,
		AssetType:       template.AssetType,
		CollectionSlug:  template.CollectionSlug,

		AvatarURL:        template.AvatarURL,
		LargeImageURL:    template.LargeImageURL,
		ImageURL:         template.ImageURL,
		AnimationURL:     template.AnimationURL,
		CardAnimationURL: template.CardAnimationURL,
	}
	err = newMech.Insert(tx, boil.Infer())
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert mech: %w", err)
	}
	tx.Commit()
	return newMechID, nil
}

//MechIDFromHash retrieve a mech ID from a hash
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

//MechIDFromHash retrieve a slice mech IDs from hash variatic
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
