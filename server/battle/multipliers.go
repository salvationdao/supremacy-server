package battle

import (
	"database/sql"
	"errors"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

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
	players     map[string]map[string]*boiler.Multiplier
	battle      *Battle
}

func NewMultiplierSystem(btl *Battle) *MultiplierSystem {
	ms := &MultiplierSystem{
		battle:      btl,
		multipliers: make(map[string]*boiler.Multiplier),
		players:     make(map[string]map[string]*boiler.Multiplier),
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
	usermultipliers, err := boiler.UserMultipliers(qm.Where(`until_battle_number > ?`, ms.battle.battle.BattleNumber)).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Msgf("unable to retrieve user's multipliers from database")
	}
	for _, m := range usermultipliers {
		pm, ok := ms.players[m.PlayerID]
		if !ok {
			pm = make(map[string]*boiler.Multiplier)
			ms.players[m.PlayerID] = pm
		}
		pm[m.Multiplier] = ms.multipliers[m.Multiplier]
	}
}

type TriggerDetails struct {
	FireCount  int
	PlayerIDs  []string
	FactionIDs []string
}

func (ms *MultiplierSystem) getGabMultiplier(mtype, testString string, num int) (*boiler.Multiplier, bool) {
	for _, m := range ms.multipliers {
		if m.MultiplierType == mtype && m.TestString == testString && m.TestNumber == num {
			return m, true
		}
	}
	return nil, false
}

func (ms *MultiplierSystem) calculate(btlEndInfo *BattleEndDetail) {

	triggers, err := boiler.BattleAbilityTriggers(
		qm.Where(`battle_id = ?`, ms.battle.battle.ID),
		qm.And(`is_all_syndicates = true`),
		qm.OrderBy(`triggered_at DESC`),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Panic().Err(err).Msgf("unable to retrieve trigger information from database")
	}

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

	newMultipliers := make(map[string]map[*boiler.Multiplier]bool)

outer:
	for triggerLabel, td := range fired {
		m1, m1ok := ms.getGabMultiplier("gab_ability", triggerLabel, 1)
		m3, m3ok := ms.getGabMultiplier("gab_ability", triggerLabel, 1)

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

	m1, _ := ms.getGabMultiplier("syndicate_win", "", 1)
	m3, _ := ms.getGabMultiplier("syndicate_win", "", 3)

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

	// how many times fired

}
