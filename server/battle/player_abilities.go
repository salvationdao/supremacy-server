package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/battle/player_abilities"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/sasha-s/go-deadlock"

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
const SupportMechGameAbility = 18

const BlackoutDurationSeconds = 15 // has to match duration specified in supremacy-gameclient/abilities.json

// PlayerAbilityManager tracks all player abilities and mech states that are active in the current battle
type PlayerAbilityManager struct {
	hiddenWarMachines map[string]time.Time // map[mech hash]expiry timestamp

	blackouts           map[string]*player_abilities.BlackoutEntry // map[timestamp-player_ability_id-owner_id]ability info
	hasBlackoutsUpdated bool

	movingMiniMechs map[string]*player_abilities.MiniMechMoveCommand // map[mech_hash]mini mech move entry

	MiniMechMoveCoooldownSeconds int

	deadlock.RWMutex
}

func NewPlayerAbilityManager() *PlayerAbilityManager {
	return &PlayerAbilityManager{
		hiddenWarMachines:            make(map[string]time.Time),
		blackouts:                    make(map[string]*player_abilities.BlackoutEntry),
		movingMiniMechs:              make(map[string]*player_abilities.MiniMechMoveCommand),
		MiniMechMoveCoooldownSeconds: db.GetIntWithDefault(db.KeyPlayerAbilityMiniMechMoveCommandCooldownSeconds, 0), // default 0; i.e. no cooldown
	}
}

func (pam *PlayerAbilityManager) GetMiniMechMove(hash string) (*player_abilities.MiniMechMoveCommand, error) {
	pam.RLock()
	defer pam.RUnlock()

	mm, ok := pam.movingMiniMechs[hash]
	if !ok {
		return nil, fmt.Errorf("Mini mech is not moving")
	}

	return mm, nil
}

func (pam *PlayerAbilityManager) DeleteMiniMechMove(hash string) {
	pam.Lock()
	defer pam.Unlock()

	_, ok := pam.movingMiniMechs[hash]
	if ok {
		delete(pam.movingMiniMechs, hash)
	}
}

func (pam *PlayerAbilityManager) CancelMiniMechMove(hash string) (*player_abilities.MiniMechMoveCommand, error) {
	pam.Lock()
	defer pam.Unlock()

	mm, ok := pam.movingMiniMechs[hash]
	if !ok {
		return nil, fmt.Errorf("Could not find mini mech move command to mark as cancelled")
	}

	mm.Cancel()
	return mm, nil
}

func (pam *PlayerAbilityManager) CompleteMiniMechMove(hash string) (*player_abilities.MiniMechMoveCommand, error) {
	pam.Lock()
	defer pam.Unlock()

	mm, ok := pam.movingMiniMechs[hash]
	if !ok {
		return nil, fmt.Errorf("Could not find mini mech move command to mark as complete")
	}

	mm.Complete()
	return mm, nil
}

func (pam *PlayerAbilityManager) MovingFactionMiniMechs(factionID string) []*player_abilities.MiniMechMoveCommand {
	pam.RLock()
	defer pam.RUnlock()

	result := []*player_abilities.MiniMechMoveCommand{}
	for _, mmm := range pam.movingMiniMechs {
		mmm.Read(func(mmmc *player_abilities.MiniMechMoveCommand) {
			if mmm.FactionID != factionID || mmm.ReachedAt.Valid || mmm.CancelledAt.Valid {
				return
			}

			result = append(result, mmm)
		})
	}

	return result
}

