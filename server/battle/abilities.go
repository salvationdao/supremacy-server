package battle

import (
	"context"
	"database/sql"
	"errors"
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
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
)

//******************************
// Game Ability setup
//******************************

const EachMechIntroSecond = 0
const InitIntroSecond = 1

type LocationDeciders struct {
	list []uuid.UUID
}

type AbilitiesSystem struct {
	battle *Battle
	// faction unique abilities
	factionUniqueAbilities map[uuid.UUID]map[string]*GameAbility // map[faction_id]map[identity]*Ability

	// gabs abilities (air craft, nuke, repair)
	battleAbilityPool *BattleAbilityPool

	// track the sups contribution of each user, use for location select
	userContributeMap map[uuid.UUID]*UserContribution

	bribe      chan *Contribution
	contribute chan *Contribution
	// location select winner list
	locationDeciders *LocationDeciders
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	factionAbilities := map[uuid.UUID]map[string]*GameAbility{}

	// initialise new gabs ability pool
	battleAbilityPool := &BattleAbilityPool{
		Stage: &GabsBribeStage{
			Phase:   BribeStageHold,
			EndTime: time.Now().AddDate(1, 0, 0), // HACK: set end time to far future to implement infinite time
		},
		BattleAbility: &server.BattleAbility{},
		Abilities:     map[uuid.UUID]*GameAbility{},
	}

	userContributeMap := map[uuid.UUID]*UserContribution{}

	for factionID := range battle.factions {
		// initialise faction unique abilities
		factionAbilities[factionID] = map[string]*GameAbility{}

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
						gamelog.L.Error().Err(err).Msg("Failed to ability sups cost to decimal")

						// set sups cost to initial price
						supsCost = decimal.New(100, 18)
					}

					currentSups, err := decimal.NewFromString(ability.CurrentSups)
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to ability current sups to decimal")

						// set current sups to initial price
						currentSups = decimal.Zero
					}

					// build the ability
					wmAbility := &GameAbility{
						ID:                  uuid.Must(uuid.FromString(ability.ID)), // generate a uuid for frontend to track sups contribution
						Identity:            wm.Hash,
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
						OfferingID:          uuid.Must(uuid.NewV4()),
					}

					// inject ability to war machines
					battle.WarMachines[i].Abilities = []*GameAbility{wmAbility}

					// store faction ability for price tracking
					factionAbilities[factionID][wmAbility.Identity] = wmAbility
				}
			}

		} else {
			// for other faction unique abilities
			abilities := map[string]*GameAbility{}
			for _, ability := range factionUniqueAbilities {

				supsCost, err := decimal.NewFromString(ability.SupsCost)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to ability sups cost to decimal")

					// set sups cost to initial price
					supsCost = decimal.New(100, 18)
				}

				currentSups, err := decimal.NewFromString(ability.CurrentSups)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to ability current sups to decimal")

					// set current sups to initial price
					currentSups = decimal.Zero
				}

				wmAbility := &GameAbility{
					ID:                  uuid.Must(uuid.FromString(ability.ID)), // generate a uuid for frontend to track sups contribution
					Identity:            ability.ID,
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
					OfferingID:          uuid.Must(uuid.NewV4()),
				}
				abilities[wmAbility.Identity] = wmAbility
			}
			factionAbilities[factionID] = abilities
		}

		// initialise user vote map in gab ability pool
		userContributeMap[factionID] = &UserContribution{
			contributionMap: map[uuid.UUID]decimal.Decimal{},
		}
	}

	as := &AbilitiesSystem{
		battle:                 battle,
		factionUniqueAbilities: factionAbilities,
		battleAbilityPool:      battleAbilityPool,
		userContributeMap:      userContributeMap,
		locationDeciders: &LocationDeciders{
			list: []uuid.UUID{},
		},
	}

	// broadcast faction unique ability
	for factionID, ga := range as.factionUniqueAbilities {
		if factionID.String() == server.ZaibatsuFactionID.String() {
			// broadcast the war machine abilities
			for identity, ability := range ga {
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineAbilitiesUpdated, identity)), []*GameAbility{ability})
			}
		} else {
			// broadcast faction ability
			for _, ability := range ga {
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionUniqueAbilitiesUpdated, factionID.String())), []*GameAbility{ability})
			}
		}
	}

	// init battle ability
	_, err := as.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to set up battle ability")
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

	main_ticker := time.NewTicker(1 * time.Second)

	as.contribute = make(chan *Contribution, 10)

	// start the battle
	for {
		select {
		case <-main_ticker.C:
			for _, abilities := range as.factionUniqueAbilities {

				// start ability price updater for each faction
				// read the stage first
				stage := as.battle.stage

				// start ticker while still in battle
				if stage == BattleStagStart {
					for _, ability := range abilities {
						// update ability price
						isTriggered := ability.FactionUniqueAbilityPriceUpdate(minPrice)

						triggeredFlag := "0"
						if isTriggered {
							triggeredFlag = "1"
							// send message to game client, if ability trigger
							as.battle.arena.Message(
								"BATTLE:ABILITY",
								&server.GameAbilityEvent{
									IsTriggered:         true,
									GameClientAbilityID: ability.GameClientAbilityID,
									ParticipantID:       ability.ParticipantID, // trigger on war machine
								},
							)

							bat := boiler.BattleAbilityTrigger{
								PlayerID:          null.StringFromPtr(nil),
								BattleID:          as.battle.ID.String(),
								FactionID:         ability.FactionID.String(),
								IsAllSyndicates:   false,
								AbilityLabel:      ability.Label,
								GameAbilityID:     ability.ID.String(),
								AbilityOfferingID: ability.OfferingID.String(),
							}
							err := bat.Insert(gamedb.StdConn, boil.Infer())
							if err != nil {
								gamelog.L.Error().Err(err).Msg("Failed to record ability triggered")
							}

							// generate new offering id for current ability
							ability.OfferingID = uuid.Must(uuid.NewV4())
						}

						// broadcast the new price
						payload := []byte{byte(GameAbilityProgressTick)}
						payload = append(payload, []byte(fmt.Sprintf("%s_%s_%s_%s", ability.Identity, ability.SupsCost.String(), ability.CurrentSups.String(), triggeredFlag))...)
						as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(fmt.Sprintf("%s,%s", HubKeyAbilityPriceUpdated, ability.Identity)), payload)

					}
				} else {
					gamelog.L.Info().Msg("exiting ability price update")
					return
				}

			}
		case cont := <-as.contribute:
			if abilities, ok := as.factionUniqueAbilities[cont.factionID]; ok {

				// check ability exists
				if ability, ok := abilities[cont.abilityIdentity]; ok {

					// return early if battle stage is invalid
					if as.battle.stage != BattleStagStart {
						return
					}

					actualSupSpent, isTriggered := ability.SupContribution(as.battle.arena.ppClient, as.battle.ID.String(), as.battle.BattleNumber, cont.userID, cont.amount)

					// cache user's sup contribution for generating location select order
					if _, ok := as.userContributeMap[cont.factionID].contributionMap[cont.userID]; !ok {
						as.userContributeMap[cont.factionID].contributionMap[cont.userID] = decimal.Zero
					}
					as.userContributeMap[cont.factionID].contributionMap[cont.userID] = as.userContributeMap[cont.factionID].contributionMap[cont.userID].Add(actualSupSpent)

					// sups contribution
					triggeredFlag := "0"
					if isTriggered {
						triggeredFlag = "1"
						// send message to game client, if ability trigger
						as.battle.arena.Message(
							"BATTLE:ABILITY",
							&server.GameAbilityEvent{
								IsTriggered:         true,
								GameClientAbilityID: ability.GameClientAbilityID,
								ParticipantID:       ability.ParticipantID, // trigger on war machine
							},
						)

						bat := boiler.BattleAbilityTrigger{
							PlayerID:          null.StringFrom(cont.userID.String()),
							BattleID:          as.battle.ID.String(),
							FactionID:         ability.FactionID.String(),
							IsAllSyndicates:   false,
							AbilityLabel:      ability.Label,
							GameAbilityID:     ability.ID.String(),
							AbilityOfferingID: ability.OfferingID.String(),
						}
						err := bat.Insert(gamedb.StdConn, boil.Infer())
						if err != nil {
							gamelog.L.Error().Err(err).Msg("Failed to record ability triggered")
						}

						// generate new offering id for current ability
						ability.OfferingID = uuid.Must(uuid.NewV4())
					}

					// broadcast the new price
					payload := []byte{byte(GameAbilityProgressTick)}
					payload = append(payload, []byte(fmt.Sprintf("%s_%s_%s_%s", ability.Identity, ability.SupsCost.String(), ability.CurrentSups.String(), triggeredFlag))...)
					as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(fmt.Sprintf("%s,%s", HubKeyAbilityPriceUpdated, ability.Identity)), payload)
				}
			}
		}
	}
}

