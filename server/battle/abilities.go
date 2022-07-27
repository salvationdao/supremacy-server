package battle

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"server"
	"server/benchmark"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sort"
	"sync"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/terror/v2"

	"github.com/volatiletech/null/v8"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
)

type AbilityRadius int

const (
	NukeRadius     AbilityRadius = 5200
	BlackoutRadius AbilityRadius = 20000
)

type AbilitiesSystem struct {
	// faction unique abilities
	_battle           *Battle
	startedAt         time.Time
	BattleAbilityPool *AbilityPool
	isClosed          atomic.Bool

	locationSelectChan chan *locationSelect

	sync.RWMutex
}

type locationSelect struct {
	userID     string
	factionID  string
	startPoint server.CellLocation
	endPoint   *server.CellLocation
}

type optIn struct {
	userID    string
	factionID string
}

func (as *AbilitiesSystem) battle() (*Battle, bool) {
	as.RLock()
	defer as.RUnlock()

	return as._battle, as._battle != nil
}

func (as *AbilitiesSystem) storeBattle(btl *Battle) {
	as.Lock()
	defer as.Unlock()
	as._battle = btl
}

func AbilitySystemIsAvailable(as *AbilitiesSystem) bool {
	if as == nil {
		gamelog.L.Debug().Msg("ability system is nil")
		return false
	}

	as.RLock()
	defer as.RUnlock()

	// no battle instance
	if as._battle == nil {
		gamelog.L.Debug().Msg("battle is nil")
		return false
	}

	// battle ended
	if as._battle.stage.Load() == BattleStageEnd {
		gamelog.L.Debug().Msg("battle ended")
		return false
	}

	// no current battle
	if as._battle.arena._currentBattle == nil {
		gamelog.L.Debug().Msg("current battle is nil")
		return false
	}

	// battle mismatch
	if as._battle != as._battle.arena._currentBattle {
		gamelog.L.Debug().Msg("battle not match")
		return false
	}

	return true
}

func (as *AbilitiesSystem) broadcastLocationSelectNotification(data *GameNotificationLocationSelect) {
	if as._battle != nil {
		as._battle.arena.BroadcastGameNotificationLocationSelect(data)
	}
}

type AbilityPool struct {
	Stage            *GabsBribeStage
	midpoint         *MidPoint // track location select
	BattleAbility    *BattleAbility
	LocationDeciders *LocationDeciders
	config           *AbilityConfig
	sync.RWMutex
}

type MidPoint struct {
	interruptedAt null.Time
	sync.Mutex
}

func (mp *MidPoint) store(t time.Time) {
	mp.Lock()
	defer mp.Unlock()

	mp.interruptedAt = null.TimeFrom(t)
}

func (mp *MidPoint) clear() {
	mp.Lock()
	defer mp.Unlock()

	mp.interruptedAt = null.TimeFromPtr(nil)
}

func (mp *MidPoint) load() (time.Time, bool) {
	mp.Lock()
	defer mp.Unlock()

	return mp.interruptedAt.Time, mp.interruptedAt.Valid
}

type BattleAbility struct {
	*boiler.BattleAbility
	OfferingID string
	sync.RWMutex
}

func (ba *BattleAbility) store(battleAbility *boiler.BattleAbility) {
	ba.Lock()
	defer ba.Unlock()

	ba.BattleAbility = battleAbility
	ba.OfferingID = uuid.Must(uuid.NewV4()).String()
}

func (ba *BattleAbility) LoadOfferingID() string {
	ba.Lock()
	defer ba.Unlock()

	return ba.OfferingID
}

func (ba *BattleAbility) LoadBattleAbility() *boiler.BattleAbility {
	ba.Lock()
	defer ba.Unlock()

	return ba.BattleAbility
}

type LocationDeciders struct {
	m               map[string][]string
	currentDeciders map[string]string
	sync.RWMutex
}

func (ld *LocationDeciders) clear() {
	ld.Lock()
	defer ld.Unlock()

	for key := range ld.m {
		ld.m[key] = []string{}
		ld.currentDeciders[key] = ""
	}
}

