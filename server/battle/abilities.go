package battle

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/terror/v2"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	sync.RWMutex
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
		return false
	}

	as.RLock()
	defer as.RUnlock()

	// no battle instance
	if as._battle == nil {
		return false
	}

	// battle ended
	if as._battle.stage.Load() == BattleStageEnd {
		return false
	}

	// no current battle
	if as._battle.arena._currentBattle == nil {
		return false
	}

	// battle mismatch
	if as._battle != as._battle.arena._currentBattle {
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
	sync.RWMutex
}

type BattleAbility struct {
	*boiler.BattleAbility
	OfferingID string
	sync.RWMutex
}

func (ba *BattleAbility) store(battleAbility *boiler.BattleAbility, offeringID string) {
	ba.Lock()
	defer ba.Unlock()

	ba.BattleAbility = battleAbility
	ba.OfferingID = offeringID
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
	m map[string][]string
	sync.RWMutex
}

func (ld *LocationDeciders) clear() {
	ld.Lock()
	defer ld.Unlock()

	for key := range ld.m {
		ld.m[key] = []string{}
	}
}

func (ld *LocationDeciders) store(factionID string, userID string) {
	ld.Lock()
	defer ld.Unlock()

	if _, ok := ld.m[factionID]; !ok {
		return
	}

	ld.m[factionID] = append(ld.m[factionID], userID)
}

func (ld *LocationDeciders) pop(factionID string) {
	ld.Lock()
	defer ld.Unlock()
	li, ok := ld.m[factionID]

	if !ok || len(li) == 0 {
		return
	}

	if len(li) == 1 {
		ld.m[factionID] = []string{}
		return
	}

	ld.m[factionID] = li[1:]
}

func (ld *LocationDeciders) length(factionID string) int {
	ld.RLock()
	defer ld.RUnlock()
	if li, ok := ld.m[factionID]; ok {
		return len(li)
	}

	return 0
}

func (ld *LocationDeciders) first(factionID string) string {
	ld.RLock()
	defer ld.RUnlock()

	if ids, ok := ld.m[factionID]; ok && len(ids) > 0 {
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
			BattleAbility: &BattleAbility{},
			LocationDeciders: &LocationDeciders{
				m: make(map[string][]string),
			},
			config: &AbilityConfig{
				BattleAbilityOptInDuration:          time.Duration(db.GetIntWithDefault(db.KeyBattleAbilityBribeDuration, 30)) * time.Second,
				BattleAbilityLocationSelectDuration: time.Duration(db.GetIntWithDefault(db.KeyBattleAbilityLocationSelectDuration, 15)) * time.Second,
				AdvanceAbilityShowUpUntilSeconds:    db.GetIntWithDefault(db.KeyAdvanceBattleAbilityShowUpUntilSeconds, 300),
				AdvanceAbilityLabel:                 db.GetStrWithDefault(db.KeyAdvanceBattleAbilityLabel, "NUKE"),
			},
		},
		isClosed: atomic.Bool{},
	}
	as.isClosed.Store(false)

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
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("FactionBattleAbilityGet failed to retrieve shit")
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
		OfferingID:             uuid.Must(uuid.NewV4()), // remove offering id to disable bribing
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

