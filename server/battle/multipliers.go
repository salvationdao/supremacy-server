package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
		multipliers: make(map[string]*boiler.Multiplier),
		players:     make(map[string]map[*boiler.Multiplier]*boiler.UserMultiplier),
	}
	ms.init()
	return ms
}

func (ms *MultiplierSystem) init() {
	multipliers, err := boiler.Multipliers().All(gamedb.StdConn)
	for _, m := range multipliers {
		ms.multipliers[m.Key] = m
	}
	if err != nil {
		gamelog.L.Panic().Err(err).Msgf("unable to retrieve multipliers from database")
	}

	// fetch all active user multipliers

	usermultipliers, err := boiler.UserMultipliers(qm.Where(`until_battle_number >= ?`, ms.battle.BattleNumber)).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Msgf("unable to retrieve user's multipliers from database")
	}
	for _, m := range usermultipliers {
		pm, ok := ms.players[m.PlayerID]
		if !ok {
			pm = make(map[*boiler.Multiplier]*boiler.UserMultiplier)
			ms.players[m.PlayerID] = pm
		}
		mlt, ok := ms.multipliers[m.MultiplierID]
		if !ok {
			gamelog.L.Error().Err(err).Msgf("unable to retrieve multiplier - this should never happen")
		}
		pm[mlt] = m
	}
}

type TriggerDetails struct {
	FireCount  int
	PlayerIDs  []string
	FactionIDs []string
}

func (ms *MultiplierSystem) getMultiplier(mtype, testString string, num int) (*boiler.Multiplier, bool) {
	for _, m := range ms.multipliers {
		if m.MultiplierType == mtype && m.TestString == testString && m.TestNumber == num {
			return m, true
		}
	}
	return nil, false
}

func (ms *MultiplierSystem) end(btlEndInfo *BattleEndDetail) {
	ms.calculate(btlEndInfo)
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
		}
		td.FireCount++
		if trigger.PlayerID.Valid {
			td.PlayerIDs = append(td.PlayerIDs, trigger.PlayerID.String)
		}
		td.FactionIDs = append(td.FactionIDs, trigger.FactionID)
	}

	// create new multipliers map
	newMultipliers := make(map[string]map[*boiler.Multiplier]bool)

	// last gab ability last three gab abilities
