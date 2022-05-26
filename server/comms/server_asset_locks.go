package comms

type AssetLockToServiceResp struct {
}

type AssetLockToServiceReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetLockHandler request a lock of an asset
func (s *S) AssetLockHandler(req AssetLockToServiceReq, resp *AssetLockToServiceResp) error {

	return nil
}

type AssetUnlockToServiceResp struct {
}

type AssetUnlockToServiceReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetUnlockHandler request a unlock of an asset
func (s *S) AssetUnlockHandler(req AssetUnlockToServiceReq, resp *AssetUnlockToServiceResp) error {

	return nil
}
