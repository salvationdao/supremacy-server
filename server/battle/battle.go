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
	"server/passport"
	"server/rpcclient"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"

	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type Battle struct {
	arena       *Arena
	stage       string
	BattleID    string        `json:"battleID"`
	MapName     string        `json:"mapName"`
	WarMachines []*WarMachine `json:"warMachines"`
	SpawnedAI   []*WarMachine `json:"SpawnedAI"`
	lastTick    *[]byte
	gameMap     *server.GameMap
	abilities   *AbilitiesSystem
	users       usersMap
	factions    map[uuid.UUID]*boiler.Faction
	multipliers *MultiplierSystem
	spoils      *SpoilsOfWar
	rpcClient   *rpcclient.XrpcClient
	startedAt   time.Time

	destroyedWarMachineMap map[byte]*WMDestroyedRecord
	*boiler.Battle
}

const HubKeyLiveVoteCountUpdated hub.HubCommandKey = "LIVE:VOTE:COUNT:UPDATED"
const HubKeyWarMachineLocationUpdated hub.HubCommandKey = "WAR:MACHINE:LOCATION:UPDATED"

func (btl *Battle) start(payload *BattleStartPayload) {
	for _, wm := range payload.WarMachines {
		for i, wm2 := range btl.WarMachines {
			if wm.Hash == wm2.Hash {
				btl.WarMachines[i].ParticipantID = wm.ParticipantID
				continue
			}
		}
	}

	// set up the abilities for current battle

	btl.abilities = NewAbilitiesSystem(btl)
	btl.multipliers = NewMultiplierSystem(btl)
	btl.spoils = NewSpoilsOfWar(btl, 5*time.Second, 5*time.Second)

	btl.startedAt = time.Now()
	btl.BroadcastUpdate()

	// insert spoil of war
	spoil := &boiler.SpoilsOfWar{
		BattleID:     btl.ID,
		BattleNumber: btl.BattleNumber,
		Amount:       decimal.New(0, 18),
		AmountSent:   decimal.New(0, 18),
	}
	err := spoil.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to insert spoils")
	}

	// broadcast spoil of war on the start of the battle
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
	btl.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeySpoilOfWarUpdated), spoilOfWarPayload)
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

