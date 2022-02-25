package api

import (
	"context"
	"fmt"
	"net/http"
	"server"
	"server/passport"
	"time"

	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/rs/zerolog"

	"github.com/ninja-syndicate/hub"

	"github.com/gofrs/uuid"
)

const TickSecond = 5

type ClientAction string

const (
	ClientOnline                ClientAction = "Online"
	ClientOffline               ClientAction = "Offline"
	ClientVoted                 ClientAction = "Applause"
	ClientPickedLocation        ClientAction = "Picked Location"
	ClientBattleRewardUpdate    ClientAction = "BattleRewardUpdate"
	ClientSupsMultiplierGet     ClientAction = "SupsMultiplierGet"
	ClientCheckMultiplierUpdate ClientAction = "CheckMultiplierUpdate"
	ClientSupsTick              ClientAction = "SupsTick"
)

type BattleRewardType string

const (
	BattleRewardTypeFaction         BattleRewardType = "Battle Faction Reward"
	BattleRewardTypeWinner          BattleRewardType = "Battle Winner Reward"
	BattleRewardTypeKill            BattleRewardType = "Battle Kill Reward"
	BattleRewardTypeAbilityExecutor BattleRewardType = "Ability Executor"
	BattleRewardTypeInfluencer      BattleRewardType = "Battle Influencer"
	BattleRewardTypeWarContributor  BattleRewardType = "War Contributor"
)

