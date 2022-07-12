package db

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

type SiloType struct {
	Type        string                `db:"type" json:"type"`
	OwnershipID string                `db:"ownership_id" json:"ownership_id"`
	StaticID    *string               `db:"static_id" json:"static_id,omitempty"`
	SkinIDStr   *string               `db:"skin_id" json:"skin_id_str,omitempty"`
	SkinID      *SiloSkin             `json:"skin,omitempty"`
	CanOpenOn   *string               `db:"can_open_on" json:"can_open_on,omitempty"`
	Accessories []MechSiloAccessories `json:"accessories,omitempty"`
}

type MechSiloAccessories struct {
	Type        string    `json:"type"`
	OwnershipID string    `json:"ownership_id"`
	StaticID    string    `json:"static_id"`
	Skin        *SiloSkin `json:"skin,omitempty"`
}

type SiloSkin struct {
	Type        string  `json:"type"`
	OwnershipID *string `json:"ownership_id,omitempty"`
	StaticID    *string `json:"static_id,omitempty"`
}

func GetUserMechHangarItems(userID string) ([]*SiloType, error) {
	q := `
	SELECT 	ci.item_type    as type,
			ci.id           as ownership_id,
       		m.model_id  	as static_id,
       		ms.blueprint_id as skin_id
	FROM collection_items ci
         	INNER JOIN mechs m on
    			m.id = ci.item_id
         	INNER JOIN mech_skin ms on
        		ms.id = coalesce(
            			m.chassis_skin_id,
            			(select default_chassis_skin_id from mech_models mm where mm.id = m.model_id)
        				)
	WHERE ci.owner_id = $1
  	AND (ci.item_type = 'mech');
	`
	rows, err := boiler.NewQuery(qm.SQL(q, userID)).Query(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*SiloType{}, nil
		}
		return nil, terror.Error(err, "failed to query for finding silos")
	}

	mechSiloType := make([]*SiloType, 0)
	defer rows.Close()
	for rows.Next() {
		mst := &SiloType{}

		err := rows.Scan(&mst.Type, &mst.OwnershipID, &mst.StaticID, &mst.SkinIDStr)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		mechSiloType = append(mechSiloType, mst)
	}

	for _, mechSilo := range mechSiloType {
		collectionItem, err := boiler.FindCollectionItem(gamedb.StdConn, mechSilo.OwnershipID)
		if err != nil {
			continue
		}
		var mechAttributes []MechSiloAccessories
		mech, err := Mech(gamedb.StdConn, collectionItem.ItemID)
		if err != nil {
			return nil, terror.Error(err, "Failed to get mech info")
		}
		if mech.IsCompleteLimited() || mech.IsCompleteGenesis() {
			mechDefaultSkin := &SiloSkin{
				Type:        "skin",
				OwnershipID: nil,
				StaticID:    mechSilo.SkinIDStr,
			}
			mechSilo.SkinIDStr = nil
			mechSilo.SkinID = mechDefaultSkin
			mechAttributes = []MechSiloAccessories{}
			continue
		}

		mechSkin := &SiloSkin{
			Type:        "skin",
			OwnershipID: nil,
			StaticID:    mechSilo.SkinIDStr,
		}

		if mech.ChassisSkinID.Valid {
			mechSkinOwnership, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.ChassisSkinID.String)).One(gamedb.StdConn)
			if err != nil {
				continue
			}

			mechSkin.OwnershipID = &mechSkinOwnership.ID
		}

		mechSilo.SkinID = mechSkin

		if len(mech.Weapons) > 0 {
			for _, weapon := range mech.Weapons {
				weaponStringID := weapon.EquippedWeaponSkinID.String
				if !weapon.EquippedWeaponSkinID.Valid {
					defaultSkin, err := boiler.BlueprintWeaponSkins(
						boiler.BlueprintWeaponSkinWhere.Label.EQ(mech.ChassisSkin.Label),
						boiler.BlueprintWeaponSkinWhere.WeaponType.EQ(weapon.WeaponType),
					).One(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
						continue
					}
					weaponStringID = defaultSkin.ID
				}

				weaponCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weapon.ID), qm.Select(boiler.CollectionItemColumns.ID)).One(gamedb.StdConn)
				if err != nil {
					continue
				}
				var weaponSkin *string
				if weapon.EquippedWeaponSkinID.Valid {
					weaponSkin = &weapon.EquippedWeaponSkinID.String
				}

				newAttribute := MechSiloAccessories{
					Type:        "weapon",
					OwnershipID: weaponCollection.ID,
					StaticID:    weapon.BlueprintID,
					Skin: &SiloSkin{
						Type:        "skin",
						OwnershipID: weaponSkin,
						StaticID:    &weaponStringID,
					},
				}

				mechAttributes = append(mechAttributes, newAttribute)
			}
		}

		if mech.PowerCoreID.Valid {
			powerCoreCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.PowerCoreID.String)).One(gamedb.StdConn)
			if err != nil {
				continue
			}

			powerCoreBlueprint, err := boiler.PowerCores(boiler.PowerCoreWhere.ID.EQ(mech.PowerCoreID.String)).One(gamedb.StdConn)
			if err != nil {
				continue
			}

			newAttribute := MechSiloAccessories{
				Type:        "power_core",
				OwnershipID: powerCoreCollection.ID,
				StaticID:    powerCoreBlueprint.BlueprintID.String,
			}

			mechAttributes = append(mechAttributes, newAttribute)
		}

		mechSilo.Accessories = mechAttributes
		mechSilo.SkinIDStr = nil
	}

	return mechSiloType, nil
}