// FactionUniqueAbilityPriceUpdate update target price on every tick
func (ga *GameAbility) FactionUniqueAbilityPriceUpdate(minPrice decimal.Decimal) bool {
	ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromFloat(0.9977))

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
		gamelog.L.Error().Err(err)
		return isTriggered
	}

	return isTriggered
}

// SupContribution contribute sups to specific game ability, return the actual sups spent and whether the ability is triggered
func (ga *GameAbility) SupContribution(ppClient *passport.Passport, battleID string, battleNumber int, userID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, bool) {

	isTriggered := false

	// calc the different
	diff := ga.SupsCost.Sub(ga.CurrentSups)

	// if players spend more thant they need, crop the spend price
	if amount.Cmp(diff) >= 0 {
		isTriggered = true
		amount = diff
	}
	now := time.Now()

	// pay sup
	txid, err := ppClient.SpendSupMessage(passport.SpendSupsReq{
		FromUserID:           userID,
		Amount:               amount.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("ability_sup_contribute|%s|%d", ga.OfferingID.String(), time.Now().UnixNano())),
		Group:                "battle",
		SubGroup:             battleID,
		Description:          "battle contribution: " + ga.Label,
		NotSafe:              true,
	})
	if err != nil {
		return decimal.Zero, false
	}

	isAllSyndicates := false
	if ga.BattleAbilityID == nil || ga.BattleAbilityID.IsNil() {
		isAllSyndicates = true
	}

	battleContrib := &boiler.BattleContribution{
		BattleID:          battleID,
		PlayerID:          userID.String(),
		AbilityOfferingID: ga.OfferingID.String(),
		DidTrigger:        isTriggered,
		FactionID:         ga.FactionID.String(),
		AbilityLabel:      ga.Label,
		IsAllSyndicates:   isAllSyndicates,
		Amount:            amount,
		ContributedAt:     now,
		TransactionID:     null.StringFrom(txid),
	}

	err = battleContrib.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("txid", txid).Err(err).Msg("unable to insert battle contrib")
	}

	tx, err := gamedb.StdConn.Begin()
	if err == nil {

		defer tx.Rollback()

		spoil, err := boiler.SpoilsOfWars(qm.Where(`battle_id = ?`, battleID)).One(tx)
		if errors.Is(err, sql.ErrNoRows) {
			spoil = &boiler.SpoilsOfWar{
				BattleID:     battleID,
				BattleNumber: battleNumber,
				Amount:       amount,
				AmountSent:   decimal.New(0, 18),
			}
		} else {
			spoil.Amount = spoil.Amount.Add(amount)
			//broadcast spoil of war total and tick here
		}

		_, err = spoil.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to insert spoil of war")
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to create tx")
			tx.Rollback()
		}
	} else {
		gamelog.L.Error().Err(err).Msg("unable to create tx to create spoil of war")
	}

	// update the current sups if not triggered
	if !isTriggered {
		ga.CurrentSups = ga.CurrentSups.Add(amount)

		// store updated price to db
		err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost.String(), ga.CurrentSups.String())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to insert faction ability sup cost update")
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
		gamelog.L.Error().Err(err)
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
	Phase   BribePhase `json:"phase"`
	EndTime time.Time  `json:"end_time"`
}

