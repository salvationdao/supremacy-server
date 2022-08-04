package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sasha-s/go-deadlock"
	"math"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const IncognitoGameAbilityID = 15
const BlackoutGameAbilityID = 16

const BlackoutDurationSeconds = 15 // has to match duration specified in supremacy-gameclient/abilities.json

type BlackoutEntry struct {
	GameCoords server.GameLocation
	CellCoords server.CellLocation
	ExpiresAt  time.Time
}

// PlayerAbilityManager tracks all player abilities and mech states that are active in the current battle
type PlayerAbilityManager struct {
	hiddenWarMachines map[string]time.Time // mech hash, expiry timestamp

	blackouts           map[string]BlackoutEntry //  timestamp-player_ability_id-owner_id, ability info
	hasBlackoutsUpdated bool

	deadlock.RWMutex
}

func NewPlayerAbilityManager() *PlayerAbilityManager {
	return &PlayerAbilityManager{
		hiddenWarMachines: make(map[string]time.Time),
		blackouts:         make(map[string]BlackoutEntry),
	}
}

func (pam *PlayerAbilityManager) ResetHasBlackoutsUpdated() {
	pam.Lock()
	defer pam.Unlock()

	pam.hasBlackoutsUpdated = false
}

func (pam *PlayerAbilityManager) HasBlackoutsUpdated() bool {
	pam.RLock()
	defer pam.RUnlock()

	return pam.hasBlackoutsUpdated
}

func (pam *PlayerAbilityManager) Blackouts() map[string]BlackoutEntry {
	pam.RLock()
	defer pam.RUnlock()

	return pam.blackouts
}

func (pam *PlayerAbilityManager) IsWarMachineInBlackout(position server.GameLocation) bool {
	pam.Lock()
	defer pam.Unlock()
	for id, b := range pam.blackouts {
		// Check if blackout is currently active, if not then delete it
		if time.Now().After(b.ExpiresAt) {
			delete(pam.blackouts, id)
			pam.hasBlackoutsUpdated = true
			continue
		}

		c1 := position
		c2 := b.GameCoords
		d := math.Sqrt(math.Pow(float64(c2.X)-float64(c1.X), 2) + math.Pow(float64(c2.Y)-float64(c1.Y), 2))
		if d < float64(BlackoutRadius) {
			return true
		}
	}
	return false
}

func (pam *PlayerAbilityManager) AddBlackout(id string, cellCoords server.CellLocation, gameCoords server.GameLocation) error {
	pam.Lock()
	defer pam.Unlock()

	_, ok := pam.blackouts[id]
	if ok {
		return fmt.Errorf("Blackout has already been cast")
	}
	pam.blackouts[id] = BlackoutEntry{
		CellCoords: cellCoords,
		GameCoords: gameCoords,
		ExpiresAt:  time.Now().Add(time.Duration(BlackoutDurationSeconds) * time.Second),
	}
	pam.hasBlackoutsUpdated = true

	return nil
}

// IsWarMachineHidden is called frequently and will remove mechs from the map if their hidden
// duration is up
func (pam *PlayerAbilityManager) IsWarMachineHidden(hash string) bool {
	pam.Lock()
	defer pam.Unlock()
	t, exists := pam.hiddenWarMachines[hash]
	if exists && time.Now().After(t) {
		delete(pam.hiddenWarMachines, hash)
		return false
	}

	return exists
}

func (pam *PlayerAbilityManager) AddHiddenWarMachineHash(hash string, duration time.Duration) error {
	pam.Lock()
	defer pam.Unlock()

	_, ok := pam.hiddenWarMachines[hash]
	if ok {
		return fmt.Errorf("War machine is already hidden")
	}
	pam.hiddenWarMachines[hash] = time.Now().Add(duration)

	return nil
}

func (pam *PlayerAbilityManager) RemoveHiddenWarMachineHash(hash string) {
	pam.Lock()
	defer pam.Unlock()

	_, ok := pam.hiddenWarMachines[hash]
	if !ok {
		return
	}
	delete(pam.hiddenWarMachines, hash)
}

