package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"server"
	"server/battle_arena"
	"time"

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

	api.Command(HubKeyFactionColour, voteHub.FactionColour)

	api.SecureUserFactionCommand(HubKeyFactionVotePrice, voteHub.FactionVotePrice)
	api.SecureUserFactionCommand(HubKeyVoteAbilityRight, voteHub.AbilityRight)
	api.SecureUserFactionCommand(HubKeyAbilityLocationSelect, voteHub.AbilityLocationSelect)

	// subscription
	api.SecureUserFactionSubscribeCommand(HubKeyVoteWinnerAnnouncement, voteHub.WinnerAnnouncementSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyVoteBattleAbilityUpdated, voteHub.BattleAbilityUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyVoteStageUpdated, voteHub.VoteStageUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyFactionWarMachineQueueUpdated, voteHub.FactionWarMachineQueueUpdateSubscribeHandler)

	return voteHub
}

const HubKeyFactionColour hub.HubCommandKey = "FACTION:COLOUR"

type FactionColourRespose struct {
	RedMountain string `json:"redMountain"`
	Boston      string `json:"boston"`
	Zaibatsu    string `json:"zaibatsu"`
}

func (vc *VoteControllerWS) FactionColour(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	if vc.API.factionMap == nil {
		return terror.Error(terror.ErrForbidden, "faction data not ready yet")
	}

	reply(&FactionColourRespose{
		RedMountain: vc.API.factionMap[server.RedMountainFactionID].Theme.Primary,
		Boston:      vc.API.factionMap[server.BostonCyberneticsFactionID].Theme.Primary,
		Zaibatsu:    vc.API.factionMap[server.ZaibatsuFactionID].Theme.Primary,
	})

	return nil
}

const HubKeyFactionVotePrice hub.HubCommandKey = "FACTION:VOTE:PRICE"

func (vc *VoteControllerWS) FactionVotePrice(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	hcd, err := vc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(terror.ErrForbidden)
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
	req := &AbilityRightVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	// get user detail
	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	hcd, err := vc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(terror.ErrForbidden)
	}

	// check voting phase first
	if vc.API.votePhaseChecker.Phase != VotePhaseVoteAbilityRight && vc.API.votePhaseChecker.Phase != VotePhaseNextVoteWin {
		return terror.Error(terror.ErrForbidden, "Error - Invalid voting phase")
	}

	// get current faction vote price
	pricePerVote := server.BigInt{Int: *big.NewInt(0)}
	pricePerVote.Add(&pricePerVote.Int, &vc.API.votePriceSystem.FactionVotePriceMap[hcd.FactionID].CurrentVotePriceSups.Int)

	totalSups := server.BigInt{Int: *big.NewInt(0)}
	totalSups.Add(&totalSups.Int, &pricePerVote.Int)
	totalSups.Mul(&totalSups.Int, big.NewInt(req.Payload.VoteAmount))

	// deliver vote
	errChan := make(chan error)

	vc.API.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
		if vs.Phase != VotePhaseVoteAbilityRight && vs.Phase != VotePhaseNextVoteWin {
			errChan <- terror.Error(terror.ErrInvalidInput, "Error - Invalid voting phase")
			return
		}

		// pay sups
		reason := fmt.Sprintf("battle:%s|vote_ability_collection:%s", vc.API.BattleArena.CurrentBattleID(), va.BattleAbility.ID)
		supTransactionReference, err := vc.API.Passport.SendHoldSupsMessage(context.Background(), userID, totalSups, req.TransactionID, reason)
		if err != nil {
			errChan <- terror.Error(err, "Error - Failed to pay sups")
			return
		}

		switch hcd.FactionID {
		case server.RedMountainFactionID:
			ftv.RedMountainTotalVote += req.Payload.VoteAmount
		case server.BostonCyberneticsFactionID:
			ftv.BostonTotalVote += req.Payload.VoteAmount
		case server.ZaibatsuFactionID:
			ftv.ZaibatsuTotalVote += req.Payload.VoteAmount
		}

		// update vote result, if it is vote ability right phase
		if vs.Phase == VotePhaseVoteAbilityRight {
			_, ok := fuvm[hcd.FactionID][userID]
			if !ok {
				fuvm[hcd.FactionID][userID] = make(map[server.TransactionReference]int64)
			}

			fuvm[hcd.FactionID][userID][supTransactionReference] = req.Payload.VoteAmount

			errChan <- nil
			return
		}

		// if next vote win, set user as winner once the transaction is successful
		transactions, err := vc.API.Passport.CommitTransactions(ctx, []server.TransactionReference{supTransactionReference})
		if err != nil {
			errChan <- terror.Error(err, "Error - Failed to check transactions")
			return
		}

		for _, chktx := range transactions {
			// return if transaction failed
			if chktx.Status == server.TransactionFailed {
				errChan <- terror.Error(terror.ErrInvalidInput, "Error - Transaction failed")
				return
			}
		}

		// set current user as winner
		vw.List = append(vw.List, userID)

		// voting phase change
		vc.API.votePhaseChecker.Phase = VotePhaseLocationSelect
		vs.Phase = VotePhaseLocationSelect
		vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

		go vc.API.BroadcastGameNotification(GameNotificationTypeText, fmt.Sprintf("User %s is selecting location for the ability %s", hcd.Username, va.FactionAbilityMap[hcd.FactionID].Label))

		// announce winner
		vc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, userID)), &WinnerSelectAbilityLocation{
			FactionAbility: *va.FactionAbilityMap[hcd.FactionID],
			EndTime:        vs.EndTime,
		})

		// broadcast current stage to faction users
		vc.API.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

		// start vote listener
		if vct.VotingStageListener.NextTick == nil {
			vct.VotingStageListener.Start()
		}

		// stop vote right result broadcaster
		if vct.AbilityRightResultBroadcaster.NextTick != nil {
			vct.AbilityRightResultBroadcaster.Stop()
		}

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err, "Failed to vote")
	}

	// store vote amount to live voting data after vote success
	vc.API.liveSupsSpend[hcd.FactionID] <- func(lvd *LiveVotingData) {
		lvd.TotalVote.Add(&lvd.TotalVote.Int, &totalSups.Int)
	}

	// add vote count to faction price channels
	vc.API.increaseFactionVoteTotal(hcd.FactionID, req.Payload.VoteAmount)

	vc.API.ClientVoted(wsc)
	reply(true)

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
	req := &AbilityLocationSelectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	hcd, err := vc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	if vc.API.votePhaseChecker.Phase != VotePhaseLocationSelect {
		return terror.Error(terror.ErrForbidden, "Error - Invalid voting phase")
	}

	errChan := make(chan error)
	vc.API.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
		// check voting phase
		if vs.Phase != VotePhaseLocationSelect {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Invalid voting phase")
			return
		}

		// check winner user id
		if vw.List[0] != userID {
			errChan <- terror.Error(terror.ErrForbidden)
			return
		}

		// broadcast notification
		go vc.API.BroadcastGameNotification(GameNotificationTypeText, fmt.Sprintf("User %s placed %s at (x: %d, y: %d)", hcd.Username, va.FactionAbilityMap[hcd.FactionID].Label, req.Payload.XIndex, req.Payload.YIndex))

		// record ability animation
		userIDString := userID.String()
		selectedX := req.Payload.XIndex
		selectedY := req.Payload.YIndex
		err = vc.API.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
			FactionID:           hcd.FactionID,
			FactionAbilityID:    va.FactionAbilityMap[hcd.FactionID].ID,
			IsSuccess:           true,
			GameClientAbilityID: va.FactionAbilityMap[hcd.FactionID].GameClientAbilityID,
			TriggeredByUserID:   &userIDString,
			TriggeredByUsername: &hcd.Username,
			TriggeredOnCellX:    &selectedX,
			TriggeredOnCellY:    &selectedY,
		})
		if err != nil {
			errChan <- terror.Error(err)
			return
		}

		// get random ability collection set
		battleAbility, factionAbilityMap, err := vc.API.BattleArena.RandomAbilityCollection()
		if err != nil {
			errChan <- terror.Error(err)
			return
		}

		vc.API.MessageBus.Send(messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

		// initialise new ability collection
		va.BattleAbility = battleAbility

		// initialise new faction ability map
		for fid, ability := range factionAbilityMap {
			va.FactionAbilityMap[fid] = ability
		}

		// broadcast next stage
		vc.API.votePhaseChecker.Phase = VotePhaseVoteCooldown
		vs.Phase = VotePhaseVoteCooldown
		vs.EndTime = time.Now().Add(CooldownInitialDurationSecond * time.Second)

		// broadcast current stage to faction users
		vc.API.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	vc.API.ClientPickedLocation(wsc)
	reply(true)

	return nil
}

