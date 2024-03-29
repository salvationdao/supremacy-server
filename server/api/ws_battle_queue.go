package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/system_messages"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

func BattleQueueController(api *API) {

	api.SecureUserFactionCommand(HubKeyPrivateBattleLobbyGet, api.PrivateBattleLobbyGet)
	api.SecureUserFactionCommand(HubKeyBattleLobbyCreate, api.BattleLobbyCreate)
	api.SecureUserFactionCommand(HubKeyBattleLobbyClose, api.BattleLobbyClose)

	api.SecureUserFactionCommand(HubKeyBattleLobbyJoin, api.BattleLobbyJoin)
	api.SecureUserFactionCommand(HubKeyBattleLobbyLeave, api.BattleLobbyLeave)
	api.SecureUserCommand(HubKeyBattleLobbyTopUpReward, api.BattleLobbyTopUpReward)

	api.SecureUserFactionCommand(HubKeyMechsStake, api.MechStake)
	api.SecureUserFactionCommand(HubKeyMechsUnstake, api.MechUnstake)

	api.SecureUserFactionCommand(HubKeyBattleLobbySupporterJoin, api.BattleLobbySupporterJoin)
	//api.SecureUserFactionCommand(HubKeyBattleLobbySupporterLeave, api.BattleLobbySupporterLeave)
}

type PrivateBattleLobbyGetRequest struct {
	Payload struct {
		AccessCode string `json:"access_code"`
	} `json:"payload"`
}

const HubKeyPrivateBattleLobbyGet = "PRIVATE:BATTLE:LOBBY:GET"

func (api *API) PrivateBattleLobbyGet(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PrivateBattleLobbyGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	bl, err := db.GetBattleLobbyViaAccessCode(req.Payload.AccessCode)
	if err != nil {
		return terror.Error(err, "Failed to load battle lobby")
	}

	filteredBattleLobbies, err := server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl})
	if err != nil {
		return err
	}

	if len(filteredBattleLobbies) > 0 {
		lobby := filteredBattleLobbies[0]
		reply(server.BattleLobbyInfoFilter(lobby, factionID, lobby.HostByID == user.ID))
	}

	return nil
}

type LobbyAccessibility string

const (
	LobbyAccessibilityPublic  LobbyAccessibility = "PUBLIC"
	LobbyAccessibilityPrivate LobbyAccessibility = "PRIVATE"
)

type BattleLobbyCreateRequest struct {
	Payload struct {
		Name              string             `json:"name"`
		Accessibility     LobbyAccessibility `json:"accessibility"`
		AccessCode        null.String        `json:"access_code"`
		EntryFee          decimal.Decimal    `json:"entry_fee"`
		FirstFactionCut   decimal.Decimal    `json:"first_faction_cut"`
		SecondFactionCut  decimal.Decimal    `json:"second_faction_cut"`
		GameMapID         null.String        `json:"game_map_id,omitempty"`
		SchedulingType    string             `json:"scheduling_type"`
		WillNotStartUntil null.Time          `json:"will_not_start_until,omitempty"`
		MaxDeployNumber   int                `json:"max_deploy_number"`
		ExtraReward       decimal.Decimal    `json:"extra_reward"`

		MechIDs        []string `json:"mech_ids"`
		InvitedUserIDs []string `json:"invited_user_ids"`
	} `json:"payload"`
}

const HubKeyBattleLobbyCreate = "BATTLE:LOBBY:CREATE"

