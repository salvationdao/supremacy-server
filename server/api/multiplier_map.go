package api

import (
	"context"
	"errors"
	"fmt"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/sasha-s/go-deadlock"
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

	BattleIDMap deadlock.Map

	// other dependencies
	UserMap     *UserMap
	Passport    *passport.Passport
	BattleArena *battle_arena.BattleArena

	ActiveMap *deadlock.Map

	// ability triggered map
	NukeAbility            *AbilityTrigger
	AirstrikeAbility       *AbilityTrigger
	RepairAbility          *AbilityTrigger
	AbilityFactionRecorder []server.FactionID

	// Citizen
}

type Multiplier struct {
	ApplauseMap       deadlock.Map
	PickedLocationMap deadlock.Map

	// battle multiplier
	WinningFactionMap deadlock.Map
	WinningUserMap    deadlock.Map
	KillMap           deadlock.Map

	// Ability Reward
	NukeRewardMap      *AbilityTriggerMap
	AirstrikeRewardMap *AbilityTriggerMap
	RepairRewardMap    *AbilityTriggerMap

	// Combo Breaker
	ComboBreakerMap *ComboBreakerMap

	// citizen
	ActiveCitizenMap *CitizenMap
}

type MultiplierAction struct {
	MultiplierValue int
	Expiry          time.Time
}

// TODO: set up sups ticker
func NewUserMultiplier(userMap *UserMap, pp *passport.Passport, ba *battle_arena.BattleArena) *UserMultiplier {
	um := &UserMultiplier{
		CurrentMaps: &Multiplier{deadlock.Map{}, deadlock.Map{}, deadlock.Map{}, deadlock.Map{}, deadlock.Map{}, &AbilityTriggerMap{}, &AbilityTriggerMap{}, &AbilityTriggerMap{}, &ComboBreakerMap{}, &CitizenMap{}},
		BattleIDMap: deadlock.Map{},
		UserMap:     userMap,
		Passport:    pp,
		BattleArena: ba,

		ActiveMap: &deadlock.Map{},

		NukeAbility: &AbilityTrigger{
			[]server.UserID{},
			deadlock.RWMutex{},
		},
		AirstrikeAbility: &AbilityTrigger{
			[]server.UserID{},
			deadlock.RWMutex{},
		},
		RepairAbility: &AbilityTrigger{
			[]server.UserID{},
			deadlock.RWMutex{},
		},
		AbilityFactionRecorder: []server.FactionID{},
	}

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if ba.BattleActive() {
				// distribute sups
				um.SupsTick()
			}
		}
	}()

	go func() {
		for {
			// check user active list
			um.UserActiveChecker()
			um.UserMultiplierUpdate()
			time.Sleep(10 * time.Second)
		}
	}()

	return um
}

