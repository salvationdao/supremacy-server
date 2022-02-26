package battle_arena

import (
	"context"
	"server"
	"server/db"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
)

const BattleCommandInitBattle BattleCommand = "BATTLE:INIT"

func (ba *BattleArena) InitNextBattle() error {
	// switch battle state to LOBBY
	ba.battle.State = server.StateLobby

	// send new battle details to game client

	// generate a new battle event
	ba.battle.ID = server.BattleID(uuid.Must(uuid.NewV4()))

	// clean up war machine destroyed record cache
	for key := range ba.battle.WarMachineDestroyedRecordMap {
		delete(ba.battle.WarMachineDestroyedRecordMap, key)
	}

	// assign a random map
	gameMap, err := db.GameMapGetRandom(ba.ctx, ba.Conn)
	if err != nil {
		ba.Log.Err(err).Msg("")
		return terror.Error(err)
	}
	ba.battle.GameMap = gameMap
	ba.battle.GameMapID = gameMap.ID

	// get NFT from battle queue
	ba.battle.WarMachines = []*server.WarMachineMetadata{}

	for len(ba.BattleQueueMap) == 0 {
		ba.Log.Info().Msg("No factions, trying again in 2 seconds")
		time.Sleep(2 * time.Second)
	}
	mechsPerFaction := gameMap.MaxSpawns / 3
	for factionID := range ba.BattleQueueMap {
		ba.battle.WarMachines = append(ba.battle.WarMachines, ba.GetBattleWarMachineFromQueue(factionID, mechsPerFaction)...)
	}

	// get Zaibatsu faction abilities to insert
	zaibatsuAbility, err := db.GetZaibatsuFactionAbility(context.Background(), ba.Conn)
	if err != nil {
		ba.Log.Err(err).Msg("Unable to get zaibatsu faction ability")
		return terror.Error(err)
	}

	if len(ba.battle.WarMachines) > 0 {
		hashes := []string{}
		for _, warMachine := range ba.battle.WarMachines {
			hashes = append(hashes, warMachine.Hash)

			if warMachine.FactionID == server.ZaibatsuFactionID {
				// if war machine is from Zaibatsu, insert the ability as faction ability
				warMachine.Abilities = append(warMachine.Abilities, &server.AbilityMetadata{
					ID:           zaibatsuAbility.ID,
					Identity:     uuid.Must(uuid.NewV4()), // track ability's price
					Colour:       zaibatsuAbility.Colour,
					GameClientID: int(zaibatsuAbility.GameClientAbilityID),
					Image:        zaibatsuAbility.ImageUrl,
					Description:  zaibatsuAbility.Description,
					Name:         zaibatsuAbility.Label,
					SupsCost:     zaibatsuAbility.SupsCost,
				})
			}
		}

		// set war machine lock request
		err := ba.passport.AssetLock(ba.ctx, hashes)
		if err != nil {
			ba.Log.Err(err).Msg("Failed to lock assets")
			// TODO: figure out how to handle this
		}
	}

	// clean up battle end message of the last battle
	ba.battle.EndedAt = nil
	ba.Events.Trigger(context.Background(), EventGameInit, nil)

	ba.Log.Info().Msgf("Initializing new battle: %s", ba.battle.ID)

	// trunc war machine name before it is send to battle
	for _, wm := range ba.battle.WarMachines {
		if len(wm.Name) > 20 {
			wm.Name = wm.Name[:20]
		}
	}

	for _, wm := range ba.battle.WarMachines {
		fillFaction(wm)
	}

	spew.Dump(ba.battle.WarMachines)

	// Setup payload
	payload := struct {
		BattleID    server.BattleID              `json:"battleID"`
		MapName     string                       `json:"mapName"`
		WarMachines []*server.WarMachineMetadata `json:"warMachines"`
	}{
		BattleID:    ba.battle.ID,
		MapName:     ba.battle.GameMap.Name,
		WarMachines: ba.battle.WarMachines,
	}

	ctx, cancel := context.WithCancel(ba.ctx)
	gameMessage := &GameMessage{
		BattleCommand: BattleCommandInitBattle,
		Payload:       payload,
		context:       ctx,
		cancel:        cancel,
	}

	// NOTE: this will potentially lock game server if game client is disconnected
	// 		 so wrap it in a go routine
	go func() {
		ba.send <- gameMessage
	}()
	return nil
}

var RedMountainFaction = &server.Faction{
	ID:    server.RedMountainFactionID,
	Label: "Red Mountain Offworld Mining Corporation",
	Theme: &server.FactionTheme{
		Primary:    "#C24242",
		Secondary:  "#FFFFFF",
		Background: "#120E0E",
	},
}

var BostonFaction = &server.Faction{
	ID:    server.BostonCyberneticsFactionID,
	Label: "Boston Cybernetics",
	Theme: &server.FactionTheme{
		Primary:    "#428EC1",
		Secondary:  "#FFFFFF",
		Background: "#080C12",
	},
}

var ZaibatsuFaction = &server.Faction{
	ID:    server.ZaibatsuFactionID,
	Label: "Zaibatsu Heavy Industries",
	Theme: &server.FactionTheme{
		Primary:    "#FFFFFF",
		Secondary:  "#000000",
		Background: "#0D0D0D",
	},
}

func fillFaction(wm *server.WarMachineMetadata) {
	switch wm.FactionID {
	case server.BostonCyberneticsFactionID:
		wm.Faction = BostonFaction
	case server.RedMountainFactionID:
		wm.Faction = RedMountainFaction
	case server.ZaibatsuFactionID:
		wm.Faction = ZaibatsuFaction
	}
}
