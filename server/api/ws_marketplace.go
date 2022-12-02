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
	"server/marketplace"
	"server/xsyn_rpcclient"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	api.SecureUserFactionCommand(HubKeyMarketplaceEventList, marketplaceHub.EventListHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesGet, marketplaceHub.SalesGetHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardGet, marketplaceHub.SalesKeycardGetHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesCreate, WithMarketLockCheck(marketplaceHub.SalesCreateHandler))
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardCreate, WithMarketLockCheck(marketplaceHub.SalesKeycardCreateHandler))
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesArchive, marketplaceHub.SalesArchiveHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardArchive, marketplaceHub.SalesKeycardArchiveHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesBuy, WithMarketLockCheck(marketplaceHub.SalesBuyHandler))
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardBuy, WithMarketLockCheck(marketplaceHub.SalesKeycardBuyHandler))
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesBid, WithMarketLockCheck(marketplaceHub.SalesBidHandler))

	return marketplaceHub
}

func WithMarketLockCheck(fn server.SecureFactionCommandFunc) server.SecureFactionCommandFunc {
	return func(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
		locked, err := boiler.BlockMarketplaces(
			boiler.BlockMarketplaceWhere.PublicAddress.EQ(user.PublicAddress.String),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error checking market lock: %s", err)
		}
		if locked == nil {
			return fn(ctx, user, user.FactionID.String, key, payload, reply)
		}

		if locked.BlockedUntil.After(time.Now()) {
			return fmt.Errorf("you are market locked until %s", locked.BlockedUntil.Format("02-01-2006"))
		}

		return fn(ctx, user, user.FactionID.String, key, payload, reply)
	}
}

const HubKeyMarketplaceSalesList = "MARKETPLACE:SALES:LIST"

type MarketplaceSalesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID             server.UserID                   `json:"user_id"`
		SortDir            db.SortByDir                    `json:"sort_dir"`
		SortBy             string                          `json:"sort_by"`
		FilterRarities     []string                        `json:"rarities"`
		FilterListingTypes []string                        `json:"listing_types"`
		FilterWeaponTypes  []string                        `json:"weapon_types"`
		FilterWeaponStats  *db.MarketplaceWeaponStatFilter `json:"weapon_stats"`
		FilterOwnedBy      []string                        `json:"owned_by"`
		Sold               bool                            `json:"sold"`
		ItemType           string                          `json:"item_type"`
		MinPrice           decimal.NullDecimal             `json:"min_price"`
		MaxPrice           decimal.NullDecimal             `json:"max_price"`
		Search             string                          `json:"search"`
		PageSize           int                             `json:"page_size"`
		Page               int                             `json:"page"`
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

	total, records, err := db.MarketplaceItemSaleList(
		req.Payload.ItemType,
		user.ID,
		factionID,
		req.Payload.Search,
		req.Payload.FilterRarities,
		req.Payload.FilterListingTypes,
		req.Payload.FilterWeaponTypes,
		req.Payload.FilterWeaponStats,
		req.Payload.FilterOwnedBy,
		req.Payload.Sold,
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
		UserID        server.UserID         `json:"user_id"`
		SortDir       db.SortByDir          `json:"sort_dir"`
		SortBy        string                `json:"sort_by"`
		FilterOwnedBy []string              `json:"owned_by"`
		MinPrice      decimal.NullDecimal   `json:"min_price"`
		MaxPrice      decimal.NullDecimal   `json:"max_price"`
		Filter        *db.ListFilterRequest `json:"filter,omitempty"`
		Search        string                `json:"search"`
		PageSize      int                   `json:"page_size"`
		Page          int                   `json:"page"`
		Sold          bool                  `json:"sold"`
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
		user.ID,
		factionID,
		req.Payload.Search,
		req.Payload.Filter,
		req.Payload.FilterOwnedBy,
		req.Payload.MinPrice,
		req.Payload.MaxPrice,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
		req.Payload.Sold,
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

const HubKeyMarketplaceEventList = "MARKETPLACE:EVENT:LIST"

type MarketplaceEventListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir   db.SortByDir `json:"sort_dir"`
		SortBy    string       `json:"sort_by"`
		EventType []string     `json:"event_type"`
		Search    string       `json:"search"`
		PageSize  int          `json:"page_size"`
		Page      int          `json:"page"`
	} `json:"payload"`
}

type MarketplaceEventListResponse struct {
	Total   int64                      `json:"total"`
	Records []*server.MarketplaceEvent `json:"records"`
}

