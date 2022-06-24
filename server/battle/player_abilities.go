package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/xsyn_rpcclient"
	"time"
)

const MechMoveCommandCreateCode = 8
const MechMoveCommandCancelCode = 9

const HubKeyMechCommandsSubscribe = "MECH:COMMANDS:SUBSCRIBE"

func (arena *Arena) MechCommandsSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	err := arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		return terror.Error(err, "Failed to get mech command logs")
	}
	return nil
}

func (arena *Arena) BroadcastFactionMechCommands(factionID string) error {
	if arena.currentBattleState() != BattleStageStart {
		return nil
	}

	ids := arena.currentBattleWarMachineIDs(factionID)
	if len(ids) == 0 {
		return nil
	}

	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.IN(helpers.UUIDArray2StrArray(ids)),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get mech command logs from db")
		return terror.Error(err, "Failed to get mech command logs")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_commands", factionID), HubKeyMechCommandsSubscribe, mmc)

	return nil
}

const HubKeyMechMoveCommandSubscribe = "MECH:MOVE:COMMAND:SUBSCRIBE"

type MechMoveCommandResponse struct {
	*boiler.MechMoveCommandLog
	RemainCooldownSeconds int `json:"remain_cooldown_seconds"`
}

func (arena *Arena) MechMoveCommandSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	hash := cctx.URLParam("hash")

	wm := arena.CurrentBattleWarMachineByHash(hash)
	if wm == nil {
		return terror.Error(terror.ErrInvalidInput, "Current mech is not on the battlefield")
	}

	// query unfinished mech move command
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("mech id", wm.ID).Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to get mech move command.")
	}

	resp := &MechMoveCommandResponse{
		RemainCooldownSeconds: 0,
	}

	if mmc != nil {
		resp.MechMoveCommandLog = mmc
		resp.RemainCooldownSeconds = 30 - int(time.Now().Sub(mmc.CreatedAt).Seconds())
		if resp.RemainCooldownSeconds < 0 {
			resp.RemainCooldownSeconds = 0
		}
	}

	reply(resp)
	return nil
}

const HubKeyMechMoveCommandCreate = "MECH:MOVE:COMMAND:CREATE"

type MechMoveCommandCreateRequest struct {
	Payload struct {
		Hash        string               `json:"mech_hash"`
		StartCoords *server.CellLocation `json:"start_coords"`
	} `json:"payload"`
}

// MechMoveCommandCreateHandler send mech move command to game client
func (arena *Arena) MechMoveCommandCreateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// check battle stage
	if arena.currentBattleState() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	req := &MechMoveCommandCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.StartCoords == nil {
		return terror.Error(fmt.Errorf("missing location"), "Missing location")
	}

	// check ownership
	wm := arena.CurrentBattleWarMachineByHash(req.Payload.Hash)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battlefield.")
	}

	if wm.OwnedByID != user.ID {
		gamelog.L.Warn().Str("mech id", wm.ID).Str("mech owner id", wm.OwnedByID).Str("current user id", user.ID).Msg("Unauthorised mech move command.")
		return terror.Error(terror.ErrForbidden, "The mech is not owned by current user.")
	}

	// check mech move command is triggered within 30 seconds
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.CreatedAt.GT(time.Now().Add(-30*time.Second)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to trigger mech move command")
	}

	if mmc != nil {
		return terror.Error(terror.ErrInvalidInput, "Command is still cooling down.")
	}

	txid, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.FromStringOrNil(user.ID),
		ToUserID:             SupremacyBattleUserID,
		Amount:               decimal.New(10, 18).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("mech_move_command|%s|%d", wm.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		SubGroup:             arena.CurrentBattle().ID,
		Description:          "mech move command: " + wm.ID,
		NotSafe:              true,
	})
	if err != nil {
		return terror.Error(err, err.Error())
	}

	now := time.Now()

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: MechMoveCommandCreateCode, // 8
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID, // trigger on war machine
		TriggeredOnCellX:    &req.Payload.StartCoords.X,
		TriggeredOnCellY:    &req.Payload.StartCoords.Y,
		EventID:             uuid.Must(uuid.NewV4()),
		GameLocation: arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(&server.CellLocation{
			X: req.Payload.StartCoords.X,
			Y: req.Payload.StartCoords.Y,
		}),
	}

	// check mech command
	arena.Message("BATTLE:ABILITY", event)

	// log mech move command
	mmc = &boiler.MechMoveCommandLog{
		MechID:        wm.ID,
		TriggeredByID: user.ID,
		CellX:         req.Payload.StartCoords.X,
		CellY:         req.Payload.StartCoords.Y,
		BattleID:      arena.CurrentBattle().ID,
		TXID:          txid,
		CreatedAt:     now,
	}

	err = mmc.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert mech move command")
		return terror.Error(err, "Failed to trigger mech move command.")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_command/%s", factionID, wm.Hash), HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
		MechMoveCommandLog:    mmc,
		RemainCooldownSeconds: 30,
	})

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to broadcast faction mech commands")
	}

	arena.BroadcastMechCommandNotification(&MechCommandNotification{
		MechID:    wm.ID,
		MechLabel: wm.Name,
		FactionID: wm.FactionID,
		Action:    MechCommandActionFired,
		FiredByUser: &UserBrief{
			ID:        uuid.FromStringOrNil(user.ID),
			Username:  user.Username.String,
			FactionID: user.FactionID.String,
			Gid:       user.Gid,
		},
	})

	reply(true)

	return nil
}

