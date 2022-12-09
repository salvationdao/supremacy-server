package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
func GetNextBattleLobby(battleLobbyIDs []string) (*boiler.BattleLobby, bool, error) {
	excludingPlayerIDs, err := playersInLobbies(battleLobbyIDs)
	if err != nil {
		return nil, false, err
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
		return nil, false, err
	}

	shouldFillAIMechs := false
	// get the system lobby which has the most queued mechs
	if bl == nil {
		excludingPlayerQuery = ""
		if len(excludingPlayerIDs) > 0 {
			excludingPlayerQuery += fmt.Sprintf(
				"AND NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s IN ( ",
				boiler.TableNames.BattleLobbiesMechs,
				boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
				boiler.BattleLobbyTableColumns.ID,
				boiler.BattleLobbiesMechTableColumns.QueuedByID,
			)
			for i, id := range excludingPlayerIDs {
				excludingPlayerQuery += "'" + id + "'"

				if i < len(excludingPlayerIDs)-1 {
					excludingPlayerQuery += ","
					continue
				}

				excludingPlayerQuery += ")"
			}
			excludingPlayerQuery += ")"
		}

		queries := []qm.QueryMod{
			qm.Select(boiler.BattleLobbyTableColumns.ID),
			qm.From(fmt.Sprintf(
				`(
					SELECT * FROM %s 
					WHERE 
						%s = TRUE AND 
						%s ISNULL AND 
						%s ISNULL 
						%s
				) %s`,
				boiler.TableNames.BattleLobbies,
				boiler.BattleLobbyTableColumns.GeneratedBySystem,
				boiler.BattleLobbyTableColumns.ReadyAt,
				boiler.BattleLobbyTableColumns.DeletedAt,
				excludingPlayerQuery,
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

			qm.GroupBy(boiler.BattleLobbyTableColumns.ID + "," + boiler.BattleLobbyTableColumns.CreatedAt),
			qm.OrderBy(fmt.Sprintf("COUNT(%s) DESC, %s", boiler.BattleLobbiesMechTableColumns.ID, boiler.BattleLobbyTableColumns.CreatedAt)),
			qm.Limit(1),
		}
		battleLobbyID := ""
		err = boiler.NewQuery(queries...).QueryRow(gamedb.StdConn).Scan(&battleLobbyID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("Failed to load battle lobby.")
			return nil, false, terror.Error(err, "Failed to load battle lobby")
		}

		if battleLobbyID != "" {
			bl, err = boiler.BattleLobbies(
				boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
				qm.Load(qm.Rels(boiler.BattleLobbyRels.BattleLobbySupporters, boiler.BattleLobbySupporterRels.Supporter, boiler.PlayerRels.ProfileAvatar)),
				qm.Load(qm.Rels(boiler.BattleLobbyRels.BattleLobbySupporterOptIns, boiler.BattleLobbySupporterOptInRels.Supporter, boiler.PlayerRels.ProfileAvatar)),
			).One(gamedb.StdConn)
			if err != nil {
				return nil, false, terror.Error(err, "Failed to load battle lobby.")
			}

			shouldFillAIMechs = true
		}
	}

	return bl, shouldFillAIMechs, nil
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

type TotalAmountExtraSups struct {
	Total decimal.Decimal `json:"total"`
}

var DISCORD_BATTLE_LOBBY_WAITING = "Waiting"
var DISCORD_BATTLE_LOBBY_QUEUE = "In Queue"
var DISCORD_BATTLE_LOBBY_BATTLE = "In Battle"
var DISCORD_BATTLE_LOBBY_END = "Battle Ended"

func GetDiscordEmbedMessage(battleLobbyID string) (*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	battleLobby, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, nil, err
	}

	if battleLobby.R == nil || battleLobby.R.HostBy == nil {
		return nil, nil, errors.New("failed to load battle lobby rels")
	}

	battleArenaBaseUrl := GetStrWithDefault(KeyBattleArenaWebURL, "https://play.supremacy.game")

	battleLobbyMechCount, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.BattleLobbyID.EQ(battleLobbyID),
		boiler.BattleLobbiesMechWhere.DeletedAt.IsNotNull(),
	).Count(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	totalSups := battleLobby.EntryFee.Mul(decimal.NewFromInt(battleLobbyMechCount))

	extraSups := &TotalAmountExtraSups{}

	err = boiler.BattleLobbyExtraSupsRewards(
		qm.Select(fmt.Sprintf("SUM (%s) as total", boiler.BattleLobbyExtraSupsRewardColumns.Amount)),
		boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
		boiler.BattleLobbyExtraSupsRewardWhere.BattleLobbyID.EQ(battleLobbyID),
	).Bind(context.Background(), gamedb.StdConn, extraSups)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load extra battle reward.")
		return nil, nil, err
	}

	totalSups = totalSups.Add(extraSups.Total)

	lobbySupporters, err := boiler.BattleLobbySupporters(
		boiler.BattleLobbySupporterWhere.BattleLobbyID.EQ(battleLobbyID),
		boiler.BattleLobbySupporterWhere.DeletedAt.IsNotNull(),
		qm.OrderBy(boiler.BattleLobbySupporterColumns.FactionID),
		qm.Load(boiler.BattleLobbySupporterRels.Faction),
		qm.Load(boiler.BattleLobbySupporterRels.Supporter),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	supportersField := ""

	if len(lobbySupporters) > 0 {
		for _, supporter := range lobbySupporters {
			if supporter.R == nil || supporter.R.Faction == nil || supporter.R.Supporter == nil {
				return nil, nil, errors.New("failed to laod supporters")
			}

			supportersField = fmt.Sprintf("%s%s#%d (%s)\n", supportersField, supporter.R.Supporter.Username.String, supporter.R.Supporter.Gid, supporter.R.Faction.Label)
		}
	} else {
		supportersField = "None"
	}

	embedMessage := embed.NewEmbed()
	gameMap := "Random"

	if battleLobby.GameMapID.Valid {
		if battleLobby.R.GameMap != nil {
			gameMap = battleLobby.R.GameMap.Name
			embedMessage.Image.URL = battleLobby.R.GameMap.BackgroundURL
		}
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Reward Split",
			Value:  "In Percentage (%)",
			Inline: true,
		},
		{
			Name:   ":first_place: 1st",
			Value:  battleLobby.FirstFactionCut.Mul(decimal.NewFromInt(100)).String() + "%",
			Inline: true,
		},
		{
			Name:   ":second_place: 2nd",
			Value:  battleLobby.SecondFactionCut.Mul(decimal.NewFromInt(100)).String() + "%",
			Inline: true,
		},
		{
			Name:   ":third_place: 3rd",
			Value:  battleLobby.ThirdFactionCut.Mul(decimal.NewFromInt(100)).String() + "%",
			Inline: true,
		},
		{
			Name:   ":moneybag: Reward Pool",
			Value:  totalSups.Div(decimal.New(1, 18)).String() + " SUPS",
			Inline: false,
		},
		{
			Name:   ":map: Map",
			Value:  gameMap,
			Inline: false,
		},
		{
			Name:   "Supporters",
			Value:  supportersField,
			Inline: false,
		},
	}

	status := fmt.Sprintf("%s %d/9", DISCORD_BATTLE_LOBBY_WAITING, battleLobbyMechCount)
	canJoin := true
	colour := "efab00"

	if battleLobby.ReadyAt.Valid && !battleLobby.AssignedToBattleID.Valid {
		canJoin = false
		status = DISCORD_BATTLE_LOBBY_QUEUE
	} else if battleLobby.AssignedToBattleID.Valid && !battleLobby.EndedAt.Valid {
		canJoin = false
		status = DISCORD_BATTLE_LOBBY_BATTLE
		colour = "951515"
	} else if battleLobby.EndedAt.Valid {
		canJoin = false
		status = DISCORD_BATTLE_LOBBY_END
		colour = "89e740"
	} else if battleLobby.ExpiresAt.Valid && battleLobby.ExpiresAt.Time.Before(time.Now()) {
		canJoin = false
		status = "Expired"
		colour = "89e740"
	}

	messageComponents := []discordgo.MessageComponent{
		discordgo.Button{
			Label: "Join Lobby",
			Style: discordgo.LinkButton,
			URL:   fmt.Sprintf("%s/lobbies?join=%s", battleArenaBaseUrl, battleLobbyID),
		},
		discordgo.Button{
			Label:    "Follow Lobby",
			Style:    discordgo.PrimaryButton,
			CustomID: "follow-lobby",
			Disabled: !canJoin,
		},
	}

	colourInt, err := strconv.ParseInt(colour, 16, 64)
	if err == nil {
		embedMessage.Color = int(colourInt)
	} else {
		gamelog.L.Err(err).Msg("Failed to parse colour")
	}

	embedMessage.Title = battleLobby.Name
	embedMessage.Author = &discordgo.MessageEmbedAuthor{
		Name: fmt.Sprintf("Hosted By: %s#%d", battleLobby.R.HostBy.Username.String, battleLobby.R.HostBy.Gid),
	}
	embedMessage.Description = fmt.Sprintf("Status: %s", status)
	embedMessage.Fields = fields
	embedMessage.URL = fmt.Sprintf("%s/lobbies?join=%s", battleArenaBaseUrl, battleLobbyID)

	return embedMessage.MessageEmbed, messageComponents, nil
}
