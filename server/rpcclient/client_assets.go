package rpcclient

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strings"

	"github.com/ninja-software/terror/v2"
)

type AssetOnChainStatusReq struct {
	AssetID string `json:"asset_ID"`
}

type AssetOnChainStatusResp struct {
	OnChainStatus server.OnChainStatus `json:"on_chain_status"`
}

// AssetOnChainStatus return an assets on chain status
func (pp *PassportXrpcClient) AssetOnChainStatus(assetID string) (server.OnChainStatus, error) {
	resp := &AssetOnChainStatusResp{}
	err := pp.XrpcClient.Call("S.AssetOnChainStatusHandler", AssetOnChainStatusReq{assetID}, resp)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			// if we get no rows error, get the mechs old ID, get the status of that, and then tell xsyn to update that
			oldMech, err := boiler.MechsOlds(boiler.MechsOldWhere.ChassisID.EQ(assetID)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Err(err).Str("assetID", assetID).Str("method", "AssetOnChainStatusHandler").Msg("rpc error - boiler.MechsOlds")
				return "", terror.Error(err)
			}
			_, err = pp.UpdateAssetID(oldMech.ID, assetID)
			if err != nil {
				return "", terror.Error(err)
			}

			// now retry!
			err = pp.XrpcClient.Call("S.AssetOnChainStatusHandler", AssetOnChainStatusReq{AssetID: assetID}, resp)
			if err != nil {
				gamelog.L.Err(err).Str("assetID", assetID).Str("method", "AssetOnChainStatusHandler").Msg("rpc error")
				return "", terror.Error(err)
			}
			return resp.OnChainStatus, nil
		}
		gamelog.L.Err(err).Str("assetID", assetID).Str("method", "AssetOnChainStatusHandler").Msg("rpc error")
		return "", terror.Error(err)
	}
	return resp.OnChainStatus, nil
}

type AssetsOnChainStatusReq struct {
	AssetIDs []string `json:"asset_ids"`
}

type AssetsOnChainStatusResp struct {
	OnChainStatuses map[string]server.OnChainStatus `json:"on_chain_statuses"`
}

// AssetsOnChainStatus return a map of assets on chain statuses map[assetID]onChainStatus
func (pp *PassportXrpcClient) AssetsOnChainStatus(assetIDs []string) (map[string]server.OnChainStatus, error) {
	resp := &AssetsOnChainStatusResp{}
	err := pp.XrpcClient.Call("S.AssetsOnChainStatusHandler", AssetsOnChainStatusReq{assetIDs}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("assetIDes", strings.Join(assetIDs, ", ")).Str("method", "AssetsOnChainStatusHandler").Msg("rpc error")
		return nil, terror.Error(err)
	}

	return resp.OnChainStatuses, nil
}

type UpdateAssetIDReq struct {
	AssetID    string `json:"asset_ID" db:"asset_ID"`
	OldAssetID string `json:"old_asset_ID" db:"old_asset_ID"`
}

type UpdateAssetIDResp struct {
	AssetID string `json:"asset_ID"`
}

// UpdateAssetID updates a purchased_items id on passport server
func (pp *PassportXrpcClient) UpdateAssetID(oldAssetID, assetID string) (string, error) {
	resp := &UpdateAssetIDResp{}
	err := pp.XrpcClient.Call("S.UpdateAssetIDHandler", UpdateAssetIDReq{
		AssetID:    assetID,
		OldAssetID: oldAssetID,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("assetID", assetID).Str("oldAssetID", oldAssetID).Str("method", "UpdateAssetIDHandler").Msg("rpc error")
		return "", terror.Error(err)
	}

	return resp.AssetID, nil
}

type UpdateAssetsIDReq struct {
	AssetsToUpdate []*UpdateAssetIDReq `json:"assets_to_update"`
}

type UpdateAssetsIDResp struct {
	Success bool `json:"success"`
}

// UpdateAssetsID updates the purchased_items id on passport server
func (pp *PassportXrpcClient) UpdateAssetsID(assetsToUpdate []*UpdateAssetIDReq) error {
	resp := &UpdateAssetsIDResp{}
	err := pp.XrpcClient.Call("S.UpdateAssetsIDHandler", UpdateAssetsIDReq{
		AssetsToUpdate: assetsToUpdate,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Interface("assetsToUpdate", assetsToUpdate).Str("method", "UpdateAssetsIDHandler").Msg("rpc error")
		return terror.Error(err)
	}

	return nil
}

type UpdateStoreItemIDsReq struct {
	StoreItemsToUpdate []*TemplatesToUpdate `json:"store_items_to_update"`
}

type TemplatesToUpdate struct {
	OldTemplateID string `json:"old_template_id"`
	NewTemplateID string `json:"new_template_id"`
}

type UpdateStoreItemIDsResp struct {
	Success bool `json:"success"`
}

// UpdateStoreItemIDs updates the store item ids on passport server
func (pp *PassportXrpcClient) UpdateStoreItemIDs(assetsToUpdate []*TemplatesToUpdate) error {
	resp := &UpdateStoreItemIDsResp{}
	err := pp.XrpcClient.Call("S.UpdateStoreItemIDsHandler", UpdateStoreItemIDsReq{
		StoreItemsToUpdate: assetsToUpdate,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Interface("assetsToUpdate", assetsToUpdate).Str("method", "UpdateStoreItemIDsHandler").Msg("rpc error")
		return terror.Error(err)
	}

	return nil
}
