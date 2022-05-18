package multipliers

import (
	"database/sql"
	"errors"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/shopspring/decimal"
)

type MultiplierSummary struct {
	PlayerID        string          `json:"player_id"`
	TotalMultiplier decimal.Decimal `json:"total_multiplier"`
}

type PlayerMultiplier struct {
	Key              string          `json:"key"`
	Description      string          `json:"description"`
	Value            decimal.Decimal `json:"value"`
	IsMultiplicative bool            `json:"is_multiplicative"`
	BattleNumber     int             `json:"battle_number"`
}

// GetPlayersMultiplierSummaryForBattle gets the summary for multipliers for all user multis in a battle
func GetPlayersMultiplierSummaryForBattle(battleNumber int) ([]*MultiplierSummary, error) {
	var result []*MultiplierSummary

	userMultipliers, err := boiler.PlayerMultipliers(
		boiler.PlayerMultiplierWhere.FromBattleNumber.LTE(battleNumber),
		boiler.PlayerMultiplierWhere.UntilBattleNumber.GT(battleNumber),
		qm.Load(boiler.PlayerMultiplierRels.Multiplier),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().
			Int("battleNumber", battleNumber).
			Err(err).
			Msg("unable to retrieve player multipliers")
		return nil, err
	}

	playerMulties := make(map[string][]*boiler.PlayerMultiplier)

	for _, multi := range userMultipliers {
		playerMulties[multi.PlayerID] = append(playerMulties[multi.PlayerID], multi)
	}

	for playerID, multis := range playerMulties {
		_, total := calculateSingleUserMultiplierValues(multis)
		result = append(result, &MultiplierSummary{
			PlayerID:        playerID,
			TotalMultiplier: total,
		})
	}

	return result, nil
}

// GetPlayerMultipliersForBattle gets the player multipliers for a player for a given battle
func GetPlayerMultipliersForBattle(playerID string, battleNumber int) ([]*PlayerMultiplier, decimal.Decimal, bool) {
	userMultipliers, err := boiler.PlayerMultipliers(
		boiler.PlayerMultiplierWhere.PlayerID.EQ(playerID),
		boiler.PlayerMultiplierWhere.FromBattleNumber.LTE(battleNumber),
		boiler.PlayerMultiplierWhere.UntilBattleNumber.GT(battleNumber),
		qm.Load(boiler.PlayerMultiplierRels.Multiplier),
	).All(gamedb.StdConn)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().
				Str("playerID.String()", playerID).
				Int("battleNumber", battleNumber).
				Err(err).
				Msg("unable to retrieve player multipliers")
			return []*PlayerMultiplier{}, decimal.Zero, false
		}
		return nil, decimal.Zero, false
	}

	playerMulties, total := calculateSingleUserMultiplierValues(userMultipliers)

	return playerMulties, total, total.GreaterThan(decimal.Zero)
}

// calculateUserMultiplierValues takes a list of user multipliers, coverts them to player multipliers (more user-friendly struct) and returns the total multiplier
func calculateSingleUserMultiplierValues(userMultipliers []*boiler.PlayerMultiplier) ([]*PlayerMultiplier, decimal.Decimal) {
	multipliers := make([]*PlayerMultiplier, len(userMultipliers))
	value := decimal.Zero
	multiplicativeValue := decimal.Zero
	for i, m := range userMultipliers {
		multipliers[i] = &PlayerMultiplier{
			Key:              m.R.Multiplier.Key,
			Description:      m.R.Multiplier.Description,
			IsMultiplicative: m.R.Multiplier.IsMultiplicative,
			BattleNumber:     m.FromBattleNumber,
		}

		if !m.R.Multiplier.IsMultiplicative {
			multipliers[i].Value = m.Value.Shift(-1)
			value = value.Add(m.Value)
			continue
		}

		multipliers[i].Value = m.Value
		multiplicativeValue = multiplicativeValue.Add(m.Value)
	}

	// set multiplicative to 1 if the value is zero
	if multiplicativeValue.Equal(decimal.Zero) {
		multiplicativeValue = decimal.NewFromInt(1)
	}

	return multipliers, value.Mul(multiplicativeValue)
}

// CalculateOneMultiWorth takes a slice of MultiplierSummary and the total spoils and returns the value of a 1x multiplier
func CalculateOneMultiWorth(multis []*MultiplierSummary, spoilsTotal decimal.Decimal) decimal.Decimal {
	totalMultis := decimal.Zero

	for _, player := range multis {
		totalMultis = totalMultis.Add(player.TotalMultiplier)
	}

	if totalMultis.IsZero() {
		return decimal.Zero
	}

	return spoilsTotal.Div(totalMultis)
}

// CalculateMultipliersWorth takes the value of a 1x multiplier and a users total multiplier and returns the value of the users total multiplier
func CalculateMultipliersWorth(oneMultiWorth decimal.Decimal, totalMultiplier decimal.Decimal) decimal.Decimal {
	return oneMultiWorth.Mul(totalMultiplier)
}

// FriendlyFormatMultiplier returns a total multiplier in the format "65.5x", example FriendlyFormatMultiplier(5665) = 566.5x
func FriendlyFormatMultiplier(multi decimal.Decimal) string {
	return multi.Shift(-1).Round(1).String()
}
