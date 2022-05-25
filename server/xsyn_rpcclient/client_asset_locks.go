package xsyn_rpcclient

import (
	"server"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

// AssetLock requests an asset to be locked to supremacy
func (pp *XsynXrpcClient) AssetLock(assetToLock *server.CollectionDetails) error {
	resp := &AssetLockToServiceResp{}
	err := pp.XrpcClient.Call("S.AssetLockToServiceHandler", AssetLockToServiceReq{
		ApiKey:         pp.ApiKey,
		CollectionSlug: assetToLock.CollectionSlug,
		TokenID:        assetToLock.TokenID,
		OwnerID:        assetToLock.OwnerID,
		Hash:           assetToLock.Hash,
	}, resp)

	if err != nil {
		gamelog.L.Err(err).Str("method", "AssetLockToServiceHandler").Interface("assetToLock", assetToLock).Msg("rpc error")
		return terror.Error(err, "Failed to lock asset to supremacy")
	}
	return nil
}

// AssetUnlock request a service unlock of an asset
func (pp *XsynXrpcClient) AssetUnlock(assetToUnlock *server.CollectionDetails) error {
	resp := &AssetUnlockToServiceResp{}
	err := pp.XrpcClient.Call("S.AssetLockToServiceHandler", AssetUnlockToServiceReq{
		ApiKey:         pp.ApiKey,
		CollectionSlug: assetToUnlock.CollectionSlug,
		TokenID:        assetToUnlock.TokenID,
		OwnerID:        assetToUnlock.OwnerID,
		Hash:           assetToUnlock.Hash,
	}, resp)

	if err != nil {
		gamelog.L.Err(err).Str("method", "AssetLockToServiceHandler").Interface("assetToUnlock", assetToUnlock).Msg("rpc error")
		return terror.Error(err, "Failed to unlock asset from supremacy")
	}

	return nil
}
