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
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
)

/***************
* Spoil of War *
***************/
func (api *API) startSpoilOfWarBroadcaster() {
	spoilOfWarBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Spoil of War Broadcaster").Level(zerolog.Disabled)
	spoilOfWarBroadcaster := tickle.New("Spoil of War Broadcaster", 5, func() (int, error) {

		amount, err := api.Passport.GetSpoilOfWarAmount(context.Background())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}

		// prepare payload
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeSpoilOfWarTick))
		payload = append(payload, []byte(amount)...)

		api.Hub.Clients(func(clients hub.ClientsList) {
			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					err := c.SendWithMessageType(payload, websocket.MessageBinary)
					if err != nil {
						api.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})

		return http.StatusOK, nil
	})

	spoilOfWarBroadcaster.Log = &spoilOfWarBroadcasterLogger
	spoilOfWarBroadcaster.Start()
}

/********************
* Vote Price System *
********************/

const VotePriceMultiplierPercentage = 10 // 10%
const VotePriceUpdaterTickSecond = 10

const VotePriceAccuracy = 10000

func (api *API) startVotePriceSystem(factions []*server.Faction, conn *pgxpool.Pool) {
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
	for _, faction := range factions {
		factionVotePrice := &FactionVotePrice{
			OuterLock:            sync.Mutex{},
			NextAccessLock:       sync.Mutex{},
			DataLock:             sync.Mutex{},
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
	votePriceTickerLogger := log_helpers.NamedLogger(api.Log, "Vote Price Ticker").Level(zerolog.TraceLevel)
	votePriceUpdater := tickle.New("Vote Price Ticker", VotePriceUpdaterTickSecond, api.votePriceUpdaterFactory(conn))
	votePriceUpdater.Log = &votePriceTickerLogger

	// initialise vote price forecaster
	votePriceForecasterLogger := log_helpers.NamedLogger(api.Log, "Vote Price Ticker").Level(zerolog.Disabled)
	votePriceForecaster := tickle.New("Vote Price Ticker", 0.5, api.votePriceForecaster)
	votePriceForecaster.Log = &votePriceForecasterLogger

	api.votePriceSystem.VotePriceUpdater = votePriceUpdater
	api.votePriceSystem.VotePriceForecaster = votePriceForecaster
}

func absoluteInt64(x int64) int64 {
	if x < 0 {
		return -x
	}

	return x
}

// votePriceHighPriorityLock for vote price system to lock all the faction vote price lock
func (api *API) votePriceHighPriorityLock() {
	var wg sync.WaitGroup

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

func (api *API) votePriceUpdaterFactory(conn *pgxpool.Pool) func() (int, error) {
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
		var wg sync.WaitGroup
		votePriceMutex := sync.Mutex{}
		factionVotePriceMap := make(map[server.FactionID][]byte)

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

				votePriceMutex.Lock()
				factionVotePriceMap[factionID] = payload
				votePriceMutex.Unlock()

				wg.Done()
			}(factionID, fvp)
		}
		wg.Wait()
		api.votePriceHighPriorityUnlock()

		// start broadcast price
		api.Hub.Clients(func(clients hub.ClientsList) {
			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					// get user faction id
					hcd, err := api.getClientDetailFromChannel(c)

					// skip, if error or no faction
					if err != nil || hcd.FactionID.IsNil() {
						return
					}

					// broadcast vote price forecast
					err = c.SendWithMessageType(factionVotePriceMap[hcd.FactionID], websocket.MessageBinary)
					if err != nil {
						api.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})

		return http.StatusOK, nil
	}
}

// votePriceForecaster
func (api *API) votePriceForecaster() (int, error) {
	// get snap shot of current value
	globalFirstTick := api.votePriceSystem.GlobalVotePerTick[0]
	globalTotalVote := api.votePriceSystem.GlobalTotalVote

	currentTotalVote := int64(0)
	factionVoteMap := make(map[server.FactionID]int64)
	for _, faction := range api.factionMap {
		cvpt := api.votePriceSystem.FactionVotePriceMap[faction.ID].CurrentVotePerTick
		factionVoteMap[faction.ID] = cvpt
		currentTotalVote += cvpt
	}

	// calculate total vote
	globalTotalVote = globalTotalVote - globalFirstTick + currentTotalVote

	var wg sync.WaitGroup
	votePriceMutex := sync.Mutex{}
	factionVotePriceMap := make(map[server.FactionID][]byte)

	// start calc faction vote price
	for _, faction := range api.factionMap {
		wg.Add(1)
		go func(faction *server.Faction) {

			// get a copy of current vote price
			newVotePrice := calVotePrice(
				globalTotalVote,
				api.votePriceSystem.FactionVotePriceMap[faction.ID].CurrentVotePriceSups,
				factionVoteMap[faction.ID],
				faction.ID,
			)

			// prepare broadcast payload
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeVotePriceForecastTick))
			payload = append(payload, []byte(newVotePrice.Int.String())...)

			votePriceMutex.Lock()
			factionVotePriceMap[faction.ID] = payload
			votePriceMutex.Unlock()
			wg.Done()
		}(faction)
	}
	wg.Wait()

	// start broadcast
	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok {
				continue
			}
			go func(c *hub.Client) {
				// get user faction id
				hcd, err := api.getClientDetailFromChannel(c)

				// skip, if error or no faction
				if err != nil || hcd.FactionID.IsNil() {
					return
				}

				// broadcast vote price forecast
				err = c.SendWithMessageType(factionVotePriceMap[hcd.FactionID], websocket.MessageBinary)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
		}
	})

	return http.StatusOK, nil
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
			priceChange.Mul(&priceChange.Int, big.NewInt(2))
		}

		votePriceSups.Add(&votePriceSups.Int, &priceChange.Int)
	} else {
		// price go down
		if votePriceSups.Cmp(big.NewInt(1000000000)) < 0 { // price floor
			priceChange = server.BigInt{Int: *big.NewInt(0)}
		} else if votePriceSups.Cmp(big.NewInt(1000000000000)) < 0 {
			priceChange.Div(&priceChange.Int, big.NewInt(4))
		} else if votePriceSups.Cmp(big.NewInt(1000000000000000000)) < 0 {
			priceChange.Div(&priceChange.Int, big.NewInt(2))
		}

		votePriceSups.Sub(&votePriceSups.Int, &priceChange.Int)
	}

	return votePriceSups
}

