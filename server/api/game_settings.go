package api

import (
	"context"
	"net/http"
	"server"
	"server/battle_arena"
	"server/helpers"

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
