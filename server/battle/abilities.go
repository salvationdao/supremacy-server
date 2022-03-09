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

const EachMechIntroSecond = 3
const InitIntroSecond = 7

type LocationDeciders struct {
	deadlock.RWMutex
	list []server.UserID
}

type AbilitiesSystem struct {
	battle *Battle
	// faction unique abilities
	factionUniqueAbilities map[uuid.UUID]map[server.GameAbilityID]*GameAbility // map[faction_id]map[identity]*Ability

	// gabs abilities (air craft, nuke, repair)
	gabsAbilityPool *GabsAbilityPool

	// track the sups contribution of each user, use for location select
	userContributeMap map[uuid.UUID]*UserContribution

	// location select winner list
	locationDeciders *LocationDeciders
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	factionAbilities := map[uuid.UUID]map[server.GameAbilityID]*GameAbility{}

	// initialise new gabs ability pool
	gabsAbilityPool := &GabsAbilityPool{
		Stage: &GabsBribeStage{
			Phase:   BribeStageHold,
			EndTime: time.Now().AddDate(1, 0, 0), // HACK: set end time to far future to implement infinite time
		},
		Abilities: map[uuid.UUID]*GameAbility{},
	}

	userContributeMap := map[uuid.UUID]*UserContribution{}

	// init battle ability
	err := gabsAbilityPool.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to set up battle ability")
		return nil
	}

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
		gabsAbilityPool:        gabsAbilityPool,
		userContributeMap:      userContributeMap,
		locationDeciders: &LocationDeciders{
			list: []server.UserID{},
		},
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

				// terminate the function when battle is end
				fmt.Println("Battle Ended")
				break
			}

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
	err := db.FactionExclusiveAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
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
		err := db.FactionExclusiveAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
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
	err = db.FactionExclusiveAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
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

type GabsAbilityPool struct {
	deadlock.Mutex // force bribe process synchronize
	Stage          *GabsBribeStage

	// pick random battle ability
	BattleAbility *server.BattleAbility

	Abilities map[uuid.UUID]*GameAbility // faction ability current, change on every bribing cycle

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
		as.gabsAbilityPool.Stage.RLock()

		// skip, if the end time of current phase haven't been reached
		if as.gabsAbilityPool.Stage.EndTime.After(time.Now()) {
			as.gabsAbilityPool.Stage.RUnlock()
			continue
		}

		// otherwise, read current bribe phase
		bribePhase := as.gabsAbilityPool.Stage.Phase

		as.gabsAbilityPool.Stage.RUnlock()

		/////////////////
		// Bribe Phase //
		/////////////////
		switch bribePhase {

		// at the end of bribing phase
		// no ability is triggered, switch to cooldown phase
		case BribeStageBribe:
			// change bribing phase
			as.gabsAbilityPool.Stage.Lock()

			// set new battle ability
			err := as.gabsAbilityPool.SetNewBattleAbility()
			if err != nil {
				gamelog.L.Err(err).Msg("Failed to set new battle ability")
			}

			as.gabsAbilityPool.Stage.Phase = BribeStageCooldown
			as.gabsAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(as.gabsAbilityPool.BattleAbility.CooldownDurationSecond) * time.Second)
			// broadcast stage to frontend
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.gabsAbilityPool.Stage)
			as.gabsAbilityPool.Stage.Unlock()

		// at the end of location select phase
		// pass the location select to next player
		case BribeStageLocationSelect:
			// get the next location decider
			userID, ok := as.nextLocationDeciderGet()
			if !ok {
				// enter cooldown phase, if there is no user left for location select

				// change bribing phase
				as.gabsAbilityPool.Stage.Lock()

				// set new battle ability
				err := as.gabsAbilityPool.SetNewBattleAbility()
				if err != nil {
					gamelog.L.Err(err).Msg("Failed to set new battle ability")
				}

				as.gabsAbilityPool.Stage.Phase = BribeStageCooldown
				as.gabsAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(as.gabsAbilityPool.BattleAbility.CooldownDurationSecond) * time.Second)
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.gabsAbilityPool.Stage)
				as.gabsAbilityPool.Stage.Unlock()
			}

			// extend location select phase duration
			as.gabsAbilityPool.Stage.Lock()
			as.gabsAbilityPool.Stage.Phase = BribeStageLocationSelect
			as.gabsAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.gabsAbilityPool.Stage)
			as.gabsAbilityPool.Stage.Unlock()

			// broadcast the next location decider
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID)), &LocationSelectAnnouncement{
				GameAbility: as.gabsAbilityPool.Abilities[as.gabsAbilityPool.TriggeredFactionID],
				EndTime:     as.gabsAbilityPool.Stage.EndTime,
			})

		// at the end of cooldown phase
		// random choose a battle ability for next bribing session
		case BribeStageCooldown:

			// change bribing phase
			as.gabsAbilityPool.Stage.Lock()
			as.gabsAbilityPool.Stage.Phase = BribeStageBribe
			as.gabsAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(BribeDurationSecond) * time.Second)
			// broadcast stage to frontend
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.gabsAbilityPool.Stage)
			as.gabsAbilityPool.Stage.Unlock()

			continue
		}
	}

	// do some thing after battle end...
}