// track user contribution of current battle
type UserContribution struct {
	contributionMap map[uuid.UUID]decimal.Decimal
}

type BattleAbilityPool struct {
	Stage *GabsBribeStage

	BattleAbility *server.BattleAbility
	Abilities     map[uuid.UUID]*GameAbility // faction ability current, change on every bribing cycle

	TriggeredFactionID uuid.UUID
}

type LocationSelectAnnouncement struct {
	GameAbility *GameAbility `json:"game_ability"`
	EndTime     time.Time    `json:"end_time"`
}

// StartGabsAbilityPoolCycle
func (as *AbilitiesSystem) StartGabsAbilityPoolCycle(waitDurationSecond int) {
	// wait for mech intro
	time.Sleep(time.Duration(waitDurationSecond) * time.Second)

	// ability price updater
	as.bribe = make(chan *Contribution, 10)

	// initial a ticker for current battle
	main_ticker := time.NewTicker(1 * time.Second)
	price_ticker := time.NewTicker(1 * time.Second)
	progress_ticker := time.NewTicker(1 * time.Second)

	// start voting stage
	as.battleAbilityPool.Stage.Phase = BribeStageBribe
	as.battleAbilityPool.Stage.EndTime = time.Now().Add(BribeDurationSecond * time.Second)
	as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

	// start ability pool cycle
	for {
		select {
		// wait for next tick
		case <-main_ticker.C:

			// check phase
			stage := as.battle.stage
			// exit the loop, when battle is ended
			if stage == BattleStageEnd {
				// change phase to hold and broadcast to user
				as.battleAbilityPool.Stage.Phase = BribeStageHold
				as.battleAbilityPool.Stage.EndTime = time.Now().AddDate(1, 0, 0) // HACK: set end time to far future to implement infinite time
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

				// stop all the ticker and exit the loop
				main_ticker.Stop()
				price_ticker.Stop()
				progress_ticker.Stop()
				gamelog.L.Info().Msg("Stop ability tickers after battle is end")
				return
			}

			// skip, if the end time of current phase haven't been reached
			if as.battleAbilityPool.Stage.EndTime.After(time.Now()) {
				continue
			}

			// otherwise, read current bribe phase
			bribePhase := as.battleAbilityPool.Stage.Phase

			/////////////////
			// Bribe Phase //
			/////////////////
			switch bribePhase {

			// at the end of bribing phase
			// no ability is triggered, switch to cooldown phase
			case BribeStageBribe:
				// change bribing phase

				// set new battle ability
				cooldownSecond, err := as.SetNewBattleAbility()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to set new battle ability")
				}

				as.battleAbilityPool.Stage.Phase = BribeStageCooldown
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
				// broadcast stage to frontend
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

			// at the end of location select phase
			// pass the location select to next player
			case BribeStageLocationSelect:
				// get the next location decider
				userID, ok := as.nextLocationDeciderGet()
				if !ok {
					// set new battle ability
					cooldownSecond, err := as.SetNewBattleAbility()
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to set new battle ability")
					}

					// enter cooldown phase, if there is no user left for location select
					as.battleAbilityPool.Stage.Phase = BribeStageCooldown
					as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
					as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
					continue
				}

				// extend location select phase duration
				as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
				// broadcast stage to frontend
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

				// broadcast the announcement to the next location decider
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, userID)), &LocationSelectAnnouncement{
					GameAbility: as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
					EndTime:     as.battleAbilityPool.Stage.EndTime,
				})

			// at the end of cooldown phase
			// random choose a battle ability for next bribing session
			case BribeStageCooldown:

				// change bribing phase
				as.battleAbilityPool.Stage.Phase = BribeStageBribe
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(BribeDurationSecond) * time.Second)
				// broadcast stage to frontend
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

				continue
			default:
				gamelog.L.Error().Msg("hit default case switch on abilities loop")
			}
		case <-price_ticker.C:
			as.BattleAbilityPriceUpdater()
		case cont := <-as.bribe:
			if factionAbility, ok := as.battleAbilityPool.Abilities[cont.factionID]; ok {

				// contribute sups
				actualSupSpent, abilityTriggered := factionAbility.SupContribution(as.battle.arena.ppClient, as.battle.ID.String(), as.battle.BattleNumber, cont.userID, cont.amount)

				// cache user contribution for location select order
				if _, ok := as.userContributeMap[cont.factionID].contributionMap[cont.userID]; !ok {
					as.userContributeMap[cont.factionID].contributionMap[cont.userID] = decimal.Zero
				}
				as.userContributeMap[cont.factionID].contributionMap[cont.userID] = as.userContributeMap[cont.factionID].contributionMap[cont.userID].Add(actualSupSpent)

				if abilityTriggered {
					// generate location select order list
					as.locationDecidersSet(cont.factionID, cont.userID)

					// change bribing phase to location select
					as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
					as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)

					// broadcast stage change
					as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

					// send message to the user who trigger the ability
					as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, cont.userID)), &LocationSelectAnnouncement{
						GameAbility: as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
						EndTime:     as.battleAbilityPool.Stage.EndTime,
					})

					// broadcast the latest result progress bar, when ability is triggered
					go as.BroadcastAbilityProgressBar()
				}
			}
		case <-progress_ticker.C:
			as.BattleAbilityProgressBar()
		}
	}
}

