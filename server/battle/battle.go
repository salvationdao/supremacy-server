package battle

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/replay"
	"server/system_messages"
	"server/xsyn_rpcclient"
	"sort"
	"strings"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog/log"

	"golang.org/x/exp/slices"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ninja-syndicate/ws"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
)

const (
	EndState      = 0
	SetupState    = 1
	IntroState    = 2
	BattlingState = 3
)

type Battle struct {
	*boiler.Battle
	arena                  *Arena
	state                  *atomic.Int32
	MapName                string        `json:"mapName"`
	WarMachines            []*WarMachine `json:"warMachines"`
	spawnedAIMux           deadlock.RWMutex
	SpawnedAI              []*WarMachine `json:"SpawnedAI"`
	warMachineIDs          []string
	lastTick               *[]byte
	lobby                  *boiler.BattleLobby
	gameMap                *server.GameMap
	battleZones            []server.BattleZone
	currentBattleZoneIndex int
	nextMapID              null.String
	rpcClient              *xsyn_rpcclient.XrpcClient
	battleMechData         []*db.BattleMechData
	startedAt              time.Time
	replaySession          *RecordingSession

	_playerAbilityManager *PlayerAbilityManager

	destroyedWarMachineMap map[string]*WMDestroyedRecord

	abilityDetails []*AbilityDetail

	MiniMapAbilityDisplayList *MiniMapAbilityDisplayList

	MapEventList *MapEventList

	deadlock.RWMutex

	// for reword calculation
	playerBattleCompleteMessage  []*PlayerBattleCompleteMessage
	stakedMechOwnerRewardMessage []*PlayerBattleCompleteMessage
	mechRewards                  []*MechReward

	// for afk checker
	introEndedAt time.Time
}

type MechBattleBrief struct {
	MechID    string      `json:"mech_id"`
	Name      string      `json:"name"`
	Tier      string      `json:"tier"`
	ImageUrl  string      `json:"image_url"`
	FactionID string      `json:"faction_id"`
	Kills     []*KillInfo `json:"kills,omitempty"`
	KilledBy  *KillInfo   `json:"killed,omitempty"`
}

type KillInfo struct {
	Name      string `json:"name"`
	FactionID string `json:"faction_id"`
	ImageUrl  string `json:"image_url"`
}

type MiniMapAbilityDisplayList struct {
	arenaID string
	list    []*MiniMapAbilityContent
	deadlock.RWMutex

	broadcastChan chan *MiniMapAbilityContent
	stop          chan bool
}

type MiniMapAbilityContent struct {
	OfferingID               string              `json:"offering_id"`
	Location                 server.CellLocation `json:"location"`
	MechID                   string              `json:"mech_id"`
	ImageUrl                 string              `json:"image_url"`
	Colour                   string              `json:"colour"`
	MiniMapDisplayEffectType string              `json:"mini_map_display_effect_type"`
	MechDisplayEffectType    string              `json:"mech_display_effect_type"`
	LocationSelectType       string              `json:"location_select_type"`
	Radius                   null.Int            `json:"radius,omitempty"`
	LaunchingAt              null.Time           `json:"launching_at,omitempty"`
	IsRemoved                bool                `json:"is_removed"`
	UpdatedAt                time.Time           `json:"-"`
}

// Add new pending ability and return a copy of current list
func (dap *MiniMapAbilityDisplayList) Add(dac *MiniMapAbilityContent) {
	// set updated at
	dac.UpdatedAt = time.Now()

	// update mini map ability list
	dap.Lock()
	defer dap.Unlock()

	lastIndex := len(dap.list) - 1
	index := slices.IndexFunc(dap.list, func(da *MiniMapAbilityContent) bool { return da.OfferingID == dac.OfferingID })
	if index != -1 {
		dap.list[index] = dap.list[lastIndex] // replace target element with the last element
		dap.list[lastIndex] = nil             // free up the memory of target element
		dap.list = dap.list[:lastIndex]       // truncate the list
	}

	dap.list = append(dap.list, dac)

	sort.Slice(dap.list, func(i, j int) bool { return dap.list[i].UpdatedAt.After(dap.list[j].UpdatedAt) })

	// broadcast changes
	select {
	case dap.broadcastChan <- dac:
	case <-time.After(100 * time.Millisecond):
		// timeout
	}
}

// Remove pending ability and return a copy of current list
func (dap *MiniMapAbilityDisplayList) Remove(offeringID string) {
	dap.Lock()
	defer dap.Unlock()

	index := slices.IndexFunc(dap.list, func(da *MiniMapAbilityContent) bool { return da.OfferingID == offeringID })
	if index == -1 {
		return
	}

	// broadcast changes
	dac := dap.list[index]
	dac.IsRemoved = true
	select {
	case dap.broadcastChan <- dac:
	case <-time.After(100 * time.Millisecond):
		// timeout
	}

	// remove data
	lastIndex := len(dap.list) - 1
	dap.list[index] = dap.list[lastIndex] // replace target element with the last element
	dap.list[lastIndex] = nil             // free up the memory of target element
	dap.list = dap.list[:lastIndex]       // truncate the list

	sort.Slice(dap.list, func(i, j int) bool { return dap.list[i].UpdatedAt.After(dap.list[j].UpdatedAt) })
}

// List a copy of current pending list
func (dap *MiniMapAbilityDisplayList) List() []*MiniMapAbilityContent {
	dap.RLock()
	defer dap.RUnlock()
	return dap.list
}

func (dap *MiniMapAbilityDisplayList) debounceBroadcastMiniMapDisplay() {
	interval := 150 * time.Millisecond
	timer := time.NewTimer(5 * time.Minute)
	var broadcastList []*MiniMapAbilityContent

	for {
		select {
		case <-dap.stop:
			return

		case dac := <-dap.broadcastChan:
			index := slices.IndexFunc(broadcastList, func(bl *MiniMapAbilityContent) bool { return bl.OfferingID == dac.OfferingID })
			if index != -1 {
				broadcastList[index] = dac
			} else {
				broadcastList = append(broadcastList, dac)
			}

			timer.Reset(interval)

		case <-timer.C:
			ws.PublishMessage(
				fmt.Sprintf("/mini_map/arena/%s/public/mini_map_ability_display_list", dap.arenaID),
				server.HubKeyMiniMapAbilityContentSubscribe,
				broadcastList,
			)

			broadcastList = []*MiniMapAbilityContent{}
		}
	}
}

type RecordingSession struct {
	ReplaySession *boiler.BattleReplay `json:"replay_session"`
	Events        []*RecordingEvents   `json:"battle_events"`
}

type RecordingEvents struct {
	Timestamp    time.Time        `json:"timestamp"`
	Notification GameNotification `json:"notification"`
}

func (btl *Battle) playerAbilityManager() *PlayerAbilityManager {
	btl.RLock()
	defer btl.RUnlock()
	return btl._playerAbilityManager
}

// storeGameMap set the game map detail from game client
func (btl *Battle) storeGameMap(gm server.GameMap, battleZones []server.BattleZone) {
	gamelog.L.Trace().Str("func", "storeGameMap").Msg("start")
	btl.Lock()
	defer btl.Unlock()

	btl.gameMap.Name = gm.Name
	btl.gameMap.ImageUrl = gm.ImageUrl
	btl.gameMap.Width = gm.Width
	btl.gameMap.Height = gm.Height
	btl.gameMap.CellsX = gm.CellsX
	btl.gameMap.CellsY = gm.CellsY
	btl.gameMap.PixelLeft = gm.PixelLeft
	btl.gameMap.PixelTop = gm.PixelTop
	btl.gameMap.DisabledCells = gm.DisabledCells
	btl.battleZones = battleZones
	gamelog.L.Trace().Str("func", "storeGameMap").Msg("end")
}

// cleanUpBattleRecord remove all the record of the battle
func cleanUpBattleRecord(battleID string) {
	now := time.Now()

	battle, err := boiler.FindBattle(gamedb.StdConn, battleID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("not able to load previous battle")
	}

	l := gamelog.L.With().Str("log_name", "battle arena").Interface("battle", battle).Str("battle.go", ":battle.go:battle.Battle()").Logger()

	// refund abilities
	go ReversePlayerAbilities(battle.ID, battle.BattleNumber)

	// stops recording process of the previous battle
	go func(battleID, arenaID string) {
		reRunBattle, err := boiler.FindBattle(gamedb.StdConn, battleID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("battle_id", battleID).Msg("Failed to get battle while stopping recording")
			return
		}
		prevReplay, err := boiler.BattleReplays(
			boiler.BattleReplayWhere.BattleID.EQ(battleID),
			boiler.BattleReplayWhere.ArenaID.EQ(arenaID),
			boiler.BattleReplayWhere.RecordingStatus.EQ(boiler.RecordingStatusRECORDING),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("battle_id", battleID).Msg("Failed to find previous replay")
			return
		}
		// url request
		err = replay.RecordReplayRequest(reRunBattle, prevReplay.ID, replay.StopRecording)
		if err != nil {
			if err != replay.ErrDontLogRecordingStatus {
				gamelog.L.Error().Err(err).Str("battle_id", battleID).Str("replay_id", prevReplay.ID).Msg("Failed to stop recording")
				return
			}
			return
		}

		// update start time
		prevReplay.StoppedAt = null.TimeFrom(time.Now())
		prevReplay.RecordingStatus = boiler.RecordingStatusSTOPPED
		_, err = prevReplay.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("battle_id", prevReplay.BattleID).Str("replay_id", prevReplay.ID).Err(err).Msg("Failed to update recording status to STOPPED while starting battle")
			return
		}
	}(battle.ID, battle.ArenaID)

	_, err = boiler.BattleMechs(boiler.BattleMechWhere.BattleID.EQ(battle.ID)).DeleteAll(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to delete delete stale battle mechs from database")
	}

	_, err = boiler.BattleWins(boiler.BattleWinWhere.BattleID.EQ(battle.ID)).DeleteAll(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to delete delete stale battle wins from database")
	}

	_, err = boiler.BattleKills(boiler.BattleKillWhere.BattleID.EQ(battle.ID)).DeleteAll(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to delete delete stale battle kills from database")
	}

	_, err = boiler.BattleHistories(boiler.BattleHistoryWhere.BattleID.EQ(battle.ID)).DeleteAll(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to delete delete stale battle histories from database")
	}

	// soft delete all the ability triggers of the previous battle
	_, err = boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.BattleID.EQ(battle.ID),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.BattleAbilityTriggerColumns.DeletedAt: null.TimeFrom(now)})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to clean up mech ability trigger logs.")
	}

	gamelog.L.Info().Interface("battle", battle).Msg("clean up unfinished battle")
}

func (btl *Battle) storePlayerAbilityManager(im *PlayerAbilityManager) {
	btl.Lock()
	defer btl.Unlock()
	btl._playerAbilityManager = im
}

func (btl *Battle) preIntro(payload *BattleStartPayload) error {
	gamelog.L.Trace().Str("func", "preIntro").Msg("start")

	btl.Lock()
	defer btl.Unlock()

	for _, pwm := range payload.WarMachines {
		index := slices.IndexFunc(btl.WarMachines, func(wm *WarMachine) bool { return pwm.Hash == wm.Hash })

		// skip, if mech not found
		if index == -1 {
			gamelog.L.Error().Str("log_name", "battle arena").Err(fmt.Errorf("didnt find matching hash"))
			continue
		}

		// otherwise, update war machine's participant id
		btl.WarMachines[index].Lock()
		btl.WarMachines[index].ParticipantID = pwm.ParticipantID
		btl.WarMachines[index].Unlock()

		gamelog.L.Trace().Interface("battle war machine", btl.WarMachines[index]).Msg("set participant id of the battle war machine")
	}

	// only broadcast battle state, after receiving the participant id from game client
	btl.BroadcastUpdate()

	gamelog.L.Trace().Str("func", "preIntro").Msg("end")
	return nil
}

