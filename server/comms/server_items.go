package comms

import (
	"encoding/json"
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

func (s *S) Asset(req AssetReq, resp *AssetResp) error {
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

	resp.Asset = &XsynAsset{
		ID:             ci.ID,
		CollectionSlug: ci.CollectionSlug,
		TokenID:        ci.TokenID,
		Tier:           ci.Tier,
		Hash:           ci.Hash,
		OwnerID:        ci.OwnerID,
		ItemType:       ci.ItemType,
		OnChainStatus:  ci.OnChainStatus,
		Data:           asJson,
	}
	return nil
}
