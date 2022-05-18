package api

import (
	"context"
	"encoding/json"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamelog"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
)

type PlayerAssetsControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerAssetsController(api *API) *PlayerAssetsControllerWS {
	pac := &PlayerAssetsControllerWS{
		API: api,
	}

	api.SecureUserCommand(HubKeyPlayerAssetMechList, pac.PlayerAssetMechListHandler)

	return pac
}

const HubKeyPlayerAssetMechList = "PLAYER:ASSET:MECH:LIST"

type PlayerAssetMechListRequest struct {
	Payload struct {
		Search   string                `json:"search"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Sort     *db.ListSortRequest   `json:"sort"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type PlayerAssetMechListResp struct {
	Total int64          `json:"total"`
	Mechs []*server.Mech `json:"mechs"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetMechListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetMechListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	total, mechs, err := db.MechList(&db.MechListOpts{
		Search:   req.Payload.Search,
		Filter:   req.Payload.Filter,
		Sort:     req.Payload.Sort,
		PageSize: req.Payload.PageSize,
		Page:     req.Payload.Page,
		OwnerID:  user.ID,
	})
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your War Machine assets, please try again or contact support.")
	}

	reply(&PlayerAssetMechListResp{
		Total: total,
		Mechs: mechs,
	})
	return nil
}
