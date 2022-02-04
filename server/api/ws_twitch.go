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

// TwitchControllerWS holds handlers for checking server status
type TwitchControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewTwitchController creates the check hub
func NewTwitchController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *TwitchControllerWS {
	twitchHub := &TwitchControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "twitch_hub"),
		API:  api,
	}

	api.Command(HubKeyTwitchJWTAuth, twitchHub.JWTAuth)
	api.SecureUserFactionCommand(HubKeyTwitchFactionAbilityFirstVote, twitchHub.FactionAbilityFirstVote)
	api.SecureUserCommand(HubKeyTwitchFactionAbilitySecondVote, twitchHub.FactionAbilitySecondVote)
	api.SecureUserFactionCommand(HubKeyTwitchActionLocationSelect, twitchHub.ActionLocationSelect)

	// subscription
	api.SecureUserSubscribeCommand(HubKeyTwitchVoteWinnerAnnouncement, twitchHub.VoteWinnerAnnouncementSubscribeHandler)

	api.SecureUserFactionSubscribeCommand(HubKeyTwitchFactionAbilityUpdated, twitchHub.FactionAbilityUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyTwitchFactionVoteStageUpdated, twitchHub.FactionVoteStageUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyTwitchFactionWarMachineQueueUpdated, twitchHub.FactionWarMachineQueueUpdateSubscribeHandler)
	return twitchHub
}

// TwitchAuthRequest authenticate a twitch user
type TwitchAuthRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TwitchToken string `json:"twitchToken"`
	} `json:"payload"`
}

const HubKeyTwitchJWTAuth = hub.HubCommandKey("TWITCH:JWT:AUTH")

func (th *TwitchControllerWS) JWTAuth(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchAuthRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	th.API.ringCheckAuthChan <- func(rca RingCheckAuthMap) {
		rca[req.Payload.TwitchToken] = wsc
	}

	// distroy the token in 30 second
	go func() {
		time.Sleep(600 * time.Second)

		th.API.ringCheckAuthChan <- func(rca RingCheckAuthMap) {
			_, ok := rca[req.Payload.TwitchToken]
			if ok {
				delete(rca, req.Payload.TwitchToken)
			}
		}
	}()

	reply(true)

	return nil
}

type TwitchActionVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionAbilityID server.FactionAbilityID `json:"factionAbilityID"`
		PointSpend       server.BigInt           `json:"pointSpend"`
	} `json:"payload"`
}

const HubKeyTwitchFactionAbilityFirstVote = hub.HubCommandKey("TWITCH:FACTION:ABILITY:FIRST:VOTE")

