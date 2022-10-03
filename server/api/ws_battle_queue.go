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
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

func BattleQueueController(api *API) {
	api.SecureUserFactionCommand(HubKeyBattleLobbyCreate, api.BattleLobbyCreate)
	api.SecureUserFactionCommand(HubKeyBattleLobbyJoin, api.BattleLobbyJoin)
	api.SecureUserFactionCommand(HubKeyBattleLobbyLeave, api.BattleLobbyLeave)

	api.SecureUserFactionCommand(HubKeyBattleLobbySupporterJoin, api.BattleLobbySupporterJoin)
	//api.SecureUserFactionCommand(HubKeyBattleLobbySupporterLeave, api.BattleLobbySupporterLeave)
}

type BattleLobbyCreateRequest struct {
	Payload struct {
		MechIDs           []string        `json:"mech_ids"`
		EntryFee          decimal.Decimal `json:"entry_fee"`
		FirstFactionCut   decimal.Decimal `json:"first_faction_cut"`
		SecondFactionCut  decimal.Decimal `json:"second_faction_cut"`
		ThirdFactionCut   decimal.Decimal `json:"third_faction_cut"`
		Password          null.String     `json:"password,omitempty"`
		GameMapID         string          `json:"game_map_id"`
		WillNotStartUntil null.Time       `json:"will_not_start_until,omitempty"`
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

	err = db.CheckMechOwnership(user.ID, req.Payload.MechIDs)
	if err != nil {
		return err
	}

	availableMechIDs, err := db.FilterCanDeployMechIDs(req.Payload.MechIDs)
	if err != nil {
		return err
	}

	if len(availableMechIDs) == 0 {
		return terror.Error(err, "The provided mechs are still under repair.")
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
		availableMechIDs, err = db.FilterOutMechAlreadyInQueue(availableMechIDs)
		if err != nil {
			return err
		}

		if len(availableMechIDs) == 0 {
			return terror.Error(fmt.Errorf("no mech to queue"), "No mech is available to queue.")
		}

		var deployedMechIDs []string
		for _, mechID := range availableMechIDs {
			if len(deployedMechIDs) == db.FACTION_MECH_LIMIT {
				break
			}
			deployedMechIDs = append(deployedMechIDs, mechID)
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
			FirstFactionCut:       req.Payload.FirstFactionCut.Div(decimal.NewFromInt(100)),
			SecondFactionCut:      req.Payload.SecondFactionCut.Div(decimal.NewFromInt(100)),
			ThirdFactionCut:       req.Payload.ThirdFactionCut.Div(decimal.NewFromInt(100)),
			EachFactionMechAmount: db.FACTION_MECH_LIMIT,
			Password:              req.Payload.Password,
			WillNotStartUntil:     req.Payload.WillNotStartUntil,
		}

		if req.Payload.Password.Valid {
			// TODO: pay sups for creating private room?

		}

		err = bl.Insert(tx, boil.Infer())
		if err != nil {
			refund(refundFuncList)
			gamelog.L.Error().Err(err).Interface("battle lobby", bl).Msg("Failed to insert battle lobby")
			return terror.Error(err, "Failed to create battle lobby")
		}

		// check user balance
		userBalance := api.Passport.UserBalanceGet(uuid.FromStringOrNil(user.ID))
		if userBalance.LessThan(bl.EntryFee.Mul(decimal.NewFromInt(int64(len(deployedMechIDs))))) {
			refund(refundFuncList)
			return terror.Error(fmt.Errorf("not enough fund"), "Not enough fund to queue the mechs")
		}

		// insert battle mechs
		for _, mechID := range deployedMechIDs {
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

				// update mech queue status
				ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, mechID), server.HubKeyPlayerAssetMechQueueSubscribe, &server.MechArenaInfo{
					Status:            server.MechArenaStatusQueue,
					CanDeploy:         false,
					BattleLobbyNumber: null.IntFrom(bl.Number),
				})
			}

			err = blm.Insert(tx, boil.Infer())
			if err != nil {
				refund(refundFuncList)
				gamelog.L.Error().Err(err).Interface("battle lobby mech", blm).Msg("Failed to insert battle lobbies mech")
				return terror.Error(err, "Failed to insert mechs into battle lobby.")
			}
		}

		err = tx.Commit()
		if err != nil {
			refund(refundFuncList)
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to create battle lobby.")
		}

		// pause mechs repair case
		err = api.ArenaManager.PauseRepairCases(deployedMechIDs)
		if err != nil {
			return err
		}

		// broadcast lobby
		go battle.BroadcastBattleLobbyUpdate(bl.ID)

		go battle.BroadcastMechQueueStatus(user.ID, deployedMechIDs...)

		go battle.BroadcastPlayerQueueStatus(user.ID)

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type BattleLobbyJoinRequest struct {
	Payload struct {
		BattleLobbyID string   `json:"battle_lobby_id"`
		MechIDs       []string `json:"mech_ids"`
		Password      string   `json:"password"`
	} `json:"payload"`
}

