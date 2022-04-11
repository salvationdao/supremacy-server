package battle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sort"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type MultiplierTypeEnum string

const SPEND_AVERAGE MultiplierTypeEnum = "spend_average"
const MOST_SUPS_LOST MultiplierTypeEnum = "most_sups_lost"
const GAB_ABILITY MultiplierTypeEnum = "gab_ability"
const COMBO_BREAKER MultiplierTypeEnum = "combo_breaker"
const PLAYER_MECH MultiplierTypeEnum = "player_mech"
const HOURS_ONLINE MultiplierTypeEnum = "hours_online"
const SYNDICATE_WIN MultiplierTypeEnum = "syndicate_win"

type MultiplierSystem struct {
	multipliers map[string]*boiler.Multiplier
	players     map[string]map[*boiler.Multiplier]*boiler.UserMultiplier
	battle      *Battle
}

func NewMultiplierSystem(btl *Battle) *MultiplierSystem {
	ms := &MultiplierSystem{
		battle:      btl,
		players:     make(map[string]map[*boiler.Multiplier]*boiler.UserMultiplier),
		multipliers: make(map[string]*boiler.Multiplier),
	}
	ms.init()
	return ms
}

func (ms *MultiplierSystem) init() {

}

type TriggerDetails struct {
	FireCount  int
	PlayerIDs  []string
	FactionIDs []string
}

func PlayerMultipliers(playerID uuid.UUID, battleSeconds decimal.Decimal, specificBattleNumber ...int) ([]*Multiplier, string) {
	var total decimal.Decimal

	queries := []qm.QueryMod{
		boiler.UserMultiplierWhere.PlayerID.EQ(playerID.String()),
		boiler.UserMultiplierWhere.ExpiresAtBattleSeconds.GTE(battleSeconds),
		qm.Load(
			boiler.UserMultiplierRels.Multiplier,
		),
	}

	// only obtaining multiplier on specific battle
	if specificBattleNumber != nil && len(specificBattleNumber) > 0 {
		queries = append(queries, boiler.UserMultiplierWhere.FromBattleNumber.EQ(specificBattleNumber[0]))
	}

	usermultipliers, err := boiler.UserMultipliers(queries...).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msgf("unable to retrieve player multipliers")
		return []*Multiplier{}, "0"
	}

	multipliers := make([]*Multiplier, len(usermultipliers))
	value := decimal.Zero
	multiplicativeValue := decimal.Zero
	for i, m := range usermultipliers {
		multipliers[i] = &Multiplier{
			Key:              m.R.Multiplier.Key,
			Description:      m.R.Multiplier.Description,
			IsMultiplicative: m.R.Multiplier.IsMultiplicative,
			ExpiresInSeconds: m.ExpiresAtBattleSeconds.Sub(battleSeconds).IntPart(),
		}

		if !m.R.Multiplier.IsMultiplicative {
			multipliers[i].Value = m.Value.Shift(-1).String()
			value = value.Add(m.Value)
			continue
		}

		multipliers[i].Value = m.Value.String()
		multiplicativeValue = multiplicativeValue.Add(m.Value)
	}

	// set multiplicative to 1 if the value is zero
	if multiplicativeValue.Equal(decimal.Zero) {
		multiplicativeValue = decimal.NewFromInt(1)
	}

	total = value.Mul(multiplicativeValue)

	if playerID.String() == "294be3d5-03be-4daa-ac6e-b9b862f79ae6" {
		multipliers = append(multipliers, &Multiplier{
			Key:              "reece \U0001F9CB\U0001F9CB\U0001F9CB",
			Value:            "\U0001F9CB",
			Description:      "no bbt for reece",
			ExpiresInSeconds: 10000000000,
		})
	}

	return multipliers, total.Shift(-1).StringFixed(1)
}

func (ms *MultiplierSystem) getMultiplier(mtype, testString string, num int) (*boiler.Multiplier, bool) {
	multiplier, err := boiler.Multipliers(
		qm.Where(`multiplier_type = ?`, mtype),
		qm.And(`test_string = ?`, testString),
		qm.And(`test_number = ?`, num),
	).One(gamedb.StdConn)

	if err != nil {
		gamelog.L.Error().Str("m,type", mtype).Err(err).Msgf("unable to retrieve multiplier from database")
		return nil, false
	}
	return multiplier, true
}

func (ms *MultiplierSystem) end(btlEndInfo *BattleEndDetail) {
	ms.calculate(btlEndInfo)
}

type PlayerContribution struct {
	FactionID string
	Amount    decimal.Decimal
}

