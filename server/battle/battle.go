package battle

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/multipliers"
	"server/rpcclient"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"go.uber.org/atomic"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"

	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type BattleStage int32

const (
	BattleStagStart = 1
	BattleStageEnd  = 0
)

type Battle struct {
	arena         *Arena
	stage         *atomic.Int32
	BattleID      string        `json:"battleID"`
	MapName       string        `json:"mapName"`
	WarMachines   []*WarMachine `json:"warMachines"`
	SpawnedAI     []*WarMachine `json:"SpawnedAI"`
	warMachineIDs []uuid.UUID   `json:"ids"`
	lastTick      *[]byte
	gameMap       *server.GameMap
	_abilities    *AbilitiesSystem
	//_battleSeconds        decimal.Decimal
	//battleSecondCloseChan chan bool
	users          usersMap
	factions       map[uuid.UUID]*boiler.Faction
	multipliers    *MultiplierSystem
	spoils         *SpoilsOfWar
	rpcClient      *rpcclient.XrpcClient
	battleMechData []*db.BattleMechData
	startedAt      time.Time

	destroyedWarMachineMap map[byte]*WMDestroyedRecord
	*boiler.Battle

	inserted bool

	viewerCountInputChan chan *ViewerLiveCount
	sync.RWMutex
}

//type BattleSeconds struct {
//	Current decimal.Decimal
//	deadlock.RWMutex
//}

//func (btl *Battle) battleSeconds() decimal.Decimal {
//	btl.RLock()
//	defer btl.RUnlock()
//	return btl._battleSeconds
//}

//func (btl *Battle) storeBattleSeconds(d decimal.Decimal) {
//	btl.Lock()
//	defer btl.Unlock()
//	btl._battleSeconds = d
//}

//func (btl *Battle) increaseBattleSeconds(i int64) decimal.Decimal {
//	btl.Lock()
//	defer btl.Unlock()
//	btl._battleSeconds = btl._battleSeconds.Add(decimal.NewFromInt(i))
//	return btl._battleSeconds
//}

func (btl *Battle) abilities() *AbilitiesSystem {
	btl.RLock()
	defer btl.RUnlock()
	return btl._abilities
}

func (btl *Battle) storeAbilities(as *AbilitiesSystem) {
	btl.Lock()
	defer btl.Unlock()
	btl._abilities = as
}

const HubKeyLiveVoteCountUpdated hub.HubCommandKey = "LIVE:VOTE:COUNT:UPDATED"
const HubKeyWarMachineLocationUpdated hub.HubCommandKey = "WAR:MACHINE:LOCATION:UPDATED"

func (btl *Battle) preIntro(payload *BattleStartPayload) error {
	bmd := make([]*db.BattleMechData, len(btl.WarMachines))
	factions := map[uuid.UUID]*boiler.Faction{}

	for i, wm := range btl.WarMachines {
		if payload.WarMachines[i].Hash == wm.Hash {
			btl.WarMachines[i].ParticipantID = payload.WarMachines[i].ParticipantID
		} else {
			for _, wm2 := range payload.WarMachines {
				if wm2.Hash == wm.Hash {
					btl.WarMachines[i].ParticipantID = wm2.ParticipantID
					break
				}
			}
		}
		wm.ParticipantID = payload.WarMachines[i].ParticipantID
		mechID, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().Str("ownerID", wm.ID).Err(err).Msg("unable to convert owner id from string")
			return terror.Error(err)
		}

		ownerID, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().Str("ownerID", wm.OwnedByID).Err(err).Msg("unable to convert owner id from string")
			return terror.Error(err)
		}

		factionID, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().Str("factionID", wm.FactionID).Err(err).Msg("unable to convert faction id from string")
			return terror.Error(err)
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
				gamelog.L.Error().
					Str("Battle ID", btl.ID).
					Str("Faction ID", factionID.String()).
					Err(err).Msg("unable to retrieve faction from database")

			}
			factions[factionID] = faction
		}
	}

	btl.factions = factions
	btl.battleMechData = bmd

	return nil
}

