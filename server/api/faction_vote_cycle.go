package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"server"
	"server/battle_arena"
	"server/helpers"
	"sort"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

/***********************
* Faction User Tracker *
***********************/

type FactionUserMap map[server.FactionID]map[server.UserID]bool
type FactionVoteValueMap map[server.FactionID]*VoteValue
type VoteValue struct {
	Weight int64
}

const MinimumVoteWeight = 1

func (api *API) startFactionUserTracker(factions []*server.Faction) {
	factionUserMap := make(FactionUserMap)
	factionVoteValueMap := make(FactionVoteValueMap)

	for _, faction := range factions {
		factionUserMap[faction.ID] = make(map[server.UserID]bool)
		factionVoteValueMap[faction.ID] = &VoteValue{
			Weight: MinimumVoteWeight,
		}
	}

	// start channel
	go func() {
		for fn := range api.factionUserTracker {
			fn(factionUserMap, factionVoteValueMap)
		}
	}()
}

// CalcVoteWeight recalculate the weight of each faction vote
func CalcVoteWeight(fum FactionUserMap, fvv FactionVoteValueMap) {
	for factionID, fv := range fvv {
		for fid, fu := range fum {
			if factionID != fid {
				fv.Weight += int64(len(fu))
			}
		}

		// minmum weight is 1
		if fv.Weight < MinimumVoteWeight {
			fv.Weight = MinimumVoteWeight
		}
	}
}

func (api *API) GetFactionVoteWage(factionID server.FactionID) int64 {
	factionWageChan := make(chan int64, 5)

	api.factionUserTracker <- func(fum FactionUserMap, fvvm FactionVoteValueMap) {
		factionWageChan <- fvvm[factionID].Weight
	}

	return <-factionWageChan
}

/************************
* Live Voting Broadcast *
************************/

type LiveVotingData struct {
	TotalVote server.BigInt
}

func (api *API) startLiveVotingDataTicker(factionID server.FactionID) {
	// live voting data broadcast
	liveVotingData := &LiveVotingData{
		TotalVote: server.BigInt{Int: *big.NewInt(0)},
	}

	// start channel
	go func() {
		for fn := range api.liveVotingData[factionID] {
			fn(liveVotingData)
		}
	}()
}

/*************
* Vote Stage *
*************/

const (
	// CooldownInitialDurationSecond the amount of second users have to wait for the next vote coming up
	CooldownInitialDurationSecond = 5

	// FirstVoteDurationSecond the amount of second users can vote the ability
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
	UserVoteMap    map[server.UserID]map[server.TransactionReference]server.BigInt
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
	AgreedCount    []server.TransactionReference `json:"AgreedCount"`
	DisagreedCount []server.TransactionReference `json:"DisagreedCount"`
	VoteValueMap   map[server.TransactionReference]int64
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
		AgreedCount:    []server.TransactionReference{},
		DisagreedCount: []server.TransactionReference{},
		VoteValueMap:   make(map[server.TransactionReference]int64),
	}

	// initialise current vote stage
	voteStage := &VoteStage{
		Phase:   VotePhaseHold,
		EndTime: time.Now(),
	}

	// start faction voting cycle tickle
	tickle.MinDurationOverride = true
	voteStageLogger := log_helpers.NamedLogger(api.Log, fmt.Sprintf("Faction %s Voting Cycle", faction.Label)).Level(zerolog.Disabled)
	voteStageListener := tickle.New(fmt.Sprintf("Faction %s Voting Cycle", faction.Label), 1, api.voteStageListenerFactory(faction.ID))
	voteStageListener.Log = &voteStageLogger

	// tickle for broadcasting second result
	secondVoteResultLogger := log_helpers.NamedLogger(api.Log, fmt.Sprintf("Faction %s Second Vote Broadcast", faction.Label)).Level(zerolog.Disabled)
	secondVoteResultBroadcaster := tickle.New(fmt.Sprintf("Faction %s Second Vote Broadcast", faction.Label), 0.5, api.secondVoteResultBroadcasterFactory(faction.ID, voteStage, secondVoteResult))
	secondVoteResultBroadcaster.Log = &secondVoteResultLogger

	tickers := &FactionVotingTicker{
		VotingStageListener:         voteStageListener,
		SecondVoteResultBroadcaster: secondVoteResultBroadcaster,
	}

	// start channel
	go func() {
		for fn := range api.factionVoteCycle[faction.ID] {
			fn(faction, voteStage, firstVoteStat, firstVoteResult, secondVoteResult, tickers)
		}
	}()
}