type PlayerAbilityUseRequest struct {
	Payload struct {
		BlueprintAbilityID string               `json:"blueprint_ability_id"`
		LocationSelectType string               `json:"location_select_type"`
		StartCoords        *server.CellLocation `json:"start_coords"` // used for LINE_SELECT and LOCATION_SELECT abilities
		EndCoords          *server.CellLocation `json:"end_coords"`   // used only for LINE_SELECT abilities
		MechHash           string               `json:"mech_hash"`    // used only for MECH_SELECT abilities
	} `json:"payload"`
}

const HubKeyPlayerAbilityUse = "PLAYER:ABILITY:USE"

var playerAbilityBucket = leakybucket.NewCollector(1, 2, true)

func (arena *Arena) PlayerAbilityUse(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	b := playerAbilityBucket.Add(user.ID, 2)
	if b == 0 {
		return terror.Error(fmt.Errorf("Too many executions. Please wait a bit before trying again."))
	}

	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return terror.Error(fmt.Errorf("wrong battle state"), "There is no battle currently to use this ability on.")
	}

	// check player is banned
	isBanned, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(true),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to load player ban")
		return terror.Error(err, "Failed to trigger ability")
	}

	if isBanned {
		return terror.Error(fmt.Errorf("player is banned for triggering ability"), "You are banned for triggering ability")
	}

	req := &PlayerAbilityUseRequest{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Msg("invalid request received")
		return terror.Error(err, "Invalid request received")
	}

	// mech command handler
	if req.Payload.LocationSelectType == "MECH_COMMAND" {
		err := arena.MechMoveCommandCreateHandler(ctx, user, factionID, key, payload, reply)
		if err != nil {
			return terror.Error(err, "Failed to fire mech command")
		}

		return nil
	}

	player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(user.ID), qm.Load(boiler.PlayerRels.Faction)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("userID", user.ID).Msg("could not find player from given user ID")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	pa, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.BlueprintID.EQ(req.Payload.BlueprintAbilityID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(player.ID),
		qm.Load(boiler.PlayerAbilityRels.Blueprint),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("blueprintAbilityID", req.Payload.BlueprintAbilityID).Msg("failed to get player ability")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	if pa.Count < 1 {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Interface("playerAbility", pa).Msg("player ability count is 0, cannot be used")
		return terror.Error(err, "You do not have any more of this ability to use.")
	}

	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the PlayerAbilityUse!", r)
		}
	}()

	// check battle end
	if arena.CurrentBattle().stage.Load() == BattleStageEnd {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("battle stage has en ended")
		return nil
	}

	bpa := pa.R.Blueprint

	userID := uuid.FromStringOrNil(user.ID)
	var event *server.GameAbilityEvent
	switch req.Payload.LocationSelectType {
	case boiler.LocationSelectTypeEnumLINE_SELECT:
		if req.Payload.StartCoords == nil || req.Payload.EndCoords == nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("no start/end coords was provided for executing ability of type %s", boiler.LocationSelectTypeEnumLINE_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when executing this ability.")
		}
		if req.Payload.StartCoords.X < 0 || req.Payload.StartCoords.Y < 0 || req.Payload.EndCoords.X < 0 || req.Payload.EndCoords.Y < 0 {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("invalid start/end coords were provided for executing %s ability", boiler.LocationSelectTypeEnumLINE_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			GameLocation:        arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(req.Payload.StartCoords),
			GameLocationEnd:     arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(req.Payload.EndCoords),
		}

	case boiler.LocationSelectTypeEnumMECH_SELECT:
		if req.Payload.MechHash == "" {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Err(err).Msgf("no mech hash was provided for executing ability of type %s", boiler.LocationSelectTypeEnumMECH_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Mech hash must be provided to execute this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			WarMachineHash:      &req.Payload.MechHash,
		}
	case boiler.LocationSelectTypeEnumLOCATION_SELECT:
		if req.Payload.StartCoords == nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("no start coords was provided for executing ability of type %s", boiler.LocationSelectTypeEnumLOCATION_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when executing this ability.")
		}
		if req.Payload.StartCoords.X < 0 || req.Payload.StartCoords.Y < 0 {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("invalid start coords were provided for executing %s ability", boiler.LocationSelectTypeEnumLOCATION_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			GameLocation:        arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(req.Payload.StartCoords),
		}
	case boiler.LocationSelectTypeEnumGLOBAL:
	default:
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Interface("request payload", req.Payload).Msg("no location select type was provided when activating a player ability")
		return terror.Error(terror.ErrInvalidInput, "Something went wrong while activating this ability. Please try again, or contact support if this issue persists.")
	}

	if event == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Interface("request payload", req.Payload).Msg("game ability event is nil for some reason")
		return terror.Error(terror.ErrInvalidInput, "Something went wrong while activating this ability. Please try again, or contact support if this issue persists.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	defer tx.Rollback()

	// Create consumed_abilities entry
	ca := boiler.ConsumedAbility{
		BattleID:            arena.CurrentBattle().BattleID,
		ConsumedBy:          player.ID,
		BlueprintID:         pa.BlueprintID,
		GameClientAbilityID: bpa.GameClientAbilityID,
		Label:               bpa.Label,
		Colour:              bpa.Colour,
		ImageURL:            bpa.ImageURL,
		Description:         bpa.Description,
		TextColour:          bpa.TextColour,
		LocationSelectType:  bpa.LocationSelectType,
		ConsumedAt:          time.Now(),
	}
	err = ca.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Interface("consumedAbility", ca).Msg("failed to created consumed ability entry")
		return err
	}

	// Update the count of the player_abilities entry
	pa.Count = pa.Count - 1
	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Interface("playerAbility", pa).Msg("failed to update player ability count")
		return err
	}

	isIncognito := bpa.GameClientAbilityID == IncognitoGameAbilityID
	// If player ability is "Incognito"
	if isIncognito {
		wm, err := boiler.CollectionItems(boiler.CollectionItemWhere.Hash.EQ(req.Payload.MechHash)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Err(err).Msgf("failed to execute INCOGNITO ability: could not get war machine from hash %s", req.Payload.MechHash)
			return terror.Error(err, "Failed to get war machine from hash")
		}

		incognitoDurationSeconds := db.GetIntWithDefault(db.KeyPlayerAbilityIncognitoDurationSeconds, 20) // default 20 seconds

		err = arena.CurrentBattle().playerAbilityManager().AddHiddenWarMachineHash(wm.Hash, time.Second*time.Duration(incognitoDurationSeconds))
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("failed to execute Incognito player ability")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue executing player ability, please try again or contact support.")
	}
	reply(true)

	if !isIncognito {
		// Tell gameclient to execute ability
		arena.CurrentBattle().arena.Message("BATTLE:ABILITY", event)
	}

	pas, err := db.PlayerAbilitiesList(user.ID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("boiler func", "PlayerAbilities").Str("ownerID", user.ID).Err(err).Msg("unable to get player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}
	ws.PublishMessage(fmt.Sprintf("/user/%s/player_abilities", userID), server.HubKeyPlayerAbilitiesList, pas)

	if bpa.GameClientAbilityID == BlackoutGameAbilityID {
		cellCoords := req.Payload.StartCoords
		gameCoords := arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(req.Payload.StartCoords)
		err = arena.CurrentBattle().playerAbilityManager().AddBlackout(fmt.Sprintf("%d-%s-%s", time.Now().UnixNano(), pa.ID, pa.OwnerID), *cellCoords, *gameCoords)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to execute Incognito player ability")
			return err
		}
	}

	return nil
}

const MechMoveCommandCreateGameAbilityID = 8
const MechMoveCommandCancelGameAbilityID = 9

const HubKeyMechCommandsSubscribe = "MECH:COMMANDS:SUBSCRIBE"

func (arena *Arena) MechCommandsSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	err := arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		return terror.Error(err, "Failed to get mech command logs")
	}
	return nil
}

