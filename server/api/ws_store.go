package api

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"server/db/boiler"
)

type StoreController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

type MysteryCrateSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Type string `json:"type"`
	} `json:"payload"`
}

func NewStoreController(api *API) *StoreController {
	sc := &StoreController{
		API: api,
	}

	return sc
}

const HubkeyMysteryCrateSubscribe = "STORE:MYSTERY:CRATE:SUBSCRIBE"

func (sc *StoreController) MysteryCrateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MysteryCrateSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	return nil
}

type MysteryCratePurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Type string `json:"type"`
	} `json:"payload"`
}

const HubkeyMysteryCratePurchase = "STORE:MYSTERY:CRATE:PURCHASE"

func (sc *StoreController) PurchaseMysteryCrateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MysteryCrateSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	//check faction id != nil
	//get random crate where faction id == user.faction_id and purchased == false and opened == false and type == req.payload.type
	//check user SUPS is more than crate.price
	//take money - tx... transaction block?
	//create collection item with ownerID = user.Id

	//update mysterycrate subscribers
	ws.PublishMessage("/store/mystery_crate", HubkeyMysteryCrateSubscribe, req.Payload)

	//send back userSettings
	reply(nil)
	return nil
}
