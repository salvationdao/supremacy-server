package api

import (
	"context"
	"encoding/json"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
)

type AdminController struct {
	API *API
}

func NewAdminController(api *API) *AdminController {
	adminHub := &AdminController{
		API: api,
	}

	api.SecureAdminCommand(HubKeyAdminFiatProductList, adminHub.FiatProductList)

	return adminHub
}

type AdminFiatProductListRequest struct {
	Payload struct {
		ProductType string `json:"product_type"`
		Search      string `json:"search"`
		PageSize    int    `json:"page_size"`
		Page        int    `json:"page"`
	} `json:"payload"`
}

type AdminFiatProductListResponse struct {
	Total   int64                 `json:"total"`
	Records []*server.FiatProduct `json:"records"`
}

const HubKeyAdminFiatProductList = "ADMIN:FIAT:PRODUCT:LIST"

func (ac *AdminController) FiatProductList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get packages, please try again."

	req := &FiatProductListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, storePackages, err := db.FiatProducts(gamedb.StdConn, nil, req.Payload.ProductType, req.Payload.Search, offset, req.Payload.PageSize)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := &FiatProductListResponse{
		Total:   total,
		Records: storePackages,
	}
	reply(resp)

	return nil
}
