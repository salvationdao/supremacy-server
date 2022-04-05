package db

import (
	"context"
	"math/rand"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// GameMapCreate create a new game map
func GameMapCreate(ctx context.Context, conn Conn, gameMap *server.GameMap) error {
	q := `
		INSERT INTO 
			game_maps (
				name, 
				image_url, 
				width, 
				height, 
				cells_x, 
				cells_y, 
				top_pixels, 
				left_pixels, 
				scale, 
				disabled_cells,
				max_spawns
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING 
			id,
			name, 
			image_url, 
			width, 
			height, 
			cells_x, 
			cells_y, 
			top_pixels, 
			left_pixels, 
			scale,
			disabled_cells,
			max_spawns
		
	`
	err := pgxscan.Get(ctx, conn, gameMap, q,
		gameMap.Name,
		gameMap.ImageUrl,
		gameMap.Width,
		gameMap.Height,
		gameMap.CellsX,
		gameMap.CellsY,
		gameMap.TopPixels,
		gameMap.LeftPixels,
		gameMap.Scale,
		gameMap.DisabledCells,
		gameMap.MaxSpawns,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// GameMapGetRandom return a game map by given id
func GameMapGetRandom(ctx context.Context, conn Conn) (*boiler.GameMap, error) {
	maps, err := boiler.GameMaps().All(gamedb.StdConn)

	if err != nil {
		return nil, terror.Error(err)
	}

	rand.Seed(time.Now().UnixNano())
	gameMap := maps[rand.Intn(len(maps))]

	return gameMap, nil
}
