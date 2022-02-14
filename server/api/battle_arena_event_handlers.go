package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"nhooyr.io/websocket"
)

const HubKeyGameSettingsUpdated = hub.HubCommandKey("GAME:SETTINGS:UPDATED")

type GameSettingsResponse struct {
	GameMap     *server.GameMap         `json:"gameMap"`
	WarMachines []*server.WarMachineNFT `json:"warMachines"`
	// WarMachineLocation []byte                  `json:"warMachineLocation"`
}

// BattleStartSignal start all the voting cycle
func (api *API) BattleStartSignal(ctx context.Context, ed *battle_arena.EventData) {
	// clean up current viewer id map
	api.viewerIDRead()

	// build faction detail to battle start
	warMachines := ed.BattleArena.WarMachines
	for _, wm := range warMachines {
		wm.Faction = ed.BattleArena.FactionMap[wm.FactionID]
	}

	// marshal payload
	gameSettingsData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameSettingsUpdated,
		Payload: &GameSettingsResponse{
			GameMap:     ed.BattleArena.GameMap,
			WarMachines: ed.BattleArena.WarMachines,
			// WarMachineLocation: ed.BattleArena.BattleHistory[0],
		},
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
				err := c.Send(gameSettingsData)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
		}
	})

	// start voting cycle, initial intro time equal: (mech_count * 3 + 7) seconds
	introSecond := len(warMachines)*3 + 7

	for factionID := range api.factionMap {
		go func(factionID server.FactionID) {
			// get initial abilities
			initialAbilities, err := db.FactionExclusiveAbilitiesByFactionID(api.ctx, api.BattleArena.Conn, factionID)
			if err != nil {
				api.Log.Err(err).Msg("Failed to query initial faction abilities")
				return
			}
			for _, ab := range initialAbilities {
				ab.Title = "FACTION_WIDE"
				ab.CurrentSups = "0"
			}

			for _, wm := range ed.BattleArena.WarMachines {
				if wm.FactionID != factionID || len(wm.Abilities) == 0 {
					continue
				}

				for _, ability := range wm.Abilities {
					initialAbilities = append(initialAbilities, &server.GameAbility{
						ID:                  server.GameAbilityID(uuid.Must(uuid.NewV4())), // generate a uuid for frontend to track sups contribution
						GameClientAbilityID: byte(ability.GameClientID),
						ImageUrl:            ability.Image,
						FactionID:           factionID,
						Label:               ability.Name,
						SupsCost:            ability.SupsCost,
						CurrentSups:         "0",
						AbilityTokenID:      ability.TokenID,
						WarMachineTokenID:   wm.TokenID,
						ParticipantID:       &wm.ParticipantID,
						WarMachineName:      wm.Name,  // for game notification
						WarMachineImage:     wm.Image, // for game notification
						Title:               wm.Name,
					})
				}
			}

			api.startGameAbilityPoolTicker(factionID, initialAbilities, introSecond)
		}(factionID)
	}

	go api.startVotingCycle(introSecond)
}

// BattleEndSignal terminate all the voting cycle
func (api *API) BattleEndSignal(ctx context.Context, ed *battle_arena.EventData) {
	// stop all the tickles in voting cycle
	go api.stopVotingCycle()
	go api.stopGameAbilityPoolTicker()

	// parse battle reward list
	api.Hub.Clients(func(clients hub.ClientsList) {
		for c := range clients {
			go func(c *hub.Client) {
				userID := server.UserID(uuid.FromStringOrNil(c.Identifier()))
				if userID.IsNil() {
					return
				}
				hcd, err := api.getClientDetailFromChannel(c)
				if err != nil || hcd.FactionID.IsNil() {
					return
				}

				brs := []BattleRewardType{}
				// check reward
				if hcd.FactionID == ed.BattleRewardList.WinnerFactionID {
					brs = append(brs, BattleRewardTypeFaction)
				}

				if _, ok := ed.BattleRewardList.WinningWarMachineOwnerIDs[userID]; ok {
					brs = append(brs, BattleRewardTypeWinner)
				}

				if _, ok := ed.BattleRewardList.ExecuteKillWarMachineOwnerIDs[userID]; ok {
					brs = append(brs, BattleRewardTypeKill)
				}

				if len(brs) == 0 {
					return
				}

				api.ClientBattleRewardUpdate(c, &ClientBattleReward{
					BattleID: ed.BattleRewardList.BattleID,
					Rewards:  brs,
				})
			}(c)
		}
	})

	// increment users' view battle count
	go func() {
		err := db.UserBattleViewRecord(ctx, api.Conn, api.viewerIDRead())
		if err != nil {
			api.Log.Err(err).Msg("Failed to record users' battle count")
			return
		}
	}()

	// trigger faction stat refresh and send result to passport server
	go func() {
		ctx := context.Background()

		// update factions' stat
		err := db.FactionStatMaterialisedViewRefresh(ctx, api.Conn)
		if err != nil {
			api.Log.Err(err).Msg("Failed to refresh materialised view")
			return
		}

		// get all the faction stat
		factionStats, err := db.FactionStatAll(ctx, api.Conn)
		if err != nil {
			api.Log.Err(err).Msg("failed to query faction stats")
			return
		}

		sendRequest := []*passport.FactionStatSend{}
		for _, factionStat := range factionStats {
			sendRequest = append(sendRequest, &passport.FactionStatSend{
				FactionStat: factionStat,
			})
		}

		// send faction stat to passport server
		err = api.Passport.FactionStatsSend(ctx, sendRequest)
		if err != nil {
			api.Log.Err(err).Msg("failed to send faction stat")
			return
		}
	}()
}

func (api *API) WarMachineDestroyedBroadcast(ctx context.Context, ed *battle_arena.EventData) {
	api.MessageBus.Send(
		messagebus.BusKey(
			fmt.Sprintf(
				"%s:%x",
				HubKeyWarMachineDestroyedUpdated,
				ed.WarMachineDestroyedRecord.DestroyedWarMachine.ParticipantID,
			),
		),
		ed.WarMachineDestroyedRecord,
	)
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

func (api *API) UpdateWarMachineQueue(ctx context.Context, ed *battle_arena.EventData) {
	if ed.WarMachineQueue == nil {
		return
	}
	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, ed.WarMachineQueue.FactionID)), ed.WarMachineQueue.WarMachines)
}
