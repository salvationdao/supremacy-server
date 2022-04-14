package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math"
	"math/rand"
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

	mul := db.GetDecimalWithDefault("queue_fee_log_multi", decimal.NewFromFloat(3.25))
	mulFloat, _ := mul.Float64()

	if server.Env() == "staging" {
		mulFloat = 0.2 + rand.Float64()*(8.0-0.2)
		minQueueCost = decimal.NewFromFloat(1.5)
	}

	// calc queue cost
	feeMultiplier := math.Log(float64(ql)) / mulFloat * 0.25
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

	// check mobile phone, if player required notifyed through mobile sms
	if msg.Payload.EnablePushNotifications && msg.Payload.MobileNumber != "" {
		mobileNumber, err := arena.sms.Lookup(msg.Payload.MobileNumber)
		if err != nil {
			gamelog.L.Warn().Str("mobile number", msg.Payload.MobileNumber).Msg("Failed to lookup mobile number through twilio api")
			return terror.Error(err)
		}

		// set the verifyed mobile number
		msg.Payload.MobileNumber = mobileNumber
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

	// check mech is still in repair
	ar, err := boiler.AssetRepairs(
		boiler.AssetRepairWhere.MechID.EQ(mech.ID),
		boiler.AssetRepairWhere.RepairCompleteAt.GT(time.Now()),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to check asset repair table")
	}

	if ar != nil {
		return terror.Error(fmt.Errorf("mech is still in repair center"), "Your mech is still in the repair center")
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
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
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

	// Tell clients to refetch war machine queue status
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueUpdatedSubscribe, factionID.String())), true)

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

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}
	defer tx.Rollback()

	canxq := `UPDATE battle_contracts SET cancelled = true WHERE id = (SELECT battle_contract_id FROM battle_queue WHERE mech_id = $1)`
	_, err = tx.Exec(canxq, mechID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to cancel battle contract. mech has left queue though.")
	}

	// Remove from queue
	bq, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mechID.String())).One(tx)
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
			factionAccUUID, _ := uuid.FromString(factionAccountID)
			syndicateBalance := arena.RPCClient.UserBalanceGet(factionAccUUID)

			if syndicateBalance.LessThanOrEqual(*originalQueueCost) {
				txid, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
					FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
					ToUserID:             factionAccUUID,
					Amount:               originalQueueCost.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("queue_fee_reversal_shortfall|%s|%d", bq.QueueFeeTXID.String, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupBattle),
					SubGroup:             "Queue",
					Description:          "Queue reversal shortfall",
					NotSafe:              false,
				})
				if err != nil {
					gamelog.L.Error().
						Str("Faction ID", factionAccountID).
						Str("Amount", originalQueueCost.StringFixed(0)).
						Err(err).
						Msg("Could not transfer money from treasury into syndicate account!!")
					return terror.Error(err, "Unable to remove your mech from the queue. Please contact support.")
				}
				gamelog.L.Warn().
					Str("Faction ID", factionAccountID).
					Str("Amount", originalQueueCost.StringFixed(0)).
					Str("TXID", txid).
					Err(err).
					Msg("Had to transfer funds to the syndicate account")
			}

			queueRefundTransactionID, err := arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().
					Str("queue_transaction_id", bq.QueueFeeTXID.String).
					Err(err).
					Msg("failed to refund users queue fee")
				return terror.Error(err, "Unable to remove your mech from the queue, please try again in five minutes or contact support.")
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

	if req.Payload.AssetHash == "" {
		return "", "", terror.Warn(fmt.Errorf("empty asset hash"), "Empty asset data, please try again or contact support.")
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

type AssetQueueManyRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		PageNumber int `json:"page_number"`
		PageSize   int `json:"page_size"`
	} `json:"payload"`
}

type AssetQueueManyResponse struct {
	AssetQueueList []*AssetQueue `json:"asset_queue_list"`
	Total          int           `json:"total"`
}

type AssetQueue struct {
	MechID           string          `json:"mech_id"`
	Hash             string          `json:"hash"`
	Position         *int64          `json:"position"`
	ContractReward   decimal.Decimal `json:"contract_reward"`
	InBattle         bool            `json:"in_battle"`
	BattleContractID string
}

const HubKeyAssetMany hub.HubCommandKey = hub.HubCommandKey("ASSET:MANY")