// SetNewBattleAbility set new battle ability and return the cooldown time
func (as *AbilitiesSystem) SetNewBattleAbility() (int, error) {
	// clean up triggered faction
	as.battleAbilityPool.TriggeredFactionID = uuid.Nil

	// initialise new gabs ability pool
	ba, err := db.BattleAbilityGetRandom(context.Background(), gamedb.Conn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get battle ability from db")
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
			gamelog.L.Error().Err(err).Msg("Failed to ability sups cost to decimal")

			// set sups cost to initial price
			supsCost = decimal.New(100, 18)
		}

		currentSups, err := decimal.NewFromString(ga.CurrentSups)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to ability current sups to decimal")

			// set current sups to initial price
			currentSups = decimal.Zero
		}

		// initialise game ability
		gameAbility := &GameAbility{
			ID:                     ga.ID,
			GameClientAbilityID:    byte(ga.GameClientAbilityID),
			ImageUrl:               ga.ImageUrl,
			Description:            ga.Description,
			FactionID:              ga.FactionID,
			Label:                  ga.Label,
			SupsCost:               supsCost,
			CurrentSups:            currentSups,
			Colour:                 ga.Colour,
			TextColour:             ga.TextColour,
			CooldownDurationSecond: ba.CooldownDurationSecond,
			OfferingID:             uuid.Must(uuid.NewV4()),
		}
		as.battleAbilityPool.Abilities[ga.FactionID] = gameAbility

		// broadcast ability update to faction users
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBattleAbilityUpdated, gameAbility.FactionID.String())), gameAbility)
	}

	// broadcast the latest result progress bar, when ability is triggered
	factionAbilityPrices := []string{}
	for factionID, ability := range as.battleAbilityPool.Abilities {
		factionAbilityPrice := fmt.Sprintf("%s_%s_%s", factionID.String(), ability.SupsCost.String(), ability.CurrentSups.String())
		factionAbilityPrices = append(factionAbilityPrices, factionAbilityPrice)
	}

	payload := []byte{byte(BattleAbilityProgressTick)}
	payload = append(payload, []byte(strings.Join(factionAbilityPrices, "|"))...)

	as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeyBattleAbilityProgressBarUpdated), payload)

	return ba.CooldownDurationSecond, nil
}

