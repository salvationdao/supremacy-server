package battle_arena

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/passport"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

type WarMachineQueue struct {
	RedMountain *FactionQueue
	Boston      *FactionQueue
	Zaibatsu    *FactionQueue
	log         *zerolog.Logger
}

type FactionQueue struct {
	*sync.Mutex
	Conn        *pgxpool.Pool
	WarMachines []*server.WarMachineMetadata

	defaultWarMachines []*server.WarMachineMetadata
	log                *zerolog.Logger
}

func NewWarMachineQueue(factions []*server.Faction, conn *pgxpool.Pool, log *zerolog.Logger, ba *BattleArena) *WarMachineQueue {
	wmq := &WarMachineQueue{
		RedMountain: &FactionQueue{&sync.Mutex{}, conn, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, log_helpers.NamedLogger(log, "Red Mountain queue")},
		Boston:      &FactionQueue{&sync.Mutex{}, conn, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, log_helpers.NamedLogger(log, "Boston queue")},
		Zaibatsu:    &FactionQueue{&sync.Mutex{}, conn, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, log_helpers.NamedLogger(log, "Zaibatsu queue")},
		log:         log_helpers.NamedLogger(log, "war machine queue"),
	}

	for _, faction := range factions {
		switch faction.ID {

		// initialise Red Mountain war machine queue
		case server.RedMountainFactionID:

			wmq.RedMountain.defaultWarMachines = ba.FillDefaultWarMachines(faction.ID, 3)
			wmq.RedMountain.Init(faction)

			// initialise Boston war machine queue
		case server.BostonCyberneticsFactionID:
			wmq.Boston.defaultWarMachines = ba.FillDefaultWarMachines(faction.ID, 3)
			wmq.Boston.Init(faction)

			// initialise Zaibatsu war machine queue
		case server.ZaibatsuFactionID:
			wmq.Zaibatsu.defaultWarMachines = ba.FillDefaultWarMachines(faction.ID, 3)
			wmq.Zaibatsu.Init(faction)
		}
	}

	return wmq
}

func (wmq *WarMachineQueue) WarMachineJoin(wmm *server.WarMachineMetadata) {
	// check faction id
	switch wmm.FactionID {
	case server.RedMountainFactionID:
		wmq.RedMountain.Join(wmm)
	case server.BostonCyberneticsFactionID:
		wmq.Boston.Join(wmm)
	case server.ZaibatsuFactionID:
		wmq.Zaibatsu.Join(wmm)
	default:
		wmq.log.Err(fmt.Errorf("No faction war machine")).Msg("NON-FACTION WAR MACHINE IS NOT ALLOWED!!!!!!!!!!!!!!!!!!!")
	}
}

// Init read war machine list from db and set up the list
func (fq *FactionQueue) Init(faction *server.Faction) {
	// read war machine queue from db
	wms, err := db.BattleQueueRead(context.Background(), fq.Conn, faction.ID)
	if err != nil {
		fq.log.Err(err).Msg("failed to read battle queue list from db")
	}

	// chuck war machines into list
	fq.WarMachines = wms

	// set up war machines' faction detail
	for _, wm := range fq.WarMachines {
		wm.Faction = faction
	}
}

func (fq *FactionQueue) Join(wmm *server.WarMachineMetadata) {
	// reject queue if already in the queue
	if index := checkWarMachineExist(fq.WarMachines, wmm.Hash); index != -1 {
		fq.log.Err(fmt.Errorf("war machine already in the queue")).Msgf("war machine %s is already in queue", wmm.Hash)
		return
	}

	// join war machine to queue
	fq.Lock()
	fq.WarMachines = append(fq.WarMachines, wmm)
	fq.Unlock()

	// insert war machine into db
	err := db.BattleQueueInsert(context.Background(), fq.Conn, wmm)
	if err != nil {
		fq.log.Err(err).Msgf("Failed to insert a copy of queue in db, token id: %s", wmm.Hash)
		return
	}
}

func (fq *FactionQueue) EnterGame(desireAmount int) []*server.WarMachineMetadata {
	newList := []*server.WarMachineMetadata{}
	fq.Lock()
	defer fq.Unlock()

	if len(fq.WarMachines) < desireAmount {
		newList = append(newList, fq.WarMachines...)

		// newList = append(newList)
		return newList
	}

	fq.WarMachines = fq.WarMachines[desireAmount-1:]

	return newList
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

func (ba *BattleArena) FillDefaultWarMachines(factionID server.FactionID, amount int) []*server.WarMachineMetadata {
	warMachines := []*server.WarMachineMetadata{}
	// add default war machine to meet the total amount
	for len(warMachines) < amount {
		wg := sync.WaitGroup{}
		wg.Add(1)

		ba.passport.GetDefaultWarMachines(context.Background(), factionID, amount, func(msg []byte) {
			defer wg.Done()
			resp := struct {
				WarMachines []*server.WarMachineMetadata `json:"payload"`
			}{}
			err := json.Unmarshal(msg, &resp)
			if err != nil {
				ba.Log.Err(err)
				return
			}
			spew.Dump(resp.WarMachines)
			warMachines = append(warMachines, resp.WarMachines...)
		})
		wg.Wait()
		time.Sleep(200 * time.Microsecond)
	}
	return warMachines
}

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
	ba.BattleQueueMap[factionID] <- func(wmq *WarMachineQueuingList) {
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

				wg := sync.WaitGroup{}
				wg.Add(1)

				ba.passport.GetDefaultWarMachines(ctx, factionID, amountToGet, func(msg []byte) {
					defer wg.Done()
					resp := struct {
						WarMachines []*server.WarMachineMetadata `json:"payload"`
					}{}
					err := json.Unmarshal(msg, &resp)
					if err != nil {
						ba.Log.Err(err)
						return
					}
					spew.Dump(resp.WarMachines)
					tempList = append(tempList, resp.WarMachines...)
				})
				wg.Wait()
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
			go ba.passport.WarMachineQueuePositionBroadcast(ba.BuildUserWarMachineQueuePosition(wmq.WarMachines, tempList, includedUserID...))

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
		go ba.passport.WarMachineQueuePositionBroadcast(ba.BuildUserWarMachineQueuePosition(wmq.WarMachines, tempList, includedUserID...))

		// return the war machines
		inGameWarMachinesChan <- tempList
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