func (btl *Battle) start() {
	var err error
	btl.startedAt = time.Now()
	if btl.inserted {
		_, err := btl.Battle.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Interface("battle", btl).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to update Battle in database")
			return
		}

		// clean up battle contributions
		contributions, err := boiler.BattleContributions(
			boiler.BattleContributionWhere.BattleID.EQ(btl.ID),
			boiler.BattleContributionWhere.RefundTransactionID.IsNull(),
			qm.OrderBy(boiler.BattleContributionColumns.PlayerID),
		).All(gamedb.StdConn)
		for _, c := range contributions {
			if !c.TransactionID.Valid {
				gamelog.L.Warn().
					Str("contribution_id", c.ID).
					Err(err).
					Msg("contribution does not have a transaction id")
				continue
			}
			contributeRefundTransactionID, err := btl.arena.RPCClient.RefundSupsMessage(c.TransactionID.String)
			if err != nil {
				gamelog.L.Error().
					Str("queue_transaction_id", c.TransactionID.String).
					Err(err).
					Msg("failed to refund users queue fee")
			}
			c.RefundTransactionID = null.StringFrom(contributeRefundTransactionID)
			if _, err := c.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleContributionColumns.RefundTransactionID)); err != nil {
				gamelog.L.Error().
					Str("battle_contributions_id", c.ID).
					Err(err).
					Msg("failed to save refund")
			}
		}

		// empty spoils of war
		sow, err := boiler.SpoilsOfWars(boiler.SpoilsOfWarWhere.BattleID.EQ(btl.ID)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().
				Str("battle_id", btl.ID).
				Err(err).
				Msg("unable to retrieve spoil of war for unfinished battle")
		} else {
			sow.Amount = decimal.Zero
			sow.AmountSent = decimal.Zero
			if _, err := sow.Update(gamedb.StdConn, boil.Infer()); err != nil {
				gamelog.L.Error().
					Str("spoils_of_war_id", sow.ID).
					Err(err).
					Msg("failed to clear spoils of war for unfinished battle")
			}
		}

		bmds, err := boiler.BattleMechs(boiler.BattleMechWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err == nil {
			_, err = bmds.DeleteAll(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to delete delete stale battle mechs from database")
			}
		}

		bws, err := boiler.BattleWins(boiler.BattleWinWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err == nil {
			_, err = bws.DeleteAll(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to delete delete stale battle wins from database")
			}
		}

		bks, err := boiler.BattleKills(boiler.BattleKillWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err == nil {
			_, err = bks.DeleteAll(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to delete delete stale battle kills from database")
			}
		}

		bhs, err := boiler.BattleHistories(boiler.BattleHistoryWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err == nil {
			_, err = bhs.DeleteAll(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to delete delete stale battle historys from database")
			}
		}
	} else {
		err := btl.Battle.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Interface("battle", btl).Str("battle.go", ":battle.go:battle.Battle()").Err(err).Msg("unable to insert Battle into database")
			return
		}

		btl.inserted = true

		// insert current users to
		btl.users.Range(func(user *BattleUser) bool {
			err = db.BattleViewerUpsert(context.Background(), gamedb.Conn, btl.ID, user.ID.String())
			if err != nil {
				gamelog.L.Error().Str("battle_id", btl.ID).Str("player_id", user.ID.String()).Err(err).Msg("to upsert battle view")
				return true
			}
			return true
		})

		err = db.QueueSetBattleID(btl.ID, btl.warMachineIDs...)
		if err != nil {
			gamelog.L.Error().Interface("mechs_ids", btl.warMachineIDs).Str("battle_id", btl.ID).Err(err).Msg("failed to set battle id in queue")
			return
		}
	}

	// start battle seconds ticker
	//btl.battleSecondCloseChan = btl.BattleSecondStartTicking()

	// insert current users to
	btl.users.Range(func(user *BattleUser) bool {
		user.Send(HubKeyUserMultiplierSignalUpdate, true)
		return true
	})

	if btl.battleMechData == nil {
		gamelog.L.Error().Str("battlemechdata", btl.ID).Msg("battle mech data failed nil check")
	}

	err = db.BattleMechs(btl.Battle, btl.battleMechData)
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID).Err(err).Msg("unable to insert battle into database")
		//TODO: something more dramatic
	}

	// set up the abilities for current battle

	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up battle spoils")
	btl.spoils = NewSpoilsOfWar(btl, 30*time.Second, 30*time.Second)
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up battle abilities")
	btl.storeAbilities(NewAbilitiesSystem(btl))
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up battle multipliers")
	btl.multipliers = NewMultiplierSystem(btl)
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting battle start to players")
	btl.BroadcastUpdate()

	// broadcast spoil of war on the start of the battle
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting spoils of war updates")

	yesterday := time.Now().Add(time.Hour * -24)
	warchests, err := boiler.SpoilsOfWars(
		boiler.SpoilsOfWarWhere.CreatedAt.GT(yesterday),
		boiler.SpoilsOfWarWhere.BattleID.NEQ(btl.ID),
		qm.And(`amount_sent < amount`),
		qm.OrderBy(fmt.Sprintf("%s %s", boiler.SpoilsOfWarColumns.CreatedAt, "DESC")),
	).All(gamedb.StdConn)

	warchest, err := boiler.SpoilsOfWars(
		boiler.SpoilsOfWarWhere.BattleID.EQ(btl.ID),
	).One(gamedb.StdConn)

	spoilOfWarPayload := []byte{byte(SpoilOfWarTick)}
	amnt := decimal.NewFromInt(0)
	for _, sow := range warchests {
		amnt = amnt.Add(sow.Amount.Sub(sow.AmountSent))
	}
	spoilOfWarPayload = append(spoilOfWarPayload, []byte(strings.Join([]string{warchest.Amount.String(), amnt.String()}, "|"))...)
	go btl.arena.messageBus.SendBinary(messagebus.BusKey(HubKeySpoilOfWarUpdated), spoilOfWarPayload)

	// handle global announcements
	ga, err := boiler.GlobalAnnouncements().One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting battle start to players")
	}

	// global announcement exists
	if ga != nil {
		const HubKeyGlobalAnnouncementSubscribe hub.HubCommandKey = "GLOBAL_ANNOUNCEMENT:SUBSCRIBE"

		// show if battle number is equal or in between the global announcement's to and from battle number
		if btl.BattleNumber >= ga.ShowFromBattleNumber.Int && btl.BattleNumber <= ga.ShowUntilBattleNumber.Int {
			go btl.arena.messageBus.Send(messagebus.BusKey(HubKeyGlobalAnnouncementSubscribe), ga)
		}

		// delete if global announcement expired/ is in the past
		if btl.BattleNumber > ga.ShowUntilBattleNumber.Int {
			_, err := boiler.GlobalAnnouncements().DeleteAll(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("Battle ID", btl.ID).Msg("unable to delete global announcement")
			}

			go btl.arena.messageBus.Send(messagebus.BusKey(HubKeyGlobalAnnouncementSubscribe), nil)
		}

	}

}

