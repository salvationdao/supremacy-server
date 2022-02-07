package battle_arena

import (
	"context"
	"server"
	"server/db"

	"github.com/ninja-software/terror/v2"
)

func (ba *BattleArena) SetFactionMap(factionMap map[server.FactionID]*server.Faction) {
	ba.battle.FactionMap = factionMap
}

// RandomAbilityCollection return random ability collection and faction ability map
func (ba *BattleArena) RandomAbilityCollection() (*server.BattleAbility, map[server.FactionID]*server.FactionAbility, error) {
	// get random collection
	battleAbility, err := db.AbilityCollectionGetRandom(ba.ctx, ba.Conn)
	if err != nil {
		return nil, nil, terror.Error(err)
	}

	// get abilities by collection id
	abilities, err := db.FactionAbilityGetByBattleAbilityID(ba.ctx, ba.Conn, battleAbility.ID)
	if err != nil {
		return nil, nil, terror.Error(err)
	}

	// build ability map
	factionAbilityMap := make(map[server.FactionID]*server.FactionAbility)
	for _, ability := range abilities {
		factionAbilityMap[ability.FactionID] = ability

		// set ability detail to battle ability
		battleAbility.Colour = ability.Colour
		battleAbility.ImageUrl = ability.ImageUrl
	}

	return battleAbility, factionAbilityMap, nil
}

const BattleAbilityCommand = BattleCommand("BATTLE:ABILITY")

type AbilityTriggerRequest struct {
	FactionID           server.FactionID
	FactionAbilityID    server.FactionAbilityID
	IsSuccess           bool
	TriggeredByUserID   *string
	TriggeredByUsername *string
	TriggeredOnCellX    *int
	TriggeredOnCellY    *int
	GameClientAbilityID byte
}

func (ba *BattleArena) FactionAbilityTrigger(atr *AbilityTriggerRequest) error {

	ctx := context.Background()
	factionAbilityEvent := &server.FactionAbilityEvent{
		FactionAbilityID:    atr.FactionAbilityID,
		IsTriggered:         atr.IsSuccess,
		TriggeredByUserID:   atr.TriggeredByUserID,
		TriggeredByUsername: atr.TriggeredByUsername,
		TriggeredOnCellX:    atr.TriggeredOnCellX,
		TriggeredOnCellY:    atr.TriggeredOnCellY,
		GameClientAbilityID: atr.GameClientAbilityID,
	}

	err := db.FactionAbilityEventCreate(ctx, ba.Conn, ba.battle.ID, factionAbilityEvent)
	if err != nil {
		return terror.Error(err)
	}

	// TODO: add possible counter animations in unreal
	if !factionAbilityEvent.IsTriggered {
		return nil
	}

	// Get the ability enum
	//fa, err := db.FactionAbilityGetRandom()

	// To get the location in game its
	//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
	//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels
	if factionAbilityEvent.TriggeredOnCellX != nil && factionAbilityEvent.TriggeredOnCellY != nil {
		factionAbilityEvent.GameLocation.X = ((*factionAbilityEvent.TriggeredOnCellX * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + ba.battle.GameMap.LeftPixels
		factionAbilityEvent.GameLocation.Y = ((*factionAbilityEvent.TriggeredOnCellY * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + ba.battle.GameMap.TopPixels
	}

	// send new battle details to game client
	ctx, cancel := context.WithCancel(ba.ctx)

	gameMessage := &GameMessage{
		BattleCommand: BattleAbilityCommand,
		Payload:       factionAbilityEvent,
		context:       ctx,
		cancel:        cancel,
	}

	ba.send <- gameMessage
	return nil
}
