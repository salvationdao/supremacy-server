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
	"server/xsyn_rpcclient"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func CalcNextQueueStatus(length int64) QueueStatusResponse {
	ql := float64(length + 1)
	queueLength := decimal.NewFromFloat(ql)

	// min cost will be one forth of the queue length
	minQueueCost := queueLength.Div(decimal.NewFromFloat(4)).Mul(decimal.New(1, 18))

	mul := db.GetDecimalWithDefault("queue_fee_log_multi", decimal.NewFromFloat(3.25))
	mulFloat, _ := mul.Float64()

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
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	// onChainStatus, err := arena.RPCClient.AssetOnChainStatus(mechID.String())
	// if err != nil {
	// 	return terror.Error(err, "Unable to get asset ownership details, please try again or contact support.")
	// }

	// if onChainStatus != server.OnChainStatusMintable && onChainStatus != server.OnChainStatusUnstakable {
	// 	return terror.Error(fmt.Errorf("asset on chain status is %s", onChainStatus), "This asset isn't on world, please transition on world.")
	// }

	mech, err := db.Mech(mechID.String())
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(fmt.Errorf("missing warmachine faction"))
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
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
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("check mech exists in queue")
		return terror.Error(err, "Failed to check whether mech is in the battle queue")
	}
	if existMech != nil {
		gamelog.L.Debug().Str("mech_id", mechID.String()).Err(err).Msg("mech already in queue")
		position, err := db.QueuePosition(mechID, uuid.FromStringOrNil(factionID))
		if err != nil {
			return terror.Error(err, "Already in queue, failed to get position. Contact support or try again.")
		}

		if position == -1 {
			return terror.Error(terror.ErrInvalidInput, "Your mech is in battle.")
		}

		return terror.Error(terror.ErrInvalidInput, fmt.Sprintf("Your mech is already in queue, current position is %d.", position))
	}

	// Get current queue length and calculate queue fee and reward
	result, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	queueStatus := CalcNextQueueStatus(result)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return fmt.Errorf(terror.Echo(err))
	}
	defer tx.Rollback()

	bc := &boiler.BattleContract{
		MechID:         mechID.String(),
		FactionID:      factionID,
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
		FactionID:        factionID,
		OwnerID:          ownerID.String(),
		BattleContractID: null.StringFrom(bc.ID),
	}

	err = bq.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to insert mech into queue")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	// get faction user account
	factionAccountID, ok := server.FactionUsers[factionID]
	if !ok {
		gamelog.L.Error().
			Str("mech ID", mech.ID).
			Str("faction ID", factionID).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	if ownerID.String() == factionAccountID {
		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().
				Str("mech ID", mech.ID).
				Str("faction ID", factionID).
				Err(err).
				Msg("unable to save battle queue join for faction owned mech")
			return err
		}
		return nil
	}

	// Charge user queue fee
	supTransactionID, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		Amount:               queueStatus.QueueCost.String(),
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
		gamelog.L.Error().Str("txID", supTransactionID).Interface("mechID", mechID).Interface("factionID", factionID).Err(err).Msg("unable to charge user for insert mech into queue")
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
	if err != nil && !errors.Is(sql.ErrNoRows, err) {
		gamelog.L.Error().
			Str("mechID", mechID.String()).
			Str("factionID", factionID).
			Err(err).Msg("unable to retrieve mech queue position")
		return terror.Error(err, "Unable to join queue, check your balance and try again.")
	}

	// Tell clients to refetch war machine queue status
	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueUpdatedSubscribe, true)

	reply(QueueJoinHandlerResponse{
		Success: true,
		Code:    "",
	})

	// Send updated battle queue status to all subscribers
	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, CalcNextQueueStatus(queueStatus.QueueLength+1))

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
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mech, err := db.Mech(mechID.String())
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if user.ID != mech.OwnerID {
		return terror.Error(terror.ErrForbidden, "Only the owners of the war machine can remove it from the queue.")
	}

	originalQueueCost, err := db.QueueFee(mechID, uuid.FromStringOrNil(factionID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to remove mech from queue")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	// Get queue position before deleting
	position, err := db.QueuePosition(mechID, uuid.FromStringOrNil(factionID))
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

	// check current battle war machine id list
	for _, wmID := range arena.currentBattleWarMachineIDs() {
		if wmID == mechID {
			gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("cannot remove battling mech from queue")
			return terror.Error(fmt.Errorf("cannot remove war machine from queue when it is in battle"), "You cannot remove war machines currently in battle.")
		}
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}
	defer tx.Rollback()

	canxq := `UPDATE battle_contracts SET cancelled = TRUE WHERE id = (SELECT battle_contract_id FROM battle_queue WHERE mech_id = $1)`
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

	factionAccountID, ok := server.FactionUsers[factionID]
	if !ok {
		gamelog.L.Error().
			Str("mech ID", mech.ID).
			Str("faction ID", factionID).
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
				txid, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
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
		}
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("queue_refund_transaction_id", bq.QueueFeeTXIDRefund.String).
				Err(err).Msg("unable to update battle queue with refund transaction details")
			return terror.Error(err, "Unable to join queue, check your balance and try again.")
		}
	}

	updateBQNq := `UPDATE battle_queue_notifications SET is_refunded = TRUE, queue_mech_id = NULL WHERE mech_id = $1`
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
	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueUpdatedSubscribe, true)

	result, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	// Send updated Battle queue status to all subscribers
	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, CalcNextQueueStatus(result))

	return nil
}

