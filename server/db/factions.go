package db

// import (
// 	"server"

// 	"github.com/georgysavva/scany/pgxscan"
// 	"github.com/ninja-software/terror/v2"
// 	"golang.org/x/net/context"
// )

// // FactionCreate create a new faction
// func FactionCreate(ctx context.Context, conn Conn, faction *server.Faction) error {
// 	q := `
// 		INSERT INTO
// 			factions (id, label, colour)
// 		VALUES
// 			($1, $2, $3)
// 		RETURNING
// 			id, label
// 	`

// 	err := pgxscan.Get(ctx, conn, faction, q, faction.ID, faction.Label, faction.Colour)
// 	if err != nil {
// 		return terror.Error(err)
// 	}

// 	return nil
// }

// // FactionGet return a faction by given id
// func FactionGet(ctx context.Context, conn Conn, factionID server.FactionID) (*server.Faction, error) {
// 	result := &server.Faction{}

// 	q := `
// 		SELECT * FROM factions where id = $1
// 	`

// 	err := pgxscan.Get(ctx, conn, result, q, factionID)
// 	if err != nil {
// 		return nil, terror.Error(err)
// 	}

// 	return result, nil
// }

// func FactionAll(ctx context.Context, conn Conn) ([]*server.Faction, error) {
// 	result := []*server.Faction{}

// 	q := `
// 		SELECT * FROM factions
// 	`

// 	err := pgxscan.Select(ctx, conn, &result, q)
// 	if err != nil {
// 		return nil, terror.Error(err)
// 	}

// 	return result, nil
// }
