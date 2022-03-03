package battle_arena

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/comms"
	"server/db"
	"server/gamelog"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
)

func (ba *BattleArena) GetCurrentState() *server.Battle {
	return ba.battle
}

//sets up functions to get and set this property not available elsewhere
func (ba *BattleArena) GetGamesToClose() int {
	return ba.gamesToClose
}

func (ba *BattleArena) PutGamesToClose(games int) {
	ba.gamesToClose = games
}

const BattleStartCommand = BattleCommand("BATTLE:START")

type BattleStartRequest struct {
	Payload struct {
		BattleID    server.BattleID `json:"battleID"`
		WarMachines []*struct {
			Hash          string `json:"hash"`
			ParticipantID byte   `json:"participantID"`
		} `json:"warMachines"`
		WarMachineLocation []byte `json:"warMachineLocation"`
	} `json:"payload"`
}

func (ba *BattleArena) BattleActive() bool {
	if ba.battle == nil {
		return false
	}
	return ba.battle.State == server.StateMatchStart
}

// BattleStartHandler start a new battle
func (ba *BattleArena) BattleStartHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	req := &BattleStartRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	if ba.gamesToClose == 0 {
		return terror.Error(fmt.Errorf("stream has closed"))
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
			if wm.Hash == wmbid.Hash {
				wm.ParticipantID = wmbid.ParticipantID
				continue outerLoop
			}
		}
	}

	// check they all have ids
	for _, wm := range ba.battle.WarMachines {
		if wm.ParticipantID == 0 {
			return terror.Error(fmt.Errorf("missing participant ID for %s  %s", wm.Name, wm.Hash))
		}
	}

	ba.Log.Info().Msgf("Battle starting: %s", req.Payload.BattleID)
	for _, wm := range ba.battle.WarMachines {
		ba.Log.Info().Msgf("War Machine: %s - %s", wm.Name, wm.Hash)
	}

	ba.battle.BattleHistory = append(ba.battle.BattleHistory, req.Payload.WarMachineLocation)

	ba.battle.SpawnedAI = []*server.WarMachineMetadata{}

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

	fmt.Println(ba.battle.ID)
	err = db.BattleWarMachineAssign(ctx, tx, ba.battle.ID, ba.battle.WarMachines)
	if err != nil {
		return terror.Error(err)
	}

	// marshal warMachineData
	b, err := json.Marshal(ba.battle.WarMachines)
	if err != nil {
		return terror.Error(err)
	}

	_, err = db.CreateBattleStateEvent(ctx, tx, ba.battle.ID, server.BattleEventBattleStart, b)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	ba.Events.Trigger(ctx, EventGameStart, &EventData{BattleArena: ba.battle})

	// switch battle state to START
	ba.battle.State = server.StateMatchStart

	//games to close gets init to -1, and means it is unset, if it is set, tick down games to close
	if ba.gamesToClose > 0 {
		ba.gamesToClose -= 1
	}

	return nil
}

const BattleEndCommand = BattleCommand("BATTLE:END")

type BattleEndRequest struct {
	Payload struct {
		BattleID                   server.BattleID           `json:"battleID"`
		WinCondition               server.BattleWinCondition `json:"winCondition"`
		WinningWarMachineMetadatas []*struct {
			Hash   string `json:"hash"`
			Health int    `json:"health"`
		} `json:"winningWarMachines"`
	} `json:"payload"`
}

type BattleRewardList struct {
	BattleID                      server.BattleID
	WinnerFactionID               server.FactionID
	WinningWarMachineOwnerIDs     map[server.UserID]bool
	ExecuteKillWarMachineOwnerIDs map[server.UserID]bool
	TopSupsSpendUsers             []server.UserID
}