func (btl *Battle) start() {
	gamelog.L.Trace().Str("func", "start").Msg("start")

	var err error

	btl.state.Store(BattlingState)
	ws.PublishMessage(fmt.Sprintf("/public/arena/%s/battle_state", btl.ArenaID), server.HubKeyBattleState, BattlingState)

	// handle global announcements
	ga, err := boiler.GlobalAnnouncements().One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting battle start to players")
	}

	// global announcement exists
	if ga != nil {
		// show if battle number is equal or in between the global announcement's to and from battle number
		if btl.BattleNumber >= ga.ShowFromBattleNumber.Int && btl.BattleNumber <= ga.ShowUntilBattleNumber.Int {
			ws.PublishMessage("/public/global_announcement", server.HubKeyGlobalAnnouncementSubscribe, ga)
		}

		// delete if global announcement expired/ is in the past
		if btl.BattleNumber > ga.ShowUntilBattleNumber.Int {
			_, err := boiler.GlobalAnnouncements().DeleteAll(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("Battle ID", btl.ID).Msg("unable to delete global announcement")
			}
			ws.PublishMessage("/public/global_announcement", server.HubKeyGlobalAnnouncementSubscribe, nil)
		}
	}

	gamelog.L.Trace().Str("func", "start").Msg("end")
}

// getGameWorldCoordinatesFromCellXY converts picked cell to the location in game
func (btl *Battle) getGameWorldCoordinatesFromCellXY(cell *server.CellLocation) *server.GameLocation {
	gameMap := btl.gameMap
	// To get the location in game its
	//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + PixelLeft
	//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + PixelTop

	return &server.GameLocation{
		X: int(cell.X.Mul(decimal.NewFromInt(server.GameClientTileSize)).Add(decimal.NewFromInt(server.GameClientTileSize / 2)).Add(decimal.NewFromInt(int64(gameMap.PixelLeft))).IntPart()),
		Y: int(cell.Y.Mul(decimal.NewFromInt(server.GameClientTileSize)).Add(decimal.NewFromInt(server.GameClientTileSize / 2)).Add(decimal.NewFromInt(int64(gameMap.PixelTop))).IntPart()),
	}
}

// getCellCoordinatesFromGameWorldXY converts location in game to a cell location
func (btl *Battle) getCellCoordinatesFromGameWorldXY(location *server.GameLocation) *server.CellLocation {
	gameMap := btl.gameMap

	return &server.CellLocation{
		X: decimal.NewFromInt(int64(location.X)).Sub(decimal.NewFromInt(int64(gameMap.PixelLeft))).Sub(decimal.NewFromInt(server.GameClientTileSize * 2)).Div(decimal.NewFromInt(server.GameClientTileSize)),
		Y: decimal.NewFromInt(int64(location.Y)).Sub(decimal.NewFromInt(int64(gameMap.PixelTop))).Sub(decimal.NewFromInt(server.GameClientTileSize * 2)).Div(decimal.NewFromInt(server.GameClientTileSize)),
	}
}

type WarMachinePosition struct {
	X int
	Y int
}