func (pam *PlayerAbilityManager) IssueMiniMechMoveCommand(hash string, factionID string, triggeredByID string, cellX int, cellY int, battleID string) (*player_abilities.MiniMechMoveCommand, error) {
	pam.Lock()
	defer pam.Unlock()

	mm, ok := pam.movingMiniMechs[hash]
	if ok {
		onCooldown := true
		mm.Read(func(mmmc *player_abilities.MiniMechMoveCommand) {
			onCooldown = time.Now().Before(mmmc.CooldownExpiry)
		})
		if onCooldown {
			return nil, fmt.Errorf("Command is still cooling down. Please wait another %f seconds.", time.Until(mm.CooldownExpiry).Seconds())
		}
	}

	newMm := &player_abilities.MiniMechMoveCommand{
		BattleID:       battleID,
		CellX:          cellX,
		CellY:          cellY,
		TriggeredByID:  triggeredByID,
		FactionID:      factionID,
		MechHash:       hash,
		CooldownExpiry: time.Now().Add(time.Duration(pam.MiniMechMoveCoooldownSeconds) * time.Second),
		CancelledAt:    null.TimeFromPtr(nil),
		ReachedAt:      null.TimeFromPtr(nil),
		CreatedAt:      time.Now(),
		IsMoving:       true,
	}
	pam.movingMiniMechs[hash] = newMm

	return newMm, nil
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

func (pam *PlayerAbilityManager) Blackouts() map[string]*player_abilities.BlackoutEntry {
	pam.RLock()
	defer pam.RUnlock()

	return pam.blackouts
}

func (pam *PlayerAbilityManager) IsWarMachineInBlackout(position server.GameLocation) bool {
	pam.Lock()
	defer pam.Unlock()
	for id, b := range pam.blackouts {
		// Check if blackout is currently active, if not then delete it
		if b.IsExpired() {
			delete(pam.blackouts, id)
			pam.hasBlackoutsUpdated = true
			continue
		}

		if b.ContainsPosition(position) {
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
	pam.blackouts[id] = &player_abilities.BlackoutEntry{
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
		ArenaID string `json:"arena_id"`

		BlueprintAbilityID string               `json:"blueprint_ability_id"`
		LocationSelectType string               `json:"location_select_type"`
		StartCoords        *server.CellLocation `json:"start_coords"` // used for LINE_SELECT and LOCATION_SELECT abilities
		EndCoords          *server.CellLocation `json:"end_coords"`   // used only for LINE_SELECT abilities
		MechHash           string               `json:"mech_hash"`    // used only for MECH_SELECT abilities
	} `json:"payload"`
}

const HubKeyPlayerAbilityUse = "PLAYER:ABILITY:USE"

func (am *ArenaManager) PlayerAbilityUse(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
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

	arena, err := am.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return terror.Error(fmt.Errorf("wrong battle state"), "There is no battle currently to use this ability on.")
	}

	if arena.currentBattleState() != BattleStageStart {
		return terror.Error(terror.ErrForbidden, "You cannot execute an ability when the battle has not started yet.")
	}

	if arena.IsRunningAIDrivenMatch() {
		return terror.Error(fmt.Errorf("no ability is allowed for AI driven match"), "Player abilities are not allowed during AI driven match.")
	}

	// mech command handler
	if req.Payload.LocationSelectType == "MECH_COMMAND" {
		err = arena.MechMoveCommandCreateHandler(ctx, user, factionID, key, payload, reply)
		if err != nil {
			return err
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
		qm.Load(boiler.PlayerAbilityRels.Owner),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("func", "PlayerAbilityUse").Str("blueprintAbilityID", req.Payload.BlueprintAbilityID).Msg("failed to get player ability")
		return terror.Error(err, "Something went wrong while activating this ability. Please try again or contact support if this issue persists.")
	}

	if pa.Count < 1 {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Interface("playerAbility", pa).Msg("player ability count is 0, cannot be used")
		return terror.Error(err, "You do not have any more of this ability to use.")
	}

	if time.Now().Before(pa.CooldownExpiresOn) {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Interface("playerAbility", pa).Msg("player ability is on cooldown")
		minutes := int(time.Until(pa.CooldownExpiresOn).Minutes())
		msg := fmt.Sprintf("Please try again in %d minutes.", minutes)
		if minutes < 1 {
			msg = fmt.Sprintf("Please try again in %d seconds.", int(time.Until(pa.CooldownExpiresOn).Seconds()))
		}
		return terror.Error(fmt.Errorf("This ability is still on cooldown. %s", msg))
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

	offeringID := uuid.Must(uuid.NewV4())

	bpa := pa.R.Blueprint

	userID := uuid.FromStringOrNil(user.ID)
	var event *server.GameAbilityEvent
	switch req.Payload.LocationSelectType {
	case boiler.LocationSelectTypeEnumLINE_SELECT:
		if req.Payload.StartCoords == nil || req.Payload.EndCoords == nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("no start/end coords was provided for executing ability of type %s", boiler.LocationSelectTypeEnumLINE_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when executing this ability.")
		}
		if req.Payload.StartCoords.X.IsNegative() || req.Payload.StartCoords.Y.IsNegative() || req.Payload.EndCoords.X.IsNegative() || req.Payload.EndCoords.Y.IsNegative() {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("invalid start/end coords were provided for executing %s ability", boiler.LocationSelectTypeEnumLINE_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             offeringID,
			FactionID:           &player.FactionID.String,
			GameLocation:        arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(req.Payload.StartCoords),
			GameLocationEnd:     arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(req.Payload.EndCoords),
		}

	case boiler.LocationSelectTypeEnumMECH_SELECT:
		if req.Payload.MechHash == "" {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Err(err).Msgf("no mech hash was provided for executing ability of type %s", boiler.LocationSelectTypeEnumMECH_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Mech hash must be provided to execute this ability.")
		}

		// check the mech is in the battlefield
		wm := arena.CurrentBattleWarMachineByHash(req.Payload.MechHash)
		if wm == nil {
			return terror.Error(fmt.Errorf("mech not found"), "The mech is not in the battlefield.")
		}

		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             offeringID,
			FactionID:           &player.FactionID.String,
			WarMachineHash:      &req.Payload.MechHash,
		}
	case boiler.LocationSelectTypeEnumMECH_SELECT_ALLIED:
		if req.Payload.MechHash == "" {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Err(err).Msgf("no mech hash was provided for executing ability of type %s", boiler.LocationSelectTypeEnumMECH_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Mech hash must be provided to execute this ability.")
		}

		// check the mech is in the battlefield
		wm := arena.CurrentBattleWarMachineByHash(req.Payload.MechHash)
		if wm == nil {
			return terror.Error(fmt.Errorf("mech not found"), "The mech is not in the battlefield.")
		}

		// check the mech is an ally mech
		if wm.FactionID != factionID {
			return terror.Error(fmt.Errorf("not ally mech"), "Must select a ally mech.")
		}

		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             offeringID,
			FactionID:           &player.FactionID.String,
			WarMachineHash:      &req.Payload.MechHash,
		}
	case boiler.LocationSelectTypeEnumMECH_SELECT_OPPONENT:
		if req.Payload.MechHash == "" {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Err(err).Msgf("no mech hash was provided for executing ability of type %s", boiler.LocationSelectTypeEnumMECH_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Mech hash must be provided to execute this ability.")
		}

		// check the mech is in the battlefield
		wm := arena.CurrentBattleWarMachineByHash(req.Payload.MechHash)
		if wm == nil {
			return terror.Error(fmt.Errorf("mech not found"), "The mech is not in the battlefield.")
		}

		// if hacker drone
		if wm.Status.IsHacked && bpa.GameClientAbilityID == 13 {
			return terror.Error(fmt.Errorf("already hacked"), "The mech is already hacked.")
		}

		// check the mech is in the
		if wm.FactionID == factionID {
			return terror.Error(fmt.Errorf("not opponent mech"), "Must select an opponent mech.")
		}

		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             offeringID,
			FactionID:           &player.FactionID.String,
			WarMachineHash:      &req.Payload.MechHash,
		}
	case boiler.LocationSelectTypeEnumLOCATION_SELECT:
		if req.Payload.StartCoords == nil {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("no start coords was provided for executing ability of type %s", boiler.LocationSelectTypeEnumLOCATION_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when executing this ability.")
		}
		if req.Payload.StartCoords.X.IsNegative() || req.Payload.StartCoords.Y.IsNegative() {
			gamelog.L.Error().Str("log_name", "battle arena").Interface("request payload", req.Payload).Msgf("invalid start coords were provided for executing %s ability", boiler.LocationSelectTypeEnumLOCATION_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             offeringID,
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

	// check if using support war machine
	if bpa.GameClientAbilityID == SupportMechGameAbility {
		btl := arena.CurrentBattle()

		swm := 0
		for _, wm := range btl.SpawnedAI {
			// add if mini mech && alive && same faction as owner
			isMiniMech := wm.AIType != nil && *wm.AIType == MiniMech
			isAlive := wm.Health > 0
			inOwnersFaction := pa.R != nil && pa.R.Owner.FactionID == null.StringFrom(wm.FactionID)
			if isMiniMech && isAlive && inOwnersFaction {
				swm++
			}
		}

		if swm >= 3 {
			gamelog.L.Debug().Msg("too many support warmachines for this faction")
			return terror.Error(fmt.Errorf("too many support warmachines for this faction"), "Only 3 active support war machines allowed in battle at one time for your faction")
		}
	}

	// Create consumed_abilities entry
	ca := boiler.ConsumedAbility{
		BattleID:            arena.CurrentBattle().ID,
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

	// Update the count and cooldown expiry of the player_abilities entry
	pa.Count = pa.Count - 1
	pa.CooldownExpiresOn = time.Now().Add(time.Second * time.Duration(bpa.CooldownSeconds))
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

	if btl := arena.CurrentBattle(); btl != nil && !isIncognito {
		// record ability on display list if needed
		if bpa.DisplayOnMiniMap {
			mma := &MiniMapAbilityContent{
				OfferingID:               offeringID.String(),
				LocationSelectType:       bpa.LocationSelectType,
				ImageUrl:                 bpa.ImageURL,
				Colour:                   bpa.Colour,
				MiniMapDisplayEffectType: bpa.MiniMapDisplayEffectType,
				MechDisplayEffectType:    bpa.MechDisplayEffectType,
			}

			switch mma.LocationSelectType {
			case boiler.LocationSelectTypeEnumLINE_SELECT, boiler.LocationSelectTypeEnumLOCATION_SELECT:
				mma.Location = *req.Payload.StartCoords
			case boiler.LocationSelectTypeEnumMECH_SELECT, boiler.LocationSelectTypeEnumMECH_SELECT_OPPONENT, boiler.LocationSelectTypeEnumMECH_SELECT_ALLIED:
				if wm := arena.CurrentBattleWarMachineByHash(req.Payload.MechHash); wm != nil {
					mma.MechID = wm.ID
				}
			}

			// set radius
			if ability := btl.abilityDetails[bpa.GameClientAbilityID]; ability != nil && ability.Radius > 0 {
				mma.Radius = null.IntFrom(ability.Radius)
			}

			// set delay seconds
			if bpa.LaunchingDelaySeconds > 0 {
				mma.LaunchingAt = null.TimeFrom(time.Now().Add(time.Duration(bpa.LaunchingDelaySeconds) * time.Second))
			}

			ws.PublishMessage(
				fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", arena.ID),
				server.HubKeyMiniMapAbilityDisplayList,
				btl.MiniMapAbilityDisplayList.Add(offeringID.String(), mma),
			)

			if bpa.AnimationDurationSeconds > 0 {
				go func(battle *Battle, bpa *boiler.BlueprintPlayerAbility) {
					time.Sleep(time.Duration(bpa.AnimationDurationSeconds) * time.Second)
					if battle != nil && battle.stage.Load() == BattleStageStart {
						if ab := battle.MiniMapAbilityDisplayList.Get(offeringID.String()); ab != nil {
							ws.PublishMessage(
								fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", arena.ID),
								server.HubKeyMiniMapAbilityDisplayList,
								battle.MiniMapAbilityDisplayList.Remove(offeringID.String()),
							)
						}
					}
				}(btl, bpa)
			}
		}

		// Tell game client to execute ability
		btl.arena.Message("BATTLE:ABILITY", event)
	}

	pas, err := db.PlayerAbilitiesList(user.ID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("boiler func", "PlayerAbilities").Str("ownerID", user.ID).Err(err).Msg("unable to get player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}
	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/player_abilities", userID), server.HubKeyPlayerAbilitiesList, pas)

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

func (am *ArenaManager) MechCommandsSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		return terror.Error(err, "Failed to get mech command logs")
	}
	return nil
}

type FactionMechCommands struct {
	BattleID string `json:"battle_id"`
	CellX    int    `json:"cell_x"`
	CellY    int    `json:"cell_y"`
	IsAI     bool   `json:"is_ai"`
}

func (arena *Arena) BroadcastFactionMechCommands(factionID string) error {
	if arena.currentBattleState() != BattleStageStart {
		return nil
	}

	ids := arena.currentBattleWarMachineIDs(factionID)
	if len(ids) == 0 {
		return nil
	}

	logs, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.IN(ids),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech command logs from db")
		return terror.Error(err, "Failed to get mech command logs")
	}

	result := []*FactionMechCommands{}
	for _, l := range logs {
		result = append(result, &FactionMechCommands{
			BattleID: l.BattleID,
			CellX:    l.CellX,
			CellY:    l.CellY,
			IsAI:     false,
		})
	}

	movingMiniMechs := arena._currentBattle.playerAbilityManager().MovingFactionMiniMechs(factionID)
	for _, mm := range movingMiniMechs {
		mm.Read(func(mmmc *player_abilities.MiniMechMoveCommand) {
			result = append(result, &FactionMechCommands{
				BattleID: mm.BattleID,
				CellX:    mm.CellX,
				CellY:    mm.CellY,
				IsAI:     true,
			})
		})
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_commands", factionID, arena.ID), HubKeyMechCommandsSubscribe, result)

	return nil
}

type MechMoveCommandResponse struct {
	*boiler.MechMoveCommandLog
	IsMiniMech bool `json:"is_mini_mech"`
}

func (am *ArenaManager) MechMoveCommandSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	arena, err := am.GetArenaFromContext(ctx)
	if err != nil {
		return err
	}

	if arena.currentBattleState() != BattleStageStart {
		return terror.Error(terror.ErrForbidden, "There is no current battle")
	}

	hash := chi.RouteContext(ctx).URLParam("hash")

	wm := arena.CurrentBattleWarMachineOrAIByHash(hash)
	if wm == nil {
		return terror.Error(terror.ErrInvalidInput, "Current mech is not on the battlefield")
	}

	resp := &MechMoveCommandResponse{}
	isMiniMech := wm.AIType != nil && *wm.AIType == MiniMech
	if !isMiniMech {
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

		if mmc != nil {
			resp.MechMoveCommandLog = mmc
		}
	} else {
		mmmc, _ := arena._currentBattle.playerAbilityManager().GetMiniMechMove(wm.Hash)

		if mmmc != nil {
			mmmc.Read(func(mmmc *player_abilities.MiniMechMoveCommand) {
				resp.MechMoveCommandLog = &boiler.MechMoveCommandLog{
					ID:            fmt.Sprintf("%s_%s", mmmc.BattleID, mmmc.MechHash),
					BattleID:      mmmc.BattleID,
					MechID:        mmmc.MechHash,
					TriggeredByID: mmmc.TriggeredByID,
					CellX:         mmmc.CellX,
					CellY:         mmmc.CellY,
					CancelledAt:   mmmc.CancelledAt,
					ReachedAt:     mmmc.ReachedAt,
					IsMoving:      mmmc.IsMoving,
				}
				resp.IsMiniMech = true
			})
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
		ArenaID string `json:"arena_id"`

		Hash          string `json:"mech_hash"`
		GameAbilityID string `json:"game_ability_id"`
	} `json:"payload"`
}

var mechAbilityBucket = leakybucket.NewCollector(1, 1, true)

func (am *ArenaManager) MechAbilityTriggerHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MechAbilityTriggerRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	arena, err := am.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	// check battle stage
	btl := arena.CurrentBattle()
	if btl == nil || btl.stage.Load() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	if arena.IsRunningAIDrivenMatch() {
		return terror.Error(fmt.Errorf("no ability is allowed for AI driven match"), "Mech abilities are not allowed during AI driven match.")
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
		lastTrigger, err := boiler.BattleAbilityTriggers(
			boiler.BattleAbilityTriggerWhere.OnMechID.EQ(null.StringFrom(wm.ID)),
			boiler.BattleAbilityTriggerWhere.GameAbilityID.EQ(req.Payload.GameAbilityID),
			boiler.BattleAbilityTriggerWhere.BattleID.EQ(btl.ID),
			boiler.BattleAbilityTriggerWhere.TriggerType.EQ(boiler.AbilityTriggerTypeMECH_ABILITY),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Failed to get last ability trigger")
		}

		if lastTrigger != nil {
			return terror.Error(fmt.Errorf("can only trigger once"), fmt.Sprintf("Repair can only be triggered once per battle."))
		}
	default:
		// get ability from db
		lastTrigger, err := boiler.BattleAbilityTriggers(
			boiler.BattleAbilityTriggerWhere.OnMechID.EQ(null.StringFrom(wm.ID)),
			boiler.BattleAbilityTriggerWhere.GameAbilityID.EQ(req.Payload.GameAbilityID),
			boiler.BattleAbilityTriggerWhere.TriggeredAt.GT(time.Now().Add(time.Duration(-abilityCooldownSeconds)*time.Second)),
			boiler.BattleAbilityTriggerWhere.TriggerType.EQ(boiler.AbilityTriggerTypeMECH_ABILITY),
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

	if ga.DisplayOnMiniMap {
		go func(arena *Arena, gameAbility *boiler.GameAbility, mechID string) {
			btl := arena.CurrentBattle()
			if btl == nil || btl.stage.Load() == BattleStageEnd {
				return
			}

			offeringID := uuid.Must(uuid.NewV4())

			mma := &MiniMapAbilityContent{
				OfferingID:               offeringID.String(),
				LocationSelectType:       ga.LocationSelectType,
				ImageUrl:                 ga.ImageURL,
				Colour:                   ga.Colour,
				MiniMapDisplayEffectType: ga.MiniMapDisplayEffectType,
				MechDisplayEffectType:    ga.MechDisplayEffectType,
				MechID:                   wm.ID,
			}

			ws.PublishMessage(
				fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", arena.ID),
				server.HubKeyMiniMapAbilityDisplayList,
				btl.MiniMapAbilityDisplayList.Add(offeringID.String(), mma),
			)

			// cancel ability after animation end
			if gameAbility.AnimationDurationSeconds > 0 {
				time.Sleep(time.Duration(gameAbility.AnimationDurationSeconds) * time.Second)
				ws.PublishMessage(
					fmt.Sprintf("/public/arena/%s/mini_map_ability_display_list", arena.ID),
					server.HubKeyMiniMapAbilityDisplayList,
					btl.MiniMapAbilityDisplayList.Remove(offeringID.String()),
				)
			}
		}(arena, ga, wm.ID)
	}

	now := time.Now()
	offeringID := uuid.Must(uuid.NewV4())

	// log mech move command
	mat := &boiler.BattleAbilityTrigger{
		OnMechID:          null.StringFrom(wm.ID),
		PlayerID:          null.StringFrom(user.ID),
		GameAbilityID:     ga.ID,
		AbilityLabel:      ga.Label,
		IsAllSyndicates:   false,
		FactionID:         factionID,
		BattleID:          btl.ID,
		AbilityOfferingID: offeringID.String(),
		TriggeredAt:       now,
		TriggerType:       boiler.AbilityTriggerTypeMECH_ABILITY,
	}

	err = mat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("mech ability trigger", mat).Err(err).Msg("Failed to insert mech ability trigger.")
		return terror.Error(err, "Failed to record mech ability trigger")
	}

	// trigger the ability
	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: byte(ga.GameClientAbilityID),
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID,
		EventID:             offeringID,
		FactionID:           &factionID,
	}

	// fire mech command
	arena.Message("BATTLE:ABILITY", event)

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
		ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech/%d/abilities/%s/cool_down_seconds", wm.FactionID, arena.ID, wm.ParticipantID, ga.ID), HubKeyWarMachineAbilitySubscribe, 86400)
	default:
		// broadcast cool down seconds
		ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech/%d/abilities/%s/cool_down_seconds", wm.FactionID, arena.ID, wm.ParticipantID, ga.ID), HubKeyWarMachineAbilitySubscribe, abilityCooldownSeconds)
	}

	return nil
}

type MechMoveCommandCreateRequest struct {
	Payload struct {
		Hash        string               `json:"mech_hash"`
		StartCoords *server.CellLocation `json:"start_coords"`
	} `json:"payload"`
}

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
	wm := arena.CurrentBattleWarMachineOrAIByHash(req.Payload.Hash)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battlefield.")
	}

	err = arena.mechCommandAuthorisedCheck(user.ID, wm)
	if err != nil {
		gamelog.L.Warn().Str("mech id", wm.ID).Str("user id", user.ID).Msg("Unauthorised mech command - create")
		return terror.Error(err, err.Error())
	}

	// Only perform mech move command db checks if war machine is not a mini mech
	isMiniMech := wm.AIType != nil && *wm.AIType == MiniMech
	if isMiniMech {
		_, err := arena._currentBattle.playerAbilityManager().IssueMiniMechMoveCommand(
			wm.Hash,
			wm.FactionID,
			user.ID,
			int(req.Payload.StartCoords.X.IntPart()),
			int(req.Payload.StartCoords.Y.IntPart()),
			arena.CurrentBattle().ID,
		)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to issue mini mech move command")
			return terror.Error(err, "Failed to trigger mini mech move command")
		}
	}

	now := time.Now()

	// register channel
	eventID := uuid.Must(uuid.NewV4())
	ch := make(chan bool)
	arena.MechCommandCheckMap.Register(eventID.String(), ch)

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: MechMoveCommandCreateGameAbilityID, // 8
		WarMachineHash:      &wm.Hash,
		ParticipantID:       &wm.ParticipantID, // trigger on war machine
		EventID:             eventID,
		GameLocation: arena.CurrentBattle().getGameWorldCoordinatesFromCellXY(&server.CellLocation{
			X: req.Payload.StartCoords.X,
			Y: req.Payload.StartCoords.Y,
		}),
	}

	// check mech command
	arena.Message("BATTLE:ABILITY", event)

	// wait for the message to come back
	select {
	case isValidLocation := <-ch:
		// remove channel after complete
		arena.MechCommandCheckMap.Remove(eventID.String())
		if !isValidLocation {
			return terror.Error(fmt.Errorf("invalid location"), "Selected location is not valid.")
		}
	case <-time.After(1500 * time.Millisecond):
		// remove channel after 1.5s timeout
		arena.MechCommandCheckMap.Remove(eventID.String())
		return terror.Error(fmt.Errorf("failed to check location is valid"), "Selected location is not valid.")
	}

	if !isMiniMech {
		// cancel any unfinished move commands of the mech
		_, err = boiler.MechMoveCommandLogs(
			boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
			boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
			boiler.MechMoveCommandLogWhere.CancelledAt.IsNull(),
			boiler.MechMoveCommandLogWhere.ReachedAt.IsNull(),
		).UpdateAll(gamedb.StdConn, boiler.M{
			boiler.MechMoveCommandLogColumns.CancelledAt: time.Now(),
			boiler.MechMoveCommandLogColumns.IsMoving:    false,
		})
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech id", wm.ID).Str("battle id", arena.CurrentBattle().ID).Err(err).Msg("Failed to cancel unfinished mech move command in db")
			return terror.Error(err, "Failed to update mech move command.")
		}

		// log mech move command
		mmc := &boiler.MechMoveCommandLog{
			ArenaID:       arena.ID,
			MechID:        wm.ID,
			TriggeredByID: user.ID,
			CellX:         int(req.Payload.StartCoords.X.IntPart()),
			CellY:         int(req.Payload.StartCoords.Y.IntPart()),
			BattleID:      arena.CurrentBattle().ID,
			CreatedAt:     now,
			IsMoving:      true,
		}
		err = mmc.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to insert mech move command")
			return terror.Error(err, "Failed to trigger mech move command.")
		}

		// check mech command quest
		arena.Manager.QuestManager.MechCommanderQuestCheck(user.ID)

		// broadcast mech command log
		ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_command/%s", wm.FactionID, arena.ID, wm.Hash), server.HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
			MechMoveCommandLog: mmc,
		})
	} else {
		mmmc, err := arena._currentBattle.playerAbilityManager().GetMiniMechMove(wm.Hash)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mini mech move command")
			return terror.Error(err, "Failed to trigger mech move command.")
		}

		mmmc.Read(func(mmmc *player_abilities.MiniMechMoveCommand) {
			// broadcast mech command log
			ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_command/%s", factionID, arena.ID, wm.Hash), server.HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
				MechMoveCommandLog: &boiler.MechMoveCommandLog{
					ArenaID:       arena.ID,
					ID:            fmt.Sprintf("%s_%s", mmmc.BattleID, mmmc.MechHash),
					BattleID:      mmmc.BattleID,
					MechID:        mmmc.MechHash,
					TriggeredByID: mmmc.TriggeredByID,
					CellX:         mmmc.CellX,
					CellY:         mmmc.CellY,
					CancelledAt:   mmmc.CancelledAt,
					ReachedAt:     mmmc.ReachedAt,
					CreatedAt:     mmmc.CreatedAt,
					IsMoving:      mmmc.IsMoving,
				},
				IsMiniMech: true,
			})
		})
	}

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
		ArenaID string `json:"arena_id"`

		Hash          string `json:"hash"`
		MoveCommandID string `json:"move_command_id"`
	} `json:"payload"`
}

