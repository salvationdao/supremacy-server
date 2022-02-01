package battle_arena

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/db"

	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
)

func (ba *BattleArena) GetCurrentState() *server.Battle {
	return ba.battle
}

const BattleStartCommand = BattleCommand("BATTLE:START")

type BattleStartRequest struct {
	Payload struct {
		BattleID       server.BattleID         `json:"battleID"`
		GameMapID      server.GameMapID        `json:"gameMapID"`
		WarMachineNFTs []*server.WarMachineNFT `json:"warMachines"`
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

	if len(req.Payload.WarMachineNFTs) <= 0 {
		return terror.Error(fmt.Errorf("cannot start battle with zero war machines"))
	}

	// game map
	gameMap, err := db.GameMapGet(ctx, ba.Conn, req.Payload.GameMapID)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Battle starting: %s", req.Payload.BattleID)
	for _, wm := range req.Payload.WarMachineNFTs {
		ba.Log.Info().Msgf("War Machine: %s - %d", wm.Name, wm.TokenID)
	}

	//save to database
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

	// reset a battle
	ba.battle.ID = req.Payload.BattleID
	ba.battle.GameMapID = req.Payload.GameMapID

	err = db.BattleStarted(ctx, tx, ba.battle)
	if err != nil {
		return terror.Error(err)
	}

	err = db.BattleWarMachineAssign(ctx, tx, ba.battle.ID, req.Payload.WarMachineNFTs)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	for _, warMachine := range req.Payload.WarMachineNFTs {
		warMachine.Faction = ba.battle.FactionMap[warMachine.FactionID]
	}

	// set the rest fo the fields of the battle
	ba.battle.GameMap = gameMap
	ba.battle.WarMachines = req.Payload.WarMachineNFTs
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
		BattleID           server.BattleID           `json:"battleID"`
		WinningWarMachines []uint64                  `json:"winningWarMachines"`
		WinCondition       server.BattleWinCondition `json:"winCondition"`
		WarMachineNFTs     []*server.WarMachineNFT   `json:"warMachines"`
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
		ba.Log.Info().Msgf("%d", id)
	}

	//save to database
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

	// update latest war machine stat
	ba.battle.WarMachines = req.Payload.WarMachineNFTs

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	ba.Events.Trigger(ctx, EventGameEnd, &EventData{BattleArena: ba.battle})

	return nil
}