//func (btl *Battle) BattleSecondStartTicking() chan bool {
//	interval := int64(1)
//	main_ticker := time.NewTicker(time.Duration(interval) * time.Second)
//	closeChan := make(chan bool)
//
//	go func() {
//		for {
//			select {
//			case <-main_ticker.C:
//				btl.increaseBattleSeconds(interval)
//			case <-closeChan:
//				main_ticker.Stop()
//				return
//			}
//		}
//	}()
//
//	return closeChan
//}

// calcTriggeredLocation convert picked cell to the location in game
func (btl *Battle) calcTriggeredLocation(abilityEvent *server.GameAbilityEvent) {
	// To get the location in game its
	//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
	//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels
	if abilityEvent.TriggeredOnCellX == nil || abilityEvent.TriggeredOnCellY == nil {
		return
	}

	abilityEvent.GameLocation.X = ((*abilityEvent.TriggeredOnCellX * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + btl.gameMap.LeftPixels
	abilityEvent.GameLocation.Y = ((*abilityEvent.TriggeredOnCellY * server.GameClientTileSize) + (server.GameClientTileSize / 2)) + btl.gameMap.TopPixels

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
		if wm.FactionID != server.RedMountainFactionID.String() || wm.Position == nil {
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
		abilityEvent.GameLocation = struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{
			X: wm.X,
			Y: wm.Y,
		}

		return
	}

	wm := rmw[rand.Intn(len(rmw))]
	// set cell
	abilityEvent.TriggeredOnCellX = &wm.X
	abilityEvent.TriggeredOnCellY = &wm.Y

	abilityEvent.GameLocation = struct {
		X int `json:"x"`
		Y int `json:"y"`
	}{
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
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the battle abilities end!", r)
		}
	}()

	gamelog.L.Info().Msgf("cleaning up abilities: %s", btl.ID)

	if btl.abilities == nil {
		gamelog.L.Error().Msg("battle did not have abilities!")
		return
	}

	btl.abilities().End()
	btl.abilities().storeBattle(nil)
	btl.storeAbilities(nil)
}
func (btl *Battle) endSpoils() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the spoils end!", r)
		}
	}()
	gamelog.L.Info().Msgf("cleaning up spoils: %s", btl.ID)

	if btl.spoils == nil {
		gamelog.L.Error().Msg("battle did not have spoils!")
		return
	}

	btl.spoils.End()
	btl.spoils = nil
}

