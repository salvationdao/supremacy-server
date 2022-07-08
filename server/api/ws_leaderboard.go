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

	return leaderboardHub
}

// --------------------------------------
const HubKeyPlayerBattlesSpectated = "LEADERBOARD:PLAYER:BATTLES:SPECTATED"

type PlayerBattlesSpectated struct {
	Player          *boiler.Player `json:"player"`
	ViewBattleCount int            `json:"view_battle_count"`
}

func (lc *LeaderboardController) GetPlayerBattlesSpectatedHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	pss, err := boiler.PlayerStats(
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
	for _, ps := range pss {
		resp = append(resp, &PlayerBattlesSpectated{
			Player:          ps.R.IDPlayer,
			ViewBattleCount: ps.ViewBattleCount,
		})
	}

	reply(resp)
	return nil
}
