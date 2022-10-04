package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"
	"time"

	"github.com/shopspring/decimal"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type StoreController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewStoreController(api *API) *StoreController {
	sc := &StoreController{
		API: api,
	}

	api.SecureUserFactionCommand(HubKeyGetMysteryCrates, sc.GetMysteryCratesHandler)
	api.SecureUserFactionCommand(HubKeyMysteryCratePurchase, sc.PurchaseMysteryCrateHandler)

	return sc
}

type MysteryCrateOwnershipResp struct {
	Owned   int `json:"owned"`
	Allowed int `json:"allowed"`
}

const HubKeyGetMysteryCrates = "STORE:MYSTERY:CRATES"

func (sc *StoreController) GetMysteryCratesHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	crates, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
		qm.Load(qm.Rels(boiler.StorefrontMysteryCrateRels.FiatProduct, boiler.FiatProductRels.FiatProductPricings)),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get mystery crate")
	}

	resp := server.StoreFrontMysteryCrateSliceFromBoiler(crates)
	reply(resp)

	return nil
}

func (sc *StoreController) MysteryCrateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	crateID := cctx.URLParam("crate_id")

	crate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.ID.EQ(crateID),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
		qm.Load(qm.Rels(boiler.StorefrontMysteryCrateRels.FiatProduct, boiler.FiatProductRels.FiatProductPricings)),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get mystery crate")
	}

	resp := server.StoreFrontMysteryCrateFromBoiler(crate)
	reply(resp)

	return nil
}

type MysteryCratePurchaseRequest struct {
	Payload struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	} `json:"payload"`
}

const HubKeyMysteryCratePurchase = "STORE:MYSTERY:CRATE:PURCHASE"

func (sc *StoreController) PurchaseMysteryCrateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MysteryCratePurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	//checks
	if user.FactionID != null.StringFrom(factionID) {
		return terror.Error(terror.ErrForbidden, "User must be enlisted in correct faction to purchase faction crate.")
	}

	//double check there are still crates available on storefront, user should not be able to buy it though
	storeCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(req.Payload.Type),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
		qm.Load(boiler.StorefrontMysteryCrateRels.Faction),
		qm.Load(qm.Rels(boiler.StorefrontMysteryCrateRels.FiatProduct, boiler.FiatProductRels.FiatProductPricings)),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get crate for purchase, please try again or contact support.")
	}

	if (storeCrate.AmountSold + req.Payload.Quantity) >= storeCrate.Amount {
		return terror.Error(fmt.Errorf("player ID: %s, attempted to purchase sold out mystery crate", user.ID), "This mystery crate is sold out!")
	}

	//check user SUPS is more than crate.price
	supPrice := decimal.Zero
	for _, s := range storeCrate.R.FiatProduct.R.FiatProductPricings {
		if s.CurrencyCode == server.FiatCurrencyCodeSUPS {
			supPrice = s.Amount
			break
		}
	}
	if supPrice.LessThanOrEqual(decimal.Zero) {
		return terror.Error(fmt.Errorf("unable to find correct pricing for crate"), "Failed to get crate for purchase, please try again or contact support.")
	}

	// -------------------------------------
	supTransactionID, err := sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		Amount:               supPrice.Mul(decimal.NewFromInt(int64(req.Payload.Quantity))).String(),
		FromUserID:           uuid.FromStringOrNil(user.ID),
		ToUserID:             uuid.FromStringOrNil(server.SupremacyGameUserID),
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_mystery_crate_purchase|%s|%d", storeCrate.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             "Mystery Crate",
		Description:          fmt.Sprintf("Purchased mystery crate id %s", storeCrate.ID),
	})
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
		if err == nil {
			err = fmt.Errorf("transaction failed")
		}
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Str("mystery_crate_id", storeCrate.ID).Err(err).Msg("unable to charge user for mystery crate purchase")
		return terror.Error(err, "Unable to process mystery crate purchase,  check your balance and try again.")
	}

	refundFunc := func() {
		refundSupTransactionID, err := sc.API.Passport.RefundSupsMessage(supTransactionID)
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to refund user for mystery crate purchase cost")
		}

		txItem := &boiler.StorePurchaseHistory{
			PlayerID:    user.ID,
			Amount:      supPrice.Mul(decimal.NewFromInt(int64(req.Payload.Quantity))),
			ItemType:    "mystery_crate",
			ItemID:      storeCrate.ID,
			Description: "refunding mystery crate due to failed transaction",
			TXID:        supTransactionID,
			RefundTXID:  null.StringFrom(refundSupTransactionID),
			RefundedAt:  null.TimeFrom(time.Now()),
		}

		err = txItem.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to insert collectionItem into purchase history table.")
		}
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing mystery crate, please try again or contact support.")
	}
	defer tx.Rollback()

	// Assign multiple crate purchases
	var resp []Reward
	var xsynAssets []*rpctypes.XsynAsset
	for i := 0; i < req.Payload.Quantity; i++ {
		assignedCrate, xsynAsset, err := assignAndRegisterPurchasedCrate(user.ID, storeCrate, tx, sc.API)
		if err != nil {
			refundFunc()
			return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
		}

		txItem := &boiler.StorePurchaseHistory{
			PlayerID:    user.ID,
			Amount:      supPrice,
			ItemType:    "mystery_crate",
			ItemID:      assignedCrate.ID,
			Description: "purchased mystery crate",
			TXID:        supTransactionID,
		}

		err = txItem.Insert(tx, boil.Infer())
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
			return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
		}

		resp = append(resp, Reward{
			Crate:       assignedCrate,
			Label:       storeCrate.MysteryCrateType,
			ImageURL:    storeCrate.ImageURL,
			LockedUntil: null.TimeFrom(assignedCrate.LockedUntil),
		})

		xsynAssets = append(xsynAssets, xsynAsset)

	}
	serverStoreCrate := server.StoreFrontMysteryCrateFromBoiler(storeCrate)

	err = sc.API.Passport.AssetRegister(xsynAssets...)
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Interface("mystery crate", "").Msg("failed to register to XSYN")
		return terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("failed to commit mystery crate transaction")
		return terror.Error(err, "Issue purchasing mystery crate, please try again or contact support.")
	}

	//update mysterycrate subscribers and update player
	ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", factionID, storeCrate.ID), server.HubKeyMysteryCrateSubscribe, serverStoreCrate)

	reply(resp)
	return nil
}