func (arena *Arena) BroadcastFactionMechCommands(factionID string) error {
	if arena.currentBattleState() != BattleStageStart {
		return nil
	}

	ids := arena.currentBattleWarMachineIDs(factionID)
	if len(ids) == 0 {
		return nil
	}

	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.IN(helpers.UUIDArray2StrArray(ids)),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech command logs from db")
		return terror.Error(err, "Failed to get mech command logs")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_commands", factionID), HubKeyMechCommandsSubscribe, mmc)

	return nil
}

const HubKeyMechMoveCommandSubscribe = "MECH:MOVE:COMMAND:SUBSCRIBE"

type MechMoveCommandResponse struct {
	*boiler.MechMoveCommandLog
	RemainCooldownSeconds int `json:"remain_cooldown_seconds"`
}

func (arena *Arena) MechMoveCommandSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	hash := cctx.URLParam("hash")

	wm := arena.CurrentBattleWarMachineByHash(hash)
	if wm == nil {
		return terror.Error(terror.ErrInvalidInput, "Current mech is not on the battlefield")
	}

	// query unfinished mech move command
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.DeletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech id", wm.ID).Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to get mech move command.")
	}

	resp := &MechMoveCommandResponse{
		RemainCooldownSeconds: 0,
	}

	if mmc != nil {
		resp.MechMoveCommandLog = mmc
		resp.RemainCooldownSeconds = MechMoveCooldownSeconds - int(time.Now().Sub(mmc.CreatedAt).Seconds())
		if resp.RemainCooldownSeconds < 0 {
			resp.RemainCooldownSeconds = 0
		}
	}

	reply(resp)
	return nil
}

