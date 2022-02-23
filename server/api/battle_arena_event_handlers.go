package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

func (api *API) BattleInitSignal(ctx context.Context, ed *battle_arena.EventData) {
	// clean up battle end information
	api.battleEndInfo = &BattleEndInfo{}

	// pass back nil to tell game ui to clean up current end battle message
	go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyBattleEndDetailUpdated), nil)
	// send new queue details

	go func() {
		for i := range api.BattleArena.BattleQueueMap {
			api.BattleArena.BattleQueueMap[i] <- func(wmq *battle_arena.WarMachineQueuingList) {
				// for each queue map
				userMap := make(map[server.UserID][]*passport.WarMachineQueuePosition)

				for i, wm := range wmq.WarMachines {
					if _, ok := userMap[wm.OwnedByID]; !ok {
						userMap[wm.OwnedByID] = []*passport.WarMachineQueuePosition{}
					}
					userMap[wm.OwnedByID] = append(userMap[wm.OwnedByID], &passport.WarMachineQueuePosition{
						WarMachineMetadata: wm,
						Position:           i + 1,
					})
				}

				for u, uwm := range userMap {
					go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueueUpdated, u)), uwm)
				}
			}
		}
	}()
}

const HubKeyGameSettingsUpdated = hub.HubCommandKey("GAME:SETTINGS:UPDATED")

type GameSettingsResponse struct {
	GameMap            *server.GameMap              `json:"gameMap"`
	WarMachines        []*server.WarMachineMetadata `json:"warMachines"`
	WarMachineLocation []byte                       `json:"warMachineLocation"`
}

// BattleStartSignal start all the voting cycle
func (api *API) BattleStartSignal(ctx context.Context, ed *battle_arena.EventData) {

	// build faction detail to battle start
	warMachines := ed.BattleArena.WarMachines
	for _, wm := range warMachines {
		wm.Faction = ed.BattleArena.FactionMap[wm.FactionID]
	}

	// marshal payload
	gameSettingsData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameSettingsUpdated,
		Payload: &GameSettingsResponse{
			GameMap:            ed.BattleArena.GameMap,
			WarMachines:        ed.BattleArena.WarMachines,
			WarMachineLocation: ed.BattleArena.BattleHistory[0],
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
			go client.Send(gameSettingsData)
		}
	})

	// start voting cycle, initial intro time equal: (mech_count * 3 + 7) seconds
	introSecond := len(warMachines)*3 + 7

	for factionID := range api.factionMap {
		go func(factionID server.FactionID) {

			initialAbilities := []*server.GameAbility{}

			if factionID != server.ZaibatsuFactionID {
				// get initial abilities
				initialAbilities, err = db.FactionExclusiveAbilitiesByFactionID(api.ctx, api.BattleArena.Conn, factionID)
				if err != nil {
					api.Log.Err(err).Msg("Failed to query initial faction abilities")
					return
				}

				for _, ab := range initialAbilities {
					ab.Title = "FACTION_WIDE"
					ab.CurrentSups = "0"
				}
			}

			for _, wm := range ed.BattleArena.WarMachines {
				if wm.FactionID != factionID || len(wm.Abilities) == 0 {
					continue
				}

				for _, ability := range wm.Abilities {
					wmAbility := &server.GameAbility{
						ID:                  server.GameAbilityID(uuid.Must(uuid.NewV4())), // generate a uuid for frontend to track sups contribution
						GameClientAbilityID: byte(ability.GameClientID),
						ImageUrl:            ability.Image,
						Description:         ability.Description,
						FactionID:           factionID,
						Label:               ability.Name,
						SupsCost:            ability.SupsCost,
						CurrentSups:         "0",
						AbilityTokenID:      ability.TokenID,
						WarMachineTokenID:   wm.TokenID,
						ParticipantID:       &wm.ParticipantID,
						Title:               wm.Name,
					}
					// if it is zaibatsu faction ability set id back
					if ability.GameClientID == 11 {
						wmAbility.ID = ability.ID
					}
					initialAbilities = append(initialAbilities, wmAbility)

				}
			}

			api.startGameAbilityPoolTicker(ctx, factionID, initialAbilities, introSecond)
		}(factionID)
	}

	go api.startVotingCycle(ctx, introSecond)
}

