package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func CalcNextQueueStatus(length int64) QueueStatusResponse {
	ql := float64(length + 1)
	queueLength := decimal.NewFromFloat(ql)

	// min cost will be one forth of the queue length
	minQueueCost := queueLength.Div(decimal.NewFromFloat(4)).Mul(decimal.New(1, 18))

	// calc queue cost
	feeMultiplier := math.Log(float64(ql)/3.25) * 0.25
	queueCost := queueLength.Mul(decimal.NewFromFloat(feeMultiplier)).Mul(decimal.New(1, 18))

	// calc contract reward
	contractReward := queueLength.Mul(decimal.New(2, 18))

	// fee never get under queue length
	if queueCost.LessThan(minQueueCost) {
		queueCost = minQueueCost
	}

	// length * 2 sups
	return QueueStatusResponse{
		QueueLength: length, // return the current queue length

		// the fee player have to pay if they want to queue their mech
		QueueCost: queueCost,

		// the reward, player will get if their mech won the battle
		ContractReward: contractReward,
	}
}

type QueueJoinHandlerResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
}

type QueueJoinRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash                   string `json:"asset_hash"`
		NeedInsured                 bool   `json:"need_insured"`
		EnablePushNotifications     bool   `json:"enable_push_notifications,omitempty"`
		MobileNumber                string `json:"mobile_number,omitempty"`
		EnableTelegramNotifications bool   `json:"enable_telegram_notifications"`
	} `json:"payload"`
}

const WSQueueJoin hub.HubCommandKey = "BATTLE:QUEUE:JOIN"

