package api

import (
	"fmt"
	"net/http"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
)

/********************
* Client Detail Map *
********************/

type HubClientDetail struct {
	ID        server.UserID
	FactionID server.FactionID
}

// startClientTracker track client state
func (api *API) startClientTracker(wsc *hub.Client) {
	// initialise online client
	hubClientDetail := &HubClientDetail{
		ID:        server.UserID(uuid.Nil),
		FactionID: server.FactionID(uuid.Nil),
	}

	go func() {
		for {
			select {
			case fn := <-api.hubClientDetail[wsc]:
				fn(hubClientDetail)
			}
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

/**********************
* Client Instance Map *
**********************/

// MinimumConnectSecond the amount of second that a user have to be connected to earn one channel point
const MinimumConnectSecond = 10

type ClientInstanceMap map[*hub.Client]bool

// SupremacyTokenState record the state of channel point of current user
type SupremacyTokenState struct {
	SupremacyToken int64
}

// startOnlineClientTracker is a channel that track online client instances
func (api *API) startOnlineClientTracker(hubClientID server.UserID, connectPoint int64) {
	clientInstanceMap := make(ClientInstanceMap)

	connectPointState := &SupremacyTokenState{connectPoint}

	// create a channel point tickle
	taskTickle := tickle.New("FactionID Channel Point Ticker", MinimumConnectSecond, api.connectPointTickleFactory(hubClientID))
	taskTickle.Log = log_helpers.NamedLogger(api.Log, "FactionID Channel Point Ticker")
	if !hubClientID.IsNil() {
		taskTickle.Start()
	}

	go func() {
		for {
			select {
			case fn := <-api.onlineClientMap[hubClientID]:
				fn(clientInstanceMap, connectPointState, taskTickle)
			}
		}
	}()
}

// connectPointTickleFactory generate a channel point tickle task for tickle
func (api *API) connectPointTickleFactory(hubClientID server.UserID) func() (int, error) {
	fn := func() (int, error) {
		// skip, if client is not signed in
		if hubClientID.IsNil() {
			return http.StatusOK, nil
		}

		// increment the client's channel point
		api.onlineClientMap[hubClientID] <- func(cim ClientInstanceMap, cps *SupremacyTokenState, t *tickle.Tickle) {
			if cps == nil {
				return
			}

			cps.SupremacyToken += 1
			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTwitchSupremacyTokenUpdated, hubClientID)), cps.SupremacyToken)
		}

		return http.StatusOK, nil
	}

	return fn
}