const HubKeyMechMoveCommandCancel = "MECH:MOVE:COMMAND:CANCEL"

type MechMoveCommandCancelRequest struct {
	Payload struct {
		Hash          string `json:"hash"`
		MoveCommandID string `json:"move_command_id"`
	} `json:"payload"`
}

// MechMoveCommandCancelHandler send cancel mech move command to game client
func (arena *Arena) MechMoveCommandCancelHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// check battle stage
	if arena.currentBattleState() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	req := &MechMoveCommandCancelRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	wm := arena.CurrentBattleWarMachineByHash(req.Payload.Hash)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battlefield.")
	}

	// check ownership
	if wm.OwnedByID != user.ID {
		gamelog.L.Warn().Str("mech owner id", wm.OwnedByID).Str("player id", user.ID).Msg("Invalid mech move cancel request")
		return terror.Error(terror.ErrForbidden, "This mech is not owned by the player.")
	}

	// get mech move command
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.ID.EQ(req.Payload.MoveCommandID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		qm.OrderBy(boiler.MechMoveCommandLogColumns.CreatedAt+" DESC"),
		qm.Load(boiler.MechMoveCommandLogRels.Mech),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("mech move command id", req.Payload.MoveCommandID).Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to cancel mech move command.")
	}

	// check mech id
	if mmc.MechID != wm.ID {
		gamelog.L.Warn().Str("mech move command id", mmc.ID).Str("expected mech id", mmc.MechID).Str("provided mech id", wm.ID).Msg("mech id mismatch")
		return terror.Error(fmt.Errorf("mech id mismatch"), "Failed to cancel mech move command")
	}

	if mmc.CancelledAt.Valid {
		return terror.Error(fmt.Errorf("move command is already cancelled"), "Mech move command is already cancelled.")
	}

	if mmc.ReachedAt.Valid {
		return terror.Error(fmt.Errorf("mech already reach the place"), "Mech already reach the commanded spot")
	}

	// cancel command
	mmc.CancelledAt = null.TimeFrom(time.Now())
	_, err = mmc.Update(gamedb.StdConn, boil.Whitelist(boiler.MechMoveCommandLogColumns.CancelledAt))
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech move command id", mmc.ID).Msg("Failed to up date mech move command in db")
		return terror.Error(err, "Failed to cancel mech move command")
	}

	// send mech move command to game client
	arena.Message("BATTLE:ABILITY", &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: MechMoveCommandCancelCode,
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID, // trigger on war machine
	})

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_command/%s", factionID, wm.Hash), HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
		MechMoveCommandLog:    mmc,
		RemainCooldownSeconds: 30 - int(time.Now().Sub(mmc.CreatedAt).Seconds()),
	})

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to broadcast faction mech commands")
	}

	arena.BroadcastMechCommandNotification(&MechCommandNotification{
		MechID:    wm.ID,
		MechLabel: wm.Name,
		FactionID: wm.FactionID,
		Action:    MechCommandActionCancel,
		FiredByUser: &UserBrief{
			ID:        uuid.FromStringOrNil(user.ID),
			Username:  user.Username.String,
			FactionID: user.FactionID.String,
			Gid:       user.Gid,
		},
	})

	reply(true)

	return nil
}
