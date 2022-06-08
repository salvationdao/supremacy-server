package battle

import (
	"fmt"
	"math"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-syndicate/ws"

	"github.com/shopspring/decimal"
)

func CalcNextQueueStatus(length int64) QueueStatusResponse {
	ql := float64(length + 1)
	queueLength := decimal.NewFromFloat(ql)

	// min cost will be one forth of the queue length
	minQueueCost := queueLength.Div(decimal.NewFromFloat(4)).Mul(decimal.New(1, 18))

	mul := db.GetDecimalWithDefault("queue_fee_log_multi", decimal.NewFromFloat(3.25))
	mulFloat, _ := mul.Float64()

	// calc queue cost
	feeMultiplier := math.Log(float64(ql)) / mulFloat * 0.25
	queueCost := queueLength.Mul(decimal.NewFromFloat(feeMultiplier)).Mul(decimal.New(1, 18))

	// calc contract reward
	contractReward := queueLength.Mul(decimal.New(2, 18))

	// fee never get under queue length
	if queueCost.LessThan(minQueueCost) {
		queueCost = minQueueCost
	}

	// length * 2 sups
	return QueueStatusResponse{
		QueueLength: length, // return the current queue length

		// the fee player have to pay if they want to queue their mech
		QueueCost: queueCost,

		// the reward, player will get if their mech won the battle
		ContractReward: contractReward,
	}
}

func BroadcastQueuePositions() {
	// get each faction queue
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get factions to broadcast queue positions")
		return
	}

	for _, fac := range factions {
		go func() {
			factionQueue, err := db.FactionQueue(fac.ID)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("failed to get faction queue to broadcast queue positions")
				return
			}

			for _, fc := range factionQueue {
				ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", fac.ID, fc.MechID.String()), WSPlayerAssetMechQueueSubscribe, fc.QueuePosition)
			}
		}()
	}
}
