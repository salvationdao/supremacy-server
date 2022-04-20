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
	return ms
}

type TriggerDetails struct {
	FireCount  int
	PlayerIDs  []string
	FactionIDs []string
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
	FactionID        string
	Amount           decimal.Decimal
	ContributorMulti decimal.Decimal
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
	newMultipliers := make(map[string]map[string][]*boiler.Multiplier)

	// average spend multipliers test
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
				FactionID:        contribution.FactionID,
				Amount:           decimal.New(0, 18),
				ContributorMulti: decimal.Zero,
			}
		}
		sums[contribution.PlayerID].Amount = sums[contribution.PlayerID].Amount.Add(contribution.Amount)

		sums[contribution.PlayerID].ContributorMulti = sums[contribution.PlayerID].ContributorMulti.Add(contribution.MultiAmount)

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

	citizenMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("citizen")).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to get citizen from multipliers table")
	}

	contributorMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("contributor")).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to get contributor from multipliers table")
	}

	totalLength := len(playerAmountList)
	if totalLength > 0 {
		for i := 0; i < totalLength; i++ {
			mult, ok := newMultipliers[playerAmountList[i].playerID]
			if !ok {
				mult = make(map[string][]*boiler.Multiplier)
			}

			mult[citizenMultiplier.ID] = append(mult[citizenMultiplier.ID], citizenMultiplier)
			newMultipliers[playerAmountList[i].playerID] = mult
		}
	}

	for playerID, sum := range sums {
		if sum.FactionID != btlEndInfo.WinningFaction.ID {
			continue
		}
		mult, ok := newMultipliers[playerID]
		if !ok {
			continue
		}

		copiedMulti := *contributorMultiplier

		copiedMulti.Value = sum.ContributorMulti.Mul(decimal.NewFromInt(10))

		mult[contributorMultiplier.ID] = append(mult[contributorMultiplier.ID], &copiedMulti)

		newMultipliers[playerID] = mult
	}

	// TODO: move this to it own function
	repairEvents, err := boiler.BattleHistories(
		boiler.BattleHistoryWhere.EventType.EQ(boiler.BattleEventPickup),
		boiler.BattleHistoryWhere.BattleID.EQ(ms.battle.BattleID),
		boiler.BattleHistoryWhere.RelatedID.IsNotNull(),
		qm.Load(boiler.BattleHistoryRels.WarMachineOne),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to get repair events")
	}

	repairContributorMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("grease monkey")).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to get repair events")
	}

	repairTriggerMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("field mechanic")).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to get repair events")
	}

	if repairContributorMultiplier != nil && repairTriggerMultiplier != nil {

		for _, repairEvent := range repairEvents {
			triggeredPlayer, err := boiler.BattleContributions(boiler.BattleContributionWhere.AbilityOfferingID.EQ(repairEvent.RelatedID.String), boiler.BattleContributionWhere.DidTrigger.EQ(true)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("event triggered", repairEvent.RelatedID.String).Err(err).Msg("Failed to get triggered player")
				continue
			}
			mechOwner, err := boiler.Players(boiler.PlayerWhere.ID.EQ(repairEvent.R.WarMachineOne.OwnerID)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to get mech owner contributors")
				continue
			}

			eventContributors, err := boiler.BattleContributions(
				qm.Select(boiler.BattleContributionColumns.PlayerID),
				boiler.BattleContributionWhere.AbilityOfferingID.EQ(repairEvent.RelatedID.String),
				boiler.BattleContributionWhere.PlayerID.NEQ(triggeredPlayer.PlayerID),
				boiler.BattleContributionWhere.FactionID.EQ(mechOwner.FactionID.String),
				qm.GroupBy(boiler.BattleContributionColumns.PlayerID),
			).All(gamedb.StdConn)

			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to find event contributors")
				continue
			}

			mult, ok := newMultipliers[triggeredPlayer.PlayerID]
			if !ok {
				mult = make(map[string][]*boiler.Multiplier)
			}
			mult[repairTriggerMultiplier.ID] = append(mult[repairTriggerMultiplier.ID], repairTriggerMultiplier)
			newMultipliers[triggeredPlayer.PlayerID] = mult

			for _, eventContributor := range eventContributors {
				mult, ok := newMultipliers[eventContributor.PlayerID]
				if !ok {
					mult = make(map[string][]*boiler.Multiplier)
				}
				mult[repairContributorMultiplier.ID] = append(mult[repairContributorMultiplier.ID], repairContributorMultiplier)
				newMultipliers[eventContributor.PlayerID] = mult
			}

		}
	}

	killedEvents, err := boiler.BattleHistories(
		boiler.BattleHistoryWhere.EventType.EQ(boiler.BattleEventKilled),
		boiler.BattleHistoryWhere.BattleID.EQ(btlEndInfo.BattleID),
		boiler.BattleHistoryWhere.RelatedID.IsNotNull(),
		qm.Load(boiler.BattleHistoryRels.WarMachineOne),
		qm.OrderBy(boiler.BattleHistoryColumns.RelatedID),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("event type", boiler.BattleEventKilled).Str("btlEndInfo.BattleID", btlEndInfo.BattleID).Err(err).Msg("Failed to get airstrike events")

	}

	airstrikeContributorMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("air support")).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get airstrike contributor from db")
	}

	airstrikeTriggerMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("air marshal")).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get airstrike trigger from db")
	}

	nukeContributorMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("now i am become death")).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get nuke contributor from db")
	}

	nukeTriggerMultiplier, err := boiler.Multipliers(boiler.MultiplierWhere.Key.EQ("destroyer of worlds")).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get nuke trigger from db")
	}

	if airstrikeContributorMultiplier != nil && airstrikeTriggerMultiplier != nil && nukeContributorMultiplier != nil && nukeTriggerMultiplier != nil {
		killStreak := 0
		var sameKilledEvents []boiler.BattleHistory = nil
		killedTotalLength := len(killedEvents)
		for i, killedEvent := range killedEvents {
			if i+1 < killedTotalLength {
				if killedEvent.RelatedID == killedEvents[i+1].RelatedID {
					sameKilledEvents = append(sameKilledEvents, *killedEvent)
					continue
				}
			}
			sameKilledEvents = append(sameKilledEvents, *killedEvent)

			triggeredPlayer, err := boiler.BattleContributions(boiler.BattleContributionWhere.AbilityOfferingID.EQ(killedEvent.RelatedID.String), boiler.BattleContributionWhere.DidTrigger.EQ(true)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("event triggered", killedEvent.RelatedID.String).Err(err).Msg("Failed to get triggered player")
				continue
			}
			for _, sameKilledEvent := range sameKilledEvents {
				mechOwner, err := boiler.Players(boiler.PlayerWhere.ID.EQ(sameKilledEvent.R.WarMachineOne.OwnerID)).One(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("WarMachineOne.OwnerID", sameKilledEvent.R.WarMachineOne.OwnerID).Err(err).Msg("Failed to get mech owner contributors")
					continue
				}

				if mechOwner.FactionID.String == triggeredPlayer.FactionID {
					killStreak -= 1
					continue
				}
				killStreak += 1
			}

			if killStreak > 0 {

				var contributedMulti boiler.Multiplier
				var triggeredMulti boiler.Multiplier

				switch triggeredPlayer.AbilityLabel {
				case "NUKE":
					triggeredMulti = *nukeTriggerMultiplier
					contributedMulti = *nukeContributorMultiplier
				case "AIRSTRIKE":
					triggeredMulti = *airstrikeTriggerMultiplier
					contributedMulti = *airstrikeContributorMultiplier
				}

				mult, ok := newMultipliers[triggeredPlayer.PlayerID]
				if ok {
					if killStreak > 3 {
						killStreak = 3
					}

					if killStreak > 1 {
						triggeredMulti.Value = triggeredMulti.Value.Mul(decimal.NewFromInt(int64(killStreak)))
					}

					mult[triggeredMulti.ID] = append(mult[triggeredMulti.ID], &triggeredMulti)
					newMultipliers[triggeredPlayer.PlayerID] = mult
				}

				eventContributors, err := boiler.BattleContributions(
					qm.Select(boiler.BattleContributionColumns.PlayerID),
					boiler.BattleContributionWhere.AbilityOfferingID.EQ(killedEvent.RelatedID.String),
					boiler.BattleContributionWhere.PlayerID.NEQ(triggeredPlayer.ID),
					boiler.BattleContributionWhere.DidTrigger.EQ(false),
					qm.GroupBy(boiler.BattleContributionColumns.PlayerID),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("ability_offering_id", killedEvent.RelatedID.String).Err(err).Msg("Failed to find event contributors")
					continue
				}
				for _, eventContributor := range eventContributors {
					mult, ok := newMultipliers[eventContributor.PlayerID]
					if !ok {
						mult = make(map[string][]*boiler.Multiplier)
					}

					mult[contributedMulti.ID] = append(mult[contributedMulti.ID], &contributedMulti)
					newMultipliers[eventContributor.PlayerID] = mult
				}
			}
			killStreak = 0
			sameKilledEvents = nil
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
					newMultipliers[topPlayerID] = map[string][]*boiler.Multiplier{}
				}
				m, ok := ms.getMultiplier("most_sups_lost", "", 0)
				if !ok {
					gamelog.L.Error().Str("most_sups_lost", topPlayerID).Err(err).Msg("unable to retrieve 'a fool and his money' from multipliers. maybe this code needs to be removed?")
					continue
				}
				if _, ok := newMultipliers[topPlayerID]; !ok {
					newMultipliers[topPlayerID] = make(map[string][]*boiler.Multiplier)
				}
				newMultipliers[topPlayerID][m.ID] = append(newMultipliers[topPlayerID][m.ID], m)
			}
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

	m3, _ := ms.getMultiplier("syndicate_win", "", 3)

	ms.battle.users.Range(func(bu *BattleUser) bool {
		if bu.FactionID == lastWin.FactionID {
			if _, ok := newMultipliers[bu.ID.String()]; !ok {
				// newMultipliers[bu.ID.String()] = map[*boiler.Multiplier]bool{}
				return true
			}
			if hatTrick {
				newMultipliers[bu.ID.String()][m3.ID] = append(newMultipliers[bu.ID.String()][m3.ID], m3)
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
			newMultipliers[wm.OwnedByID] = make(map[string][]*boiler.Multiplier)
		}
		newMultipliers[wm.OwnedByID][m1.ID] = append(newMultipliers[wm.OwnedByID][m1.ID], m1)

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
			newMultipliers[wm.OwnedByID] = make(map[string][]*boiler.Multiplier)
		}
		newMultipliers[wm.OwnedByID][m3.ID] = append(newMultipliers[wm.OwnedByID][m3.ID], m3)
	}

	// insert multipliers
	for pid, mlts := range newMultipliers {
		for _, mlt := range mlts {
			// if it is a citizen multi
			for _, m := range mlt {
				mlt := &boiler.UserMultiplier{
					PlayerID:          pid,
					FromBattleNumber:  ms.battle.BattleNumber,
					UntilBattleNumber: ms.battle.BattleNumber + 1,
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
}
