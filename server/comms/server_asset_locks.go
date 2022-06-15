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
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(req.OwnerID),
		boiler.CollectionItemWhere.Hash.EQ(req.Hash),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetUnlockFromSupremacyHandler")
		return err
	}

	if collectionItem.XsynLocked {
		return nil
	}

	collectionItem.XsynLocked = true
	collectionItem.LockedToMarketplace = false
	_, err = collectionItem.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetUnlockFromSupremacyHandler")
		return err
	}

	// TODO: store transfer event ID

	itemUUID, err := uuid.FromString(collectionItem.ItemID)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("convert asset id to uuid - AssetUnlockFromSupremacyHandler")
		return err
	}

	if collectionItem.LockedToMarketplace {
		err = db.MarketplaceSaleArchiveByItemID(gamedb.StdConn, itemUUID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetUnlockFromSupremacyHandler")
			return terror.Error(err, "Failed to unlock asset from supremacy")
		}
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
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(req.OwnerID),
		boiler.CollectionItemWhere.Hash.EQ(req.Hash),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetLockToSupremacyHandler")
		return err
	}

	if !collectionItem.XsynLocked && collectionItem.MarketLocked == req.MarketLocked {
		return nil
	}

	collectionItem.XsynLocked = false
	collectionItem.MarketLocked = req.MarketLocked
	_, err = collectionItem.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetLockToSupremacyHandler")
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

	err := db.UpdateKeycardReductionAmount(req.OwnerID, req.TokenID)
	if err != nil {
		return terror.Error(err, "Failed to update amount")
	}
	// TODO: store transfer event ID

	return nil
}