func (api *API) BattleLobbyCreate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// check if initial mechs count is over the limit
	if len(req.Payload.MechIDs) > req.Payload.MaxDeployNumber {
		return terror.Error(fmt.Errorf("mech more than 3"), "Amount of deployed war machine has exceeded the limit.")
	}

	availableMechIDs, err := MechAuthorisationFilter(user, factionID, req.Payload.MechIDs)
	if err != nil {
		return err
	}

	availableMechIDs, err = db.OverDamagedMechFilter(availableMechIDs)
	if err != nil {
		return err
	}

	// entry fee check
	if req.Payload.EntryFee.IsNegative() {
		return terror.Error(fmt.Errorf("negative entry fee"), "Entry fee cannot be negative.")
	}

	if req.Payload.ExtraReward.IsNegative() {
		return terror.Error(fmt.Errorf("negative extra reward"), "Extra reward cannot be negative.")
	}

	// reward cut check
	if req.Payload.FirstFactionCut.IsNegative() || req.Payload.SecondFactionCut.IsNegative() {
		return terror.Error(fmt.Errorf("negative reward cut"), "Reward cut must not be less than zero.")
	}

	if !req.Payload.FirstFactionCut.Add(req.Payload.SecondFactionCut).LessThanOrEqual(decimal.NewFromInt(100)) {
		return terror.Error(fmt.Errorf("total must be 100"), "The total of the reward cut must be equal to 100.")
	}

	publicExhibitionLobbyExpireAfterSecond := db.GetIntWithDefault(db.KeyPublicExhibitionLobbyExpireAfterDurationSecond, 1800)
	lobbyHostingLimit := db.GetIntWithDefault(db.KeyLobbyHostingMaximumAmount, 5)

	// calculate total cost
	totalCost := req.Payload.ExtraReward
	if req.Payload.EntryFee.IsPositive() {
		for _ = range availableMechIDs {
			totalCost = totalCost.Add(req.Payload.EntryFee)
		}
	}

	var bl *boiler.BattleLobby
	// start process
	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// check user balance
		if totalCost.IsPositive() {
			userBalance := api.Passport.UserBalanceGet(uuid.FromStringOrNil(user.ID))
			if userBalance.LessThan(totalCost) {
				return terror.Error(fmt.Errorf("insufficent user balance"), "Insufficient user balance.")
			}
		}

		// check hosted lobbies limitation, if it is a public lobby
		if req.Payload.Accessibility == LobbyAccessibilityPublic {
			hostBattleLobbies, err := boiler.BattleLobbies(
				boiler.BattleLobbyWhere.HostByID.EQ(user.ID),
				boiler.BattleLobbyWhere.AccessCode.IsNull(),
				boiler.BattleLobbyWhere.EndedAt.IsNull(),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load hosted battle lobbies.")
				return terror.Error(err, "Failed to check hosted lobby amount")
			}

			if hostBattleLobbies != nil && len(hostBattleLobbies) >= lobbyHostingLimit {
				return terror.Error(fmt.Errorf("exceed lobby host limit"), "You have exceed the lobby hosting limit.")
			}
		}

		// filter out mechs which are already in queue
		availableMechIDs, err = db.NonQueuedMechFilter(availableMechIDs)
		if err != nil {
			return err
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

		bl = &boiler.BattleLobby{
			HostByID:              user.ID,
			Name:                  req.Payload.Name,
			GameMapID:             req.Payload.GameMapID,
			EntryFee:              req.Payload.EntryFee.Mul(decimal.New(1, 18)),
			FirstFactionCut:       req.Payload.FirstFactionCut.Div(decimal.NewFromInt(100)),
			SecondFactionCut:      req.Payload.SecondFactionCut.Div(decimal.NewFromInt(100)),
			ThirdFactionCut:       decimal.NewFromInt(100).Sub(req.Payload.FirstFactionCut).Sub(req.Payload.SecondFactionCut).Div(decimal.NewFromInt(100)),
			EachFactionMechAmount: db.FACTION_MECH_LIMIT,
			MaxDeployPerPlayer:    req.Payload.MaxDeployNumber,
			WillNotStartUntil:     req.Payload.WillNotStartUntil,
			ExpiresAt:             null.TimeFrom(time.Now().Add(time.Duration(publicExhibitionLobbyExpireAfterSecond) * time.Second)),
		}

		if bl.WillNotStartUntil.Valid {
			bl.ExpiresAt = bl.WillNotStartUntil
		}

		if req.Payload.Accessibility == LobbyAccessibilityPrivate && req.Payload.AccessCode.Valid && req.Payload.AccessCode.String != "" {
			bl.AccessCode = req.Payload.AccessCode
			bl.ExpiresAt = null.TimeFromPtr(nil)
		}

		err = bl.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("battle lobby", bl).Msg("Failed to insert battle lobby")
			return terror.Error(err, "Failed to create battle lobby")
		}

		// refund func list
		var refundFuncList []func()

		// loop through refund
		refund := func(fns []func()) {
			for _, fn := range fns {
				fn()
			}
		}

		paidTxID := ""

		if req.Payload.ExtraReward.GreaterThan(decimal.Zero) {

			amount := req.Payload.ExtraReward.Mul(decimal.New(1, 18))

			paidTxID, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(user.ID)),
				ToUserID:             uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
				Amount:               amount.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("battle_lobby_extra_reward|%s|%d", bl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          "adding extra sups reward for battle lobby.",
			})
			if err != nil {
				return terror.Error(err, "Failed to pay sups on entering battle lobby.")
			}

			// append refund func
			refundFuncList = append(refundFuncList, func() {
				_, err = api.Passport.RefundSupsMessage(paidTxID)
				if err != nil {
					gamelog.L.Error().Err(err).Str("entry tx id", paidTxID).Msg("Failed to refund transaction id")
				}
			})

			esr := &boiler.BattleLobbyExtraSupsReward{
				BattleLobbyID: bl.ID,
				OfferedByID:   user.ID,
				Amount:        amount,
				PaidTXID:      paidTxID,
			}
			err = esr.Insert(tx, boil.Infer())
			if err != nil {
				refund(refundFuncList)
				gamelog.L.Error().Err(err).Msg("Failed to insert extra sups reward.")
				return terror.Error(err, "Failed to pay extra sups reward.")
			}

		}

		if len(deployedMechIDs) > 0 {
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
					QueuedByID:    user.ID,
					FactionID:     factionID,
				}

				if bl.EntryFee.GreaterThan(decimal.Zero) {
					paidTxID, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
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
					blm.PaidTXID = null.StringFrom(paidTxID)

					// append refund func
					refundFuncList = append(refundFuncList, func() {
						_, err = api.Passport.RefundSupsMessage(paidTxID)
						if err != nil {
							gamelog.L.Error().Err(err).Str("entry tx id", paidTxID).Msg("Failed to refund transaction id")
						}
					})
				}

				err = blm.Insert(tx, boil.Infer())
				if err != nil {
					refund(refundFuncList)
					gamelog.L.Error().Err(err).Interface("battle lobby mech", blm).Msg("Failed to insert battle lobbies mech")
					return terror.Error(err, "Failed to insert mechs into battle lobby.")
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			refund(refundFuncList)
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to create battle lobby.")
		}

		if len(deployedMechIDs) > 0 {
			// pause mechs repair case
			err = api.ArenaManager.PauseRepairCases(deployedMechIDs)
			if err != nil {
				return err
			}

			// broadcast update
			api.ArenaManager.MechDebounceBroadcastChan <- deployedMechIDs
			go battle.BroadcastPlayerQueueStatus(user.ID)

			api.ArenaManager.FactionStakedMechDashboardKeyChan <- []string{battle.FactionStakedMechDashboardKeyQueue}

		}

		// broadcast lobby

		api.ArenaManager.BattleLobbyDebounceBroadcastChan <- []string{bl.ID}

		// send invitations
		if len(req.Payload.InvitedUserIDs) > 0 {
			go func(host *boiler.Player, battleLobby *boiler.BattleLobby, invitedPlayerIDs []string) {
				b, err := json.Marshal(battleLobby)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("battle lobby", battleLobby).Msg("failed to marshal battle lobby data.")
					return
				}

				for _, playerID := range invitedPlayerIDs {
					// build system message
					msg := &boiler.SystemMessage{
						PlayerID: playerID,
						SenderID: server.SupremacyBattleUserID,
						DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeBattleLobbyInvitation)),
						Title:    "Lobby Invitation",
						Message:  fmt.Sprintf("%s #%d invited you to join lobby %s", host.Username.String, host.Gid, bl.Name),
						Data:     null.JSONFrom(b),
					}

					err = msg.Insert(gamedb.StdConn, boil.Infer())
					if err != nil {
						gamelog.L.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new system message into db")
						return
					}

					ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", playerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
				}
			}(user, bl, req.Payload.InvitedUserIDs)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if bl != nil {
		if !bl.AccessCode.Valid && !bl.GeneratedBySystem {
			// go api.Discord.SendBattleLobbyCreateMessage(bl.ID)
		}
	}

	reply(true)

	return nil
}

