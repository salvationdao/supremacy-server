package asset

import (
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
)

func SyncAssetOwners(rpcClient *xsyn_rpcclient.XsynXrpcClient) {
	lastTransferEvent := db.GetIntWithDefault(db.KeyLastTransferEventID, 0)

	transferEvents, err := rpcClient.GetTransferEvents(int64(lastTransferEvent))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get transfer events from xsyn")
	} else {
		sort.Slice(transferEvents, func(i, k int) bool {
			return transferEvents[i].TransferEventID < transferEvents[k].TransferEventID
		})
		for _, te := range transferEvents {
			exists, err := boiler.Players(boiler.PlayerWhere.ID.EQ(te.ToUserID)).Exists(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("transfer event", te).Int64("transfer event id", te.TransferEventID).Msg("failed to check if user exists in transfer event")
				break
			}

			if !exists {
				userUUID := server.UserID(uuid.Must(uuid.FromString(te.ToUserID)))
				user, err := rpcClient.UserGet(userUUID)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("transfer event", te).Int64("transfer event id", te.TransferEventID).Msg("failed to get new user in transfer event")
					break
				}
				_, err = db.PlayerRegister(uuid.UUID(userUUID), user.Username, uuid.FromStringOrNil(user.FactionID.String), common.HexToAddress(user.PublicAddress.String))
				if err != nil {
					gamelog.L.Error().Err(err).Interface("transfer event", te).Int64("transfer event id", te.TransferEventID).Msg("failed to get register new user in transfer event")
					break
				}
			}

			_, err = boiler.CollectionItems(
				boiler.CollectionItemWhere.Hash.EQ(te.AssetHast),
			).UpdateAll(gamedb.StdConn, boiler.M{
				"owner_id": te.ToUserID,
			})
			if err != nil {
				gamelog.L.Error().Err(err).Interface("transfer event", te).Int64("transfer event id", te.TransferEventID).Msg("failed to transfer collection item")
				break
			}
			db.PutInt(db.KeyLastTransferEventID, int(te.TransferEventID))
		}
	}
}

func RegisterAllNewAssets(pp *xsyn_rpcclient.XsynXrpcClient) {
	updatedMechs := db.GetBoolWithDefault("INSERTED_NEW_ASSETS_MECHS", false)
	if !updatedMechs {
		var mechIDs []string

		mechCollections, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech)).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to get mech collection items for RegisterAllNewAssets")
			return
		}
		for _, m := range mechCollections {
			if m.OwnerID == "2fa1a63e-a4fa-4618-921f-4b4d28132069" {
				continue
			}
			mechIDs = append(mechIDs, m.ItemID)
		}

		mechs, err := db.Mechs(mechIDs...)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to get mechs for RegisterAllNewAssets")
			return
		}

		var mechsToInsert []*server.Mech

		// go through each mech and set if genesis or limited
		for _, m := range mechs {
			if m.OwnerID == "2fa1a63e-a4fa-4618-921f-4b4d28132069" && m.GenesisTokenID.Int64 == 356 {
				continue
			}

			if m.GenesisTokenID.Valid {
				m.TokenID = m.GenesisTokenID.Int64
				m.CollectionSlug = "supremacy-genesis"
			} else if m.LimitedReleaseTokenID.Valid {
				m.TokenID = m.LimitedReleaseTokenID.Int64
				m.CollectionSlug = "supremacy-limited-release"
			}

			mechsToInsert = append(mechsToInsert, m)
		}

		err = pp.AssetsRegister(rpctypes.ServerMechsToXsynAsset(mechsToInsert)) // register new mechs
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue inserting new mechs to xsyn for RegisterAllNewAssets")
			return
		}
		gamelog.L.Info().Msg("Successfully inserted new asset mechs")
		db.PutBool("INSERTED_NEW_ASSETS_MECHS", true)
	}
	return
}
