package fiat

import (
	"database/sql"
	"fmt"
	"math/rand"
	"server"
	"time"

	"server/db"
	"server/db/boiler"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ProductItemPackageContents struct {
	Mech        *server.Mech         `json:"mech,omitempty"`
	MechSkins   []*server.MechSkin   `json:"mech_skins,omitempty"`
	Weapons     []*server.Weapon     `json:"weapon,omitempty"`
	WeaponSkins []*server.WeaponSkin `json:"weapon_skins,omitempty"`
	PowerCore   *server.PowerCore    `json:"power_core,omitempty"`
}

// SendMysteryCrateToUser sends out the mystery crate to user.
func SendMysteryCrateToUser(conn *sql.Tx, pp *xsyn_rpcclient.XsynXrpcClient, userID string, productID string, quantity int) error {
	pl := gamelog.L.With().Str("fiat_product_id", productID).Int("quantity", quantity).Logger()
	errMsg := "Could not give package item, try again or contact support."

	storeCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.FiatProductID.EQ(productID),
	).One(conn)
	if err != nil {
		return terror.Error(err, "Failed to get crate for purchase, please try again or contact support.")
	}

	// TODO: issue refund or somehow let staff know about someone winning first?
	if (storeCrate.AmountSold + quantity) >= storeCrate.Amount {
		errMsg = fmt.Sprintf("player ID: %s, attempted to purchase sold out mystery crate", userID)
		err := fmt.Errorf(errMsg)
		pl.Error().Err(err).
			Msg(fmt.Sprintf(errMsg, userID))
		return terror.Error(err, "This mystery crate is sold out!")
	}

	// Assign multiple crate purchases
	var xsynAssets []*rpctypes.XsynAsset
	for i := 0; i < quantity; i++ {
		xsynAsset, err := assignAndRegisterPurchasedCrate(userID, storeCrate, conn)
		if err != nil {
			return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
		}

		// TODO: Not sure if we need one for fiat... maybe?
		// txItem := &boiler.StorePurchaseHistory{
		// 	PlayerID:    user.ID,
		// 	Amount:      storeCrate.Price,
		// 	ItemType:    "mystery_crate",
		// 	ItemID:      assignedCrate.ID,
		// 	Description: "purchased mystery crate",
		// 	TXID:        supTransactionID,
		// }

		// err = txItem.Insert(tx, boil.Infer())
		// if err != nil {
		// 	refundFunc()
		// 	gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
		// 	return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
		// }

		// resp = append(resp, Reward{
		// 	Crate:       assignedCrate,
		// 	Label:       storeCrate.MysteryCrateType,
		// 	ImageURL:    storeCrate.ImageURL,
		// 	LockedUntil: null.TimeFrom(assignedCrate.LockedUntil),
		// })

		xsynAssets = append(xsynAssets, xsynAsset)

	}

	err = pp.AssetRegister(xsynAssets...)
	if err != nil {
		pl.Error().Msg("failed to register to XSYN")
		return terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	return nil
}

// TODO: This is basically duplicated code from ws_store.go
func assignAndRegisterPurchasedCrate(userID string, storeCrate *boiler.StorefrontMysteryCrate, tx *sql.Tx) (*rpctypes.XsynAsset, error) {
	availableCrates, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.FactionID.EQ(storeCrate.FactionID),
		boiler.MysteryCrateWhere.Type.EQ(storeCrate.MysteryCrateType),
		boiler.MysteryCrateWhere.Purchased.EQ(false),
		boiler.MysteryCrateWhere.Opened.EQ(false),
		qm.Load(boiler.MysteryCrateRels.Blueprint),
	).All(tx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get available crates, please try again or contact support.")
	}

	faction, err := boiler.FindFaction(tx, storeCrate.FactionID)
	if err != nil {
		return nil, terror.Error(err, "Failed to find faction, please try again or contact support.")
	}

	//randomly assigning crate to user
	rand.Seed(time.Now().UnixNano())
	assignedCrate := availableCrates[rand.Intn(len(availableCrates))]

	//update purchased value
	assignedCrate.Purchased = true

	// set newly bought crates openable on staging/dev (this is so people cannot open already purchased crates and see what is in them)
	if server.IsDevelopmentEnv() || server.IsStagingEnv() {
		assignedCrate.LockedUntil = time.Now()
	}

	_, err = assignedCrate.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to update assigned crate information")
		return nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	collectionItem, err := db.InsertNewCollectionItem(tx,
		"supremacy-general",
		boiler.ItemTypeMysteryCrate,
		assignedCrate.ID,
		"",
		userID,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to insert into collection items")
		return nil, terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}
	storeCrate.AmountSold = storeCrate.AmountSold + 1
	_, err = storeCrate.Update(tx, boil.Whitelist(boiler.StorefrontMysteryCrateColumns.AmountSold))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
		return nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	//register
	assignedCrateServer := server.MysteryCrateFromBoiler(assignedCrate, collectionItem, null.String{})
	xsynAsset := rpctypes.ServerMysteryCrateToXsynAsset(assignedCrateServer, faction.Label)

	return xsynAsset, nil
}

