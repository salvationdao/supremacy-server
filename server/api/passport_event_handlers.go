package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/passport"

	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/tickle"
)

type PassportUserOnlineStatusRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID server.UserID `json:"userID"`
		Status bool          `json:"status"`
	} `json:"payload"`
}

func (api *API) PassportUserOnlineStatusHandler(ctx context.Context, payload []byte) {
	req := &PassportUserOnlineStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user online handler request")
	}

	// TODO: maybe add a difference between passport online and gameserver online
	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, req.Payload.UserID)), req.Payload.Status)
}

type PassportUserUpdatedRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		User *server.User `json:"user"`
	} `json:"payload"`
}

func (api *API) PassportUserUpdatedHandler(ctx context.Context, payload []byte) {
	req := &PassportUserUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user updated handler request")
	}

	// update faction
	if !req.Payload.User.FactionID.IsNil() {
		// get faction
		faction := api.factionMap[req.Payload.User.FactionID]

		req.Payload.User.Faction = faction

		if clientMap, ok := api.onlineClientMap[req.Payload.User.ID]; ok {
			clientMap <- func(cim ClientInstanceMap, cps *SupremacyTokenState, t *tickle.Tickle) {
				for client, ok := range cim {
					if ok {
						hubClient, ok := api.hubClientDetail[client]
						if ok {
							hubClient <- func(hcd *HubClientDetail) {
								hcd.FactionID = faction.ID
							}
						}
					}
				}
			}
		}

	}

	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUser, req.Payload.User.ID)), req.Payload.User)
}