func (btl *Battle) endCreateStats(payload *BattleEndPayload, winningWarMachines []*WarMachine) *BattleEndDetail {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the creation of ending info: endCreateStats!", r)
		}
	}()
	gamelog.L.Info().Msgf("battle end: looping TopSupsContributeFactions: %s", btl.ID)
	topFactionContributorBoilers, err := db.TopSupsContributeFactions(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top faction contributors")
	}
	gamelog.L.Info().Msgf("battle end: looping topPlayerContributorsBoilers: %s", btl.ID)
	topPlayerContributorsBoilers, err := db.TopSupsContributors(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top player contributors")
	}
	gamelog.L.Info().Msgf("battle end: looping MostFrequentAbilityExecutors: %s", btl.ID)
	topPlayerExecutorsBoilers, err := db.MostFrequentAbilityExecutors(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top player executors")
	}

	topFactionContributors := []*Faction{}
	gamelog.L.Info().Msgf("battle end: looping topFactionContributorBoilers: %s", btl.ID)
	for _, f := range topFactionContributorBoilers {
		topFactionContributors = append(topFactionContributors, &Faction{
			ID:    f.ID,
			Label: f.Label,
			Theme: &FactionTheme{
				Primary:    f.PrimaryColor,
				Secondary:  f.SecondaryColor,
				Background: f.BackgroundColor,
			},
		})
	}
	topPlayerContributors := []*BattleUser{}

	gamelog.L.Info().Msgf("battle end: looping topPlayerContributorsBoilers: %s", btl.ID)
	for _, p := range topPlayerContributorsBoilers {
		factionID := uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))
		if p.FactionID.Valid {
			factionID = uuid.Must(uuid.FromString(p.FactionID.String))
		}

		topPlayerContributors = append(topPlayerContributors, &BattleUser{
			ID:            uuid.Must(uuid.FromString(p.ID)),
			Username:      p.Username.String,
			FactionID:     factionID.String(),
			FactionColour: btl.factions[factionID].PrimaryColor,
			FactionLogoID: FactionLogos[factionID.String()],
		})
	}

	gamelog.L.Info().Msgf("battle end: looping topPlayerExecutorsBoilers: %s", btl.ID)
	topPlayerExecutors := []*BattleUser{}
	for _, p := range topPlayerExecutorsBoilers {
		factionID := uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))
		if p.FactionID.Valid {
			factionID = uuid.Must(uuid.FromString(p.FactionID.String))
		}
		topPlayerExecutors = append(topPlayerExecutors, &BattleUser{
			ID:            uuid.Must(uuid.FromString(p.ID)),
			Username:      p.Username.String,
			FactionID:     factionID.String(),
			FactionColour: btl.factions[factionID].PrimaryColor,
			FactionLogoID: FactionLogos[factionID.String()],
		})
	}

	gamelog.L.Debug().
		Int("top_faction_contributors", len(topFactionContributors)).
		Int("top_player_executors", len(topPlayerExecutors)).
		Int("top_player_contributors", len(topPlayerContributors)).
		Msg("get top players and factions")

	return &BattleEndDetail{
		BattleID:                     btl.ID,
		BattleIdentifier:             btl.Battle.BattleNumber,
		StartedAt:                    btl.Battle.StartedAt,
		EndedAt:                      btl.Battle.EndedAt.Time,
		WinningCondition:             payload.WinCondition,
		WinningFaction:               winningWarMachines[0].Faction,
		WinningWarMachines:           winningWarMachines,
		TopSupsContributeFactions:    topFactionContributors,
		TopSupsContributors:          topPlayerContributors,
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
			gamelog.L.Error().Str("Battle ID", btl.ID).Msg("unable to match war machine to battle with hash")
			continue
		}
		mechId, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert mech id to uuid")
			continue
		}
		ownedById, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert owned id to uuid")
			continue
		}
		factionId, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().
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

		contract, err := boiler.BattleContracts(boiler.BattleContractWhere.BattleID.EQ(
			null.StringFrom(btl.BattleID)),
			boiler.BattleContractWhere.MechID.EQ(mws[i].MechID.String()),
			boiler.BattleContractWhere.Cancelled.EQ(null.BoolFrom(false)),
		).One(gamedb.StdConn)

		if err != nil && errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().
				Str("Battle ID", btl.ID).
				Str("Mech ID", wm.ID).
				Err(err).
				Msg("no contract in database")

			continue
		} else if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID).
				Str("Mech ID", wm.ID).
				Err(err).
				Msg("failed to retrieve contract")
			continue
		}

		contract.DidWin = null.BoolFrom(true)
		factionAccountID, ok := server.FactionUsers[factionId.String()]
		if !ok {
			gamelog.L.Error().
				Str("Battle ID", btl.ID).
				Str("faction ID", wm.FactionID).
				Msg("unable to get hard coded syndicate player ID from faction ID")
		} else {
			//do contract payout for winning mech
			gamelog.L.Info().
				Str("Battle ID", btl.ID).
				Str("Faction ID", wm.FactionID).
				Str("Faction Account ID", factionAccountID).
				Str("Player ID", wm.OwnedByID).
				Str("Contract ID", contract.ID).
				Str("Amount", contract.ContractReward.StringFixed(0)).
				Msg("paying out mech winnings from contract reward")

			factID := uuid.Must(uuid.FromString(factionAccountID))
			syndicateBalance := btl.arena.RPCClient.UserBalanceGet(factID)

			if syndicateBalance.LessThanOrEqual(contract.ContractReward) {
				txid, err := btl.arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
					FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
					ToUserID:             factID,
					Amount:               contract.ContractReward.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("contract_rewards|%s|%d", contract.ID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupBattle),
					SubGroup:             wmwin.Hash,
					Description:          fmt.Sprintf("Mech won battle #%d", btl.BattleNumber),
					NotSafe:              false,
				})
				if err != nil {
					gamelog.L.Error().
						Str("Faction ID", factionAccountID).
						Str("Amount", contract.ContractReward.StringFixed(0)).
						Err(err).
						Msg("Could not transfer money from treasury into syndicate account!!")
					continue
				}
				gamelog.L.Warn().
					Str("Faction ID", factionAccountID).
					Str("Amount", contract.ContractReward.StringFixed(0)).
					Str("TXID", txid).
					Err(err).
					Msg("Had to transfer funds to the syndicate account")
			}

			// pay sups
			txid, err := btl.arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
				FromUserID:           factID,
				ToUserID:             uuid.Must(uuid.FromString(contract.PlayerID)),
				Amount:               contract.ContractReward.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("contract_rewards|%s|%d", contract.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             wmwin.Hash,
				Description:          fmt.Sprintf("Mech won battle #%d", btl.BattleNumber),
				NotSafe:              false,
			})
			if err != nil {
				gamelog.L.Error().
					Str("Battle ID", btl.ID).
					Str("faction ID", wm.FactionID).
					Str("Player ID", wm.OwnedByID).
					Err(err).
					Msg("unable to transfer funds to winning mech owner")
				continue
			}

			contract.PaidOut = true
			contract.TransactionID = null.StringFrom(txid)
			_, err = contract.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().
					Str("Battle ID", btl.ID).
					Str("faction ID", wm.FactionID).
					Str("Player ID", wm.OwnedByID).
					Str("TX ID", txid).
					Err(err).
					Msg("unable to save transaction ID on contract")
				continue
			}
		}
	}
	err := db.WinBattle(btl.ID, payload.WinCondition, mws...)
	if err != nil {
		gamelog.L.Error().
			Str("Battle ID", btl.ID).
			Err(err).
			Msg("unable to store mech wins")
	}
}

