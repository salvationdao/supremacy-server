package api

import (
	"context"
	"github.com/ninja-syndicate/ws"
	"server/db/boiler"
)

func NewFactionPassController(api *API) {
	api.SecureUserFactionCommand(HubKeyFactionPassPurchase, api.FactionPassPurchase)
}

type FactionPassPurchaseRequest struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

const HubKeyFactionPassPurchase = "FACTION:PASS:PURCHASE"

func (api *API) FactionPassPurchase(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	return nil
}
