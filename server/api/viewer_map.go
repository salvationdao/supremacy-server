package api

import (
	"context"
	"fmt"
	"server"
	"server/battle_arena"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

/********************
* Viewer Live Count *
********************/

type ViewerCount struct {
	Count int64
}

type ViewerLiveCount struct {
	FactionViewerMap map[server.FactionID]*ViewerCount
	ViewerIDMap      sync.Map
	NetMessageBus    *messagebus.NetBus
}

func NewViewerLiveCount(nmb *messagebus.NetBus, factions []*server.Faction) *ViewerLiveCount {
	vlc := &ViewerLiveCount{
		FactionViewerMap: make(map[server.FactionID]*ViewerCount),
		ViewerIDMap:      sync.Map{},
	}

	vlc.FactionViewerMap[server.FactionID(uuid.Nil)] = &ViewerCount{0}

	for _, f := range factions {
		vlc.FactionViewerMap[f.ID] = &ViewerCount{0}
	}

	go func() {
		for {
			// broadcast to users
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeViewerLiveCountTick))

			payload = append(payload, []byte(fmt.Sprintf(
				"B_%d|R_%d|Z_%d|O_%d",
				vlc.FactionViewerMap[server.BostonCyberneticsFactionID].Count,
				vlc.FactionViewerMap[server.RedMountainFactionID].Count,
				vlc.FactionViewerMap[server.ZaibatsuFactionID].Count,
				vlc.FactionViewerMap[server.FactionID(uuid.Nil)].Count,
			))...)

			nmb.Send(context.Background(), messagebus.NetBusKey(HubKeyViewerLiveCountUpdated), payload)

			// sleep one second
			time.Sleep(1 * time.Second)
		}
	}()

	return vlc
}

func (vcm *ViewerLiveCount) Add(factionID server.FactionID) {
	if fvm, ok := vcm.FactionViewerMap[factionID]; ok {
		atomic.AddInt64(&fvm.Count, 1)
	}
}

func (vcm *ViewerLiveCount) Sub(factionID server.FactionID) {
	if fvm, ok := vcm.FactionViewerMap[factionID]; ok {
		atomic.AddInt64(&fvm.Count, -1)
	}
}

func (vcm *ViewerLiveCount) Swap(oldFactionID, newFactionID server.FactionID) {
	if fvm, ok := vcm.FactionViewerMap[oldFactionID]; ok {
		atomic.AddInt64(&fvm.Count, -1)
	}
	if fvm, ok := vcm.FactionViewerMap[newFactionID]; ok {
		atomic.AddInt64(&fvm.Count, 1)
	}
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
type UserMap struct {
	*ViewerLiveCount
	ClientMap map[string]*UserClientMap
	sync.RWMutex
}

type UserClientMap struct {
	User      *server.User
	ClientMap map[*hub.Client]bool
	sync.RWMutex
}

func NewUserMap(vlc *ViewerLiveCount) *UserMap {
	return &UserMap{
		vlc,
		make(map[string]*UserClientMap),
		sync.RWMutex{},
	}
}

func (um *UserMap) UserRegister(wsc *hub.Client, user *server.User) {
	um.RWMutex.Lock()
	defer um.RWMutex.Unlock()
	um.ViewerLiveCount.IDRecord(user.ID)
	hcm, ok := um.ClientMap[wsc.Identifier()]
	if !ok {
		hcm = &UserClientMap{
			&server.User{},
			make(map[*hub.Client]bool),
			sync.RWMutex{},
		}

		// set up user
		hcm.RWMutex.Lock()
		defer hcm.RWMutex.Unlock()
		hcm.User.ID = user.ID
		hcm.User.Username = user.Username
		hcm.User.FirstName = user.FirstName
		hcm.User.LastName = user.LastName
		hcm.User.AvatarID = user.AvatarID
		hcm.User.FactionID = user.FactionID
		hcm.User.Faction = user.Faction
		hcm.ClientMap[wsc] = true

		um.ClientMap[wsc.Identifier()] = hcm
		return
	}

	hcm.RWMutex.Lock()
	defer hcm.RWMutex.Unlock()
	if _, ok := hcm.ClientMap[wsc]; !ok {
		hcm.ClientMap[wsc] = true
	}
}

func (um *UserMap) GetUserDetail(wsc *hub.Client) *server.User {
	if wsc.Identifier() == "" {
		return nil
	}
	um.RWMutex.RLock()
	defer um.RWMutex.RUnlock()
	cm, ok := um.ClientMap[wsc.Identifier()]
	if !ok {
		return nil
	}
	return cm.User
}

func (um *UserMap) Update(user *server.User) []*hub.Client {
	hcs := []*hub.Client{}
	um.RWMutex.Lock()
	defer um.RWMutex.Unlock()
	hcm, ok := um.ClientMap[user.ID.String()]
	if !ok {
		return nil
	}

	hcm.RWMutex.Lock()
	hcm.User.ID = user.ID
	hcm.User.Username = user.Username
	hcm.User.FirstName = user.FirstName
	hcm.User.LastName = user.LastName
	hcm.User.AvatarID = user.AvatarID
	hcm.User.FactionID = user.FactionID
	hcm.User.Faction = user.Faction
	hcm.RWMutex.Unlock()

	hcm.RWMutex.RLock()
	for cl := range hcm.ClientMap {
		hcs = append(hcs, cl)
	}
	hcm.RWMutex.RUnlock()

	// return broadcast list
	return hcs
}

func (um *UserMap) Remove(wsc *hub.Client) bool {
	if wsc.Identifier() == "" {
		return false
	}

	um.RWMutex.Lock()
	defer um.RWMutex.Unlock()
	hcm, ok := um.ClientMap[wsc.Identifier()]
	if !ok {
		return false
	}

	hcm.RWMutex.Lock()
	defer hcm.RWMutex.Unlock()
	delete(hcm.ClientMap, wsc)

	if len(hcm.ClientMap) == 0 {
		delete(um.ClientMap, wsc.Identifier())
		return true
	}

	return false
}

func (um *UserMap) GetUserDetailByID(userID server.UserID) (*server.User, error) {
	um.RWMutex.RLock()
	defer um.RWMutex.RUnlock()
	hcm, ok := um.ClientMap[userID.String()]
	if !ok {
		return nil, terror.Error(fmt.Errorf("user not found"))
	}
	return hcm.User, nil
}

func (um *UserMap) GetClientsByUserID(userID server.UserID) []*hub.Client {
	hcs := []*hub.Client{}
	um.RWMutex.RLock()
	defer um.RWMutex.RUnlock()
	hcm, ok := um.ClientMap[userID.String()]
	if !ok {
		return hcs
	}

	hcm.RWMutex.RLock()
	defer hcm.RWMutex.RUnlock()
	for cl := range hcm.ClientMap {
		hcs = append(hcs, cl)
	}

	return hcs
}
