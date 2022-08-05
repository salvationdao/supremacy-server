package battle

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/system_messages"
	"server/xsyn_rpcclient"
	"sort"
	"strings"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
)

type BattleStage int32

const (
	BattleStageStart = 1
	BattleStageEnd   = 0
)

type Battle struct {
	arena                  *Arena
	stage                  *atomic.Int32
	BattleID               string        `json:"battleID"`
	MapName                string        `json:"mapName"`
	WarMachines            []*WarMachine `json:"warMachines"`
	spawnedAIMux           deadlock.RWMutex
	SpawnedAI              []*WarMachine `json:"SpawnedAI"`
	warMachineIDs          []uuid.UUID
	lastTick               *[]byte
	gameMap                *server.GameMap
	battleZones            []server.BattleZone
	currentBattleZoneIndex int
	_abilities             *AbilitiesSystem
	users                  usersMap
	factions               map[uuid.UUID]*boiler.Faction
	rpcClient              *xsyn_rpcclient.XrpcClient
	battleMechData         []*db.BattleMechData
	startedAt              time.Time

	_playerAbilityManager *PlayerAbilityManager

	destroyedWarMachineMap map[string]*WMDestroyedRecord
	*boiler.Battle

	inserted bool

	viewerCountInputChan chan *ViewerLiveCount
	deadlock.RWMutex
}

func (btl *Battle) AbilitySystem() *AbilitiesSystem {
	btl.RLock()
	defer btl.RUnlock()
	return btl._abilities
}

func (btl *Battle) playerAbilityManager() *PlayerAbilityManager {
	btl.RLock()
	defer btl.RUnlock()
	return btl._playerAbilityManager
}

func (btl *Battle) storeAbilities(as *AbilitiesSystem) {
	btl.Lock()
	defer btl.Unlock()
	btl._abilities = as
}

// storeGameMap set the game map detail from game client
func (btl *Battle) storeGameMap(gm server.GameMap, battleZones []server.BattleZone) {
	gamelog.L.Trace().Str("func", "storeGameMap").Msg("start")
	btl.Lock()
	defer btl.Unlock()

	btl.gameMap.ImageUrl = gm.ImageUrl
	btl.gameMap.Width = gm.Width
	btl.gameMap.Height = gm.Height
	btl.gameMap.CellsX = gm.CellsX
	btl.gameMap.CellsY = gm.CellsY
	btl.gameMap.LeftPixels = gm.LeftPixels
	btl.gameMap.TopPixels = gm.TopPixels
	btl.gameMap.DisabledCells = gm.DisabledCells
	btl.battleZones = battleZones
	gamelog.L.Trace().Str("func", "storeGameMap").Msg("end")
}

func (btl *Battle) setBattleQueue() error {
	l := gamelog.L.With().Str("log_name", "battle arena").Interface("battle", btl).Str("battle.go", ":battle.go:battle.Battle()").Logger()
	if btl.inserted {
		btl.Battle.StartedAt = time.Now()
		_, err := btl.Battle.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleColumns.StartedAt))
		if err != nil {
			l.Error().Err(err).Msg("unable to update Battle in database")
			return err
		}

		_, err = boiler.BattleMechs(boiler.BattleMechWhere.BattleID.EQ(btl.ID)).DeleteAll(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("unable to delete delete stale battle mechs from database")
		}

		_, err = boiler.BattleWins(boiler.BattleWinWhere.BattleID.EQ(btl.ID)).DeleteAll(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("unable to delete delete stale battle wins from database")
		}

		_, err = boiler.BattleKills(boiler.BattleKillWhere.BattleID.EQ(btl.ID)).DeleteAll(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("unable to delete delete stale battle kills from database")
		}

		_, err = boiler.BattleHistories(boiler.BattleHistoryWhere.BattleID.EQ(btl.ID)).DeleteAll(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("unable to delete delete stale battle histories from database")
		}

	} else {
		err := btl.Battle.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("unable to insert Battle into database")
			return err
		}

		gamelog.L.Debug().Msg("Inserted battle into db")
		btl.inserted = true

		// insert current users to
		btl.users.Range(func(user *BattleUser) bool {
			err = db.BattleViewerUpsert(btl.ID, user.ID.String())
			if err != nil {
				l.Error().Str("player_id", user.ID.String()).Err(err).Msg("to upsert battle view")
				return true
			}
			return true
		})

		err = db.QueueSetBattleID(btl.ID, btl.warMachineIDs...)
		if err != nil {
			l.Error().Interface("mechs_ids", btl.warMachineIDs).Err(err).Msg("failed to set battle id in queue")
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue-update", server.RedMountainFactionID), WSPlayerAssetMechQueueUpdateSubscribe, true)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue-update", server.BostonCyberneticsFactionID), WSPlayerAssetMechQueueUpdateSubscribe, true)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue-update", server.ZaibatsuFactionID), WSPlayerAssetMechQueueUpdateSubscribe, true)
	}

	return nil
}

func (btl *Battle) storePlayerAbilityManager(im *PlayerAbilityManager) {
	btl.Lock()
	defer btl.Unlock()
	btl._playerAbilityManager = im
}

func (btl *Battle) warMachineUpdateFromGameClient(payload *BattleStartPayload) ([]*db.BattleMechData, map[uuid.UUID]*boiler.Faction, error) {
	bmd := make([]*db.BattleMechData, len(btl.WarMachines))
	factions := map[uuid.UUID]*boiler.Faction{}

	for i, wm := range btl.WarMachines {
		wm.Lock() // lock mech detail
		for ii, pwm := range payload.WarMachines {
			if wm.Hash == pwm.Hash {
				wm.ParticipantID = pwm.ParticipantID
				break
			}
			if ii == len(payload.WarMachines)-1 {
				gamelog.L.Error().Str("log_name", "battle arena").Err(fmt.Errorf("didnt find matching hash"))
			}
		}
		wm.Unlock()

		gamelog.L.Trace().Interface("battle war machine", wm).Msg("battle war machine")

		mechID, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("ownerID", wm.ID).Err(err).Msg("unable to convert owner id from string")
			return nil, nil, err
		}

		ownerID, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("ownerID", wm.OwnedByID).Err(err).Msg("unable to convert owner id from string")
			return nil, nil, err
		}

		factionID, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("factionID", wm.FactionID).Err(err).Msg("unable to convert faction id from string")
			return nil, nil, err
		}

		bmd[i] = &db.BattleMechData{
			MechID:    mechID,
			OwnerID:   ownerID,
			FactionID: factionID,
		}

		_, ok := factions[factionID]
		if !ok {
			faction, err := boiler.FindFaction(gamedb.StdConn, factionID.String())
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").
					Str("Battle ID", btl.ID).
					Str("Faction ID", factionID.String()).
					Err(err).Msg("unable to retrieve faction from database")

			}
			factions[factionID] = faction
		}
	}

	return bmd, factions, nil
}

