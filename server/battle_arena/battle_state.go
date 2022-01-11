package battle_arena

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"time"

	"github.com/ninja-software/terror/v2"
)

func (ba *BattleArena) GetCurrentState() *server.Battle {
	return ba.battle
}

const BattleStartCommand = BattleCommand("BATTLE:START")

type BattleStartRequest struct {
	Payload struct {
		BattleID    server.BattleID      `json:"battleId"`
		GameMapID   server.GameMapID     `json:"gameMapID"`
		WarMachines []*server.WarMachine `json:"warMachines"`
	} `json:"payload"`
}

// BattleStartHandler start a new battle
func (ba *BattleArena) BattleStartHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &BattleStartRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	if req.Payload.BattleID.IsNil() {
		return terror.Error(fmt.Errorf("missing battle id"))
	}

	if req.Payload.GameMapID.IsNil() {
		return terror.Error(fmt.Errorf("missing map id"))
	}

	if len(req.Payload.WarMachines) <= 0 {
		return terror.Error(fmt.Errorf("cannot start battle with zero war machines"))
	}

	// game map
	gameMap, err := db.GameMapGet(ctx, ba.Conn, req.Payload.GameMapID)
	if err != nil {
		return terror.Error(err)
	}

	/*************************************
	* TODO: Replace this with queue system
	*************************************/
	// factions
	factions, err := db.FactionAll(ctx, ba.Conn)
	if err != nil {
		return terror.Error(err)
	}

	warMachines, err := db.WarMachineAll(ctx, ba.Conn)
	if err != nil {
		return terror.Error(err)
	}

	warMachineIDs := []server.WarMachineID{}
	for i, warMachine := range warMachines {
		warMachineIDs = append(warMachineIDs, warMachine.ID)
		index := i % len(factions)
		warMachine.Faction = factions[index]
		warMachine.FactionID = &warMachine.Faction.ID
		warMachine.Position = &server.Vector3{
			X: 100 + 45*i,
			Y: 100 + 45*i,
			Z: 0,
		}
		warMachine.Rotation = 100 * i % 360
	}
	/*************************************
	* TODO: Replace this with queue system
	*************************************/

	ba.Log.Info().Msgf("Battle starting: %s", req.Payload.BattleID)
	for _, wm := range req.Payload.WarMachines {
		ba.Log.Info().Msgf("War Machine: %s - %s", wm.Name, wm.ID)
	}

	battle := &server.Battle{
		ID:        req.Payload.BattleID,
		GameMapID: gameMap.ID,
	}

	//save to database
	tx, err := ba.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	defer tx.Rollback(ctx)

	err = db.BattleStarted(ctx, tx, battle)
	if err != nil {
		return terror.Error(err)
	}

	err = db.BattleWarMachineAssign(ctx, tx, battle.ID, warMachineIDs)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	// set battle
	ba.battle.ID = battle.ID
	ba.battle.GameMapID = battle.GameMapID
	ba.battle.StartedAt = battle.StartedAt
	ba.battle.GameMap = gameMap
	ba.battle.WarMachines = warMachines
	ba.battle.WinningCondition = nil
	ba.battle.EndedAt = nil

	ba.Events.Trigger(ctx, EventGameStart, &EventData{BattleArena: ba.battle})

	// start dummy war machine moving
	go ba.FakeWarMachinePositionUpdate()
	return nil
}

const BattleEndCommand = BattleCommand("BATTLE:END")

type BattleEndRequest struct {
	Payload struct {
		BattleID           server.BattleID           `json:"battleId"`
		WinningWarMachines []server.WarMachineID     `json:"winningWarMachines"`
		WinCondition       server.BattleWinCondition `json:"winCondition"`
	} `json:"payload"`
}

func (ba *BattleArena) BattleEndHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &BattleEndRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Battle ending: %s", req.Payload.BattleID)
	ba.Log.Info().Msg("Winning War Machines")
	for _, id := range req.Payload.WinningWarMachines {
		ba.Log.Info().Msg(id.String())
	}

	//save to database
	tx, err := ba.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	defer tx.Rollback(ctx)

	err = db.BattleEnded(ctx, tx, req.Payload.BattleID, req.Payload.WinCondition)
	if err != nil {
		return terror.Error(err)
	}

	// assign winner war machine
	if len(req.Payload.WinningWarMachines) > 0 {
		err = db.BattleWinnerWarMachinesSet(ctx, ba.Conn, req.Payload.BattleID, req.Payload.WinningWarMachines)
		if err != nil {
			return terror.Error(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	ba.Events.Trigger(ctx, EventGameEnd, &EventData{BattleArena: ba.battle})

	go func() {
		time.Sleep(10 * time.Second)
		err := ba.InitNextBattle()
		if err != nil {
			ba.Log.Err(err).Msgf("error initializing new battle")
			return
		}
	}()

	return nil
}
