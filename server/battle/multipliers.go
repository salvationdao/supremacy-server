package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
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

}

type TriggerDetails struct {
	FireCount  int
	PlayerIDs  []string
	FactionIDs []string
}

func (ms *MultiplierSystem) PlayerMultipliers(playerID uuid.UUID) ([]*Multiplier, string) {
	var total decimal.Decimal

	usermultipliers, err := boiler.Multipliers(
		qm.InnerJoin("user_multipliers um on um.multiplier_id = multipliers.id"),
		qm.Where(`um.player_id = ?`, playerID.String()),
		qm.And(`um.until_battle_number >= ?`, ms.battle.BattleNumber)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msgf("unable to retrieve player multipliers")
		return []*Multiplier{}, "0"
	}

	multipliers := make([]*Multiplier, len(usermultipliers))
	for i, m := range usermultipliers {
		multipliers[i] = &Multiplier{
			Key:         m.Key,
			Value:       fmt.Sprintf("%sx", m.Value.Shift(-1).String()),
			Description: m.Description,
		}
		total = total.Add(m.Value)
	}

	if playerID.String() == "294be3d5-03be-4daa-ac6e-b9b862f79ae6" {
		multipliers = append(multipliers, &Multiplier{
			Key:         "reece 🍭🍭🍭",
			Value:       "-🍭",
			Description: "no lollipop for reece",
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
	/*
		SELECT battle_id, faction_id
		FROM "battle_wins"
		group by battle_id, faction_id
		order by max(created_at) asc
		limit 3;
	*/
	lastWins, err := boiler.BattleWins(
		qm.Select(boiler.BattleWinColumns.BattleID, boiler.BattleWinColumns.FactionID),
		qm.GroupBy(boiler.BattleWinColumns.BattleID+","+boiler.BattleWinColumns.FactionID),
		qm.OrderBy("MAX("+boiler.BattleWinColumns.CreatedAt+") DESC"),
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

	isGabs := map[string]bool{}

	for _, contribution := range contributions {
		factions[contribution.PlayerID] = contribution.FactionID
		if _, ok := sums[contribution.PlayerID]; !ok {
			sums[contribution.PlayerID] = decimal.New(0, 18)
		}
		sums[contribution.PlayerID] = sums[contribution.PlayerID].Add(contribution.Amount)
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

	for _, m := range multipliers {
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
