package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// VoteControllerWS holds handlers for checking server status
type VoteControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewVoteController creates the check hub
func NewVoteController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *VoteControllerWS {
	voteHub := &VoteControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "vote_hub"),
		API:  api,
	}

	api.SecureUserFactionCommand(HubKeyFactionVotePrice, voteHub.FactionVotePrice)
	api.SecureUserFactionCommand(HubKeyVoteAbilityRight, voteHub.AbilityRight)
	api.SecureUserFactionCommand(HubKeyAbilityLocationSelect, voteHub.AbilityLocationSelect)

	// subscription
	api.SecureUserFactionSubscribeCommand(HubKeyVoteWinnerAnnouncement, voteHub.WinnerAnnouncementSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyVoteBattleAbilityUpdated, voteHub.BattleAbilityUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyVoteStageUpdated, voteHub.VoteStageUpdateSubscribeHandler)

	// net message subscription
	api.NetSubscribeCommand(HubKeyLiveVoteUpdated, voteHub.LiveVoteUpdateSubscribeHandler)
	api.NetSubscribeCommand(HubKeyWarMachineLocationUpdated, voteHub.WarMachineLocationUpdateSubscribeHandler)
	api.NetSubscribeCommand(HubKeyViewerLiveCountUpdated, voteHub.ViewerLiveCountUpdateSubscribeHandler)
	api.NetSubscribeCommand(HubKeySpoilOfWarUpdated, voteHub.SpoilOfWarUpdateSubscribeHandler)
	api.NetSecureUserFactionSubscribeCommand(HubKeyAbilityRightRatioUpdated, voteHub.AbilityRightRatioUpdateSubscribeHandler)
	api.NetSecureUserFactionSubscribeCommand(HubKeyFactionAbilityPriceUpdated, voteHub.FactionAbilityPriceUpdateSubscribeHandler)
	api.NetSecureUserFactionSubscribeCommand(HubKeyFactionVotePriceUpdated, voteHub.FactionVotePriceUpdateSubscribeHandler)

	return voteHub
}

const HubKeyFactionVotePrice hub.HubCommandKey = "FACTION:VOTE:PRICE"

func (vc *VoteControllerWS) FactionVotePrice(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	gamelog.L.Info().Str("fn", "FactionVotePrice").RawJSON("req", payload).Msg("ws handler")
	hcd := vc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return terror.Error(terror.ErrForbidden)
	}
	if hcd == nil {
		return terror.Error(fmt.Errorf("hub client details returned nil"), "Error while getting vote price, please contact support.")
	}
	if vc.API.votePriceSystem == nil {
		return terror.Error(fmt.Errorf("nil vote price system"), "Error finding voting system, please contact support.")
	}
	if _, ok := vc.API.votePriceSystem.FactionVotePriceMap[hcd.FactionID]; !ok {
		return terror.Error(fmt.Errorf("unable to find faction id %s in FactionVotePriceMap", hcd.FactionID), "Error finding faction vote details, please contact support.")
	}
	reply(vc.API.votePriceSystem.FactionVotePriceMap[hcd.FactionID].CurrentVotePriceSups.Int.String())

	return nil
}

type AbilityRightVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		VoteAmount int64 `json:"voteAmount"` // 1, 10, 100
	} `json:"payload"`
}

const HubKeyVoteAbilityRight hub.HubCommandKey = "VOTE:ABILITY:RIGHT"

