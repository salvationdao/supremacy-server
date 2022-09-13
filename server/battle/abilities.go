package battle

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"server"
	"server/battle/player_abilities"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/terror/v2"

	"github.com/volatiletech/null/v8"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
)

type AbilityRadius int

const (
	BlackoutRadius AbilityRadius = player_abilities.BlackoutRadius
)

type AbilitiesSystem struct {
	arenaID string

	// faction unique abilities
	_battle           *Battle
	startedAt         time.Time
	BattleAbilityPool *AbilityPool
	isClosed          atomic.Bool

	locationSelectChan chan *locationSelect

	deadlock.RWMutex
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
	BattleAbility    *BattleAbility
	LocationDeciders *LocationDeciders
	config           *AbilityConfig
	deadlock.RWMutex
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
	deadlock.RWMutex
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
	currentDeciders map[string]bool
	deadlock.RWMutex
}

func (ld *LocationDeciders) clear() {
	ld.Lock()
	defer ld.Unlock()

	for key := range ld.m {
		ld.m[key] = []string{}
	}

	for key := range ld.currentDeciders {
		delete(ld.currentDeciders, key)
	}
}

func (ld *LocationDeciders) maxSelectorCount() int {
	ld.RLock()
	defer ld.RUnlock()

	return len(ld.currentDeciders)
}

func (ld *LocationDeciders) length(factionID string) int {
	ld.RLock()
	defer ld.RUnlock()
	if li, ok := ld.m[factionID]; ok {
		return len(li)
	}

	return 0
}

func (ld *LocationDeciders) setSelector(factionID string, playerID string) {
	ld.Lock()
	defer ld.Unlock()

	if _, ok := ld.m[factionID]; !ok {
		return
	}
	ld.m[factionID] = append(ld.m[factionID], playerID)

	ld.currentDeciders[playerID] = true
}

func (ld *LocationDeciders) rangeSelectors(fn func(playerID string)) {
	ld.RLock()
	defer ld.RUnlock()

	for pid, ok := range ld.currentDeciders {
		if !ok || pid == "" {
			continue
		}
		fn(pid)
	}
}

func (ld *LocationDeciders) canTrigger(playerID string, shouldRemove bool) bool {
	ld.Lock()
	defer ld.Unlock()

	canTrigger, ok := ld.currentDeciders[playerID]
	if !ok {
		return false
	}

	if !canTrigger {
		return false
	}

	// set to false, if remove is set
	if shouldRemove {
		ld.currentDeciders[playerID] = false
	}

	return true
}

func (ld *LocationDeciders) hasSelector() bool {
	ld.RLock()
	defer ld.RUnlock()

	for _, canTrigger := range ld.currentDeciders {
		if canTrigger {
			return true
		}
	}

	return false
}

type AbilityConfig struct {
	BattleAbilityOptInDuration          time.Duration
	BattleAbilityLocationSelectDuration time.Duration
	DeadlyAbilityShowUpUntilSeconds     int
	FirstAbilityLabel                   string
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	// initialise new gabs ability pool
	as := &AbilitiesSystem{
		arenaID:   battle.ArenaID,
		_battle:   battle,
		startedAt: time.Now(),
		BattleAbilityPool: &AbilityPool{
			Stage: &GabsBribeStage{
				Phase:   atomic.NewInt32(BribeStageHold),
				endTime: time.Now().AddDate(1, 0, 0), // HACK: set end time to far future to implement infinite time
			},
			BattleAbility: &BattleAbility{},
			LocationDeciders: &LocationDeciders{
				m:               make(map[string][]string),
				currentDeciders: make(map[string]bool),
			},
			config: &AbilityConfig{
				BattleAbilityOptInDuration:          time.Duration(db.GetIntWithDefault(db.KeyBattleAbilityBribeDuration, 20)) * time.Second,
				BattleAbilityLocationSelectDuration: time.Duration(db.GetIntWithDefault(db.KeyBattleAbilityLocationSelectDuration, 20)) * time.Second,
				DeadlyAbilityShowUpUntilSeconds:     db.GetIntWithDefault(db.KeyAdvanceBattleAbilityShowUpUntilSeconds, 300),
				FirstAbilityLabel:                   db.GetStrWithDefault(db.KeyFirstBattleAbilityLabel, "LANDMINE"),
			},
		},
		isClosed:           atomic.Bool{},
		locationSelectChan: make(chan *locationSelect),
	}
	as.isClosed.Store(false)

	as.BattleAbilityPool.LocationDeciders.m[server.RedMountainFactionID] = []string{}
	as.BattleAbilityPool.LocationDeciders.m[server.BostonCyberneticsFactionID] = []string{}
	as.BattleAbilityPool.LocationDeciders.m[server.ZaibatsuFactionID] = []string{}

	// init battle ability
	_, err := as.SetNewBattleAbility(true)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set up battle ability")
		return nil
	}

	// bribe cycle
	go as.StartGabsAbilityPoolCycle(false)

	return as
}

