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
	"strings"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
)

//******************************
// Game Ability setup
//******************************
type LocationDeciders struct {
	list []uuid.UUID
}

type LiveCount struct {
	deadlock.Mutex
	TotalVotes decimal.Decimal `json:"total_votes"`
}

func (lc *LiveCount) AddSups(amount decimal.Decimal) {
	lc.Lock()
	lc.TotalVotes = lc.TotalVotes.Add(amount)
	lc.Unlock()
}

func (lc *LiveCount) ReadTotal() string {
	lc.Lock()
	defer lc.Unlock()

	value := lc.TotalVotes.String()
	lc.TotalVotes = decimal.Zero

	return value
}

type AbilitiesSystem struct {
	battle *Battle
	// faction unique abilities
	factionUniqueAbilities map[uuid.UUID]map[string]*GameAbility // map[faction_id]map[identity]*Ability

	// gabs abilities (air craft, nuke, repair)
	battleAbilityPool *BattleAbilityPool

	bribe      chan *Contribution
	contribute chan *Contribution
	// location select winner list
	locationDeciders *LocationDeciders

	end       chan bool
	endGabs   chan bool
	liveCount *LiveCount
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
			gamelog.L.Error().Str("Battle ID", battle.ID).Err(err).Msg("unable to retrieve game abilities")
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
					wmAbility := GameAbility{
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
					battle.WarMachines[i].Abilities = []GameAbility{wmAbility}

					// store faction ability for price tracking
					factionAbilities[factionID][wmAbility.Identity] = &wmAbility
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

				wmAbility := GameAbility{
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
				abilities[wmAbility.Identity] = &wmAbility
			}
			factionAbilities[factionID] = abilities
		}

		// initialise user vote map in gab ability pool
		userContributeMap[factionID] = &UserContribution{
			contributionMap: map[uuid.UUID]decimal.Decimal{},
		}
	}

	as := &AbilitiesSystem{
		bribe:                  make(chan *Contribution),
		contribute:             make(chan *Contribution),
		battle:                 battle,
		factionUniqueAbilities: factionAbilities,
		battleAbilityPool:      battleAbilityPool,
		locationDeciders: &LocationDeciders{
			list: []uuid.UUID{},
		},
		liveCount: &LiveCount{
			TotalVotes: decimal.Zero,
		},
		end:     make(chan bool),
		endGabs: make(chan bool),
	}

	// broadcast faction unique ability
	for factionID, ga := range as.factionUniqueAbilities {
		if factionID.String() == server.ZaibatsuFactionID.String() {
			// broadcast the war machine abilities
			for identity, ability := range ga {
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineAbilitiesUpdated, identity)), []GameAbility{*ability})
			}
		} else {
			// broadcast faction ability
			for _, ability := range ga {
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionUniqueAbilitiesUpdated, factionID.String())), []GameAbility{*ability})
			}
		}
	}

	// init battle ability
	_, err := as.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to set up battle ability")
		return nil
	}

	// start ability cycle
	go as.FactionUniqueAbilityUpdater()

	// bribe cycle
	go as.StartGabsAbilityPoolCycle(false)

	return as
}

// ***********************************
// Faction Unique Ability Contribution
// ***********************************

