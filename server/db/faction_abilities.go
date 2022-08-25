package db

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"server/db/boiler"
	"server/gamedb"
	"time"
)

// BattleAbilityGetRandom return three random abilities
func BattleAbilityGetRandom(firstAbilityLabel string, includeDeadlyAbilities bool) (*boiler.BattleAbility, error) {
	queries := []qm.QueryMod{}
	if firstAbilityLabel != "" {
		// only get first ability, if flag is set
		queries = append(queries, boiler.BattleAbilityWhere.Label.EQ(firstAbilityLabel))
	} else if !includeDeadlyAbilities {
		// exclude deadly abilities, if not set
		queries = append(queries, boiler.BattleAbilityWhere.KillingPowerLevel.NEQ(boiler.AbilityKillingPowerLevelDEADLY))
	}

	battleAbilities, err := boiler.BattleAbilities(queries...).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	battleAbility := battleAbilities[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(battleAbilities))]

	return battleAbility, nil
}