func (btl *Battle) preIntro(payload *BattleStartPayload) error {
	gamelog.L.Trace().Str("func", "preIntro").Msg("start")

	btl.Lock()
	defer btl.Unlock()

	bmd, factions, err := btl.warMachineUpdateFromGameClient(payload)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to update war machine from game client data")
		return err
	}

	btl.factions = factions
	btl.battleMechData = bmd

	btl.BroadcastUpdate()
	gamelog.L.Trace().Str("func", "preIntro").Msg("end")
	return nil
}

func (btl *Battle) start() {
	gamelog.L.Trace().Str("func", "start").Msg("start")

	var err error

	if btl.battleMechData == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("battlemechdata", btl.ID).Msg("battle mech data failed nil check")
	}

	err = db.BattleMechs(btl.Battle, btl.battleMechData)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("Battle ID", btl.ID).Err(err).Msg("unable to insert battle into database")
		//TODO: something more dramatic
	}

	gamelog.L.Debug().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up battle AbilitySystem()")
	btl.storeAbilities(NewAbilitiesSystem(btl))
	gamelog.L.Debug().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting battle start to players")
	btl.BroadcastUpdate()

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
	//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
	//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels
	return &server.GameLocation{
		X: ((cell.X * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + gameMap.LeftPixels,
		Y: ((cell.Y * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + gameMap.TopPixels,
	}
}

// getCellCoordinatesFromGameWorldXY converts location in game to a cell location
func (btl *Battle) getCellCoordinatesFromGameWorldXY(location *server.GameLocation) *server.CellLocation {
	gameMap := btl.gameMap
	return &server.CellLocation{
		X: (location.X - gameMap.LeftPixels - server.GameClientTileSize*2) / server.GameClientTileSize,
		Y: (location.Y - gameMap.TopPixels - server.GameClientTileSize*2) / server.GameClientTileSize,
	}
}

type WarMachinePosition struct {
	X int
	Y int
}

func (btl *Battle) spawnReinforcementNearMech(abilityEvent *server.GameAbilityEvent) {
	// only calculate reinforcement location
	if abilityEvent.GameClientAbilityID != 10 {
		return
	}

	// get snapshots of the red mountain war machines health and postion
	rmw := []WarMachinePosition{}
	aliveWarMachines := []WarMachinePosition{}
	for _, wm := range btl.WarMachines {
		// store red mountain war machines
		if wm.FactionID != server.RedMountainFactionID || wm.Position == nil {
			continue
		}

		// get snapshot of current war machine
		x := wm.Position.X
		y := wm.Position.Y

		rmw = append(rmw, WarMachinePosition{
			X: x,
			Y: y,
		})

		// store alive red mountain war machines
		if wm.Health <= 0 || wm.Health >= 10000 {
			continue
		}
		aliveWarMachines = append(aliveWarMachines, WarMachinePosition{
			X: x,
			Y: y,
		})
	}

	// should never happen, but just in case
	if len(rmw) == 0 {
		gamelog.L.Warn().Str("ability_trigger_offering_id", abilityEvent.EventID.String()).Msg("No Red Mountain mech in the battle to locate reinforcement bot, which should never happen...")
		return
	}

	if len(aliveWarMachines) > 0 {
		// random pick one of the red mountain postion
		wm := aliveWarMachines[rand.Intn(len(aliveWarMachines))]

		// set cell
		abilityEvent.TriggeredOnCellX = &wm.X
		abilityEvent.TriggeredOnCellY = &wm.Y
		abilityEvent.GameLocation = &server.GameLocation{
			X: wm.X,
			Y: wm.Y,
		}

		return
	}

	wm := rmw[rand.Intn(len(rmw))]
	// set cell
	abilityEvent.TriggeredOnCellX = &wm.X
	abilityEvent.TriggeredOnCellY = &wm.Y

	abilityEvent.GameLocation = &server.GameLocation{
		X: wm.X,
		Y: wm.Y,
	}
}

func (btl *Battle) isOnline(userID uuid.UUID) bool {
	_, ok := btl.users.User(userID)
	return ok
}

func (btl *Battle) endAbilities() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the battle AbilitySystem() end!", r)
		}
	}()

	gamelog.L.Debug().Msgf("cleaning up AbilitySystem(): %s", btl.ID)

	if btl.AbilitySystem() == nil {
		gamelog.L.Error().Str("log_name", "battle arena").Msg("battle did not have AbilitySystem()!")
		return
	}

	btl.AbilitySystem().End()
	btl.AbilitySystem().storeBattle(nil)
	btl.storeAbilities(nil)
}

func (btl *Battle) endCreateStats(payload *BattleEndPayload, winningWarMachines []*WarMachine) *BattleEndDetail {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the creation of ending info: endCreateStats!", r)
		}
	}()

	gamelog.L.Debug().Msgf("battle end: looping MostFrequentAbilityExecutors: %s", btl.ID)
	topPlayerExecutorsBoilers, err := db.MostFrequentAbilityExecutors(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top player executors")
	}

	gamelog.L.Debug().Msgf("battle end: looping topPlayerExecutorsBoilers: %s", btl.ID)
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

	// winning factions
	winningFaction := winningWarMachines[0].Faction

	// get winning faction order
	winningFactionIDOrder := []string{winningFaction.ID}

	factionIDs, err := db.FactionMechDestroyedOrderGet(btl.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load mech destroy order.")
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

	gamelog.L.Debug().
		Int("top_player_executors", len(topPlayerExecutors)).
		Msg("get top players and factions")

	return &BattleEndDetail{
		BattleID:                     btl.ID,
		BattleIdentifier:             btl.Battle.BattleNumber,
		StartedAt:                    btl.Battle.StartedAt,
		EndedAt:                      btl.Battle.EndedAt.Time,
		WinningCondition:             payload.WinCondition,
		WinningFaction:               winningFaction,
		WinningFactionIDOrder:        winningFactionIDOrder,
		WinningWarMachines:           winningWarMachines,
		MostFrequentAbilityExecutors: topPlayerExecutors,
	}
}

func (btl *Battle) processWinners(payload *BattleEndPayload) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the battle end processWinners!", r)
		}
	}()
	mws := make([]*db.MechWithOwner, len(payload.WinningWarMachines))

	for i, wmwin := range payload.WinningWarMachines {
		var wm *WarMachine
		for _, w := range btl.WarMachines {
			if w.Hash == wmwin.Hash {
				wm = w
				break
			}
		}
		if wm == nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("Battle ID", btl.ID).Msg("unable to match war machine to battle with hash")
			continue
		}
		mechId, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("Battle ID", btl.ID).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert mech id to uuid")
			continue
		}
		ownedById, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("Battle ID", btl.ID).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert owned id to uuid")
			continue
		}
		factionId, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("Battle ID", btl.ID).
				Str("faction ID", wm.FactionID).
				Err(err).
				Msg("unable to convert faction id to uuid")
			continue
		}
		mws[i] = &db.MechWithOwner{
			OwnerID:   ownedById,
			MechID:    mechId,
			FactionID: factionId,
		}
	}
	err := db.WinBattle(btl.ID, payload.WinCondition, mws...)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").
			Str("Battle ID", btl.ID).
			Err(err).
			Msg("unable to store mech wins")
	}
}

