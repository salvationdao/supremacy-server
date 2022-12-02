package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
		return terror.Error(err, "Invalid request received")
	}

	req.Payload.Sort.Column = boiler.BattleReplayColumns.CreatedAt
	req.Payload.Sort.Table = boiler.TableNames.BattleReplays

	if req.Payload.Page < 1 {
		return terror.Error(fmt.Errorf("less than one page"), "Page should start from 1.")
	}

	limit := req.Payload.PageSize
	offset := (req.Payload.Page - 1) * req.Payload.PageSize

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
		BattleNumber int `json:"battle_number"`
		ArenaGID     int `json:"arena_gid"`
	} `json:"payload"`
}

type BattleReplayDetailsResponse struct {
	BattleReplay *server.BattleReplay `json:"battle_replay"`
	Mechs        []*server.Mech       `json:"mechs"`
}

const HubKeyGetReplayDetails = "GET:REPLAY:DETAILS"

func (rc *ReplayController) GetBattleReplayDetails(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleReplayDetailsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	battleReplayResp := &BattleReplayDetailsResponse{}

	battleReplay, err := boiler.BattleReplays(
		boiler.BattleReplayWhere.IsCompleteBattle.EQ(true),
		qm.Where(
			fmt.Sprintf(
				"EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s = ? )",
				boiler.TableNames.Battles,
				qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.ID),
				qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.BattleID),
				qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.BattleNumber),
			),
			req.Payload.BattleNumber,
		),
		qm.Load(boiler.BattleReplayRels.Battle),
		qm.Load(qm.Rels(boiler.BattleReplayRels.Battle, boiler.BattleRels.GameMap)),
		qm.Where(
			fmt.Sprintf(
				"EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s = ? )",
				boiler.TableNames.BattleArena,
				qm.Rels(boiler.TableNames.BattleArena, boiler.BattleArenaColumns.ID),
				qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.ArenaID),
				qm.Rels(boiler.TableNames.BattleArena, boiler.BattleArenaColumns.Gid),
			),
			req.Payload.ArenaGID,
		),
		qm.Load(boiler.BattleReplayRels.Arena),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find battle replay")
	}
	battleReplayResp.BattleReplay = server.BattleReplayFromBoilerWithEvent(battleReplay)

	battleMechs, err := boiler.BattleMechs(boiler.BattleMechWhere.BattleID.EQ(battleReplay.BattleID)).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get battle mechs")
	}
	var MechIDs []string

	for _, battleMech := range battleMechs {
		MechIDs = append(MechIDs, battleMech.MechID)
	}

	mechs, err := db.Mechs(MechIDs...)
	if err != nil {
		return terror.Error(err, "Failed to get mech ids")
	}

	battleReplayResp.Mechs = mechs

	reply(battleReplayResp)

	return nil
}
