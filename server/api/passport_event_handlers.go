package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/passport"
	"strconv"

	"github.com/ninja-software/hub/v3/ext/messagebus"
	"github.com/ninja-software/terror/v2"
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

	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, req.Payload.User.ID)), req.Payload.User)
}

type PassportUserSupsUpdatedRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID server.UserID `json:"userID"`
		Sups   server.BigInt `json:"sups"`
	} `json:"payload"`
}

func (api *API) PassportUserSupsUpdatedHandler(ctx context.Context, payload []byte) {
	req := &PassportUserSupsUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user sups updated request")
	}

	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsUpdated, req.Payload.UserID)), req.Payload.Sups.String())
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
		api.battleQueueMap[req.Payload.WarMachineNFT.FactionID] <- func(wmq *warMachineQueuingList) {
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
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, req.Payload.WarMachineNFT.FactionID)), wmq.WarMachines)
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
			api.Passport.WarMachineQueuePosition(ctx, fmt.Sprintf("war_machine_queue_position_%s", req.Payload.WarMachineNFT.OwnedByID), []*passport.UserWarMachineQueuePosition{
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
		api.battleQueueMap[req.Payload.WarMachineNFT.FactionID] <- func(wmq *warMachineQueuingList) {
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

				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, req.Payload.WarMachineNFT.FactionID)), wmq.WarMachines[:maxLength])
			}

			api.Passport.WarMachineQueuePosition(context.Background(), fmt.Sprintf("war_machine_queue_position_%s", req.Payload.WarMachineNFT.OwnedByID), BuildUserWarMachineQueuePosition(wmq.WarMachines))
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

	api.battleQueueMap[req.Payload.FactionID] <- func(wmq *warMachineQueuingList) {
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
		api.Passport.WarMachineQueuePosition(ctx, fmt.Sprintf("war_machine_queue_position_%s", req.Payload.UserID), []*passport.UserWarMachineQueuePosition{
			{
				UserID:                   req.Payload.UserID,
				WarMachineQueuePositions: warMachineQueuePosition,
			},
		})
	}
}