func (th *TwitchControllerWS) FactionAbilityFirstVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchActionVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	if req.Payload.PointSpend.Cmp(big.NewInt(0)) <= 0 {
		return terror.Error(terror.ErrInvalidInput, "Error - Can not spend negative amount of sups")
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	if hubClientDetail.FactionID.IsNil() {
		return terror.Error(terror.ErrForbidden, "Error - Can only vote after joining one of the three factions")
	}

	// declare amount
	bigInt := big.NewInt(0)
	amount := server.BigInt{Int: *bigInt.Add(bigInt, &req.Payload.PointSpend.Int)}

	errChan := make(chan error)

	// check vote cycle of current user's faction
	th.API.factionVoteCycle[hubClientDetail.FactionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		if vs.Phase != VotePhaseFirstVote && vs.Phase != VotePhaseTie {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Invalid voting stage")
			return
		}

		// check action exists
		ability, ok := fvs[req.Payload.FactionAbilityID]
		if !ok {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Action not exists")
			return
		}

		reason := fmt.Sprintf("battle:%s|voteaction:%s", th.API.BattleArena.CurrentBattleID(), req.Payload.FactionAbilityID)
		supTransactionReference, err := th.API.Passport.SendHoldSupsMessage(context.Background(), userID, amount, req.TransactionID, reason)
		if err != nil {
			errChan <- terror.Error(err, "Error - Failed to spend sups")
			return
		}

		// store vote amount to live voting data
		th.API.liveVotingData[f.ID] <- func(lvd *LiveVotingData) {
			lvd.TotalVote.Add(&lvd.TotalVote.Int, &amount.Int)
		}

		// calc total vote count of the current sups spent
		voteCount := server.BigInt{Int: *req.Payload.PointSpend.Div(&req.Payload.PointSpend.Int, &ability.FactionAbility.SupsCost.Int)}

		// update vote result if it is in first vote phase
		if vs.Phase == VotePhaseFirstVote {
			_, ok = fvs[req.Payload.FactionAbilityID].UserVoteMap[userID]
			if !ok {
				fvs[req.Payload.FactionAbilityID].UserVoteMap[userID] = make(map[server.TransactionReference]server.BigInt)
			}

			fvs[req.Payload.FactionAbilityID].UserVoteMap[userID][supTransactionReference] = voteCount

			errChan <- nil
			return
		}

		// if TIE, directly set user as winner once the transaction is successful

		// commit the transactions and check status
		transactions, err := th.API.Passport.CommitTransactions(ctx, []server.TransactionReference{supTransactionReference})
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
		fvr.factionAbilityID = req.Payload.FactionAbilityID
		fvr.hubClientID = []server.UserID{userID}

		// update vote phase
		vs.Phase = VotePhaseSecondVote
		vs.EndTime = time.Now().Add(SecondVoteDurationSecond * time.Second)

		// broadcast current stage to current faction users
		th.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

		// broadcast second vote candidate to all the connected clients
		broadcastData, err := json.Marshal(&BroadcastPayload{
			Key: HubKeyTwitchNotification,
			Payload: &TwitchNotification{
				Type: TwitchNotificationTypeSecondVote,
				Data: &secondVoteCandidate{
					Faction:        f,
					FactionAbility: fvs[req.Payload.FactionAbilityID].FactionAbility,
					EndTime:        vs.EndTime,
				},
			},
		})
		if err == nil {
			th.API.Hub.Clients(func(clients hub.ClientsList) {
				for client, ok := range clients {
					if !ok {
						continue
					}
					go func(c *hub.Client) {
						err := c.Send(broadcastData)
						if err != nil {
							th.API.Log.Err(err).Msg("failed to send broadcast")
						}
					}(client)
				}
			})
		}

		// restart vote ticker
		if t.VotingStageListener.NextTick == nil {
			t.VotingStageListener.Start()
		}

		// start second vote broadcaster
		if t.SecondVoteResultBroadcaster.NextTick == nil {
			t.SecondVoteResultBroadcaster.Start()
		}

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	th.API.ClientVoted(wsc)

	reply(true)
	return nil
}

const HubKeyTwitchFactionAbilitySecondVote = hub.HubCommandKey("TWITCH:FACTION:ABILITY:SECOND:VOTE")

type TwitchActionSecondVote struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID        server.FactionID        `json:"factionID"`
		FactionAbilityID server.FactionAbilityID `json:"factionAbilityID"`
		IsAgreed         bool                    `json:"isAgreed"`
	} `json:"payload"`
}

