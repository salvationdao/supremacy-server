package db

import (
	"database/sql"
	"errors"
	"math/rand"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

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

// GameMapGetRandom return a game map by given id
func GameMapGetByID(id string) (*boiler.GameMap, error) {

	mapQueries := []qm.QueryMod{
		boiler.GameMapWhere.ID.EQ(id),
		qm.Select(
			boiler.GameMapColumns.ID,
			boiler.GameMapColumns.Name,
		),
		boiler.GameMapWhere.DisabledAt.IsNull(),
	}

	gameMap, err := boiler.GameMaps(mapQueries...).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return gameMap, nil
}
