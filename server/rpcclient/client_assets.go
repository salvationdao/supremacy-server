package rpcclient

import (
	"server/gamelog"
	"strings"
)

type AssetOnChainStatusReq struct {
	AssetHash string `json:"asset_hash"`
}

type AssetOnChainStatusResp struct {
	OnChainStatus string `json:"on_chain_status"`
}

// AssetOnChainStatus return an assets on chain status
func (pp *PassportXrpcClient) AssetOnChainStatus(assetHash string) string {
	resp := &AssetOnChainStatusResp{}
	err := pp.XrpcClient.Call("S.AssetOnChainStatusHandler", AssetOnChainStatusReq{assetHash}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("assetHash", assetHash).Str("method", "AssetOnChainStatusHandler").Msg("rpc error")
		return ""
	}

	return resp.OnChainStatus
}

type AssetsOnChainStatusReq struct {
	AssetHashes []string `json:"asset_hashes"`
}

type AssetsOnChainStatusResp struct {
	OnChainStatuses map[string]string `json:"on_chain_statuses"`
}

// AssetsOnChainStatus return a map of assets on chain statuses map[assetHash]onChainStatus
func (pp *PassportXrpcClient) AssetsOnChainStatus(assetHashes []string) map[string]string {
	resp := &AssetsOnChainStatusResp{}
	err := pp.XrpcClient.Call("S.AssetsOnChainStatusHandler", AssetsOnChainStatusReq{assetHashes}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("assetHashes", strings.Join(assetHashes, ", ")).Str("method", "AssetsOnChainStatusHandler").Msg("rpc error")
		return nil
	}

	return resp.OnChainStatuses
}
