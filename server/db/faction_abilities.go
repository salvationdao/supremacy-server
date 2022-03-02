package db

import (
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// BattleAbilityCreate create ability collection
func BattleAbilityCreate(ctx context.Context, conn Conn, battleAbility *server.BattleAbility) error {
	q := `
		INSERT INTO
			battle_abilities (label, description, cooldown_duration_second)
		VALUES
			($1, $2, $3)
		RETURNING
			id, label, description, cooldown_duration_second
	`
	err := pgxscan.Get(ctx, conn, battleAbility, q,
		battleAbility.Label,
		battleAbility.Description,
		battleAbility.CooldownDurationSecond,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// GameAbilityCreate create a new faction action
func GameAbilityCreate(ctx context.Context, conn Conn, gameAbility *server.GameAbility) error {
	q := `
		INSERT INTO
			game_abilities (game_client_ability_id, faction_id, label, sups_cost, battle_ability_id, colour, text_colour, description, image_url)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id, game_client_ability_id, faction_id, label, sups_cost, battle_ability_id, colour, text_colour, description, image_url
	`

	err := pgxscan.Get(ctx, conn, gameAbility, q,
		gameAbility.GameClientAbilityID,
		gameAbility.FactionID,
		gameAbility.Label,
		gameAbility.SupsCost,
		gameAbility.BattleAbilityID,
		gameAbility.Colour,
		gameAbility.TextColour,
		gameAbility.Description,
		gameAbility.ImageUrl,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// AbilityCollectionGetRandom return three random abilities
func AbilityCollectionGetRandom(ctx context.Context, conn Conn) (*server.BattleAbility, error) {
	result := &server.BattleAbility{}
	q := `
		SELECT * FROM battle_abilities
		ORDER BY RANDOM()
		LIMIT 1;
	`
	err := pgxscan.Get(ctx, conn, result, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// GameAbility return faction ability by given id
func GameAbility(ctx context.Context, conn Conn, id server.GameAbilityID) (*server.GameAbility, error) {
	result := &server.GameAbility{}
	q := `
		SELECT * FROM game_abilities
		where id = $1;
	`
	err := pgxscan.Get(ctx, conn, result, q, id)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

func GetZaibatsuFactionAbility(ctx context.Context, conn Conn) (*server.GameAbility, error) {
	result := &server.GameAbility{}
	q := `
		SELECT * FROM game_abilities
		WHERE game_client_ability_id = 11
	`

	err := pgxscan.Get(ctx, conn, result, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

func ZaibatsuFactionAbilityUpdate(ctx context.Context, conn Conn, supsCost string) error {
	q := `
	UPDATE 
		game_abilities
	SET
		sups_cost = $1
	WHERE game_client_ability_id = 11
	`

	_, err := conn.Exec(ctx, q, supsCost)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionAbilityGetByBattleAbilityID return game ability by given collection id
func FactionAbilityGetByBattleAbilityID(ctx context.Context, conn Conn, battleAbilityID server.BattleAbilityID) ([]*server.GameAbility, error) {
	result := []*server.GameAbility{}
	q := `
		SELECT * FROM game_abilities
		where battle_ability_id = $1;
	`
	err := pgxscan.Select(ctx, conn, &result, q, battleAbilityID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// FactionExclusiveAbilitiesByFactionID return exclusive abilities of a faction which is not battle abilities
func FactionExclusiveAbilitiesByFactionID(ctx context.Context, conn Conn, factionID server.FactionID) ([]*server.GameAbility, error) {
	result := []*server.GameAbility{}
	q := `
		SELECT * FROM game_abilities
		where faction_id = $1 AND battle_ability_id ISNULL;
	`
	err := pgxscan.Select(ctx, conn, &result, q, factionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// FactionExclusiveAbilitiesSupsCostUpdate update faction exclusive ability
func FactionExclusiveAbilitiesSupsCostUpdate(ctx context.Context, conn Conn, gameAbility *server.GameAbility) error {
	q := `
		UPDATE 
			game_abilities
		SET
			sups_cost = $2
		where 
			id = $1;
	`
	_, err := conn.Exec(ctx, q,
		gameAbility.ID,
		gameAbility.SupsCost,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
