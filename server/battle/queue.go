package battle

import (
	"math"
	"server/db"

	"github.com/shopspring/decimal"
)

func CalcNextQueueStatus(length int64) QueueStatusResponse {
	ql := float64(length + 1)
	queueAddage := db.GetIntWithDefault(db.QueueLengthAdd, 100)
	ql = ql + float64(queueAddage)

	queueLength := decimal.NewFromFloat(ql)

	// min cost will be one forth of the queue length or the floor

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
