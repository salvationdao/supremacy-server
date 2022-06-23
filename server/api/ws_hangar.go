package api

import (
	"context"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"server/db"
	"server/db/boiler"
)

type HangarController struct {
	API *API
}

func NewHangarController(api *API) *HangarController {
	hc := &HangarController{
		API: api,
	}

	api.SecureUserCommand(HubKeyGetHangarItems, hc.GetUserHangarItems)

	return hc
}

type GetHangarItemResponse struct {
	Faction null.String    `json:"faction"`
	Silos   []*db.SiloType `json:"silos"`
}

const HubKeyGetHangarItems = "GET:HANGAR:ITEMS"

func (hc *HangarController) GetUserHangarItems(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user not enlisted"), "Please enlist to view hangar items")
	}

	mechItems, err := db.GetUserMechHangarItems(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get users mech hangar details")
	}

	mysteryCrateItems, err := db.GetUserMysteryCrateHangarItems(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get users mystery crate hangar details")
	}

	hangarResp := &GetHangarItemResponse{
		Faction: user.FactionID,
		Silos:   make([]*db.SiloType, 0),
	}

	hangarResp.Silos = append(hangarResp.Silos, mechItems...)
	hangarResp.Silos = append(hangarResp.Silos, mysteryCrateItems...)

	reply(hangarResp)

	return nil
}
