package api

import (
	"context"
	"server"
	"server/battle_arena"
	"server/passport"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

type ClientAction string

const (
	ClientOnline                ClientAction = "Online"
	ClientOffline               ClientAction = "Offline"
	ClientVoted                 ClientAction = "Applause"
	ClientPickedLocation        ClientAction = "Picked Location"
	ClientBattleRewardUpdate    ClientAction = "BattleRewardUpdate"
	ClientSupsMultiplierGet     ClientAction = "SupsMultiplierGet"
	ClientCheckMultiplierUpdate ClientAction = "CheckMultiplierUpdate"
	ClientSupsTick              ClientAction = "SupsTick"
)

type BattleRewardType string

const (
	BattleRewardTypeFaction         BattleRewardType = "Battle Faction Reward"
	BattleRewardTypeWinner          BattleRewardType = "Battle Winner Reward"
	BattleRewardTypeKill            BattleRewardType = "Battle Kill Reward"
	BattleRewardTypeAbilityExecutor BattleRewardType = "Ability Executor"
	BattleRewardTypeInfluencer      BattleRewardType = "Battle Influencer"
	BattleRewardTypeWarContributor  BattleRewardType = "War Contributor"
)

type UserMultiplier struct {
	CurrentMaps *Multiplier
	CheckMaps   *Multiplier

	BattleIDMap sync.Map

	// other dependencies
	UserMap     *UserMap
	Passport    *passport.Passport
	BattleArena *battle_arena.BattleArena

	ActiveMap *sync.Map
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
func NewUserMultiplier(userMap *UserMap, pp *passport.Passport, ba *battle_arena.BattleArena) *UserMultiplier {
	um := &UserMultiplier{
		CurrentMaps: &Multiplier{sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}},
		CheckMaps:   &Multiplier{sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}, sync.Map{}},
		BattleIDMap: sync.Map{},
		UserMap:     userMap,
		Passport:    pp,
		BattleArena: ba,

		ActiveMap: &sync.Map{},
	}

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if ba.BattleActive() {
				// check user active list
				um.UserActiveChecker()

				// distribute sups
				um.SupsTick()
			}
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			um.UserMultiplierUpdate()
		}
	}()

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

	go um.UserSupsMultiplierToPassport(userID, nil)
}

func (um *UserMultiplier) Voted(userID server.UserID) {
	userIDStr := userID.String()

	// update user active check
	um.ActiveMap.Store(userIDStr, time.Now())

	um.CurrentMaps.ApplauseMap.Store(userIDStr, &MultiplierAction{
		MultiplierValue: 50,
		Expiry:          time.Now().Add(time.Minute * 30),
	})
}

