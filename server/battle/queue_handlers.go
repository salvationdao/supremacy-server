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

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"

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

	pos, err := db.GetFactionQueueLength(factionID)
	if err != nil {
		l.Warn().Err(err).Msg("unable to retrieve queue position")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/queue", factionID), WSQueueStatusSubscribe, QueueStatusResponse{
		QueuePosition: pos + 1,
		QueueCost:     decimal.Zero,
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

	playerBan, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BanMechQueue.EQ(true),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to find banned user")
	}

	if playerBan != nil {
		return terror.Error(fmt.Errorf("user is banned from queing mechs"), "User is banned from queuing mech")
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

	queueCount, err := db.GetPlayerQueueCount(user.ID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("userID", user.ID).Err(err).Msg("failed to check player queue count")
		return terror.Error(err, "Something went wrong while attempting to queue your mech(s). Please try again or contact support if this problem persists.")
	}
	queueLimit := db.GetIntWithDefault(db.KeyPlayerQueueLimit, 10)
	if queueCount >= int64(queueLimit) {
		return terror.Error(terror.ErrForbidden, fmt.Sprintf("You cannot have more than %d mechs in queue at the same time. Please wait before queueing any more mechs.", queueLimit))
	}
	if (int64(len(mcis)) + queueCount) > int64(queueLimit) {
		return terror.Error(terror.ErrForbidden, fmt.Sprintf("You cannot have more than %d mechs in queue at the same time. You currently have %d mechs in queue. Please remove at least %d mechs from your selection and try again.", queueLimit, queueCount, len(mcis)-(queueLimit-int(queueCount))))
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
		qm.Load(boiler.RepairCaseRels.RepairOffers, boiler.RepairOfferWhere.ClosedAt.IsNull()),
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

	deployedMechIDs := []string{}
	now := time.Now()
	nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)

	// insert mech from the input order
	for _, mechID := range req.Payload.MechIDs {
		idx := slices.IndexFunc(mcis, func(mci *boiler.CollectionItem) bool {
			return mci.ItemID == mechID
		})
		if idx == -1 {
			continue
		}
		mci := mcis[idx]

		err = func() error {
			tx, err = gamedb.StdConn.Begin()
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("unable to begin tx")
				return fmt.Errorf(terror.Echo(err))
			}
			defer tx.Rollback()

			// Insert into battle queue
			bq := &boiler.BattleQueue{
				MechID:    mci.ItemID,
				QueuedAt:  time.Now(),
				FactionID: factionID,
				OwnerID:   mci.OwnerID,
			}
			err = bq.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").
					Interface("mech id", mci.ItemID).
					Err(err).Msg("unable to insert mech into battle queue")
				return terror.Error(err, "Unable to join queue, contact support or try again.")
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
						return err
					}
				}
			}

			// Commit transaction
			err = tx.Commit()
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").
					Interface("mech id", mci.ItemID).
					Err(err).Msg("unable to commit mech insertion into queue")

				return terror.Error(err, "Unable to join queue, contact support or try again.")
			}

			deployedMechIDs = append(deployedMechIDs, mci.ItemID)

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

		// broadcast queue detail
		go func() {
			qs, err := db.GetNextBattle(ctx)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech arena status")
				return
			}

			ws.PublishMessage("/public/arena/upcomming_battle", HubKeyNextBattleDetails, qs)
		}()

		if err != nil {
			// error out if no mech is deployed
			if len(deployedMechIDs) == 0 {
				return terror.Error(err, "Failed to deploy mech.")
			}

			// otherwise, break the loop and broadcast the partially deployed mechs
			break
		}
	}

	// clean up repair slots, if any mechs are successfully deployed and in the bay
	if len(deployedMechIDs) > 0 {
		// wrap it in go routine, the channel will not slow down the deployment process
		go func(playerID string, mechIDs []string) {
			_ = am.SendRepairFunc(func() error {
				tx, err = gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
					return terror.Error(err, "Failed to start db transaction")
				}

				defer tx.Rollback()

				count, err := boiler.PlayerMechRepairSlots(
					boiler.PlayerMechRepairSlotWhere.MechID.IN(mechIDs),
					boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
				).UpdateAll(
					tx,
					boiler.M{
						boiler.PlayerMechRepairSlotColumns.Status:         boiler.RepairSlotStatusDONE,
						boiler.PlayerMechRepairSlotColumns.SlotNumber:     0,
						boiler.PlayerMechRepairSlotColumns.NextRepairTime: null.TimeFromPtr(nil),
					},
				)
				if err != nil {
					gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to update repair slot.")
					return terror.Error(err, "Failed to update repair slot")
				}

				// update remain slots and broadcast
				resp := []*boiler.PlayerMechRepairSlot{}
				if count > 0 {
					pms, err := boiler.PlayerMechRepairSlots(
						boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(playerID),
						boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
						qm.OrderBy(boiler.PlayerMechRepairSlotColumns.SlotNumber),
					).All(tx)
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to load player mech repair slots.")
						return terror.Error(err, "Failed to load repair slots")
					}

					for i, pm := range pms {
						shouldUpdate := false

						// check slot number
						if pm.SlotNumber != i+1 {
							pm.SlotNumber = i + 1
							shouldUpdate = true
						}

						if pm.SlotNumber == 1 {
							if pm.Status != boiler.RepairSlotStatusREPAIRING {
								pm.Status = boiler.RepairSlotStatusREPAIRING
								shouldUpdate = true
							}

							if !pm.NextRepairTime.Valid {
								pm.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
								shouldUpdate = true
							}
						} else {
							if pm.Status != boiler.RepairSlotStatusPENDING {
								pm.Status = boiler.RepairSlotStatusPENDING
								shouldUpdate = true
							}

							if pm.NextRepairTime.Valid {
								pm.NextRepairTime = null.TimeFromPtr(nil)
								shouldUpdate = true
							}
						}

						if shouldUpdate {
							_, err = pm.Update(tx,
								boil.Whitelist(
									boiler.PlayerMechRepairSlotColumns.SlotNumber,
									boiler.PlayerMechRepairSlotColumns.Status,
									boiler.PlayerMechRepairSlotColumns.NextRepairTime,
								),
							)
							if err != nil {
								gamelog.L.Error().Err(err).Interface("repair slot", pm).Msg("Failed to update repair slot.")
								return terror.Error(err, "Failed to update repair slot")
							}
						}

						resp = append(resp, pm)
					}
				}

				err = tx.Commit()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
					return terror.Error(err, "Failed to commit db transaction.")
				}

				// broadcast new list, if changed
				if count > 0 {
					ws.PublishMessage(fmt.Sprintf("/secure/user/%s/repair_bay", playerID), server.HubKeyMechRepairSlots, resp)
				}

				return nil
			})
		}(user.ID, deployedMechIDs)
	}

	reopeningDate, err := time.Parse(time.RFC3339, "2022-09-08T08:00:00+08:00")
	if err != nil {
		gamelog.L.Error().Str("func", "Load").Msg("failed to get reopening date time")
		return terror.Error(err, "Failed to parse reopen date.")
	}
	// restart idle arenas, if it is not prod env or the time has passed reopen date
	if !server.IsProductionEnv() || time.Now().After(db.GetTimeWithDefault(db.KeyProdReopeningDate, reopeningDate)) {
		for _, arena := range am.IdleArenas() {
			// trigger begin battle when arena is idle
			go arena.BeginBattle()
		}
	}

	// Send updated battle queue status to all subscribers
	go CalcNextQueueStatus(factionID)

	if len(deployedMechIDs) != len(req.Payload.MechIDs) {
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
	QueuePosition int64           `json:"queue_position"`
	QueueCost     decimal.Decimal `json:"queue_cost"`
}

const WSQueueStatusSubscribe = "BATTLE:QUEUE:STATUS:SUBSCRIBE"
const WSPlayerAssetMechQueueUpdateSubscribe = "PLAYER:ASSET:MECH:QUEUE:UPDATE"
const WSPlayerAssetMechQueueSubscribe = "PLAYER:ASSET:MECH:QUEUE:SUBSCRIBE"
