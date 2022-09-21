package fiat

import (
	"fmt"

	"github.com/ninja-software/terror/v2"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
)

// StripeProducts grabs a list of products from given product ids.
func StripeProducts(stripeClient *client.API, ids ...*string) ([]*stripe.Product, []*stripe.Price, error) {
	prices := []*stripe.Price{}
	products := []*stripe.Product{}
	params := &stripe.ProductListParams{
		IDs: ids,
	}
	productIterator := stripeClient.Products.List(params)
	for productIterator.Next() {
		product := productIterator.Product()
		products = append(products, product)

		price, err := StripeProductPlanPrice(stripeClient, product.ID, stripe.String(product.DefaultPrice.ID))
		if err != nil {
			return nil, nil, terror.Error(err, "get price failed")
		}
		prices = append(prices, price)
	}
	err := productIterator.Err()
	if err != nil {
		return nil, nil, terror.Error(err, "get product failed")
	}

	return products, prices, nil
}

// ProductPlanPrice grabs the active price from specific product.
// Optionally, you can specify which price to use instead.
func StripeProductPlanPrice(stripeClient *client.API, id string, priceID *string) (*stripe.Price, error) {
	// Find Specific Price if given
	if priceID != nil {
		price, err := stripeClient.Prices.Get(*priceID, nil)
		if err != nil {
			return nil, terror.Error(err, "get price failed")
		}
		if price.Product.ID != id {
			return nil, terror.Error(fmt.Errorf("price not found"), "Price Not Found")
		}
		return price, nil
	}

	// Find Active Price Instead
	priceParams := &stripe.PriceListParams{
		Active: stripe.Bool(true),
	}
	priceParams.Filters.AddFilter("product", "", id)
	priceIterator := stripeClient.Prices.List(priceParams)

	if !priceIterator.Next() {
		err := priceIterator.Err()
		if err != nil {
			return nil, terror.Error(err, "get price failed")
		}
	}

	price := priceIterator.Price()

	return price, nil
}

// StripeProductPlan loads a specific product from Stripe API
func StripeProductPlan(stripeClient *client.API, id string) (*stripe.Product, *stripe.Price, error) {
	product, err := stripeClient.Products.Get(id, nil)
	if err != nil {
		return nil, nil, terror.Error(err, "get product failed")
	}

	price, err := StripeProductPlanPrice(stripeClient, product.ID, nil)
	if err != nil {
		return nil, nil, terror.Error(err, "get price failed")
	}

	return product, price, nil
}

// StripeChargeList grabs list of charges from customer and payment intent.
func StripeChargeList(stripeClient *client.API, paymentIntentID string) ([]*stripe.Charge, error) {
	chargeListParams := &stripe.ChargeListParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	chargeIterator := stripeClient.Charges.List(chargeListParams)

	output := []*stripe.Charge{}
	for chargeIterator.Next() {
		output = append(output, chargeIterator.Charge())
	}
	err := chargeIterator.Err()
	if err != nil {
		return nil, terror.Error(err, "get charge failed")
	}

	return output, nil
}

// StripePaymentIntentGet retrieves a specific payment.
func StripePaymentIntentGet(stripeClient *client.API, paymentIntentID string) (*stripe.PaymentIntent, error) {
	output, err := stripeClient.PaymentIntents.Get(paymentIntentID, nil)
	if err != nil {
		return nil, terror.Error(err, "get payment intent failed")
	}
	return output, nil
}

// StripeCheckoutSessionList retrieves checkout session list from given payment intent.
func StripeCheckoutSessionList(stripeClient *client.API, paymentIntentID string) ([]*stripe.CheckoutSession, error) {
	checkoutSessionListParams := &stripe.CheckoutSessionListParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	checkoutSessionListParams.AddExpand("data.line_items")
	checkoutSessionIterator := stripeClient.CheckoutSessions.List(checkoutSessionListParams)

	output := []*stripe.CheckoutSession{}
	for checkoutSessionIterator.Next() {
		output = append(output, checkoutSessionIterator.CheckoutSession())
	}
	err := checkoutSessionIterator.Err()
	if err != nil {
		return nil, terror.Error(err, "get checkout sessions failed")
	}

	return output, nil
}