func (ld *LocationDeciders) currentDecider(factionID string) string {
	ld.RLock()
	defer ld.RUnlock()

	id, ok := ld.currentDeciders[factionID]
	if !ok {
		return ""
	}
	return id
}

func (ld *LocationDeciders) currentDeciderClear(factionID string) {
	ld.RLock()
	defer ld.RUnlock()

	ld.currentDeciders[factionID] = ""
}

func (ld *LocationDeciders) maxSelectorAmount() int {
	ld.RLock()
	defer ld.RUnlock()

	if len(ld.m) == 0 {
		return 0
	}

	lengths := []int{}
	for _, m := range ld.m {
		lengths = append(lengths, len(m))
	}

	sort.Slice(lengths, func(i, j int) bool {
		return lengths[i] > lengths[j]
	})

	return lengths[0]
}

func (ld *LocationDeciders) store(factionID string, userID string) {
	ld.Lock()
	defer ld.Unlock()

	if _, ok := ld.m[factionID]; !ok {
		return
	}

	ld.m[factionID] = append(ld.m[factionID], userID)
}

func (ld *LocationDeciders) length(factionID string) int {
	ld.RLock()
	defer ld.RUnlock()
	if li, ok := ld.m[factionID]; ok {
		return len(li)
	}

	return 0
}

func (ld *LocationDeciders) pop(factionID string) {
	ld.Lock()
	defer ld.Unlock()
	ids, ok := ld.m[factionID]
	if !ok || len(ids) == 0 {
		return
	}

	if len(ids) == 1 {
		ld.m[factionID] = []string{}
		return
	}

	ld.m[factionID] = ids[1:]
}

func (ld *LocationDeciders) first(factionID string) string {
	ld.Lock()
	defer ld.Unlock()

	// update current decider as well
	ld.currentDeciders[factionID] = ""

	if ids, ok := ld.m[factionID]; ok && len(ids) > 0 {
		ld.currentDeciders[factionID] = ids[0]
		return ids[0]
	}

	return ""
}

type AbilityConfig struct {
	BattleAbilityOptInDuration          time.Duration
	BattleAbilityLocationSelectDuration time.Duration
	AdvanceAbilityShowUpUntilSeconds    int
	AdvanceAbilityLabel                 string
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	// initialise new gabs ability pool
	as := &AbilitiesSystem{
		_battle:   battle,
		startedAt: time.Now(),
		BattleAbilityPool: &AbilityPool{
			Stage: &GabsBribeStage{
				Phase:   atomic.NewInt32(BribeStageHold),
				endTime: time.Now().AddDate(1, 0, 0), // HACK: set end time to far future to implement infinite time
			},
			midpoint: &MidPoint{
				interruptedAt: null.TimeFromPtr(nil),
			},
			BattleAbility: &BattleAbility{},
			LocationDeciders: &LocationDeciders{
				m:               make(map[string][]string),
				currentDeciders: make(map[string]string),
			},
			config: &AbilityConfig{
				BattleAbilityOptInDuration:          time.Duration(db.GetIntWithDefault(db.KeyBattleAbilityBribeDuration, 5)) * time.Second,
				BattleAbilityLocationSelectDuration: time.Duration(db.GetIntWithDefault(db.KeyBattleAbilityLocationSelectDuration, 15)) * time.Second,
				AdvanceAbilityShowUpUntilSeconds:    db.GetIntWithDefault(db.KeyAdvanceBattleAbilityShowUpUntilSeconds, 300),
				AdvanceAbilityLabel:                 db.GetStrWithDefault(db.KeyAdvanceBattleAbilityLabel, "NUKE"),
			},
		},
		isClosed:           atomic.Bool{},
		locationSelectChan: make(chan *locationSelect),
	}
	as.isClosed.Store(false)

	btl, ok := as.battle()
	if ok {
		for factionID := range btl.factions {
			as.BattleAbilityPool.LocationDeciders.m[factionID.String()] = []string{}
			as.BattleAbilityPool.LocationDeciders.currentDeciders[factionID.String()] = ""
		}
	}

	// init battle ability
	_, err := as.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set up battle ability")
		return nil
	}

	// bribe cycle
	go as.StartGabsAbilityPoolCycle(false)

	return as
}