const HubKeyBattleLobbyJoin = "BATTLE:LOBBY:JOIN"

func (api *API) BattleLobbyJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	err = db.CheckMechOwnership(user.ID, req.Payload.MechIDs)
	if err != nil {
		return err
	}

	availableMechIDs, err := db.FilterCanDeployMechIDs(req.Payload.MechIDs)
	if err != nil {
		return err
	}

	if len(availableMechIDs) == 0 {
		return terror.Error(err, "The provided mechs are still under repair.")
	}

	bl, err := boiler.FindBattleLobby(gamedb.StdConn, req.Payload.BattleLobbyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("battle lobby id", req.Payload.BattleLobbyID).Msg("Failed to query battle lobby")
		return terror.Error(err, "Failed to load battle lobby")
	}

	if bl == nil {
		return terror.Error(fmt.Errorf("battle lobby not exist"), "Battle lobby does not exist.")
	}

	if bl.EndedAt.Valid {
		return terror.Error(fmt.Errorf("lobby is ended"), "The battle lobby is already closed.")
	}

	if bl.ReadyAt.Valid {
		return terror.Error(fmt.Errorf("lobby is full"), "The battle lobby is already full.")
	}

	// check password, if needed
	if bl.Password.Valid && req.Payload.Password != bl.Password.String {
		return terror.Error(fmt.Errorf("incorrect password"), "The password is incorrect.")
	}

	affectedLobbyIDs := []string{bl.ID}

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		now := time.Now()
		// filter out mechs which are already in queue
		availableMechIDs, err = db.FilterOutMechAlreadyInQueue(availableMechIDs)
		if err != nil {
			return err
		}

		if len(availableMechIDs) == 0 {
			return terror.Error(fmt.Errorf("mechs ara already in the queue"), "All the mechs are already in queue.")
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

		if bl.ReadyAt.Valid {
			return terror.Error(fmt.Errorf("battle lobby is already full"), "The battle lobby is already full.")
		}

		var battleLobbyMechs []*boiler.BattleLobbiesMech
		deployedMechIDs := []string{}
		availableSlotCount := bl.EachFactionMechAmount
		// check available slot
		if bl.R != nil {
			for _, blm := range bl.R.BattleLobbiesMechs {
				if blm.FactionID == factionID {
					availableSlotCount -= 1
				}

				// record the mechs in the battle lobby
				battleLobbyMechs = append(battleLobbyMechs, blm)
			}

			// return error, if not enough slots
			if availableSlotCount == 0 {
				return terror.Error(fmt.Errorf("battle lobby is already full"), "The battle lobby is already full.")
			}
		}

		// filled mech in remain slots
		for _, mechID := range availableMechIDs {
			// break, if no slot left
			if len(deployedMechIDs) == availableSlotCount {
				break
			}

			deployedMechIDs = append(deployedMechIDs, mechID)
		}

		// mark battle lobby is ready, if it is full
		lobbyReady := bl.EachFactionMechAmount*3 == (len(bl.R.BattleLobbiesMechs) + len(deployedMechIDs))

		// queue mech
		var tx *sql.Tx
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to queue your mech.")
		}

		defer tx.Rollback()

		// create empty function placeholder
		refundFns := []func(){}
		refund := func(fns []func()) {
			for _, fn := range fns {
				fn()
			}
		}

		for _, mechID := range deployedMechIDs {
			blm := &boiler.BattleLobbiesMech{
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

				refundFns = append(refundFns, func() {
					_, err = api.Passport.RefundSupsMessage(entryTxID)
					if err != nil {
						gamelog.L.Error().Err(err).Str("entry tx id", entryTxID).Msg("Failed to refund mech queue fee.")
					}
				})
			}

			err = blm.Insert(tx, boil.Infer())
			if err != nil {
				refund(refundFns)
				gamelog.L.Error().Err(err).Interface("battle lobby mech", blm).Msg("Failed to insert battle lobbies mech")
				return terror.Error(err, "Failed to insert mechs into battle lobby.")
			}

			// record the mechs in the battle lobby
			battleLobbyMechs = append(battleLobbyMechs, blm)
		}

		// mark battle lobby to ready
		if lobbyReady {
			bl.ReadyAt = null.TimeFrom(now)
			_, err = bl.Update(tx, boil.Whitelist(boiler.BattleLobbyColumns.ReadyAt))
			if err != nil {
				refund(refundFns)
				gamelog.L.Error().Interface("battle lobby", bl).Err(err).Msg("Failed to update battle lobby.")
				return terror.Error(err, "Failed to mark battle lobby to ready.")
			}

			_, err = bl.BattleLobbiesMechs(
				boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			).UpdateAll(tx, boiler.M{boiler.BattleLobbiesMechColumns.LockedAt: null.TimeFrom(now)})
			if err != nil {
				refund(refundFns)
				gamelog.L.Error().Interface("battle lobby", bl).Err(err).Msg("Failed to lock battle lobby mechs.")
				return terror.Error(err, "Failed to lock mechs in the battle lobby.")
			}

			// generate another system lobby
			if bl.GeneratedBySystem {
				newBattleLobby := &boiler.BattleLobby{
					HostByID:              bl.HostByID,
					EntryFee:              bl.EntryFee, // free to join
					FirstFactionCut:       bl.FirstFactionCut,
					SecondFactionCut:      bl.SecondFactionCut,
					ThirdFactionCut:       bl.ThirdFactionCut,
					EachFactionMechAmount: bl.EachFactionMechAmount,
					GameMapID:             bl.GameMapID,
					GeneratedBySystem:     true,
				}

				err = newBattleLobby.Insert(tx, boil.Infer())
				if err != nil {
					refund(refundFns)
					gamelog.L.Error().Err(err).Msg("Failed to insert public battle lobbies.")
					return terror.Error(err, "Failed to insert new system battle lobby.")
				}

				affectedLobbyIDs = append(affectedLobbyIDs, newBattleLobby.ID)
			}

		}

		err = tx.Commit()
		if err != nil {
			refund(refundFns)
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to queue your mech.")
		}

		// pause mechs repair case
		err = api.ArenaManager.PauseRepairCases(deployedMechIDs)
		if err != nil {
			return err
		}

		// kick
		if lobbyReady {
			api.ArenaManager.KickIdleArenas()
		}

		// broadcast mech queue position
		go func(battleLobby *boiler.BattleLobby, currentDeployedMechIDs []string, allLobbyMechs []*boiler.BattleLobbiesMech) {
			battle.BroadcastMechQueueStatus(user.ID, deployedMechIDs...)

			mai := &server.MechArenaInfo{
				Status:            server.MechArenaStatusQueue,
				CanDeploy:         false,
				BattleLobbyNumber: null.IntFrom(battleLobby.Number),
			}

			// only broadcast queue status change for deployed mechs, if battle lobby is not ready
			if !battleLobby.ReadyAt.Valid {
				for _, mechID := range currentDeployedMechIDs {
					// update mech queue status
					ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, mechID), server.HubKeyPlayerAssetMechQueueSubscribe, mai)
				}

				return
			}

			// otherwise, get battle lobby queue position
			battleLobbies, err := boiler.BattleLobbies(
				boiler.BattleLobbyWhere.ReadyAt.LTE(bl.ReadyAt),
				boiler.BattleLobbyWhere.AssignedToBattleID.IsNull(),
				boiler.BattleLobbyWhere.EndedAt.IsNull(),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load battle lobby queue position")
				return
			}

			// in case, race condition
			if battleLobbies == nil {
				return
			}

			// set battle lobby queue position
			mai.BattleLobbyQueuePosition = null.IntFrom(len(battleLobbies))
			// broadcast status change for all the lobby mechs
			for _, blm := range allLobbyMechs {
				// update mech queue status
				ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", blm.FactionID, blm.MechID), server.HubKeyPlayerAssetMechQueueSubscribe, mai)
			}

		}(bl, deployedMechIDs, battleLobbyMechs)

		// broadcast battle lobby
		go battle.BroadcastBattleLobbyUpdate(affectedLobbyIDs...)

		// broadcast player queue status
		go battle.BroadcastPlayerQueueStatus(user.ID)

		// terminate repair bay
		// wrap it in go routine, the channel will not slow down the deployment process
		go func(playerID string, mechIDs []string) {
			// clean up repair slots, if any mechs are successfully deployed and in the bay
			nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)
			now = time.Now()
			_ = api.ArenaManager.SendRepairFunc(func() error {
				tx, err = gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
					return terror.Error(err, "Failed to start db transaction")
				}

				defer tx.Rollback()

				var count int64
				count, err = boiler.PlayerMechRepairSlots(
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
					var pms boiler.PlayerMechRepairSlotSlice
					pms, err = boiler.PlayerMechRepairSlots(
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

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type BattleLobbyLeaveRequest struct {
	Payload struct {
		MechIDs []string `json:"mech_ids"`
	} `json:"payload"`
}

const HubKeyBattleLobbyLeave = "BATTLE:LOBBY:LEAVE"

func (api *API) BattleLobbyLeave(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyLeaveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	err = db.CheckMechOwnership(user.ID, req.Payload.MechIDs)
	if err != nil {
		return err
	}

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		now := time.Now()

		var blms boiler.BattleLobbiesMechSlice
		blms, err = boiler.BattleLobbiesMechs(
			boiler.BattleLobbiesMechWhere.MechID.IN(req.Payload.MechIDs),
			boiler.BattleLobbiesMechWhere.LockedAt.IsNull(),
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			qm.Load(boiler.BattleLobbiesMechRels.BattleLobby),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Strs("mech id list", req.Payload.MechIDs).Msg("Failed to load battle lobbies mech.")
			return terror.Error(err, "Failed to load battle lobby queuing records.")
		}

		if blms == nil {
			return terror.Error(fmt.Errorf("no mech in queue"), "No mech is in queue.")
		}

		var tx *sql.Tx
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to leave battle lobby.")
		}

		defer tx.Rollback()

		var refundFns []func()
		refund := func(fns []func()) {
			for _, fn := range fns {
				fn()
			}
		}

		changedBattleLobbyIDs := []string{}
		leftMechIDs := []string{}
		for _, blm := range blms {
			if blm.R == nil || blm.R.BattleLobby == nil {
				continue
			}

			bl := blm.R.BattleLobby

			blm.DeletedAt = null.TimeFrom(now)

			// refund entry fee
			if bl.EntryFee.GreaterThan(decimal.Zero) {
				refundTxID := ""
				refundTxID, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
					ToUserID:             uuid.Must(uuid.FromString(user.ID)),
					Amount:               bl.EntryFee.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("refund_for_leaving_battle_lobby|%s|%s|%d", blm.MechID, bl.ID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupSupremacy),
					SubGroup:             string(server.TransactionGroupBattle),
					Description:          fmt.Sprintf("refund entry fee of joining battle lobby #%d.", bl.Number),
				})
				if err != nil {
					refund(refundFns)
					gamelog.L.Error().Err(err).
						Str("from user id", server.SupremacyBattleUserID).
						Str("to user id", user.ID).
						Str("amount", bl.EntryFee.StringFixed(0)).
						Msg("Failed to refund battle lobby entry fee.")
					return terror.Error(err, "Failed to refund battle lobby entry fee.")
				}

				// append refund func
				refundFns = append(refundFns, func() {
					_, err = api.Passport.RefundSupsMessage(refundTxID)
					if err != nil {
						gamelog.L.Error().Err(err).Str("transaction id", refundTxID).Msg("Failed to refund transaction.")
						return
					}
				})

				blm.RefundTXID = null.StringFrom(refundTxID)
			}

			_, err = blm.Update(tx, boil.Whitelist(boiler.BattleLobbiesMechColumns.RefundTXID, boiler.BattleLobbiesMechColumns.DeletedAt))
			if err != nil {
				refund(refundFns)
				return terror.Error(err, "Failed to archive ")
			}

			leftMechIDs = append(leftMechIDs, blm.MechID)
			if slices.Index(changedBattleLobbyIDs, blm.BattleLobbyID) == -1 {
				changedBattleLobbyIDs = append(changedBattleLobbyIDs, blm.BattleLobbyID)
			}
		}

		err = tx.Commit()
		if err != nil {
			refund(refundFns)
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to leave battle lobby.")
		}

		// load changed battle lobbies
		var bls boiler.BattleLobbySlice
		bls, err = boiler.BattleLobbies(
			boiler.BattleLobbyWhere.ID.IN(changedBattleLobbyIDs),
			qm.Load(boiler.BattleLobbyRels.BattleLobbiesMechs, boiler.BattleLobbiesMechWhere.DeletedAt.IsNull()),
		).All(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to load changed battle lobbies")
		}

		lobbyIDs := []string{}
		for _, bl := range bls {
			lobbyIDs = append(lobbyIDs, bl.ID)

			// broadcast new battle lobby status, if the lobby is generated by system or still have mech queued in it
			if bl.GeneratedBySystem || (bl.R != nil && bl.R.BattleLobbiesMechs != nil && len(bl.R.BattleLobbiesMechs) > 0) {
				continue
			}

			// otherwise, soft delete battle lobby
			bl.DeletedAt = null.TimeFrom(now)
			_, err = bl.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleLobbyColumns.DeletedAt))
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to soft delete battle lobby")
				continue
			}
		}

		go battle.BroadcastBattleLobbyUpdate(lobbyIDs...)

		// restart repair case
		go func() {
			err = api.ArenaManager.RestartRepairCases(leftMechIDs)
			if err != nil {
				gamelog.L.Error().Err(err).Strs("mech id list", leftMechIDs).Msg("Failed to restart repair cases")
			}
		}()

		// broadcast player queue status
		go battle.BroadcastPlayerQueueStatus(user.ID)

		// broadcast new mech stat
		go func() {
			battle.BroadcastMechQueueStatus(user.ID, leftMechIDs...)

			cis, err := boiler.CollectionItems(
				boiler.CollectionItemWhere.ItemID.IN(leftMechIDs),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Strs("collection item id list", leftMechIDs).Msg("Failed to load collection items")
				return
			}

			for _, ci := range cis {
				queueDetails, err := db.GetCollectionItemStatus(*ci)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech arena status")
					continue
				}

				ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, ci.ItemID), server.HubKeyPlayerAssetMechQueueSubscribe, queueDetails)
			}
		}()

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type BattleBountyCreateRequest struct {
	Payload struct {
		MechID string          `json:"mech_id"`
		Amount decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

// subscriptions

func (api *API) BattleETASubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	bs, err := boiler.Battles(
		boiler.BattleWhere.EndedAt.IsNotNull(),
		qm.OrderBy(boiler.BattleColumns.BattleNumber+" DESC"),
		qm.Limit(100),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load latest 100 battles")
		return terror.Error(err, "Failed to load ended battles")
	}

	if bs == nil || len(bs) == 0 {
		// NOTE: this will only happen on DEV or STAGING (after reset), so hardcode default ETA to 5 minutes
		reply(300)
		return nil
	}

	var totalDuration time.Duration
	for _, b := range bs {
		totalDuration += b.EndedAt.Time.Sub(b.StartedAt)
	}

	reply(int(totalDuration.Seconds()) / len(bs))

	return nil
}

func (api *API) BattleLobbyListUpdate(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// return all the unfinished lobbies
	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.EndedAt.IsNull(),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle lobbies.")
		return terror.Error(err, "Failed to load battle lobbies.")
	}

	resp, err := server.BattleLobbiesFromBoiler(bls)
	if err != nil {
		return err
	}


	reply(resp)

	return nil
}


