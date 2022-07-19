package db

import (
	"fmt"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type (
	SalePlayerAbilityColumn      string
	PlayerAbilityColumn          string
	BlueprintPlayerAbilityColumn string
)

func (p SalePlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.SalePlayerAbilityColumns.ID,
		boiler.SalePlayerAbilityColumns.BlueprintID,
		boiler.SalePlayerAbilityColumns.CurrentPrice:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid sale player ability column"))
}

func (p PlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.PlayerAbilityColumns.ID,
		boiler.PlayerAbilityColumns.OwnerID:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid player ability column"))
}

func (p BlueprintPlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.BlueprintPlayerAbilityColumns.ID,
		boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID,
		boiler.BlueprintPlayerAbilityColumns.Label,
		boiler.BlueprintPlayerAbilityColumns.Colour,
		boiler.BlueprintPlayerAbilityColumns.ImageURL,
		boiler.BlueprintPlayerAbilityColumns.Description,
		boiler.BlueprintPlayerAbilityColumns.TextColour,
		boiler.BlueprintPlayerAbilityColumns.LocationSelectType:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid blueprint player ability column"))
}

type SaleAbilityDetailed struct {
	*boiler.SalePlayerAbility
	Ability *boiler.BlueprintPlayerAbility `json:"ability,omitempty"`
}

type DetailedPlayerAbility struct {
	*boiler.PlayerAbility
	Ability boiler.BlueprintPlayerAbility `json:"ability"`
}

// PlayerAbilitiesList returns a list of tallied player abilities, ordered by last purchased date from the player_abilities table.
// It excludes player abilities with a count of 0
func PlayerAbilitiesList(
	userID string,
) ([]*DetailedPlayerAbility, error) {
	pas, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.OwnerID.EQ(userID),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.PlayerAbilityColumns.Count)),
		qm.Load(boiler.PlayerAbilityRels.Blueprint),
		boiler.PlayerAbilityWhere.Count.GT(0),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	result := []*DetailedPlayerAbility{}
	for _, p := range pas {
		result = append(result, &DetailedPlayerAbility{
			PlayerAbility: p,
			Ability:       *p.R.Blueprint,
		})
	}

	return result, nil
}
