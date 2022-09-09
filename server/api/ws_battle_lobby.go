package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

func BattleLobbyController(api *API) {
	api.SecureUserFactionCommand(HubKeyBattleLobbyCreate, api.BattleLobbyCreate)
	api.SecureUserFactionCommand(HubKeyBattleLobbyJoin, api.BattleLobbyJoin)
}

type BattleLobbyCreateRequest struct {
	Payload struct {
		MechIDs          []string        `json:"mechIDs"`
		EntryFee         decimal.Decimal `json:"entry_fee"`
		FirstFactionCut  decimal.Decimal `json:"first_faction_cut"`
		SecondFactionCut decimal.Decimal `json:"second_faction_cut"`
		ThirdFactionCut  decimal.Decimal `json:"third_faction_cut"`
		Password         string          `json:"password"`
		GameMapID        string          `json:"game_map_id"`
	} `json:"payload"`
}

const HubKeyBattleLobbyCreate = "BATTLE:LOBBY:CREATE"

func (api *API) BattleLobbyCreate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// initial mech amount check
	if len(req.Payload.MechIDs) == 0 {
		return terror.Error(fmt.Errorf("mech id list not provided"), "Initial mech is not provided.")
	}

	// check if initial mechs count is over the limit
	if len(req.Payload.MechIDs) > db.FACTION_MECH_LIMIT {
		return terror.Error(fmt.Errorf("mech more than 3"), "Maximum 3 mech per faction.")
	}

	relatedRepairCases, err := api.ArenaManager.CheckMechCanQueue(user.ID, req.Payload.MechIDs)
	if err != nil {
		return err
	}

	// entry fee check
	if req.Payload.EntryFee.IsNegative() {
		return terror.Error(fmt.Errorf("negative entry fee"), "Entry fee cannot be negative.")
	}

	// reward cut check
	if req.Payload.FirstFactionCut.IsNegative() || req.Payload.SecondFactionCut.IsNegative() || req.Payload.ThirdFactionCut.IsNegative() {
		return terror.Error(fmt.Errorf("negative reward cut"), "Reward cut must not be less than zero.")
	}

	if !req.Payload.FirstFactionCut.Add(req.Payload.SecondFactionCut).Add(req.Payload.ThirdFactionCut).Equal(decimal.NewFromInt(100)) {
		return terror.Error(fmt.Errorf("total must be 100"), "The total of the reward cut must equal 100.")
	}

	// start process
	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// check mech in queue
		err = api.ArenaManager.CheckMechAlreadyInQueue(req.Payload.MechIDs)
		if err != nil {
			return err
		}

		var tx *sql.Tx

		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to create battle lobby.")
		}

		defer tx.Rollback()

		// refund func list
		var refundFuncList []func()

		// loop through refund
		refund := func(fns []func()) {
			for _, fn := range fns {
				fn()
			}
		}

		bl := &boiler.BattleLobby{
			HostByID:              user.ID,
			EntryFee:              req.Payload.EntryFee,
			FirstFactionCut:       req.Payload.FirstFactionCut,
			SecondFactionCut:      req.Payload.SecondFactionCut,
			ThirdFactionCut:       req.Payload.ThirdFactionCut,
			EachFactionMechAmount: db.FACTION_MECH_LIMIT,
		}

		if req.Payload.Password != "" {
			bl.Password = null.StringFrom(req.Payload.Password)

			// TODO: pay sups for private room??

		}

		err = bl.Insert(tx, boil.Infer())
		if err != nil {
			refund(refundFuncList)
			gamelog.L.Error().Err(err).Interface("battle lobby", bl).Msg("Failed to insert battle lobby")
			return terror.Error(err, "Failed to create battle lobby")
		}

		// check user balance
		userBalance := api.Passport.UserBalanceGet(uuid.FromStringOrNil(user.ID))
		if userBalance.LessThan(bl.EntryFee.Mul(decimal.NewFromInt(int64(len(req.Payload.MechIDs))))) {
			refund(refundFuncList)
			return terror.Error(fmt.Errorf("not enough fund"), "Not enough fund to queue the mechs")
		}

		// insert battle mechs
		for _, mechID := range req.Payload.MechIDs {
			blm := boiler.BattleLobbiesMech{
				BattleLobbyID: bl.ID,
				MechID:        mechID,
				OwnerID:       user.ID,
				FactionID:     factionID,
			}

			if bl.EntryFee.GreaterThan(decimal.Zero) {
				var entryTxID string

				entryTxID, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.Must(uuid.FromString(user.ID)),
					ToUserID:             uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
					Amount:               bl.EntryFee.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("enter_battle_lobby_fee|%s|%s|%d", mechID, bl.ID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupSupremacy),
					SubGroup:             string(server.TransactionGroupBattle),
					Description:          "entry fee of joining battle lobby.",
				})
				if err != nil {
					refund(refundFuncList)
					gamelog.L.Error().
						Str("player_id", user.ID).
						Str("mech id", mechID).
						Str("amount", bl.EntryFee.StringFixed(0)).
						Err(err).Msg("Failed to pay sups on entering battle lobby.")
					return terror.Error(err, "Failed to pay sups on entering battle lobby.")
				}
				blm.PaidTXID = null.StringFrom(entryTxID)

				// append refund func
				refundFuncList = append(refundFuncList, func() {
					_, err = api.Passport.RefundSupsMessage(entryTxID)
					if err != nil {
						gamelog.L.Error().Err(err).Str("entry tx id", entryTxID).Msg("Failed to refund transaction id")
					}
				})
			}

			err = blm.Insert(tx, boil.Infer())
			if err != nil {
				refund(refundFuncList)
				gamelog.L.Error().Err(err).Interface("battle lobby mech", blm).Msg("Failed to insert battle lobbies mech")
				return terror.Error(err, "Failed to insert mechs into battle lobby.")
			}

			// stop repair offers, if there is any
			if index := slices.IndexFunc(relatedRepairCases, func(rc *boiler.RepairCase) bool { return rc.MechID == mechID }); index != -1 {
				rc := relatedRepairCases[index]
				// cancel all the existing offer
				if rc.R != nil && rc.R.RepairOffers != nil {
					ids := []string{}
					for _, ro := range rc.R.RepairOffers {
						ids = append(ids, ro.ID)
					}

					err = api.ArenaManager.SendRepairFunc(func() error {
						err = api.ArenaManager.CloseRepairOffers(ids, boiler.RepairFinishReasonSTOPPED, boiler.RepairAgentFinishReasonEXPIRED)
						if err != nil {
							return err
						}

						return nil
					})
					if err != nil {
						refund(refundFuncList)
						return err
					}
				}
			}

		}

		err = tx.Commit()
		if err != nil {
			refund(refundFuncList)
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to create battle lobby.")
		}

		// broadcast lobby
		go api.ArenaManager.BroadcastBattleLobbyUpdate(bl.ID)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type BattleLobbyJoinRequest struct {
	Payload struct {
		BattleLobbyID string `json:"battle_lobby_id"`
		MechID        string `json:"mech_id"`
		Password      string `json:"password"`
	} `json:"payload"`
}

