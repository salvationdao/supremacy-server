package battle

import (
	"context"
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
	"server/rpcclient"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"

	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type BattleStage int32

const (
	BattleStagStart = 1
	BattleStageEnd  = 0
)

type Battle struct {
	arena          *Arena
	stage          *atomic.Int32
	BattleID       string        `json:"battleID"`
	MapName        string        `json:"mapName"`
	WarMachines    []*WarMachine `json:"warMachines"`
	SpawnedAI      []*WarMachine `json:"SpawnedAI"`
	warMachineIDs  []uuid.UUID   `json:"ids"`
	lastTick       *[]byte
	gameMap        *server.GameMap
	_abilities     *AbilitiesSystem
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

	viewerCountInputChan chan (*ViewerLiveCount)
	sync.RWMutex
}

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
	btl.spoils = NewSpoilsOfWar(btl, 5*time.Second, 5*time.Second)
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up battle abilities")
	btl.storeAbilities(NewAbilitiesSystem(btl))
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Spinning up battle multipliers")
	btl.multipliers = NewMultiplierSystem(btl)
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting battle start to players")
	btl.BroadcastUpdate()

	// broadcast spoil of war on the start of the battle
	gamelog.L.Info().Int("battle_number", btl.BattleNumber).Str("battle_id", btl.ID).Msg("Broadcasting spoils of war updates")

	sows, err := db.LastTwoSpoilOfWarAmount()
	if err != nil || len(sows) == 0 {
		gamelog.L.Error().Err(err).Msg("Failed to get last two spoil of war amount")
		return
	}

	spoilOfWarPayload := []byte{byte(SpoilOfWarTick)}
	spoilOfWarStr := []string{}
	for _, sow := range sows {
		spoilOfWarStr = append(spoilOfWarStr, sow.String())
	}
	spoilOfWarPayload = append(spoilOfWarPayload, []byte(strings.Join(spoilOfWarStr, "|"))...)
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the battle abilities end!")
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the spoils end!")
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Msg("panic! panic! panic! Panic at the creation of ending info: endCreateStats!")
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the battle end processWinners!")
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
			gamelog.L.Warn().
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
				Err(err).
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
				Err(err).
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the sorting up ending war machines!")
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
					MechID:    w.ID,
					TotalWins: 1,
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

			ms.TotalWins = ms.TotalWins + 1
			_, err = ms.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", ms).
					Msg("unable to update mech stat")
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the ending of multis! btl.endMultis!")
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
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the ending of end broadcast!")
		}
	}()
	btl.endInfoBroadcast(*endInfo)
}

func (btl *Battle) end(payload *BattleEndPayload) {
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the battle end!")
			exists, err := boiler.BattleExists(gamedb.StdConn, btl.ID)
			if err != nil {
				gamelog.L.Panic().Err(err).Msg("Panicing. Unable to even check if battle id exists")
			}
			if exists {

			}
		}
	}()

	btl.EndedAt = null.TimeFrom(time.Now())
	_, err := btl.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID).Time("EndedAt", btl.EndedAt.Time).Msg("unable to update database for endat battle")
	}

	btl.endAbilities()
	btl.endSpoils()

	winningWarMachines := btl.endWarMachines(payload)
	endInfo := btl.endCreateStats(payload, winningWarMachines)

	btl.processWinners(payload)
	btl.endMultis(endInfo)

	notifications, err := boiler.BattleQueueNotifications(
		boiler.BattleQueueNotificationWhere.QueueMechID.IsNotNull(),
		boiler.BattleQueueNotificationWhere.BattleID.EQ(null.StringFrom(btl.BattleID)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Str("Battle ID", btl.ID).Str("battle_id", payload.BattleID).Msg("Failed to get battle queue notifications with a queue mech id")
	}

	for _, n := range notifications {
		n.QueueMechID = null.NewString("", false)
		_, err = n.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Panic().Err(err).Str("Battle ID", btl.ID).Str("battle_id", payload.BattleID).Msg("Failed to remove queue mech id from battle queue notifications.")
		}
	}

	_, err = boiler.BattleQueues(boiler.BattleQueueWhere.BattleID.EQ(null.StringFrom(btl.BattleID))).DeleteAll(gamedb.StdConn)
	if err != nil {
		gamelog.L.Panic().Err(err).Str("Battle ID", btl.ID).Str("battle_id", payload.BattleID).Msg("Failed to remove mechs from battle queue.")
	}

	gamelog.L.Info().Msgf("battle has been cleaned up, sending broadcast %s", btl.ID)
	btl.endBroadcast(endInfo)
}