func (th *TwitchControllerWS) FactionAbilitySecondVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchActionSecondVote{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	// get user detail
	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err, "Failed to read user detail")
	}

	// check faction exists
	factionVoteCycle, ok := th.API.factionVoteCycle[req.Payload.FactionID]
	if !ok {
		return terror.Error(terror.ErrInvalidInput, "Error - FactionID voting cycle not exists")
	}

	errChan := make(chan error)

	factionVoteCycle <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		if vs.Phase != VotePhaseSecondVote {
			errChan <- terror.Error(terror.ErrInvalidInput, "Error - Invalid voting phase")
			return
		}

		if fvr.factionAbilityID != req.Payload.FactionAbilityID {
			errChan <- terror.Error(terror.ErrInvalidInput, "Error - Invalid action id")
			return
		}

		// get ability basic sups cost as vote value
		amount := fvs[fvr.factionAbilityID].FactionAbility.SupsCost
		reason := fmt.Sprintf("battle:%s|voteaction:%s", th.API.BattleArena.CurrentBattleID(), req.Payload.FactionAbilityID)
		supTransactionReference, err := th.API.Passport.SendHoldSupsMessage(context.Background(), userID, amount, req.TransactionID, reason)
		if err != nil {
			th.API.Log.Err(err).Msg("failed to spend sups")
			return
		}

		// store vote amount to live voting data
		th.API.liveVotingData[f.ID] <- func(lvd *LiveVotingData) {
			lvd.TotalVote.Add(&lvd.TotalVote.Int, &amount.Int)
		}

		// add vote to result
		if req.Payload.IsAgreed {
			svs.AgreedCount = append(svs.AgreedCount, supTransactionReference)
		} else {
			svs.DisagreedCount = append(svs.DisagreedCount, supTransactionReference)
		}

		// vote weight

		// basic vote weight
		voteWage := int64(5)

		// get another 5 point if user join a faction
		if !hubClientDetail.FactionID.IsNil() {
			voteWage += 10
		}

		// get faction bonus
		voteWage += th.API.GetFactionVoteWage(hubClientDetail.FactionID)

		// add vote wage to the list
		svs.VoteValueMap[supTransactionReference] = voteWage

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	reply(true)

	return nil
}

const HubKeyTwitchActionLocationSelect = hub.HubCommandKey("TWITCH:ACTION:LOCATION:SELECT")

type TwitchLocationSelect struct {
	*hub.HubCommandRequest
	Payload struct {
		XIndex int `json:"x"`
		YIndex int `json:"y"`
	} `json:"payload"`
}

func (th *TwitchControllerWS) ActionLocationSelect(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchLocationSelect{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	if _, ok := th.API.factionVoteCycle[hubClientDetail.FactionID]; !ok {
		return terror.Error(terror.ErrInvalidInput, "Currently no action cycle")
	}

	errChan := make(chan error)
	th.API.factionVoteCycle[hubClientDetail.FactionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		if vs.Phase != VotePhaseLocationSelect {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Invalid voting phase")
			return
		}

		if fvr.hubClientID[0] != userID {
			errChan <- terror.Error(terror.ErrForbidden)
			return
		}

		// broadcast notification to all the connected clients
		broadcastData, err := json.Marshal(&BroadcastPayload{
			Key: HubKeyTwitchNotification,
			Payload: &TwitchNotification{
				Type: TwitchNotificationTypeText,
				Data: fmt.Sprintf("User %s placed %s at (x: %d, y: %d)", hubClientDetail.Username, fvs[fvr.factionAbilityID].FactionAbility.Label, req.Payload.XIndex, req.Payload.YIndex),
			},
		})
		if err == nil {
			th.API.Hub.Clients(func(clients hub.ClientsList) {
				for client, ok := range clients {
					if !ok {
						continue
					}
					go func(c *hub.Client) {
						err := c.Send(broadcastData)
						if err != nil {
							th.API.Log.Err(err).Msg("failed to send broadcast")
						}
					}(client)
				}
			})
		}

		if t.SecondVoteResultBroadcaster.NextTick != nil {
			t.SecondVoteResultBroadcaster.Stop()
		}

		// broadcast current stage to current faction users
		th.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

		userDetails, err := th.API.getClientDetailFromChannel(wsc)
		if err != nil {
			errChan <- terror.Error(err)
			return
		}

		userIDString := userID.String()
		selectedX := req.Payload.XIndex
		selectedY := req.Payload.YIndex

		// signal ability animation
		err = th.API.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
			FactionID:           f.ID,
			FactionAbilityID:    fvr.factionAbilityID,
			IsSuccess:           true,
			GameClientAbilityID: fvs[fvr.factionAbilityID].FactionAbility.GameClientAbilityID,
			TriggeredByUserID:   &userIDString,
			TriggeredByUsername: &userDetails.Username,
			TriggeredOnCellX:    &selectedX,
			TriggeredOnCellY:    &selectedY,
		})
		if err != nil {
			errChan <- terror.Error(err)
			return
		}

		// broadcast next stage
		vs.Phase = VotePhaseVoteCooldown
		vs.EndTime = time.Now().Add(time.Duration(fvs[fvr.factionAbilityID].FactionAbility.CooldownDurationSecond) * time.Second)

		// broadcast current stage to current faction users
		th.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	reply(true)
	th.API.ClientPickedLocation(wsc)
	return nil
}

