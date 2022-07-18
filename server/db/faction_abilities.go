package db

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// BattleAbilityGetRandom return three random abilities
func BattleAbilityGetRandom(abilityLabel string) (*boiler.BattleAbility, error) {
	queries := []qm.QueryMod{}

	if abilityLabel != "" {
		queries = append(queries, boiler.BattleAbilityWhere.Label.EQ(abilityLabel))
	}

	battleAbilities, err := boiler.BattleAbilities(queries...).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("query", queries).Msg("Failed to query battle abilities")
		return nil, err
	}

	if battleAbilities == nil {
		gamelog.L.Error().Str("ability label", abilityLabel).Msg("No battle ability found")
		return nil, terror.Error(fmt.Errorf("no ability found"), "No ability found")
	}

	battleAbility := battleAbilities[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(battleAbilities))]

	return battleAbility, nil
}

// FactionAbilitiesSupsCostUpdate update faction exclusive ability
func FactionAbilitiesSupsCostUpdate(gameAbilityID string, supsCost decimal.Decimal, currentSups decimal.Decimal) error {
	supsCost = supsCost.RoundDown(0)
	currentSups = currentSups.RoundDown(0)
	asc := boiler.GameAbility{
		ID:          gameAbilityID,
		SupsCost:    supsCost.String(),
		CurrentSups: currentSups.String(),
	}

	_, err := asc.Update(gamedb.StdConn, boil.Whitelist(boiler.GameAbilityColumns.SupsCost, boiler.GameAbilityColumns.CurrentSups))
	if err != nil {
		return err
	}

	return nil
}
