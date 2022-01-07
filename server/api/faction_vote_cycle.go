package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"server"
	"server/battle_arena"
	"server/helpers"
	"sort"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/rs/zerolog"
)

/*************
* Vote Stage *
*************/

const (
	// CooldownDurationSecond the amount of second users have to wait for the next vote comming up
	CooldownDurationSecond = 5

	// FirstVoteDurationSecond the amount of second users can vote the actions
	FirstVoteDurationSecond = 10

	// SecondVoteDurationSecond the amount of seconds users can vote the second vote
	SecondVoteDurationSecond = 10

	// LocationSelectDurationSecond the amount of second the winner user can select the location
	LocationSelectDurationSecond = 20
)

type VoteStage struct {
	Phase   VotePhase `json:"phase"`
	EndTime time.Time `json:"endTime"`
}

type VotePhase string

const (
	VotePhaseHold           VotePhase = "HOLD" // Waiting on signal
	VotePhaseFirstVote      VotePhase = "FIRST_VOTE"
	VotePhaseTie            VotePhase = "TIE"
	VotePhaseSecondVote     VotePhase = "SECOND_VOTE"
	VotePhaseLocationSelect VotePhase = "LOCATION_SELECT"
	VotePhaseVoteCooldown   VotePhase = "VOTE_COOLDOWN"
)

type FactionVotingTicker struct {
	VotingStageListener         *tickle.Tickle
	SecondVoteResultBroadcaster *tickle.Tickle
}

/**************
* Vote Result *
**************/

type FirstVoteAction struct {
	FactionAction *server.FactionAction
	UserVoteMap   map[server.UserID]int64
}

type FirstVoteState map[server.FactionActionID]*FirstVoteAction

type FirstVoteResult struct {
	factionActionID server.FactionActionID
	hubClientID     []server.UserID
}

type secondVoteCandidate struct {
	Faction       *server.Faction       `json:"faction"`
	FactionAction *server.FactionAction `json:"factionAction"`
	EndTime       time.Time             `json:"endTime"`
}

type secondVoteResult struct {
	AgreedCount    int64 `json:"AgreedCount"`
	DisagreedCount int64 `json:"DisagreedCount"`
}

/***********
* Channels *
***********/

func (api *API) startFactionVoteCycle(faction *server.Faction) {
	// initialise first vote stat
	firstVoteStat := make(FirstVoteState)

	// initialise first vote result
	firstVoteResult := &FirstVoteResult{
		factionActionID: server.FactionActionID(uuid.Nil),
		hubClientID:     []server.UserID{},
	}

	// initialise second vote stat
	secondVoteResult := &secondVoteResult{
		AgreedCount:    0,
		DisagreedCount: 0,
	}

	// initialise current vote stage
	voteStage := &VoteStage{
		Phase:   VotePhaseVoteCooldown,
		EndTime: time.Now().Add(CooldownDurationSecond * time.Second),
	}

	// start faction voting cycle tickle
	tickle.MinDurationOverride = true
	voteStageLogger := log_helpers.NamedLogger(api.Log, "FactionID Voting Cycle").Level(zerolog.Disabled)

	voteStageListener := tickle.New("FactionID Voting Cycle", 1, api.voteStageListenerFactory(faction.ID))
	voteStageListener.Log = &voteStageLogger

	// tickle for broadcasting second result
	secondVoteResultLogger := log_helpers.NamedLogger(api.Log, "FactionID Second Vote Broadcast").Level(zerolog.Disabled)

	secondVoteResultBroadcaster := tickle.New("FactionID Second Vote Broadcast", 0.5, api.secondVoteResultBroadcasterFactory(faction.ID))
	secondVoteResultBroadcaster.Log = &secondVoteResultLogger

	tickers := &FactionVotingTicker{
		VotingStageListener:         voteStageListener,
		SecondVoteResultBroadcaster: secondVoteResultBroadcaster,
	}

	// add event handlers in here
	api.BattleArena.Events.AddEventHandler(battle_arena.Event(fmt.Sprintf("%s:%s", faction.ID, battle_arena.EventAnamationEnd)), api.startVotingCycleFactory(faction.ID))

	// start channel
	go func() {
		for {
			select {
			case fn := <-api.factionVoteCycle[faction.ID]:
				fn(faction, voteStage, firstVoteStat, firstVoteResult, secondVoteResult, tickers)
			}
		}
	}()

}

