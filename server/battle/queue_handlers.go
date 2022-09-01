package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/battle_queue"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type QueueJoinHandlerResponse struct {
	Success bool `json:"success"`
}

type QueueJoinRequest struct {
	Payload struct {
		MechIDs []string `json:"mech_ids"`
	} `json:"payload"`
}

func CalcNextQueueStatus(factionID string) {
	l := gamelog.L.With().Str("func", "CalcNextQueueStatus").Str("factionID", factionID).Logger()

	mwts, err := db.GetMinimumQueueWaitTimeSecondsFromFactionID(factionID)
	if err != nil {
		l.Warn().Err(err).Msg("unable to retrieve estimated queue time")
	}

	abl, err := db.GetAverageBattleLengthSeconds()
	if err != nil {
		l.Warn().Err(err).Msg("unable to retrieve average game length")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, QueueStatusResponse{
		MinimumWaitTimeSeconds:   mwts,
		AverageGameLengthSeconds: abl,
		QueueCost:                db.GetDecimalWithDefault(db.KeyBattleQueueFee, decimal.New(250, 18)),
	})
}

const WSQueueJoin = "BATTLE:QUEUE:JOIN"

func (am *ArenaManager) QueueJoinHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &QueueJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	if len(req.Payload.MechIDs) == 0 {
		return terror.Error(fmt.Errorf("mech id list not provided"), "Mech id list is not provided.")
	}

	mcis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(req.Payload.MechIDs),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech ids", req.Payload.MechIDs).Err(err).Msg("unable to retrieve mech collection item from hash")
		return err
	}

	if len(mcis) != len(req.Payload.MechIDs) {
		return terror.Error(fmt.Errorf("contain non-mech assest"), "The list contains non-mech asset.")
	}

	for _, mci := range mcis {
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
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Msg("war machine is not available for queuing")
			return fmt.Errorf("mech is cannot be used")
		}

		if mci.OwnerID != user.ID {
			return terror.Error(fmt.Errorf("does not own the mech"), "This mech is not owned by you")
		}
	}

	// Check if any of the mechs exist in the battle queue backlog
	backlogMech, err := boiler.BattleQueueBacklogs(boiler.BattleQueueBacklogWhere.MechID.IN(req.Payload.MechIDs)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech_ids", req.Payload.MechIDs).Err(err).Msg("failed to check if mech exists in queue backlog")
		return terror.Error(err, "Failed to check whether or not mech is in the battle queue backlog")
	}

	if backlogMech != nil {
		gamelog.L.Debug().Str("mech_id", backlogMech.MechID).Err(err).Msg("mech already in queue backlog")
		return terror.Error(fmt.Errorf("mech already in queue backlog"), "Your mech is already in queue")
	}

	// Check if any of the mechs exist in the battle queue
	existMech, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.IN(req.Payload.MechIDs)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech_ids", req.Payload.MechIDs).Err(err).Msg("failed to check if mech exists in queue")
		return terror.Error(err, "Failed to check whether mech is in the battle queue")
	}

	if existMech != nil {
		gamelog.L.Debug().Str("mech_id", existMech.MechID).Err(err).Msg("mech already in queue")
		position, err := db.MechQueuePosition(existMech.MechID, factionID)
		if err != nil {
			return terror.Error(err, "Already in queue, failed to get position. Contact support or try again.")
		}

		if position.QueuePosition == 0 {
			return terror.Error(fmt.Errorf("war machine already in battle"))
		}

		return terror.Error(fmt.Errorf("your mech is already in queue, current position is %d", position.QueuePosition))
	}

	// check mech is still in repair
	rcs, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(req.Payload.MechIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
		qm.Load(boiler.RepairCaseRels.RepairOffers),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech ids", req.Payload.MechIDs).Msg("Failed to get repair case")
		return terror.Error(err, "Failed to queue mech.")
	}

	if rcs != nil && len(rcs) > 0 {
		canDeployRatio := db.GetDecimalWithDefault(db.KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))

		for _, rc := range rcs {
			totalBlocks := db.TotalRepairBlocks(rc.MechID)

			// broadcast current mech stat if repair is above can deploy ratio
			if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).GreaterThan(canDeployRatio) {
				// if mech has more than half of the block to repair
				return terror.Error(fmt.Errorf("mech is not fully recovered"), "One of your mechs is still under repair.")
			}
		}
	}

	var tx *sql.Tx
	paidTxID := ""

	deployedMechs := []*boiler.CollectionItem{}
	for _, mci := range mcis {
		err = func() error {
			tx, err = gamedb.StdConn.Begin()
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

			// Insert into battle queue backlog
			bqb := &boiler.BattleQueueBacklog{
				MechID:    mci.ItemID,
				QueuedAt:  time.Now(),
				FactionID: factionID,
				OwnerID:   mci.OwnerID,
				FeeID:     null.StringFrom(bqf.ID),
			}

			err = bqb.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").
					Interface("mech id", mci.ItemID).
					Err(err).Msg("unable to insert mech into battle queue backlog")
				return terror.Error(err, "Unable to join queue, contact support or try again.")
			}

			// pay battle queue fee
			paidTxID, err = am.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
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
				// refund queue fee
				_, err = am.RPCClient.RefundSupsMessage(paidTxID)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("txID", paidTxID).Err(err).Msg("failed to refund queue fee")
				}
			}

			bqf.PaidTXID = null.StringFrom(paidTxID)
			_, err = bqf.Update(tx, boil.Whitelist(boiler.BattleQueueFeeColumns.PaidTXID))
			if err != nil {
				refundFunc()
				gamelog.L.Error().Interface("battle queue fee", bqf).Err(err).Msg("Failed to update battle queue fee transaction id")
				return terror.Error(err, "Failed to update queue fee transaction id")
			}

			// stop repair offers, if there is any
			if index := slices.IndexFunc(rcs, func(rc *boiler.RepairCase) bool { return rc.MechID == mci.ItemID }); index != -1 {
				rc := rcs[index]
				// cancel all the existing offer
				if rc.R != nil && rc.R.RepairOffers != nil {
					ids := []string{}
					for _, ro := range rc.R.RepairOffers {
						ids = append(ids, ro.ID)
					}

					err = am.SendRepairFunc(func() error {
						err = am.CloseRepairOffers(ids, boiler.RepairFinishReasonSTOPPED, boiler.RepairAgentFinishReasonEXPIRED)
						if err != nil {
							return err
						}

						return nil
					})
					if err != nil {
						refundFunc()
						return err
					}
				}
			}

			// Commit transaction
			err = tx.Commit()
			if err != nil {
				refundFunc()
				gamelog.L.Error().Str("log_name", "battle arena").
					Interface("mech id", mci.ItemID).
					Err(err).Msg("unable to commit mech insertion into queue")

				return terror.Error(err, "Unable to join queue, contact support or try again.")
			}
			deployedMechs = append(deployedMechs, mci)

			am.BattleQueueManager.Deploy <- &battle_queue.Deploy{
				FactionID: factionID,
				StartBattles: func() {
					for _, a := range am.AvailableArenas() {
						a.startBattle()
					}
				},
			}

			// broadcast queue detail
			go func(mechID string) {
				collectionItem, err := boiler.CollectionItems(
					boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
					boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
					boiler.CollectionItemWhere.ItemID.EQ(mechID),
				).One(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech collection item")
					return
				}

				qs, err := db.GetCollectionItemStatus(*collectionItem)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech arena status")
					return
				}

				ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, mechID), WSPlayerAssetMechQueueSubscribe, qs)
			}(mci.ItemID)

			return nil
		}()
		if err != nil {
			// error out if no mech is deployed
			if len(deployedMechs) == 0 {
				return terror.Error(err, "Failed to deploy mech.")
			}

			// otherwise, break the loop and broadcast the partially deployed mechs
			break
		}
	}

	// Send updated battle queue status to all subscribers
	go CalcNextQueueStatus(factionID)

	if len(deployedMechs) != len(req.Payload.MechIDs) {
		return terror.Error(fmt.Errorf("not all the mechs are deployed"), "Mechs are partially deployed.")
	}

	reply(QueueJoinHandlerResponse{
		Success: true,
	})

	return nil
}

const WSMechArenaStatusUpdate = "PLAYER:ASSET:MECH:STATUS:UPDATE"

type AssetUpdateRequest struct {
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

func (am *ArenaManager) AssetUpdateRequest(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	msg := &AssetUpdateRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.EQ(msg.Payload.MechID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find mech from db")
	}

	mechStatus, err := db.GetCollectionItemStatus(*collectionItem)
	if err != nil {
		return terror.Error(err, "Failed to get mech status")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, msg.Payload.MechID), WSPlayerAssetMechQueueSubscribe, mechStatus)
	return nil
}

type QueueStatusResponse struct {
	MinimumWaitTimeSeconds   int64           `json:"minimum_wait_time_seconds"`
	AverageGameLengthSeconds int64           `json:"average_game_length_seconds"`
	QueueCost                decimal.Decimal `json:"queue_cost"`
}

const WSQueueStatusSubscribe = "BATTLE:QUEUE:STATUS:SUBSCRIBE"
const WSPlayerAssetMechQueueUpdateSubscribe = "PLAYER:ASSET:MECH:QUEUE:UPDATE"
const WSPlayerAssetMechQueueSubscribe = "PLAYER:ASSET:MECH:QUEUE:SUBSCRIBE"