type BattleLobbyCloseRequest struct {
	Payload struct {
		BattleLobbyID string `json:"battle_lobby_id"`
	} `json:"payload"`
}

const HubKeyBattleLobbyClose = "BATTLE:LOBBY:CLOSE"

func (api *API) BattleLobbyClose(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	return nil
}

type BattleLobbyJoinRequest struct {
	Payload struct {
		BattleLobbyID string   `json:"battle_lobby_id"`
		MechIDs       []string `json:"mech_ids"`
		AccessCode    string   `json:"access_code"`
	} `json:"payload"`
}

const HubKeyBattleLobbyJoin = "BATTLE:LOBBY:JOIN"

func (api *API) BattleLobbyJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	availableMechIDs, err := MechAuthorisationFilter(user, factionID, req.Payload.MechIDs)
	if err != nil {
		return err
	}

	availableMechIDs, err = db.OverDamagedMechFilter(availableMechIDs)
	if err != nil {
		return err
	}

	if len(availableMechIDs) == 0 {
		return terror.Error(fmt.Errorf("no available mech"), "The provided mechs are not queueable.")
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

	blm, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.BattleLobbyID.EQ(bl.ID),
		boiler.BattleLobbiesMechWhere.QueuedByID.EQ(user.ID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to check queued mechs")
		return terror.Error(err, "Failed to check queued mechs")
	}

	// check password
	// if provided password is incorrect and this is the first time the players queue their mech in the lobby
	if bl.AccessCode.Valid && req.Payload.AccessCode != bl.AccessCode.String && blm == nil {
		return terror.Error(fmt.Errorf("incorrect password"), "The password is incorrect.")
	}

	affectedLobbyIDs := []string{bl.ID}

	// queued limit
	queueLimit := db.GetIntWithDefault(db.KeyPlayerQueueLimit, 10)

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// queue limit check
		blms, err := boiler.BattleLobbiesMechs(
			boiler.BattleLobbiesMechWhere.QueuedByID.EQ(user.ID),
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to load player battle queue mechs")
			return terror.Error(err, "Failed to load player battle queue mechs")
		}

		remainingQueueLimit := queueLimit - len(blms)

		if remainingQueueLimit <= 0 {
			return terror.Error(fmt.Errorf("reach queuing limit"), "You have reached the queuing limit.")
		}

		now := time.Now()
		// filter out mechs which are already in queue
		availableMechIDs, err = db.NonQueuedMechFilter(availableMechIDs)
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

		// check available amount

		if bl == nil {
			return terror.Error(fmt.Errorf("battle lobby not exist"), "Battle lobby does not exist.")
		}

		if bl.ExpiresAt.Valid && bl.ExpiresAt.Time.Before(time.Now()) {
			return terror.Error(fmt.Errorf("battle lobby is expired"), "Battle lobby has already expired.")
		}

		if bl.ReadyAt.Valid {
			return terror.Error(fmt.Errorf("battle lobby is already full"), "The battle lobby is already full.")
		}

		var battleLobbyMechs []*boiler.BattleLobbiesMech
		deployedMechIDs := []string{}
		availableSlotCount := bl.EachFactionMechAmount
		availableMaxDeployCount := bl.MaxDeployPerPlayer
		// check available slot
		if bl.R != nil {
			for _, battleLobbyMech := range bl.R.BattleLobbiesMechs {
				if battleLobbyMech.FactionID == factionID {
					availableSlotCount -= 1
				}

				if battleLobbyMech.QueuedByID == user.ID {
					availableMaxDeployCount -= 1
					remainingQueueLimit -= 1
				}

				// record the mechs in the battle lobby
				battleLobbyMechs = append(battleLobbyMechs, battleLobbyMech)
			}

			// return error, if not enough slots
			if availableSlotCount <= 0 {
				return terror.Error(fmt.Errorf("battle lobby is already full"), "The battle lobby is already full.")
			}

			if availableMaxDeployCount <= 0 {
				return terror.Error(fmt.Errorf("reach max deploy count"), "You have reached the max deploy count.")
			}

			if remainingQueueLimit <= 0 {
				return terror.Error(fmt.Errorf("reach queuing limit"), "You have reached the queuing limit.")
			}

		}

		// filled mech in remain slots
		for _, mechID := range availableMechIDs {
			// break, if no slot left or already reach max deploy count
			if len(deployedMechIDs) == availableSlotCount || len(deployedMechIDs) == availableMaxDeployCount {
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
				QueuedByID:    user.ID,
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
			bl.AccessCode = null.StringFromPtr(nil)
			_, err = bl.Update(tx, boil.Whitelist(boiler.BattleLobbyColumns.ReadyAt, boiler.BattleLobbyColumns.AccessCode))
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
					Name:                  helpers.GenerateAdjectiveName(),
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

				// set default lobby
				amount := db.GetDecimalWithDefault(db.KeySystemLobbyDefaultExtraReward, decimal.New(100, 18))

				if amount.GreaterThan(decimal.Zero) {
					paidTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
						FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
						ToUserID:             uuid.FromStringOrNil(server.SupremacyBattleUserID),
						Amount:               amount.StringFixed(0),
						TransactionReference: server.TransactionReference(fmt.Sprintf("top_up_system_lobby_default_reward|%s|%d", newBattleLobby.ID, time.Now().UnixNano())),
						Group:                string(server.TransactionGroupSupremacy),
						SubGroup:             string(server.TransactionGroupBattle),
						Description:          fmt.Sprintf("top up system lobby default reward %s.", newBattleLobby.ID),
					})
					if err != nil {
						return terror.Error(err, "Failed to top up reward.")
					}

					blr := &boiler.BattleLobbyExtraSupsReward{
						BattleLobbyID: newBattleLobby.ID,
						OfferedByID:   server.SupremacyBattleUserID,
						Amount:        amount,
						PaidTXID:      paidTXID,
					}

					err = blr.Insert(tx, boil.Infer())
					if err != nil {
						gamelog.L.Error().Err(err).Interface("battle lobby reward", blr).Msg("Failed to add battle lobby reward.")
						return terror.Error(err, "Failed to insert default system reward.")
					}
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

		if bl.GeneratedBySystem {
			go api.ArenaManager.FactionBattleLobbyMechsChecker(factionID)
		}

		if len(deployedMechIDs) > 0 {
			// pause mechs repair case
			err = api.ArenaManager.PauseRepairCases(deployedMechIDs)
			if err != nil {
				return err
			}

			api.ArenaManager.FactionStakedMechDashboardKeyChan <- []string{battle.FactionStakedMechDashboardKeyQueue}
		}

		// kick
		if lobbyReady {
			go api.ArenaManager.KickIdleArenas()

			for _, lm := range battleLobbyMechs {
				ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", factionID, lm.MechID), server.HubKeyPlayerAssetMechQueueSubscribe, &server.MechArenaInfo{
					Status:              server.MechArenaStatusQueue,
					CanDeploy:           false,
					BattleLobbyIsLocked: true,
				})
			}
		}

		api.ArenaManager.MechDebounceBroadcastChan <- deployedMechIDs

		// broadcast battle lobby
		api.ArenaManager.BattleLobbyDebounceBroadcastChan <- affectedLobbyIDs

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

		if !bl.AccessCode.Valid && !bl.IsAiDrivenMatch {
			// go api.Discord.SendBattleLobbyEditMessage(bl.ID, "")
		}
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

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		now := time.Now()

		var blms boiler.BattleLobbiesMechSlice
		blms, err = boiler.BattleLobbiesMechs(
			boiler.BattleLobbiesMechWhere.MechID.IN(req.Payload.MechIDs),
			boiler.BattleLobbiesMechWhere.QueuedByID.EQ(user.ID),
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

			// skip, if the battle lobby is expired.
			// NOTE: all the mechs in the expired lobbies SHOULD be evacuated from expire lobby func
			if bl.ExpiresAt.Valid && bl.ExpiresAt.Time.Before(time.Now()) {
				continue
			}

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

		playerLeftLobbies := []*server.BattleLobby{}
		lobbyIDs := []string{}
		for _, bl := range bls {
			lobbyIDs = append(lobbyIDs, bl.ID)

			// skip, if the player is the host of the lobby
			if bl.HostByID == user.ID {
				continue
			}

			// skip, if the player still has mechs queued in the lobby
			if bl.R != nil && bl.R.BattleLobbiesMechs != nil && slices.IndexFunc(bl.R.BattleLobbiesMechs, func(blm *boiler.BattleLobbiesMech) bool { return blm.QueuedByID == user.ID }) != -1 {
				continue
			}

			// otherwise, append the lobbies that the player has left
			playerLeftLobbies = append(playerLeftLobbies, &server.BattleLobby{
				BattleLobby: &boiler.BattleLobby{
					ID:        bl.ID,
					DeletedAt: null.TimeFrom(time.Now()),
				},
			})
		}

		// broadcast the lobbies player have left
		if len(playerLeftLobbies) > 0 {
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/involved_battle_lobbies", user.ID), server.HubKeyInvolvedBattleLobbyListUpdate, playerLeftLobbies)
		}

		api.ArenaManager.BattleLobbyDebounceBroadcastChan <- lobbyIDs

		// restart repair case
		go func() {
			err = api.ArenaManager.RestartRepairCases(leftMechIDs)
			if err != nil {
				gamelog.L.Error().Err(err).Strs("mech id list", leftMechIDs).Msg("Failed to restart repair cases")
			}

			api.ArenaManager.FactionStakedMechDashboardKeyChan <- []string{battle.FactionStakedMechDashboardKeyQueue}
		}()

		// broadcast player queue status
		go battle.BroadcastPlayerQueueStatus(user.ID)

		// broadcast new mech stat
		api.ArenaManager.MechDebounceBroadcastChan <- leftMechIDs

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

func (api *API) BattleLobbyListUpdate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// return all the unfinished lobbies
	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.EndedAt.IsNull(),
		boiler.BattleLobbyWhere.AccessCode.IsNull(),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
		qm.Load(
			qm.Rels(
				boiler.BattleLobbyRels.BattleLobbySupporters,
				boiler.BattleLobbySupporterRels.Supporter,
				boiler.PlayerRels.ProfileAvatar,
			),
		),
		qm.Load(
			qm.Rels(
				boiler.BattleLobbyRels.BattleLobbySupporterOptIns,
				boiler.BattleLobbySupporterOptInRels.Supporter,
				boiler.PlayerRels.ProfileAvatar,
			),
		),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards,
			boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
			boiler.BattleLobbyExtraSupsRewardWhere.DeletedAt.IsNull(),
		),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle lobbies.")
		return terror.Error(err, "Failed to load battle lobbies.")
	}

	resp, err := server.BattleLobbiesFromBoiler(bls)
	if err != nil {
		return err
	}

	reply(server.BattleLobbiesFactionFilter(resp, factionID, false))

	return nil
}

