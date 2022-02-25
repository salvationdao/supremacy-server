package api

import (
	"server"
	"server/battle_arena"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

type UserMultiplier struct {
	CurrentMaps *Multiplier
	CheckMaps   *Multiplier

	BattleIDMap sync.Map
	UserMap     *UserMap
}

type Multiplier struct {
	OnlineMap         sync.Map
	ApplauseMap       sync.Map
	PickedLocationMap sync.Map

	// battle multiplier
	WinningFactionMap sync.Map
	WinningUserMap    sync.Map
	KillMap           sync.Map
}

type MultiplierAction struct {
	MultiplierValue int
	Expiry          time.Time
}

// TODO: set up sups ticker
func NewUserMultiplier(userMap *UserMap) *UserMultiplier {
	um := &UserMultiplier{
		CurrentMaps: &Multiplier{sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}},
		CheckMaps:   &Multiplier{sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}},
		BattleIDMap: sync.Map{},
		UserMap:     userMap,
	}

	return um
}

// Online handle user online multiplier
func (um *UserMultiplier) Online(userID server.UserID) {
	userIDStr := userID.String()
	now := time.Now()
	// go through check map and get non expired multiplier
	um.CheckMaps.OnlineMap.Delete(userIDStr)
	um.CurrentMaps.OnlineMap.Store(userIDStr, &MultiplierAction{
		MultiplierValue: 100,
		Expiry:          now.AddDate(1, 0, 0),
	})

	// applause map
	if v, ok := um.CheckMaps.ApplauseMap.LoadAndDelete(userIDStr); ok {
		if m := v.(*MultiplierAction); m.Expiry.After(now) {
			um.CurrentMaps.ApplauseMap.Store(userIDStr, m)
		}
	}

	// picked location map
	if v, ok := um.CheckMaps.PickedLocationMap.LoadAndDelete(userIDStr); ok {
		if m := v.(*MultiplierAction); m.Expiry.After(now) {
			um.CurrentMaps.PickedLocationMap.Store(userIDStr, m)
		}
	}

	// move battle map
	um.BattleIDMap.Range(func(key, value interface{}) bool {
		battleIDStr := key.(string)

		// winning faction map
		if v, ok := um.CheckMaps.WinningFactionMap.LoadAndDelete(battleIDStr + "_" + userIDStr); ok {
			if m := v.(*MultiplierAction); m.Expiry.After(now) {
				um.CurrentMaps.WinningFactionMap.Store(battleIDStr+"_"+userIDStr, m)
			}
		}

		// winner map
		if v, ok := um.CheckMaps.WinningUserMap.LoadAndDelete(battleIDStr + "_" + userIDStr); ok {
			if m := v.(*MultiplierAction); m.Expiry.After(now) {
				um.CurrentMaps.WinningUserMap.Store(battleIDStr+"_"+userIDStr, m)
			}
		}

		// kill map
		if v, ok := um.CheckMaps.KillMap.LoadAndDelete(battleIDStr + "_" + userIDStr); ok {
			if m := v.(*MultiplierAction); m.Expiry.After(now) {
				um.CurrentMaps.KillMap.Store(battleIDStr+"_"+userIDStr, m)
			}
		}
		return true
	})
}

// Offline remove all the user related multiplier action in current map
func (um *UserMultiplier) Offline(userID server.UserID) {
	userIDStr := userID.String()

	um.CurrentMaps.OnlineMap.Delete(userIDStr)
	um.CurrentMaps.ApplauseMap.Delete(userIDStr)
	um.CurrentMaps.PickedLocationMap.Delete(userIDStr)

	um.BattleIDMap.Range(func(key, value interface{}) bool {
		um.CurrentMaps.WinningFactionMap.Range(func(key, value interface{}) bool {
			if strings.HasSuffix(key.(string), userIDStr) {
				um.CurrentMaps.WinningFactionMap.Delete(userIDStr)
			}

			return true
		})

		um.CurrentMaps.WinningUserMap.Range(func(key, value interface{}) bool {
			if strings.HasSuffix(key.(string), userIDStr) {
				um.CurrentMaps.WinningUserMap.Delete(userIDStr)
			}

			return true
		})

		um.CurrentMaps.KillMap.Range(func(key, value interface{}) bool {
			if strings.HasSuffix(key.(string), userIDStr) {
				um.CurrentMaps.KillMap.Delete(userIDStr)
			}

			return true
		})

		return true
	})
}

func (um *UserMultiplier) Voted(userID server.UserID) {
	userIDStr := userID.String()
	um.CurrentMaps.ApplauseMap.Store(userIDStr, &MultiplierAction{
		MultiplierValue: 50,
		Expiry:          time.Now().Add(time.Minute * 30),
	})
}