func (btl *Battle) endWarMachines(payload *BattleEndPayload) []*WarMachine {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the sorting up ending war machines!", r)
		}
	}()
	winningWarMachines := make([]*WarMachine, len(payload.WinningWarMachines))

	gamelog.L.Info().Msgf("battle end: looping WinningWarMachines: %s", btl.ID)
	for i := range payload.WinningWarMachines {
		for _, w := range btl.WarMachines {
			if w.Hash == payload.WinningWarMachines[i].Hash {
				winningWarMachines[i] = w
				break
			}
		}
		if winningWarMachines[i] == nil {
			gamelog.L.Error().Str("Battle ID", btl.ID).Msg("unable to match war machine to battle with hash")
		}
	}

	if len(winningWarMachines) == 0 || winningWarMachines[0] == nil {
		gamelog.L.Panic().Str("Battle ID", btl.ID).Msg("no winning war machines")
	} else {
		for _, w := range winningWarMachines {
			// update battle_mechs to indicate survival
			bm, err := boiler.FindBattleMech(gamedb.StdConn, btl.ID, w.ID)
			if err != nil {
				gamelog.L.Error().
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

			bqn, err := boiler.BattleQueueNotifications(boiler.BattleQueueNotificationWhere.MechID.EQ(bm.MechID), qm.OrderBy(boiler.BattleQueueNotificationColumns.SentAt+" DESC")).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("bm.MechID", bm.MechID).Err(err).Msg("failed to get BattleQueueNotifications")
			} else {
				if bqn.TelegramNotificationID.Valid {
					// killed a war machine
					msg := fmt.Sprintf("Your War machine %s is Victorious! 🎉", w.Name)
					err := btl.arena.telegram.Notify(bqn.TelegramNotificationID.String, msg)
					if err != nil {
						gamelog.L.Error().Str("bqn.TelegramNotificationID.String", bqn.TelegramNotificationID.String).Err(err).Msg("failed to send notification")
					}
				}
			}
		}

		// update battle_mechs to indicate faction win
		bms, err := boiler.BattleMechs(boiler.BattleMechWhere.FactionID.EQ(winningWarMachines[0].FactionID), boiler.BattleMechWhere.BattleID.EQ(btl.ID)).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().
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
			gamelog.L.Error().
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
			gamelog.L.Error().
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

func (btl *Battle) endMultis(endInfo *BattleEndDetail) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the ending of multis! btl.endMultis!", r)
		}
	}()
	gamelog.L.Info().Msgf("cleaning up multipliers: %s", btl.ID)

	if btl.multipliers == nil {
		gamelog.L.Error().Msg("battle did not have multipliers!")
		return
	}

	btl.multipliers.end(endInfo)
}
func (btl *Battle) endBroadcast(endInfo *BattleEndDetail) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the ending of end broadcast!", r)
		}
	}()
	btl.endInfoBroadcast(*endInfo)
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

	//btl.battleSecondCloseChan <- true
	//btl.Battle.EndedBattleSeconds = decimal.NullDecimal{btl.battleSeconds(), true}
	btl.Battle.EndedAt = null.TimeFrom(time.Now())
	_, err := btl.Battle.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID).Time("EndedAt", btl.EndedAt.Time).Msg("unable to update database for endat battle")
	}

	btl.endAbilities()
	btl.endSpoils()

	winningWarMachines := btl.endWarMachines(payload)
	endInfo := btl.endCreateStats(payload, winningWarMachines)

	btl.processWinners(payload)
	btl.endMultis(endInfo)

	btl.insertUserSpoils(endInfo)

	_, err = boiler.BattleQueues(boiler.BattleQueueWhere.BattleID.EQ(null.StringFrom(btl.BattleID))).DeleteAll(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Str("Battle ID", btl.ID).Str("battle_id", payload.BattleID).Msg("Failed to remove mechs from battle queue.")
	}

	gamelog.L.Info().Msgf("battle has been cleaned up, sending broadcast %s", btl.ID)
	btl.endBroadcast(endInfo)
}

// insertUserSpoils gets the spoils for given battle, gets players multis for given battle and calculates and inserts the user spoils for this battle
func (btl *Battle) insertUserSpoils(btlEndInfo *BattleEndDetail) {
	// get battle sow
	spoils, err := boiler.SpoilsOfWars(boiler.SpoilsOfWarWhere.BattleID.EQ(btlEndInfo.BattleID)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("btlEndInfo.BattleID", btlEndInfo.BattleID).
			Int("btlEndInfo.BattleIdentifier", btlEndInfo.BattleIdentifier).
			Err(err).
			Msg("issue getting SpoilsOfWars")
		return
	}

	// get player multies
	playerMultis, err := multipliers.GetPlayersMultiplierSummaryForBattle(btlEndInfo.BattleIdentifier)
	if err != nil {
		gamelog.L.Error().
			Str("btlEndInfo.BattleID", btlEndInfo.BattleID).
			Int("btlEndInfo.BattleIdentifier", btlEndInfo.BattleIdentifier).
			Err(err).
			Msg("issue getting PlayerMultipliers")
		return
	}

	oneMultiWorth := multipliers.CalculateOneMultiWorth(playerMultis, spoils.Amount)

	totalAssignedToPlayers := decimal.Zero

	for _, player := range playerMultis {
		playerTotalSow := multipliers.CalculateMultipliersWorth(oneMultiWorth, player.TotalMultiplier)
		userSpoils := &boiler.UserSpoilsOfWar{
			PlayerID:                 player.PlayerID,
			BattleID:                 btlEndInfo.BattleID,
			TotalMultiplierForBattle: int(player.TotalMultiplier.IntPart()),
			TotalSow:                 playerTotalSow,
			PaidSow:                  decimal.Zero,
			TickAmount:               playerTotalSow.Div(decimal.NewFromInt(int64(spoils.MaxTicks))),
			LostSow:                  decimal.Zero,
		}

		err := userSpoils.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Interface("userSpoils", userSpoils).
				Err(err).
				Msg("issue inserting userSpoils")
		}
		totalAssignedToPlayers = totalAssignedToPlayers.Add(playerTotalSow)
	}

	if totalAssignedToPlayers.Equal(spoils.Amount) { // we gucci
		return
	} else if totalAssignedToPlayers.GreaterThan(spoils.Amount) { // if we assigned too much, panic because shit broke
		gamelog.L.Panic().
			Str("totalAssignedToPlayers", totalAssignedToPlayers.String()).
			Str("spoils.Amount", spoils.Amount.String()).
			Err(fmt.Errorf("assigned more sups than what is in the spoils")).
			Msg("issue assigning spoils")
	} else if totalAssignedToPlayers.LessThan(spoils.Amount) { // we didn't give them all out
		gamelog.L.Error().
			Str("totalAssignedToPlayers", totalAssignedToPlayers.String()).
			Str("spoils.Amount", spoils.Amount.String()).
			Err(fmt.Errorf("assigned less sups than what is in the spoils")).
			Msg("issue assigning spoils")
	}
}