func (api *API) BattleLobbyUpdate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	battleLobbyID := chi.RouteContext(ctx).URLParam("battle_lobby_id")
	if battleLobbyID == "" {
		return nil
	}

	bl, err := db.GetBattleLobbyViaID(battleLobbyID)
	if err != nil {
		return err
	}

	filteredBattleLobbies, err := server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl})
	if err != nil {
		return err
	}

	if len(filteredBattleLobbies) > 0 {
		lobby := filteredBattleLobbies[0]

		isInvolved := lobby.HostByID == user.ID
		if !isInvolved && bl.R != nil {
			for _, mech := range bl.R.BattleLobbiesMechs {
				if mech.QueuedByID == user.ID {
					isInvolved = true
					break
				}
			}
		}
		reply(server.BattleLobbyInfoFilter(lobby, factionID, isInvolved))
	}

	return nil
}

func (api *API) PrivateBattleLobbyUpdate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	accessCode := chi.RouteContext(ctx).URLParam("access_code")
	if accessCode == "" {
		return nil
	}

	bl, err := db.GetBattleLobbyViaAccessCode(accessCode)
	if err != nil {
		return terror.Error(err, "Failed to load battle lobby")
	}

	filteredBattleLobbies, err := server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl})
	if err != nil {
		return err
	}

	if len(filteredBattleLobbies) > 0 {
		lobby := filteredBattleLobbies[0]
		reply(server.BattleLobbyInfoFilter(lobby, factionID, lobby.HostByID == user.ID))
	}

	return nil
}