func (arena *Arena) AssetManyHandler(ctx context.Context, hubc *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error {
	req := &AssetQueueManyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp := &AssetQueueManyResponse{
		AssetQueueList: []*AssetQueue{},
	}

	// get the list of player's mechs (id, hash, created_at)
	mechs, err := boiler.Mechs(
		qm.Select(boiler.MechColumns.ID, boiler.MechColumns.Hash, boiler.MechColumns.CreatedAt),
		boiler.MechWhere.OwnerID.EQ(hubc.Identifier()),
		qm.OrderBy(boiler.MechColumns.CreatedAt),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player id", hubc.Identifier()).Err(err).Msg("Failed to get player's mechs")
		return terror.Error(err, "Failed to get mech data")
	}

	// reply empty
	if len(mechs) == 0 {
		reply(resp)
		return nil
	}

	// set total
	resp.Total = len(mechs)

	// calc mech id list
	mechIDs := []string{}
	for _, mech := range mechs {
		mechIDs = append(mechIDs, mech.ID)
	}

	// get queue position
	queuePosition, err := db.MechQueuePosition(userFactionID.String(), hubc.Identifier())
	if err != nil {
		gamelog.L.Error().Str("player id", hubc.Identifier()).Err(err).Msg("Failed to get player mech position")
		return terror.Error(err, "Failed to get mech position")
	}

	// get player's in-battle mech
	bqs, err := boiler.BattleQueues(
		qm.Select(
			boiler.BattleQueueColumns.ID,
			boiler.BattleQueueColumns.BattleContractID,
			boiler.BattleQueueColumns.MechID,
		),
		boiler.BattleQueueWhere.OwnerID.EQ(hubc.Identifier()),
		boiler.BattleQueueWhere.BattleID.IsNotNull(),
		boiler.BattleQueueWhere.BattleContractID.IsNotNull(),
		qm.Load(boiler.BattleQueueRels.Mech),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get player's in-battle mechs")
	}

	// manually handle mech pagination

	// resort the order
	newList := []*AssetQueue{}

	// insert in-battle mech
	for _, bq := range bqs {
		newList = append(newList, &AssetQueue{
			MechID:           bq.MechID,
			Hash:             bq.R.Mech.Hash,
			InBattle:         true,
			BattleContractID: bq.BattleContractID.String,
		})

	}

	// fill queued mech
	for _, qp := range queuePosition {
		newList = append(newList, &AssetQueue{
			MechID:           qp.MechID.String(),
			Position:         &qp.QueuePosition,
			BattleContractID: qp.BattleContractID,
		})
	}

	// fill mech detail
	for _, mech := range mechs {
		existOnIndex := -1

		// check whether it is in queue
		for i, nl := range newList {
			if mech.ID == nl.MechID {
				existOnIndex = i
				break
			}
		}

		if existOnIndex >= 0 {
			// update mech hash
			newList[existOnIndex].Hash = mech.Hash
		} else {
			// append to new list
			newList = append(newList, &AssetQueue{
				MechID: mech.ID,
				Hash:   mech.Hash,
			})
		}
	}

	offset := req.Payload.PageNumber * req.Payload.PageSize
	limit := req.Payload.PageSize

	for i, nl := range newList {
		if i < offset {
			continue
		}

		resp.AssetQueueList = append(resp.AssetQueueList, nl)

		if len(resp.AssetQueueList) >= limit {
			break
		}
	}

	// fetch battle contract for contract reward
	battleContractIDs := []string{}
	for _, bc := range resp.AssetQueueList {
		if bc.BattleContractID != "" {
			battleContractIDs = append(battleContractIDs, bc.BattleContractID)
		}
	}

	// if there contain any battle contract id
	if len(battleContractIDs) > 0 {
		bcs, err := boiler.BattleContracts(
			qm.Select(
				boiler.BattleContractColumns.ID,
				boiler.BattleContractColumns.MechID,
				boiler.BattleContractColumns.ContractReward,
			),
			boiler.BattleContractWhere.ID.IN(battleContractIDs),
		).All(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to get battle contract")
		}

		for _, bc := range bcs {
			for _, asset := range resp.AssetQueueList {
				if asset.MechID == bc.MechID {
					asset.ContractReward = bc.ContractReward
					break
				}
			}
		}
	}

	reply(resp)

	return nil
}
