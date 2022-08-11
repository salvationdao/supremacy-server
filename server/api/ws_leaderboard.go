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

func NewLeaderboardController(api *API) {
	api.Command(HubKeyLeaderboardRounds, api.GetLeaderboardRoundsHandler)
	api.Command(HubKeyPlayerBattlesSpectatedLeaderboard, api.GetPlayerBattlesSpectatedLeaderboardHandler)
	api.Command(HubKeyPlayerMechSurvivesLeaderboard, api.GetPlayerMechSurvivesLeaderboardHandler)
	api.Command(HubKeyPlayerMechKillsLeaderboard, api.GetPlayerMechKillsLeaderboardHandler)
	api.Command(HubKeyPlayerAbilityKillsLeaderboard, api.GetPlayerAbilityKillsLeaderboardHandler)
	api.Command(HubKeyPlayerAbilityTriggersLeaderboard, api.GetPlayerAbilityTriggersLeaderboardHandler)
	api.Command(HubKeyPlayerMechsOwnedLeaderboard, api.GetPlayerMechsOwnedLeaderboardHandler)
	api.Command(HubKeyPlayerRepairBlockLeaderboard, api.GetPlayerRepairBlockLeaderboardHandler)
}

const HubKeyLeaderboardRounds = "LEADERBOARD:ROUNDS"

func (api *API) GetLeaderboardRoundsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
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
const HubKeyPlayerBattlesSpectatedLeaderboard = "LEADERBOARD:PLAYER:BATTLE:SPECTATED"

type LeaderboardRequest struct {
	Payload struct {
		RoundID null.String `json:"round_id"`
	} `json:"payload"`
}

func (api *API) GetPlayerBattlesSpectatedLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopBattleViewers(req.Payload.RoundID)
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
const HubKeyPlayerMechSurvivesLeaderboard = "LEADERBOARD:PLAYER:MECH:SURVIVES"

func (api *API) GetPlayerMechSurvivesLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.GetPlayerMechSurvives(req.Payload.RoundID)

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
const HubKeyPlayerMechKillsLeaderboard = "LEADERBOARD:PLAYER:MECH:KILLS"

func (api *API) GetPlayerMechKillsLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopMechKillPlayers(req.Payload.RoundID)
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
const HubKeyPlayerAbilityKillsLeaderboard = "LEADERBOARD:PLAYER:ABILITY:KILLS"

func (api *API) GetPlayerAbilityKillsLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopAbilityKillPlayers(req.Payload.RoundID)
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
const HubKeyPlayerAbilityTriggersLeaderboard = "LEADERBOARD:PLAYER:ABILITY:TRIGGERS"

func (api *API) GetPlayerAbilityTriggersLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopAbilityTriggerPlayers(req.Payload.RoundID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load player mech kill count")
		return err
	}

	reply(resp)
	return nil
}

/**
* Get top 10 players who repair the most blocks
 */

const HubKeyPlayerRepairBlockLeaderboard = "LEADERBOARD:PLAYER:REPAIR:BLOCK"

func (api *API) GetPlayerRepairBlockLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &LeaderboardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resp, err := db.TopRepairBlockPlayers(req.Payload.RoundID)
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
const HubKeyPlayerMechsOwnedLeaderboard = "LEADERBOARD:PLAYER:MECHS:OWNED"

func (api *API) GetPlayerMechsOwnedLeaderboardHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	resp, err := db.GetPlayerMechsOwned()

	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get leaderboard player mechs owned.")
		return terror.Error(err, "Failed to get leaderboard player mechs owned.")
	}

	reply(resp)
	return nil
}
