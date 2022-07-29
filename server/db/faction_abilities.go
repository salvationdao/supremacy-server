package db

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"server/db/boiler"
	"server/gamedb"
	"time"
)

// BattleAbilityGetRandom return three random abilities
func BattleAbilityGetRandom(excludedAbility string) (*boiler.BattleAbility, error) {
	queries := []qm.QueryMod{}
	if excludedAbility != "" {
		queries = append(queries, boiler.BattleAbilityWhere.Label.NEQ(excludedAbility))
	}

	battleAbilities, err := boiler.BattleAbilities(queries...).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	battleAbility := battleAbilities[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(battleAbilities))]

	return battleAbility, nil
}
