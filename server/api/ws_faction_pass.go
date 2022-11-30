package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

func (api *API) exchangeRatesUpdater() {
	// update the exchange rate before setting up the loop
	api.exchangeRateUpdate()

	interval := 1 * time.Minute
	timer := time.NewTimer(interval)

	for {
		<-timer.C

		// update the exchange rate before setting up the loop
		api.exchangeRateUpdate()

		// NOTE: reset timer everytime the process is finish to avoid overlap
		timer.Reset(interval)
	}
}

func (api *API) exchangeRateUpdate() {
	// fetch exchange rates
	exchangeRates, err := api.Passport.GetCurrentRates()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load exchange rate.")
		return
	}

	// NOTE: put any exchange related functions below...

	// update faction passes
	go factionPassPriceUpdate(exchangeRates)

}

// factionPassPriceUpdate update the price of the faction pass
func factionPassPriceUpdate(exchangeRates *xsyn_rpcclient.GetExchangeRatesResp) {
	l := gamelog.L.With().
		Str("func", "updateFactionPassPrice").
		Str("sup_to_usd_rate", exchangeRates.SUPtoUSD.String()).
		Str("eth_to_usd_rate", exchangeRates.ETHtoUSD.String()).
		Str("bnb_to_usd_rate", exchangeRates.BNBtoUSD.String()).
		Logger()

	factionPasses, err := boiler.FactionPasses().All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load faction pass list.")
		return
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		l.Error().Err(err).Msg("Failed to start db transaction.")
		return
	}

	defer tx.Rollback()

	for _, factionPass := range factionPasses {
		factionPass.UsdPrice = factionPass.EthPriceWei.Div(decimal.New(1, 18)).Mul(exchangeRates.ETHtoUSD).Round(2)
		factionPass.SupsPrice = factionPass.UsdPrice.Div(exchangeRates.SUPtoUSD).Mul(decimal.New(1, 18)).Round(0)

		_, err = factionPass.Update(tx, boil.Whitelist(boiler.FactionPassColumns.UsdPrice, boiler.FactionPassColumns.SupsPrice))
		if err != nil {
			l.Error().Err(err).Interface("faction pass", factionPass).Msg("Failed to update faction pass")
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		l.Error().Err(err).Msg("Failed to commit db transaction.")
		return
	}

	ws.PublishMessage("/secure/faction_pass_list", HubKeyFactionPassList, factionPasses)

}

func NewFactionPassController(api *API) {
	api.SecureUserFactionCommand(HubKeyFactionPassSupsPurchase, api.FactionPassSupsPurchase)
	api.SecureUserFactionCommand(HubKeyFactionPassStripePaymentIntent, api.FactionPassStripePaymentIntent)
	api.SecureUserFactionCommand(HubKeyFactionPassStripePaymentClaim, api.FactionPassPaymentClaim)
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
		supsCost := fp.SupsPrice
		actualPrice := supsCost.Mul(decimal.NewFromInt(100).Sub(fp.DiscountPercentage).Div(decimal.NewFromInt(100)))

		// refund reward
		paidTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.FromStringOrNil(user.ID),
			ToUserID:             uuid.UUID(server.XsynTreasuryUserID),
			Amount:               actualPrice.String(),
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
			FactionPassID:         fp.ID,
			PurchasedByID:         user.ID,
			PurchaseMethod:        boiler.PaymentMethodsSups,
			ExpendFactionPassDays: fp.LastForDays,
			SupsPaid:              actualPrice,
			SupsPurchaseTXID:      null.StringFrom(paidTXID),
			PaymentStatus:         PaymentStatusSuccess,
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

const PaymentStatusSuccess = "SUCCESS"
const PaymentStatusFail = "FAIL"
const PaymentStatusPending = "PENDING"

type FactionPassPaymentClaimRequest struct {
	Payload struct {
		PaymentIntentID string `json:"payment_intent_id"`
	} `json:"payload"`
}

const HubKeyFactionPassStripePaymentClaim = "FACTION:PASS:STRIPE:PAYMENT:CLAIM"

func (api *API) FactionPassPaymentClaim(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &FactionPassPaymentClaimRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	l := gamelog.L.With().Str("func", "FactionPassPaymentClaim").Str("faction pass payment intent id", req.Payload.PaymentIntentID).Logger()

	// generate payment intent
	pi, err := paymentintent.Get(req.Payload.PaymentIntentID, nil)
	if err != nil {
		l.Error().Err(err).Msg("could not get payment intent from payment intent ID")
		return terror.Error(err, "Failed to generate claim id.")
	}

	factionPassID := pi.Metadata["faction_pass_id"]
	playerID := pi.Metadata["player_id"]

	if playerID != user.ID {
		return terror.Error(fmt.Errorf("user id not the same"), "The player id does not match.")
	}

	// load faction pass
	fp, err := boiler.FindFactionPass(gamedb.StdConn, factionPassID)
	if err != nil {
		return terror.Error(err, "Failed to load faction pass.")
	}

	fpp := boiler.FactionPassPurchaseLog{
		FactionPassID:         fp.ID,
		PurchasedByID:         user.ID,
		PurchaseMethod:        boiler.PaymentMethodsStripe,
		UsdPaid:               decimal.NewFromInt(pi.Amount).Div(decimal.NewFromInt(100)),
		StripePaymentIntentID: null.StringFrom(req.Payload.PaymentIntentID),
		PaymentStatus:         PaymentStatusPending,
		ExpendFactionPassDays: fp.LastForDays,
	}

	err = fpp.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert faction pass purchase log.")
	}

	reply(fpp)
	return nil
}

