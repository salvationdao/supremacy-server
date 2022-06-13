package xsyn_rpcclient

import (
	"server/gamelog"
	"time"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-software/terror/v2"
)

type AssetTransferOwnershipResp struct {
	TransferEventID int64 `json:"transfer_event_id"`
}

type AssetTransferOwnershipReq struct {
	ApiKey               string      `json:"api_key,omitempty"`
	FromOwnerID          string      `json:"from_owner_id,omitempty"`
	ToOwnerID            string      `json:"to_owner_id,omitempty"`
	Hash                 string      `json:"hash,omitempty"`
	RelatedTransactionID null.String `json:"related_transaction_id"`
}

// TransferAsset transfers an assets' ownership on xsyn
func (pp *XsynXrpcClient) TransferAsset(
	fromOwnerID,
	toOwnerID,
	hash string,
	relatedTransactionID null.String,
	updateLatestHandledTransferEvent func(rpcClient *XsynXrpcClient, eventID int64),
) error {
	resp := &AssetTransferOwnershipResp{}
	err := pp.XrpcClient.Call("S.AssetTransferOwnershipHandler", AssetTransferOwnershipReq{
		ApiKey:               pp.ApiKey,
		FromOwnerID:          fromOwnerID,
		ToOwnerID:            toOwnerID,
		Hash:                 hash,
		RelatedTransactionID: relatedTransactionID,
	}, resp)

	if err != nil {
		gamelog.L.Err(err).Str("method", "AssetTransferOwnershipHandler").
			Str("FromOwnerID", fromOwnerID).
			Str("ToOwnerID", toOwnerID).
			Str("Hash", hash).
			Str("RelatedTransactionID", relatedTransactionID.String).
			Msg("rpc error")
		return terror.Error(err, "Failed to transfer asset on xsyn")
	}
	if updateLatestHandledTransferEvent != nil {
		updateLatestHandledTransferEvent(pp, resp.TransferEventID)
	}
	return nil
}

type TransferEvent struct {
	TransferEventID int64       `json:"transfer_event_id"`
	AssetHash       string      `json:"asset_hash,omitempty"`
	FromUserID      string      `json:"from_user_id,omitempty"`
	ToUserID        string      `json:"to_user_id,omitempty"`
	TransferredAt   time.Time   `json:"transferred_at"`
	TransferTXID    null.String `json:"transfer_tx_id"`
	OwnedService    null.String `json:"owned_service"`
}

type GetAssetTransferEventsResp struct {
	TransferEvents []*TransferEvent `json:"transfer_events"`
}

type GetAssetTransferEventsReq struct {
	ApiKey      string `json:"api_key"`
	FromEventID int64  `json:"from_event_id"`
}

// GetTransferEvents requests all the asset transfer events on xsyn
func (pp *XsynXrpcClient) GetTransferEvents(fromEventID int64) ([]*TransferEvent, error) {
	resp := &GetAssetTransferEventsResp{}
	err := pp.XrpcClient.Call("S.GetAssetTransferEventsHandler", GetAssetTransferEventsReq{
		ApiKey:      pp.ApiKey,
		FromEventID: fromEventID,
	}, resp)

	if err != nil {
		gamelog.L.Err(err).Str("method", "GetAssetTransferEventsHandler").Int64("fromEventID", fromEventID).Msg("rpc error")
		return nil, terror.Error(err, "Failed to get transfer events from xsyn")
	}

	return resp.TransferEvents, nil
}