type MechStakeRequest struct {
	Payload struct {
		MechIDs []string `json:"mech_ids"`
	} `json:"payload"`
}

const HubKeyMechsStake = "MECHS:STAKE"

func (api *API) MechStake(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MechStakeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	l := gamelog.L.With().Str("func", "MechStake").Str("player id", user.ID).Strs("staked mech id list", req.Payload.MechIDs).Logger()

	if !user.FactionPassExpiresAt.Valid || user.FactionPassExpiresAt.Time.Before(time.Now()) {
		return terror.Error(fmt.Errorf("required faction pass"), "Faction pass is required.")
	}

	mqas, err := db.MechsQueueAuthorisationDataGet(req.Payload.MechIDs)
	if err != nil {
		return err
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		l.Error().Err(err).Msg("Failed to start db transaction.")
		return terror.Error(err, "Failed to stake your mechs")
	}

	defer tx.Rollback()

	stakedMechIDs := []string{}
	for _, mqa := range mqas {
		if mqa.OwnerID != user.ID || mqa.StakedOnFactionID.Valid || mqa.XsynLocked || mqa.LockedToMarketplace || !mqa.PowerCoreID.Valid || !mqa.HasWeapon {
			continue
		}

		// stake mech
		sm := &boiler.StakedMech{
			MechID:    mqa.MechID,
			OwnerID:   user.ID,
			FactionID: factionID,
		}

		err = sm.Insert(tx, boil.Infer())
		if err != nil {
			l.Error().Interface("mech", sm).Err(err).Msg("Failed to insert staked mech.")
			return terror.Error(err, "Failed to insert staked mech")
		}

		stakedMechIDs = append(stakedMechIDs, sm.MechID)
	}

	if len(stakedMechIDs) == 0 {
		return terror.Error(fmt.Errorf("no mech is staked"), "None of the provided mechs is stakedable, please check the status of the provided mechs.")
	}

	err = tx.Commit()
	if err != nil {
		l.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to stake your mechs")
	}

	// send to debounce broadcast channel
	api.ArenaManager.MechDebounceBroadcastChan <- stakedMechIDs

	// update faction staked mech count
	api.ArenaManager.FactionStakedMechDashboardKeyChan <- []string{battle.FactionStakedMechDashboardKeyStaked}

	reply(true)
	return nil
}

