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
	// CooldownInitialDurationSecond the amount of second users have to wait for the next vote comming up
	CooldownInitialDurationSecond = 5

	// FirstVoteDurationSecond the amount of second users can vote the abilitys
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
	FactionAbility *server.FactionAbility
	UserVoteMap    map[server.UserID]int64
}

type FirstVoteState map[server.FactionAbilityID]*FirstVoteAction

type FirstVoteResult struct {
	factionAbilityID server.FactionAbilityID
	hubClientID      []server.UserID
}

type secondVoteCandidate struct {
	Faction        *server.Faction        `json:"faction"`
	FactionAbility *server.FactionAbility `json:"factionAbility"`
	EndTime        time.Time              `json:"endTime"`
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
		factionAbilityID: server.FactionAbilityID(uuid.Nil),
		hubClientID:      []server.UserID{},
	}

	// initialise second vote stat
	secondVoteResult := &secondVoteResult{
		AgreedCount:    0,
		DisagreedCount: 0,
	}

	// initialise current vote stage
	voteStage := &VoteStage{
		Phase:   VotePhaseHold,
		EndTime: time.Now(),
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
		cooldownSecond := CooldownInitialDurationSecond
		if !fvr.factionAbilityID.IsNil() {
			cooldownSecond = fvs[fvr.factionAbilityID].FactionAbility.CooldownDurationSecond
		}
		vs.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)

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

			// handle the ability of the end of each phase
			switch vs.Phase {

			// at the end of first vote
			case VotePhaseFirstVote:
				parseFirstVoteResult(fvs, fvr)

				// enter TIE phase if no result
				if fvr.factionAbilityID.IsNil() || len(fvr.hubClientID) == 0 {
					vs.Phase = VotePhaseTie
					// broadcast TIE phase to faction users
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

					// stop ticker
					if t.VotingStageListener.NextTick != nil {
						t.VotingStageListener.Stop()
					}
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
							Faction:        f,
							FactionAbility: fvs[fvr.factionAbilityID].FactionAbility,
							EndTime:        vs.EndTime,
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
				// stop second vote broadcaster
				if t.SecondVoteResultBroadcaster.NextTick != nil {
					t.SecondVoteResultBroadcaster.Stop()
				}

				// broadcast latest vote result to all the connected clients
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

				// enter ability select
				if svs.AgreedCount >= svs.DisagreedCount {
					vs.Phase = VotePhaseLocationSelect
					vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

					// broadcast current stage to current faction users
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

					// unicast to the winner
					winnerClientID := fvr.hubClientID[0]
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchVoteWinnerAnnouncement, winnerClientID)),
						&secondVoteCandidate{
							Faction:        f,
							FactionAbility: fvs[fvr.factionAbilityID].FactionAbility,
							EndTime:        vs.EndTime,
						},
					)

					return
				}

				// pause the whole voting cycle, wait until animation finish
				vs.Phase = VotePhaseHold
				if t.VotingStageListener.NextTick != nil {
					t.VotingStageListener.Stop()
				}

				// broadcast current stage to current faction users
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

				// broadcast counterred notification to all the connected clients
				broadcastData, err = json.Marshal(&BroadcastPayload{
					Key: HubKeyTwitchNotification,
					Payload: &TwitchNotification{
						Type: TwitchNotificationTypeText,
						Data: fmt.Sprintf("Action %s from Faction %s have been countered", fvs[fvr.factionAbilityID].FactionAbility.Label, f.Label),
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

				// signal ability countered animation
				api.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
					FactionID:        f.ID,
					FactionAbilityID: fvr.factionAbilityID,
					IsSuccess:        false,
				})

				// at the end of ability select
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
							Data: fmt.Sprintf("Action %s from Faction %s have been cancelled, due to no one select the location.", fvs[fvr.factionAbilityID].FactionAbility.Label, f.Label),
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

					// signal ability countered animation
					api.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
						FactionID:        f.ID,
						FactionAbilityID: fvr.factionAbilityID,
						IsSuccess:        false,
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
						Faction:        f,
						FactionAbility: fvs[fvr.factionAbilityID].FactionAbility,
						EndTime:        vs.EndTime,
					},
				)

			// at the end of cooldown
			case VotePhaseVoteCooldown:
				// query faction ability
				abilities, err := api.BattleArena.FactionAbilitiesQuery(f.ID)
				if err != nil {
					api.Log.Err(err).Msgf("Failed to query abilities for faction %s", f.Label)
				}

				// initialise first vote state
				for abilityKey, fv := range fvs {
					for hubClientKey := range fv.UserVoteMap {
						delete(fv.UserVoteMap, hubClientKey)
					}
					delete(fvs, abilityKey)
				}

				// set first vote abilitys
				for _, ability := range abilities {
					fvs[ability.ID] = &FirstVoteAction{
						FactionAbility: ability,
						UserVoteMap:    make(map[server.UserID]int64),
					}
				}

				// initialise first vote result
				fvr = &FirstVoteResult{
					factionAbilityID: server.FactionAbilityID(uuid.Nil),
					hubClientID:      []server.UserID{},
				}

				// initialise second vote state
				svs.AgreedCount = 0
				svs.DisagreedCount = 0

				// start a new vote

				// broadcast first vote options
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionAbilityUpdated, f.ID)), abilities)

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