func (vc *VoteControllerWS) AbilityRight(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	//TODO ALEX: fix
	//gamelog.L.Info().Str("fn", "AbilityRight").RawJSON("req", payload).Msg("ws handler")
	//req := &AbilityRightVoteRequest{}
	//err := json.Unmarshal(payload, req)
	//if err != nil {
	//	return terror.Error(err, "There was a problem parsing the data")
	//}
	//
	//// get user detail
	//userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	//if userID.IsNil() {
	//	return terror.Error(errors.New("user ID is nil"), "There was a problem getting your user")
	//}
	//
	////TODO ALEX: fix
	////if vc.API.BattleArena.GetCurrentState().State != server.StateMatchStart {
	////	gamelog.L.Warn().Str("battle_arena_state", string(vc.API.BattleArena.GetCurrentState().State)).Str("want", string(server.StateMatchStart)).Msg("wrong game state")
	////	return nil
	////}
	//
	//// check voting phase first
	//vc.API.votePhaseChecker.RLock()
	//if vc.API.votePhaseChecker.Phase != VotePhaseVoteAbilityRight && vc.API.votePhaseChecker.Phase != VotePhaseNextVoteWin {
	//	gamelog.L.
	//		Warn().
	//		Str("server_phase", string(vc.API.votePhaseChecker.Phase)).
	//		Msg("wrong vote phase (can only vote in next vote win or ability vote right phase)")
	//	vc.API.votePhaseChecker.RUnlock()
	//	return nil
	//}
	//vc.API.votePhaseChecker.RUnlock()
	//
	//if req.Payload.VoteAmount <= 0 {
	//	gamelog.L.Warn().Int64("amt", req.Payload.VoteAmount).Msg("negative or zero vote amount")
	//	return nil
	//}
	//
	//hcd := vc.API.UserMap.GetUserDetail(wsc)
	//if hcd == nil {
	//	return terror.Error(err, "Could not get user details")
	//}
	//
	//// get current faction vote price
	//pricePerVote := server.BigInt{Int: *big.NewInt(0)}
	//pricePerVote.Add(&pricePerVote.Int, &vc.API.votePriceSystem.FactionVotePriceMap[hcd.FactionID].CurrentVotePriceSups.Int)
	//
	//totalSups := server.BigInt{Int: *big.NewInt(0)}
	//totalSups.Add(&totalSups.Int, &pricePerVote.Int)
	//totalSups.Mul(&totalSups.Int, big.NewInt(req.Payload.VoteAmount))
	//
	//// deliver vote
	//
	//vc.API.VotingCycle(func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
	//	reason := fmt.Sprintf("battle:%vote_ability_right:%s", vc.API.BattleArena.CurrentBattleID(), va.BattleAbility.ID)
	//	go vc.API.Passport.SpendSupMessage(passport.SpendSupsReq{
	//		FromUserID:           userID,
	//		Amount:               totalSups.String(),
	//		TransactionReference: server.TransactionReference(fmt.Sprintf("%s|%s", reason, uuid.Must(uuid.NewV4()))),
	//		Group:                "battle",
	//		SubGroup:             vc.API.BattleArena.CurrentBattleID().String(),
	//		Description:          "battle vote.",
	//		NotSafe:              true,
	//	}, func(transaction string) {
	//		// check voting phase first
	//		vc.API.votePhaseChecker.RLock()
	//		if vc.API.votePhaseChecker.Phase != VotePhaseVoteAbilityRight && vc.API.votePhaseChecker.Phase != VotePhaseNextVoteWin {
	//			go vc.API.Passport.ReleaseTransactions([]string{transaction})
	//			gamelog.L.
	//				Warn().
	//				Str("server_phase", string(vc.API.votePhaseChecker.Phase)).
	//				Msg("wrong vote phase (can only vote in next vote win or ability vote right phase)")
	//			vc.API.votePhaseChecker.RUnlock()
	//			return
	//		}
	//		vc.API.votePhaseChecker.RUnlock()
	//
	//		fts.Lock()
	//		fts.Transactions = append(fts.Transactions, transaction)
	//
	//		vc.API.liveSupsSpend[hcd.FactionID].Lock()
	//		vc.API.liveSupsSpend[hcd.FactionID].TotalVote.Add(&vc.API.liveSupsSpend[hcd.FactionID].TotalVote.Int, &totalSups.Int)
	//		vc.API.liveSupsSpend[hcd.FactionID].Unlock()
	//
	//		vc.API.increaseFactionVoteTotal(hcd.FactionID, req.Payload.VoteAmount)
	//		// go vc.API.ClientVoted(wsc)
	//		vc.API.UserMultiplier.Voted(userID)
	//
	//		switch hcd.FactionID {
	//		case server.RedMountainFactionID:
	//			ftv.RedMountainTotalVote += req.Payload.VoteAmount
	//		case server.BostonCyberneticsFactionID:
	//			ftv.BostonTotalVote += req.Payload.VoteAmount
	//		case server.ZaibatsuFactionID:
	//			ftv.ZaibatsuTotalVote += req.Payload.VoteAmount
	//		}
	//
	//		// update vote result, if it is vote ability right phase
	//		vc.API.votePhaseChecker.RLock()
	//		if vc.API.votePhaseChecker.Phase == VotePhaseVoteAbilityRight {
	//			_, ok := fuvm[hcd.FactionID][userID]
	//			if !ok {
	//				fuvm[hcd.FactionID][userID] = 0
	//			}
	//
	//			fuvm[hcd.FactionID][userID] += req.Payload.VoteAmount
	//			vc.API.votePhaseChecker.RUnlock()
	//			fts.Unlock()
	//
	//			go wsc.SendWithMessageType(getRatio(ftv.RedMountainTotalVote, ftv.BostonTotalVote, ftv.ZaibatsuTotalVote), websocket.MessageBinary)
	//
	//			return
	//		}
	//		vc.API.votePhaseChecker.RUnlock()
	//		fts.Unlock()
	//
	//		// record user vote map
	//		if _, ok := uvm[userID]; !ok {
	//			uvm[userID] = 0
	//		}
	//		uvm[userID] += req.Payload.VoteAmount
	//
	//		// set current user as winner
	//		vw.List = append(vw.List, userID)
	//
	//		// voting phase change
	//		vc.API.votePhaseChecker.Lock()
	//		vc.API.votePhaseChecker.Phase = VotePhaseLocationSelect
	//		vc.API.votePhaseChecker.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)
	//		vc.API.votePhaseChecker.Unlock()
	//
	//		go vc.API.BroadcastGameNotificationAbility(GameNotificationTypeBattleAbility, &GameNotificationAbility{
	//			User:    hcd.Brief(),
	//			Ability: va.FactionAbilityMap[hcd.FactionID].Brief(),
	//		})
	//
	//		// announce winner
	//		go vc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, userID)), &WinnerSelectAbilityLocation{
	//			GameAbility: va.FactionAbilityMap[hcd.FactionID],
	//			EndTime:     vc.API.votePhaseChecker.EndTime,
	//		})
	//
	//		// broadcast current stage to faction users
	//		go vc.API.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), vc.API.votePhaseChecker)
	//
	//		// start vote listener
	//		if vct.VotingStageListener.NextTick == nil || vct.VotingStageListener.NextTick.Before(time.Now()) {
	//			vct.VotingStageListener.Start()
	//		}
	//
	//		// stop vote right result broadcaster
	//		if vct.AbilityRightResultBroadcaster.NextTick != nil {
	//			vct.AbilityRightResultBroadcaster.Stop()
	//		}
	//	}, func(err error) {})
	//})
	return nil
}

type AbilityLocationSelectRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		XIndex int `json:"x"`
		YIndex int `json:"y"`
	} `json:"payload"`
}

const HubKeyAbilityLocationSelect hub.HubCommandKey = "ABILITY:LOCATION:SELECT"

func (vc *VoteControllerWS) AbilityLocationSelect(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	gamelog.L.Info().Str("fn", "AbilityLocationSelect").RawJSON("req", payload).Msg("ws handler")
	return nil
	//TODO ALEX: reimplement
	//req := &AbilityLocationSelectRequest{}
	//err := json.Unmarshal(payload, req)
	//if err != nil {
	//	return terror.Error(err)
	//}
	//
	//userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	//if userID.IsNil() {
	//	return terror.Error(terror.ErrInvalidInput)
	//}
	//
	//hcd := vc.API.UserMap.GetUserDetail(wsc)
	//if hcd == nil {
	//	return terror.Error(fmt.Errorf("user not found"))
	//}
	//
	//if vc.API.votePhaseChecker.Phase != VotePhaseLocationSelect {
	//	return terror.Error(terror.ErrForbidden, "Error - Invalid voting phase")
	//}
	//
	//vc.API.VotingCycle(func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
	//
	//	// check voting phase
	//	vc.API.votePhaseChecker.RLock()
	//	if vc.API.votePhaseChecker.Phase != VotePhaseLocationSelect {
	//		vc.API.votePhaseChecker.RUnlock()
	//		return
	//	}
	//	vc.API.votePhaseChecker.RUnlock()
	//
	//	// check winner user id
	//	if vw.List[0] != userID {
	//		return
	//	}
	//
	//	// record ability animation
	//	selectedX := req.Payload.XIndex
	//	selectedY := req.Payload.YIndex
	//
	//	err = vc.API.BattleArena.GameAbilityTrigger(&server.GameAbilityEvent{
	//		GameAbilityID:       &va.FactionAbilityMap[hcd.FactionID].ID,
	//		IsTriggered:         true,
	//		GameClientAbilityID: va.FactionAbilityMap[hcd.FactionID].GameClientAbilityID,
	//		TriggeredByUserID:   &userID,
	//		TriggeredByUsername: &hcd.Username,
	//		TriggeredOnCellX:    &selectedX,
	//		TriggeredOnCellY:    &selectedY,
	//	})
	//
	//	if err != nil {
	//		return
	//	}
	//
	//	// record triggered abilities
	//	vc.API.UserMultiplier.AbilityTriggered(hcd.FactionID, userID, va.FactionAbilityMap[hcd.FactionID])
	//
	//	// clean up the transactions after ability is triggered
	//	fts.Lock()
	//	defer fts.Unlock()
	//
	//	// broadcast notification
	//	go vc.API.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
	//		Type:        LocationSelectTypeTrigger,
	//		CurrentUser: hcd.Brief(),
	//		X:           &req.Payload.XIndex,
	//		Y:           &req.Payload.YIndex,
	//		Ability:     va.BattleAbility.Brief(),
	//	})
	//
	//	// get random ability collection set
	//	battleAbility, factionAbilityMap, err := vc.API.BattleArena.RandomBattleAbility()
	//	if err != nil {
	//		return
	//	}
	//
	//	go vc.API.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)
	//
	//	// initialise new ability collection
	//	va.BattleAbility = battleAbility
	//
	//	// initialise new game ability map
	//	for fid, ability := range factionAbilityMap {
	//		va.FactionAbilityMap[fid] = ability
	//	}
	//
	//	// broadcast next stage
	//	vc.API.votePhaseChecker.Lock()
	//	vc.API.votePhaseChecker.Phase = VotePhaseVoteCooldown
	//	vc.API.votePhaseChecker.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)
	//	vc.API.votePhaseChecker.Unlock()
	//
	//	// stop vote price update when cooldown
	//	if vc.API.votePriceSystem.VotePriceUpdater.NextTick != nil {
	//		vc.API.votePriceSystem.VotePriceUpdater.Stop()
	//	}
	//
	//	// broadcast current stage to faction users
	//	go vc.API.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), vc.API.votePhaseChecker)
	//
	//	vc.API.UserMultiplier.PickedLocation(userID)
	//})
	return nil
}

