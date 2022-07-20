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
	api.Command(HubKeyPlayerMechKills, leaderboardHub.GetPlayerMechKillsHandler)
	api.Command(HubKeyPlayerAbilityKills, leaderboardHub.GetPlayerAbilityKillsHandler)
	api.Command(HubKeyPlayerAbilityTriggers, leaderboardHub.GetPlayerAbilityTriggersHandler)
	api.Command(HubKeyPlayerBattleContributions, leaderboardHub.GetPlayerBattleContributionsHandler)
	api.Command(HubKeyPlayerMechsOwned, leaderboardHub.GetPlayerMechsOwnedHandler)

	return leaderboardHub
}

/**
* Get top players battles spectated
 */
const HubKeyPlayerBattlesSpectated = "LEADERBOARD:PLAYER:BATTLE:SPECTATED"

type PlayerBattlesSpectated struct {
	Player          *boiler.Player `json:"player"`
	ViewBattleCount int            `json:"view_battle_count"`
}

func (lc *LeaderboardController) GetPlayerBattlesSpectatedHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	rows, err := boiler.PlayerStats(
		qm.Select(
			boiler.PlayerStatColumns.ID,
			boiler.PlayerStatColumns.ViewBattleCount,
		),
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
* Get top players mech survivals based on the mechs they own
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

/**
* Get top players mech kills
 */
const HubKeyPlayerMechKills = "LEADERBOARD:PLAYER:MECH:KILLS"

type PlayerMechKills struct {
	Player        *boiler.Player `json:"player"`
	MechKillCount int            `json:"mech_kill_count"`
}

func (lc *LeaderboardController) GetPlayerMechKillsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	rows, err := boiler.PlayerStats(
		qm.Select(
			boiler.PlayerStatColumns.ID,
			boiler.PlayerStatColumns.MechKillCount,
		),
		qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.MechKillCount, db.SortByDirDesc)),
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
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player mech kills.")
		return terror.Error(err, "Failed to get leaderboard player mech kills.")
	}

	resp := []*PlayerMechKills{}
	for _, row := range rows {
		resp = append(resp, &PlayerMechKills{
			Player:        row.R.IDPlayer,
			MechKillCount: row.MechKillCount,
		})
	}

	reply(resp)
	return nil
}

/**
* Get top players ability kills
 */
const HubKeyPlayerAbilityKills = "LEADERBOARD:PLAYER:ABILITY:KILLS"

type PlayerAbilityKills struct {
	Player           *boiler.Player `json:"player"`
	AbilityKillCount int            `json:"ability_kill_count"`
}

func (lc *LeaderboardController) GetPlayerAbilityKillsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	rows, err := boiler.PlayerStats(
		qm.Select(
			boiler.PlayerStatColumns.ID,
			boiler.PlayerStatColumns.AbilityKillCount,
		),
		qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.AbilityKillCount, db.SortByDirDesc)),
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
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player ability kills.")
		return terror.Error(err, "Failed to get leaderboard player ability kills.")
	}

	resp := []*PlayerAbilityKills{}
	for _, row := range rows {
		resp = append(resp, &PlayerAbilityKills{
			Player:           row.R.IDPlayer,
			AbilityKillCount: row.AbilityKillCount,
		})
	}

	reply(resp)
	return nil
}

/**
* Get top players ability triggers
 */
const HubKeyPlayerAbilityTriggers = "LEADERBOARD:PLAYER:ABILITY:TRIGGERS"

type PlayerAbilityTriggers struct {
	Player                *boiler.Player `json:"player"`
	TotalAbilityTriggered int            `json:"total_ability_triggered"`
}

func (lc *LeaderboardController) GetPlayerAbilityTriggersHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	rows, err := boiler.PlayerStats(
		qm.Select(
			boiler.PlayerStatColumns.ID,
			boiler.PlayerStatColumns.TotalAbilityTriggered,
		),
		qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.TotalAbilityTriggered, db.SortByDirDesc)),
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
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player ability triggers.")
		return terror.Error(err, "Failed to get leaderboard player ability triggers.")
	}

	resp := []*PlayerAbilityTriggers{}
	for _, row := range rows {
		resp = append(resp, &PlayerAbilityTriggers{
			Player:                row.R.IDPlayer,
			TotalAbilityTriggered: row.TotalAbilityTriggered,
		})
	}

	reply(resp)
	return nil
}

/**
* Get top players most mech survivals based on the mechs they own
 */
const HubKeyPlayerBattleContributions = "LEADERBOARD:PLAYER:BATTLE:CONTRIBUTIONS"

func (lc *LeaderboardController) GetPlayerBattleContributionsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	resp, err := db.GetPlayerBattleContributions()

	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player battle contributions.")
		return terror.Error(err, "Failed to get leaderboard player battle contributions.")
	}

	reply(resp)
	return nil
}

/**
* Get top players most mech survivals based on the mechs they own
 */
const HubKeyPlayerMechsOwned = "LEADERBOARD:PLAYER:MECHS:OWNED"

func (lc *LeaderboardController) GetPlayerMechsOwnedHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	resp, err := db.GetPlayerMechsOwned()

	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player mechs owned.")
		return terror.Error(err, "Failed to get leaderboard player mechs owned.")
	}

	reply(resp)
	return nil
}
