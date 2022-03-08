package api

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"server"
	"server/battle_arena"
	"server/db"
	"sort"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
)

/***************
* Spoil of War *
***************/
func (api *API) startSpoilOfWarBroadcaster(ctx context.Context) {
	tickle.MinDurationOverride = true
	spoilOfWarBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Spoil of War Broadcaster").Level(zerolog.Disabled)
	spoilOfWarBroadcaster := tickle.New("Spoil of War Broadcaster", 5, func() (int, error) {

		result := api.Passport.GetSpoilOfWarAmount()
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeSpoilOfWarTick))
		payload = append(payload, []byte(result)...)

		api.NetMessageBus.Send(ctx, messagebus.NetBusKey(HubKeySpoilOfWarUpdated), payload)

		return http.StatusOK, nil
	})

	spoilOfWarBroadcaster.Log = &spoilOfWarBroadcasterLogger
	spoilOfWarBroadcaster.Start()
}

/********************
* Vote Price System *
********************/

const VotePriceMultiplierPercentage = 10 // 1%
const VotePriceUpdaterTickSecond = 10

const VotePriceAccuracy = 10000

func (api *API) startVotePriceSystem(ctx context.Context, conn *pgxpool.Pool) {
	// initialise value
	api.votePriceSystem = &VotePriceSystem{
		GlobalVotePerTick:   []int64{},
		GlobalTotalVote:     0,
		FactionVotePriceMap: make(map[server.FactionID]*FactionVotePrice),
	}

	// fill up array with 0
	for i := 0; i < 100; i++ {
		api.votePriceSystem.GlobalVotePerTick = append(api.votePriceSystem.GlobalVotePerTick, 0)
	}

	// initialise faction vote price map
	for _, faction := range api.factionMap {
		factionVotePrice := &FactionVotePrice{
			OuterLock:            deadlock.Mutex{},
			NextAccessLock:       deadlock.Mutex{},
			DataLock:             deadlock.Mutex{},
			CurrentVotePriceSups: server.BigInt{Int: *big.NewInt(0)},
			CurrentVotePerTick:   0,
		}

		// parse previous price
		prevPrice := server.BigInt{Int: *big.NewInt(0)}
		prevPrice.SetString(faction.VotePrice, 10)

		// set previous price to current price
		factionVotePrice.CurrentVotePriceSups.Add(&factionVotePrice.CurrentVotePriceSups.Int, &prevPrice.Int)

		// assign to map
		api.votePriceSystem.FactionVotePriceMap[faction.ID] = factionVotePrice
	}

	// initialise vote price ticker
	tickle.MinDurationOverride = true
	votePriceTickerLogger := log_helpers.NamedLogger(api.Log, "Vote Price Ticker").Level(zerolog.Disabled)
	votePriceUpdater := tickle.New("Vote Price Ticker", VotePriceUpdaterTickSecond, api.votePriceUpdaterFactory(ctx, conn))
	votePriceUpdater.Log = &votePriceTickerLogger

	api.votePriceSystem.VotePriceUpdater = votePriceUpdater
}

func absoluteInt64(x int64) int64 {
	if x < 0 {
		return -x
	}

	return x
}

// votePriceHighPriorityLock for vote price system to lock all the faction vote price lock
func (api *API) votePriceHighPriorityLock() {
	var wg deadlock.WaitGroup

	for _, fvp := range api.votePriceSystem.FactionVotePriceMap {
		wg.Add(1)

		go func(fvp *FactionVotePrice) {
			fvp.NextAccessLock.Lock()
			fvp.DataLock.Lock()
			fvp.NextAccessLock.Unlock()
			wg.Done()
		}(fvp)
	}

	wg.Wait()
}

// votePriceHighPriorityUnlock for vote price system to unlock all the faction vote price lock
func (api *API) votePriceHighPriorityUnlock() {
	for _, fvp := range api.votePriceSystem.FactionVotePriceMap {
		go func(fvp *FactionVotePrice) {
			fvp.DataLock.Unlock()
		}(fvp)
	}
}

// votePriceLowPriorityLock for vote price system to lock the given faction vote price
func (api *API) votePriceLowPriorityLock(factionID server.FactionID) {
	api.votePriceSystem.FactionVotePriceMap[factionID].OuterLock.Lock()
	api.votePriceSystem.FactionVotePriceMap[factionID].NextAccessLock.Lock()
	api.votePriceSystem.FactionVotePriceMap[factionID].DataLock.Lock()
	api.votePriceSystem.FactionVotePriceMap[factionID].NextAccessLock.Unlock()
}