outer:
	for triggerLabel, td := range fired {
		m1, m1ok := ms.getMultiplier("gab_ability", triggerLabel, 1)
		m3, m3ok := ms.getMultiplier("gab_ability", triggerLabel, 1)

		if !m1ok && !m3ok {
			continue
		}
		if m1ok {
			_, ok := newMultipliers[td.PlayerIDs[0]]
			if !ok {
				newMultipliers[td.PlayerIDs[0]] = map[*boiler.Multiplier]bool{}
			}
			newMultipliers[td.PlayerIDs[0]][m1] = true
		}

		if m3ok && td.FireCount < 3 {
			for i := 1; i < len(td.PlayerIDs); i++ {
				if td.PlayerIDs[i] != td.PlayerIDs[i-1] {
					continue outer
				}
			}

			triggers, err := boiler.BattleAbilityTriggers(
				qm.Where(`is_all_syndicates = true`),
				qm.And(`trigger_label = ?`, triggerLabel),
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
			gamelog.L.Info().Interface("td.PlayerIds", td.PlayerIDs).Msg("someone did the last 3!")
			newMultipliers[td.PlayerIDs[0]][m3] = true
		} else {
			for i := 1; i < len(td.PlayerIDs); i++ {
				if td.PlayerIDs[i] != td.PlayerIDs[i-1] {
					continue outer
				}
			}
			//if it makes it here, it's because it was the last 3
			gamelog.L.Info().Interface("td.PlayerIds", td.PlayerIDs).Msg("someone did the last 3!")
			newMultipliers[td.PlayerIDs[0]][m3] = true
		}

	}

	// check for syndicate wins

	lastWins, err := boiler.BattleWins(
		qm.Distinct(boiler.BattleWinColumns.BattleID),
		qm.OrderBy(boiler.BattleWinColumns.CreatedAt, "DESC"),
		qm.Limit(3),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to retrieve last 3 winning factions")
	}

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
		if bu.FactionID == lastWins[0].FactionID {
			if _, ok := newMultipliers[bu.ID.String()]; !ok {
				newMultipliers[bu.ID.String()] = map[*boiler.Multiplier]bool{}
			}
			newMultipliers[bu.ID.String()][m1] = true
			if hatTrick {
				newMultipliers[bu.ID.String()][m3] = true
			}
		}
		return true
	})

	// average spend multipliers test

	total := decimal.New(0, 18)
	sums := map[string]decimal.Decimal{}
	factions := map[string]string{}
	abilitySums := map[string]map[string]decimal.Decimal{}

	for _, contribution := range contributions {
		factions[contribution.PlayerID] = contribution.FactionID
		if _, ok := sums[contribution.PlayerID]; !ok {
			sums[contribution.PlayerID] = decimal.New(0, 18)
		}
		sums[contribution.PlayerID] = sums[contribution.PlayerID].Add(contribution.Amount)
		total = total.Add(contribution.Amount)

		if _, ok := abilitySums[contribution.AbilityOfferingID]; !ok {
			abilitySums[contribution.AbilityOfferingID] = map[string]decimal.Decimal{}
		}
		if _, ok := abilitySums[contribution.AbilityOfferingID][contribution.PlayerID]; !ok {
			abilitySums[contribution.AbilityOfferingID][contribution.PlayerID] = decimal.New(0, 18)
		}
		amnt := abilitySums[contribution.AbilityOfferingID][contribution.PlayerID]
		abilitySums[contribution.AbilityOfferingID][contribution.PlayerID] = amnt.Add(contribution.Amount)
	}

	for _, m := range ms.multipliers {
		if m.MultiplierType == "spend_average" {
			for playerID, amount := range sums {
				perc := total.Mul(decimal.New(100-int64(m.TestNumber), 18).Div(decimal.New(100, 18)))
				if amount.GreaterThanOrEqual(perc) {
					if _, ok := newMultipliers[playerID]; !ok {
						newMultipliers[playerID] = map[*boiler.Multiplier]bool{}
					}
					newMultipliers[playerID][m] = true
				}
			}
		}
	}

	// fool and his money
	for abilityID, abPlayers := range abilitySums {
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
					newMultipliers[topPlayerID] = map[*boiler.Multiplier]bool{}
				}
				m, ok := ms.getMultiplier("most_sups_lost", "", 0)
				if !ok {
					gamelog.L.Error().Str("most_sups_lost", topPlayerID).Err(err).Msg("unable to retrieve 'a fool and his money' from multipliers. maybe this code needs to be removed?")
					continue
				}
				newMultipliers[topPlayerID][m] = true
			}
		}
	}

	// mech owner multipliers

winwar:
	for _, wm := range btlEndInfo.WinningWarMachines {
		if _, ok := newMultipliers[wm.OwnedByID]; !ok {
			newMultipliers[wm.OwnedByID] = map[*boiler.Multiplier]bool{}
		}

		m1, ok := ms.getMultiplier("player_mech", "", 1)
		if !ok {
			gamelog.L.Error().Str("playerID", wm.OwnedByID).Err(err).Msg("unable to retrieve 'player_mech / mech win x1' from multipliers. maybe this code needs to be removed?")
			continue
		}
		newMultipliers[wm.OwnedByID][m1] = true

		for i := 0; i < 3; i++ {
			if lastWins[i].OwnerID != wm.OwnedByID {
				continue winwar
			}
		}

		m3, ok := ms.getMultiplier("player_mech", "", 3)
		if !ok {
			gamelog.L.Error().Str("playerID", wm.OwnedByID).Err(err).Msg("unable to retrieve 'player_mech / mech win x3' from multipliers. maybe this code needs to be removed?")
			continue
		}

		newMultipliers[wm.OwnedByID][m3] = true
	}

	// insert multipliers

	for pid, mlts := range newMultipliers {
		for m := range mlts {
			mlt := &boiler.UserMultiplier{
				PlayerID:          pid,
				FromBattleNumber:  ms.battle.BattleNumber,
				UntilBattleNumber: ms.battle.BattleNumber + m.ForGames,
				MultiplierID:      m.ID,
				Value:             m.Value,
			}
			err := mlt.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("playerID", pid).Interface("user_multiplier", mlt).Err(err).Msg("unable to insert user multiplier at battle end")
				continue
			}
		}
	}
}