// startVotingCycleFactory create a function to handle voting start event
func (api *API) startVotingCycleFactory(factionID server.FactionID) func(ctx context.Context, ed *battle_arena.EventData) {
	fn := func(ctx context.Context, ed *battle_arena.EventData) {
		api.startVotingCycle(factionID)
	}
	return fn
}

// startVotingCycle start the voting cycle of the faction
func (api *API) startVotingCycle(factionID server.FactionID) {
	api.factionVoteCycle[factionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		vs.Phase = VotePhaseVoteCooldown
		vs.EndTime = time.Now().Add(CooldownDurationSecond * time.Second)

		// broadcast current stage to current faction users
		api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

		if t.VotingStageListener.NextTick == nil {
			t.VotingStageListener.Start()
		}
	}
}

// pauseVotingCycle pause the voting cycle of the faction
func (api *API) pauseVotingCycle(factionID server.FactionID) {
	api.factionVoteCycle[factionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		vs.Phase = VotePhaseHold

		if t.VotingStageListener.NextTick != nil {
			t.VotingStageListener.Stop()
		}

		if t.SecondVoteResultBroadcaster.NextTick != nil {
			t.SecondVoteResultBroadcaster.Stop()
		}
	}
}

// secondVoteResultBroadcasterFactory generate the function for broadcasting the second vote result
func (api *API) secondVoteResultBroadcasterFactory(factionID server.FactionID) func() (int, error) {
	fn := func() (int, error) {
		api.factionVoteCycle[factionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
			if vs.Phase != VotePhaseSecondVote {
				return
			}

			// broadcast notification to all the connected clients
			broadcastData, err := json.Marshal(&BroadcastPayload{
				Key:     HubKeyTwitchFactionSecondVoteUpdated,
				Payload: calcSecondVoteResult(f.ID, svs),
			})
			if err == nil {
				api.Hub.Clients(func(clients hub.ClientsList) {

					for client, ok := range clients {
						if !ok {
							continue
						}
						go client.Send(broadcastData)
					}
				})
			}
		}

		return http.StatusOK, nil
	}

	return fn
}

