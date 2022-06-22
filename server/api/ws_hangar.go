package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

type HangarController struct {
	API *API
}

func NewHangarController(api *API) *HangarController {
	hc := &HangarController{
		API: api,
	}

	api.SecureUserCommand(HubKeyGetHangarItems, hc.GetUserHangarItems)
	api.SecureUserFactionCommand(HubKeyOpenCrate, hc.OpenCrateHandler)

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

type OpenCrateRequest struct {
	Payload struct {
		id string `json:"id"`
	} `json:"payload"`
}

type OpenCrateResponse struct {
	Faction null.String    `json:"faction"`
	Silos   []*db.SiloType `json:"silos"`
}

const HubKeyOpenCrate = "CRATE:OPEN"

func (hc *HangarController) OpenCrateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &OpenCrateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.id),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Could not find collection item, try again or contact support.")
	}

	//checks
	if collectionItem.OwnerID != user.ID {
		return terror.Error(fmt.Errorf("user: %s attempted to claim crate: %s belonging to owner: %s", user.ID, req.Payload.id, collectionItem.OwnerID), "This crate does not belong to this user, try again or contact support.")
	}
	if collectionItem.MarketLocked {
		return terror.Error(fmt.Errorf("user: %s attempted to claim crate: %s while market locked", user.ID, req.Payload.id), "This crate is still on Marketplace, try again or contact support.")
	}
	if collectionItem.XsynLocked {
		return terror.Error(fmt.Errorf("user: %s attempted to claim crate: %s while XSYN locked", user.ID, req.Payload.id), "This crate is locked to XSYN, move asset to Supremacy and try again.")
	}

	crate, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.ID.EQ(collectionItem.ItemID),
		boiler.MysteryCrateWhere.FactionID.EQ(factionID),
		boiler.MysteryCrateWhere.LockedUntil.LTE(time.Now()),
		qm.Load(boiler.MysteryCrateRels.MysteryCrateBlueprints),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Could not find crate, try again or contact support.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return fmt.Errorf("start tx: %w", err)
	}
	defer func() {
		tx.Rollback()
	}()

	crate.Opened = true

	for _, blueprintItem := range crate.R.MysteryCrateBlueprints {
		if blueprintItem.BlueprintType == boiler.TemplateItemTypeMECH {

		}
		if blueprintItem.BlueprintType == boiler.TemplateItemTypeMECH_SKIN {

		}
		if blueprintItem.BlueprintType == boiler.TemplateItemTypeWEAPON {

		}
		if blueprintItem.BlueprintType == boiler.TemplateItemTypeWEAPON_SKIN {

		}
		if blueprintItem.BlueprintType == boiler.TemplateItemTypePOWER_CORE {

		}
		if blueprintItem.BlueprintType == boiler.TemplateItemTypeMECH {

		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		gamelog.L.Error().Err(err).Msg("failed to open mystery crate")
		return terror.Error(err, "Could not open mystery crate, please try again or contact support.")
	}

	resp := OpenCrateResponse{}
	reply(resp)

	return nil
}
