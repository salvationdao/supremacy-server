package db

import (
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// ShoppingCartGC clears all existing shopping cart that have expired.
func ShoppingCartGC(conn boil.Executor) ([]string, error) {
	output := []string{}
	q := fmt.Sprintf(
		`DELETE FROM %s WHERE %s <= NOW() AND %s = false RETURNING %s`,
		boiler.TableNames.ShoppingCarts,
		boiler.ShoppingCartColumns.ExpiresAt,
		boiler.ShoppingCartColumns.Locked,
		boiler.ShoppingCartColumns.UserID,
	)

	rows, err := conn.Query(q)
	if err != nil {
		return output, terror.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return output, terror.Error(err)
		}
		output = append(output, userID)
	}

	return output, nil
}

// ShoppingCartCreate gets or creates a new shopping cart for user.
func ShoppingCartCreate(conn boil.Executor, userID string, expiresAt time.Time) (*boiler.ShoppingCart, error) {
	item := &boiler.ShoppingCart{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	err := item.Insert(conn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	return item, nil
}

// ShoppingCartGetByUser gets a user's shopping cart.
func ShoppingCartGetByUser(conn boil.Executor, userID string) (*boiler.ShoppingCart, []*boiler.ShoppingCartItem, error) {
	cart, err := boiler.ShoppingCarts(
		boiler.ShoppingCartWhere.UserID.EQ(userID),
		boiler.ShoppingCartWhere.ExpiresAt.GT(time.Now()),
		qm.Load(qm.Rels(boiler.ShoppingCartRels.ShoppingCartItems, boiler.ShoppingCartItemRels.Product, boiler.FiatProductRels.StorefrontMysteryCrate)),
		qm.Load(qm.Rels(boiler.ShoppingCartRels.ShoppingCartItems, boiler.ShoppingCartItemRels.Product, boiler.FiatProductRels.FiatProductPricings)),
	).One(conn)
	if err != nil {
		return nil, nil, terror.Error(err)
	}
	return cart, cart.R.ShoppingCartItems, nil
}

// ShoppingCartDelete deletes the a user's shopping cart.
func ShoppingCartDelete(conn boil.Executor, id string) error {
	item := &boiler.ShoppingCart{
		ID: id,
	}
	_, err := item.Delete(conn)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartDeleteByUser deletes the a user's shopping cart.
func ShoppingCartDeleteByUser(conn boil.Executor, userID string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, boiler.TableNames.ShoppingCarts, boiler.ShoppingCartColumns.UserID)
	_, err := conn.Exec(q, userID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartToggleLock locks or unlocks a shopping cart (used for dealing with checkout process).
func ShoppingCartToggleLock(conn boil.Executor, id string, locked bool) error {
	item := &boiler.ShoppingCart{
		ID:     id,
		Locked: locked,
	}
	_, err := item.Update(conn, boil.Whitelist(boiler.ShoppingCartColumns.Locked))
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartUpdateExpiry updates the cart's expiry time.
func ShoppingCartUpdateExpiry(conn boil.Executor, id string, expiresAt time.Time) error {
	item := &boiler.ShoppingCart{
		ID:        id,
		ExpiresAt: expiresAt,
	}
	_, err := item.Update(conn, boil.Whitelist(boiler.ShoppingCartColumns.ExpiresAt))
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartItemAdd adds a single cart item into the shopping cart.
func ShoppingCartItemAdd(conn boil.Executor, shoppingCartID string, productID string, quantity int) error {
	item := &boiler.ShoppingCartItem{
		ShoppingCartID: shoppingCartID,
		ProductID:      productID,
		Quantity:       quantity,
	}
	err := item.Insert(conn, boil.Infer())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartItemQtyUpdate updates a cart item's quantity.
func ShoppingCartItemQtyUpdate(conn boil.Executor, shoppingCartID string, itemID string, quantity int) error {
	q := fmt.Sprintf(
		`UPDATE %s SET %s = $3 WHERE %s = $1 AND %s = $2`,
		boiler.TableNames.ShoppingCartItems,
		boiler.ShoppingCartItemColumns.Quantity,
		boiler.ShoppingCartItemColumns.ShoppingCartID,
		boiler.ShoppingCartItemColumns.ID,
	)
	_, err := conn.Exec(q, shoppingCartID, itemID, quantity)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartItemQtyIncrementUpdate updates a cart item's quantity via increment.
func ShoppingCartItemQtyIncrementUpdate(conn boil.Executor, shoppingCartID string, itemID string, quantityIncrement int) error {
	q := fmt.Sprintf(
		`UPDATE %[1]s SET %[2]s = %[2]s + $3 WHERE %[3]s = $1 AND %[4]s = $2`,
		boiler.TableNames.ShoppingCartItems,
		boiler.ShoppingCartItemColumns.Quantity,
		boiler.ShoppingCartItemColumns.ShoppingCartID,
		boiler.ShoppingCartItemColumns.ID,
	)
	_, err := conn.Exec(q, shoppingCartID, itemID, quantityIncrement)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ShoppingCartItemDelete deletes a single cart item.
func ShoppingCartItemDelete(conn boil.Executor, shoppingCartID string, itemID string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1 AND %s = $2`, boiler.TableNames.ShoppingCartItems, boiler.ShoppingCartItemColumns.ShoppingCartID, boiler.ShoppingCartItemColumns.ID)
	result, err := conn.Exec(q, shoppingCartID, itemID)
	if err != nil {
		return terror.Error(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return terror.Error(err)
	}
	if affected != 1 {
		return terror.Error(fmt.Errorf("cart item not found"))
	}
	return nil
}
