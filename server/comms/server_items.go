package comms

import (
	"encoding/json"
	"fmt"
	"github.com/volatiletech/null/v8"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"

	"github.com/ninja-software/terror/v2"
)

func (s *S) Asset(req rpctypes.AssetReq, resp *rpctypes.AssetResp) error {
	gamelog.L.Debug().Msg("comms.Asset")

	ci, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(req.AssetID.String())).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("req.AssetID.String()", req.AssetID.String()).Msg(" failed to get collection item in Asset rpc call ")
		return terror.Error(err)
	}

	var item any

	switch ci.ItemType {
	case boiler.ItemTypeUtility:
		item, err = db.Utility(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Utility in Asset rpc call ")
			return terror.Error(err)
		}
	case boiler.ItemTypeWeapon:
		item, err = db.Weapon(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Weapon in Asset rpc call ")
			return terror.Error(err)
		}
	case boiler.ItemTypeMech:
		item, err = db.Mech(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get Mech in Asset rpc call ")
			return terror.Error(err)
		}
	case boiler.ItemTypeMechSkin:
		item, err = db.MechSkin(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get MechSkin in Asset rpc call ")
			return terror.Error(err)
		}
	case boiler.ItemTypeMechAnimation:
		item, err = db.MechAnimation(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get MechAnimation in Asset rpc call ")
			return terror.Error(err)
		}
	case boiler.ItemTypePowerCore:
		item, err = db.PowerCore(ci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("ci.ItemID", ci.ItemID).Msg(" failed to get PowerCore in Asset rpc call ")
			return terror.Error(err)
		}
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
		OnChainStatus:  ci.OnChainStatus,
		Data:           asJson,
		//Name: //TODO?
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

func (s *S) GenesisOrLimitedMech(req *GenesisOrLimitedMechReq, resp *GenesisOrLimitedMechResp) error {
	gamelog.L.Trace().Msg("comms.GenesisOrLimitedMech")
	var mech *server.Mech

	//if req.TokenID
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println(req.TokenID)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	switch req.CollectionSlug {
	case "supremacy-genesis":
		mechBoiler, err := boiler.Mechs(
			boiler.MechWhere.GenesisTokenID.EQ(null.Int64From(int64(req.TokenID))),
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
		mech = server.MechFromBoiler(mechBoiler, collection)
	case "supremacy-limited-release":
		mechBoiler, err := boiler.Mechs(
			boiler.MechWhere.LimitedReleaseTokenID.EQ(null.Int64From(int64(req.TokenID))),
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
		mech = server.MechFromBoiler(mechBoiler, collection)
	default:
		err := fmt.Errorf("invalid collection slug")
		gamelog.L.Error().Err(err).Str("req.CollectionSlug", req.CollectionSlug).Msg("collection slug is invalid")
		return err
	}

	resp.Asset = rpctypes.ServerMechsToXsynAsset([]*server.Mech{mech})[0]
	return nil
}
