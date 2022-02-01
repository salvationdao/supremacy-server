package api

import (
	"context"
	"encoding/json"
	"net/http"
	"server"
	"server/battle_arena"
	"server/helpers"

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

const HubKeyWarMachinePositionUpdated hub.HubCommandKey = hub.HubCommandKey("WARMACHINE:UPDATED")

type WarMachineState struct {
	TokenID        uint64          `json:"tokenID"`
	Position       *server.Vector3 `json:"position"`
	Rotation       int             `json:"rotation"`
	RemainHitPoint int             `json:"remainHitPoint"`
	RemainShield   int             `json:"remainShield"`
}

func (api *API) UpdateWarMachineState(ctx context.Context, ed *battle_arena.EventData) {
	if len(ed.BattleArena.WarMachines) == 0 {
		return
	}

	warMachineStates := []*WarMachineState{}

	for _, warmachine := range ed.BattleArena.WarMachines {
		warMachineStates = append(warMachineStates, &WarMachineState{
			TokenID:        warmachine.TokenID,
			Position:       warmachine.Position,
			Rotation:       warmachine.Rotation,
			RemainHitPoint: warmachine.RemainHitPoint,
			RemainShield:   warmachine.MaxShield,
		})
	}

	// parse broadcast data
	data, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyWarMachinePositionUpdated,
		Payload: warMachineStates,
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
