package db

import (
	"database/sql"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

type SiloType struct {
	Type           string                `db:"type" json:"type"`
	OwnershipID    string                `db:"ownership_id" json:"ownership_id"`
	MechID         *string               `db:"mech_id" json:"mech_id,omitempty"`
	SkinIDStr      *string               `db:"skin_id" json:"skin_id_str,omitempty"`
	SkinID         *SiloSkin             `json:"skin,omitempty"`
	MysteryCrateID *string               `db:"mystery_crate_id" json:"mystery_crate_id,omitempty"`
	CanOpenOn      *string               `db:"can_open_on" json:"can_open_on,omitempty"`
	Accessories    []MechSiloAccessories `json:"accessories,omitempty"`
}

type MechSiloAccessories struct {
	Type        string    `json:"type"`
	OwnershipID string    `json:"ownership_id"`
	StaticID    string    `json:"static_id"`
	Skin        *SiloSkin `json:"skin,omitempty"`
}

type SiloSkin struct {
	OwnershipID *string `json:"ownership_id,omitempty"`
	SkinID      *string `json:"skin_id,omitempty"`
}

func GetUserMechHangarItems(userID string) ([]*SiloType, error) {
	q := `
	SELECT 	ci.item_type    as type,
			ci.id           as ownership_id,
       		m.model_id  	as mech_id,
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

		err := rows.Scan(&mst.Type, &mst.OwnershipID, &mst.MechID, &mst.SkinIDStr)
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
		mech, err := Mech(gamedb.StdConn, collectionItem.ItemID)
		if err != nil {
			return nil, terror.Error(err, "Failed to get mech info")
		}
		if mech.IsCompleteLimited() || mech.IsCompleteGenesis() {
			continue
		}
		var mechAttributes []MechSiloAccessories

		mechSkin := &SiloSkin{
			SkinID: mechSilo.SkinIDStr,
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
				newAttribute := MechSiloAccessories{
					Type:        "weapon",
					OwnershipID: weaponCollection.ID,
					StaticID:    weapon.BlueprintID,
					Skin: &SiloSkin{
						OwnershipID: &weapon.EquippedWeaponSkinID.String,
						SkinID:      &weaponStringID,
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
		err := rows.Scan(&mst.Type, &mst.OwnershipID, &mst.MysteryCrateID, &canOpenOn)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		canOpenOnStr := canOpenOn.Format(time.UnixDate)

		mst.CanOpenOn = &canOpenOnStr

		mechSiloType = append(mechSiloType, mst)
	}

	return mechSiloType, nil
}