type PlayerReward struct {
	PlayerID              string                         `json:"player_id"`
	RewardedSups          decimal.Decimal                `json:"rewarded_sups"`
	RewardedPlayerAbility *boiler.BlueprintPlayerAbility `json:"rewarded_player_ability"`
	FactionRank           string                         `json:"faction_rank"`
}
type MechReward struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	FactionID    string          `json:"faction_id"`
	AvatarURL    string          `json:"avatar_url"`
	OwnerID      string          `json:"owner_id"`
	RewardedSups decimal.Decimal `json:"rewarded_sups"`
}

// RewardBattleMechOwners give reward to war machine owner
func (btl *Battle) RewardBattleMechOwners(winningFactionOrder []string) ([]*PlayerReward, []*MechReward) {
	playerRewards := []*PlayerReward{}
	mechRewars := []*MechReward{}

	abilityRewardPlayers := []string{}

	// get sups pool
	bqs, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.BattleID.EQ(null.StringFrom(btl.ID)),
		qm.Load(boiler.BattleQueueRels.Fee),
		qm.Load(boiler.BattleQueueRels.Owner),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle id", btl.ID).Msg("Failed to load battle queue fees")
		return []*PlayerReward{}, []*MechReward{}
	}

	totalSups := decimal.Zero
	for _, bq := range bqs {
		if bq.R != nil && bq.R.Fee != nil {
			totalSups = totalSups.Add(bq.R.Fee.Amount)
		}
	}

	if totalSups.Equal(decimal.Zero) {
		gamelog.L.Debug().Msg("No sups to distribute.")
		return []*PlayerReward{}, []*MechReward{}
	}

	// get players per faction
	playerPerFaction := decimal.Zero
	for _, bq := range bqs {
		if bq.FactionID != server.RedMountainFactionID {
			continue
		}
		// if owner is not AI
		if bq.R != nil && bq.R.Owner != nil && !bq.R.Owner.IsAi {
			playerPerFaction = playerPerFaction.Add(decimal.NewFromInt(1))
		}
	}

	firstRankSupsRewardRatio := db.GetDecimalWithDefault(db.KeyFirstRankFactionRewardRatio, decimal.NewFromFloat(0.5))
	secondRankSupsRewardRatio := db.GetDecimalWithDefault(db.KeySecondRankFactionRewardRatio, decimal.NewFromFloat(0.3))
	thirdRankSupsRewardRatio := db.GetDecimalWithDefault(db.KeyThirdRankFactionRewardRatio, decimal.NewFromFloat(0.2))

	// reward sups
	taxRatio := db.GetDecimalWithDefault(db.KeyBattleRewardTaxRatio, decimal.NewFromFloat(0.025))
	for i, factionID := range winningFactionOrder {
		switch i {
		case 0: // winning faction
			for _, bq := range bqs {
				if bq.FactionID == factionID && bq.R != nil && bq.R.Fee != nil && bq.R.Owner != nil && !bq.R.Owner.IsAi {
					pw := btl.RewardPlayerSups(
						bq.R.Fee,
						totalSups.Mul(firstRankSupsRewardRatio).Div(playerPerFaction),
						taxRatio,
					)

					// record mech reward
					if m := btl.arena.CurrentBattleWarMachineByID(bq.MechID); m != nil {
						mechRewars = append(mechRewars, &MechReward{
							ID:           m.ID,
							FactionID:    m.FactionID,
							Name:         m.Name,
							Label:        m.Label,
							AvatarURL:    m.ImageAvatar,
							RewardedSups: pw.RewardedSups,
							OwnerID:      bq.OwnerID,
						})
					}

					// append or update player rewards
					exist := false
					for _, pr := range playerRewards {
						if pr.PlayerID == pw.PlayerID {
							pr.RewardedSups = pr.RewardedSups.Add(pw.RewardedSups)
							exist = true
						}
					}
					if !exist {
						// fill war machine
						pw.FactionRank = "FIRST"
						playerRewards = append(playerRewards, pw)
					}

				}
			}

		case 1: // second faction
			for _, bq := range bqs {
				if bq.FactionID == factionID && bq.R != nil && bq.R.Fee != nil && bq.R.Owner != nil && !bq.R.Owner.IsAi {
					pw := btl.RewardPlayerSups(
						bq.R.Fee,
						totalSups.Mul(secondRankSupsRewardRatio).Div(playerPerFaction),
						taxRatio,
					)

					// record mech reward
					if m := btl.arena.CurrentBattleWarMachineByID(bq.MechID); m != nil {
						mechRewars = append(mechRewars, &MechReward{
							ID:           m.ID,
							FactionID:    m.FactionID,
							Name:         m.Name,
							Label:        m.Label,
							AvatarURL:    m.ImageAvatar,
							RewardedSups: pw.RewardedSups,
							OwnerID:      bq.OwnerID,
						})
					}

					// append or update player rewards
					exist := false
					for _, pr := range playerRewards {
						if pr.PlayerID == pw.PlayerID {
							pr.RewardedSups = pr.RewardedSups.Add(pw.RewardedSups)
							exist = true
						}
					}
					if !exist {
						pw.FactionRank = "SECOND"
						playerRewards = append(playerRewards, pw)
					}

				}
			}

		case 2: // lose faction
			for _, bq := range bqs {
				if bq.FactionID == factionID && bq.R != nil && bq.R.Fee != nil && bq.R.Owner != nil && !bq.R.Owner.IsAi {
					pw := btl.RewardPlayerSups(
						bq.R.Fee,
						totalSups.Mul(thirdRankSupsRewardRatio).Div(playerPerFaction),
						taxRatio,
					)

					// record mech reward
					if m := btl.arena.CurrentBattleWarMachineByID(bq.MechID); m != nil {
						mechRewars = append(mechRewars, &MechReward{
							ID:           m.ID,
							FactionID:    m.FactionID,
							Name:         m.Name,
							Label:        m.Label,
							AvatarURL:    m.ImageAvatar,
							RewardedSups: pw.RewardedSups,
							OwnerID:      bq.OwnerID,
						})
					}

					// append or update player rewards
					exist := false
					for _, pr := range playerRewards {
						if pr.PlayerID == pw.PlayerID {
							pr.RewardedSups = pr.RewardedSups.Add(pw.RewardedSups)
							exist = true
						}
					}
					if !exist {
						pw.FactionRank = "THIRD"
						playerRewards = append(playerRewards, pw)
					}

					// add player ability reward list
					exists := false
					for _, pid := range abilityRewardPlayers {
						if pid == bq.OwnerID {
							exists = true
							break
						}
					}
					if !exists {
						abilityRewardPlayers = append(abilityRewardPlayers, bq.OwnerID)
					}
				}
			}
		}
	}

	// reward player abilities
	pws := btl.RewardPlayerAbility(abilityRewardPlayers)
	for _, pw := range pws {
		for _, pr := range playerRewards {
			if pr.PlayerID == pw.PlayerID {
				pr.RewardedPlayerAbility = pw.RewardedPlayerAbility
				break
			}
		}
	}

	return playerRewards, mechRewars
}

