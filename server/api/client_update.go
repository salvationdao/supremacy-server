package api

import (
	"context"
	"fmt"
	"net/http"
	"server"
	"time"

	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"

	"github.com/ninja-syndicate/hub"

	"github.com/gofrs/uuid"
)

const TickSecond = 3

type ClientAction string

const (
	ClientOnline             ClientAction = "Online"
	ClientOffline            ClientAction = "Offline"
	ClientVoted              ClientAction = "Voted"
	ClientPickedLocation     ClientAction = "PickedLocation"
	ClientBattleRewardUpdate ClientAction = "BattleRewardUpdate"
)

type BattleRewardType string

const (
	BattleRewardTypeFaction = "BattleFactionReward"
	BattleRewardTypeWinner  = "BattleWinnerReward"
	BattleRewardTypeKill    = "BattleKillReward"
)

type ClientUpdate struct {
	Client           *hub.Client
	Action           ClientAction
	BattleReward     *ClientBattleReward
	NoClientLeftChan chan bool
}

type ClientBattleReward struct {
	BattleID server.BattleID
	Rewards  []BattleRewardType
}

type ClientMultiplier struct {
	clients           map[*hub.Client]bool
	MultiplierActions map[string]*MultiplierAction
}

type MultiplierAction struct {
	MultiplierValue int
	Expiry          time.Time
}

func (api *API) ValidateAndSendActiveClients(clientMap map[server.UserID][]*MultiplierAction) (int, error) {
	return 0, nil
}

func (api *API) ClientListener() {
	// this map stores online clients and their multipliers
	// [userID][action-name]
	// to avoid floats we use % based multipliers with a base of 100
	clientMultiplierMap := make(map[server.UserID]*ClientMultiplier)
	tickle.MinDurationOverride = true
	// start ticker for watch to earn
	taskTickle := tickle.New("FactionID Channel Point Ticker", TickSecond, func() (int, error) {
		ctx := context.Background()
		userMap := make(map[int][]server.UserID)

		for uid, actionSlice := range clientMultiplierMap {
			userMultiplier := 0
			for key, actn := range actionSlice.MultiplierActions {
				if actn.Expiry.After(time.Now()) {
					userMultiplier += actn.MultiplierValue
				} else {
					delete(actionSlice.MultiplierActions, key)
				}
			}

			_, ok := userMap[userMultiplier]
			if !ok {
				userMap[userMultiplier] = []server.UserID{}
			}

			userMap[userMultiplier] = append(userMap[userMultiplier], uid)
		}

		api.Passport.SendTickerMessage(ctx, userMap)

		return http.StatusOK, nil
	})
	taskTickle.Log = log_helpers.NamedLogger(api.Log, "FactionID Channel Point Ticker")
	taskTickle.DisableLogging = true
	taskTickle.Start()

listenLoop:
	for {
		msg := <-api.onlineClientMap
		uid, err := uuid.FromString(msg.Client.Identifier())
		if uid.IsNil() {
			continue
		}

		userID := server.UserID(uid)
		if err != nil {
			api.Log.Err(err).Msg("unable to marshall client identifier as uuid")
		}

		switch msg.Action {
		case ClientOnline:
			if _, ok := clientMultiplierMap[userID]; !ok {
				clientMultiplierMap[userID] = &ClientMultiplier{
					clients:           make(map[*hub.Client]bool),
					MultiplierActions: make(map[string]*MultiplierAction),
				}

				clientMultiplierMap[userID].MultiplierActions[string(ClientOnline)] = &MultiplierAction{
					MultiplierValue: 100,
					Expiry:          time.Now().AddDate(1, 0, 0),
				}
			}

			clientMultiplierMap[userID].clients[msg.Client] = true

		case ClientVoted:
			// check for existing voting action, then bump the time if exists
			multiplier, ok := clientMultiplierMap[userID].MultiplierActions[string(ClientVoted)]
			if ok {
				multiplier.Expiry = time.Now().Add(time.Minute * 30)
				continue listenLoop
			}

			multiplier = &MultiplierAction{
				MultiplierValue: 50,
				Expiry:          time.Now().Add(time.Minute * 30),
			}

			clientMultiplierMap[userID].MultiplierActions[string(ClientVoted)] = multiplier

		case ClientPickedLocation:
			// check for existing voting action, then bump the time if exists
			multiplier, ok := clientMultiplierMap[userID].MultiplierActions[string(ClientPickedLocation)]
			if ok {
				multiplier.Expiry = time.Now().Add(time.Minute * 30)
				continue listenLoop
			}

			multiplier = &MultiplierAction{
				MultiplierValue: 50,
				Expiry:          time.Now().Add(time.Minute * 30),
			}

			clientMultiplierMap[userID].MultiplierActions[string(ClientPickedLocation)] = multiplier

		case ClientBattleRewardUpdate:
			for _, reward := range msg.BattleReward.Rewards {
				switch reward {
				case BattleRewardTypeFaction:
					clientMultiplierMap[userID].MultiplierActions[fmt.Sprintf("%s:%s", BattleRewardTypeFaction, msg.BattleReward.BattleID)] = &MultiplierAction{
						MultiplierValue: 75,
						Expiry:          time.Now().Add(time.Minute * 5),
					}
				case BattleRewardTypeWinner:
					clientMultiplierMap[userID].MultiplierActions[fmt.Sprintf("%s:%s", BattleRewardTypeWinner, msg.BattleReward.BattleID)] = &MultiplierAction{
						MultiplierValue: 225,
						Expiry:          time.Now().Add(time.Minute * 5),
					}
				case BattleRewardTypeKill:
					clientMultiplierMap[userID].MultiplierActions[fmt.Sprintf("%s:%s", BattleRewardTypeKill, msg.BattleReward.BattleID)] = &MultiplierAction{
						MultiplierValue: 150,
						Expiry:          time.Now().Add(time.Minute * 5),
					}
				}
			}

		case ClientOffline:
			clientMap, ok := clientMultiplierMap[userID]
			if !ok {
				api.Log.Err(err)
				msg.NoClientLeftChan <- true
				continue listenLoop
			}

			delete(clientMap.clients, msg.Client)

			if len(clientMap.clients) == 0 {
				delete(clientMultiplierMap, userID)
				msg.NoClientLeftChan <- true
				continue listenLoop
			}
			msg.NoClientLeftChan <- false

		default:
			api.Log.Err(fmt.Errorf("unknown client action")).Msgf("unknown client action: %s", msg.Action)
		}
	}
}

func (api *API) ClientOnline(c *hub.Client) {
	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientOnline,
	}
}

func (api *API) ClientOffline(c *hub.Client) bool {
	noClientLeftChan := make(chan bool)
	api.onlineClientMap <- &ClientUpdate{
		Client:           c,
		Action:           ClientOffline,
		NoClientLeftChan: noClientLeftChan,
	}
	isNoClient := <-noClientLeftChan
	return isNoClient
}

func (api *API) ClientVoted(c *hub.Client) {
	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientVoted,
	}
}

func (api *API) ClientPickedLocation(c *hub.Client) {
	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientPickedLocation,
	}
}

func (api *API) ClientBattleRewardUpdate(c *hub.Client, cbr *ClientBattleReward) {
	api.onlineClientMap <- &ClientUpdate{
		Client:       c,
		Action:       ClientBattleRewardUpdate,
		BattleReward: cbr,
	}
}
