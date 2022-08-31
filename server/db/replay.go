package db

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"strconv"
)

func ReplayList(
	Search string,
	Sort *ListSortRequest,
	Limit int,
	Offset int,
) (int64, []*server.BattleReplay, error) {
	var queryMods []qm.QueryMod

	queryMods = append(queryMods,
		qm.Select(
			boiler.BattleReplayColumns.ID,
			boiler.BattleReplayColumns.BattleID,
			boiler.BattleReplayColumns.IsCompleteBattle,
			boiler.BattleReplayColumns.StreamID,
			boiler.BattleReplayColumns.ArenaID,
			boiler.BattleReplayColumns.RecordingStatus,
			boiler.BattleReplayColumns.CreatedAt,
			boiler.BattleReplayColumns.StoppedAt,
			boiler.BattleReplayColumns.StartedAt,
		),
		boiler.BattleReplayWhere.IsCompleteBattle.EQ(true),
		boiler.BattleReplayWhere.StreamID.IsNotNull(),
	)

	if Search != "" {
		// check if battle number is a valid number
		number, err := strconv.Atoi(Search)
		if err != nil {
			return 0, nil, terror.Error(err, "Failed to get battle number from search. Please ensure you use a real number")
		}

		queryMods = append(queryMods,
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s AND %s = ?",
					boiler.TableNames.Battles,
					qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.ID),
					qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.BattleID),
					qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.BattleNumber),
				),
				number,
			),
		)
	}

	count, err := boiler.BattleReplays(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err, "Failed to get count")
	}

	if Sort != nil && Sort.IsValid() {
		queryMods = append(queryMods, Sort.GenQueryMod())
	}

	queryMods = append(queryMods, qm.Limit(Limit))
	queryMods = append(queryMods, qm.Offset(Offset))
	queryMods = append(queryMods,
		qm.Load(boiler.BattleReplayRels.Battle),
		qm.Load(qm.Rels(boiler.BattleReplayRels.Battle, boiler.BattleRels.GameMap)),
	)

	brs, err := boiler.BattleReplays(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err, "Failed to get battle replays")
	}

	return count, server.BattleReplaySliceFromBoilerNoEvent(brs), nil
}