var BribeStages = [4]string{"HOLD", "BRIBE",
	"LOCATION_SELECT",
	"COOLDOWN"}

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
		ws.PublishMessage("/battle/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
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

			// skip, if the end time of current phase haven't been reached
			if as.BattleAbilityPool.Stage.EndTime().After(time.Now()) {
				continue
			}

			switch as.BattleAbilityPool.Stage.Phase.Load() {

			// return early if it is in hold stage
			case BribeStageHold:
				continue

			// at the end of bribing phase
			// no ability is triggered, switch to cooldown phase
			case BribeStageOptIn:
				if !AbilitySystemIsAvailable(as) {
					// fire close
					continue
				}
				// set new battle ability
				cooldownSecond, err := as.SetNewBattleAbility()
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
				}

				as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
				// broadcast stage to frontend
				ws.PublishMessage("/battle/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

			// at the end of location select phase
			// pass the location select to next player
			case BribeStageLocationSelect:
				if !AbilitySystemIsAvailable(as) {
					continue
				}

				ba := as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
				// get game ability
				ga, err := ba.GameAbilities().One(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Str("battle ability", ba.Label).Err(err).Msg("Failed to load game ability for notification")
				}

				// get the next location decider
				currentUserID, nextUserID, ok := as.nextLocationDeciderGet()
				if !ok {
					if as.BattleAbilityPool == nil {
						gamelog.L.Error().Str("log_name", "battle arena").Msg("ability pool is nil")
						continue
					}

					// broadcast no ability
					as.broadcastLocationSelectNotification(&GameNotificationLocationSelect{
						Type: LocationSelectTypeCancelledNoPlayer,
						Ability: &AbilityBrief{
							Label:    ga.Label,
							ImageUrl: ga.ImageURL,
							Colour:   ga.Colour,
						},
					})

					// set new battle ability
					cooldownSecond, err := as.SetNewBattleAbility()
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
					}
					// enter cooldown phase, if there is no user left for location select
					as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
					as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
					ws.PublishMessage("/battle/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)
					continue
				}

				notification := &GameNotificationLocationSelect{
					Type: LocationSelectTypeFailedTimeout,
					Ability: &AbilityBrief{
						Label:    ga.Label,
						ImageUrl: ga.ImageURL,
						Colour:   ga.Colour,
					},
				}

				// get current player
				currentPlayer, err := BuildUserDetailWithFaction(currentUserID)
				if err == nil {
					notification.CurrentUser = currentPlayer
				}

				// get next player
				nextPlayer, err := BuildUserDetailWithFaction(nextUserID)
				if err == nil {
					notification.NextUser = nextPlayer
				}

				go as.broadcastLocationSelectNotification(notification)

				// extend location select phase duration
				as.BattleAbilityPool.Stage.Phase.Store(BribeStageLocationSelect)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(as.BattleAbilityPool.config.BattleAbilityLocationSelectDuration))
				// broadcast stage to frontend
				ws.PublishMessage("/battle/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

				// broadcast the announcement to the next location decider
				ws.PublishMessage(fmt.Sprintf("/user/%s", nextUserID), HubKeyBribingWinnerSubscribe, struct {
					GameAbility *boiler.GameAbility `json:"game_ability"`
					EndTime     time.Time           `json:"end_time"`
				}{
					GameAbility: ga,
					EndTime:     as.BattleAbilityPool.Stage.EndTime(),
				})

			// at the end of cooldown phase
			// random choose a battle ability for next bribing session
			case BribeStageCooldown:

				// change bribing phase
				as.BattleAbilityPool.Stage.Phase.Store(BribeStageOptIn)
				as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(as.BattleAbilityPool.config.BattleAbilityOptInDuration))
				// broadcast stage to frontend
				ws.PublishMessage("/battle/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

				continue
			default:
				gamelog.L.Error().Str("log_name", "battle arena").Msg("hit default case switch on abilities loop")
			}
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

	// shuffle players
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(bao), func(i, j int) { bao[i], bao[j] = bao[j], bao[i] })

	for _, ba := range bao {
		// skip, if player is not connected
		if !ws.IsConnected(ba.PlayerID) {
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
		if !exist {
			continue
		}

		//
		if as.BattleAbilityPool.LocationDeciders.length(ba.FactionID) == 2 {
			continue
		}

		// set player ability
		as.BattleAbilityPool.LocationDeciders.store(ba.FactionID, ba.PlayerID)
	}

	playerList, err := db.PlayerFactionContributionList(battleID, factionID, abilityOfferingID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("battle_id", battleID).Str("faction_id", factionID).Err(err).Msg("failed to get player list")
	}

	// sort the order of the list
	tempList := []uuid.UUID{}
	for _, tid := range triggerByUserID {
		tempList = append(tempList, tid)
	}
	for _, pid := range playerList {
		exists := false
		for _, tid := range triggerByUserID {
			if pid == tid {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		tempList = append(tempList, pid)
	}

	// get location select limited players
	punishedPlayers, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get limited select players from db")
	}
	// initialise location select list
	as.locationDeciders.list = []uuid.UUID{}

	for _, pid := range tempList {
		isPunished := false
		// check user is banned
		for _, pp := range punishedPlayers {

			if pp.BannedPlayerID == pid.String() {
				isPunished = true
				break
			}
		}

		// append to the list if player is not punished
		if !isPunished {
			as.locationDeciders.list = append(as.locationDeciders.list, pid)
		}
	}
}

// nextLocationDeciderGet return the uuid of the next player to select the location for ability
func (as *AbilitiesSystem) nextLocationDeciderGet() (uuid.UUID, uuid.UUID, bool) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the nextLocationDeciderGet!", r)
		}
	}()
	if as.locationDeciders == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Msg("nil check failed as.locationDeciders")

		return uuid.UUID(uuid.Nil), uuid.UUID(uuid.Nil), false
	}

	// clean up the location select list if there is no user left to select location
	if len(as.locationDeciders.list) <= 1 {
		gamelog.L.Error().Str("log_name", "battle arena").Msg("no as.locationDeciders <= 1")
		as.locationDeciders.list = []uuid.UUID{}
		return uuid.UUID(uuid.Nil), uuid.UUID(uuid.Nil), false
	}

	currentUserID := as.locationDeciders.list[0]
	nextUserID := as.locationDeciders.list[1]

	// remove the first user from the list
	as.locationDeciders.list = as.locationDeciders.list[1:]

	return currentUserID, nextUserID, true
}