type QueueStatusResponse struct {
	QueueLength    int64           `json:"queue_length"`
	QueueCost      decimal.Decimal `json:"queue_cost"`
	ContractReward decimal.Decimal `json:"contract_reward"`
}

const WSQueueStatusSubscribe = "BATTLE:QUEUE:STATUS:SUBSCRIBE"

func (arena *Arena) QueueStatusSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	result, err := db.QueueLength(uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Interface("factionID", user.FactionID.String).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	reply(CalcNextQueueStatus(result))

	return nil
}

const WSQueueUpdatedSubscribe = "BATTLE:QUEUE:UPDATED"

type AssetQueueStatusRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

type AssetQueueStatusResponse struct {
	QueuePosition  *int64           `json:"queue_position"` // in-game: -1; in queue: > 0; not in queue: nil
	ContractReward *decimal.Decimal `json:"contract_reward"`
}

const WSAssetQueueStatus = "ASSET:QUEUE:STATUS"

func (arena *Arena) AssetQueueStatusHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetQueueStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	mechID, err := db.MechIDFromHash(req.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", req.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	mech, err := db.Mech(mechID.String())
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return err
	}

	if mech.OwnerID != user.ID {
		gamelog.L.Error().Str("mech_id", mechID.String()).Str("mech owner id", mech.OwnerID).Str("player id", user.ID).Err(err).Msg("player does not own the mech")
		return err
	}

	position, err := db.QueuePosition(mechID, uuid.FromStringOrNil(factionID))
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		reply(AssetQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}
	if err != nil {
		return err
	}

	contractReward, err := db.QueueContract(mechID, uuid.FromStringOrNil(factionID))
	if err != nil {
		gamelog.L.Error().Str("mech id", mechID.String()).Str("faction id", factionID).Err(err).Msg("unable to get contract reward")
		return err
	}

	reply(AssetQueueStatusResponse{
		&position,
		contractReward,
	})

	return nil
}

const WSAssetQueueStatusList = "ASSET:QUEUE:STATUS:LIST"

type AssetQueueStatusItem struct {
	MechID        string `json:"mech_id"`
	QueuePosition int64  `json:"queue_position"`
}