// SetNewBattleAbility set new battle ability and return the cooldown time
func (as *AbilitiesSystem) SetNewBattleAbility() (int, error) {
	if !AbilitySystemIsAvailable(as) {
		return 30, terror.Error(fmt.Errorf("ability system is closed"), "Ability system is closed.")
	}

	// offering id
	offeringID := as.BattleAbilityPool.BattleAbility.LoadOfferingID()

	go func(offeringID string) {
		if offeringID == "" {
			return
		}
		bao, err := boiler.BattleAbilityOptInLogs(
			boiler.BattleAbilityOptInLogWhere.BattleAbilityOfferingID.EQ(offeringID),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("offering id", offeringID).Err(err).Msg("Failed to get battle ability opt in logs")
			return
		}

		for _, ba := range bao {
			ws.PublishMessage(fmt.Sprintf("/user/%s/battle_ability/check_opt_in", ba.PlayerID), HubKeyBattleAbilityOptInCheck, false)
		}

	}(offeringID)

	as.BattleAbilityPool.Lock()
	defer as.BattleAbilityPool.Unlock()

	excludedAbility := as.BattleAbilityPool.config.AdvanceAbilityLabel
	if int(time.Now().Sub(as.startedAt).Seconds()) > as.BattleAbilityPool.config.AdvanceAbilityShowUpUntilSeconds {
		// bring in advance battle ability
		excludedAbility = ""
	}

	// initialise new gabs ability pool
	ba, err := db.BattleAbilityGetRandom(excludedAbility)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get battle ability from db")
		return 30, err
	}

	as.BattleAbilityPool.BattleAbility.store(ba)

	// broadcast battle ability to non-login or non-faction players
	ga, err := boiler.GameAbilities(
		boiler.GameAbilityWhere.BattleAbilityID.EQ(null.StringFrom(ba.ID)),
		boiler.GameAbilityWhere.FactionID.EQ(server.RedMountainFactionID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle ability id", ba.ID).Str("faction id", server.RedMountainFactionID).Msg("failed to retrieve faction battle ability")
		return ba.CooldownDurationSecond, err
	}

	ws.PublishMessage("/public/battle_ability", HubKeyBattleAbilityUpdated, GameAbility{
		ID:                     ga.ID,
		GameClientAbilityID:    byte(ga.GameClientAbilityID),
		ImageUrl:               ga.ImageURL,
		Description:            ga.Description,
		FactionID:              ga.FactionID,
		Label:                  ga.Label,
		Colour:                 ga.Colour,
		TextColour:             ga.TextColour,
		CooldownDurationSecond: ba.CooldownDurationSecond,
	})

	return ba.CooldownDurationSecond, nil
}

// ***************************
// Gabs Abilities Voting Cycle
// ***************************

const (
	BribeStageHold           int32 = 0
	BribeStageOptIn          int32 = 1
	BribeStageLocationSelect int32 = 2
	BribeStageCooldown       int32 = 3
)

var BribeStages = [4]string{
	"HOLD",
	"OPT_IN",
	"LOCATION_SELECT",
	"COOLDOWN",
}

type GabsBribeStage struct {
	Phase   *atomic.Int32 `json:"phase"`
	endTime time.Time     `json:"end_time"`
	sync.RWMutex
}

func (p *GabsBribeStage) EndTime() time.Time {
	p.RLock()
	defer p.RUnlock()
	return p.endTime
}

func (p *GabsBribeStage) StoreEndTime(t time.Time) {
	p.Lock()
	defer p.Unlock()

	p.endTime = t
}

func (p *GabsBribeStage) Normalise() *GabsBribeStageNormalised {
	return &GabsBribeStageNormalised{
		Phase:   BribeStages[p.Phase.Load()],
		EndTime: p.endTime,
	}
}