func assignAndRegisterPurchasedCrate(userID string, storeCrate *boiler.StorefrontMysteryCrate, tx *sql.Tx, api *API) (*server.MysteryCrate, *rpctypes.XsynAsset, error) {
	availableCrates, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.FactionID.EQ(storeCrate.FactionID),
		boiler.MysteryCrateWhere.Type.EQ(storeCrate.MysteryCrateType),
		boiler.MysteryCrateWhere.Purchased.EQ(false),
		boiler.MysteryCrateWhere.Opened.EQ(false),
		qm.Load(boiler.MysteryCrateRels.Blueprint),
	).All(tx)
	if err != nil {
		return nil, nil, terror.Error(err, "Failed to get available crates, please try again or contact support.")
	}

	faction, err := boiler.FindFaction(tx, storeCrate.FactionID)
	if err != nil {
		return nil, nil, terror.Error(err, "Failed to find faction, please try again or contact support.")
	}

	//randomly assigning crate to user
	rand.Seed(time.Now().UnixNano())
	assignedCrate := availableCrates[rand.Intn(len(availableCrates))]

	//update purchased value
	assignedCrate.Purchased = true

	// set newly bought crates openable on staging/dev (this is so people cannot open already purchased crates and see what is in them)
	if server.IsDevelopmentEnv() || server.IsStagingEnv() {
		assignedCrate.LockedUntil = time.Now()
	}

	_, err = assignedCrate.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to update assigned crate information")
		return nil, nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	collectionItem, err := db.InsertNewCollectionItem(tx,
		"supremacy-general",
		boiler.ItemTypeMysteryCrate,
		assignedCrate.ID,
		"",
		userID,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to insert into collection items")
		return nil, nil, terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}
	storeCrate.AmountSold = storeCrate.AmountSold + 1
	_, err = storeCrate.Update(tx, boil.Whitelist(boiler.StorefrontMysteryCrateColumns.AmountSold))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
		return nil, nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	//register
	assignedCrateServer := server.MysteryCrateFromBoiler(assignedCrate, collectionItem, null.String{})
	xsynAsset := rpctypes.ServerMysteryCrateToXsynAsset(assignedCrateServer, faction.Label)

	return assignedCrateServer, xsynAsset, nil
}
