package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"math/rand"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

type StoreController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

type MysteryCrateSubscribeRequest struct {
	*hub.HubCommandRequest
}

func NewStoreController(api *API) *StoreController {
	sc := &StoreController{
		API: api,
	}

	api.SecureUserFactionCommand(HubkeyMysteryCrateSubscribe, sc.MysteryCrateSubscribeHandler)
	api.SecureUserFactionCommand(HubkeyMysteryCratePurchase, sc.PurchaseMysteryCrateHandler)

	return sc
}

const HubkeyMysteryCrateSubscribe = "STORE:MYSTERY:CRATE:SUBSCRIBE"

func (sc *StoreController) MysteryCrateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MysteryCrateSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if user == nil {
		return terror.Error(terror.ErrForbidden, "User must be logged in to view crates.")
	}

	if factionID == "" {
		return terror.Error(terror.ErrForbidden, "User must be enlisted in a faction to view crates.")
	}

	crates, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get mystery crate")
	}

	reply(crates)

	return nil
}

type MysteryCratePurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Type string `json:"type"`
	} `json:"payload"`
}

const HubkeyMysteryCratePurchase = "STORE:MYSTERY:CRATE:PURCHASE"

func (sc *StoreController) PurchaseMysteryCrateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MysteryCratePurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	//checks
	if user == nil {
		return terror.Error(terror.ErrForbidden, "User must be logged in to purchase crates.")
	}
	if factionID == "" {
		return terror.Error(terror.ErrForbidden, "User must be enlisted in a faction to purchase crates.")
	}
	if user.FactionID != null.StringFrom(factionID) {
		return terror.Error(terror.ErrForbidden, "User must be enlisted in correct faction to purchase faction crate.")
	}
	//double check there are still crates available on storefront, user should not be able to buy it though
	storeCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(req.Payload.Type),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get crate for purchase, please try again or contact support.")
	}

	if storeCrate.AmountSold >= storeCrate.Amount {
		return terror.Error(fmt.Errorf("player ID: %s, attempted to purchase sold out mystery crate", user.ID), "This mystery crate is sold out!")
	}
	//check user SUPS is more than crate.price

	//get random crate where faction id == user.faction_id and purchased == false and opened == false and type == req.payload.type
	availableCrates, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.FactionID.EQ(factionID),
		boiler.MysteryCrateWhere.Type.EQ(req.Payload.Type),
		boiler.MysteryCrateWhere.Opened.EQ(false),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get available crates, please try again or contact support.")
	}

	//randomly assigning crate to user
	rand.Seed(time.Now().UnixNano())
	assignedCrate := availableCrates[rand.Intn(len(availableCrates))]

	// -------------------------------------
	supTransactionID, err := sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		Amount:               storeCrate.Price.String(),
		FromUserID:           uuid.FromStringOrNil(user.ID),
		ToUserID:             battle.SupremacyUserID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_mystery_crate_purchase|%s|%d", assignedCrate.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             "Mystery Crate",
		Description:          fmt.Sprintf("Purchased mystery crate %s", assignedCrate.Label),
		NotSafe:              true,
	})
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
		if err == nil {
			err = fmt.Errorf("transaction failed")
		}
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Str("mystery_crate_id", assignedCrate.ID).Err(err).Msg("unable to charge user for mystery crate purchase")
		return terror.Error(err, "Unable to process mystery crate purchase,  check your balance and try again.")
	}

	refundFunc := func() {
		refundSupTransactionID, err := sc.API.Passport.RefundSupsMessage(supTransactionID)
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to refund user for mystery crate purchase cost")
		}

		txItem := &boiler.StorePurchaseHistory{
			PlayerID:    user.ID,
			Amount:      storeCrate.Price,
			ItemType:    "LOOTBOX",
			ItemID:      assignedCrate.ID,
			Description: "refunding mystery crate due to failed transaction",
			TXID:        supTransactionID,
			RefundTXID:  null.StringFrom(refundSupTransactionID),
			RefundedAt:  null.TimeFrom(time.Now()),
		}

		err = txItem.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to insert item into purchase history table.")
		}
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing mystery crate, please try again or contact support.")
	}
	defer tx.Rollback()

	//update purchased value
	assignedCrate.Purchased = true
	_, err = assignedCrate.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("unable to update assigned crate information")
		return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	err = db.InsertNewCollectionItem(tx, "supremacy-general", "mystery_crate", assignedCrate.ID, "MEGA", user.ID, null.StringFrom(""), null.StringFrom(""), null.StringFrom(""), null.StringFrom(""), null.StringFrom(""), null.StringFrom(""), null.StringFrom(""))
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to insert into collection items")
		return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	storeCrate.AmountSold = storeCrate.AmountSold + 1
	_, err = storeCrate.Update(gamedb.StdConn, boil.Whitelist(boiler.StorefrontMysteryCrateColumns.AmountSold))
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
		return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	txItem := &boiler.StorePurchaseHistory{
		PlayerID:    user.ID,
		Amount:      storeCrate.Price,
		ItemType:    "mystery_crate",
		ItemID:      assignedCrate.ID,
		Description: "purchased mystery crate",
		TXID:        supTransactionID,
	}

	err = txItem.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
		return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("failed to commit mystery crate transaction")
		return terror.Error(err, "Issue purchasing mystery crate, please try again or contact support.")
	}
	//-------------------------------------

	//update mysterycrate subscribers and update player
	ws.PublishMessage(fmt.Sprintf("/faction/%s/store/mystery_crate", factionID), HubkeyMysteryCrateSubscribe, nil)

	reply(true)
	return nil
}
