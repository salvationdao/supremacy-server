package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type QueueJoinHandlerResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
}

type QueueJoinRequest struct {
	Payload struct {
		AssetHash                   string `json:"asset_hash"`
		NeedInsured                 bool   `json:"need_insured"`
		EnablePushNotifications     bool   `json:"enable_push_notifications,omitempty"`
		MobileNumber                string `json:"mobile_number,omitempty"`
		EnableTelegramNotifications bool   `json:"enable_telegram_notifications"`
	} `json:"payload"`
}

const WSQueueJoin = "BATTLE:QUEUE:JOIN"

func (arena *Arena) QueueJoinHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	msg := &QueueJoinRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	mech, err := db.Mech(gamedb.StdConn, mechID.String())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	if mech.XsynLocked {
		err := fmt.Errorf("mech is locked to xsyn locked")
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("war machine is xsyn locked")
		return err
	}

	if mech.CollectionItem.LockedToMarketplace {
		err := fmt.Errorf("mech is listed in marketplace")
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("war machine is listed in marketplace")
		return err
	}

	if !mech.BattleReady {
		err := fmt.Errorf("mech is cannot be used")
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("war machine is not available for queuing")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(fmt.Errorf("missing warmachine faction"))
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return err
	}

	if !mech.IsDefault && mech.OwnerID != user.ID {
		return terror.Error(fmt.Errorf("does not own the mech"), "Current mech does not own by you")
	}

	// check mech is still in repair
	ar, err := boiler.MechRepairs(
		boiler.MechRepairWhere.MechID.EQ(mech.ID),
		boiler.MechRepairWhere.RepairCompleteAt.GT(time.Now()),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to check asset repair table")
	}

	if ar != nil {
		return terror.Error(fmt.Errorf("mech is still in repair center"), "Your mech is still in the repair center")
	}

	// Insert mech into queue
	existMech, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mechID.String())).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("check mech exists in queue")
		return terror.Error(err, "Failed to check whether mech is in the battle queue")
	}
	if existMech != nil {
		gamelog.L.Debug().Str("mech_id", mechID.String()).Err(err).Msg("mech already in queue")
		position, err := db.MechQueuePosition(mechID.String(), factionID)
		if err != nil {
			return terror.Error(err, "Already in queue, failed to get position. Contact support or try again.")
		}

		if position.QueuePosition == 0 {
			return terror.Error(fmt.Errorf("war machine already in battle"))
		}

		return terror.Error(fmt.Errorf("your mech is already in queue, current position is %d", position.QueuePosition))
	}

	// Get current queue length and calculate queue fee and reward
	result, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	queueStatus := CalcNextQueueStatus(result)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("unable to begin tx")
		return fmt.Errorf(terror.Echo(err))
	}
	defer tx.Rollback()

	bq := &boiler.BattleQueue{
		MechID:    mechID.String(),
		QueuedAt:  time.Now(),
		FactionID: factionID,
		OwnerID:   ownerID.String(),
	}

	err = bq.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Interface("mech", mech).
			Err(err).Msg("unable to insert mech into queue")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	// get faction user account
	factionAccountID, ok := server.FactionUsers[factionID]
	if !ok {
		gamelog.L.Error().Str("log_name", "battle arena").
			Str("mech ID", mech.ID).
			Str("faction ID", factionID).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	if ownerID.String() == factionAccountID {
		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("mech ID", mech.ID).
				Str("faction ID", factionID).
				Err(err).
				Msg("unable to save battle queue join for faction owned mech")
			return err
		}
		return nil
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Interface("mech", mech).
			Err(err).Msg("unable to commit mech insertion into queue")
		if bq.QueueFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
			}
		}
		if bq.QueueNotificationFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
			}
		}

		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	reply(QueueJoinHandlerResponse{
		Success: true,
		Code:    "",
	})

	// Send updated battle queue status to all subscribers
	go func() {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, CalcNextQueueStatus(queueStatus.QueueLength+1))

		queueDetails, err := db.MechArenaStatus(user.ID, mechID.String(), factionID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech arena status")
			return
		}

		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, mechID), WSPlayerAssetMechQueueSubscribe, queueDetails)
	}()

	return nil
}

const WSMechArenaStatusUpdate = "PLAYER:ASSET:MECH:STATUS:UPDATE"

type AssetUpdateRequest struct {
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

func (arena *Arena) AssetUpdateRequest(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	msg := &AssetUpdateRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	queueDetails, err := db.MechArenaStatus(user.ID, msg.Payload.MechID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Invalid request received.")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, msg.Payload.MechID), WSPlayerAssetMechQueueSubscribe, queueDetails)
	return nil
}

type QueueLeaveRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

const WSQueueLeave = "BATTLE:QUEUE:LEAVE"

func (arena *Arena) QueueLeaveHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	msg := &QueueLeaveRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mech, err := db.Mech(gamedb.StdConn, mechID.String())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if user.ID != mech.OwnerID {
		return terror.Error(terror.ErrForbidden, "Only the owners of the war machine can remove it from the queue.")
	}

	// Get queue position before deleting
	position, err := db.MechQueuePosition(mechID.String(), factionID)
	if errors.Is(sql.ErrNoRows, err) {
		gamelog.L.Warn().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("tried to remove already unqueued mech from queue")
		return terror.Warn(fmt.Errorf("unable to find war machine in battle queue, ensure machine isn't already removed and contact support"))
	}
	if err != nil {
		gamelog.L.Warn().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to get mech position")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if position.QueuePosition == 0 {
		return terror.Error(fmt.Errorf("cannot remove war machine from queue when it is in battle"))
	}

	// check current battle war machine id list
	for _, wmID := range arena.currentBattleWarMachineIDs() {
		if wmID == mechID {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("cannot remove battling mech from queue")
			return terror.Error(fmt.Errorf("cannot remove war machine from queue when it is in battle"), "You cannot remove war machines currently in battle.")
		}
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}
	defer tx.Rollback()

	// Remove from queue
	bq, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mechID.String())).One(tx)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Err(err).
			Str("mech_id", mechID.String()).
			Msg("unable to get existing mech from queue")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	_, err = bq.Delete(tx)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Interface("mech", mech).
			Err(err).Msg("unable to remove mech from queue")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Interface("mech", mech).
			Err(err).Msg("unable to commit mech deletion from queue")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	reply(true)

	result, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	// Send updated battle queue status to all subscribers
	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, CalcNextQueueStatus(result))
	gamelog.L.Debug().Str("factionID", factionID).Str("mechID", mechID.String()).Msg("published message on queue leave")

	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue-update", factionID), WSPlayerAssetMechQueueUpdateSubscribe, true)

	return nil
}

type QueueStatusResponse struct {
	QueueLength int64 `json:"queue_length"`
}

const WSQueueStatusSubscribe = "BATTLE:QUEUE:STATUS:SUBSCRIBE"

func (arena *Arena) QueueStatusSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	result, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("factionID", user.FactionID.String).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	reply(CalcNextQueueStatus(result))
	return nil
}

const WSPlayerAssetMechQueueUpdateSubscribe = "PLAYER:ASSET:MECH:QUEUE:UPDATE"
const WSPlayerAssetMechQueueSubscribe = "PLAYER:ASSET:MECH:QUEUE:SUBSCRIBE"

func (arena *Arena) PlayerAssetMechQueueSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	mechID := cctx.URLParam("mech_id")

	queueDetails, err := db.MechArenaStatus(user.ID, mechID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Invalid request received.")
	}

	reply(queueDetails)
	return nil
}