// SetNewBattleAbility query
func (bap *GabsAbilityPool) SetNewBattleAbility() error {
	// clean up triggered faction
	bap.TriggeredFactionID = uuid.Nil

	// initialise new gabs ability pool
	ba, err := db.BattleAbilityGetRandom(context.Background(), gamedb.Conn)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to get battle ability from db")
		return terror.Error(err)
	}
	bap.BattleAbility = ba

	// get faction battle abilities
	gabsAbilities, err := db.FactionBattleAbilityGet(context.Background(), gamedb.Conn, ba.ID)
	if err != nil {
		return terror.Error(err)
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
		bap.Abilities[ga.FactionID] = &GameAbility{
			ID:                  ga.ID,
			GameClientAbilityID: byte(ga.GameClientAbilityID),
			ImageUrl:            ga.ImageUrl,
			Description:         ga.Description,
			FactionID:           ga.FactionID,
			Label:               ga.Label,
			SupsCost:            supsCost,
			CurrentSups:         currentSups,
		}
	}

	return nil
}

func (as *AbilitiesSystem) BattleAbilityContribution(factionID uuid.UUID, userID server.UserID, amount decimal.Decimal) error {
	as.gabsAbilityPool.Lock()
	defer as.gabsAbilityPool.Unlock()

	// check current battle stage
	as.battle.Stage.RLock()
	as.gabsAbilityPool.Stage.RLock()
	// return early if battle stage or bribing stage are invalid
	if as.battle.Stage.Stage != BattleStagStart || as.gabsAbilityPool.Stage.Phase != BribeStageBribe {
		as.gabsAbilityPool.Stage.RUnlock()
		as.battle.Stage.RUnlock()
		return nil
	}
	as.gabsAbilityPool.Stage.RUnlock()
	as.battle.Stage.RUnlock()

	// check faction ability exists
	if factionAbility, ok := as.gabsAbilityPool.Abilities[uuid.UUID(factionID)]; ok {
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
			as.gabsAbilityPool.Stage.Lock()
			as.gabsAbilityPool.Stage.Phase = BribeStageLocationSelect
			as.gabsAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
			// broadcast stage change
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.gabsAbilityPool.Stage)
			as.gabsAbilityPool.Stage.Unlock()

			// send message to the user who trigger the ability
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID)), &LocationSelectAnnouncement{
				GameAbility: as.gabsAbilityPool.Abilities[as.gabsAbilityPool.TriggeredFactionID],
				EndTime:     as.gabsAbilityPool.Stage.EndTime,
			})
		}
	}

	return nil
}

// locationSelectListSet set a user list for location select for current ability triggered
func (as *AbilitiesSystem) locationSelectListSet(factionID uuid.UUID, triggerByUserID server.UserID) {
	// lock the user id list while setting
	as.locationDeciders.Lock()
	defer as.locationDeciders.Unlock()

	// set triggered faction id
	as.gabsAbilityPool.TriggeredFactionID = factionID

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
	as.locationDeciders.list = []server.UserID{triggerByUserID}

	// set location select order
	for _, uss := range list {
		// skip the user who trigger the location
		if uss.userID == triggerByUserID {
			continue
		}
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

	err = as.BattleAbilityContribution(uuid.UUID(factionID), userID, amount)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func (as *AbilitiesSystem) BribeStageGet() *GabsBribeStage {
	return as.gabsAbilityPool.Stage
}

func (as *AbilitiesSystem) LocationSelect(userID server.UserID, x int, y int) error {
	// check current user is the winner
	as.locationDeciders.RLock()
	defer as.locationDeciders.RUnlock()

	// check eligibility
	if len(as.locationDeciders.list) <= 0 || as.locationDeciders.list[0] != userID {
		return terror.Error(terror.ErrForbidden)
	}

	// trigger location select
	as.battle.arena.Message(
		"BATTLE:ABILITY",
		&server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: as.gabsAbilityPool.Abilities[as.gabsAbilityPool.TriggeredFactionID].GameClientAbilityID,
			TriggeredOnCellX:    &x,
			TriggeredOnCellY:    &y,
			TriggeredByUserID:   &userID,
			// TODO: Get user name
		},
	)

	// TODO: store ability event data

	return nil
}