// MechMoveCommandCancelHandler send cancel mech move command to game client
func (am *ArenaManager) MechMoveCommandCancelHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	req := &MechMoveCommandCancelRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	arena, err := am.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	// check battle stage
	if arena.currentBattleState() == BattleStageEnd {
		return terror.Error(terror.ErrInvalidInput, "Current battle is ended.")
	}

	wm := arena.CurrentBattleWarMachineOrAIByHash(req.Payload.Hash)
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

	isMiniMech := wm.AIType != nil && *wm.AIType == MiniMech
	offeringID := uuid.Must(uuid.NewV4())
	if !isMiniMech {
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
		mmc.IsMoving = false
		_, err = mmc.Update(gamedb.StdConn, boil.Whitelist(boiler.MechMoveCommandLogColumns.CancelledAt, boiler.MechMoveCommandLogColumns.IsMoving))
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
			EventID:             offeringID,
		})

		ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_command/%s", factionID, arena.ID, wm.Hash), server.HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
			MechMoveCommandLog: mmc,
		})
	} else {
		mmmc, err := arena._currentBattle.playerAbilityManager().CancelMiniMechMove(wm.Hash)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Interface("warmachine", wm).Msg("Failed to cancel mini mech move command")
			return terror.Error(err, "Failed to cancel mini mech move command")
		}

		// send mech move command to game client
		arena.Message("BATTLE:ABILITY", &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: MechMoveCommandCancelGameAbilityID,
			WarMachineHash:      &wm.Hash,
			ParticipantID:       &wm.ParticipantID, // trigger on war machine
			EventID:             offeringID,
		})

		ws.PublishMessage(fmt.Sprintf("/faction/%s/arena/%s/mech_command/%s", factionID, arena.ID, wm.Hash), server.HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
			MechMoveCommandLog: &boiler.MechMoveCommandLog{
				ID:            fmt.Sprintf("%s_%s", mmmc.BattleID, mmmc.MechHash),
				BattleID:      mmmc.BattleID,
				MechID:        mmmc.MechHash,
				TriggeredByID: mmmc.TriggeredByID,
				CellX:         mmmc.CellX,
				CellY:         mmmc.CellY,
				CancelledAt:   mmmc.CancelledAt,
				ReachedAt:     mmmc.ReachedAt,
				CreatedAt:     mmmc.CreatedAt,
				IsMoving:      mmmc.IsMoving,
			},
			IsMiniMech: true,
		})
	}

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to broadcast faction mech commands")
	}

	reply(true)

	return nil
}

type AbilityOptInRequest struct {
	Payload struct {
		ArenaID string `json:"arena_id"`
	} `json:"payload"`
}

const HubKeyBattleAbilityOptIn = "BATTLE:ABILITY:OPT:IN"

var optInBucket = leakybucket.NewCollector(1, 1, true)

func (am *ArenaManager) BattleAbilityOptIn(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if optInBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many Requests")
	}

	req := &AbilityOptInRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	arena, err := am.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
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
		BattleID:                btl.ID,
		PlayerID:                user.ID,
		BattleAbilityOfferingID: offeringID,
		FactionID:               factionID,
		BattleAbilityID:         ba.ID,
	}
	err = bao.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to opt in battle ability")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/arena/%s/battle_ability/check_opt_in", user.ID, arena.ID), HubKeyBattleAbilityOptInCheck, true)

	return nil
}
