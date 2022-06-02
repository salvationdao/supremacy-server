package comms

import (
	"fmt"
	"github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db"
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
	Label        string
	Description  string
	ImageURL     string
	AnimationURL null.String
	KeycardGroup string
	Syndicate    null.String
	Count        int
}

// KeycardTransferToXsynHandler transfer keycard to xsyn
func (s *S) KeycardTransferToXsynHandler(req Asset1155LockToSupremacyReq, resp *Asset1155FromSupremacyResp) error {
	b := TransferBucket.Add(fmt.Sprintf("%s_%d", req.OwnerID, req.TokenID), 1)
	if b == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many request made for transfer")
	}
	asset, err := boiler.PlayerKeycards(
		boiler.PlayerKeycardWhere.PlayerID.EQ(req.OwnerID),
		qm.InnerJoin(
			fmt.Sprintf(`%s ON %s = %s AND %s = $1`,
				boiler.TableNames.BlueprintKeycards,
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
				qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.KeycardTokenID),
			),
			req.TokenID,
		),
		qm.Load(boiler.PlayerKeycardRels.BlueprintKeycard),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to create or get player keycard")
	}

	asset.Count -= req.Amount
	if asset.Count < 0 {
		return terror.Error(err, "Amount less than 0 after transfer")
	}

	_, err = asset.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerKeycardColumns.Count))
	if err != nil {
		return terror.Error(err, "Failed to update amount")
	}

	resp.Label = asset.R.BlueprintKeycard.Label
	resp.Description = asset.R.BlueprintKeycard.Description
	resp.ImageURL = asset.R.BlueprintKeycard.ImageURL
	resp.Syndicate = asset.R.BlueprintKeycard.Syndicate
	resp.KeycardGroup = asset.R.BlueprintKeycard.KeycardGroup
	resp.Count = asset.Count
	// TODO: store transfer event ID

	return nil
}