func (arena *Arena) QueueJoinHandler(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	msg := &QueueJoinRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return terror.Error(err)
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(fmt.Errorf("missing warmachine faction"))
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return terror.Error(err)
	}

	if mech.OwnerID != wsc.Identifier() {
		return terror.Error(fmt.Errorf("does not own the mech"), "Current mech does not own by you")
	}

	// Get current queue length and calculate queue fee and reward
	result, err := db.QueueLength(factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return terror.Error(err)
	}

	queueStatus := CalcNextQueueStatus(result)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return fmt.Errorf(terror.Echo(err))
	}
	defer tx.Rollback()

	var position int64

	// Insert mech into queue
	exists, err := boiler.BattleQueueExists(tx, mechID.String())
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("check mech exists in queue")
	}
	if exists {
		gamelog.L.Debug().Str("mech_id", mechID.String()).Err(err).Msg("mech already in queue")
		position, err = db.QueuePosition(mechID, factionID)
		if err != nil {
			return terror.Error(err, "Already in queue, failed to get position. Contact support or try again.")
		}
		reply(true)
		return nil
	}

	bc := &boiler.BattleContract{
		MechID:         mechID.String(),
		FactionID:      factionID.String(),
		PlayerID:       ownerID.String(),
		ContractReward: queueStatus.ContractReward,
		Fee:            queueStatus.QueueCost,
	}
	err = bc.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("contractReward", queueStatus.ContractReward.String()).
			Str("queueFee", queueStatus.QueueCost.String()).
			Err(err).Msg("unable to create battle contract")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	bq := &boiler.BattleQueue{
		MechID:           mechID.String(),
		QueuedAt:         time.Now(),
		FactionID:        factionID.String(),
		OwnerID:          ownerID.String(),
		BattleContractID: null.StringFrom(bc.ID),
	}

	notifications := msg.Payload.EnablePushNotifications || msg.Payload.MobileNumber != "" || msg.Payload.EnableTelegramNotifications
	if !notifications {
		bq.Notified = true
	}
	err = bq.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to insert mech into queue")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}
	factionAccountID, ok := server.FactionUsers[factionID.String()]
	if !ok {
		gamelog.L.Error().
			Str("mech ID", mech.ID).
			Str("faction ID", factionID.String()).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	// Charge user queue fee
	supTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
		Amount:               queueStatus.QueueCost.StringFixed(18),
		FromUserID:           ownerID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queueing_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		SubGroup:             "Queue",
		Description:          "Queued mech to battle arena",
		NotSafe:              true,
	})
	if err != nil {
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Interface("mechID", mechID).Interface("factionID", factionID.String()).Err(err).Msg("unable to charge user for insert mech into queue")
		return terror.Error(err, "Unable to process queue fee,  check your balance and try again.")
	}

	bq.QueueFeeTXID = null.StringFrom(supTransactionID)
	_, err = bq.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Str("tx_id", supTransactionID).
			Err(err).Msg("unable to update battle queue with queue transaction id")
		if bq.QueueFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
			}
		}

		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	shortcode := ""
	bqn := &boiler.BattleQueueNotification{}
	// Charge queue notification fee, if enabled (10% of queue cost)
	if !bq.Notified {
		notifyCost := queueStatus.QueueCost.Mul(decimal.NewFromFloat(0.1))
		notifyTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
			Amount:               notifyCost.String(),
			FromUserID:           ownerID,
			ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
			TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queue_notification_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupBattle),
			SubGroup:             "Queue",
			Description:          "Notification surcharge for queued mech in arena",
			NotSafe:              true,
		})
		if err != nil {
			gamelog.L.Error().Str("txID", notifyTransactionID).Err(err).Msg("unable to charge user for sms notification for mech in queue")

			if bq.QueueFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
				}
			}
			// Abort transaction if charge fails
			return terror.Error(err, "Unable to process notification fee, please check your balance and try again.")
		}
		bq.QueueNotificationFeeTXID = null.StringFrom(notifyTransactionID)
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("tx_id", notifyTransactionID).
				Err(err).Msg("unable to update battle queue with queue notification transaction id")
			if bq.QueueFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
				}
			}
			if bq.QueueNotificationFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
				}
			}

			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		//insert notification into db
		bqn = &boiler.BattleQueueNotification{
			MechID:            mechID.String(),
			QueueMechID:       null.StringFrom(mechID.String()),
			MobileNumber:      null.StringFrom(msg.Payload.MobileNumber),
			PushNotifications: msg.Payload.EnablePushNotifications,
			Fee:               notifyCost,
		}

		if msg.Payload.EnableTelegramNotifications {
			telegramNotification, err := arena.telegram.NotificationCreate(mechID.String(), bqn)
			if err != nil {
				gamelog.L.Error().
					Str("mechID", mechID.String()).
					Str("playerID", ownerID.String()).
					Err(err).Msg("unable to create telegram notification")
				if bq.QueueFeeTXID.Valid {
					_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
					if err != nil {
						gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
					}
				}
				if bq.QueueNotificationFeeTXID.Valid {
					_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
					if err != nil {
						gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
					}
				}
				return terror.Error(err, "Unable create telegram notification. Contact support.")
			}
			bqn.TelegramNotificationID = null.StringFrom(telegramNotification.ID)
			shortcode = telegramNotification.Shortcode
		}

		err = bqn.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Interface("mech", mech).
				Err(err).Msg("unable to insert queue notification for mech")
			if bq.QueueFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
				}
			}
			if bq.QueueNotificationFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
				}
			}

			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to commit mech insertion into queue")
		if bq.QueueFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
			}
		}
		if bq.QueueNotificationFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
			}
		}

		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	// Get mech current queue position
	position, err = db.QueuePosition(mechID, factionID)
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), AssetQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}
	if err != nil {
		gamelog.L.Error().
			Str("mechID", mechID.String()).
			Str("factionID", factionID.String()).
			Err(err).Msg("unable to retrieve mech queue position")
		return terror.Error(err, "Unable to join queue, check your balance and try again.")
	}

	// reply with shortcode if telegram notifs enabled
	if bqn.TelegramNotificationID.Valid && shortcode != "" {
		reply(QueueJoinHandlerResponse{
			Success: true,
			Code:    shortcode,
		})
	} else {
		reply(QueueJoinHandlerResponse{
			Success: true,
			Code:    "",
		})
	}

	// Send updated war machine queue status to subscriber
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), AssetQueueStatusResponse{
		&position,
		&queueStatus.ContractReward,
	})

	// Send updated battle queue status to all subscribers
	nextQueueStatus := CalcNextQueueStatus(queueStatus.QueueLength + 1)
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatusSubscribe, factionID.String())), nextQueueStatus)

	return nil
}

type QueueLeaveRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

const WSQueueLeave hub.HubCommandKey = "BATTLE:QUEUE:LEAVE"