func (btl *Battle) end(payload *BattleEndPayload) {
	btl.EndedAt = null.TimeFrom(time.Now())
	_, err := btl.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID).Time("EndedAt", btl.EndedAt.Time).Msg("unable to update database for endat battle")
	}

	winningWarMachines := make([]*WarMachine, len(payload.WinningWarMachines))

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

	if winningWarMachines[0] == nil {
		gamelog.L.Panic().Str("Battle ID", btl.ID).Msg("no winning war machines")
	}

	topFactionContributorBoilers, err := db.TopSupsContributeFactions(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top faction contributors")
	}
	topPlayerContributorsBoilers, err := db.TopSupsContributors(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top player contributors")
	}
	topPlayerExecutorsBoilers, err := db.MostFrequentAbilityExecutors(uuid.Must(uuid.FromString(payload.BattleID)))
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", payload.BattleID).Msg("get top player executors")
	}

	topFactionContributors := []*Faction{}
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
	for _, p := range topPlayerContributorsBoilers {
		factionID := uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))
		if p.FactionID.Valid {
			factionID = uuid.Must(uuid.FromString(p.FactionID.String))
		}

		topPlayerContributors = append(topPlayerContributors, &BattleUser{
			ID:            uuid.Must(uuid.FromString(p.ID)),
			Username:      p.Username.String,
			FactionID:     winningWarMachines[0].FactionID,
			FactionColour: btl.factions[factionID].PrimaryColor,
			FactionLogoID: FactionLogos[p.FactionID.String],
		})
	}

	topPlayerExecutors := []*BattleUser{}
	for _, p := range topPlayerExecutorsBoilers {
		factionID := uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))
		if p.FactionID.Valid {
			factionID = uuid.Must(uuid.FromString(p.FactionID.String))
		}
		topPlayerExecutors = append(topPlayerExecutors, &BattleUser{
			ID:            uuid.Must(uuid.FromString(p.ID)),
			Username:      p.Username.String,
			FactionID:     p.FactionID.String,
			FactionColour: btl.factions[factionID].PrimaryColor,
			FactionLogoID: FactionLogos[p.FactionID.String],
		})
	}

	gamelog.L.Debug().
		Int("top_faction_contributors", len(topFactionContributors)).
		Int("top_player_executors", len(topPlayerExecutors)).
		Int("top_player_contributors", len(topPlayerContributors)).
		Msg("get top players and factions")

	endInfo := &BattleEndDetail{
		BattleID:                     btl.ID,
		BattleIdentifier:             btl.Battle.BattleNumber,
		StartedAt:                    btl.Battle.StartedAt,
		EndedAt:                      btl.Battle.EndedAt.Time,
		WinningCondition:             payload.WinCondition,
		WinningFaction:               winningWarMachines[0].Faction,
		WinningWarMachines:           winningWarMachines,
		TopSupsContributeFactions:    topFactionContributors,
		TopSupsContributors:          topPlayerExecutors,
		MostFrequentAbilityExecutors: topPlayerExecutors,
	}

	btl.stage = BattleStageEnd

	mws := make([]*db.MechWithOwner, len(payload.WinningWarMachines))

	err = db.ClearQueueByBattle(btl.ID)
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID).Msg("unable to clear queue for battle")
	}

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
		).One(gamedb.StdConn)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Warn().
					Str("Battle ID", btl.ID).
					Str("Mech ID", wm.ID).
					Err(err).
					Msg("no contract in database")
				continue
			}
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
			syndicateBalance := btl.arena.ppClient.UserBalanceGet(factID)

			if syndicateBalance.LessThanOrEqual(contract.ContractReward) {
				txid, err := btl.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
					FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
					ToUserID:             factID,
					Amount:               contract.ContractReward.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("contract_rewards|%s|%d", contract.ID, time.Now().UnixNano())),
					Group:                "battle",
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
			txid, err := btl.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
				FromUserID:           factID,
				ToUserID:             uuid.Must(uuid.FromString(contract.PlayerID)),
				Amount:               contract.ContractReward.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("contract_rewards|%s|%d", contract.ID, time.Now().UnixNano())),
				Group:                "battle",
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
	err = db.WinBattle(btl.ID, payload.WinCondition, mws...)
	if err != nil {
		gamelog.L.Error().
			Str("Battle ID", btl.ID).
			Err(err).
			Msg("unable to store mech wins")
	}

	btl.spoils.End()
	btl.multipliers.end(endInfo)
	btl.endInfoBroadcast(*endInfo)

	go func(id string) {
		// update user stat
		err = db.UserStatsRefresh(context.Background(), gamedb.Conn)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", id).
				Err(err).
				Msg("unable to refresh users stats")
		}

		us, err := db.UserStatsAll(context.Background(), gamedb.Conn)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", id).
				Err(err).
				Msg("unable to get users stats")
		}

		for _, u := range us {
			go btl.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, u.ID.String())), u)
		}
	}(btl.ID)
}

const HubKeyBattleEndDetailUpdated hub.HubCommandKey = "BATTLE:END:DETAIL:UPDATED"

func (btl *Battle) endInfoBroadcast(info BattleEndDetail) {
	btl.users.Range(func(user *BattleUser) bool {
		m, total := btl.multipliers.PlayerMultipliers(user.ID, 1)

		info.MultiplierUpdate = &MultiplierUpdate{
			UserMultipliers:  m,
			TotalMultipliers: fmt.Sprintf("%sx", total),
		}

		user.Send(HubKeyBattleEndDetailUpdated, info)
		return true
	})

	multipliers, err := db.PlayerMultipliers(btl.BattleNumber + 1)
	if err != nil {
		gamelog.L.Error().Str("battle number #", strconv.Itoa(btl.BattleNumber+1)).Err(err).Msg("Failed to get player multipliers from db")
		return
	}
	for _, m := range multipliers {
		m.TotalMultiplier = m.TotalMultiplier.Shift(-1)
	}

	// get the citizen list
	citizenPlayerIDs, err := db.CitizenPlayerIDs(btl.BattleNumber + 1)
	if err != nil {
		gamelog.L.Error().Str("battle number #", strconv.Itoa(btl.BattleNumber+1)).Err(err).Msg("Failed to get citizen player id list from db")
		return
	}

	go btl.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyMultiplierMapSubscribe), &MultiplierMapResponse{
		Multipliers:      multipliers,
		CitizenPlayerIDs: citizenPlayerIDs,
	})

	// broadcast spoil of war on the end of the battle
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
	go btl.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeySpoilOfWarUpdated), spoilOfWarPayload)
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
}

