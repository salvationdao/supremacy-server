package api

import (
	"context"
	"encoding/json"
	"net/http"
	"server"
	"server/battle_arena"
	"server/helpers"

	"github.com/ninja-software/hub/v2"
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

const HubKeyWarMachinePositionUpdated hub.HubCommandKey = hub.HubCommandKey("WARMACHINE:UPDATED")

func (api *API) UpdateWarMachinePosition(ctx context.Context, ed *battle_arena.EventData) {

	if len(ed.BattleArena.WarMachines) == 0 {
		return
	}

	positions := []*server.WarMachineNFT{}

	for _, warmachine := range ed.BattleArena.WarMachines {
		positions = append(positions, &server.WarMachineNFT{
			TokenID:  warmachine.TokenID,
			Position: warmachine.Position,
			Rotation: warmachine.Rotation,
		})
	}

	// parse broadcast data
	data, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyWarMachinePositionUpdated,
		Payload: positions,
	})
	if err != nil {
		return
	}

	// broadcast game settings to all the connected clients
	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok {
				continue
			}
			go func(c *hub.Client) {
				err := c.Send(data)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
		}
	})
}