func (arena *Arena) QueueLeaveHandler(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	msg := &QueueLeaveRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden, "Only the owners of the war machine can remove it from the queue.")
	}

	if userID != ownerID {
		return terror.Error(terror.ErrForbidden, "Only the owners of the war machine can remove it from the queue.")
	}

	originalQueueCost, err := db.QueueFee(mechID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to remove mech from queue")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	// Get queue position before deleting
	position, err := db.QueuePosition(mechID, factionID)
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		gamelog.L.Warn().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("tried to remove already unqueued mech from queue")
		return terror.Warn(err, "Unable to find war machine in battle queue, ensure machine isn't already removed and contact support.")
	}
	if err != nil {
		gamelog.L.Warn().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to get mech position")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if position == -1 {
		// If mech is currently in battle
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("cannot remove battling mech from queue")
		return terror.Error(fmt.Errorf("cannot remove war machine from queue when it is in battle"), "You cannot remove war machines currently in battle.")
	}

	canxq := `UPDATE battle_contracts SET cancelled = true WHERE id = (SELECT battle_contract_id FROM battle_queue WHERE mech_id = $1)`
	_, err = gamedb.StdConn.Exec(canxq, mechID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to cancel battle contract. mech has left queue though.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}
	defer tx.Rollback()

	// Remove from queue
	bq, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mechID.String())).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("mech_id", mechID.String()).
			Msg("unable to get existing mech from queue")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	factionAccountID, ok := server.FactionUsers[factionID.String()]
	if !ok {
		gamelog.L.Error().
			Str("mech ID", mech.ID).
			Str("faction ID", factionID.String()).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	// refund queue fee if not already refunded
	if !bq.QueueFeeTXIDRefund.Valid {
		// check if they have a transaction ID
		if bq.QueueFeeTXID.Valid && bq.QueueFeeTXID.String != "" {
			queueRefundTransactionID, err := arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().
					Str("queue_transaction_id", bq.QueueFeeTXID.String).
					Err(err).
					Msg("failed to refund users queue fee")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueFeeTXIDRefund = null.StringFrom(queueRefundTransactionID)
		} else {
			// TODO: Eventually all battle queues will have transaction ids to refund against, but legency queue will not. So keeping below until all legacy queues have passed
			// Refund user queue fee
			queueRefundTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
				Amount:               originalQueueCost.StringFixed(18),
				FromUserID:           uuid.Must(uuid.FromString(factionAccountID)),
				ToUserID:             ownerID,
				TransactionReference: server.TransactionReference(fmt.Sprintf("refund_war_machine_queueing_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             "Queue",
				Description:          "Refunded battle arena queueing fee",
				NotSafe:              true,
			})
			if err != nil {
				// Abort transaction if refund fails
				gamelog.L.Error().Str("txID", queueRefundTransactionID).Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to charge user for insert mech into queue")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueFeeTXIDRefund = null.StringFrom(queueRefundTransactionID)
		}
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("queue_refund_transaction_id", bq.QueueFeeTXIDRefund.String).
				Err(err).Msg("unable to update battle queue with refund transaction details")
			return terror.Error(err, "Unable to join queue, check your balance and try again.")
		}
	}

	updateBQNq := `UPDATE battle_queue_notifications SET is_refunded = true, queue_mech_id = null WHERE mech_id = $1`
	_, err = gamedb.StdConn.Exec(updateBQNq, mechID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to update battle_queue_notifications table during refund")
	}

	// Refund queue notification fee, if enabled and not already refunded
	if !bq.Notified && !bq.QueueNotificationFeeTXIDRefund.Valid {
		if bq.QueueNotificationFeeTXID.Valid && bq.QueueNotificationFeeTXID.String != "" {
			queueNotificationRefundTransactionID, err := arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
			if err != nil {
				gamelog.L.Error().
					Err(err).
					Str("queue_notification_transaction_id", bq.QueueNotificationFeeTXID.String).
					Msg("failed to refund users notification fee")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueNotificationFeeTXIDRefund = null.StringFrom(queueNotificationRefundTransactionID)
		} else {
			// TODO: Eventually all battle queues will have transaction ids to refund against, but legency queue will not. So keeping below until all legacy queues have passed
			notifyCost := originalQueueCost.Mul(decimal.NewFromFloat(0.1))
			queueNotificationRefundTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
				Amount:               notifyCost.StringFixed(18),
				FromUserID:           uuid.Must(uuid.FromString(factionAccountID)),
				ToUserID:             ownerID,
				TransactionReference: server.TransactionReference(fmt.Sprintf("refund_war_machine_queue_notification_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             "Queue",
				Description:          "Refunded notification surcharge for queued mech in arena",
				NotSafe:              true,
			})
			if err != nil {
				// Abort transaction if charge fails
				gamelog.L.Error().Str("txID", queueNotificationRefundTransactionID).Err(err).Msg("unable to refund user for notification for mech in queue")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueNotificationFeeTXIDRefund = null.StringFrom(queueNotificationRefundTransactionID)
		}
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("queue_notification_refund_transaction_id", bq.QueueNotificationFeeTXIDRefund.String).
				Err(err).Msg("unable to update battle queue with notification refund transaction details")
			return terror.Error(err, "Unable to leave queue, try again or contact support.")
		}
	}

	_, err = bq.Delete(tx)
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to remove mech from queue")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to commit mech deletion from queue")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	reply(true)

	// Tell clients to refetch war machine queue status
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueUpdatedSubscribe, factionID.String())), true)

	result, err := db.QueueLength(factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}
	nextQueueStatus := CalcNextQueueStatus(result)

	// Send updated Battle queue status to all subscribers
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatusSubscribe, factionID.String())), nextQueueStatus)

	return nil
}

type QueueStatusResponse struct {
	QueueLength    int64           `json:"queue_length"`
	QueueCost      decimal.Decimal `json:"queue_cost"`
	ContractReward decimal.Decimal `json:"contract_reward"`
}