func (arena *Arena) mechCommandAuthorisedCheck(userID string, wm *WarMachine) error {
	// check ownership
	if wm.OwnedByID != userID {
		gamelog.L.Warn().Str("mech id", wm.ID).Str("mech owner id", wm.OwnedByID).Str("current user id", userID).Msg("Unauthorised mech move command.")
		return terror.Error(terror.ErrForbidden, "The mech is not owned by current user.")
	}

	// TODO: check is general?

	// TODO: check is renter?

	return nil
}

const HubKeyWarMachineAbilityTrigger = "WAR:MACHINE:ABILITY:TRIGGER"

type MechAbilityTriggerRequest struct {
	Payload struct {
		Hash          string `json:"mech_hash"`
		GameAbilityID string `json:"game_ability_id"`
	} `json:"payload"`
}

var mechAbilityBucket = leakybucket.NewCollector(1, 1, true)

func (arena *Arena) MechAbilityTriggerHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// check battle stage
	if arena.currentBattleState() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	req := &MechAbilityTriggerRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if mechAbilityBucket.Add(req.Payload.Hash, 1) == 0 {
		return terror.Error(fmt.Errorf("too many request"), "Too many mech ability request.")
	}

	// get mech
	wm := arena.CurrentBattleWarMachineByHash(req.Payload.Hash)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battlefield.")
	}

	err = arena.mechCommandAuthorisedCheck(user.ID, wm)
	if err != nil {
		gamelog.L.Warn().Str("mech id", wm.ID).Str("user id", user.ID).Msg("Unauthorised mech command - create")
		return terror.Error(err, err.Error())
	}

	// get current battle
	bn := arena.currentBattleNumber()
	if bn == -1 {
		return terror.Error(fmt.Errorf("current battle is cleaned up"), "Current battle is cleaned up.")
	}

	// get ability
	a, err := boiler.FindGameAbility(gamedb.StdConn, req.Payload.GameAbilityID)
	if err != nil {
		return terror.Error(err, "Failed to load game ability")
	}

	// get cooldown timer
	abilityCooldownSeconds := db.GetIntWithDefault(db.KeyMechAbilityCoolDownSeconds, 30)

	// validate the ability can be triggered
	switch a.Label {
	case "REPAIR":
		// get ability from db
		lastTrigger, err := boiler.MechAbilityTriggerLogs(
			boiler.MechAbilityTriggerLogWhere.MechID.EQ(wm.ID),
			boiler.MechAbilityTriggerLogWhere.GameAbilityID.EQ(req.Payload.GameAbilityID),
			boiler.MechAbilityTriggerLogWhere.BattleNumber.EQ(bn),
			boiler.MechAbilityTriggerLogWhere.DeletedAt.IsNull(),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Failed to get last ability trigger")
		}

		if lastTrigger != nil {
			return terror.Error(fmt.Errorf("can only trigger once"), fmt.Sprintf("Repair can only be triggered once per battle."))
		}
	default:

		// get ability from db
		lastTrigger, err := boiler.MechAbilityTriggerLogs(
			boiler.MechAbilityTriggerLogWhere.MechID.EQ(wm.ID),
			boiler.MechAbilityTriggerLogWhere.GameAbilityID.EQ(req.Payload.GameAbilityID),
			boiler.MechAbilityTriggerLogWhere.CreatedAt.GT(time.Now().Add(time.Duration(-abilityCooldownSeconds)*time.Second)),
			boiler.MechAbilityTriggerLogWhere.DeletedAt.IsNull(),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Failed to get last ability trigger")
		}

		if lastTrigger != nil {
			return terror.Error(fmt.Errorf("ability is still cooling down"), fmt.Sprintf("The ability is still cooling down."))
		}
	}

	// get game ability
	ga, err := boiler.FindGameAbility(gamedb.StdConn, req.Payload.GameAbilityID)
	if err != nil {
		gamelog.L.Error().Str("ability id", req.Payload.GameAbilityID).Msg("Failed to get game ability from db")
		return terror.Error(err, "Failed to load game ability")
	}

	if ga.FactionID != wm.FactionID {
		return terror.Error(fmt.Errorf("invalid faction id"), "Targeted game ability is not from the same faction")
	}

	if ga.Level != boiler.AbilityLevelPLAYER {
		return terror.Error(fmt.Errorf("non player ability ability"), "Targeted game ability is not a player level ability.")
	}

	// trigger the ability
	now := time.Now()
	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: byte(ga.GameClientAbilityID),
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID,
		EventID:             uuid.Must(uuid.NewV4()),
	}

	// fire mech command
	arena.Message("BATTLE:ABILITY", event)

	// log mech move command
	mat := &boiler.MechAbilityTriggerLog{
		MechID:        wm.ID,
		TriggeredByID: user.ID,
		GameAbilityID: ga.ID,
		BattleNumber:  bn,
		CreatedAt:     now,
	}

	err = mat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("mech ability trigger", mat).Err(err).Msg("Failed to insert mech ability trigger.")
		return terror.Error(err, "Failed to record mech ability trigger")
	}

	// send notification
	arena.BroadcastGameNotificationWarMachineAbility(&GameNotificationWarMachineAbility{
		User: &UserBrief{
			ID:        uuid.FromStringOrNil(user.ID),
			Username:  user.Username.String,
			FactionID: user.FactionID.String,
		},
		Ability: &AbilityBrief{
			Label:    ga.Label,
			ImageUrl: ga.ImageURL,
			Colour:   ga.Colour,
		},
		WarMachine: &WarMachineBrief{
			ParticipantID: wm.ParticipantID,
			Hash:          wm.Hash,
			ImageUrl:      wm.Image,
			ImageAvatar:   wm.ImageAvatar,
			Name:          wm.Name,
			FactionID:     wm.FactionID,
		},
	})

	switch a.Label {
	case "REPAIR":
		// HACK: set cool down to 1 day, to implement once per battle
		ws.PublishMessage(fmt.Sprintf("/faction/%s/mech/%d/abilities/%s/cool_down_seconds", wm.FactionID, wm.ParticipantID, ga.ID), HubKeyWarMachineAbilitySubscribe, 86400)
	default:
		// broadcast cool down seconds
		ws.PublishMessage(fmt.Sprintf("/faction/%s/mech/%d/abilities/%s/cool_down_seconds", wm.FactionID, wm.ParticipantID, ga.ID), HubKeyWarMachineAbilitySubscribe, abilityCooldownSeconds)

	}

	return nil
}

