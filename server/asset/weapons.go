package asset

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server/db/boiler"
)

func TransferWeaponToNewOwner(
	conn boil.Executor,
	weaponID,
	toID string,
	xsynLocked bool,
	assetHidden null.String,
	xsynAssetTransfer func(colItems []*boiler.CollectionItem) error,
) error {
	itemIDsToTransfer := []string{}

	// update mech owner
	updated, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(weaponID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
	).UpdateAll(conn, boiler.M{
		"owner_id": toID,
		"xsyn_locked": xsynLocked,
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return fmt.Errorf("expected to update 1 weapon but updated %d", updated)
	}
	// get equipped mech weapon skins
	mWpnSkin, err := boiler.WeaponSkins(
		boiler.WeaponSkinWhere.EquippedOn.EQ(null.StringFrom(weaponID)),
	).All(conn)
	if err != nil {
		return err
	}
	for _, itm := range mWpnSkin {
		itemIDsToTransfer = append(itemIDsToTransfer, itm.ID)
	}

	// update!
	_, err = boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToTransfer),
	).UpdateAll(conn, boiler.M{
		"owner_id": toID,
		"asset_hidden": assetHidden,
	})
	if err != nil {
		return err
	}

	// now lets also transfer all the assets on xsyn too!
	colItems, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToTransfer),
	).All(conn)
	if err != nil {
		return err
	}
	err = xsynAssetTransfer(colItems)
	if err != nil {
		return err
	}

	return nil
}