type ViewerLiveCount struct {
	RedMountain int64 `json:"red_mountain"`
	Boston      int64 `json:"boston"`
	Zaibatsu    int64 `json:"zaibatsu"`
	Other       int64 `json:"other"`
}

func (btl *Battle) userOnline(user *BattleUser, wsc *hub.Client) {
	u, ok := btl.users.User(user.ID)
	if !ok {
		user.wsClient[wsc] = true
		btl.users.Add(user)
	} else {
		u.Lock()
		u.wsClient[wsc] = true
		u.Unlock()
	}

	err := db.BattleViewerUpsert(context.Background(), gamedb.Conn, btl.ID, wsc.Identifier())
	if err != nil {
		gamelog.L.Error().Err(err)
	}

	resp := &ViewerLiveCount{
		RedMountain: 0,
		Boston:      0,
		Zaibatsu:    0,
		Other:       0,
	}

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

	btl.users.Send(HubKeyViewerLiveCountUpdated, resp)
}

func (btl *Battle) updatePayload() *GameSettingsResponse {
	var lt []byte
	if btl.lastTick != nil {
		lt = *btl.lastTick
	}
	return &GameSettingsResponse{
		GameMap:            btl.gameMap,
		WarMachines:        btl.WarMachines,
		SpawnedAI:          btl.SpawnedAI,
		WarMachineLocation: lt,
	}
}

const HubKeyGameSettingsUpdated = hub.HubCommandKey("GAME:SETTINGS:UPDATED")
const HubKeyGameUserOnline = hub.HubCommandKey("GAME:ONLINE")

func (btl *Battle) BroadcastUpdate() {
	btl.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameSettingsUpdated), btl.updatePayload())
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

	btl.arena.netMessageBus.Send(context.Background(), messagebus.NetBusKey(HubKeyWarMachineLocationUpdated), payload)

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

type JoinPayload struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash   string `json:"asset_hash"`
		NeedInsured bool   `json:"need_insured"`
	} `json:"payload"`
}

func (arena *Arena) Join(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	msg := &JoinPayload{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return err
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return err
	}

	// Get current queue length and calculate queue fee and reward
	result, err := db.QueueLength(factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	queueLength := decimal.NewFromInt(result + 1)
	queueCost := decimal.New(25, 16)     // 0.25 sups
	contractReward := decimal.New(2, 18) // 2 sups
	if queueLength.GreaterThan(decimal.NewFromInt(0)) {
		queueCost = queueLength.Mul(decimal.New(25, 16))     // 0.25x queue length
		contractReward = queueLength.Mul(decimal.New(2, 18)) // 2x queue length
	}

	// Charge user queue fee
	txid, err := arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
		Amount:               queueCost.StringFixed(18),
		FromUserID:           ownerID,
		ToUserID:             SupremacyBattleUserID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queueing_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
		Group:                "Battle",
		SubGroup:             "Queue",
		Description:          "Queued mech to battle arena",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("txID", txid).Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to charge user for insert mech into queue")
		return err
	}

	// Insert mech into queue
	position, err := db.JoinQueue(&db.BattleMechData{
		MechID:    mechID,
		OwnerID:   ownerID,
		FactionID: uuid.UUID(factionID),
	},
		contractReward,
		queueCost,
	)
	if err != nil {
		gamelog.L.Error().Interface("factionID", mech.FactionID).Err(err).Msg("unable to insert mech into queue")
		return err
	}

	if position == -1 {
		arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mechID)), WarMachineQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}

	reply(position)

	// Send updated battle queue status to all subscribers
	arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatus, factionID.String())), QueueStatusResponse{
		result,
		queueCost,
		contractReward,
	})

	// Send updated war machine queue status to all subscribers
	arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mechID)), WarMachineQueueStatusResponse{
		&position,
		&contractReward,
	})

	return err
}

type LeaveQueueRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

