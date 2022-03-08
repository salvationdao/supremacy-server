package battle_arena

import (
	"context"
	"server"
	"server/comms"
	"server/db"
	"server/passport"
	"time"

	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
)

const BattleCommandInitBattle BattleCommand = "BATTLE:INIT"

func (ba *BattleArena) InitNextBattle() error {
	ba.Events.Trigger(context.Background(), EventGameInit, nil)

	// switch battle state to LOBBY
	ba.battle.State = server.StateLobby

	// send new battle details to game client

	// generate a new battle event
	ba.battle.ID = server.BattleID(uuid.Must(uuid.NewV4()))

	// clean up war machine destroyed record cache
	for key := range ba.battle.WarMachineDestroyedRecordMap {
		delete(ba.battle.WarMachineDestroyedRecordMap, key)
	}

	// assign a random map
	gameMap, err := db.GameMapGetRandom(ba.ctx, ba.Conn)
	if err != nil {
		ba.Log.Err(err).Msg("")
		return terror.Error(err)
	}
	ba.battle.GameMap = gameMap
	ba.battle.GameMapID = gameMap.ID

	// get NFT from battle queue
	ba.battle.WarMachines = []*server.WarMachineMetadata{}

	for ba.battle == nil || len(ba.battle.FactionMap) == 0 {
		ba.Log.Info().Msg("No factions, trying again in 2 seconds")
		time.Sleep(2 * time.Second)
	}

	mechsPerFaction := gameMap.MaxSpawns / 3
	ba.battle.WarMachines = append(ba.battle.WarMachines, ba.WarMachineQueue.RedMountain.GetWarMachineForEnterGame(mechsPerFaction)...)
	ba.battle.WarMachines = append(ba.battle.WarMachines, ba.WarMachineQueue.Boston.GetWarMachineForEnterGame(mechsPerFaction)...)
	ba.battle.WarMachines = append(ba.battle.WarMachines, ba.WarMachineQueue.Zaibatsu.GetWarMachineForEnterGame(mechsPerFaction)...)

	// broadcast warmachine stat to passport
	broadcastList := []*comms.WarMachineQueueStat{}
	// Red mountain
	for i, wm := range ba.WarMachineQueue.RedMountain.QueuingWarMachines {
		position := i + 1
		broadcastList = append(broadcastList, &comms.WarMachineQueueStat{Hash: wm.Hash, Position: &position, ContractReward: wm.ContractReward})
	}
	for _, wm := range ba.WarMachineQueue.RedMountain.InGameWarMachines {
		position := -1
		broadcastList = append(broadcastList, &comms.WarMachineQueueStat{Hash: wm.Hash, Position: &position, ContractReward: wm.ContractReward})
	}
	ba.passport.FactionQueueCostUpdate(&passport.FactionQueuePriceUpdateReq{
		FactionID:     server.RedMountainFactionID,
		QueuingLength: ba.WarMachineQueue.RedMountain.QueuingLength(),
	})

	// release in game the mechs

	// Boston
	for i, wm := range ba.WarMachineQueue.Boston.QueuingWarMachines {
		position := i + 1
		broadcastList = append(broadcastList, &comms.WarMachineQueueStat{Hash: wm.Hash, Position: &position, ContractReward: wm.ContractReward})
	}
	for _, wm := range ba.WarMachineQueue.Boston.InGameWarMachines {
		position := -1
		broadcastList = append(broadcastList, &comms.WarMachineQueueStat{Hash: wm.Hash, Position: &position, ContractReward: wm.ContractReward})
	}
	ba.passport.FactionQueueCostUpdate(&passport.FactionQueuePriceUpdateReq{
		FactionID:     server.BostonCyberneticsFactionID,
		QueuingLength: ba.WarMachineQueue.Boston.QueuingLength(),
	})

	// Zaibatsu
	for i, wm := range ba.WarMachineQueue.Zaibatsu.QueuingWarMachines {
		position := i + 1
		broadcastList = append(broadcastList, &comms.WarMachineQueueStat{Hash: wm.Hash, Position: &position, ContractReward: wm.ContractReward})
	}
	for _, wm := range ba.WarMachineQueue.Zaibatsu.InGameWarMachines {
		position := -1
		broadcastList = append(broadcastList, &comms.WarMachineQueueStat{Hash: wm.Hash, Position: &position, ContractReward: wm.ContractReward})
	}
	ba.passport.FactionQueueCostUpdate(&passport.FactionQueuePriceUpdateReq{
		FactionID:     server.ZaibatsuFactionID,
		QueuingLength: ba.WarMachineQueue.Zaibatsu.QueuingLength(),
	})

	// broadcast position change
	ba.passport.WarMachineQueuePositionBroadcast(broadcastList)

	// get Zaibatsu faction abilities to insert
	zaibatsuAbility, err := db.GetZaibatsuFactionAbility(context.Background(), ba.Conn)
	if err != nil {
		ba.Log.Err(err).Msg("Unable to get zaibatsu faction ability")
		return terror.Error(err)
	}

	if len(ba.battle.WarMachines) > 0 {
		for _, warMachine := range ba.battle.WarMachines {
			// HACK: clean up war machine ability before stacking it
			warMachine.Abilities = []*server.AbilityMetadata{}
			if warMachine.FactionID == server.ZaibatsuFactionID {
				// if war machine is from Zaibatsu, insert the ability as faction ability
				warMachine.Abilities = append(warMachine.Abilities, &server.AbilityMetadata{
					ID:           zaibatsuAbility.ID,
					Identity:     uuid.Must(uuid.NewV4()), // track ability's price
					Colour:       zaibatsuAbility.Colour,
					TextColour:   zaibatsuAbility.TextColour,
					GameClientID: int(zaibatsuAbility.GameClientAbilityID),
					Image:        zaibatsuAbility.ImageUrl,
					Description:  zaibatsuAbility.Description,
					Name:         zaibatsuAbility.Label,
					SupsCost:     zaibatsuAbility.SupsCost,
				})
			}
		}
	}

	// clean up battle end message of the last battle
	ba.battle.EndedAt = nil

	ba.Log.Info().Msgf("Initializing new battle: %s", ba.battle.ID)

	// trunc war machine name before it is send to battle
	for _, wm := range ba.battle.WarMachines {
		if len(wm.Name) == 0 {
			wm.Name = wm.Hash
		} else if len(wm.Name) > 20 {
			wm.Name = wm.Name[:20]
		}
	}

	for _, wm := range ba.battle.WarMachines {
		fillFaction(wm)
	}

	// Setup payload
	payload := struct {
		BattleID    server.BattleID              `json:"battle_id"`
		MapName     string                       `json:"map_name"`
		WarMachines []*server.WarMachineMetadata `json:"war_machines"`
	}{
		BattleID:    ba.battle.ID,
		MapName:     ba.battle.GameMap.Name,
		WarMachines: ba.battle.WarMachines,
	}

	ctx, cancel := context.WithCancel(ba.ctx)
	gameMessage := &GameMessage{
		BattleCommand: BattleCommandInitBattle,
		Payload:       payload,
		context:       ctx,
		cancel:        cancel,
	}

	// NOTE: this will potentially lock game server if game client is disconnected
	// 		 so wrap it in a go routine
	go func() {
		ba.send <- gameMessage
	}()
	return nil
}

func fillFaction(wm *server.WarMachineMetadata) {
	switch wm.FactionID {
	case server.BostonCyberneticsFactionID:
		wm.Faction = BostonFaction
	case server.RedMountainFactionID:
		wm.Faction = RedMountainFaction
	case server.ZaibatsuFactionID:
		wm.Faction = ZaibatsuFaction
	}
}
