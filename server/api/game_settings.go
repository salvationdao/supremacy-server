package api

import (
	"context"
	"net/http"
	"server"
	"server/battle_arena"
	"server/helpers"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"nhooyr.io/websocket"

	"github.com/ninja-syndicate/hub"
)

// GetGameSettings return current game settings
func (api *API) GetGameSettings(w http.ResponseWriter, r *http.Request) (int, error) {
	resp := &GameSettingsResponse{
		GameMap:     &server.GameMap{},
		WarMachines: []*server.WarMachineNFT{},
	}

	if api.BattleArena.GetCurrentState() != nil {
		resp.GameMap = api.BattleArena.GetCurrentState().GameMap
		resp.WarMachines = api.BattleArena.GetCurrentState().WarMachines
	}

	return helpers.EncodeJSON(w, resp)
}

func (api *API) UpdateWarMachinePosition(ctx context.Context, ed *battle_arena.EventData) {
	if len(ed.BattleArena.WarMachines) == 0 {
		return
	}

	// broadcast game settings to all the connected clients
	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok {
				continue
			}
			go func(c *hub.Client) {
				err := c.SendWithMessageType(ed.WarMachineLocation, websocket.MessageBinary)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
		}
	})
}

func (api *API) UpdateWarMachineHitPoint(ctx context.Context, ed *battle_arena.EventData) {
	if len(ed.BattleArena.WarMachines) == 0 {
		return
	}

	// broadcast game settings to all the connected clients
	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok {
				continue
			}
			go func(c *hub.Client) {
				err := c.SendWithMessageType(ed.WarMachineHitPoint, websocket.MessageBinary)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
		}
	})
}

// WinningFactionViewerIDsGet return the list of viewer id with in the winning faction
func (api *API) WinningFactionViewerIDsGet(ctx context.Context, ed *battle_arena.EventData) {
	if ed.WinnerFactionViewers == nil {
		api.Log.Err(terror.ErrInvalidInput).Msg("Winner faction view request is missing")
		return
	}

	viewerIDs := []server.UserID{}

	for wsc, hubClientDetailChan := range api.hubClientDetail {

		// get hub client detail
		detailChan := make(chan *HubClientDetail)
		hubClientDetailChan <- func(hcd *HubClientDetail) {
			detailChan <- hcd
		}
		detail := *<-detailChan

		// skip, if current client is not enlisted in the winning faction
		if detail.FactionID != ed.WinnerFactionViewers.WinnerFactionID {
			continue
		}

		// check the client is already in the list
		exists := false
		for _, viewerID := range viewerIDs {
			if viewerID.String() == wsc.Identifier() {
				exists = true
				break
			}
		}

		// append the client to the list
		if !exists {
			viewerIDs = append(viewerIDs, server.UserID(uuid.Must(uuid.FromString(wsc.Identifier()))))
		}
	}

	// return result
	ed.WinnerFactionViewers.CallbackChannel <- viewerIDs
}