func (ms *MultiplierSystem) calculate(btlEndInfo *BattleEndDetail) {
	//fetch data
	//fetch contributions
	contributions, err := boiler.BattleContributions(qm.Where(`battle_id = ?`, ms.battle.ID)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Panic().Err(err).Msgf("unable to retrieve trigger information from database")
	}

	//fetch triggers
	triggers, err := boiler.BattleAbilityTriggers(
		qm.Where(`battle_id = ?`, ms.battle.ID),
		qm.And(`is_all_syndicates = true`),
		qm.OrderBy(`triggered_at DESC`),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Panic().Err(err).Msgf("unable to retrieve trigger information from database")
	}

	//sort triggers by player / faction
	fired := make(map[string]*TriggerDetails)
	for _, trigger := range triggers {
		td, ok := fired[trigger.AbilityLabel]
		if !ok {
			td = &TriggerDetails{FireCount: 0, PlayerIDs: []string{}, FactionIDs: []string{}}
			fired[trigger.AbilityLabel] = td
		}
		td.FireCount++
		if trigger.PlayerID.Valid {
			td.PlayerIDs = append(td.PlayerIDs, trigger.PlayerID.String)
		}
		td.FactionIDs = append(td.FactionIDs, trigger.FactionID)
	}

	// create new multipliers map
	newMultipliers := make(map[string]map[string]*boiler.Multiplier)

	// average spend multipliers test
	total := decimal.New(0, 18)
	sums := map[string]*PlayerContribution{}
	factions := map[string]string{}
	abilitySums := map[string]map[string]decimal.Decimal{}

	isGabs := map[string]bool{}

	for _, contribution := range contributions {
		if contribution.PlayerID == server.XsynTreasuryUserID.String() {
			continue
		}
		factions[contribution.PlayerID] = contribution.FactionID
		if _, ok := sums[contribution.PlayerID]; !ok {
			sums[contribution.PlayerID] = &PlayerContribution{
				FactionID: contribution.FactionID,
				Amount:    decimal.New(0, 18),
			}
		}
		sums[contribution.PlayerID].Amount = sums[contribution.PlayerID].Amount.Add(contribution.Amount)
		total = total.Add(contribution.Amount)

		isGabs[contribution.AbilityOfferingID] = contribution.IsAllSyndicates

		if _, ok := abilitySums[contribution.AbilityOfferingID]; !ok {
			abilitySums[contribution.AbilityOfferingID] = map[string]decimal.Decimal{}
		}
		if _, ok := abilitySums[contribution.AbilityOfferingID][contribution.PlayerID]; !ok {
			abilitySums[contribution.AbilityOfferingID][contribution.PlayerID] = decimal.New(0, 18)
		}
		amnt := abilitySums[contribution.AbilityOfferingID][contribution.PlayerID]
		abilitySums[contribution.AbilityOfferingID][contribution.PlayerID] = amnt.Add(contribution.Amount)
	}

	multipliers, err := boiler.Multipliers().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Msg("unable to retrieve multipliers from db")
		return
	}

	removedPlayerAmount := 0
	// citizen tag
	playerAmountList := []struct {
		playerID  string
		factionID string
		amount    decimal.Decimal
	}{}
	for playerID, pc := range sums {

		if pc.Amount.LessThan(decimal.New(3, 18)) {
			removedPlayerAmount += 1
			continue
		}

		playerAmountList = append(playerAmountList, struct {
			playerID  string
			factionID string
			amount    decimal.Decimal
		}{playerID, pc.FactionID, pc.Amount})
	}

	gamelog.L.Info().Msgf("%d players are removed from the list, due to spending less than 3 sups", removedPlayerAmount)

	//caching this value in memory- instantiating variable
	citizenMulti := &boiler.Multiplier{}
	totalLength := len(playerAmountList)
	if totalLength > 0 {
		// sort the total
		sort.Slice(playerAmountList, func(i, j int) bool { return playerAmountList[i].amount.GreaterThan(playerAmountList[j].amount) })

		// top 80% of contributors will become citizens
		citizenAmount := totalLength * 95 / 100
		if citizenAmount == 0 {
			citizenAmount = 1
		}

		// top 95% of contributors and their faction win, will become citizens
		winningFactionCitizenAmount := totalLength * 95 / 100
		if winningFactionCitizenAmount == 0 {
			winningFactionCitizenAmount = 1
		}

		// top 50% of contributors will become supporters
		supportAmount := totalLength * 40 / 100
		if supportAmount == 0 {
			supportAmount = 1
		}

		// top 25% of contributors will become contributors
		contributorAmount := totalLength * 20 / 100
		if contributorAmount == 0 {
			contributorAmount = 1
		}

		// top 10% of contributors will become super contributors
		superContributorAmount := totalLength * 10 / 100
		if superContributorAmount == 0 {
			superContributorAmount = 1
		}

		for _, m := range multipliers {
			if m.MultiplierType == "spend_average" {
				switch m.Key {
				case "citizen":
					//this battle's citizen multiplier
					citizenMulti = m
					for i := 0; i < winningFactionCitizenAmount; i++ {
						// skip, if the user is not from the winning faction and fall into 80% - 95% range
						if i >= citizenAmount && playerAmountList[i].factionID != btlEndInfo.WinningFaction.ID {
							continue
						}

						if _, ok := newMultipliers[playerAmountList[i].playerID]; !ok {
							newMultipliers[playerAmountList[i].playerID] = map[string]*boiler.Multiplier{}
						}

						//still creating new multiplier for players who already have it (thus far, still stacks)
						newMultipliers[playerAmountList[i].playerID][m.ID] = m
					}
				case "supporter":
					for i := 0; i < supportAmount; i++ {
						if _, ok := newMultipliers[playerAmountList[i].playerID]; !ok {
							newMultipliers[playerAmountList[i].playerID] = map[string]*boiler.Multiplier{}
						}
						newMultipliers[playerAmountList[i].playerID][m.ID] = m
					}
				case "contributor":
					for i := 0; i < contributorAmount; i++ {
						if _, ok := newMultipliers[playerAmountList[i].playerID]; !ok {
							newMultipliers[playerAmountList[i].playerID] = map[string]*boiler.Multiplier{}
						}
						newMultipliers[playerAmountList[i].playerID][m.ID] = m
					}
				case "super contributor":
					for i := 0; i < superContributorAmount; i++ {
						if _, ok := newMultipliers[playerAmountList[i].playerID]; !ok {
							newMultipliers[playerAmountList[i].playerID] = map[string]*boiler.Multiplier{}
						}
						newMultipliers[playerAmountList[i].playerID][m.ID] = m
					}
				}

			}
		}
	}

	//// checking if they have citizen this round
	//citizenIDs, err := db.CitizenPlayerIDs(ms.battle.BattleNumber)
	//if err != nil {
	//	gamelog.L.Error().Str("battle number", strconv.Itoa(ms.battle.BattleNumber)).Err(err).Msg("Failed to get citizen ids for next round")
	//}
	//
	//for _, id := range citizenIDs {
	//	// if citizen in last round is not in the list, add them
	//	if _, ok := newMultipliers[id.String()]; !ok {
	//		//adding to newMultiplier map- key = player ID value = new map of Multiplier.
	//		newMultipliers[id.String()] = map[string]*boiler.Multiplier{}
	//
	//		//setting the citizenMultiplier equal to the cached citizen Multi which is equal to the current battle's multi
	//		newMultipliers[id.String()][citizenMulti.ID] = citizenMulti
	//	}
	//}

	// fool and his money
	for abilityID, abPlayers := range abilitySums {
		if !isGabs[abilityID] {
			continue
		}
		topPlayerAmount := decimal.New(0, 18)
		topPlayerID := ""
		for playerID, amount := range abPlayers {
			if amount.GreaterThan(topPlayerAmount) {
				topPlayerAmount = amount
				topPlayerID = playerID
			}
		}
		if topPlayerID != "" {
			abilityTrigger, err := boiler.BattleAbilityTriggers(
				qm.Where(fmt.Sprintf("%s = ?", boiler.BattleAbilityTriggerColumns.AbilityOfferingID), abilityID)).One(gamedb.StdConn)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Str("ability_offering_id", abilityID).Err(err).Msg("unable to retrieve trigger event for ability contribution")
				continue
			}
			if abilityTrigger.FactionID != factions[topPlayerID] {
				if _, ok := newMultipliers[topPlayerID]; !ok {
					newMultipliers[topPlayerID] = map[string]*boiler.Multiplier{}
				}
				m, ok := ms.getMultiplier("most_sups_lost", "", 0)
				if !ok {
					gamelog.L.Error().Str("most_sups_lost", topPlayerID).Err(err).Msg("unable to retrieve 'a fool and his money' from multipliers. maybe this code needs to be removed?")
					continue
				}
				if _, ok := newMultipliers[topPlayerID]; !ok {
					newMultipliers[topPlayerID] = make(map[string]*boiler.Multiplier)
				}
				newMultipliers[topPlayerID][m.ID] = m
			}
		}
	}

	gab_triggers, err := boiler.BattleAbilityTriggers(
		qm.Where(`battle_id = ?`, ms.battle.ID),
		qm.OrderBy(`triggered_at DESC`),
		qm.Where(`is_all_syndicates = true`),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to retrieve trigger event for battle")
	}

	for _, tr := range gab_triggers {
		if tr.PlayerID.String != "" {
			_, ok := newMultipliers[tr.PlayerID.String]
			if !ok {
				// skip if player not a citizen
				// newMultipliers[td.PlayerIDs[0]] = map[*boiler.Multiplier]bool{}
				continue
			}
			m1, m1ok := ms.getMultiplier("gab_ability", tr.AbilityLabel, 1)
			if !m1ok {
				continue
			}

			if _, ok := newMultipliers[tr.PlayerID.String]; !ok {
				newMultipliers[tr.PlayerID.String] = make(map[string]*boiler.Multiplier)
			}
			newMultipliers[tr.PlayerID.String][m1.ID] = m1
		}
	}

	// last three gab abilities
