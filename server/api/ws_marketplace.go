package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// MarketplaceController holds handlers for marketplace
type MarketplaceController struct {
	API *API
}

// NewMarketplaceController creates the marketplace hub
func NewMarketplaceController(api *API) *MarketplaceController {
	marketplaceHub := &MarketplaceController{
		API: api,
	}

	api.SecureUserFactionCommand(HubKeyMarketplaceSalesList, marketplaceHub.SalesListHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardList, marketplaceHub.SalesListKeycardHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesGet, marketplaceHub.SalesGetHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardGet, marketplaceHub.SalesKeycardGetHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesCreate, marketplaceHub.SalesCreateHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardCreate, marketplaceHub.SalesKeycardCreateHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesArchive, marketplaceHub.SalesArchiveHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardArchive, marketplaceHub.SalesKeycardArchiveHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesBuy, marketplaceHub.SalesBuyHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardBuy, marketplaceHub.SalesKeycardBuyHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesBid, marketplaceHub.SalesBidHandler)

	return marketplaceHub
}

const HubKeyMarketplaceSalesList = "MARKETPLACE:SALES:LIST"

type MarketplaceSalesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID             server.UserID       `json:"user_id"`
		SortDir            db.SortByDir        `json:"sort_dir"`
		SortBy             string              `json:"sort_by"`
		FilterRarities     []string            `json:"rarities"`
		FilterListingTypes []string            `json:"listing_types"`
		ItemType           string              `json:"item_type"`
		MinPrice           decimal.NullDecimal `json:"min_price"`
		MaxPrice           decimal.NullDecimal `json:"max_price"`
		Search             string              `json:"search"`
		PageSize           int                 `json:"page_size"`
		Page               int                 `json:"page"`
	} `json:"payload"`
}

type MarketplaceSalesListResponse struct {
	Total   int64                         `json:"total"`
	Records []*server.MarketplaceSaleItem `json:"records"`
}

func (fc *MarketplaceController) SalesListHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	var filters *db.ListFilterRequest
	if req.Payload.ItemType != "" {
		if req.Payload.ItemType != boiler.ItemTypeMech && req.Payload.ItemType != boiler.ItemTypeMysteryCrate {
			return terror.Error(fmt.Errorf("invalid item type"), "Invalid item type received.")
		}
		filters = &db.ListFilterRequest{
			LinkOperator: db.LinkOperatorTypeAnd,
			Items: []*db.ListFilterRequestItem{
				{
					Table:    boiler.TableNames.CollectionItems,
					Column:   boiler.CollectionItemColumns.ItemType,
					Value:    req.Payload.ItemType,
					Operator: db.OperatorValueTypeEquals,
				},
			},
		}
	}

	total, records, err := db.MarketplaceItemSaleList(
		factionID,
		req.Payload.Search,
		filters,
		req.Payload.FilterRarities,
		req.Payload.FilterListingTypes,
		req.Payload.MinPrice,
		req.Payload.MaxPrice,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of items for sale")
		return terror.Error(err, "Failed to get list of items for sale")
	}

	resp := &MarketplaceSalesListResponse{
		Total:   total,
		Records: records,
	}
	reply(resp)

	return nil
}

const HubKeyMarketplaceSalesKeycardList = "MARKETPLACE:SALES:KEYCARD:LIST"

type MarketplaceSalesKeycardListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID   server.UserID         `json:"user_id"`
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   string                `json:"sort_by"`
		MinPrice decimal.NullDecimal   `json:"min_price"`
		MaxPrice decimal.NullDecimal   `json:"max_price"`
		Filter   *db.ListFilterRequest `json:"filter,omitempty"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type MarketplaceSalesListKeycardResponse struct {
	Total   int64                             `json:"total"`
	Records []*server.MarketplaceSaleItem1155 `json:"records"`
}

func (fc *MarketplaceController) SalesListKeycardHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesKeycardListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	filters := req.Payload.Filter
	if filters == nil {
		filters = &db.ListFilterRequest{
			LinkOperator: db.LinkOperatorTypeAnd,
			Items:        []*db.ListFilterRequestItem{},
		}
	}
	if req.Payload.MinPrice.Valid {
		filters.Items = append(filters.Items, &db.ListFilterRequestItem{
			Table:    boiler.TableNames.ItemKeycardSales,
			Column:   boiler.ItemKeycardSaleColumns.BuyoutPrice,
			Value:    req.Payload.MinPrice.Decimal.String(),
			Operator: db.OperatorValueTypeGreaterOrEqual,
		})
	}
	if req.Payload.MaxPrice.Valid {
		filters.Items = append(filters.Items, &db.ListFilterRequestItem{
			Table:    boiler.TableNames.ItemKeycardSales,
			Column:   boiler.ItemKeycardSaleColumns.BuyoutPrice,
			Value:    req.Payload.MaxPrice.Decimal.String(),
			Operator: db.OperatorValueTypeLessOrEqual,
		})
	}

	total, records, err := db.MarketplaceItemKeycardSaleList(
		req.Payload.Search,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of items for sale")
		return terror.Error(err, "Failed to get list of items for sale")
	}

	resp := &MarketplaceSalesListKeycardResponse{
		Total:   total,
		Records: records,
	}
	reply(resp)

	return nil
}

const (
	HubKeyMarketplaceSalesGet        = "MARKETPLACE:SALES:GET"
	HubKeyMarketplaceSalesKeycardGet = "MARKETPLACE:SALES:KEYCARD:GET"
)

type MarketplaceSalesGetRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID uuid.UUID `json:"id"`
	} `json:"payload"`
}

func (fc *MarketplaceController) SalesGetHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	resp, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Sale Item not found.")
	}
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get of items for sale")
		return terror.Error(err, "Failed to get item.")
	}

	reply(resp)

	return nil
}

func (fc *MarketplaceController) SalesKeycardGetHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	resp, err := db.MarketplaceItemKeycardSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Sale Item not found.")
	}
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get keycard for sale")
		return terror.Error(err, "Failed to get keycard.")
	}

	reply(resp)

	return nil
}

const HubKeyMarketplaceSalesCreate = "MARKETPLACE:SALES:CREATE"

type MarketplaceSalesCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemType             string              `json:"item_type"`
		ItemID               uuid.UUID           `json:"item_id"`
		HasBuyout            bool                `json:"has_buyout"`
		HasAuction           bool                `json:"has_auction"`
		HasDutchAuction      bool                `json:"has_dutch_auction"`
		AskingPrice          decimal.NullDecimal `json:"asking_price"`
		AuctionReservedPrice decimal.NullDecimal `json:"auction_reserved_price"`
		AuctionCurrentPrice  decimal.NullDecimal `json:"auction_current_price"`
		DutchAuctionDropRate decimal.NullDecimal `json:"dutch_auction_drop_rate"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesCreateHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing create sale item, try again or contact support."
	req := &MarketplaceSalesCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player requesting to sell item")
		return terror.Error(err, errMsg)
	}

	factionID, err := uuid.FromString(fID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Err(err).
			Msg("Player is not in a faction")
		return terror.Error(err, errMsg)
	}

	// Check price input
	if req.Payload.HasBuyout || req.Payload.HasDutchAuction {
		if !req.Payload.AskingPrice.Valid {
			return terror.Error(terror.ErrInvalidInput, "Asking Price is required.")
		}
	}
	if req.Payload.HasDutchAuction {
		if !req.Payload.AuctionReservedPrice.Valid {
			return terror.Error(terror.ErrInvalidInput, "Reversed Auction Price is required.")
		}
		if !req.Payload.DutchAuctionDropRate.Valid {
			return terror.Error(terror.ErrInvalidInput, "Drop Rate is required.")
		}
	}

	// Check if allowed to sell item
	if req.Payload.ItemType != boiler.ItemTypeMech && req.Payload.ItemType != boiler.ItemTypeMysteryCrate {
		return terror.Error(fmt.Errorf("invalid item type"), "Invalid Item Type input received.")
	}
	var collectionItemID uuid.UUID
	err = boiler.CollectionItems(
		qm.Select(boiler.CollectionItemColumns.ID),
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.ItemID.String()),
		boiler.CollectionItemWhere.ItemType.EQ(req.Payload.ItemType),
	).QueryRow(gamedb.StdConn).Scan(&collectionItemID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("item_type", req.Payload.ItemType).
			Err(err).
			Msg("Failed to get collection item.")
		if errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Item not found.")
		}
		return terror.Error(err, errMsg)
	}
	if collectionItemID == uuid.Nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("item_type", req.Payload.ItemType).
			Err(err).
			Msg("Unable to parse collection item id")
		return terror.Error(err, errMsg)
	}

	alreadySelling, err := db.MarketplaceCheckCollectionItem(collectionItemID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to check if already selling collection item.")
		if errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Item not found.")
		}
		return terror.Error(err, errMsg)
	}
	if alreadySelling {
		return terror.Error(fmt.Errorf("item is already for sale on marketplace"), "Item is already for sale on Marketplace.")
	}

	// Process listing fee
	factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}

	balance := mp.API.Passport.UserBalanceGet(userID)
	feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(10))
	if req.Payload.HasBuyout {
		feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingBuyoutFee, decimal.NewFromInt(5)))
	}
	if req.Payload.AuctionReservedPrice.Valid {
		feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingAuctionReserveFee, decimal.NewFromInt(5)))
	}
	feePrice = feePrice.Mul(decimal.New(1, 18))

	if balance.Sub(feePrice).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups to list item.")
	}

	// Pay Listing Fees
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               feePrice.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|%s|%s|%d", req.Payload.ItemType, req.Payload.ItemID.String(), time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace List Item Fee: %s (%s)", req.Payload.ItemID.String(), req.Payload.ItemType),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process marketplace fee transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Marketplace Fee.")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Begin transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to start db transaction (new sale item).")
		return terror.Error(err, errMsg)
	}
	defer tx.Rollback()

	// Create Sales Item
	// TODO: Add listing hours option back with fee rates applied
	endAt := time.Now().Add(time.Hour * 24)
	obj, err := db.MarketplaceSaleCreate(
		tx,
		userID,
		factionID,
		txid,
		endAt,
		collectionItemID,
		req.Payload.HasBuyout,
		req.Payload.AskingPrice,
		req.Payload.HasAuction,
		req.Payload.AuctionReservedPrice,
		req.Payload.AuctionCurrentPrice,
		req.Payload.HasDutchAuction,
		req.Payload.DutchAuctionDropRate,
	)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, errMsg)
	}

	// Unlock Item
	collectionItem := boiler.CollectionItem{
		ID:                  collectionItemID.String(),
		LockedToMarketplace: true,
	}
	_, err = collectionItem.Update(tx, boil.Whitelist(
		boiler.CollectionItemColumns.ID,
		boiler.CollectionItemColumns.LockedToMarketplace,
	))
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, errMsg)
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to commit db transaction (new sale item).")
		return terror.Error(err, errMsg)
	}

	reply(obj)

	return nil
}