func (btl *Battle) handleBattleEnd(payload *BattleEndPayload) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the creation of ending info: handleBattleEnd!", r)
		}
	}()

	sublogger := log.With().
		Str("case", "battle_end").
		Str("battle_id", btl.ID).
		Logger()
	now := time.Now()
	sublogger.Debug().Msg("locking arena")
	btl.arena.Manager.Lock()

	// close battle
	sublogger.Debug().Str("correlation_id", "db16df7e-e53c-42eb-ad50-80570f27e835").Msg("close battle")
	btl.Battle.EndedAt = null.TimeFrom(now)
	_, err := btl.Battle.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleColumns.EndedAt))
	if err != nil {
		sublogger.Error().Str("log_name", "battle arena").Interface("battle", btl.Battle).Msg("Failed to up date end_at of current battle.")
	}

	sublogger.Debug().Str("correlation_id", "6ae3f5a4-9b80-4a55-b0c3-2351db96cd11").Msg("close battle lobby")
	oldLobbyID := btl.arena.currentLobbyID.Load()
	// close battle lobby
	btl.lobby.EndedAt = null.TimeFrom(now)
	_, err = btl.lobby.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleLobbyColumns.EndedAt))
	if err != nil {
		sublogger.Error().Str("log_name", "battle arena").Interface("lobby", btl.lobby).Msg("Failed to update ended_at of the battle lobby.")
	}

	sublogger.Debug().Str("correlation_id", "832be7a9-7c2b-44cc-bce7-ddb85848e2e0").Msg("close battle lobby mechs")
	// close battle lobby mechs
	_, err = btl.lobby.BattleLobbiesMechs().UpdateAll(gamedb.StdConn, boiler.M{boiler.BattleLobbiesMechColumns.EndedAt: null.TimeFrom(now)})
	if err != nil {
		sublogger.Error().Str("log_name", "battle arena").Interface("lobby", btl.lobby).Msg("Failed to update ended_at of the battle lobby mechs.")
	}

	sublogger.Debug().Str("correlation_id", "695def8d-385d-4202-8993-fe88ef101da6").Msg("pre-assign next battle lobby")
	// pre-assign next battle lobby
	btl.arena.assignBattleLobby()
	newLobbyID := btl.arena.currentLobbyID.Load()
	btl.arena.Manager.Unlock()

	// broadcast lobby changes
	btl.arena.Manager.BattleLobbyDebounceBroadcastChan <- []string{oldLobbyID, newLobbyID}

	// start the
	winningWarMachines := []*WarMachine{}
	var winningFaction *Faction
	winningFactionID := ""

	sublogger.Debug().Msgf("battle end: looping WinningWarMachines: %s", btl.ID)
	for _, wwm := range payload.WinningWarMachines {
		idx := slices.IndexFunc(btl.WarMachines, func(wm *WarMachine) bool { return wm.Hash == wwm.Hash })
		if idx == -1 {
			sublogger.Error().Str("log_name", "battle arena").Str("Battle ID", btl.ID).Msg("unable to match war machine to battle with hash")
		}
		wm := btl.WarMachines[idx]

		winningWarMachines = append(winningWarMachines, wm)
		winningFaction = btl.WarMachines[idx].Faction
		winningFactionID = winningFaction.ID

		// insert battle win
		mw := &boiler.BattleWin{
			BattleID:     btl.ID,
			WinCondition: payload.WinCondition,
			MechID:       wm.ID,
			OwnerID:      wm.OwnedByID,
			FactionID:    wm.FactionID,
		}
		err = mw.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			sublogger.Error().Str("db func", "WinBattle").Err(err).Msg("unable to commit tx")
		}
	}

	sublogger.Debug().Str("correlation_id", "3bb93c52-9366-4af9-bcc4-9b3f33f5cda0").Msg("load all the mech stats")
	// load all the mech stats
	mechStats, err := boiler.MechStats(boiler.MechStatWhere.MechID.IN(btl.warMachineIDs)).All(gamedb.StdConn)
	if err != nil {
		sublogger.Error().Str("log_name", "battle arena").
			Interface("mech id list", btl.warMachineIDs).
			Err(err).Msg("unable to retrieve mech stats from database")
	}

	sublogger.Debug().Str("correlation_id", "bf2b1fc5-c189-4e43-98a3-34bd4a58e45d").Msg("load all the battle mechs")
	// load all the battle mechs
	battleMechs, err := boiler.BattleMechs(boiler.BattleMechWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
	if err != nil {
		sublogger.Error().Str("log_name", "battle arena").
			Str("battleID", btl.ID).
			Str("db func", "endWarMachines").
			Err(err).Msg("unable to retrieve winning faction battle mechs from database")
	}

	sublogger.Debug().Str("correlation_id", "7fe00358-6d0e-46da-9cb4-ff80ba033406").Msg("start updating")
	// start updating
	for _, bm := range battleMechs {
		// get mech
		idx := slices.IndexFunc(btl.WarMachines, func(wm *WarMachine) bool { return wm.ID == bm.MechID })
		if idx == -1 {
			continue
		}
		wm := btl.WarMachines[idx]

		// get mech stat
		ms := &boiler.MechStat{
			MechID: bm.MechID,
		}
		idx = slices.IndexFunc(mechStats, func(mechStat *boiler.MechStat) bool { return mechStat.MechID == ms.MechID })
		if idx == -1 {
			// insert, if not exists
			err = ms.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				sublogger.Warn().Err(err).
					Interface("boiler.MechStat", ms).
					Msg("unable to create mech stat")
				continue
			}

			// append to the list
			mechStats = append(mechStats, ms)

			// assign current index
			idx = len(mechStats) - 1
		}

		// override mech stat
		ms = mechStats[idx]

		updateBattleMechCols := []string{}
		updateMechStatCols := []string{}

		// if faction won
		if bm.FactionID == winningFactionID {
			// notify winning players
			prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(bm.PilotedByID)).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				sublogger.Error().Str("log_name", "battle arena").Err(err).Str("player_id", bm.PilotedByID).Msg("unable to get player prefs")
				continue
			}

			if prefs != nil && prefs.TelegramID.Valid && prefs.EnableTelegramNotifications {
				// killed a war machine
				msg := fmt.Sprintf("Your War machine %s is Victorious! 🎉", wm.Name)
				err := btl.arena.Manager.telegram.Notify(prefs.TelegramID.Int64, msg)
				if err != nil {
					sublogger.Error().Str("log_name", "battle arena").Str("telegramID", fmt.Sprintf("%v", prefs.TelegramID)).Err(err).Msg("failed to send notification")
				}
			}

			// update battle mech
			bm.FactionWon = null.BoolFrom(true)
			updateBattleMechCols = append(updateBattleMechCols, boiler.BattleMechColumns.FactionWon)

			ms.TotalWins += 1
			updateMechStatCols = append(updateMechStatCols, boiler.MechStatColumns.TotalWins)

			// if survived
			if slices.IndexFunc(winningWarMachines, func(wm *WarMachine) bool { return bm.MechID == wm.ID }) != -1 {
				bm.MechSurvived = null.BoolFrom(true)
				updateBattleMechCols = append(updateBattleMechCols, boiler.BattleMechColumns.MechSurvived)

				// update mech stat
				ms.BattlesSurvived += 1
				updateMechStatCols = append(updateMechStatCols, boiler.MechStatColumns.BattlesSurvived)
			}
		} else {
			// if faction loss
			ms.TotalLosses += 1
			updateMechStatCols = append(updateMechStatCols, boiler.MechStatColumns.TotalLosses)
		}

		// update battle mech, if needed
		if len(updateBattleMechCols) > 0 {
			_, err = bm.Update(gamedb.StdConn, boil.Whitelist(updateBattleMechCols...))
			if err != nil {
				sublogger.Error().Str("log_name", "battle arena").
					Interface("battle mech", bm).
					Strs("updated columns", updateBattleMechCols).
					Err(err).Msg("unable to update battle mech.")
			}
		}

		// update mech stat, if needed
		if len(updateMechStatCols) > 0 {
			_, err = ms.Update(gamedb.StdConn, boil.Whitelist(updateMechStatCols...))
			if err != nil {
				sublogger.Error().Str("log_name", "battle arena").
					Interface("mech stat", ms).
					Strs("updated columns", updateMechStatCols).
					Err(err).Msg("unable to update mech stat.")
			}
		}
	}

	sublogger.Debug().Str("correlation_id", "73ba9f8a-7a77-470a-ad0b-1d48f636694d").Msg("record faction win/loss count")
	// record faction win/loss count
	err = db.FactionAddWinLossCount(winningFactionID)
	if err != nil {
		sublogger.Panic().Str("Battle ID", btl.ID).Str("winning_faction_id", winningFactionID).Msg("Failed to update faction win/loss count")
	}

	sublogger.Debug().Msgf("battle end: looping MostFrequentAbilityExecutors: %s", btl.ID)
	topPlayerExecutorsBoilers, err := db.MostFrequentAbilityExecutors(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		sublogger.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top player executors")
	}

	sublogger.Debug().Msgf("battle end: looping topPlayerExecutorsBoilers: %s", btl.ID)
	topPlayerExecutors := []*BattleUser{}
	for _, p := range topPlayerExecutorsBoilers {
		factionID := uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))
		if p.FactionID.Valid {
			factionID = uuid.Must(uuid.FromString(p.FactionID.String))
		}
		topPlayerExecutors = append(topPlayerExecutors, &BattleUser{
			ID:        uuid.Must(uuid.FromString(p.ID)),
			Username:  p.Username.String,
			FactionID: factionID.String(),
		})
	}

	sublogger.Debug().Str("correlation_id", "fd962257-8eeb-4190-a22d-412c4adae811").Msg("get winning faction order")
	// get winning faction order
	winningFactionIDOrder := []string{winningFactionID}
	factionIDs, err := db.FactionMechDestroyedOrderGet(btl.ID)
	if err != nil {
		sublogger.Error().Err(err).Msg("Failed to load mech destroy order.")
	}

	for _, fid := range factionIDs {
		exist := false
		for _, wid := range winningFactionIDOrder {
			if wid == fid {
				exist = true
			}
		}

		if !exist {
			winningFactionIDOrder = append(winningFactionIDOrder, fid)
		}
	}

	sublogger.Debug().Str("correlation_id", "7cff4d31-55c7-4306-bd3e-cf66669159fb").Msg("declare rewards")
	// declare rewards
	btl.playerBattleCompleteMessage = []*PlayerBattleCompleteMessage{}
	btl.mechRewards = []*MechReward{}

	sublogger.Debug().Str("correlation_id", "874bfebd-55b0-4f40-8a67-b13908d04f25").Msg("reward mech owners")
	// reward mech owners
	btl.RewardBattleMechOwners(winningFactionIDOrder)

	sublogger.Debug().Str("correlation_id", "6fea54ab-5dc3-408b-bae9-8fb454cb92b7").Msg("end info")
	// end info
	endInfo := &BattleEndDetail{
		BattleID:                     btl.ID,
		BattleIdentifier:             btl.Battle.BattleNumber,
		StartedAt:                    btl.Battle.StartedAt,
		EndedAt:                      btl.Battle.EndedAt.Time,
		WinningCondition:             payload.WinCondition,
		WinningFaction:               winningFaction,
		WinningFactionIDOrder:        winningFactionIDOrder,
		WinningWarMachines:           winningWarMachines,
		MostFrequentAbilityExecutors: topPlayerExecutors,
		MechRewards:                  btl.mechRewards,
	}

	sublogger.Debug().Str("correlation_id", "64283805-3a9b-4660-8a1b-dbb7f95d3eb5").Msg("publish message")
	ws.PublishMessage(fmt.Sprintf("/public/arena/%s/battle_end_result", btl.ArenaID), HubKeyBattleEndDetailUpdated, endInfo)

	// cache battle end detail
	btl.arena.LastBattleResult = endInfo

	// broadcast player mech status change
	btl.arena.Manager.MechDebounceBroadcastChan <- btl.warMachineIDs

	sublogger.Debug().Str("correlation_id", "905e7228-9d92-4c4b-b4a7-5491325d2ffe").Msg("broadcast battle eta")
	// broadcast battle eta
	go func() {
		bs, err := boiler.Battles(
			boiler.BattleWhere.EndedAt.IsNotNull(),
			qm.OrderBy(boiler.BattleColumns.BattleNumber+" DESC"),
			qm.Limit(100),
		).All(gamedb.StdConn)
		if err != nil {
			sublogger.Error().Err(err).Msg("Failed to load latest 100 battles")
			return
		}

		var totalDuration time.Duration
		for _, b := range bs {
			totalDuration += b.EndedAt.Time.Sub(b.StartedAt)
		}

		ws.PublishMessage("/secure/battle_eta", server.HubKeyBattleETAUpdate, int(totalDuration.Seconds())/len(bs))
	}()

	sublogger.Debug().Str("correlation_id", "90c52aaf-4ecd-48ca-9e53-f118f147c3ea").Msg("broadcast battle complete system messages")
	// broadcast battle complete system messages
	go func(battle *Battle) {
		// broadcast end info
		for _, msg := range battle.playerBattleCompleteMessage {
			// get mechs data
			for _, bm := range battleMechs {
				// skip, if player is not the owner
				if bm.PilotedByID != msg.PlayerID {
					continue
				}

				mbb := &MechBattleBrief{
					MechID:    bm.MechID,
					FactionID: bm.FactionID,
				}

				if idx := slices.IndexFunc(btl.WarMachines, func(wm *WarMachine) bool { return wm.ID == bm.MechID }); idx != -1 {
					wm := btl.WarMachines[idx]
					mbb.Name = wm.Label
					if wm.Name != "" {
						mbb.Name = wm.Name
					}

					mbb.Tier = wm.Tier
					mbb.ImageUrl = wm.ImageAvatar

					for _, destroyedMechRecord := range btl.destroyedWarMachineMap {
						destroyedMech := destroyedMechRecord.DestroyedWarMachine
						killerMech := destroyedMechRecord.KilledByWarMachine
						killerUser := destroyedMechRecord.KilledByUser

						killInfo := &KillInfo{
							Name:      destroyedMechRecord.KilledBy,
							FactionID: destroyedMechRecord.KillerFactionID,
						}

						// if destroyed mech is current mech
						if destroyedMech.Hash == wm.Hash {
							if killerMech != nil {
								killInfo.Name = killerMech.Name
								killInfo.ImageUrl = killerMech.ImageAvatar
							} else if killerUser != nil {
								killInfo.Name = fmt.Sprintf("%s %s", killerUser.Username, destroyedMechRecord.KilledBy)
							}
							mbb.KilledBy = killInfo // set kill by info
							continue
						} else if killerMech != nil && killerMech.Hash == wm.Hash {
							// if current mech is the killer mech

							killInfo.Name = destroyedMech.Name
							killInfo.FactionID = destroyedMech.FactionID
							killInfo.ImageUrl = destroyedMech.ImageAvatar
							mbb.Kills = append(mbb.Kills, killInfo)
							continue
						}

					}
				}

				msg.MechBattleBriefs = append(msg.MechBattleBriefs, mbb)
			}

			// send battle reward system message
			b, err := json.Marshal(msg)
			if err != nil {
				sublogger.Error().Interface("player reward data", msg).Err(err).Msg("Failed to marshal player reward data into json.")
				break
			}
			sysMsg := boiler.SystemMessage{
				PlayerID: msg.PlayerID,
				SenderID: server.SupremacyBattleUserID,
				DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeMechBattleComplete)),
				Title:    "Battle Complete",
				Message:  fmt.Sprintf("Summary of the battle #%d.", battle.BattleNumber),
				Data:     null.JSONFrom(b),
			}
			err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				sublogger.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
				break
			}
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", msg.PlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
		}
	}(btl)

	sublogger.Debug().Str("correlation_id", "8975f574-51f0-4d8d-a21b-0e36ff229dd5").Msg("broadcast staked mech owner reward system messages")
	// broadcast staked mech owner reward system messages
	go func(battle *Battle) {
		// broadcast end info
		for _, msg := range battle.stakedMechOwnerRewardMessage {
			// get mechs data
			for _, bm := range battleMechs {
				// skip, if player does not have the mech
				index := slices.IndexFunc(msg.MechBattleBriefs, func(mbb *MechBattleBrief) bool { return bm.MechID == mbb.MechID })
				if index == -1 {
					continue
				}

				mbb := msg.MechBattleBriefs[index]
				mbb.FactionID = bm.FactionID

				if idx := slices.IndexFunc(btl.WarMachines, func(wm *WarMachine) bool { return wm.ID == bm.MechID }); idx != -1 {
					wm := btl.WarMachines[idx]
					mbb.Name = wm.Label
					if wm.Name != "" {
						mbb.Name = wm.Name
					}

					mbb.Tier = wm.Tier
					mbb.ImageUrl = wm.ImageAvatar

					for _, destroyedMechRecord := range btl.destroyedWarMachineMap {
						destroyedMech := destroyedMechRecord.DestroyedWarMachine
						killerMech := destroyedMechRecord.KilledByWarMachine
						killerUser := destroyedMechRecord.KilledByUser

						killInfo := &KillInfo{
							Name:      destroyedMechRecord.KilledBy,
							FactionID: destroyedMechRecord.KillerFactionID,
						}

						// if destroyed mech is current mech
						if destroyedMech.Hash == wm.Hash {
							if killerMech != nil {
								killInfo.Name = killerMech.Name
								killInfo.ImageUrl = killerMech.ImageAvatar
							} else if killerUser != nil {
								killInfo.Name = fmt.Sprintf("%s %s", killerUser.Username, destroyedMechRecord.KilledBy)
							}
							mbb.KilledBy = killInfo // set kill by info
							continue
						} else if killerMech != nil && killerMech.Hash == wm.Hash {
							// if current mech is the killer mech

							killInfo.Name = destroyedMech.Name
							killInfo.FactionID = destroyedMech.FactionID
							killInfo.ImageUrl = destroyedMech.ImageAvatar
							mbb.Kills = append(mbb.Kills, killInfo)
							continue
						}

					}
				}

				msg.MechBattleBriefs = append(msg.MechBattleBriefs, mbb)
			}

			// send battle reward system message
			b, err := json.Marshal(msg)
			if err != nil {
				sublogger.Error().Interface("player reward data", msg).Err(err).Msg("Failed to marshal player reward data into json.")
				break
			}
			sysMsg := boiler.SystemMessage{
				PlayerID: msg.PlayerID,
				SenderID: server.SupremacyBattleUserID,
				DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeMechBattleComplete)),
				Title:    "Battle Complete",
				Message:  fmt.Sprintf("Staked mech owner reward from the battle #%d.", battle.BattleNumber),
				Data:     null.JSONFrom(b),
			}
			err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				sublogger.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
				break
			}
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", msg.PlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
		}
	}(btl)

	sublogger.Debug().Str("correlation_id", "3d5f2de1-dcc9-4a99-b05a-936fdc2a6b35").Msg("record staked mechs battle logs")
	// record staked mechs battle logs
	go func(battle *Battle) {
		stakedMechs, err := boiler.StakedMechs(
			boiler.StakedMechWhere.MechID.IN(battle.warMachineIDs),
		).All(gamedb.StdConn)
		if err != nil {
			return
		}

		for _, stakedMech := range stakedMechs {
			smb := boiler.StakedMechBattleLog{
				BattleID:     btl.ID,
				StakedMechID: stakedMech.MechID,
				OwnerID:      stakedMech.OwnerID,
				FactionID:    stakedMech.FactionID,
			}
			err = smb.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				sublogger.Error().Err(err).Interface("staked mech battle log", smb).Msg("Failed to insert staked mech battle record.")
			}
		}
	}(btl)
}

type PlayerBattleCompleteMessage struct {
	PlayerID string `json:"player_id"`

	BattleReward     *BattleReward      `json:"battle_reward,omitempty"`
	MechBattleBriefs []*MechBattleBrief `json:"mech_battle_briefs,omitempty"`
}

type BattleReward struct {
	RewardedSups          decimal.Decimal                `json:"rewarded_sups"`
	RewardedSupsBonus     decimal.Decimal                `json:"rewarded_sups_bonus"`
	RewardedPlayerAbility *boiler.BlueprintPlayerAbility `json:"rewarded_player_ability"`
	FactionRank           string                         `json:"faction_rank"`
}

type MechReward struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Label             string          `json:"label"`
	FactionID         string          `json:"faction_id"`
	AvatarURL         string          `json:"avatar_url"`
	OwnerID           string          `json:"owner_id"`
	RewardedSups      decimal.Decimal `json:"rewarded_sups"`
	RewardedSupsBonus decimal.Decimal `json:"rewarded_sups_bonus"`
	IsAFK             bool            `json:"is_afk"`
}

// RewardBattleMechOwners give reward to war machine owner
func (btl *Battle) RewardBattleMechOwners(winningFactionOrder []string) {
	// load reward from entry fee
	totalSups := btl.lobby.EntryFee.Mul(decimal.NewFromInt(int64(len(btl.warMachineIDs))))

	blms, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.BattleLobbyID.EQ(btl.lobby.ID),
		boiler.BattleLobbiesMechWhere.MechID.IN(btl.warMachineIDs),
		qm.Load(boiler.BattleLobbiesMechRels.QueuedBy),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle lobby id", btl.lobby.ID).Strs("mech id list", btl.warMachineIDs).Msg("Failed to load mechs from battle lobby")
		return
	}

	extraBattleRewards, err := btl.lobby.BattleLobbyExtraSupsRewards(
		boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load extra battle reward.")
	}

	for _, ebr := range extraBattleRewards {
		totalSups = totalSups.Add(ebr.Amount)
	}

	// reward sups
	taxRatio := db.GetDecimalWithDefault(db.KeyBattleRewardTaxRatio, decimal.NewFromFloat(0.025))

	afkMechIDs := btl.AFKChecker()

	for i, factionID := range winningFactionOrder {
		switch i {
		case 0: // winning faction
			for _, blm := range blms {
				if blm.FactionID == factionID && blm.R != nil && blm.R.QueuedBy != nil {
					player := blm.R.QueuedBy
					btl.RewardMechOwner(
						blm.MechID,
						player,
						"FIRST",
						totalSups.Mul(btl.lobby.FirstFactionCut).Div(decimal.NewFromInt(3)),
						taxRatio,
						blm,
						slices.Index(afkMechIDs, blm.MechID) != -1,
						false,
					)
				}
			}

		case 1: // second faction
			for _, blm := range blms {
				if blm.FactionID == factionID && blm.R != nil && blm.R.QueuedBy != nil {
					player := blm.R.QueuedBy
					btl.RewardMechOwner(
						blm.MechID,
						player,
						"SECOND",
						totalSups.Mul(btl.lobby.SecondFactionCut).Div(decimal.NewFromInt(3)),
						taxRatio,
						blm,
						slices.Index(afkMechIDs, blm.MechID) != -1,
						false,
					)
				}
			}

		case 2: // lose faction
			for _, blm := range blms {
				if blm.FactionID == factionID && blm.R != nil && blm.R.QueuedBy != nil {
					player := blm.R.QueuedBy
					btl.RewardMechOwner(
						blm.MechID,
						player,
						"THIRD",
						totalSups.Mul(btl.lobby.ThirdFactionCut).Div(decimal.NewFromInt(3)),
						taxRatio,
						blm,
						slices.Index(afkMechIDs, blm.MechID) != -1,
						true,
					)
				}
			}
		}
	}
}

