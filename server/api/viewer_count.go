package api

import (
	"context"
	"github.com/ninja-syndicate/ws"
	"server/gamelog"
	"time"
)

func (api *API) LiveViewerCount(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	reply(0)
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

	interval := 500 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		select {
		case <-api.ViewerUpdateChan:
			timer.Reset(interval)
		case <-ws.ClientDisconnectedChan:
			timer.Reset(interval)
		case <-timer.C:
			// return total amount of tracked player
			ws.PublishMessage("/public/live_viewer_count", HubKeyViewerLiveCountUpdated, len(ws.TrackedIdents()))
		}
	}
}