// Online handle user online multiplier
func (um *UserMultiplier) Online(userID server.UserID) {
	userIDStr := userID.String()
	now := time.Now()
	um.ActiveMap.Store(userIDStr, now)

	// load multipliers from db
	sm, err := db.UserMultiplierGet(context.Background(), um.BattleArena.Conn, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		fmt.Println("Failed to read user multipliers from db", err.Error())
		return
	}

	for _, s := range sm {
		if CitizenTag(s.Key).IsCitizen() {
			continue
		}
		switch s.Key {
		case string(ClientVoted):
			if s.ExpiredAt.Before(now) {
				continue
			}
			um.CurrentMaps.ApplauseMap.Store(userIDStr, &MultiplierAction{
				MultiplierValue: s.Value,
				Expiry:          s.ExpiredAt,
			})
		case string(ClientPickedLocation):
			if s.ExpiredAt.Before(now) {
				continue
			}
			um.CurrentMaps.PickedLocationMap.Store(userIDStr, &MultiplierAction{
				MultiplierValue: s.Value,
				Expiry:          s.ExpiredAt,
			})
		default:
			strs := strings.Split(s.Key, "_")
			if len(strs) < 2 {
				continue
			}

			if len(strs) == 3 { // user id, title, time
				userID := strs[0]
				title := strs[1]
				timestp := strs[2]
				switch strs[1] {
				case "Air Marshal", "Air Support":
					um.CurrentMaps.AirstrikeRewardMap.Store(fmt.Sprintf("%s_%s_%s", title, userID, timestp), &MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					})
				case "Now I am become Death", "Destroyer of worlds":
					um.CurrentMaps.NukeRewardMap.Store(fmt.Sprintf("%s_%s_%s", title, userID, timestp), &MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					})
				case "Grease Monkey", "Field Mechanic":
					um.CurrentMaps.RepairRewardMap.Store(fmt.Sprintf("%s_%s_%s", title, userID, timestp), &MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					})
				case "Combo breaker":
					um.CurrentMaps.ComboBreakerMap.Store(fmt.Sprintf("%s_%s_%s", title, userID, timestp), &MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					})
				}
			}

			brk := strs[0]
			battleID := strs[1]

			switch brk {
			case string(BattleRewardTypeFaction):
				if s.ExpiredAt.Before(now) {
					continue
				}
				um.CurrentMaps.WinningFactionMap.Store(
					battleID+"_"+userIDStr,
					&MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					},
				)
			case string(BattleRewardTypeWinner):
				if s.ExpiredAt.Before(now) {
					continue
				}
				um.CurrentMaps.WinningUserMap.Store(
					battleID+"_"+userIDStr,
					&MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					},
				)
			case string(BattleRewardTypeKill):
				if s.ExpiredAt.Before(now) {
					continue
				}
				um.CurrentMaps.KillMap.Store(
					battleID+"_"+userIDStr,
					&MultiplierAction{
						MultiplierValue: s.Value,
						Expiry:          s.ExpiredAt,
					},
				)
			}
		}
	}

	um.CurrentMaps.ActiveCitizenMap.Store(userIDStr+"_"+string(CitizenTagCitizen), &MultiplierAction{
		MultiplierValue: 100,
		Expiry:          time.Now().AddDate(1, 0, 0),
	})
}