const HubKeyMechsUnstake = "MECHS:UNSTAKE"

func (api *API) MechUnstake(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MechStakeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	l := gamelog.L.With().Str("func", "MechUnstake").Str("user id", user.ID).Strs("mech ids", req.Payload.MechIDs).Logger()

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		now := time.Now()

		var unstakedMechs []*boiler.StakedMech
		unstakedMechs, err = boiler.StakedMechs(
			boiler.StakedMechWhere.OwnerID.EQ(user.ID),
			boiler.StakedMechWhere.MechID.IN(req.Payload.MechIDs),
		).All(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("Failed to load staked mechs.")
			return terror.Error(err, "Failed to load staked mechs.")
		}

		if len(unstakedMechs) == 0 {
			return nil
		}

		unstakedMechIDs := []string{}
		var unstakedMechList []*db.MechBrief
		for _, unstakedMech := range unstakedMechs {
			unstakedMechIDs = append(unstakedMechIDs, unstakedMech.MechID)
			unstakedMechList = append(unstakedMechList, &db.MechBrief{
				ID:       unstakedMech.MechID,
				IsStaked: false,
			})
		}

		var blms boiler.BattleLobbiesMechSlice
		blms, err = boiler.BattleLobbiesMechs(
			boiler.BattleLobbiesMechWhere.MechID.IN(unstakedMechIDs),
			boiler.BattleLobbiesMechWhere.LockedAt.IsNull(),
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			qm.Load(boiler.BattleLobbiesMechRels.BattleLobby),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Strs("mech id list", unstakedMechIDs).Msg("Failed to load battle lobbies mech.")
			return terror.Error(err, "Failed to load battle lobby queuing records.")
		}

		// pull mechs out of the pending lobbies
		var tx *sql.Tx
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to leave battle lobby.")
		}

		defer tx.Rollback()

		// unstake mechs
		_, err = boiler.StakedMechs(boiler.StakedMechWhere.MechID.IN(unstakedMechIDs)).DeleteAll(tx)
		if err != nil {
			l.Error().Err(err).Msg("Failed to unstake mechs.")
			return terror.Error(err, "Failed to unstake mechs.")
		}

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

		// restart repair case
		go func() {
			err = api.ArenaManager.RestartRepairCases(leftMechIDs)
			if err != nil {
				gamelog.L.Error().Err(err).Strs("mech id list", leftMechIDs).Msg("Failed to restart repair cases")
			}

			api.ArenaManager.MechDebounceBroadcastChan <- leftMechIDs
		}()

		// load changed battle lobbies
		api.ArenaManager.BattleLobbyDebounceBroadcastChan <- changedBattleLobbyIDs

		// broadcast mech update
		api.ArenaManager.MechDebounceBroadcastChan <- unstakedMechIDs

		api.ArenaManager.FactionStakedMechDashboardKeyChan <- []string{battle.FactionStakedMechDashboardKeyStaked}

		// tell frontend to clean up the unstaked mechs from the list
		ws.PublishMessage(fmt.Sprintf("/faction/%s/staked_mechs", factionID), server.HubKeyFactionStakedMechs, unstakedMechList)

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

func (api *API) NextBattleDetails(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := api.ArenaManager.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	//ws.PublishMessage(fmt.Sprintf("/arena/%s/upcoming_battle", arena.ID), server.HubKeyNextBattleDetails, arena.GetLobbyDetails())
	reply(arena.GetLobbyDetails())
	return nil
}

type BattleLobbySupporterJoinRequest struct {
	Payload struct {
		BattleLobbyID string `json:"battle_lobby_id"`
		AccessCode    string `json:"access_code"`
	} `json:"payload"`
}

const HubKeyBattleLobbySupporterJoin = "BATTLE:LOBBY:SUPPORTER:JOIN"

func (api *API) BattleLobbySupporterJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "").Str("user id", user.ID).Logger()
	req := &BattleLobbySupporterJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	l = l.With().Interface("payload", req.Payload).Logger()

	bl := &boiler.BattleLobby{}

	// add support to battle lobby
	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// check lobby exists
		bl, err = db.GetBattleLobbyViaID(req.Payload.BattleLobbyID)
		if err != nil {
			return err
		}

		if !bl.ReadyAt.Valid {
			return terror.Error(fmt.Errorf("players caon only opt-in when lobby is ready"), "Player cannot opt in before the lobby is ready.")
		}

		// check if they have a mech in the battle
		if bl.R != nil && bl.R.BattleLobbiesMechs != nil {
			if slices.IndexFunc(bl.R.BattleLobbiesMechs, func(blm *boiler.BattleLobbiesMech) bool { return blm.QueuedByID == user.ID }) != -1 {
				return terror.Error(fmt.Errorf("mech owner cannot become a suppporter"), "Cannot opt in to support your own war machine.")
			}

			if slices.IndexFunc(bl.R.BattleLobbySupporters, func(bls *boiler.BattleLobbySupporter) bool { return bls.SupporterID == user.ID }) != -1 {
				return terror.Error(fmt.Errorf("already is a supporter"), "You have already joined as a supporter.")
			}

			if slices.IndexFunc(bl.R.BattleLobbySupporterOptIns, func(bls *boiler.BattleLobbySupporterOptIn) bool { return bls.SupporterID == user.ID }) != -1 {
				return terror.Error(fmt.Errorf("already is a supporter"), "You have already joined as a supporter")
			}
		}

		// if they provide the access code, bypass the opting in and straight up assign them as a supporter.
		if bl.AccessCode.Valid {
			// add them as a supporter
			bls := &boiler.BattleLobbySupporter{
				SupporterID:   user.ID,
				BattleLobbyID: bl.ID,
				FactionID:     factionID,
			}
			err = bls.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return err
			}

			api.ArenaManager.BattleLobbyDebounceBroadcastChan <- []string{bl.ID}

			go api.ArenaManager.BroadcastLobbyUpdate(bl.AssignedToArenaID.String)

			return nil
		}

		// add them as a supporter
		bls := &boiler.BattleLobbySupporterOptIn{
			SupporterID:   user.ID,
			BattleLobbyID: bl.ID,
			FactionID:     factionID,
		}
		err = bls.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}

		api.ArenaManager.BattleLobbyDebounceBroadcastChan <- []string{bl.ID}

		go api.ArenaManager.BroadcastLobbyUpdate(bl.AssignedToArenaID.String)

		return nil
	})
	if err != nil {
		return err
	}

	if bl != nil {
		if !bl.AccessCode.Valid && !bl.IsAiDrivenMatch {
			// go api.Discord.SendBattleLobbyEditMessage(bl.ID, "")
		}

	}

	reply(true)
	return nil
}

