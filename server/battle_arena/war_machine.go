package battle_arena

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/db"

	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
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
	var destroyedWarMachine *server.WarMachineMetadata
	for _, wm := range ba.battle.WarMachines {
		if wm.Hash == req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineHash {
			// set health to 0
			wm.Health = 0
			destroyedWarMachine = wm
			break
		}
	}
	if destroyedWarMachine == nil {
		return terror.Error(fmt.Errorf("destroyed war machine %s does not exist", req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineHash))
	}

	var killByWarMachine *server.WarMachineMetadata
	if req.Payload.DestroyedWarMachineEvent.KillByWarMachineHash != nil {
		for _, wm := range ba.battle.WarMachines {
			if wm.Hash == *req.Payload.DestroyedWarMachineEvent.KillByWarMachineHash {
				killByWarMachine = wm
			}
		}
		if destroyedWarMachine == nil {
			return terror.Error(fmt.Errorf("killer war machine %s does not exist", *req.Payload.DestroyedWarMachineEvent.KillByWarMachineHash))
		}
	}

	ba.Log.Info().Msgf("Battle Update: %s - War Machine Destroyed: %s", req.Payload.BattleID, req.Payload.DestroyedWarMachineEvent.DestroyedWarMachineHash)

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

	// prepare destroyed record
	destroyedRecord := &server.WarMachineDestroyedRecord{
		DestroyedWarMachine: destroyedWarMachine,
		KilledByWarMachine:  killByWarMachine,
		KilledBy:            req.Payload.DestroyedWarMachineEvent.KilledBy,
		DamageRecords:       []*server.DamageRecord{},
	}

	// calc total damage and merge the duplicated damage source
	totalDamage := 0
	newDamageHistory := []*server.DamageHistory{}
	for _, damage := range req.Payload.DestroyedWarMachineEvent.DamageHistory {
		totalDamage += damage.Amount
		// check instigator token id exist in the list
		if damage.InstigatorHash != "" {
			exists := false
			for _, hist := range newDamageHistory {
				if hist.InstigatorHash == damage.InstigatorHash {
					hist.Amount += damage.Amount
					exists = true
					break
				}
			}
			if !exists {
				newDamageHistory = append(newDamageHistory, &server.DamageHistory{
					Amount:         damage.Amount,
					InstigatorHash: damage.InstigatorHash,
					SourceName:     damage.SourceName,
					SourceHash:     damage.SourceHash,
				})
			}
			continue
		}
		// check source name
		exists := false
		for _, hist := range newDamageHistory {
			if hist.SourceName == damage.SourceName {
				hist.Amount += damage.Amount
				exists = true
				break
			}
		}
		if !exists {
			newDamageHistory = append(newDamageHistory, &server.DamageHistory{
				Amount:         damage.Amount,
				InstigatorHash: damage.InstigatorHash,
				SourceName:     damage.SourceName,
				SourceHash:     damage.SourceHash,
			})
		}
	}

	// get total damage amount for calculating percentage
	for _, damage := range newDamageHistory {
		damageRecord := &server.DamageRecord{
			SourceName: damage.SourceName,
			Amount:     (damage.Amount * 1000000 / totalDamage) / 100,
		}
		if damage.InstigatorHash != "" {
			for _, wm := range ba.battle.WarMachines {
				if wm.Hash == damage.InstigatorHash {
					damageRecord.CausedByWarMachine = wm
				}
			}
		}
		destroyedRecord.DamageRecords = append(destroyedRecord.DamageRecords, damageRecord)
	}

	// cache record in battle, for future subscription
	ba.battle.WarMachineDestroyedRecordMap[destroyedWarMachine.ParticipantID] = destroyedRecord

	// send event to hub clients
	ba.Events.Trigger(ctx, EventWarMachineDestroyed, &EventData{
		WarMachineDestroyedRecord: destroyedRecord,
	})

	return nil
}

func (ba *BattleArena) WarMachinesTick(ctx context.Context, payload []byte) {
	// Save to history
	ba.battle.BattleHistory = append(ba.battle.BattleHistory, payload)

	// broadcast
	ba.Events.Trigger(ctx, EventWarMachinePositionChanged, &EventData{
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