// BattleEndSignal terminate all the voting cycle
func (api *API) BattleEndSignal(ctx context.Context, ed *battle_arena.EventData) {

	battleStat := api.BattleArena.GetCurrentState()
	api.battleEndInfo.BattleID = battleStat.ID
	api.battleEndInfo.StartedAt = battleStat.StartedAt
	api.battleEndInfo.EndedAt = *battleStat.EndedAt
	// stop all the tickles in voting cycle
	go api.stopGameAbilityPoolTicker()

	battleViewers := api.viewerIDRead()
	// increment users' view battle count
	err := db.UserBattleViewUpsert(ctx, api.Conn, battleViewers)
	if err != nil {
		api.Log.Err(err).Msg("Failed to record users' battle count")
		return
	}

	userVoteList := api.stopVotingCycle(ctx)
	// start preparing ending broadcast data
	if len(userVoteList) > 0 {
		// insert user vote list to db
		err := db.UserBattleVoteCountInsert(context.Background(), api.Conn, api.BattleArena.CurrentBattleID(), userVoteList)
		if err != nil {
			api.Log.Err(err).Msg("Failed to record battle user vote")
			return
		}
	}

	// get the user who spend most sups during the battle from passport
	wg := sync.WaitGroup{}

	wg.Add(1)
	err = api.Passport.TopSupsContributorsGet(ctx, ed.BattleArena.StartedAt, time.Now(), func(msg []byte) {
		resp := &passport.SupremacyTopSupsContributorResponse{}
		err := json.Unmarshal(msg, resp)
		if err != nil {
			return
		}

		for _, topUser := range resp.Payload.TopSupsContributors {
			if !topUser.FactionID.IsNil() {
				topUser.Faction = api.factionMap[topUser.FactionID]
			}
			api.battleEndInfo.TopSupsContributors = append(api.battleEndInfo.TopSupsContributors, topUser.Brief())
		}

		for _, topFaction := range resp.Payload.TopSupsContributeFactions {
			fmt.Println(topFaction.Label)
			api.battleEndInfo.TopSupsContributeFactions = append(api.battleEndInfo.TopSupsContributeFactions, topFaction.Brief())
		}

		wg.Done()
	})
	wg.Wait()

	if err != nil {
		api.Log.Err(err).Msg("Failed to get top sups contributors from passport")
		return
	}

	// get most frequent trigger ability user
	us, err := db.UsersMostFrequentTriggerAbility(ctx, api.Conn, ed.BattleArena.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		api.Log.Err(err).Msg("Failed to get most frequent trigger ability user")
		return
	}

	userIDs := []server.UserID{}
	if len(us) <= 5 {
		for _, u := range us {
			userIDs = append(userIDs, u.ID)
		}
	} else {
		for len(userIDs) < 5 {
			userIDs = append(userIDs, us[len(userIDs)].ID)
		}
	}

	if len(userIDs) > 0 {
		wg := sync.WaitGroup{}
		wg.Add(1)
		api.Passport.UsersGet(userIDs, func(msg []byte) {
			defer wg.Done()
			resp := &passport.GetUsers{}
			err := json.Unmarshal(msg, resp)
			if err != nil {
				return
			}

			for _, userID := range userIDs {
				for _, user := range resp.Users {
					if user.ID == userID {
						if !user.FactionID.IsNil() {
							user.Faction = api.factionMap[user.FactionID]
						}
						api.battleEndInfo.MostFrequentAbilityExecutors = append(api.battleEndInfo.MostFrequentAbilityExecutors, user.Brief())
						break
					}
				}
			}
		})
		wg.Wait()
	}

	// set up rest of battle end info
	api.battleEndInfo.BattleIdentifier = ed.BattleArena.Identifier
	api.battleEndInfo.WinningCondition = *ed.BattleArena.WinningCondition
	api.battleEndInfo.WinningWarMachines = []*server.WarMachineBrief{}

	if len(ed.BattleArena.WinningWarMachines) > 0 {
		api.battleEndInfo.WinningFaction = ed.BattleArena.WinningWarMachines[0].Faction.Brief()
		for _, wm := range ed.BattleArena.WinningWarMachines {
			api.battleEndInfo.WinningWarMachines = append(api.battleEndInfo.WinningWarMachines, wm.Brief())
		}
	}

	b, err := json.Marshal(api.battleEndInfo)
	if err != nil {
		api.Log.Err(err).Msg("Failed to marshal battle end information")
		return
	}

	_, err = db.CreateBattleStateEvent(ctx, api.Conn, api.BattleArena.CurrentBattleID(), server.BattleEventBattleEnd, b)
	if err != nil {
		api.Log.Err(err).Msg("Failed to store battle end event")
		return
	}

	// broadcast battle end info back to game ui
	go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyBattleEndDetailUpdated), api.battleEndInfo)

	// refresh user stat
	if len(battleViewers) > 0 {
		go func() {
			err := db.UserStatMaterialisedViewRefresh(ctx, api.Conn)
			if err != nil {
				api.Log.Err(err).Msg("Failed to refresh user stats")
				return
			}

			// get user stat
			userStats, err := db.UserStatMany(ctx, api.Conn, battleViewers)
			if err != nil {
				api.Log.Err(err).Msg("Failed to get users' stat")
				return
			}

			userStatSends := []*passport.UserStatSend{}
			for _, us := range userStats {
				userStatSends = append(userStatSends, &passport.UserStatSend{
					Stat: us,
				})
			}

			go api.Passport.UserStatSend(ctx, userStatSends)
		}()
	}

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

				// TODO: set sups multiplier for these three rewards
				for _, executor := range api.battleEndInfo.MostFrequentAbilityExecutors {
					if executor.ID == userID {
						brs = append(brs, BattleRewardTypeAbilityExecutor)
						break
					}
				}

				for _, supsContributor := range api.battleEndInfo.TopSupsContributors {
					if supsContributor.ID == userID {
						brs = append(brs, BattleRewardTypeWarContributor)
						break
					}
				}

				if len(brs) == 0 {
					return
				}

				api.ClientBattleRewardUpdate(c, &ClientBattleReward{
					BattleID: api.BattleArena.CurrentBattleID(),
					Rewards:  brs,
				})
			}(c)
		}
	})

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

	// record destroyed event
	wmd := &WarMachineDestroyedEventRecord{
		KilledBy:            ed.WarMachineDestroyedRecord.KilledBy,
		DestroyedWarMachine: ed.WarMachineDestroyedRecord.DestroyedWarMachine.Brief(),
	}
	if ed.WarMachineDestroyedRecord.KilledByWarMachine != nil {
		wmd.KilledByWarMachine = ed.WarMachineDestroyedRecord.KilledByWarMachine.Brief()
	}

	go api.MessageBus.Send(ctx,
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

	api.NetMessageBus.Send(ctx, messagebus.NetBusKey(HubKeyWarMachineLocationUpdated), ed.WarMachineLocation)
}

func (api *API) UpdateWarMachineQueue(ctx context.Context, ed *battle_arena.EventData) {
	if ed.WarMachineQueue == nil {
		return
	}
	go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, ed.WarMachineQueue.FactionID)), ed.WarMachineQueue.WarMachines)
}
