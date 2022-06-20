package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

const HubKeyMechMoveCommand = "MECH:MOVE:COMMAND"

type MechMovementRequest struct {
	Payload struct {
		ParticipantID int `json:"participant_id"`
		X             int `json:"x"`
		Y             int `json:"y"`
	} `json:"payload"`
}

var mechCommandBucket = leakybucket.NewCollector(0.0333, 1, true)

func (arena *Arena) MechMoveCommandHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MechMovementRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check ownership
	wm := arena.CurrentBattleWarMachine(req.Payload.ParticipantID)
	if wm == nil {
		return terror.Error(fmt.Errorf("required mech not found"), "Targeted mech is not on the battle list.")
	}

	if wm.OwnedByID != user.ID {
		gamelog.L.Warn().Str("mech owner id", wm.OwnedByID).Str("current user id", user.ID).Msg("Unauthorised mech moving reques.t")
		return terror.Error(terror.ErrForbidden, "The mech is not owned by current user.")
	}

	// check rate limit
	b := mechCommandBucket.Add(wm.ID, 1)
	if b == 0 {
		return terror.Error(err, "Mech move command can only be triggered once every 30 seconds.")
	}

	// TODO: pay sups

	// TODO: send position to game client, and check command string
	arena.Message("MECH_MOVE_COMMAND", &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: 8,
		ParticipantID:       &wm.ParticipantID, // trigger on war machine
		TriggeredOnCellX:    &req.Payload.X,
		TriggeredOnCellY:    &req.Payload.Y,
	})

	// log mech move command
	mvl := &boiler.MechMoveCommandLog{
		MechID:        wm.ID,
		TriggeredByID: user.ID,
		X:             req.Payload.X,
		Y:             req.Payload.Y,
	}

	err = mvl.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert mech move command")
		return terror.Error(err, "Failed to trigger mech move command.")
	}

	return nil
}
