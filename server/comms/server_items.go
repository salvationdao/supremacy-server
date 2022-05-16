package comms

import (
	"encoding/json"
	"fmt"
	"server/db"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

//type AssetReq struct {
//	AssetID uuid.UUID
//}
//
//type AssetResp struct {
//	Asset *XsynAsset
//}

func (s *S) Asset(req AssetReq, resp *AssetResp) error {
	gamelog.L.Debug().Msg("comms.Mech")
	result, err := db.Mech(req.AssetID)
	if err != nil {
		return terror.Error(err)
	}

	jsn, _ := json.Marshal(result)
	fmt.Println(string(jsn))

	//resp.MechContainer = ServerMechToApiV1(result)
	return nil
}
