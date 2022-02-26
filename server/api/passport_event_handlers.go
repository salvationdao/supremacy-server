package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

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
		return
	}

	detail, err := api.UserMap.GetUserDetailByID(req.Payload.User.ID)
	if err != nil {
		api.Log.Err(err)
		return
	}

	// prepare broadcast data
	req.Payload.User.Faction = api.factionMap[req.Payload.User.FactionID]
	// send
	resp := struct {
		Key           hub.HubCommandKey `json:"key"`
		TransactionID string            `json:"transactionID"`
		Payload       interface{}       `json:"payload"`
	}{
		Key:           HubKeyUserSubscribe,
		TransactionID: "userUpdate",
		Payload:       req.Payload.User,
	}
	broadcastData, err := json.Marshal(resp)
	if err != nil {
		api.Hub.Log.Err(err).Errorf("send: issue marshalling resp")
		return
	}

	// update user
	detail.FirstName = req.Payload.User.FirstName
	detail.LastName = req.Payload.User.LastName
	detail.Username = req.Payload.User.Username
	detail.AvatarID = req.Payload.User.AvatarID

	if detail.FactionID == req.Payload.User.FactionID {
		for _, cl := range api.UserMap.Update(detail) {
			go cl.Send(broadcastData)

		}
		return
	}

	// update user faction
	oldFactionID := detail.FactionID
	newFactionID := req.Payload.User.FactionID
	detail.FactionID = req.Payload.User.FactionID
	if !detail.FactionID.IsNil() {
		detail.Faction = api.factionMap[detail.FactionID]
	}
	for _, cl := range api.UserMap.Update(detail) {
		api.viewerLiveCount.Swap(oldFactionID, newFactionID)
		go cl.Send(broadcastData)

	}

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
		return
	}

	detail, err := api.UserMap.GetUserDetailByID(req.Payload.UserID)
	if err != nil {
		api.Log.Err(err)
		return
	}

	oldFactionID := detail.FactionID
	newFactionID := req.Payload.FactionID
	detail.FactionID = req.Payload.FactionID
	detail.Faction = api.factionMap[detail.FactionID]

	// prepare broadcast data
	resp := struct {
		Key           hub.HubCommandKey `json:"key"`
		TransactionID string            `json:"transactionID"`
		Payload       interface{}       `json:"payload"`
	}{
		Key:           HubKeyUserSubscribe,
		TransactionID: "userUpdate",
		Payload:       detail,
	}
	broadcastData, err := json.Marshal(resp)
	if err != nil {
		api.Hub.Log.Err(err).Errorf("send: issue marshalling resp")
		return
	}

	fmt.Println("old faction id", oldFactionID)
	fmt.Println("new faction id", newFactionID)
	for _, cl := range api.UserMap.Update(detail) {
		api.viewerLiveCount.Swap(oldFactionID, newFactionID)
		go cl.Send(broadcastData)
	}
}

type BattleQueueJoinRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	} `json:"payload"`
}

func (api *API) PassportBattleQueueJoinHandler(ctx context.Context, payload []byte) {
	req := &BattleQueueJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue join request")
		return
	}

	if !req.Payload.WarMachineMetadata.FactionID.IsNil() {
		api.BattleArena.BattleQueueMap[req.Payload.WarMachineMetadata.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
			// skip if the war machine already join the queue
			if checkWarMachineExist(wmq.WarMachines, req.Payload.WarMachineMetadata.Hash) != -1 {
				api.Log.Err(terror.ErrInvalidInput).Msgf("Asset %s is already in the queue", req.Payload.WarMachineMetadata.Hash)
				return
			}

			// fire a freeze command to the passport server
			err := api.Passport.AssetFreeze(req.Payload.WarMachineMetadata.Hash)
			if err != nil {
				api.Log.Err(err).Msgf("Failed to freeze asset %s", req.Payload.WarMachineMetadata.Hash)
				return
			}

			// insert war machine into db
			err = db.BattleQueueInsert(ctx, api.Conn, req.Payload.WarMachineMetadata)
			if err != nil {
				api.Log.Err(err).Msgf("Failed to insert a copy of queue in db, token id: %s", req.Payload.WarMachineMetadata.Hash)
				return
			}

			wmq.WarMachines = append(wmq.WarMachines, req.Payload.WarMachineMetadata)

			// broadcast next 5 queuing war machines to twitch ui
			if len(wmq.WarMachines) <= 5 {
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, req.Payload.WarMachineMetadata.FactionID)), wmq.WarMachines)
			}

			// broadcast war machine queue position update
			warMachineQueuePosition := []*passport.WarMachineQueuePosition{}
			for i, wm := range wmq.WarMachines {
				if wm.OwnedByID != req.Payload.WarMachineMetadata.OwnedByID {
					continue
				}
				warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
					WarMachineMetadata: wm,
					Position:           i,
				})
			}

			// fire a war machine queue passport request
			go api.Passport.WarMachineQueuePositionBroadcast([]*passport.UserWarMachineQueuePosition{
				{
					UserID:                   req.Payload.WarMachineMetadata.OwnedByID,
					WarMachineQueuePositions: warMachineQueuePosition,
				},
			})
			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueueUpdated, req.Payload.WarMachineMetadata.OwnedByID)), warMachineQueuePosition)

		}
	}
}

type BattleQueueReleaseRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	} `json:"payload"`
}

func (api *API) PassportBattleQueueReleaseHandler(ctx context.Context, payload []byte) {
	req := &BattleQueueReleaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue release request")
		return
	}

	if !req.Payload.WarMachineMetadata.FactionID.IsNil() {
		api.BattleArena.BattleQueueMap[req.Payload.WarMachineMetadata.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
			// check war machine is in the queue
			index := checkWarMachineExist(wmq.WarMachines, req.Payload.WarMachineMetadata.Hash)
			if index < 0 {
				api.Log.Err(terror.ErrInvalidInput).Msgf("Asset %s is not in the queue", req.Payload.WarMachineMetadata.Hash)
				return
			}

			// fire a freeze command to the passport server
			api.Passport.AssetRelease([]*server.WarMachineMetadata{wmq.WarMachines[index]})

			copy(wmq.WarMachines[index:], wmq.WarMachines[index+1:])   // Shift wmq.WarMachines[i+1:] left one index.
			wmq.WarMachines[len(wmq.WarMachines)-1] = nil              // wmq.WarMachinesse wmq.WarMachinesst element (write zero vwmq.WarMachineslue).
			wmq.WarMachines = wmq.WarMachines[:len(wmq.WarMachines)-1] // Truncate slice.

			// remove the war machine queue copy in db
			err = db.BattleQueueRemove(ctx, api.Conn, req.Payload.WarMachineMetadata)
			if err != nil {
				api.Log.Err(err).Msgf("failed to remove war machine queue in db, token id: %s", req.Payload.WarMachineMetadata.Hash)
				return
			}

			// broadcast next 5 queuing war machines to twitch ui
			if index <= 5 {
				maxLength := 5
				if len(wmq.WarMachines) < maxLength {
					maxLength = len(wmq.WarMachines)
				}

				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, req.Payload.WarMachineMetadata.FactionID)), wmq.WarMachines[:maxLength])
			}
			result := api.BattleArena.BuildUserWarMachineQueuePosition(wmq.WarMachines, []*server.WarMachineMetadata{}, req.Payload.WarMachineMetadata.OwnedByID)
			go api.Passport.WarMachineQueuePositionBroadcast(result)

			warMachineQueuePosition := make([]*passport.WarMachineQueuePosition, 0)
			for _, qp := range result {
				if qp.UserID != req.Payload.WarMachineMetadata.OwnedByID {
					continue
				}
				warMachineQueuePosition = qp.WarMachineQueuePositions
			}
			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueueUpdated, req.Payload.WarMachineMetadata.OwnedByID)), warMachineQueuePosition)

		}

	}
}

// checkWarMachineExist return true if war machine already exist in the list
func checkWarMachineExist(list []*server.WarMachineMetadata, hash string) int {
	for i, wm := range list {
		if wm.Hash == hash {
			return i
		}
	}

	return -1
}

type AssetInsurancePayRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		FactionID server.FactionID `json:"factionID"`
		AssetHash string           `json:"assetHash"`
	} `json:"payload"`
}

func (api *API) PassportAssetInsurancePayHandler(ctx context.Context, payload []byte) {
	req := &AssetInsurancePayRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue release request")
		return
	}

	if !req.Payload.FactionID.IsNil() {
		api.BattleArena.BattleQueueMap[req.Payload.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
			// check war machine is in the queue
			index := checkWarMachineExist(wmq.WarMachines, req.Payload.AssetHash)
			if index < 0 {
				api.Log.Err(terror.ErrInvalidInput).Msgf("Asset %s is not in the queue", req.Payload.AssetHash)
				return
			}

			targetWarMachine := wmq.WarMachines[index]

			// calc insurance amount
			insuranceCost := server.BigInt{Int: *big.NewInt(0)}
			insuranceCost.Div(&targetWarMachine.ContractReward, big.NewInt(10))

			err = api.Passport.AssetInsurancePay(
				targetWarMachine.OwnedByID,
				targetWarMachine.FactionID,
				insuranceCost,
				server.TransactionReference(
					fmt.Sprintf(
						"pay_insurance_for_%s|%s",
						targetWarMachine.Name,
						time.Now(),
					),
				),
			)
			if err != nil {
				api.Log.Err(err).Msg(err.Error())
				return
			}

			// set isInsured flag to true
			targetWarMachine.IsInsured = true

			// update war machine copy in battle queue
			err = db.BattleQueueWarMachineUpdate(ctx, api.Conn, targetWarMachine)
			if err != nil {
				api.Log.Err(err).Msgf("failed to update war machine in db, token id: %s", req.Payload.AssetHash)
				return
			}

			// broadcast war machine queue
			warMachineQueuePosition := []*passport.WarMachineQueuePosition{}
			for i, wm := range wmq.WarMachines {
				if wm.OwnedByID != targetWarMachine.OwnedByID {
					continue
				}
				warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
					WarMachineMetadata: wm,
					Position:           i,
				})
			}

			go api.Passport.WarMachineQueuePositionBroadcast([]*passport.UserWarMachineQueuePosition{
				{
					UserID:                   targetWarMachine.OwnedByID,
					WarMachineQueuePositions: warMachineQueuePosition,
				},
			})

		}
	}
}

type UserSupsMultiplierGetRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID    server.UserID `json:"userID"`
		SessionID hub.SessionID `json:"sessionID"`
	} `json:"payload"`
}

func (api *API) PassportUserSupsMultiplierGetHandler(ctx context.Context, payload []byte) {
	req := &UserSupsMultiplierGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport faction stat get request")
		return
	}

	// api.ClientSupsMultipliersGet(req.Payload.UserID)

	api.UserMultiplier.PushUserMultiplierToPassport(req.Payload.UserID)
}

type UserStatGetRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID    server.UserID `json:"userID"`
		SessionID hub.SessionID `json:"sessionID"`
	} `json:"payload"`
}

func (api *API) PassportUserStatGetHandler(ctx context.Context, payload []byte) {
	req := &UserStatGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling user stat get request")
		return
	}

	if req.Payload.UserID.IsNil() {
		api.Log.Err(err).Msg("User id is required")
		return
	}

	userStat, err := db.UserStatGet(ctx, api.Conn, req.Payload.UserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		api.Log.Err(err).Msg("Failed to get user stat")
		return
	}

	if userStat == nil {
		// build a empty user stat if there is no user stat in db
		userStat = &server.UserStat{
			ID:                    req.Payload.UserID,
			ViewBattleCount:       0,
			TotalVoteCount:        0,
			TotalAbilityTriggered: 0,
			KillCount:             0,
		}
	}

	api.Passport.UserStatSend(ctx, []*passport.UserStatSend{
		{
			ToUserSessionID: &req.Payload.SessionID,
			Stat:            userStat,
		},
	})

}

type FactionStatGetRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID    *server.UserID   `json:"userID,omitempty"`
		SessionID *hub.SessionID   `json:"sessionID,omitempty"`
		FactionID server.FactionID `json:"factionID"`
	} `json:"payload"`
}

