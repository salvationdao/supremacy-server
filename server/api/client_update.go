package api

import (
	"context"
	"fmt"
	"server"
	"time"

	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"

	"github.com/ninja-syndicate/hub"

	"github.com/gofrs/uuid"
)

const TickTime = 2

type ClientAction string

const (
	ClientOnline         ClientAction = "Online"
	ClientOffline        ClientAction = "Offline"
	ClientVoted          ClientAction = "Voted"
	ClientPickedLocation ClientAction = "PickedLocation"
)

type ClientUpdate struct {
	Client *hub.Client
	Action ClientAction
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
	clientMultiplierMap := make(map[server.UserID][]*MultiplierAction)
	tickle.MinDurationOverride = true
	// start ticker for watch to earn
	taskTickle := tickle.New("FactionID Channel Point Ticker", TickTime, func() (int, error) {
		ctx := context.Background()
		userMap := make(map[int][]server.UserID)

		for uid, actionSlice := range clientMultiplierMap {
			userMultiplier := 0
			for _, actn := range actionSlice {
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

		_, err := api.Passport.SendTickerMessage(ctx, userMap)
		if err != nil {
			return 0, terror.Error(err)
		}

		return 0, nil
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
				clientMultiplierMap[userID] = []*MultiplierAction{{
					Name:            ClientOnline,
					MultiplierValue: 100,
					Expiry:          time.Now().AddDate(1, 0, 0),
				}}
			}

		case ClientVoted:
			if _, ok := clientMultiplierMap[userID]; !ok { // if not okay add online one
				clientMultiplierMap[userID] = []*MultiplierAction{{
					Name:            ClientOnline,
					MultiplierValue: 100,
					Expiry:          time.Now().AddDate(1, 0, 0),
				}}
			}

			// check for existing voting action, then bump the time if exists
			for _, multipler := range clientMultiplierMap[userID] {
				if multipler.Name == ClientVoted {
					multipler.Expiry = time.Now().Add(time.Minute * 30)
					continue listenLoop
				}
			}

			// if voting actions didn't exist, add it
			clientMultiplierMap[userID] = append(clientMultiplierMap[userID], &MultiplierAction{
				Name:            ClientVoted,
				MultiplierValue: 50,
				Expiry:          time.Now().Add(time.Minute * 30),
			})
		case ClientPickedLocation:
			if _, ok := clientMultiplierMap[userID]; !ok { // if not okay add online one
				clientMultiplierMap[userID] = []*MultiplierAction{{
					Name:            ClientOnline,
					MultiplierValue: 100,
					Expiry:          time.Now().AddDate(1, 0, 0),
				}}
			}

			// check for existing voting action, then bump the time if exists
			for _, multiplier := range clientMultiplierMap[userID] {
				if multiplier.Name == ClientPickedLocation {
					multiplier.Expiry = time.Now().Add(time.Minute * 30)
					continue listenLoop
				}
			}

			// if voting actions didn't exist, add it
			clientMultiplierMap[userID] = append(clientMultiplierMap[userID], &MultiplierAction{
				Name:            ClientPickedLocation,
				MultiplierValue: 50,
				Expiry:          time.Now().Add(time.Minute * 30),
			})
		case ClientOffline:
			delete(clientMultiplierMap, userID)
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
