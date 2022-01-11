package battle_arena

import (
	"context"
	"server"
	"server/db"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

const BattleCommandInitBattle BattleCommand = "BATTLE:INIT"

func (ba *BattleArena) InitNextBattle() error {

	// generate a new battle event
	newBattle := &server.Battle{
		ID: server.BattleID(uuid.Must(uuid.NewV4())),
	}

	// assign a random map
	gameMap, err := db.GameMapGetRandom(ba.ctx, ba.Conn)
	if err != nil {
		return terror.Error(err)
	}
	newBattle.GameMap = gameMap

	// assign war machines
	// NOTE: there will be a queue system for users to register their NFTs to join the battle
	//       currently just randomly assign faction to warmachine
	warMachines, err := db.WarMachineAll(ba.ctx, ba.Conn)
	if err != nil {
		return terror.Error(err)
	}

	newBattle.WarMachines = warMachines

	ba.Log.Info().Msgf("Initializing new battle: %s", ba.battle.ID)

	// send new battle details to game client
	ctx, cancel := context.WithCancel(ba.ctx)

	gameMessage := &GameMessage{
		BattleCommand: BattleCommandInitBattle,
		Payload:       newBattle,
		context:       ctx,
		cancel:        cancel,
	}

	ba.send <- gameMessage
	return nil
}
