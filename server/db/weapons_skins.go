package db

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewWeaponSkin(ownerID uuid.UUID, blueprintWeaponSkin *server.BlueprintWeaponSkin) (*server.WeaponSkin, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	//getting blueprintWeaponSkin model to get default skin id to get image url on blueprint blueprintWeaponSkin skins
	weaponModel, err := boiler.WeaponModels(
		boiler.WeaponModelWhere.ID.EQ(blueprintWeaponSkin.WeaponModelID),
		qm.Load(boiler.WeaponModelRels.DefaultSkin),
	).One(tx)
	if err != nil {
		return nil, terror.Error(err)
	}

	if weaponModel.R == nil || weaponModel.R.DefaultSkin == nil {
		return nil, terror.Error(fmt.Errorf("could not find default skin relationship to blueprintWeaponSkin"), "Could not find blueprintWeaponSkin default skin relationship, try again or contact support")
	}

	//should only have one in the arr
	bpws := weaponModel.R.DefaultSkin

	newWeaponSkin := boiler.WeaponSkin{
		BlueprintID:   blueprintWeaponSkin.ID,
		OwnerID:       ownerID.String(),
		Label:         blueprintWeaponSkin.Label,
		WeaponType:    blueprintWeaponSkin.WeaponType,
		EquippedOn:    null.String{},
		Tier:          blueprintWeaponSkin.Tier,
		CreatedAt:     blueprintWeaponSkin.CreatedAt,
		WeaponModelID: blueprintWeaponSkin.WeaponModelID,
	}

	err = newWeaponSkin.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	//change img, avatar etc. here but have to get it from blueprint blueprintWeaponSkin skins
	_, err = InsertNewCollectionItem(tx,
		blueprintWeaponSkin.Collection,
		boiler.ItemTypeWeaponSkin,
		newWeaponSkin.ID,
		blueprintWeaponSkin.Tier,
		ownerID.String(),
		bpws.ImageURL,
		bpws.CardAnimationURL,
		bpws.AvatarURL,
		bpws.LargeImageURL,
		bpws.BackgroundColor,
		bpws.AnimationURL,
		bpws.YoutubeURL,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	return WeaponSkin(newWeaponSkin.ID)
}

func WeaponSkin(id string) (*server.WeaponSkin, error) {
	boilerWeaponSkin, err := boiler.FindWeaponSkin(gamedb.StdConn, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(gamedb.StdConn)
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