// Offline remove all the user related multiplier action in current map
func (um *UserMultiplier) Offline(userID server.UserID) {
	userIDStr := userID.String()

	um.CurrentMaps.ApplauseMap.Delete(userIDStr)
	um.CurrentMaps.PickedLocationMap.Delete(userIDStr)

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
		um.CurrentMaps.WinningFactionMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CurrentMaps.WinningFactionMap.Delete(key)
			return true
		})
	}()

	go func() {
		um.CurrentMaps.WinningUserMap.Range(func(key, value interface{}) bool {
			if !strings.HasPrefix(key.(string), battleIDStr) {
				return true
			}
			um.CurrentMaps.WinningUserMap.Delete(key)
			return true
		})
	}()

	go func() {
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

	// ability
	um.CurrentMaps.NukeRewardMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.NukeRewardMap.Delete(key)
			return true
		}

		// append user to the ticking list
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		userID := server.UserID(uuid.FromStringOrNil(userStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userStr)
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

	// ability
	um.CurrentMaps.AirstrikeRewardMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.AirstrikeRewardMap.Delete(key)
			return true
		}

		// append user to the ticking list
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		userID := server.UserID(uuid.FromStringOrNil(userStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userStr)
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

	um.CurrentMaps.RepairRewardMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.RepairRewardMap.Delete(key)
			return true
		}

		// append user to the ticking list
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		userID := server.UserID(uuid.FromStringOrNil(userStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userStr)
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

	um.CurrentMaps.ComboBreakerMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.ComboBreakerMap.Delete(key)
			return true
		}

		// append user to the ticking list
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		userID := server.UserID(uuid.FromStringOrNil(userStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userStr)
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

	// citizen
	um.CurrentMaps.ActiveCitizenMap.Range(func(key, value interface{}) bool {
		m := value.(*MultiplierAction)
		// clean up, if expired
		if m.Expiry.Before(now) {
			um.CurrentMaps.ActiveCitizenMap.Delete(key)
			return true
		}

		// append user to the ticking list
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		userID := server.UserID(uuid.FromStringOrNil(userStr))

		// skip, if user is not active
		multiplierValue := m.MultiplierValue
		remain := um.UserRemainRate(now, userStr)
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

// UserMultiplierGet push the multiplier actions list of the user to passport user
func (um *UserMultiplier) UserMultiplierGet(userID server.UserID) []*server.SupsMultiplier {
	uidStr := userID.String()
	mas := make(map[string]*MultiplierAction)
	now := time.Now()

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

	result := []*server.SupsMultiplier{}
	remain := um.UserRemainRate(now, uidStr)
	for key, sm := range mas {
		m := sm.MultiplierValue
		m = m * remain / 100

		result = append(result, &server.SupsMultiplier{
			Key:       key,
			Value:     m,
			ExpiredAt: sm.Expiry,
		})
	}

	return result
}

func (um *UserMultiplier) UserSupsMultiplierToPassport(userID server.UserID, supsMultiplierMap map[string]*MultiplierAction, multiplier int) {
	userSupsMultiplierSend := &server.UserSupsMultiplierSend{
		ToUserID:        userID,
		SupsMultipliers: []*server.SupsMultiplier{},
	}

	for key, sm := range supsMultiplierMap {
		m := sm.MultiplierValue
		m = m * multiplier / 100

		userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &server.SupsMultiplier{
			Key:       key,
			Value:     m,
			ExpiredAt: sm.Expiry,
		})
	}

	go um.Passport.UserSupsMultiplierSend(context.Background(), []*server.UserSupsMultiplierSend{userSupsMultiplierSend})
}

func (um *UserMultiplier) UserMultiplierUpdate() {
	// map[userID]map[reward text] &MultiplierAction
	diff := make(map[string]map[string]*MultiplierAction)
	now := time.Now()

	um.CurrentMaps.ActiveCitizenMap.Range(func(key, value interface{}) bool {
		keys := strings.Split(key.(string), "_")
		userIDstr := keys[0]
		title := keys[1]
		currentValue := value.(*MultiplierAction)
		// store different
		d, ok := diff[userIDstr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}

		d[title] = currentValue
		diff[userIDstr] = d

		// update check map
		return true
	})

	// check current map with check map, add any different from the cache
	um.CurrentMaps.ApplauseMap.Range(func(key, value interface{}) bool {
		uidStr := key.(string)
		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			um.CurrentMaps.ApplauseMap.Delete(key)
			return true
		}

		// store different
		d, ok := diff[uidStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(ClientVoted)] = currentValue
		diff[uidStr] = d

		return true
	})

	// check current map with check map, add any different from the cache
	um.CurrentMaps.PickedLocationMap.Range(func(key, value interface{}) bool {
		uidStr := key.(string)
		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			um.CurrentMaps.PickedLocationMap.Delete(key)
			return true
		}
		// store different
		d, ok := diff[uidStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(ClientPickedLocation)] = currentValue
		diff[uidStr] = d
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
				um.CurrentMaps.WinningFactionMap.Delete(key)
				return true
			}
			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(BattleRewardTypeFaction)+"_"+battleID] = currentValue
			diff[uidStr] = d

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
				um.CurrentMaps.WinningUserMap.Delete(key)
				return true
			}
			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(BattleRewardTypeWinner)+"_"+battleID] = currentValue
			diff[uidStr] = d

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
				um.CurrentMaps.KillMap.Delete(key)
				return true
			}
			// store different
			d, ok := diff[uidStr]
			if !ok {
				d = make(map[string]*MultiplierAction)
			}
			d[string(BattleRewardTypeKill)+"_"+battleID] = currentValue
			diff[uidStr] = d

			return true
		})

		return true
	})

	// claim abilities
	um.CurrentMaps.AirstrikeRewardMap.Range(func(key, value interface{}) bool {
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		title := keys[1]
		timeStp := keys[2]

		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			um.CurrentMaps.AirstrikeRewardMap.Delete(key)
			return true
		}
		// store different
		d, ok := diff[userStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(title)+"_"+userStr+"_"+timeStp] = currentValue

		return true
	})

	um.CurrentMaps.NukeRewardMap.Range(func(key, value interface{}) bool {
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		title := keys[1]
		timeStp := keys[2]

		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			um.CurrentMaps.NukeRewardMap.Delete(key)
			return true
		}
		// store different
		d, ok := diff[userStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(title)+"_"+userStr+"_"+timeStp] = currentValue

		return true
	})

	um.CurrentMaps.RepairRewardMap.Range(func(key, value interface{}) bool {
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		title := keys[1]
		timeStp := keys[2]

		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			um.CurrentMaps.RepairRewardMap.Delete(key)
			return true
		}
		// store different
		d, ok := diff[userStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(title)+"_"+userStr+"_"+timeStp] = currentValue

		return true
	})

	// Combo breaker
	um.CurrentMaps.ComboBreakerMap.Range(func(key, value interface{}) bool {
		keys := strings.Split(key.(string), "_") // user id, title, timestamp
		userStr := keys[0]
		title := keys[1]
		timeStp := keys[2]

		currentValue := value.(*MultiplierAction)
		if currentValue.Expiry.Before(now) {
			um.CurrentMaps.ComboBreakerMap.Delete(key)
			return true
		}

		// store different
		d, ok := diff[userStr]
		if !ok {
			d = make(map[string]*MultiplierAction)
		}
		d[string(title)+"_"+userStr+"_"+timeStp] = currentValue

		return true
	})

	userSupsMultiplierSends := []*server.UserSupsMultiplierSend{}
	for userID, ma := range diff {
		// update user remain rate
		remainRate := um.UserRemainRate(now, userID)
		if remainRate == 0 {
			continue
		}

		uid := server.UserID(uuid.FromStringOrNil(userID))
		userSupsMultiplierSend := &server.UserSupsMultiplierSend{
			ToUserID:        uid,
			SupsMultipliers: []*server.SupsMultiplier{},
		}

		for key, sm := range ma {
			m := sm.MultiplierValue
			m = m * remainRate / 100

			userSupsMultiplierSend.SupsMultipliers = append(userSupsMultiplierSend.SupsMultipliers, &server.SupsMultiplier{
				Key:       key,
				Value:     m,
				ExpiredAt: sm.Expiry,
			})
		}

		userSupsMultiplierSends = append(userSupsMultiplierSends, userSupsMultiplierSend)

	}

	// broadcast to user
	go um.Passport.UserSupsMultiplierSend(context.Background(), userSupsMultiplierSends)

	// store in db
	err := db.UserMultiplierStore(context.Background(), um.BattleArena.Conn, userSupsMultiplierSends)
	if err != nil {
		um.BattleArena.Log.Err(err)
		return
	}
}