// FactionUniqueAbilityUpdater update ability price every 10 seconds
func (as *AbilitiesSystem) FactionUniqueAbilityUpdater() {
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Msg("Panic! Panic! Panic! Panic at the FactionUniqueAbilityUpdater!")

			as.FactionUniqueAbilityUpdater()
		}
	}()
	minPrice := decimal.New(1, 18)

	main_ticker := time.NewTicker(1 * time.Second)

	live_vote_ticker := time.NewTicker(1 * time.Second)

	defer func() {
		main_ticker.Stop()
		live_vote_ticker.Stop()
	}()

	// start the battle
	for {
		select {
		case <-as.end:
			as.battle.stage = BattleStageEnd
			gamelog.L.Info().Msg("exiting ability price update")

			// get spoil of war
			sows, err := db.LastTwoSpoilOfWarAmount()
			if err != nil || len(sows) == 0 {
				gamelog.L.Error().Err(err).Msg("Failed to get last two spoil of war amount")
				return
			}

			// broadcast the spoil of war
			payload := []byte{byte(SpoilOfWarTick)}
			spoilOfWarStr := []string{}
			for _, sow := range sows {
				spoilOfWarStr = append(spoilOfWarStr, sow.String())
			}
			if len(spoilOfWarStr) > 0 {
				payload = append(payload, []byte(spoilOfWarStr[0]+"|0")...)
				as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeySpoilOfWarUpdated), payload)
			}

			gamelog.L.Info().Msgf("abilities system has been cleaned up: %s", as.battle.ID)

			// previously caused panic so wrapping in recover
			func() {
				defer func() {
					if err := recover(); err != nil {
						gamelog.L.Error().Interface("err", err).Msg("Panic! Panic! Panic! Panic at the cleaning up abilities channels!")
					}
				}()
				close(as.end)
				close(as.contribute)
			}()

			return
		case <-main_ticker.C:
			if as.battle == nil || as.battle.arena.currentBattle == nil || as.battle.arena.currentBattle.BattleNumber != as.battle.BattleNumber {
				return
			}
			// terminate ticker if battle mismatch
			if as.battle != as.battle.arena.currentBattle {
				main_ticker.Stop()
				live_vote_ticker.Stop()
				gamelog.L.Info().Msg("clean up ability price ticker update")
				return
			}

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

							event := &server.GameAbilityEvent{
								EventID:             ability.OfferingID,
								IsTriggered:         true,
								GameClientAbilityID: ability.GameClientAbilityID,
								ParticipantID:       ability.ParticipantID, // trigger on war machine
								WarMachineHash:      &ability.WarMachineHash,
							}

							// send message to game client, if ability trigger
							as.battle.arena.Message(
								"BATTLE:ABILITY",
								event,
							)

							bat := boiler.BattleAbilityTrigger{
								PlayerID:          null.StringFromPtr(nil),
								BattleID:          as.battle.ID,
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

							// get ability faction
							faction, err := db.FactionGet(ability.FactionID.String())
							if err != nil {
								gamelog.L.Error().Err(err).Msg("failed to get player faction")
							} else {

								//build notification
								gameNotification := &GameNotificationWarMachineAbility{
									Ability: &AbilityBrief{
										Label:    ability.Label,
										ImageUrl: ability.ImageUrl,
										Colour:   ability.Colour,
									},
								}

								// broadcast notification
								if ability.ParticipantID == nil {
									as.battle.arena.BroadcastGameNotificationAbility(GameNotificationTypeFactionAbility, GameNotificationAbility{
										Ability: gameNotification.Ability,
									})

								} else {
									// filled war machine detail
									for _, wm := range as.battle.WarMachines {
										if wm.ParticipantID == *ability.ParticipantID {
											gameNotification.WarMachine = &WarMachineBrief{
												ParticipantID: wm.ParticipantID,
												Hash:          wm.Hash,
												ImageUrl:      wm.Image,
												ImageAvatar:   wm.ImageAvatar,
												Name:          wm.Name,
												Faction: &FactionBrief{
													ID:         faction.ID,
													Label:      faction.Label,
													LogoBlobID: FactionLogos[faction.ID],
													Theme: &FactionTheme{
														Primary:    faction.PrimaryColor,
														Secondary:  faction.SecondaryColor,
														Background: faction.BackgroundColor,
													},
												},
											}
											break
										}
									}

									as.battle.arena.BroadcastGameNotificationWarMachineAbility(gameNotification)
								}
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
		case cont := <-as.contribute:
			if as.factionUniqueAbilities == nil {
				continue
			}
			if abilities, ok := as.factionUniqueAbilities[cont.factionID]; ok {
				// check ability exists
				if ability, ok := abilities[cont.abilityIdentity]; ok {

					// return early if battle stage is invalid
					if as.battle.stage != BattleStagStart {
						return
					}

					actualSupSpent, isTriggered := ability.SupContribution(as.battle.arena.ppClient, as.battle.ID, as.battle.BattleNumber, cont.userID, cont.amount)

					as.liveCount.AddSups(actualSupSpent)

					// sups contribution
					triggeredFlag := "0"
					if isTriggered {
						triggeredFlag = "1"
						// send message to game client, if ability trigger

						event := &server.GameAbilityEvent{
							IsTriggered:         true,
							GameClientAbilityID: ability.GameClientAbilityID,
							ParticipantID:       ability.ParticipantID, // trigger on war machine
							WarMachineHash:      &ability.WarMachineHash,
							EventID:             ability.OfferingID,
						}

						as.battle.arena.Message(
							"BATTLE:ABILITY",
							event,
						)

						bat := boiler.BattleAbilityTrigger{
							PlayerID:          null.StringFrom(cont.userID.String()),
							BattleID:          as.battle.ID,
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

						_, err = db.UserStatAddTotalAbilityTriggered(cont.userID.String())
						if err != nil {
							gamelog.L.Error().Str("player_id", cont.userID.String()).Err(err).Msg("failed to update user ability triggered amount")
						}

						// get player
						player, err := boiler.FindPlayer(gamedb.StdConn, cont.userID.String())
						if err != nil {
							gamelog.L.Error().Err(err).Msg("failed to get player")
						} else {

							// get user faction
							faction, err := db.FactionGet(player.FactionID.String)
							if err != nil {
								gamelog.L.Error().Err(err).Msg("failed to get player faction")
							} else {

								//build notification
								gameNotification := &GameNotificationWarMachineAbility{
									User: &UserBrief{
										ID:        cont.userID,
										Username:  player.Username.String,
										FactionID: player.FactionID.String,
										Faction: &FactionBrief{
											ID:         faction.ID,
											Label:      faction.Label,
											LogoBlobID: FactionLogos[faction.ID],
											Theme: &FactionTheme{
												Primary:    faction.PrimaryColor,
												Secondary:  faction.SecondaryColor,
												Background: faction.BackgroundColor,
											},
										},
									},
									Ability: &AbilityBrief{
										Label:    ability.Label,
										ImageUrl: ability.ImageUrl,
										Colour:   ability.Colour,
									},
								}

								// broadcast notification
								if ability.ParticipantID == nil {
									as.battle.arena.BroadcastGameNotificationAbility(GameNotificationTypeFactionAbility, GameNotificationAbility{
										Ability: gameNotification.Ability,
										User:    gameNotification.User,
									})

								} else {

									// filled war machine detail
									for _, wm := range as.battle.WarMachines {
										if wm.ParticipantID == *ability.ParticipantID {
											gameNotification.WarMachine = &WarMachineBrief{
												ParticipantID: wm.ParticipantID,
												Hash:          wm.Hash,
												ImageUrl:      wm.Image,
												ImageAvatar:   wm.ImageAvatar,
												Name:          wm.Name,
												Faction: &FactionBrief{
													ID:         faction.ID,
													Label:      faction.Label,
													LogoBlobID: FactionLogos[faction.ID],
													Theme: &FactionTheme{
														Primary:    faction.PrimaryColor,
														Secondary:  faction.SecondaryColor,
														Background: faction.BackgroundColor,
													},
												},
											}
											break
										}
									}

									as.battle.arena.BroadcastGameNotificationWarMachineAbility(gameNotification)
								}
							}
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

		case <-live_vote_ticker.C:
			if as.liveCount == nil {
				continue
			}

			total := as.liveCount.ReadTotal()

			// broadcast
			payload := []byte{byte(LiveVotingTick)}
			payload = append(payload, []byte(total)...)
			as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeyLiveVoteCountUpdated), payload)

			if as.battle.stage != BattleStagStart {
				continue
			}

			// get spoil of war
			sows, err := db.LastTwoSpoilOfWarAmount()
			if err != nil || len(sows) == 0 {
				gamelog.L.Error().Err(err).Msg("Failed to get last two spoil of war amount")
				continue
			}

			// broadcast the spoil of war
			payload = []byte{byte(SpoilOfWarTick)}
			spoilOfWarStr := []string{}
			for _, sow := range sows {
				spoilOfWarStr = append(spoilOfWarStr, sow.String())
			}
			payload = append(payload, []byte(strings.Join(spoilOfWarStr, "|"))...)
			as.battle.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeySpoilOfWarUpdated), payload)
		}
	}
}