// parseFirstVoteResult return the most voted ability and the user who contribute the most vote on that ability
func parseFirstVoteResult(fvs FirstVoteState, fvr *FirstVoteResult) {
	// initialise first vote result
	fvr.factionAbilityID = server.FactionAbilityID(uuid.Nil)
	fvr.hubClientID = []server.UserID{}

	type voter struct {
		id        server.UserID
		totalVote int64
	}
	type ability struct {
		id        server.FactionAbilityID
		totalVote int64
		voters    []*voter
	}
	abilities := []*ability{}
	for abilityID, fv := range fvs {
		ability := &ability{
			id:        abilityID,
			totalVote: 0,
			voters:    []*voter{},
		}

		for voterID, voteCount := range fv.UserVoteMap {
			voter := &voter{
				id:        voterID,
				totalVote: voteCount,
			}

			ability.totalVote += voteCount
			ability.voters = append(ability.voters, voter)
		}

		// skip if no vote
		if ability.totalVote == 0 {
			continue
		}

		abilities = append(abilities, ability)
	}

	// exit, if no ability
	if len(abilities) == 0 {
		return
	}

	// if only one ability
	if len(abilities) == 1 {
		// exit, if no voters
		if len(abilities[0].voters) == 0 {
			return
		}

		// if only one voter
		if len(abilities[0].voters) == 1 {
			fvr.factionAbilityID = abilities[0].id
			fvr.hubClientID = []server.UserID{abilities[0].voters[0].id}
			return
		}

		// sort voters
		sort.Slice(abilities[0].voters, func(i, j int) bool {
			return abilities[0].voters[i].totalVote > abilities[0].voters[j].totalVote
		})

		// exit, if tie on vote
		if abilities[0].voters[0].totalVote == abilities[0].voters[1].totalVote {
			return
		}

		// set first vote result
		fvr.factionAbilityID = abilities[0].id
		for _, voter := range abilities[0].voters {
			fvr.hubClientID = append(fvr.hubClientID, voter.id)
		}
		return
	}

	// sort ability list
	sort.Slice(abilities, func(i, j int) bool {
		return abilities[i].totalVote > abilities[j].totalVote
	})

	// exit, if no voters
	if len(abilities[0].voters) == 0 {
		return
	}

	// if only one voter in current ability
	if len(abilities[0].voters) == 1 {
		fvr.factionAbilityID = abilities[0].id
		fvr.hubClientID = []server.UserID{abilities[0].voters[0].id}
		return
	}

	// exit, if tie on ability
	if abilities[0].totalVote == abilities[1].totalVote {
		return
	}

	// sort voters
	sort.Slice(abilities[0].voters, func(i, j int) bool {
		return abilities[0].voters[i].totalVote > abilities[0].voters[j].totalVote
	})

	// exit, if tie on user vote
	if abilities[0].voters[0].totalVote == abilities[0].voters[1].totalVote {
		return
	}

	// set first vote result
	fvr.factionAbilityID = abilities[0].id
	for _, voter := range abilities[0].voters {
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
				Faction:        f,
				FactionAbility: fvs[fvr.factionAbilityID].FactionAbility,
				EndTime:        vs.EndTime,
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
