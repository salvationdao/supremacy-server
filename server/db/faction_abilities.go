package db

import (
	"math/rand"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// BattleAbilityGetRandom return three random abilities
func BattleAbilityGetRandom() (*boiler.BattleAbility, error) {
	battleAbilities, err := boiler.BattleAbilities().All(gamedb.StdConn)
	if err != nil {
		return nil, err
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
