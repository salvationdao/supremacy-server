package battle_arena

import (
	"context"
	"gameserver"

	"github.com/gofrs/uuid"
)

const BattleCommandInitBattle BattleCommand = "BATTLE:INIT"

func (ba *BattleArena) InitNextBattle() error {

	// TODO: get the next battle details from ????
	newBattle := &gameserver.Battle{
		ID:          gameserver.BattleID(uuid.Must(uuid.NewV4())),
		WarMachines: ba.passport.GetWarMachines(),
		Map:         gameserver.FakeGameMaps[0],
	}
	ba.battle = newBattle

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
