package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/passport"
	"sort"
	"strings"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
)

//******************************
// Game Ability setup
//******************************

const EachMechIntroSecond = 1
const InitIntroSecond = 1

type LocationDeciders struct {
	deadlock.RWMutex
	list []server.UserID
}

type AbilitiesSystem struct {
	battle *Battle
	// faction unique abilities
	factionUniqueAbilities map[uuid.UUID]map[server.GameAbilityID]*GameAbility // map[faction_id]map[identity]*Ability

	// gabs abilities (air craft, nuke, repair)
	battleAbilityPool *BattleAbilityPool

	// track the sups contribution of each user, use for location select
	userContributeMap map[uuid.UUID]*UserContribution

	// location select winner list
	locationDeciders *LocationDeciders
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	factionAbilities := map[uuid.UUID]map[server.GameAbilityID]*GameAbility{}

	// initialise new gabs ability pool
	battleAbilityPool := &BattleAbilityPool{
		Stage: &GabsBribeStage{
			Phase:   BribeStageHold,
			EndTime: time.Now().AddDate(1, 0, 0), // HACK: set end time to far future to implement infinite time
		},
		BattleAbilityMx: deadlock.RWMutex{},
		BattleAbility:   &server.BattleAbility{},
		Abilities:       map[uuid.UUID]*GameAbility{},
	}

	userContributeMap := map[uuid.UUID]*UserContribution{}

	for factionID := range battle.factions {
		// initialise faction unique abilities
		factionAbilities[factionID] = map[server.GameAbilityID]*GameAbility{}

		// faction unique abilities
		factionUniqueAbilities, err := boiler.GameAbilities(qm.Where("faction_id = ?", factionID.String()), qm.And("battle_ability_id ISNULL")).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("Battle ID", battle.ID.String()).Err(err).Msg("unable to retrieve game abilities")
		}

		// for zaibatsu unique abilities
		if factionID.String() == server.ZaibatsuFactionID.String() {

			for _, ability := range factionUniqueAbilities {
				for i, wm := range battle.WarMachines {
					// skip if mech is not zaibatsu mech
					if wm.FactionID != factionID.String() {
						continue
					}

					supsCost, err := decimal.NewFromString(ability.SupsCost)
					if err != nil {
						gamelog.L.Err(err).Msg("Failed to ability sups cost to decimal")

						// set sups cost to initial price
						supsCost = decimal.New(100, 18)
					}

					currentSups, err := decimal.NewFromString(ability.CurrentSups)
					if err != nil {
						gamelog.L.Err(err).Msg("Failed to ability current sups to decimal")

						// set current sups to initial price
						currentSups = decimal.Zero
					}

					// build the ability
					wmAbility := &GameAbility{
						ID:                  server.GameAbilityID(uuid.Must(uuid.FromString(ability.ID))), // generate a uuid for frontend to track sups contribution
						GameClientAbilityID: byte(ability.GameClientAbilityID),
						ImageUrl:            ability.ImageURL,
						Description:         ability.Description,
						FactionID:           factionID,
						Label:               ability.Label,
						SupsCost:            supsCost,
						CurrentSups:         currentSups,
						WarMachineHash:      wm.Hash,
						ParticipantID:       &wm.ParticipantID,
						Title:               wm.Name,
						Colour:              ability.Colour,
						TextColour:          ability.TextColour,
					}

					// inject ability to war machines
					battle.WarMachines[i].Abilities = []*GameAbility{wmAbility}

					// store faction ability for price tracking
					factionAbilities[factionID][wmAbility.ID] = wmAbility
				}
			}

		} else {

			// for other faction unique abilities
			abilities := map[server.GameAbilityID]*GameAbility{}
			for _, ability := range factionUniqueAbilities {

				supsCost, err := decimal.NewFromString(ability.SupsCost)
				if err != nil {
					gamelog.L.Err(err).Msg("Failed to ability sups cost to decimal")

					// set sups cost to initial price
					supsCost = decimal.New(100, 18)
				}

				currentSups, err := decimal.NewFromString(ability.CurrentSups)
				if err != nil {
					gamelog.L.Err(err).Msg("Failed to ability current sups to decimal")

					// set current sups to initial price
					currentSups = decimal.Zero
				}

				wmAbility := &GameAbility{
					ID:                  server.GameAbilityID(uuid.Must(uuid.FromString(ability.ID))), // generate a uuid for frontend to track sups contribution
					GameClientAbilityID: byte(ability.GameClientAbilityID),
					ImageUrl:            ability.ImageURL,
					Description:         ability.Description,
					FactionID:           factionID,
					Label:               ability.Label,
					SupsCost:            supsCost,
					CurrentSups:         currentSups,
					Colour:              ability.Colour,
					TextColour:          ability.TextColour,
					Title:               "FACTION_WIDE",
				}
				abilities[wmAbility.ID] = wmAbility
			}
			factionAbilities[factionID] = abilities
		}

		// initialise user vote map in gab ability pool
		userContributeMap[factionID] = &UserContribution{
			contributionMap: map[server.UserID]decimal.Decimal{},
		}
	}

	as := &AbilitiesSystem{
		battle:                 battle,
		factionUniqueAbilities: factionAbilities,
		battleAbilityPool:      battleAbilityPool,
		userContributeMap:      userContributeMap,
		locationDeciders: &LocationDeciders{
			list: []server.UserID{},
		},
	}

	// init battle ability
	_, err := as.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to set up battle ability")
		return nil
	}

	// calc the intro time, mech_amount * 3 + 7 second
	waitDurationSecond := len(battle.WarMachines)*EachMechIntroSecond + InitIntroSecond

	// start ability cycle
	go as.FactionUniqueAbilityUpdater(waitDurationSecond)

	// bribe cycle
	go as.StartGabsAbilityPoolCycle(waitDurationSecond)

	return as
}

