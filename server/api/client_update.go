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

const TickTime = 3

type ClientAction string

const (
	ClientOnline          ClientAction = "Online"
	ClientOffline         ClientAction = "Offline"
	ClientVoted           ClientAction = "Voted"
	ClientPickedLocation  ClientAction = "PickedLocation"
	ClientMultiplierValue ClientAction = "MultiplierValue"
)

type ClientUpdate struct {
	Client            *hub.Client
	Action            ClientAction
	MultipleValueChan chan int
}

type ClientMultiplier struct {
	clients           map[*hub.Client]bool
	MultiplierActions []*MultiplierAction
}

type MultiplierAction struct {
	Name            ClientAction
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
	taskTickle := tickle.New("FactionID Channel Point Ticker", TickTime, func() (int, error) {
		ctx := context.Background()
		userMap := make(map[int][]server.UserID)

		for uid, actionSlice := range clientMultiplierMap {
			userMultiplier := 0
			for _, actn := range actionSlice.MultiplierActions {
				if actn.Expiry.After(time.Now()) {
					userMultiplier += actn.MultiplierValue
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
					clients: make(map[*hub.Client]bool),
					MultiplierActions: []*MultiplierAction{{
						Name:            ClientOnline,
						MultiplierValue: 100,
						Expiry:          time.Now().AddDate(1, 0, 0),
					}},
				}
			}

			clientMultiplierMap[userID].clients[msg.Client] = true

		case ClientVoted:
			// check for existing voting action, then bump the time if exists
			for _, multipler := range clientMultiplierMap[userID].MultiplierActions {
				if multipler.Name == ClientVoted {
					multipler.Expiry = time.Now().Add(time.Minute * 30)
					continue listenLoop
				}
			}

			// if voting actions didn't exist, add it
			clientMultiplierMap[userID].MultiplierActions = append(clientMultiplierMap[userID].MultiplierActions, &MultiplierAction{
				Name:            ClientVoted,
				MultiplierValue: 50,
				Expiry:          time.Now().Add(time.Minute * 30),
			})
		case ClientPickedLocation:
			// check for existing voting action, then bump the time if exists
			for _, multiplier := range clientMultiplierMap[userID].MultiplierActions {
				if multiplier.Name == ClientPickedLocation {
					multiplier.Expiry = time.Now().Add(time.Minute * 30)
					continue listenLoop
				}
			}

			// if voting actions didn't exist, add it
			clientMultiplierMap[userID].MultiplierActions = append(clientMultiplierMap[userID].MultiplierActions, &MultiplierAction{
				Name:            ClientPickedLocation,
				MultiplierValue: 50,
				Expiry:          time.Now().Add(time.Minute * 30),
			})

		case ClientMultiplierValue:
			userMultiplier := 0
			for _, actn := range clientMultiplierMap[userID].MultiplierActions {
				if actn.Expiry.After(time.Now()) {
					userMultiplier += actn.MultiplierValue
				}
			}
			msg.MultipleValueChan <- userMultiplier

		case ClientOffline:
			clientMap, ok := clientMultiplierMap[userID]
			if !ok {
				api.Log.Err(err)
				continue listenLoop
			}

			delete(clientMap.clients, msg.Client)

			if len(clientMap.clients) == 0 {
				delete(clientMultiplierMap, userID)
			}

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

func (api *API) ClientOffline(c *hub.Client) {
	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientOffline,
	}
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

func (api *API) ClientMultiplierValueGet(c *hub.Client) int {
	multiValueChan := make(chan int, 5)
	api.onlineClientMap <- &ClientUpdate{
		Client:            c,
		Action:            ClientMultiplierValue,
		MultipleValueChan: multiValueChan,
	}
	return <-multiValueChan
}