func (api *API) PassportFactionStatGetHandler(ctx context.Context, payload []byte) {
	req := &FactionStatGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport faction stat get request")
		return
	}

	if req.Payload.FactionID.IsNil() {
		api.Log.Err(terror.ErrInvalidInput).Msg("Faction id is empty")
		return
	}

	factionStat := &server.FactionStat{
		ID: req.Payload.FactionID,
	}

	err = db.FactionStatGet(ctx, api.Conn, factionStat)
	if err != nil {
		api.Log.Err(err).Msgf("Failed to get faction %s stat", req.Payload.FactionID)
		return
	}

	err = api.Passport.FactionStatsSend(ctx, []*passport.FactionStatSend{
		{
			FactionStat:     factionStat,
			ToUserID:        req.Payload.UserID,
			ToUserSessionID: req.Payload.SessionID,
		},
	})
	if err != nil {
		api.Log.Err(err).Msgf("Failed to send faction %s stat", req.Payload.FactionID)
		return
	}

}

type WarMachineQueuePositionRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		FactionID server.FactionID `json:"factionID"`
		UserID    server.UserID    `json:"userID"`
	} `json:"payload"`
}

// PassportWarMachineQueuePositionHandler return the list of user's war machines in the queue
func (api *API) PassportWarMachineQueuePositionHandler(ctx context.Context, payload []byte) {
	req := &WarMachineQueuePositionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport battle queue position")
		return
	}

	warMachineQueuePosition := []*passport.WarMachineQueuePosition{}

	api.BattleArena.BattleQueueMap[req.Payload.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
		for i, wm := range wmq.WarMachines {
			if wm.OwnedByID != req.Payload.UserID {
				continue
			}
			warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
				WarMachineMetadata: wm,
				Position:           i,
			})
		}
	}

	// get in game war machine
	for _, wm := range api.BattleArena.InGameWarMachines() {
		if wm.OwnedByID != req.Payload.UserID {
			continue
		}
		warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
			WarMachineMetadata: wm,
			Position:           -1,
		})
	}

	// fire a war machine queue passport request
	if len(warMachineQueuePosition) > 0 {
		go api.Passport.WarMachineQueuePositionBroadcast([]*passport.UserWarMachineQueuePosition{
			{
				UserID:                   req.Payload.UserID,
				WarMachineQueuePositions: warMachineQueuePosition,
			},
		})
	}

}

type AuthRingCheckRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		User                *server.User  `json:"user"`
		SessionID           hub.SessionID `json:"sessionID"`
		GameserverSessionID string        `json:"gameserverSessionID"`
	} `json:"payload"`
}

func (api *API) AuthRingCheckHandler(ctx context.Context, payload []byte) {
	req := &AuthRingCheckRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport auth ring check")
		return
	}

	if req.Payload.GameserverSessionID == "" {
		api.Log.Err(fmt.Errorf("No auth ring check key is provided"))
		return
	}

	client, err := api.RingCheckAuthMap.Check(req.Payload.GameserverSessionID)
	if err != nil {
		api.Log.Err(err)
		return
	}

	client.SetIdentifier(req.Payload.User.ID.String())

	// go api.ClientOnline(client)
	api.UserMultiplier.Online(req.Payload.User.ID)

	if !req.Payload.User.FactionID.IsNil() {
		api.viewerLiveCount.Swap(server.FactionID(uuid.Nil), req.Payload.User.FactionID)
		req.Payload.User.Faction = api.factionMap[req.Payload.User.FactionID]
	}

	api.UserMap.UserRegister(client, req.Payload.User)

	// send user id and faction id back to twitch ui client
	resp := struct {
		Key           hub.HubCommandKey `json:"key"`
		TransactionID string            `json:"transactionID"`
		Payload       interface{}       `json:"payload"`
	}{
		Key:           HubKeyUserSubscribe,
		TransactionID: "authRingCheck",
		Payload:       req.Payload.User,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		api.Hub.Log.Err(err).Errorf("send: issue marshalling resp")
		return
	}

	go client.Send(b)

	// send request to passport server to upgrade the gamebar user
	api.Passport.UpgradeUserConnection(req.Payload.SessionID)
}