// ***********************************
// Faction Unique Ability Contribution
// ***********************************

// FactionUniqueAbilityUpdater update ability price every 10 seconds
func (as *AbilitiesSystem) FactionUniqueAbilityUpdater(waitDurationSecond int) {
	// wait for mech intro
	time.Sleep(time.Duration(waitDurationSecond) * time.Second)

	minPrice := decimal.New(1, 18)

	// start the battle
	for _, abilities := range as.factionUniqueAbilities {

		// start ability price updater for each faction
		go func(abilities map[server.GameAbilityID]*GameAbility) {
			for {
				// read the stage first
				as.battle.Stage.RLock()
				stage := as.battle.Stage.Stage
				as.battle.Stage.RUnlock()

				// start ticker while still in battle
				if stage == BattleStagStart {
					for _, ability := range abilities {
						// lock current ability price update
						ability.Lock()

						// check battle stage before reduce update ability price
						as.battle.Stage.RLock()
						if as.battle.Stage.Stage == BattleStageEnd {
							as.battle.Stage.RUnlock()
							ability.Unlock()
							return
						}
						as.battle.Stage.RUnlock()

						// update ability price
						if ability.TargetPriceUpdate(minPrice) {
							// send message to game client, if ability trigger
							as.battle.arena.Message(
								"BATTLE:ABILITY",
								&server.GameAbilityEvent{
									IsTriggered:         true,
									GameClientAbilityID: ability.GameClientAbilityID,
									ParticipantID:       ability.ParticipantID, // trigger on war machine
								},
							)

						}
						ability.Unlock()

					}

					time.Sleep(10 * time.Second)
					continue
				}

				break
			}

			// terminate the function when battle is end
			fmt.Println("Exit battle price updater")
			// do something after battle end...

		}(abilities)

	}
}