// SetNewBattleAbility set new battle ability and return the cooldown time
func (as *AbilitiesSystem) SetNewBattleAbility(isFirst bool) (int, error) {
	if !AbilitySystemIsAvailable(as) {
		return 30, terror.Error(fmt.Errorf("ability system is closed"), "Ability system is closed.")
	}

	// offering id
	offeringID := as.BattleAbilityPool.BattleAbility.LoadOfferingID()

	// uncheck all the opted in players
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
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/arena/%s/battle_ability/check_opt_in", ba.PlayerID, as.arenaID), HubKeyBattleAbilityOptInCheck, false)
		}

	}(offeringID)

	as.BattleAbilityPool.Lock()
	defer as.BattleAbilityPool.Unlock()

	firstAbilityLabel := as.BattleAbilityPool.config.FirstAbilityLabel
	if !isFirst {
		firstAbilityLabel = ""
	}

	includeDeadlyAbilities := false
	if int(time.Now().Sub(as.startedAt).Seconds()) > as.BattleAbilityPool.config.DeadlyAbilityShowUpUntilSeconds {
		// bring in advance battle ability
		includeDeadlyAbilities = true
	}

	// initialise new gabs ability pool
	ba, err := db.BattleAbilityGetRandom(firstAbilityLabel, includeDeadlyAbilities)
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

	ws.PublishMessage(fmt.Sprintf("/public/arena/%s/battle_ability", as.arenaID), HubKeyBattleAbilityUpdated, GameAbility{
		ID:                     ga.ID,
		GameClientAbilityID:    byte(ga.GameClientAbilityID),
		ImageUrl:               ga.ImageURL,
		Description:            ba.Description,
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
	"COOLDOWN"}

type GabsBribeStage struct {
	Phase   *atomic.Int32 `json:"phase"`
	endTime time.Time     `json:"end_time"`
	deadlock.RWMutex
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
		ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
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
				continue
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

				as.locationDecidersSet()

				// get another ability if no one opt in
				if as.BattleAbilityPool.LocationDeciders.maxSelectorCount() == 0 {
					// set new battle ability
					cooldownSecond, err := as.SetNewBattleAbility(false)
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
					}

					as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
					as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
					// broadcast stage to frontend
					ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
					continue
				}

				// broadcast the announcement to the next location decider
				ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
				// announce winner
				ga, err := boiler.GameAbilities(
					boiler.GameAbilityWhere.BattleAbilityID.EQ(null.StringFrom(ba.ID)),
				).One(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("battle ability", ba.Label).Err(err).Msg("Failed to load game ability for notification")
				}

				endTime := time.Now().Add(as.BattleAbilityPool.config.BattleAbilityLocationSelectDuration)

				// assign ability user
				as.BattleAbilityPool.LocationDeciders.rangeSelectors(func(playerID string) {
					ws.PublishMessage(fmt.Sprintf("/secure/user/%s", playerID), HubKeyBribingWinnerSubscribe, struct {
						GameAbility *boiler.GameAbility `json:"game_ability"`
						EndTime     time.Time           `json:"end_time"`
					}{
						GameAbility: ga,
						EndTime:     endTime,
					})
				})

				// change stage
				as.BattleAbilityPool.Stage.Phase.Store(BribeStageLocationSelect)
				as.BattleAbilityPool.Stage.StoreEndTime(endTime)
				// broadcast stage to frontend
				ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

			// at the end of location select phase
			// pass the location select to next player
			case BribeStageLocationSelect:
				if !AbilitySystemIsAvailable(as) {
					continue
				}

				// set new battle ability
				cooldownSecond, err := as.SetNewBattleAbility(false)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
				}

				as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
				// broadcast stage to frontend
				ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
				continue

			// at the end of cooldown phase
			// random choose a battle ability for next bribing session
			case BribeStageCooldown:

				// change bribing phase
				as.BattleAbilityPool.Stage.Phase.Store(BribeStageOptIn)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(as.BattleAbilityPool.config.BattleAbilityOptInDuration))
				// broadcast stage to frontend
				ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

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

			// check battle phase
			if as.BattleAbilityPool.Stage.Phase.Load() != BribeStageLocationSelect {
				continue
			}

			// check eligibility
			if !as.BattleAbilityPool.LocationDeciders.canTrigger(ls.userID, true) {
				continue
			}

			offeringID := uuid.Must(uuid.NewV4())

			// start ability trigger process
			ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
			ga, err := ba.GameAbilities(
				boiler.GameAbilityWhere.FactionID.EQ(ls.factionID),
			).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("battle ability id", ba.ID).Msg("Failed to get game ability")
				return
			}

			// get player detail
			player, err := boiler.FindPlayer(gamedb.StdConn, ls.userID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("player id", player.ID).Msg("Failed to get player detail")
				return
			}

			faction, err := player.Faction().One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("player id", player.ID).Msg("Failed to get faction detail")
				return
			}

			go as.launchAbility(ls, offeringID, ga, player, faction)

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

			// enter cool down, when every selector fire the ability
			if !as.BattleAbilityPool.LocationDeciders.hasSelector() {
				// set new battle ability
				cooldownSecond, err := as.SetNewBattleAbility(false)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
				}

				as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
				// broadcast stage to frontend
				ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
				continue
			}
		}
	}
}