func (ba *BattleArena) BattleEndHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {
	// switch battle state to END
	ba.battle.State = server.StateMatchEnd

	req := &BattleEndRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	if req.Payload.BattleID.IsNil() {
		return terror.Error(fmt.Errorf("missing battle id"))
	}

	if req.Payload.BattleID != ba.battle.ID {
		return terror.Error(fmt.Errorf("mismatch battleID, expected %s, got %s", ba.battle.ID.String(), req.Payload.BattleID.String()))
	}

	// check battle is started
	_, err = db.BattleGet(ctx, ba.Conn, req.Payload.BattleID)
	if err != nil {
		return terror.Error(err, "current battle has not started yet.")
	}

	ba.Log.Info().Msgf("Battle ending: %s", req.Payload.BattleID)
	ba.Log.Info().Msg("Winning War Machines")
	for _, warMachine := range req.Payload.WinningWarMachineMetadatas {
		ba.Log.Info().Msgf("%s", warMachine.Hash)
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

	// record battle end
	err = db.BattleEnded(ctx, tx, req.Payload.BattleID, req.Payload.WinCondition)
	if err != nil {
		return terror.Error(err)
	}

	now := time.Now()
	ba.battle.EndedAt = &now

	// prepare battle reward request
	battleRewardList := &BattleRewardList{
		BattleID:                      req.Payload.BattleID,
		WinningWarMachineOwnerIDs:     make(map[server.UserID]bool),
		ExecuteKillWarMachineOwnerIDs: make(map[server.UserID]bool),
	}

	// cache in game war machines
	winningMachines := []*server.WarMachineMetadata{}
	wmq := []*comms.WarMachineQueueStat{}

	for _, bwm := range ba.battle.WarMachines {
		l := gamelog.GameLog.Debug().Str("hash", bwm.Hash).Str("battle_id", req.Payload.BattleID.String())
		// get contract reward from queuing
		winningHashes := []string{}
		for _, meta := range req.Payload.WinningWarMachineMetadatas {
			winningHashes = append(winningHashes, meta.Hash)
		}

		l.Str("winning_hashes", strings.Join(winningHashes, ",")).Msg("process warmachine from battle result")

		assetQueueStat, err := db.AssetQueuingStat(ctx, tx, bwm.Hash)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return terror.Error(err, "failed to get contract reward from db")
		}
		// default war machines
		isDefaultWarMachine := false
		if assetQueueStat == nil {
			isDefaultWarMachine = true
		}
		if assetQueueStat != nil {
			l.Str("amt", assetQueueStat.ContractReward).Msg("received assetQueueStat")
		} else {
			l.Msg("assetQueueStat is nil (probably default war machine)")
		}

		assetRepairRecord := &server.AssetRepairRecord{
			Hash:       bwm.Hash,
			RepairMode: server.RepairModeFast,
		}

		if !isDefaultWarMachine && !assetQueueStat.IsInsured {
			assetRepairRecord.RepairMode = server.RepairModeStandard
		}

		// if war machines win
		health, exists := WarMachineExistInList(req.Payload.WinningWarMachineMetadatas, bwm.Hash)
		l.Bool("in_list", exists).Msg("war machine list status")
		if exists {
			bwm.Health = health
			winningMachines = append(winningMachines, bwm)
			battleRewardList.WinnerFactionID = bwm.FactionID
			battleRewardList.WinningWarMachineOwnerIDs[bwm.OwnedByID] = true

			// skip if it is default war machine
			if isDefaultWarMachine {
				continue
			}

			// pay queuing contract reward
			if assetQueueStat != nil {
				l.Msg("asset is in queue, pay rewards")
				ba.passport.AssetContractRewardRedeem(
					bwm.OwnedByID,
					bwm.FactionID,
					assetQueueStat.ContractReward,
					server.TransactionReference(
						fmt.Sprintf(
							"redeem_faction_contract_reward|%s|%s",
							bwm.Name,
							time.Now(),
						),
					),
				)
			} else {
				l.Str("battle_id", req.Payload.BattleID.String()).Msg("asset is not in queue, skip reward")
			}

			// calc asset repair complete time
			assetRepairRecord.ExpectCompletedAt = calcRepairCompleteTime(bwm.MaxHealth, bwm.Health, assetQueueStat.IsInsured, now)
			err := db.AssetRepairInsert(ctx, tx, assetRepairRecord)
			if err != nil {
				return terror.Error(err)
			}

			err = db.BattleQueueRemove(ctx, tx, bwm.Hash)
			if err != nil {
				ba.Log.Err(err).Msgf("Failed to remove battle queue cache in db, token id: %s ", bwm.Hash)
			}
			wmq = append(wmq, &comms.WarMachineQueueStat{Hash: bwm.Hash})

			// broadcast repair stat
			ba.passport.AssetRepairStat(assetRepairRecord)

			continue
		}

		// skip if it is default war machine
		if isDefaultWarMachine {
			continue
		}

		// if loss, store mech in asset repair db
		assetRepairRecord.ExpectCompletedAt = calcRepairCompleteTime(bwm.MaxHealth, 0, assetQueueStat.IsInsured, now)
		err = db.AssetRepairInsert(ctx, tx, assetRepairRecord)
		if err != nil {
			return terror.Error(err)
		}

		err = db.BattleQueueRemove(ctx, tx, bwm.Hash)
		if err != nil {
			ba.Log.Err(err).Msgf("Failed to remove battle queue cache in db, token id: %s ", bwm.Hash)
		}
		wmq = append(wmq, &comms.WarMachineQueueStat{Hash: bwm.Hash})

		// broadcast repair stat
		ba.passport.AssetRepairStat(assetRepairRecord)

	}
	// broadcast queuing stat to passport server

	ba.passport.WarMachineQueuePositionBroadcast(wmq)

	ba.battle.WarMachines = []*server.WarMachineMetadata{}
	ba.battle.WinningWarMachines = winningMachines
	ba.battle.WinningCondition = (*string)(&req.Payload.WinCondition)

	// assign winner war machine
	if len(winningMachines) > 0 {
		err = db.BattleWinnerWarMachinesSet(ctx, tx, req.Payload.BattleID, winningMachines)
		if err != nil {
			return terror.Error(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	// set war machine durability
	for _, warMachine := range ba.battle.WarMachines {
		warMachine.Durability = 100 * warMachine.Health / warMachine.MaxHealth
	}

	// execute kill war machine owner id
	destoryedEvents, err := db.WarMachineDestroyedEventGetByBattleID(ctx, ba.Conn, req.Payload.BattleID)
	if err != nil {
		return terror.Error(err)
	}

	for _, event := range destoryedEvents {
		if event.KillByWarMachineHash == nil {
			continue
		}
		for _, warMachine := range ba.battle.WarMachines {
			if *event.KillByWarMachineHash != warMachine.Hash {
				continue
			}

			battleRewardList.ExecuteKillWarMachineOwnerIDs[warMachine.OwnedByID] = true
		}
	}

	ba.passport.TransferBattleFundToSupsPool()

	// trigger battle end
	ba.Events.Trigger(ctx, EventGameEnd, &EventData{
		BattleArena:      ba.battle,
		BattleRewardList: battleRewardList,
	})

	//checks if games left until close equals 0, if so, return early and do not init next battle
	if ba.gamesToClose == 0 {
		return nil
	}

	go func() {
		time.Sleep(25 * time.Second)
		err := ba.InitNextBattle()
		if err != nil {
			ba.Log.Err(err).Msg("Failed to initialise next battle")
		}
	}()

	return nil
}

const BattleReadyCommand = BattleCommand("BATTLE:READY")

// BattleReadyHandler gets called when the game client is ready for a new battle
func (ba *BattleArena) BattleReadyHandler(ctx context.Context, payload []byte, reply ReplyFunc) error {

	return nil
}

func WarMachineExistInList(wms []*struct {
	Hash   string `json:"hash"`
	Health int    `json:"health"`
}, hash string) (int, bool) {
	for _, wm := range wms {
		if wm.Hash == hash {
			return wm.Health, true
		}
	}

	return 0, false
}

// calcRepairCompleteTime
func calcRepairCompleteTime(maxHealth, health int, isInsured bool, now time.Time) time.Time {
	secondForEachPoint := 18
	if !isInsured {
		secondForEachPoint = 864
	}

	recoverPoint := (100 - health*100/maxHealth)

	return now.Add(time.Duration(recoverPoint) * time.Duration(secondForEachPoint) * time.Second)

}