const HubKeyMarketplaceSalesKeycardCreate = "MARKETPLACE:SALES:KEYCARD:CREATE"

type HubKeyMarketplaceSalesKeycardCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemID      uuid.UUID       `json:"item_id"`
		AskingPrice decimal.Decimal `json:"asking_price"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesKeycardCreateHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing create keycard sale item, try again or contact support."
	req := &HubKeyMarketplaceSalesKeycardCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player requesting to sell item")
		return terror.Error(err, errMsg)
	}

	factionID, err := uuid.FromString(fID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Err(err).
			Msg("Player is not in a faction")
		return terror.Error(err, errMsg)
	}

	factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}

	// Check if can sell any keycards
	keycard, err := db.PlayerKeycard(req.Payload.ItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Player Keycard not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", req.Payload.ItemID.String()).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get player's keycard")
		return terror.Error(err, errMsg)
	}
	numKeycardsSelling, err := db.MarketplaceCountKeycards(req.Payload.ItemID)
	if err != nil {
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", req.Payload.ItemID.String()).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to check number of keycards in marketplace")
		return terror.Error(err, errMsg)
	}
	if keycard.Count <= numKeycardsSelling {
		return terror.Error(fmt.Errorf("all keycards are on marketplace"), "Your keycard(s) are already for sale on Marketplace.")
	}

	// Process fee
	balance := mp.API.Passport.UserBalanceGet(userID)

	feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(10)).Mul(decimal.New(1, 18))

	if balance.Sub(feePrice).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups to list item.")
	}

	// Pay sup
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               feePrice.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|keycard|%s|%d", req.Payload.ItemID.String(), time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace List Item Fee: %s (keycard)", req.Payload.ItemID.String()),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process marketplace fee transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Marketplace Fee.")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Create Sales Item
	// TODO: Add listing hours option back with fee rates applied
	endAt := time.Now().Add(time.Hour * 24)
	obj, err := db.MarketplaceKeycardSaleCreate(
		userID,
		factionID,
		txid,
		endAt,
		req.Payload.ItemID,
		req.Payload.AskingPrice,
	)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	reply(obj)

	return nil
}

const (
	HubKeyMarketplaceSalesArchive        = "MARKETPLACE:SALES:ARCHIVE"
	HubKeyMarketplaceSalesKeycardArchive = "MARKETPLACE:SALES:KEYCARD:ARCHIVE"
)

type MarketplaceSalesCancelRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID uuid.UUID `json:"id"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesArchiveHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue cancelling sale item, try again or contact support."
	req := &MarketplaceSalesCancelRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check whether user can cancel sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if saleItem.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user.")
	}
	if saleItem.SoldBy.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}

	// Cancel item
	err = db.MarketplaceSaleArchive(gamedb.StdConn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	return nil
}

