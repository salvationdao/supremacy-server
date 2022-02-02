package battle_arena

import (
	"context"
	"server"
	"server/db"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

const BattleCommandInitBattle BattleCommand = "BATTLE:INIT"

func (ba *BattleArena) InitNextBattle() error {

	ctx := context.Background()
	// generate a new battle event
	ba.battle.ID = server.BattleID(uuid.Must(uuid.NewV4()))

	// assign a random map
	gameMap, err := db.GameMapGetRandom(ba.ctx, ba.Conn)
	if err != nil {
		ba.Log.Err(err).Msg("")
		return terror.Error(err)
	}
	ba.battle.GameMap = gameMap
	ba.battle.GameMapID = gameMap.ID

	// get NFT from battle queue
	ba.battle.WarMachines = []*server.WarMachineNFT{}

	for len(ba.BattleQueueMap) == 0 {
		ba.Log.Info().Msg("No factions, trying again in 2 seconds")
		time.Sleep(2 * time.Second)
	}

	for factionID := range ba.BattleQueueMap {
		ba.battle.WarMachines = append(ba.battle.WarMachines, ba.GetBattleWarMachineFromQueue(factionID)...)
	}

	if len(ba.battle.WarMachines) > 0 {
		tokenIDs := []uint64{}
		for _, warMachine := range ba.battle.WarMachines {
			tokenIDs = append(tokenIDs, warMachine.TokenID)
		}

		// set war machine lock request
		err := ba.passport.AssetLock(ctx, "asset_lock", tokenIDs)
		if err != nil {
			ba.Log.Err(err).Msg("Failed to lock assets")
			// TODO: figure out how to handle this
		}
	}

	ba.Log.Info().Msgf("Initializing new battle: %s", ba.battle.ID)

	// send new battle details to game client
	ctx, cancel := context.WithCancel(ba.ctx)

	// Setup payload
	payload := struct {
		BattleID    server.BattleID         `json:"battleID"`
		MapName     string                  `json:"mapName"`
		WarMachines []*server.WarMachineNFT `json:"warMachines"`
	}{
		BattleID:    ba.battle.ID,
		MapName:     ba.battle.GameMap.Name,
		WarMachines: ba.battle.WarMachines,
	}

	gameMessage := &GameMessage{
		BattleCommand: BattleCommandInitBattle,
		Payload:       payload,
		context:       ctx,
		cancel:        cancel,
	}

	ba.send <- gameMessage
	return nil
}

//
//func (ba *BattleArena) InitTestBattle() error {
//
//	//ba.passport
//	// Get factions
//	factionRedMountain := &server.Faction{
//		ID:    server.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))),
//		Label: "Red Mountain Offworld Mining Corporation",
//		Theme: &server.FactionTheme{
//			Primary:    "#C24242",
//			Secondary:  "#FFFFFF",
//			Background: "#0D0404",
//		},
//	}
//	factionBoston := &server.Faction{
//		ID:    server.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))),
//		Label: "Boston Cybernetics",
//		Theme: &server.FactionTheme{
//			Primary:    "#428EC1",
//			Secondary:  "#FFFFFF",
//			Background: "#050A12",
//		},
//	}
//	factionZaibatsu := &server.Faction{
//		ID:    server.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))),
//		Label: "Zaibatsu Heavy Industries",
//		Theme: &server.FactionTheme{
//			Primary:    "#FFFFFF",
//			Secondary:  "#FFFFFF",
//			Background: "#0D0D0D",
//		},
//	}
//
//	// Get test WarMachines
//	warMachineNFTs := []*server.WarMachineNFT{
//		{
//			TokenID:     1,
//			Name:        "Tenshi Mk1",
//			FactionID:   factionZaibatsu.ID,
//			Faction:     factionZaibatsu,
//			WeaponNames: []string{"Sniper Rifle", "Laser Sword", "Rocket Pods"},
//		},
//		{
//			TokenID:     2,
//			Name:        "Olympus Mons LY07",
//			FactionID:   factionRedMountain.ID,
//			Faction:     factionRedMountain,
//			WeaponNames: []string{"Dual Cannons", "Rocket Pods"},
//		},
//		{
//			TokenID:     3,
//			Name:        "Law Enforcer X-1000",
//			FactionID:   factionBoston.ID,
//			Faction:     factionBoston,
//			WeaponNames: []string{"Plasma Rifle", "Sword"},
//		},
//		{
//			TokenID:     4,
//			Name:        "Tenshi Mk1 B",
//			FactionID:   factionZaibatsu.ID,
//			Faction:     factionZaibatsu,
//			WeaponNames: []string{"Sniper Rifle", "Laser Sword", "Rocket Pods"},
//		},
//		{
//			TokenID:     5,
//			Name:        "Olympus Mons LY07 B",
//			FactionID:   factionRedMountain.ID,
//			Faction:     factionRedMountain,
//			WeaponNames: []string{"Dual Cannons", "Rocket Pods"},
//		},
//		{
//			TokenID:     6,
//			Name:        "Law Enforcer X-1000 B",
//			FactionID:   factionBoston.ID,
//			Faction:     factionBoston,
//			WeaponNames: []string{"Plasma Rifle", "Sword"},
//		},
//	}
//
//	// generate a new battle event
//	newBattle := &server.Battle{
//		ID: server.BattleID(uuid.Must(uuid.NewV4())),
//	}
//
//	// assign a random map
//	gameMap, err := db.GameMapGetRandom(ba.ctx, ba.Conn)
//	if err != nil {
//		return terror.Error(err)
//	}
//
//	newBattle.GameMap = gameMap
//	newBattle.GameMapID = gameMap.ID
//
//	newBattle.WarMachines = warMachineNFTs
//
//	ba.Log.Info().Msgf("Initializing new battle: %s", newBattle.ID)
//
//	ba.battle = newBattle
//
//	// Setup payload
//	payload := struct {
//		BattleID    server.BattleID         `json:"battleID"`
//		MapName     string                  `json:"mapName"`
//		WarMachines []*server.WarMachineNFT `json:"warMachines"`
//	}{
//		BattleID:    newBattle.ID,
//		MapName:     newBattle.GameMap.Name,
//		WarMachines: newBattle.WarMachines,
//	}
//
//	// send new battle details to game client
//	ctx, cancel := context.WithCancel(ba.ctx)
//
//	gameMessage := &GameMessage{
//		BattleCommand: BattleCommandInitBattle,
//		Payload:       payload,
//		context:       ctx,
//		cancel:        cancel,
//	}
//
//	ba.send <- gameMessage
//	return nil
//}