// votePriceLowPriorityUnlock for vote price system to unlock the given faction vote price
func (api *API) votePriceLowPriorityUnlock(factionID server.FactionID) {
	api.votePriceSystem.FactionVotePriceMap[factionID].DataLock.Unlock()
	api.votePriceSystem.FactionVotePriceMap[factionID].OuterLock.Unlock()
}

// increaseFactionVoteTotal
func (api *API) increaseFactionVoteTotal(factionID server.FactionID, voteCount int64) {
	if fvp, ok := api.votePriceSystem.FactionVotePriceMap[factionID]; ok {
		api.votePriceLowPriorityLock(factionID)
		fvp.CurrentVotePerTick += voteCount
		api.votePriceLowPriorityUnlock(factionID)
	}
}

// vote price ticker

func (api *API) votePriceUpdaterFactory(ctx context.Context, conn *pgxpool.Pool) func() (int, error) {
	return func() (int, error) {
		api.votePriceHighPriorityLock()

		// sum current vote per tick from all the faction
		totalCurrentVote := int64(0)
		for _, fvp := range api.votePriceSystem.FactionVotePriceMap {
			totalCurrentVote += fvp.CurrentVotePerTick
		}

		// calculate total vote per tick
		api.votePriceSystem.GlobalTotalVote = api.votePriceSystem.GlobalTotalVote - api.votePriceSystem.GlobalVotePerTick[0] + totalCurrentVote

		// shift one index
		api.votePriceSystem.GlobalVotePerTick = append(api.votePriceSystem.GlobalVotePerTick[1:], totalCurrentVote)

		// async calculate new faction vote price
		var wg deadlock.WaitGroup

		for factionID, fvp := range api.votePriceSystem.FactionVotePriceMap {
			wg.Add(1)

			go func(factionID server.FactionID, fvp *FactionVotePrice) {
				newVotePrice := calVotePrice(
					api.votePriceSystem.GlobalTotalVote,
					fvp.CurrentVotePriceSups,
					fvp.CurrentVotePerTick,
					factionID,
				)

				// store new vote price to db
				err := db.FactionVotePriceUpdate(context.Background(), conn, &server.Faction{
					ID:        factionID,
					VotePrice: newVotePrice.String(),
				})
				if err != nil {
					api.Log.Err(err).Msg("failed to store new faction vote price into db")
				}

				// set current vote price
				fvp.CurrentVotePriceSups = server.BigInt{Int: newVotePrice.Int}

				// reset vote per tick for next round
				fvp.CurrentVotePerTick = 0

				// prepare broadcast payload
				payload := []byte{}
				payload = append(payload, byte(battle_arena.NetMessageTypeVotePriceTick))
				payload = append(payload, []byte(fvp.CurrentVotePriceSups.Int.String())...)

				// broadcase
				api.NetMessageBus.Send(ctx, messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionVotePriceUpdated, factionID)), payload)

				wg.Done()
			}(factionID, fvp)
		}
		wg.Wait()
		api.votePriceHighPriorityUnlock()

		return http.StatusOK, nil
	}
}

