package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle_arena"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
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

	api.Command(HubKeyTwitchAuth, twitchHub.Authentication)
	api.SecureUserCommand(HubKeyTwitchFactionAbilityFirstVote, twitchHub.FactionAbilityFirstVote)
	api.SecureUserCommand(HubKeyTwitchFactionAbilitySecondVote, twitchHub.FactionAbilitySecondVote)
	api.SecureUserFactionCommand(HubKeyTwitchActionLocationSelect, twitchHub.ActionLocationSelect)

	// subscription
	api.SecureUserSubscribeCommand(HubKeyTwitchVoteWinnerAnnouncement, twitchHub.VoteWinnerAnnouncementSubscribeHandler)

	api.SecureUserFactionSubscribeCommand(HubKeyTwitchFactionAbilityUpdated, twitchHub.FactionAbilityUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyTwitchFactionVoteStageUpdated, twitchHub.FactionVoteStageUpdateSubscribeHandler)
	return twitchHub
}

// getClaimsFromTwitchToken verifies token from Twitch
func getClaimsFromTwitchToken(token string, jwtSecret []byte) (*TwitchJWTClaims, error) {
	// Get claims
	claims := &TwitchJWTClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, terror.Error(terror.ErrBadClaims, "Invalid token")
	}

	return claims, nil
}

// TwitchJWTClaims is the payload of a JWT sent by the Twitch extension
type TwitchJWTClaims struct {
	OpaqueUserID    string `json:"opaque_user_id,omitempty"`
	TwitchAccountID string `json:"user_id"`
	ChannelID       string `json:"channel_id,omitempty"`
	Role            string `json:"role"`
	jwt.StandardClaims
}

const HubKeyTwitchAuth = hub.HubCommandKey("TWITCH:AUTH")

// TwitchAuthRequest authenticate a twitch user
type TwitchAuthRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TwitchToken string `json:"twitchToken"`
	} `json:"payload"`
}

type UserInfo struct {
	UserID string `json:"userID"`
}

type TwitchAuthResponse struct {
	UserInfo *UserInfo `json:"userInfo"`
}

func (th *TwitchControllerWS) Authentication(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchAuthRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	user, err := th.API.Passport.TwitchAuth(ctx, req.Payload.TwitchToken, req.TransactionID)
	if err != nil {
		return terror.Error(err, "Unable to load user")
	}

	// update client detail
	th.API.hubClientDetail[wsc] <- func(hcd *HubClientDetail) {
		hcd.ID = user.ID
		hcd.FactionID = user.FactionID
	}

	// remove client from default online client map
	th.API.onlineClientMap[server.UserID(uuid.Nil)] <- func(cim ClientInstanceMap, t *tickle.Tickle) {
		if _, ok := cim[wsc]; ok {
			delete(cim, wsc)
		}
	}

	// add client to new online client map
	currentOnlineClientMap, ok := th.API.onlineClientMap[user.ID]
	if !ok {
		currentOnlineClientMap = make(chan func(ClientInstanceMap, *tickle.Tickle))
		th.API.onlineClientMap[user.ID] = currentOnlineClientMap
		go th.API.startOnlineClientTracker(user.ID)
	}

	currentOnlineClientMap <- func(cim ClientInstanceMap, t *tickle.Tickle) {
		// add instance
		if _, ok := cim[wsc]; !ok {
			cim[wsc] = true
		}
	}

	reply(user)

	return nil
}

type TwitchActionVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionAbilityID server.FactionAbilityID `json:"factionAbilityID"`
		PointSpend       int                     `json:"pointSpend"`
	} `json:"payload"`
}

const HubKeyTwitchFactionAbilityFirstVote = hub.HubCommandKey("TWITCH:FACTION:ABILITY:FIRST:VOTE")

func (th *TwitchControllerWS) FactionAbilityFirstVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchActionVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	if hubClientDetail.FactionID.IsNil() {
		return terror.Error(terror.ErrForbidden, "Error - Con only vote after joining one of the three factions")
	}

	errChan := make(chan error)

	// check vote cycle of current user's faction
	th.API.factionVoteCycle[hubClientDetail.FactionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		if vs.Phase != VotePhaseFirstVote && vs.Phase != VotePhaseTie {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Invalid voting stage")
			return
		}

		// check action exists
		firstVoteAction, ok := fvs[req.Payload.FactionAbilityID]
		if !ok {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Action not exists")
			return
		}

		// check channel point payment success
		isSuccessPaidChan := make(chan bool)

		// check current user's channel point is sufficient
		th.API.onlineClientMap[hubClientDetail.ID] <- func(cim ClientInstanceMap, t *tickle.Tickle) {
			isSuccess, err := th.API.Passport.UserSupsUpdate(context.Background(), hubClientDetail.ID, int64(-req.Payload.PointSpend), "test")
			if err != nil {
				th.API.Log.Err(err).Msg("failed to spend sups")
				isSuccessPaidChan <- false
				return
			}

			if !isSuccess {
				th.API.Log.Err(err).Msg("failed to spend sups")
				isSuccessPaidChan <- false
				return
			}

			isSuccessPaidChan <- true
		}

		// if not success terminate the function
		isSuccessPaid := <-isSuccessPaidChan
		if !isSuccessPaid {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Insufficient channel point")
			return
		}

		// update vote result
		currentClientVote, ok := firstVoteAction.UserVoteMap[hubClientDetail.ID]
		if !ok {
			currentClientVote = 0
		}

		currentClientVote += int64(req.Payload.PointSpend)
		firstVoteAction.UserVoteMap[hubClientDetail.ID] = currentClientVote
		fvs[req.Payload.FactionAbilityID] = firstVoteAction

		// if TIE check winner
		if vs.Phase == VotePhaseTie {

			parseFirstVoteResult(fvs, fvr)

			// if winner exists, enter second vote
			if !fvr.factionAbilityID.IsNil() && len(fvr.hubClientID) > 0 {
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
							FactionAbility: firstVoteAction.FactionAbility,
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
							go client.Send(broadcastData)
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
			}
		}

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	reply(true)

	return nil
}

