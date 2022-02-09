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
		BattleID    server.BattleID `json:"battleID"`
		WarMachines []*struct {
			TokenID       uint64 `json:"tokenID"`
			ParticipantID byte   `json:"participantID"`
		} `json:"warMachines"`
		WarMachineLocation []byte `json:"warMachineLocation"`
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

	if len(req.Payload.WarMachines) <= 0 {
		return terror.Error(fmt.Errorf("cannot start battle with zero war machines"))
	}

	if len(req.Payload.WarMachines) != len(ba.battle.WarMachines) {
		return terror.Error(fmt.Errorf("mismatch warmachine count, expected %d, got %d", len(ba.battle.WarMachines), len(req.Payload.WarMachines)))
	}

	if req.Payload.BattleID != ba.battle.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", ba.battle.ID.String(), req.Payload.BattleID.String()))
	}

	// assign the participantID to the war machines
outerLoop:
	for _, wm := range ba.battle.WarMachines {
		for _, wmbid := range req.Payload.WarMachines {
			if wm.TokenID == wmbid.TokenID {
				wm.ParticipantID = wmbid.ParticipantID
				continue outerLoop
			}
		}
	}

	// check they all have ids
	for _, wm := range ba.battle.WarMachines {
		if wm.ParticipantID == 0 {
			return terror.Error(fmt.Errorf("missing participant ID for %s  %d", wm.Name, wm.TokenID))
		}
	}

	ba.Log.Info().Msgf("Battle starting: %s", req.Payload.BattleID)
	for _, wm := range ba.battle.WarMachines {
		ba.Log.Info().Msgf("War Machine: %s - %d", wm.Name, wm.TokenID)
	}

	ba.battle.BattleHistory = append(ba.battle.BattleHistory, req.Payload.WarMachineLocation)

	// save to database
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

	// Record battle to database
	err = db.BattleStarted(ctx, tx, ba.battle)
	if err != nil {
		return terror.Error(err)
	}

	err = db.BattleWarMachineAssign(ctx, tx, ba.battle.ID, ba.battle.WarMachines)
	if err != nil {
		return terror.Error(err)
	}

	_, err = db.CreateBattleStateEvent(ctx, tx, ba.battle.ID, server.BattleEventBattleStart)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	ba.Events.Trigger(ctx, EventGameStart, &EventData{BattleArena: ba.battle})

	return nil
}

const BattleEndCommand = BattleCommand("BATTLE:END")

type BattleEndRequest struct {
	Payload struct {
		BattleID              server.BattleID           `json:"battleID"`
		WinCondition          server.BattleWinCondition `json:"winCondition"`
		WinningWarMachineNFTs []*struct {
			TokenID uint64 `json:"tokenID"`
			Health  int    `json:"health"`
		} `json:"winningWarMachines"`
	} `json:"payload"`
}

type BattleRewardList struct {
	BattleID                      server.BattleID
	WinnerFactionID               server.FactionID
	WinningWarMachineOwnerIDs     map[server.UserID]bool
	ExecuteKillWarMachineOwnerIDs map[server.UserID]bool
}

func (ba *BattleArena) BattleEndHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &BattleEndRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	ba.Log.Info().Msgf("Battle ending: %s", req.Payload.BattleID)
	ba.Log.Info().Msg("Winning War Machines")
	for _, warMachine := range req.Payload.WinningWarMachineNFTs {
		ba.Log.Info().Msgf("%d", warMachine.TokenID)
	}

	// prepare battle reward request
	battleRewardList := &BattleRewardList{
		BattleID:                      req.Payload.BattleID,
		WinningWarMachineOwnerIDs:     make(map[server.UserID]bool),
		ExecuteKillWarMachineOwnerIDs: make(map[server.UserID]bool),
	}

	winningMachines := []*server.WarMachineNFT{}

	for _, wm := range req.Payload.WinningWarMachineNFTs {
		for _, bwm := range ba.battle.WarMachines {
			if wm.TokenID == bwm.TokenID {
				bwm.Health = wm.Health
				winningMachines = append(winningMachines, bwm)
				battleRewardList.WinnerFactionID = bwm.FactionID
				battleRewardList.WinningWarMachineOwnerIDs[bwm.OwnedByID] = true
			}
		}
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
	if len(req.Payload.WinningWarMachineNFTs) > 0 {
		err = db.BattleWinnerWarMachinesSet(ctx, ba.Conn, req.Payload.BattleID, winningMachines)
		if err != nil {
			return terror.Error(err)
		}
	}

	_, err = db.CreateBattleStateEvent(ctx, tx, ba.battle.ID, server.BattleEventBattleEnd)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	// set war machine durability
	for _, warMachine := range ba.battle.WarMachines {
		warMachine.Durability = 100 * warMachine.Health / warMachine.MaxHealth
	}

	//release war machine
	if len(ba.battle.WarMachines) > 0 {
		ba.passport.AssetRelease(
			ctx,
			fmt.Sprintf("release_asset|battleID:%s", ba.battle.ID),
			ba.battle.WarMachines,
		)
	}

	// execute kill war machine owner id
	destoryedEvents, err := db.WarMachineDestroyedEventGetByBattleID(ctx, ba.Conn, req.Payload.BattleID)
	if err != nil {
		return terror.Error(err)
	}

	for _, event := range destoryedEvents {
		if event.KillByWarMachineID == nil {
			continue
		}
		for _, warMachine := range ba.battle.WarMachines {
			if *event.KillByWarMachineID != warMachine.TokenID {
				continue
			}

			battleRewardList.ExecuteKillWarMachineOwnerIDs[warMachine.OwnedByID] = true
		}
	}

	err = ba.passport.TransferBattleFundToSupsPool(ctx, fmt.Sprintf("transfer_battle_fund_to_sup_pool|%s", req.Payload.BattleID))
	if err != nil {
		return terror.Error(err, "Failed to distribute battle reward")
	}

	// trigger battle end
	ba.Events.Trigger(ctx, EventGameEnd, &EventData{
		BattleArena:      ba.battle,
		BattleRewardList: battleRewardList,
	})

	return nil
}

const BattleReadyCommand = BattleCommand("BATTLE:READY")

// BattleReadyHandler gets called when the game client is ready for a new battle
func (ba *BattleArena) BattleReadyHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {

	err := ba.InitNextBattle()
	if err != nil {
		ba.Log.Err(err).Msg("Failed to initialise next battle")
		return terror.Error(err)
	}
	return nil
}