type GabsBribeStageNormalised struct {
	Phase   string    `json:"phase"`
	EndTime time.Time `json:"end_time"`
}

func (p *GabsBribeStage) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Normalise())
}

func (as *AbilitiesSystem) StartGabsAbilityPoolCycle(resume bool) {
	// start voting stage
	if !resume {
		as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)

		cooldownSeconds := 20
		if ab := as.BattleAbilityPool.BattleAbility.LoadBattleAbility(); ab != nil {
			cooldownSeconds = ab.CooldownDurationSecond
		}

		as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSeconds) * time.Second))
		ws.PublishMessage("/public/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
	}

	// initial a ticker for current battle
	mainTicker := time.NewTicker(1 * time.Second)

	// stop ticker, when func close
	defer func(as *AbilitiesSystem, mainTicker *time.Ticker) {
		mainTicker.Stop()
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic at the StartGabsAbilityPoolCycle!", r)

			if AbilitySystemIsAvailable(as) {
				as.StartGabsAbilityPoolCycle(true)
			}
		}
	}(as, mainTicker)

	// start ability pool cycle
	for {
		select {
		// wait for next tick
		case <-mainTicker.C:
			if as.isClosed.Load() {

				return
			}

			if !AbilitySystemIsAvailable(as) {
				// fire close
				as.isClosed.Store(true)
				continue
			}
			now := time.Now()

			// skip, if the end time of current phase haven't been reached
			if as.BattleAbilityPool.Stage.EndTime().After(now) {
				if it, ok := as.BattleAbilityPool.midpoint.load(); !ok || it.After(now) {
					continue
				}
			}

			switch as.BattleAbilityPool.Stage.Phase.Load() {

			// return early if it is in hold stage
			case BribeStageHold:
				continue

			// at the end of opt in phase
			// no ability is triggered, switch to cooldown phase
			case BribeStageOptIn:
				if !AbilitySystemIsAvailable(as) {
					// fire close
					continue
				}

				bm := benchmark.New()
				bm.Start("location select deciders")
				as.locationDecidersSet()
				bm.End("location select deciders")
				bm.Alert(100)

				maxTargetingRound := as.BattleAbilityPool.LocationDeciders.maxSelectorAmount()

				// get another ability if no one opt in
				if maxTargetingRound == 0 {
					// set new battle ability
					cooldownSecond, err := as.SetNewBattleAbility()
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
					}

					as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
					as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
					// broadcast stage to frontend
					ws.PublishMessage("/public/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
					continue
				}

				// broadcast the announcement to the next location decider
				ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
				// announce winner
				gas, err := boiler.GameAbilities(
					boiler.GameAbilityWhere.BattleAbilityID.EQ(null.StringFrom(ba.ID)),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("battle ability", ba.Label).Err(err).Msg("Failed to load game ability for notification")
				}

				newMidPoint := now.Add(as.BattleAbilityPool.config.BattleAbilityLocationSelectDuration)
				// assign ability user
				for _, ga := range gas {
					userID := as.BattleAbilityPool.LocationDeciders.first(ga.FactionID)
					if userID != "" {
						ws.PublishMessage(fmt.Sprintf("/user/%s", userID), HubKeyBribingWinnerSubscribe, struct {
							GameAbility *boiler.GameAbility `json:"game_ability"`
							EndTime     time.Time           `json:"end_time"`
						}{
							GameAbility: ga,
							EndTime:     newMidPoint,
						})

						// broadcast faction notification
					}
				}

				// set midpoint
				as.BattleAbilityPool.midpoint.store(newMidPoint)

				// change stage
				as.BattleAbilityPool.Stage.Phase.Store(BribeStageLocationSelect)
				as.BattleAbilityPool.Stage.StoreEndTime(now.Add(time.Duration(maxTargetingRound) * as.BattleAbilityPool.config.BattleAbilityLocationSelectDuration))
				// broadcast stage to frontend
				ws.PublishMessage("/public/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

			// at the end of location select phase
			// pass the location select to next player
			case BribeStageLocationSelect:
				if !AbilitySystemIsAvailable(as) {
					continue
				}

				// pop out the first decider
				for factionID := range as.BattleAbilityPool.LocationDeciders.currentDeciders {
					as.BattleAbilityPool.LocationDeciders.pop(factionID)
				}

				// if no selector in the pool, get into cool down stage
				if as.BattleAbilityPool.LocationDeciders.maxSelectorAmount() == 0 {
					as.BattleAbilityPool.midpoint.clear()

					// set new battle ability
					cooldownSecond, err := as.SetNewBattleAbility()
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
					}

					as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
					as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
					// broadcast stage to frontend
					ws.PublishMessage("/public/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
					continue
				}

				newMidPoint := now.Add(as.BattleAbilityPool.config.BattleAbilityLocationSelectDuration)

				as.BattleAbilityPool.midpoint.store(newMidPoint)

				ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
				// get game ability
				gas, err := boiler.GameAbilities(
					boiler.GameAbilityWhere.BattleAbilityID.EQ(null.StringFrom(ba.ID)),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("battle ability", ba.Label).Err(err).Msg("Failed to load game ability for notification")
				}

				// assign ability user
				wg := sync.WaitGroup{}
				for _, ga := range gas {
					wg.Add(1)
					go func(ga *boiler.GameAbility) {
						lastUserID := as.BattleAbilityPool.LocationDeciders.currentDecider(ga.FactionID)
						if lastUserID != "" {
							// send failed select notification
							notification := &GameNotificationLocationSelect{
								Type: LocationSelectTypeFailedTimeout,
								Ability: &AbilityBrief{
									Label:    gas[0].Label,
									ImageUrl: gas[0].ImageURL,
									Colour:   gas[0].Colour,
								},
							}

							// get current player
							currentPlayer, err := BuildUserDetailWithFaction(uuid.FromStringOrNil(lastUserID))
							if err == nil {
								notification.CurrentUser = currentPlayer
							}

							go as.broadcastLocationSelectNotification(notification)

							// get next location decider
							userID := as.BattleAbilityPool.LocationDeciders.first(ga.FactionID)
							if userID != "" {
								// get next player
								nextPlayer, err := BuildUserDetailWithFaction(uuid.FromStringOrNil(userID))
								if err == nil {
									notification.NextUser = nextPlayer
								}

								ws.PublishMessage(fmt.Sprintf("/user/%s", userID), HubKeyBribingWinnerSubscribe, struct {
									GameAbility *boiler.GameAbility `json:"game_ability"`
									EndTime     time.Time           `json:"end_time"`
								}{
									GameAbility: ga,
									EndTime:     newMidPoint,
								})

								// broadcast faction notification
							}
						}
						wg.Done()
					}(ga)
				}
				wg.Wait()

			// at the end of cooldown phase
			// random choose a battle ability for next bribing session
			case BribeStageCooldown:

				// change bribing phase
				as.BattleAbilityPool.Stage.Phase.Store(BribeStageOptIn)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(as.BattleAbilityPool.config.BattleAbilityOptInDuration))
				// broadcast stage to frontend
				ws.PublishMessage("/public/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

				continue
			default:
				gamelog.L.Error().Str("log_name", "battle arena").Msg("hit default case switch on abilities loop")
			}

		// location select chan
		case ls := <-as.locationSelectChan:
			// check battle end
			if !AbilitySystemIsAvailable(as) {
				gamelog.L.Warn().Str("func", "LocationSelect").Msg("Ability system is not available.")
				continue
			}

			btl, ok := as.battle()
			if !ok {
				continue
			}

			// check eligibility
			if as.BattleAbilityPool.LocationDeciders.currentDecider(ls.factionID) != ls.userID {
				continue
			}

			offeringID := uuid.Must(uuid.NewV4())
			userUUID := uuid.FromStringOrNil(ls.userID)

			// clear up current faction location decider
			as.BattleAbilityPool.LocationDeciders.currentDeciderClear(ls.factionID)

			// start ability trigger process
			ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
			ga, err := boiler.GameAbilities(
				boiler.GameAbilityWhere.BattleAbilityID.EQ(null.StringFrom(ba.ID)),
			).One(gamedb.StdConn)
			if err != nil {
				continue
			}

			// get player detail
			player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(ls.userID)).One(gamedb.StdConn)
			if err != nil {
				continue
			}

			faction, err := boiler.Factions(boiler.FactionWhere.ID.EQ(ls.factionID)).One(gamedb.StdConn)
			if err != nil {
				continue
			}

			event := &server.GameAbilityEvent{
				IsTriggered:         true,
				GameClientAbilityID: byte(ga.GameClientAbilityID),
				TriggeredByUserID:   &userUUID,
				TriggeredByUsername: &player.Username.String,
				EventID:             offeringID,
				FactionID:           &ls.factionID,
			}

			event.GameLocation = btl.getGameWorldCoordinatesFromCellXY(&ls.startPoint)

			if ga.LocationSelectType == boiler.LocationSelectTypeEnumLINE_SELECT && ls.endPoint != nil {
				event.GameLocationEnd = btl.getGameWorldCoordinatesFromCellXY(ls.endPoint)
			}

			// trigger location select
			btl.arena.Message("BATTLE:ABILITY", event)

			bat := boiler.BattleAbilityTrigger{
				PlayerID:          null.StringFrom(ls.userID),
				BattleID:          btl.ID,
				FactionID:         ga.FactionID,
				IsAllSyndicates:   true,
				AbilityLabel:      ga.Label,
				GameAbilityID:     ga.ID,
				AbilityOfferingID: offeringID.String(),
			}
			err = bat.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Interface("battle_ability_trigger", bat).Err(err).Msg("Failed to record ability triggered")
			}

			_, err = db.UserStatAddTotalAbilityTriggered(ls.userID)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", ls.userID).Err(err).Msg("failed to update user ability triggered amount")
			}

			btl.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
				Type: LocationSelectTypeTrigger,
				Ability: &AbilityBrief{
					Label:    ga.Label,
					ImageUrl: ga.ImageURL,
					Colour:   ga.Colour,
				},
				CurrentUser: &UserBrief{
					ID:        userUUID,
					Username:  player.Username.String,
					FactionID: player.FactionID.String,
					Gid:       player.Gid,
					Faction: &Faction{
						ID:    faction.ID,
						Label: faction.Label,
						Theme: &Theme{
							PrimaryColor:    faction.PrimaryColor,
							SecondaryColor:  faction.SecondaryColor,
							BackgroundColor: faction.BackgroundColor,
						},
					},
				},
			})
		}
	}
}

