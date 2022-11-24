package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

func NewFactionPassController(api *API) {
	api.SecureUserFactionCommand(HubKeyFactionPassSupsPurchase, api.FactionPassSupsPurchase)
}

type FactionPassPurchaseSupsRequest struct {
	Payload struct {
		FactionPassID string `json:"faction_pass_id"`
		PaymentType   string `json:"payment_type"`
	} `json:"payload"`
}

const HubKeyFactionPassSupsPurchase = "FACTION:PASS:SUPS:PURCHASE"

func (api *API) FactionPassSupsPurchase(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &FactionPassPurchaseSupsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	l := gamelog.L.With().Str("func", "FactionPassSupsPurchase").Str("faction pass id", req.Payload.FactionPassID).Logger()

	// load faction pass
	fp, err := boiler.FindFactionPass(gamedb.StdConn, req.Payload.FactionPassID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("Failed to load faction pass from db")
		return terror.Error(err, "Failed to load faction pass")
	}

	if fp == nil {
		return terror.Error(fmt.Errorf("faction pass does not exist"), "Faction pass does not exist.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		l.Error().Err(err).Msg("Failed to start db transaction.")
		return terror.Error(err, "Failed to purchase faction pass.")
	}

	defer tx.Rollback()

	err = user.Reload(tx)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load user.")
		return terror.Error(err, "Failed to load user.")
	}

	// update player's faction pass
	startTime := time.Now()
	if user.FactionPassExpiresAt.Valid && user.FactionPassExpiresAt.Time.After(startTime) {
		// set start time to expiry date, if player's faction pass hasn't expired yet
		startTime = user.FactionPassExpiresAt.Time
	}

	user.FactionPassExpiresAt = null.TimeFrom(startTime.Add(time.Duration(fp.LastForDays*24) * time.Hour))
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.FactionPassExpiresAt))
	if err != nil {
		l.Error().Err(err).Interface("player", user).Msg("Failed to update the expiry date of faction pass.")
		return terror.Error(err, "Failed to update the expiry date of faction pass.")
	}

	refund := func() {
		return
	}

	switch req.Payload.PaymentType {
	case boiler.PaymentMethodsSups:
		amount := fp.SupsCost.Mul(decimal.NewFromInt(100).Sub(fp.SupsDiscountPercentage).Div(decimal.NewFromInt(100)))

		// refund reward
		paidTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.FromStringOrNil(user.ID),
			ToUserID:             uuid.UUID(server.XsynTreasuryUserID),
			Amount:               amount.String(),
			TransactionReference: server.TransactionReference(fmt.Sprintf("purchase_faction_pass|%s|%d", fp.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupFactionPass),
			Description:          fmt.Sprintf("purchase a '%s' faction.", fp.Label),
		})

		refund = func() {
			_, err = api.Passport.RefundSupsMessage(paidTXID)
			if err != nil {
				l.Error().Err(err).Msg("Failed to refund purchase faction pass.")
			}
		}

		// record faction pass log
		fpl := boiler.FactionPassPurchaseLog{
			FactionPassID:  fp.ID,
			PurchasedByID:  user.ID,
			PurchaseMethod: boiler.PaymentMethodsSups,
			Price:          fp.SupsCost,
			Discount:       fp.SupsDiscountPercentage,
			PurchaseTXID:   null.StringFrom(paidTXID),
		}

		err = fpl.Insert(tx, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("Failed to record faction pass log")
			return terror.Error(err, "Failed to purchase faction pass.")
		}

	case boiler.PaymentMethodsStripe:
	case boiler.PaymentMethodsEth:
	default:
		return terror.Error(fmt.Errorf("payment type does not exist"), "Payment type does not exist")
	}

	err = tx.Commit()
	if err != nil {
		refund()
		l.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to purchase faction pass.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/faction_pass_expiry_date", user.ID), HubKeyPlayerFactionPassExpiryDate, user.FactionPassExpiresAt)

	reply(true)
	return nil
}

const HubKeyFactionPassList = "FACTION:PASS:LIST"