// RewardPlayerSups reward player sups
func (btl *Battle) RewardPlayerSups(queueFee *boiler.BattleQueueFee, supsReward decimal.Decimal, taxRatio decimal.Decimal) *PlayerReward {
	playerID := queueFee.PaidByID
	tax := supsReward.Mul(taxRatio)
	challengeFund := decimal.New(1, 18)

	l := gamelog.L.With().Str("function", "RewardPlayerSups").Logger()

	// record
	pw := &PlayerReward{
		PlayerID:     queueFee.PaidByID,
		RewardedSups: supsReward,
	}

	// pay battle queue fee
	payoutTXID, err := btl.arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(server.SupremacyBattleUserID)),
		ToUserID:             uuid.Must(uuid.FromString(playerID)),
		Amount:               supsReward.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("battle_reward|%s|%d", btl.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          fmt.Sprintf("reward from battle #%d.", btl.BattleNumber),
	})
	if err != nil {
		l.Error().Err(err).
			Str("from", server.SupremacyBattleUserID).
			Str("to", playerID).
			Str("amount", supsReward.StringFixed(0)).
			Msg("Failed to pay player battel reward")
	}
	queueFee.PayoutTXID = null.StringFrom(payoutTXID)

	// pay reward tax
	taxTXID, err := btl.arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(playerID)),
		ToUserID:             uuid.UUID(server.XsynTreasuryUserID),
		Amount:               tax.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("battle_reward_tax|%s|%d", btl.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          fmt.Sprintf("reward tax from battle #%d.", btl.BattleNumber),
	})
	if err != nil {
		l.Error().Err(err).
			Str("from", playerID).
			Str("to", server.XsynTreasuryUserID.String()).
			Str("amount", tax.StringFixed(0)).
			Msg("Failed to pay player battle reward")
	}
	queueFee.TaxTXID = null.StringFrom(taxTXID)

	// pay challenge fund
	challengeFundTXID, err := btl.arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(playerID)),
		ToUserID:             uuid.Must(uuid.FromString(server.SupremacyChallengeFundUserID)),
		Amount:               challengeFund.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("supremacy_challenge_fund|%s|%d", btl.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          fmt.Sprintf("challenge fund from battle #%d.", btl.BattleNumber),
	})
	if err != nil {
		l.Error().Err(err).
			Str("from", playerID).
			Str("to", server.SupremacyChallengeFundUserID).
			Str("amount", challengeFund.StringFixed(0)).
			Msg("Failed to pay player battle reward")
	}
	queueFee.ChallengeFundTXID = null.StringFrom(challengeFundTXID)

	_, err = queueFee.Update(gamedb.StdConn, boil.Whitelist(
		boiler.BattleQueueFeeColumns.PayoutTXID,
		boiler.BattleQueueFeeColumns.TaxTXID,
		boiler.BattleQueueFeeColumns.ChallengeFundTXID,
	))
	if err != nil {
		l.Error().Err(err).Interface("queue fee", queueFee).Msg("Failed to update payout, tax and challenge fund transaction id")
	}

	return pw
}

// RewardPlayerAbility reward mech owners from lose faction one player ability
func (btl *Battle) RewardPlayerAbility(playerIDs []string) []*PlayerReward {
	pws := []*PlayerReward{}

	if len(playerIDs) == 0 {
		return pws
	}

	bpas, err := boiler.SalePlayerAbilities(
		boiler.SalePlayerAbilityWhere.RarityWeight.GT(0),
		qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to refresh pool of sale abilities from db")
	}

	for _, pid := range playerIDs {
		// load existing player abilities
		pas, err := boiler.PlayerAbilities(
			boiler.PlayerAbilityWhere.OwnerID.EQ(pid),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("player id", pid).Msg("Failed to load player abilities")
			continue
		}

		availableAbilities := []*boiler.SalePlayerAbility{}
		for _, bpa := range bpas {
			isAvailable := true
			for _, pa := range pas {
				if pa.BlueprintID != bpa.ID {
					continue
				}

				// if player has the ability, check ability is reach the limit
				if pa.Count >= bpa.R.Blueprint.InventoryLimit {
					isAvailable = false
				}

				break
			}

			// collect available abilities
			if isAvailable {
				availableAbilities = append(availableAbilities, bpa)
			}
		}

		// skip, if no player ability is available
		if len(availableAbilities) == 0 {
			sysMsg := boiler.SystemMessage{
				PlayerID: pid,
				SenderID: server.SupremacyBattleUserID,
				DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeMechOwnerBattleReward)),
				Title:    "Battle Reward",
				Message:  "Unable to reward you new player ability due to your inventory is full.",
			}
			err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
				break
			}
			ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", pid), server.HubKeySystemMessageListUpdatedSubscribe, true)

			continue
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

		err = db.PlayerAbilityAssign(pid, ability.BlueprintID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("player id", pid).Str("ability id", ability.ID).Msg("Failed to assign ability to the player")
			continue
		}

		pws = append(pws, &PlayerReward{
			PlayerID:              pid,
			RewardedPlayerAbility: ability.R.Blueprint,
		})
	}

	return pws
}

func (btl *Battle) processWarMachineRepair() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at register mech repair cases", r)
		}
	}()
	for _, wm := range btl.WarMachines {
		wm.RLock()
		mechID := wm.ID
		modelID := wm.ModelID
		maxHealth := wm.MaxHealth
		health := wm.Health
		ownerID := wm.OwnedByID
		wm.RUnlock()

		go func() {
			// skip, if player is AI
			p, err := boiler.FindPlayer(gamedb.StdConn, ownerID)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load mech owner detail")
				return
			}

			if p.IsAi {
				return
			}

			// register mech repair case
			err = RegisterMechRepairCase(mechID, modelID, maxHealth, health)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to register mech repair")
			}
		}()
	}
}

