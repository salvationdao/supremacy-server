package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// StripeCustomerIDByUser gets the customer's stripe customer id (if available).
func StripeCustomerIDByUser(conn boil.Executor, userID string) (*string, error) {
	var output null.String
	err := boiler.Players(
		qm.Select(boiler.PlayerColumns.StripeCustomerID),
		boiler.PlayerWhere.ID.EQ(userID),
	).QueryRow(conn).Scan(&output)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if output.Valid {
		return &output.String, nil
	}
	return nil, nil
}

// UserByStripeCustomer gets the customer's stripe customer id (if available).
func UserByStripeCustomer(conn boil.Executor, stripeCustomerID string) (string, error) {
	var output string
	err := boiler.Players(
		qm.Select(boiler.PlayerColumns.ID),
		boiler.PlayerWhere.StripeCustomerID.EQ(null.StringFrom(stripeCustomerID)),
	).QueryRow(conn).Scan(&output)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return output, nil
}

///////////////
//  Product  //
///////////////

var fiatProductQueryMods = []qm.QueryMod{
	qm.Select(
		boiler.FiatProductTableColumns.ID,
		boiler.FiatProductTableColumns.FactionID,
		boiler.FiatProductTableColumns.ProductType,
		boiler.FiatProductTableColumns.Name,
		boiler.FiatProductTableColumns.Description,
		fmt.Sprintf(
			`(
				CASE %s
					WHEN '%s' THEN %s
				END
			) AS avatar_url`,
			boiler.FiatProductTableColumns.ProductType,
			boiler.FiatProductTypesMysteryCrate,
			boiler.StorefrontMysteryCrateTableColumns.AvatarURL,
		),
	),
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s AND %s = ?",
			boiler.TableNames.StorefrontMysteryCrates,
			boiler.StorefrontMysteryCrateTableColumns.FiatProductID,
			boiler.FiatProductTableColumns.ID,
			boiler.FiatProductTableColumns.ProductType,
		),
		boiler.FiatProductTypesMysteryCrate,
	),
	qm.Load(boiler.FiatProductRels.FiatProductPricings),
}

// FiatProduct gets a single fiat product item by id.
func FiatProduct(conn boil.Executor, id string) (*server.FiatProduct, error) {
	output := &server.FiatProduct{}
	err := boiler.FiatProducts(
		append(fiatProductQueryMods, boiler.FiatProductWhere.ID.EQ(id))...,
	).QueryRow(conn).Scan(
		&output.ID,
		&output.FactionID,
		&output.ProductType,
		&output.Name,
		&output.Description,
		&output.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// FiatProducts gets a list of available fiat products to purchase by faction.
func FiatProducts(conn boil.Executor, factionID *string, productType string, offset int, pageSize int) (int64, []*server.FiatProduct, error) {
	queryMods := []qm.QueryMod{
		boiler.FiatProductWhere.ProductType.EQ(productType),
	}
	if factionID != nil {
		queryMods = append(queryMods, boiler.FiatProductWhere.FactionID.EQ(*factionID))
	}

	// Get total rows
	total, err := boiler.FiatProducts(queryMods...).Count(conn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.FiatProduct{}, nil
	}

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	// Get products
	queryMods = append(queryMods, fiatProductQueryMods...)
	output := []*server.FiatProduct{}
	result := boiler.FiatProducts(queryMods...).QueryP(conn)
	if err != nil {
		return 0, nil, err
	}
	for result.Next() {
		row := &server.FiatProduct{}
		err := result.Scan(
			&row.ID,
			&row.FactionID,
			&row.ProductType,
			&row.Name,
			&row.Description,
			&row.AvatarURL,
		)
		if err != nil {
			return 0, nil, err
		}
		output = append(output, row)
	}
	return total, output, nil
}

//////////////////////
//  Orders History  //
//////////////////////

var fiatOrderQueryMods = []qm.QueryMod{
	qm.Select(
		boiler.OrderColumns.ID,
		boiler.OrderColumns.OrderNumber,
		boiler.OrderColumns.UserID,
		boiler.OrderColumns.OrderStatus,
		boiler.OrderColumns.PaymentMethod,
		boiler.OrderColumns.TXNReference,
		boiler.OrderColumns.CreatedAt,
	),
	qm.Load(boiler.OrderRels.OrderItems),
}

// FiatOrder gets a specific order.
func FiatOrder(conn boil.Executor, id string) (*server.FiatOrder, error) {
	queryMods := []qm.QueryMod{
		boiler.OrderWhere.ID.EQ(id),
	}
	o, err := boiler.Orders(append(queryMods, fiatOrderQueryMods...)...).One(conn)
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.FiatOrder{
		ID:            o.ID,
		OrderNumber:   o.OrderNumber,
		UserID:        o.UserID,
		OrderStatus:   o.OrderStatus,
		PaymentMethod: o.PaymentMethod,
		TXNReference:  o.TXNReference,
		Currency:      o.Currency,
		Items:         []*server.FiatOrderItem{},
		CreatedAt:     o.CreatedAt,
	}
	if o.R != nil && len(o.R.OrderItems) > 0 {
		for _, oi := range o.R.OrderItems {
			output.Items = append(output.Items, &server.FiatOrderItem{
				ID:            oi.ID,
				OrderID:       oi.OrderID,
				FiatProductID: oi.FiatProductID,
				Name:          oi.Name,
				Description:   oi.Description,
				Quantity:      oi.Quantity,
				Amount:        oi.Amount,
			})
		}
	}
	return output, nil
}

// FiatOrderList lists all orders made from a specific user in paginated format.
func FiatOrderList(conn boil.Executor, userID string, offset int, pageSize int) (int64, []*server.FiatOrder, error) {
	queryMods := []qm.QueryMod{
		boiler.OrderWhere.UserID.EQ(userID),
	}

	// Get total rows
	total, err := boiler.Orders(queryMods...).Count(conn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.FiatOrder{}, nil
	}

	// Sort
	queryMods = append(queryMods, qm.OrderBy(boiler.OrderColumns.CreatedAt+" DESC"))

	// Limit offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	queryMods = append(queryMods, fiatOrderQueryMods...)

	records, err := boiler.Orders(queryMods...).All(conn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	// Populate results
	output := []*server.FiatOrder{}
	for _, o := range records {
		row := &server.FiatOrder{
			ID:            o.ID,
			OrderNumber:   o.OrderNumber,
			UserID:        o.UserID,
			OrderStatus:   o.OrderStatus,
			PaymentMethod: o.PaymentMethod,
			TXNReference:  o.TXNReference,
			Currency:      o.Currency,
			Items:         []*server.FiatOrderItem{},
			CreatedAt:     o.CreatedAt,
		}
		if o.R != nil && len(o.R.OrderItems) > 0 {
			for _, oi := range o.R.OrderItems {
				row.Items = append(row.Items, &server.FiatOrderItem{
					ID:            oi.ID,
					OrderID:       oi.OrderID,
					FiatProductID: oi.FiatProductID,
					Name:          oi.Name,
					Description:   oi.Description,
					Quantity:      oi.Quantity,
					Amount:        oi.Amount,
				})
			}
		}
		output = append(output, row)
	}

	return total, output, nil
}