type Contribution struct {
	factionID       uuid.UUID
	userID          uuid.UUID
	amount          decimal.Decimal
	abilityIdentity string
}

// locationDecidersSet set a user list for location select for current ability triggered
func (as *AbilitiesSystem) locationDecidersSet(factionID uuid.UUID, triggerByUserID ...uuid.UUID) {
	// set triggered faction id
	as.battleAbilityPool.TriggeredFactionID = factionID

	type userSupSpent struct {
		userID   uuid.UUID
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
	as.locationDeciders.list = []uuid.UUID{}

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
func (as *AbilitiesSystem) nextLocationDeciderGet() (uuid.UUID, bool) {
	// clean up the location select list if there is no user left to select location
	if len(as.locationDeciders.list) <= 1 {
		as.locationDeciders.list = []uuid.UUID{}
		return uuid.UUID(uuid.Nil), false
	}

	// remove the first user from the list
	as.locationDeciders.list = as.locationDeciders.list[1:]

	return as.locationDeciders.list[0], true
}

// ***********************************
// Ability Progression bar Broadcaster
// ***********************************

// 1 tick per second, each tick reduce 0,978 of current price

func (as *AbilitiesSystem) BattleAbilityPriceUpdater() {

	// check battle stage
	stage := as.battle.stage
	// exit the loop, when battle is ended
	if stage == BattleStageEnd {
		return
	}

	// check bribing stage
	if as.battleAbilityPool.Stage.Phase != BribeStageBribe {
		// skip if the stage is invalid
		return
	}

	// update price
	for factionID, ability := range as.battleAbilityPool.Abilities {
		// reduce price
		ability.SupsCost = ability.SupsCost.Mul(decimal.NewFromFloat(0.978))

		// cap minmum price at 1 sup
		if ability.SupsCost.Cmp(decimal.New(1, 18)) <= 0 {
			ability.SupsCost = decimal.New(1, 18)
		}

		// if ability not triggered, store ability's new target price to database, and continue
		if ability.SupsCost.Cmp(ability.CurrentSups) > 0 {
			// store updated price to db
			err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ability.ID, ability.SupsCost.String(), ability.CurrentSups.String())
			if err != nil {
				gamelog.L.Error().Err(err)
			}
			continue
		}

		// if ability triggered
		ability.SupsCost = ability.SupsCost.Mul(decimal.NewFromInt(2))
		ability.CurrentSups = decimal.Zero
		err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ability.ID, ability.SupsCost.String(), ability.CurrentSups.String())
		if err != nil {
			gamelog.L.Error().Err(err)
		}

		// broadcast the progress bar
		as.BroadcastAbilityProgressBar()

		// set location deciders list
		as.locationDecidersSet(factionID)

		// if no user online, enter cooldown and exit the loop
		if len(as.locationDeciders.list) == 0 {
			// change bribing phase

			// set new battle ability
			cooldownSecond, err := as.SetNewBattleAbility()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to set new battle ability")
			}

			as.battleAbilityPool.Stage.Phase = BribeStageCooldown
			as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
			as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

			return
		}

		// if there is user, assign location decider and exit the loop

		// change bribing phase to location select
		as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
		as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
		// broadcast stage change
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

		// broadcast the announcement to the next location decider
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, as.locationDeciders.list[0])), &LocationSelectAnnouncement{
			GameAbility: as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
			EndTime:     as.battleAbilityPool.Stage.EndTime,
		})

		return
	}

	// broadcast the progress bar
	go as.BroadcastAbilityProgressBar()
}

