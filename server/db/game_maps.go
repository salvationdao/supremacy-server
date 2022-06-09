package db

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
				disabled_cells,
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
			disabled_cells,
		
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
		gameMap.DisabledCells,
	)
	if err != nil {
		return err
	}

	return nil
}

// GameMapGetRandom return a game map by given id
func GameMapGetRandom(allowLastMap bool) (*boiler.GameMap, error) {

	mapQueries := []qm.QueryMod{
		qm.Select(
			boiler.GameMapColumns.ID,
			boiler.GameMapColumns.Name,
		),
		boiler.GameMapWhere.DisabledAt.IsNull(),
	}

	mapCount, err := boiler.GameMaps(mapQueries...).Count(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	if mapCount == 1 {
		gameMap, err := boiler.GameMaps(mapQueries...).All(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		return gameMap[0], nil
	}

	if !allowLastMap {
		lastBattle, err := boiler.Battles(
			qm.Select(
				boiler.BattleColumns.ID,
				boiler.BattleColumns.GameMapID,
			),
			boiler.BattleWhere.EndedAt.IsNotNull(),
			qm.OrderBy("ended_at desc"),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		if lastBattle != nil {
			mapQueries = append(mapQueries, boiler.GameMapWhere.ID.NEQ(lastBattle.GameMapID))
		}
	}

	maps, err := boiler.GameMaps(mapQueries...).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	gameMap := maps[rand.Intn(len(maps))]

	return gameMap, nil
}
