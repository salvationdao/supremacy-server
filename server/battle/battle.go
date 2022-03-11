package battle

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"strconv"
	"strings"
	"time"

	"github.com/volatiletech/null/v8"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type Battle struct {
	arena       *Arena
	stage       string
	ID          uuid.UUID     `json:"battleID" db:"id"`
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
}

func (btl *Battle) isOnline(userID uuid.UUID) bool {
	_, ok := btl.users.User(userID)
	return ok
}

func (btl *Battle) end(payload *BattleEndPayload) {

	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")
	fmt.Println("end")

	btl.Battle.EndedAt = null.TimeFrom(time.Now())
	_, err := btl.Battle.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("Battle ID", btl.ID.String()).Time("EndedAt", btl.Battle.EndedAt.Time).Msg("unable to update database for endat battle")
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
			gamelog.L.Error().Str("Battle ID", btl.ID.String()).Msg("unable to match war machine to battle with hash")
		}
	}

	if winningWarMachines[0] == nil {
		gamelog.L.Panic().Str("Battle ID", btl.ID.String()).Msg("no winning war machines")
	}

	fakedUsers := []*BattleUser{
		{
			ID:            uuid.Must(uuid.NewV4()),
			Username:      "FakeUser1",
			FactionID:     winningWarMachines[0].FactionID,
			FactionColour: btl.factions[uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))].PrimaryColor,
			FactionLogoID: FactionLogos[winningWarMachines[0].FactionID],
		},
		{
			ID:            uuid.Must(uuid.NewV4()),
			Username:      "FakeUser2",
			FactionID:     winningWarMachines[0].FactionID,
			FactionColour: btl.factions[uuid.Must(uuid.FromString(winningWarMachines[0].FactionID))].PrimaryColor,
			FactionLogoID: FactionLogos[winningWarMachines[0].FactionID],
		},
	}

	fakedFactions := make([]*Faction, 2)
	i := 0
	for _, faction := range btl.factions {
		fakedFactions[i] = &Faction{
			ID:    faction.ID,
			Label: faction.Label,
			Theme: &FactionTheme{
				Primary:    faction.PrimaryColor,
				Secondary:  faction.SecondaryColor,
				Background: faction.BackgroundColor,
			},
		}
		if i == 1 {
			break
		}
		i++
	}

	endInfo := &BattleEndDetail{
		BattleID:                     btl.ID.String(),
		BattleIdentifier:             btl.Battle.BattleNumber,
		StartedAt:                    btl.Battle.StartedAt,
		EndedAt:                      btl.Battle.EndedAt.Time,
		WinningCondition:             payload.WinCondition,
		WinningFaction:               winningWarMachines[0].Faction,
		WinningWarMachines:           winningWarMachines,
		TopSupsContributors:          fakedUsers,
		TopSupsContributeFactions:    fakedFactions,
		MostFrequentAbilityExecutors: fakedUsers,
	}

	ids := make([]uuid.UUID, len(btl.WarMachines))
	err = db.ClearQueue(ids...)
	if err != nil {
		gamelog.L.Error().Interface("ids", ids).Err(err).Msg("db.ClearQueue() returned error")
		return
	}

	btl.stage = BattleStageEnd

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
			gamelog.L.Error().Str("Battle ID", btl.ID.String()).Msg("unable to match war machine to battle with hash")
			return
		}
		mechId, err := uuid.FromString(wm.ID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID.String()).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert mech id to uuid")
			return
		}
		ownedById, err := uuid.FromString(wm.OwnedByID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID.String()).
				Str("mech ID", wm.ID).
				Err(err).
				Msg("unable to convert owned id to uuid")
			return
		}
		factionId, err := uuid.FromString(wm.FactionID)
		if err != nil {
			gamelog.L.Error().
				Str("Battle ID", btl.ID.String()).
				Str("faction ID", wm.FactionID).
				Err(err).
				Msg("unable to convert faction id to uuid")
			return
		}
		mws[i] = &db.MechWithOwner{
			OwnerID:   ownedById,
			MechID:    mechId,
			FactionID: factionId,
		}
	}
	err = db.WinBattle(btl.ID, payload.WinCondition, mws...)
	if err != nil {
		gamelog.L.Error().
			Str("Battle ID", btl.ID.String()).
			Err(err).
			Msg("unable to store mech wins")
		return
	}

	btl.multipliers.end(endInfo)
	btl.spoils.End()
	btl.endInfoBroadcast(*endInfo)
	err = db.UserStatsRefresh(context.Background(), gamedb.Conn)
	if err != nil {
		gamelog.L.Error().
			Str("Battle ID", btl.ID.String()).
			Err(err).
			Msg("unable to refresh users stats")
		return
	}

	us, err := db.UserStatsAll(context.Background(), gamedb.Conn)
	if err != nil {
		gamelog.L.Error().
			Str("Battle ID", btl.ID.String()).
			Err(err).
			Msg("unable to get users stats")
		return
	}

	go func() {
		for _, u := range us {
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________")
			fmt.Println("_________", u)

			go btl.arena.messageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, u.ID.String())), u)
		}
	}()

}

const HubKeyBattleEndDetailUpdated hub.HubCommandKey = "BATTLE:END:DETAIL:UPDATED"