const HubKeyFactionPassStripePaymentIntent = "FACTION:PASS:STRIPE:PAYMENT:INTENT"

func (api *API) FactionPassStripePaymentIntent(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	factionPassID := chi.RouteContext(ctx).URLParam("faction_pass_id")
	if factionPassID == "" {
		return fmt.Errorf("faction pass id is required")
	}

	l := gamelog.L.With().Str("func", "FactionPassStripePaymentIntent").Str("faction pass id", factionPassID).Logger()

	// load faction pass
	fp, err := boiler.FindFactionPass(gamedb.StdConn, factionPassID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("Failed to load faction pass from db")
		return terror.Error(err, "Failed to load faction pass")
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(fp.UsdPrice.Mul(decimal.NewFromInt(100)).IntPart()),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Description: stripe.String(fmt.Sprintf("Purchase of %s faction pass", fp.Label)),
	}

	params.AddMetadata("sale_type", "faction pass")
	params.AddMetadata("faction_pass_id", fp.ID)
	params.AddMetadata("player_id", user.ID)

	pi, err := paymentintent.New(params)
	if err != nil {
		l.Error().Err(err).Msg("unable to generate stripe payment intent")
		return terror.Error(err, "Unable to generate payment.")
	}

	reply(struct {
		ClientSecret    string `json:"client_secret"`
		PaymentIntentID string `json:"payment_intent_id"`
	}{
		ClientSecret:    pi.ClientSecret,
		PaymentIntentID: pi.ID,
	})

	return nil
}

// SUBSCRIPTION

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

// FACTION STAKED MECH DASHBOARD

func (api *API) FactionMostPopularStakedMech(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "FactionMostPopularStakedMech").Logger()

	queries := []qm.QueryMod{
		qm.Select(boiler.StakedMechBattleLogTableColumns.StakedMechID),
		qm.From(boiler.TableNames.StakedMechBattleLogs),
		boiler.StakedMechBattleLogWhere.FactionID.EQ(factionID),
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

	if mechID == "" {
		reply(nil)
		return nil
	}

	mechs, err := db.LobbyMechsBrief("", mechID)
	if err != nil || len(mechs) == 0 {
		return terror.Error(err, "Failed to load most popular staked mech.")
	}

	reply(mechs[0])
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
	total, err := boiler.StakedMechs(
		boiler.StakedMechWhere.FactionID.EQ(factionID),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.RepairCases,
			boiler.RepairCaseTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.CompletedAt,
			boiler.RepairCaseTableColumns.PausedAt,
			boiler.RepairCaseTableColumns.DeletedAt,
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
		boiler.BattleLobbiesMechWhere.AssignedToBattleID.IsNull(),
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

func (api *API) FactionInRepairBayStakedMechCount(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	queries := []qm.QueryMod{
		qm.Select(
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.BlocksRequiredRepair,
			boiler.RepairCaseTableColumns.BlocksRepaired,
		),

		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s = '%s') %s",
			boiler.TableNames.StakedMechs,
			boiler.StakedMechTableColumns.FactionID,
			factionID,
			boiler.TableNames.StakedMechs,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s != '%s'",
			boiler.TableNames.PlayerMechRepairSlots,
			boiler.PlayerMechRepairSlotTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.PlayerMechRepairSlotTableColumns.Status,
			boiler.RepairSlotStatusDONE,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.RepairCases,
			boiler.RepairCaseTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.CompletedAt,
			boiler.RepairCaseTableColumns.DeletedAt,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to repair bay mechs from db.")
		return terror.Error(err, "Failed to load mech detail from db.")
	}

	resp := &server.FactionStakedMechRepairBayResponse{}
	for rows.Next() {
		mechID := ""
		requiredRepairedBlocks := 0
		repairedBlocks := 0

		err = rows.Scan(&mechID, &requiredRepairedBlocks, &repairedBlocks)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan mech repair detail.")
			return terror.Error(err, "Failed to scan mech repair detail.")
		}

		resp.MechCount += 1
		resp.TotalRequiredRepairedBlocks += requiredRepairedBlocks
		resp.TotalRepairedBlocks += repairedBlocks
	}

	reply(resp)
	return nil
}