// FactionUniqueAbilityPriceUpdate update target price on every tick
func (ga *GameAbility) FactionUniqueAbilityPriceUpdate(minPrice decimal.Decimal) bool {
	ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromFloat(0.9977))

	// if target price hit 1 sup, set it to 1 sup
	if ga.SupsCost.LessThanOrEqual(decimal.New(1, 18)) {
		ga.SupsCost = decimal.New(1, 18)
	}

	isTriggered := false

	// if the target price hit current price
	if ga.SupsCost.LessThanOrEqual(ga.CurrentSups) {
		// trigger the ability
		isTriggered = true

		// double the target price
		ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromInt(2))

		// reset current sups to zero
		ga.CurrentSups = decimal.Zero

	}

	// store updated price to db
	err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost, ga.CurrentSups)
	if err != nil {
		gamelog.L.Error().
			Str("ability_id", ga.ID.String()).
			Str("sups_cost", ga.SupsCost.StringFixed(4)).
			Str("current_sups", ga.CurrentSups.StringFixed(4)).
			Err(err).Msg("could not update faction ability cost")
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

	amount = amount.Truncate(0)

	// pay sup
	txid, err := ppClient.SpendSupMessage(passport.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             SupremacyBattleUserID,
		Amount:               amount.StringFixed(18),
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

	// update faction contribute
	err = db.FactionAddContribute(ga.FactionID.String(), amount)
	if err != nil {
		gamelog.L.Error().Str("txid", txid).Err(err).Msg("unable to update faction contribution")
	}

	amount = amount.Truncate(0)

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
			err = spoil.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to insert spoils")
			}
		} else {
			spoil.Amount = spoil.Amount.Add(amount)
			_, err = spoil.Update(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to insert spoil of war")
			}
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
		err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost, ga.CurrentSups)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to insert faction ability sup cost update")
			return amount, false
		}
		return amount, false
	}
	// increase price as the twice amount for normal value
	ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromInt(2))
	ga.CurrentSups = decimal.Zero

	// store updated price to db
	err = db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ga.ID, ga.SupsCost, ga.CurrentSups)
	if err != nil {
		gamelog.L.Error().
			Str("ability_id", ga.ID.String()).
			Str("sups_cost", ga.SupsCost.StringFixed(4)).
			Str("current_sups", ga.CurrentSups.StringFixed(4)).
			Err(err).Msg("could not update faction ability cost")
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
	deadlock.RWMutex
	contributionMap map[uuid.UUID]decimal.Decimal
}