// TargetPriceUpdate update target price on every tick
func (ga *GameAbility) TargetPriceUpdate(minPrice decimal.Decimal) bool {
	ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromFloat(0.9772))

	// if target price hit 1 sup, set it to 1 sup
	if ga.SupsCost.Cmp(minPrice) <= 0 {
		ga.SupsCost = decimal.New(1, 18)
	}

	isTriggered := false

	// if the target price hit current price
	if ga.SupsCost.Cmp(ga.CurrentSups) <= 0 {
		// trigger the ability
		isTriggered = true

		// double the target price
		ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromInt(2))

		// reset current sups to zero
		ga.CurrentSups = decimal.Zero

	}

	// store updated price to db
	err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
	if err != nil {
		gamelog.L.Err(err)
		return isTriggered
	}

	return isTriggered
}

// FactionUniqueAbilityContribute contribute sups to specific faction unique ability
func (as *AbilitiesSystem) FactionUniqueAbilityContribute(factionID uuid.UUID, abilityID server.GameAbilityID, userID server.UserID, amount decimal.Decimal) {
	// check faction unique ability exist
	if abilities, ok := as.factionUniqueAbilities[factionID]; ok {

		// check ability exists
		if ability, ok := abilities[abilityID]; ok {

			// lock price update
			ability.Lock()
			// check battle stage
			as.battle.Stage.RLock()

			// return early if battle stage is invalid
			if as.battle.Stage.Stage != BattleStagStart {
				as.battle.Stage.RUnlock()
				ability.Unlock()
				return
			}
			as.battle.Stage.RUnlock()

			actualSupSpent, isTriggered := ability.SupContribution(as.battle.arena.ppClient, as.battle.ID.String(), userID, amount)

			// cache user's sup contribution for generating location select order
			as.userContributeMap[factionID].Lock()
			if _, ok := as.userContributeMap[factionID].contributionMap[userID]; !ok {
				as.userContributeMap[factionID].contributionMap[userID] = decimal.Zero
			}
			as.userContributeMap[factionID].contributionMap[userID] = as.userContributeMap[factionID].contributionMap[userID].Add(actualSupSpent)
			as.userContributeMap[factionID].Unlock()

			// sups contribution
			if isTriggered {
				// send message to game client, if ability trigger
				as.battle.arena.Message(
					"BATTLE:ABILITY",
					&server.GameAbilityEvent{
						IsTriggered:         true,
						GameClientAbilityID: ability.GameClientAbilityID,
						ParticipantID:       ability.ParticipantID, // trigger on war machine
					},
				)
			}

			ability.Unlock()
		}
	}
}

// SupContribution contribute sups to specific game ability, return the actual sups spent and whether the ability is triggered
func (ga *GameAbility) SupContribution(ppClient *passport.Passport, battleID string, userID server.UserID, amount decimal.Decimal) (decimal.Decimal, bool) {

	isTriggered := false

	// calc the different
	diff := ga.SupsCost.Sub(ga.CurrentSups)

	// if players spend more thant they need, crop the spend price
	if amount.Cmp(diff) >= 0 {
		isTriggered = true
		amount = diff
	}

	// pay sup
	err := ppClient.SpendSupMessage(passport.SpendSupsReq{
		FromUserID:           userID,
		Amount:               amount.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("ability_sup_contribute|%s", uuid.Must(uuid.NewV4()))),
		Group:                "battle",
		SubGroup:             battleID,
		Description:          "battle vote.",
		NotSafe:              true,
	})
	if err != nil {
		return decimal.Zero, false
	}

	// update the current sups if not triggered
	if !isTriggered {
		ga.CurrentSups = ga.CurrentSups.Add(amount)

		// store updated price to db
		err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
		if err != nil {
			gamelog.L.Err(err)
			return amount, false
		}
		return amount, false
	}

	// otherwise update target price and reset the current price
	ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromInt(2))
	ga.CurrentSups = decimal.Zero

	// store updated price to db
	err = db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
	if err != nil {
		gamelog.L.Err(err)
		return amount, true
	}

	return amount, true
}

