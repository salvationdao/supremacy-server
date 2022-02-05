package api

import (
	"fmt"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
)

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