const HubKeyBattleEndDetailUpdated hub.HubCommandKey = "BATTLE:END:DETAIL:UPDATED"

func (btl *Battle) endInfoBroadcast(info BattleEndDetail) {
	btl.users.Range(func(user *BattleUser) bool {

		m, total, _ := multipliers.GetPlayerMultipliersForBattle(user.ID.String(), btl.BattleNumber)

		info.MultiplierUpdate = &MultiplierUpdate{
			Battles: []*MultiplierUpdateBattles{
				{
					BattleNumber:     btl.BattleNumber,
					TotalMultipliers: multipliers.FriendlyFormatMultiplier(total),
					UserMultipliers:  m,
				},
			},
		}

		user.Send(HubKeyBattleEndDetailUpdated, info)

		// broadcast user stat to user
		go func(user *BattleUser) {
			us, err := db.UserStatsGet(user.ID.String())
			if err != nil {
				gamelog.L.Error().Str("player_id", user.ID.String()).Err(err).Msg("Failed to get user stats")
			}
			if us != nil {
				btl.arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, us.ID)), us)
			}
		}(user)

		return true
	})

}

type BroadcastPayload struct {
	Key     hub.HubCommandKey `json:"key"`
	Payload interface{}       `json:"payload"`
}

type GameSettingsResponse struct {
	GameMap            *server.GameMap `json:"game_map"`
	WarMachines        []*WarMachine   `json:"war_machines"`
	SpawnedAI          []*WarMachine   `json:"spawned_ai"`
	WarMachineLocation []byte          `json:"war_machine_location"`
	BattleIdentifier   int             `json:"battle_identifier"`
}

type ViewerLiveCount struct {
	RedMountain int64 `json:"red_mountain"`
	Boston      int64 `json:"boston"`
	Zaibatsu    int64 `json:"zaibatsu"`
	Other       int64 `json:"other"`
}

func (btl *Battle) userOnline(user *BattleUser, wsc *hub.Client) {
	exists := false
	u, ok := btl.users.User(user.ID)
	if !ok {
		user.wsClient[wsc] = true
		btl.users.Add(user)
	} else {
		// do not upsert battle viewer or broadcast viewer count if user is already counted
		u.Lock()
		u.wsClient[wsc] = true
		u.Unlock()
		exists = true
	}

	if btl.inserted {
		err := db.BattleViewerUpsert(context.Background(), gamedb.Conn, btl.ID, wsc.Identifier())
		if err != nil {
			gamelog.L.Error().
				Str("battle_id", btl.ID).
				Str("player_id", wsc.Identifier()).
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

	if !exists {
		// send result to broadcast debounce function
		btl.viewerCountInputChan <- resp
	} else {
		// broadcast result to current user only if the user already exists
		btl.users.Send(HubKeyViewerLiveCountUpdated, resp, user.ID)
	}
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
			if btl != btl.arena.currentBattle() {
				timer.Stop()
				checker.Stop()
				gamelog.L.Info().Msg("Clean up live count debounce function due to battle missmatch")
				return
			}
		}
	}
}

func UpdatePayload(btl *Battle) *GameSettingsResponse {
	var lt []byte
	if btl.lastTick != nil {
		lt = *btl.lastTick
	}
	if btl == nil {
		return nil
	}
	return &GameSettingsResponse{
		BattleIdentifier:   btl.BattleNumber,
		GameMap:            btl.gameMap,
		WarMachines:        btl.WarMachines,
		SpawnedAI:          btl.SpawnedAI,
		WarMachineLocation: lt,
	}
}

const HubKeyGameSettingsUpdated = hub.HubCommandKey("GAME:SETTINGS:UPDATED")
const HubKeyGameUserOnline = hub.HubCommandKey("GAME:ONLINE")

func (btl *Battle) BroadcastUpdate() {
	btl.arena.messageBus.Send(messagebus.BusKey(HubKeyGameSettingsUpdated), UpdatePayload(btl))
}

func (btl *Battle) Tick(payload []byte) {
	// Save to history
	// btl.BattleHistory = append(btl.BattleHistory, payload)

	broadcast := false
	// broadcast
	if btl.lastTick == nil {
		broadcast = true
	}
	btl.lastTick = &payload

	btl.arena.messageBus.SendBinary(messagebus.BusKey(HubKeyWarMachineLocationUpdated), payload)

	// Update game settings (so new players get the latest position, health and shield of all warmachines)
	count := payload[1]
	var c byte
	offset := 2
	for c = 0; c < count; c++ {
		participantID := payload[offset]
		offset++

		// Get Warmachine Index
		warMachineIndex := -1
		for i, wmn := range btl.WarMachines {
			if wmn.ParticipantID == participantID {
				warMachineIndex = i
				break
			}
		}

		// Get Sync byte (tells us which data was updated for this warmachine)
		syncByte := payload[offset]
		offset++

		// Position + Yaw
		if syncByte >= 100 {
			x := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			y := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4
			rotation := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
			offset += 4

			if warMachineIndex != -1 {
				if btl.WarMachines[warMachineIndex].Position == nil {
					btl.WarMachines[warMachineIndex].Position = &server.Vector3{}
				}
				btl.WarMachines[warMachineIndex].Position.X = x
				btl.WarMachines[warMachineIndex].Position.Y = y
				btl.WarMachines[warMachineIndex].Rotation = rotation
			}
		}
		// Health
		if syncByte == 1 || syncByte == 11 || syncByte == 101 || syncByte == 111 {
			health := binary.BigEndian.Uint32(payload[offset : offset+4])
			offset += 4
			if warMachineIndex != -1 {
				btl.WarMachines[warMachineIndex].Health = health
			}
		}
		// Shield
		if syncByte == 10 || syncByte == 11 || syncByte == 110 || syncByte == 111 {
			shield := binary.BigEndian.Uint32(payload[offset : offset+4])
			offset += 4
			if warMachineIndex != -1 {
				btl.WarMachines[warMachineIndex].Shield = shield
			}
		}
	}
	if broadcast {
		btl.BroadcastUpdate()
	}
}