const HubKeyBattleEndDetailUpdated hub.HubCommandKey = "BATTLE:END:DETAIL:UPDATED"

func (btl *Battle) endInfoBroadcast(info BattleEndDetail) {
	btl.users.Range(func(user *BattleUser) bool {
		m, total := btl.multipliers.PlayerMultipliers(user.ID, 0)

		info.MultiplierUpdate = &MultiplierUpdate{
			UserMultipliers:  m,
			TotalMultipliers: fmt.Sprintf("%sx", total),
		}

		user.Send(HubKeyBattleEndDetailUpdated, info)

		us, err := db.UserStatsGet(user.ID.String())
		if err != nil {
			gamelog.L.Error().Str("player_id", user.ID.String()).Err(err).Msg("Failed to get user stats")
		}

		// broadcast user stat to user
		if us != nil {
			go btl.arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, us.ID)), us)
		}

		return true
	})

	multipliers, err := db.PlayerMultipliers(btl.BattleNumber)
	if err != nil {
		gamelog.L.Error().Str("battle number #", strconv.Itoa(btl.BattleNumber)).Err(err).Msg("Failed to get player multipliers from db")
		return
	}

	for _, m := range multipliers {
		m.TotalMultiplier = m.TotalMultiplier.Shift(-1)
	}
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

		userIDs := btl.users.OnlineUserIDs()
		if len(userIDs) > 0 {
			uss, err := boiler.UserStats(
				boiler.UserStatWhere.ID.IN(userIDs),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to get user stats from db")
			}

			if uss != nil {
				btl.users.Send(HubKeyUserStatChatSubscribe, uss, user.ID)
			}
		}

	}
}

func (btl *Battle) debounceSendingViewerCount(cb func(result ViewerLiveCount, btl *Battle)) {
	defer func() {
		if err := recover(); err != nil {
			gamelog.L.Error().Interface("err", err).Stack().Msg("panic! panic! panic! Panic at the debounceSendingViewerCount!")
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

				userIDs := btl.users.OnlineUserIDs()
				if len(userIDs) > 0 {
					uss, err := boiler.UserStats(
						qm.Select(boiler.UserStatColumns.ID, boiler.UserStatColumns.KillCount),
						boiler.UserStatWhere.ID.IN(userIDs),
					).All(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg("Failed to get user stats from db")
					}

					if uss != nil {
						btl.users.Send(HubKeyUserStatChatSubscribe, uss)
					}
				}
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

type QueueJoinHandlerResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
}

type QueueJoinRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash                   string `json:"asset_hash"`
		NeedInsured                 bool   `json:"need_insured"`
		EnablePushNotifications     bool   `json:"enable_push_notifications,omitempty"`
		MobileNumber                string `json:"mobile_number,omitempty"`
		EnableTelegramNotifications bool   `json:"enable_telegram_notifications"`
	} `json:"payload"`
}

const WSQueueJoin hub.HubCommandKey = "BATTLE:QUEUE:JOIN"