type ClientUpdate struct {
	UserID           server.UserID
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
	supsTickerLogger := log_helpers.NamedLogger(api.Log, "Sups Ticker").Level(zerolog.Disabled)
	supsTicker := tickle.New("Sups Ticker", TickSecond, func() (int, error) {
		api.onlineClientMap <- &ClientUpdate{
			UserID: server.UserID(uuid.Must(uuid.NewV4())), // HACK: to pass the user id check
			Action: ClientSupsTick,
		}

		return http.StatusOK, nil
	})
	supsTicker.Log = &supsTickerLogger
	supsTicker.Start()

	// send multiplier changes every second to passport server
	cachedUserMultiplierAction := make(map[server.UserID]map[string]*MultiplierAction)

	supsMultiplierCheckerLogger := log_helpers.NamedLogger(api.Log, "Sups Multiplier Checker").Level(zerolog.Disabled)
	supsMultiplierChecker := tickle.New("Sups Multiplier Checker", 1, func() (int, error) {

		api.onlineClientMap <- &ClientUpdate{
			UserID: server.UserID(uuid.Must(uuid.NewV4())), // HACK: to pass the user id check
			Action: ClientCheckMultiplierUpdate,
		}

		return http.StatusOK, nil
	})
	supsMultiplierChecker.Log = &supsMultiplierCheckerLogger
	supsMultiplierChecker.Start()

listenLoop:
	for {
		msg := <-api.onlineClientMap
		userID := msg.UserID
		// if user id is nil, check id from client
		if userID.IsNil() {
			if msg.Client == nil {
				continue
			}

			uid := uuid.FromStringOrNil(msg.Client.Identifier())
			if uid.IsNil() {
				continue
			}
			userID = server.UserID(uid)
		}

		switch msg.Action {
		//done
		case ClientOnline:
			if _, ok := clientMultiplierMap[userID]; !ok {
				clientMultiplierMap[userID] = &ClientMultiplier{
					clients:           make(map[*hub.Client]bool),
					MultiplierActions: make(map[string]*MultiplierAction),
				}

				now := time.Now()

				// check client is in the old list
				if mas, ok := cachedUserMultiplierAction[userID]; ok {
					for key, ma := range mas {
						// if multiplier in old map is expired, delete and skip it
						if ma.Expiry.After(now) {
							// add non-expired multipliers to client
							clientMultiplierMap[userID].MultiplierActions[key] = &MultiplierAction{
								MultiplierValue: ma.MultiplierValue,
								Expiry:          ma.Expiry,
							}
						}

						// delete cache to trigger multiplier update
						delete(mas, key)
					}
				}

				// update online client to correct time
				clientMultiplierMap[userID].MultiplierActions[string(ClientOnline)] = &MultiplierAction{
					MultiplierValue: 100,
					Expiry:          now.AddDate(1, 0, 0),
				}

			}

			clientMultiplierMap[userID].clients[msg.Client] = true

			// done
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

			// done
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

			// done
		case ClientBattleRewardUpdate:
			for _, reward := range msg.BattleReward.Rewards {
				switch reward {
				case BattleRewardTypeFaction:
					clientMultiplierMap[userID].MultiplierActions[fmt.Sprintf("%s_%s", BattleRewardTypeFaction, msg.BattleReward.BattleID)] = &MultiplierAction{
						MultiplierValue: 1000,
						Expiry:          time.Now().Add(time.Minute * 5),
					}
				case BattleRewardTypeWinner:
					clientMultiplierMap[userID].MultiplierActions[fmt.Sprintf("%s_%s", BattleRewardTypeWinner, msg.BattleReward.BattleID)] = &MultiplierAction{
						MultiplierValue: 500,
						Expiry:          time.Now().Add(time.Minute * 5),
					}
				case BattleRewardTypeKill:
					clientMultiplierMap[userID].MultiplierActions[fmt.Sprintf("%s_%s", BattleRewardTypeKill, msg.BattleReward.BattleID)] = &MultiplierAction{
						MultiplierValue: 500,
						Expiry:          time.Now().Add(time.Minute * 5),
					}
				}
			}

			// done
		case ClientSupsTick:
			userMap := make(map[int][]server.UserID)
			now := time.Now()

			for uid, actionSlice := range clientMultiplierMap {
				userMultiplier := 0
				for key, actn := range actionSlice.MultiplierActions {
					if actn.Expiry.After(now) {
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

			api.Passport.SendTickerMessage(userMap)

		case ClientSupsMultiplierGet:
			clientMap, ok := clientMultiplierMap[userID]
			if !ok {
				go api.UserSupsMultiplierToPassport(userID, nil)
				continue listenLoop
			}

			// send user's sups multipliers to passport server
			go api.UserSupsMultiplierToPassport(userID, clientMap.MultiplierActions)

		case ClientCheckMultiplierUpdate:
			api.clientMapUpdatedChecker(clientMultiplierMap, cachedUserMultiplierAction)

		case ClientOffline:
			clientMap, ok := clientMultiplierMap[userID]
			if !ok {
				api.Log.Err(fmt.Errorf("client not exists"))
				msg.NoClientLeftChan <- true
				continue listenLoop
			}

			delete(clientMap.clients, msg.Client)

			if len(clientMap.clients) == 0 {
				delete(clientMultiplierMap, userID)
				msg.NoClientLeftChan <- true

				// send user's sups multipliers to passport server
				go api.UserSupsMultiplierToPassport(userID, nil)
				continue listenLoop
			}
			msg.NoClientLeftChan <- false
		default:
			api.Log.Err(fmt.Errorf("unknown client action")).Msgf("unknown client action: %s", msg.Action)
		}
	}
}

func (api *API) ClientOnline(c *hub.Client) {
	// skip, if user not login
	if c.Identifier() == "" {
		return
	}

	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientOnline,
	}
}

func (api *API) ClientOffline(c *hub.Client) bool {
	// skip, if user not login
	if c.Identifier() == "" {
		return false
	}

	// otherwise erase the client from
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
	// skip, if user not login
	if c.Identifier() == "" {
		return
	}

	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientVoted,
	}
}

func (api *API) ClientPickedLocation(c *hub.Client) {
	// skip, if user not login
	if c.Identifier() == "" {
		return
	}

	api.onlineClientMap <- &ClientUpdate{
		Client: c,
		Action: ClientPickedLocation,
	}
}

func (api *API) ClientBattleRewardUpdate(c *hub.Client, cbr *ClientBattleReward) {
	// skip, if user not login
	if c.Identifier() == "" {
		return
	}

	api.onlineClientMap <- &ClientUpdate{
		Client:       c,
		Action:       ClientBattleRewardUpdate,
		BattleReward: cbr,
	}
}

func (api *API) ClientSupsMultipliersGet(userID server.UserID) {
	api.onlineClientMap <- &ClientUpdate{
		UserID: userID,
		Action: ClientSupsMultiplierGet,
	}
}

func (api *API) UserSupsMultiplierToPassport(userID server.UserID, supsMultiplierMap map[string]*MultiplierAction) {
	userSupsMultiplierSend := &passport.UserSupsMultiplierSend{
		ToUserID:        userID,
		SupsMultipliers: []*passport.SupsMultiplier{},
	}

	for key, sm := range supsMultiplierMap {
		userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &passport.SupsMultiplier{
			Key:       key,
			Value:     sm.MultiplierValue,
			ExpiredAt: sm.Expiry,
		})
	}

	go api.Passport.UserSupsMultiplierSend(context.Background(), []*passport.UserSupsMultiplierSend{userSupsMultiplierSend})
}

