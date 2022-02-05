package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle_arena"
	"server/passport"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

// type PassportUserOnlineStatusRequest struct {
// 	Key     passport.Event `json:"key"`
// 	Payload struct {
// 		UserID server.UserID `json:"userID"`
// 		Status bool          `json:"status"`
// 	} `json:"payload"`
// }

// func (api *API) PassportUserOnlineStatusHandler(ctx context.Context, payload []byte) {
// 	req := &PassportUserOnlineStatusRequest{}
// 	err := json.Unmarshal(payload, req)
// 	if err != nil {
// 		api.Log.Err(err).Msg("error unmarshalling passport user online handler request")
// 	}

// 	// TODO: maybe add a difference between passport online and gameserver online
// 	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, req.Payload.UserID)), req.Payload.Status)
// }

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

	uid := req.Payload.User.ID.String()

	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok || client.Identifier() != uid {
				continue
			}

			go func(c *hub.Client) {
				// update client detail
				api.hubClientDetail[c] <- func(hcd *HubClientDetail) {
					hcd.FirstName = req.Payload.User.FirstName
					hcd.LastName = req.Payload.User.LastName
					hcd.Username = req.Payload.User.Username

					if hcd.FactionID == req.Payload.User.FactionID {
						return
					}

					// if faction id has changed, send the updated user to twitch ui
					hcd.FactionID = req.Payload.User.FactionID

					user := &server.User{
						ID:        req.Payload.User.ID,
						FactionID: req.Payload.User.FactionID,
					}

					if !req.Payload.User.FactionID.IsNil() {
						user.Faction = api.factionMap[req.Payload.User.FactionID]
					}

					// send
					resp := struct {
						Key           hub.HubCommandKey `json:"key"`
						TransactionID string            `json:"transactionID"`
						Payload       interface{}       `json:"payload"`
					}{
						Key:           HubKeyUserSubscribe,
						TransactionID: "userUpdate",
						Payload:       user,
					}

					b, err := json.Marshal(resp)
					if err != nil {
						api.Hub.Log.Err(err).Errorf("send: issue marshalling resp")
						return
					}

					err = c.Send(b)
					if err != nil {
						api.Log.Err(err).Msg("Failed to send auth response back to twitch client")
						return
					}
				}
			}(client)
		}
	})
}

type PassportUserEnlistFactionRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID    server.UserID    `json:"userID"`
		FactionID server.FactionID `json:"factionID"`
	} `json:"payload"`
}

func (api *API) PassportUserEnlistFactionHandler(ctx context.Context, payload []byte) {
	req := &PassportUserEnlistFactionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user updated handler request")
	}

	uid := req.Payload.UserID.String()

	// prepare broadcast data
	faction := api.factionMap[req.Payload.FactionID]
	user := &server.User{
		ID:        req.Payload.UserID,
		FactionID: req.Payload.FactionID,
		Faction:   faction,
	}
	// send
	resp := struct {
		Key           hub.HubCommandKey `json:"key"`
		TransactionID string            `json:"transactionID"`
		Payload       interface{}       `json:"payload"`
	}{
		Key:           HubKeyUserSubscribe,
		TransactionID: "userUpdate",
		Payload:       user,
	}
	broadcastData, err := json.Marshal(resp)
	if err != nil {
		api.Hub.Log.Err(err).Errorf("send: issue marshalling resp")
		return
	}

	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok || client.Identifier() != uid {
				continue
			}

			go func(c *hub.Client) {
				api.hubClientDetail[c] <- func(hcd *HubClientDetail) {
					// update client facton id
					hcd.FactionID = req.Payload.FactionID
				}

				err = c.Send(broadcastData)
				if err != nil {
					api.Log.Err(err).Msg("Failed to send auth response back to client")
					return
				}
			}(client)
		}
	})
}

type BattleQueueJoinRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		WarMachineNFT *server.WarMachineNFT `json:"warMachineNFT"`
	} `json:"payload"`
}