func (fc *MarketplaceController) EventListHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceEventListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, records, err := db.MarketplaceEventList(
		user.ID,
		req.Payload.Search,
		req.Payload.EventType,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of items for sale")
		return terror.Error(err, "Failed to get list of items for sale")
	}

	resp := &MarketplaceEventListResponse{
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

	if resp.FactionID != factionID {
		return terror.Error(fmt.Errorf("you can only access your syndicates marketplace"))
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
		ListingDurationHours time.Duration       `json:"listing_duration_hours"`
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
	hasBuyout := false
	hasAuction := false
	hasDutchAuction := false

	if req.Payload.AskingPrice.Valid {
		if req.Payload.AskingPrice.Decimal.LessThanOrEqual(decimal.Zero) {
			return terror.Error(fmt.Errorf("invalid asking price"), "Invalid asking price received.")
		}
		hasBuyout = true
	}
	if req.Payload.DutchAuctionDropRate.Valid {
		if req.Payload.DutchAuctionDropRate.Decimal.LessThanOrEqual(decimal.Zero) {
			return terror.Error(fmt.Errorf("invalid drop rate"), "Invalid drop rate received.")
		}
		hasDutchAuction = true
		hasBuyout = false
	}
	if req.Payload.AuctionCurrentPrice.Valid || (req.Payload.AuctionReservedPrice.Valid && !req.Payload.DutchAuctionDropRate.Valid) {
		if req.Payload.AuctionCurrentPrice.Valid && req.Payload.AuctionCurrentPrice.Decimal.LessThan(decimal.Zero) {
			return terror.Error(fmt.Errorf("invalid auction current price"), "Invalid auction current price received.")
		}
		if req.Payload.AuctionReservedPrice.Valid && req.Payload.AuctionReservedPrice.Decimal.LessThan(decimal.Zero) {
			return terror.Error(fmt.Errorf("invalid auction reserved price"), "Invalid auction reserved price received.")
		}
		hasAuction = true
	}

	if !hasBuyout && !hasAuction && !hasDutchAuction {
		return terror.Error(fmt.Errorf("invalid sales input received"), "Unable to determine listing sale type from given input.")
	}

	// Check if allowed to sell item
	if req.Payload.ItemType != boiler.ItemTypeMech && req.Payload.ItemType != boiler.ItemTypeMysteryCrate && req.Payload.ItemType != boiler.ItemTypeWeapon {
		return terror.Error(fmt.Errorf("invalid item type"), "Invalid Item Type input received.")
	}

	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.ItemID.String()),
		boiler.CollectionItemWhere.ItemType.EQ(req.Payload.ItemType),
	).One(gamedb.StdConn)
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

	if collectionItem.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user.")
	}

	if collectionItem.MarketLocked {
		return terror.Error(fmt.Errorf("unable to list assets staked with old staking contract"))
	}

	if collectionItem.XsynLocked {
		return terror.Error(fmt.Errorf("asset does not live on supremacy"))
	}

	ciUUID := uuid.FromStringOrNil(collectionItem.ID)

	if ciUUID.IsNil() {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("item_type", req.Payload.ItemType).
			Err(err).
			Msg("Unable to parse collection item id")
		return terror.Error(err, errMsg)
	}

	// check if opened
	if collectionItem.ItemType == boiler.ItemTypeMysteryCrate {
		crate, err := boiler.MysteryCrates(
			boiler.MysteryCrateWhere.ID.EQ(collectionItem.ItemID),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().
				Str("user_id", user.ID).
				Str("item_id", req.Payload.ItemID.String()).
				Str("item_type", req.Payload.ItemType).
				Err(err).
				Msg("unable to check whether crate is opened")
			return err
		}
		if crate.Opened {
			return fmt.Errorf("unable to list opened crates")
		}
	}

	// check if weapon is equipped
	if collectionItem.ItemType == boiler.ItemTypeWeapon {
		equipped, err := db.CheckWeaponAttached(collectionItem.ItemID)
		if err != nil {
			gamelog.L.Error().
				Str("user_id", user.ID).
				Str("item_id", req.Payload.ItemID.String()).
				Str("item_type", req.Payload.ItemType).
				Err(err).
				Msg("unable to check whether weapon is attached")
			return err
		}
		if equipped {
			return fmt.Errorf("cannot sell weapons attached to a war machine")
		}
	}

	// check if queue
	if collectionItem.ItemType == boiler.ItemTypeMech {
		blm, err := boiler.BattleLobbiesMechs(
			boiler.BattleLobbiesMechWhere.MechID.EQ(collectionItem.ItemID),
			boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Str("mech id", collectionItem.ItemID).Msg("Failed to check mech queue.")
			return terror.Error(err, "Failed to check mech queue.")
		}

		if blm != nil {
			return fmt.Errorf("cannot sell war machine which is already in battle lobby")
		}
	}

	alreadySelling, err := db.MarketplaceCheckCollectionItem(ciUUID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("item_type", req.Payload.ItemType).
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
	//factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	//if !ok {
	//	err = fmt.Errorf("failed to get hard coded syndicate player id")
	//	gamelog.L.Error().
	//		Str("player_id", user.ID).
	//		Str("faction_id", user.FactionID.String).
	//		Err(err).
	//		Msg("unable to get hard coded syndicate player ID from faction ID")
	//	return terror.Error(err, errMsg)
	//}

	// balance := mp.API.Passport.UserBalanceGet(userID)
	// feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(10))
	// if hasBuyout {
	// 	feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingBuyoutFee, decimal.NewFromInt(5)))
	// }
	// if req.Payload.AuctionReservedPrice.Valid {
	// 	feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingAuctionReserveFee, decimal.NewFromInt(5)))
	// }
	// if req.Payload.ListingDurationHours > 24 {
	// 	listingDurationFee := (req.Payload.ListingDurationHours/24 - 1) * 5
	// 	feePrice = feePrice.Add(decimal.NewFromInt(int64(listingDurationFee)))
	// }

	// feePrice = feePrice.Mul(decimal.New(1, 18))

	// if balance.Sub(feePrice).LessThan(decimal.Zero) {
	// 	err = fmt.Errorf("insufficient funds")
	// 	gamelog.L.Error().
	// 		Str("user_id", user.ID).
	// 		Str("balance", balance.String()).
	// 		Str("item_type", req.Payload.ItemType).
	// 		Str("item_id", req.Payload.ItemID.String()).
	// 		Err(err).
	// 		Msg("Player does not have enough sups.")
	// 	return terror.Error(err, "You do not have enough sups to list item.")
	// }

	// // Pay Listing Fees
	// txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
	// 	FromUserID:           userID,
	// 	ToUserID:             uuid.Must(uuid.FromString(server.SupremacyChallengeFundUserID)), // NOTE: send fees to challenge fund for now. (was faction account)
	// 	Amount:               feePrice.String(),
	// 	TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|%s|%s|%d", req.Payload.ItemType, req.Payload.ItemID.String(), time.Now().UnixNano())),
	// 	Group:                string(server.TransactionGroupSupremacy),
	// 	SubGroup:             string(server.TransactionGroupMarketplace),
	// 	Description:          fmt.Sprintf("Marketplace List Item Fee: %s (%s)", req.Payload.ItemID.String(), req.Payload.ItemType),
	// })
	// if err != nil {
	// 	err = fmt.Errorf("failed to process marketplace fee transaction")
	// 	gamelog.L.Error().
	// 		Str("user_id", user.ID).
	// 		Str("balance", balance.String()).
	// 		Str("item_type", req.Payload.ItemType).
	// 		Str("item_id", req.Payload.ItemID.String()).
	// 		Err(err).
	// 		Msg("Failed to process transaction for Marketplace Fee.")
	// 	return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	// }

	// // trigger challenge fund update
	// defer func() {
	// 	mp.API.ArenaManager.ChallengeFundUpdateChan <- true
	// }()

	// Begin transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", req.Payload.ItemType).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to start db transaction (new sale item).")
		return terror.Error(err, errMsg)
	}
	defer tx.Rollback()

	// Create Sales Item
	endAt := time.Now()
	if mp.API.Config.Environment == "staging" {
		endAt = endAt.Add(time.Minute * 5)
	} else {
		endAt = endAt.Add(time.Hour * req.Payload.ListingDurationHours)
	}
	obj, err := db.MarketplaceSaleCreate(
		tx,
		userID,
		factionID,
		null.String{},
		endAt,
		ciUUID,
		hasBuyout,
		req.Payload.AskingPrice,
		hasAuction,
		req.Payload.AuctionReservedPrice,
		req.Payload.AuctionCurrentPrice,
		hasDutchAuction,
		req.Payload.DutchAuctionDropRate,
	)
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", req.Payload.ItemType).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, errMsg)
	}

	// Lock Item
	collectionItem.LockedToMarketplace = true
	_, err = collectionItem.Update(tx, boil.Whitelist(
		boiler.CollectionItemColumns.ID,
		boiler.CollectionItemColumns.LockedToMarketplace,
	))
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", req.Payload.ItemType).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, errMsg)
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", req.Payload.ItemType).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to commit db transaction (new sale item).")
		return terror.Error(err, errMsg)
	}

	reply(obj)

	// Log Event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventCreated, user.ID, decimal.NullDecimal{}, obj.ID, boiler.TableNames.ItemSales)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", req.Payload.ItemType).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg(fmt.Sprintf("Failed to log create sale item event (%s).", req.Payload.ItemType))
	}

	// Broadcast Queue Status Market
	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ID.EQ(ciUUID.String()),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("collection item id", ciUUID.String()).Err(err).Msg("Failed to get collection item from db")
	}

	if ci.ItemType == boiler.ItemTypeMech {
		mp.API.ArenaManager.MechDebounceBroadcastChan <- []string{ci.ItemID}
	}

	return nil
}

