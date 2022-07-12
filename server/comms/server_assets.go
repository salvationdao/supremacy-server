package comms

import (
	"encoding/json"
	"fmt"
	"server"
	"server/asset"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-software/terror/v2"
)

func (s *S) AssetHandler(req rpctypes.AssetReq, resp *rpctypes.AssetResp) error {
	gamelog.L.Debug().Msg("comms.Asset")

	ci, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(req.AssetID.String())).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("req.AssetID.String()", req.AssetID.String()).Msg(" failed to get collection item in Asset rpc call ")
		return terror.Error(err)
	}

	var item any
	var name string

	switch ci.ItemType {
	case boiler.ItemTypeUtility:
		obj, err := db.Utility(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Utility in Asset rpc call ")
			return terror.Error(err)
		}
		item = obj
		name = obj.Label
	case boiler.ItemTypeWeapon:
		obj, err := db.Weapon(nil, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Weapon in Asset rpc call ")
			return terror.Error(err)
		}
		item = obj
		name = obj.Label
	case boiler.ItemTypeMech:
		obj, err := db.Mech(gamedb.StdConn, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Mech in Asset rpc call ")
			return terror.Error(err)
		}
		item = obj
		name = obj.Label
	case boiler.ItemTypeMechSkin:
		obj, err := db.MechSkin(nil, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get MechSkin in Asset rpc call ")
			return terror.Error(err)
		}
		item = obj
		name = obj.Label
	case boiler.ItemTypeMechAnimation:
		obj, err := db.MechAnimation(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get MechAnimation in Asset rpc call ")
			return terror.Error(err)
		}
		item = obj
		name = obj.Label
	case boiler.ItemTypePowerCore:
		obj, err := db.PowerCore(nil, ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get PowerCore in Asset rpc call ")
			return terror.Error(err)
		}
		item = obj
		name = obj.Label
	default:
		err := fmt.Errorf("invalid type")
		gamelog.L.Error().Err(err).Interface("ci", ci).Msg("invalid item type in Asset rpc call ")
		return terror.Error(err)
	}

	asJson, err := json.Marshal(item)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("item", item).Msg(" failed to marshall item in Asset rpc call ")
		return terror.Error(err)
	}

	resp.Asset = &rpctypes.XsynAsset{
		ID:             ci.ID,
		CollectionSlug: ci.CollectionSlug,
		TokenID:        ci.TokenID,
		Tier:           ci.Tier,
		Hash:           ci.Hash,
		OwnerID:        ci.OwnerID,
		AssetType:      null.StringFrom(ci.ItemType),
		Data:           asJson,
		Name:           name,
		XsynLocked:     ci.XsynLocked,
	}
	return nil
}

type GenesisOrLimitedMechReq struct {
	CollectionSlug string
	TokenID        int
}

type GenesisOrLimitedMechResp struct {
	Asset *rpctypes.XsynAsset
}

func (s *S) GenesisOrLimitedMechHandler(req *GenesisOrLimitedMechReq, resp *GenesisOrLimitedMechResp) error {
	gamelog.L.Trace().Msg("comms.GenesisOrLimitedMechHandler")
	var mech *server.Mech

	switch req.CollectionSlug {
	case "supremacy-genesis":
		mechBoiler, err := boiler.Mechs(
			boiler.MechWhere.GenesisTokenID.EQ(null.Int64From(int64(req.TokenID))),
			qm.Load(boiler.MechRels.Model),
			qm.Load(qm.Rels(boiler.MechRels.Model, boiler.MechModelRels.DefaultChassisSkin)),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Int("req.TokenID", req.TokenID).Msg("failed to find genesis mech")
			return err
		}

		collection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mechBoiler.ID)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("mechBoiler.ID", mechBoiler.ID).Msg("failed to find collection item")
			return err
		}

		var skinCollection *boiler.CollectionItem
		if mechBoiler.ChassisSkinID.Valid {
			skinCollection, err = boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mechBoiler.ChassisSkinID.String)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("mechBoiler.ChassisSkinID.String", mechBoiler.ChassisSkinID.String).Msg("failed to find skin collection item")
				return err
			}
		}

		mech = server.MechFromBoiler(mechBoiler, collection, skinCollection)
	case "supremacy-limited-release":
		mechBoiler, err := boiler.Mechs(
			boiler.MechWhere.LimitedReleaseTokenID.EQ(null.Int64From(int64(req.TokenID))),
			qm.Load(boiler.MechRels.Model),
			qm.Load(qm.Rels(boiler.MechRels.Model, boiler.MechModelRels.DefaultChassisSkin)),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Int("req.TokenID", req.TokenID).Msg("failed to find limited release mech")
			return err
		}

		collection, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mechBoiler.ID)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("mechBoiler.ID", mechBoiler.ID).Msg("failed to find collection item")
			return err
		}

		var skinCollection *boiler.CollectionItem
		if mechBoiler.ChassisSkinID.Valid {
			skinCollection, err = boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mechBoiler.ChassisSkinID.String)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("mechBoiler.ChassisSkinID.String", mechBoiler.ChassisSkinID.String).Msg("failed to find skin collection item")
				return err
			}
		}

		mech = server.MechFromBoiler(mechBoiler, collection, skinCollection)
	default:
		err := fmt.Errorf("invalid collection slug")
		gamelog.L.Error().Err(err).Str("req.CollectionSlug", req.CollectionSlug).Msg("collection slug is invalid")
		return err
	}

	resp.Asset = rpctypes.ServerMechsToXsynAsset([]*server.Mech{mech})[0]
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
}

func (s *S) AssetTransferHandler(req *AssetTransferReq, resp *AssetTransferResp) error {
	asset.HandleTransferEvent(s.passportRPC, req.TransferEvent, false)
	return nil
}
