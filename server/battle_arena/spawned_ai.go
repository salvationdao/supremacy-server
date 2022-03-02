package battle_arena

import (
	"context"
	"encoding/json"
	"fmt"
	"server"

	"github.com/ninja-software/terror/v2"
)

const AISpawnedCommand = BattleCommand("BATTLE:AI_SPAWNED")

type AISpawnedRequest struct {
	Payload struct {
		BattleID       server.BattleID        `json:"battleID"`
		SpawnedAIEvent *server.SpawnedAIEvent `json:"spawnedAIEvent"`
	} `json:"payload"`
}

func (ba *BattleArena) AISpawnedHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &AISpawnedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	// check battle id
	if req.Payload.BattleID != ba.battle.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", ba.battle.ID.String(), req.Payload.BattleID.String()))
	}

	if req.Payload.SpawnedAIEvent == nil {
		return terror.Error(fmt.Errorf("missing Spawned AI event"))
	}

	// get spawned AI
	spawnedAI := server.WarMachineMetadata{
		ParticipantID: req.Payload.SpawnedAIEvent.ParticipantID,
		Name:          req.Payload.SpawnedAIEvent.Name,
		Model:         req.Payload.SpawnedAIEvent.Model,
		Skin:          req.Payload.SpawnedAIEvent.Skin,
		MaxHealth:     req.Payload.SpawnedAIEvent.MaxHealth,
		Health:        req.Payload.SpawnedAIEvent.MaxHealth,
		MaxShield:     req.Payload.SpawnedAIEvent.MaxShield,
		Shield:        req.Payload.SpawnedAIEvent.MaxShield,
		FactionID:     req.Payload.SpawnedAIEvent.FactionID,
		Position:      req.Payload.SpawnedAIEvent.Position,
		Rotation:      req.Payload.SpawnedAIEvent.Rotation,
	}

	aiFaction, ok := ba.battle.FactionMap[spawnedAI.FactionID]
	if ok {
		spawnedAI.Faction = aiFaction
	}

	ba.Log.Info().Msgf("Battle Update: %s - AI Spawned: %d", req.Payload.BattleID, spawnedAI.ParticipantID)

	// cache record in battle, for future subscription
	ba.battle.SpawnedAI = append(ba.battle.SpawnedAI, &spawnedAI)

	// send event to hub clients
	ba.Events.Trigger(ctx, EventAISpawned, &EventData{
		SpawnedAI: &spawnedAI,
	})

	return nil
}