func (as *AbilitiesSystem) BribeStageGet() *GabsBribeStageNormalised {
	if as.BattleAbilityPool != nil {
		return as.BattleAbilityPool.Stage.Normalise()
	}
	return nil
}

func (as *AbilitiesSystem) FactionBattleAbilityGet(factionID string) (*GameAbility, error) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the FactionBattleAbilityGet!", r)
		}
	}()
	if as.BattleAbilityPool == nil {
		return nil, fmt.Errorf("BattleAbilityPool is nil, fid: %s", factionID)
	}
	if as.BattleAbilityPool.Abilities == nil {
		return nil, fmt.Errorf("BattleAbilityPool.Abilities is nil, fid: %s", factionID)
	}

	ability, ok := as.BattleAbilityPool.Abilities.Load(factionID)
	if !ok {
		gamelog.L.Warn().Str("func", "FactionBattleAbilityGet").Msg("unable to retrieve abilities for faction")
		return nil, fmt.Errorf("game ability does not exist for faction %s", factionID)
	}

	return ability, nil
}

func (as *AbilitiesSystem) LocationSelect(userID uuid.UUID, startPoint server.CellLocation, endPoint *server.CellLocation) error {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the LocationSelect!", r)
		}
	}()
	// check battle end
	if as.battle().stage.Load() == BattleStageEnd {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("battle stage has en ended")
		return nil
	}

	// check eligibility
	if len(as.locationDeciders.list) <= 0 || as.locationDeciders.list[0] != userID {
		return terror.Error(terror.ErrForbidden)
	}

	ability, _ := as.BattleAbilityPool.Abilities.Load(as.BattleAbilityPool.TriggeredFactionID.Load())
	// get player detail
	player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(userID.String())).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "player not exists")
	}

	faction, err := boiler.Factions(boiler.FactionWhere.ID.EQ(as.BattleAbilityPool.TriggeredFactionID.Load())).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "player not exists")
	}

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: ability.GameClientAbilityID,
		TriggeredByUserID:   &userID,
		TriggeredByUsername: &player.Username.String,
		EventID:             ability.OfferingID,
		FactionID:           &faction.ID,
	}

	event.GameLocation = as.battle().getGameWorldCoordinatesFromCellXY(&startPoint)

	if ability.LocationSelectType == boiler.LocationSelectTypeEnumLINE_SELECT && endPoint != nil {
		event.GameLocationEnd = as.battle().getGameWorldCoordinatesFromCellXY(endPoint)
	}

	// trigger location select
	as.battle().arena.Message("BATTLE:ABILITY", event)

	bat := boiler.BattleAbilityTrigger{
		PlayerID:          null.StringFrom(userID.String()),
		BattleID:          as.battle().ID,
		FactionID:         ability.FactionID,
		IsAllSyndicates:   true,
		AbilityLabel:      ability.Label,
		GameAbilityID:     ability.ID,
		AbilityOfferingID: ability.OfferingID.String(),
	}
	err = bat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Interface("battle_ability_trigger", bat).Err(err).Msg("Failed to record ability triggered")
	}

	_, err = db.UserStatAddTotalAbilityTriggered(userID.String())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", userID.String()).Err(err).Msg("failed to update user ability triggered amount")
	}

	as.battle().arena.BroadcastGameNotificationLocationSelect(&GameNotificationLocationSelect{
		Type: LocationSelectTypeTrigger,
		Ability: &AbilityBrief{
			Label:    ability.Label,
			ImageUrl: ability.ImageUrl,
			Colour:   ability.Colour,
		},
		CurrentUser: &UserBrief{
			ID:        userID,
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

	//// enter the cooldown phase
	cooldownSecond, err := as.SetNewBattleAbility()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to set new battle ability")
	}

	as.BattleAbilityPool.Stage.Phase.Store(BribeStageCooldown)
	as.BattleAbilityPool.Stage.StoreEndTime(time.Now().Add(time.Duration(cooldownSecond) * time.Second))
	// broadcast stage to frontend
	ws.PublishMessage("/battle/bribe_stage", HubKeyBribeStageUpdateSubscribe, as.BattleAbilityPool.Stage)

	return nil
}

func (as *AbilitiesSystem) End() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic at the abilities.End!", r)
		}
	}()

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
