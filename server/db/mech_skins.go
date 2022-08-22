package db

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
)

// InsertNewMechSkin if modelID is nil it will return images of a random mech in this skin
func InsertNewMechSkin(tx boil.Executor, ownerID uuid.UUID, skin *server.BlueprintMechSkin, modelID *string) (*server.MechSkin, error) {
	// first insert the skin
	newSkin := boiler.MechSkin{
		BlueprintID:           skin.ID,
		GenesisTokenID:        skin.GenesisTokenID,
		LimitedReleaseTokenID: skin.LimitedReleaseTokenID,
		Level: skin.DefaultLevel,
	}

	err := newSkin.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("newSkin", newSkin).Msg("failed to insert")
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		skin.Collection,
		boiler.ItemTypeMechSkin,
		newSkin.ID,
		skin.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechSkin(tx, newSkin.ID, modelID)
}

// MechSkin if modelID is nil it will return images of a random mech in this skin
func MechSkin(trx boil.Executor, id string, modelID *string) (*server.MechSkin, error) {
	boilerMech, err := boiler.MechSkins(
		boiler.MechSkinWhere.ID.EQ(id),
		qm.Load(boiler.MechSkinRels.Blueprint),
	).One(trx)
	if err != nil {
		return nil, err
	}

	boilerMechCollectionDetails, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(id),
	).One(trx)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		boiler.MechModelSkinCompatibilityWhere.BlueprintMechSkinID.EQ(boilerMech.BlueprintID),
	}

	// if nil was passed in, we get a random one
	if modelID != nil && *modelID != "" {
		queryMods = append(queryMods, boiler.MechModelSkinCompatibilityWhere.MechModelID.EQ(*modelID))
	}

	mechSkinCompatabilityMatrix, err := boiler.MechModelSkinCompatibilities(
		queryMods...,
	).One(trx)
	if err != nil {
		return nil, err
	}

	return server.MechSkinFromBoiler(boilerMech, boilerMechCollectionDetails, mechSkinCompatabilityMatrix), nil
}

func MechSkins(id ...string) ([]*server.MechSkin, error) {
	var skins []*server.MechSkin
	boilerMechSkins, err := boiler.MechSkins(boiler.MechSkinWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, ms := range boilerMechSkins {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(ms.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		skins = append(skins, server.MechSkinFromBoiler(ms, boilerMechCollectionDetails, nil))
	}
	return skins, nil
}
