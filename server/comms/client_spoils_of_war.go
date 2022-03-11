package comms

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	FromUserID           uuid.UUID
	ToUserID             uuid.UUID
	Amount               decimal.Decimal
	TransactionReference string
	Group                string
	SubGroup             string
	Description          string

	NotSafe bool `json:"not_safe"`
}

// type FlushSupsReq struct {
// 	Transactions []*Transaction
// }
// type FlushSupsResp struct {
// }

// // Tables to deal with
// // battle_contributions
// // battle_ability_triggers
// // multipliers
// // user_multipliers
// // game_abilities

// func ProcessSpoilsOfWar(c *C, battleID uuid.UUID) error {
// 	// resp := &FlushSupsReq{}
// 	// err := c.Call("S.FlushSups", &FlushSupsReq{}, resp)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// contributions, err := boiler.BattleContributions(
// 	// 	boiler.BattleContributionWhere.BattleID.EQ(battleID.String()),
// 	// ).All(gamedb.StdConn)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// for _, contribution := range contributions {

// 	// }
// 	return nil
// }