func (um *UserMultiplier) UserActiveChecker() {
	now := time.Now()
	um.ActiveMap.Range(func(key, value interface{}) bool {
		userIDstr := key.(string)
		lastTime, ok := value.(time.Time)
		if !ok {
			um.ActiveMap.Delete(userIDstr)
			return true
		}
		if now.Sub(lastTime).Minutes() >= 30 {
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

type AbilityTrigger struct {
	UserIDArray []server.UserID
	deadlock.RWMutex
}

type AbilityTriggerMap struct {
	deadlock.Map
}

func (atm *AbilityTriggerMap) Set(userID string, title string, isCombo bool) {
	value := 500
	expiredAt := time.Now().Add(1 * time.Minute)
	if isCombo {
		value = 1000
		expiredAt = time.Now().Add(30 * time.Minute)
	}
	atm.Store(fmt.Sprintf("%s_%s_%s", userID, title, time.Now().String()), &MultiplierAction{
		MultiplierValue: value,
		Expiry:          expiredAt,
	})
}

type ComboBreakerMap struct {
	deadlock.Map
}

func (cbm *ComboBreakerMap) Set(userID string) {
	cbm.Store(fmt.Sprintf("%s_Combo breaker_%s", userID, time.Now().String()), &MultiplierAction{
		MultiplierValue: 500,
		Expiry:          time.Now().Add(3 * time.Minute),
	})
}

func (um *UserMultiplier) AbilityTriggered(factionID server.FactionID, triggerUserID server.UserID, ability *server.GameAbility) {
	switch ability.Label {
	case "AIRSTRIKE":
		um.AirstrikeAbility.Lock()
		um.AirstrikeAbility.UserIDArray = append(um.AirstrikeAbility.UserIDArray, triggerUserID)

		// if the array is longer than three
		if len(um.AirstrikeAbility.UserIDArray) == 4 {
			um.AirstrikeAbility.UserIDArray = um.AirstrikeAbility.UserIDArray[1:]
		}

		isCombo := len(um.AirstrikeAbility.UserIDArray) == 3
		for _, userID := range um.AirstrikeAbility.UserIDArray {
			if userID != triggerUserID {
				isCombo = false
				break
			}
		}

		if isCombo {
			// set user as the winner of last three airstrike
			// tripple kill
			um.CurrentMaps.AirstrikeRewardMap.Set(triggerUserID.String(), "Air Marshal", true)
		} else {
			// set user as air support
			um.CurrentMaps.AirstrikeRewardMap.Set(triggerUserID.String(), "Air Support", false)
		}

		// combo breaker
		if um.IsComboBreaker(factionID) {
			um.CurrentMaps.ComboBreakerMap.Set(triggerUserID.String())
		}

		um.AirstrikeAbility.Unlock()
	case "NUKE":
		um.NukeAbility.Lock()
		um.NukeAbility.UserIDArray = append(um.NukeAbility.UserIDArray, triggerUserID)
		// if the array is longer than three
		if len(um.NukeAbility.UserIDArray) == 4 {
			um.NukeAbility.UserIDArray = um.NukeAbility.UserIDArray[1:]
		}

		isCombo := len(um.NukeAbility.UserIDArray) == 3
		for _, userID := range um.NukeAbility.UserIDArray {
			if userID != triggerUserID {
				isCombo = false
				break
			}
		}

		if isCombo {
			// set user as the winner of last three airstrike
			um.CurrentMaps.NukeRewardMap.Set(triggerUserID.String(), "Destroyer of worlds", true)
		} else {
			// set user as
			um.CurrentMaps.NukeRewardMap.Set(triggerUserID.String(), "Now I am become Death", false)
		}

		// combo breaker
		if um.IsComboBreaker(factionID) {
			um.CurrentMaps.ComboBreakerMap.Set(triggerUserID.String())
		}

		um.NukeAbility.Unlock()
	case "REPAIR":
		um.RepairAbility.Lock()
		um.RepairAbility.UserIDArray = append(um.RepairAbility.UserIDArray, triggerUserID)
		// if the array is longer than three
		if len(um.RepairAbility.UserIDArray) == 4 {
			um.RepairAbility.UserIDArray = um.RepairAbility.UserIDArray[1:]
		}

		isCombo := len(um.RepairAbility.UserIDArray) == 3
		for _, userID := range um.RepairAbility.UserIDArray {
			if userID != triggerUserID {
				isCombo = false
				break
			}
		}

		if isCombo {
			// set user as the winner of last three airstrike
			um.CurrentMaps.RepairRewardMap.Set(triggerUserID.String(), "Field Mechanic", true)
		} else {
			// set user as
			um.CurrentMaps.RepairRewardMap.Set(triggerUserID.String(), "Grease Monkey", false)
		}

		// combo breaker
		if um.IsComboBreaker(factionID) {
			um.CurrentMaps.ComboBreakerMap.Set(triggerUserID.String())
		}

		um.RepairAbility.Unlock()
	}
}

func (um *UserMultiplier) IsComboBreaker(triggerFactionID server.FactionID) bool {
	um.AbilityFactionRecorder = append(um.AbilityFactionRecorder, triggerFactionID)

	// check faction
	if len(um.AbilityFactionRecorder) == 5 {
		um.AbilityFactionRecorder = um.AbilityFactionRecorder[1:]
	}

	isComboBreaker := len(um.AbilityFactionRecorder) == 4

	for i, factionID := range um.AbilityFactionRecorder {
		// don't count the last one
		if i == 4 {
			break
		}

		// if faction triggere last three round
		if triggerFactionID == factionID {
			isComboBreaker = false
			break
		}
	}

	return isComboBreaker
}

type CitizenTag string

const (
	CitizenTagSuperContributor    CitizenTag = "Super Contributor"    // top 10%
	CitizenTagContributor         CitizenTag = "Contributor"          // top 25%
	CitizenTagSupporter           CitizenTag = "Supporter"            // top 50%
	CitizenTagCitizen             CitizenTag = "Citizen"              // top 80%
	CitizenTagUnproductiveCitizen CitizenTag = "Unproductive Citizen" // other 20%
)

func (e CitizenTag) IsCitizen() bool {
	switch e {
	case CitizenTagSuperContributor,
		CitizenTagContributor,
		CitizenTagSupporter,
		CitizenTagCitizen,
		CitizenTagUnproductiveCitizen:
		return true
	}

	return false
}

type CitizenMap struct {
	deadlock.Map
}

func (cm *CitizenMap) Clear() {
	cm.Range(func(key, value interface{}) bool {
		cm.Delete(key)
		return true
	})
}

// this map will not be stored in db
func (cm *CitizenMap) BulkSet(userIDs []*server.User, title string, ma *MultiplierAction) {
	for _, userID := range userIDs {
		key := userID.ID.String() + "_" + title
		cm.Store(key, ma)
	}
}

func (um *UserMultiplier) NewCitizenOrder(users []*server.User) {
	if len(users) == 0 {
		return
	}

	// clear the map
	um.CurrentMaps.ActiveCitizenMap.Clear()

	expiredAt := time.Now().AddDate(1, 0, 0)

	// calculate
	superContributors := []*server.User{}
	contributors := []*server.User{}
	supporters := []*server.User{}
	citizens := []*server.User{}

	// calc the top 10% amount
	superContributorAmount := len(users) / 10
	if superContributorAmount > 0 {
		superContributors = append(superContributors, users[:superContributorAmount]...)
		users = users[superContributorAmount:] // clear queue
	} else {
		superContributors = append(superContributors, users...)
		users = []*server.User{} // clear queue
	}
	um.CurrentMaps.ActiveCitizenMap.BulkSet(superContributors, "Super Contributor", &MultiplierAction{
		MultiplierValue: 1000,
		Expiry:          expiredAt,
	})

	// calc the top 25% amount
	contributorAmount := len(users)/4 - superContributorAmount
	if contributorAmount > 0 {
		contributors = append(contributors, users[:contributorAmount]...)
		users = users[contributorAmount:]
	} else {
		contributors = append(contributors, users...)
		users = []*server.User{} // clear queue
	}
	um.CurrentMaps.ActiveCitizenMap.BulkSet(contributors, "Contributor", &MultiplierAction{
		MultiplierValue: 500,
		Expiry:          expiredAt,
	})

	// calc the top 50% amount
	supporterAmount := len(users)/2 - superContributorAmount - contributorAmount
	if supporterAmount > 0 {
		supporters = append(supporters, users[:supporterAmount]...)
		users = users[supporterAmount:]
	} else {
		supporters = append(supporters, users...)
		users = []*server.User{} // clear queue
	}
	um.CurrentMaps.ActiveCitizenMap.BulkSet(supporters, "Supporter", &MultiplierAction{
		MultiplierValue: 250,
		Expiry:          expiredAt,
	})

	// calc the top 80% amount
	citizenAmount := len(users)*4/5 - superContributorAmount - contributorAmount - supporterAmount
	if citizenAmount > 0 {
		citizens = append(citizens, users[:citizenAmount]...)
		users = users[citizenAmount:]
	} else {
		citizens = append(citizens, users...)
		users = []*server.User{} // clear queue
	}
	um.CurrentMaps.ActiveCitizenMap.BulkSet(citizens, "Citizen", &MultiplierAction{
		MultiplierValue: 100,
		Expiry:          expiredAt,
	})

	// store in inactive citizen
	// calc the rest of 20% amount
	um.CurrentMaps.ActiveCitizenMap.BulkSet(users, "Unproductive Citizen", &MultiplierAction{
		MultiplierValue: 50,
		Expiry:          expiredAt,
	})
}