func (api *API)  NextBattleDetails(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	battleLobbyIDs := api.ArenaManager.GetCurrentBattleLobbyIDs()
	bl, err := db.GetNextBattleLobby(battleLobbyIDs)
	if err != nil {
		return err
	}

	if bl == nil {
		return nil
	}

	resp, err := server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl})
	if err != nil {
		return err
	}

	if len(resp) != 1 {
		return fmt.Errorf("unable to retrieve upcoming lobby details")
	}

	ws.PublishMessage("/public/upcoming_battle", server.HubKeyNextBattleDetails, resp[0])
	return nil
}

type BattleLobbySupporterJoinRequest struct {
	Payload struct {
		BattleLobbyID string `json:"battle_lobby_id"`
	} `json:"payload"`
}

const HubKeyBattleLobbySupporterJoin = "BATTLE:LOBBY:SUPPORTER:JOIN"

func (api *API) BattleLobbySupporterJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	L := gamelog.L.With().Str("func", "").Str("user id", user.ID).Logger()
	req := &BattleLobbySupporterJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	L = L.With().Interface("payload", req.Payload).Logger()

	// add support to battle lobby
	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// todo, figure out rules for when they are allowed to join as a supporter
		// check lobby exists
		bl, err := boiler.BattleLobbies(
			boiler.BattleLobbyWhere.ID.EQ(req.Payload.BattleLobbyID),
			qm.Load(boiler.BattleLobbyRels.BattleLobbySupporters,
				boiler.BattleLobbySupporterWhere.FactionID.EQ(factionID),
				),
			).One(gamedb.StdConn)
		if err != nil {
			return err
		}

		if bl == nil {
			return fmt.Errorf("lobby id: %s does not exist", req.Payload.BattleLobbyID)
		}

		// add them as a supporter
		bls := &boiler.BattleLobbySupporter{
			SupporterID:   user.ID,
			BattleLobbyID: bl.ID,
			FactionID: factionID,
		}
		err = bls.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		L.Error().Err(err).Msg("failed to insert new battle lobby supporter")
		return err
	}

	reply(true)
	return nil
}
