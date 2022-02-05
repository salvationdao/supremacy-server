package db

import (
	"fmt"
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// AbilityCollectionCreate create ability collection
func AbilityCollectionCreate(ctx context.Context, conn Conn, abilityCollection *server.AbilityCollection) error {
	q := `
		INSERT INTO
			ability_collections (label, colour, image_url, cooldown_duration_second)
		VALUES
			($1, $2, $3, $4)
		RETURNING
			id, label, colour, image_url, cooldown_duration_second
	`
	err := pgxscan.Get(ctx, conn, abilityCollection, q,
		abilityCollection.Label,
		abilityCollection.Colour,
		abilityCollection.ImageUrl,
		abilityCollection.CooldownDurationSecond,
	)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

// FactionAbilityCreate create a new faction action
func FactionAbilityCreate(ctx context.Context, conn Conn, factionAbility *server.FactionAbility) error {
	q := `
		INSERT INTO
			faction_abilities (game_client_ability_id, faction_id, label, usd_cent_cost, collection_id)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING
			id, game_client_ability_id, faction_id, label, usd_cent_cost, collection_id
	`

	err := pgxscan.Get(ctx, conn, factionAbility, q,
		factionAbility.GameClientAbilityID,
		factionAbility.FactionID,
		factionAbility.Label,
		factionAbility.USDCentCost,
		factionAbility.CollectionID,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// AbilityCollectionGetRandom return three random abilities
func AbilityCollectionGetRandom(ctx context.Context, conn Conn) (*server.AbilityCollection, error) {
	result := &server.AbilityCollection{}
	q := `
		SELECT * FROM ability_collections
		ORDER BY RANDOM()
		LIMIT 1;
	`
	err := pgxscan.Get(ctx, conn, result, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// FactionAbilityGetByCollectionID return faction ability by given collection id
func FactionAbilityGetByCollectionID(ctx context.Context, conn Conn, collectionID server.AbilityCollectionID) ([]*server.FactionAbility, error) {
	result := []*server.FactionAbility{}
	q := `
		SELECT * FROM faction_abilities
		where collection_id = $1;
	`
	err := pgxscan.Select(ctx, conn, &result, q, collectionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}
