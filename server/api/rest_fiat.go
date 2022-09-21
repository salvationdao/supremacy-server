package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/fiat"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// StripeWebhook handles all event messages from stripe such as processin  paid invoices.
func (f *FiatController) StripeWebhook(w http.ResponseWriter, r *http.Request) (int, error) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("error reading request body")
		return http.StatusServiceUnavailable, terror.Error(err)
	}

	l := gamelog.L.With().Str("payload", string(payload)).Logger()

	// Secure webhook - https://dashboard.stripe.com/webhooks
	endpointSecret := f.API.StripeWebhookSecret
	signatureHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		l.Error().Err(err).Msg("webhook signature verification failed")
		return http.StatusOK, terror.Error(err)
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "invoice.paid":
		// Parse invoice
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			l.Error().Err(err).Msg("error parsing webhook JSON")
			return http.StatusBadRequest, terror.Error(err)
		}
		l = l.With().Str("invoice_id", invoice.ID).Logger()
		if invoice.Customer == nil {
			l.Error().Msg("stripe customer missing on successful paid invoice")
			return http.StatusBadRequest, terror.Error(fmt.Errorf("package type required"), "product type missing on invoice payload")
		}
		l = l.With().Str("customer_id", invoice.Customer.ID).Logger()

		// Associate User with Stripe Customer
		stripeCustomerID := invoice.Customer.ID
		userID, err := db.UserByStripeCustomer(gamedb.StdConn, stripeCustomerID)
		if err != nil {
			l.Error().Err(err).Msg("failed to find player")
			return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to find user"), "unable to find user associated with stripe customer account")
		}
		l = l.With().Str("user_id", userID).Logger()

		// Record Payment
		stripeFiatProducts := map[string]*server.FiatProduct{}

		order := &boiler.Order{
			UserID:        userID,
			OrderStatus:   boiler.OrderStatusesPending,
			PaymentMethod: boiler.PaymentMethodsStripe,
			TXNReference:  invoice.ID,
			Currency:      server.FiatCurrencyCodeUSD,
		}
		err = order.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("failed to record order")
			return http.StatusInternalServerError, terror.Error(err, "failed to record order")
		}

		for i, item := range invoice.Lines.Data {
			fiatProductID, ok := item.Metadata["fiat_product_id"]
			if !ok {
				l.Error().Err(err).Msg("failed to get product info")
				return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
			}
			product, ok := stripeFiatProducts[fiatProductID]
			if !ok {
				product, err = db.FiatProduct(gamedb.StdConn, fiatProductID)
				if err != nil {
					l.Error().Err(err).Msg("failed to get product info")
					return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
				}
				stripeFiatProducts[fiatProductID] = product
			}
			if product == nil {
				l.Error().Err(err).Msg("failed to find fiat product info")
				return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
			}

			orderItem := &boiler.OrderItem{
				OrderID:       order.ID,
				FiatProductID: product.ID,
				Name:          product.Name,
				Description:   product.Description,
				Quantity:      int(item.Quantity),
				Amount:        decimal.NewFromInt(item.Price.UnitAmount),
			}
			err = orderItem.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				l.Error().Err(err).Msg(fmt.Sprintf("failed to record order item %d", i))
				return http.StatusInternalServerError, terror.Error(err, "failed to record order item")
			}
		}

		// Start transaction
		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			l.Error().Err(err).Msg("failed to start transaction")
			return http.StatusInternalServerError, terror.Error(err, "Failed to give out items to user.")
		}
		defer tx.Rollback()

		// Handle giving out items to user
		for _, item := range invoice.Lines.Data {
			fiatProductID, ok := item.Metadata["fiat_product_id"]
			if !ok {
				l.Error().Err(err).Msg("failed to get product info")
				return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
			}
			product, ok := stripeFiatProducts[fiatProductID]
			if !ok {
				product, err = db.FiatProduct(gamedb.StdConn, fiatProductID)
				if err != nil {
					l.Error().Err(err).Msg("failed to get product info")
					return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
				}
				stripeFiatProducts[fiatProductID] = product
			}
			if product == nil {
				l.Error().Err(err).Msg("failed to find fiat product info")
				return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
			}

			if product.ProductType == boiler.FiatProductTypesStarterPackage {
				for i := 0; i < int(item.Quantity); i++ {
					if product.ProductType == server.FiatProductTypeGeneric {
						err = fiat.SendStarterPackageContentsToUser(tx, f.API.Passport, userID, product.ID)
						if err != nil {
							l.Error().Err(err).Msg("failed to send out package contents to user.")
							return http.StatusInternalServerError, terror.Error(err, "failed to give user the package contents")
						}
					}
				}
			} else if product.ProductType == boiler.FiatProductTypesMysteryCrate {
				err = fiat.SendMysteryCrateToUser(tx, f.API.Passport, userID, product.ID, int(item.Quantity))
				if err != nil {
					l.Error().Err(err).Msg("failed to send out mystery crate contents to user.")
					return http.StatusInternalServerError, terror.Error(err, "failed to give user the package contents")
				}
			}

			// TODO: Handle other product types :/
		}

		// Clear user's cart
		err = db.ShoppingCartDeleteByUser(tx, userID)
		if err != nil {
			l.Error().Err(err).Msg("failed to clear user's shopping cart")
			return http.StatusInternalServerError, terror.Error(err, "Failed to clear user's shopping cart.")
		}

		// Commit Transaction
		err = tx.Commit()
		if err != nil {
			l.Error().Err(err).Msg("failed to commit transaction")
			return http.StatusInternalServerError, terror.Error(err, "Failed to give out item to user.")
		}

		// Mark as paid
		order.OrderStatus = boiler.OrderStatusesCompleted
		order.UpdatedAt = time.Now()
		_, err = order.Update(gamedb.StdConn, boil.Whitelist(boiler.OrderColumns.OrderStatus, boiler.OrderColumns.UpdatedAt))
		if err != nil {
			l.Error().Err(err).Msg("failed to mark order as completed")
			return http.StatusInternalServerError, terror.Error(err, "Failed to order as completed.")
		}

		f.publishUpdatedCart(userID, nil)
	}

	return http.StatusOK, nil
}