func (arena *Arena) Leave(ctx context.Context, wsc *hub.Client, payload []byte, factionID uuid.UUID, reply hub.ReplyFunc) error {
	msg := &LeaveQueueRequest{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue leave")
		return err
	}

	mechID, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	bq, err := boiler.FindBattleQueue(gamedb.StdConn, mechID.String())
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("probably not in queue")
		return err
	}

	mech, err := db.Mech(mechID)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("unable to retrieve mech")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechID.String()).Err(err).Msg("mech's owner player has no faction")
		return err
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return err
	}

	userID := uuid.FromStringOrNil(wsc.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	if userID != ownerID {
		return terror.Error(terror.ErrForbidden, "user is not mech owner")
	}

	originalQueueCost, err := db.QueueFee(mechID, factionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to remove mech from queue")
		return err
	}

	position, err := db.LeaveQueue(&db.BattleMechData{
		MechID:    mechID,
		OwnerID:   ownerID,
		FactionID: uuid.UUID(factionID),
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to remove mech from queue")
		return err
	}

	if position == -1 {
		arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mech.ID)), WarMachineQueueStatusResponse{
			nil,
			nil,
		})
		return nil
	}

	// Refund user queue fee
	txid, err := arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
		Amount:               originalQueueCost.StringFixed(18),
		FromUserID:           SupremacyBattleUserID,
		ToUserID:             ownerID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("refund_war_machine_queueing_fee|%s|%d", msg.Payload.AssetHash, time.Now().UnixNano())),
		Group:                "Battle",
		SubGroup:             "Queue",
		Description:          "Refunded battle arena queueing fee",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("txID", txid).Interface("mechID", mechID).Interface("factionID", mech.FactionID).Err(err).Msg("unable to charge user for insert mech into queue")
		return err
	}

	reply(true)

	result, err := db.QueueLength(factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to retrieve queue length")
		return err
	}

	queueLength := decimal.NewFromInt(result)
	queueCost := decimal.New(25, 16)     // 0.25 sups
	contractReward := decimal.New(2, 18) // 2 sups
	if queueLength.GreaterThan(decimal.NewFromInt(0)) {
		queueCost = queueLength.Mul(decimal.New(25, 16))     // 0.25x queue length
		contractReward = queueLength.Mul(decimal.New(2, 18)) // 2x queue length
	}

	// Send updated Battle queue status to all subscribers
	arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatus, factionID.String())), QueueStatusResponse{
		result,
		queueCost,
		contractReward,
	})

	mechsAfterIDs, err := db.AllMechsAfter(int(position), bq.QueuedAt, factionID)
	if err != nil {
		gamelog.L.Error().Interface("factionID", factionID).Err(err).Msg("unable to get mechs after")
		return err
	}

	// Send updated war machine queue status to all subscribers
	for _, m := range mechsAfterIDs {
		contractReward, err := db.QueueContract(m.MechID, factionID)
		if err != nil {
			gamelog.L.Error().Interface("mechID", mechID).Interface("factionID", factionID).Err(err).Msg("unable to get mechs contract reward")
			return err
		}
		arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, m.MechID)), WarMachineQueueStatusResponse{
			&m.QueuePosition,
			contractReward,
		})
	}
	arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mechID)), WarMachineQueueStatusResponse{
		nil,
		nil,
	})

	return err
}

type QueueStatusResponse struct {
	QueueLength    int64           `json:"queue_length"`
	QueueCost      decimal.Decimal `json:"queue_cost"`
	ContractReward decimal.Decimal `json:"contract_reward"`
}

func (arena *Arena) QueueStatus(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSQueueStatus, factionID.String())), nil
}

type WarMachineQueueStatusRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

type WarMachineQueueStatusResponse struct {
	QueuePosition  *int64           `json:"queue_position"` // in-game: -1; in queue: > 0; not in queue: nil
	ContractReward *decimal.Decimal `json:"contract_reward"`
}

