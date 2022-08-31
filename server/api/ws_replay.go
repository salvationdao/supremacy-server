package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ReplayController struct {
	API *API
}

func NewReplayController(api *API) *ReplayController {
	rc := &ReplayController{
		API: api,
	}

	api.Command(HubKeyGetAllReplays, rc.GetAllBattleReplays)
	api.Command(HubKeyGetReplayDetails, rc.GetBattleReplayDetails)
	return rc
}

type BattleReplayGetRequest struct {
	Payload struct {
		Search   string              `json:"search"`
		Sort     *db.ListSortRequest `json:"sort"`
		Page     int                 `json:"page"`
		PageSize int                 `json:"page_size"`
	} `json:"payload"`
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
		fmt.Println(string(payload))
		return terror.Error(err, "Invalid request received")
	}

	req.Payload.Sort.Column = boiler.BattleReplayColumns.CreatedAt
	req.Payload.Sort.Table = boiler.TableNames.BattleReplays

	limit := req.Payload.PageSize
	offset := req.Payload.Page * req.Payload.PageSize

	brs := []*server.BattleReplay{}

	count, replays, err := db.ReplayList(req.Payload.Search, req.Payload.Sort, limit, offset)
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
	Payload struct {
		ReplayID string `json:"replay_id"`
	} `json:"payload"`
}

const HubKeyGetReplayDetails = "GET:REPLAY:DETAILS"

func (rc *ReplayController) GetBattleReplayDetails(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleReplayDetailsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.ReplayID == "" {
		return terror.Error(err, "Invalid replay id")
	}

	replay, err := boiler.BattleReplays(
		boiler.BattleReplayWhere.ID.EQ(req.Payload.ReplayID),
		qm.Load(boiler.BattleReplayRels.Battle),
		qm.Load(qm.Rels(boiler.BattleReplayRels.Battle, boiler.BattleRels.GameMap)),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find battle replay")
	}

	reply(server.BattleReplayFromBoilerWithEvent(replay))

	return nil
}