func (arena *Arena) QueueJoinHandler(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	msg := &QueueJoinRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return terror.Error(err)
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(fmt.Errorf("missing warmachine faction"))
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return terror.Error(err)
	}

	if mech.OwnerID != wsc.Identifier() {
		return terror.Error(fmt.Errorf("does not own the mech"), "Current mech does not own by you")
	}

	// Get current queue length and calculate queue fee and reward
	result, err := db.QueueLength(factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return terror.Error(err)
	}

	queueLength := decimal.NewFromInt(result + 1)
	queueCost := decimal.New(25, 16)     // 0.25 sups
	contractReward := decimal.New(2, 18) // 2 sups
	if queueLength.GreaterThan(decimal.NewFromInt(0)) {
		queueCost = queueLength.Mul(decimal.New(25, 16))     // 0.25x queue length
		contractReward = queueLength.Mul(decimal.New(2, 18)) // 2x queue length
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return fmt.Errorf(terror.Echo(err))
	}
	defer tx.Rollback()

	var position int64

	// Insert mech into queue
	exists, err := boiler.BattleQueueExists(tx, mechID.String())
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("check mech exists in queue")
	}
	if exists {
		gamelog.L.Debug().Str("mech_id", mechID.String()).Err(err).Msg("mech already in queue")
		position, err = db.QueuePosition(mechID, factionID)
		if err != nil {
			return terror.Error(err, "Already in queue, failed to get position. Contact support or try again.")
		}
		reply(true)
		return nil
	}

	bc := &boiler.BattleContract{
		MechID:         mechID.String(),
		FactionID:      factionID.String(),
		PlayerID:       ownerID.String(),
		ContractReward: contractReward,
		Fee:            queueCost,
	}
	err = bc.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("contractReward", contractReward.String()).
			Str("queueFee", queueCost.String()).
			Err(err).Msg("unable to create battle contract")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	bq := &boiler.BattleQueue{
		MechID:           mechID.String(),
		QueuedAt:         time.Now(),
		FactionID:        factionID.String(),
		OwnerID:          ownerID.String(),
		BattleContractID: null.StringFrom(bc.ID),
	}

	notifications := msg.Payload.EnablePushNotifications || msg.Payload.MobileNumber != "" || msg.Payload.EnableTelegramNotifications
	if !notifications {
		bq.Notified = true
	}
	err = bq.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to insert mech into queue")
		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}
	factionAccountID, ok := server.FactionUsers[factionID.String()]
	if !ok {
		gamelog.L.Error().
			Str("mech ID", mech.ID).
			Str("faction ID", factionID.String()).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	// Charge user queue fee
	supTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
		Amount:               queueCost.StringFixed(18),
		FromUserID:           ownerID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queueing_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		SubGroup:             "Queue",
		Description:          "Queued mech to battle arena",
		NotSafe:              true,
	})
	if err != nil {
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Interface("mechID", mechID).Interface("factionID", factionID.String()).Err(err).Msg("unable to charge user for insert mech into queue")
		return terror.Error(err, "Unable to process queue fee,  check your balance and try again.")
	}

	bq.QueueFeeTXID = null.StringFrom(supTransactionID)
	_, err = bq.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Str("tx_id", supTransactionID).
			Err(err).Msg("unable to update battle queue with queue transaction id")
		if bq.QueueFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
			}
		}

		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	shortcode := ""
	bqn := &boiler.BattleQueueNotification{}
	// Charge queue notification fee, if enabled (10% of queue cost)
	if !bq.Notified {
		notifyCost := queueCost.Mul(decimal.NewFromFloat(0.1))
		notifyTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
			Amount:               notifyCost.String(),
			FromUserID:           ownerID,
			ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
			TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queue_notification_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupBattle),
			SubGroup:             "Queue",
			Description:          "Notification surcharge for queued mech in arena",
			NotSafe:              true,
		})
		if err != nil {
			gamelog.L.Error().Str("txID", notifyTransactionID).Err(err).Msg("unable to charge user for sms notification for mech in queue")

			if bq.QueueFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
				}
			}
			// Abort transaction if charge fails
			return terror.Error(err, "Unable to process notification fee, please check your balance and try again.")
		}
		bq.QueueNotificationFeeTXID = null.StringFrom(notifyTransactionID)
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("tx_id", notifyTransactionID).
				Err(err).Msg("unable to update battle queue with queue notification transaction id")
			if bq.QueueFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
				}
			}
			if bq.QueueNotificationFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
				}
			}

			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

		//insert notification into db
		bqn = &boiler.BattleQueueNotification{
			MechID:            mechID.String(),
			QueueMechID:       null.StringFrom(mechID.String()),
			MobileNumber:      null.StringFrom(msg.Payload.MobileNumber),
			PushNotifications: msg.Payload.EnablePushNotifications,
			Fee:               notifyCost,
		}

		if msg.Payload.EnableTelegramNotifications {
			telegramNotification, err := arena.telegram.NotificationCreate(mechID.String(), bqn)
			if err != nil {
				gamelog.L.Error().
					Str("mechID", mechID.String()).
					Str("playerID", ownerID.String()).
					Err(err).Msg("unable to create telegram notification")
				if bq.QueueFeeTXID.Valid {
					_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
					if err != nil {
						gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
					}
				}
				if bq.QueueNotificationFeeTXID.Valid {
					_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
					if err != nil {
						gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
					}
				}
				return terror.Error(err, "Unable create telegram notification. Contact support.")
			}
			bqn.TelegramNotificationID = null.StringFrom(telegramNotification.ID)
			shortcode = telegramNotification.Shortcode
		}

		err = bqn.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Interface("mech", mech).
				Err(err).Msg("unable to insert queue notification for mech")
			if bq.QueueFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
				}
			}
			if bq.QueueNotificationFeeTXID.Valid {
				_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
				if err != nil {
					gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
				}
			}

			return terror.Error(err, "Unable to join queue, contact support or try again.")
		}

	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to commit mech insertion into queue")
		if bq.QueueFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("txID", bq.QueueFeeTXID.String).Err(err).Msg("failed to refund queue fee")
			}
		}
		if bq.QueueNotificationFeeTXID.Valid {
			_, err = arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
			if err != nil {
				gamelog.L.Error().Str("txID", bq.QueueNotificationFeeTXID.String).Err(err).Msg("failed to refund queue notification fee")
			}
		}

		return terror.Error(err, "Unable to join queue, contact support or try again.")
	}

	// Get mech current queue position
	position, err = db.QueuePosition(mechID, factionID)
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), AssetQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}
	if err != nil {
		gamelog.L.Error().
			Str("mechID", mechID.String()).
			Str("factionID", factionID.String()).
			Err(err).Msg("unable to retrieve mech queue position")
		return terror.Error(err, "Unable to join queue, check your balance and try again.")
	}

	nextQueueLength := queueLength.Add(decimal.NewFromInt(1))
	nextQueueCost := decimal.New(25, 16)     // 0.25 sups
	nextContractReward := decimal.New(2, 18) // 2 sups
	if nextQueueLength.GreaterThan(decimal.NewFromInt(0)) {
		nextQueueCost = nextQueueLength.Mul(decimal.New(25, 16))     // 0.25x queue length
		nextContractReward = nextQueueLength.Mul(decimal.New(2, 18)) // 2x queue length
	}

	// reply with shortcode if telegram notifs enabled
	if bqn.TelegramNotificationID.Valid && shortcode != "" {
		reply(QueueJoinHandlerResponse{
			Success: true,
			Code:    shortcode,
		})
	} else {
		reply(QueueJoinHandlerResponse{
			Success: true,
			Code:    "",
		})
	}

	// Send updated battle queue status to all subscribers
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatusSubscribe, factionID.String())), QueueStatusResponse{
		result + 1,
		nextQueueCost,
		nextContractReward,
	})

	// Send updated war machine queue status to subscriber
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), AssetQueueStatusResponse{
		&position,
		&contractReward,
	})

	return nil
}

type QueueLeaveRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

const WSQueueLeave hub.HubCommandKey = "BATTLE:QUEUE:LEAVE"

func (arena *Arena) QueueLeaveHandler(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	msg := &QueueLeaveRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden, "Only the owners of the war machine can remove it from the queue.")
	}

	if userID != ownerID {
		return terror.Error(terror.ErrForbidden, "Only the owners of the war machine can remove it from the queue.")
	}

	originalQueueCost, err := db.QueueFee(mechID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to remove mech from queue")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	// Get queue position before deleting
	position, err := db.QueuePosition(mechID, factionID)
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		gamelog.L.Warn().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("tried to remove already unqueued mech from queue")
		return terror.Warn(err, "Unable to find war machine in battle queue, ensure machine isn't already removed and contact support.")
	}
	if err != nil {
		gamelog.L.Warn().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to get mech position")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	if position == -1 {
		// If mech is currently in battle
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("cannot remove battling mech from queue")
		return terror.Error(fmt.Errorf("cannot remove war machine from queue when it is in battle"), "You cannot remove war machines currently in battle.")
	}

	canxq := `UPDATE battle_contracts SET cancelled = true WHERE id = (SELECT battle_contract_id FROM battle_queue WHERE mech_id = $1)`
	_, err = gamedb.StdConn.Exec(canxq, mechID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to cancel battle contract. mech has left queue though.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}
	defer tx.Rollback()

	// Remove from queue
	bq, err := boiler.BattleQueues(boiler.BattleQueueWhere.MechID.EQ(mechID.String())).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("mech_id", mechID.String()).
			Msg("unable to get existing mech from queue")
		return terror.Error(err, "Issue leaving queue, try again or contact support.")
	}

	factionAccountID, ok := server.FactionUsers[factionID.String()]
	if !ok {
		gamelog.L.Error().
			Str("mech ID", mech.ID).
			Str("faction ID", factionID.String()).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	// refund queue fee if not already refunded
	if !bq.QueueFeeTXIDRefund.Valid {
		// check if they have a transaction ID
		if bq.QueueFeeTXID.Valid && bq.QueueFeeTXID.String != "" {
			queueRefundTransactionID, err := arena.RPCClient.RefundSupsMessage(bq.QueueFeeTXID.String)
			if err != nil {
				gamelog.L.Error().
					Str("queue_transaction_id", bq.QueueFeeTXID.String).
					Err(err).
					Msg("failed to refund users queue fee")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueFeeTXIDRefund = null.StringFrom(queueRefundTransactionID)
		} else {
			// TODO: Eventually all battle queues will have transaction ids to refund against, but legency queue will not. So keeping below until all legacy queues have passed
			// Refund user queue fee
			queueRefundTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
				Amount:               originalQueueCost.StringFixed(18),
				FromUserID:           uuid.Must(uuid.FromString(factionAccountID)),
				ToUserID:             ownerID,
				TransactionReference: server.TransactionReference(fmt.Sprintf("refund_war_machine_queueing_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             "Queue",
				Description:          "Refunded battle arena queueing fee",
				NotSafe:              true,
			})
			if err != nil {
				// Abort transaction if refund fails
				gamelog.L.Error().Str("txID", queueRefundTransactionID).Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to charge user for insert mech into queue")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueFeeTXIDRefund = null.StringFrom(queueRefundTransactionID)
		}
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("queue_refund_transaction_id", bq.QueueFeeTXIDRefund.String).
				Err(err).Msg("unable to update battle queue with refund transaction details")
			return terror.Error(err, "Unable to join queue, check your balance and try again.")
		}
	}

	updateBQNq := `UPDATE battle_queue_notifications SET is_refunded = true, queue_mech_id = null WHERE mech_id = $1`
	_, err = gamedb.StdConn.Exec(updateBQNq, mechID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to update battle_queue_notifications table during refund")
	}

	// Refund queue notification fee, if enabled and not already refunded
	if !bq.Notified && !bq.QueueNotificationFeeTXIDRefund.Valid {
		if bq.QueueNotificationFeeTXID.Valid && bq.QueueNotificationFeeTXID.String != "" {
			queueNotificationRefundTransactionID, err := arena.RPCClient.RefundSupsMessage(bq.QueueNotificationFeeTXID.String)
			if err != nil {
				gamelog.L.Error().
					Err(err).
					Str("queue_notification_transaction_id", bq.QueueNotificationFeeTXID.String).
					Msg("failed to refund users notification fee")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueNotificationFeeTXIDRefund = null.StringFrom(queueNotificationRefundTransactionID)
		} else {
			// TODO: Eventually all battle queues will have transaction ids to refund against, but legency queue will not. So keeping below until all legacy queues have passed
			notifyCost := originalQueueCost.Mul(decimal.NewFromFloat(0.1))
			queueNotificationRefundTransactionID, err := arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
				Amount:               notifyCost.StringFixed(18),
				FromUserID:           uuid.Must(uuid.FromString(factionAccountID)),
				ToUserID:             ownerID,
				TransactionReference: server.TransactionReference(fmt.Sprintf("refund_war_machine_queue_notification_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             "Queue",
				Description:          "Refunded notification surcharge for queued mech in arena",
				NotSafe:              true,
			})
			if err != nil {
				// Abort transaction if charge fails
				gamelog.L.Error().Str("txID", queueNotificationRefundTransactionID).Err(err).Msg("unable to refund user for notification for mech in queue")
				return terror.Error(err, "Unable to process refund, try again or contact support.")
			}
			bq.QueueNotificationFeeTXIDRefund = null.StringFrom(queueNotificationRefundTransactionID)
		}
		_, err = bq.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("queue_notification_refund_transaction_id", bq.QueueNotificationFeeTXIDRefund.String).
				Err(err).Msg("unable to update battle queue with notification refund transaction details")
			return terror.Error(err, "Unable to leave queue, try again or contact support.")
		}
	}

	_, err = bq.Delete(tx)
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to remove mech from queue")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Err(err).Msg("unable to commit mech deletion from queue")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	reply(true)

	result, err := db.QueueLength(factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return terror.Error(err, "Unable to leave queue, try again or contact support.")
	}

	nextQueueLength := decimal.NewFromInt(result + 1)
	nextQueueCost := decimal.New(25, 16)     // 0.25 sups
	nextContractReward := decimal.New(2, 18) // 2 sups
	if nextQueueLength.GreaterThan(decimal.NewFromInt(0)) {
		nextQueueCost = nextQueueLength.Mul(decimal.New(25, 16))     // 0.25x queue length
		nextContractReward = nextQueueLength.Mul(decimal.New(2, 18)) // 2x queue length
	}

	// Send updated Battle queue status to all subscribers
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatusSubscribe, factionID.String())), QueueStatusResponse{
		result,
		nextQueueCost,
		nextContractReward,
	})

	// Tell clients to refetch war machine queue status
	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueUpdatedSubscribe, factionID.String())), true)

	return nil
}