outer:
	for triggerLabel, td := range fired {

		m3, m3ok := ms.getMultiplier("gab_ability", triggerLabel, 1)
		if !m3ok {
			continue
		}
		if len(td.PlayerIDs) < 3 {
			continue
		}
		if td.FireCount < 3 {
			for i := 1; i < len(td.PlayerIDs); i++ {
				if td.PlayerIDs[i] != td.PlayerIDs[i-1] {
					continue outer
				}
			}

			triggers, err := boiler.BattleAbilityTriggers(
				qm.Where(`is_all_syndicates = true`),
				qm.And(`ability_label = ?`, triggerLabel),
				qm.OrderBy(`triggered_at DESC`),
				qm.Limit(3),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msgf("unable to retrieve last three triggers information from database")
				continue outer
			}
			for i := 1; i < len(triggers); i++ {
				if triggers[i].PlayerID != triggers[i-1].PlayerID {
					continue outer
				}
			}
			//if it makes it here, it's because it was the last 3
			gamelog.L.Info().Interface("td.PlayerIds", td.PlayerIDs).Str("triggerLabel", triggerLabel).Msg("someone did the last 3!")
			if _, ok := newMultipliers[td.PlayerIDs[0]]; !ok {
				newMultipliers[td.PlayerIDs[0]] = map[string]*boiler.Multiplier{}
			}
			newMultipliers[td.PlayerIDs[0]][m3.ID] = m3
		} else {
			if len(td.PlayerIDs) < 3 {
				return
			}
			for i := 1; i < len(td.PlayerIDs); i++ {
				if td.PlayerIDs[i] != td.PlayerIDs[i-1] {
					continue outer
				}
			}
			//if it makes it here, it's because it was the last 3
			gamelog.L.Info().Interface("td.PlayerIds", td.PlayerIDs).Msg("someone did the last 3!")

			if _, ok := newMultipliers[td.PlayerIDs[0]]; !ok {
				newMultipliers[td.PlayerIDs[0]] = map[string]*boiler.Multiplier{}
			}
			newMultipliers[td.PlayerIDs[0]][m3.ID] = m3
		}

	}

	// check for syndicate wins
	lastWin, err := boiler.BattleWins(
		boiler.BattleWinWhere.BattleID.EQ(ms.battle.ID),
		qm.Limit(1),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to retrieve last win")
	}

	lastWins := []struct {
		BattleID  string            `boil:"battle_id"`
		FactionID string            `boil:"faction_id"`
		OwnerIDs  types.StringArray `boil:"owner_ids"`
	}{}

	err = boiler.NewQuery(
		qm.Select("battle_id, faction_id, array_agg(owner_id) as owner_ids, max(created_at)"),
		qm.From(boiler.TableNames.BattleWins),
		qm.GroupBy("battle_id, faction_id"),
		qm.OrderBy(`max(created_at) DESC`),
		qm.Limit(3),
	).Bind(context.Background(), gamedb.StdConn, &lastWins)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to retrieve last wins")
	}
	// set syndicate win

	hatTrick := true
	for i := 1; i < len(lastWins); i++ {
		if lastWins[i].FactionID != lastWins[i-1].FactionID {
			hatTrick = false
			break
		}
	}

	m1, _ := ms.getMultiplier("syndicate_win", "", 1)
	m3, _ := ms.getMultiplier("syndicate_win", "", 3)

	ms.battle.users.Range(func(bu *BattleUser) bool {
		if bu.FactionID == lastWin.FactionID {
			if _, ok := newMultipliers[bu.ID.String()]; !ok {
				// newMultipliers[bu.ID.String()] = map[*boiler.Multiplier]bool{}
				return true
			}
			newMultipliers[bu.ID.String()][m1.ID] = m1
			if hatTrick {
				newMultipliers[bu.ID.String()][m3.ID] = m3
				return true
			}
		}
		return true
	})

	// mech owner multipliers