// AFKChecker return a list of id of the AFK mechs
func (btl *Battle) AFKChecker() []string {
	minimumMechActionCountStrict := db.GetIntWithDefault(db.KeyMinimumMechActionCountStrict, 5)
	minimumMechActionCountMild := db.GetIntWithDefault(db.KeyMinimumMechActionCountMild, 3)
	minimumMechActionCountLoose := db.GetIntWithDefault(db.KeyMinimumMechActionCountLoose, 2)

	// get mech command
	mcs, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.BattleID.EQ(btl.ID),
		qm.OrderBy(boiler.MechMoveCommandLogColumns.CreatedAt),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load mech move command.")
		return []string{}
	}

	bas, err := boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.BattleID.EQ(btl.ID),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle ability triggers.")
		return []string{}
	}

	// collect afk mech id
	afkMechIDs := []string{}
	for _, wm := range btl.WarMachines {

		// record ident
		mechID := wm.ID
		pilotID := wm.OwnedByID

		// count the amount of mechs the player is piloting in the match
		pilotedMechAmount := 0
		for _, mech := range btl.WarMachines {
			if mech.OwnedByID != pilotID {
				continue
			}
			pilotedMechAmount += 1
		}

		// set min action criteria base on the amount of piloted mechs
		minActionCount := minimumMechActionCountStrict
		switch pilotedMechAmount {
		case 1:
			minActionCount = minimumMechActionCountStrict
		case 2:
			minActionCount = minimumMechActionCountMild
		default:
			minActionCount = minimumMechActionCountLoose
		}

		// accumulate the times of actions the player has triggered in the match
		actionCount := 0

		// count the number of mech move command
		for _, mc := range mcs {
			if mc.MechID == mechID {
				actionCount += 1
			}
		}

		// count the number of triggered abilities
		for _, ba := range bas {
			if ba.PlayerID.String == pilotID {
				actionCount += 1
			}
		}

		// append the mech to the AFK list, if it does not meet the criteria
		if actionCount < minActionCount {
			afkMechIDs = append(afkMechIDs, mechID)
		}
	}

	return afkMechIDs
}

func (btl *Battle) RewardMechOwner(
	mechID string,
	owner *boiler.Player,
	ranking string,
	rewardedSups decimal.Decimal,
	taxRatio decimal.Decimal,
	battleLobbiesMech *boiler.BattleLobbiesMech,
	isAFK bool,
	rewardAbility bool,
) {
	// trigger challenge fund update
	defer func() {
		btl.arena.Manager.ChallengeFundUpdateChan <- true
	}()

	// reward staked mech
	rewardedSups = btl.rewardStakedMech(mechID, rewardedSups, taxRatio)

	l := gamelog.L.With().Str("function", "RewardMechOwner").Logger()
	pw := &BattleReward{
		RewardedSups:      rewardedSups,
		RewardedSupsBonus: decimal.Zero,
		FactionRank:       ranking,
	}

	updateCols := []string{}

	// reward sups
	if pw.RewardedSups.GreaterThan(decimal.Zero) {
		tax := rewardedSups.Mul(taxRatio)
		challengeFund := decimal.New(1, 18)

		// if player is AI, pay reward back to treasury fund, and return
		if owner.IsAi {
			payoutTXID, err := btl.arena.Manager.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
				ToUserID:             uuid.UUID(server.XsynTreasuryUserID),
				Amount:               rewardedSups.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("battle_reward|%s|%d", btl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          fmt.Sprintf("reward from battle #%d.", btl.BattleNumber),
			})
			if err != nil {
				l.Error().Err(err).
					Str("from", server.SupremacyBattleUserID).
					Str("to", owner.ID).
					Str("amount", rewardedSups.StringFixed(0)).
					Msg("Failed to pay player battel reward")
			}
			battleLobbiesMech.PayoutTXID = null.StringFrom(payoutTXID)
			updateCols = append(updateCols, boiler.BattleLobbiesMechColumns.PayoutTXID)
		} else if !isAFK {
			// otherwise, pay battle reward to the actual player
			payoutTXID, err := btl.arena.Manager.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
				ToUserID:             uuid.Must(uuid.FromString(owner.ID)),
				Amount:               rewardedSups.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("battle_reward|%s|%d", btl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          fmt.Sprintf("reward from battle #%d.", btl.BattleNumber),
			})
			if err != nil {
				l.Error().Err(err).
					Str("from", server.SupremacyBattleUserID).
					Str("to", owner.ID).
					Str("amount", rewardedSups.StringFixed(0)).
					Msg("Failed to pay player battle reward")
			}
			battleLobbiesMech.PayoutTXID = null.StringFrom(payoutTXID)
			updateCols = append(updateCols, boiler.BattleLobbiesMechColumns.PayoutTXID)

			// pay reward tax
			taxTXID, err := btl.arena.Manager.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(owner.ID)),
				ToUserID:             uuid.FromStringOrNil(server.SupremacyChallengeFundUserID), // NOTE: send fees to challenge fund for now. (was treasury)
				Amount:               tax.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("battle_reward_tax|%s|%d", btl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          fmt.Sprintf("reward tax from battle #%d.", btl.BattleNumber),
			})
			if err != nil {
				l.Error().Err(err).
					Str("from", owner.ID).
					Str("to", server.SupremacyChallengeFundUserID).
					Str("amount", tax.StringFixed(0)).
					Msg("Failed to pay player battle reward")
			}
			battleLobbiesMech.TaxTXID = null.StringFrom(taxTXID)
			updateCols = append(updateCols, boiler.BattleLobbiesMechColumns.TaxTXID)

			// pay challenge fund
			challengeFundTXID, err := btl.arena.Manager.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(owner.ID)),
				ToUserID:             uuid.Must(uuid.FromString(server.SupremacyChallengeFundUserID)),
				Amount:               challengeFund.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("supremacy_challenge_fund|%s|%d", btl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          fmt.Sprintf("challenge fund from battle #%d.", btl.BattleNumber),
			})
			if err != nil {
				l.Error().Err(err).
					Str("from", owner.ID).
					Str("to", server.SupremacyChallengeFundUserID).
					Str("amount", challengeFund.StringFixed(0)).
					Msg("Failed to pay player battle reward")
			}
			battleLobbiesMech.ChallengeFundTXID = null.StringFrom(challengeFundTXID)
			updateCols = append(updateCols, boiler.BattleLobbiesMechColumns.ChallengeFundTXID)
		}
	}

	if len(updateCols) > 0 {
		_, err := battleLobbiesMech.Update(gamedb.StdConn, boil.Whitelist(updateCols...))
		if err != nil {
			l.Error().Err(err).Interface("queue fee", battleLobbiesMech).Msg("Failed to update payout, tax and challenge fund transaction id")
		}
	}

	// record mech reward
	if m := btl.arena.CurrentBattleWarMachineByID(mechID); m != nil {
		btl.mechRewards = append(btl.mechRewards, &MechReward{
			ID:                m.ID,
			FactionID:         m.FactionID,
			Name:              m.Name,
			Label:             m.Label,
			AvatarURL:         m.ImageAvatar,
			RewardedSups:      pw.RewardedSups,
			RewardedSupsBonus: pw.RewardedSupsBonus,
			OwnerID:           owner.ID,
			IsAFK:             !owner.IsAi && isAFK, // if the owner is not AI, and the mech has no move command issued during the battle
		})
	}

	index := slices.IndexFunc(btl.playerBattleCompleteMessage, func(pr *PlayerBattleCompleteMessage) bool { return pr.PlayerID == owner.ID })
	if index == -1 {
		btl.playerBattleCompleteMessage = append(btl.playerBattleCompleteMessage, &PlayerBattleCompleteMessage{
			PlayerID: owner.ID,
		})
		index = len(btl.playerBattleCompleteMessage) - 1
	}

	pbm := btl.playerBattleCompleteMessage[index]
	if pbm.BattleReward == nil {
		pbm.BattleReward = pw
	} else {
		pbm.BattleReward.RewardedSups = pbm.BattleReward.RewardedSups.Add(rewardedSups)
	}

	// skip ability reward, if
	// 1. the player is AI
	// 2. the player is not eligible
	// 3. the player has already got an ability
	if owner.IsAi || !rewardAbility || pbm.BattleReward.RewardedPlayerAbility != nil {
		return
	}

	// start rewarding ability
	availableAbilities, err := boiler.SalePlayerAbilities(
		boiler.SalePlayerAbilityWhere.RarityWeight.GT(0),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.BlueprintPlayerAbilities,
				qm.Rels(boiler.TableNames.BlueprintPlayerAbilities, boiler.BlueprintPlayerAbilityColumns.ID),
				qm.Rels(boiler.TableNames.SalePlayerAbilities, boiler.SalePlayerAbilityColumns.BlueprintID),
			),
		),
		qm.Where(
			fmt.Sprintf(
				"NOT EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s = ? AND %s >= %s)",
				boiler.TableNames.PlayerAbilities,
				qm.Rels(boiler.TableNames.PlayerAbilities, boiler.PlayerAbilityColumns.BlueprintID),
				qm.Rels(boiler.TableNames.SalePlayerAbilities, boiler.SalePlayerAbilityColumns.BlueprintID),
				qm.Rels(boiler.TableNames.PlayerAbilities, boiler.PlayerAbilityColumns.OwnerID),
				qm.Rels(boiler.TableNames.PlayerAbilities, boiler.PlayerAbilityColumns.Count),
				qm.Rels(boiler.TableNames.BlueprintPlayerAbilities, boiler.BlueprintPlayerAbilityColumns.InventoryLimit),
			),
			owner.ID,
		),

		qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to refresh pool of sale abilities from db")
		return
	}

	// skip, if no player ability is available
	if availableAbilities == nil || len(availableAbilities) == 0 {
		sysMsg := boiler.SystemMessage{
			PlayerID: owner.ID,
			SenderID: server.SupremacyBattleUserID,
			DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeMechOwnerBattleReward)),
			Title:    "Battle Reward",
			Message:  "Unable to reward you new player ability due to your inventory is full.",
		}
		err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
			return
		}
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", owner.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)
		return
	}

	// create the pool
	pool := []*boiler.SalePlayerAbility{}
	for _, aa := range availableAbilities {
		for i := 0; i < aa.RarityWeight; i++ {
			pool = append(pool, aa)
		}
	}

	// randomly assign an ability
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	rand.Seed(time.Now().UnixNano())
	ability := availableAbilities[rand.Intn(len(availableAbilities))]

	err = db.PlayerAbilityAssign(pbm.PlayerID, ability.BlueprintID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("player id", owner.ID).Str("ability id", ability.ID).Msg("Failed to assign ability to the player")
		return
	}

	pbm.BattleReward.RewardedPlayerAbility = ability.R.Blueprint
}

