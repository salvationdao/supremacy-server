package api

import (
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/terror/v2"
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
