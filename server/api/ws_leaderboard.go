package api

import (
	"context"

	"github.com/ninja-syndicate/ws"
)

// func NewLeaderboardController(api *API) {
// api.Command(HubKeyBattleMechHistoryList, api.BattleMechHistoryListHandler)
// 	api.SecureUserCommand(HubKeyPlayerAssetMechList, lc.PlayerAssetMechListHandler)
// 	api.SecureUserCommand(HubKeyPlayerAssetWeaponList, lc.PlayerAssetWeaponListHandler)
// 	api.SecureUserCommand(HubKeyPlayerAssetMysteryCrateList, lc.PlayerAssetMysteryCrateListHandler)
// 	api.SecureUserCommand(HubKeyPlayerAssetMysteryCrateGet, lc.PlayerAssetMysteryCrateGetHandler)
// 	api.SecureUserFactionCommand(HubKeyPlayerAssetMechDetail, lc.PlayerAssetMechDetail)
// 	api.SecureUserFactionCommand(HubKeyPlayerAssetWeaponDetail, lc.PlayerAssetWeaponDetail)
// 	api.SecureUserCommand(HubKeyPlayerAssetKeycardList, lc.PlayerAssetKeycardListHandler)
// 	api.SecureUserCommand(HubKeyPlayerAssetKeycardGet, lc.PlayerAssetKeycardGetHandler)
// 	api.SecureUserCommand(HubKeyPlayerAssetRename, lc.PlayerMechRenameHandler)
// 	api.SecureUserFactionCommand(HubKeyOpenCrate, lc.OpenCrateHandler)

// }


func (api *API) Leaderboard(ctx context.Context, key string ,payload []byte, reply ws.ReplyFunc)error{


	return nil
}
