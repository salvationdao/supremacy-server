package api

import (
	"context"
	"fmt"
	"net/http"
	"server"
	"server/battle_arena"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

/********************
* Viewer Live Count *
********************/

type ViewerLiveCount map[server.FactionID]*ViewerCount

type ViewerCount struct {
	Count int64
}

// used for tracking user
type ViewerIDMap map[server.UserID]bool

func (api *API) initialiseViewerLiveCount(ctx context.Context, factions []*server.Faction) {
	vlc := make(ViewerLiveCount)

	vim := make(ViewerIDMap)

	vlc[server.FactionID(uuid.Nil)] = &ViewerCount{
		Count: 0,
	}

	for _, f := range factions {
		vlc[f.ID] = &ViewerCount{
			Count: 0,
		}
	}

	// declare live sups spend broadcaster
	tickle.MinDurationOverride = true
	viewLiveCountBroadcasterLogger := log_helpers.NamedLogger(api.Log, "View Live Count Broadcaster").Level(zerolog.Disabled)
	viewLiveCountBroadcaster := tickle.New("Live Sups spend Broadcaster", 1, func() (int, error) {
		// prepare payload
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeViewerLiveCountTick))
		payload = append(payload, []byte(fmt.Sprintf(
			"B_%d|R_%d|Z_%d|O_%d",
			vlc[server.BostonCyberneticsFactionID].Count,
			vlc[server.RedMountainFactionID].Count,
			vlc[server.ZaibatsuFactionID].Count,
			vlc[server.FactionID(uuid.Nil)].Count,
		))...)

		api.NetMessageBus.Send(ctx, messagebus.NetBusKey(HubKeyViewerLiveCountUpdated), payload)

		return http.StatusOK, nil
	})
	viewLiveCountBroadcaster.Log = &viewLiveCountBroadcasterLogger
	viewLiveCountBroadcaster.Start()

	go func() {
		for fn := range api.viewerLiveCount {
			fn(vlc, vim)
		}
	}()
}

func (api *API) viewerLiveCountAdd(factionID server.FactionID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount, vim ViewerIDMap) {
		vlc[factionID].Count += 1
	}
}

func (api *API) viewerLiveCountRemove(factionID server.FactionID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount, vim ViewerIDMap) {
		vlc[factionID].Count -= 1
	}
}

func (api *API) viewerLiveCountSwap(oldFactionID, newFactionID server.FactionID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount, vim ViewerIDMap) {
		vlc[oldFactionID].Count -= 1
		vlc[newFactionID].Count += 1
	}
}

func (api *API) viewerIDRecord(userID server.UserID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount, vim ViewerIDMap) {
		vim[userID] = true
	}
}

func (api *API) viewerIDRead() []server.UserID {
	userIDChan := make(chan []server.UserID)
	api.viewerLiveCount <- func(vlc ViewerLiveCount, vim ViewerIDMap) {
		userIDs := []server.UserID{}
		for userID := range vim {
			userIDs = append(userIDs, userID)
			delete(vim, userID)
		}
		userIDChan <- userIDs
	}
	return <-userIDChan
}

/**********************
* Auth Ring Check Map *
**********************/

type RingCheckAuthMap map[string]*hub.Client

func (api *API) startAuthRignCheckListener() {

	ringCheckAuthMap := make(RingCheckAuthMap)

	go func() {
		for fn := range api.ringCheckAuthChan {
			fn(ringCheckAuthMap)
		}
	}()

}

/********************
* Client Detail Map *
********************/

// startClientTracker track client state
func (api *API) startClientTracker() {
	wscMap := make(map[*hub.Client]*server.User)
	go func() {
		for fn := range api.hubClientDetail {
			fn(wscMap)
		}
	}()
}

// register client detail channel
func (api *API) hubClientDetailRegister(wsc *hub.Client) {
	hcd := &server.User{
		FactionID: server.FactionID(uuid.Nil),
	}
	api.hubClientDetail <- func(m map[*hub.Client]*server.User) {
		if _, ok := m[wsc]; !ok {
			m[wsc] = hcd
		}
	}
}

// remove hub client channel
func (api *API) hubClientDetailRemove(wsc *hub.Client) {
	api.hubClientDetail <- func(m map[*hub.Client]*server.User) {
		delete(m, wsc)
	}
}

// getClientDetailFromChannel return a client detail from client detail channel
func (api *API) getClientDetailFromChannel(wsc *hub.Client) (*server.User, error) {
	detailChan := make(chan *server.User)
	api.hubClientDetail <- func(m map[*hub.Client]*server.User) {
		hcd, ok := m[wsc]
		if !ok {
			detailChan <- nil
			return
		}

		detailChan <- hcd

	}
	result := <-detailChan

	if result == nil {
		return nil, terror.Error(terror.ErrInvalidInput, "Error - Current hub client is not on the map")
	}

	return result, nil
}

// getClientDetailFromUserID return hub client detail by given user id
func (api *API) getClientDetailFromUserID(userID server.UserID) (*server.User, error) {
	// get winner username
	for _, wsc := range api.Hub.FindClients(func(usersClients *hub.Client) bool {
		return usersClients.Identifier() == userID.String()
	}) {
		// get all the hub client instance
		clientDetail, err := api.getClientDetailFromChannel(wsc)
		if err != nil {
			continue
		}

		return clientDetail, nil
	}

	return nil, terror.Error(fmt.Errorf("No hub client found"))
}
