package comms

import (
	"fmt"
	"github.com/gofrs/uuid"
	"server"
	"server/asset"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"
	"strings"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-software/terror/v2"
)

func (s *S) AssetHandler(req rpctypes.AssetReq, resp *rpctypes.AssetResp) error {
	gamelog.L.Debug().Msg("comms.Asset")

	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.Hash.EQ(req.AssetHash),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("AssetHash", req.AssetHash).Msg(" failed to get collection item in Asset rpc call ")
		return terror.Error(err)
	}

	switch ci.ItemType {
	case boiler.ItemTypeMysteryCrate:
		idAsUUID, err := uuid.FromString(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get mystery crate in Asset rpc call ")
			return terror.Error(err)
		}

		obj, err := db.PlayerMysteryCrate(idAsUUID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get mystery crate in Asset rpc call ")
			return terror.Error(err)
		}

		// oof we forgot to store the original faction id on the crate so we need to do a string check...
		factionName := ""
		if strings.Contains(obj.Label, "Red Mountain") {
			factionName = " Red Mountain Offworld Mining Corporation"
		} else if strings.Contains(obj.Label, "Boston") {
			factionName = "Boston Cybernetics"
		} else if strings.Contains(obj.Label, "Zaibatsu") {
			factionName = "Zaibatsu Heavy Industries"
		}

		resp.Asset = rpctypes.ServerMysteryCrateToXsynAsset(obj, factionName)
	case boiler.ItemTypeUtility:
		obj, err := db.Utility(gamedb.StdConn, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Utility in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerUtilitiesToXsynAsset([]*server.Utility{obj})[0]
	case boiler.ItemTypeWeapon:
		obj, err := db.Weapon(gamedb.StdConn, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Weapon in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerWeaponsToXsynAsset([]*server.Weapon{obj})[0]
	case boiler.ItemTypeWeaponSkin:
		obj, err := db.WeaponSkin(gamedb.StdConn, ci.ItemID, nil)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Weapon skin in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerWeaServerWeaponSkinsToXsynAsset(gamedb.StdConn, []*server.WeaponSkin{obj})[0]
	case boiler.ItemTypeMech:
		obj, err := db.Mech(gamedb.StdConn, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Mech in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerMechsToXsynAsset([]*server.Mech{obj})[0]
	case boiler.ItemTypeMechSkin:
		obj, err := db.MechSkin(gamedb.StdConn, ci.ItemID, nil)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get MechSkin in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerMechSkinsToXsynAsset(gamedb.StdConn, []*server.MechSkin{obj})[0]
	case boiler.ItemTypeMechAnimation:
		obj, err := db.MechAnimation(gamedb.StdConn, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get MechAnimation in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerMechAnimationsToXsynAsset([]*server.MechAnimation{obj})[0]
	case boiler.ItemTypePowerCore:
		obj, err := db.PowerCore(gamedb.StdConn, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get PowerCore in Asset rpc call ")
			return terror.Error(err)
		}
		resp.Asset = rpctypes.ServerPowerCoresToXsynAsset([]*server.PowerCore{obj})[0]
	default:
		err := fmt.Errorf("invalid type")
		gamelog.L.Error().Err(err).Interface("ci", ci).Msg("invalid item type in Asset rpc call ")
		return terror.Error(err)
	}

	return nil
}

type NFT1155DetailsReq struct {
	TokenID        int    `json:"token_id"`
	CollectionSlug string `json:"collection_slug"`
}

type NFT1155DetailsResp struct {
	Label        string      `json:"label"`
	Description  string      `json:"description"`
	ImageURL     string      `json:"image_url"`
	AnimationUrl null.String `json:"animation_url"`
	Group        string      `json:"group"`
	Syndicate    null.String `json:"syndicate"`
}

func (s *S) Get1155Details(req *NFT1155DetailsReq, resp *NFT1155DetailsResp) error {
	asset, err := boiler.BlueprintKeycards(
		boiler.BlueprintKeycardWhere.KeycardTokenID.EQ(req.TokenID),
		boiler.BlueprintKeycardWhere.Collection.EQ(req.CollectionSlug),
	).One(gamedb.StdConn)
	if err != nil {
		return err
	}

	resp.Syndicate = asset.Syndicate
	resp.Label = asset.Label
	resp.Description = asset.Description
	resp.ImageURL = asset.ImageURL
	resp.AnimationUrl = asset.AnimationURL
	resp.Group = asset.KeycardGroup

	return nil
}

type AssetTransferReq struct {
	TransferEvent *xsyn_rpcclient.TransferEvent `json:"transfer_event"`
}

type AssetTransferResp struct {
	OtherTransferredAssetHashes []string `json:"other_transferred_asset_hashes"`
}

func (s *S) AssetTransferHandler(req *AssetTransferReq, resp *AssetTransferResp) error {
	resp.OtherTransferredAssetHashes = asset.HandleTransferEvent(s.passportRPC, req.TransferEvent, false)
	return nil
}
