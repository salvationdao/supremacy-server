package api

import (
	"context"
	"encoding/json"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

// LeaderboardController holds handlers for leaderboard
type LeaderboardController struct {
	API *API
}

func NewLeaderboardController(api *API) *LeaderboardController {
	leaderboardHub := &LeaderboardController{
		API: api,
	}

	api.Command(HubKeyLeaderboardRounds, leaderboardHub.GetLeaderboardRoundsHandler)
	api.Command(HubKeyPlayerBattlesSpectated, leaderboardHub.GetPlayerBattlesSpectatedHandler)
	api.Command(HubKeyPlayerMechSurvives, leaderboardHub.GetPlayerMechSurvivesHandler)
	api.Command(HubKeyPlayerMechKills, leaderboardHub.GetPlayerMechKillsHandler)
	api.Command(HubKeyPlayerAbilityKills, leaderboardHub.GetPlayerAbilityKillsHandler)
	api.Command(HubKeyPlayerAbilityTriggers, leaderboardHub.GetPlayerAbilityTriggersHandler)
	api.Command(HubKeyPlayerMechsOwned, leaderboardHub.GetPlayerMechsOwnedHandler)

	return leaderboardHub
}

const HubKeyLeaderboardRounds = "LEADERBOARD:ROUNDS"

func (lc *LeaderboardController) GetLeaderboardRoundsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	rs, err := boiler.Rounds(
		boiler.RoundWhere.IsInit.EQ(false),
		qm.OrderBy(boiler.RoundColumns.CreatedAt+" DESC"),
		qm.Limit(10),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get leaderboard rounds")
	}

	reply(rs)
	return nil
}

/**
* Get top players battles spectated
 */
const HubKeyPlayerBattlesSpectated = "LEADERBOARD:PLAYER:BATTLE:SPECTATED"

type LeaderboardRequest struct {
	StartTime null.Time `json:"start_time"`
	EndTime   null.Time `json:"end_time"`
}

func (lc *LeaderboardController) GetPlayerBattlesSpectatedHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopBattleViewers(req.StartTime, req.EndTime)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load top battle spectators")
		return err
	}

	reply(resp)
	return nil
}

/**
* Get top players mech survivals based on the mechs they own
 */
const HubKeyPlayerMechSurvives = "LEADERBOARD:PLAYER:MECH:SURVIVES"

func (lc *LeaderboardController) GetPlayerMechSurvivesHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.GetPlayerMechSurvives(req.StartTime, req.EndTime)

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

func (lc *LeaderboardController) GetPlayerMechKillsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopMechKillPlayers(req.StartTime, req.EndTime)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load player mech kill count")
		return err
	}

	reply(resp)
	return nil
}

/**
* Get top players ability kills
 */
const HubKeyPlayerAbilityKills = "LEADERBOARD:PLAYER:ABILITY:KILLS"

func (lc *LeaderboardController) GetPlayerAbilityKillsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopAbilityKillPlayers(req.StartTime, req.EndTime)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load player mech kill count")
		return err
	}

	reply(resp)
	return nil
}

/**
* Get top players ability triggers
 */
const HubKeyPlayerAbilityTriggers = "LEADERBOARD:PLAYER:ABILITY:TRIGGERS"

func (lc *LeaderboardController) GetPlayerAbilityTriggersHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopAbilityTriggerPlayers(req.StartTime, req.EndTime)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load player mech kill count")
		return err
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
