package comms

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

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
	_, err = collectionItem.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to lock asset - AssetUnlockFromSupremacyHandler")
		return err
	}

	// TODO: store transfer event ID

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