type BattleAbilityPool struct {
	Stage *GabsBribeStage

	BattleAbility *server.BattleAbility
	Abilities     map[uuid.UUID]*GameAbility // faction ability current, change on every bribing cycle

	TriggeredFactionID uuid.UUID
}

type LocationSelectAnnouncement struct {
	GameAbility GameAbility `json:"game_ability"`
	EndTime     time.Time   `json:"end_time"`
}

// StartGabsAbilityPoolCycle
func (as *AbilitiesSystem) StartGabsAbilityPoolCycle(resume bool) {
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Msg("Panic! Panic! Panic! Panic at the StartGabsAbilityPoolCycle!")

			as.StartGabsAbilityPoolCycle(true)
		}
	}()

	// initial a ticker for current battle
	main_ticker := time.NewTicker(1 * time.Second)
	price_ticker := time.NewTicker(1 * time.Second)
	progress_ticker := time.NewTicker(1 * time.Second)

	defer func() {
		price_ticker.Stop()
		main_ticker.Stop()
	}()

	end_progress := make(chan bool)

	// start voting stage
	if !resume {
		as.battleAbilityPool.Stage.Phase = BribeStageBribe
		as.battleAbilityPool.Stage.EndTime = time.Now().Add(BribeDurationSecond * time.Second)
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)
	}
	bn := as.battle.BattleNumber

	go func() {
		defer progress_ticker.Stop()
		for {
			select {
			case <-end_progress:
				close(end_progress)
				return
			case <-progress_ticker.C:
				if as.battle == nil || as.battle.arena.currentBattle == nil || as.battle.arena.currentBattle.BattleNumber != bn {
					close(end_progress)
					return
				}
				as.BattleAbilityProgressBar()
			}
		}
	}()

	// start ability pool cycle
	for {
		select {
		// wait for next tick
		case <-as.endGabs:
			end_progress <- true
			close(as.endGabs)
			return
		case <-main_ticker.C:
			if as.battle == nil || as.battle.arena.currentBattle == nil || as.battle.arena.currentBattle.BattleNumber != bn {
				return
			}
			// terminate ticker if battle mismatch
			if as.battle != as.battle.arena.currentBattle {
				gamelog.L.Info().Msg("clean up mismatch bribing ticker")
				as.endGabs <- true
				continue
			}
			// check phase
			stage := as.battle.stage
			// exit the loop, when battle is ended
			if stage == BattleStageEnd {
				// change phase to hold and broadcast to user
				as.battleAbilityPool.Stage.Phase = BribeStageHold
				as.battleAbilityPool.Stage.EndTime = time.Now().AddDate(1, 0, 0) // HACK: set end time to far future to implement infinite time
				as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

				// stop all the ticker and exit the loop
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
				currentUserID, nextUserID, ok := as.nextLocationDeciderGet()
				if !ok {

					if as == nil {
						gamelog.L.Error().Msg("abilities are nil")
						continue
					}
					if as.battleAbilityPool == nil {
						gamelog.L.Error().Msg("ability pool is nil")
						continue
					}

					if as.battleAbilityPool.Abilities == nil {
						gamelog.L.Error().Msg("abilities map in battle ability pool is nil")
						continue
					}

					ability := as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID]

					// broadcast no ability
					as.battle.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
						Type: LocationSelectTypeCancelledNoPlayer,
						Ability: &AbilityBrief{
							Label:    ability.Label,
							ImageUrl: ability.ImageUrl,
							Colour:   ability.Colour,
						},
					})

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

				if as == nil {
					gamelog.L.Error().Msg("abilities are nil")
					continue
				}
				if as.battleAbilityPool == nil {
					gamelog.L.Error().Msg("ability pool is nil")
					continue
				}
				if as.battleAbilityPool.Abilities == nil {
					gamelog.L.Error().Msg("abilities map in battle ability pool is nil")
					continue
				}

				ab, ok := as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID]
				if !ok {
					gamelog.L.Error().
						Str("triggered faction id", as.battleAbilityPool.TriggeredFactionID.String()).
						Msg("nothing for triggered faction id")
					continue

				}

				notification := &GameNotificationLocationSelect{
					Type: LocationSelectTypeFailedTimeout,
					Ability: &AbilityBrief{
						Label:    ab.Label,
						ImageUrl: ab.ImageUrl,
						Colour:   ab.Colour,
					},
				}

				// get current player
				currentPlayer, err := BuildUserDetailWithFaction(currentUserID)
				if err == nil {
					notification.CurrentUser = currentPlayer
				}

				// get next player
				nextPlayer, err := BuildUserDetailWithFaction(nextUserID)
				if err == nil {
					notification.NextUser = nextPlayer
				}
				go as.battle.arena.BroadcastGameNotificationLocationSelect(notification)

				// extend location select phase duration
				as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
				// broadcast stage to frontend
				go as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

				// broadcast the announcement to the next location decider
				go as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, nextUserID)), &LocationSelectAnnouncement{
					GameAbility: *as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
					EndTime:     as.battleAbilityPool.Stage.EndTime,
				})

			// at the end of cooldown phase
			// random choose a battle ability for next bribing session
			case BribeStageCooldown:

				// change bribing phase
				as.battleAbilityPool.Stage.Phase = BribeStageBribe
				as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(BribeDurationSecond) * time.Second)
				// broadcast stage to frontend
				go as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

				continue
			default:
				gamelog.L.Error().Msg("hit default case switch on abilities loop")
			}
		case <-price_ticker.C:
			if as.battle == nil || as.battle.arena.currentBattle == nil || as.battle.arena.currentBattle.BattleNumber != bn {
				return
			}
			as.BattleAbilityPriceUpdater()
		case cont := <-as.bribe:
			if as.battle == nil || as.battle.arena.currentBattle == nil || as.battle.arena.currentBattle.BattleNumber != bn {
				return
			}

			// skip, if the bribe stage is incorrect
			if as.battleAbilityPool == nil || as.battleAbilityPool.Stage == nil || as.battleAbilityPool.Stage.Phase != BribeStageBribe {
				continue
			}

			if factionAbility, ok := as.battleAbilityPool.Abilities[cont.factionID]; ok {

				// contribute sups
				actualSupSpent, abilityTriggered := factionAbility.SupContribution(as.battle.arena.ppClient, as.battle.ID, as.battle.BattleNumber, cont.userID, cont.amount)

				as.liveCount.AddSups(actualSupSpent)

				if abilityTriggered {
					// generate location select order list
					as.locationDecidersSet(as.battle.ID, cont.factionID, cont.userID)

					// change bribing phase to location select
					as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
					as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)

					// broadcast stage change
					as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

					// send message to the user who trigger the ability
					as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, cont.userID)), &LocationSelectAnnouncement{
						GameAbility: *as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
						EndTime:     as.battleAbilityPool.Stage.EndTime,
					})

					notification := GameNotificationAbility{
						Ability: &AbilityBrief{
							Label:    factionAbility.Label,
							ImageUrl: factionAbility.ImageUrl,
							Colour:   factionAbility.Colour,
						},
					}
					// get player
					currentUser, err := BuildUserDetailWithFaction(cont.userID)
					if err == nil {
						notification.User = currentUser
					}
					as.battle.arena.BroadcastGameNotificationAbility(GameNotificationTypeBattleAbility, notification)

					// broadcast the latest result progress bar, when ability is triggered
					as.BroadcastAbilityProgressBar()
				}
			}
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
		gamelog.L.Error().Err(err).Msg("FactionBattleAbilityGet failed to retrieve shit")
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
func (as *AbilitiesSystem) locationDecidersSet(battleID string, factionID uuid.UUID, triggerByUserID ...uuid.UUID) {
	// set triggered faction id
	as.battleAbilityPool.TriggeredFactionID = factionID

	type userSupSpent struct {
		userID   uuid.UUID
		supSpent decimal.Decimal
	}

	playerList, err := db.PlayerFactionContributionList(battleID, factionID)
	if err != nil {
		gamelog.L.Error().Str("battle_id", battleID).Str("faction_id", factionID.String()).Err(err).Msg("failed to get player list")
	}

	// initialise location select list
	as.locationDeciders.list = []uuid.UUID{}

	for _, tid := range triggerByUserID {
		as.locationDeciders.list = append(as.locationDeciders.list, tid)
	}

	for _, pid := range playerList {
		exists := false
		for _, tid := range triggerByUserID {
			if pid == tid {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		as.locationDeciders.list = append(as.locationDeciders.list, pid)
	}
}

// nextLocationDeciderGet return the uuid of the next player to select the location for ability
func (as *AbilitiesSystem) nextLocationDeciderGet() (uuid.UUID, uuid.UUID, bool) {
	if as.locationDeciders == nil {
		gamelog.L.Error().Msg("nil check failed as.locationDeciders")

		return uuid.UUID(uuid.Nil), uuid.UUID(uuid.Nil), false
	}

	// clean up the location select list if there is no user left to select location
	if len(as.locationDeciders.list) <= 1 {
		gamelog.L.Error().Msg("no as.locationDeciders <= 1")
		as.locationDeciders.list = []uuid.UUID{}
		return uuid.UUID(uuid.Nil), uuid.UUID(uuid.Nil), false
	}

	currentUserID := as.locationDeciders.list[0]
	nextUserID := as.locationDeciders.list[1]

	// remove the first user from the list
	as.locationDeciders.list = as.locationDeciders.list[1:]

	return currentUserID, nextUserID, true
}

// ***********************************
// Ability Progression bar Broadcaster
// ***********************************

// 1 tick per second, each tick reduce 0.93304 of current price (drop the price to half in 10 second)

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
		ability.SupsCost = ability.SupsCost.Mul(decimal.NewFromFloat(0.93304))

		// cap minmum price at 1 sup
		if ability.SupsCost.Cmp(decimal.New(1, 18)) <= 0 {
			ability.SupsCost = decimal.New(1, 18)
		}

		// if ability not triggered, store ability's new target price to database, and continue
		if ability.SupsCost.Cmp(ability.CurrentSups) > 0 {
			// store updated price to db
			err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ability.ID, ability.SupsCost, ability.CurrentSups)
			if err != nil {
				gamelog.L.Error().
					Str("ability_id", ability.ID.String()).
					Str("sups_cost", ability.SupsCost.StringFixed(4)).
					Str("current_sups", ability.CurrentSups.StringFixed(4)).
					Err(err).Msg("could not update faction ability cost")
			}
			continue
		}

		// if ability triggered
		ability.SupsCost = ability.SupsCost.Mul(decimal.NewFromInt(2))
		ability.CurrentSups = decimal.Zero
		err := db.FactionAbilitiesSupsCostUpdate(context.Background(), gamedb.Conn, ability.ID, ability.SupsCost, ability.CurrentSups)
		if err != nil {
			gamelog.L.Error().
				Str("ability_id", ability.ID.String()).
				Str("sups_cost", ability.SupsCost.StringFixed(4)).
				Str("current_sups", ability.CurrentSups.StringFixed(4)).
				Err(err).Msg("could not update faction ability cost")
		}

		// broadcast the progress bar
		as.BroadcastAbilityProgressBar()

		// get player
		as.battle.arena.BroadcastGameNotificationAbility(GameNotificationTypeBattleAbility, GameNotificationAbility{
			Ability: &AbilityBrief{
				Label:    ability.Label,
				ImageUrl: ability.ImageUrl,
				Colour:   ability.Colour,
			},
		})

		// set location deciders list
		as.locationDecidersSet(as.battle.ID, factionID)

		// if no user online, enter cooldown and exit the loop
		if len(as.locationDeciders.list) == 0 {

			// broadcast no ability
			as.battle.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
				Type: LocationSelectTypeCancelledNoPlayer,
				Ability: &AbilityBrief{
					Label:    ability.Label,
					ImageUrl: ability.ImageUrl,
					Colour:   ability.Colour,
				},
			})

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

		notification := GameNotificationAbility{
			Ability: &AbilityBrief{
				Label:    ability.Label,
				ImageUrl: ability.ImageUrl,
				Colour:   ability.Colour,
			},
		}
		// get player
		currentUser, err := BuildUserDetailWithFaction(as.locationDeciders.list[0])
		if err == nil {
			notification.User = currentUser
		}
		as.battle.arena.BroadcastGameNotificationAbility(GameNotificationTypeBattleAbility, notification)

		// if there is user, assign location decider and exit the loop
		// change bribing phase to location select
		as.battleAbilityPool.Stage.Phase = BribeStageLocationSelect
		as.battleAbilityPool.Stage.EndTime = time.Now().Add(time.Duration(LocationSelectDurationSecond) * time.Second)
		// broadcast stage change
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeGabsBribeStageUpdateSubscribe), as.battleAbilityPool.Stage)

		// broadcast the announcement to the next location decider
		as.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeGabsBribingWinnerSubscribe, as.locationDeciders.list[0])), &LocationSelectAnnouncement{
			GameAbility: *as.battleAbilityPool.Abilities[as.battleAbilityPool.TriggeredFactionID],
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
	if as.battleAbilityPool == nil || as.battleAbilityPool.Abilities == nil {
		return
	}
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
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Msg("panic! panic! panic! Panic at the AbilityContribute!")
		}
	}()
	if as == nil || as.battle == nil || as.battle.stage != BattleStagStart || as.factionUniqueAbilities == nil {
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
func (as *AbilitiesSystem) FactionUniqueAbilitiesGet(factionID uuid.UUID) []GameAbility {
	abilities := []GameAbility{}
	for _, ga := range as.factionUniqueAbilities[factionID] {
		abilities = append(abilities, *ga)
	}

	if len(abilities) == 0 {
		return nil
	}

	return abilities
}

// FactionUniqueAbilityGet return the faction unique ability for the given faction
func (as *AbilitiesSystem) WarMachineAbilitiesGet(factionID uuid.UUID, hash string) []GameAbility {
	abilities := []GameAbility{}
	if as == nil {
		gamelog.L.Error().Str("factionID", factionID.String()).Str("hash", hash).Msg("nil pointer found as")
		return abilities
	}
	if as.factionUniqueAbilities == nil {
		gamelog.L.Error().Str("factionID", factionID.String()).Str("hash", hash).Msg("nil pointer found as.factionUniqueAbilities")
		return abilities
	}
	// NOTE: just pass down the faction unique abilities for now
	if fua, ok := as.factionUniqueAbilities[factionID]; ok {
		for h, ga := range fua {
			if h == hash {
				abilities = append(abilities, *ga)
			}
		}
	}

	if len(abilities) == 0 {
		return nil
	}

	return abilities
}

func (as *AbilitiesSystem) BribeGabs(factionID uuid.UUID, userID uuid.UUID, amount decimal.Decimal) {
	if as == nil || as.battle == nil || as.battle.stage != BattleStagStart {
		gamelog.L.Error().
			Bool("nil checks as", as == nil).
			Msg("unable to retrieve abilities for faction")
		return
	}

	if as.battleAbilityPool.Stage.Phase != BribeStageBribe {
		gamelog.L.Warn().
			Msg("unable to retrieve abilities for faction")
	}

	cont := &Contribution{
		factionID,
		userID,
		amount,
		"",
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				gamelog.L.Error().Interface("err", err).Msg("panic! panic! panic! Panic at the gabsbribe!")
			}
		}()

		as.bribe <- cont
	}()
}

