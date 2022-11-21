package api

import (
	"context"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server/db/boiler"
	"server/gamedb"
)

func NewFactionPassController(api *API) {
	api.SecureUserFactionCommand(HubKeyFactionPassPurchase, api.FactionPassPurchase)
}

type FactionPassPurchaseRequest struct {
	Payload struct {
	} `json:"payload"`
}

const HubKeyFactionPassPurchase = "FACTION:PASS:PURCHASE"

func (api *API) FactionPassPurchase(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	return nil
}

func (api *API) FactionPassList(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	fps, err := boiler.FactionPasses().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load faction passes")
	}
	reply(fps)
	return nil
}