type MechMoveCommandCreateRequest struct {
	Payload struct {
		Hash        string               `json:"mech_hash"`
		StartCoords *server.CellLocation `json:"start_coords"`
	} `json:"payload"`
}

const MechMoveCooldownSeconds = 5

// MechMoveCommandCreateHandler send mech move command to game client
func (arena *Arena) MechMoveCommandCreateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// check battle stage
	if arena.currentBattleState() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	req := &MechMoveCommandCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.StartCoords == nil {
		return terror.Error(fmt.Errorf("missing location"), "Missing location")
	}

	// get mech
	wm := arena.CurrentBattleWarMachineByHash(req.Payload.Hash)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battlefield.")
	}

	err = arena.mechCommandAuthorisedCheck(user.ID, wm)
	if err != nil {
		gamelog.L.Warn().Str("mech id", wm.ID).Str("user id", user.ID).Msg("Unauthorised mech command - create")
		return terror.Error(err, err.Error())
	}

	// check cell is disabled or not
	disableCells := arena.currentDisableCells()
	if disableCells == nil {
		return terror.Error(fmt.Errorf("no disabeld cells provided"), "The selected cell is disabled.")
	}

	selectedCell := int64(req.Payload.StartCoords.X + req.Payload.StartCoords.Y*arena.CurrentBattle().gameMap.CellsX)
	for _, dc := range disableCells {
		if dc == selectedCell {
			return terror.Error(fmt.Errorf("cell disabled"), "The selected cell is disabled.")
		}
	}

	// check mech move command is triggered within 5 seconds
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.CreatedAt.GT(time.Now().Add(-MechMoveCooldownSeconds*time.Second)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to trigger mech move command")
	}

	if mmc != nil {
		return terror.Error(fmt.Errorf("Command is still cooling down."))
	}

	now := time.Now()

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: MechMoveCommandCreateGameAbilityID, // 8
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID, // trigger on war machine
		TriggeredOnCellX:    &req.Payload.StartCoords.X,
		TriggeredOnCellY:    &req.Payload.StartCoords.Y,
		EventID:             uuid.Must(uuid.NewV4()),
		GameLocation: arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(&server.CellLocation{
			X: req.Payload.StartCoords.X,
			Y: req.Payload.StartCoords.Y,
		}),
	}

	// check mech command
	arena.Message("BATTLE:ABILITY", event)

	// cancel any unfinished move commands of the mech
	_, err = boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
	).UpdateAll(gamedb.StdConn, boiler.M{boiler.MechMoveCommandLogColumns.CancelledAt: time.Now()})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech id", wm.ID).Str("battle id", arena.CurrentBattle().ID).Err(err).Msg("Failed to cancel unfinished mech move command in db")
		return terror.Error(err, "Failed to update mech move command.")
	}

	// log mech move command
	mmc = &boiler.MechMoveCommandLog{
		MechID:        wm.ID,
		TriggeredByID: user.ID,
		CellX:         req.Payload.StartCoords.X,
		CellY:         req.Payload.StartCoords.Y,
		BattleID:      arena.CurrentBattle().ID,
		CreatedAt:     now,
	}

	err = mmc.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to insert mech move command")
		return terror.Error(err, "Failed to trigger mech move command.")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_command/%s", factionID, wm.Hash), HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
		MechMoveCommandLog:    mmc,
		RemainCooldownSeconds: MechMoveCooldownSeconds,
	})

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to broadcast faction mech commands")
	}

	reply(true)

	return nil
}

