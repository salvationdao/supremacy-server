package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/xsyn_rpcclient"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// IncognitoManager tracks all war machines that are currently hidden from the map
type IncognitoManager struct {
	_incognitoWarMachineIDs map[string]struct{}

	sync.RWMutex
}

func NewIncognitoManager() *IncognitoManager {
	return &IncognitoManager{
		_incognitoWarMachineIDs: make(map[string]struct{}),
	}
}

func (iwmm *IncognitoManager) IsWarMachineHidden(hash string) bool {
	_, ok := iwmm._incognitoWarMachineIDs[hash]
	return ok
}

func (iwmm *IncognitoManager) AddHiddenWarMachineHash(hash string) error {
	iwmm.RLock()
	defer iwmm.RUnlock()

	_, ok := iwmm._incognitoWarMachineIDs[hash]
	if ok {
		return fmt.Errorf("War machine is already hidden")
	}
	iwmm._incognitoWarMachineIDs[hash] = struct{}{}

	return nil
}

func (iwmm *IncognitoManager) RemoveHiddenWarMachineHash(hash string) error {
	iwmm.RLock()
	defer iwmm.RUnlock()

	_, ok := iwmm._incognitoWarMachineIDs[hash]
	if !ok {
		return fmt.Errorf("Cannot unhide war machine that is not already hidden")
	}
	delete(iwmm._incognitoWarMachineIDs, hash)

	return nil
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

const IncognitoGameAbilityID = 15
const BlackoutGameAbilityID = 16

const HubKeyPlayerAbilityUse = "PLAYER:ABILITY:USE"

func (arena *Arena) PlayerAbilityUse(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// skip, if current not battle
	if arena.CurrentBattle() == nil {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Msg("no current battle")
		return terror.Error(terror.ErrForbidden, "There is no battle currently to use this ability on.")
	}

	req := &PlayerAbilityUseRequest{}
	err := json.Unmarshal(payload, req)
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

	if pa.OwnerID != player.ID {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Str("ability ownerID", pa.OwnerID).Str("blueprintAbilityID", req.Payload.BlueprintAbilityID).Msgf("player %s tried to execute an ability that wasn't theirs", player.ID)
		return terror.Error(terror.ErrForbidden, "You do not have permission to activate this ability.")
	}

	if !player.FactionID.Valid || player.FactionID.String == "" {
		gamelog.L.Warn().Str("func", "PlayerAbilityUse").Str("ability ownerID", pa.OwnerID).Str("blueprintAbilityID", req.Payload.BlueprintAbilityID).Msgf("player %s tried to execute an ability but they aren't part of a faction", player.ID)
		return terror.Error(terror.ErrForbidden, "You must be enrolled in a faction in order to use this ability.")
	}

	if pa.Count < 1 {
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("player ability count is 0, cannot be used")
		return terror.Error(err, "You do not have any more of this ability to use.")
	}

	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the PlayerAbilityUse!", r)
		}
	}()

	currentBattle := arena.CurrentBattle()
	// check battle end
	if currentBattle.stage.Load() == BattleStageEnd {
		gamelog.L.Warn().Str("func", "LocationSelect").Msg("battle stage has en ended")
		return nil
	}

	bpa := pa.R.Blueprint

	userID := uuid.FromStringOrNil(user.ID)
	var event *server.GameAbilityEvent
	switch req.Payload.LocationSelectType {
	case boiler.LocationSelectTypeEnumLINE_SELECT:
		if req.Payload.StartCoords == nil || req.Payload.EndCoords == nil {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("no start/end coords was provided for executing ability of type %s", boiler.LocationSelectTypeEnumLINE_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when executing this ability.")
		}
		if req.Payload.StartCoords.X < 0 || req.Payload.StartCoords.Y < 0 || req.Payload.EndCoords.X < 0 || req.Payload.EndCoords.Y < 0 {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("invalid start/end coords were provided for executing %s ability", boiler.LocationSelectTypeEnumLINE_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			GameLocation:        currentBattle.getGameWorldCoordinatesFromCellXY(req.Payload.StartCoords),
			GameLocationEnd:     currentBattle.getGameWorldCoordinatesFromCellXY(req.Payload.EndCoords),
		}

	case boiler.LocationSelectTypeEnumMECH_SELECT:
		if req.Payload.MechHash == "" {
			gamelog.L.Error().Interface("request payload", req.Payload).Err(err).Msgf("no mech hash was provided for executing ability of type %s", boiler.LocationSelectTypeEnumMECH_SELECT)
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
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("no start coords was provided for executing ability of type %s", boiler.LocationSelectTypeEnumLOCATION_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Coordinates must be provided when executing this ability.")
		}
		if req.Payload.StartCoords.X < 0 || req.Payload.StartCoords.Y < 0 {
			gamelog.L.Error().Interface("request payload", req.Payload).Msgf("invalid start coords were provided for executing %s ability", boiler.LocationSelectTypeEnumLOCATION_SELECT)
			return terror.Error(terror.ErrInvalidInput, "Invalid coordinates provided when executing this ability.")
		}
		event = &server.GameAbilityEvent{
			IsTriggered:         true,
			GameClientAbilityID: byte(bpa.GameClientAbilityID),
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &player.Username.String,
			EventID:             uuid.FromStringOrNil(pa.ID), // todo: change this?
			FactionID:           &player.FactionID.String,
			GameLocation:        currentBattle.getGameWorldCoordinatesFromCellXY(req.Payload.StartCoords),
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
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	defer tx.Rollback()

	// Create consumed_abilities entry
	ca := boiler.ConsumedAbility{
		BattleID:            currentBattle.BattleID,
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
		gamelog.L.Error().Err(err).Interface("consumedAbility", ca).Msg("failed to created consumed ability entry")
		return err
	}

	// Update the count of the player_abilities entry
	pa.Count = pa.Count - 1
	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to update player ability count")
		return err
	}

	isIncognito := bpa.GameClientAbilityID == IncognitoGameAbilityID
	// If player ability is "Incognito"
	if isIncognito {
		wm, err := boiler.CollectionItems(boiler.CollectionItemWhere.Hash.EQ(req.Payload.MechHash)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Interface("request payload", req.Payload).Err(err).Msgf("failed to execute INCOGNITO ability: could not get war machine from hash %s", req.Payload.MechHash)
			return terror.Error(err, "Failed to get war machine from hash")
		}

		im := arena.CurrentBattle().incognitoManager()
		err = im.AddHiddenWarMachineHash(wm.Hash)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to execute Incognito player ability")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue executing player ability, please try again or contact support.")
	}
	reply(true)

	if !isIncognito {
		// Tell gameclient to execute ability
		currentBattle.arena.Message("BATTLE:ABILITY", event)
	}

	pas, err := db.PlayerAbilitiesList(user.ID)
	if err != nil {
		gamelog.L.Error().Str("boiler func", "PlayerAbilities").Str("ownerID", user.ID).Err(err).Msg("unable to get player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}
	ws.PublishMessage(fmt.Sprintf("/user/%s/player_abilities", userID), server.HubKeyPlayerAbilitiesList, pas)

	if bpa.GameClientAbilityID == BlackoutGameAbilityID {
		ws.PublishMessage("/public/minimap", HubKeyMinimapUpdatesSubscribe, MinimapUpdatesSubscribeResponse{
			Duration: 3000,
			Radius:   int(BlackoutRadius),
			Coords:   *req.Payload.StartCoords,
		})
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
		gamelog.L.Error().Err(err).Msg("Failed to get mech command logs from db")
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
	// check environment
	if os.Getenv("GAMESERVER_ENVIRONMENT") == "production" {
		return nil
	}

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
		gamelog.L.Error().Str("mech id", wm.ID).Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to get mech move command.")
	}

	resp := &MechMoveCommandResponse{
		RemainCooldownSeconds: 0,
	}

	if mmc != nil {
		resp.MechMoveCommandLog = mmc
		resp.RemainCooldownSeconds = 30 - int(time.Now().Sub(mmc.CreatedAt).Seconds())
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

type MechMoveCommandCreateRequest struct {
	Payload struct {
		Hash        string               `json:"mech_hash"`
		StartCoords *server.CellLocation `json:"start_coords"`
	} `json:"payload"`
}

// MechMoveCommandCreateHandler send mech move command to game client
func (arena *Arena) MechMoveCommandCreateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// check environment
	if os.Getenv("GAMESERVER_ENVIRONMENT") == "production" {
		gamelog.L.Warn().Msg("Mech move command is not allowed in prod environment")
		return terror.Error(terror.ErrForbidden, "Mech move command is not allowed in prod environment")
	}

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

	// check mech move command is triggered within 30 seconds
	mmc, err := boiler.MechMoveCommandLogs(
		boiler.MechMoveCommandLogWhere.MechID.EQ(wm.ID),
		boiler.MechMoveCommandLogWhere.BattleID.EQ(arena.CurrentBattle().ID),
		boiler.MechMoveCommandLogWhere.CreatedAt.GT(time.Now().Add(-30*time.Second)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to get mech move command from db")
		return terror.Error(err, "Failed to trigger mech move command")
	}

	if mmc != nil {
		return terror.Error(terror.ErrInvalidInput, "Command is still cooling down.")
	}

	txid, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.FromStringOrNil(user.ID),
		ToUserID:             SupremacyBattleUserID,
		Amount:               decimal.New(10, 18).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("mech_move_command|%s|%d", wm.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		SubGroup:             arena.CurrentBattle().ID,
		Description:          "mech move command: " + wm.ID,
		NotSafe:              true,
	})
	if err != nil {
		return terror.Error(err, err.Error())
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
		gamelog.L.Error().Str("mech id", wm.ID).Str("battle id", arena.CurrentBattle().ID).Err(err).Msg("Failed to cancel unfinished mech move command in db")
		return terror.Error(err, "Failed to update mech move command.")
	}

	// log mech move command
	mmc = &boiler.MechMoveCommandLog{
		MechID:        wm.ID,
		TriggeredByID: user.ID,
		CellX:         req.Payload.StartCoords.X,
		CellY:         req.Payload.StartCoords.Y,
		BattleID:      arena.CurrentBattle().ID,
		TXID:          txid,
		CreatedAt:     now,
	}

	err = mmc.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert mech move command")
		return terror.Error(err, "Failed to trigger mech move command.")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/mech_command/%s", factionID, wm.Hash), HubKeyMechMoveCommandSubscribe, &MechMoveCommandResponse{
		MechMoveCommandLog:    mmc,
		RemainCooldownSeconds: 30,
	})

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to broadcast faction mech commands")
	}

	arena.BroadcastMechCommandNotification(&MechCommandNotification{
		MechID:       wm.ID,
		MechLabel:    wm.Name,
		MechImageUrl: wm.ImageAvatar,
		FactionID:    wm.FactionID,
		Action:       MechCommandActionFired,
		FiredByUser: &UserBrief{
			ID:        uuid.FromStringOrNil(user.ID),
			Username:  user.Username.String,
			FactionID: user.FactionID.String,
			Gid:       user.Gid,
		},
	})

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
	// check environment
	if os.Getenv("GAMESERVER_ENVIRONMENT") == "production" {
		gamelog.L.Warn().Msg("Mech move command is not allowed in prod environment")
		return terror.Error(terror.ErrForbidden, "Mech move command is not allowed in prod environment")
	}

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
		gamelog.L.Error().Str("mech move command id", req.Payload.MoveCommandID).Err(err).Msg("Failed to get mech move command from db")
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
		gamelog.L.Error().Err(err).Str("mech move command id", mmc.ID).Msg("Failed to up date mech move command in db")
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
		RemainCooldownSeconds: 30 - int(time.Now().Sub(mmc.CreatedAt).Seconds()),
	})

	err = arena.BroadcastFactionMechCommands(factionID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to broadcast faction mech commands")
	}

	arena.BroadcastMechCommandNotification(&MechCommandNotification{
		MechID:       wm.ID,
		MechLabel:    wm.Name,
		MechImageUrl: wm.ImageAvatar,
		FactionID:    wm.FactionID,
		Action:       MechCommandActionCancel,
		FiredByUser: &UserBrief{
			ID:        uuid.FromStringOrNil(user.ID),
			Username:  user.Username.String,
			FactionID: user.FactionID.String,
			Gid:       user.Gid,
		},
	})

	reply(true)

	return nil
}