const WSQueueStatusSubscribe hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:STATUS:SUBSCRIBE")

func (arena *Arena) QueueStatusSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find faction from user id")
		return "", "", terror.Error(err)
	}

	if needProcess {
		result, err := db.QueueLength(factionID)
		if err != nil {
			gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
			return "", "", terror.Error(err)
		}

		reply(CalcNextQueueStatus(result))
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatusSubscribe, factionID.String())), nil
}

const WSQueueUpdatedSubscribe hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:UPDATED")

func (arena *Arena) QueueUpdatedSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find faction from user id")
		return "", "", terror.Error(err)
	}

	if needProcess {
		reply(true)
	}
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueUpdatedSubscribe, factionID)), nil
}

type AssetQueueStatusRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

type AssetQueueStatusResponse struct {
	QueuePosition  *int64           `json:"queue_position"` // in-game: -1; in queue: > 0; not in queue: nil
	ContractReward *decimal.Decimal `json:"contract_reward"`
}

const WSAssetQueueStatus hub.HubCommandKey = hub.HubCommandKey("ASSET:QUEUE:STATUS")

func (arena *Arena) AssetQueueStatusHandler(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	req := &AssetQueueStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	mechID, err := db.MechIDFromHash(req.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", req.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(err)
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return terror.Error(err)
	}

	mechFactionID, err := GetPlayerFactionID(ownerID)
	if err != nil || mechFactionID.IsNil() {
		gamelog.L.Error().Str("userID", ownerID.String()).Err(err).Msg("unable to find faction from owner id")
		return terror.Error(err)
	}

	position, err := db.QueuePosition(mechID, mechFactionID)
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		reply(AssetQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}
	if err != nil {
		return terror.Error(err)
	}

	contractReward, err := db.QueueContract(mechID, mechFactionID)
	if err != nil {
		gamelog.L.Error().Str("mechID", mechID.String()).Str("mechFactionID", mechFactionID.String()).Err(err).Msg("unable to get contract reward")
		return terror.Error(err)
	}

	reply(AssetQueueStatusResponse{
		&position,
		contractReward,
	})

	return nil
}

const WSAssetQueueStatusList hub.HubCommandKey = hub.HubCommandKey("ASSET:QUEUE:STATUS:LIST")

type AssetQueueStatusItem struct {
	MechID        string `json:"mech_id"`
	QueuePosition int64  `json:"queue_position"`
}

func (arena *Arena) AssetQueueStatusListHandler(ctx context.Context, hub *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error {
	userID, err := uuid.FromString(hub.Identifier())
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	queueList, err := db.QueueOwnerList(userID)
	if err != nil {
		return terror.Error(err, "Failed to list mechs in queue")
	}

	resp := []*AssetQueueStatusItem{}
	for _, q := range queueList {
		obj := &AssetQueueStatusItem{
			MechID:        q.MechID.String(),
			QueuePosition: q.QueuePosition,
		}
		resp = append(resp, obj)
	}

	reply(resp)

	return nil
}

const WSAssetQueueStatusSubscribe hub.HubCommandKey = hub.HubCommandKey("ASSET:QUEUE:STATUS:SUBSCRIBE")

func (arena *Arena) AssetQueueStatusSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &AssetQueueStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	mechID, err := db.MechIDFromHash(req.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", req.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return "", "", terror.Error(err)
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return "", "", terror.Error(err)
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return "", "", terror.Error(err)
	}

	if mech.OwnerID != wsc.Identifier() {
		gamelog.L.Warn().Str("player id", wsc.Identifier()).Str("mech id", mechID.String()).Msg("Someone attempt to subscribe on a mech's queuing status which is not belong to them")
		return "", "", terror.Error(terror.ErrForbidden, "Cannot subscribe on mech which is not belong to you")
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return "", "", terror.Error(err)
	}

	factionID, err := GetPlayerFactionID(ownerID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", ownerID.String()).Err(err).Msg("unable to find faction from owner id")
		return "", "", terror.Error(err)
	}

	if needProcess {
		position, err := db.QueuePosition(mechID, factionID)
		if errors.Is(sql.ErrNoRows, err) {
			// If mech is not in queue
			reply(AssetQueueStatusResponse{
				nil,
				nil,
			})
			return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), nil
		}
		if err != nil {
			gamelog.L.Error().Str("mechID", mechID.String()).Str("factionID", factionID.String()).Err(err).Msg("unable to get mech queue position")
			return "", "", terror.Error(err)
		}

		contractReward, err := db.QueueContract(mechID, factionID)
		if err != nil {
			gamelog.L.Error().Str("mechID", mechID.String()).Str("factionID", factionID.String()).Err(err).Msg("unable to get contract reward")
			return "", "", terror.Error(err)
		}

		reply(AssetQueueStatusResponse{
			&position,
			contractReward,
		})
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), nil
}