func (api *API) FactionPassList(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	fps, err := boiler.FactionPasses(
		qm.OrderBy(boiler.FactionPassColumns.LastForDays),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load faction passes")
	}
	reply(fps)
	return nil
}

// DASHBOARD

func (api *API) FactionMVPStakedMech(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "FactionMVPStakedMech").Logger()

	queries := []qm.QueryMod{
		qm.Select(boiler.StakedMechBattleLogTableColumns.StakedMechID),
		qm.From(boiler.TableNames.StakedMechBattleLogs),
		qm.GroupBy(boiler.StakedMechBattleLogTableColumns.StakedMechID),
		qm.OrderBy(fmt.Sprintf("COUNT(%s) DESC", boiler.StakedMechBattleLogTableColumns.ID)),
		qm.Limit(1),
	}

	mechID := ""
	err := boiler.NewQuery(queries...).QueryRow(gamedb.StdConn).Scan(&mechID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("Failed to load faction MVP staked mech")
		return terror.Error(err, "Failed to load faction MVP staked mech.")
	}

	reply(mechID)
	return nil
}

func (api *API) FactionStakeMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	total, err := boiler.StakedMechs(
		boiler.StakedMechWhere.FactionID.EQ(factionID),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load staked mech count.")
		return terror.Error(err, "Failed to load staked mech count.")
	}

	reply(total)
	return nil
}

func (api *API) FactionQueuedStakedMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	total, err := boiler.BattleLobbiesMechs(
		qm.Where(fmt.Sprintf(
			"EXISTS ( SELECT 1 FROM %s WHERE %s = %s )",
			boiler.TableNames.StakedMechs,
			boiler.StakedMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.MechID,
		)),
		boiler.BattleLobbiesMechWhere.FactionID.EQ(factionID),
		boiler.BattleLobbiesMechWhere.LockedAt.IsNull(),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load total staked mech in queue.")
		return terror.Error(err, "Failed to load total staked mech in queue")
	}

	reply(total)
	return nil
}

func (api *API) FactionDamagedStakedMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	total, err := boiler.RepairCases(
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
		qm.Where(fmt.Sprintf(
			"EXISTS ( SELECT 1 FROM %s WHERE %s = %s )",
			boiler.TableNames.StakedMechs,
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.MechID,
		)),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load the count of damaged staked mech.")
		return terror.Error(err, "Failed to load the count of damaged staked mech.")
	}

	reply(total)
	return nil
}

func (api *API) FactionBattleReadyStakedMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	total, err := boiler.BattleLobbiesMechs(
		qm.Where(fmt.Sprintf(
			"EXISTS ( SELECT 1 FROM %s WHERE %s = %s )",
			boiler.TableNames.StakedMechs,
			boiler.StakedMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.MechID,
		)),
		boiler.BattleLobbiesMechWhere.FactionID.EQ(factionID),
		boiler.BattleLobbiesMechWhere.LockedAt.IsNotNull(),
		boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load total staked mech in queue.")
		return terror.Error(err, "Failed to load total staked mech in queue")
	}

	reply(total)
	return nil
}

func (api *API) FactionInBattleStakedMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	total, err := boiler.BattleLobbiesMechs(
		qm.Where(fmt.Sprintf(
			"EXISTS ( SELECT 1 FROM %s WHERE %s = %s )",
			boiler.TableNames.StakedMechs,
			boiler.StakedMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.MechID,
		)),
		boiler.BattleLobbiesMechWhere.FactionID.EQ(factionID),
		boiler.BattleLobbiesMechWhere.LockedAt.IsNotNull(),
		boiler.BattleLobbiesMechWhere.AssignedToBattleID.IsNotNull(),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load total staked mech in queue.")
		return terror.Error(err, "Failed to load total staked mech in queue")
	}

	reply(total)
	return nil
}

func (api *API) FactionBattledStakedMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	total, err := boiler.StakedMechBattleLogs(
		boiler.StakedMechBattleLogWhere.FactionID.EQ(factionID),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load total battled staked mech count.")
		return terror.Error(err, "Failed to load battled staked mech count.")
	}

	reply(total)
	return nil
}