winwar:
	for _, wm := range btlEndInfo.WinningWarMachines {
		if _, ok := newMultipliers[wm.OwnedByID]; !ok {
			// newMultipliers[wm.OwnedByID] = map[*boiler.Multiplier]bool{}
			continue
		}

		m1, ok := ms.getMultiplier("player_mech", "", 1)
		if !ok {
			gamelog.L.Error().Str("playerID", wm.OwnedByID).Err(err).Msg("unable to retrieve 'player_mech / mech win x1' from multipliers. maybe this code needs to be removed?")
			continue
		}
		if _, ok := newMultipliers[wm.OwnedByID]; !ok {
			newMultipliers[wm.OwnedByID] = make(map[string]*boiler.Multiplier)
		}
		newMultipliers[wm.OwnedByID][m1.ID] = m1

		if hatTrick {
			if len(lastWins) < 3 {
				gamelog.L.Error().Interface("lastwins", lastWins).Msg("last wins is less than 3 - this should never happen")
				continue winwar
			}
			for _, lastWinItem := range lastWins {
				found := false
				for _, lastWinOwnerID := range lastWinItem.OwnerIDs {
					if lastWinOwnerID == wm.OwnedByID {
						found = true
						break
					}
				}
				if !found {
					continue winwar
				}
			}
		}

		m3, ok := ms.getMultiplier("player_mech", "", 3)
		if !ok {
			gamelog.L.Error().Str("playerID", wm.OwnedByID).Err(err).Msg("unable to retrieve 'player_mech / mech win x3' from multipliers. maybe this code needs to be removed?")
			continue
		}

		if _, ok := newMultipliers[wm.OwnedByID]; !ok {
			newMultipliers[wm.OwnedByID] = make(map[string]*boiler.Multiplier)
		}
		newMultipliers[wm.OwnedByID][m3.ID] = m3
	}

	// insert multipliers
	playersWithCitizenAlready := make(map[string]bool)
	battleEndSeconds := ms.battle.battleSeconds()
	for pid, mlts := range newMultipliers {
		for multiID, m := range mlts {
			// if it is a citizen multi
			if citizenMulti.ID == multiID {
				if _, ok := playersWithCitizenAlready[pid]; ok {
					continue
				}
				playersWithCitizenAlready[pid] = true

				//um, err := boiler.UserMultipliers(
				//	//setting query conditions
				//	// check the player has citizen multi that ends this round
				//	boiler.UserMultiplierWhere.PlayerID.EQ(pid),
				//	boiler.UserMultiplierWhere.UntilBattleNumber.GT(ms.battle.BattleNumber),
				//	boiler.UserMultiplierWhere.FromBattleNumber.LTE(ms.battle.BattleNumber),
				//	boiler.UserMultiplierWhere.MultiplierID.EQ(citizenMulti.ID),
				//).One(gamedb.StdConn)
				//if err != nil && !errors.Is(err, sql.ErrNoRows) {
				//	gamelog.L.Error().Str("player id", pid).Err(err).Msg("Unable to get player citizen multiplier from last round")
				//	continue
				//}
				//
				//if um != nil {
				//	// if we get user multiplier back, update the db untilBattleNumber to +2, extending the duration +1 battle.
				//	um.UntilBattleNumber = um.UntilBattleNumber + 1
				//	_, err = um.Update(gamedb.StdConn, boil.Infer())
				//	if err != nil {
				//		gamelog.L.Error().Str("player id", pid).Err(err).Msg("Unable to extend player citizen multi")
				//		continue
				//	}
				//	//continues the loop, does not put extra multiplier in- does not stack
				//	continue
				//}
			}

			mlt := &boiler.UserMultiplier{
				PlayerID:                pid,
				FromBattleNumber:        ms.battle.BattleNumber,
				UntilBattleNumber:       ms.battle.BattleNumber + m.ForGames,
				MultiplierID:            m.ID,
				Value:                   m.Value,
				ObtainedAtBattleSeconds: battleEndSeconds,
				ExpiresAtBattleSeconds:  battleEndSeconds.Add(decimal.NewFromInt(int64(m.RemainSeconds))),
			}
			err := mlt.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("playerID", pid).Interface("user_multiplier", mlt).Err(err).Msg("unable to insert user multiplier at battle end")
				continue
			}
		}
	}
}