func calVotePrice(globalTotalVote int64, currentVotePrice server.BigInt, currentVotePerTick int64, factionID server.FactionID) server.BigInt {

	// get a copy of current vote price
	votePriceSups := server.BigInt{Int: *big.NewInt(0)}
	votePriceSups.Add(&votePriceSups.Int, &currentVotePrice.Int)

	// calculate max price change
	maxPriceChange := server.BigInt{Int: *big.NewInt(0)}
	maxPriceChange.Add(&maxPriceChange.Int, &votePriceSups.Int)
	maxPriceChange.Mul(&maxPriceChange.Int, big.NewInt(VotePriceMultiplierPercentage))
	maxPriceChange.Div(&maxPriceChange.Int, big.NewInt(100))

	// if no vote
	if currentVotePerTick == 0 {
		// reduce maximum price change
		votePriceSups.Sub(&votePriceSups.Int, &maxPriceChange.Int)

		return votePriceSups

	}
	// priceChange := currentFactionPrice * multiplier * (1 - abs(vpt-avpt)/avpt)
	priceChange := server.BigInt{Int: *big.NewInt(0)}

	// calc price change ratio "abs(vpt-avpt)/avpt"
	ratio := VotePriceAccuracy * absoluteInt64(currentVotePerTick*300-globalTotalVote) * 100 / (globalTotalVote * 3)

	// if ratio is greater than or equal to 1 or ratio is equal to 0,
	// direct set price change to the max price change
	if ratio >= VotePriceAccuracy*100 || ratio == 0 {
		priceChange.Add(&priceChange.Int, &maxPriceChange.Int)
	} else {
		// otherwise calc the current price change
		priceChange.Add(&priceChange.Int, &maxPriceChange.Int)
		priceChange.Mul(&priceChange.Int, big.NewInt(VotePriceAccuracy*100-ratio))
		priceChange.Div(&priceChange.Int, big.NewInt(VotePriceAccuracy*100))
	}

	// check current vote per tick is over average
	if currentVotePerTick*300 > globalTotalVote {
		// price go up
		if votePriceSups.Cmp(big.NewInt(1000000000000000000)) < 0 {
			priceChange.Mul(&priceChange.Int, big.NewInt(4))
		}

		votePriceSups.Add(&votePriceSups.Int, &priceChange.Int)
	} else {
		// price go down
		if votePriceSups.Cmp(big.NewInt(1000000000)) < 0 { // price floor
			priceChange = server.BigInt{Int: *big.NewInt(0)}
		} else if votePriceSups.Cmp(big.NewInt(1000000000000)) < 0 {
			priceChange.Div(&priceChange.Int, big.NewInt(5))
		} else if votePriceSups.Cmp(big.NewInt(1000000000000000000)) < 0 {
			priceChange.Div(&priceChange.Int, big.NewInt(3))
		}

		priceChange.Div(&priceChange.Int, big.NewInt(2))

		votePriceSups.Sub(&votePriceSups.Int, &priceChange.Int)
	}

	if votePriceSups.Cmp(big.NewInt(1000000000000)) <= 0 {
		// set minimum price of voting
		return server.BigInt{Int: *big.NewInt(1000000000000)}
	}

	return votePriceSups
}

/************************
* Live Voting Broadcast *
************************/

/***************
* Voting Cycle *
***************/

const (
	// VoteAbilityRightDurationSecond the amount of second users can vote the ability
	VoteAbilityRightDurationSecond = 30

	// LocationSelectDurationSecond the amount of second the winner user can select the location
	LocationSelectDurationSecond = 15
)

type VotePhase string

const (
	VotePhaseHold             VotePhase = "HOLD" // Waiting on signal
	VotePhaseWaitMechIntro    VotePhase = "WAIT_MECH_INTRO"
	VotePhaseVoteCooldown     VotePhase = "VOTE_COOLDOWN"
	VotePhaseVoteAbilityRight VotePhase = "VOTE_ABILITY_RIGHT"
	VotePhaseNextVoteWin      VotePhase = "NEXT_VOTE_WIN"
	VotePhaseLocationSelect   VotePhase = "LOCATION_SELECT"
)

/**************
* Vote Struct *
**************/

type VotePhaseChecker struct {
	deadlock.RWMutex
	Phase   VotePhase `json:"phase"`
	EndTime time.Time `json:"endTime"`
}

type VoteAbility struct {
	BattleAbility     *server.BattleAbility
	FactionAbilityMap map[server.FactionID]*server.GameAbility
	deadlock.Mutex
}

type FactionUserVoteMap map[server.FactionID]map[server.UserID]int64

type FactionTransactions struct {
	Transactions []string
	deadlock.Mutex
}

type FactionTotalVote struct {
	RedMountainTotalVote int64
	BostonTotalVote      int64
	ZaibatsuTotalVote    int64
}

type VotingCycleTicker struct {
	VotingStageListener           *tickle.Tickle
	AbilityRightResultBroadcaster *tickle.Tickle
}

type VoteWinner struct {
	List []server.UserID
}

type WinnerSelectAbilityLocation struct {
	GameAbility *server.GameAbility `json:"gameAbility"`
	EndTime     time.Time           `json:"endTime"`
}

// use for tracking total vote of each user per battle
type UserVoteMap map[server.UserID]int64

