package battle_arena

import (
	"context"
	"server"
	"server/db"
	"time"

	"github.com/ninja-software/terror/v2"
)

func (ba *BattleArena) SetFactionMap(factionMap map[server.FactionID]server.Faction) {
	ba.battle.FactionMap = factionMap
}

// FactionAbilitiesQuery return 3 random abilities from faction
func (ba *BattleArena) FactionAbilitiesQuery(factionID server.FactionID) ([]*server.FactionAbility, error) {
	factionAbilities, err := db.FactionAbilityGetRandom(ba.ctx, ba.Conn, factionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return factionAbilities, nil
}

type AbilityTriggerRequest struct {
	FactionID         server.FactionID
	FactionAbilityID  server.FactionAbilityID
	IsSuccess         bool
	TriggeredByUserID *string
	TriggeredOnCellX  *int
	TriggeredOnCellY  *int
}

func (ba *BattleArena) FactionAbilityTrigger(atr *AbilityTriggerRequest) error {
	ctx := context.Background()
	factionAbilityEvent := &server.FactionAbilityEvent{
		FactionAbilityID:  atr.FactionAbilityID,
		IsTriggered:       atr.IsSuccess,
		TriggeredByUserID: atr.TriggeredByUserID,
		TriggeredOnCellX:  atr.TriggeredOnCellX,
		TriggeredOnCellY:  atr.TriggeredOnCellY,
	}

	err := db.FactionAbilityEventCreate(ctx, ba.Conn, ba.battle.ID, factionAbilityEvent)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func (ba *BattleArena) FakeWarMachinePositionUpdate() {
	i := 1
	for {
		for _, warMachine := range ba.battle.WarMachines {
			// do update
			scale := 1
			if i%2 == 0 {
				scale = -1
			}

			warMachine.Rotation = i % 360
			warMachine.Position.X += 10 * scale
			warMachine.Position.Y -= 10 * scale
		}

		// broadcast
		ba.Events.Trigger(context.Background(), EventWarMachinePositionChanged, &EventData{
			BattleArena: ba.battle,
		})

		time.Sleep(250 * time.Millisecond)
		i++
	}

}
