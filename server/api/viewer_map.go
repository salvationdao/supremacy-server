package api

import (
	"context"
	"fmt"
	"server"
	"server/battle_arena"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
)

/********************
* Viewer Live Count *
********************/

type ViewerCount struct {
	Count int64
}

type ViewerLiveCount struct {
	FactionViewerMap map[server.FactionID]*ViewerCount
	ViewerIDMap      deadlock.Map
	NetMessageBus    *messagebus.NetBus
}

func NewViewerLiveCount(nmb *messagebus.NetBus) *ViewerLiveCount {
	vlc := &ViewerLiveCount{
		FactionViewerMap: make(map[server.FactionID]*ViewerCount),
		ViewerIDMap:      deadlock.Map{},
	}

	vlc.FactionViewerMap[server.FactionID(uuid.Nil)] = &ViewerCount{0}
	vlc.FactionViewerMap[server.RedMountainFactionID] = &ViewerCount{0}
	vlc.FactionViewerMap[server.BostonCyberneticsFactionID] = &ViewerCount{0}
	vlc.FactionViewerMap[server.ZaibatsuFactionID] = &ViewerCount{0}

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
	deadlock.Map
}

func NewRingCheckMap() *RingCheckAuthMap {
	return &RingCheckAuthMap{
		deadlock.Map{},
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
	deadlock.RWMutex
}

type UserClientMap struct {
	User      *server.User
	ClientMap map[*hub.Client]bool
	deadlock.RWMutex
}

func NewUserMap(vlc *ViewerLiveCount) *UserMap {
	return &UserMap{
		vlc,
		make(map[string]*UserClientMap),
		deadlock.RWMutex{},
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
			deadlock.RWMutex{},
		}

		// set up user
		hcm.RWMutex.Lock()
		hcm.User.ID = user.ID
		hcm.User.Username = user.Username
		hcm.User.FirstName = user.FirstName
		hcm.User.LastName = user.LastName
		hcm.User.AvatarID = user.AvatarID
		hcm.User.FactionID = user.FactionID
		hcm.User.Faction = user.Faction
		hcm.ClientMap[wsc] = true
		hcm.RWMutex.Unlock()

		um.ClientMap[wsc.Identifier()] = hcm
		return
	}

	hcm.RWMutex.Lock()
	if _, ok := hcm.ClientMap[wsc]; !ok {
		hcm.ClientMap[wsc] = true
	}
	hcm.RWMutex.Unlock()
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
	delete(hcm.ClientMap, wsc)
	hcm.RWMutex.Unlock()

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
	for cl := range hcm.ClientMap {
		hcs = append(hcs, cl)
	}
	hcm.RWMutex.RUnlock()

	return hcs
}
