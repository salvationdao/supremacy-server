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

type FiatProductColumn string

const (
	FiatProductColumnCreatedAt   FiatProductColumn = "created_at"
	FiatProductColumnName        FiatProductColumn = "name"
	FiatProductColumnFaction     FiatProductColumn = "faction_id"
	FiatProductColumnDescription FiatProductColumn = "description"
	FiatProductColumnProductType FiatProductColumn = "product_type"
)

func (c FiatProductColumn) IsValid() error {
	switch c {
	case FiatProductColumnCreatedAt,
		FiatProductColumnName,
		FiatProductColumnFaction,
		FiatProductColumnDescription,
		FiatProductColumnProductType:
		return nil

	}
	return terror.Error(fmt.Errorf("invalid sort fiat product column"))
}

func (c FiatProductColumn) ColumnName() string {
	if c == FiatProductColumnFaction {
		return boiler.FactionTableColumns.Label
	}
	return qm.Rels(boiler.TableNames.FiatProducts, string(c))
}

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
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Factions,
			boiler.FactionTableColumns.ID,
			boiler.FiatProductTableColumns.FactionID,
		),
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

	// Get pricing
	pricing, err := boiler.FiatProductPricings(
		boiler.FiatProductPricingWhere.FiatProductID.EQ(output.ID),
	).All(conn)
	if err != nil {
		return nil, err
	}

	output.Pricing = []*server.FiatProductPricing{}
	for _, p := range pricing {
		item := &server.FiatProductPricing{
			CurrencyCode: p.CurrencyCode,
			Amount:       p.Amount,
		}
		output.Pricing = append(output.Pricing, item)
	}

	// Get product items
	productItems, err := boiler.FiatProductItems(
		boiler.FiatProductItemWhere.ProductID.EQ(output.ID),
		qm.Load(boiler.FiatProductItemRels.ProductItemFiatProductItemBlueprints),
	).All(conn)
	if err != nil {
		return nil, err
	}

	output.Items = []*server.FiatProductItem{}
	for _, pi := range productItems {
		item := &server.FiatProductItem{
			ID:         pi.ID,
			Name:       pi.Name,
			ItemType:   pi.ItemType,
			Blueprints: []*server.FiatProductItemBlueprint{},
		}
		for _, bp := range pi.R.ProductItemFiatProductItemBlueprints {
			bpItem := &server.FiatProductItemBlueprint{
				ID: bp.ID,
			}
			if bp.MechBlueprintID.Valid {
				bpItem.MechBlueprintID = bp.MechBlueprintID.String
			}
			if bp.MechAnimationBlueprintID.Valid {
				bpItem.MechAnimationBlueprintID = bp.MechAnimationBlueprintID.String
			}
			if bp.MechSkinBlueprintID.Valid {
				bpItem.MechSkinBlueprintID = bp.MechSkinBlueprintID.String
			}
			if bp.UtilityBlueprintID.Valid {
				bpItem.UtilityBlueprintID = bp.UtilityBlueprintID.String
			}
			if bp.WeaponBlueprintID.Valid {
				bpItem.WeaponBlueprintID = bp.WeaponBlueprintID.String
			}
			if bp.WeaponSkinBlueprintID.Valid {
				bpItem.WeaponSkinBlueprintID = bp.WeaponSkinBlueprintID.String
			}
			if bp.AmmoBlueprintID.Valid {
				bpItem.AmmoBlueprintID = bp.AmmoBlueprintID.String
			}
			if bp.PowerCoreBlueprintID.Valid {
				bpItem.PowerCoreBlueprintID = bp.PowerCoreBlueprintID.String
			}
			if bp.PlayerAbilityBlueprintID.Valid {
				bpItem.PlayerAbilityBlueprintID = bp.PlayerAbilityBlueprintID.String
			}
			item.Blueprints = append(item.Blueprints, bpItem)
		}

		output.Items = append(output.Items, item)
	}

	return output, nil
}

type FiatProductFilter struct {
	FactionID   []string `json:"faction_id"`
	ProductType []string `json:"product_type"`
}

// FiatProducts gets a list of available fiat products to purchase by faction.
func FiatProducts(conn boil.Executor, filters *FiatProductFilter, search string, sortBy string, sortDir SortByDir, offset int, pageSize int) (int64, []*server.FiatProduct, error) {
	queryMods := []qm.QueryMod{}

	// Filters
	if filters != nil {
		if len(filters.FactionID) > 0 {
			queryMods = append(queryMods, boiler.FiatProductWhere.FactionID.IN(filters.FactionID))
		}
		if len(filters.ProductType) > 0 {
			queryMods = append(queryMods, boiler.FiatProductWhere.ProductType.IN(filters.ProductType))
		}
	}
	if search != "" {
		xsearch := ParseQueryText(search, true)
		queryMods = append(queryMods,
			qm.And(
				fmt.Sprintf(
					`(
						(to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
					)`,
					boiler.FiatProductTableColumns.Name,
					boiler.FiatProductTableColumns.Description,
					boiler.FiatProductTableColumns.ProductType,
					boiler.FactionTableColumns.Label,
				),
				xsearch,
				xsearch,
				xsearch,
				xsearch,
			),
		)
	}

	// Get total rows
	total, err := boiler.FiatProducts(queryMods...).Count(conn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.FiatProduct{}, nil
	}

	// Sort by
	if sortBy != "" {
		sortByColumn := FiatProductColumn(sortBy)
		err = sortByColumn.IsValid()
		if err != nil {
			sortByColumn = FiatProductColumnCreatedAt
		}
		if !sortDir.IsValid() {
			sortDir = SortByDirDesc
		}
		queryMods = append(queryMods, qm.OrderBy(sortByColumn.ColumnName()+" "+string(sortDir)))
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