func (api *API) clientMapUpdatedChecker(newClientMap map[server.UserID]*ClientMultiplier, oldClientMap map[server.UserID]map[string]*MultiplierAction) {
	// send different
	sendDiff := []*passport.UserSupsMultiplierSend{}

	// loop through new user multiplier
	for userID, newMultiplier := range newClientMap {

		// prepare broadcast data to send
		userSupsMultiplierSend := &passport.UserSupsMultiplierSend{
			ToUserID:        userID,
			SupsMultipliers: []*passport.SupsMultiplier{},
		}

		// find old client map with user id
		oldMultiplier, ok := oldClientMap[userID]
		if !ok {
			// create a new copy of the new multiplier map
			oldClientMap[userID] = make(map[string]*MultiplierAction)

			for key, nm := range newMultiplier.MultiplierActions {
				// add a copy of multiplier action to old client map
				oldClientMap[userID][key] = &MultiplierAction{
					MultiplierValue: nm.MultiplierValue,
					Expiry:          nm.Expiry,
				}

				// add a copy of multiplier action to user sups send
				userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &passport.SupsMultiplier{
					Key:       key,
					Value:     nm.MultiplierValue,
					ExpiredAt: nm.Expiry,
				})
			}

			// append send data
			sendDiff = append(sendDiff, userSupsMultiplierSend)

			// skip
			continue
		}

		// if exists, update the old multiplier map with the new one
		for key, nm := range newMultiplier.MultiplierActions {

			// find multiplier action with key
			multiplier, ok := oldMultiplier[key]
			if !ok {

				// add a copy to old map if not exist
				oldMultiplier[key] = &MultiplierAction{
					MultiplierValue: nm.MultiplierValue,
					Expiry:          nm.Expiry,
				}

				userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &passport.SupsMultiplier{
					Key:       key,
					Value:     nm.MultiplierValue,
					ExpiredAt: nm.Expiry,
				})

				continue
			}

			// update copy if expiry is different
			if multiplier.Expiry.String() != nm.Expiry.String() {
				// add a copy to old map if not exist
				oldMultiplier[key].Expiry = nm.Expiry

				// record change
				userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &passport.SupsMultiplier{
					Key:       key,
					Value:     nm.MultiplierValue,
					ExpiredAt: nm.Expiry,
				})
			}
		}

		// append to send list if user sups multiplier is changed
		if len(userSupsMultiplierSend.SupsMultipliers) > 0 {
			sendDiff = append(sendDiff, userSupsMultiplierSend)
		}
	}

	if len(sendDiff) > 0 {
		// broadcast multiplier change
		go api.Passport.UserSupsMultiplierSend(context.Background(), sendDiff)
	}
}