// startVotingCycle start the voting cycle of the faction
func (api *API) startVotingCycle(factionID server.FactionID) {
	api.factionVoteCycle[factionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
		vs.Phase = VotePhaseVoteCooldown
		vs.EndTime = time.Now().Add(time.Duration(CooldownInitialDurationSecond) * time.Second)

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
func (api *API) secondVoteResultBroadcasterFactory(factionID server.FactionID, vs *VoteStage, svs *secondVoteResult) func() (int, error) {
	fn := func() (int, error) {
		if vs.Phase != VotePhaseSecondVote {
			return http.StatusOK, nil
		}

		// get a copy of second vote
		secondVoteResult := *svs

		// remove the unsuccessful agreed votes
		agreedVoteValue := int64(0)
		for _, txRef := range secondVoteResult.AgreedCount {
			if voteValue, ok := svs.VoteValueMap[txRef]; ok {
				agreedVoteValue += voteValue
			}
		}
		// remove the unsuccessful disagree votes
		disagreedVoteValue := int64(0)
		for _, txRef := range secondVoteResult.DisagreedCount {
			if voteValue, ok := svs.VoteValueMap[txRef]; ok {
				disagreedVoteValue += voteValue
			}
		}

		// broadcast notification to all the connected clients
		broadcastData, err := json.Marshal(&BroadcastPayload{
			Key:     HubKeyTwitchFactionSecondVoteUpdated,
			Payload: calcSecondVoteResult(factionID, agreedVoteValue, disagreedVoteValue),
		})

		if err == nil {
			api.Hub.Clients(func(clients hub.ClientsList) {

				for client, ok := range clients {
					if !ok {
						continue
					}
					go func(c *hub.Client) {
						err := c.Send(broadcastData)
						if err != nil {
							api.Log.Err(err).Msg("failed to send broadcast")
						}
					}(client)
				}
			})
		}

		return http.StatusOK, nil
	}

	return fn
}

// voteStageListenerFactory generate a vote stage listener function use in tickle
func (api *API) voteStageListenerFactory(factionID server.FactionID) func() (int, error) {
	fn := func() (int, error) {
		api.factionVoteCycle[factionID] <- func(f *server.Faction, vs *VoteStage, fvs FirstVoteState, fvr *FirstVoteResult, svs *secondVoteResult, t *FactionVotingTicker) {
			ctx := context.Background()
			// skip if it does not reach the end time or current phase is TIE
			if vs.EndTime.After(time.Now()) || vs.Phase == VotePhaseHold || vs.Phase == VotePhaseTie {
				return
			}

			// handle the ability of the end of each phase
			switch vs.Phase {

			// at the end of first vote
			case VotePhaseFirstVote:
				// get all the tx
				var txRefs []server.TransactionReference
				for _, votes := range fvs {
					for _, txMap := range votes.UserVoteMap {
						for tx := range txMap {
							txRefs = append(txRefs, tx)
						}
					}
				}

				// commit the transactions and check the status
				transactions, err := api.Passport.CommitTransactions(ctx, txRefs)
				if err != nil {
					api.Log.Err(err).Msg("failed to check transactions")
					return
				}

				parseFirstVoteResult(fvs, fvr, transactions)

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
							go func(c *hub.Client) {
								err := c.Send(broadcastData)
								if err != nil {
									api.Log.Err(err).Msg("failed to send broadcast")
								}
							}(client)
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

				// validate the agreed votes
				agreeTx, err := api.Passport.CommitTransactions(ctx, svs.AgreedCount)
				if err != nil {
					api.Log.Err(err).Msg("failed to check transactions")
					return
				}

				// remove the unsuccessful agreed votes
				agreedVoteValue := int64(0)
				for _, tx := range agreeTx {
					if tx.Status == server.TransactionSuccess {
						if voteValue, ok := svs.VoteValueMap[tx.TransactionReference]; ok {
							agreedVoteValue += voteValue
						}
					}
				}

				// validate the disagreed votes
				disagreeTx, err := api.Passport.CommitTransactions(ctx, svs.DisagreedCount)
				if err != nil {
					api.Log.Err(err).Msg("failed to check transactions")
					return
				}

				// remove the unsuccessful disagree votes
				disagreedVoteValue := int64(0)
				for _, tx := range disagreeTx {
					if tx.Status == server.TransactionSuccess {
						if voteValue, ok := svs.VoteValueMap[tx.TransactionReference]; ok {
							disagreedVoteValue += voteValue
						}
					}
				}

				// broadcast the latest vote result to all the connected clients
				broadcastData, err := json.Marshal(&BroadcastPayload{
					Key:     HubKeyTwitchFactionSecondVoteUpdated,
					Payload: calcSecondVoteResult(f.ID, agreedVoteValue, disagreedVoteValue),
				})
				if err == nil {
					api.Hub.Clients(func(clients hub.ClientsList) {

						for client, ok := range clients {
							if !ok {
								continue
							}
							go func(c *hub.Client) {
								err := c.Send(broadcastData)
								if err != nil {
									api.Log.Err(err).Msg("failed to send broadcast")
								}
							}(client)
						}
					})
				}

				// enter ability select
				if len(svs.AgreedCount) >= len(svs.DisagreedCount) {
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

				// broadcast current stage to current faction users
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

				// broadcast countered notification to all the connected clients
				broadcastData, err = json.Marshal(&BroadcastPayload{
					Key: HubKeyTwitchNotification,
					Payload: &TwitchNotification{
						Type: TwitchNotificationTypeText,
						Data: fmt.Sprintf("Action %s from Faction %s has been countered", fvs[fvr.factionAbilityID].FactionAbility.Label, f.Label),
					},
				})
				if err == nil {
					api.Hub.Clients(func(clients hub.ClientsList) {

						for client, ok := range clients {
							if !ok {
								continue
							}
							go func(c *hub.Client) {
								err := c.Send(broadcastData)
								if err != nil {
									api.Log.Err(err).Msg("failed to send broadcast")
								}
							}(client)
						}
					})
				}

				// signal ability countered animation
				err = api.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
					FactionID:           f.ID,
					FactionAbilityID:    fvr.factionAbilityID,
					GameClientAbilityID: fvs[fvr.factionAbilityID].FactionAbility.GameClientAbilityID,
					IsSuccess:           false,
				})
				if err != nil {
					api.Log.Err(err).Msg("failed to call FactionAbilityTrigger")
					return
				}

				// broadcast next stage
				vs.Phase = VotePhaseVoteCooldown
				vs.EndTime = time.Now().Add(time.Duration(fvs[fvr.factionAbilityID].FactionAbility.CooldownDurationSecond) * time.Second)

				// broadcast current stage to current faction users
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

				// at the end of ability select
			case VotePhaseLocationSelect:
				if len(fvr.hubClientID) > 1 {
					fvr.hubClientID = fvr.hubClientID[1:]
				} else {
					fvr.hubClientID = []server.UserID{}
				}

				if len(fvr.hubClientID) == 0 {
					if t.SecondVoteResultBroadcaster.NextTick != nil {
						t.SecondVoteResultBroadcaster.Stop()
					}

					// broadcast countered notification
					broadcastData, err := json.Marshal(&BroadcastPayload{
						Key: HubKeyTwitchNotification,
						Payload: &TwitchNotification{
							Type: TwitchNotificationTypeText,
							Data: fmt.Sprintf("Action %s from Faction %s has been cancelled, due to no one selecting the location.", fvs[fvr.factionAbilityID].FactionAbility.Label, f.Label),
						},
					})
					if err != nil {
						api.Log.Err(err).Msg("marshal broadcast payload")
						return
					}

					api.Hub.Clients(func(clients hub.ClientsList) {
						for client, ok := range clients {
							if !ok {
								continue
							}
							go func(c *hub.Client) {
								err := c.Send(broadcastData)
								if err != nil {
									api.Log.Err(err).Msg("failed to send broadcast")
								}
							}(client)
						}
					})

					// signal ability countered animation
					err = api.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
						FactionID:           f.ID,
						FactionAbilityID:    fvr.factionAbilityID,
						GameClientAbilityID: fvs[fvr.factionAbilityID].FactionAbility.GameClientAbilityID,
						IsSuccess:           false,
					})
					if err != nil {
						api.Log.Err(err).Msg("Failed to call FactionAbilityTrigger")
						return
					}

					// broadcast next stage
					vs.Phase = VotePhaseVoteCooldown
					vs.EndTime = time.Now().Add(time.Duration(fvs[fvr.factionAbilityID].FactionAbility.CooldownDurationSecond) * time.Second)

					// broadcast current stage to current faction users
					api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchFactionVoteStageUpdated, f.ID)), vs)

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

				// set first vote abilities
				for _, ability := range abilities {
					fvs[ability.ID] = &FirstVoteAction{
						FactionAbility: ability,
						UserVoteMap:    make(map[server.UserID]map[server.TransactionReference]server.BigInt),
					}
				}

				// initialise first vote result
				fvr = &FirstVoteResult{
					factionAbilityID: server.FactionAbilityID(uuid.Nil),
					hubClientID:      []server.UserID{},
				}

				// initialise second vote state
				svs.AgreedCount = []server.TransactionReference{}
				svs.DisagreedCount = []server.TransactionReference{}
				for key := range svs.VoteValueMap {
					delete(svs.VoteValueMap, key)
				}

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

// parseFirstVoteResult return the most voted ability and the user who contributed the most vote on that ability
func parseFirstVoteResult(fvs FirstVoteState, fvr *FirstVoteResult, checkedTransactions []*server.Transaction) {
	// initialise first vote result
	fvr.factionAbilityID = server.FactionAbilityID(uuid.Nil)
	fvr.hubClientID = []server.UserID{}

	// check there is vote exists
	hasVote := false
	for _, fv := range fvs {
		if len(fv.UserVoteMap) > 0 {
			hasVote = true
			break
		}
	}

	// exit check if there is no vote
	if !hasVote {
		return
	}

	// start parsing vote result
	type voter struct {
		id         server.UserID
		totalVotes server.BigInt
	}
	type ability struct {
		id         server.FactionAbilityID
		totalVotes server.BigInt
		voters     []*voter
	}

	abilities := []*ability{}
	for abilityID, fv := range fvs {
		ability := &ability{
			id:         abilityID,
			totalVotes: server.BigInt{Int: *big.NewInt(0)},
			voters:     []*voter{},
		}

		// here we count the total amount of sups spent on that ability
		for voterID, txMap := range fv.UserVoteMap {
			voter := &voter{
				id:         voterID,
				totalVotes: server.BigInt{Int: *big.NewInt(0)},
			}
			// tally their votes
			for tx, voteCount := range txMap {
				// validate transfer was successful here then add it
				for _, chktx := range checkedTransactions {
					if tx == chktx.TransactionReference && chktx.Status == server.TransactionSuccess {
						voter.totalVotes.Add(&voter.totalVotes.Int, &voteCount.Int)
						continue
					}
				}
			}
			// append the voter and sups spent on that vote
			ability.totalVotes.Add(&ability.totalVotes.Int, &voter.totalVotes.Int)
			ability.voters = append(ability.voters, voter)
		}

		// skip if no vote
		if ability.totalVotes.BitLen() == 0 {
			continue
		}

		abilities = append(abilities, ability)
	}

	// skip if there is no ability selected
	if len(abilities) == 0 {
		return
	}

	// if only one ability
	if len(abilities) == 1 {
		// if only one voter
		if len(abilities[0].voters) == 1 {
			fvr.factionAbilityID = abilities[0].id
			fvr.hubClientID = []server.UserID{abilities[0].voters[0].id}
			return
		}

		// sort voters
		sort.Slice(abilities[0].voters, func(i, j int) bool {
			return abilities[0].voters[i].totalVotes.Cmp(&abilities[0].voters[j].totalVotes.Int) == 1
		})

		// set first vote result
		fvr.factionAbilityID = abilities[0].id
		for _, voter := range abilities[0].voters {
			fvr.hubClientID = append(fvr.hubClientID, voter.id)
		}
		return
	}

	// sort ability list
	sort.Slice(abilities, func(i, j int) bool {
		return abilities[i].totalVotes.Cmp(&abilities[j].totalVotes.Int) == 1
	})

	// if only one voter in current ability
	if len(abilities[0].voters) == 1 {
		fvr.factionAbilityID = abilities[0].id
		fvr.hubClientID = []server.UserID{abilities[0].voters[0].id}
		return
	}

	// sort voters
	sort.Slice(abilities[0].voters, func(i, j int) bool {
		return abilities[0].voters[i].totalVotes.Cmp(&abilities[0].voters[j].totalVotes.Int) == 1
	})

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
func calcSecondVoteResult(factionID server.FactionID, agreedVoteValue, disagreedVoteValue int64) *secondVoteResultResponse {
	resp := &secondVoteResultResponse{
		FactionID: factionID,
		Result:    0.5,
	}
	if agreedVoteValue > 0 || disagreedVoteValue > 0 {
		resp.Result = float64(disagreedVoteValue) / float64(agreedVoteValue+disagreedVoteValue)
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