/***********************
* Voting Cycle Channel *
***********************/

// StartVotingCycle start voting cycle ticker
func (api *API) StartVotingCycle(ctx context.Context) {
	// initialise current vote stage
	api.votePhaseChecker = &VotePhaseChecker{
		deadlock.RWMutex{},
		VotePhaseHold,
		time.Now(),
	}

	// initialise vote ability
	voteAbility := &VoteAbility{
		BattleAbility:     &server.BattleAbility{},
		FactionAbilityMap: make(map[server.FactionID]*server.GameAbility),
	}

	// initial faction user voting map
	factionUserVoteMap := make(FactionUserVoteMap)
	for _, f := range api.factionMap {
		factionUserVoteMap[f.ID] = make(map[server.UserID]int64)
		voteAbility.FactionAbilityMap[f.ID] = &server.GameAbility{}
	}

	// initial faction transactions
	factionTransactions := &FactionTransactions{
		Transactions: []string{},
	}

	// initialise faction total vote
	factionTotalVote := &FactionTotalVote{
		RedMountainTotalVote: 0,
		BostonTotalVote:      0,
		ZaibatsuTotalVote:    0,
	}

	// initialise vote winner
	voteWinner := &VoteWinner{
		List: []server.UserID{},
	}

	// initialise user vote map
	userVoteMap := make(UserVoteMap)

	// start faction voting cycle tickle
	tickle.MinDurationOverride = true
	voteStageLogger := log_helpers.NamedLogger(api.Log, "Voting Cycle Tracker").Level(zerolog.Disabled)
	voteStageListener := tickle.New("Voting Cycle Tracker", 1, api.voteStageListenerFactory(ctx))
	voteStageListener.Log = &voteStageLogger

	// start faction voting cycle tickle
	abilityRightResultLogger := log_helpers.NamedLogger(api.Log, "Ability Right Result Broadcaster").Level(zerolog.Disabled)
	abilityRightResultBroadcaster := tickle.New("Ability Right Result Broadcaster", 0.5, api.abilityRightResultBroadcasterFactory(ctx, factionTotalVote))
	abilityRightResultBroadcaster.Log = &abilityRightResultLogger

	tickers := &VotingCycleTicker{
		VotingStageListener:           voteStageListener,
		AbilityRightResultBroadcaster: abilityRightResultBroadcaster,
	}

	api.VotingCycle = func(fn func(*VoteAbility, FactionUserVoteMap, *FactionTransactions, *FactionTotalVote, *VoteWinner, *VotingCycleTicker, UserVoteMap)) {
		fn(voteAbility, factionUserVoteMap, factionTransactions, factionTotalVote, voteWinner, tickers, userVoteMap)
	}
}

/******************************
* Ability Right Result Ticker *
******************************/

// abilityRightResultBroadcasterFactory generate the function for broadcasting the ability right result
func (api *API) abilityRightResultBroadcasterFactory(ctx context.Context, ftv *FactionTotalVote) func() (int, error) {
	return func() (int, error) {
		// save a snapshot of current faction total vote
		factionTotalVote := *ftv

		// initialise ratio data
		ratioData := "333333,333333,333333"

		// if any faction have more than 1 vote, start calculate
		if factionTotalVote.BostonTotalVote > 0 || factionTotalVote.RedMountainTotalVote > 0 || factionTotalVote.ZaibatsuTotalVote > 0 {
			// calc ratio
			totalVote := factionTotalVote.BostonTotalVote + factionTotalVote.RedMountainTotalVote + factionTotalVote.ZaibatsuTotalVote

			// calc ratio
			redMountainRatio := factionTotalVote.RedMountainTotalVote * 10000 * 100 / totalVote
			bostonRatio := factionTotalVote.BostonTotalVote * 10000 * 100 / totalVote
			zaibatsuRatio := factionTotalVote.ZaibatsuTotalVote * 10000 * 100 / totalVote

			ratioData = fmt.Sprintf("%d,%d,%d", redMountainRatio, bostonRatio, zaibatsuRatio)
		}

		// prepare broadcast data
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeAbilityRightRatioTick))
		payload = append(payload, []byte(ratioData)...)

		api.NetMessageBus.Send(ctx, messagebus.NetBusKey(HubKeyAbilityRightRatioUpdated), payload)

		return http.StatusOK, nil
	}
}

