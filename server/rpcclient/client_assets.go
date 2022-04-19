package rpcclient

import (
	"server"
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