func (btl *Battle) endWarMachines(payload *BattleEndPayload) []*WarMachine {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the sorting up ending war machines!", r)
		}
	}()
	winningWarMachines := make([]*WarMachine, len(payload.WinningWarMachines))

	gamelog.L.Debug().Msgf("battle end: looping WinningWarMachines: %s", btl.ID)
	for i := range payload.WinningWarMachines {
		for _, w := range btl.WarMachines {
			if w.Hash == payload.WinningWarMachines[i].Hash {
				winningWarMachines[i] = w
				break
			}
		}
		if winningWarMachines[i] == nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("Battle ID", btl.ID).Msg("unable to match war machine to battle with hash")
		}
	}

	if len(winningWarMachines) == 0 || winningWarMachines[0] == nil {
		gamelog.L.Panic().Str("Battle ID", btl.ID).Msg("no winning war machines")
	} else {
		for _, w := range winningWarMachines {
			// update battle_mechs to indicate survival
			bm, err := boiler.FindBattleMech(gamedb.StdConn, btl.ID, w.ID)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").
					Str("battleID", btl.ID).
					Str("mechID", w.ID).
					Str("db func", "endWarMachines").
					Err(err).Msg("unable to retrieve battle mech from database")
				continue
			}

			bm.MechSurvived = null.BoolFrom(true)
			_, err = bm.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.BattleMech", bm).
					Msg("unable to update winning battle mech")
			}

			// update mech_stats, total_wins column
			ms, err := boiler.MechStats(boiler.MechStatWhere.MechID.EQ(w.ID)).One(gamedb.StdConn)
			if errors.Is(err, sql.ErrNoRows) {
				// If mech stats not exist then create it
				newMs := boiler.MechStat{
					MechID:          w.ID,
					BattlesSurvived: 1,
				}
				err := newMs.Insert(gamedb.StdConn, boil.Infer())
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", newMs).
					Msg("unable to create mech stat")
				continue
			} else if err != nil {
				gamelog.L.Warn().Err(err).
					Str("mechID", w.ID).
					Msg("unable to get mech stat")
				continue
			}

			ms.BattlesSurvived = ms.BattlesSurvived + 1
			_, err = ms.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", ms).
					Msg("unable to update mech stat")
			}

			prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(bm.OwnerID)).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("player_id", bm.OwnerID).Msg("unable to get player prefs")
				continue
			}

			if prefs != nil && prefs.TelegramID.Valid && prefs.EnableTelegramNotifications {
				// killed a war machine
				msg := fmt.Sprintf("Your War machine %s is Victorious! 🎉", w.Name)
				err := btl.arena.telegram.Notify(prefs.TelegramID.Int64, msg)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("telegramID", fmt.Sprintf("%v", prefs.TelegramID)).Err(err).Msg("failed to send notification")
				}
			}

		}

		// update battle_mechs to indicate faction win
		bms, err := boiler.BattleMechs(boiler.BattleMechWhere.FactionID.EQ(winningWarMachines[0].FactionID), boiler.BattleMechWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("battleID", btl.ID).
				Str("factionID", winningWarMachines[0].FactionID).
				Str("db func", "endWarMachines").
				Err(err).Msg("unable to retrieve faction battle mechs from database")
		}
		_, err = bms.UpdateAll(gamedb.StdConn, boiler.M{
			"faction_won": true,
		})
		if err != nil {
			gamelog.L.Warn().Err(err).
				Interface("boiler.BattleMech", bms).
				Msg("unable to update faction battle mechs")
		}

		// update mech_stats total_wins (total faction wins)
		wonBms, err := boiler.BattleMechs(boiler.BattleMechWhere.FactionID.EQ(winningWarMachines[0].FactionID), boiler.BattleMechWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("battleID", btl.ID).
				Str("factionID", winningWarMachines[0].FactionID).
				Str("db func", "endWarMachines").
				Err(err).Msg("unable to retrieve winning faction battle mechs from database")
		}
		for _, w := range wonBms {
			// update mech_stats, total_losses column
			wms, err := boiler.MechStats(boiler.MechStatWhere.MechID.EQ(w.MechID)).One(gamedb.StdConn)
			if errors.Is(err, sql.ErrNoRows) {
				// If mech stats not exist then create it
				newMs := boiler.MechStat{
					MechID:    w.MechID,
					TotalWins: 1,
				}
				err := newMs.Insert(gamedb.StdConn, boil.Infer())
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", newMs).
					Msg("unable to create mech stat")
				continue
			} else if err != nil {
				gamelog.L.Warn().Err(err).
					Str("mechID", w.MechID).
					Msg("unable to get mech stat")
				continue
			}

			wms.TotalWins = wms.TotalWins + 1
			_, err = wms.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", wms).
					Msg("unable to update mech stat")
			}
		}

		// update mech_stats total_losses
		lostBms, err := boiler.BattleMechs(boiler.BattleMechWhere.FactionID.NEQ(winningWarMachines[0].FactionID), boiler.BattleMechWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("battleID", btl.ID).
				Str("factionID", winningWarMachines[0].FactionID).
				Str("db func", "endWarMachines").
				Err(err).Msg("unable to retrieve losing faction battle mechs from database")
		}
		for _, l := range lostBms {
			// update mech_stats, total_losses column
			lms, err := boiler.MechStats(boiler.MechStatWhere.MechID.EQ(l.MechID)).One(gamedb.StdConn)
			if errors.Is(err, sql.ErrNoRows) {
				// If mech stats not exist then create it
				newMs := boiler.MechStat{
					MechID:      l.MechID,
					TotalLosses: 1,
				}
				err := newMs.Insert(gamedb.StdConn, boil.Infer())
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", newMs).
					Msg("unable to create mech stat")
				continue
			} else if err != nil {
				gamelog.L.Warn().Err(err).
					Str("mechID", l.MechID).
					Msg("unable to get mech stat")
				continue
			}

			lms.TotalLosses = lms.TotalLosses + 1
			_, err = lms.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", lms).
					Msg("unable to update mech stat")
			}
		}

		// record faction win/loss count
		err = db.FactionAddWinLossCount(winningWarMachines[0].FactionID)
		if err != nil {
			gamelog.L.Panic().Str("Battle ID", btl.ID).Str("winning_faction_id", winningWarMachines[0].FactionID).Msg("Failed to update faction win/loss count")
		}
	}

	return winningWarMachines
}

const HubKeyBattleEndDetailUpdated = "BATTLE:END:DETAIL:UPDATED"

func (btl *Battle) endBroadcast(endInfo *BattleEndDetail, playerRewardRecords []*PlayerReward, mechRewardRecords []*MechReward) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the ending of end broadcast!", r)
		}
	}()
	for _, prr := range playerRewardRecords {
		// send battle reward system message
		b, err := json.Marshal(prr)
		if err != nil {
			gamelog.L.Error().Interface("player reward data", prr).Err(err).Msg("Failed to marshal player reward data into json.")
			break
		}
		sysMsg := boiler.SystemMessage{
			PlayerID: prr.PlayerID,
			SenderID: server.SupremacyBattleUserID,
			DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeMechOwnerBattleReward)),
			Title:    "Battle Reward",
			Message:  fmt.Sprintf("Your faction is the %s rank in the battle #%d.", prr.FactionRank, btl.BattleNumber),
			Data:     null.JSONFrom(b),
		}
		err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
			break
		}
		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", prr.PlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	}

	endInfo.MechRewards = mechRewardRecords

	ws.PublishMessage("/public/battle_end_result", HubKeyBattleEndDetailUpdated, endInfo)
}

func (btl *Battle) end(payload *BattleEndPayload) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the battle end!", r)

			exists, err := boiler.BattleExists(gamedb.StdConn, btl.ID)
			if err != nil {
				gamelog.L.Panic().Err(err).Msg("Panicing. Unable to even check if battle id exists")
			}
			if exists {

			}
		}
	}()

	btl.Battle.EndedAt = null.TimeFrom(time.Now())
	_, err := btl.Battle.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("Battle ID", btl.ID).Time("EndedAt", btl.EndedAt.Time).Msg("unable to update database for endat battle")
	}

	btl.endAbilities()

	winningWarMachines := btl.endWarMachines(payload)
	endInfo := btl.endCreateStats(payload, winningWarMachines)
	playerRewardRecords, mechRewardRecords := btl.RewardBattleMechOwners(endInfo.WinningFactionIDOrder)
	btl.processWinners(payload)

	btl.processWarMachineRepair()

	// TODO: we can remove this after a while
	_, err = boiler.BattleQueueNotifications(
		boiler.BattleQueueNotificationWhere.QueueMechID.IsNotNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{"queue_mech_id": nil})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("failed to update battle queue notifications")
	}

	// broadcast system message to mech owners
	q, err := boiler.BattleQueues(boiler.BattleQueueWhere.BattleID.EQ(null.StringFrom(btl.BattleID))).All(gamedb.StdConn)
	go system_messages.BroadcastMechBattleCompleteMessage(q, btl.BattleID)

	_, err = q.DeleteAll(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Str("Battle ID", btl.ID).Str("battle_id", payload.BattleID).Msg("Failed to remove mechs from battle queue.")
	}

	gamelog.L.Info().Msgf("battle has been cleaned up, sending broadcast %s", btl.ID)
	btl.endBroadcast(endInfo, playerRewardRecords, mechRewardRecords)
}

