package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
)

type ReplayController struct {
	API *API
}

func NewReplayController(api *API) *ReplayController {
	rc := &ReplayController{
		API: api,
	}

	api.Command(HubKeyGetAllReplays, rc.GetAllBattleReplays)
	api.Command(HubKeyGetAllReplays, rc.GetBattleReplayDetails)
	return rc
}

type BattleReplayGetRequest struct {
	Search string              `json:"search"`
	Sort   *db.ListSortRequest `json:"sort"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type BattleReplayResponse struct {
	Total        int64                  `json:"total"`
	BattleReplay []*server.BattleReplay `json:"battle_replays"`
}

const HubKeyGetAllReplays = "GET:BATTLE:REPLAYS"

func (rc *ReplayController) GetAllBattleReplays(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleReplayGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Limit <= 0 || req.Offset < 0 {
		return terror.Error(fmt.Errorf("invalid limit and offset"), "Invalid limit and offset")
	}

	brs := []*server.BattleReplay{}

	count, replays, err := db.ReplayList(req.Search, req.Sort, req.Limit, req.Offset)
	if err != nil {
		return err
	}

	if replays != nil {
		brs = replays
	}

	reply(&BattleReplayResponse{
		Total:        count,
		BattleReplay: brs,
	})

	return nil
}

type BattleReplayDetailsRequest struct {
	ReplayID string `json:"replay_id"`
}

const HubKeyGetReplayDetails = "GET:REPLAY:DETAILS"

func (rc *ReplayController) GetBattleReplayDetails(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleReplayDetailsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.ReplayID == "" {
		return terror.Error(err, "Invalid replay id")
	}

	replay, err := boiler.FindBattleReplay(gamedb.StdConn, req.ReplayID)
	if err != nil {
		return terror.Error(err, "Failed to find battle replay")
	}

	reply(server.BattleReplayFromBoilerWithEvent(replay))

	return nil
}
