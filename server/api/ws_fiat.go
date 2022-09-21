package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/stripe/stripe-go/v72"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type FiatController struct {
	API *API
}

func NewFiatController(api *API) *FiatController {
	fiatHub := &FiatController{
		API: api,
	}

	api.SecureUserCommand(HubKeyFiatBillingHistoryList, fiatHub.BillingHistoryListHandler)
	api.SecureUserCommand(HubKeyFiatBillingHistoryGet, fiatHub.BillingHistoryGetHandler)
	api.SecureUserFactionCommand(HubKeyFiatProductGet, fiatHub.ProductGetHandler)
	api.SecureUserFactionCommand(HubKeyFiatProductList, fiatHub.ProductListHandler)

	api.SecureUserFactionCommand(HubKeyShoppingCartAddItem, fiatHub.ShoppingCartAddItemHandler)
	api.SecureUserCommand(HubKeyShoppingCartUpdateItem, fiatHub.ShoppingCartUpdateItemHandler)
	api.SecureUserCommand(HubKeyShoppingCartRemoveItem, fiatHub.ShoppingCartRemoveItemHandler)
	api.SecureUserCommand(HubKeyShoppingCartClear, fiatHub.ShoppingCartClearHandler)

	api.SecureUserCommand(HubKeyFiatCheckoutSetup, fiatHub.CheckoutSetupHandler)

	return fiatHub
}

type FiatProductGetRequest struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

const HubKeyFiatProductGet = "FIAT:PRODUCT:GET"

func (f *FiatController) ProductGetHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get packages, please try again."

	req := &FiatProductGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	product, err := db.FiatProduct(gamedb.StdConn, req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(fmt.Errorf("product not found"), "Product not found.")
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if product.FactionID != factionID {
		return terror.Error(fmt.Errorf("product does not belong to player's faction"), "Invalid Product ID received.")
	}

	reply(product)

	return nil
}

type FiatProductListRequest struct {
	Payload struct {
		ProductType string `json:"product_type"`
		Search      string `json:"search"`
		PageSize    int    `json:"page_size"`
		Page        int    `json:"page"`
	} `json:"payload"`
}

type FiatProductListResponse struct {
	Total   int64                 `json:"total"`
	Records []*server.FiatProduct `json:"records"`
}

const HubKeyFiatProductList = "FIAT:PRODUCT:LIST"

func (f *FiatController) ProductListHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get packages, please try again."

	req := &FiatProductListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, storePackages, err := db.FiatProducts(gamedb.StdConn, factionID, req.Payload.ProductType, offset, req.Payload.PageSize)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := &FiatProductListResponse{
		Total:   total,
		Records: storePackages,
	}
	reply(resp)

	return nil
}

