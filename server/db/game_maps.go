package db

import (
	"context"
	"server"

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
				disabled_cells
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
			disabled_cells
		
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
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// GameMapGet return a game map by given id
func GameMapGet(ctx context.Context, conn Conn, id server.GameMapID) (*server.GameMap, error) {
	gameMap := &server.GameMap{}

	q := `
		SELECT * FROM game_maps WHERE id = $1 
	`
	err := pgxscan.Get(ctx, conn, gameMap, q, id)
	if err != nil {
		return nil, terror.Error(err)
	}

	return gameMap, nil
}

// GameMapGetRamdom return a game map by given id
func GameMapGetRandom(ctx context.Context, conn Conn) (*server.GameMap, error) {
	gameMap := &server.GameMap{}

	q := `
		SELECT * FROM game_maps
		ORDER BY RANDOM()
		LIMIT 1
	`

	err := pgxscan.Get(ctx, conn, gameMap, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return gameMap, nil
}
