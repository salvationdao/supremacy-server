package battle_arena

import (
	"context"
	"server"
	"server/db"
	"time"

	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
)

const BattleCommandInitBattle BattleCommand = "BATTLE:INIT"

func (ba *BattleArena) InitNextBattle() error {
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

	for len(ba.BattleQueueMap) == 0 {
		ba.Log.Info().Msg("No factions, trying again in 2 seconds")
		time.Sleep(2 * time.Second)
	}
	mechsPerFaction := gameMap.MaxSpawns / 3

	for factionID := range ba.BattleQueueMap {
		ba.battle.WarMachines = append(ba.battle.WarMachines, ba.GetBattleWarMachineFromQueue(factionID, mechsPerFaction)...)
	}

	if len(ba.battle.WarMachines) > 0 {
		tokenIDs := []uint64{}
		for _, warMachine := range ba.battle.WarMachines {
			tokenIDs = append(tokenIDs, warMachine.TokenID)
		}

		// set war machine lock request
		err := ba.passport.AssetLock(ba.ctx, tokenIDs)
		if err != nil {
			ba.Log.Err(err).Msg("Failed to lock assets")
			// TODO: figure out how to handle this
		}
	}

	// clean up battle end message of the last battle
	ba.battle.EndedAt = nil
	ba.Events.Trigger(context.Background(), EventGameInit, nil)

	ba.Log.Info().Msgf("Initializing new battle: %s", ba.battle.ID)

	// Setup payload
	payload := struct {
		BattleID    server.BattleID              `json:"battleID"`
		MapName     string                       `json:"mapName"`
		WarMachines []*server.WarMachineMetadata `json:"warMachines"`
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

	ba.send <- gameMessage
	return nil
}
