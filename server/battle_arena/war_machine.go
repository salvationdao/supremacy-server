package battle_arena

import (
	"context"
	"encoding/json"
	"errors"
	"server"
	"server/db"

	"github.com/gofrs/uuid"

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

	if req.Payload.DestroyedWarMachineEvent.RelatedEventIDString != "" {
		relatedEventuuid, err := uuid.FromString(req.Payload.DestroyedWarMachineEvent.RelatedEventIDString)
		if err != nil {
			return terror.Error(err)
		}
		vid := server.EventID(relatedEventuuid)
		req.Payload.DestroyedWarMachineEvent.RelatedEventID = &vid
	}

	err = db.WarMachineDestroyedEventCreate(ctx, tx, req.Payload.BattleID, req.Payload.DestroyedWarMachineEvent)
	if err != nil {
		return terror.Error(err)
	}

	// TODO: MAKE TREAD SAFE
	for _, wm := range ba.battle.WarMachines {
		if wm.TokenID == req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineID {
			wm.RemainingHitPoints = 0
		}
	}

	// TODO: Add kill assists
	//if len(assistedWarMachineIDs) > 0 {
	//	err = db.WarMachineDestroyedEventAssistedWarMachineSet(ctx, tx, req.Payload.DestroyedWarMachineEvent.ID, assistedWarMachineIDs)
	//	if err != nil {
	//		return terror.Error(err)
	//	}
	//}

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

func (ba *BattleArena) WarMachinePositionUpdate(payload []byte) {
	ba.battle.BattleHistory = append(ba.battle.BattleHistory, payload)

	// broadcast
	ba.Events.Trigger(context.Background(), EventWarMachinePositionChanged, &EventData{
		BattleArena:        ba.battle,
		WarMachineLocation: payload,
	})
}

func (ba *BattleArena) WarMachineHitPointUpdate(payload []byte) {
	ba.battle.BattleHistory = append(ba.battle.BattleHistory, payload)

	// broadcast
	ba.Events.Trigger(context.Background(), EventWarMachineHitPointChanged, &EventData{
		BattleArena:        ba.battle,
		WarMachineHitPoint: payload,
	})
}
