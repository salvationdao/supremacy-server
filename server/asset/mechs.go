package asset

import (
	"fmt"
	"server/db/boiler"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TransferMechToNewOwner(
	conn boil.Executor,
	mechID,
	toID string,
	xsynLocked bool,
	assetHidden null.String,
) ([]*boiler.CollectionItem, error) {
	itemIDsToTransfer := []string{}

	// update mech owner
	updated, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).UpdateAll(conn, boiler.M{
		"owner_id":    toID,
		"xsyn_locked": xsynLocked,
	})
	if err != nil {
		return nil, err
	}
	if updated != 1 {
		return nil, fmt.Errorf("expected to update 1 mech but updated %d", updated)
	}

	// get equipped mech skin
	mSkins, err := boiler.MechSkins(
		boiler.MechSkinWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(conn)
	if err != nil {
		return nil, err
	}
	for _, itm := range mSkins {
		itemIDsToTransfer = append(itemIDsToTransfer, itm.ID)
	}

	// get equipped mech power core
	mPc, err := boiler.PowerCores(
		boiler.PowerCoreWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(conn)
	if err != nil {
		return nil, err
	}
	for _, itm := range mPc {
		itemIDsToTransfer = append(itemIDsToTransfer, itm.ID)
	}

	// get equipped mech animations
	mAnim, err := boiler.MechAnimations(
		boiler.MechAnimationWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(conn)
	if err != nil {
		return nil, err
	}
	for _, itm := range mAnim {
		itemIDsToTransfer = append(itemIDsToTransfer, itm.ID)
	}

	// get equipped mech weapons
	mWpn, err := boiler.Weapons(
		boiler.WeaponWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(conn)
	if err != nil {
		return nil, err
	}
	for _, itm := range mWpn {
		itemIDsToTransfer = append(itemIDsToTransfer, itm.ID)
		// get equipped mech weapon skins
		mWpnSkin, err := boiler.WeaponSkins(
			boiler.WeaponSkinWhere.EquippedOn.EQ(null.StringFrom(itm.ID)),
		).All(conn)
		if err != nil {
			return nil, err
		}
		for _, wItem := range mWpnSkin {
			itemIDsToTransfer = append(itemIDsToTransfer, wItem.ID)
		}
	}

	// get equipped mech utilities
	mUtil, err := boiler.Utilities(
		boiler.UtilityWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(conn)
	if err != nil {
		return nil, err
	}
	for _, itm := range mUtil {
		itemIDsToTransfer = append(itemIDsToTransfer, itm.ID)
	}

	// update!
	_, err = boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToTransfer),
	).UpdateAll(conn, boiler.M{
		"owner_id":     toID,
		"asset_hidden": assetHidden,
	})
	if err != nil {
		return nil, err
	}

	// now lets also transfer all the assets on xsyn too!
	colItems, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToTransfer),
	).All(conn)
	if err != nil {
		return nil, err
	}

	return colItems, nil
}
