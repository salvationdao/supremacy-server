package api

import (
	"context"
	"database/sql"
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// LeaderboardController holds handlers for leaderboard
type LeaderboardController struct {
	API *API
}

func NewLeaderboardController(api *API) *LeaderboardController {
	leaderboardHub := &LeaderboardController{
		API: api,
	}

	api.Command(HubKeyPlayerBattlesSpectated, leaderboardHub.GetPlayerBattlesSpectatedHandler)
	api.Command(HubKeyPlayerMechSurvives, leaderboardHub.GetPlayerMechSurvivesHandler)

	return leaderboardHub
}

/**
* Get players battles spectated
 */
const HubKeyPlayerBattlesSpectated = "LEADERBOARD:PLAYER:BATTLES:SPECTATED"

type PlayerBattlesSpectated struct {
	Player          *boiler.Player `json:"player"`
	ViewBattleCount int            `json:"view_battle_count"`
}

func (lc *LeaderboardController) GetPlayerBattlesSpectatedHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	rows, err := boiler.PlayerStats(
		qm.Select(boiler.PlayerStatColumns.ViewBattleCount),
		qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.ViewBattleCount, db.SortByDirDesc)),
		qm.Limit(10),
		qm.Load(
			boiler.PlayerStatRels.IDPlayer,
			qm.Select(
				boiler.PlayerColumns.ID,
				boiler.PlayerColumns.Username,
				boiler.PlayerColumns.FactionID,
				boiler.PlayerColumns.Gid,
				boiler.PlayerColumns.Rank,
			),
		),
	).All(gamedb.StdConn)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player battles spectated.")
		return terror.Error(err, "Failed to get leaderboard player battles spectated.")
	}

	resp := []*PlayerBattlesSpectated{}
	for _, row := range rows {
		resp = append(resp, &PlayerBattlesSpectated{
			Player:          row.R.IDPlayer,
			ViewBattleCount: row.ViewBattleCount,
		})
	}

	reply(resp)
	return nil
}

/**
* Get players most mech survivals based on the mechs they own
 */
const HubKeyPlayerMechSurvives = "LEADERBOARD:PLAYER:MECH:SURVIVES"

func (lc *LeaderboardController) GetPlayerMechSurvivesHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	resp, err := db.GetPlayerMechSurvives()

	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player mech survives.")
		return terror.Error(err, "Failed to get leaderboard player mech survives.")
	}

	reply(resp)
	return nil
}