type GameSettingsResponse struct {
	GameMap            *server.GameMap    `json:"game_map"`
	BattleZone         *server.BattleZone `json:"battle_zone"`
	WarMachines        []*WarMachine      `json:"war_machines"`
	SpawnedAI          []*WarMachine      `json:"spawned_ai"`
	WarMachineLocation []byte             `json:"war_machine_location"`
	BattleIdentifier   int                `json:"battle_identifier"`
	AbilityDetails     []*AbilityDetail   `json:"ability_details"`
}

type ViewerLiveCount struct {
	RedMountain int64 `json:"red_mountain"`
	Boston      int64 `json:"boston"`
	Zaibatsu    int64 `json:"zaibatsu"`
	Other       int64 `json:"other"`
}

func (btl *Battle) UserOnline(user *BattleUser) *ViewerLiveCount {
	_, ok := btl.users.User(user.ID)
	if !ok {
		btl.users.Add(user)
	}

	if btl.inserted {
		err := db.BattleViewerUpsert(btl.ID, user.ID.String())
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("battle_id", btl.ID).
				Str("player_id", user.ID.String()).
				Err(err).
				Msg("could not upsert battle viewer")
		}
	}

	resp := &ViewerLiveCount{
		RedMountain: 0,
		Boston:      0,
		Zaibatsu:    0,
		Other:       0,
	}

	// TODO: optimise at some point
	btl.users.Range(func(user *BattleUser) bool {
		if faction, ok := FactionNames[user.FactionID]; ok {
			switch faction {
			case "RedMountain":
				resp.RedMountain++
			case "Boston":
				resp.Boston++
			case "Zaibatsu":
				resp.Zaibatsu++
			default:
				resp.Other++
			}
		} else {
			resp.Other++
		}
		return true
	})

	btl.viewerCountInputChan <- resp

	return resp
}

func (btl *Battle) debounceSendingViewerCount(cb func(result ViewerLiveCount, btl *Battle)) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the debounceSendingViewerCount!", r)
		}
	}()

	var result *ViewerLiveCount
	interval := 500 * time.Millisecond
	timer := time.NewTimer(interval)
	checker := time.NewTicker(1 * time.Second)
	for {
		select {
		case result = <-btl.viewerCountInputChan:
			timer.Reset(interval)
		case <-timer.C:
			if result != nil {
				cb(*result, btl)
			}
		case <-checker.C:
			if btl != btl.arena.CurrentBattle() {
				timer.Stop()
				checker.Stop()
				gamelog.L.Info().Msg("Clean up live count debounce function due to battle missmatch")
				return
			}
		}
	}
}

func GameSettingsPayload(btl *Battle) *GameSettingsResponse {
	var lt []byte
	if btl.lastTick != nil {
		lt = *btl.lastTick
	}
	if btl == nil {
		return nil
	}

	// Indexes correspond to the game_client_ability_id in the db
	abilityDetails := make([]*AbilityDetail, 20)
	// Nuke
	abilityDetails[1] = &AbilityDetail{
		Radius: 5200,
	}
	// EMP
	abilityDetails[12] = &AbilityDetail{
		Radius: 10000,
	}
	// BLACKOUT
	abilityDetails[16] = &AbilityDetail{
		Radius: 20000,
	}

	// Current Battle Zone
	var battleZone *server.BattleZone
	if len(btl.battleZones) > 0 {
		if btl.currentBattleZoneIndex >= len(btl.battleZones) {
			btl.currentBattleZoneIndex = 0
		}
		battleZone = &btl.battleZones[btl.currentBattleZoneIndex]
	}

	return &GameSettingsResponse{
		BattleIdentifier:   btl.BattleNumber,
		GameMap:            btl.gameMap,
		BattleZone:         battleZone,
		WarMachines:        btl.WarMachines,
		SpawnedAI:          btl.SpawnedAI,
		WarMachineLocation: lt,
		AbilityDetails:     abilityDetails,
	}
}

const HubKeyGameSettingsUpdated = "GAME:SETTINGS:UPDATED"

func (btl *Battle) BroadcastUpdate() {
	ws.PublishMessage("/public/game_settings", HubKeyGameSettingsUpdated, GameSettingsPayload(btl))
}

