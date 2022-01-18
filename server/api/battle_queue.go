package api

import (
	"context"
	"fmt"
	"server"
	"server/passport"

	"github.com/ninja-software/hub/v2/ext/messagebus"
)

const MaxInGameWarmachinePerFaction = 3

type warMachineQueuingList struct {
	WarMachines []*server.WarMachineNFT
}

func (api *API) startBattleQueue(factionID server.FactionID) {
	warMachineNFTs := &warMachineQueuingList{
		WarMachines: []*server.WarMachineNFT{},
	}

	go func() {
		for fn := range api.battleQueueMap[factionID] {
			fn(warMachineNFTs)
		}
	}()
}

func (api *API) GetBattleWarMachineFromQueue(factionID server.FactionID) []*server.WarMachineNFT {
	inGameWarMachinesChan := make(chan []*server.WarMachineNFT)

	api.battleQueueMap[factionID] <- func(wmq *warMachineQueuingList) {
		tempList := []*server.WarMachineNFT{}
		// if queuing war machines is less than maximum in game war machine amount
		if len(wmq.WarMachines) <= MaxInGameWarmachinePerFaction {

			// get all the the war machines
			tempList = append(tempList, wmq.WarMachines...)

			// clear up the queuing list
			wmq.WarMachines = []*server.WarMachineNFT{}

			// TODO: add default war machine to meet the total amount

			// broadcast next 5 queuing war machines to twitch ui
			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, factionID)), []*server.WarMachineNFT{})

			// broadcast empty queue for all the passport clients
			api.Passport.WarMachineQueuePositionClear(context.Background(), fmt.Sprintf("war_machine_position_clear_%s", factionID), factionID)

			inGameWarMachinesChan <- tempList
			return
		}

		// get first 3 war machines
		for i := 0; i < MaxInGameWarmachinePerFaction; i++ {
			tempList = append(tempList, wmq.WarMachines[i])
		}

		// delete it from the queue list
		wmq.WarMachines = wmq.WarMachines[MaxInGameWarmachinePerFaction-1:]

		// broadcast next 5 queuing war machines to twitch ui
		maxLength := 5
		if len(wmq.WarMachines) < maxLength {
			maxLength = len(wmq.WarMachines)
		}

		api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, factionID)), wmq.WarMachines[:maxLength])

		// broadcast war machine queue position update
		api.Passport.WarMachineQueuePosition(context.Background(), "war_machine_queue_position", BuildUserWarMachineQueuePosition(wmq.WarMachines))

		// return the war machines
		inGameWarMachinesChan <- tempList
	}

	return <-inGameWarMachinesChan
}

func BuildUserWarMachineQueuePosition(wmn []*server.WarMachineNFT) []*passport.UserWarMachineQueuePosition {
	result := []*passport.UserWarMachineQueuePosition{}
	queuePositionMap := make(map[server.UserID][]*passport.WarMachineQueuePosition)
	for i, wm := range wmn {
		qp, ok := queuePositionMap[wm.OwnedByID]
		if !ok {
			qp = []*passport.WarMachineQueuePosition{}
		}

		qp = append(qp, &passport.WarMachineQueuePosition{
			WarMachineNFT: wm,
			Position:      i,
		})
		queuePositionMap[wm.OwnedByID] = qp
	}

	if len(queuePositionMap) == 0 {
		return []*passport.UserWarMachineQueuePosition{}
	}

	for userID, qp := range queuePositionMap {
		result = append(result, &passport.UserWarMachineQueuePosition{
			UserID:                   userID,
			WarMachineQueuePositions: qp,
		})
	}

	return result
}