func (mp *MarketplaceController) SalesKeycardArchiveHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue cancelling sale item, try again or contact support."
	req := &MarketplaceSalesCancelRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check whether user can cancel sale item
	saleItem, err := db.MarketplaceItemKeycardSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if saleItem.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user.")
	}
	if saleItem.SoldBy.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}

	// Cancel item
	err = db.MarketplaceKeycardSaleArchive(gamedb.StdConn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	return nil
}

const (
	HubKeyMarketplaceSalesBuy        = "MARKETPLACE:SALES:BUY"
	HubKeyMarketplaceSalesKeycardBuy = "MARKETPLACE:SALES:KEYCARD:BUY"
)

type MarketplaceSalesBuyRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID uuid.UUID `json:"id"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesBuyHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue buying sale item, try again or contact support."
	req := &MarketplaceSalesBuyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if saleItem.FactionID != fID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user's faction.")
	}
	if saleItem.SoldBy.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}
	if saleItem.CollectionItem.XsynLocked || saleItem.CollectionItem.MarketLocked {
		return terror.Error(fmt.Errorf("item is locked"), "Item is no longer for sale.")
	}
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve buyer's user id.")
		return terror.Error(err, errMsg)
	}

	// Calculate Cost depending on sale type
	saleType := "buyout"
	saleItemCost := saleItem.BuyoutPrice.Decimal
	if saleItem.DutchAuction {
		saleType = "dutch_auction"
		if !saleItem.DutchAuctionDropRate.Valid {
			gamelog.L.Error().
				Str("user_id", user.ID).
				Str("item_sale_id", req.Payload.ID.String()).
				Msg("Dutch Auction Drop rate is missing.")
			return terror.Error(fmt.Errorf("dutch auction drop rate is missing"), errMsg)
		}
		minutesLapse := decimal.NewFromFloat(math.Floor(time.Now().Sub(saleItem.CreatedAt).Minutes()))
		dutchAuctionAmount := saleItem.BuyoutPrice.Decimal.Sub(saleItem.DutchAuctionDropRate.Decimal.Mul(minutesLapse))
		if dutchAuctionAmount.GreaterThanOrEqual(saleItem.AuctionCurrentPrice.Decimal) {
			saleItemCost = dutchAuctionAmount
		} else {
			saleItemCost = saleItem.AuctionCurrentPrice.Decimal
		}
	}

	salesCutPercentageFee := db.GetDecimalWithDefault(db.KeyMarketplaceSaleCutPercentageFee, decimal.NewFromFloat(0.3))

	balance := mp.API.Passport.UserBalanceGet(userID)
	if balance.Sub(saleItemCost).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	// Pay sales cut fee amount to faction account
	factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}
	feeTXID, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               saleItemCost.Mul(salesCutPercentageFee).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item_fee:%s|%s|%d", saleType, saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Fee: %s", saleItem.ID),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to process sales cut fee transaction for Purchase Sale Item.")
		return terror.Error(err, errMsg)
	}

	// Give sales cut amount to seller
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(saleItem.OwnerID)),
		Amount:               saleItemCost.Mul(decimal.NewFromInt(1).Sub(salesCutPercentageFee)).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item:%s|%s|%d", saleType, saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Payment (%d%% cut): %s", salesCutPercentageFee.Mul(decimal.NewFromInt(100)).IntPart(), saleItem.ID),
		NotSafe:              true,
	})
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		err = fmt.Errorf("failed to process payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, errMsg)
	}

	// Start transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to start purchase sale item db transaction.")
		return terror.Error(err, errMsg)
	}
	defer tx.Rollback()

	// Update sale item
	saleItemRecord := &boiler.ItemSale{
		ID:          saleItem.ID,
		SoldAt:      null.TimeFrom(time.Now()),
		SoldFor:     decimal.NewNullDecimal(saleItemCost),
		SoldTXID:    null.StringFrom(txid),
		SoldFeeTXID: null.StringFrom(feeTXID),
		SoldBy:      null.StringFrom(user.ID),
		UpdatedAt:   time.Now(),
	}
	_, err = saleItemRecord.Update(tx,
		boil.Whitelist(
			boiler.ItemSaleColumns.SoldAt,
			boiler.ItemSaleColumns.SoldFor,
			boiler.ItemSaleColumns.SoldTXID,
			boiler.ItemSaleColumns.SoldFeeTXID,
			boiler.ItemSaleColumns.SoldBy,
			boiler.ItemSaleColumns.UpdatedAt,
		))
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		err = fmt.Errorf("failed to complete payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, errMsg)
	}

	// Transfer ownership of asset
	if saleItem.CollectionItemType == boiler.ItemTypeMech {
		err = db.ChangeMechOwner(tx, req.Payload.ID)
		if err != nil {
			mp.API.Passport.RefundSupsMessage(feeTXID)
			mp.API.Passport.RefundSupsMessage(txid)
			gamelog.L.Error().
				Str("from_user_id", user.ID).
				Str("to_user_id", saleItem.OwnerID).
				Str("balance", balance.String()).
				Str("cost", saleItemCost.String()).
				Str("item_sale_id", req.Payload.ID.String()).
				Err(err).
				Msg("Failed to Transfer Mech to New Owner")
			return terror.Error(err, errMsg)
		}
	} else if saleItem.CollectionItemType == boiler.ItemTypeMysteryCrate {
		err = db.ChangeMysteryCrateOwner(tx, req.Payload.ID)
		if err != nil {
			mp.API.Passport.RefundSupsMessage(feeTXID)
			mp.API.Passport.RefundSupsMessage(txid)
			gamelog.L.Error().
				Str("from_user_id", user.ID).
				Str("to_user_id", saleItem.OwnerID).
				Str("balance", balance.String()).
				Str("cost", saleItemCost.String()).
				Str("item_sale_id", req.Payload.ID.String()).
				Err(err).
				Msg("Failed to Transfer Mystery Crate to New Owner")
			return terror.Error(err, errMsg)
		}
	}

	// Unlock Listed Item
	collectionItem := boiler.CollectionItem{
		ID:                  saleItem.CollectionItemID,
		LockedToMarketplace: false,
	}
	_, err = collectionItem.Update(tx, boil.Whitelist(
		boiler.CollectionItemColumns.ID,
		boiler.CollectionItemColumns.LockedToMarketplace,
	))
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		err = fmt.Errorf("failed to complete payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to unlock marketplace listed collection item.")
		return terror.Error(err, errMsg)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to commit purchase sale item db transaction.")
		return terror.Error(err, errMsg)
	}

	// success
	reply(true)

	return nil
}

func (mp *MarketplaceController) SalesKeycardBuyHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue buying sale item, try again or contact support."
	req := &MarketplaceSalesBuyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemKeycardSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if saleItem.SoldBy.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}

	// Pay item
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve buyer's user id.")
		return terror.Error(err, errMsg)
	}

	saleItemCost := saleItem.BuyoutPrice

	balance := mp.API.Passport.UserBalanceGet(userID)
	if balance.Sub(saleItemCost).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	salesCutPercentageFee := db.GetDecimalWithDefault(db.KeyMarketplaceSaleCutPercentageFee, decimal.NewFromFloat(0.3))

	// Pay sales cut fee amount to faction account
	factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}
	feeTXID, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               saleItemCost.Mul(salesCutPercentageFee).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item_fee:buyout|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Fee: %s", saleItem.ID),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to process sales cut fee transaction for Purchase Sale Item.")
		return terror.Error(err, errMsg)
	}

	// Give sales cut amount to seller
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(saleItem.OwnerID)),
		Amount:               saleItemCost.Mul(decimal.NewFromInt(1).Sub(salesCutPercentageFee)).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item_keycard|buyout|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Payment (%d%% cut): %s", salesCutPercentageFee.Mul(decimal.NewFromInt(100)).IntPart(), saleItem.ID),
		NotSafe:              true,
	})
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		err = fmt.Errorf("failed to process payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// Begin transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to start purchase sale item db transaction.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}
	defer tx.Rollback()

	// Update sale item
	saleItemRecord := &boiler.ItemKeycardSale{
		ID:          saleItem.ID,
		SoldAt:      null.TimeFrom(time.Now()),
		SoldFor:     decimal.NewNullDecimal(saleItemCost),
		SoldTXID:    null.StringFrom(txid),
		SoldFeeTXID: null.StringFrom(feeTXID),
		SoldBy:      null.StringFrom(user.ID),
	}

	_, err = saleItemRecord.Update(tx, boil.Whitelist(
		boiler.ItemKeycardSaleColumns.SoldAt,
		boiler.ItemKeycardSaleColumns.SoldFor,
		boiler.ItemKeycardSaleColumns.SoldTXID,
		boiler.ItemKeycardSaleColumns.SoldFeeTXID,
		boiler.ItemKeycardSaleColumns.SoldBy,
	))
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		err = fmt.Errorf("failed to complete payment transaction")
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to update to Keycard Sale Item.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// Transfer ownership of asset
	err = db.ChangeKeycardOwner(tx, req.Payload.ID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to Transfer Keycard to New Owner")
		return terror.Error(err, "Failed to process transaction for Purchase Sale Item.")
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("from_user_id", user.ID).
			Str("to_user_id", saleItem.OwnerID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to commit purchase sale item db transaction.")
		return terror.Error(err, "Failed to process transaction for Purchase Sale Item.")
	}

	// Success
	reply(true)

	return nil
}

const HubKeyMarketplaceSalesBid = "MARKETPLACE:SALES:BID"

type MarketplaceSalesBidRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID     uuid.UUID       `json:"id"`
		Amount decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesBidHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue placing bid, try again or contact support."
	req := &MarketplaceSalesBidRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get Faction Account sending bid amount to
	factionAccountID, ok := server.FactionUsers[fID]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if !saleItem.Auction {
		return terror.Error(fmt.Errorf("item is not up for auction"), "Item is not up for auction.")
	}
	if saleItem.SoldBy.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}
	if saleItem.FactionID != fID {
		return terror.Error(fmt.Errorf("item does not belong to user's faction"), "Item does not belong to user's faction.")
	}
	if saleItem.CollectionItem.XsynLocked || saleItem.CollectionItem.MarketLocked {
		return terror.Error(fmt.Errorf("item is locked"), "Item is no longer for sale.")
	}
	bidAmount := req.Payload.Amount.Mul(decimal.New(1, 18))
	if bidAmount.LessThanOrEqual(saleItem.AuctionCurrentPrice.Decimal) {
		return terror.Error(fmt.Errorf("bid amount less than current bid amount"), "Invalid bid amount, must be above the current bid price.")
	}

	// Pay bid amount
	balance := mp.API.Passport.UserBalanceGet(userID)
	if balance.Sub(req.Payload.Amount).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("balance", balance.String()).
			Str("cost", req.Payload.Amount.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               bidAmount.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item:auction_bid|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Bid Item: %s", saleItem.ID),
		NotSafe:              true,
	})

	// Start Transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("current_auction_price", saleItem.AuctionCurrentPrice.Decimal.String()).
			Str("bid_amount", req.Payload.Amount.String()).
			Err(err).
			Msg("Failed to cancel previous bid(s).")
		return terror.Error(err, errMsg)
	}
	defer tx.Rollback()

	// Cancel all other bids before placing in the next new bid
	refundTxnIDs, err := db.MarketplaceSaleCancelBids(tx, req.Payload.ID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("current_auction_price", saleItem.AuctionCurrentPrice.Decimal.String()).
			Str("bid_amount", req.Payload.Amount.String()).
			Err(err).
			Msg("Failed to cancel previous bid(s).")
		return terror.Error(err, errMsg)
	}

	// Place Bid
	_, err = db.MarketplaceSaleBidHistoryCreate(tx, req.Payload.ID, userID, req.Payload.Amount, txid)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("current_auction_price", saleItem.AuctionCurrentPrice.Decimal.String()).
			Str("bid_amount", req.Payload.Amount.String()).
			Err(err).
			Msg("Unable to place bid.")
		return terror.Error(err, errMsg)
	}

	err = db.MarketplaceSaleAuctionSync(tx, req.Payload.ID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("bid_amount", req.Payload.Amount.String()).
			Err(err).
			Msg("Unable to update current auction price.")
		return terror.Error(err, errMsg)
	}

	// Refund other bids
	for _, bidTxID := range refundTxnIDs {
		refundTxID, err := mp.API.Passport.RefundSupsMessage(bidTxID)
		if err != nil {
			gamelog.L.Error().
				Str("item_sale_id", req.Payload.ID.String()).
				Str("txid", bidTxID).
				Err(err).
				Msg("Unable to refund cancelled bid.")
			continue
		}
		err = db.MarketplaceSaleBidHistoryRefund(tx, req.Payload.ID, bidTxID, refundTxID, false)
		if err != nil {
			gamelog.L.Error().
				Str("item_sale_id", req.Payload.ID.String()).
				Str("txid", bidTxID).
				Str("refund_tx_id", refundTxID).
				Err(err).
				Msg("Unable to update cancelled bid refund tx id.")
			continue
		}
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("bid_amount", req.Payload.Amount.String()).
			Err(err).
			Msg("Unable to update current auction price.")
		return terror.Error(err, errMsg)
	}

	reply(true)

	// Broadcast new current price
	totalBids, err := boiler.ItemSalesBidHistories(boiler.ItemSalesBidHistoryWhere.ItemSaleID.EQ(req.Payload.ID.String())).Count(gamedb.StdConn)
	if err != nil {
		// No need to abort failure
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_sale_id", req.Payload.ID.String()).
			Str("bid_amount", req.Payload.Amount.String()).
			Err(err).
			Msg("Unable to get current total bids.")
		return nil
	}

	resp := &SaleItemUpdate{
		AuctionCurrentPrice: req.Payload.Amount.Mul(decimal.New(1, 18)).String(),
		TotalBids:           totalBids,
		LastBid: server.MarketplaceBidder{
			ID:            null.StringFrom(user.ID),
			FactionID:     user.FactionID,
			Username:      user.Username,
			PublicAddress: user.PublicAddress,
			Gid:           null.IntFrom(user.Gid),
		},
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/marketplace/%s", fID, req.Payload.ID.String()), HubKeyMarketplaceSalesItemUpdate, resp)

	return nil
}

const HubKeyMarketplaceSalesItemUpdate = "MARKETPLACE:SALES:ITEM:UPDATE"

type SaleItemUpdate struct {
	AuctionCurrentPrice string                   `json:"auction_current_price"`
	TotalBids           int64                    `json:"total_bids"`
	LastBid             server.MarketplaceBidder `json:"last_bid,omitempty"`
}

func (mp *MarketplaceController) SalesItemUpdateSubscriber(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	itemSaleID := cctx.URLParam("id")
	if itemSaleID == "" {
		return terror.Error(fmt.Errorf("item sale id is required"), "Item Sale ID is required.")
	}

	itemSaleUUID, err := uuid.FromString(itemSaleID)
	if err != nil {
		return terror.Error(fmt.Errorf("item sale id is invalid"), "Item Slale ID is invalid.")
	}

	// TODO: Update this when Keycards are available for auction
	saleItem, err := db.MarketplaceItemSale(itemSaleUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Unable to find item sale id.")
	}
	if err != nil {
		return terror.Error(err, "Unable to get latest item sale update.")
	}

	resp := &SaleItemUpdate{
		AuctionCurrentPrice: saleItem.AuctionCurrentPrice.Decimal.String(),
		TotalBids:           saleItem.TotalBids,
		LastBid:             saleItem.LastBid,
	}

	reply(resp)

	return nil
}
