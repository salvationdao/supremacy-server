package battle

import (
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

func (ms *MultiplierSystem) calculate() {

}
