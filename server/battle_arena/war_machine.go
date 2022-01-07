package battle_arena

import (
	"context"
	"encoding/json"
	"gameserver"
	"gameserver/db"

	"github.com/ninja-software/terror/v2"
)

const WarMachineDestroyedCommand = BattleCommand("BATTLE:WAR_MACHINE_DESTROYED")

type WarMachineDestroyedRequest struct {
	Payload struct {
		BattleID                 gameserver.BattleID            `json:"battleId"`
		DestroyedWarMachineEvent gameserver.WarMachineDestroyed `json:"destroyedWarMachineEvent"`
	} `json:"payload"`
}

func (ba *BattleArena) WarMachineDestroyedHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &WarMachineDestroyedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Battle Update: %s - War Machine Destroyed: %s", req.Payload.BattleID, req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineID)

	// save to database
	tx, err := ba.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	err = db.WarMachineDestroyed(ctx, tx, req.Payload.BattleID, req.Payload.DestroyedWarMachineEvent)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	// send event to hub clients
	ba.Events.Trigger(ctx, EventWarMachineDestroyed, &EventData{
		WarMachineDestroyed: &req.Payload.DestroyedWarMachineEvent,
	})
	return nil
}