const HubKeyMarketplaceSalesKeycardCreate = "MARKETPLACE:SALES:KEYCARD:CREATE"

type HubKeyMarketplaceSalesKeycardCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemID               uuid.UUID       `json:"item_id"`
		AskingPrice          decimal.Decimal `json:"asking_price"`
		ListingDurationHours time.Duration   `json:"listing_duration_hours"`
	} `json:"payload"`
}

type AttributeInner struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
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

	//factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	//if !ok {
	//	err = fmt.Errorf("failed to get hard coded syndicate player id")
	//	gamelog.L.Error().
	//		Str("player_id", user.ID).
	//		Str("faction_id", user.FactionID.String).
	//		Err(err).
	//		Msg("unable to get hard coded syndicate player ID from faction ID")
	//	return terror.Error(err, errMsg)
	//}

	// Check if can sell any keycards
	keycards, err := db.PlayerKeycards("", req.Payload.ItemID.String())
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

	if len(keycards) == 0 {
		return terror.Error(fmt.Errorf("keycard not found"), "Keycard not found.")
	}

	keycard := keycards[0]

	if keycard.Count < 1 {
		return terror.Error(fmt.Errorf("all keycards are on marketplace"), "Your keycard(s) are already for sale on Marketplace.")
	}

	if keycard.PlayerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user.")
	}

	keycardBlueprint, err := boiler.BlueprintKeycards(boiler.BlueprintKeycardWhere.ID.EQ(keycard.BlueprintKeycardID)).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get blueprint keycard")
	}

	var assetJson types.JSON

	if !keycardBlueprint.Syndicate.Valid {
		keycardBlueprint.Syndicate.String = "N/A"
	}

	inner := &AttributeInner{
		TraitType: "Syndicate",
		Value:     keycardBlueprint.Syndicate.String,
	}

	err = assetJson.Marshal(inner)
	if err != nil {
		return terror.Error(err, "Failed to get marshal keycard attribute data")
	}

	keycardUpdate := &xsyn_rpcclient.Asset1155CountUpdateSupremacyReq{
		ApiKey:         mp.API.Passport.ApiKey,
		TokenID:        keycardBlueprint.KeycardTokenID,
		Address:        user.PublicAddress.String,
		CollectionSlug: keycardBlueprint.Collection,
		Amount:         1,
		ImageURL:       keycardBlueprint.ImageURL,
		AnimationURL:   keycardBlueprint.AnimationURL,
		KeycardGroup:   keycardBlueprint.KeycardGroup,
		Attributes:     assetJson,
		IsAdd:          false,
	}
	_, err = mp.API.Passport.UpdateKeycardCountXSYN(keycardUpdate)
	if err != nil {
		return terror.Error(err, "Failed to update XSYN asset count")
	}

	// // Process fee
	// balance := mp.API.Passport.UserBalanceGet(userID)

	// feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(10)).Mul(decimal.New(1, 18))
	// if req.Payload.ListingDurationHours > 24 {
	// 	listingDurationFee := (req.Payload.ListingDurationHours/24 - 1) * 5
	// 	feePrice = feePrice.Add(decimal.NewFromInt(int64(listingDurationFee)))
	// }

	// if balance.Sub(feePrice).LessThan(decimal.Zero) {
	// 	err = fmt.Errorf("insufficient funds")
	// 	gamelog.L.Error().
	// 		Str("user_id", user.ID).
	// 		Str("balance", balance.String()).
	// 		Str("item_id", req.Payload.ItemID.String()).
	// 		Err(err).
	// 		Msg("Player does not have enough sups.")
	// 	return terror.Error(err, "You do not have enough sups to list item.")
	// }

	// // Pay sup
	// txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
	// 	FromUserID:           userID,
	// 	ToUserID:             uuid.Must(uuid.FromString(server.SupremacyChallengeFundUserID)), // NOTE: send fees to challenge fund for now. (was faction account)
	// 	Amount:               feePrice.String(),
	// 	TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|keycard|%s|%d", req.Payload.ItemID.String(), time.Now().UnixNano())),
	// 	Group:                string(server.TransactionGroupSupremacy),
	// 	SubGroup:             string(server.TransactionGroupMarketplace),
	// 	Description:          fmt.Sprintf("Marketplace List Item Fee: %s (keycard)", req.Payload.ItemID.String()),
	// })
	// if err != nil {
	// 	gamelog.L.Error().
	// 		Str("user_id", user.ID).
	// 		Str("balance", balance.String()).
	// 		Str("item_id", req.Payload.ItemID.String()).
	// 		Err(err).
	// 		Msg("Failed to process transaction for Marketplace Fee.")
	// 	err = fmt.Errorf("failed to process marketplace fee transaction")
	// 	return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	// }

	// // trigger challenge fund update
	// defer func() {
	// 	mp.API.ArenaManager.ChallengeFundUpdateChan <- true
	// }()

	// Start transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to start db transaction (add player keycard sale item listing)")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Deduct Keycard Count
	err = db.DecrementPlayerKeycardCount(tx, req.Payload.ItemID)
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	// Create Sales Item
	endAt := time.Now().Add(time.Hour * req.Payload.ListingDurationHours)
	obj, err := db.MarketplaceKeycardSaleCreate(
		tx,
		userID,
		factionID,
		null.String{},
		endAt,
		req.Payload.ItemID,
		req.Payload.AskingPrice,
	)
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		// mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	reply(obj)

	// Log Event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventCreated, user.ID, decimal.NullDecimal{}, obj.ID, boiler.TableNames.ItemKeycardSales)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to log create sale item event (keycards).")
	}

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

	l := gamelog.L.With().
		Str("user_id", user.ID).
		Str("item_sale_id", req.Payload.ID.String()).Logger()

	// Check whether user can cancel sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		l.Error().Err(err).Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if saleItem.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user.")
	}
	if saleItem.SoldTo.ID.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}

	// Refund Item if auction
	if saleItem.Auction && saleItem.LastBid.ID.Valid {
		lastBid, err := boiler.ItemSalesBidHistories(
			qm.Select(
				boiler.ItemSalesBidHistoryColumns.BidderID,
				boiler.ItemSalesBidHistoryColumns.BidTXID,
				boiler.ItemSalesBidHistoryColumns.BidPrice,
			),
			boiler.ItemSalesBidHistoryWhere.ItemSaleID.EQ(saleItem.ID),
			boiler.ItemSalesBidHistoryWhere.CancelledAt.IsNull(),
			qm.Load(boiler.ItemSalesBidHistoryRels.Bidder),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return terror.Error(terror.ErrUnauthorised, "Unable to check last auction bid.")
		}
		if lastBid != nil {
			if lastBid.R == nil || lastBid.R.Bidder == nil {
				l.Error().Err(err).Str("bidTID", lastBid.BidTXID).Msg("unable to get find faction account")
				return terror.Error(fmt.Errorf("unable to find bidder's faction"), errMsg)
			}
			factionAccountID, ok := server.FactionUsers[lastBid.R.Bidder.FactionID.String]
			if !ok {
				l.Error().Err(err).Str("bidTID", lastBid.BidTXID).Msg("unable to get find faction account")
				return terror.Error(err, errMsg)
			}
			factID := uuid.Must(uuid.FromString(factionAccountID))
			syndicateBalance := mp.API.Passport.UserBalanceGet(factID)
			if syndicateBalance.LessThanOrEqual(lastBid.BidPrice) {
				txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
					ToUserID:             factID,
					Amount:               lastBid.BidPrice.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("bid_refunds|%s|%d", lastBid.BidderID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupSupremacy),
					SubGroup:             string(server.TransactionGroupMarketplace),
					Description:          fmt.Sprintf("Bid Refund for Player: %s (item sale: %s)", lastBid.BidderID, saleItem.ID),
				})
				if err != nil {
					l.Error().
						Str("Faction ID", factionAccountID).
						Str("Amount", lastBid.BidPrice.StringFixed(0)).
						Err(err).
						Msg("Could not transfer money from treasury into syndicate account!!")
					return terror.Error(err, errMsg)
				}
				l.Warn().
					Str("Faction ID", factionAccountID).
					Str("Amount", lastBid.BidPrice.StringFixed(0)).
					Str("TXID", txid).
					Err(err).
					Msg("Had to transfer funds to the syndicate account")
			}

			rtxid, err := mp.API.Passport.RefundSupsMessage(lastBid.BidTXID)
			if err != nil {
				return terror.Error(err, errMsg)
			}
			err = db.MarketplaceSaleBidHistoryRefund(gamedb.StdConn, req.Payload.ID, lastBid.BidTXID, rtxid, true)
			if err != nil {
				return terror.Error(err, errMsg)
			}
			err = db.MarketplaceAddEvent(boiler.MarketplaceEventBidRefund, lastBid.BidderID, decimal.NewNullDecimal(lastBid.BidPrice), saleItem.ID, boiler.TableNames.ItemSales)
			if err != nil {
				l.Error().
					Str("txid", lastBid.BidTXID).
					Str("refund_tx_id", rtxid).
					Err(err).
					Msg("Failed to log bid refund event.")
			}
		}
	}

	// Cancel item
	err = db.MarketplaceSaleArchive(gamedb.StdConn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	err = db.MarketplaceSaleItemUnlock(gamedb.StdConn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	// Log Event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventCancelled, user.ID, decimal.NullDecimal{}, req.Payload.ID.String(), boiler.TableNames.ItemSales)
	if err != nil {
		l.Error().
			Err(err).
			Msg("Failed to log cancelled sale item event.")
	}

	// Broadcast Queue Status Market
	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ID.EQ(saleItem.CollectionItemID),
	).One(gamedb.StdConn)
	if err != nil {
		l.Error().Str("collection item id", saleItem.CollectionItemID).Err(err).Msg("Failed to get collection item from db")
	}

	if ci.ItemType == boiler.ItemTypeMech {
		mp.API.ArenaManager.MechDebounceBroadcastChan <- []string{ci.ItemID}
	}

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
	if saleItem.SoldTo.ID.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}

	playerKeycardID, err := uuid.FromString(saleItem.ItemID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Cancel item
	err = db.MarketplaceKeycardSaleArchive(gamedb.StdConn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	keycardBlueprint, err := boiler.BlueprintKeycards(boiler.BlueprintKeycardWhere.ID.EQ(saleItem.Keycard.ID)).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get blueprint keycard")
	}

	// Return Keycard Counts (Game Server)
	err = db.IncrementPlayerKeycardCount(gamedb.StdConn, playerKeycardID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Return Keycard Counts (Xsyn)
	var assetJson types.JSON

	if !keycardBlueprint.Syndicate.Valid {
		keycardBlueprint.Syndicate.String = "N/A"
	}

	inner := &AttributeInner{
		TraitType: "Syndicate",
		Value:     keycardBlueprint.Syndicate.String,
	}

	err = assetJson.Marshal(inner)
	if err != nil {
		return terror.Error(err, "Failed to get marshal keycard attribute data")
	}
	keycardUpdate := &xsyn_rpcclient.Asset1155CountUpdateSupremacyReq{
		ApiKey:         mp.API.Passport.ApiKey,
		TokenID:        keycardBlueprint.KeycardTokenID,
		Address:        user.PublicAddress.String,
		CollectionSlug: keycardBlueprint.Collection,
		Amount:         1,
		ImageURL:       keycardBlueprint.ImageURL,
		AnimationURL:   keycardBlueprint.AnimationURL,
		KeycardGroup:   keycardBlueprint.KeycardGroup,
		Attributes:     assetJson,
		IsAdd:          true,
	}
	_, err = mp.API.Passport.UpdateKeycardCountXSYN(keycardUpdate)
	if err != nil {
		return terror.Error(err, "Failed to update XSYN asset count")
	}

	reply(true)

	// Log Event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventCancelled, user.ID, decimal.NullDecimal{}, req.Payload.ID.String(), boiler.TableNames.ItemKeycardSales)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ID.String()).
			Err(err).
			Msg("Failed to log cancelled sale item event (keycards).")
	}

	return nil
}