/************************
* Live Voting Broadcast *
************************/

type LiveVotingData struct {
	TotalVote server.BigInt
}

func (api *API) startLiveVotingDataTicker(factionID server.FactionID) {
	// live voting data broadcast
	liveSupsSpend := &LiveVotingData{
		TotalVote: server.BigInt{Int: *big.NewInt(0)},
	}

	// start channel
	go func() {
		for fn := range api.liveSupsSpend[factionID] {
			fn(liveSupsSpend)
		}
	}()
}

/***************
* Voting Cycle *
***************/

const (
	// CooldownInitialDurationSecond the amount of second users have to wait for the next vote coming up
	CooldownInitialDurationSecond = 15

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
	Phase VotePhase
}

type VoteStage struct {
	Phase   VotePhase `json:"phase"`
	EndTime time.Time `json:"endTime"`
}

type VoteAbility struct {
	BattleAbility     *server.BattleAbility
	FactionAbilityMap map[server.FactionID]*server.FactionAbility
}

type FactionUserVoteMap map[server.FactionID]map[server.UserID]map[server.TransactionReference]int64

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
	FactionAbility server.FactionAbility `json:"factionAbility"`
	EndTime        time.Time             `json:"endTime"`
}

/***********************
* Voting Cycle Channel *
***********************/