func (as *AbilitiesSystem) launchAbility(ls *locationSelect, offeringID uuid.UUID, gameAbility *boiler.GameAbility, player *boiler.Player, faction *boiler.Faction) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at launching ability.", r)
		}
	}()

	btl, ok := as.battle()
	if !ok {
		return
	}

	if gameAbility.DisplayOnMiniMap {
		mma := &MiniMapAbilityContent{
			OfferingID: offeringID.String(),
			Location: server.CellLocation{
				X: ls.startPoint.X,
				Y: ls.startPoint.Y,
			},
			LocationSelectType:       gameAbility.LocationSelectType,
			ImageUrl:                 gameAbility.ImageURL,
			Colour:                   gameAbility.Colour,
			MiniMapDisplayEffectType: gameAbility.MiniMapDisplayEffectType,
			MechDisplayEffectType:    gameAbility.MechDisplayEffectType,
		}

		// if delay second is greater than zero
		if gameAbility.LaunchingDelaySeconds > 0 {
			duration := time.Duration(gameAbility.LaunchingDelaySeconds) * time.Second

			mma.LaunchingAt = null.TimeFrom(time.Now().Add(duration))

			// add ability onto pending list, and broadcast
			ws.PublishMessage(
				fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", as.arenaID),
				server.HubKeyMiniMapAbilityDisplayList,
				btl.MiniMapAbilityDisplayList.Add(offeringID.String(), mma),
			)

			// time sleep
			time.Sleep(duration)

			// check ability system and battle are still available after time sleep
			if !AbilitySystemIsAvailable(as) {
				return
			}
			btl, ok = as.battle()
			if !ok {
				return
			}
		}

		mma.LaunchingAt = null.TimeFromPtr(nil)
		if ability := btl.abilityDetails[gameAbility.GameClientAbilityID]; ability != nil && ability.Radius > 0 {
			mma.Radius = null.IntFrom(ability.Radius)
		}
		// broadcast changes
		ws.PublishMessage(
			fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", as.arenaID),
			server.HubKeyMiniMapAbilityDisplayList,
			btl.MiniMapAbilityDisplayList.Add(offeringID.String(), mma),
		)

		if gameAbility.AnimationDurationSeconds > 0 {
			go func(battle *Battle, abilityContent *MiniMapAbilityContent, animationSeconds int) {
				time.Sleep(time.Duration(animationSeconds) * time.Second)

				if battle != nil && battle.stage.Load() == BattleStageStart {
					if ab := battle.MiniMapAbilityDisplayList.Get(offeringID.String()); ab != nil {
						ws.PublishMessage(
							fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", battle.ArenaID),
							server.HubKeyMiniMapAbilityDisplayList,
							battle.MiniMapAbilityDisplayList.Remove(offeringID.String()),
						)
					}
				}
			}(btl, mma, gameAbility.AnimationDurationSeconds)
		}
	}

	userUUID := uuid.FromStringOrNil(ls.userID)

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: byte(gameAbility.GameClientAbilityID),
		TriggeredByUserID:   &userUUID,
		TriggeredByUsername: &player.Username.String,
		EventID:             offeringID,
		FactionID:           &ls.factionID,
	}

	event.GameLocation = btl.getGameWorldCoordinatesFromCellXY(&ls.startPoint)

	if gameAbility.LocationSelectType == boiler.LocationSelectTypeEnumLINE_SELECT && ls.endPoint != nil {
		event.GameLocationEnd = btl.getGameWorldCoordinatesFromCellXY(ls.endPoint)
	}

	// trigger location select
	btl.arena.Message("BATTLE:ABILITY", event)

	btl.arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
		Type: LocationSelectTypeTrigger,
		Ability: &AbilityBrief{
			Label:    gameAbility.Label,
			ImageUrl: gameAbility.ImageURL,
			Colour:   gameAbility.Colour,
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

// locationDecidersSet set a user list for location select for current ability triggered
func (as *AbilitiesSystem) locationDecidersSet() {
	// clear current location deciders
	as.BattleAbilityPool.LocationDeciders.clear()

	if !AbilitySystemIsAvailable(as) {
		return
	}

	// check ability is advance
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

	// get maximum selector count
	maximumCommanderCount := as.BattleAbilityPool.BattleAbility.MaximumCommanderCount

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

				// skip, if player is banned
				if slices.IndexFunc(pbs, func(pb *boiler.PlayerBan) bool { return pb.BannedPlayerID == ba.PlayerID }) >= 0 {
					continue
				}

				// check player is inactive (no mouse movement)
				if slices.IndexFunc(aps, func(ap *boiler.PlayerActiveLog) bool { return ap.PlayerID == ba.PlayerID }) == -1 {
					continue
				}

				// set location decider list
				as.BattleAbilityPool.LocationDeciders.setSelector(ba.FactionID, ba.PlayerID)

				// exit, if maximum reached
				if as.BattleAbilityPool.LocationDeciders.length(ba.FactionID) >= maximumCommanderCount {
					break
				}
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

	if as.BattleAbilityPool.Stage.Phase.Load() != BribeStageLocationSelect {
		return nil
	}

	if !as.BattleAbilityPool.LocationDeciders.canTrigger(userID, false) {
		return nil
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
	ws.PublishMessage(fmt.Sprintf("/public/arena/%s/bribe_stage", as.arenaID), HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
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
