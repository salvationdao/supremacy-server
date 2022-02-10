package api

import (
	"fmt"
	"net/http"
	"server"
	"server/battle_arena"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
)

/********************
* Viewer Live Count *
********************/

type ViewerLiveCount map[server.FactionID]*ViewerCount

type ViewerCount struct {
	Count int64
}

func (api *API) initialiseViewerLiveCount(factions []*server.Faction) {
	vlc := make(ViewerLiveCount)

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

		api.Hub.Clients(func(clients hub.ClientsList) {
			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					err := c.SendWithMessageType(payload, websocket.MessageBinary)
					if err != nil {
						api.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})

		return http.StatusOK, nil
	})
	viewLiveCountBroadcaster.Log = &viewLiveCountBroadcasterLogger
	viewLiveCountBroadcaster.Start()

	go func() {
		for fn := range api.viewerLiveCount {
			fn(vlc)
		}
	}()
}

func (api *API) viewerLiveCountAdd(factionID server.FactionID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount) {
		vlc[factionID].Count += 1
	}
}

func (api *API) viewerLiveCountRemove(factionID server.FactionID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount) {
		vlc[factionID].Count -= 1
	}
}

func (api *API) viewerLiveCountSwap(oldFactionID, newFactionID server.FactionID) {
	api.viewerLiveCount <- func(vlc ViewerLiveCount) {
		vlc[oldFactionID].Count -= 1
		vlc[newFactionID].Count += 1
	}
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

type HubClientDetail struct {
	FactionID server.FactionID
	FirstName string
	LastName  string
	Username  string
	avatarID  *server.BlobID
}

// startClientTracker track client state
func (api *API) startClientTracker(wsc *hub.Client) {
	// initialise online client
	hubClientDetail := &HubClientDetail{
		FactionID: server.FactionID(uuid.Nil),
	}

	go func() {
		for fn := range api.hubClientDetail[wsc] {
			fn(hubClientDetail)
		}
	}()
}

// getClientDetailFromChannel return a client detail from client detail channel
func (api *API) getClientDetailFromChannel(wsc *hub.Client) (*HubClientDetail, error) {
	hubClientDetailChan, ok := api.hubClientDetail[wsc]
	if !ok {
		return nil, terror.Error(terror.ErrInvalidInput, "Error - Current hub client is not on the map")
	}

	detailChan := make(chan *HubClientDetail)
	hubClientDetailChan <- func(hcd *HubClientDetail) {
		detailChan <- hcd
	}

	result := *<-detailChan

	return &result, nil
}

// getClientDetailFromUserID return hub client detail by given user id
func (api *API) getClientDetailFromUserID(userID server.UserID) (*HubClientDetail, error) {
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