func (as *AbilitiesSystem) BribeStageGet() *GabsBribeStage {
	if as.battleAbilityPool != nil {
		return as.battleAbilityPool.Stage
	}
	return nil
}

func (as *AbilitiesSystem) FactionBattleAbilityGet(factionID uuid.UUID) (GameAbility, error) {
	if as.battleAbilityPool == nil {
		return GameAbility{}, fmt.Errorf("battleAbilityPool is nil, fid: %s", factionID.String())
	}
	if as.battleAbilityPool.Abilities == nil {
		return GameAbility{}, fmt.Errorf("battleAbilityPool.Abilities is nil, fid: %s", factionID.String())
	}

	ability, ok := as.battleAbilityPool.Abilities[factionID]
	if !ok {
		gamelog.L.Warn().Str("func", "FactionBattleAbilityGet").Msg("unable to retrieve abilities for faction")
		return GameAbility{}, fmt.Errorf("game ability does not exist for faction %s", factionID.String())
	}

	return *ability, nil
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

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: ability.GameClientAbilityID,
		TriggeredOnCellX:    &x,
		TriggeredOnCellY:    &y,
		TriggeredByUserID:   &userID,
		TriggeredByUsername: &player.Username.String,
		EventID:             ability.OfferingID,
	}

	as.battle.calcTriggeredLocation(event)

	// trigger location select
	as.battle.arena.Message("BATTLE:ABILITY", event)

	bat := boiler.BattleAbilityTrigger{
		PlayerID:          null.StringFrom(userID.String()),
		BattleID:          as.battle.ID,
		FactionID:         ability.FactionID.String(),
		IsAllSyndicates:   true,
		AbilityLabel:      ability.Label,
		GameAbilityID:     ability.ID.String(),
		AbilityOfferingID: ability.OfferingID.String(),
	}
	err = bat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("battle_ability_trigger", bat).Err(err).Msg("Failed to record ability triggered")
	}

	_, err = db.UserStatAddTotalAbilityTriggered(userID.String())
	if err != nil {
		gamelog.L.Error().Str("player_id", userID.String()).Err(err).Msg("failed to update user ability triggered amount")
	}

	as.battle.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
		Type: LocationSelectTypeTrigger,
		X:    &x,
		Y:    &y,
		Ability: &AbilityBrief{
			Label:    ability.Label,
			ImageUrl: ability.ImageUrl,
			Colour:   ability.Colour,
		},
		CurrentUser: &UserBrief{
			ID:        userID,
			Username:  player.Username.String,
			FactionID: player.FactionID.String,
			Faction: &FactionBrief{
				ID:         faction.ID,
				Label:      faction.Label,
				LogoBlobID: FactionLogos[as.battleAbilityPool.TriggeredFactionID.String()],
				Theme: &FactionTheme{
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

func (as *AbilitiesSystem) End() {
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Msg("Panic! Panic! Panic! Panic at the abilities.End!")
			as.battle = nil
		}
	}()
	as.end <- true
	as.endGabs <- true
	as.battle = nil
}

func BuildUserDetailWithFaction(userID uuid.UUID) (*UserBrief, error) {
	userBrief := &UserBrief{}

	user, err := boiler.FindPlayer(gamedb.StdConn, userID.String())
	if err != nil {
		gamelog.L.Error().Str("player_id", userID.String()).Err(err).Msg("failed to get player from db")
		return nil, terror.Error(err)
	}

	userBrief.ID = userID
	userBrief.Username = user.Username.String

	if !user.FactionID.Valid {
		return userBrief, nil
	}

	userBrief.FactionID = user.FactionID.String

	faction, err := db.FactionGet(user.FactionID.String)
	if err != nil {
		gamelog.L.Error().Str("player_id", userID.String()).Str("faction_id", user.FactionID.String).Err(err).Msg("failed to get player faction from db")
		return userBrief, nil
	}

	userBrief.Faction = &FactionBrief{
		ID:         faction.ID,
		Label:      faction.Label,
		LogoBlobID: FactionLogos[faction.ID],
		Theme: &FactionTheme{
			Primary:    faction.PrimaryColor,
			Secondary:  faction.SecondaryColor,
			Background: faction.BackgroundColor,
		},
	}

	return userBrief, nil
}