func GetUserMysteryCrateHangarItems(userID string) ([]*SiloType, error) {
	q := `
	SELECT 	ci.item_type		 	as type,
			ci.id    				as ownership_id,
			smc.id 					as mystery_crate_id,
			mc.locked_until        	as can_open_on
	FROM 	collection_items ci
         	INNER JOIN mystery_crate mc on
    			mc.id = ci.item_id AND mc.opened = false
         	INNER JOIN storefront_mystery_crates smc on
            	smc.mystery_crate_type = mc."type"
        	AND smc.faction_id = mc.faction_id
	WHERE ci.owner_id = $1
  			AND ci.item_type = 'mystery_crate';
	`
	rows, err := boiler.NewQuery(qm.SQL(q, userID)).Query(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*SiloType{}, nil
		}
		return nil, terror.Error(err, "failed to query for finding silos")
	}

	mechSiloType := make([]*SiloType, 0)
	defer rows.Close()
	for rows.Next() {
		mst := &SiloType{}
		var canOpenOn time.Time
		err := rows.Scan(&mst.Type, &mst.OwnershipID, &mst.StaticID, &canOpenOn)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		canOpenOnStr := canOpenOn.Format("2006-01-02T15:04:05.000Z")

		mst.CanOpenOn = &canOpenOnStr

		mechSiloType = append(mechSiloType, mst)
	}

	return mechSiloType, nil
}

func GetUserWeaponHangarItems(userID string) ([]*SiloType, error) {
	ownedWeapons, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
		boiler.CollectionItemWhere.OwnerID.EQ(userID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get user owned weapons")
	}

	var weaponHangarSilo []*SiloType
	for _, ownedWeapon := range ownedWeapons {
		weapon, err := Weapon(gamedb.StdConn, ownedWeapon.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("owned weapon col id", ownedWeapon.ID).Msg("Failed to get weapon")
			continue
		}

		if weapon.EquippedOn.Valid || !weapon.EquippedWeaponSkinID.Valid {
			continue
		}

		weaponBlueprint, err := boiler.BlueprintWeapons(boiler.BlueprintWeaponWhere.ID.EQ(weapon.BlueprintID), qm.Load(boiler.BlueprintWeaponRels.WeaponModel)).One(gamedb.StdConn)
		if err != nil {
			continue
		}

		weaponSilo := &SiloType{
			Type:        ownedWeapon.ItemType,
			OwnershipID: ownedWeapon.ID,
			StaticID:    &weaponBlueprint.R.WeaponModel.ID,
		}

		weaponSkin, err := boiler.WeaponSkins(boiler.WeaponSkinWhere.ID.EQ(weapon.EquippedWeaponSkinID.String)).One(gamedb.StdConn)
		if err != nil {
			continue
		}

		weaponSkinOwnership, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weaponSkin.ID)).One(gamedb.StdConn)
		if err != nil {
			continue
		}

		weaponSkinSilo := &SiloSkin{
			Type:        "skin",
			OwnershipID: &weaponSkinOwnership.ID,
			StaticID:    &weaponSkin.BlueprintID,
		}

		weaponSilo.SkinID = weaponSkinSilo

		weaponHangarSilo = append(weaponHangarSilo, weaponSilo)

	}

	return weaponHangarSilo, nil
}

