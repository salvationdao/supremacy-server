package battle

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/passport"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
)

const EachMechIntroSecond = 3
const InitIntroSecond = 7

type AbilitiesSystem struct {
	battle *Battle
	// faction unique abilities
	factionUniqueAbilities map[uuid.UUID]map[string]*GameAbility // map[faction_id]map[identity]*Ability

	// gabs abilities (air craft, nuke, repair)
	factionGabsAbilities map[uuid.UUID]map[string]*GameAbility
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	factionAbilities := map[uuid.UUID]map[string]*GameAbility{}
	gabAbilities := map[uuid.UUID]map[string]*GameAbility{}

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
						Identity:            wm.ID,
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

					//TODO: if more mech abilities shit will fail
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
				}
				abilities[wmAbility.Identity] = wmAbility
			}
			factionAbilities[factionID] = abilities
		}

		// GABS abilities
		gabAbilities[factionID] = map[string]*GameAbility{}
		factionGabsAbilities, err := boiler.GameAbilities(qm.Where("faction_id = ?", factionID.String()), qm.And("battle_ability_id NOTNULL")).All(gamedb.StdConn)
		for _, ability := range factionGabsAbilities {
			supsCost, err := decimal.NewFromString(ability.SupsCost)
			if err != nil {
				gamelog.L.Err(err).Msg("Failed to ability sups cost to decimal")

				// set sups cost to initial price
				supsCost = decimal.New(1, 18)
			}

			currentSups, err := decimal.NewFromString(ability.CurrentSups)
			if err != nil {
				gamelog.L.Err(err).Msg("Failed to ability current sups to decimal")

				// set current sups to initial price
				currentSups = decimal.Zero
			}

			gabAbilities[factionID][ability.ID] = &GameAbility{
				ID:                  server.GameAbilityID(uuid.Must(uuid.FromString(ability.ID))),
				Identity:            ability.ID,
				GameClientAbilityID: byte(ability.GameClientAbilityID),
				ImageUrl:            ability.ImageURL,
				Description:         ability.Description,
				FactionID:           factionID,
				Label:               ability.Label,
				SupsCost:            supsCost,
				CurrentSups:         currentSups,
			}
		}
	}

	as := &AbilitiesSystem{
		battle:                 battle,
		factionUniqueAbilities: factionAbilities,
		factionGabsAbilities:   gabAbilities,
	}

	// calc the intro time, mech_amount *3 + 7 second
	waitDurationSecond := len(battle.WarMachines)*EachMechIntroSecond + InitIntroSecond

	// as.VoteCycle = as.VotingCycleTicker(waitDurationSecond)
	as.FactionUniqueAbilityUpdater(waitDurationSecond)

	return as
}

// FactionUniqueAbilityUpdater update ability price every 10 seconds
func (as *AbilitiesSystem) FactionUniqueAbilityUpdater(waitDurationSecond int) {
	// wait for mech intro
	time.Sleep(time.Duration(waitDurationSecond) * time.Second)

	minPrice := decimal.New(1, 18)

	// start the battle
	for _, abilities := range as.factionUniqueAbilities {

		// start ability price updater for each faction
		go func(abilities map[string]*GameAbility) {
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
				return
			}
		}(abilities)

	}
}

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
	return isTriggered
}

// FactionUniqueAbilityContribute contribute sups to specific faction unique ability
func (as *AbilitiesSystem) FactionUniqueAbilityContribute(factionID uuid.UUID, abilityIdentity string, userID server.UserID, amount decimal.Decimal) {
	if abilities, ok := as.factionUniqueAbilities[factionID]; ok {
		if ability, ok := abilities[abilityIdentity]; ok {
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

			// sups contribution
			if ability.SupContribution(as.battle.arena.ppClient, as.battle.ID.String(), userID, amount) {
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

// SupContribution contribute sups to specific game ability
func (ga *GameAbility) SupContribution(ppClient *passport.Passport, battleID string, userID server.UserID, amount decimal.Decimal) bool {
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
		return false
	}

	// update the current sups if not triggered
	if !isTriggered {
		ga.CurrentSups = ga.CurrentSups.Add(amount)
		return false
	}

	// otherwise update target price and reset the current price
	ga.SupsCost = ga.SupsCost.Mul(decimal.NewFromInt(2))
	ga.CurrentSups = decimal.Zero

	return isTriggered
}