func (arena *Arena) WarMachineQueueStatus(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &WarMachineQueueStatusRequest{}
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

	position, err := db.QueuePosition(mechID, factionID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			reply(WarMachineQueueStatusResponse{
				nil,
				nil,
			})
			return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mechID)), nil
		}
		return "", "", terror.Error(err)
	}

	if position == -1 {
		reply(WarMachineQueueStatusResponse{
			nil,
			nil,
		})
		return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mechID)), nil
	}

	contractReward, err := db.QueueContract(mechID, factionID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	mechInBattle, err := db.MechBattleStatus(mechID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if mechInBattle {
		position = -1
	}

	reply(WarMachineQueueStatusResponse{
		&position,
		contractReward,
	})

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", WSWarMachineQueueStatus, mechID)), nil
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
			break
		}
	}
	if destroyedWarMachine == nil {
		gamelog.L.Warn().Str("hash", dHash).Msg("can't match destroyed mech with battle state")
		return
	}

	var killByWarMachine *WarMachine
	if dp.DestroyedWarMachineEvent.KillByWarMachineHash != "" {
		for _, wm := range btl.WarMachines {
			if wm.Hash == dp.DestroyedWarMachineEvent.KillByWarMachineHash {
				killByWarMachine = wm
			}
		}
		if destroyedWarMachine == nil {
			gamelog.L.Warn().Str("killed_by_hash", dp.DestroyedWarMachineEvent.KillByWarMachineHash).Msg("can't match killer mech with battle state")
			return
		}
	}

	gamelog.L.Info().Msgf("battle Update: %s - War Machine Destroyed: %s", btl.ID, dHash)

	// save to database
	//tx, err := ba.Conn.Begin(ctx)
	//if err != nil {
	//	return terror.Error(err)
	//}
	//
	//defer func(tx pgx.Tx, ctx context.Context) {
	//	err := tx.Rollback(ctx)
	//	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
	//		ba.Log.Err(err).Msg("error rolling back")
	//	}
	//}(tx, ctx)

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

	//err = db.WarMachineDestroyedEventCreate(ctx, tx, dp.BattleID, dp.DestroyedWarMachineEvent)
	//if err != nil {
	//	return terror.Error(err)
	//}

	//err = db.StoreBattleEvent(&db.BattleEvent{})

	// TODO: Add kill assists
	//if len(assistedWarMachineIDs) > 0 {
	//	err = db.WarMachineDestroyedEventAssistedWarMachineSet(ctx, tx, dp.DestroyedWarMachineEvent.ID, assistedWarMachineIDs)
	//	if err != nil {
	//		return terror.Error(err)
	//	}
	//}

	//err = tx.Commit(ctx)
	//if err != nil {
	//	return terror.Error(err)
	//}

	_, err = db.UpdateBattleMech(btl.ID, warMachineID, false, true, killByWarMachineID)
	if err != nil {
		gamelog.L.Error().
			Str("battle_id", btl.ID).
			Interface("mech_id", warMachineID).
			Bool("killed", true).
			Msg("can't update battle mech")
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
	btl.arena.messageBus.Send(context.Background(),
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
		KilledBy:            wmd.KilledBy,
	})

}

func (btl *Battle) Load() error {
	q, err := db.LoadBattleQueue(context.Background(), 3)
	if err != nil {
		gamelog.L.Warn().Str("battle_id", btl.ID).Err(err).Msg("unable to load out queue")
		return err
	}

	if len(q) < 9 {
		gamelog.L.Warn().Msg("not enough mechs to field a battle. replacing with default battle.")

		err = btl.DefaultMechs()
		if err != nil {
			gamelog.L.Warn().Str("battle_id", btl.ID).Err(err).Msg("unable to load default mechs")
			return err
		}
		return nil
	}

	ids := make([]uuid.UUID, len(q))
	for i, bq := range q {
		ids[i], err = uuid.FromString(bq.MechID)
		if err != nil {
			gamelog.L.Warn().Str("mech_id", bq.MechID).Msg("failed to convert mech id string to uuid")
			return err
		}
	}

	mechs, err := db.Mechs(ids...)
	if err != nil {
		gamelog.L.Warn().Interface("mechs_ids", ids).Str("battle_id", btl.ID).Err(err).Msg("failed to retrieve mechs from mech ids")
		return err
	}
	btl.WarMachines = btl.MechsToWarMachines(mechs)

	err = db.QueueSetBattleID(btl.ID, ids...)
	if err != nil {
		gamelog.L.Error().Interface("mechs_ids", ids).Str("battle_id", btl.ID).Err(err).Msg("failed to set battle id in queue")
		return err
	}

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
