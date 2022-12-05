package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func GetBattleLobbyViaIDs(lobbyIDs []string) ([]*boiler.BattleLobby, error) {
	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.IN(lobbyIDs),
		qm.Load(boiler.BattleLobbyRels.GameMap),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.BattleLobbiesMechs),
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
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards,
			boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
			boiler.BattleLobbyExtraSupsRewardWhere.DeletedAt.IsNull(),
		),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return bls, nil
}

func GetBattleLobbyViaID(lobbyID string) (*boiler.BattleLobby, error) {
	// get next lobby
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(lobbyID),
		qm.Load(boiler.BattleLobbyRels.GameMap),
		qm.Load(boiler.BattleLobbyRels.BattleLobbiesMechs),
		qm.Load(boiler.BattleLobbyRels.HostBy),
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
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards,
			boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
			boiler.BattleLobbyExtraSupsRewardWhere.DeletedAt.IsNull(),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to load battle lobby from db")
		return nil, err
	}

	if bl == nil {
		return nil, terror.Error(fmt.Errorf("battle lobby not found"), "Battle lobby does not exist.")
	}

	return bl, nil
}

func GetBattleLobbyViaAccessCode(accessCode string) (*boiler.BattleLobby, error) {
	// get next lobby
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.AccessCode.EQ(null.StringFrom(accessCode)),
		qm.Load(boiler.BattleLobbyRels.GameMap),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.BattleLobbiesMechs),
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
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards,
			boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
			boiler.BattleLobbyExtraSupsRewardWhere.DeletedAt.IsNull(),
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
	if err != nil {
		return nil, err
	}
	// build excluding player query
	excludingPlayerQuery := ""
	if len(excludingPlayerIDs) > 0 {
		excludingPlayerQuery += fmt.Sprintf("AND %s NOT IN(", boiler.BattleLobbiesMechTableColumns.QueuedByID)
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

// PlayersInLobbies takes a list of battle lobby ids, and return a list of users in them battle lobbies (excluding AI player)
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
				boiler.BattleLobbiesMechColumns.QueuedByID,
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
				"(SELECT %s, %s FROM %s WHERE %s ISNULL AND %s ISNULL AND EXISTS(SELECT 1 FROM %s WHERE %s = %s AND %s = FALSE)) _blm ON _blm.%s = _bl.%s",
				boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
				boiler.BattleLobbiesMechTableColumns.QueuedByID,
				boiler.TableNames.BattleLobbiesMechs,
				boiler.BattleLobbiesMechTableColumns.RefundTXID,
				boiler.BattleLobbiesMechTableColumns.DeletedAt,
				boiler.TableNames.Players,
				boiler.PlayerTableColumns.ID,
				boiler.BattleLobbiesMechTableColumns.QueuedByID,
				boiler.PlayerTableColumns.IsAi,
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

type MechInLobby struct {
	ArenaID           string
	MechLabel         string
	MechID            string
	MechName          string
	QueuedByID        string
	StakedMechOwnerID null.String
}

func GetMechsInLobby(lobbyID string) ([]*MechInLobby, error) {
	queries := []qm.QueryMod{
		qm.Select(
			boiler.BattleLobbyTableColumns.AssignedToArenaID,
			boiler.BlueprintMechTableColumns.Label,
			boiler.MechTableColumns.ID,
			boiler.MechTableColumns.Name,
			boiler.BattleLobbiesMechTableColumns.QueuedByID,
			fmt.Sprintf(
				"(SELECT %s FROM %s WHERE %s = %s) AS stake_mech_owner_id",
				boiler.StakedMechTableColumns.OwnerID,
				boiler.TableNames.StakedMechs,
				boiler.StakedMechTableColumns.MechID,
				boiler.MechTableColumns.ID,
			),
		),
		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s = '%s') %s",
			boiler.TableNames.BattleLobbies,
			boiler.BattleLobbyTableColumns.ID,
			lobbyID,
			boiler.TableNames.BattleLobbies,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.BattleLobbyTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Mechs,
			boiler.MechTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.MechID,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			boiler.BlueprintMechTableColumns.ID,
			boiler.MechTableColumns.BlueprintID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to load battle lobby mechs")
	}

	data := []*MechInLobby{}
	for rows.Next() {
		mib := &MechInLobby{}

		err = rows.Scan(&mib.ArenaID, &mib.MechLabel, &mib.MechID, &mib.MechName, &mib.QueuedByID, &mib.StakedMechOwnerID)
		if err != nil {
			return nil, terror.Error(err, "Failed to scan battle lobby mech")
		}

		data = append(data, mib)
	}

	return data, nil
}