// ***************************
// Gabs Abilities Voting Cycle
// ***************************

const (
	// BribeDurationSecond the amount of second players can bribe GABS
	BribeDurationSecond = 30
	// LocationSelectDurationSecond the amount of second the winner user can select the location
	LocationSelectDurationSecond = 15
	// CooldownDurationSecond the amount of second players have to wait for next bribe phase
	CooldownDurationSecond = 20
)

type BribePhase string

const (
	BribeStageBribe          BribePhase = "BRIBE"
	BribeStageLocationSelect BribePhase = "LOCATION_SELECT"
	BribeStageCooldown       BribePhase = "COOLDOWN"
	BribeStageHold           BribePhase = "HOLD"
)

type GabsBribeStage struct {
	deadlock.RWMutex
	Phase   BribePhase `json:"phase"`
	EndTime time.Time  `json:"endTime"`
}

// track user contribution of current battle
type UserContribution struct {
	deadlock.Mutex
	contributionMap map[server.UserID]decimal.Decimal
}

type BattleAbilityPool struct {
	deadlock.Mutex // force bribe process synchronize
	Stage          *GabsBribeStage

	BattleAbilityMx deadlock.RWMutex // lock for ability changes
	BattleAbility   *server.BattleAbility
	Abilities       map[uuid.UUID]*GameAbility // faction ability current, change on every bribing cycle

	TriggeredFactionID uuid.UUID
}

type LocationSelectAnnouncement struct {
	GameAbility *GameAbility `json:"gameAbility"`
	EndTime     time.Time    `json:"endTime"`
}

// StartGabsAbilityPoolCycle
func (as *AbilitiesSystem) StartGabsAbilityPoolCycle(waitDurationSecond int) {
	// wait for mech intro
	time.Sleep(time.Duration(waitDurationSecond) * time.Second)

	// start voting stage
	as.battleAbilityPool.Stage.Lock()
	as.battleAbilityPool.Stage.Phase = BribeStageBribe
	as.battleAbilityPool.Stage.EndTime = time.Now().Add(BribeDurationSecond * time.Second)
	as.battleAbilityPool.Stage.Unlock()

	// progress bar broadcaster
	go as.BattleAbilityProgressBar()
	// ability price updater
	go as.BAttleAbilityPriceUpdater()

	// initial a ticker for current battle
	ticker := time.NewTicker(1 * time.Second)

	// start ability pool cycle
	for {
		// wait for next tick
		<-ticker.C

		// check phase
		as.battle.Stage.RLock()
		stage := as.battle.Stage.Stage
		// exit the loop, when battle is ended
		if stage == BattleStageEnd {
			ticker.Stop()
			as.battle.Stage.RUnlock()
			break
		}
		as.battle.Stage.RUnlock()

		// otherwise check bribing phase
		as.battleAbilityPool.Stage.RLock()

		// skip, if the end time of current phase haven't been reached
		if as.battleAbilityPool.Stage.EndTime.After(time.Now()) {
			as.battleAbilityPool.Stage.RUnlock()
			continue
		}

		// otherwise, read current bribe phase
		bribePhase := as.battleAbilityPool.Stage.Phase

		as.battleAbilityPool.Stage.RUnlock()

		/////////////////
		// Bribe Phase //
		/////////////////
		switch bribePhase {

		// at the end of bribing phase
		// no ability is triggered, switch to cooldown phase
		case BribeStageBribe:
			// change bribing phase
			as.battleAbilityPool.Stage.Lock()

			// set new battle ability
			cooldownSecond, err := as.SetNewBattleAbility()
			if err != nil {
				gamelog.L.Err(err).Msg("Failed to set new battle ability")
			}

			as.battleAbilityPool.Stage.Phase = BribeStageCooldown
			as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
			// broadcast stage to frontend
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
			as.battleAbilityPool.Stage.Unlock()

		// at the end of location select phase
		// pass the location select to next player
		case BribeStageLocationSelect:
			// get the next location decider
			userID, ok := as.nextLocationDeciderGet()
			if !ok {
				// enter cooldown phase, if there is no user left for location select

				// change bribing phase
				as.battleAbilityPool.Stage.Lock()

				// set new battle ability
				cooldownSecond, err := as.SetNewBattleAbility()
				if err != nil {
					gamelog.L.Err(err).Msg("Failed to set new battle ability")
				}

				as.battleAbilityPool.Stage.Phase = BribeStageCooldown
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
				as.battleAbilityPool.Stage.Unlock()
				continue
			}

			// extend location select phase duration
			as.battleAbilityPool.Stage.Lock()
			as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
			as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
			// broadcast stage to frontend
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
			as.battleAbilityPool.Stage.Unlock()

			// broadcast the announcement to the next location decider
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID)), &LocationSelectAnnouncement{
				GameAbility: as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
				EndTime:     as.battleAbilityPool.Stage.EndTime,
			})

		// at the end of cooldown phase
		// random choose a battle ability for next bribing session
		case BribeStageCooldown:

			// change bribing phase
			as.battleAbilityPool.Stage.Lock()
			as.battleAbilityPool.Stage.Phase = BribeStageBribe
			as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(BribeDurationSecond) * time.Second)
			// broadcast stage to frontend
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
			as.battleAbilityPool.Stage.Unlock()

			continue
		}
	}

	// do some thing after battle end...
	fmt.Println("Exit bribing ticker")
}

