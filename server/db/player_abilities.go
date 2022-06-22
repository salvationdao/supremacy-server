package db

import (
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
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
		boiler.SalePlayerAbilityColumns.CurrentPrice,
		boiler.SalePlayerAbilityColumns.AvailableUntil:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid sale player ability column"))
}

func (p PlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.PlayerAbilityColumns.ID,
		boiler.PlayerAbilityColumns.OwnerID,
		boiler.PlayerAbilityColumns.GameClientAbilityID,
		boiler.PlayerAbilityColumns.Label,
		boiler.PlayerAbilityColumns.Colour,
		boiler.PlayerAbilityColumns.ImageURL,
		boiler.PlayerAbilityColumns.Description,
		boiler.PlayerAbilityColumns.TextColour,
		boiler.PlayerAbilityColumns.LocationSelectType,
		boiler.PlayerAbilityColumns.PurchasedAt:
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

// CurrentSaleAbilitiesList returns a list of abilities that are currently on sale from the sale_player_abilities table.
func CurrentSaleAbilitiesList() ([]*SaleAbilityDetailed, error) {
	spas, err := boiler.SalePlayerAbilities(
		boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now())),
		qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	detailedSaleAbilities := []*SaleAbilityDetailed{}
	for _, s := range spas {
		detailedSaleAbilities = append(detailedSaleAbilities, &SaleAbilityDetailed{
			SalePlayerAbility: s,
			Ability:           s.R.Blueprint,
		})
	}

	return detailedSaleAbilities, nil
}

type TalliedPlayerAbility struct {
	BlueprintID     string                         `json:"blueprint_id" boil:"blueprint_id"`
	Count           int                            `json:"count" boil:"count"`
	LastPurchasedAt time.Time                      `json:"last_purchased_at" boil:"last_purchased_at"`
	Ability         *boiler.BlueprintPlayerAbility `json:"ability,omitempty"`
}

// TalliedPlayerAbilitiesList returns a list of tallied player abilities, ordered by last purchased date from the player_abilities table.
func TalliedPlayerAbilitiesList(
	userID string,
) ([]*TalliedPlayerAbility, error) {
	talliedPlayerAbilities := []*TalliedPlayerAbility{}
	err := boiler.PlayerAbilities(
		qm.Select(boiler.PlayerAbilityColumns.BlueprintID,
			fmt.Sprintf("count(%s)", boiler.PlayerAbilityColumns.BlueprintID),
			fmt.Sprintf("max(%s) as last_purchased_at", boiler.PlayerAbilityColumns.PurchasedAt)),
		qm.GroupBy(boiler.PlayerAbilityColumns.BlueprintID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(userID),
		qm.OrderBy("last_purchased_at desc"),
	).Bind(nil, gamedb.StdConn, &talliedPlayerAbilities)
	if err != nil {
		return nil, err
	}

	abilityIDs := []string{}
	for _, p := range talliedPlayerAbilities {
		abilityIDs = append(abilityIDs, p.BlueprintID)
	}

	bpas, err := boiler.BlueprintPlayerAbilities(
		boiler.BlueprintPlayerAbilityWhere.ID.IN(abilityIDs),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, t := range talliedPlayerAbilities {
		for _, b := range bpas {
			if b.ID == t.BlueprintID {
				t.Ability = b
				break
			}
		}
	}

	return talliedPlayerAbilities, nil
}