// locationDecidersSet set a user list for location select for current ability triggered
func (as *AbilitiesSystem) locationDecidersSet() {
	// clear current location deciders
	as.BattleAbilityPool.LocationDeciders.clear()

	if !AbilitySystemIsAvailable(as) {
		return
	}

	// check ability is advance
	locationSelectorAmount := 2
	if as.BattleAbilityPool.BattleAbility.Label == as.BattleAbilityPool.config.AdvanceAbilityLabel {
		// drop selector amount to one
		locationSelectorAmount = 1
	}

	offeringID := as.BattleAbilityPool.BattleAbility.LoadOfferingID()

	aps, err := boiler.PlayerActiveLogs(
		qm.Select(boiler.PlayerActiveLogColumns.PlayerID),
		boiler.PlayerActiveLogWhere.InactiveAt.IsNull(),
		boiler.PlayerActiveLogWhere.ActiveAt.GT(time.Now().AddDate(0, 0, -1)),
		qm.GroupBy(boiler.PlayerActiveLogColumns.PlayerID),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load active player list")
	}

	// load ability log
	bao, err := boiler.BattleAbilityOptInLogs(
		boiler.BattleAbilityOptInLogWhere.BattleAbilityOfferingID.EQ(offeringID),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle ability opt list")
		return
	}

	// get location select limited players
	pbs, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get limited select players from db")
	}

	// shuffle players
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(bao), func(i, j int) { bao[i], bao[j] = bao[j], bao[i] })

	wg := sync.WaitGroup{}
	for factionID := range as.BattleAbilityPool.LocationDeciders.m {
		wg.Add(1)
		go func(factionID string) {
			for _, ba := range bao {
				if ba.FactionID != factionID {
					continue
				}

				// skip, if player is not connected
				if !ws.IsConnected(ba.PlayerID) {
					continue
				}

				// check user is banned
				isBanned := false
				for _, pb := range pbs {
					if pb.BannedPlayerID == ba.PlayerID {
						isBanned = true
						break
					}
				}

				// skip, if player is banned
				if isBanned {
					continue
				}

				// check player is active (no mouse movement)
				exist := false
				for _, ap := range aps {
					if ap.PlayerID == ba.PlayerID {
						exist = true
						break
					}
				}

				// skip, if player does not exist
				if !exist {
					continue
				}

				fmt.Println("player still active", ba.PlayerID)

				if as.BattleAbilityPool.LocationDeciders.length(ba.FactionID) >= locationSelectorAmount {
					continue
				}

				// set location decider list
				as.BattleAbilityPool.LocationDeciders.store(ba.FactionID, ba.PlayerID)
			}

			wg.Done()
		}(factionID)
	}
	// wait until all process done
	wg.Wait()

	return
}