func (as *AbilitiesSystem) BattleAbilityProgressBar() {
	// check battle stage
	stage := as.battle.stage
	// exit the loop, when battle is ended
	if stage == BattleStageEnd {
		return
	}

	// check bribing stage
	if as.battleAbilityPool.Stage.Phase != BribeStageBribe {
		// skip if the stage is invalid
		return
	}

	go as.BroadcastAbilityProgressBar()
}

func (as *AbilitiesSystem) BroadcastAbilityProgressBar() {
	factionAbilityPrices := []string{}
	for factionID, ability := range as.battleAbilityPool.Abilities {
		factionAbilityPrice := fmt.Sprintf("%s_%s_%s", factionID.String(), ability.SupsCost.String(), ability.CurrentSups.String())
		factionAbilityPrices = append(factionAbilityPrices, factionAbilityPrice)
	}

	payload := []byte{byte(BattleAbilityProgressTick)}
	payload = append(payload, []byte(strings.Join(factionAbilityPrices, "|"))...)

	as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeyBattleAbilityProgressBarUpdated), payload)
}

// *********************
// Handlers
// *********************
func (as *AbilitiesSystem) AbilityContribute(factionID uuid.UUID, userID uuid.UUID, abilityIdentity string, amount decimal.Decimal) {
	if as.battle.stage != BattleStagStart {
		return
	}

	cont := &Contribution{
		factionID,
		userID,
		amount,
		abilityIdentity,
	}

	as.contribute <- cont
}

// FactionUniqueAbilityGet return the faction unique ability for the given faction
func (as *AbilitiesSystem) FactionUniqueAbilitiesGet(factionID uuid.UUID) []*GameAbility {
	abilities := []*GameAbility{}
	for _, ga := range as.factionUniqueAbilities[factionID] {
		abilities = append(abilities, ga)
	}

	if len(abilities) == 0 {
		return nil
	}

	return abilities
}

