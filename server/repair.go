package server

import (
	"github.com/shopspring/decimal"
)

type RepairGameBlock struct {
	ID              string                   `json:"id"`
	Type            string                   `json:"type"`
	Key             string                   `json:"key"`
	SpeedMultiplier decimal.Decimal          `json:"speed_multiplier"`
	TotalScore      int                      `json:"total_score"`
	Dimension       RepairGameBlockDimension `json:"dimension"`
}

type RepairGameBlockDimension struct {
	Width decimal.Decimal `json:"width"`
	Depth decimal.Decimal `json:"depth"`
}