// SendStarterPackageContentsToUser sends out the package contents to user.
func SendStarterPackageContentsToUser(conn *sql.Tx, pp *xsyn_rpcclient.XsynXrpcClient, userID string, productID string) error {
	errMsg := "Could not give package item, try again or contact support."

	productItems, err := boiler.FiatProductItems(
		boiler.FiatProductItemWhere.ProductID.EQ(productID),
	).All(conn)
	if err != nil {
		return terror.Error(err, "failed to get product items")
	}

	user, err := boiler.FindPlayer(conn, userID)
	if err != nil {
		return terror.Error(err, "failed to get user")
	}

	xsynAsserts := []*rpctypes.XsynAsset{}

	for _, item := range productItems {
		pl := gamelog.L.With().Interface("package", item).Logger()
		items := &ProductItemPackageContents{} // used for mech and weapon package item types
		singleItem := item.ItemType == boiler.FiatProductItemTypesSingleItem
		blueprintItems, err := item.ProductItemFiatProductItemBlueprints().All(conn)
		if err != nil {
			return terror.Error(err, "Could not get blueprints, try again or contact support.")
		}

		// Single Items
		if singleItem {
			// TODO: Redo this again :(
			continue
		}

		// Items
		blueprintMechs := []string{}
		blueprintMechSkins := []string{}
		blueprintWeapons := []string{}
		blueprintWeaponSkins := []string{}
		blueprintPowercores := []string{}

		for _, blueprintItem := range blueprintItems {
			if blueprintItem.MechBlueprintID.Valid {
				blueprintMechs = append(blueprintMechs, blueprintItem.MechBlueprintID.String)
			} else if blueprintItem.WeaponBlueprintID.Valid {
				blueprintWeapons = append(blueprintWeapons, blueprintItem.WeaponBlueprintID.String)
			} else if blueprintItem.MechSkinBlueprintID.Valid {
				blueprintMechSkins = append(blueprintMechSkins, blueprintItem.MechSkinBlueprintID.String)
			} else if blueprintItem.WeaponSkinBlueprintID.Valid {
				blueprintWeaponSkins = append(blueprintWeaponSkins, blueprintItem.WeaponSkinBlueprintID.String)
			} else if blueprintItem.PowerCoreBlueprintID.Valid {
				blueprintPowercores = append(blueprintPowercores, blueprintItem.PowerCoreBlueprintID.String)
			}
		}

		for _, blueprintItemID := range blueprintMechs {
			mechSkinBlueprints, err := db.BlueprintMechSkinSkins(conn, blueprintMechSkins)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get mech blueprint from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get mech, try again or contact support.")
			}

			// insert the non default skin with the mech
			rarerSkinIndex := 0
			for i, skin := range mechSkinBlueprints {
				if skin.Tier != "COLOSSAL" {
					rarerSkinIndex = i
				}
			}

			mechBlueprint, err := db.BlueprintMech(blueprintItemID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get mech from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get mech, try again or contact support.")
			}

			insertedMech, insertedMechSkin, err := db.InsertNewMechAndSkin(conn, uuid.FromStringOrNil(user.ID), mechBlueprint, mechSkinBlueprints[rarerSkinIndex])
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to insert mech from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get mech, try again or contact support.")
			}
			items.Mech = insertedMech
			items.MechSkins = append(items.MechSkins, insertedMechSkin)

			// remove the already inserted skin
			mechSkinBlueprints = append(mechSkinBlueprints[:rarerSkinIndex], mechSkinBlueprints[rarerSkinIndex+1:]...)

			// insert the rest of the skins
			for _, skin := range mechSkinBlueprints {
				mechSkin, err := db.InsertNewMechSkin(conn, uuid.FromStringOrNil(user.ID), skin, &insertedMech.BlueprintID)
				if err != nil {
					pl.Error().Err(err).
						Msg(fmt.Sprintf("failed to insert mech skin from product item: %s, for user: %s", item.ID, user.ID))
					return terror.Error(err, "Could not get mech, try again or contact support.")
				}
				items.MechSkins = append(items.MechSkins, mechSkin)
			}
		}

		for _, blueprintItemID := range blueprintWeapons {
			weaponSkinBlueprints, err := db.BlueprintWeaponSkins(blueprintWeaponSkins)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get weapon skin from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get weapon, try again or contact support.")
			}

			// insert the non default skin with the weapon
			rarerSkinIndex := 0
			for i, skin := range weaponSkinBlueprints {
				if skin.Tier != "COLOSSAL" {
					rarerSkinIndex = i
				}
			}

			bp, err := db.BlueprintWeapon(blueprintItemID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get weapon blueprint from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get weapon, try again or contact support.")
			}

			weapon, weaponSkin, err := db.InsertNewWeapon(conn, uuid.FromStringOrNil(user.ID), bp, weaponSkinBlueprints[rarerSkinIndex])
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to insert weapon from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get weapon, try again or contact support.")
			}
			items.Weapons = append(items.Weapons, weapon)
			items.WeaponSkins = append(items.WeaponSkins, weaponSkin)

			for i, bpws := range blueprintWeaponSkins {
				if bpws == weaponSkin.BlueprintID {
					blueprintWeaponSkins = append(blueprintWeaponSkins[:i], blueprintWeaponSkins[i+1:]...)
					break
				}
			}
		}

		for _, blueprintItemID := range blueprintWeaponSkins {
			bp, err := db.BlueprintWeaponSkin(blueprintItemID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get weapon skin blueprint from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get weapon skin, try again or contact support.")
			}
			weaponSkin, err := db.InsertNewWeaponSkin(conn, uuid.FromStringOrNil(user.ID), bp, nil)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to insert weapon skin from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get weapon skin, try again or contact support.")
			}
			items.WeaponSkins = append(items.WeaponSkins, weaponSkin)
		}

		for _, blueprintItemID := range blueprintPowercores {
			bp, err := db.BlueprintPowerCore(blueprintItemID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get powercore blueprint from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get powercore, try again or contact support.")
			}

			powerCore, err := db.InsertNewPowerCore(conn, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to insert powercore from product item: %s, for user: %s", item.ID, user.ID))
				return terror.Error(err, "Could not get powercore, try again or contact support.")
			}
			items.PowerCore = powerCore
		}

		// Attach parts to items
		if item.ItemType == boiler.FiatProductItemTypesMechPackage {
			eod, err := db.MechEquippedOnDetails(conn, items.Mech.ID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get MechEquippedOnDetails when giving product item mech: %s", item.ID))
				return terror.Error(err, errMsg)
			}
			rarerSkin := items.MechSkins[0]
			for _, skin := range items.MechSkins {
				if skin.Tier != "COLOSSAL" {
					rarerSkin = skin
				}
			}

			rarerSkin.EquippedOn = null.StringFrom(items.Mech.ID)
			rarerSkin.EquippedOnDetails = eod
			xsynAsserts = append(xsynAsserts, rpctypes.ServerMechSkinsToXsynAsset(items.MechSkins)...)

			//attach powercore to mech - mech
			err = db.AttachPowerCoreToMech(conn, user.ID, items.Mech.ID, items.PowerCore.ID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to attach powercore to check when giving product item mech: %s", item.ID))
				return terror.Error(err, errMsg)
			}
			items.PowerCore.EquippedOn = null.StringFrom(items.Mech.ID)
			items.PowerCore.EquippedOnDetails = eod
			xsynAsserts = append(xsynAsserts, rpctypes.ServerPowerCoresToXsynAsset([]*server.PowerCore{items.PowerCore})...)

			//attach weapons
			for i, weapon := range items.Weapons {
				err := db.AttachWeaponToMech(conn, user.ID, items.Mech.ID, weapon.ID)
				if err != nil {
					pl.Error().Err(err).
						Msg(fmt.Sprintf("failed to attach weapon to check when giving product item mech: %s", item.ID))
					return terror.Error(err, errMsg)
				}
				weapon.EquippedOn = null.StringFrom(items.Mech.ID)
				weapon.EquippedOnDetails = eod

				wod, err := db.WeaponEquippedOnDetails(conn, items.Weapons[0].ID)
				if err != nil {
					pl.Error().Err(err).
						Msg(fmt.Sprintf("failed to get WeaponEquippedOnDetails to check when giving product item mech: %s", item.ID))
					return terror.Error(err, errMsg)
				}
				weapon.WeaponSkin = items.WeaponSkins[i]
				weapon.WeaponSkin.EquippedOn = null.StringFrom(items.Weapons[i].ID)
				weapon.WeaponSkin.EquippedOnDetails = wod
				xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponSkinsToXsynAsset([]*server.WeaponSkin{items.WeaponSkins[i]})...)
			}
			xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponsToXsynAsset(items.Weapons)...)

			mech, err := db.Mech(conn, items.Mech.ID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get final mech when giving product item mech: %s", item.ID))
				return terror.Error(err, errMsg)
			}
			mech.ChassisSkin = rarerSkin
			xsynAsserts = append(xsynAsserts, rpctypes.ServerMechsToXsynAsset([]*server.Mech{mech})...)
		}

		if item.ItemType == boiler.FiatProductItemTypesWeaponPackage {
			wod, err := db.WeaponEquippedOnDetails(conn, items.Weapons[0].ID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get WeaponEquippedOnDetails when giving product item weapon: %s", item.ID))
				return terror.Error(err, errMsg)
			}

			//attach weapon_skin to weapon -weapon
			if len(items.Weapons) != 1 {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("invalid amount of weapons when giving product item weapon: %s", item.ID))
				return terror.Error(err, errMsg)
			}
			items.WeaponSkins[0].EquippedOn = null.StringFrom(items.Weapons[0].ID)
			items.WeaponSkins[0].EquippedOnDetails = wod
			xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponSkinsToXsynAsset([]*server.WeaponSkin{items.WeaponSkins[0]})...)

			weapon, err := db.Weapon(conn, items.Weapons[0].ID)
			if err != nil {
				pl.Error().Err(err).
					Msg(fmt.Sprintf("failed to get final weapon when giving product item weapon: %s", item.ID))
				return terror.Error(err, errMsg)
			}
			xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponsToXsynAsset([]*server.Weapon{weapon})...)
		}
	}

	// Register new assets
	err = pp.AssetsRegister(xsynAsserts) // register new assets
	if err != nil {
		gamelog.L.Error().Err(err).Msg("issue inserting new assets to xsyn for RegisterAllNewAssets")
		return terror.Error(err, "Could not give product items, try again or contact support.")
	}

	// Update Amount Sold
	q := fmt.Sprintf(
		`UPDATE %[1]s SET %[2]s = %[2]s + 1 WHERE %[3]s = $1`,
		boiler.TableNames.FiatProducts,
		boiler.FiatProductColumns.AmountSold,
		boiler.FiatProductColumns.ID,
	)
	_, err = conn.Exec(q, productID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("product_id", productID).Msg("issue updating amount sold for product")
		return terror.Error(err, "Could not give product items, try again or contact support.")
	}
	return nil
}
