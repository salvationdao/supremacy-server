package comms

import (
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AssetUnlockFromSupremacyResp struct {
}

type AssetUnlockFromSupremacyReq struct {
	ApiKey          string `json:"api_key,omitempty"`
	OwnerID         string `json:"owner_id,omitempty"`
	Hash            string `json:"hash,omitempty"`
	TransferEventID int64  `json:"transfer_event_id"`
}

// AssetUnlockFromSupremacyHandler request a lock of an asset
func (s *S) AssetUnlockFromSupremacyHandler(req AssetUnlockFromSupremacyReq, resp *AssetUnlockFromSupremacyResp) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetUnlockFromSupremacyHandler")
		return err
	}
	defer tx.Rollback()

	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(req.OwnerID),
		boiler.CollectionItemWhere.Hash.EQ(req.Hash),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetUnlockFromSupremacyHandler")
		return err
	}

	if collectionItem.XsynLocked {
		return nil
	}

	// check if asset is equipped // TODO after composable stuff comes out, we can change this to just unequipped it
	switch collectionItem.ItemType {
	case "utility":
		ult, err := boiler.FindUtility(tx, collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - boiler.FindUtility - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
		if ult.EquippedOn.Valid && ult.EquippedOn.String != "" {
			return fmt.Errorf("asset is equipped to another object, unequip first to transfer")
		}
	case "weapon":
		wpn, err := boiler.FindWeapon(tx, collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - boiler.FindWeapon - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
		if wpn.EquippedOn.Valid && wpn.EquippedOn.String != "" {
			return fmt.Errorf("asset is equipped to another object, unequip first to transfer")
		}
		// we need to set the "asset_hidden" on all the equipped assets to this weapon
		err = db.WeaponSetAllEquippedAssetsAsHidden(tx, collectionItem.ItemID, null.StringFrom("Equipped on asset that doesn't live on Supremacy."))
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - WeaponSetAllEquippedAssetsAsHidden - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
	case "mech_skin":
		ms, err := boiler.FindMechSkin(tx, collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - boiler.FindMechSkin - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
		if ms.EquippedOn.Valid && ms.EquippedOn.String != "" {
			return fmt.Errorf("asset is equipped to another object, unequip first to transfer")
		}
	case "mech_animation":
		ma, err := boiler.FindMechAnimation(tx, collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - boiler.FindMechAnimation - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
		if ma.EquippedOn.Valid && ma.EquippedOn.String != "" {
			return fmt.Errorf("asset is equipped to another object, unequip first to transfer")
		}
	case "power_core":
		pc, err := boiler.FindPowerCore(tx, collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - boiler.FindPowerCore - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
		if pc.EquippedOn.Valid && pc.EquippedOn.String != "" {
			return fmt.Errorf("asset is equipped to another object, unequip first to transfer")
		}
	case "weapon_skin":
		ws, err := boiler.FindWeaponSkin(tx, collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - boiler.FindWeaponSkin - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
		if ws.EquippedOn.Valid && ws.EquippedOn.String != "" {
			return fmt.Errorf("asset is equipped to another object, unequip first to transfer")
		}
	case "mystery_crate":
		//		these can't be equipped so all gucci
	case "mech":
		// we need to set the "asset_hidden" on all the equipped assets to this mech
		err = db.MechSetAllEquippedAssetsAsHidden(tx, collectionItem.ItemID, null.StringFrom("Equipped on asset that doesn't live on Supremacy."))
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - MechSetAllEquippedAssetsAsHidden - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
	}

	// TODO: store transfer event ID

	itemUUID, err := uuid.FromString(collectionItem.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("convert asset id to uuid - AssetUnlockFromSupremacyHandler")
		return err
	}

	if collectionItem.LockedToMarketplace {
		err = db.MarketplaceSaleArchiveByItemID(tx, itemUUID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
	}

	collectionItem.XsynLocked = true
	collectionItem.LockedToMarketplace = false
	_, err = collectionItem.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetUnlockFromSupremacyHandler")
		return err
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - failed to commit - AssetUnlockFromSupremacyHandler")
		return err
	}

	return nil
}

type AssetLockToSupremacyResp struct {
}

type AssetLockToSupremacyReq struct {
	ApiKey          string `json:"api_key,omitempty"`
	OwnerID         string `json:"owner_id,omitempty"`
	Hash            string `json:"hash,omitempty"`
	TransferEventID int64  `json:"transfer_event_id"`
	MarketLocked    bool   `json:"market_locked"`
}

// AssetLockToSupremacyHandler locks an asset to supremacy
func (s *S) AssetLockToSupremacyHandler(req AssetLockToSupremacyReq, resp *AssetLockToSupremacyResp) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetLockToSupremacyHandler")
		return err
	}
	defer tx.Rollback()

	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(req.OwnerID),
		boiler.CollectionItemWhere.Hash.EQ(req.Hash),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetLockToSupremacyHandler")
		return err
	}

	if !collectionItem.XsynLocked && collectionItem.MarketLocked == req.MarketLocked {
		return nil
	}

	switch collectionItem.ItemType {
	case "utility":
	case "weapon":
		// we need to set the "asset_hidden" on all the equipped assets to this weapon
		err = db.WeaponSetAllEquippedAssetsAsHidden(tx, collectionItem.ItemID, null.String{
			String: "",
			Valid:  false,
		})
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - WeaponSetAllEquippedAssetsAsHidden - AssetLockToSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
	case "mech_skin":
	case "mech_animation":
	case "power_core":
	case "weapon_skin":
	case "mystery_crate":
	case "mech":
		// we need to set the "asset_hidden" to null on all the equipped assets to this mech
		err = db.MechSetAllEquippedAssetsAsHidden(tx, collectionItem.ItemID, null.String{
			String: "",
			Valid:  false,
		})
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Str("collectionItem.ItemID", collectionItem.ItemID).Msg("failed to unlock asset - MechSetAllEquippedAssetsAsHidden - AssetLockToSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
	}

	collectionItem.XsynLocked = false
	collectionItem.MarketLocked = req.MarketLocked
	_, err = collectionItem.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetLockToSupremacyHandler")
		return err
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - failed to commit - AssetLockToSupremacyHandler")
		return err
	}
	// TODO: store transfer event ID

	return nil
}

type Asset1155LockToSupremacyResp struct {
}

type Asset1155LockToSupremacyReq struct {
	ApiKey          string `json:"api_key,omitempty"`
	OwnerID         string `json:"owner_id,omitempty"`
	Amount          int    `json:"amount"`
	TokenID         int    `json:"token_id"`
	TransferEventID int64  `json:"transfer_event_id"`
}

var TransferBucket = leakybucket.NewCollector(0.5, 1, true)

// KeycardTransferToSupremacyHandler transfer keycard to supremacy
func (s *S) KeycardTransferToSupremacyHandler(req Asset1155LockToSupremacyReq, resp *AssetLockToSupremacyResp) error {
	b := TransferBucket.Add(fmt.Sprintf("%s_%d", req.OwnerID, req.TokenID), 1)
	if b == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many request made for transfer")
	}
	asset, err := db.CreateOrGetKeycard(req.OwnerID, req.TokenID)
	if err != nil {
		return terror.Error(err, "Failed to create or get player keycard")
	}

	asset.Count += req.Amount

	_, err = asset.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerKeycardColumns.Count))
	if err != nil {
		return terror.Error(err, "Failed to update amount")
	}

	// TODO: store transfer event ID

	return nil
}

type Asset1155FromSupremacyResp struct {
	Label        string      `json:"label"`
	Description  string      `json:"description"`
	ImageURL     string      `json:"image_url"`
	AnimationURL null.String `json:"animation_url"`
	KeycardGroup string      `json:"keycard_group"`
	Syndicate    null.String `json:"syndicate"`
	Count        int         `json:"count"`
}

// KeycardTransferToXsynHandler transfer keycard to xsyn
func (s *S) KeycardTransferToXsynHandler(req Asset1155LockToSupremacyReq, resp *Asset1155FromSupremacyResp) error {
	b := TransferBucket.Add(fmt.Sprintf("%s_%d", req.OwnerID, req.TokenID), 1)
	if b == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many request made for transfer")
	}

	err := db.UpdateKeycardReductionAmount(req.OwnerID, req.TokenID, req.Amount)
	if err != nil {
		return terror.Error(err, "Failed to update amount")
	}
	// TODO: store transfer event ID

	return nil
}
