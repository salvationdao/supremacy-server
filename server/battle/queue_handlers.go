package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
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

func CalcNextQueueStatus(length int64) QueueStatusResponse {
	return QueueStatusResponse{
		QueueLength: length, // return the current queue length
		QueueCost:   db.GetDecimalWithDefault(db.KeyBattleQueueFee, decimal.New(250, 18)),
	}
}

const WSQueueJoin = "BATTLE:QUEUE:JOIN"

func (arena *Arena) QueueJoinHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	msg := &QueueJoinRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	mci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.Hash.EQ(msg.Payload.AssetHash),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech collection item from hash")
		return err
	}

	if mci.XsynLocked {
		err := fmt.Errorf("mech is locked to xsyn locked")
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is xsyn locked")
		return err
	}

	if mci.LockedToMarketplace {
		err := fmt.Errorf("mech is listed in marketplace")
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is listed in marketplace")
		return err
	}

	battleReady, err := db.MechBattleReady(mci.ItemID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle ready status")
		return err
	}

	if !battleReady {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is not available for queuing")
		return fmt.Errorf("mech is cannot be used")
	}

	if mci.OwnerID != user.ID {
		return terror.Error(fmt.Errorf("does not own the mech"), "This mech is not owned by you")
	}

	// Check mech exist in the battle queue
	existMech, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mci.ItemID)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("check mech exists in queue")
		return terror.Error(err, "Failed to check whether mech is in the battle queue")
	}
	if existMech != nil {
		gamelog.L.Debug().Str("mech_id", mci.ItemID).Err(err).Msg("mech already in queue")
		position, err := db.MechQueuePosition(mci.ItemID, factionID)
		if err != nil {
			return terror.Error(err, "Already in queue, failed to get position. Contact support or try again.")
		}

		if position.QueuePosition == 0 {
			return terror.Error(fmt.Errorf("war machine already in battle"))
		}

		return terror.Error(fmt.Errorf("your mech is already in queue, current position is %d", position.QueuePosition))
	}

	// check mech is still in repair
	rc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mci.ItemID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
		qm.Load(boiler.RepairCaseRels.RepairOffers),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mech id", mci.ItemID).Msg("Failed to get repair case")
		return terror.Error(err, "Failed to queue mech.")
	}

	if rc != nil {
		canDeployRatio := db.GetDecimalWithDefault(db.KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))

		// broadcast current mech stat if repair is above can deploy ratio
		if decimal.NewFromInt(int64(rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(rc.BlocksRequiredRepair))).LessThan(canDeployRatio) {
			// if mech has more than half of the block to repair
			return terror.Error(fmt.Errorf("mech is not fully recovered"), "Your mech is still under repair.")
		}
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

	bqf := &boiler.BattleQueueFee{
		MechID:   mci.ItemID,
		PaidByID: user.ID,
		Amount:   db.GetDecimalWithDefault(db.KeyBattleQueueFee, decimal.New(250, 18)),
	}

	err = bqf.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert battle queue fee.")
	}

	bq := &boiler.BattleQueue{
		MechID:    mci.ItemID,
		QueuedAt:  time.Now(),
		FactionID: factionID,
		OwnerID:   mci.OwnerID,
		FeeID:     null.StringFrom(bqf.ID),
	}

	err = bq.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Interface("mech id", mci.ItemID).
			Err(err).Msg("unable to insert mech into queue")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	// pay battle queue fee
	paidTxID, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(user.ID)),
		ToUserID:             uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
		Amount:               bqf.Amount.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("battle_queue_fee|%s|%d", mci.ItemID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          "queue mech to join the battle arena.",
	})
	if err != nil {
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("mech id", mci.ItemID).
			Str("amount", bqf.Amount.StringFixed(0)).
			Err(err).Msg("Failed to pay sups on queuing mech.")
		return terror.Error(err, "Failed to pay sups on queuing mech.")
	}

	refundFunc := func() {
		_, err = arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
			ToUserID:             uuid.Must(uuid.FromString(user.ID)),
			Amount:               bqf.Amount.StringFixed(0),
			TransactionReference: server.TransactionReference(fmt.Sprintf("refund_battle_queue_fee|%s|%d", mci.ItemID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupBattle),
			Description:          "refund the mech queuing fee.",
		})
		if err != nil {
			gamelog.L.Error().
				Str("player_id", user.ID).
				Str("mech id", mci.ItemID).
				Str("amount", bqf.Amount.StringFixed(0)).
				Err(err).Msg("Failed to refund sups on queuing mech.")
		}
	}

	// do not return, if error occur.
	bq.QueueFeeTXID = null.StringFrom(paidTxID)
	_, err = bq.Update(tx, boil.Whitelist(boiler.BattleQueueColumns.QueueFeeTXID))
	if err != nil {
		refundFunc() // refund player
		gamelog.L.Error().Err(err).Msg("Failed to record queue fee tx id")
		return terror.Error(err, "Failed to update battle queue")
	}

	if rc != nil {
		// otherwise, cancel all the existing offer
		if rc.R != nil && rc.R.RepairOffers != nil {
			ids := []string{}
			for _, ro := range rc.R.RepairOffers {
				ids = append(ids, ro.ID)
			}

			select {
			case arena.RepairOfferCloseChan <- &RepairOfferClose{
				OfferIDs:          ids,
				OfferClosedReason: boiler.RepairFinishReasonSTOPPED,
				AgentClosedReason: boiler.RepairAgentFinishReasonEXPIRED,
			}:
			case <-time.After(5 * time.Second):
				return terror.Error(fmt.Errorf("failed to terminate repair case"), "Failed to close the repair case of the mech.")
			}
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Interface("mech id", mci.ItemID).
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

	bqf.PaidTXID = null.StringFrom(paidTxID)
	_, err = bqf.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueFeeColumns.PaidTXID))
	if err != nil {
		gamelog.L.Error().Interface("battle queue fee", bqf).Err(err).Msg("Failed to update battle queue fee transaction id")
	}

	reply(QueueJoinHandlerResponse{
		Success: true,
		Code:    "",
	})

	// Send updated battle queue status to all subscribers
	go func() {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, CalcNextQueueStatus(queueStatus.QueueLength+1))

		queueDetails, err := db.MechArenaStatus(user.ID, mci.ItemID, factionID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech arena status")
			return
		}

		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, mci.ItemID), WSPlayerAssetMechQueueSubscribe, queueDetails)
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

type QueueStatusResponse struct {
	QueueLength int64           `json:"queue_length"`
	QueueCost   decimal.Decimal `json:"queue_cost"`
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