const HubKeyMechMoveCommandCancel = "MECH:MOVE:COMMAND:CANCEL"

type MechMoveCommandCancelRequest struct {
	Payload struct {
		Hash          string `json:"hash"`
		MoveCommandID string `json:"move_command_id"`
	} `json:"payload"`
}

// MechMoveCommandCancelHandler send cancel mech move command to game client
func (arena *Arena) MechMoveCommandCancelHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// check battle stage
	if arena.currentBattleState() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	req := &MechMoveCommandCancelRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	wm := arena.CurrentBattleWarMachineByHash(req.Payload.Hash)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battlefield.")
	}

	// check ownership
	if wm.OwnedByID != user.ID {
		gamelog.L.Warn().Str("mech owner id", wm.OwnedByID).Str("player id", user.ID).Msg("Invalid mech move cancel request")
		return terror.Error(terror.ErrForbidden, "This mech is not owned by the player.")
	}

	err = arena.mechCommandAuthorisedCheck(user.ID, wm)
	if err != nil {
		gamelog.L.Warn().Str("mech id", wm.ID).Str("user id", user.ID).Msg("Unauthorised mech command - cancel")
		return terror.Error(err, err.Error())
	}

	// get mech move command
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.ID.EQ(req.Payload.MoveCommandID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		qm.OrderBy(boiler.MechMoveCommandLogColumns.CreatedAt+" DESC"),
		qm.Load(boiler.MechMoveCommandLogRels.Mech),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("mech move command id", req.Payload.MoveCommandID).Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to cancel mech move command.")
	}

	// check mech id
	if mmc.MechID != wm.ID {
		gamelog.L.Warn().Str("mech move command id", mmc.ID).Str("expected mech id", mmc.MechID).Str("provided mech id", wm.ID).Msg("mech id mismatch")
		return terror.Error(fmt.Errorf("mech id mismatch"), "Failed to cancel mech move command")
	}

	if mmc.CancelledAt.Valid {
		return terror.Error(fmt.Errorf("move command is already cancelled"), "Mech move command is already cancelled.")
	}

	if mmc.ReachedAt.Valid {
		return terror.Error(fmt.Errorf("mech already reach the place"), "Mech already reach the commanded spot")
	}

	// cancel command
	mmc.CancelledAt = null.TimeFrom(time.Now())
	_, err = mmc.Update(gamedb.StdConn, boil.Whitelist(boiler.MechMoveCommandLogColumns.CancelledAt))
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("mech move command id", mmc.ID).Msg("Failed to up date mech move command in db")
		return terror.Error(err, "Failed to cancel mech move command")
	}

	// send mech move command to game client
	arena.Message("BATTLE:ABILITY", &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: MechMoveCommandCancelGameAbilityID,
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID, // trigger on war machine
	})

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_command/%s", factionID, wm.Hash), HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
		MechMoveCommandLog:    mmc,
		RemainCooldownSeconds: MechMoveCooldownSeconds - int(time.Now().Sub(mmc.CreatedAt).Seconds()),
	})

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to broadcast faction mech commands")
	}

	reply(true)

	return nil
}