func (um *UserMultiplier) PickedLocation(userID server.UserID) {
	userIDStr := userID.String()

	um.CurrentMaps.PickedLocationMap.Store(userIDStr, &MultiplierAction{
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
	um.BattleIDMap.Delete(battleIDStr)
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

// sups tick
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

		// skip, if user is not active
		userIDstr := key.(string)
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userIDstr)
		// return if no remain
		if remain == 0 {
			return true
		}
		multiplierValue = multiplierValue * remain / 100

		// append user to the ticking list
		userID := server.UserID(uuid.FromStringOrNil(key.(string)))
		if _, ok := userMap[multiplierValue]; !ok {
			userMap[multiplierValue] = []server.UserID{}
		}
		userMap[multiplierValue] = append(userMap[multiplierValue], userID)
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

		// skip, if user is not active
		userIDstr := key.(string)
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userIDstr)
		// return if no remain
		if remain == 0 {
			return true
		}
		multiplierValue = multiplierValue * remain / 100

		// append user to the ticking list
		userID := server.UserID(uuid.FromStringOrNil(key.(string)))
		if _, ok := userMap[multiplierValue]; !ok {
			userMap[multiplierValue] = []server.UserID{}
		}
		userMap[multiplierValue] = append(userMap[multiplierValue], userID)
		return true
	})
	um.CurrentMaps.PickedLocationMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.PickedLocationMap.Delete(key)
			return true
		}

		// skip, if user is not active
		userIDstr := key.(string)
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userIDstr)
		// return if no remain
		if remain == 0 {
			return true
		}
		multiplierValue = multiplierValue * remain / 100

		// append user to the ticking list
		userID := server.UserID(uuid.FromStringOrNil(key.(string)))
		if _, ok := userMap[multiplierValue]; !ok {
			userMap[multiplierValue] = []server.UserID{}
		}
		userMap[multiplierValue] = append(userMap[multiplierValue], userID)
		return true
	})

	um.CurrentMaps.WinningFactionMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.WinningFactionMap.Delete(key)
			return true
		}

		// append user to the ticking list
		k := key.(string)
		uidStr := strings.Split(k, "_")[1]
		userID := server.UserID(uuid.FromStringOrNil(uidStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, uidStr)
		// return if no remain
		if remain == 0 {
			return true
		}
		multiplierValue = multiplierValue * remain / 100

		if _, ok := userMap[multiplierValue]; !ok {
			userMap[multiplierValue] = []server.UserID{}
		}
		userMap[multiplierValue] = append(userMap[multiplierValue], userID)
		return true
	})

	um.CurrentMaps.WinningUserMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.WinningUserMap.Delete(key)
			return true
		}

		// append user to the ticking list
		k := key.(string)
		uidStr := strings.Split(k, "_")[1]
		userID := server.UserID(uuid.FromStringOrNil(uidStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, uidStr)
		// return if no remain
		if remain == 0 {
			return true
		}
		multiplierValue = multiplierValue * remain / 100

		if _, ok := userMap[multiplierValue]; !ok {
			userMap[multiplierValue] = []server.UserID{}
		}
		userMap[multiplierValue] = append(userMap[multiplierValue], userID)
		return true
	})

	um.CurrentMaps.KillMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.KillMap.Delete(key)
			return true
		}

		// append user to the ticking list
		k := key.(string)
		uidStr := strings.Split(k, "_")[1]
		userID := server.UserID(uuid.FromStringOrNil(uidStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, uidStr)
		// return if no remain
		if remain == 0 {
			return true
		}
		multiplierValue = multiplierValue * remain / 100

		if _, ok := userMap[multiplierValue]; !ok {
			userMap[multiplierValue] = []server.UserID{}
		}
		userMap[multiplierValue] = append(userMap[multiplierValue], userID)
		return true
	})

	um.Passport.SendTickerMessage(userMap)
}

// PushUserMultiplierToPassport push the multiplier actions list of the user to passport user
func (um *UserMultiplier) PushUserMultiplierToPassport(userID server.UserID) {
	uidStr := userID.String()
	mas := make(map[string]*MultiplierAction)
	now := time.Now()

	// online
	if value, ok := um.CurrentMaps.OnlineMap.Load(uidStr); ok {
		ma := value.(*MultiplierAction)
		if ma.Expiry.After(now) {
			mas[string(ClientOnline)] = ma
		}
	}
	if value, ok := um.CurrentMaps.ApplauseMap.Load(uidStr); ok {
		ma := value.(*MultiplierAction)
		if ma.Expiry.After(now) {
			mas[string(ClientVoted)] = ma
		}
	}
	if value, ok := um.CurrentMaps.PickedLocationMap.Load(uidStr); ok {
		ma := value.(*MultiplierAction)
		if ma.Expiry.After(now) {
			mas[string(ClientPickedLocation)] = ma
		}
	}

	// battle rewards
	um.BattleIDMap.Range(func(key, value interface{}) bool {
		battleID := key.(string)

		if value, ok := um.CurrentMaps.WinningFactionMap.Load(battleID + "_" + uidStr); ok {
			ma := value.(*MultiplierAction)
			if ma.Expiry.After(now) {
				mas[string(BattleRewardTypeFaction)+"_"+battleID] = ma
			}
		}

		if value, ok := um.CurrentMaps.WinningUserMap.Load(battleID + "_" + uidStr); ok {
			ma := value.(*MultiplierAction)
			if ma.Expiry.After(now) {
				mas[string(BattleRewardTypeWinner)+"_"+battleID] = ma
			}
		}

		if value, ok := um.CurrentMaps.KillMap.Load(battleID + "_" + uidStr); ok {
			ma := value.(*MultiplierAction)
			if ma.Expiry.After(now) {
				mas[string(BattleRewardTypeKill)+"_"+battleID] = ma
			}
		}

		return true
	})

	if len(mas) == 0 {
		go um.UserSupsMultiplierToPassport(userID, nil)
		return
	}

	go um.UserSupsMultiplierToPassport(userID, mas)
}