// SetNewBattleAbility set new battle ability and return the cooldown time
func (as *AbilitiesSystem) SetNewBattleAbility() (int, error) {

	// lock the ability while changing
	as.battleAbilityPool.BattleAbilityMx.Lock()
	defer as.battleAbilityPool.BattleAbilityMx.Unlock()

	// clean up triggered faction
	as.battleAbilityPool.TriggeredFactionID = uuid.Nil

	// initialise new gabs ability pool
	ba, err := db.BattleAbilityGetRandom(context.Background(), gamedb.Conn)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to get battle ability from db")
		return 0, terror.Error(err)
	}
	as.battleAbilityPool.BattleAbility = ba

	// get faction battle abilities
	gabsAbilities, err := db.FactionBattleAbilityGet(context.Background(), gamedb.Conn, ba.ID)
	if err != nil {
		return ba.CooldownDurationSecond, terror.Error(err)
	}

	// set battle abilities of each faction
	for _, ga := range gabsAbilities {
		supsCost, err := decimal.NewFromString(ga.SupsCost)
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to ability sups cost to decimal")

			// set sups cost to initial price
			supsCost = decimal.New(100, 18)
		}

		currentSups, err := decimal.NewFromString(ga.CurrentSups)
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to ability current sups to decimal")

			// set current sups to initial price
			currentSups = decimal.Zero
		}

		// initialise game ability
		gameAbility := &GameAbility{
			ID:                  ga.ID,
			GameClientAbilityID: byte(ga.GameClientAbilityID),
			ImageUrl:            ga.ImageUrl,
			Description:         ga.Description,
			FactionID:           ga.FactionID,
			Label:               ga.Label,
			SupsCost:            supsCost,
			CurrentSups:         currentSups,
		}
		as.battleAbilityPool.Abilities[ga.FactionID] = gameAbility

		// broadcast ability update to faction users
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBattleAbilityUpdated, gameAbility.FactionID.String())), gameAbility)
	}

	return ba.CooldownDurationSecond, nil
}