// StartVotingCycle start voting cycle ticker
func (api *API) StartVotingCycle(factions []*server.Faction) {
	// initialise current vote stage
	api.votePhaseChecker = &VotePhaseChecker{
		Phase: VotePhaseHold,
	}
	voteStage := &VoteStage{
		Phase:   VotePhaseHold,
		EndTime: time.Now(),
	}

	// initialise vote ability
	voteAbility := &VoteAbility{
		BattleAbility:     &server.BattleAbility{},
		FactionAbilityMap: make(map[server.FactionID]*server.FactionAbility),
	}

	// initial faction user voting map
	factionUserVoteMap := make(FactionUserVoteMap)
	for _, f := range factions {
		factionUserVoteMap[f.ID] = make(map[server.UserID]map[server.TransactionReference]int64)
		voteAbility.FactionAbilityMap[f.ID] = &server.FactionAbility{}
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

	// start faction voting cycle tickle
	tickle.MinDurationOverride = true
	voteStageLogger := log_helpers.NamedLogger(api.Log, "Voting Cycle Tracker").Level(zerolog.Disabled)
	voteStageListener := tickle.New("Voting Cycle Tracker", 1, api.voteStageListenerFactory())
	voteStageListener.Log = &voteStageLogger

	// start faction voting cycle tickle
	abilityRightResultLogger := log_helpers.NamedLogger(api.Log, "Ability Right Result Broadcaster").Level(zerolog.Disabled)
	abilityRightResultBroadcaster := tickle.New("Ability Right Result Broadcaster", 0.5, api.abilityRightResultBroadcasterFactory(factionTotalVote))
	abilityRightResultBroadcaster.Log = &abilityRightResultLogger

	tickers := &VotingCycleTicker{
		VotingStageListener:           voteStageListener,
		AbilityRightResultBroadcaster: abilityRightResultBroadcaster,
	}

	// start channel
	go func() {
		for fn := range api.votingCycle {
			fn(voteStage, voteAbility, factionUserVoteMap, factionTotalVote, voteWinner, tickers)
		}
	}()
}

/******************************
* Ability Right Result Ticker *
******************************/

// abilityRightResultBroadcasterFactory generate the function for broadcasting the ability right result
func (api *API) abilityRightResultBroadcasterFactory(ftv *FactionTotalVote) func() (int, error) {
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

		// broadcast back to client
		api.Hub.Clients(func(clients hub.ClientsList) {
			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					err := c.SendWithMessageType(payload, websocket.MessageBinary)
					if err != nil {
						api.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})

		return http.StatusOK, nil
	}
}

/**********************
* Voting Stage Ticker *
**********************/

// startVotingCycle start voting cycle tickles
func (api *API) startVotingCycle(introSecond int) {
	api.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
		api.votePhaseChecker.Phase = VotePhaseWaitMechIntro
		vs.Phase = VotePhaseWaitMechIntro
		vs.EndTime = time.Now().Add(time.Duration(introSecond) * time.Second)

		// broadcast current stage to faction users
		api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

		if vct.VotingStageListener.NextTick == nil || vct.VotingStageListener.NextTick.Before(time.Now()) {
			vct.VotingStageListener.Start()
		}
	}
}

// stopVotingCycle pause voting cycle tickles
func (api *API) stopVotingCycle() {
	api.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
		vs.Phase = VotePhaseHold
		// broadcast current stage to faction users
		api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

		if vct.VotingStageListener.NextTick != nil {
			vct.VotingStageListener.Stop()
		}

		if vct.AbilityRightResultBroadcaster.NextTick != nil {
			vct.AbilityRightResultBroadcaster.Stop()
		}

		if api.votePriceSystem.VotePriceUpdater.NextTick != nil {
			api.votePriceSystem.VotePriceUpdater.Stop()
		}

		if api.votePriceSystem.VotePriceForecaster.NextTick != nil {
			api.votePriceSystem.VotePriceForecaster.Stop()
		}

		// get all the left over transaction
		var txRefs []server.TransactionReference
		for _, factionVotes := range fuvm {
			for _, userVotes := range factionVotes {
				for txRef := range userVotes {
					txRefs = append(txRefs, txRef)
				}
			}
		}

		// commit the transactions
		if len(txRefs) > 0 {
			_, err := api.Passport.ReleaseTransactions(context.Background(), txRefs)
			if err != nil {
				api.Log.Err(err).Msg("failed to Release transactions")
				return
			}
		}
	}
}