func (api *API) PlayerInvolvedBattleLobbies(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	// load involved battle lobby
	bls, err := boiler.BattleLobbies(
		qm.Where(fmt.Sprintf(
			`
				%[1]s IN (
					SELECT DISTINCT (%[1]s) FROM (
						SELECT %[1]s FROM %[2]s WHERE %[3]s = '%[4]s' AND %[5]s ISNULL AND %[6]s ISNULL
						UNION
						SELECT %[7]s AS id FROM %[8]s WHERE %[9]s = '%[4]s' AND %[10]s ISNULL AND %[11]s IS NULL AND %[12]s ISNULL
					) %[2]s
				)
			`,
			boiler.BattleLobbyTableColumns.ID,                  // 1
			boiler.TableNames.BattleLobbies,                    // 2
			boiler.BattleLobbyTableColumns.HostByID,            // 3
			user.ID,                                            // 4
			boiler.BattleLobbyTableColumns.EndedAt,             // 5
			boiler.BattleLobbyTableColumns.DeletedAt,           // 6
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID, // 7
			boiler.TableNames.BattleLobbiesMechs,               // 8
			boiler.BattleLobbiesMechTableColumns.QueuedByID,    // 9
			boiler.BattleLobbiesMechTableColumns.EndedAt,       // 10
			boiler.BattleLobbiesMechTableColumns.RefundTXID,    // 11
			boiler.BattleLobbiesMechTableColumns.DeletedAt,     // 12
		)),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
		qm.Load(
			qm.Rels(
				boiler.BattleLobbyRels.BattleLobbySupporters,
				boiler.BattleLobbySupporterRels.Supporter,
				boiler.PlayerRels.ProfileAvatar,
			),
		),
		qm.Load(
			qm.Rels(
				boiler.BattleLobbyRels.BattleLobbySupporterOptIns,
				boiler.BattleLobbySupporterOptInRels.Supporter,
				boiler.PlayerRels.ProfileAvatar,
			),
		),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards,
			boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
			boiler.BattleLobbyExtraSupsRewardWhere.DeletedAt.IsNull(),
		),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load battle lobby")
	}

	resp, err := server.BattleLobbiesFromBoiler(bls)
	if err != nil {
		return err
	}

	reply(server.BattleLobbiesFactionFilter(resp, factionID, true))

	return nil
}

type BattleLobbyTopUpRewardRequest struct {
	Payload struct {
		BattleLobbyID string          `json:"battle_lobby_id"`
		Amount        decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

const HubKeyBattleLobbyTopUpReward = "BATTLE:LOBBY:TOP:UP:REWARD"

func (api *API) BattleLobbyTopUpReward(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyTopUpRewardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !req.Payload.Amount.IsPositive() {
		return terror.Error(fmt.Errorf("amount less than zero"), "Reward must be greater than zero.")
	}

	l := gamelog.L.With().Str("func", "BattleLobbyTopUpReward").Str("user id", user.ID).Str("battle lobby id", req.Payload.BattleLobbyID).Str("amount", req.Payload.Amount.Mul(decimal.New(1, 18)).String()).Logger()
	var bl *boiler.BattleLobby
	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		bl, err = boiler.BattleLobbies(
			boiler.BattleLobbyWhere.ID.EQ(req.Payload.BattleLobbyID),
		).One(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to load battle lobby.")
		}

		if bl.EndedAt.Valid {
			return terror.Error(fmt.Errorf("battle already ended"), "Failed to battle is already ended.")
		}

		if bl.AssignedToBattleID.Valid {
			return terror.Error(fmt.Errorf("battle has already started"), "The battle has already started.")
		}

		amount := req.Payload.Amount.Mul(decimal.New(1, 18))

		var paidTxID string
		paidTxID, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.Must(uuid.FromString(user.ID)),
			ToUserID:             uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
			Amount:               amount.StringFixed(0),
			TransactionReference: server.TransactionReference(fmt.Sprintf("battle_lobby_extra_reward|%s|%d", bl.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupBattle),
			Description:          "adding extra sups reward for battle lobby.",
		})
		if err != nil {
			return terror.Error(err, "Failed to add sups reward, check your balance and try again.")
		}

		blr := &boiler.BattleLobbyExtraSupsReward{
			BattleLobbyID: bl.ID,
			OfferedByID:   user.ID,
			Amount:        amount,
			PaidTXID:      paidTxID,
		}

		err = blr.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Interface("battle lobby reward", blr).Msg("Failed to add battle lobby reward.")
			return terror.Error(err, "Failed to add battle lobby reward.")
		}

		if !bl.AccessCode.Valid && !bl.IsAiDrivenMatch {
			// go api.Discord.SendBattleLobbyEditMessage(bl.ID, "")
		}
		api.ArenaManager.BattleLobbyDebounceBroadcastChan <- []string{bl.ID}
		return nil
	})

	reply(true)

	return nil
}