// rewardStakedMech staked mech function
func (btl *Battle) rewardStakedMech(mechID string, rewardedSups decimal.Decimal, taxRatio decimal.Decimal) decimal.Decimal {
	if rewardedSups.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	remainSups := rewardedSups

	stakedMechReward := remainSups.Div(decimal.NewFromInt(2))
	tax := stakedMechReward.Mul(taxRatio)

	// reward sups for the owner of staked mech
	index := slices.IndexFunc(btl.WarMachines, func(wm *WarMachine) bool { return wm.ID == mechID })
	if index == -1 {
		return remainSups
	}

	// check if the mech staked mech
	sm, err := boiler.StakedMechs(
		boiler.StakedMechWhere.MechID.EQ(btl.WarMachines[index].ID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Warn().Err(err).Msg("Failed to load staked mech from id")
		return remainSups
	}

	// if staked mech not exists
	if sm == nil {
		return remainSups
	}

	// no reward for player who is the owner of the staked mech
	if sm.OwnerID == btl.WarMachines[index].OwnedByID {
		return remainSups
	}

	// reward the owner of the staked mech
	_, err = btl.arena.Manager.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
		ToUserID:             uuid.Must(uuid.FromString(sm.OwnerID)),
		Amount:               stakedMechReward.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("staked_mech_battle_winning_reward|%s|%s|%d", btl.ID, sm.OwnerID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          fmt.Sprintf("staked mech winning from battle #%d.", btl.BattleNumber),
	})
	if err != nil {
		gamelog.L.Error().Err(err).
			Str("from", server.XsynTreasuryUserID.String()).
			Str("to", sm.OwnerID).
			Str("amount", stakedMechReward.StringFixed(0)).
			Msg("Failed to pay the owner of the staked mech winning battle reward")
		return remainSups
	}

	remainSups = remainSups.Sub(stakedMechReward)

	// tax the reward
	_, err = btl.arena.Manager.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(sm.OwnerID)),
		ToUserID:             uuid.UUID(server.XsynTreasuryUserID),
		Amount:               tax.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("tax_staked_mech_winning_reward|%s|%s|%d", btl.ID, sm.OwnerID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          fmt.Sprintf("reward tax from battle #%d.", btl.BattleNumber),
	})
	if err != nil {
		gamelog.L.Error().Err(err).
			Str("from", sm.OwnerID).
			Str("to", server.XsynTreasuryUserID.String()).
			Str("amount", tax.StringFixed(0)).
			Msg("Failed to pay the owner of the staked mech winning battle reward")
		return remainSups
	}

	index = slices.IndexFunc(btl.stakedMechOwnerRewardMessage, func(pr *PlayerBattleCompleteMessage) bool { return pr.PlayerID == sm.OwnerID })
	if index == -1 {
		btl.stakedMechOwnerRewardMessage = append(btl.stakedMechOwnerRewardMessage, &PlayerBattleCompleteMessage{
			PlayerID:         sm.OwnerID,
			MechBattleBriefs: []*MechBattleBrief{},
		})
		index = len(btl.stakedMechOwnerRewardMessage) - 1
	}

	pbm := btl.stakedMechOwnerRewardMessage[index]
	// add sups
	if pbm.BattleReward == nil {
		pbm.BattleReward = &BattleReward{RewardedSups: stakedMechReward}
	} else {
		pbm.BattleReward.RewardedSups = pbm.BattleReward.RewardedSups.Add(stakedMechReward)
	}

	// append mechs
	pbm.MechBattleBriefs = append(pbm.MechBattleBriefs, &MechBattleBrief{MechID: sm.OwnerID})

	return remainSups
}

func (btl *Battle) processWarMachineRepair() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at register mech repair cases", r)
		}
	}()

	// soft delete all the incomplete repair cases of the mechs
	_, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(btl.warMachineIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.RepairCaseColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Strs("mech id list", btl.warMachineIDs).Err(err).Msg("Failed to delete incomplete repair cases.")
	}

	// generate repair case for damaged war machines
	for _, wm := range btl.WarMachines {
		wm.RLock()
		mechID := wm.ID
		modelID := wm.ModelID
		maxHealth := wm.MaxHealth
		health := wm.Health
		wm.RUnlock()

		go func() {
			ci, err := boiler.CollectionItems(
				boiler.CollectionItemWhere.ItemID.EQ(mechID),
				qm.Load(boiler.CollectionItemRels.Owner),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Err(err).Msg("Failed to load mech owner detail")
				return
			}

			if ci.R == nil || ci.R.Owner == nil || ci.R.Owner.IsAi {
				return
			}

			// register mech repair case
			err = RegisterMechRepairCase(mechID, modelID, maxHealth, health)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to register mech repair")
			}

			// update faction staked mech damaged status
			btl.arena.Manager.FactionStakedMechDashboardKeyChan <- []string{FactionStakedMechDashboardKeyDamaged}
		}()
	}
}

const HubKeyBattleEndDetailUpdated = "BATTLE:END:DETAIL:UPDATED"

func (btl *Battle) end(payload *BattleEndPayload) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the battle end!", r)
		}
	}()

	sublogger := log.With().
		Str("case", "battle_end").
		Str("battle_id", btl.ID).
		Logger()

	// stop mini map ability display list
	sublogger.Debug().Msg("stopping minimap...")
	btl.MiniMapAbilityDisplayList.stop <- true
	sublogger.Debug().Msg("minimap stopped")

	// pre-assign next battle lobby
	sublogger.Debug().Msg("pre-assign next battle lobby")
	btl.arena.beginBattleMux.Lock()
	defer btl.arena.beginBattleMux.Unlock()

	// if the battle is not an AI driven
	if !btl.lobby.IsAiDrivenMatch {
		sublogger.Debug().Msg("non AI match, setup repairs")
		// assigning repair case
		btl.processWarMachineRepair()
	}

	// clean up current battle
	sublogger.Debug().Msg("handle battle end")
	btl.handleBattleEnd(payload)
	sublogger.Info().Msgf("battle has been cleaned up, sending broadcast %s", btl.ID)

	// reactivate idle arenas
	sublogger.Debug().Msg("kick arenas")
	go btl.arena.Manager.KickIdleArenas()

	if !btl.arena._currentBattle.lobby.IsAiDrivenMatch && !btl.arena._currentBattle.lobby.AccessCode.Valid {
		sublogger.Debug().Msg("notify discord")
		go btl.arena.Manager.DiscordSession.SendBattleLobbyEditMessage(btl.arena._currentBattle.lobby.ID, btl.arena.Name)
	}

	sublogger.Debug().Msg("send message to faction stakers")
	btl.arena.Manager.FactionStakedMechDashboardKeyChan <- []string{FactionStakedMechDashboardKeyQueue, FactionStakedMechDashboardKeyMVP}
}

type GameSettingsResponse struct {
	GameMap            *server.GameMap    `json:"game_map"`
	BattleZone         *server.BattleZone `json:"battle_zone"`
	WarMachines        []*WarMachine      `json:"war_machines"`
	SpawnedAI          []*WarMachine      `json:"spawned_ai"`
	WarMachineLocation []byte             `json:"war_machine_location"`
	BattleIdentifier   int                `json:"battle_identifier"`
	BattleID           string             `json:"battle_id"`
	AbilityDetails     []*AbilityDetail   `json:"ability_details"`

	ServerTime      time.Time `json:"server_time"` // time for frontend to adjust the different
	IsAIDrivenMatch bool      `json:"is_ai_driven_match"`
}

func GameSettingsPayload(btl *Battle) *GameSettingsResponse {
	var lt []byte
	if btl.lastTick != nil {
		lt = *btl.lastTick
	}
	if btl == nil {
		return nil
	}

	// Current Battle Zone
	var battleZone *server.BattleZone
	if len(btl.battleZones) > 0 {
		if btl.currentBattleZoneIndex >= len(btl.battleZones) {
			btl.currentBattleZoneIndex = 0
		}
		battleZone = &btl.battleZones[btl.currentBattleZoneIndex]
	}

	wms := []*WarMachine{}
	for _, w := range btl.WarMachines {
		wCopy := &WarMachine{
			ID:                 w.ID,
			Hash:               w.Hash,
			OwnedByID:          w.OwnedByID,
			OwnerUsername:      w.OwnerUsername,
			Name:               w.Name,
			Label:              w.Label,
			ParticipantID:      w.ParticipantID,
			FactionID:          w.FactionID,
			MaxHealth:          w.MaxHealth,
			MaxShield:          w.MaxShield,
			Health:             w.Health,
			AIType:             w.AIType,
			ModelID:            w.ModelID,
			ModelName:          w.ModelName,
			SkinName:           w.SkinName,
			Speed:              w.Speed,
			Faction:            w.Faction,
			Tier:               w.Tier,
			PowerCore:          w.PowerCore,
			Weapons:            w.Weapons,
			Utility:            w.Utility,
			Image:              w.Image,
			ImageAvatar:        w.ImageAvatar,
			Position:           w.Position,
			Rotation:           w.Rotation,
			IsHidden:           w.IsHidden,
			Shield:             w.Shield,
			ShieldRechargeRate: w.ShieldRechargeRate,
			Stats:              w.Stats,
			Status:             w.Status,
		}
		// Hidden/Incognito
		if wCopy.Position != nil {
			hideMech := btl._playerAbilityManager.IsWarMachineHidden(wCopy.Hash)
			hideMech = hideMech || btl._playerAbilityManager.IsWarMachineInBlackout(server.GameLocation{
				X: wCopy.Position.X,
				Y: wCopy.Position.Y,
			})
			if hideMech {
				wCopy.IsHidden = true
				wCopy.Position = &server.Vector3{
					X: -1,
					Y: -1,
					Z: -1,
				}
			}
		}
		wms = append(wms, wCopy)
	}

	ais := []*WarMachine{}
	for _, w := range btl.SpawnedAI {
		wCopy := &WarMachine{
			ID:                 w.ID,
			Hash:               w.Hash,
			OwnedByID:          w.OwnedByID,
			OwnerUsername:      w.OwnerUsername,
			Name:               w.Name,
			Label:              w.Label,
			ParticipantID:      w.ParticipantID,
			FactionID:          w.FactionID,
			MaxHealth:          w.MaxHealth,
			MaxShield:          w.MaxShield,
			Health:             w.Health,
			AIType:             w.AIType,
			ModelID:            w.ModelID,
			ModelName:          w.ModelName,
			SkinName:           w.SkinName,
			Speed:              w.Speed,
			Faction:            w.Faction,
			Tier:               w.Tier,
			PowerCore:          w.PowerCore,
			Weapons:            w.Weapons,
			Utility:            w.Utility,
			Image:              w.Image,
			ImageAvatar:        w.ImageAvatar,
			Position:           w.Position,
			Rotation:           w.Rotation,
			IsHidden:           w.IsHidden,
			Shield:             w.Shield,
			ShieldRechargeRate: w.ShieldRechargeRate,
			Stats:              w.Stats,
			Status:             w.Status,
		}

		// Hidden/Incognito
		if wCopy.Position != nil {
			hideMech := btl._playerAbilityManager.IsWarMachineHidden(wCopy.Hash)
			hideMech = hideMech || btl._playerAbilityManager.IsWarMachineInBlackout(server.GameLocation{
				X: wCopy.Position.X,
				Y: wCopy.Position.Y,
			})
			if hideMech {
				wCopy.IsHidden = true
				wCopy.Position = &server.Vector3{
					X: -1,
					Y: -1,
					Z: -1,
				}
			}
		}
		ais = append(ais, wCopy)
	}

	return &GameSettingsResponse{
		BattleIdentifier:   btl.BattleNumber,
		GameMap:            btl.gameMap,
		BattleZone:         battleZone,
		WarMachines:        wms,
		SpawnedAI:          ais,
		WarMachineLocation: lt,
		AbilityDetails:     btl.abilityDetails,
		ServerTime:         time.Now(),
		IsAIDrivenMatch:    btl.lobby.IsAiDrivenMatch,
		BattleID:           btl.ID,
	}
}

const HubKeyBattleAISpawned = "BATTLE:AI:SPAWNED:SUBSCRIBE"
const HubKeyGameSettingsUpdated = "GAME:SETTINGS:UPDATED"

func (btl *Battle) BroadcastUpdate() {
	ws.PublishMessage(fmt.Sprintf("/public/arena/%s/game_settings", btl.ArenaID), HubKeyGameSettingsUpdated, GameSettingsPayload(btl))
}

