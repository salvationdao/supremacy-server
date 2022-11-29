package fiat

import (
	"fmt"
	"server"
	"server/benchmark"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v72/client"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type FiatController struct {
	Passport     *xsyn_rpcclient.XsynXrpcClient
	StripeClient *client.API
}

func NewFiatController(pp *xsyn_rpcclient.XsynXrpcClient, sc *client.API) *FiatController {
	f := &FiatController{pp, sc}
	go f.Run()
	return f
}

func (f *FiatController) Run() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the FiatController!", r)
		}
	}()

	mainTicker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-mainTicker.C:
			bm := benchmark.New()

			bm.Start("update_storefront_sup_prices")
			f.processStorefrontSupPrices()
			bm.End("update_storefront_sup_prices")
			bm.Start("gc_shopping_cart")
			f.processShoppingCartGarbageCollection()
			bm.End("gc_shopping_cart")
			bm.Alert(60000)
		}
	}
}

func (f *FiatController) processStorefrontSupPrices() {
	l := gamelog.L.With().Str("task", "processStorefrontSupPrices").Logger()
	currentRates, err := f.Passport.GetCurrentRates()
	if err != nil {
		l.Error().Err(err).Msg("failed to get current sup price")
		return
	}

	fiatToSupConversionCut := db.GetDecimalWithDefault(db.KeyFiatToSUPCut, decimal.NewFromInt(1).Div(decimal.NewFromInt(5))) // 20% by default
	l = l.With().
		Str("sup_to_usd_rate", currentRates.SUPtoUSD.String()).
		Str("eth_to_usd_rate", currentRates.ETHtoUSD.String()).
		Str("bnb_to_usd_rate", currentRates.BNBtoUSD.String()).
		Str("fiat_to_sup_conversion_cut", fiatToSupConversionCut.String()).
		Logger()

	// Update any Fiat Items with SUPS pricing
	products, err := boiler.FiatProducts(
		qm.Load(boiler.FiatProductRels.FiatProductPricings),
		qm.Load(boiler.FiatProductRels.StorefrontMysteryCrate),
		qm.Where(fmt.Sprintf(
			`EXISTS (
				SELECT 1 
				FROM %s _p
				WHERE _p.%s = %s AND _p.%s = ?
			)`,
			boiler.TableNames.FiatProductPricings,
			boiler.FiatProductPricingColumns.FiatProductID,
			boiler.FiatProductColumns.ID,
			boiler.FiatProductPricingColumns.CurrencyCode,
		), server.FiatCurrencyCodeSUPS),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to load fiat products")
		return
	}

	// Update Crate SUP pricing to be 20% cheaper than Fiat Pricing
	for _, p := range products {
		pl := l.With().Str("fiat_product_id", p.ID).Logger()

		var (
			priceUSD    decimal.NullDecimal
			pricingSUPS *boiler.FiatProductPricing
		)
		for _, price := range p.R.FiatProductPricings {
			if price.CurrencyCode == server.FiatCurrencyCodeUSD {
				priceUSD = decimal.NewNullDecimal(price.Amount)
			} else if price.CurrencyCode == server.FiatCurrencyCodeSUPS {
				pricingSUPS = price
			}
		}
		if !priceUSD.Valid {
			pl.Error().Msg("unable to find USD pricing to use for converting to SUPS")
			continue
		}
		if pricingSUPS == nil {
			pl.Error().Msg("unable to find SUP pricing record to update")
			continue
		}

		convertedPrice := priceUSD.Decimal.Div(decimal.NewFromInt(100))
		pl = pl.With().Str("fiat_price", convertedPrice.String()).Logger()

		convertedPrice = convertedPrice.Div(currentRates.SUPtoUSD).
			Mul(decimal.New(1, 0).Sub(fiatToSupConversionCut)).
			Mul(decimal.New(1, 18))
		pl = pl.With().Str("converted_sup_price", convertedPrice.String()).Logger()

		pricingSUPS.Amount = convertedPrice
		pricingSUPS.UpdatedAt = time.Now()
		_, err := pricingSUPS.Update(gamedb.StdConn, boil.Whitelist(boiler.FiatProductPricingColumns.Amount, boiler.FiatProductPricingColumns.UpdatedAt))
		if err != nil {
			pl.Error().Err(err).Msg("failed to update fiat sup pricing")
			continue
		}

		// Broadcast new storefront mystery crate payload
		if p.R.StorefrontMysteryCrate != nil {
			resp := server.StoreFrontMysteryCrateFromBoiler(p.R.StorefrontMysteryCrate)
			resp.FiatProduct = server.FiatProductFromBoiler(p)
			resp.Price = convertedPrice
			ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", p.FactionID, p.R.StorefrontMysteryCrate.ID), server.HubKeyMysteryCrateSubscribe, resp)
		}

		pl.Debug().Msg("Fiat Product sup price updated")
	}
}

func (f *FiatController) processShoppingCartGarbageCollection() {
	affected, err := db.ShoppingCartGC(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to process expired shopping cart")
	}
	for _, userID := range affected {
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/shopping_cart_expired", userID), server.HubKeyShoppingCartExpired, true)
	}
	gamelog.L.Debug().Int("num_deleted", len(affected)).Msg("shopping cart garbage collection completed")
}
