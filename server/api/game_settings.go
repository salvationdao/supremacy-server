package api

import (
	"context"
	"encoding/json"
	"gameserver"
	"gameserver/battle_arena"
	"gameserver/helpers"
	"net/http"

	"github.com/ninja-software/hub/v2"
)

// GetGameSettings return current game settings
func (api *API) GetGameSettings(w http.ResponseWriter, r *http.Request) (int, error) {
	resp := &GameSettingsResponse{
		GameMap:     &gameserver.GameMap{},
		WarMachines: []*gameserver.WarMachine{},
	}

	if api.BattleArena.GetCurrentState() != nil {
		resp.GameMap = api.BattleArena.GetCurrentState().Map
		resp.WarMachines = api.BattleArena.GetCurrentState().WarMachines
	}

	return helpers.EncodeJSON(w, resp)
}

const HubKeyWarMachinePositionUpdated hub.HubCommandKey = hub.HubCommandKey("WARMACHINE:UPDATED")

func (api *API) UpdateWarMachinePosition(ctx context.Context, ed *battle_arena.EventData) {

	if len(ed.BattleArena.WarMachines) == 0 {
		return
	}

	positions := []*gameserver.WarMachine{}

	for _, warmachine := range ed.BattleArena.WarMachines {
		positions = append(positions, &gameserver.WarMachine{
			ID:                 warmachine.ID,
			WarMachinePosition: warmachine.WarMachinePosition,
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
			go client.Send(data)
		}
	})
}
