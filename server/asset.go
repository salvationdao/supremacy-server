package server

import "github.com/volatiletech/sqlboiler/v4/types"

// Asset is a generic Asset struct, used for xsyn
type Asset struct {
	ID             string     `json:"id"`
	CollectionSlug string     `json:"collection_id"`
	TokenID        int64      `json:"external_token_id"`
	Tier           string     `json:"tier"`
	Hash           string     `json:"hash"`
	OwnerID        string     `json:"owner_id"`
	Data           types.JSON `json:"data"`
	OnChainStatus  string     `json:"on_chain_status"`
}