func (as *AbilitiesSystem) BribeStageGet() *GabsBribeStageNormalised {
	if as.BattleAbilityPool != nil {
		return as.BattleAbilityPool.Stage.Normalise()
	}
	return nil
}

func (as *AbilitiesSystem) LocationSelect(userID string, factionID string, startPoint server.CellLocation, endPoint *server.CellLocation) error {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the LocationSelect!", r)
		}
	}()
	// check battle end
	if !AbilitySystemIsAvailable(as) {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("Ability system is not available.")
		return nil
	}

	// check eligibility
	if as.BattleAbilityPool.LocationDeciders.currentDecider(factionID) != userID {
		return terror.Error(terror.ErrForbidden, "Not eligible to target location.")
	}

	as.locationSelectChan <- &locationSelect{userID, factionID, startPoint, endPoint}

	return nil
}

func (as *AbilitiesSystem) End() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic at the abilities.End!", r)
		}
	}()

	as.BattleAbilityPool.Stage.Phase.Store(BribeStageHold)
	as.BattleAbilityPool.Stage.StoreEndTime(time.Now().AddDate(1, 0, 0))
	// broadcast stage to frontend
	ws.PublishMessage("/public/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
	as.isClosed.Store(true)

	as.storeBattle(nil)
}

func BuildUserDetailWithFaction(userID uuid.UUID) (*UserBrief, error) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the BuildUserDetailWithFaction!", r)
		}
	}()
	userBrief := &UserBrief{}

	user, err := boiler.FindPlayer(gamedb.StdConn, userID.String())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", userID.String()).Err(err).Msg("failed to get player from db")
		return nil, err
	}

	userBrief.ID = userID
	userBrief.Username = user.Username.String
	userBrief.Gid = user.Gid

	if !user.FactionID.Valid {
		return userBrief, nil
	}

	userBrief.FactionID = user.FactionID.String

	faction, err := boiler.Factions(boiler.FactionWhere.ID.EQ(user.FactionID.String)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", userID.String()).Str("faction_id", user.FactionID.String).Err(err).Msg("failed to get player faction from db")
		return userBrief, nil
	}

	userBrief.Faction = &Faction{
		ID:    faction.ID,
		Label: faction.Label,
		Theme: &Theme{
			PrimaryColor:    faction.PrimaryColor,
			SecondaryColor:  faction.SecondaryColor,
			BackgroundColor: faction.BackgroundColor,
		},
	}

	return userBrief, nil
}
