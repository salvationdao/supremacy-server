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
	"server/passport"
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

	for _, wm := range ba.battle.WarMachines {
		err = db.ContractRewardInsert(ctx, tx, ba.battle.ID, wm.ContractReward, wm.Hash)
		if err != nil {
			return terror.Error(err)
		}
	}

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

func SendForRepairs(ctx context.Context, tx db.Conn, battleID server.BattleID, ppclient *passport.Passport, maxHealth int, health int, hash string) error {
	isDefault := db.IsDefaultWarMachine(ctx, tx, hash)
	if isDefault {
		gamelog.GameLog.Warn().
			Str("fn", "SendForRepairs").
			Str("battle_id", battleID.String()).
			Str("hash", hash).
			Msg("skip, is default war machine")
		return nil
	}

	gamelog.GameLog.Debug().Str("fn", "SendForRepairs").Str("battle_id", battleID.String()).Str("hash", hash).Msg("send participant to repairs")

	isInsured, err := db.IsInsured(ctx, tx, hash)
	if err != nil {
		return fmt.Errorf("get is insured: %w", err)
	}
	repairMode := server.RepairModeStandard
	if isInsured {
		gamelog.GameLog.Warn().
			Str("fn", "SendForRepairs").
			Str("battle_id", battleID.String()).
			Str("hash", hash).
			Msg("is insured, fast repair")
		repairMode = server.RepairModeFast
	}
	record := &server.AssetRepairRecord{
		Hash:              hash,
		RepairMode:        repairMode,
		ExpectCompletedAt: calcRepairCompleteTime(maxHealth, health, isInsured, time.Now()),
	}
	err = db.AssetRepairInsert(ctx, tx, record)
	if err != nil {
		return fmt.Errorf("insert asset repair record: %w", err)
	}

	gamelog.GameLog.Debug().
		Str("fn", "SendForRepairs").
		Str("battle_id", battleID.String()).
		Str("hash", hash).
		Msg("broadcast repair event")
	ppclient.AssetRepairStat(record)
	return nil
}

func RemoveParticipant(ctx context.Context, tx db.Conn, battleID server.BattleID, hash string) error {
	// Remove from queue
	// Skip default mechs (won't be in queue)
	isDefault := db.IsDefaultWarMachine(ctx, tx, hash)
	if isDefault {
		return nil
	}

	gamelog.GameLog.Debug().Str("fn", "RemoveParticipant").Str("battle_id", battleID.String()).Str("hash", hash).Msg("remove participant from queue")

	err := db.BattleQueueRemove(ctx, tx, hash)
	if err != nil {
		return fmt.Errorf("remove from queue: %w", err)
	}
	return nil
}

func PayWinners(ctx context.Context, tx db.Conn, ppclient *passport.Passport, battleID server.BattleID, winnerHash string) error {
	gamelog.GameLog.Debug().Str("fn", "PayWinners").Str("battle_id", battleID.String()).Str("winning_hash", winnerHash).Msg("attempt to pay winner from queue")

	// Payout

	// Skip default mechs (won't be in queue)
	isDefault := db.IsDefaultWarMachine(ctx, tx, winnerHash)
	if isDefault {
		gamelog.GameLog.Warn().
			Str("fn", "PayWinners").
			Str("battle_id", battleID.String()).
			Str("winning_hash", winnerHash).
			Msg("skip, is default war machine")
		return nil
	}

	gamelog.GameLog.Debug().
		Str("fn", "PayWinners").
		Str("battle_id", battleID.String()).
		Str("winning_hash", winnerHash).Msg("not default mech, prepare to pay winner from queue")

	reward, err := db.ContractRewardGet(ctx, tx, winnerHash)
	if err != nil {
		return fmt.Errorf("get reward from db: %w", err)
	}

	ownedByID, factionID, mechName, err := db.MechMetadata(ctx, tx, winnerHash)
	if err != nil {
		return fmt.Errorf("get metadata: %w", err)
	}

	gamelog.GameLog.Debug().
		Str("fn", "PayWinners").
		Str("battle_id", battleID.String()).
		Str("winning_hash", winnerHash).Msg("send to rpc")

	err = ppclient.AssetContractRewardRedeem(
		ownedByID,
		factionID,
		reward,
		server.TransactionReference(
			fmt.Sprintf(
				"redeem_faction_contract_reward|%s|%s",
				mechName,
				time.Now(),
			),
		),
		battleID.String(),
	)
	if err != nil {
		return fmt.Errorf("request redeem contract reward: %w", err)
	}
	err = db.ContractRewardMarkIsPaid(ctx, tx, battleID, winnerHash)
	if err != nil {
		return fmt.Errorf("insert reward: %w", err)
	}
	// Insert into battle ID
	return nil
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

	for _, meta := range req.Payload.WinningWarMachineMetadatas {
		gamelog.GameLog.Debug().Str("battle_id", req.Payload.BattleID.String()).Str("hash", meta.Hash).Msg("process winning mech")
		err = PayWinners(ctx, tx, ba.passport, req.Payload.BattleID, meta.Hash)
		if err != nil {
			gamelog.GameLog.Err(err).Str("battle_id", req.Payload.BattleID.String()).Str("hash", meta.Hash).Msg("failed to pay winners")
			return fmt.Errorf("PayWinners: %w", err)
		}
	}

	for _, bwm := range ba.battle.WarMachines {
		gamelog.GameLog.Debug().Str("battle_id", req.Payload.BattleID.String()).Str("hash", bwm.Hash).Msg("process participating mech")
		err = SendForRepairs(ctx, tx, req.Payload.BattleID, ba.passport, bwm.MaxHealth, bwm.Health, bwm.Hash)
		if err != nil {
			gamelog.GameLog.Err(err).Str("battle_id", req.Payload.BattleID.String()).Str("hash", bwm.Hash).Msg("failed to send for repairs")
			return fmt.Errorf("SendForRepairs: %w", err)
		}
		err = RemoveParticipant(ctx, tx, req.Payload.BattleID, bwm.Hash)
		if err != nil {
			gamelog.GameLog.Err(err).Str("battle_id", req.Payload.BattleID.String()).Str("hash", bwm.Hash).Msg("failed to remove participants")
			return fmt.Errorf("RemoveParticipant: %w", err)
		}
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