const HubKeyTwitchFactionAbilitySecondVote hub.HubCommandKey = hub.HubCommandKey("TWITCH:FACTION:ABILITY:SECOND:VOTE")

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

	hubClientDetail, err := th.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
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

		isSuccessPaidChan := make(chan bool)

		th.API.onlineClientMap[hubClientDetail.ID] <- func(cim ClientInstanceMap, t *tickle.Tickle) {
			isSuccess, err := th.API.Passport.UserSupsUpdate(context.Background(), hubClientDetail.ID, -1, "test")
			if err != nil {
				th.API.Log.Err(err).Msg("failed to spend sups")
				isSuccessPaidChan <- false
				return
			}

			if !isSuccess {
				th.API.Log.Err(err).Msg("failed to spend sups")
				isSuccessPaidChan <- false
				return
			}

			isSuccessPaidChan <- true
		}

		isSuccessPaid := <-isSuccessPaidChan
		if !isSuccessPaid {
			errChan <- terror.Error(terror.ErrInvalidInput, "Error - Insufficient connect point")
			return
		}

		// add up to result
		if req.Payload.IsAgreed {
			svs.AgreedCount += 1
		} else {
			svs.DisagreedCount += 1
		}

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	reply(true)

	return nil
}

const HubKeyTwitchActionLocationSelect hub.HubCommandKey = hub.HubCommandKey("TWITCH:ACTION:LOCATION:SELECT")

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

		if fvr.hubClientID[0] != hubClientDetail.ID {
			errChan <- terror.Error(terror.ErrForbidden)
			return
		}

		// broadcast notification to all the connected clients
		broadcastData, err := json.Marshal(&BroadcastPayload{
			Key: HubKeyTwitchNotification,
			Payload: &TwitchNotification{
				Type: TwitchNotificationTypeText,
				Data: fmt.Sprintf("User %s select x: %d, y: %d", hubClientDetail.ID, req.Payload.XIndex, req.Payload.YIndex),
			},
		})
		if err == nil {
			th.API.Hub.Clients(func(clients hub.ClientsList) {
				for client, ok := range clients {
					if !ok {
						continue
					}
					go client.Send(broadcastData)
				}
			})
		}

		// pause the whole voting cycle, wait until animation finish
		vs.Phase = VotePhaseHold
		if t.VotingStageListener.NextTick != nil {
			t.VotingStageListener.Stop()
		}

		if t.SecondVoteResultBroadcaster.NextTick != nil {
			t.SecondVoteResultBroadcaster.Stop()
		}

		// broadcast current stage to current faction users
		th.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

		userName := hubClientDetail.ID.String()
		selectedX := req.Payload.XIndex
		selectedY := req.Payload.YIndex

		// signal ability animation
		th.API.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
			FactionID:        f.ID,
			FactionAbilityID: fvr.factionAbilityID,
			IsSuccess:        true,
			TriggeredByUser:  &userName,
			TriggeredOnCellX: &selectedX,
			TriggeredOnCellY: &selectedY,
		})

		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	reply(true)

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

const HubKeyTwitchNotification hub.HubCommandKey = hub.HubCommandKey("TWITCH:NOTIFICATION")

const HubKeyTwitchFactionSecondVoteUpdated hub.HubCommandKey = hub.HubCommandKey("TWITCH:FACTION:SECOND:VOTE:UPDATED")

const HubKeyTwitchVoteWinnerAnnouncement hub.HubCommandKey = hub.HubCommandKey("TWITCH:VOTE:WINNER:ANNOUNCEMENT")

// EvenUpdateSubscribeHandler to subscribe to game event
func (th *TwitchControllerWS) VoteWinnerAnnouncementSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchVoteWinnerAnnouncement, hubClientDetail.ID))

	return req.TransactionID, busKey, nil
}

const HubKeyTwitchFactionVoteStageUpdated hub.HubCommandKey = hub.HubCommandKey("TWITCH:FACTION:VOTE:STAGE:UPDATED")

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

const HubKeyTwitchFactionAbilityUpdated hub.HubCommandKey = hub.HubCommandKey("TWITCH:FACTION:ABILITY:UPDATED")

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
			abilities := []*server.FactionAbility{}
			for _, firstVoteAction := range fvs {
				abilities = append(abilities, firstVoteAction.FactionAbility)
			}
			reply(abilities)
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionAbilityUpdated, hubClientDetail.FactionID))

	return req.TransactionID, busKey, nil
}