/***************
* Subscription *
***************/

type GameNotificationType string

const (
	GameNotificationTypeText       GameNotificationType = "TEXT"
	GameNotificationTypeSecondVote GameNotificationType = "SECOND_VOTE"
)

type GameNotification struct {
	Type GameNotificationType `json:"type"`
	Data interface{}          `json:"data"`
}

const HubKeyGameNotification hub.HubCommandKey = "GAME:NOTIFICATION"

const HubKeyVoteWinnerAnnouncement hub.HubCommandKey = "VOTE:WINNER:ANNOUNCEMENT"

// WinnerAnnouncementSubscribeHandler subscribe on vote winner to pick location
func (vc *VoteControllerWS) WinnerAnnouncementSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	vc.API.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
		if vs.Phase == VotePhaseHold {
			return
		}
		reply(va.BattleAbility)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), nil
}

const HubKeyVoteStageUpdated hub.HubCommandKey = "VOTE:STAGE:UPDATED"

// VoteStageUpdateSubscribeHandler to subscribe on vote stage
func (vc *VoteControllerWS) VoteStageUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	vc.API.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
		reply(vs)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyVoteStageUpdated), nil
}

const HubKeyFactionWarMachineQueueUpdated hub.HubCommandKey = "FACTION:WAR:MACHINE:QUEUE:UPDATED"

// FactionWarMachineQueueUpdateSubscribeHandler
func (vc *VoteControllerWS) FactionWarMachineQueueUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hubClientDetail, err := vc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if battleQueue, ok := vc.API.BattleArena.BattleQueueMap[hubClientDetail.FactionID]; ok {
		battleQueue <- func(wmql *battle_arena.WarMachineQueuingList) {
			maxLength := 5
			if len(wmql.WarMachines) < maxLength {
				maxLength = len(wmql.WarMachines)
			}

			reply(wmql.WarMachines[:maxLength])
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, hubClientDetail.FactionID))

	return req.TransactionID, busKey, nil
}