func (btl *Battle) endInfoBroadcast(info BattleEndDetail) {
	btl.users.Range(func(user *BattleUser) bool {
		m, total := btl.multipliers.PlayerMultipliers(user.ID)

		info.MultiplierUpdate = &MultiplierUpdate{
			UserMultipliers:  m,
			TotalMultipliers: fmt.Sprintf("%sx", total),
		}

		user.Send(HubKeyBattleEndDetailUpdated, info)
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

	err := db.BattleViewerUpsert(context.Background(), gamedb.Conn, btl.ID.String(), wsc.Identifier())
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
				btl.WarMachines[warMachineIndex].Position.X = y
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
	span := tracer.StartSpan("ws.Command", tracer.ResourceName(string(WSJoinQueue)))
	defer span.Finish()

	msg := &JoinPayload{}
	err := json.Unmarshal(payload, msg)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}

	mechId, err := db.MechIDFromHash(msg.Payload.AssetHash)
	if err != nil {
		gamelog.L.Error().Str("hash", msg.Payload.AssetHash).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	mech, err := db.Mech(mechId)
	if err != nil {
		gamelog.L.Error().Str("mech_id", mechId.String()).Err(err).Msg("unable to retrieve mech id from hash")
		return err
	}

	if mech.Faction == nil {
		gamelog.L.Error().Str("mech_id", mechId.String()).Err(err).Msg("mech's owner player has no faction")
		return err
	}

	ownerID, err := uuid.FromString(mech.OwnerID)
	if err != nil {
		gamelog.L.Error().Str("ownerID", mech.OwnerID).Err(err).Msg("unable to convert owner id from string")
		return err
	}

	pos, err := db.JoinQueue(&db.BattleMechData{
		MechID:    mechId,
		OwnerID:   ownerID,
		FactionID: uuid.UUID(factionID),
	})

	if err != nil {
		gamelog.L.Error().Interface("factionID", mech.FactionID).Err(err).Msg("unable to insert mech into queue")
		return err
	}

	reply(pos)

	return err
}

func (btl *Battle) Destroyed(dp *BattleWMDestroyedPayload) {
	// check destroyed war machine exist
	if btl.ID.String() != dp.BattleID {
		gamelog.L.Warn().Str("battle.ID", btl.ID.String()).Str("gameclient.ID", dp.BattleID).Msg("battle state does not match game client state")
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
			Str("battle_id", btl.ID.String()).
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
					Str("battle_id", btl.ID.String()).
					Msg("can't create uuid from non-empty related event idf")
			}
			dp.DestroyedWarMachineEvent.RelatedEventID = relatedEventuuid
		}

		evt := &db.BattleEvent{
			BattleID:  btl.ID,
			WM1:       warMachineID,
			WM2:       killByWarMachineID,
			EventType: db.Btlevnt_Killed,
			CreatedAt: time.Now(),
			RelatedID: dp.DestroyedWarMachineEvent.RelatedEventIDString,
		}

		_, err = db.StoreBattleEvent(btl.ID, dp.DestroyedWarMachineEvent.RelatedEventID, warMachineID, killByWarMachineID, db.Btlevnt_Killed, time.Now())
		if err != nil {
			gamelog.L.Warn().
				Interface("event_data", evt).
				Str("battle_id", btl.ID.String()).
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
			Str("battle_id", btl.ID.String()).
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
		gamelog.L.Warn().Str("battle_id", btl.ID.String()).Err(err).Msg("unable to load out queue")
		return err
	}

	if len(q) < 9 {
		gamelog.L.Warn().Msg("not enough mechs to field a battle. replacing with default battle.")

		err = btl.DefaultMechs()
		if err != nil {
			gamelog.L.Warn().Str("battle_id", btl.ID.String()).Err(err).Msg("unable to load default mechs")
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
		gamelog.L.Warn().Interface("mechs_ids", ids).Str("battle_id", btl.ID.String()).Err(err).Msg("failed to retrieve mechs from mech ids")
		return err
	}
	btl.WarMachines = btl.MechsToWarMachines(mechs)

	return nil
}

func (btl *Battle) MechsToWarMachines(mechs []*server.MechContainer) []*WarMachine {
	warmachines := make([]*WarMachine, len(mechs))
	for i, mech := range mechs {
		label := mech.Faction.Label
		if label == "" {
			gamelog.L.Warn().Interface("faction_id", mech.Faction.ID).Str("battle_id", btl.ID.String()).Msg("mech faction is an empty label")
		}
		if len(label) > 10 {
			words := strings.Split(label, " ")
			label = ""
			for _, word := range words {
				label = label + string([]rune(word)[0])
			}
		}

		weaponNames := make([]string, len(mech.Weapons))
		for k, wpn := range mech.Weapons {
			i, err := strconv.Atoi(k)
			if err != nil {
				gamelog.L.Warn().Str("key", k).Interface("weapon", wpn).Str("battle_id", btl.ID.String()).Msg("mech weapon's key is not an int")
			}
			weaponNames[i] = wpn.Label
		}

		model, ok := ModelMap[mech.Chassis.Model]
		if !ok {
			model = "WREX"
		}

		warmachines[i] = &WarMachine{
			ID:            mech.ID,
			Name:          mech.Name,
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
			Skin:               mech.Chassis.Skin,
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

var ModelMap = map[string]string{
	"Law Enforcer X-1000": "XFVS",
	"Olympus Mons LY07":   "BXSD",
	"Tenshi Mk1":          "WREX",
}