// voteStageListenerFactory is the main vote stage handler
func (api *API) voteStageListenerFactory() func() (int, error) {
	return func() (int, error) {
		api.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker) {
			ctx := context.Background()
			// skip if it does not reach the end time or current phase is TIE
			if vs.EndTime.After(time.Now()) || vs.Phase == VotePhaseHold || vs.Phase == VotePhaseNextVoteWin {
				return
			}

			switch vs.Phase {
			// at the end of wait mech intro
			case VotePhaseWaitMechIntro:

				// get random ability collection set
				battleAbility, factionAbilityMap, err := api.BattleArena.RandomAbilityCollection()
				if err != nil {
					api.Log.Err(err)
				}

				api.MessageBus.Send(messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

				// initialise new ability collection
				va.BattleAbility = battleAbility

				// initialise new faction ability map
				for fid, ability := range factionAbilityMap {
					va.FactionAbilityMap[fid] = ability
				}

				// start vote ticker
				api.votePhaseChecker.Phase = VotePhaseVoteCooldown

				vs.Phase = VotePhaseVoteCooldown
				vs.EndTime = time.Now().Add(CooldownInitialDurationSecond * time.Second)

				// broadcast current stage to faction users
				api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

			// at the end of ability right voting
			case VotePhaseVoteAbilityRight:
				if api.votePriceSystem.VotePriceUpdater.NextTick != nil {
					api.votePriceSystem.VotePriceUpdater.Stop()
				}

				if api.votePriceSystem.VotePriceForecaster.NextTick != nil {
					api.votePriceSystem.VotePriceForecaster.Stop()
				}

				// get all the tx
				var txRefs []server.TransactionReference
				for _, factionVotes := range fuvm {
					for _, userVotes := range factionVotes {
						for txRef := range userVotes {
							txRefs = append(txRefs, txRef)
						}
					}
				}

				// if no vote, enter next vote win phase
				if len(txRefs) == 0 {
					api.votePhaseChecker.Phase = VotePhaseNextVoteWin
					vs.Phase = VotePhaseNextVoteWin
					// broadcast current stage to faction users
					api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

					// stop ticker
					if vct.VotingStageListener.NextTick != nil {
						vct.VotingStageListener.Stop()
					}
					return
				}

				// otherwise, commit the transactions and check the status
				checkedTransactions, err := api.Passport.CommitTransactions(ctx, txRefs)
				if err != nil {
					api.Log.Err(err).Msg("failed to check transactions")
					return
				}

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

					for userID, userVotes := range factionUserVote {
						// record voter
						voter := &voter{
							id:         userID,
							totalVotes: 0,
						}

						// sum total successful vote
						for txRef, voteCount := range userVotes {
							for _, chktx := range checkedTransactions {
								if txRef == chktx.TransactionReference && chktx.Status == server.TransactionSuccess {
									factionVote.totalVotes += voteCount
									voter.totalVotes += voteCount
									continue
								}
							}
						}

						// append voter to faction vote
						factionVote.voters = append(factionVote.voters, voter)
					}

					// if no vote skip current faction
					if factionVote.totalVotes == 0 {
						continue
					}

					// append current result to faction vote list
					factionVotes = append(factionVotes, factionVote)
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
						Type: LocationSelectTypeCancelled,
						Ability: &AbilityBrief{
							Label:    va.BattleAbility.Label,
							ImageUrl: va.BattleAbility.ImageUrl,
						},
						Reason: "NO_PLAYER_SELECT_LOCATION",
					})

					// voting phase change
					api.votePhaseChecker.Phase = VotePhaseVoteCooldown
					vs.Phase = VotePhaseVoteCooldown
					vs.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)

					// broadcast current stage to faction users
					api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

					return
				}

				// voting phase change
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				vs.Phase = VotePhaseLocationSelect
				vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

				go api.BroadcastGameNotificationAbility(GameNotificationTypeBattleAbility, &GameNotificationAbility{
					User: &UserBrief{
						Username: hcd.Username,
						AvatarID: hcd.avatarID,
						Faction: &FactionBrief{
							Label:      api.factionMap[hcd.FactionID].Label,
							Theme:      api.factionMap[hcd.FactionID].Theme,
							LogoBlobID: api.factionMap[hcd.FactionID].LogoBlobID,
						},
					},
				})

				// announce winner
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, winnerClientID)), &WinnerSelectAbilityLocation{
					FactionAbility: *va.FactionAbilityMap[hcd.FactionID],
					EndTime:        vs.EndTime,
				})

				// broadcast current stage to faction users
				api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

				// stop broadcaster when the vote right is done
				if vct.AbilityRightResultBroadcaster.NextTick != nil {
					vct.AbilityRightResultBroadcaster.Stop()
				}

			// at the end of location select
			case VotePhaseLocationSelect:
				currentUser, err := api.getClientDetailFromUserID(vw.List[0])
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
						Type: LocationSelectTypeCancelled,
						Ability: &AbilityBrief{
							Label:    va.BattleAbility.Label,
							ImageUrl: va.BattleAbility.ImageUrl,
						},
						Reason: "NO_PLAYER",
					})

					// get random ability collection set
					battleAbility, factionAbilityMap, err := api.BattleArena.RandomAbilityCollection()
					if err != nil {
						api.Log.Err(err)
					}

					api.MessageBus.Send(messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

					// initialise new ability collection
					va.BattleAbility = battleAbility

					// initialise new faction ability map
					for fid, ability := range factionAbilityMap {
						va.FactionAbilityMap[fid] = ability
					}

					// voting phase change
					api.votePhaseChecker.Phase = VotePhaseVoteCooldown
					vs.Phase = VotePhaseVoteCooldown
					vs.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)

					// broadcast current stage to faction users
					api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

					return
				}

				// otherwise, choose next winner
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				vs.Phase = VotePhaseLocationSelect
				vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

				// otherwise announce another winner
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, winnerClientID)), &WinnerSelectAbilityLocation{
					FactionAbility: *va.FactionAbilityMap[nextUser.FactionID],
					EndTime:        vs.EndTime,
				})

				// broadcast winner select location
				go api.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
					Type: LocationSelectTypeFailed,
					Ability: &AbilityBrief{
						Label:    va.BattleAbility.Label,
						ImageUrl: va.BattleAbility.ImageUrl,
					},
					Reason: "TIMEOUT",
					CurrentUser: &UserBrief{
						Username: currentUser.Username,
						AvatarID: currentUser.avatarID,
						Faction: &FactionBrief{
							Label:      api.factionMap[currentUser.FactionID].Label,
							Theme:      api.factionMap[currentUser.FactionID].Theme,
							LogoBlobID: api.factionMap[currentUser.FactionID].LogoBlobID,
						},
					},
					NextUser: &UserBrief{
						Username: nextUser.Username,
						AvatarID: nextUser.avatarID,
						Faction: &FactionBrief{
							Label:      api.factionMap[nextUser.FactionID].Label,
							Theme:      api.factionMap[nextUser.FactionID].Theme,
							LogoBlobID: api.factionMap[nextUser.FactionID].LogoBlobID,
						},
					},
				})

				// broadcast current stage to faction users
				api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

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

				api.votePhaseChecker.Phase = VotePhaseVoteAbilityRight
				vs.Phase = VotePhaseVoteAbilityRight
				vs.EndTime = time.Now().Add(VoteAbilityRightDurationSecond * time.Second)

				// broadcast current stage to faction users
				api.MessageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), vs)

				// start tracking vote right result
				if vct.AbilityRightResultBroadcaster.NextTick == nil || vct.AbilityRightResultBroadcaster.NextTick.Before(time.Now()) {
					vct.AbilityRightResultBroadcaster.Start()
				}

				if api.votePriceSystem.VotePriceUpdater.NextTick == nil || api.votePriceSystem.VotePriceUpdater.NextTick.Before(time.Now()) {
					api.votePriceSystem.VotePriceUpdater.Start()
				}

				if api.votePriceSystem.VotePriceForecaster.NextTick == nil || api.votePriceSystem.VotePriceForecaster.NextTick.Before(time.Now()) {
					api.votePriceSystem.VotePriceForecaster.Start()
				}

			}
		}

		return http.StatusOK, nil
	}
}

// getNextWinnerDetail get next winner detail from vote winner list
func (api *API) getNextWinnerDetail(vw *VoteWinner) (*HubClientDetail, server.UserID) {
	for len(vw.List) > 0 {
		winnerClientID := vw.List[0]
		// broadcast winner notification
		hubClientDetail, err := api.getClientDetailFromUserID(winnerClientID)
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