const HubKeyBattleAbilityOptIn = "BATTLE:ABILITY:OPT:IN"

var optInBucket = leakybucket.NewCollector(1, 1, true)

func (arena *Arena) BattleAbilityOptIn(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if optInBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many Requests")
	}

	btl := arena.CurrentBattle()
	if btl == nil {
		return terror.Error(fmt.Errorf("battle is endded"), "Battle has not started yet.")
	}

	as := btl.AbilitySystem()
	if as == nil {
		return terror.Error(fmt.Errorf("ability system is closed"), "Ability system is closed.")
	}

	if !AbilitySystemIsAvailable(as) {
		return terror.Error(fmt.Errorf("ability system si not available"), "Ability is not ready.")
	}

	if as.BattleAbilityPool.Stage.Phase.Load() != BribeStageOptIn {
		return terror.Error(fmt.Errorf("invlid phase"), "It is not in the stage for player to opt in.")
	}

	ba := *as.BattleAbilityPool.BattleAbility.LoadBattleAbility()
	offeringID := as.BattleAbilityPool.BattleAbility.LoadOfferingID()

	bao := boiler.BattleAbilityOptInLog{
		BattleID:                btl.BattleID,
		PlayerID:                user.ID,
		BattleAbilityOfferingID: offeringID,
		FactionID:               factionID,
		BattleAbilityID:         ba.ID,
	}
	err := bao.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to opt in battle ability")
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s/battle_ability/check_opt_in", user.ID), HubKeyBattleAbilityOptInCheck, true)

	return nil
}