func (api *API) PassportBattleQueueJoinHandler(ctx context.Context, payload []byte) {
	req := &BattleQueueJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue join request")
	}

	if !req.Payload.WarMachineNFT.FactionID.IsNil() {
		api.BattleArena.BattleQueueMap[req.Payload.WarMachineNFT.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
			// skip if the war machine already join the queue
			if checkWarMachineExist(wmq.WarMachines, req.Payload.WarMachineNFT.TokenID) != -1 {
				api.Log.Err(terror.ErrInvalidInput).Msgf("Asset %d is already in the queue", req.Payload.WarMachineNFT.TokenID)
				return
			}

			// fire a freeze command to the passport server
			err := api.Passport.AssetFreeze(ctx, "asset_freeze"+strconv.Itoa(int(req.Payload.WarMachineNFT.TokenID)), req.Payload.WarMachineNFT.TokenID)
			if err != nil {
				api.Log.Err(err).Msgf("Failed to freeze asset %d", req.Payload.WarMachineNFT.TokenID)
				return
			}

			wmq.WarMachines = append(wmq.WarMachines, req.Payload.WarMachineNFT)

			// broadcast next 5 queuing war machines to twitch ui
			if len(wmq.WarMachines) <= 5 {
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, req.Payload.WarMachineNFT.FactionID)), wmq.WarMachines)
			}

			// broadcast war machine queue position update
			warMachineQueuePosition := []*passport.WarMachineQueuePosition{}
			for i, wm := range wmq.WarMachines {
				if wm.OwnedByID != req.Payload.WarMachineNFT.OwnedByID {
					continue
				}
				warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
					WarMachineNFT: wm,
					Position:      i,
				})
			}

			// fire a war machine queue passport request
			api.Passport.WarMachineQueuePositionBroadcast(ctx, []*passport.UserWarMachineQueuePosition{
				{
					UserID:                   req.Payload.WarMachineNFT.OwnedByID,
					WarMachineQueuePositions: warMachineQueuePosition,
				},
			})
		}
	}
}

type BattleQueueReleaseRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		WarMachineNFT *server.WarMachineNFT `json:"warMachineNFT"`
	} `json:"payload"`
}

func (api *API) PassportBattleQueueReleaseHandler(ctx context.Context, payload []byte) {
	req := &BattleQueueReleaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue release request")
	}

	if !req.Payload.WarMachineNFT.FactionID.IsNil() {
		api.BattleArena.BattleQueueMap[req.Payload.WarMachineNFT.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
			// check war machine is in the queue
			index := checkWarMachineExist(wmq.WarMachines, req.Payload.WarMachineNFT.TokenID)
			if index < 0 {
				api.Log.Err(terror.ErrInvalidInput).Msgf("Asset %d is not in the queue", req.Payload.WarMachineNFT.TokenID)
				return
			}

			// fire a freeze command to the passport server
			api.Passport.AssetRelease(ctx, "asset_release"+strconv.Itoa(int(req.Payload.WarMachineNFT.TokenID)), []*server.WarMachineNFT{wmq.WarMachines[index]})

			copy(wmq.WarMachines[index:], wmq.WarMachines[index+1:])   // Shift wmq.WarMachines[i+1:] left one index.
			wmq.WarMachines[len(wmq.WarMachines)-1] = nil              // wmq.WarMachinesse wmq.WarMachinesst element (write zero vwmq.WarMachineslue).
			wmq.WarMachines = wmq.WarMachines[:len(wmq.WarMachines)-1] // Truncate slice.

			// broadcast next 5 queuing war machines to twitch ui
			if index <= 5 {
				maxLength := 5
				if len(wmq.WarMachines) < maxLength {
					maxLength = len(wmq.WarMachines)
				}

				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, req.Payload.WarMachineNFT.FactionID)), wmq.WarMachines[:maxLength])
			}

			api.Passport.WarMachineQueuePositionBroadcast(context.Background(), battle_arena.BuildUserWarMachineQueuePosition(wmq.WarMachines))
		}
	}
}

// checkWarMachineExist return true if war machine already exist in the list
func checkWarMachineExist(list []*server.WarMachineNFT, tokenID uint64) int {
	for i, wm := range list {
		if wm.TokenID == tokenID {
			return i
		}
	}

	return -1
}

type WarMachineQueuePositionRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		FactionID server.FactionID `json:"factionID"`
		UserID    server.UserID    `json:"userID"`
	} `json:"payload"`
}