type QueueStatusResponse struct {
	QueueLength    int64           `json:"queue_length"`
	QueueCost      decimal.Decimal `json:"queue_cost"`
	ContractReward decimal.Decimal `json:"contract_reward"`
}

const WSQueueStatusSubscribe hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:STATUS:SUBSCRIBE")

func (arena *Arena) QueueStatusSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find faction from user id")
		return "", "", terror.Error(err)
	}

	if needProcess {
		result, err := db.QueueLength(factionID)
		if err != nil {
			gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
			return "", "", terror.Error(err)
		}

		queueLength := decimal.NewFromInt(result + 1)
		queueCost := decimal.New(25, 16)     // 0.25 sups
		contractReward := decimal.New(2, 18) // 2 sups
		if queueLength.GreaterThan(decimal.NewFromInt(0)) {
			queueCost = queueLength.Mul(decimal.New(25, 16))     // 0.25x queue length
			contractReward = queueLength.Mul(decimal.New(2, 18)) // 2x queue length
		}

		reply(QueueStatusResponse{
			result,
			queueCost,
			contractReward,
		})
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatusSubscribe, factionID.String())), nil
}

const WSQueueUpdatedSubscribe hub.HubCommandKey = hub.HubCommandKey("BATTLE:QUEUE:UPDATED")