func (arena *Arena) AssetQueueStatusListHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	queueList, err := db.QueueOwnerList(uuid.FromStringOrNil(user.ID))
	if err != nil {
		return terror.Error(err, "Failed to list war machines in queue")
	}

	// check their on world / off world status
	assetIDsToCheck := []string{}
	for _, q := range queueList {
		assetIDsToCheck = append(assetIDsToCheck, q.MechID.String())
	}

	assetMap, err := arena.RPCClient.AssetsOnChainStatus(assetIDsToCheck)
	if err != nil {
		return terror.Error(err, "Unable to get asset ownership details, please try again or contact support.")
	}

	resp := []*AssetQueueStatusItem{}
	for _, q := range queueList {
		if onChainStatus, ok := assetMap[q.MechID.String()]; ok && (onChainStatus == server.OnChainStatusMintable || onChainStatus == server.OnChainStatusUnstakable) {
			obj := &AssetQueueStatusItem{
				MechID:        q.MechID.String(),
				QueuePosition: q.QueuePosition,
			}
			resp = append(resp, obj)
		}
	}

	reply(resp)
	return nil
}

type AssetQueueManyRequest struct {
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

const HubKeyAssetMany = "ASSET:MANY"

// THIS IS A LEGACY HANDLER, will be replaced
func (arena *Arena) AssetManyHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetQueueManyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp := &AssetQueueManyResponse{
		AssetQueueList: []*AssetQueue{},
	}

	type mechDetailsBrief struct {
		ID   string
		Hash string
	}

	var allMechs []*mechDetailsBrief

	// get the list of player's mechs (id, hash, created_at)
	query := `	
		SELECT m.id, ci.hash
		FROM mechs m
		INNER JOIN collection_items ci on ci.item_id = m.id
		WHERE owner_id = $1
		`

	rows, err := gamedb.StdConn.Query(query, user.ID)
	if err != nil {
		return terror.Error(err, "Issue retrieving your mechs, please try again or contact support.")
	}
	defer rows.Close()
	for rows.Next() {
		newMech := &mechDetailsBrief{}
		err := rows.Scan(&newMech.ID, &newMech.Hash)
		if err != nil {
			gamelog.L.Error().Err(err).Str("query", query).Str("hubc.Identifier()", user.ID).Msg("unable to scan mech details into struct")
			return terror.Error(err, "Issue composing your mechs, please try again or contact support.")
		}
		allMechs = append(allMechs, newMech)
	}

	var mechs []*mechDetailsBrief

	// reply empty
	if len(allMechs) == 0 {
		reply(resp)
		return nil
	}

	// calc mech id list
	mechIDs := []string{}
	for _, mech := range allMechs {
		mechIDs = append(mechIDs, mech.ID)
	}

	assetMap, err := arena.RPCClient.AssetsOnChainStatus(mechIDs)
	if err != nil {
		return terror.Error(err, "Unable to get asset ownership details, please try again or contact support.")
	}

	for _, m := range allMechs {
		if onChainStatus, ok := assetMap[m.ID]; ok && (onChainStatus == server.OnChainStatusMintable || onChainStatus == server.OnChainStatusUnstakable) {
			mechs = append(mechs, m)
		}
	}

	// set total
	resp.Total = len(mechs)

	// get queue position
	queuePosition, err := db.MechQueuePositions(factionID, user.ID)
	if err != nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player mech position")
		return terror.Error(err, "Failed to get mech position")
	}

	// get player's in-battle mech
	inBattleMechs, err := boiler.BattleQueues(
		qm.Select(
			boiler.BattleQueueColumns.ID,
			boiler.BattleQueueColumns.BattleContractID,
			boiler.BattleQueueColumns.MechID,
		),
		boiler.BattleQueueWhere.OwnerID.EQ(user.ID),
		boiler.BattleQueueWhere.BattleID.IsNotNull(),
		boiler.BattleQueueWhere.BattleContractID.IsNotNull(),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get player's in-battle mechs")
	}

	// manually handle mech pagination

	// resort the order
	newList := []*AssetQueue{}

	// insert in-battle mech
	for _, bq := range inBattleMechs {
		for _, m := range mechs {
			if m.ID == bq.MechID {
				newList = append(newList, &AssetQueue{
					MechID:           m.ID,
					Hash:             m.Hash,
					InBattle:         true,
					BattleContractID: bq.BattleContractID.String,
				})
			}
		}
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
