package battle_arena

import (
	"context"
	"fmt"
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
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("fjdsijfksdjfkadsjflkasdjf;lidsajf;klasdjfoiadsjfoiajdsfijads;flijdsa;oifjsad;of")
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	for factionID := range ba.BattleQueueMap {
		ba.battle.WarMachines = append(ba.battle.WarMachines, ba.GetBattleWarMachineFromQueue(factionID, mechsPerFaction)...)
	}

	fmt.Println("EXit dfasdlkfjlkdsjfkldsajglkadsjg;lkdsajgk;lasdjglk;adsjkg;as")

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

		fmt.Println("Enter Asset!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		// set war machine lock request
		err := ba.passport.AssetLock(ba.ctx, hashes)
		fmt.Println("Enter Assetdfslkjsdlkfjdslkjfkdslfj;lksdafjk;ldsajgfk;lsadjf;lksdjf;lkasjf;")
		if err != nil {
			ba.Log.Err(err).Msg("Failed to lock assets")
			// TODO: figure out how to handle this
		}
	}

	// clean up battle end message of the last battle
	ba.battle.EndedAt = nil
	ba.Events.Trigger(context.Background(), EventGameInit, nil)

	ba.Log.Info().Msgf("Initializing new battle: %s", ba.battle.ID)

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

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	spew.Dump(gameMessage)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// NOTE: this will potentially lock game server if game client is disconnected
	// 		 so wrap it in a go routine
	go func() {

		fmt.Println("fired to game client 4395345983475987349857438957230572348572439857984035789432750")
		ba.send <- gameMessage
		fmt.Println("fired 44444444444444444444444444444444444")
	}()
	return nil
}
