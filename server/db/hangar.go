package db

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
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
	SELECT
    distinct on ( ms.blueprint_id) ms.blueprint_id as skin_id,
                                   ci.item_type    as type,
                                   ci.id           as ownership_id,
                                   m.model_id  	as static_id
	FROM collection_items ci
         INNER JOIN mechs m on
        	m.id = ci.item_id
         INNER JOIN mech_skin ms on
        	ms.id = coalesce(
            	m.chassis_skin_id,
            	(select default_chassis_skin_id from mech_models mm where mm.id = m.model_id)
        	)
	WHERE ci.owner_id = $1
  	AND ci.item_type = 'mech'
  	AND ci.xsyn_locked=false
	ORDER BY ms.blueprint_id, m.genesis_token_id NULLS FIRST, m.limited_release_token_id NULLS FIRST;
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
		mst := SiloType{}

		err := rows.Scan(&mst.SkinIDStr, &mst.Type, &mst.OwnershipID, &mst.StaticID)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		mechSiloType = append(mechSiloType, &mst)
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
				weaponSkinBlueprintID := ""
				var weaponSkinCollectionID *string

				weaponBlueprintFromMechSkin, err := boiler.BlueprintWeaponSkins(
					boiler.BlueprintWeaponSkinWhere.Label.EQ(mech.ChassisSkin.Label),
					boiler.BlueprintWeaponSkinWhere.WeaponType.EQ(weapon.WeaponType),
				).One(gamedb.StdConn)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
					return nil, err
				}
				if err == nil {
					weaponSkinBlueprintID = weaponBlueprintFromMechSkin.ID
				} else if !weapon.EquippedWeaponSkinID.Valid {
					weaponModel, err := boiler.WeaponModels(
						boiler.WeaponModelWhere.ID.EQ(mech.ModelID),
					).One(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
						continue
					}
					weaponSkinBlueprintID = weaponModel.DefaultSkinID
				} else {
					skinBP, err := boiler.FindWeaponSkin(gamedb.StdConn, weapon.EquippedWeaponSkinID.String)
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
						continue
					}
					weaponSkinBlueprintID = skinBP.BlueprintID

					weaponSkinCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weapon.EquippedWeaponSkinID.String), qm.Select(boiler.CollectionItemColumns.ID)).One(gamedb.StdConn)
					if err != nil {
						continue
					}
					weaponSkinCollectionID = &weaponSkinCollection.ID
				}

				weaponCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weapon.ID), qm.Select(boiler.CollectionItemColumns.ID)).One(gamedb.StdConn)
				if err != nil {
					continue
				}

				newAttribute := MechSiloAccessories{
					Type:        "weapon",
					OwnershipID: weaponCollection.ID,
					StaticID:    weapon.WeaponModelID.String,
					Skin: &SiloSkin{
						Type:        "skin",
						OwnershipID: weaponSkinCollectionID,
						StaticID:    &weaponSkinBlueprintID,
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
          ci.id    					as ownership_id,
          smc.id 					as mystery_crate_id,
          mc.locked_until        	as can_open_on
	FROM collection_items ci
            INNER JOIN mystery_crate mc on
            mc.id = ci.item_id AND mc.opened = false
    	INNER JOIN storefront_mystery_crates smc on
    	smc.mystery_crate_type = mc."type"
    	AND smc.faction_id = mc.faction_id
	WHERE ci.owner_id = $1
  		AND ci.item_type = 'mystery_crate'
  		AND ci.xsyn_locked=false
	ORDER BY mc.type;
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

		canOpenOnStr := canOpenOn.UTC().Format("2006-01-02T15:04:05.000Z")

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

func GetUserMechHangarItemsWithMechID(mech *server.Mech, userID string, trx boil.Executor) (*SiloType, error) {
	if trx == nil {
		return nil, terror.Error(fmt.Errorf("not tx provided"), "Failed to get user mech hangar items")
	}
	mechSiloType := &SiloType{
		Type:        "mech",
		OwnershipID: mech.CollectionItemID,
		StaticID:    &mech.Model.ID,
	}

	var mechAttributes []MechSiloAccessories

	if mech.IsCompleteLimited() || mech.IsCompleteGenesis() {

		if mech.ChassisSkin == nil {
			return nil, terror.Error(fmt.Errorf("default mech with no chassis skin"), "Fail to get chassis skin for genesis or limited mech")
		}

		mechDefaultSkin := &SiloSkin{
			Type:        "skin",
			OwnershipID: nil,
			StaticID:    &mech.ChassisSkin.BlueprintID,
		}
		mechSiloType.SkinID = mechDefaultSkin
		mechAttributes = []MechSiloAccessories{}
		return mechSiloType, nil
	}

	mechSkin := &SiloSkin{
		Type:        "skin",
		OwnershipID: nil,
		StaticID:    &mech.ChassisSkin.BlueprintID,
	}

	if mech.ChassisSkinID.Valid {
		mechSkinOwnership, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.ChassisSkinID.String)).One(trx)
		if err != nil {
			return nil, terror.Error(err, "Failed to get mech skin ownership")
		}

		mechSkin.OwnershipID = &mechSkinOwnership.ID
	}

	mechSiloType.SkinID = mechSkin

	if len(mech.Weapons) > 0 {
		for _, weapon := range mech.Weapons {
			weaponSkinBlueprintID := ""
			var weaponSkinCollectionID *string

			weaponBlueprintFromMechSkin, err := boiler.BlueprintWeaponSkins(
				boiler.BlueprintWeaponSkinWhere.Label.EQ(mech.ChassisSkin.Label),
				boiler.BlueprintWeaponSkinWhere.WeaponType.EQ(weapon.WeaponType),
			).One(trx)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
				return nil, err
			}
			if err == nil {
				weaponSkinBlueprintID = weaponBlueprintFromMechSkin.ID
			} else if !weapon.EquippedWeaponSkinID.Valid {
				weaponModel, err := boiler.WeaponModels(
					boiler.WeaponModelWhere.ID.EQ(mech.ModelID),
				).One(trx)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
					continue
				}
				weaponSkinBlueprintID = weaponModel.DefaultSkinID
			} else {
				skinBP, err := boiler.FindWeaponSkin(trx, weapon.EquippedWeaponSkinID.String)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to get default skin for weapon skin for hangar")
					continue
				}
				weaponSkinBlueprintID = skinBP.BlueprintID

				weaponSkinCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weapon.EquippedWeaponSkinID.String), qm.Select(boiler.CollectionItemColumns.ID)).One(trx)
				if err != nil {
					continue
				}
				weaponSkinCollectionID = &weaponSkinCollection.ID
			}

			weaponCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weapon.ID), qm.Select(boiler.CollectionItemColumns.ID)).One(trx)
			if err != nil {
				continue
			}

			newAttribute := MechSiloAccessories{
				Type:        "weapon",
				OwnershipID: weaponCollection.ID,
				StaticID:    weapon.WeaponModelID.String,
				Skin: &SiloSkin{
					Type:        "skin",
					OwnershipID: weaponSkinCollectionID,
					StaticID:    &weaponSkinBlueprintID,
				},
			}

			mechAttributes = append(mechAttributes, newAttribute)
		}
	}

	if mech.PowerCoreID.Valid {
		powerCoreCollection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.PowerCoreID.String)).One(trx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, terror.Error(err, "Failed to get collection item for power core")
		}

		powerCoreBlueprint, err := boiler.PowerCores(boiler.PowerCoreWhere.ID.EQ(mech.PowerCoreID.String)).One(trx)
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

