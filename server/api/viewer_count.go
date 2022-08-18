package api

import (
	"context"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

func (api *API) LiveViewerCount(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	reply(&ViewerLiveCount{})
	api.ViewerUpdateChan <- true
	return nil
}

type ViewerLiveCount struct {
	RedMountain int64 `json:"red_mountain"`
	Boston      int64 `json:"boston"`
	Zaibatsu    int64 `json:"zaibatsu"`
	Other       int64 `json:"other"`
}

const HubKeyViewerLiveCountUpdated = "VIEWER:LIVE:COUNT:UPDATED"

func (api *API) debounceSendingViewerCount() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the debounceSendingViewerCount!", r)
		}
	}()

	interval := 1 * time.Second
	timer := time.NewTimer(interval)
	for {
		select {
		case <-api.ViewerUpdateChan:
			timer.Reset(interval)
		case <-ws.ClientDisconnectedChan:
			timer.Reset(interval)
		case <-timer.C:
			// get user ids from ws connection
			playerIDs := ws.TrackedIdents()

			// cal current online player
			if len(playerIDs) > 0 {
				ps, err := boiler.Players(
					boiler.PlayerWhere.ID.IN(playerIDs),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Debug().Strs("playerIDs", playerIDs).Msg("Failed to query players.")
					continue
				}

				result := &ViewerLiveCount{}
				for _, p := range ps {
					switch p.FactionID.String {
					case server.RedMountainFactionID:
						result.RedMountain += 1
					case server.BostonCyberneticsFactionID:
						result.Boston += 1
					case server.ZaibatsuFactionID:
						result.Zaibatsu += 1
					default:
						result.Other += 1
					}
				}
				ws.PublishMessage("/public/live_viewer_count", HubKeyViewerLiveCountUpdated, result)
			}
		}
	}
}