func (arena *Arena) reset() {
	gamelog.L.Warn().Msg("arena state resetting")
}

func (btl *Battle) Destroyed(dp *BattleWMDestroyedPayload) {
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
				gamelog.L.Error().Str("faction_id", wm.FactionID).Err(err).Msg("failed to update faction death count")
			}
			break
		}
	}
	if destroyedWarMachine == nil {
		gamelog.L.Warn().Str("hash", dHash).Msg("can't match destroyed mech with battle state")
		return
	}
	bqn, err := boiler.BattleQueueNotifications(boiler.BattleQueueNotificationWhere.MechID.EQ(destroyedWarMachine.ID), qm.OrderBy(boiler.BattleQueueNotificationColumns.SentAt+" DESC")).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("destroyedWarMachine.ID", destroyedWarMachine.ID).Err(err).Msg("failed to get BattleQueueNotifications")
	}
	if bqn != nil && bqn.TelegramNotificationID.Valid {
		// killed a war machine
		msg := fmt.Sprintf("Your War machine %s has been destroyed ☠️", destroyedWarMachine.Name)
		err := btl.arena.telegram.Notify(bqn.TelegramNotificationID.String, msg)
		if err != nil {
			gamelog.L.Error().Str("bqn.TelegramNotificationID.String", bqn.TelegramNotificationID.String).Err(err).Msg("failed to send telegram notification")
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
						gamelog.L.Error().Str("player_id", wm.OwnedByID).Err(err).Msg("Failed to update user mech kill count")
					}

					// add faction kill count
					err = db.FactionAddMechKillCount(killByWarMachine.FactionID)
					if err != nil {
						gamelog.L.Error().Str("faction_id", killByWarMachine.FactionID).Err(err).Msg("failed to update faction mech kill count")
					}
					bqn, err := boiler.BattleQueueNotifications(boiler.BattleQueueNotificationWhere.MechID.EQ(wm.ID), qm.OrderBy(boiler.BattleQueueNotificationColumns.SentAt+" DESC")).One(gamedb.StdConn)
					if err != nil {
						if !errors.Is(err, sql.ErrNoRows) {
							gamelog.L.Error().Str("wm.ID", wm.ID).Err(err).Msg("failed to get BattleQueueNotifications")
						}
					} else {
						if bqn.TelegramNotificationID.Valid {
							// killed a war machine
							msg := fmt.Sprintf("Your War machine destroyed %s \U0001F9BE ", destroyedWarMachine.Name)
							err := btl.arena.telegram.Notify(bqn.TelegramNotificationID.String, msg)
							if err != nil {
								gamelog.L.Error().Str("bqn.TelegramNotificationID.String", bqn.TelegramNotificationID.String).Err(err).Msg("failed to send notification")
							}
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
			gamelog.L.Error().Str("related event id", dp.DestroyedWarMachineEvent.RelatedEventIDString).Err(err).Msg("Failed get ability from offering id")
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
					gamelog.L.Error().Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to subtract user ability kill count")
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
					gamelog.L.Error().Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to insert player last seven days kill record- (TEAM KILL)")
				}

				// subtract faction kill count
				err = db.FactionSubtractAbilityKillCount(abl.FactionID)
				if err != nil {
					gamelog.L.Error().Str("faction_id", abl.FactionID).Err(err).Msg("Failed to subtract user ability kill count")
				}

			} else {
				// update user kill
				_, err := db.UserStatAddAbilityKill(abl.PlayerID.String)
				if err != nil {
					gamelog.L.Error().Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to add user ability kill count")
				}

				// insert a team kill record to last seven days kills
				lastSevenDaysKill := boiler.PlayerKillLog{
					PlayerID:  abl.PlayerID.String,
					FactionID: abl.FactionID,
					BattleID:  btl.BattleID,
				}
				err = lastSevenDaysKill.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to insert player last seven days kill record- (ABILITY KILL)")
				}

				// add faction kill count
				err = db.FactionAddAbilityKillCount(abl.FactionID)
				if err != nil {
					gamelog.L.Error().Str("faction_id", abl.FactionID).Err(err).Msg("Failed to add faction ability kill count")
				}
			}

			// broadcast player stat to the player
			us, err := db.UserStatsGet(currentUser.ID.String())
			if err != nil {
				gamelog.L.Error().Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to get player current stat")
			}
			if us != nil {
				btl.arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, us.ID)), us)
			}
		}

	}

	gamelog.L.Info().Msgf("battle Update: %s - War Machine Destroyed: %s", btl.ID, dHash)

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
		gamelog.L.Error().
			Err(err).
			Str("battle_id", btl.ID).
			Interface("mech_id", warMachineID).
			Bool("killed", true).
			Msg("can't update battle mech")

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
			Faction: &FactionBrief{
				ID:    destroyedWarMachine.FactionID,
				Label: destroyedWarMachine.Faction.Label,
				Theme: destroyedWarMachine.Faction.Theme,
			},
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
						Faction: &FactionBrief{
							ID:    wm.FactionID,
							Label: wm.Faction.Label,
							Theme: wm.Faction.Theme,
						},
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
			Faction: &FactionBrief{
				ID:    killByWarMachine.FactionID,
				Label: killByWarMachine.Faction.Label,
				Theme: killByWarMachine.Faction.Theme,
			},
		}
	}

	// cache destroyed war machine
	btl.destroyedWarMachineMap[wmd.DestroyedWarMachine.ParticipantID] = wmd

	// broadcast destroy detail
	btl.arena.messageBus.Send(
		messagebus.BusKey(
			fmt.Sprintf(
				"%s:%x",
				HubKeyWarMachineDestroyedUpdated,
				destroyedWarMachine.ParticipantID,
			),
		),
		wmd,
	)

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

}