const (
	HubKeyMarketplaceSalesBuy        = "MARKETPLACE:SALES:BUY"
	HubKeyMarketplaceSalesKeycardBuy = "MARKETPLACE:SALES:KEYCARD:BUY"
)

type MarketplaceSalesBuyRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID     uuid.UUID       `json:"id"`
		Amount decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesBuyHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "SalesBuyHandler").Str("user_id", user.ID).Str("faction_id", fID).Logger()
	errMsg := "Issue buying sale item, try again or contact support."
	req := &MarketplaceSalesBuyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	l = l.With().Str("item_sale_id", req.Payload.ID.String()).Logger()

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		l.Error().Err(err).Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}

	l = l.With().Interface("saleItem", saleItem).Logger()

	if saleItem.FactionID != fID {
		l.Error().Err(terror.ErrUnauthorised).Msg("wrong faction id")
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user's faction.")
	}
	if saleItem.SoldTo.ID.Valid {
		l.Error().Err(fmt.Errorf("item is sold")).Msg("item already sold")
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}
	if saleItem.CollectionItem.XsynLocked || saleItem.CollectionItem.MarketLocked {
		err = fmt.Errorf("item is locked")
		l.Error().Err(err).Msg("item already sold")
		return terror.Error(err, "Item is no longer for sale.")
	}
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve buyer's user id")
		return terror.Error(err, errMsg)
	}

	// Calculate Cost depending on sale type
	saleType := "buyout"
	saleItemCost := saleItem.BuyoutPrice.Decimal
	if saleItem.DutchAuction {
		saleType = "dutch_auction"
		if !saleItem.DutchAuctionDropRate.Valid {
			err = fmt.Errorf("dutch auction drop rate is missing")
			l.Error().Err(err).Msg("dutch auction drop rate is missing")
			return terror.Error(err, errMsg)
		}
		minutesLapse := decimal.NewFromFloat(math.Floor(time.Now().Sub(saleItem.CreatedAt).Minutes()))
		dutchAuctionAmount := saleItem.BuyoutPrice.Decimal.Sub(saleItem.DutchAuctionDropRate.Decimal.Mul(minutesLapse))
		if saleItem.AuctionReservedPrice.Valid {
			if dutchAuctionAmount.GreaterThanOrEqual(saleItem.AuctionReservedPrice.Decimal) {
				saleItemCost = dutchAuctionAmount
			} else {
				saleItemCost = saleItem.AuctionReservedPrice.Decimal
			}
		} else {
			if dutchAuctionAmount.LessThanOrEqual(decimal.Zero) {
				saleItemCost = decimal.New(1, 18)
			} else {
				saleItemCost = dutchAuctionAmount
			}
		}
	}

	l = l.With().Str("saleItemCost", saleItemCost.String()).Logger()

	if !saleItemCost.Equal(req.Payload.Amount.Mul(decimal.New(1, 18))) {
		err = fmt.Errorf("amount does not match current price")
		l.Error().Err(err).Msg("amount does not match current price")
		return terror.Error(err, "Prices do not match up, please try again.")
	}

	salesCutPercentageFee := db.GetDecimalWithDefault(db.KeyMarketplaceSaleCutPercentageFee, decimal.NewFromFloat(0.1))

	balance := mp.API.Passport.UserBalanceGet(userID)
	l = l.With().Str("userBalance", balance.String()).Logger()
	if balance.Sub(saleItemCost).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		l.Error().Err(err).Msg("player does not have enough sups")
		return terror.Error(err, "You do not have enough sups.")
	}

	// Pay sales cut fee amount to faction account
	//factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	//if !ok {
	//	err = fmt.Errorf("failed to get hard coded syndicate player id")
	//	l.Error().Err(err).Msg("unable to get hard coded syndicate player ID from faction ID")
	//	return terror.Error(err, errMsg)
	//}
	feeTXID, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(server.SupremacyChallengeFundUserID)), // NOTE: send fees to challenge fund for now. (was faction account)
		Amount:               saleItemCost.Mul(salesCutPercentageFee).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item_fee:%s|%s|%d", saleType, saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Fee: %s", saleItem.ID),
	})
	if err != nil {
		err = fmt.Errorf("failed to process payment transaction")
		l.Error().Msg("failed to process sales cut fee transaction for purchase sale item")
		return terror.Error(err, errMsg)
	}

	// trigger challenge fund update
	defer func() {
		mp.API.ArenaManager.ChallengeFundUpdateChan <- true
	}()

	// Give sales cut amount to seller
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(saleItem.OwnerID)),
		Amount:               saleItemCost.Mul(decimal.NewFromInt(1).Sub(salesCutPercentageFee)).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item:%s|%s|%d", saleType, saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Payment (%d%% cut): %s", salesCutPercentageFee.Mul(decimal.NewFromInt(100)).IntPart(), saleItem.ID),
	})
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		err = fmt.Errorf("failed to process payment transaction")
		l.Error().Err(err).Msg("failed to process transaction for purchase sale item")
		return terror.Error(err, errMsg)
	}

	rpcAssetTransferRollback, err := marketplace.TransferAssetsToXsyn(gamedb.StdConn, mp.API.Passport, saleItem.OwnerID, userID.String(), txid, saleItem.CollectionItem.Hash, saleItem.ID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		l.Error().Msg("Failed to start purchase sale item rpc TransferAsset.")
		return terror.Error(err, errMsg)
	}

	// Start transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		rpcAssetTransferRollback()
		l.Error().Msg("Failed to start purchase sale item db transaction.")
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
		SoldTo:      null.StringFrom(user.ID),
		UpdatedAt:   time.Now(),
	}
	_, err = saleItemRecord.Update(tx,
		boil.Whitelist(
			boiler.ItemSaleColumns.SoldAt,
			boiler.ItemSaleColumns.SoldFor,
			boiler.ItemSaleColumns.SoldTXID,
			boiler.ItemSaleColumns.SoldFeeTXID,
			boiler.ItemSaleColumns.SoldTo,
			boiler.ItemSaleColumns.UpdatedAt,
		))
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		rpcAssetTransferRollback()
		err = fmt.Errorf("failed to complete payment transaction")
		l.Error().Err(err).Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, errMsg)
	}

	err = marketplace.HandleMarketplaceAssetTransfer(tx, mp.API.Passport, req.Payload.ID.String())
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		rpcAssetTransferRollback()
		l.Error().Err(err).Msg("Failed to Transfer Mech to New Owner")
		return terror.Error(err, errMsg)
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
		rpcAssetTransferRollback()
		err = fmt.Errorf("failed to complete payment transaction")
		l.Error().Err(err).Msg("Failed to unlock marketplace listed collection item.")
		return terror.Error(err, errMsg)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		rpcAssetTransferRollback()
		l.Error().Err(err).Msg("Failed to commit purchase sale item db transaction.")
		return terror.Error(err, errMsg)
	}

	// Log event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventPurchase, user.ID, decimal.NewNullDecimal(saleItemCost), saleItem.ID, boiler.TableNames.ItemSales)
	if err != nil {
		l.Error().Err(err).Msg("failed to log purchase event")
	}
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventSold, saleItem.OwnerID, decimal.NewNullDecimal(saleItemCost), saleItem.ID, boiler.TableNames.ItemSales)
	if err != nil {
		l.Error().Err(err).Msg("failed to log sold event")
	}

	// Refund bids
	bids, err := db.MarketplaceSaleCancelBids(gamedb.StdConn, uuid.Must(uuid.FromString(saleItem.ID)), "Item bought out")
	if err != nil {
		l.Error().Err(err).Msg("marketplace sale cancel bids error refunding bids")
		return err
	}
	for _, b := range bids {
		factionAccountID, ok := server.FactionUsers[b.FactionID.String]
		if !ok {
			l.Error().Err(err).Str("bidTID", b.TXID).Msg("unable to get find faction account")
		}
		factID := uuid.Must(uuid.FromString(factionAccountID))
		syndicateBalance := mp.API.Passport.UserBalanceGet(factID)
		if syndicateBalance.LessThanOrEqual(b.Amount) {
			txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
				ToUserID:             factID,
				Amount:               b.Amount.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("bid_refunds|%s|%d", b.BidderID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupMarketplace),
				Description:          fmt.Sprintf("Bid Refund for Player ID: %s (item sale: %s)", b.BidderID, saleItem.ID),
			})
			if err != nil {
				gamelog.L.Error().
					Str("Faction ID", factionAccountID).
					Str("Amount", b.Amount.StringFixed(0)).
					Err(err).
					Msg("Could not transfer money from treasury into syndicate account!!")
				continue
			}
			gamelog.L.Warn().
				Str("Faction ID", factionAccountID).
				Str("Amount", b.Amount.StringFixed(0)).
				Str("TXID", txid).
				Err(err).
				Msg("Had to transfer funds to the syndicate account")
		}

		_, err = mp.API.Passport.RefundSupsMessage(b.TXID)
		if err != nil {
			l.Error().Str("txID", b.TXID).Err(err).Msg("error refunding bids")
		}
		err = db.MarketplaceAddEvent(boiler.MarketplaceEventBidRefund, b.BidderID, decimal.NewNullDecimal(b.Amount), saleItem.ID, boiler.TableNames.ItemSales)
		if err != nil {
			l.Error().Str("txID", b.TXID).Err(err).Msg("failed to log bid refund event")
		}
	}

	// broadcast status change if item is a mech
	if saleItem.CollectionItemType == boiler.ItemTypeMech {
		ci, err := boiler.CollectionItems(
			boiler.CollectionItemWhere.ID.EQ(saleItem.CollectionItemID),
		).One(gamedb.StdConn)
		if err != nil {
			l.Error().Str("collection item id", saleItem.CollectionItemID).Err(err).Msg("failed to get collection item from db")
		}

		if ci != nil {
			mp.API.ArenaManager.MechDebounceBroadcastChan <- []string{ci.ItemID}

			ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", saleItem.FactionID, ci.ItemID), server.HubKeyPlayerAssetMechQueueSubscribe, &server.MechArenaInfo{
				Status: server.MechArenaStatusSold,
			})
		}
	}

	// success
	reply(true)
	return nil
}