// FactionUniqueAbilityGet return the faction unique ability for the given faction
func (as *AbilitiesSystem) WarMachineAbilitiesGet(factionID uuid.UUID, hash string) []*GameAbility {
	abilities := []*GameAbility{}
	// NOTE: just pass down the faction unique abilities for now
	if fua, ok := as.factionUniqueAbilities[factionID]; ok {
		for h, ga := range fua {
			if h == hash {
				abilities = append(abilities, ga)
			}
		}
	}

	if len(abilities) == 0 {
		return nil
	}

	return abilities
}

func (as *AbilitiesSystem) BribeGabs(factionID uuid.UUID, userID uuid.UUID, amount decimal.Decimal) {
	if as.battle.stage != BattleStagStart || as.battleAbilityPool.Stage.Phase != BribeStageBribe {
		return
	}

	cont := &Contribution{
		factionID,
		userID,
		amount,
		"",
	}

	as.bribe <- cont
}

func (as *AbilitiesSystem) BribeStageGet() *GabsBribeStage {
	if as.battleAbilityPool != nil {
		return as.battleAbilityPool.Stage
	}
	return nil
}

func (as *AbilitiesSystem) FactionBattleAbilityGet(factionID uuid.UUID) *GameAbility {
	if as.battleAbilityPool == nil || as.battleAbilityPool.Abilities == nil {
		return nil
	}
	ability, ok := as.battleAbilityPool.Abilities[factionID]
	if !ok {
		gamelog.L.Warn().Str("func", "FactionBattleAbilityGet").Msg("unable to retrieve abilities for faction")
		return nil
	}

	return ability
}

func (as *AbilitiesSystem) LocationSelect(userID uuid.UUID, x int, y int) error {
	// check battle end
	if as.battle.stage == BattleStageEnd {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("battle stage has en ended")
		return nil
	}

	// check eligibility
	if len(as.locationDeciders.list) <= 0 || as.locationDeciders.list[0] != userID {
		return terror.Error(terror.ErrForbidden)
	}

	ability := as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID]

	// get player detail
	player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(userID.String())).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "player not exists")
	}

	faction, err := db.FactionGet(as.battleAbilityPool.TriggeredFactionID.String())
	if err != nil {
		return terror.Error(err, "player not exists")
	}

	// trigger location select
	as.battle.arena.Message(
		"BATTLE:ABILITY",
		&server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: ability.GameClientAbilityID,
			TriggeredOnCellX:    &x,
			TriggeredOnCellY:    &y,
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
		},
	)

	bat := boiler.BattleAbilityTrigger{
		PlayerID:          null.StringFrom(userID.String()),
		BattleID:          as.battle.ID.String(),
		FactionID:         ability.FactionID.String(),
		IsAllSyndicates:   true,
		AbilityLabel:      ability.Label,
		GameAbilityID:     ability.ID.String(),
		AbilityOfferingID: ability.OfferingID.String(),
	}
	err = bat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to record ability triggered")
	}

	as.battle.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
		Type: LocationSelectTypeTrigger,
		X:    &x,
		Y:    &y,
		Ability: &server.AbilityBrief{
			Label:    ability.Label,
			ImageUrl: ability.ImageUrl,
			Colour:   ability.Colour,
		},
		CurrentUser: &server.UserBrief{
			ID:       userID,
			Username: player.Username.String,
			Faction: &server.FactionBrief{
				Label:      faction.Label,
				LogoBlobID: server.BlobID(uuid.FromStringOrNil(FactionLogos[as.battleAbilityPool.TriggeredFactionID.String()])),
				Theme: &server.FactionTheme{
					Primary:    faction.PrimaryColor,
					Secondary:  faction.SecondaryColor,
					Background: faction.BackgroundColor,
				},
			},
		},
	})

	// enter the cooldown phase
	cooldownSecond, err := as.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to set new battle ability")
	}

	as.battleAbilityPool.Stage.Phase = BribeStageCooldown
	as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(cooldownSecond) * time.Second)
	// broadcast stage to frontend
	as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

	return nil
}