func (btl *Battle) Tick(payload []byte) {
	gamelog.L.Trace().Str("func", "Tick").Msg("start")
	defer gamelog.L.Trace().Str("func", "Tick").Msg("end")
	if len(payload) < 1 {
		gamelog.L.Error().Str("log_name", "battle arena").Err(fmt.Errorf("len(payload) < 1")).Interface("payload", payload).Msg("len(payload) < 1")
		return
	}

	if btl.state.Load() != BattlingState {
		return
	}

	btl.lastTick = &payload

	// return if the war machines list is not ready
	if len(btl.WarMachines) == 0 {
		return
	}
	// return, if any war machines have 0 as their participant id
	if btl.WarMachines[0].ParticipantID == 0 {
		return
	}

	// collect war machine stat
	var wmss []*WarMachineStat

	// Update game settings (so new players get the latest position, health and shield of all warmachines)
	count := int(payload[1])
	offset := 2
	for c := 0; c < count; c++ {
		participantID := payload[offset]
		offset++

		// Get Sync byte (tells us which data was updated for this warmachine)
		syncByte := payload[offset]
		booleans := helpers.UnpackBooleansFromByte(syncByte)
		offset++

		// Get Warmachine Index
		warMachineIndex := -1
		var warmachine *WarMachine
		if participantID > 100 {
			// find Spawned AI
			btl.spawnedAIMux.RLock()
			for i, wmn := range btl.SpawnedAI {
				if checkWarMachineByParticipantID(wmn, int(participantID)) {
					warMachineIndex = i
					break
				}
			}
			btl.spawnedAIMux.RUnlock()

			if warMachineIndex == -1 {
				gamelog.L.Warn().Err(fmt.Errorf("aiSpawnedIndex == -1")).
					Str("participantID", fmt.Sprintf("%d", participantID)).Msg("unable to find warmachine participant ID for Spawned AI")

				tickSkipToWarmachineEnd(payload, &offset, booleans)
				continue
			}
			warmachine = btl.SpawnedAI[warMachineIndex]
		} else {
			// Mech
			for i, wmn := range btl.WarMachines {
				if checkWarMachineByParticipantID(wmn, int(participantID)) {
					warMachineIndex = i
					break
				}
			}
			if warMachineIndex == -1 {
				gamelog.L.Warn().Err(fmt.Errorf("warMachineIndex == -1")).
					Str("participantID", fmt.Sprintf("%d", participantID)).Msg("unable to find warmachine participant ID war machine - returning")

				tickSkipToWarmachineEnd(payload, &offset, booleans)
				continue
			}
			warmachine = btl.WarMachines[warMachineIndex]
		}

		// Get Current Mech State
		warmachine.Lock()
		wms := &WarMachineStat{
			ParticipantID: int(warmachine.ParticipantID),
			Position:      warmachine.Position,
			Rotation:      warmachine.Rotation,
			Health:        warmachine.Health,
			Shield:        warmachine.Shield,
			IsHidden:      false,
		}

		if wms.Position == nil {
			wms.Position = &server.Vector3{}
		}

		// Position + Yaw
		if booleans[0] {
			x := int(helpers.BytesToInt(payload[offset : offset+4]))
			offset += 4
			y := int(helpers.BytesToInt(payload[offset : offset+4]))
			offset += 4
			rotation := int(helpers.BytesToInt(payload[offset : offset+4]))
			offset += 4

			if warmachine.Position == nil {
				warmachine.Position = &server.Vector3{}
			}
			warmachine.Position.X = x
			warmachine.Position.Y = y
			wms.Position = warmachine.Position
			warmachine.Rotation = rotation
			wms.Rotation = rotation
		}
		// Health
		if booleans[1] {
			health := binary.BigEndian.Uint32(payload[offset : offset+4])
			offset += 4
			warmachine.Health = health
			wms.Health = health
		}
		// Shield
		if booleans[2] {
			shield := binary.BigEndian.Uint32(payload[offset : offset+4])
			offset += 4
			warmachine.Shield = shield
			wms.Shield = shield
		}

		// Energy
		if booleans[3] {
			offset += 4
		}

		// Weapon Ammo
		if booleans[4] {
			updatedWeapons := int(payload[offset])
			offset += 1

			for i := 0; i < updatedWeapons; i++ {
				weaponIndex := int(payload[offset])
				offset += 1

				currentAmmo := binary.BigEndian.Uint32(payload[offset : offset+4])
				offset += 4

				for w := range warmachine.Weapons {
					if warmachine.Weapons[w].SocketIndex == weaponIndex {
						warmachine.Weapons[w].CurrentAmmo = int(currentAmmo)
						break
					}
				}
			}
		}

		if booleans[5] {
			weaponSystemCurrentPower := helpers.BytesToFloat(payload[offset : offset+4])
			shieldSystemCurrentPower := helpers.BytesToFloat(payload[offset+4 : offset+8])
			movementSystemCurrentPower := helpers.BytesToFloat(payload[offset+8 : offset+12])
			offset += 12

			warmachine.PowerCore.WeaponSystemCurrentPower = weaponSystemCurrentPower
			warmachine.PowerCore.ShieldSystemCurrentPower = shieldSystemCurrentPower
			warmachine.PowerCore.MovementSystemCurrentPower = movementSystemCurrentPower
		}

		warmachine.Unlock()

		// Hidden/Incognito
		if btl.playerAbilityManager().IsWarMachineHidden(warmachine.Hash) {
			wms.IsHidden = true
			wms.Position = &server.Vector3{
				X: -1,
				Y: -1,
				Z: -1,
			}
		} else if btl.playerAbilityManager().IsWarMachineInBlackout(server.GameLocation{
			X: wms.Position.X,
			Y: wms.Position.Y,
		}) {
			wms.IsHidden = true
			wms.Position = &server.Vector3{
				X: -1,
				Y: -1,
				Z: -1,
			}
		}

		// If Mech is a regular type OR is a mini mech
		if participantID < 100 || btl.IsMechOfType(int(participantID), MiniMech) {
			wmss = append(wmss, wms)
		}
	}

	if len(wmss) > 0 {
		select {
		case btl.arena.WarMachineStatBroadcastChan <- wmss:
		default: // skip, if the channel is full
		}
	}

	if btl.playerAbilityManager().HasBlackoutsUpdated() {
		var minimapUpdates []MinimapEvent
		for id, b := range btl.playerAbilityManager().Blackouts() {
			minimapUpdates = append(minimapUpdates, MinimapEvent{
				ID:            id,
				GameAbilityID: BlackoutGameAbilityID,
				Duration:      BlackoutDurationSeconds,
				Radius:        int(BlackoutRadius),
				Coords:        b.CellCoords,
			})
		}

		btl.playerAbilityManager().ResetHasBlackoutsUpdated()

		ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/public/minimap", btl.ArenaID), server.HubKeyMiniMapUpdateSubscribe, minimapUpdates)
	}

	// Map Events
	if len(payload) > offset {
		mapEventCount := int(payload[offset])
		if mapEventCount > 0 {
			// Pass map events straight to frontend clients
			mapEvents := payload[offset:]
			ws.PublishBytes(fmt.Sprintf("/mini_map/arena/%s/public/minimap_events", btl.ArenaID), server.BinaryKeyMiniMapEvents, mapEvents)

			// Unpack and save static events for sending to newly joined frontend clients (ie: landmine, pickup locations and the hive status)
			//btl.MapEventList.MapEventsUnpack(mapEvents)
		}
	}
}

