package marketplace

import (
	"server/asset"
	"server/db"
	"server/gamelog"
	"server/xsyn_rpcclient"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type TransferAssetRollbackFunc func()

// TransferAssets transfers an item sale's collection item to new owner.
func TransferAssets(
	conn boil.Executor,
	passport *xsyn_rpcclient.XsynXrpcClient,
	fromUserID string,
	toUserID string,
	relatedTransactionID string,
	hash string,
	itemSaleID string,
) (TransferAssetRollbackFunc, error) {
	otherAssets, err := db.MarketplaceGetOtherAssets(conn, itemSaleID)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = passport.TransferAsset(
		fromUserID,
		toUserID,
		hash,
		null.StringFrom(relatedTransactionID),
		func(rpcClient *xsyn_rpcclient.XsynXrpcClient, eventID int64) {
			asset.UpdateLatestHandledTransferEvent(rpcClient, eventID)
		},
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	rpcAssetTransferRollbackFuncs := []TransferAssetRollbackFunc{
		func() {
			err := passport.TransferAsset(
				toUserID,
				fromUserID,
				hash,
				null.StringFrom(relatedTransactionID),
				func(rpcClient *xsyn_rpcclient.XsynXrpcClient, eventID int64) {
					asset.UpdateLatestHandledTransferEvent(rpcClient, eventID)
				},
			)
			if err != nil {
				gamelog.L.Error().
					Str("from_user_id", fromUserID).
					Str("to_user_id", toUserID).
					Str("item_sale_id", itemSaleID).
					Err(err).
					Msg("Failed to start purchase sale item rpc TransferAsset rollback.")
			}
		},
	}

	for _, attachedHash := range otherAssets {
		err = passport.TransferAsset(
			fromUserID,
			toUserID,
			attachedHash,
			null.StringFrom(relatedTransactionID),
			func(rpcClient *xsyn_rpcclient.XsynXrpcClient, eventID int64) {
				asset.UpdateLatestHandledTransferEvent(rpcClient, eventID)
			},
		)
		if err != nil {
			return nil, terror.Error(err)
		}
		rpcAssetTransferRollbackFuncs = append(rpcAssetTransferRollbackFuncs, func() {
			err := passport.TransferAsset(
				toUserID,
				fromUserID,
				attachedHash,
				null.StringFrom(relatedTransactionID),
				func(rpcClient *xsyn_rpcclient.XsynXrpcClient, eventID int64) {
					asset.UpdateLatestHandledTransferEvent(rpcClient, eventID)
				},
			)
			if err != nil {
				gamelog.L.Error().
					Str("from_user_id", fromUserID).
					Str("to_user_id", toUserID).
					Str("item_sale_id", itemSaleID).
					Err(err).
					Msg("Failed to start purchase sale item rpc TransferAsset rollback (attachment).")
			}
		})
	}

	rollbackFunc := func() {
		for _, fn := range rpcAssetTransferRollbackFuncs {
			fn()
		}
	}

	return rollbackFunc, nil
}