func (btl *Battle) Tick(payload []byte) {
	gamelog.L.Trace().Str("func", "Tick").Msg("start")
	defer gamelog.L.Trace().Str("func", "Tick").Msg("end")

	if len(payload) < 1 {
		gamelog.L.Error().Str("log_name", "battle arena").Err(fmt.Errorf("len(payload) < 1")).Interface("payload", payload).Msg("len(payload) < 1")
		return
	}

	if btl.stage.Load() == BattleStageEnd {
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

	// collect ws message
	wsMessages := []ws.Message{}

	// Update game settings (so new players get the latest position, health and shield of all warmachines)
	count := payload[1]
	var c byte
	offset := 2
	for c = 0; c < count; c++ {
		participantID := payload[offset]
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
				return
			}
			warmachine = btl.WarMachines[warMachineIndex]
		}
		// Get Sync byte (tells us which data was updated for this warmachine)
		syncByte := payload[offset]
		booleans := helpers.UnpackBooleansFromByte(syncByte)
		offset++
		warmachine.Lock()
		wms := WarMachineStat{
			ParticipantID: int(warmachine.ParticipantID),
			Position:      warmachine.Position,
			Rotation:      warmachine.Rotation,
			Health:        warmachine.Health,
			Shield:        warmachine.Shield,
			IsHidden:      false,
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
		warmachine.Unlock()
		// Energy
		if booleans[3] {
			offset += 4
		}

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

		if participantID < 100 {
			wsMessages = append(wsMessages, ws.Message{
				URI:     fmt.Sprintf("/public/mech/%d", participantID),
				Key:     HubKeyWarMachineStatUpdated,
				Payload: wms,
			})
		}
	}

	if len(wsMessages) > 0 {
		gamelog.L.Trace().Str("func", "Tick").Msg("batch sending")
		ws.PublishBatchMessages("/public/mech", wsMessages)
		gamelog.L.Trace().Str("func", "Tick").Msg("batch sent")
	}

	if btl.playerAbilityManager().HasBlackoutsUpdated() {
		minimapUpdates := []MinimapEvent{}
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
		ws.PublishMessage("/public/minimap", HubKeyMinimapUpdatesSubscribe, minimapUpdates)
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
	dHash := dp.DestroyedWarMachineEvent.DestroyedWarMachineHash
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
	if destroyedWarMachine == nil {
		gamelog.L.Warn().Str("hash", dHash).Msg("can't match destroyed mech with battle state")
		return
	}

	prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(destroyedWarMachine.OwnedByID)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Str("destroyedWarMachine.ID", destroyedWarMachine.ID).Err(err).Msg("failed to get player preferences")

	}

	if prefs != nil && prefs.TelegramID.Valid && prefs.EnableTelegramNotifications {
		// killed a war machine
		msg := fmt.Sprintf("Your War machine %s has been destroyed ☠️", destroyedWarMachine.Name)
		err := btl.arena.telegram.Notify(prefs.TelegramID.Int64, msg)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("playerID", prefs.PlayerID).Str("telegramID", fmt.Sprintf("%v", prefs.TelegramID)).Err(err).Msg("failed to send notification")
		}
	}

	var killedByUser *UserBrief
	var killByWarMachine *WarMachine
	if dp.DestroyedWarMachineEvent.KillByWarMachineHash != "" {
		for _, wm := range btl.WarMachines {
			if wm.Hash == dp.DestroyedWarMachineEvent.KillByWarMachineHash {
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
						err := btl.arena.telegram.Notify(prefs.TelegramID.Int64, msg)
						if err != nil {
							gamelog.L.Error().Str("log_name", "battle arena").Str("playerID", prefs.PlayerID).Str("telegramID", fmt.Sprintf("%v", prefs.TelegramID)).Err(err).Msg("failed to send notification")
						}
					}
				}
			}
		}
		if destroyedWarMachine == nil {
			gamelog.L.Warn().Str("killed_by_hash", dp.DestroyedWarMachineEvent.KillByWarMachineHash).Msg("can't match killer mech with battle state")
			return
		}
	} else if dp.DestroyedWarMachineEvent.RelatedEventIDString != "" {
		// check related event id
		var abl *boiler.BattleAbilityTrigger
		var err error
		retAbl := func() (*boiler.BattleAbilityTrigger, error) {
			abl, err := boiler.BattleAbilityTriggers(boiler.BattleAbilityTriggerWhere.AbilityOfferingID.EQ(dp.DestroyedWarMachineEvent.RelatedEventIDString)).One(gamedb.StdConn)
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
			gamelog.L.Error().Str("log_name", "battle arena").Str("related event id", dp.DestroyedWarMachineEvent.RelatedEventIDString).Err(err).Msg("Failed get ability from offering id")
		}
		// get ability via offering id

		if abl != nil && abl.PlayerID.Valid {
			currentUser, err := BuildUserDetailWithFaction(uuid.FromStringOrNil(abl.PlayerID.String))
			if err == nil {
				// update kill by user and killed by information
				killedByUser = currentUser
				dp.DestroyedWarMachineEvent.KilledBy = fmt.Sprintf("(%s)", abl.AbilityLabel)
			}

			// update player ability kills and faction kills
			if strings.EqualFold(destroyedWarMachine.FactionID, abl.FactionID) {
				// update user kill
				_, err := db.UserStatSubtractAbilityKill(abl.PlayerID.String)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to subtract user ability kill count")
				}

				// insert a team kill record to last seven days kills
				lastSevenDaysKill := boiler.PlayerKillLog{
					PlayerID:   abl.PlayerID.String,
					FactionID:  abl.FactionID,
					BattleID:   btl.BattleID,
					IsTeamKill: true,
				}
				err = lastSevenDaysKill.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to insert player last seven days kill record- (TEAM KILL)")
				}

				// subtract faction kill count
				err = db.FactionSubtractAbilityKillCount(abl.FactionID)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("faction_id", abl.FactionID).Err(err).Msg("Failed to subtract user ability kill count")
				}

				// sent instance to system ban manager
				go btl.arena.SystemBanManager.SendToTeamKillCourtroom(abl.PlayerID.String, dp.DestroyedWarMachineEvent.RelatedEventIDString)

			} else {
				// update user kill
				_, err := db.UserStatAddAbilityKill(abl.PlayerID.String)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to add user ability kill count")
				}

				// insert a team kill record to last seven days kills
				lastSevenDaysKill := boiler.PlayerKillLog{
					PlayerID:  abl.PlayerID.String,
					FactionID: abl.FactionID,
					BattleID:  btl.BattleID,
				}
				err = lastSevenDaysKill.Insert(gamedb.StdConn, boil.Infer())
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
				ws.PublishMessage(fmt.Sprintf("/user/%s/stat", us.ID), server.HubKeyUserStatSubscribe, us)
			}
		}

	}

	gamelog.L.Debug().Msgf("battle Update: %s - War Machine Destroyed: %s", btl.ID, dHash)

	var warMachineID uuid.UUID
	var killByWarMachineID uuid.UUID
	ids, err := db.MechIDsFromHash(destroyedWarMachine.Hash, dp.DestroyedWarMachineEvent.KillByWarMachineHash)

	if err != nil || len(ids) == 0 {
		gamelog.L.Warn().
			Str("hashes", fmt.Sprintf("%s, %s", destroyedWarMachine.Hash, dp.DestroyedWarMachineEvent.KillByWarMachineHash)).
			Str("battle_id", btl.ID).
			Err(err).
			Msg("can't retrieve mech ids")

	} else {
		warMachineID = ids[0]
		if len(ids) > 1 {
			killByWarMachineID = ids[1]
		}

		//TODO: implement related id
		if dp.DestroyedWarMachineEvent.RelatedEventIDString != "" {
			relatedEventuuid, err := uuid.FromString(dp.DestroyedWarMachineEvent.RelatedEventIDString)
			if err != nil {
				gamelog.L.Warn().
					Str("relatedEventuuid", dp.DestroyedWarMachineEvent.RelatedEventIDString).
					Str("battle_id", btl.ID).
					Msg("can't create uuid from non-empty related event idf")
			}
			dp.DestroyedWarMachineEvent.RelatedEventID = relatedEventuuid
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

		if dp.DestroyedWarMachineEvent.RelatedEventIDString != "" {
			bh.RelatedID = null.StringFrom(dp.DestroyedWarMachineEvent.RelatedEventIDString)
		}

		err = bh.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Warn().
				Interface("event_data", bh).
				Str("battle_id", btl.ID).
				Err(err).
				Msg("unable to store mech event data")
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
	for _, damage := range dp.DestroyedWarMachineEvent.DamageHistory {
		totalDamage += damage.Amount
		// check instigator token id exist in the list
		if damage.InstigatorHash != "" {
			exists := false
			for _, hist := range newDamageHistory {
				if hist.InstigatorHash == damage.InstigatorHash {
					hist.Amount += damage.Amount
					exists = true
					break
				}
			}
			if !exists {
				newDamageHistory = append(newDamageHistory, &DamageHistory{
					Amount:         damage.Amount,
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
				hist.Amount += damage.Amount
				exists = true
				break
			}
		}
		if !exists {
			newDamageHistory = append(newDamageHistory, &DamageHistory{
				Amount:         damage.Amount,
				InstigatorHash: damage.InstigatorHash,
				SourceName:     damage.SourceName,
				SourceHash:     damage.SourceHash,
			})
		}
	}

	wmd := &WMDestroyedRecord{
		DestroyedWarMachine: &WarMachineBrief{
			ParticipantID: destroyedWarMachine.ParticipantID,
			ImageUrl:      destroyedWarMachine.Image,
			ImageAvatar:   destroyedWarMachine.ImageAvatar, // TODO: should be imageavatar
			Name:          destroyedWarMachine.Name,
			Hash:          destroyedWarMachine.Hash,
			FactionID:     destroyedWarMachine.FactionID,
		},
		KilledBy: dp.DestroyedWarMachineEvent.KilledBy,
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
					damageRecord.CausedByWarMachine = &WarMachineBrief{
						ParticipantID: wm.ParticipantID,
						ImageUrl:      wm.Image,
						ImageAvatar:   wm.ImageAvatar,
						Name:          wm.Name,
						Hash:          wm.Hash,
						FactionID:     wm.FactionID,
					}
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
			Name:          killByWarMachine.Name,
			Hash:          killByWarMachine.Hash,
			FactionID:     killByWarMachine.FactionID,
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
		boiler.MechMoveCommandLogWhere.BattleID.EQ(btl.BattleID),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechMoveCommandLogColumns.CancelledAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech id", destroyedWarMachine.ID).Str("battle id", btl.BattleID).Err(err).Msg("Failed to clean up mech move command.")
	}

	// broadcast changes
	err = btl.arena.BroadcastFactionMechCommands(destroyedWarMachine.FactionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to broadcast faction mech commands")
	}

}

func (btl *Battle) Load() error {
	gamelog.L.Trace().Str("func", "Load").Msg("start")
	q, err := db.LoadBattleQueue(context.Background(), 3)
	ids := make([]string, len(q))
	if err != nil {
		gamelog.L.Warn().Str("battle_id", btl.ID).Err(err).Msg("unable to load out queue")
		gamelog.L.Trace().Str("func", "Load").Msg("end")
		return err
	}

	if len(q) < 9 {
		gamelog.L.Warn().Msg("not enough mechs to field a battle. replacing with default battle.")

		// build the mechs
		err = btl.QueueDefaultMechs(btl.GenerateDefaultQueueRequest(q))
		if err != nil {
			gamelog.L.Warn().Str("battle_id", btl.ID).Err(err).Msg("unable to load default mechs")
			gamelog.L.Trace().Str("func", "Load").Msg("end")
			return err
		}

		gamelog.L.Trace().Str("func", "Load").Msg("end")
		return btl.Load()
	}

	for i, bq := range q {
		ids[i] = bq.MechID
	}

	mechs, err := db.Mechs(ids...)
	if errors.Is(err, db.ErrNotAllMechsReturned) || len(mechs) != len(ids) {
		for _, m := range mechs {
			for i, v := range ids {
				if v == m.ID {
					ids = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
		_, err = boiler.BattleQueues(boiler.BattleQueueWhere.MechID.IN(ids)).DeleteAll(gamedb.StdConn)
		if err != nil {
			gamelog.L.Panic().Strs("mechIDs", ids).Err(err).Msg("unable to delete mech from queue")
		}

		gamelog.L.Trace().Str("func", "Load").Msg("end")
		return btl.Load()
	}

	if err != nil {
		gamelog.L.Warn().Interface("mechs_ids", ids).Str("battle_id", btl.ID).Err(err).Msg("failed to retrieve mechs from mech ids")
		gamelog.L.Trace().Str("func", "Load").Msg("end")
		return err
	}
	btl.WarMachines = btl.MechsToWarMachines(mechs)
	uuids := make([]uuid.UUID, len(q))
	mechIDs := make([]string, len(q))
	for i, bq := range q {
		mechIDs[i] = bq.MechID
		uuids[i], err = uuid.FromString(bq.MechID)
		if err != nil {
			gamelog.L.Warn().Str("mech_id", bq.MechID).Msg("failed to convert mech id string to uuid")
			gamelog.L.Trace().Str("func", "Load").Msg("end")
			return err
		}
	}

	// set mechs current health
	rcs, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(mechIDs),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load mech repair cases.")
	}

	if rcs != nil {
		for _, rc := range rcs {
			for _, wm := range btl.WarMachines {
				if rc.MechID == wm.ID {
					totalBlocks := db.TotalRepairBlocks(rc.MechID)
					wm.Health = wm.MaxHealth * uint32(totalBlocks-(rc.BlocksRequiredRepair-rc.BlocksRepaired)) / uint32(totalBlocks)
					break
				}
			}
		}

		_, err = rcs.UpdateAll(gamedb.StdConn, boiler.M{boiler.RepairCaseColumns.CompletedAt: null.TimeFrom(time.Now())})
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to update mech repair cases.")
		}
	}

	btl.warMachineIDs = uuids
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
			ID:          mech.ID,
			Hash:        mech.Hash,
			OwnedByID:   mech.OwnerID,
			Name:        TruncateString(mech.Name, 20),
			Label:       mech.Label,
			FactionID:   mech.FactionID.String,
			MaxHealth:   uint32(mech.MaxHitpoints),
			Health:      uint32(mech.MaxHitpoints),
			Speed:       mech.Speed,
			Tier:        mech.Tier,
			Image:       mech.ChassisSkin.ImageURL.String,
			ImageAvatar: mech.ChassisSkin.AvatarURL.String,

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
			Weapons:   WeaponsFromServer(mech.Weapons),
			Utility:   UtilitiesFromServer(mech.Utility),
			Stats: &Stats{
				TotalWins:       mech.Stats.TotalWins,
				TotalDeaths:     mech.Stats.TotalDeaths,
				TotalKills:      mech.Stats.TotalKills,
				BattlesSurvived: mech.Stats.BattlesSurvived,
				TotalLosses:     mech.Stats.TotalLosses,
			},
		}
		// set shield (assume for frontend, not game client)
		for _, utl := range mech.Utility {
			if utl.Type == boiler.UtilityTypeSHIELD && utl.Shield != nil {
				newWarMachine.Shield = uint32(utl.Shield.Hitpoints)
				newWarMachine.MaxShield = uint32(utl.Shield.Hitpoints)
				newWarMachine.ShieldRechargeRate = uint32(utl.Shield.RechargeRate)
			}
		}

		// add owner username
		if mech.Owner != nil {
			newWarMachine.OwnerUsername = fmt.Sprintf("%s#%d", mech.Owner.Username, mech.Owner.Gid)
		}

		// check model
		if mech.Model != nil {
			model, ok := ModelMap[mech.Model.Label]
			if !ok {
				model = "WREX"
			}
			newWarMachine.Model = model
			newWarMachine.ModelID = mech.ModelID
		}

		// check model skin
		if mech.ChassisSkin != nil {
			mappedSkin, ok := SubmodelSkinMap[mech.ChassisSkin.Label]
			if ok {
				newWarMachine.Skin = mappedSkin
			}
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
