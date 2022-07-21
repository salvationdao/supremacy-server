package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/battle"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func NewWSWarMachineController(api *API) {
	api.SecureUserFactionCommand(HubKeyGetMysteryCrates, api.WarMachineRepair)

}

type WarMachineRepairRequest struct {
	Payload struct {
		MechID     string `json:"mech_id"`
		RepairType string `json:"repair_type"`
	} `json:"payload"`
}

const HubKeyWarMachineRepair = "WAR:MACHINE:REPAIR"

func (api *API) WarMachineRepair(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &WarMachineRepairRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	mech, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.MechID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("item type", boiler.ItemTypeMech).Str("mech id", req.Payload.MechID).Msg("Failed to query war machine collection item")
		return terror.Error(err, "Failed to load war machine detail.")
	}

	if mech.OwnerID != user.ID {
		return terror.Error(fmt.Errorf("do not own the mech"), "The mech is not owned by you.")
	}

	switch req.Payload.RepairType {
	case boiler.MechRepairLogTypeSTART_STANDARD_REPAIR:
		err = api.BattleArena.RepairSystem.StartStandardRepair(user.ID, mech.ID)
		if err != nil {
			return err
		}
	case boiler.MechRepairLogTypeSTART_FAST_REPAIR:
		err = api.BattleArena.RepairSystem.StartFastRepair(user.ID, mech.ID)
		if err != nil {
			return err
		}
	default:
		return terror.Error(fmt.Errorf("invalid repair type"), fmt.Sprintf("Repair type '%s' does not exist.", req.Payload.RepairType))
	}

	reply(true)
	return nil
}

// subscriptions

func (api *API) WarMachineRepairStatusSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	mechID := cctx.URLParam("mech_id")
	if mechID == "" {
		return terror.Error(fmt.Errorf("id not provided"), "Mech id is not provided")
	}

	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
		qm.Load(boiler.CollectionItemRels.Owner),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("item id (mech)", mechID).Err(err).Msg("Failed to load collection item.")
		return terror.Error(err, "Failed to load mech.")
	}

	if ci.R == nil || ci.R.Owner == nil || ci.R.Owner.FactionID.String != factionID {
		return terror.Error(fmt.Errorf("mech is not in your faction"), "Mech is not in your.")
	}

	mrc, err := boiler.MechRepairCases(
		boiler.MechRepairCaseWhere.MechID.EQ(mechID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to query mech repair case from db.")
		return terror.Error(err, "Failed to load war machine repair status.")
	}

	// send nil if mech is not repairing
	if mrc == nil || mrc.EndedAt.Valid {
		reply(nil)
		return nil
	}

	// build mech repair status
	mrs := battle.MechRepairStatus{
		RepairStatus: mrc.Status,
	}

	if mrc.ExpectedEndAt.Valid && mrc.StartedAt.Valid {
		// add a second delay
		mrs.RemainSeconds = null.IntFrom(1 + int(mrc.ExpectedEndAt.Time.Sub(mrc.StartedAt.Time).Seconds()))
	}

	reply(mrs)

	return nil
}