func (arena *Arena) warMachinePositionBroadcaster() {

	// broadcast war machine stat every 250 millisecond
	ticker := time.NewTicker(330 * time.Millisecond)
	var warMachineStats []*WarMachineStat

	exitChan := make(chan bool, 2)
	l := deadlock.RWMutex{}

	go func() {
		for {
			select {
			case stats := <-arena.WarMachineStatBroadcastChan:
				l.Lock()
				// update war machine stats
				for _, stat := range stats {
					index := slices.IndexFunc(warMachineStats, func(wms *WarMachineStat) bool {
						return wms.ParticipantID == stat.ParticipantID
					})

					// append, if not exists
					if index == -1 {
						warMachineStats = append(warMachineStats, stat)
						continue
					}

					// replace, if exits
					warMachineStats[index] = stat
				}
				l.Unlock()

				// trigger everytime when a battle is ended
			case <-arena.WarMachineStatBroadcastResetChan:
				l.Lock()
				warMachineStats = []*WarMachineStat{}
				l.Unlock()
			case <-exitChan:
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			l.RLock()
			if warMachineStats == nil || len(warMachineStats) == 0 {
				l.RUnlock()
				continue
			}

			// otherwise broadcast current data
			ws.PublishBytes(fmt.Sprintf("/mini_map/arena/%s/public/mech_stats", arena.ID), server.BinaryKeyWarMachineStats, PackWarMachineStatsInBytes(warMachineStats))
			l.RUnlock()

			// triggered when arena is disconnected
		case <-arena.WarMachineStatBroadcastStopChan:
			exitChan <- true
			return
		}
	}

}

func PackWarMachineStatsInBytes(warMachineStats []*WarMachineStat) []byte {
	payload := []byte{byte(len(warMachineStats))}

	// repack current data
	for _, wms := range warMachineStats {
		// push participant id into the array
		payload = append(payload, byte(wms.ParticipantID))

		// push location x into the array
		payload = append(payload, helpers.IntToBytes(int32(wms.Position.X))...)

		// push location x into the array
		payload = append(payload, helpers.IntToBytes(int32(wms.Position.Y))...)

		// push location x into the array
		payload = append(payload, helpers.IntToBytes(int32(wms.Rotation))...)

		// push current health into the array
		health := make([]byte, 4)
		binary.BigEndian.PutUint32(health, wms.Health)
		payload = append(payload, health...)

		// push current shield into the array
		shield := make([]byte, 4)
		binary.BigEndian.PutUint32(shield, wms.Shield)
		payload = append(payload, shield...)

		// is hidden
		if wms.IsHidden {
			payload = append(payload, byte(1))
			continue
		}

		// not hidden
		payload = append(payload, byte(0))
	}

	return payload
}

func tickSkipToWarmachineEnd(payload []byte, offset *int, booleans []bool) {
	if booleans[0] {
		*offset += 12
	}
	if booleans[1] {
		*offset += 4
	}
	if booleans[2] {
		*offset += 4
	}
	if booleans[3] {
		*offset += 4
	}
	if booleans[4] {
		*offset += 4
		updatedWeapons := int(payload[*offset])
		*offset += 1
		for i := 0; i < updatedWeapons; i++ {
			*offset += 5
		}
	}
	if booleans[5] {
		*offset += 12
	}
}

func (arena *Arena) reset() {
	gamelog.L.Warn().Msg("arena state resetting")
}

func (btl *Battle) Destroyed(dp *BattleWMDestroyedPayload) {
	gamelog.L.Trace().Str("func", "Destroyed").Msg("start")
	defer gamelog.L.Trace().Str("func", "Destroyed").Msg("end")

	// check destroyed war machine exist
	if btl.ID != dp.BattleID {
		gamelog.L.Warn().Str("battle.ID", btl.ID).Str("gameclient.ID", dp.BattleID).Msg("battle state does not match game client state")
		btl.arena.reset()
		return
	}

	var destroyedWarMachine *WarMachine
	dHash := dp.DestroyedWarMachineHash
	for i, wm := range btl.WarMachines {
		if wm.Hash == dHash {
			// set health to 0
			btl.WarMachines[i].Health = 0
			destroyedWarMachine = wm

			err := db.FactionAddDeathCount(wm.FactionID)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("faction_id", wm.FactionID).Err(err).Msg("failed to update faction death count")
			}
			break
		}
	}
	for _, aiwm := range btl.SpawnedAI {
		if aiwm.Hash == dHash {
			destroyedWarMachine = aiwm
		}
	}
	if destroyedWarMachine == nil {
		gamelog.L.Warn().Str("hash", dHash).Msg("can't match destroyed mech with battle state")
		return
	}

	isAI := destroyedWarMachine.AIType != nil
	if !isAI {
		prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(destroyedWarMachine.OwnedByID)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Str("destroyedWarMachine.ID", destroyedWarMachine.ID).Err(err).Msg("failed to get player preferences")
		}

		if prefs != nil && prefs.TelegramID.Valid && prefs.EnableTelegramNotifications {
			// killed a war machine
			msg := fmt.Sprintf("Your War machine %s has been destroyed ☠️", destroyedWarMachine.Name)
			err := btl.arena.Manager.telegram.Notify(prefs.TelegramID.Int64, msg)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("playerID", prefs.PlayerID).Str("telegramID", fmt.Sprintf("%v", prefs.TelegramID)).Err(err).Msg("failed to send notification")
			}
		}

		var killedByUser *UserBrief
		var killByWarMachine *WarMachine
		if dp.KilledByWarMachineHash != "" {
			for _, wm := range btl.WarMachines {
				if wm.Hash == dp.KilledByWarMachineHash {
					killByWarMachine = wm
					// update user kill
					if wm.OwnedByID != "" {
						_, err := db.UserStatAddMechKill(wm.OwnedByID)
						if err != nil {
							gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", wm.OwnedByID).Err(err).Msg("Failed to update user mech kill count")
						}

						// add faction kill count
						err = db.FactionAddMechKillCount(killByWarMachine.FactionID)
						if err != nil {
							gamelog.L.Error().Str("log_name", "battle arena").Str("faction_id", killByWarMachine.FactionID).Err(err).Msg("failed to update faction mech kill count")
						}

						prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(wm.OwnedByID)).One(gamedb.StdConn)
						if err != nil && !errors.Is(err, sql.ErrNoRows) {
							gamelog.L.Error().Str("log_name", "battle arena").Str("wm.ID", wm.ID).Err(err).Msg("failed to get player preferences")

						}

						if prefs != nil && prefs.TelegramID.Valid && prefs.EnableTelegramNotifications {
							// killed a war machine
							msg := fmt.Sprintf("Your War machine destroyed %s \U0001F9BE ", destroyedWarMachine.Name)
							err := btl.arena.Manager.telegram.Notify(prefs.TelegramID.Int64, msg)
							if err != nil {
								gamelog.L.Error().Str("log_name", "battle arena").Str("playerID", prefs.PlayerID).Str("telegramID", fmt.Sprintf("%v", prefs.TelegramID)).Err(err).Msg("failed to send notification")
							}
						}
					}
					break
				}
			}
		} else if dp.RelatedEventIDString != "" {
			// check related event id
			var abl *boiler.BattleAbilityTrigger
			var err error
			retAbl := func() (*boiler.BattleAbilityTrigger, error) {
				abl, err := boiler.BattleAbilityTriggers(boiler.BattleAbilityTriggerWhere.AbilityOfferingID.EQ(dp.RelatedEventIDString)).One(gamedb.StdConn)
				return abl, err
			}

			retries := 0
			for abl == nil {
				abl, err = retAbl()
				if errors.Is(err, sql.ErrNoRows) {
					if retries >= 5 {
						break
					}
					retries++
					time.Sleep(1 * time.Second)
					continue
				} else if err != nil {
					break
				}
			}

			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("related event id", dp.RelatedEventIDString).Err(err).Msg("Failed get ability from offering id")
			}
			// get ability via offering id

			if abl != nil && abl.PlayerID.Valid {
				currentUser, err := BuildUserDetailWithFaction(uuid.FromStringOrNil(abl.PlayerID.String))
				if err == nil {
					// update kill by user and killed by information
					killedByUser = currentUser
					dp.KilledBy = fmt.Sprintf("(%s)", abl.AbilityLabel)
				}

				// update player ability kills and faction kills
				if strings.EqualFold(destroyedWarMachine.FactionID, abl.FactionID) {
					// load game ability
					ga, err := abl.GameAbility().One(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Str("game ability id", abl.GameAbilityID).Err(err).Msg("Failed to load game ability.")
					}

					// only check team kill, if needed.
					if ga != nil && ga.ShouldCheckTeamKill {
						// if ability ignore self kill or the kill is not self kill
						if !ga.IgnoreSelfKill || abl.PlayerID.String != destroyedWarMachine.OwnedByID {
							// update user kill
							_, err := db.UserStatSubtractAbilityKill(abl.PlayerID.String)
							if err != nil {
								gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to subtract user ability kill count")
							}

							// insert a team kill record to last seven days kills
							pkl := boiler.PlayerKillLog{
								PlayerID:          abl.PlayerID.String,
								FactionID:         abl.FactionID,
								BattleID:          btl.ID,
								IsTeamKill:        true,
								AbilityOfferingID: null.StringFrom(dp.RelatedEventIDString),
								GameAbilityID:     null.StringFrom(abl.GameAbilityID),
							}
							err = pkl.Insert(gamedb.StdConn, boil.Infer())
							if err != nil {
								gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to insert player last seven days kill record- (TEAM KILL)")
							}

							// subtract faction kill count
							err = db.FactionSubtractAbilityKillCount(abl.FactionID)
							if err != nil {
								gamelog.L.Error().Str("log_name", "battle arena").Str("faction_id", abl.FactionID).Err(err).Msg("Failed to subtract user ability kill count")
							}

							// sent instance to system ban manager
							go btl.arena.Manager.SystemBanManager.SendToTeamKillCourtroom(abl.PlayerID.String, dp.RelatedEventIDString)

						}
					}
				} else {
					// update user kill
					_, err := db.UserStatAddAbilityKill(abl.PlayerID.String)
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to add user ability kill count")
					}

					// insert a team kill record to last seven days kills
					pkl := boiler.PlayerKillLog{
						PlayerID:          abl.PlayerID.String,
						FactionID:         abl.FactionID,
						BattleID:          btl.ID,
						AbilityOfferingID: null.StringFrom(dp.RelatedEventIDString),
						GameAbilityID:     null.StringFrom(abl.GameAbilityID),
						IsVerified:        true,
					}
					err = pkl.Insert(gamedb.StdConn, boil.Infer())
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to insert player last seven days kill record- (ABILITY KILL)")
					}

					// add faction kill count
					err = db.FactionAddAbilityKillCount(abl.FactionID)
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Str("faction_id", abl.FactionID).Err(err).Msg("Failed to add faction ability kill count")
					}
				}

				// broadcast player stat to the player
				us, err := db.UserStatsGet(currentUser.ID.String())
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to get player current stat")
				}
				if us != nil {
					ws.PublishMessage(fmt.Sprintf("/secure/user/%s/stat", us.ID), server.HubKeyUserStatSubscribe, us)
				}
			}

		}

		gamelog.L.Debug().Msgf("battle Update: %s - War Machine Destroyed: %s", btl.ID, dHash)

		var warMachineID uuid.UUID
		var killByWarMachineID uuid.UUID
		ids, err := db.MechIDsFromHash(destroyedWarMachine.Hash, dp.KilledByWarMachineHash)

		if err != nil || len(ids) == 0 {
			gamelog.L.Warn().
				Str("hashes", fmt.Sprintf("%s, %s", destroyedWarMachine.Hash, dp.KilledByWarMachineHash)).
				Str("battle_id", btl.ID).
				Err(err).
				Msg("can't retrieve mech ids")

		} else {
			warMachineID = ids[0]
			if len(ids) > 1 {
				killByWarMachineID = ids[1]
			}

			bh := &boiler.BattleHistory{
				BattleID:        btl.ID,
				WarMachineOneID: warMachineID.String(),
				EventType:       db.Btlevnt_Killed.String(),
			}

			// record killer war machine if exists
			if !killByWarMachineID.IsNil() {
				bh.WarMachineTwoID = null.StringFrom(killByWarMachineID.String())
			}

			if dp.RelatedEventIDString != "" {
				bh.BattleAbilityOfferingID = null.StringFrom(dp.RelatedEventIDString)
			}

			err = bh.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().
					Interface("event_data", bh).
					Str("battle_id", btl.ID).
					Err(err).
					Msg("unable to store mech event data")
			}

			// check player obtain mech kill quest
			if killByWarMachine != nil {
				btl.arena.Manager.QuestManager.MechKillQuestCheck(killByWarMachine.OwnedByID)
			}

			// check player obtain ability kill quest, if it is not a team kill
			if killedByUser != nil && destroyedWarMachine.FactionID != killedByUser.FactionID {
				// check player quest reward
				btl.arena.Manager.QuestManager.AbilityKillQuestCheck(killedByUser.ID.String())
			}
		}

		_, err = db.UpdateKilledBattleMech(btl.ID, warMachineID, destroyedWarMachine.OwnedByID, destroyedWarMachine.FactionID, killByWarMachineID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Err(err).
				Str("battle_id", btl.ID).
				Interface("mech_id", warMachineID).
				Bool("killed", true).
				Msg("can't update battle mech")
			gamelog.L.Trace().Str("func", "Destroyed").Msg("end")
			return
		}

		// calc total damage and merge the duplicated damage source
		totalDamage := 0
		newDamageHistory := []*DamageHistory{}
		for _, damage := range dp.DamageHistory {
			damageAmount := int(damage.Amount.IntPart())

			totalDamage += damageAmount
			// check instigator token id exist in the list
			if damage.InstigatorHash != "" {
				exists := false
				for _, hist := range newDamageHistory {
					if hist.InstigatorHash == damage.InstigatorHash {
						hist.Amount += damageAmount
						exists = true
						break
					}
				}
				if !exists {
					newDamageHistory = append(newDamageHistory, &DamageHistory{
						Amount:         damageAmount,
						InstigatorHash: damage.InstigatorHash,
						SourceName:     damage.SourceName,
						SourceHash:     damage.SourceHash,
					})
				}
				continue
			}
			// check source name
			exists := false
			for _, hist := range newDamageHistory {
				if hist.SourceName == damage.SourceName {
					hist.Amount += damageAmount
					exists = true
					break
				}
			}
			if !exists {
				newDamageHistory = append(newDamageHistory, &DamageHistory{
					Amount:         damageAmount,
					InstigatorHash: damage.InstigatorHash,
					SourceName:     damage.SourceName,
					SourceHash:     damage.SourceHash,
				})
			}
		}

		// initial destroy record
		wmd := &WMDestroyedRecord{
			KilledBy: dp.KilledBy,
		}

		// set destroyed war machine
		wmd.DestroyedWarMachine = &WarMachineBrief{
			ParticipantID: destroyedWarMachine.ParticipantID,
			ImageUrl:      destroyedWarMachine.Image,
			ImageAvatar:   destroyedWarMachine.ImageAvatar,
			Name:          destroyedWarMachine.Label,
			Hash:          destroyedWarMachine.Hash,
			FactionID:     destroyedWarMachine.FactionID,
		}
		if destroyedWarMachine.Name != "" {
			wmd.DestroyedWarMachine.Name = destroyedWarMachine.Name
		}

		//
		if killByWarMachine != nil {
			wmd.KillerFactionID = killByWarMachine.FactionID
		} else if killedByUser != nil {
			wmd.KilledByUser = killedByUser
			wmd.KillerFactionID = killedByUser.FactionID
		}

		// get total damage amount for calculating percentage
		for _, damage := range newDamageHistory {
			damageRecord := &DamageRecord{
				SourceName: damage.SourceName,
				Amount:     (damage.Amount * 1000000 / totalDamage) / 100,
			}
			if damage.InstigatorHash != "" {
				for _, wm := range btl.WarMachines {
					if wm.Hash == damage.InstigatorHash {
						wmb := &WarMachineBrief{
							ParticipantID: wm.ParticipantID,
							ImageUrl:      wm.Image,
							ImageAvatar:   wm.ImageAvatar,
							Name:          wm.Label,
							Hash:          wm.Hash,
							FactionID:     wm.FactionID,
						}
						if wm.Name != "" {
							wmb.Name = wm.Name
						}

						damageRecord.CausedByWarMachine = wmb

					}
				}
			}
			wmd.DamageRecords = append(wmd.DamageRecords, damageRecord)
		}

		if killByWarMachine != nil {
			wmd.KilledByWarMachine = &WarMachineBrief{
				ParticipantID: killByWarMachine.ParticipantID,
				ImageUrl:      killByWarMachine.Image,
				ImageAvatar:   killByWarMachine.ImageAvatar,
				Name:          killByWarMachine.Label,
				Hash:          killByWarMachine.Hash,
				FactionID:     killByWarMachine.FactionID,
			}
			if killByWarMachine.Name != "" {
				wmd.KilledByWarMachine.Name = killByWarMachine.Name
			}
		}

		// cache destroyed war machine
		btl.destroyedWarMachineMap[destroyedWarMachine.ID] = wmd

		// check the "?" show up in killed by
		if wmd.KilledBy == "?" {
			// check whether there is a battle ability in the damage records
			for _, dr := range wmd.DamageRecords {
				if strings.ToLower(dr.SourceName) == "nuke" || strings.ToLower(dr.SourceName) == "airstrike" {
					wmd.KilledBy = dr.SourceName
					break
				}
			}
		}

		// broadcast notification
		btl.arena.BroadcastGameNotificationWarMachineDestroyed(&WarMachineDestroyedEventRecord{
			DestroyedWarMachine: wmd.DestroyedWarMachine,
			KilledByWarMachine:  wmd.KilledByWarMachine,
			KilledByUser:        killedByUser,
			KilledBy:            wmd.KilledBy,
		})

		// clear up unfinished mech move command of the destroyed mech
		_, err = boiler.MechMoveCommandLogs(
			boiler.MechMoveCommandLogWhere.MechID.EQ(destroyedWarMachine.ID),
			boiler.MechMoveCommandLogWhere.BattleID.EQ(btl.ID),
		).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechMoveCommandLogColumns.CancelledAt: null.TimeFrom(time.Now())})
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech id", destroyedWarMachine.ID).Str("battle id", btl.ID).Err(err).Msg("Failed to clean up mech move command.")
		}
	}

	// tell frontend to cancel mech move command
	mmc := &MechMoveCommandResponse{
		MechMoveCommandLog: &boiler.MechMoveCommandLog{
			ID:          btl.ID + destroyedWarMachine.Hash,
			BattleID:    btl.ID,
			MechID:      destroyedWarMachine.ID,
			CancelledAt: null.TimeFrom(time.Now()),
		},
		IsMiniMech: false,
	}

	fmc := &FactionMechCommand{
		ID:         mmc.ID,
		BattleID:   btl.ID,
		IsMiniMech: false,
		IsEnded:    true,
	}
	if isAI && *destroyedWarMachine.AIType == MiniMech {
		btl.arena._currentBattle.playerAbilityManager().DeleteMiniMechMove(destroyedWarMachine.Hash)
		fmc.IsMiniMech = true

		// tell frontend to cancel mech move command
		mmc = &MechMoveCommandResponse{
			MechMoveCommandLog: &boiler.MechMoveCommandLog{
				ID:          btl.ID + destroyedWarMachine.Hash,
				BattleID:    btl.ID,
				CancelledAt: null.TimeFrom(time.Now()),
			},
			IsMiniMech: false,
		}
	}

	// broadcast faction mech commands
	ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/faction/%s/mech_command/%s", btl.ArenaID, destroyedWarMachine.FactionID, destroyedWarMachine.Hash), server.HubKeyMechCommandUpdateSubscribe, mmc)
	ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/faction/%s/mech_commands", btl.ArenaID, destroyedWarMachine.FactionID), server.HubKeyFactionMechCommandUpdateSubscribe, []*FactionMechCommand{fmc})
}