/***************
* Subscription *
***************/

const HubKeyVoteWinnerAnnouncement hub.HubCommandKey = "VOTE:WINNER:ANNOUNCEMENT"

// WinnerAnnouncementSubscribeHandler subscribe on vote winner to pick location
func (vc *VoteControllerWS) WinnerAnnouncementSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "WinnerAnnouncementSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, userID))

	return req.TransactionID, busKey, nil
}

const HubKeyVoteBattleAbilityUpdated hub.HubCommandKey = "VOTE:BATTLE:ABILITY:UPDATED"

// BattleAbilityUpdateSubscribeHandler to subscribe to game event
func (vc *VoteControllerWS) BattleAbilityUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "BattleAbilityUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	//TODO ALEX: fix
	//req := &hub.HubCommandRequest{}
	//err := json.Unmarshal(payload, req)
	//if err != nil {
	//	return "", "", terror.Error(err, "Invalid request received")
	//}
	//
	//// only pass ability when battle started and vote phase is not on hold
	//if vc.API.BattleArena.GetCurrentState().State == server.StateMatchStart {
	//	vc.API.votePhaseChecker.RLock()
	//	defer vc.API.votePhaseChecker.RUnlock()
	//	if vc.API.votePhaseChecker.Phase != VotePhaseHold {
	//		vc.API.VotingCycle(func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
	//			reply(va.BattleAbility)
	//		})
	//	}
	//}
	//
	//return req.TransactionID, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), nil
	return "", "", nil
}