func (api *API) PassportWarMachineQueuePositionHandler(ctx context.Context, payload []byte) {
	req := &WarMachineQueuePositionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue release request")
	}

	warMachineQueuePositionChan := make(chan []*passport.WarMachineQueuePosition)

	api.BattleArena.BattleQueueMap[req.Payload.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
		warMachineQueuePosition := []*passport.WarMachineQueuePosition{}
		for i, wm := range wmq.WarMachines {
			if wm.OwnedByID != req.Payload.UserID {
				continue
			}
			warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
				WarMachineNFT: wm,
				Position:      i,
			})
		}

		warMachineQueuePositionChan <- warMachineQueuePosition
	}

	warMachineQueuePosition := <-warMachineQueuePositionChan

	// fire a war machine queue passport request
	if len(warMachineQueuePosition) > 0 {
		api.Passport.WarMachineQueuePositionBroadcast(ctx, []*passport.UserWarMachineQueuePosition{
			{
				UserID:                   req.Payload.UserID,
				WarMachineQueuePositions: warMachineQueuePosition,
			},
		})
	}
}

type AuthedTwitchExtensionRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		User                server.User   `json:"user"`
		SessionID           hub.SessionID `json:"sessionID"`
		TwitchExtensionJWT  string        `json:"twitchExtensionJWT"`
		GameserverSessionID string        `json:"gameserverSessionID"`
	} `json:"payload"`
}

func (api *API) AuthRingCheckHandler(ctx context.Context, payload []byte) {
	req := &AuthedTwitchExtensionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport auth ring check")
		return
	}

	if req.Payload.TwitchExtensionJWT == "" && req.Payload.GameserverSessionID == "" {
		api.Log.Err(fmt.Errorf("Not auth key provided"))
		return
	}

	ringCheckKey := req.Payload.TwitchExtensionJWT
	if ringCheckKey == "" {
		ringCheckKey = req.Payload.GameserverSessionID
	}

	api.ringCheckAuthChan <- func(rca RingCheckAuthMap) {
		hubClient, ok := rca[ringCheckKey]
		if !ok {
			return
		}

		if req.Payload.GameserverSessionID != "" && req.Payload.GameserverSessionID != string(hubClient.SessionID) {
			api.Log.Err(fmt.Errorf("Session id does not match"))
			return
		}
		// reset session id for security
		hubClient.SessionID = hub.SessionID(uuid.Must(uuid.NewV4()).String())

		hubClientDetail, ok := api.hubClientDetail[hubClient]
		if !ok {
			return
		}

		// set hub client detail
		hubClientDetail <- func(hcd *HubClientDetail) {
			hcd.FactionID = req.Payload.User.FactionID
			hcd.Username = req.Payload.User.Username
			hcd.FirstName = req.Payload.User.FirstName
			hcd.LastName = req.Payload.User.LastName
		}

		// set user id
		hubClient.SetIdentifier(req.Payload.User.ID.String())

		// set user online
		api.ClientOnline(hubClient)

		// parse user response
		user := &server.User{
			ID: req.Payload.User.ID,
		}
		if !req.Payload.User.FactionID.IsNil() {
			user.FactionID = req.Payload.User.FactionID
			user.Faction = api.factionMap[req.Payload.User.FactionID]
		}

		// send user id and faction id back to twitch ui client
		resp := struct {
			Key           hub.HubCommandKey `json:"key"`
			TransactionID string            `json:"transactionID"`
			Payload       interface{}       `json:"payload"`
		}{
			Key:           HubKeyUserSubscribe,
			TransactionID: "authRingCheck",
			Payload:       user,
		}

		b, err := json.Marshal(resp)
		if err != nil {
			api.Hub.Log.Err(err).Errorf("send: issue marshalling resp")
			return
		}

		err = hubClient.Send(b)
		if err != nil {
			api.Log.Err(err).Msg("Failed to send auth response back to twitch client")
			return
		}

		// send request to passport server to upgrade the gamebar user
		err = api.Passport.UpgradeUserConnection(ctx, req.Payload.SessionID, string(req.Payload.SessionID))
		if err != nil {
			api.Log.Err(err).Msg("Failed to upgrade passport hub client level")
			return
		}

		// delete jwt from map
		delete(rca, req.Payload.TwitchExtensionJWT)
	}
}
