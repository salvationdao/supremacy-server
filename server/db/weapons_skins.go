package db

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func InsertNewWeaponSkin(tx *sql.Tx, ownerID uuid.UUID, blueprintWeaponSkin *server.BlueprintWeaponSkin) (*server.WeaponSkin, error) {
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

	return WeaponSkin(tx, newWeaponSkin.ID)
}

func WeaponSkin(trx boil.Executor, id string) (*server.WeaponSkin, error) {
	boilerWeaponSkin, err := boiler.FindWeaponSkin(trx, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(trx)
	if err != nil {
		return nil, err
	}

	return server.WeaponSkinFromBoiler(boilerWeaponSkin, boilerMechCollectionDetails), nil
}

func WeaponSkins(id ...string) ([]*server.WeaponSkin, error) {
	var weaponSkins []*server.WeaponSkin
	boilerWeaponSkins, err := boiler.WeaponSkins(boiler.WeaponWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, bws := range boilerWeaponSkins {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(bws.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		weaponSkins = append(weaponSkins, server.WeaponSkinFromBoiler(bws, boilerMechCollectionDetails))
	}

	return weaponSkins, nil
}