func (arena *Arena) QueueUpdatedSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput)
	}

	factionID, err := GetPlayerFactionID(userID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", userID.String()).Err(err).Msg("unable to find faction from user id")
		return "", "", terror.Error(err)
	}

	if needProcess {
		reply(true)
	}
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueUpdatedSubscribe, factionID)), nil
}

type AssetQueueStatusRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

type AssetQueueStatusResponse struct {
	QueuePosition  *int64           `json:"queue_position"` // in-game: -1; in queue: > 0; not in queue: nil
	ContractReward *decimal.Decimal `json:"contract_reward"`
}

const WSAssetQueueStatus hub.HubCommandKey = hub.HubCommandKey("ASSET:QUEUE:STATUS")

func (arena *Arena) AssetQueueStatusHandler(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	req := &AssetQueueStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	mechID, err := db.MechIDFromHash(req.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", req.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return terror.Error(err)
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return terror.Error(err)
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return terror.Error(err)
	}

	mechFactionID, err := GetPlayerFactionID(ownerID)
	if err != nil || mechFactionID.IsNil() {
		gamelog.L.Error().Str("userID", ownerID.String()).Err(err).Msg("unable to find faction from owner id")
		return terror.Error(err)
	}

	position, err := db.QueuePosition(mechID, mechFactionID)
	if errors.Is(sql.ErrNoRows, err) {
		// If mech is not in queue
		reply(AssetQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}
	if err != nil {
		return terror.Error(err)
	}

	contractReward, err := db.QueueContract(mechID, mechFactionID)
	if err != nil {
		gamelog.L.Error().Str("mechID", mechID.String()).Str("mechFactionID", mechFactionID.String()).Err(err).Msg("unable to get contract reward")
		return terror.Error(err)
	}

	reply(AssetQueueStatusResponse{
		&position,
		contractReward,
	})

	return nil
}

const WSAssetQueueStatusList hub.HubCommandKey = hub.HubCommandKey("ASSET:QUEUE:STATUS:LIST")

type AssetQueueStatusItem struct {
	MechID        string `json:"mech_id"`
	QueuePosition int64  `json:"queue_position"`
}

func (arena *Arena) AssetQueueStatusListHandler(ctx context.Context, hub *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error {
	userID, err := uuid.FromString(hub.Identifier())
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	queueList, err := db.QueueOwnerList(userID)
	if err != nil {
		return terror.Error(err, "Failed to list mechs in queue")
	}

	resp := []*AssetQueueStatusItem{}
	for _, q := range queueList {
		obj := &AssetQueueStatusItem{
			MechID:        q.MechID.String(),
			QueuePosition: q.QueuePosition,
		}
		resp = append(resp, obj)
	}

	reply(resp)

	return nil
}

const WSAssetQueueStatusSubscribe hub.HubCommandKey = hub.HubCommandKey("ASSET:QUEUE:STATUS:SUBSCRIBE")

func (arena *Arena) AssetQueueStatusSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &AssetQueueStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	mechID, err := db.MechIDFromHash(req.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", req.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return "", "", terror.Error(err)
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return "", "", terror.Error(err)
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return "", "", terror.Error(err)
	}

	if mech.OwnerID != wsc.Identifier() {
		gamelog.L.Warn().Str("player id", wsc.Identifier()).Str("mech id", mechID.String()).Msg("Someone attempt to subscribe on a mech's queuing status which is not belong to them")
		return "", "", terror.Error(terror.ErrForbidden, "Cannot subscribe on mech which is not belong to you")
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return "", "", terror.Error(err)
	}

	factionID, err := GetPlayerFactionID(ownerID)
	if err != nil || factionID.IsNil() {
		gamelog.L.Error().Str("userID", ownerID.String()).Err(err).Msg("unable to find faction from owner id")
		return "", "", terror.Error(err)
	}

	if needProcess {
		position, err := db.QueuePosition(mechID, factionID)
		if errors.Is(sql.ErrNoRows, err) {
			// If mech is not in queue
			reply(AssetQueueStatusResponse{
				nil,
				nil,
			})
			return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), nil
		}
		if err != nil {
			gamelog.L.Error().Str("mechID", mechID.String()).Str("factionID", factionID.String()).Err(err).Msg("unable to get mech queue position")
			return "", "", terror.Error(err)
		}

		contractReward, err := db.QueueContract(mechID, factionID)
		if err != nil {
			gamelog.L.Error().Str("mechID", mechID.String()).Str("factionID", factionID.String()).Err(err).Msg("unable to get contract reward")
			return "", "", terror.Error(err)
		}

		reply(AssetQueueStatusResponse{
			&position,
			contractReward,
		})
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSAssetQueueStatusSubscribe, mechID)), nil
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
				}
			}
		}
		if destroyedWarMachine == nil {
			gamelog.L.Warn().Str("killed_by_hash", dp.DestroyedWarMachineEvent.KillByWarMachineHash).Msg("can't match killer mech with battle state")
			return
		}
	} else if dp.DestroyedWarMachineEvent.RelatedEventIDString != "" {
		// check related event id

		// get ability via offering id
		abl, err := boiler.BattleAbilityTriggers(boiler.BattleAbilityTriggerWhere.AbilityOfferingID.EQ(dp.DestroyedWarMachineEvent.RelatedEventIDString)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("related event id", dp.DestroyedWarMachineEvent.RelatedEventIDString).Err(err).Msg("Failed get ability from offering id")
		}

		if abl != nil && abl.PlayerID.Valid {
			currentUser, err := BuildUserDetailWithFaction(uuid.FromStringOrNil(abl.PlayerID.String))
			if err == nil {
				// update kill by user and killed by information
				killedByUser = currentUser
				dp.DestroyedWarMachineEvent.KilledBy = fmt.Sprintf("(%s)", abl.AbilityLabel)
			}
			if strings.EqualFold(destroyedWarMachine.FactionID, abl.FactionID) {
				// update user kill
				_, err := db.UserStatSubtractAbilityKill(abl.PlayerID.String)
				if err != nil {
					gamelog.L.Error().Str("player_id", abl.PlayerID.String).Err(err).Msg("Failed to subtract user ability kill count")
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

				// add faction kill count
				err = db.FactionAddAbilityKillCount(abl.FactionID)
				if err != nil {
					gamelog.L.Error().Str("faction_id", abl.FactionID).Err(err).Msg("Failed to add faction ability kill count")
				}
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