func (btl *Battle) Load(battleLobby *boiler.BattleLobby) error {
	lms, err := battleLobby.BattleLobbiesMechs().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle lobby id", battleLobby.ID).Msg("Failed to load mechs from battle lobby")
		return terror.Error(err, "Failed to load mech from battle lobby")
	}

	btl.warMachineIDs = []string{}

	involvedPlayerIDs := []string{}
	// insert battle mechs
	for _, blm := range lms {
		bmd := boiler.BattleMech{
			BattleID:    btl.ID,
			MechID:      blm.MechID,
			PilotedByID: blm.QueuedByID,
			FactionID:   blm.FactionID,
		}
		err = bmd.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Interface("battle mech", bmd).Str("db func", "Battle").Err(err).Msg("unable to insert battle Mech into database")
			return err
		}

		btl.warMachineIDs = append(btl.warMachineIDs, blm.MechID)
		involvedPlayerIDs = append(involvedPlayerIDs, blm.QueuedByID)
	}

	mechs, err := db.Mechs(btl.warMachineIDs...)
	if err != nil {
		gamelog.L.Error().Strs("mech ids", btl.warMachineIDs).Err(err).Msg("Failed to load mech detail")
		return terror.Error(err, "Failed to load mech details")
	}

	// override owner with the players who queue
	ps, err := boiler.Players(
		boiler.PlayerWhere.ID.IN(involvedPlayerIDs),
		qm.Load(boiler.PlayerRels.Faction),
		qm.Load(qm.Rels(boiler.PlayerRels.Faction, boiler.FactionRels.FactionPalette)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Strs("player id list", involvedPlayerIDs).Err(err).Msg("Failed to players")
		return terror.Error(err, "Failed to load players")
	}

	for _, mech := range mechs {
		index := slices.IndexFunc(lms, func(lm *boiler.BattleLobbiesMech) bool { return lm.MechID == mech.ID })
		if index == -1 {
			continue
		}
		queuedByID := lms[index].QueuedByID

		// get player
		index = slices.IndexFunc(ps, func(p *boiler.Player) bool { return p.ID == queuedByID })
		if index == -1 {
			continue
		}

		queuedBy := ps[index]
		mech.OwnerID = queuedByID
		mech.Owner = &server.User{
			Username: queuedBy.Username.String,
			Gid:      queuedBy.Gid,
		}

		// overwrite the faction of the mech with the faction of the player who queue it
		mech.FactionID = queuedBy.FactionID
		mech.Faction.SetFromBoilerFaction(queuedBy.R.Faction)
	}

	btl.WarMachines = btl.MechsToWarMachines(mechs)

	// set mechs current health
	rcs, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(btl.warMachineIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load mech repair cases.")
	}

	for _, rc := range rcs {
		for _, wm := range btl.WarMachines {
			if rc.MechID == wm.ID {
				totalBlocks := db.TotalRepairBlocks(rc.MechID)
				wm.Health = wm.MaxHealth * uint32(totalBlocks-(rc.BlocksRequiredRepair-rc.BlocksRepaired)) / uint32(totalBlocks)
				wm.damagedBlockCount = rc.BlocksRequiredRepair - rc.BlocksRepaired
				break
			}
		}
	}

	gamelog.L.Trace().Str("func", "Load").Msg("end")
	return nil
}

var SubmodelSkinMap = map[string]string{
	"Crystal Blue":       "CrystalBlue",
	"Rust Bucket":        "RustBucket",
	"Dune":               "Dune",
	"Dynamic Yellow":     "DynamicYellow",
	"Molten":             "Molten",
	"Mystermech":         "MysterMech",
	"Nebula":             "Nebula",
	"Sleek":              "Sleek",
	"Blue White":         "BlueWhite",
	"BioHazard":          "BioHazard",
	"Cyber":              "Cyber",
	"Light Blue Police":  "LightBluePolice",
	"Vintage":            "Vintage",
	"Red White":          "RedWhite",
	"Red Hex":            "RedHex",
	"Desert":             "Desert",
	"Navy":               "Navy",
	"Nautical":           "Nautical",
	"Military":           "Military",
	"Irradiated":         "Irradiated",
	"Evo":                "EVA-02",
	"Beetle":             "Beetle",
	"Villain":            "Villain",
	"Green Yellow":       "GreenYellow",
	"Red Blue":           "RedBlue",
	"White Gold":         "WhiteGold",
	"Vector":             "Vector",
	"Cherry Blossom":     "CherryBlossom",
	"Warden":             "Warden",
	"Gumdan":             "Gundam",
	"White Gold Pattern": "WhiteGoldPattern",
	"Evangelic":          "Evangelion",
	"Evangelica":         "Evangelion",
	"Chalky Neon":        "ChalkyNeon",
	"Black Digi":         "BlackDigi",
	"Purple Haze":        "PurpleHaze",
	"Destroyer":          "Destroyer",
	"Static":             "Static",
	"Neon":               "Neon",
	"Gold":               "Gold",
	"Slava Ukraini":      "Ukraine",
	"Ukraine":            "Ukraine",
}

func (btl *Battle) MechsToWarMachines(mechs []*server.Mech) []*WarMachine {
	var warMachines []*WarMachine
	for _, mech := range mechs {
		if !mech.FactionID.Valid {
			gamelog.L.Error().Str("log_name", "battle arena").Err(fmt.Errorf("mech without a faction"))
		}
		newWarMachine := &WarMachine{
			ID:                      mech.ID,
			Hash:                    mech.Hash,
			OwnedByID:               mech.OwnerID,
			Name:                    TruncateString(mech.Name, 20),
			Label:                   mech.Label,
			FactionID:               mech.FactionID.String,
			MaxHealth:               uint32(mech.BoostedMaxHitpoints),
			Health:                  uint32(mech.BoostedMaxHitpoints),
			Speed:                   mech.BoostedSpeed,
			Tier:                    mech.Tier,
			Image:                   mech.ImageURL.String,
			ImageAvatar:             mech.AvatarURL.String,
			Shield:                  uint32(mech.Shield),
			MaxShield:               uint32(mech.Shield),
			ShieldRechargeRate:      uint32(mech.BoostedShieldRechargeRate),
			ShieldRechargeDelay:     mech.ShieldRechargeDelay.InexactFloat64(),
			ShieldRechargePowerCost: uint32(mech.ShieldRechargePowerCost),
			ShieldTypeID:            mech.ShieldTypeID,
			ShieldTypeLabel:         mech.ShieldTypeLabel,
			ShieldTypeDescription:   mech.ShieldTypeDescription,
			HeightMeters:            mech.HeightMeters.InexactFloat64(),

			Faction: &Faction{
				ID:    mech.Faction.ID,
				Label: mech.Faction.Label,
				Theme: &Theme{
					PrimaryColor:    mech.Faction.PrimaryColor,
					SecondaryColor:  mech.Faction.SecondaryColor,
					BackgroundColor: mech.Faction.BackgroundColor,
				},
			},

			PowerCore: PowerCoreFromServer(mech.PowerCore),
			Weapons:   WeaponsFromServer(mech.Weapons, mech.BlueprintWeaponIDsWithSkinInheritance, mech.ChassisSkin.BlueprintWeaponSkinID, mech.InheritAllWeaponSkins),
			Utility:   UtilitiesFromServer(mech.Utility),
			Stats: &Stats{
				TotalWins:       mech.Stats.TotalWins,
				TotalDeaths:     mech.Stats.TotalDeaths,
				TotalKills:      mech.Stats.TotalKills,
				BattlesSurvived: mech.Stats.BattlesSurvived,
				TotalLosses:     mech.Stats.TotalLosses,
			},
			Status:   &Status{},
			Position: &server.Vector3{},
		}

		// add owner username
		if mech.Owner != nil {
			newWarMachine.OwnerUsername = fmt.Sprintf("%s#%d", mech.Owner.Username, mech.Owner.Gid)
		}

		newWarMachine.ModelName = mech.Label
		newWarMachine.ModelID = mech.BlueprintID

		// check model skin
		if mech.ChassisSkin != nil {
			newWarMachine.SkinName = mech.ChassisSkin.Label
			newWarMachine.SkinID = mech.ChassisSkin.BlueprintID
		}

		warMachines = append(warMachines, newWarMachine)
		gamelog.L.Debug().Interface("mech", mech).Interface("newWarMachine", newWarMachine).Msg("converted mech to warmachine")
	}

	sort.Slice(warMachines, func(i, k int) bool {
		return warMachines[i].FactionID == warMachines[k].FactionID
	})

	return warMachines
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	// This code cannot support Japanese
	// orgLen := len(str)
	// if orgLen <= length {
	//     return str
	// }
	// return str[:length]

	// Support Japanese
	// Ref: Range loops https://blog.golang.org/strings
	truncated := ""
	count := 0
	for _, char := range str {
		truncated += string(char)
		count++
		if count >= length {
			break
		}
	}
	return truncated
}

var ModelMap = map[string]string{
	"Law Enforcer X-1000": "XFVS",
	"Olympus Mons LY07":   "BXSD",
	"Tenshi Mk1":          "WREX",
	"BXSD":                "BXSD",
	"XFVS":                "XFVS",
	"WREX":                "WREX",
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

	faction, err := boiler.Factions(boiler.FactionWhere.ID.EQ(user.FactionID.String), qm.Load(boiler.FactionRels.FactionPalette)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", userID.String()).Str("faction_id", user.FactionID.String).Err(err).Msg("failed to get player faction from db")
		return userBrief, nil
	}

	userBrief.Faction = &Faction{
		ID:    faction.ID,
		Label: faction.Label,
		Theme: &Theme{
			PrimaryColor:    faction.R.FactionPalette.Primary,
			SecondaryColor:  faction.R.FactionPalette.Text,
			BackgroundColor: faction.R.FactionPalette.Background,
		},
	}

	return userBrief, nil
}
