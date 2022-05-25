package xsyn_rpcclient

import (
	"server"
	"server/gamelog"
	"server/rpctypes"
	"strings"

	"github.com/volatiletech/null/v8"
)

type AssetOnChainStatusReq struct {
	AssetID string `json:"asset_ID"`
}

type AssetOnChainStatusResp struct {
	OnChainStatus server.OnChainStatus `json:"on_chain_status"`
}

// AssetOnChainStatus return an assets on chain status
func (pp *XsynXrpcClient) AssetOnChainStatus(assetID string) (server.OnChainStatus, error) {
	resp := &AssetOnChainStatusResp{}
	err := pp.XrpcClient.Call("S.AssetOnChainStatusHandler", AssetOnChainStatusReq{assetID}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("assetID", assetID).Str("method", "AssetOnChainStatusHandler").Msg("rpc error")
		return "", err
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
func (pp *XsynXrpcClient) AssetsOnChainStatus(assetIDs []string) (map[string]server.OnChainStatus, error) {
	resp := &AssetsOnChainStatusResp{}
	err := pp.XrpcClient.Call("S.AssetsOnChainStatusHandler", AssetsOnChainStatusReq{assetIDs}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("assetIDes", strings.Join(assetIDs, ", ")).Str("method", "AssetsOnChainStatusHandler").Msg("rpc error")
		return nil, err
	}

	return resp.OnChainStatuses, nil
}

type RegisterAssetReq struct {
	Asset *rpctypes.XsynAsset `json:"asset"`
}

type RegisterAssetResp struct {
	Success bool `json:"success"`
}

// AssetRegister registers a item on xsyn
func (pp *XsynXrpcClient) AssetRegister(ass *rpctypes.XsynAsset) error {
	resp := &RegisterAssetResp{}
	err := pp.XrpcClient.Call("S.AssetRegisterHandler", RegisterAssetReq{
		ass,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Interface("asset", ass).Msg("rpc error - S.AssetRegisterHandler")
		return err
	}

	return nil
}

type RegisterAssetsReq struct {
	Assets []*rpctypes.XsynAsset `json:"assets"`
}

type RegisterAssetsResp struct {
	Success bool `json:"success"`
}

// AssetsRegister registers items on xsyn
func (pp *XsynXrpcClient) AssetsRegister(ass []*rpctypes.XsynAsset) error {
	resp := &RegisterAssetsResp{}
	err := pp.XrpcClient.Call("S.AssetsRegisterHandler", RegisterAssetsReq{
		ass,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Interface("asset", ass).Msg("rpc error - S.AssetRegisterHandler")
		return err
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
func (pp *XsynXrpcClient) UpdateStoreItemIDs(assetsToUpdate []*TemplatesToUpdate) error {
	resp := &UpdateStoreItemIDsResp{}
	err := pp.XrpcClient.Call("S.UpdateStoreItemIDsHandler", UpdateStoreItemIDsReq{
		StoreItemsToUpdate: assetsToUpdate,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Interface("assetsToUpdate", assetsToUpdate).Str("method", "UpdateStoreItemIDsHandler").Msg("rpc error")
		return err
	}

	return nil
}

type UpdateUser1155AssetReq struct {
	PublicAddress string               `json:"public_address"`
	AssetData     []Supremacy1155Asset `json:"asset_data"`
}

type Supremacy1155Asset struct {
	BlueprintID    string
	Label          string                      `json:"label"`
	Description    string                      `json:"description"`
	CollectionSlug string                      `json:"collection_slug"`
	TokenID        int                         `json:"token_id"`
	Count          int                         `json:"count"`
	ImageURL       string                      `json:"image_url"`
	AnimationURL   string                      `json:"animation_url"`
	KeycardGroup   string                      `json:"keycard_group"`
	Attributes     []SupremacyKeycardAttribute `json:"attributes"`
}

type SupremacyKeycardAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value,omitempty"`
}

type UpdateUser1155AssetResp struct {
	UserID        string      `json:"user_id"`
	Username      string      `json:"username"`
	FactionID     null.String `json:"faction_id"`
	PublicAddress null.String `json:"public_address"`
}

func (pp *XsynXrpcClient) UpdateKeycardItem(keycardUpdate UpdateUser1155AssetReq) (*UpdateUser1155AssetResp, error) {
	resp := &UpdateUser1155AssetResp{}
	err := pp.XrpcClient.Call("S.InsertUser1155Asset", keycardUpdate, resp)
	if err != nil {
		gamelog.L.Err(err).Str("user_address", keycardUpdate.PublicAddress).Msg("rpc error")
		return nil, err
	}

	return resp, nil
}
