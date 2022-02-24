package api

import (
	"context"
	"fmt"
	"server"
	"server/battle_arena"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

/********************
* Viewer Live Count *
********************/

type ViewerLiveCount struct {
	FactionViewerMap sync.Map
	ViewerIDMap      sync.Map
	NetMessageBus    *messagebus.NetBus
}

func NewViewerLiveCount(nmb *messagebus.NetBus, factions []*server.Faction) *ViewerLiveCount {
	vlc := &ViewerLiveCount{
		FactionViewerMap: sync.Map{},
		ViewerIDMap:      sync.Map{},
	}

	vlc.FactionViewerMap.Store(server.FactionID(uuid.Nil), 0)

	for _, f := range factions {
		vlc.FactionViewerMap.Store(f.ID, 0)
	}

	go func() {
		for {
			// broadcast to users
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeViewerLiveCountTick))

			bc, _ := vlc.FactionViewerMap.Load(server.BostonCyberneticsFactionID)
			rc, _ := vlc.FactionViewerMap.Load(server.RedMountainFactionID)
			zc, _ := vlc.FactionViewerMap.Load(server.ZaibatsuFactionID)
			oc, _ := vlc.FactionViewerMap.Load(server.FactionID(uuid.Nil))
			payload = append(payload, []byte(fmt.Sprintf(
				"B_%d|R_%d|Z_%d|O_%d",
				bc.(int),
				rc.(int),
				zc.(int),
				oc.(int),
			))...)

			nmb.Send(context.Background(), messagebus.NetBusKey(HubKeyViewerLiveCountUpdated), payload)

			// sleep one second
			time.Sleep(1 * time.Second)
		}
	}()

	return vlc
}

func (vcm *ViewerLiveCount) Add(factionID server.FactionID) {
	c, _ := vcm.FactionViewerMap.Load(factionID)
	c = c.(int) + 1
	vcm.FactionViewerMap.Store(factionID, c)
}

func (vcm *ViewerLiveCount) Remove(factionID server.FactionID) {
	c, _ := vcm.FactionViewerMap.Load(factionID)
	c = c.(int) - 1
	vcm.FactionViewerMap.Store(factionID, c)
}

func (vcm *ViewerLiveCount) Swap(oldFactionID, newFactionID server.FactionID) {
	c, _ := vcm.FactionViewerMap.Load(oldFactionID)
	c = c.(int) - 1
	vcm.FactionViewerMap.Store(oldFactionID, c)

	c, _ = vcm.FactionViewerMap.Load(newFactionID)
	c = c.(int) + 1
	vcm.FactionViewerMap.Store(newFactionID, c)
}

func (vcm *ViewerLiveCount) IDRecord(userID server.UserID) {
	vcm.ViewerIDMap.Store(userID, true)
}

func (vcm *ViewerLiveCount) IDRead() []server.UserID {
	userIDs := []server.UserID{}
	vcm.ViewerIDMap.Range(func(key interface{}, value interface{}) bool {
		userIDs = append(userIDs, key.(server.UserID))

		vcm.ViewerIDMap.Delete(key)
		return true
	})

	return userIDs
}

/**********************
* Auth Ring Check Map *
**********************/

type RingCheckAuthMap struct {
	sync.Map
}

func NewRingCheckMap() *RingCheckAuthMap {
	return &RingCheckAuthMap{
		sync.Map{},
	}
}

func (rcm *RingCheckAuthMap) Record(key string, cl *hub.Client) {
	rcm.Store(key, cl)
}

func (rcm *RingCheckAuthMap) Check(key string) (*hub.Client, error) {
	value, ok := rcm.LoadAndDelete(key)
	if !ok {
		return nil, terror.Error(fmt.Errorf("hub client not found"))
	}

	return value.(*hub.Client), nil
}

/********************
* Client Detail Map *
********************/

type ClientDetailMap struct {
	sync.Map
}

func NewClientDetailMap() *ClientDetailMap {
	return &ClientDetailMap{
		sync.Map{},
	}
}

func (cdm *ClientDetailMap) Register(wsc *hub.Client) {
	user := &server.User{}
	cdm.Store(wsc, user)
}

func (cdm *ClientDetailMap) GetDetail(wsc *hub.Client) (*server.User, error) {
	user, ok := cdm.Load(wsc)
	if !ok {
		return nil, terror.Error(fmt.Errorf("client not found"))
	}

	return user.(*server.User), nil
}

func (cdm *ClientDetailMap) Update(wsc *hub.Client, us *server.User) {
	cdm.Store(wsc, us)
}

func (cdm *ClientDetailMap) Remove(wsc *hub.Client) {
	cdm.Delete(wsc)
}

func (cdm *ClientDetailMap) GetDetailByUserID(userID server.UserID) (*server.User, error) {
	var user *server.User
	cdm.Range(func(key, value interface{}) bool {
		u := value.(*server.User)
		if u.ID == userID {
			user = u
			return false
		}
		return true
	})

	if user == nil {
		return nil, terror.Error(fmt.Errorf("user not found"))
	}
	return user, nil
}

type ClientDetail struct {
	hubClient *hub.Client
	detail    *server.User
}

func (cdm *ClientDetailMap) GetDetailsByUserID(userID server.UserID) []*ClientDetail {
	clients := []*ClientDetail{}
	cdm.Range(func(key, value interface{}) bool {
		u := value.(*server.User)
		if u.ID == userID {
			clients = append(clients, &ClientDetail{
				hubClient: key.(*hub.Client),
				detail:    u,
			})
		}
		return true
	})
	return clients
}
