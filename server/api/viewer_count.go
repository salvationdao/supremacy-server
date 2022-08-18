package api

import (
	"context"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"server/db"
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

			viewerCount := int64(len(ws.TrackedIdents()))

			// multiplier
			if viewerCount > 0 {
				viewerCount = decimal.NewFromInt(viewerCount).
					Mul(db.GetDecimalWithDefault(db.KeyViewerCountMultiplierPercentage, decimal.NewFromInt(120))).
					Div(decimal.NewFromInt(100)).
					Ceil().IntPart()
			}

			ws.PublishMessage("/public/live_viewer_count", HubKeyViewerLiveCountUpdated, viewerCount)
		}
	}
}
