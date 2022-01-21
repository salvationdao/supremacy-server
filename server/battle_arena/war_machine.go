package battle_arena

import (
	"context"
	"encoding/json"
	"errors"
	"server"
	"server/db"

	"github.com/jackc/pgx/v4"

	"github.com/ninja-software/terror/v2"
)

const WarMachineDestroyedCommand = BattleCommand("BATTLE:WAR_MACHINE_DESTROYED")

type WarMachineDestroyedRequest struct {
	Payload struct {
		BattleID                 server.BattleID                  `json:"battleID"`
		DestroyedWarMachineEvent *server.WarMachineDestroyedEvent `json:"destroyedWarMachineEvent"`
	} `json:"payload"`
}

func (ba *BattleArena) WarMachineDestroyedHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &WarMachineDestroyedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Battle Update: %s - War Machine Destroyed: %d", req.Payload.BattleID, req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineID)

	// save to database
	tx, err := ba.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ba.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	assistedWarMachineIDs := req.Payload.DestroyedWarMachineEvent.AssistedWarMachineIDs

	err = db.WarMachineDestroyedEventCreate(ctx, tx, req.Payload.BattleID, req.Payload.DestroyedWarMachineEvent)
	if err != nil {
		return terror.Error(err)
	}

	if len(assistedWarMachineIDs) > 0 {
		err = db.WarMachineDestroyedEventAssistedWarMachineSet(ctx, tx, req.Payload.DestroyedWarMachineEvent.ID, assistedWarMachineIDs)
		if err != nil {
			return terror.Error(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	// send event to hub clients
	ba.Events.Trigger(ctx, EventWarMachineDestroyed, &EventData{
		WarMachineDestroyedEvent: req.Payload.DestroyedWarMachineEvent,
	})
	return nil
}
