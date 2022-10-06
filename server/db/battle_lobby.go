package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func GetBattleLobby(lobbyID string) (*boiler.BattleLobby, error) {
	// get next lobby
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(lobbyID),
		qm.Load(
			qm.Rels(boiler.BattleLobbyRels.GameMap),
			),
		qm.Load(
			qm.Rels(
				boiler.BattleLobbyRels.BattleLobbySupporters,
				boiler.BattleLobbySupporterRels.Supporter,
				boiler.PlayerRels.ProfileAvatar,
			),
		),
		qm.Load(
			qm.Rels(
				boiler.BattleLobbyRels.BattleLobbySupporterOptIns,
				boiler.BattleLobbySupporterOptInRels.Supporter,
				boiler.PlayerRels.ProfileAvatar,
			),
		),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return bl, nil
}

// GetNextBattleLobby finds the next upcoming battle
func GetNextBattleLobby(battleLobbyIDs []string) (*boiler.BattleLobby, error) {
	excludingPlayerIDs, err := playersInLobbies(battleLobbyIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	// build excluding player query
	excludingPlayerQuery := ""
	if len(excludingPlayerIDs) > 0 {
		excludingPlayerQuery += fmt.Sprintf("AND %s NOT IN(", boiler.BattleLobbiesMechTableColumns.OwnerID)
		for i, id := range excludingPlayerIDs {
			excludingPlayerQuery += "'" + id + "'"

			if i < len(excludingPlayerIDs)-1 {
				excludingPlayerQuery += ","
				continue
			}

			excludingPlayerQuery += ")"
		}
	}

	// get next lobby
	bl, err := boiler.BattleLobbies(
		qm.Where(fmt.Sprintf(
			"(SELECT COUNT(%s) FROM %s WHERE %s = %s AND %s NOTNULL AND %s ISNULL AND %s ISNULL AND %s ISNULL %s) = 9",
			boiler.BattleLobbiesMechTableColumns.ID,
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.BattleLobbyTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.LockedAt,
			boiler.BattleLobbiesMechTableColumns.EndedAt,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
			excludingPlayerQuery,
		)),
		boiler.BattleLobbyWhere.ReadyAt.IsNotNull(),
		qm.Where(fmt.Sprintf(
			"(%[1]s ISNULL OR %[1]s <= NOW())",
			boiler.BattleLobbyTableColumns.WillNotStartUntil,
		)),
		qm.OrderBy(fmt.Sprintf(
			"%s NULLS LAST, %s",
			boiler.BattleLobbyTableColumns.WillNotStartUntil,
			boiler.BattleLobbyTableColumns.ReadyAt,
		)),
		qm.Load(qm.Rels(boiler.BattleLobbyRels.BattleLobbySupporters, boiler.BattleLobbySupporterRels.Supporter, boiler.PlayerRels.ProfileAvatar)),
		qm.Load(qm.Rels(boiler.BattleLobbyRels.BattleLobbySupporterOptIns, boiler.BattleLobbySupporterOptInRels.Supporter, boiler.PlayerRels.ProfileAvatar)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return bl, nil
}

// PlayersInLobbies takes a list of battle lobby ids, and return a list of users in them battle lobbies
func playersInLobbies(battleLobbyIDs []string) ([]string, error) {
	players := []string{}
	if len(battleLobbyIDs) > 0 {
		battleLobbyQuery := ""
		if battleLobbyIDs != nil && len(battleLobbyIDs) > 0 {
			battleLobbyQuery += fmt.Sprintf("AND %s IN(", boiler.BattleLobbyColumns.ID)
			for i, id := range battleLobbyIDs {
				battleLobbyQuery += "'" + id + "'"

				if i < len(battleLobbyIDs)-1 {
					battleLobbyQuery += ","
					continue
				}

				battleLobbyQuery += ")"
			}
		}

		rows, err := boiler.NewQuery(
			qm.Select(fmt.Sprintf(
				"DISTINCT(_blm.%s)",
				boiler.BattleLobbiesMechColumns.OwnerID,
			)),
			qm.From(fmt.Sprintf(
				"(SELECT %s FROM %s WHERE %s NOTNULL AND %s ISNULL AND %s ISNULL %s) _bl",
				boiler.BattleLobbyColumns.ID,
				boiler.TableNames.BattleLobbies,
				boiler.BattleLobbyColumns.ReadyAt,
				boiler.BattleLobbyColumns.EndedAt,
				boiler.BattleLobbyColumns.DeletedAt,
				battleLobbyQuery,
			)),
			qm.InnerJoin(fmt.Sprintf(
				"(SELECT %s, %s FROM %s WHERE %s ISNULL AND %s ISNULL) _blm ON _blm.%s = _bl.%s",
				boiler.BattleLobbiesMechColumns.BattleLobbyID,
				boiler.BattleLobbiesMechColumns.OwnerID,
				boiler.TableNames.BattleLobbiesMechs,
				boiler.BattleLobbiesMechColumns.RefundTXID,
				boiler.BattleLobbiesMechColumns.DeletedAt,
				boiler.BattleLobbiesMechColumns.BattleLobbyID,
				boiler.BattleLobbyColumns.ID,
			)),
		).Query(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("Failed to load battle lobby")
			return []string{}, err
		}

		for rows.Next() {
			playerID := ""
			err = rows.Scan(&playerID)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to scan existing player id")
				return players, err
			}

			players = append(players, playerID)
		}
	}

	return players, nil
}