func (mp *MarketplaceController) SalesKeycardBuyHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "SalesKeycardBuyHandler").Str("user_id", user.ID).Str("faction_id", fID).Logger()

	errMsg := "Issue buying sale item, try again or contact support."
	req := &MarketplaceSalesBuyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	l = l.With().Str("item_sale_id", req.Payload.ID.String()).Logger()

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemKeycardSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve sale item")
		return terror.Error(err, errMsg)
	}
	if saleItem.SoldTo.ID.Valid {
		return terror.Error(fmt.Errorf("item is sold"), "Item has already being sold.")
	}

	// Pay item
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("Unable to retrieve buyer's user id.")
		return terror.Error(err, errMsg)
	}

	saleItemCost := saleItem.BuyoutPrice
	if !saleItemCost.Equal(req.Payload.Amount.Mul(decimal.New(1, 18))) {
		return terror.Error(fmt.Errorf("amount does not match current price"), "Prices do not match up, please try again.")
	}

	l = l.With().Str("saleItemCost", saleItemCost.String()).Logger()

	balance := mp.API.Passport.UserBalanceGet(userID)
	l = l.With().Str("balance", balance.String()).Logger()
	if balance.Sub(saleItemCost).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		l.Warn().Err(err).Msg("player does not have enough sups")
		return terror.Error(err, "You do not have enough sups.")
	}

	salesCutPercentageFee := db.GetDecimalWithDefault(db.KeyMarketplaceSaleCutPercentageFee, decimal.NewFromFloat(0.1))

	// Pay sales cut fee amount to faction account
	//factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	//if !ok {
	//	err = fmt.Errorf("failed to get hard coded syndicate player id")
	//	l.Error().Err(err).Msg("unable to get hard coded syndicate player ID from faction ID")
	//	return terror.Error(err, errMsg)
	//}
	feeTXID, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(server.SupremacyChallengeFundUserID)), // NOTE: send fees to challenge fund for now. (was faction account)
		Amount:               saleItemCost.Mul(salesCutPercentageFee).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item_fee:buyout|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Buy Item Fee: %s", saleItem.ID),
	})
	if err != nil {
		err = fmt.Errorf("failed to process payment transaction")
		l.Error().Err(err).Msg("failed to process sales cut fee transaction for purchase sale item")
		return terror.Error(err, errMsg)
	}

	// trigger challenge fund update
	defer func() {
		mp.API.ArenaManager.ChallengeFundUpdateChan <- true
	}()

	keycardBlueprint, err := boiler.BlueprintKeycards(boiler.BlueprintKeycardWhere.ID.EQ(saleItem.Keycard.ID)).One(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to get blueprint keycard")
		return terror.Error(err, "Failed to get blueprint keycard")
	}

	var assetJson types.JSON

	if !keycardBlueprint.Syndicate.Valid {
		keycardBlueprint.Syndicate.String = "N/A"
	}

	inner := &AttributeInner{
		TraitType: "Syndicate",
		Value:     keycardBlueprint.Syndicate.String,
	}

	err = assetJson.Marshal(inner)
	if err != nil {
		l.Error().Err(err).Msg("failed to marshall keycard attributes")
		return terror.Error(err, "Failed to get marshal keycard attribute data")
	}
	keycardUpdate := &xsyn_rpcclient.Asset1155CountUpdateSupremacyReq{
		ApiKey:         mp.API.Passport.ApiKey,
		TokenID:        keycardBlueprint.KeycardTokenID,
		Address:        user.PublicAddress.String,
		CollectionSlug: keycardBlueprint.Collection,
		Amount:         1,
		ImageURL:       keycardBlueprint.ImageURL,
		AnimationURL:   keycardBlueprint.AnimationURL,
		KeycardGroup:   keycardBlueprint.KeycardGroup,
		Attributes:     assetJson,
		IsAdd:          true,
	}
	_, err = mp.API.Passport.UpdateKeycardCountXSYN(keycardUpdate)
	if err != nil {
		l.Error().Err(err).Msg("failed to update xsyn count")
		return terror.Error(err, "Failed to update XSYN asset count")
	}

	removeKeycardFunc := func() {
		_, err := mp.API.Passport.UpdateKeycardCountXSYN(&xsyn_rpcclient.Asset1155CountUpdateSupremacyReq{
			ApiKey:         mp.API.Passport.ApiKey,
			TokenID:        keycardUpdate.TokenID,
			Address:        user.PublicAddress.String,
			CollectionSlug: keycardUpdate.CollectionSlug,
			Amount:         keycardUpdate.Amount,
			ImageURL:       keycardUpdate.ImageURL,
			AnimationURL:   keycardUpdate.AnimationURL,
			KeycardGroup:   keycardUpdate.KeycardGroup,
			Attributes:     keycardUpdate.Attributes,
			IsAdd:          false,
		})
		if err != nil {
			l.Error().Err(err).Interface("keycardUpdate", keycardUpdate).Msg("retract of keycard failed")
		}
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
	})
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		removeKeycardFunc()
		err = fmt.Errorf("failed to process payment transaction")
		l.Error().Err(err).Msg("failed to process transaction for purchase sale item")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// Begin transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		removeKeycardFunc()
		l.Error().Err(err).Msg("failed to start purchase sale item db transaction")
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
		SoldTo:      null.StringFrom(user.ID),
	}

	_, err = saleItemRecord.Update(tx, boil.Whitelist(
		boiler.ItemKeycardSaleColumns.SoldAt,
		boiler.ItemKeycardSaleColumns.SoldFor,
		boiler.ItemKeycardSaleColumns.SoldTXID,
		boiler.ItemKeycardSaleColumns.SoldFeeTXID,
		boiler.ItemKeycardSaleColumns.SoldTo,
	))
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		err = fmt.Errorf("failed to complete payment transaction")
		l.Error().Err(err).Msg("failed to update to keycard sale item")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// Transfer ownership of asset
	err = db.ChangeKeycardOwner(tx, req.Payload.ID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		removeKeycardFunc()
		l.Error().Err(err).Msg("failed to Transfer keycard to new owner")
		return terror.Error(err, "Failed to process transaction for Purchase Sale Item.")
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(feeTXID)
		mp.API.Passport.RefundSupsMessage(txid)
		removeKeycardFunc()
		l.Error().Err(err).Msg("failed to commit purchase sale item db transaction")
		return terror.Error(err, "Failed to process transaction for Purchase Sale Item.")
	}

	// Log Event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventPurchase, user.ID, decimal.NewNullDecimal(saleItemCost), saleItem.ID, boiler.TableNames.ItemKeycardSales)
	if err != nil {
		l.Error().Err(err).Msg("failed to log purchase event")
	}
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventSold, saleItem.OwnerID, decimal.NewNullDecimal(saleItemCost), saleItem.ID, boiler.TableNames.ItemKeycardSales)
	if err != nil {
		l.Error().Err(err).Msg("failed to log sold event")
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
	l := gamelog.L.With().Str("func", "SalesBidHandler").Str("user_id", user.ID).Str("faction_id", fID).Logger()

	errMsg := "Issue placing bid, try again or contact support."
	req := &MarketplaceSalesBidRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse user id to uuid")
		return terror.Error(err, errMsg)
	}

	// Get Faction Account sending bid amount to
	factionAccountID, ok := server.FactionUsers[fID]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		l.Error().Err(err).Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}

	l = l.With().Str("item_sale_id", req.Payload.ID.String()).Logger()

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("unable to retrieve sale item")
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve sale item")
		return terror.Error(err, errMsg)
	}

	l = l.With().Interface("saleItem", saleItem).Logger()

	if !saleItem.Auction {
		err = fmt.Errorf("item is not up for auction")
		l.Error().Err(err).Msg("not a valid auction")
		return terror.Error(err, "Item is not up for auction.")
	}
	if saleItem.SoldTo.ID.Valid {
		err = fmt.Errorf("item is sold")
		l.Warn().Err(err).Msg("item already sold")
		return terror.Error(err, "Item has already being sold.")
	}
	if saleItem.FactionID != fID {
		err = fmt.Errorf("item does not belong to users faction")
		l.Error().Err(err).Msg("item does not belong to users faction")
		return terror.Error(err, "Item does not belong to user's faction.")
	}
	if saleItem.CollectionItem.XsynLocked || saleItem.CollectionItem.MarketLocked {
		err = fmt.Errorf("item is locked")
		l.Error().Err(err).Bool("xsynLocked", saleItem.CollectionItem.XsynLocked).Bool("marketLocked", saleItem.CollectionItem.MarketLocked).Msg("item is locked")
		return terror.Error(err, "Item is no longer for sale.")
	}
	bidAmount := req.Payload.Amount.Mul(decimal.New(1, 18))
	if bidAmount.LessThanOrEqual(saleItem.AuctionCurrentPrice.Decimal) {
		err = fmt.Errorf("bid amount less than current bid amount")
		l.Warn().Err(err).Str("bidAmount", bidAmount.String()).Msg("bid amount less than current bid amount")
		return terror.Error(err, "Invalid bid amount, must be above the current bid price.")
	}

	// Check if bid amount is greater than Dutch Auction drop rate
	if saleItem.DutchAuction {
		if !saleItem.DutchAuctionDropRate.Valid {
			err = fmt.Errorf("dutch auction drop rate is missing")
			l.Error().Err(err).Msg("dutch auction drop rate is missing")
			return terror.Error(err, errMsg)
		}
		minutesLapse := decimal.NewFromFloat(math.Floor(time.Now().Sub(saleItem.CreatedAt).Minutes()))
		dutchAuctionAmount := saleItem.BuyoutPrice.Decimal.Sub(saleItem.DutchAuctionDropRate.Decimal.Mul(minutesLapse))
		if saleItem.AuctionReservedPrice.Valid {
			if dutchAuctionAmount.LessThan(saleItem.AuctionReservedPrice.Decimal) {
				dutchAuctionAmount = saleItem.AuctionReservedPrice.Decimal
			}
		} else {
			if dutchAuctionAmount.LessThanOrEqual(decimal.Zero) {
				dutchAuctionAmount = decimal.New(1, 18)
			}
		}
		if dutchAuctionAmount.LessThanOrEqual(bidAmount) {
			return terror.Error(fmt.Errorf("bid amount is less than dutch auction dropped price"), "Bid Amount is cheaper than Dutch Auction Dropped Price, buy the item instead.")
		}
	}
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               bidAmount.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item:auction_bid|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupMarketplace),
		Description:          fmt.Sprintf("Marketplace Bid Item: %s", saleItem.ID),
	})
	if err != nil {
		l.Error().Err(err).Msg("payment failed")
		return terror.Error(err, "Issue making bid transaction.")
	}

	// Start Transaction
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		l.Error().Err(err).Msg("failed to start tx")
		return terror.Error(err, errMsg)
	}
	defer tx.Rollback()

	// Cancel all other bids before placing in the next new bid
	refundBids, err := db.MarketplaceSaleCancelBids(tx, req.Payload.ID, "New Bid")
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		l.Error().Err(err).Msg("failed to cancel previous bids")
		return terror.Error(err, errMsg)
	}

	// Place Bid
	_, err = db.MarketplaceSaleBidHistoryCreate(tx, req.Payload.ID, userID, req.Payload.Amount, txid)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		l.Error().Err(err).Msg("unable to place bid")
		return terror.Error(err, errMsg)
	}

	err = db.MarketplaceSaleAuctionSync(tx, req.Payload.ID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		l.Error().Err(err).Msg("unable to update current auction price")
		return terror.Error(err, errMsg)
	}

	// Refund other bids
	for _, b := range refundBids {
		factionAccountID, ok := server.FactionUsers[b.FactionID.String]
		if !ok {
			l.Error().Err(err).Str("bidTID", b.TXID).Msg("unable to get find faction account")
		}
		factID := uuid.Must(uuid.FromString(factionAccountID))
		syndicateBalance := mp.API.Passport.UserBalanceGet(factID)
		if syndicateBalance.LessThanOrEqual(b.Amount) {
			txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
				ToUserID:             factID,
				Amount:               b.Amount.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("bid_refunds|%s|%d", b.BidderID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupMarketplace),
				Description:          fmt.Sprintf("Bid Refund for Player ID: %s (item sale: %s)", b.BidderID, saleItem.ID),
			})
			if err != nil {
				gamelog.L.Error().
					Str("Faction ID", factionAccountID).
					Str("Amount", b.Amount.StringFixed(0)).
					Err(err).
					Msg("Could not transfer money from treasury into syndicate account!!")
				continue
			}
			gamelog.L.Warn().
				Str("Faction ID", factionAccountID).
				Str("Amount", b.Amount.StringFixed(0)).
				Str("TXID", txid).
				Err(err).
				Msg("Had to transfer funds to the syndicate account")
		}

		refundTxID, err := mp.API.Passport.RefundSupsMessage(b.TXID)
		if err != nil {
			l.Error().Err(err).Str("bidTID", b.TXID).Msg("unable to refund cancelled bid")
			continue
		}
		err = db.MarketplaceSaleBidHistoryRefund(tx, req.Payload.ID, b.TXID, refundTxID, false)
		if err != nil {
			l.Error().Err(err).Str("bidTID", b.TXID).Str("refundTxID", refundTxID).Msg("unable to update cancelled bid refund tx id")
			continue
		}
		err = db.MarketplaceAddEvent(boiler.MarketplaceEventBidRefund, b.BidderID, decimal.NewNullDecimal(b.Amount), saleItem.ID, boiler.TableNames.ItemSales)
		if err != nil {
			l.Error().Err(err).Str("bidTID", b.TXID).Str("refundTxID", refundTxID).Msg("failed to log bid refund event")
		}
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		l.Error().Err(err).Msg("unable to update current auction price")
		return terror.Error(err, errMsg)
	}

	reply(true)

	// Broadcast new current price
	totalBids, err := boiler.ItemSalesBidHistories(boiler.ItemSalesBidHistoryWhere.ItemSaleID.EQ(req.Payload.ID.String())).Count(gamedb.StdConn)
	if err != nil {
		// No need to abort failure
		l.Error().Err(err).Msg("unable to get current total bids")
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

	// Log Event
	err = db.MarketplaceAddEvent(boiler.MarketplaceEventBid, user.ID, decimal.NewNullDecimal(req.Payload.Amount.Mul(decimal.New(1, 18))), saleItem.ID, boiler.TableNames.ItemSales)
	if err != nil {
		l.Error().Err(err).Msg("failed to log bid event")
	}

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