const HubKeyBattleLobbyJoin = "BATTLE:LOBBY:JOIN"

func (api *API) BattleLobbyJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	relatedRepairCases, err := api.ArenaManager.CheckMechCanQueue(user.ID, []string{req.Payload.MechID})
	if err != nil {
		return err
	}

	bl, err := boiler.FindBattleLobby(gamedb.StdConn, req.Payload.BattleLobbyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("battle lobby id", req.Payload.BattleLobbyID).Msg("Failed to query battle lobby")
		return terror.Error(err, "Failed to load battle lobby")
	}

	if bl == nil {
		return terror.Error(fmt.Errorf("battle lobby not exist"), "Battle lobby does not exist.")
	}

	// check password, if needed
	if bl.Password.Valid && req.Payload.Password != bl.Password.String {
		return terror.Error(fmt.Errorf("incorrect password"), "The password is incorrect.")
	}

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// check mech already in queue
		err = api.ArenaManager.CheckMechAlreadyInQueue([]string{req.Payload.MechID})
		if err != nil {
			return err
		}

		// check whether the lobby is still available
		bl, err = boiler.BattleLobbies(
			boiler.BattleLobbyWhere.ID.EQ(req.Payload.BattleLobbyID),
			qm.Load(
				boiler.BattleLobbyRels.BattleLobbiesMechs,
				boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
				boiler.BattleLobbiesMechWhere.DeletedAt.IsNull(),
			),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Str("battle lobby id", req.Payload.BattleLobbyID).Msg("Failed to query battle lobby")
			return terror.Error(err, "Failed to load battle lobby")
		}

		if bl == nil {
			return terror.Error(fmt.Errorf("battle lobby not exist"), "Battle lobby does not exist.")
		}

		if bl.ReadyAt.Valid || bl.FinishedAt.Valid {
			return terror.Error(fmt.Errorf("battle lobby is already full"), "The battle lobby is already full.")
		}

		// check available slot
		if bl.R != nil && bl.R.BattleLobbiesMechs != nil {
			availableSlotCount := bl.EachFactionMechAmount
			for _, blm := range bl.R.BattleLobbiesMechs {
				if blm.FactionID == factionID {
					availableSlotCount -= 1
				}
			}
			if availableSlotCount == 0 {
				return terror.Error(fmt.Errorf("battle lobby is already full"), "The battle lobby is already full.")
			}
		}

		// queue mech
		var tx *sql.Tx
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to queue your mech.")
		}

		defer tx.Rollback()

		// create empty function placeholder
		refundFunc := func() {}

		blm := &boiler.BattleLobbiesMech{
			BattleLobbyID: bl.ID,
			MechID:        req.Payload.MechID,
			OwnerID:       user.ID,
			FactionID:     factionID,
		}

		if bl.EntryFee.GreaterThan(decimal.Zero) {
			var entryTxID string
			entryTxID, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(user.ID)),
				ToUserID:             uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
				Amount:               bl.EntryFee.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("enter_battle_lobby_fee|%s|%s|%d", blm.MechID, bl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          "entry fee of joining battle lobby.",
			})
			if err != nil {
				gamelog.L.Error().
					Str("player_id", user.ID).
					Str("mech id", blm.MechID).
					Str("amount", bl.EntryFee.StringFixed(0)).
					Err(err).Msg("Failed to pay sups on entering battle lobby.")
				return terror.Error(err, "Failed to pay sups on entering battle lobby.")
			}
			blm.PaidTXID = null.StringFrom(entryTxID)

			refundFunc = func() {
				_, err = api.Passport.RefundSupsMessage(entryTxID)
				if err != nil {
					gamelog.L.Error().Err(err).Str("entry tx id", entryTxID).Msg("Failed to refund mech queue fee.")
				}
			}
		}

		err = blm.Insert(tx, boil.Infer())
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Interface("battle lobby mech", blm).Msg("Failed to insert battle lobbies mech")
			return terror.Error(err, "Failed to insert mechs into battle lobby.")
		}

		// mark battle lobby to ready, if the queue is full
		if bl.R != nil && bl.R.BattleLobbiesMechs != nil && len(bl.R.BattleLobbyBounties)+1 == db.FACTION_MECH_LIMIT*3 {
			bl.ReadyAt = null.TimeFrom(time.Now())
			_, err = bl.Update(tx, boil.Whitelist(boiler.BattleLobbyColumns.ReadyAt))
			if err != nil {
				refundFunc()
				gamelog.L.Error().Err(err).Msg("Failed to update battle lobby.")
				return terror.Error(err, "Failed to mark battle lobby to ready.")
			}
		}

		// stop repair offers, if there is any
		if index := slices.IndexFunc(relatedRepairCases, func(rc *boiler.RepairCase) bool { return rc.MechID == req.Payload.MechID }); index != -1 {
			rc := relatedRepairCases[index]
			// cancel all the existing offer
			if rc.R != nil && rc.R.RepairOffers != nil {
				ids := []string{}
				for _, ro := range rc.R.RepairOffers {
					ids = append(ids, ro.ID)
				}

				err = api.ArenaManager.SendRepairFunc(func() error {
					err = api.ArenaManager.CloseRepairOffers(ids, boiler.RepairFinishReasonSTOPPED, boiler.RepairAgentFinishReasonEXPIRED)
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

		err = tx.Commit()
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to queue your mech.")
		}

		// broadcast battle lobby
		go api.ArenaManager.BroadcastBattleLobbyUpdate(bl.ID)

		// clean up repair slots, if any mechs are successfully deployed and in the bay
		go func(playerID string, mechID string) {
			nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)
			now := time.Now()
			_ = api.ArenaManager.SendRepairFunc(func() error {
				tx, err = gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
					return terror.Error(err, "Failed to start db transaction")
				}

				defer tx.Rollback()

				count, err := boiler.PlayerMechRepairSlots(
					boiler.PlayerMechRepairSlotWhere.MechID.EQ(mechID),
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
					gamelog.L.Error().Err(err).Str("mech id list", mechID).Msg("Failed to update repair slot.")
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
		}(user.ID, req.Payload.MechID)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// subscriptions

func (api *API) BattleLobbyListUpdate(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// return all the unfinished lobbies
	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.FinishedAt.IsNull(),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbiesMechs,
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			boiler.BattleLobbiesMechWhere.DeletedAt.IsNull(),
		),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle lobbies.")
		return terror.Error(err, "Failed to load battle lobbies.")
	}

	reply(server.BattleLobbiesFromBoiler(bls))

	return nil
}
