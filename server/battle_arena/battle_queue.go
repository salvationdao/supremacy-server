package battle_arena

import (
	"context"
	"fmt"
	"server"
	"server/passport"
	"time"
)

const MaxInGameWarmachinePerFaction = 2

type WarMachineQueuingList struct {
	WarMachines []*server.WarMachineNFT
}

func (ba *BattleArena) startBattleQueue(factionID server.FactionID) {
	warMachineNFTs := &WarMachineQueuingList{
		WarMachines: []*server.WarMachineNFT{},
	}

	for {
		fn := <-ba.BattleQueueMap[factionID]
		fn(warMachineNFTs)
	}
}

func (ba *BattleArena) GetBattleWarMachineFromQueue(factionID server.FactionID) []*server.WarMachineNFT {
	inGameWarMachinesChan := make(chan []*server.WarMachineNFT)
	ba.BattleQueueMap[factionID] <- func(wmq *WarMachineQueuingList) {
		ctx := context.Background()
		tempList := []*server.WarMachineNFT{}
		// if queuing war machines is less than maximum in game war machine amount get all and fill rest with defaults
		if len(wmq.WarMachines) <= MaxInGameWarmachinePerFaction {
			// get all the war machines
			tempList = append(tempList, wmq.WarMachines...)

			// clear up the queuing list
			wmq.WarMachines = []*server.WarMachineNFT{}

			// add default war machine to meet the total amount
			for len(tempList) < MaxInGameWarmachinePerFaction {
				amountToGet := MaxInGameWarmachinePerFaction - len(tempList)
				result, err := ba.passport.GetDefaultWarMachines(ctx, factionID, amountToGet)
				if err != nil {
					ba.Log.Err(err).Msg("issue getting default war machines")
					// TODO: figure how what to do if this errors
				}

				tempList = append(tempList, result...)
				time.Sleep(2 * time.Second)
			}

			// broadcast next 5 queuing war machines to twitch ui
			// api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, factionID)), []*server.WarMachineNFT{})

			// broadcast empty queue for all the passport client
			ba.passport.WarMachineQueuePositionClear(context.Background(), fmt.Sprintf("war_machine_position_clear_%s", factionID), factionID)

			inGameWarMachinesChan <- tempList
			return
		}

		// get first war machines up to the max
		for i := 0; i < MaxInGameWarmachinePerFaction; i++ {
			tempList = append(tempList, wmq.WarMachines[i])
		}

		// delete it from the queue list
		wmq.WarMachines = wmq.WarMachines[MaxInGameWarmachinePerFaction-1:]

		//// broadcast next 5 queuing war machines to twitch ui
		//maxLength := 5
		//if len(wmq.WarMachines) < maxLength {
		//	maxLength = len(wmq.WarMachines)
		//}

		//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, factionID)), wmq.WarMachines[:maxLength])

		// broadcast war machine queue position update
		ba.passport.WarMachineQueuePositionBroadcast(context.Background(), BuildUserWarMachineQueuePosition(wmq.WarMachines))

		// return the war machines
		inGameWarMachinesChan <- tempList
	}

	return <-inGameWarMachinesChan
}

func BuildUserWarMachineQueuePosition(wmn []*server.WarMachineNFT) []*passport.UserWarMachineQueuePosition {
	var result []*passport.UserWarMachineQueuePosition
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