func GetUserMechHangarItemsWithMechID(userID, mechID string, trx boil.Executor) (*SiloType, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	q := `
	SELECT 	ci.item_type    as type,
			ci.id           as ownership_id,
       		m.model_id  	as static_id,
       		ms.blueprint_id as skin_id
	FROM collection_items ci
         	INNER JOIN mechs m on
    			m.id = ci.item_id
         	INNER JOIN mech_skin ms on
        		ms.id = coalesce(
            			m.chassis_skin_id,
            			(select default_chassis_skin_id from mech_models mm where mm.id = m.model_id)
        				)
	WHERE ci.owner_id = $1
  	AND (ci.item_type = 'mech')
	AND (m.id = '$2');
	`
	rows, err := boiler.NewQuery(qm.SQL(q, userID, mechID)).Query(tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, terror.Error(err, "failed to query for finding silos")
	}

	mechSiloType := &SiloType{}
	defer rows.Close()
	for rows.Next() {

		err := rows.Scan(&mechSiloType.Type, &mechSiloType.OwnershipID, &mechSiloType.StaticID, &mechSiloType.SkinIDStr)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

	}

	collectionItem, err := boiler.FindCollectionItem(tx, mechSiloType.OwnershipID)
	if err != nil {
		return nil, terror.Error(err, "Failed to get collection item for")
	}
	var mechAttributes []MechSiloAccessories
	mech, err := Mech(tx, collectionItem.ItemID)
	if err != nil {
		return nil, terror.Error(err, "Failed to get mech info")
	}
	if mech.IsCompleteLimited() || mech.IsCompleteGenesis() {
		mechDefaultSkin := &SiloSkin{
			Type:        "skin",
			OwnershipID: nil,
			StaticID:    mechSiloType.SkinIDStr,
		}
		mechSiloType.SkinIDStr = nil
		mechSiloType.SkinID = mechDefaultSkin
		mechAttributes = []MechSiloAccessories{}
		return mechSiloType, nil
	}

	mechSkin := &SiloSkin{
		Type:        "skin",
		OwnershipID: nil,
		StaticID:    mechSiloType.SkinIDStr,
	}

	if mech.ChassisSkinID.Valid {
		mechSkinOwnership, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.ChassisSkinID.String)).One(tx)
		if err != nil {
			return nil, terror.Error(err, "Failed to get mech skin ownership")
		}

		mechSkin.OwnershipID = &mechSkinOwnership.ID
	}

	mechSiloType.SkinID = mechSkin

	if len(mech.Weapons) > 0 {
		for _, weapon := range mech.Weapons {
			weaponStringID := weapon.EquippedWeaponSkinID.String
			if !weapon.EquippedWeaponSkinID.Valid {
				defaultSkin, err := boiler.BlueprintWeaponSkins(
					boiler.BlueprintWeaponSkinWhere.Label.EQ(mech.ChassisSkin.Label),
					boiler.BlueprintWeaponSkinWhere.WeaponType.EQ(weapon.WeaponType),
				).One(tx)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
					continue
				}
				weaponStringID = defaultSkin.ID
			}

			weaponCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weapon.ID), qm.Select(boiler.CollectionItemColumns.ID)).One(tx)
			if err != nil {
				continue
			}
			var weaponSkin *string
			if weapon.EquippedWeaponSkinID.Valid {
				weaponSkin = &weapon.EquippedWeaponSkinID.String
			}

			newAttribute := MechSiloAccessories{
				Type:        "weapon",
				OwnershipID: weaponCollection.ID,
				StaticID:    weapon.BlueprintID,
				Skin: &SiloSkin{
					Type:        "skin",
					OwnershipID: weaponSkin,
					StaticID:    &weaponStringID,
				},
			}

			mechAttributes = append(mechAttributes, newAttribute)
		}
	}

	if mech.PowerCoreID.Valid {
		powerCoreCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.PowerCoreID.String)).One(tx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, terror.Error(err, "Failed to get collection item for power core")
		}

		powerCoreBlueprint, err := boiler.PowerCores(boiler.PowerCoreWhere.ID.EQ(mech.PowerCoreID.String)).One(tx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, terror.Error(err, "Failed to get blueprint info for power core")
		}

		newAttribute := MechSiloAccessories{
			Type:        "power_core",
			OwnershipID: powerCoreCollection.ID,
			StaticID:    powerCoreBlueprint.BlueprintID.String,
		}

		mechAttributes = append(mechAttributes, newAttribute)
	}

	mechSiloType.Accessories = mechAttributes
	mechSiloType.SkinIDStr = nil

	return mechSiloType, nil
}

func GetUserWeaponHangarItemsWithID(userID, weaponID string, trx boil.Executor) (*SiloType, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	ownedWeapon, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
		boiler.CollectionItemWhere.OwnerID.EQ(userID),
		boiler.CollectionItemWhere.ID.EQ(weaponID),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get user owned weapons")
	}

	weapon, err := Weapon(tx, ownedWeapon.ItemID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("owned weapon col id", ownedWeapon.ID).Msg("Failed to get weapon")
		return nil, err
	}

	if weapon.EquippedOn.Valid || !weapon.EquippedWeaponSkinID.Valid {
		return nil, terror.Error(fmt.Errorf("weapon not availiable in hangar"), "Weapon not available on hangar by itself")
	}

	weaponBlueprint, err := boiler.BlueprintWeapons(boiler.BlueprintWeaponWhere.ID.EQ(weapon.BlueprintID), qm.Load(boiler.BlueprintWeaponRels.WeaponModel)).One(tx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get blueprint weapon")
	}

	weaponSilo := &SiloType{
		Type:        ownedWeapon.ItemType,
		OwnershipID: ownedWeapon.ID,
		StaticID:    &weaponBlueprint.R.WeaponModel.ID,
	}

	weaponSkin, err := boiler.WeaponSkins(boiler.WeaponSkinWhere.ID.EQ(weapon.EquippedWeaponSkinID.String)).One(tx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get weapon skin")
	}

	weaponSkinOwnership, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weaponSkin.ID)).One(tx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get weapon skin ownership")
	}

	weaponSkinSilo := &SiloSkin{
		Type:        "skin",
		OwnershipID: &weaponSkinOwnership.ID,
		StaticID:    &weaponSkin.BlueprintID,
	}

	weaponSilo.SkinID = weaponSkinSilo

	return weaponSilo, nil
}