func getRatio(rmVote, bVote, zVote int64) []byte {
	// calc ratio
	totalVote := bVote + rmVote + zVote

	// calc ratio
	redMountainRatio := rmVote * 10000 * 100 / totalVote
	bostonRatio := bVote * 10000 * 100 / totalVote
	zaibatsuRatio := zVote * 10000 * 100 / totalVote

	ratioData := fmt.Sprintf("%d,%d,%d", redMountainRatio, bostonRatio, zaibatsuRatio)

	payload := []byte{}
	payload = append(payload, byte(battle_arena.NetMessageTypeAbilityRightRatioTick))
	payload = append(payload, []byte(ratioData)...)

	return payload
}

/**********************
* Voting Stage Ticker *
**********************/

// startVotingCycle start voting cycle tickles
func (api *API) startVotingCycle(ctx context.Context, introSecond int) {
	api.VotingCycle(func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
		api.votePhaseChecker.Lock()
		api.votePhaseChecker.Phase = VotePhaseWaitMechIntro
		api.votePhaseChecker.EndTime = time.Now().Add(time.Duration(introSecond) * time.Second)
		api.votePhaseChecker.Unlock()

		// broadcast current stage to faction users
		go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

		vct.VotingStageListener.Start()
	})
}

// stopVotingCycle pause voting cycle tickles
func (api *API) stopVotingCycle(ctx context.Context) []*server.BattleUserVote {
	userVoteCounts := []*server.BattleUserVote{}
	api.VotingCycle(func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
		api.votePhaseChecker.Lock()
		api.votePhaseChecker.Phase = VotePhaseHold
		api.votePhaseChecker.Unlock()

		// broadcast current stage to faction users
		go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

		if vct.VotingStageListener.NextTick != nil {
			vct.VotingStageListener.Stop()
		}

		if vct.AbilityRightResultBroadcaster.NextTick != nil {
			vct.AbilityRightResultBroadcaster.Stop()
		}

		// stop vote price update when game end
		if api.votePriceSystem.VotePriceUpdater.NextTick != nil {
			api.votePriceSystem.VotePriceUpdater.Stop()
		}

		fts.Lock()
		defer fts.Unlock()
		if len(fts.Transactions) > 0 {
			// commit the transactions
			api.Passport.ReleaseTransactions(fts.Transactions)
		}

		for userID, voteCount := range uvm {
			// get a copy of the user vote map
			userVoteCounts = append(userVoteCounts, &server.BattleUserVote{
				UserID:    userID,
				VoteCount: voteCount,
			})

			// delete current user vote
			delete(uvm, userID)
		}
	})

	return userVoteCounts
}