// MechAuthorisationFilter return error if any mechs is not available to queue
func MechAuthorisationFilter(player *boiler.Player, factionID string, mechIDs []string) ([]string, error) {
	if len(mechIDs) == 0 {
		return []string{}, nil
	}

	l := gamelog.L.With().Str("func", "MechAuthorisationFilter").Logger()

	mqas, err := db.MechsQueueAuthorisationDataGet(mechIDs)
	if err != nil {
		return nil, err
	}

	availableList := []string{}
	for _, mqa := range mqas {
		if mqa.LockedToMarketplace {
			l.Debug().Err(err).Msg("mech is locked in market place")
			continue
		}

		if mqa.XsynLocked {
			l.Debug().Err(err).Msg("mech is locked in Xsyn")
			continue
		}

		if !mqa.IsAvailable {
			l.Debug().Err(err).Msg("mech is currently not available")
			continue
		}

		if !mqa.PowerCoreID.Valid {
			l.Debug().Err(err).Msg("mech does not have power core.")
			continue
		}

		if !mqa.HasWeapon {
			l.Debug().Err(err).Msg("mech does not have weapons equipped.")
			continue
		}

		if mqa.StakedOnFactionID.Valid {
			if !player.FactionPassExpiresAt.Valid || player.FactionPassExpiresAt.Time.Before(time.Now()) {
				return nil, terror.Error(fmt.Errorf("faction pass is expired"), "Required faction pass to queue staked mechs.")
			}
			// check faction id if the mech is staked in faction list
			if mqa.StakedOnFactionID.String != factionID || mqa.OwnerID == player.ID {
				continue
			}
		} else {
			// otherwise, check owner id
			if mqa.OwnerID != player.ID {
				continue
			}
		}

		availableList = append(availableList, mqa.MechID)
	}
	return availableList, nil
}

func (api *API) PlayerBrowserAlert(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	// send battle lobby mechs for now
	queries := []qm.QueryMod{
		qm.Select(
			boiler.BattleLobbyTableColumns.AssignedToArenaID,
			boiler.BlueprintMechTableColumns.Label,
			boiler.MechTableColumns.ID,
			boiler.MechTableColumns.Name,
			boiler.BattleLobbiesMechTableColumns.QueuedByID,
			fmt.Sprintf(
				"(SELECT %s FROM %s WHERE %s = %s) AS stake_mech_owner_id",
				boiler.StakedMechTableColumns.OwnerID,
				boiler.TableNames.StakedMechs,
				boiler.StakedMechTableColumns.MechID,
				boiler.MechTableColumns.ID,
			),
		),
		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s NOTNULL AND %s ISNULL AND %s ISNULL) %s",
			boiler.TableNames.BattleLobbies,
			boiler.BattleLobbyTableColumns.AssignedToArenaID,
			boiler.BattleLobbyTableColumns.EndedAt,
			boiler.BattleLobbyTableColumns.DeletedAt,
			boiler.TableNames.BattleLobbies,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.BattleLobbyTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Mechs,
			boiler.MechTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.MechID,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			boiler.BlueprintMechTableColumns.ID,
			boiler.MechTableColumns.BlueprintID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load battle lobby mechs")
	}

	data := []*server.BattleLobbyMechsAlert{}
	for rows.Next() {
		arenaID := ""
		mechLabel := ""
		mechID := ""
		mechName := ""
		queuedByID := ""
		mechOwnerID := null.String{}

		err = rows.Scan(&arenaID, &mechLabel, &mechID, &mechName, &queuedByID, &mechOwnerID)
		if err != nil {
			return terror.Error(err, "Failed to scan battle lobby mech")
		}

		if queuedByID != user.ID && (!mechOwnerID.Valid || mechOwnerID.String != user.ID) {
			continue
		}

		index := slices.IndexFunc(data, func(bla *server.BattleLobbyMechsAlert) bool { return bla.ArenaID == arenaID })
		if index == -1 {
			arena, err := api.ArenaManager.GetArena(arenaID)
			if err != nil {
				continue
			}

			data = append(data, &server.BattleLobbyMechsAlert{
				ArenaID:    arena.ID,
				ArenaName:  arena.Name,
				MechAlerts: []*server.MechAlert{},
			})

			index = len(data) - 1
		}

		if slices.IndexFunc(data[index].MechAlerts, func(ma *server.MechAlert) bool { return ma.ID == mechID }) == -1 {
			ma := &server.MechAlert{
				ID:   mechID,
				Name: mechName,
			}
			if ma.Name == "" {
				ma.Name = mechLabel
			}

			data[index].MechAlerts = append(data[index].MechAlerts, ma)
		}
	}

	if len(data) > 0 {
		reply(&server.PlayerBrowserAlertStruct{
			Title: "MECH_IN_BATTLE",
			Data:  data,
		})
	}

	return nil
}