func (um *UserMultiplier) ClientPickedLocation(userID server.UserID) {
	userIDStr := userID.String()
	um.CurrentMaps.ApplauseMap.Store(userIDStr, &MultiplierAction{
		MultiplierValue: 50,
		Expiry:          time.Now().Add(time.Minute * 30),
	})
}

func (um *UserMultiplier) ClientBattleRewardUpdate(brl *battle_arena.BattleRewardList) {
	battleIDStr := brl.BattleID.String()
	now := time.Now()
	// remove battle id from battle id map
	go um.CleanUpBattleReward(battleIDStr)

	um.BattleIDMap.Store(battleIDStr, true)
	// give winning war machine
	for wid := range brl.WinningWarMachineOwnerIDs {
		um.CurrentMaps.WinningUserMap.Store(battleIDStr+"_"+wid.String(), &MultiplierAction{
			MultiplierValue: 500,
			Expiry:          now.Add(time.Minute * 5),
		})
	}

	// send war machine kill reward
	for kid := range brl.ExecuteKillWarMachineOwnerIDs {
		um.CurrentMaps.KillMap.Store(battleIDStr+"_"+kid.String(), &MultiplierAction{
			MultiplierValue: 500,
			Expiry:          now.Add(time.Minute * 5),
		})
	}

	// loop through current online user and provide them winning faction reward
	um.UserMap.RLock()
	for userIDstr, cm := range um.UserMap.ClientMap {
		if cm.User.FactionID != brl.WinnerFactionID {
			continue
		}

		// store user to winning faction map
		um.CurrentMaps.WinningFactionMap.Store(battleIDStr+"_"+userIDstr, &MultiplierAction{
			MultiplierValue: 1000,
			Expiry:          time.Now().Add(time.Minute * 5),
		})
	}
	um.UserMap.RUnlock()
}

func (um *UserMultiplier) CleanUpBattleReward(battleIDStr string) {
	time.Sleep(5 * time.Minute)
	// remove all the battle related reward from check map and current map
	// NOTE: check map should be cleaned up first,
	//		 otherwise it will rewrite the value back to current map in the check function

	go func() {
		um.CheckMaps.WinningFactionMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CheckMaps.WinningFactionMap.Delete(key)
			return true
		})

		um.CurrentMaps.WinningFactionMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CurrentMaps.WinningFactionMap.Delete(key)
			return true
		})
	}()

	go func() {
		um.CheckMaps.WinningUserMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CheckMaps.WinningUserMap.Delete(key)
			return true
		})

		um.CurrentMaps.WinningUserMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CurrentMaps.WinningUserMap.Delete(key)
			return true
		})
	}()

	go func() {
		um.CheckMaps.KillMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CheckMaps.KillMap.Delete(key)
			return true
		})

		um.CurrentMaps.KillMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CurrentMaps.KillMap.Delete(key)
			return true
		})
	}()
}

func (um *UserMultiplier) SupsTick() {
	userMap := make(map[int][]server.UserID)
	now := time.Now()

	// check online reward
	um.CurrentMaps.OnlineMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.OnlineMap.Delete(key)
			return true
		}

		// append user to the ticking list
		userID := server.UserID(uuid.FromStringOrNil(key.(string)))
		if _, ok := userMap[m.MultiplierValue]; !ok {
			userMap[m.MultiplierValue] = []server.UserID{}
		}
		userMap[m.MultiplierValue] = append(userMap[m.MultiplierValue], userID)
		return true
	})

	// check applause reward
	um.CurrentMaps.ApplauseMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.ApplauseMap.Delete(key)
			return true
		}

		// append user to the ticking list
		userID := server.UserID(uuid.FromStringOrNil(key.(string)))
		if _, ok := userMap[m.MultiplierValue]; !ok {
			userMap[m.MultiplierValue] = []server.UserID{}
		}
		userMap[m.MultiplierValue] = append(userMap[m.MultiplierValue], userID)
		return true
	})
	um.CurrentMaps.PickedLocationMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.PickedLocationMap.Delete(key)
			return true
		}

		// append user to the ticking list
		userID := server.UserID(uuid.FromStringOrNil(key.(string)))
		if _, ok := userMap[m.MultiplierValue]; !ok {
			userMap[m.MultiplierValue] = []server.UserID{}
		}
		userMap[m.MultiplierValue] = append(userMap[m.MultiplierValue], userID)
		return true
	})

	um.CurrentMaps.WinningFactionMap.Range(func(key, value interface{}) bool {

		return true
	})

	um.CurrentMaps.WinningUserMap.Range(func(key, value interface{}) bool {

		return true
	})

	um.CurrentMaps.KillMap.Range(func(key, value interface{}) bool {

		return true
	})

}
