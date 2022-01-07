package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle_arena"
	"strings"
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
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	ExtensionSecret []byte
}

// NewTwitchController creates the check hub
func NewTwitchController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, twitchExtensionSecret []byte) *TwitchControllerWS {
	twitchHub := &TwitchControllerWS{
		Conn:            conn,
		Log:             log_helpers.NamedLogger(log, "twitch_hub"),
		API:             api,
		ExtensionSecret: twitchExtensionSecret,
	}

	api.Command(HubKeyTwitchAuth, twitchHub.Authentication)
	api.SecureUserCommand(HubKeyTwitchFactionActionFirstVote, twitchHub.FactionActionFirstVote)
	api.SecureUserCommand(HubKeyTwitchFactionActionSecondVote, twitchHub.FactionActionSecondVote)
	api.SecureUserFactionCommand(HubKeyTwitchActionLocationSelect, twitchHub.ActionLocationSelect)

	// subscription
	api.SecureUserSubscribeCommand(HubKeyTwitchConnectPointUpdated, twitchHub.ConnectPointUpdateSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyTwitchVoteWinnerAnnouncement, twitchHub.VoteWinnerAnnouncementSubscribeHandler)

	api.SecureUserFactionSubscribeCommand(HubKeyTwitchFactionActionUpdated, twitchHub.FactionActionUpdateSubscribeHandler)
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

	claims, err := getClaimsFromTwitchToken(req.Payload.TwitchToken, th.ExtensionSecret)
	if err != nil {
		return terror.Error(err)
	}

	if strings.HasPrefix(claims.OpaqueUserID, "U") && claims.TwitchAccountID != "" {
		// TODO: get users' identity from passport
		//user := th.API.Passport.FakeUserLoginWithFaction(claims.UserID)
		user := th.API.Passport.FakeUserLoginWithoutFaction(claims.TwitchAccountID)

		// update client detail
		th.API.hubClientDetail[wsc] <- func(hcd *HubClientDetail) {
			hcd.ID = user.ID
			hcd.FactionID = user.FactionID
		}

		// remove client from default online client map
		th.API.onlineClientMap[server.UserID(uuid.Nil)] <- func(cim ClientInstanceMap, cps *ConnectPointState, t *tickle.Tickle) {
			if _, ok := cim[wsc]; ok {
				delete(cim, wsc)
			}
		}

		// add client to new online client map
		currentOnlineClientMap, ok := th.API.onlineClientMap[user.ID]
		if !ok {
			currentOnlineClientMap = make(chan func(ClientInstanceMap, *ConnectPointState, *tickle.Tickle))
			th.API.onlineClientMap[user.ID] = currentOnlineClientMap
			go th.API.startOnlineClientTracker(user.ID, user.ConnectPoint)
		}

		currentOnlineClientMap <- func(cim ClientInstanceMap, cps *ConnectPointState, t *tickle.Tickle) {
			// add instance
			if _, ok := cim[wsc]; !ok {
				cim[wsc] = true
			}
		}

		reply(user)
	}
	return nil
}

type TwitchActionVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionActionID server.FactionActionID `json:"factionActionID"`
		PointSpend      int                    `json:"pointSpend"`
	} `json:"payload"`
}

const HubKeyTwitchFactionActionFirstVote = hub.HubCommandKey("TWITCH:FACTION:ACTION:FIRST:VOTE")

func (th *TwitchControllerWS) FactionActionFirstVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		firstVoteAction, ok := fvs[req.Payload.FactionActionID]
		if !ok {
			errChan <- terror.Error(terror.ErrForbidden, "Error - Action not exists")
			return
		}

		// check channel point payment success
		isSuccessPaidChan := make(chan bool)

		// check current user's channel point is sufficient
		th.API.onlineClientMap[hubClientDetail.ID] <- func(cim ClientInstanceMap, cps *ConnectPointState, t *tickle.Tickle) {
			// reduce the fund
			if cps.ConnectPoint < int64(req.Payload.PointSpend) {
				isSuccessPaidChan <- false
				return
			}

			cps.ConnectPoint -= int64(req.Payload.PointSpend)

			// broadcast connect point update
			th.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchConnectPointUpdated, hubClientDetail.ID)), cps.ConnectPoint)

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
		fvs[req.Payload.FactionActionID] = firstVoteAction

		// if TIE check winner
		if vs.Phase == VotePhaseTie {

			parseFirstVoteResult(fvs, fvr)

			// if winner exists, enter second vote
			if !fvr.factionActionID.IsNil() && len(fvr.hubClientID) > 0 {
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
							Faction:       f,
							FactionAction: firstVoteAction.FactionAction,
							EndTime:       vs.EndTime,
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

const HubKeyTwitchFactionActionSecondVote hub.HubCommandKey = hub.HubCommandKey("TWITCH:FACTION:ACTION:SECOND:VOTE")

type TwitchActionSecondVote struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID       server.FactionID       `json:"factionID"`
		FactionActionID server.FactionActionID `json:"factionActionID"`
		IsAgreed        bool                   `json:"isAgreed"`
	} `json:"payload"`
}

func (th *TwitchControllerWS) FactionActionSecondVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

		if fvr.factionActionID != req.Payload.FactionActionID {
			errChan <- terror.Error(terror.ErrInvalidInput, "Error - Invalid action id")
			return
		}

		isSuccessPaidChan := make(chan bool)

		th.API.onlineClientMap[hubClientDetail.ID] <- func(cim ClientInstanceMap, cps *ConnectPointState, t *tickle.Tickle) {
			if cps.ConnectPoint <= 0 {
				isSuccessPaidChan <- false
				return
			}

			cps.ConnectPoint -= 1

			// broadcast connect point update
			th.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchConnectPointUpdated, hubClientDetail.ID)), cps.ConnectPoint)

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

		// signal action countered animation
		th.API.BattleArena.FactionActionTrigger(&battle_arena.ActionTriggerRequest{
			FactionID:       f.ID,
			FactionActionID: fvr.factionActionID,
			IsSuccess:       true,
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

const HubKeyTwitchConnectPointUpdated hub.HubCommandKey = hub.HubCommandKey("TWITCH:CONNECT:POINT:UPDATED")

// EvenUpdateSubscribeHandler to subscribe to game event
func (th *TwitchControllerWS) ConnectPointUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

	// return current channel point
	th.API.onlineClientMap[hubClientDetail.ID] <- func(cim ClientInstanceMap, cps *ConnectPointState, t *tickle.Tickle) {
		reply(cps.ConnectPoint)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchConnectPointUpdated, hubClientDetail.ID))

	return req.TransactionID, busKey, nil
}

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

const HubKeyTwitchFactionActionUpdated hub.HubCommandKey = hub.HubCommandKey("TWITCH:FACTION:ACTION:UPDATED")

// FactionActionUpdateSubscribeHandler to subscribe to game event
func (th *TwitchControllerWS) FactionActionUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
			actions := []*server.FactionAction{}
			for _, firstVoteAction := range fvs {
				actions = append(actions, firstVoteAction.FactionAction)
			}
			reply(actions)
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionActionUpdated, hubClientDetail.FactionID))

	return req.TransactionID, busKey, nil
}