func (as *AbilitiesSystem) BattleAbilityBribing(factionID uuid.UUID, userID server.UserID, amount decimal.Decimal) error {
	// lock bribe to force the bribe synchronize
	as.battleAbilityPool.Lock()
	defer as.battleAbilityPool.Unlock()

	// check current battle stage
	as.battle.Stage.RLock()
	as.battleAbilityPool.Stage.RLock()
	// return early if battle stage or bribing stage are invalid
	if as.battle.Stage.Stage != BattleStagStart || as.battleAbilityPool.Stage.Phase != BribeStageBribe {
		as.battleAbilityPool.Stage.RUnlock()
		as.battle.Stage.RUnlock()
		return nil
	}
	as.battleAbilityPool.Stage.RUnlock()
	as.battle.Stage.RUnlock()

	// check faction ability exists
	if factionAbility, ok := as.battleAbilityPool.Abilities[uuid.UUID(factionID)]; ok {

		// lock ability
		factionAbility.Lock()
		defer factionAbility.Unlock()

		// contribute sups
		actualSupSpent, abilityTriggered := factionAbility.SupContribution(as.battle.arena.ppClient, as.battle.ID.String(), userID, amount)

		// cache user contribution for location select order
		as.userContributeMap[factionID].Lock()
		if _, ok := as.userContributeMap[factionID].contributionMap[userID]; !ok {
			as.userContributeMap[factionID].contributionMap[userID] = decimal.Zero
		}
		as.userContributeMap[factionID].contributionMap[userID] = as.userContributeMap[factionID].contributionMap[userID].Add(actualSupSpent)
		as.userContributeMap[factionID].Unlock()

		if abilityTriggered {
			// generate location select order list
			as.locationSelectListSet(factionID, userID)

			// change bribing phase to location select
			as.battleAbilityPool.Stage.Lock()
			as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
			as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
			// broadcast stage change
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
			as.battleAbilityPool.Stage.Unlock()

			// send message to the user who trigger the ability
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID)), &LocationSelectAnnouncement{
				GameAbility: as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
				EndTime:     as.battleAbilityPool.Stage.EndTime,
			})
		}
	}

	return nil
}

