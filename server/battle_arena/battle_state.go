package battle_arena

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/passport"
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
		WarMachines []*server.WarMachine `json:"warMachines"`
		MapName     string               `json:"mapName"`
	} `json:"payload"`
}

func (ba *BattleArena) BattleStartHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &BattleStartRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	if req.Payload.BattleID.IsNil() {
		return terror.Error(fmt.Errorf("missing battle id"))
	}

	if req.Payload.MapName == "" {
		return terror.Error(fmt.Errorf("missing map name"))
	}

	if len(req.Payload.WarMachines) <= 0 {
		return terror.Error(fmt.Errorf("cannot start battle with zero war machines"))
	}

	// TODO: add get map via name
	theMap := &server.GameMap{}

	for _, mp := range server.FakeGameMaps {
		if mp.Name == req.Payload.MapName {
			theMap = mp
			break
		}
	}

	if theMap == nil {
		return terror.Error(fmt.Errorf("unable to find map %s", req.Payload.MapName))
	}

	req.Payload.WarMachines = ba.passport.GetWarMachines()

	factions := passport.FakeFactions

	// match faction into war machine
	for _, warMachine := range req.Payload.WarMachines {
		for _, faction := range factions {
			if warMachine.FactionID == faction.ID {
				warMachine.Faction = faction
				break
			}
		}
	}

	ba.battle = &server.Battle{
		ID:          req.Payload.BattleID,
		WarMachines: req.Payload.WarMachines,
		StartedAt:   time.Now(),
		Map:         theMap,
	}

	ba.Log.Info().Msgf("Battle starting: %s", req.Payload.BattleID)
	for _, wm := range req.Payload.WarMachines {
		ba.Log.Info().Msgf("War Machine: %s - %s", wm.Name, wm.ID)
	}

	//save to database
	tx, err := ba.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	err = db.BattleStarted(ctx, tx, req.Payload.BattleID, req.Payload.WarMachines)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	ba.Events.Trigger(ctx, EventGameStart, &EventData{BattleArena: ba.battle})

	// start dummy war machine moving
	go ba.FakeWarMachinePositionUpdate()
	return nil
}

const BattleEndCommand = BattleCommand("BATTLE:END")

type BattleEndRequest struct {
	Payload struct {
		BattleID           server.BattleID           `json:"battleId"`
		WinningWarMachines []*server.WarMachineID    `json:"winningWarMachines"`
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

	err = db.BattleEnded(ctx, tx, req.Payload.BattleID, req.Payload.WinningWarMachines, req.Payload.WinCondition)
	if err != nil {
		return terror.Error(err)
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
