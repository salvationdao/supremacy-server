package battle_arena

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
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

	// check battle id
	if req.Payload.BattleID != ba.battle.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", ba.battle.ID.String(), req.Payload.BattleID.String()))
	}

	// check destroyed war machine exist
	exists := false
	for _, wm := range ba.battle.WarMachines {
		if wm.TokenID == req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineID {
			// set health to 0
			wm.Health = 0
			exists = true
			break
		}
	}
	if !exists {
		return terror.Error(fmt.Errorf("destroyed war machine %d does not exist", req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineID))
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

func (ba *BattleArena) WarMachinesTick(payload []byte) {
	// Save to history
	ba.battle.BattleHistory = append(ba.battle.BattleHistory, payload)

	// broadcast
	ba.Events.Trigger(context.Background(), EventWarMachinePositionChanged, &EventData{
		BattleArena:        ba.battle,
		WarMachineLocation: payload,
	})

	// Update game settings (so new players get the latest position, health and shield of all warmachines)
	count := payload[1]
	var c byte
	offset := 2
	for c = 0; c < count; c++ {
		participantID := payload[offset]
		offset++

		// Get Warmachine Index
		warMachineIndex := -1
		for i, wmn := range ba.battle.WarMachines {
			if wmn.ParticipantID == participantID {
				warMachineIndex = i
				break
			}
		}

		// Get Sync byte (tells us which data was updated for this warmachine)
		syncByte := payload[offset]
		offset++

		// Position + Yaw
		if syncByte >= 100 {
			x := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			y := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			rotation := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4

			if warMachineIndex != -1 {
				if ba.battle.WarMachines[warMachineIndex].Position == nil {
					ba.battle.WarMachines[warMachineIndex].Position = &server.Vector3{}
				}
				ba.battle.WarMachines[warMachineIndex].Position.X = x
				ba.battle.WarMachines[warMachineIndex].Position.X = y
				ba.battle.WarMachines[warMachineIndex].Rotation = rotation
			}
		}
		// Health
		if syncByte == 1 || syncByte == 11 || syncByte == 101 || syncByte == 111 {
			health := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			if warMachineIndex != -1 {
				ba.battle.WarMachines[warMachineIndex].Health = health
			}
		}
		// Shield
		if syncByte == 10 || syncByte == 11 || syncByte == 110 || syncByte == 111 {
			shield := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			if warMachineIndex != -1 {
				ba.battle.WarMachines[warMachineIndex].Shield = shield
			}
		}
	}
}
