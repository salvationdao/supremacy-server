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

// RandomBattleAbility return random ability collection and game ability map
func (ba *BattleArena) RandomBattleAbility() (*server.BattleAbility, map[server.FactionID]*server.GameAbility, error) {
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
	factionAbilityMap := make(map[server.FactionID]*server.GameAbility)
	for _, ability := range abilities {
		factionAbilityMap[ability.FactionID] = ability

		// set ability detail to battle ability
		battleAbility.Colour = ability.Colour
		battleAbility.ImageUrl = ability.ImageUrl
	}

	return battleAbility, factionAbilityMap, nil
}

const BattleAbilityCommand = BattleCommand("BATTLE:ABILITY")

func (ba *BattleArena) GameAbilityTrigger(gameAbilityEvent *server.GameAbilityEvent) error {
	ctx := context.Background()

	err := db.GameAbilityEventCreate(ctx, ba.Conn, ba.battle.ID, gameAbilityEvent)
	if err != nil {
		return terror.Error(err)
	}

	// To get the location in game its
	//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
	//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels
	if gameAbilityEvent.TriggeredOnCellX != nil && gameAbilityEvent.TriggeredOnCellY != nil {
		gameAbilityEvent.GameLocation.X = ((*gameAbilityEvent.TriggeredOnCellX * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + ba.battle.GameMap.LeftPixels
		gameAbilityEvent.GameLocation.Y = ((*gameAbilityEvent.TriggeredOnCellY * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + ba.battle.GameMap.TopPixels
	}

	// send new battle details to game client
	ctx, cancel := context.WithCancel(ba.ctx)

	gameMessage := &GameMessage{
		BattleCommand: BattleAbilityCommand,
		Payload:       gameAbilityEvent,
		context:       ctx,
		cancel:        cancel,
	}

	// NOTE: this will potentially lock game server if game client is disconnected
	// 		 so wrap it in a go routine
	func() {
		ba.send <- gameMessage
	}()

	return nil
}
