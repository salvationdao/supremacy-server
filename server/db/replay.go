package db

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
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
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.ID),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.BattleID),
			boiler.BattleReplayColumns.BattleID,
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.IsCompleteBattle),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.StreamID),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.ArenaID),
			fmt.Sprintf("%s as arena_id", qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.ArenaID)),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.RecordingStatus),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.CreatedAt),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.StoppedAt),
			qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.StartedAt),
		),
		boiler.BattleReplayWhere.IsCompleteBattle.EQ(true),
		boiler.BattleReplayWhere.StreamID.IsNotNull(),
	)

	if Search != "" {
		queryMods = append(queryMods,
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s",
					boiler.TableNames.Battles,
					qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.ID),
					qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.BattleID),
				),
			),
			qm.Where(
				fmt.Sprintf(
					"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s||%s::TEXT ILIKE ?)",
					boiler.TableNames.GameMaps,
					qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.GameMapID),
					qm.Rels(boiler.TableNames.GameMaps, boiler.GameMapColumns.ID),
					qm.Rels(boiler.TableNames.GameMaps, boiler.GameMapColumns.Name),
					qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.BattleNumber),
				),
				"%"+Search+"%",
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
		qm.Load(qm.Rels(boiler.BattleReplayRels.Battle, boiler.BattleRels.GameMap)),
	)
	queryMods = append(queryMods,
		qm.Load(boiler.BattleReplayRels.Arena),
	)

	brs, err := boiler.BattleReplays(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err, "Failed to get battle replays")
	}

	return count, server.BattleReplaySliceFromBoilerNoEvent(brs), nil
}