// locationSelectListSet set a user list for location select for current ability triggered
func (as *AbilitiesSystem) locationSelectListSet(factionID uuid.UUID, triggerByUserID ...server.UserID) {
	// lock the user id list while setting
	as.locationDeciders.Lock()
	defer as.locationDeciders.Unlock()

	// set triggered faction id
	as.battleAbilityPool.TriggeredFactionID = factionID

	type userSupSpent struct {
		userID   server.UserID
		supSpent decimal.Decimal
	}

	list := []*userSupSpent{}
	for userID, contribution := range as.userContributeMap[factionID].contributionMap {
		list = append(list, &userSupSpent{userID, contribution})
	}

	// sort order
	sort.Slice(list, func(i, j int) bool {
		return list[i].supSpent.GreaterThan(list[j].supSpent)
	})

	// initialise location select list
	as.locationDeciders.list = []server.UserID{}

	for _, tid := range triggerByUserID {
		as.locationDeciders.list = append(as.locationDeciders.list, tid)
	}

	// set location select order
	for _, uss := range list {
		// skip the user who trigger the location
		exists := false
		for _, tid := range triggerByUserID {
			if uss.userID == tid {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		// append user to the list
		as.locationDeciders.list = append(as.locationDeciders.list, uss.userID)
	}
}

// nextLocationDeciderGet return the uuid of the next player to select the location for ability
func (as *AbilitiesSystem) nextLocationDeciderGet() (server.UserID, bool) {
	// lock the user id list while setting
	as.locationDeciders.Lock()
	defer as.locationDeciders.Unlock()

	// clean up the location select list if there is not user left to select location
	if len(as.locationDeciders.list) <= 1 {
		as.locationDeciders.list = []server.UserID{}
		return server.UserID(uuid.Nil), false
	}

	// remove the first user from the list
	as.locationDeciders.list = as.locationDeciders.list[1:]

	return as.locationDeciders.list[0], true
}

// ***********************************
// Ability Progression bar Broadcaster
// ***********************************

// 1 tick per second, each tick reduce 0,978 of current price

func (as *AbilitiesSystem) BAttleAbilityPriceUpdater() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		<-ticker.C
		// check battle stage
		as.battle.Stage.RLock()
		stage := as.battle.Stage.Stage
		// exit the loop, when battle is ended
		if stage == BattleStageEnd {
			ticker.Stop()
			as.battle.Stage.RUnlock()
			break
		}
		as.battle.Stage.RUnlock()

		// check bribing stage
		as.battleAbilityPool.Stage.RLock()
		if as.battleAbilityPool.Stage.Phase != BribeStageBribe {
			// skip if the stage is invalid
			as.battleAbilityPool.Stage.RUnlock()
			continue
		}
		as.battleAbilityPool.Stage.RUnlock()

		// update price
		as.battleAbilityPool.BattleAbilityMx.Lock()
		defer as.battleAbilityPool.BattleAbilityMx.Unlock()

		for factionID, ability := range as.battleAbilityPool.Abilities {
			// lock ability while updating
			ability.Lock()

			// reduce price
			ability.SupsCost = ability.SupsCost.Mul(decimal.NewFromFloat(0.978))

			// cap minmum price at 1 sup
			if ability.SupsCost.Cmp(decimal.New(1, 18)) <= 0 {
				ability.SupsCost = decimal.New(1, 18)
			}

			// if ability not triggered, store ability to database
			if ability.SupsCost.Cmp(ability.CurrentSups) > 0 {
				// store updated price to db
				err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ability.ID, ability.SupsCost.String(), ability.CurrentSups.String())
				if err != nil {
					gamelog.L.Err(err)
				}
				ability.Unlock()
				continue
			}

			// if ability triggered
			ability.SupsCost = ability.SupsCost.Mul(decimal.NewFromInt(2))
			ability.CurrentSups = decimal.Zero
			err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ability.ID, ability.SupsCost.String(), ability.CurrentSups.String())
			if err != nil {
				gamelog.L.Err(err)
			}

			// change st
			// generate location select order list
			as.locationSelectListSet(factionID)

			// if no user online, enter cooldown and exit the loop
			if len(as.locationDeciders.list) == 0 {
				// change bribing phase
				as.battleAbilityPool.Stage.Lock()

				// set new battle ability
				cooldownSecond, err := as.SetNewBattleAbility()
				if err != nil {
					gamelog.L.Err(err).Msg("Failed to set new battle ability")
				}

				as.battleAbilityPool.Stage.Phase = BribeStageCooldown
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
				as.battleAbilityPool.Stage.Unlock()

				ability.Unlock()
				break
			}

			// if there is user, assign location decider and exit the loop

			// change bribing phase to location select
			as.battleAbilityPool.Stage.Lock()
			as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
			as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
			// broadcast stage change
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
			as.battleAbilityPool.Stage.Unlock()

			// broadcast the announcement to the next location decider
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, as.locationDeciders.list[0])), &LocationSelectAnnouncement{
				GameAbility: as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
				EndTime:     as.battleAbilityPool.Stage.EndTime,
			})

			ability.Unlock()
			break
		}
	}
}

