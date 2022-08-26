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
		boiler.SalePlayerAbilityColumns.BlueprintID:
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

func PlayerAbilityAssign(playerID string, blueprintID string) error {
	q := `
		INSERT INTO 
		    player_abilities (owner_id, blueprint_id, count)
		VALUES 
		    ($1, $2, 1)
		ON CONFLICT 
		    (owner_id, blueprint_id)
		DO UPDATE SET
			count = player_abilities.count + 1;
	`

	_, err := gamedb.StdConn.Exec(q, playerID, blueprintID)
	if err != nil {
		return err
	}

	return nil
}

type AbilityLabel struct {
	Label               string `db:"label"`
	GameClientAbilityID int    `db:"game_client_ability_id"`
}

func AbilityLabelList() ([]*AbilityLabel, error) {
	q := fmt.Sprintf(
		`
			SELECT DISTINCT (ta.label), ta.game_client_ability_id
			FROM (
					SELECT DISTINCT(UPPER(%[2]s)) AS label, %[3]s
			      	FROM %[1]s
			      	UNION
			      	SELECT DISTINCT(UPPER(%[5]s)) AS label, %[6]s
			      	FROM %[4]s
			) ta
			ORDER BY ta.game_client_ability_id DESC;
		`,
		boiler.TableNames.BlueprintPlayerAbilities,               // 1
		boiler.BlueprintPlayerAbilityColumns.Label,               // 2
		boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID, // 3
		boiler.TableNames.GameAbilities,                          // 4
		boiler.GameAbilityColumns.Label,                          // 5
		boiler.GameAbilityColumns.GameClientAbilityID,            // 6
	)

	rows, err := gamedb.StdConn.Query(q)
	if err != nil {
		return nil, terror.Error(err, "Failed to query abilities")
	}

	resp := []*AbilityLabel{}
	for rows.Next() {
		al := &AbilityLabel{}
		err = rows.Scan(&al.Label, &al.GameClientAbilityID)
		if err != nil {
			return nil, terror.Error(err, "Failed to scan ability label")
		}

		resp = append(resp, al)
	}

	return resp, nil
}