func (btl *Battle) Load() error {
	q, err := db.LoadBattleQueue(context.Background(), 3)
	ids := make([]uuid.UUID, len(q))
	if err != nil {
		gamelog.L.Warn().Str("battle_id", btl.ID).Err(err).Msg("unable to load out queue")
		return terror.Error(err)
	}

	if len(q) < 9 {
		gamelog.L.Warn().Msg("not enough mechs to field a battle. replacing with default battle.")

		err = btl.DefaultMechs()
		if err != nil {
			gamelog.L.Warn().Str("battle_id", btl.ID).Err(err).Msg("unable to load default mechs")
			return terror.Error(err)
		}

		return nil
	}

	for i, bq := range q {
		ids[i], err = uuid.FromString(bq.MechID)
		if err != nil {
			gamelog.L.Warn().Str("mech_id", bq.MechID).Msg("failed to convert mech id string to uuid")
			return terror.Error(err)
		}
	}

	mechs, err := db.Mechs(ids...)
	if err != nil {
		gamelog.L.Warn().Interface("mechs_ids", ids).Str("battle_id", btl.ID).Err(err).Msg("failed to retrieve mechs from mech ids")
		return terror.Error(err)
	}
	btl.WarMachines = btl.MechsToWarMachines(mechs)
	btl.warMachineIDs = ids

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

func (btl *Battle) MechsToWarMachines(mechs []*server.MechContainer) []*WarMachine {
	warmachines := make([]*WarMachine, len(mechs))
	for i, mech := range mechs {
		label := mech.Faction.Label
		if label == "" {
			gamelog.L.Warn().Interface("faction_id", mech.Faction.ID).Str("battle_id", btl.ID).Msg("mech faction is an empty label")
		}
		if len(label) > 10 {
			words := strings.Split(label, " ")
			label = ""
			for i, word := range words {
				if i == 0 {
					label = word
					continue
				}
				if i%1 == 0 {
					label = label + " " + word
					continue
				}
				label = label + "\n" + word
			}
		}

		weaponNames := make([]string, len(mech.Weapons))
		for k, wpn := range mech.Weapons {
			i, err := strconv.Atoi(k)
			if err != nil {
				gamelog.L.Warn().Str("key", k).Interface("weapon", wpn).Str("battle_id", btl.ID).Msg("mech weapon's key is not an int")
			}
			weaponNames[i] = wpn.Label
		}

		model, ok := ModelMap[mech.Chassis.Model]
		if !ok {
			model = "WREX"
		}

		mechName := mech.Name

		if len(mechName) < 3 {
			owner, err := mech.Owner().One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Warn().Str("mech_id", mech.ID).Msg("unable to retrieve mech's owner")
			} else {
				mechName = owner.Username.String
				if mechName == "" {
					mechName = fmt.Sprintf("%s%s%s", "🦾", mech.Hash, "🦾")
				}
			}
		}
		skin := mech.Chassis.Skin
		mappedSkin, ok := SubmodelSkinMap[mech.Chassis.Skin]
		if ok {
			skin = mappedSkin
		}
		warmachines[i] = &WarMachine{
			ID:            mech.ID,
			Name:          TruncateString(mechName, 20),
			Hash:          mech.Hash,
			ParticipantID: 0,
			FactionID:     mech.Faction.ID,
			MaxHealth:     uint32(mech.Chassis.MaxHitpoints),
			Health:        uint32(mech.Chassis.MaxHitpoints),
			MaxShield:     uint32(mech.Chassis.MaxShield),
			Shield:        uint32(mech.Chassis.MaxShield),
			Stat:          nil,
			OwnedByID:     mech.OwnerID,
			ImageAvatar:   mech.AvatarURL,
			Faction: &Faction{
				ID:    mech.Faction.ID,
				Label: label,
				Theme: &FactionTheme{
					Primary:    mech.Faction.PrimaryColor,
					Secondary:  mech.Faction.SecondaryColor,
					Background: mech.Faction.BackgroundColor,
				},
			},
			Speed:              mech.Chassis.Speed,
			Model:              model,
			Skin:               skin,
			ShieldRechargeRate: float64(mech.Chassis.ShieldRechargeRate),
			Durability:         mech.Chassis.MaxHitpoints,
			WeaponHardpoint:    mech.Chassis.WeaponHardpoints,
			TurretHardpoint:    mech.Chassis.TurretHardpoints,
			UtilitySlots:       mech.Chassis.UtilitySlots,
			Description:        nil,
			ExternalUrl:        "",
			Image:              mech.ImageURL,
			PowerGrid:          1,
			CPU:                1,
			WeaponNames:        weaponNames,
			Tier:               mech.Tier,
		}
		gamelog.L.Debug().Str("mech_id", mech.ID).Str("model", model).Str("skin", mech.Chassis.Skin).Msg("converted mech to warmachine")
	}

	sort.Slice(warmachines, func(i, k int) bool {
		return warmachines[i].FactionID == warmachines[k].FactionID
	})

	return warmachines
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
