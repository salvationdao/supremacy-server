package comms

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AssetLockResp struct {
}

type AssetLockReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetLockHandler request a lock of an asset
func (s *S) AssetLockHandler(req AssetLockReq, resp *AssetLockResp) error {
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(req.OwnerID),
		boiler.CollectionItemWhere.TokenID.EQ(req.TokenID),
		boiler.CollectionItemWhere.Hash.EQ(req.Hash),
		boiler.CollectionItemWhere.CollectionSlug.EQ(req.CollectionSlug),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetLockHandler")
		return err
	}

	if collectionItem.XsynLocked {
		return nil
	}

	collectionItem.XsynLocked = true
	_, err = collectionItem.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to lock asset - AssetLockHandler")
		return err
	}

	return nil
}

type AssetUnlockResp struct {
}

type AssetUnlockReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetUnlockHandler request an unlock of an asset
func (s *S) AssetUnlockHandler(req AssetUnlockReq, resp *AssetUnlockResp) error {
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(req.OwnerID),
		boiler.CollectionItemWhere.TokenID.EQ(req.TokenID),
		boiler.CollectionItemWhere.Hash.EQ(req.Hash),
		boiler.CollectionItemWhere.CollectionSlug.EQ(req.CollectionSlug),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to find asset - AssetUnlockHandler")
		return err
	}

	if !collectionItem.XsynLocked {
		return nil
	}

	collectionItem.XsynLocked = false
	_, err = collectionItem.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("req", req).Msg("failed to unlock asset - AssetUnlockHandler")
		return err
	}

	return nil
}