// voteStageListenerFactory generate a vote stage listener function use in tickle
func (api *API) voteStageListenerFactory(factionID server.FactionID) func() (int, error) {
	fn := func() (int, error) {
		api.factionVoteCycle[factionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {

			// skip if does not reach the end time or current phase is TIE
			if vs.EndTime.After(time.Now()) || vs.Phase == VotePhaseHold || vs.Phase == VotePhaseTie {
				return
			}

			// handle the action of the end of each phase
			switch vs.Phase {

			// at the end of first vote
			case VotePhaseFirstVote:
				parseFirstVoteResult(fvs, fvr)

				// enter TIE phase if no result
				if fvr.factionActionID.IsNil() || len(fvr.hubClientID) == 0 {
					vs.Phase = VotePhaseTie
					// broadcast TIE phase to faction users
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)
					return
				}

				// otherwise, enter second vote stage
				vs.Phase = VotePhaseSecondVote
				vs.EndTime = time.Now().Add(SecondVoteDurationSecond * time.Second)

				// broadcast current stage to current faction users
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

				// broadcast second vote candidate
				broadcastData, err := json.Marshal(&BroadcastPayload{
					Key: HubKeyTwitchNotification,
					Payload: &TwitchNotification{
						Type: TwitchNotificationTypeSecondVote,
						Data: &secondVoteCandidate{
							Faction:       f,
							FactionAction: fvs[fvr.factionActionID].FactionAction,
							EndTime:       vs.EndTime,
						},
					},
				})
				if err == nil {
					api.Hub.Clients(func(clients hub.ClientsList) {
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

			// at the end of second vote
			case VotePhaseSecondVote:

				// enter action select
				if svs.AgreedCount >= svs.DisagreedCount {
					vs.Phase = VotePhaseLocationSelect
					vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

					// broadcast current stage to current faction users
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

					// unicast to the winner
					winnerClientID := fvr.hubClientID[0]
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchVoteWinnerAnnouncement, winnerClientID)),
						&secondVoteCandidate{
							Faction:       f,
							FactionAction: fvs[fvr.factionActionID].FactionAction,
							EndTime:       vs.EndTime,
						},
					)

					// stop second vote broadcaster
					if t.SecondVoteResultBroadcaster.NextTick != nil {
						t.SecondVoteResultBroadcaster.Stop()
					}

					return
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
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

				// broadcast counterred notification to all the connected clients
				broadcastData, err := json.Marshal(&BroadcastPayload{
					Key: HubKeyTwitchNotification,
					Payload: &TwitchNotification{
						Type: TwitchNotificationTypeText,
						Data: fmt.Sprintf("Action %s from Faction %s have been countered", fvs[fvr.factionActionID].FactionAction.Label, f.Label),
					},
				})
				if err == nil {
					api.Hub.Clients(func(clients hub.ClientsList) {

						for client, ok := range clients {
							if !ok {
								continue
							}
							go client.Send(broadcastData)
						}
					})
				}

				// signal action countered animation
				api.BattleArena.FactionActionTrigger(&battle_arena.ActionTriggerRequest{
					FactionID:       f.ID,
					FactionActionID: fvr.factionActionID,
					IsSuccess:       false,
				})

				// at the end of action select
			case VotePhaseLocationSelect:
				if len(fvr.hubClientID) > 1 {
					fvr.hubClientID = fvr.hubClientID[1:]
				} else {
					fvr.hubClientID = []server.UserID{}
				}

				if len(fvr.hubClientID) == 0 {

					vs.Phase = VotePhaseHold
					if t.VotingStageListener.NextTick != nil {
						t.VotingStageListener.Stop()
					}

					if t.SecondVoteResultBroadcaster.NextTick != nil {
						t.SecondVoteResultBroadcaster.Stop()
					}

					// broadcast current stage to current faction users
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

					// broadcast counterred notification
					broadcastData, err := json.Marshal(&BroadcastPayload{
						Key: HubKeyTwitchNotification,
						Payload: &TwitchNotification{
							Type: TwitchNotificationTypeText,
							Data: fmt.Sprintf("Action %s from Faction %s have been cancelled, due to no one select the location.", fvs[fvr.factionActionID].FactionAction.Label, f.Label),
						},
					})
					if err == nil {
						api.Hub.Clients(func(clients hub.ClientsList) {
							for client, ok := range clients {
								if !ok {
									continue
								}
								go client.Send(broadcastData)
							}
						})
					}

					// signal action countered animation
					api.BattleArena.FactionActionTrigger(&battle_arena.ActionTriggerRequest{
						FactionID:       f.ID,
						FactionActionID: fvr.factionActionID,
						IsSuccess:       false,
					})

					return
				}

				vs.Phase = VotePhaseLocationSelect
				vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

				// broadcast current stage to current faction users
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

				// rotate to the second winner if the first winner does not pick the location
				winnerClientID := fvr.hubClientID[0]
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchVoteWinnerAnnouncement, winnerClientID)),
					&secondVoteCandidate{
						Faction:       f,
						FactionAction: fvs[fvr.factionActionID].FactionAction,
						EndTime:       vs.EndTime,
					},
				)

			// at the end of cooldown
			case VotePhaseVoteCooldown:
				// TODO: query actions from battle arena server
				actions := server.FactionActions

				// initialise first vote state
				for actionKey, fv := range fvs {
					for hubClientKey := range fv.UserVoteMap {
						delete(fv.UserVoteMap, hubClientKey)
					}
					delete(fvs, actionKey)
				}

				// set first vote actions
				for _, action := range actions {
					fvs[action.ID] = &FirstVoteAction{
						FactionAction: action,
						UserVoteMap:   make(map[server.UserID]int64),
					}
				}

				// initialise first vote result
				fvr = &FirstVoteResult{
					factionActionID: server.FactionActionID(uuid.Nil),
					hubClientID:     []server.UserID{},
				}

				// initialise second vote state
				svs.AgreedCount = 0
				svs.DisagreedCount = 0

				// start a new vote

				// broadcast first vote options
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionActionUpdated, f.ID)), actions)

				vs.Phase = VotePhaseFirstVote
				vs.EndTime = time.Now().Add(FirstVoteDurationSecond * time.Second)

				// broadcast current stage to current faction users
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)
			}
		}

		return http.StatusOK, nil
	}

	return fn
}