type FiatBillingHistoryListRequest struct {
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   string                `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter,omitempty"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type FiatBillingHistoryListResponse struct {
	Total   int64               `json:"total"`
	Records []*server.FiatOrder `json:"records"`
}

const HubKeyFiatBillingHistoryList = "FIAT:BILLING_HISTORY:LIST"

func (f *FiatController) BillingHistoryListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Could not get list of past billing transactions, try again or contact support."
	req := &FiatBillingHistoryListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, items, err := db.FiatOrderList(gamedb.StdConn, user.ID, offset, req.Payload.PageSize)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get billing history list")
		return terror.Error(err, errMsg)
	}

	resp := &FiatBillingHistoryListResponse{
		Total:   total,
		Records: items,
	}
	reply(resp)

	return nil
}

type FiatBillingHistoryGetRequest struct {
	Payload struct {
		ID string `json:"id"`
	}
}

const HubKeyFiatBillingHistoryGet = "FIAT:BILLING_HISTORY:GET"

func (f *FiatController) BillingHistoryGetHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Could not get past billing transactions, try again or contact support."
	req := &FiatBillingHistoryGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	item, err := db.FiatOrder(gamedb.StdConn, req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Order is not found, try again or contact support.")
	}
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get billing history list")
		return terror.Error(err, errMsg)
	}

	reply(item)

	return nil
}

type FiatCheckoutSetupRequest struct {
	Payload struct {
		ReceiptEmail string `json:"receipt_email"`
	} `json:"payload"`
}

const HubKeyFiatCheckoutSetup = "FIAT:CHECKOUT:SETUP"

func (f *FiatController) CheckoutSetupHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Something went wrong with beginning checkout process."
	req := &FiatCheckoutSetupRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check shopping cart
	cart, cartItems, err := db.ShoppingCartGetByUser(gamedb.StdConn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	shoppingCart := server.ShoppingCartFromBoiler(cart, cartItems)

	// Setup Stripe Customer
	player, err := boiler.FindPlayer(gamedb.StdConn, user.ID)
	if err != nil {
		return terror.Error(err, "failed to load user")
	}

	if !player.StripeCustomerID.Valid {
		customerParams := &stripe.CustomerParams{}
		if req.Payload.ReceiptEmail != "" {
			customerParams.Email = stripe.String(req.Payload.ReceiptEmail)
		}
		customer, err := f.API.StripeClient.Customers.New(customerParams)
		if err != nil {
			return terror.Error(err, errMsg)
		}

		player.StripeCustomerID = null.StringFrom(customer.ID)
		_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.StripeCustomerID))
		if err != nil {
			return terror.Error(err, "failed to update user")
		}
	}

	// Prepare invoice
	for _, item := range shoppingCart.Items {
		if item.Product == nil {
			return terror.Error(fmt.Errorf("cart item is missing data"), errMsg)
		}

		// TODO: Handle different currencies?
		var priceUSD *server.FiatProductPricing
		for _, price := range item.Product.Pricing {
			if price.CurrencyCode == server.FiatCurrencyCodeUSD {
				priceUSD = price
				break
			}
		}
		if priceUSD == nil {
			return terror.Error(fmt.Errorf("unable to find USD pricing"), errMsg)
		}

		invoiceItemParams := &stripe.InvoiceItemParams{
			Customer:    stripe.String(player.StripeCustomerID.String),
			Currency:    stripe.String(priceUSD.CurrencyCode),
			UnitAmount:  stripe.Int64(priceUSD.Amount.IntPart()),
			Quantity:    stripe.Int64(int64(item.Quantity)),
			Description: stripe.String(item.Product.Name),
		}
		invoiceItemParams.AddMetadata("fiat_product_id", item.Product.ID)
		_, err = f.API.StripeClient.InvoiceItems.New(invoiceItemParams)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	invoiceParams := &stripe.InvoiceParams{
		Customer: stripe.String(player.StripeCustomerID.String),
	}
	invoice, err := f.API.StripeClient.Invoices.New(invoiceParams)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Setup Payment Intent
	invoice, err = f.API.StripeClient.Invoices.FinalizeInvoice(invoice.ID, nil)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	paymentIntent, err := f.API.StripeClient.PaymentIntents.Get(invoice.PaymentIntent.ID, nil)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(paymentIntent.ClientSecret)

	return nil
}

type ShoppingCartAddItemRequest struct {
	Payload struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"payload"`
}

const HubKeyShoppingCartAddItem = "FIAT:SHOPPING_CART:ITEM:ADD"

func (f *FiatController) ShoppingCartAddItemHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to update shopping cart, please try again."

	req := &ShoppingCartAddItemRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.ProductID == "" {
		return terror.Error(fmt.Errorf("product id required"), "Product ID is required.")
	}
	if req.Payload.Quantity <= 0 {
		return terror.Error(fmt.Errorf("quantity is required"), "Quantity is required.")
	}

	// Grab product
	product, err := db.FiatProduct(gamedb.StdConn, req.Payload.ProductID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Package not found, please try again.")
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if product.FactionID != factionID {
		return terror.Error(fmt.Errorf("product does not belong to player's faction"), "Invalid Product ID received.")
	}

	// Add to cart
	isNewCart := false
	cartItemID := ""
	cartExpiryAt := time.Now().Add(time.Minute * 30)
	cart, cartItems, err := db.ShoppingCartGetByUser(gamedb.StdConn, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			isNewCart = true
			cart, err = db.ShoppingCartCreate(gamedb.StdConn, user.ID, cartExpiryAt)
			if err != nil {
				return terror.Error(err, errMsg)
			}
		} else {
			return terror.Error(err, errMsg)
		}
	}
	for _, item := range cartItems {
		// TODO: Handle cart item attributes as unique eg. Shirt Sizes :/
		if item.ProductID == req.Payload.ProductID {
			cartItemID = item.ID
			break
		}
	}

	if cartItemID != "" {
		err = db.ShoppingCartItemQtyIncrementUpdate(gamedb.StdConn, cart.ID, cartItemID, req.Payload.Quantity)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	} else {
		err = db.ShoppingCartItemAdd(gamedb.StdConn, cart.ID, product.ID, req.Payload.Quantity)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	if !isNewCart {
		err = db.ShoppingCartUpdateExpiry(gamedb.StdConn, cart.ID, cartExpiryAt)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	reply(true)

	f.publishUpdatedCart(user.ID, nil)

	return nil
}

type ShoppingCartItemUpdateItemRequest struct {
	Payload struct {
		ID       string `json:"id"`
		Quantity int    `json:"quantity"`
	} `json:"payload"`
}

const HubKeyShoppingCartUpdateItem = "FIAT:SHOPPING_CART:ITEM:UPDATE"

func (f *FiatController) ShoppingCartUpdateItemHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to update shopping cart, please try again."

	req := &ShoppingCartItemUpdateItemRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.ID == "" {
		return terror.Error(fmt.Errorf("cart item id required"), "Item to update is required.")
	}
	if req.Payload.Quantity <= 0 {
		return terror.Error(fmt.Errorf("quantity is required"), "Quantity is required.")
	}

	cart, _, err := db.ShoppingCartGetByUser(gamedb.StdConn, user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(fmt.Errorf("shopping cart is empty"), "Shopping cart is empty.")
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}
	err = db.ShoppingCartItemQtyUpdate(gamedb.StdConn, cart.ID, req.Payload.ID, req.Payload.Quantity)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	err = db.ShoppingCartUpdateExpiry(gamedb.StdConn, cart.ID, time.Now().Add(time.Minute*30))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	f.publishUpdatedCart(user.ID, nil)

	return nil
}

type ShoppingCartItemRemoveItemRequest struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

const HubKeyShoppingCartRemoveItem = "FIAT:SHOPPING_CART:ITEM:REMOVE"

func (f *FiatController) ShoppingCartRemoveItemHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to update shopping cart, please try again."

	req := &ShoppingCartItemRemoveItemRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.ID == "" {
		return terror.Error(fmt.Errorf("cart item id required"), "Item to remove is required.")
	}

	cart, _, err := db.ShoppingCartGetByUser(gamedb.StdConn, user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(fmt.Errorf("shopping cart is empty"), "Shopping cart is empty.")
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}
	err = db.ShoppingCartItemDelete(gamedb.StdConn, cart.ID, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	err = db.ShoppingCartUpdateExpiry(gamedb.StdConn, cart.ID, time.Now().Add(time.Minute*30))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	f.publishUpdatedCart(user.ID, nil)

	return nil
}

const HubKeyShoppingCartClear = "FIAT:SHOPPING_CART:CLEAR"

func (f *FiatController) ShoppingCartClearHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to update shopping cart, please try again."

	cart, _, err := db.ShoppingCartGetByUser(gamedb.StdConn, user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(fmt.Errorf("shopping cart is empty"), "Shopping cart is empty.")
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}

	err = db.ShoppingCartDelete(gamedb.StdConn, cart.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/shopping_cart_updated", user.ID), server.HubKeyShoppingCartUpdated, nil)

	return nil
}

func (f *FiatController) publishUpdatedCart(userID string, reply ws.ReplyFunc) error {
	errMsg := "Unable to load user's shopping cart."
	cart, cartItems, err := db.ShoppingCartGetByUser(gamedb.StdConn, userID)
	if errors.Is(err, sql.ErrNoRows) {
		if reply != nil {
			reply(nil)
		} else {
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/shopping_cart_updated", userID), server.HubKeyShoppingCartUpdated, nil)
		}
		return nil
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if len(cartItems) == 0 {
		if reply != nil {
			reply(nil)
		} else {
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/shopping_cart_updated", userID), server.HubKeyShoppingCartUpdated, nil)
		}
		return nil
	}
	resp := server.ShoppingCartFromBoiler(cart, cartItems)
	if reply != nil {
		reply(resp)
	} else {
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/shopping_cart_updated", userID), server.HubKeyShoppingCartUpdated, resp)
	}

	return nil
}

func (f *FiatController) ShoppingCartUpdatedSubscriber(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	err := f.publishUpdatedCart(user.ID, reply)
	if err != nil {
		return terror.Error(err, "Failed to listen for any shopping cart updates.")
	}
	return nil
}
