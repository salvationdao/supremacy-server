package db

import (
	"fmt"
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// FactionAbilityCreate create a new faction action
func FactionAbilityCreate(ctx context.Context, conn Conn, factionAbility *server.FactionAbility) error {
	q := `
		INSERT INTO
			faction_abilities (faction_id, label, type, colour, supremacy_token_cost, image_url, cooldown_duration_second)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id, faction_id, label, type, colour, supremacy_token_cost, image_url, cooldown_duration_second
	`

	err := pgxscan.Get(ctx, conn, factionAbility, q,
		factionAbility.FactionID,
		factionAbility.Label,
		factionAbility.Type,
		factionAbility.Colour,
		factionAbility.SupremacyTokenCost,
		factionAbility.ImageUrl,
		factionAbility.CooldownDurationSecond,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionAbilityGetRandom return three random abilities
func FactionAbilityGetRandom(ctx context.Context, conn Conn, factionID server.FactionID) ([]*server.FactionAbility, error) {
	result := []*server.FactionAbility{}
	q := `
		SELECT * FROM faction_abilities
		WHERE faction_id = $1
		ORDER BY RANDOM()
		LIMIT 3;
	`
	err := pgxscan.Select(ctx, conn, &result, q, factionID)
	if err != nil {
		fmt.Println(err)
		return nil, terror.Error(err)
	}

	return result, nil
}