// parseFirstVoteResult return the most voted action and the user who contribute the most vote on that action
func parseFirstVoteResult(fvs FirstVoteState, fvr *FirstVoteResult) {
	// initialise first vote result
	fvr.factionActionID = server.FactionActionID(uuid.Nil)
	fvr.hubClientID = []server.UserID{}

	type voter struct {
		id        server.UserID
		totalVote int64
	}
	type action struct {
		id        server.FactionActionID
		totalVote int64
		voters    []*voter
	}
	actionList := []*action{}
	for actionID, fv := range fvs {
		action := &action{
			id:        actionID,
			totalVote: 0,
			voters:    []*voter{},
		}

		for voterID, voteCount := range fv.UserVoteMap {
			voter := &voter{
				id:        voterID,
				totalVote: voteCount,
			}

			action.totalVote += voteCount
			action.voters = append(action.voters, voter)
		}

		// skip if no vote
		if action.totalVote == 0 {
			continue
		}

		actionList = append(actionList, action)
	}

	// exit, if no action
	if len(actionList) == 0 {
		return
	}

	// if only one action
	if len(actionList) == 1 {
		// exit, if no voters
		if len(actionList[0].voters) == 0 {
			return
		}

		// if only one voter
		if len(actionList[0].voters) == 1 {
			fvr.factionActionID = actionList[0].id
			fvr.hubClientID = []server.UserID{actionList[0].voters[0].id}
			return
		}

		// sort voters
		sort.Slice(actionList[0].voters, func(i, j int) bool {
			return actionList[0].voters[i].totalVote > actionList[0].voters[j].totalVote
		})

		// exit, if tie on vote
		if actionList[0].voters[0].totalVote == actionList[0].voters[1].totalVote {
			return
		}

		// set first vote result
		fvr.factionActionID = actionList[0].id
		for _, voter := range actionList[0].voters {
			fvr.hubClientID = append(fvr.hubClientID, voter.id)
		}
		return
	}

	// sort action list
	sort.Slice(actionList, func(i, j int) bool {
		return actionList[i].totalVote > actionList[j].totalVote
	})

	// exit, if no voters
	if len(actionList[0].voters) == 0 {
		return
	}

	// if only one voter in current action
	if len(actionList[0].voters) == 1 {
		fvr.factionActionID = actionList[0].id
		fvr.hubClientID = []server.UserID{actionList[0].voters[0].id}
		return
	}

	// exit, if tie on action
	if actionList[0].totalVote == actionList[1].totalVote {
		return
	}

	// sort voters
	sort.Slice(actionList[0].voters, func(i, j int) bool {
		return actionList[0].voters[i].totalVote > actionList[0].voters[j].totalVote
	})

	// exit, if tie on user vote
	if actionList[0].voters[0].totalVote == actionList[0].voters[1].totalVote {
		return
	}

	// set first vote result
	fvr.factionActionID = actionList[0].id
	for _, voter := range actionList[0].voters {
		fvr.hubClientID = append(fvr.hubClientID, voter.id)
	}
}

type secondVoteResultResponse struct {
	FactionID server.FactionID `json:"factionID"`
	Result    float64          `json:"result"`
}

// calc second vote result
func calcSecondVoteResult(factionID server.FactionID, svs *secondVoteResult) *secondVoteResultResponse {
	resp := &secondVoteResultResponse{
		FactionID: factionID,
		Result:    0.5,
	}
	if svs.AgreedCount > 0 || svs.DisagreedCount > 0 {
		resp.Result = float64(svs.DisagreedCount) / float64(svs.AgreedCount+svs.DisagreedCount)
	}

	return resp
}

// GetSecondVotes return a list of second vote in current round
func (api *API) GetSecondVotes(w http.ResponseWriter, r *http.Request) (int, error) {
	resp := []*secondVoteCandidate{}
	for _, factionVoteCycle := range api.factionVoteCycle {
		secondVoteChan := make(chan *secondVoteCandidate)
		factionVoteCycle <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svr *secondVoteResult, fvt *FactionVotingTicker) {
			if vs.Phase != VotePhaseSecondVote {
				secondVoteChan <- nil
				return
			}

			secondVoteChan <- &secondVoteCandidate{
				Faction:       f,
				FactionAction: fvs[fvr.factionActionID].FactionAction,
				EndTime:       vs.EndTime,
			}
		}

		secondVote := <-secondVoteChan
		if secondVote != nil {
			resp = append(resp, secondVote)
		}
	}

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].EndTime.After(resp[j].EndTime)
	})
	return helpers.EncodeJSON(w, resp)
}
