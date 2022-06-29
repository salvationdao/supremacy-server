package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
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

const HubKeyMysteryCrateOwnershipSubscribe = "STORE:MYSTERY:CRATE:OWNERSHIP:SUBSCRIBE"

func (sc *StoreController) MysteryCrateOwnershipSubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	BroadcastCrateOwnershipUpdate(user.ID)
	return nil
}

func BroadcastCrateOwnershipUpdate(userID string) {
	// check keycards
	allowedToBuy := 0
	keycards, err := boiler.PlayerKeycards(
		boiler.PlayerKeycardWhere.PlayerID.EQ(userID),
		qm.Load(boiler.PlayerKeycardRels.BlueprintKeycard),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("userID", userID).Msg("failed to get keycards for user")
		return
	}

	for _, kc := range keycards {
		if kc.R.BlueprintKeycard.KeycardGroup == "KEYCARD" {
			allowedToBuy += kc.Count * 10
		}
	}

	// check owned mystery crates
	mysteryCrateCount, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(userID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("userID", userID).Msg("failed to get mysteryCrateCount for user")
		return
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s/mystery_crates", userID), HubKeyMysteryCrateOwnershipSubscribe, &MysteryCrateOwnershipResp{
		Owned:   int(mysteryCrateCount),
		Allowed: allowedToBuy,
	})
}

const HubKeyGetMysteryCrates = "STORE:MYSTERY:CRATES"

func (sc *StoreController) GetMysteryCratesHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	crates, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get mystery crate")
	}

	resp := server.StoreFrontMysteryCrateSliceFromBoiler(crates)
	reply(resp)

	return nil
}

const HubKeyMysteryCrateSubscribe = "STORE:MYSTERY:CRATE:SUBSCRIBE"

func (sc *StoreController) MysteryCrateSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	crateID := cctx.URLParam("crate_id")

	crate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.ID.EQ(crateID),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
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

	// check keycards
	allowedToBuy := 0
	keycards, err := boiler.PlayerKeycards(
		boiler.PlayerKeycardWhere.PlayerID.EQ(user.ID),
		qm.Load(boiler.PlayerKeycardRels.BlueprintKeycard),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err)
	}

	for _, kc := range keycards {
		if kc.R.BlueprintKeycard.KeycardGroup == "KEYCARD" {
			allowedToBuy += kc.Count * 10
		}
	}

	//check owned mystery crates
	mysteryCrateCount, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
	).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err)
	}

	if int64(allowedToBuy) < (mysteryCrateCount + int64(req.Payload.Quantity)) {
		return terror.Error(fmt.Errorf("Unable to purchase %d mystery crates, owned: %d, allowed: %d", req.Payload.Quantity, mysteryCrateCount, allowedToBuy))
	}

	//double check there are still crates available on storefront, user should not be able to buy it though
	storeCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(req.Payload.Type),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
		qm.Load(boiler.StorefrontMysteryCrateRels.Faction),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get crate for purchase, please try again or contact support.")
	}

	if (storeCrate.AmountSold + req.Payload.Quantity) >= storeCrate.Amount {
		return terror.Error(fmt.Errorf("player ID: %s, attempted to purchase sold out mystery crate", user.ID), "This mystery crate is sold out!")
	}
	//check user SUPS is more than crate.price

	// -------------------------------------
	supTransactionID, err := sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		Amount:               storeCrate.Price.Mul(decimal.NewFromInt(int64(req.Payload.Quantity))).String(),
		FromUserID:           uuid.FromStringOrNil(user.ID),
		ToUserID:             battle.SupremacyUserID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_mystery_crate_purchase|%s|%d", storeCrate.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             "Mystery Crate",
		Description:          fmt.Sprintf("Purchased mystery crate id %s", storeCrate.ID),
		NotSafe:              true,
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
			Amount:      storeCrate.Price.Mul(decimal.NewFromInt(int64(req.Payload.Quantity))),
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

	// Assign multiple crate purchases
	var resp []Reward
	for i := 0; i < req.Payload.Quantity; i++ {
		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Msg("unable to begin tx")
			return terror.Error(err, "Issue purchasing mystery crate, please try again or contact support.")
		}
		defer tx.Rollback()

		assignedCrate, err := assignAndRegisterPurchasedCrate(user.ID, storeCrate, tx, sc.API)
		if err != nil {
			refundFunc()
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

		err = txItem.Insert(tx, boil.Infer())
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
			return terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
		}

		err = tx.Commit()
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Msg("failed to commit mystery crate transaction")
			return terror.Error(err, "Issue purchasing mystery crate, please try again or contact support.")
		}

		resp = append(resp, Reward{
			Label:       storeCrate.MysteryCrateType,
			ImageURL:    storeCrate.ImageURL,
			LockedUntil: null.TimeFrom(assignedCrate.LockedUntil),
		})

	}
	serverStoreCrate := server.StoreFrontMysteryCrateFromBoiler(storeCrate)

	//update mysterycrate subscribers and update player
	ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", factionID, storeCrate.ID), HubKeyMysteryCrateSubscribe, serverStoreCrate)
	go BroadcastCrateOwnershipUpdate(user.ID)

	reply(resp)
	return nil
}

func assignAndRegisterPurchasedCrate(userID string, storeCrate *boiler.StorefrontMysteryCrate, tx *sql.Tx, api *API) (*server.MysteryCrate, error) {
	availableCrates, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.FactionID.EQ(storeCrate.FactionID),
		boiler.MysteryCrateWhere.Type.EQ(storeCrate.MysteryCrateType),
		boiler.MysteryCrateWhere.Purchased.EQ(false),
		boiler.MysteryCrateWhere.Opened.EQ(false),
	).All(tx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get available crates, please try again or contact support.")
	}

	faction, err := boiler.FindFaction(tx, storeCrate.FactionID)
	if err != nil {
		return nil, terror.Error(err, "Failed to find faction, please try again or contact support.")
	}

	//randomly assigning crate to user
	rand.Seed(time.Now().UnixNano())
	assignedCrate := availableCrates[rand.Intn(len(availableCrates))]

	//update purchased value
	assignedCrate.Purchased = true
	_, err = assignedCrate.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to update assigned crate information")
		return nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	collectionItem, err := db.InsertNewCollectionItem(tx,
		"supremacy-general",
		boiler.ItemTypeMysteryCrate,
		assignedCrate.ID,
		"",
		userID,
		storeCrate.ImageURL,
		storeCrate.CardAnimationURL,
		storeCrate.AvatarURL,
		storeCrate.LargeImageURL,
		storeCrate.BackgroundColor,
		storeCrate.AnimationURL,
		storeCrate.YoutubeURL,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to insert into collection items")
		return nil, terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}
	storeCrate.AmountSold = storeCrate.AmountSold + 1
	_, err = storeCrate.Update(tx, boil.Whitelist(boiler.StorefrontMysteryCrateColumns.AmountSold))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to update crate amount sold")
		return nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	//register
	assignedCrateServer := server.MysteryCrateFromBoiler(assignedCrate, collectionItem, null.String{})
	xsynAsset := rpctypes.ServerMysteryCrateToXsynAsset(assignedCrateServer, faction.Label)

	err = api.Passport.AssetRegister(xsynAsset)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", assignedCrate).Msg("failed to register to XSYN")
		return nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	return assignedCrateServer, nil
}