/***************
* Subscription *
***************/

type TwitchNotificationType string

const (
	TwitchNotificationTypeText       TwitchNotificationType = "TEXT"
	TwitchNotificationTypeSecondVote TwitchNotificationType = "SECOND_VOTE"
)

type TwitchNotification struct {
	Type TwitchNotificationType `json:"type"`
	Data interface{}            `json:"data"`
}

const HubKeyTwitchNotification hub.HubCommandKey = "TWITCH:NOTIFICATION"
const HubKeyTwitchFactionSecondVoteUpdated hub.HubCommandKey = "TWITCH:FACTION:SECOND:VOTE:UPDATED"
const HubKeyTwitchVoteWinnerAnnouncement hub.HubCommandKey = "TWITCH:VOTE:WINNER:ANNOUNCEMENT"

func (th *TwitchControllerWS) VoteWinnerAnnouncementSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchVoteWinnerAnnouncement, userID))

	return req.TransactionID, busKey, nil
}

const HubKeyTwitchFactionVoteStageUpdated hub.HubCommandKey = "TWITCH:FACTION:VOTE:STAGE:UPDATED"

// FactionVoteStageUpdateSubscribeHandler to subscribe to game event
func (th *TwitchControllerWS) FactionVoteStageUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if voteCycle, ok := th.API.factionVoteCycle[hubClientDetail.FactionID]; ok {
		voteCycle <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
			reply(vs)
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, hubClientDetail.FactionID))
	return req.TransactionID, busKey, nil
}

const HubKeyTwitchFactionWarMachineQueueUpdated hub.HubCommandKey = "TWITCH:FACTION:WAR:MACHINE:QUEUE:UPDATED"

func (th *TwitchControllerWS) FactionWarMachineQueueUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if battleQueue, ok := th.API.BattleArena.BattleQueueMap[hubClientDetail.FactionID]; ok {
		battleQueue <- func(wmql *battle_arena.WarMachineQueuingList) {
			maxLength := 5
			if len(wmql.WarMachines) < maxLength {
				maxLength = len(wmql.WarMachines)
			}

			reply(wmql.WarMachines[:maxLength])
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionWarMachineQueueUpdated, hubClientDetail.FactionID))

	return req.TransactionID, busKey, nil
}

const HubKeyTwitchLiveVotingDataUpdated hub.HubCommandKey = "TWITCH:LIVE:VOTING:DATA:UPDATED"

const HubKeyTwitchFactionAbilityUpdated = hub.HubCommandKey("TWITCH:FACTION:ABILITY:UPDATED")

// FactionAbilityUpdateSubscribeHandler to subscribe to game event
func (th *TwitchControllerWS) FactionAbilityUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if voteCycle, ok := th.API.factionVoteCycle[hubClientDetail.FactionID]; ok {
		voteCycle <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
			var abilities []*server.FactionAbility
			for _, firstVoteAction := range fvs {
				abilities = append(abilities, firstVoteAction.FactionAbility)
			}
			reply(abilities)
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionAbilityUpdated, hubClientDetail.FactionID))

	return req.TransactionID, busKey, nil
}