// voteStageListenerFactory is the main vote stage handler
func (api *API) voteStageListenerFactory(ctx context.Context) func() (int, error) {
	return func() (int, error) {
		api.VotingCycle(func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
			// skip if it does not reach the end time or current phase is TIE
			api.votePhaseChecker.RLock()
			if api.votePhaseChecker.EndTime.After(time.Now()) ||
				api.votePhaseChecker.Phase == VotePhaseHold ||
				api.votePhaseChecker.Phase == VotePhaseNextVoteWin {
				api.votePhaseChecker.RUnlock()
				return
			}
			api.votePhaseChecker.RUnlock()

			switch api.votePhaseChecker.Phase {
			// at the end of wait mech intro
			case VotePhaseWaitMechIntro:

				// get random ability collection set
				battleAbility, factionAbilityMap, err := api.BattleArena.RandomBattleAbility()
				if err != nil {
					api.Log.Err(err)
				}

				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

				// initialise new ability collection
				va.BattleAbility = battleAbility

				// initialise new game ability map
				for fid, ability := range factionAbilityMap {
					va.FactionAbilityMap[fid] = ability
				}

				// start vote ticker
				api.votePhaseChecker.Lock()
				api.votePhaseChecker.Phase = VotePhaseVoteCooldown
				api.votePhaseChecker.EndTime = time.Now().Add(time.Duration(battleAbility.CooldownDurationSecond) * time.Second)
				api.votePhaseChecker.Unlock()

				// broadcast current stage to faction users
				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

			// at the end of ability right voting
			case VotePhaseVoteAbilityRight:
				// stop broadcaster when the vote right is done
				if vct.AbilityRightResultBroadcaster.NextTick != nil {
					vct.AbilityRightResultBroadcaster.Stop()
				}

				fts.Lock()
				// if no vote, enter next vote win phase
				if len(fts.Transactions) == 0 {
					api.votePhaseChecker.Lock()
					api.votePhaseChecker.Phase = VotePhaseNextVoteWin
					api.votePhaseChecker.Unlock()
					// broadcast current stage to faction users
					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

					// stop ticker
					if vct.VotingStageListener.NextTick != nil {
						vct.VotingStageListener.Stop()
					}
					fts.Unlock()
					return
				}

				fts.Transactions = []string{}
				fts.Unlock()

				// HACK: tell user enter location select stage, while committing transactions
				// commit process may take a noticeale time, so user won't fell the vote system freeze
				api.votePhaseChecker.Lock()
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				api.votePhaseChecker.Unlock()
				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

				// parse ability vote result
				type voter struct {
					id         server.UserID
					totalVotes int64
				}
				type factionVote struct {
					factionID  server.FactionID
					totalVotes int64
					voters     []*voter
				}

				factionVotes := []*factionVote{}
				for factionID, factionUserVote := range fuvm {
					// record faction vote
					factionVote := &factionVote{
						factionID:  factionID,
						totalVotes: 0,
						voters:     []*voter{},
					}

					for userID, totalVotes := range factionUserVote {
						// add user to user vote map
						if _, ok := uvm[userID]; !ok {
							uvm[userID] = totalVotes
						}

						// add total vote to faction vote
						factionVote.totalVotes += totalVotes

						// append voter to faction vote
						factionVote.voters = append(factionVote.voters, &voter{
							id:         userID,
							totalVotes: totalVotes,
						})
					}

					// if no vote skip current faction
					if factionVote.totalVotes == 0 {
						continue
					}

					// append current result to faction vote list
					factionVotes = append(factionVotes, factionVote)
				}

				// enter next vote win phase, there is no valid transaction
				if len(factionVotes) == 0 {
					api.votePhaseChecker.Lock()
					api.votePhaseChecker.Phase = VotePhaseNextVoteWin
					api.votePhaseChecker.Unlock()
					// broadcast current stage to faction users
					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

					// stop ticker
					if vct.VotingStageListener.NextTick != nil {
						vct.VotingStageListener.Stop()
					}
					return
				}

				// sort faction votes
				sort.Slice(factionVotes, func(i, j int) bool {
					return factionVotes[i].totalVotes > factionVotes[j].totalVotes
				})

				// sort voters
				sort.Slice(factionVotes[0].voters, func(i, j int) bool {
					return factionVotes[0].voters[i].totalVotes > factionVotes[0].voters[j].totalVotes
				})

				// append voter to list
				for _, v := range factionVotes[0].voters {
					vw.List = append(vw.List, v.id)
				}

				// unicast to the winner
				hcd, winnerClientID := api.getNextWinnerDetail(vw)
				if hcd == nil {
					// if no winner left, enter cooldown phase
					go api.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
						Type:    LocationSelectTypeCancelledNoPlayer,
						Ability: va.BattleAbility.Brief(),
					})

					// voting phase change
					api.votePhaseChecker.Lock()
					api.votePhaseChecker.Phase = VotePhaseVoteCooldown
					api.votePhaseChecker.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)
					api.votePhaseChecker.Unlock()

					// stop vote price update when cooldown
					if api.votePriceSystem.VotePriceUpdater.NextTick != nil {
						api.votePriceSystem.VotePriceUpdater.Stop()
					}

					// broadcast current stage to faction users
					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

					return
				}

				// voting phase change
				api.votePhaseChecker.Lock()
				endTime := time.Now().Add(LocationSelectDurationSecond * time.Second)
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				api.votePhaseChecker.EndTime = endTime
				api.votePhaseChecker.Unlock()

				go api.BroadcastGameNotificationAbility(GameNotificationTypeBattleAbility, &GameNotificationAbility{
					User:    hcd.Brief(),
					Ability: va.FactionAbilityMap[hcd.FactionID].Brief(),
				})

				// announce winner
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, winnerClientID)), &WinnerSelectAbilityLocation{
					GameAbility: va.FactionAbilityMap[hcd.FactionID],
					EndTime:     endTime,
				})

				// broadcast current stage to faction users
				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

			// at the end of location select
			case VotePhaseLocationSelect:
				currentUser, err := api.UserMap.GetUserDetailByID(vw.List[0])
				if err != nil {
					api.Log.Err(err).Msg("failed to get user")
				}
				// pop out the first user of the list
				if len(vw.List) > 1 {
					vw.List = vw.List[1:]
				} else {
					vw.List = []server.UserID{}
				}

				// get next winner
				nextUser, winnerClientID := api.getNextWinnerDetail(vw)
				if nextUser == nil {
					// if no winner left, enter cooldown phase
					go api.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
						Type:    LocationSelectTypeCancelledNoPlayer,
						Ability: va.BattleAbility.Brief(),
					})

					// get random ability collection set
					battleAbility, factionAbilityMap, err := api.BattleArena.RandomBattleAbility()
					if err != nil {
						api.Log.Err(err)
					}

					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

					// initialise new ability collection
					va.BattleAbility = battleAbility

					// initialise new game ability map
					for fid, ability := range factionAbilityMap {
						va.FactionAbilityMap[fid] = ability
					}

					// voting phase change
					api.votePhaseChecker.Lock()
					api.votePhaseChecker.Phase = VotePhaseVoteCooldown
					api.votePhaseChecker.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)
					api.votePhaseChecker.Unlock()

					// stop vote price update in when cooldown
					if api.votePriceSystem.VotePriceUpdater.NextTick != nil {
						api.votePriceSystem.VotePriceUpdater.Stop()
					}

					// broadcast current stage to faction users
					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

					return
				}

				// otherwise, choose next winner
				api.votePhaseChecker.Lock()
				endTime := time.Now().Add(LocationSelectDurationSecond * time.Second)
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				api.votePhaseChecker.EndTime = endTime
				api.votePhaseChecker.Unlock()

				// otherwise announce another winner
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, winnerClientID)), &WinnerSelectAbilityLocation{
					GameAbility: va.FactionAbilityMap[nextUser.FactionID],
					EndTime:     endTime,
				})

				// broadcast winner select location
				go api.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
					Type:        LocationSelectTypeFailedTimeout,
					Ability:     va.BattleAbility.Brief(),
					CurrentUser: currentUser.Brief(),
					NextUser:    nextUser.Brief(),
				})

				// broadcast current stage to faction users
				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

			// at the end of cooldown
			case VotePhaseVoteCooldown:
				// initialise faction user vote map
				for _, fuv := range fuvm {
					for uid := range fuv {
						delete(fuv, uid)
					}
				}

				// initialise faction total vote
				ftv.RedMountainTotalVote = 0
				ftv.BostonTotalVote = 0
				ftv.ZaibatsuTotalVote = 0

				// initialise vote winner
				vw.List = []server.UserID{}

				api.votePhaseChecker.Lock()
				api.votePhaseChecker.Phase = VotePhaseVoteAbilityRight
				api.votePhaseChecker.EndTime = time.Now().Add(VoteAbilityRightDurationSecond * time.Second)
				api.votePhaseChecker.Unlock()

				// broadcast current stage to faction users
				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

				// start tracking vote right result
				if vct.AbilityRightResultBroadcaster.NextTick == nil || vct.AbilityRightResultBroadcaster.NextTick.Before(time.Now()) {
					vct.AbilityRightResultBroadcaster.Start()
				}

				if api.votePriceSystem.VotePriceUpdater.NextTick == nil || api.votePriceSystem.VotePriceUpdater.NextTick.Before(time.Now()) {
					api.votePriceSystem.VotePriceUpdater.Start()
				}

			}
		})

		return http.StatusOK, nil
	}
}

// getNextWinnerDetail get next winner detail from vote winner list
func (api *API) getNextWinnerDetail(vw *VoteWinner) (*server.User, server.UserID) {
	for len(vw.List) > 0 {
		winnerClientID := vw.List[0]
		// broadcast winner notification
		hubClientDetail, err := api.UserMap.GetUserDetailByID(winnerClientID)
		if err != nil {
			// pop out current user, if the user is not online
			if len(vw.List) > 1 {
				vw.List = vw.List[1:]
			} else {
				vw.List = []server.UserID{}
			}
			continue
		}

		return hubClientDetail, winnerClientID
	}

	return nil, server.UserID{}
}