const HubKeyVoteStageUpdated hub.HubCommandKey = "VOTE:STAGE:UPDATED"

// VoteStageUpdateSubscribeHandler to subscribe on vote stage
func (vc *VoteControllerWS) VoteStageUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	gamelog.L.Info().Str("fn", "VoteStageUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	reply(vc.API.votePhaseChecker)

	return req.TransactionID, messagebus.BusKey(HubKeyVoteStageUpdated), nil
}

/***************************
* Net Message Subscription *
***************************/

const HubKeyLiveVoteUpdated hub.HubCommandKey = "LIVE:VOTE:UPDATED"

// AbilityRightRatioUpdateSubscribeHandler to subscribe on ability right ratio update
func (vc *VoteControllerWS) LiveVoteUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "LiveVoteUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	return messagebus.NetBusKey(HubKeyLiveVoteUpdated), nil
}

const HubKeyWarMachineLocationUpdated hub.HubCommandKey = "WAR:MACHINE:LOCATION:UPDATED"

// AbilityRightRatioUpdateSubscribeHandler to subscribe on ability right ratio update
func (vc *VoteControllerWS) WarMachineLocationUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "WarMachineLocationUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	return messagebus.NetBusKey(HubKeyWarMachineLocationUpdated), nil
}

const HubKeyViewerLiveCountUpdated hub.HubCommandKey = "VIEWER:LIVE:COUNT:UPDATED"

func (vc *VoteControllerWS) ViewerLiveCountUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "ViewerLiveCountUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	return messagebus.NetBusKey(HubKeyViewerLiveCountUpdated), nil
}

const HubKeySpoilOfWarUpdated hub.HubCommandKey = "SPOIL:OF:WAR:UPDATED"

func (vc *VoteControllerWS) SpoilOfWarUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "SpoilOfWarUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	return messagebus.NetBusKey(HubKeySpoilOfWarUpdated), nil
}

const HubKeyAbilityRightRatioUpdated hub.HubCommandKey = "ABILITY:RIGHT:RATIO:UPDATED"

// AbilityRightRatioUpdateSubscribeHandler to subscribe on ability right ratio update
func (vc *VoteControllerWS) AbilityRightRatioUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "AbilityRightRatioUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	busKey := messagebus.NetBusKey(HubKeyAbilityRightRatioUpdated)
	return busKey, nil
}

const HubKeyFactionAbilityPriceUpdated hub.HubCommandKey = "FACTION:ABILITY:PRICE:UPDATED"

func (vc *VoteControllerWS) FactionAbilityPriceUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "FactionAbilityPriceUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	// get user faction
	hcd := vc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return "", terror.Error(fmt.Errorf("user not found"))
	}

	busKey := messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilityPriceUpdated, hcd.FactionID))

	return busKey, nil
}

const HubKeyFactionVotePriceUpdated hub.HubCommandKey = "FACTION:VOTE:PRICE:UPDATED"

func (vc *VoteControllerWS) FactionVotePriceUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte) (messagebus.NetBusKey, error) {
	gamelog.L.Info().Str("fn", "FactionVotePriceUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	// get user faction
	hcd := vc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return "", terror.Error(fmt.Errorf("user not found"))
	}

	busKey := messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionVotePriceUpdated, hcd.FactionID))

	return busKey, nil
}