func GetUserWeaponHangarItemsWithID(weapon *server.Weapon, userID string, trx boil.Executor) (*SiloType, error) {
	if trx == nil {
		return nil, terror.Error(fmt.Errorf("no tx provided"), "Failed to get weapon hangar details")
	}

	if weapon.EquippedOn.Valid || !weapon.EquippedWeaponSkinID.Valid {
		return nil, terror.Error(fmt.Errorf("weapon not availiable in hangar"), "Weapon not available on hangar by itself")
	}

	weaponBlueprint, err := boiler.BlueprintWeapons(boiler.BlueprintWeaponWhere.ID.EQ(weapon.BlueprintID), qm.Load(boiler.BlueprintWeaponRels.WeaponModel)).One(trx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get blueprint weapon")
	}

	weaponSilo := &SiloType{
		Type:        weapon.CollectionItem.ItemType,
		OwnershipID: weapon.CollectionItemID,
		StaticID:    &weaponBlueprint.R.WeaponModel.ID,
	}

	weaponSkin, err := boiler.WeaponSkins(boiler.WeaponSkinWhere.ID.EQ(weapon.EquippedWeaponSkinID.String)).One(trx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get weapon skin")
	}

	weaponSkinOwnership, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(weaponSkin.ID)).One(trx)
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