func (as *AbilitiesSystem) BattleAbilityProgressBar() {
	ticker := time.NewTicker(200 * time.Microsecond)

	for {
		<-ticker.C
		// check battle stage
		as.battle.Stage.RLock()
		stage := as.battle.Stage.Stage
		// exit the loop, when battle is ended
		if stage == BattleStageEnd {
			ticker.Stop()
			as.battle.Stage.RUnlock()
			break
		}
		as.battle.Stage.RUnlock()

		// check bribing stage
		as.battleAbilityPool.Stage.RLock()
		if as.battleAbilityPool.Stage.Phase != BribeStageBribe {
			// skip if the stage is invalid
			as.battleAbilityPool.Stage.RUnlock()
			continue
		}
		as.battleAbilityPool.Stage.RUnlock()

		// gether faction ability price
		as.battleAbilityPool.BattleAbilityMx.RLock()
		factionAbilityPrices := []string{}
		for factionID, ability := range as.battleAbilityPool.Abilities {
			factionAbilityPrice := fmt.Sprintf("%s_%s_%s", factionID.String(), ability.SupsCost.String(), ability.CurrentSups.String())
			factionAbilityPrices = append(factionAbilityPrices, factionAbilityPrice)
		}
		as.battleAbilityPool.BattleAbilityMx.RUnlock()

		// broadcast to frontend
		data, err := json.Marshal(strings.Join(factionAbilityPrices, "|"))
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to parse ability progress bar")
			continue
		}

		as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeyFactionProgressBarUpdated), data)
	}

	fmt.Println("Stop ability progress bar broadcast")
}

// *********************
// Handlers
// *********************

type GameAbilityContributeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		GameAbilityID server.GameAbilityID `json:"gameAbilityID"`
		Amount        int64                `json:"amount"` // 1, 25, 100
	} `json:"payload"`
}

const HubKeFactionUniqueAbilityContribute hub.HubCommandKey = "FACTION:UNIQUE:ABILITY:CONTRIBUTE"

func (as *AbilitiesSystem) AbilityContribute(wsc *hub.Client, payload []byte, factionID server.FactionID) error {
	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	supsAmount := decimal.New(req.Payload.Amount, 18)

	as.FactionUniqueAbilityContribute(uuid.UUID(factionID), req.Payload.GameAbilityID, userID, supsAmount)

	return nil
}

type BribeGabRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount int64 `json:"amount"` // 1, 25, 100
	} `json:"payload"`
}

func (as *AbilitiesSystem) BribeGabs(wsc *hub.Client, payload []byte, factionID server.FactionID) error {
	req := &BribeGabRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	amount := decimal.New(req.Payload.Amount, 18)

	err = as.BattleAbilityBribing(uuid.UUID(factionID), userID, amount)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func (as *AbilitiesSystem) BribeStageGet() *GabsBribeStage {
	return as.battleAbilityPool.Stage
}

func (as *AbilitiesSystem) FactionBattleAbilityGet(factionID uuid.UUID) *GameAbility {
	as.battleAbilityPool.BattleAbilityMx.RLock()
	defer as.battleAbilityPool.BattleAbilityMx.Unlock()

	ability, ok := as.battleAbilityPool.Abilities[factionID]
	if !ok {
		return nil
	}

	return ability
}

func (as *AbilitiesSystem) LocationSelect(userID server.UserID, x int, y int) error {
	// check current user is the winner
	as.locationDeciders.RLock()
	defer as.locationDeciders.RUnlock()

	// check battle end
	as.battle.Stage.RLock()
	if as.battle.Stage.Stage == BattleStageEnd {
		as.battle.Stage.RUnlock()
		return nil
	}
	as.battle.Stage.RUnlock()

	// check eligibility
	if len(as.locationDeciders.list) <= 0 || as.locationDeciders.list[0] != userID {
		return terror.Error(terror.ErrForbidden)
	}

	ability := as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID]

	// trigger location select
	as.battle.arena.Message(
		"BATTLE:ABILITY",
		&server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: ability.GameClientAbilityID,
			TriggeredOnCellX:    &x,
			TriggeredOnCellY:    &y,
			TriggeredByUserID:   &userID,
			// TODO: Get user name
		},
	)

	// TODO: store ability event data
	as.battle.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
		Type: LocationSelectTypeTrigger,
		X:    &x,
		Y:    &y,
		Ability: &server.AbilityBrief{
			Label:    ability.Label,
			ImageUrl: ability.ImageUrl,
			Colour:   ability.Colour,
		},
		// TODO: Get current user
	})

	return nil
}
