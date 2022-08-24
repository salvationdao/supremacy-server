package db

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
)

func InsertNewWeaponSkin(tx *sql.Tx, ownerID uuid.UUID, blueprintWeaponSkin *server.BlueprintWeaponSkin, modelID *string) (*server.WeaponSkin, error) {
	newWeaponSkin := boiler.WeaponSkin{
		BlueprintID:   blueprintWeaponSkin.ID,
		EquippedOn:    null.String{},
		CreatedAt:     blueprintWeaponSkin.CreatedAt,
	}

	err := newWeaponSkin.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		blueprintWeaponSkin.Collection,
		boiler.ItemTypeWeaponSkin,
		newWeaponSkin.ID,
		blueprintWeaponSkin.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return WeaponSkin(tx, newWeaponSkin.ID, modelID)
}

func WeaponSkin(tx boil.Executor, id string, blueprintID *string) (*server.WeaponSkin, error) {
	boilerWeaponSkin, err := boiler.WeaponSkins(
		boiler.WeaponSkinWhere.ID.EQ(id),
		qm.Load(boiler.WeaponSkinRels.Blueprint),
		).One(tx)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		boiler.WeaponModelSkinCompatibilityWhere.BlueprintWeaponSkinID.EQ(boilerWeaponSkin.BlueprintID),
	}

	if blueprintID != nil && *blueprintID != "" {
		queryMods = append(queryMods, boiler.WeaponModelSkinCompatibilityWhere.WeaponModelID.EQ(*blueprintID))
	}

	weaponSkinCompatMatrix, err := boiler.WeaponModelSkinCompatibilities(
		queryMods...
		).One(tx)
	if err != nil {
		return nil, err
	}
	return server.WeaponSkinFromBoiler(boilerWeaponSkin, boilerMechCollectionDetails, weaponSkinCompatMatrix), nil
}
