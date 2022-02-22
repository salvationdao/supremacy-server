package battle_arena

import (
	"context"
	"errors"
	"server"
	"server/db"
	"server/passport"
	"time"
)

type WarMachineQueuingList struct {
	WarMachines []*server.WarMachineMetadata
}

func (ba *BattleArena) startBattleQueue(factionID server.FactionID) {
	warMachineMetadatas := &WarMachineQueuingList{
		WarMachines: []*server.WarMachineMetadata{},
	}

	// read war machine queue from db
	wms, err := db.BattleQueueRead(ba.ctx, ba.Conn, factionID)
	if err != nil {
		ba.Log.Err(err).Msg("failed to read battle queue list from db")
	}

	// assign war machine list
	warMachineMetadatas.WarMachines = wms

	go func() {
		for {
			fn := <-ba.BattleQueueMap[factionID]
			fn(warMachineMetadatas)
		}
	}()
}

func (ba *BattleArena) GetBattleWarMachineFromQueue(factionID server.FactionID, warMachinePerBattle int) []*server.WarMachineMetadata {
	inGameWarMachinesChan := make(chan []*server.WarMachineMetadata)
	select {
	case ba.BattleQueueMap[factionID] <- func(wmq *WarMachineQueuingList) {
		ctx := context.Background()
		tempList := []*server.WarMachineMetadata{}

		// if queuing war machines is less than maximum in game war machine amount get all and fill rest with defaults
		if len(wmq.WarMachines) <= warMachinePerBattle {
			// get all the war machines
			tempList = append(tempList, wmq.WarMachines...)

			// cache included user
			includedUserID := []server.UserID{}
			for _, tl := range tempList {
				includedUserID = append(includedUserID, tl.OwnedByID)
			}

			// clear up the queuing list
			wmq.WarMachines = []*server.WarMachineMetadata{}

			// add default war machine to meet the total amount
			for len(tempList) < warMachinePerBattle {
				amountToGet := warMachinePerBattle - len(tempList)
				result, err := ba.passport.GetDefaultWarMachines(ctx, factionID, amountToGet)
				if err != nil {
					ba.Log.Err(err).Msg("issue getting default war machines")
					// TODO: figure how what to do if this errors
				}

				tempList = append(tempList, result...)
				time.Sleep(200 * time.Microsecond)
			}

			// broadcast next 5 queuing war machines to game ui
			ba.Events.Trigger(context.Background(), EventWarMachineQueueUpdated, &EventData{
				WarMachineQueue: &WarMachineQueueUpdateEvent{
					FactionID:   factionID,
					WarMachines: []*server.WarMachineMetadata{},
				},
			})

			// broadcast empty queue for all the passport client
			go ba.passport.WarMachineQueuePositionBroadcast(context.Background(), ba.BuildUserWarMachineQueuePosition(wmq.WarMachines, tempList, includedUserID...))

			inGameWarMachinesChan <- tempList
			return
		}

		// get first war machines up to the max
		for i := 0; i < warMachinePerBattle; i++ {
			tempList = append(tempList, wmq.WarMachines[i])
		}

		// cache included user
		includedUserID := []server.UserID{}
		for _, tl := range tempList {
			includedUserID = append(includedUserID, tl.OwnedByID)
		}

		// delete it from the queue list
		wmq.WarMachines = wmq.WarMachines[warMachinePerBattle-1:]

		// broadcast next 5 queuing war machines to twitch ui
		maxLength := 5
		if len(wmq.WarMachines) < maxLength {
			maxLength = len(wmq.WarMachines)
		}

		ba.Events.Trigger(context.Background(), EventWarMachineQueueUpdated, &EventData{
			WarMachineQueue: &WarMachineQueueUpdateEvent{
				FactionID:   factionID,
				WarMachines: wmq.WarMachines[:maxLength],
			},
		})

		// broadcast war machine queue position update
		go ba.passport.WarMachineQueuePositionBroadcast(context.Background(), ba.BuildUserWarMachineQueuePosition(wmq.WarMachines, tempList, includedUserID...))

		// return the war machines
		inGameWarMachinesChan <- tempList
	}:

	case <-time.After(10 * time.Second):
		ba.Log.Err(errors.New("timeout on channel send exceeded"))
		panic("Client Battle Reward Update")
	}

	return <-inGameWarMachinesChan
}

func (ba *BattleArena) BuildUserWarMachineQueuePosition(queuingList []*server.WarMachineMetadata, pendingList []*server.WarMachineMetadata, mustIncludeUserIDs ...server.UserID) []*passport.UserWarMachineQueuePosition {
	result := []*passport.UserWarMachineQueuePosition{}
	queuePositionMap := make(map[server.UserID][]*passport.WarMachineQueuePosition)

	// from queuing list
	for i, wm := range queuingList {
		qp, ok := queuePositionMap[wm.OwnedByID]
		if !ok {
			qp = []*passport.WarMachineQueuePosition{}
		}

		qp = append(qp, &passport.WarMachineQueuePosition{
			WarMachineMetadata: wm,
			Position:           i,
		})
		queuePositionMap[wm.OwnedByID] = qp
	}

	// from pending list
	for _, wm := range pendingList {
		qp, ok := queuePositionMap[wm.OwnedByID]
		if !ok {
			qp = []*passport.WarMachineQueuePosition{}
		}

		qp = append(qp, &passport.WarMachineQueuePosition{
			WarMachineMetadata: wm,
			Position:           -1,
		})
		queuePositionMap[wm.OwnedByID] = qp
	}

	// from in game list
	for _, wm := range ba.battle.WarMachines {
		qp, ok := queuePositionMap[wm.OwnedByID]
		if !ok {
			qp = []*passport.WarMachineQueuePosition{}
		}

		qp = append(qp, &passport.WarMachineQueuePosition{
			WarMachineMetadata: wm,
			Position:           -1,
		})

		queuePositionMap[wm.OwnedByID] = qp
	}

	if len(queuePositionMap) > 0 {
		for userID, qp := range queuePositionMap {
			result = append(result, &passport.UserWarMachineQueuePosition{
				UserID:                   userID,
				WarMachineQueuePositions: qp,
			})
		}
	}

	if len(mustIncludeUserIDs) > 0 {
		for _, uid := range mustIncludeUserIDs {
			exists := false
			for _, r := range result {
				if r.UserID == uid {
					exists = true
				}
			}
			if !exists {
				result = append(result, &passport.UserWarMachineQueuePosition{
					UserID:                   uid,
					WarMachineQueuePositions: []*passport.WarMachineQueuePosition{},
				})
			}
		}
	}

	return result
}