func (um *UserMultiplier) UserSupsMultiplierToPassport(userID server.UserID, supsMultiplierMap map[string]*MultiplierAction) {
	userSupsMultiplierSend := &passport.UserSupsMultiplierSend{
		ToUserID:        userID,
		SupsMultipliers: []*passport.SupsMultiplier{},
	}

	for key, sm := range supsMultiplierMap {
		userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &passport.SupsMultiplier{
			Key:       key,
			Value:     sm.MultiplierValue,
			ExpiredAt: sm.Expiry,
		})
	}

	go um.Passport.UserSupsMultiplierSend(context.Background(), []*passport.UserSupsMultiplierSend{userSupsMultiplierSend})
}

func (um *UserMultiplier) UserMultiplierUpdate() {
	// map[userID]map[reward text] &MultiplierAction
	diff := make(map[string]map[string]*MultiplierAction)
	now := time.Now()

	// check current map with check map, add any different from the cache
	um.CurrentMaps.OnlineMap.Range(func(key, value interface{}) bool {
		uidStr := key.(string)
		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			return true
		}

		// get data from check map
		_, ok := um.CheckMaps.OnlineMap.Load(uidStr)
		// record, if not exists
		if !ok {
			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(ClientOnline)] = currentValue
			diff[uidStr] = d

			// update check map
			um.CheckMaps.OnlineMap.Store(uidStr, currentValue)
			return true
		}

		// store different
		d, ok := diff[uidStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(ClientOnline)] = currentValue
		diff[uidStr] = d
		// update check map
		um.CheckMaps.OnlineMap.Store(uidStr, currentValue)

		return true
	})

	// check current map with check map, add any different from the cache
	um.CurrentMaps.ApplauseMap.Range(func(key, value interface{}) bool {
		uidStr := key.(string)
		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			return true
		}
		// get data from check map
		_, ok := um.CheckMaps.ApplauseMap.Load(uidStr)
		// record, if not exists
		if !ok {
			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(ClientVoted)] = currentValue
			diff[uidStr] = d
			// update check map
			um.CheckMaps.ApplauseMap.Store(uidStr, currentValue)

			return true
		}

		// store different
		d, ok := diff[uidStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(ClientVoted)] = currentValue
		diff[uidStr] = d
		// update check map
		um.CheckMaps.ApplauseMap.Store(uidStr, currentValue)

		return true
	})

	// check current map with check map, add any different from the cache
	um.CurrentMaps.PickedLocationMap.Range(func(key, value interface{}) bool {
		uidStr := key.(string)
		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			return true
		}
		// get data from check map
		_, ok := um.CheckMaps.PickedLocationMap.Load(uidStr)
		// record, if not exists
		if !ok {
			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(ClientPickedLocation)] = currentValue
			diff[uidStr] = d
			// update check map
			um.CheckMaps.PickedLocationMap.Store(uidStr, currentValue)

			return true
		}

		// store different
		d, ok := diff[uidStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(ClientPickedLocation)] = currentValue
		diff[uidStr] = d
		// update check map
		um.CheckMaps.PickedLocationMap.Store(uidStr, currentValue)

		return true
	})

	// battle rewards
	um.BattleIDMap.Range(func(key, value interface{}) bool {
		battleID := key.(string)
		// check current map with check map, add any different from the cache
		um.CurrentMaps.WinningFactionMap.Range(func(key, value interface{}) bool {
			innerBattleID := strings.Split(key.(string), "_")[0]
			// check inner battle id is the same
			if innerBattleID != battleID {
				return true
			}

			// check value
			uidStr := strings.Split(key.(string), "_")[1]
			currentValue := value.(*MultiplierAction)
			if currentValue.Expiry.Before(now) {
				return true
			}
			// get data from check map
			_, ok := um.CheckMaps.WinningFactionMap.Load(battleID + "_" + uidStr)
			// record, if not exists
			if !ok {
				// store different
				d, ok := diff[uidStr]
				if !ok {
					d = make(map[string]*MultiplierAction)
				}
				d[string(BattleRewardTypeFaction)+"_"+battleID] = currentValue
				diff[uidStr] = d
				// update check map
				um.CheckMaps.WinningFactionMap.Store(battleID+"_"+uidStr, currentValue)

				return true
			}

			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(BattleRewardTypeFaction)+"_"+battleID] = currentValue
			diff[uidStr] = d
			// update check map
			um.CheckMaps.WinningFactionMap.Store(battleID+"_"+uidStr, currentValue)

			return true
		})

		// check current map with check map, add any different from the cache
		um.CurrentMaps.WinningUserMap.Range(func(key, value interface{}) bool {
			innerBattleID := strings.Split(key.(string), "_")[0]
			// check inner battle id is the same
			if innerBattleID != battleID {
				return true
			}

			uidStr := strings.Split(key.(string), "_")[1]
			currentValue := value.(*MultiplierAction)
			if currentValue.Expiry.Before(now) {
				return true
			}
			// get data from check map
			_, ok := um.CheckMaps.WinningUserMap.Load(battleID + "_" + uidStr)
			// record, if not exists
			if !ok {
				// store different
				d, ok := diff[uidStr]
				if !ok {
					d = make(map[string]*MultiplierAction)
				}
				d[string(BattleRewardTypeWinner)+"_"+battleID] = currentValue
				diff[uidStr] = d
				// update check map
				um.CheckMaps.WinningUserMap.Store(battleID+"_"+uidStr, currentValue)

				return true
			}

			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(BattleRewardTypeWinner)+"_"+battleID] = currentValue
			diff[uidStr] = d
			// update check map
			um.CheckMaps.WinningUserMap.Store(battleID+"_"+uidStr, currentValue)

			return true
		})

		// check current map with check map, add any different from the cache
		um.CurrentMaps.KillMap.Range(func(key, value interface{}) bool {
			innerBattleID := strings.Split(key.(string), "_")[0]
			// check inner battle id is the same
			if innerBattleID != battleID {
				return true
			}

			uidStr := strings.Split(key.(string), "_")[1]
			currentValue := value.(*MultiplierAction)
			if currentValue.Expiry.Before(now) {
				return true
			}
			// get data from check map
			_, ok := um.CheckMaps.KillMap.Load(battleID + "_" + uidStr)
			// record, if not exists
			if !ok {
				// store different
				d, ok := diff[uidStr]
				if !ok {
					d = make(map[string]*MultiplierAction)
				}
				d[string(BattleRewardTypeKill)+"_"+battleID] = currentValue
				diff[uidStr] = d
				// update check map
				um.CheckMaps.KillMap.Store(battleID+"_"+uidStr, currentValue)

				return true
			}

			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(BattleRewardTypeKill)+"_"+battleID] = currentValue
			diff[uidStr] = d
			// update check map
			um.CheckMaps.KillMap.Store(battleID+"_"+uidStr, currentValue)

			return true
		})

		return true
	})

	for userID, ma := range diff {
		// update user remain rate
		remainRate := um.UserRemainRate(now, userID)
		if remainRate == 0 || remainRate == 100 {
			continue
		}
		for _, m := range ma {
			m.MultiplierValue = m.MultiplierValue * remainRate / 100
		}
		uid := server.UserID(uuid.FromStringOrNil(userID))
		go um.UserSupsMultiplierToPassport(uid, ma)
	}
}

func (um *UserMultiplier) UserActiveChecker() {
	now := time.Now()
	um.ActiveMap.Range(func(key, value interface{}) bool {
		userIDstr := key.(string)
		lastTime, ok := value.(time.Time)
		if !ok {
			um.ActiveMap.Delete(userIDstr)
		}
		if now.Sub(lastTime).Minutes() >= 20 {
			// remove from active map
			um.ActiveMap.Delete(userIDstr)
			return true
		}
		return true
	})
}

func (um *UserMultiplier) UserRemainRate(now time.Time, userID string) int {
	value, ok := um.ActiveMap.Load(userID)
	if !ok {
		return 0
	}

	lastValue, ok := value.(time.Time)
	if !ok {
		return 0
	}

	lastMinute := int(now.Sub(lastValue).Minutes())

	if lastMinute >= 30 {
		return 0
	}

	if lastMinute <= 10 {
		return 100
	}

	remainRate := lastMinute - 10

	return 100 - (remainRate/2)*10
}
