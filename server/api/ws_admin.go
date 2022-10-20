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

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AdminController struct {
	API *API
}

func NewAdminController(api *API) *AdminController {
	adminHub := &AdminController{
		API: api,
	}

	api.SecureAdminCommand(HubKeyAdminFiatProductGet, adminHub.FiatProductGet)
	api.SecureAdminCommand(HubKeyAdminFiatProductList, adminHub.FiatProductList)
	api.SecureAdminCommand(HubKeyAdminFiatProductCreate, adminHub.FiatProductCreate)
	api.SecureAdminCommand(HubKeyAdminFiatProductUpdate, adminHub.FiatProductUpdate)
	api.SecureAdminCommand(HubKeyAdminFiatBlueprintMechList, adminHub.FiatBlueprintMechList)
	api.SecureAdminCommand(HubKeyAdminFiatBlueprintMechSkinList, adminHub.FiatBlueprintMechSkinList)
	api.SecureAdminCommand(HubKeyAdminFiatBlueprintWeaponSkinList, adminHub.FiatBlueprintWeaponSkinList)
	api.SecureAdminCommand(HubKeyAdminFiatBlueprintMechAnimationList, adminHub.FiatBlueprintMechAnimationList)

	return adminHub
}

type AdminFiatProductGetRequest struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

const HubKeyAdminFiatProductGet = "ADMIN:FIAT:PRODUCT:GET"

func (ac *AdminController) FiatProductGet(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get packages, please try again."

	req := &AdminFiatProductGetRequest{}
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

	reply(product)

	return nil
}

type AdminFiatProductListRequest struct {
	Payload struct {
		Filters  *db.FiatProductFilter `json:"filters"`
		Search   string                `json:"search"`
		SortBy   string                `json:"sort_by"`
		SortDir  db.SortByDir          `json:"sort_dir"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type AdminFiatProductListResponse struct {
	Total   int64                 `json:"total"`
	Records []*server.FiatProduct `json:"records"`
}

const HubKeyAdminFiatProductList = "ADMIN:FIAT:PRODUCT:LIST"

func (ac *AdminController) FiatProductList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get packages, please try again."

	req := &AdminFiatProductListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, storePackages, err := db.FiatProducts(gamedb.StdConn, req.Payload.Filters, req.Payload.Search, req.Payload.SortBy, req.Payload.SortDir, offset, req.Payload.PageSize)
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

type AdminFiatProductCreateRequest struct {
	Payload struct {
		Name                      string   `json:"name"`
		Description               string   `json:"description"`
		Factions                  []string `json:"factions"`
		ProductType               string   `json:"product_type"`
		MechBlueprintIDs          []string `json:"mech_blueprint_ids"`
		MechSkinBlueprintIDs      []string `json:"mech_skin_blueprint_ids"`
		MechAnimationBlueprintIDs []string `json:"mech_animation_blueprint_ids"`
		WeaponSkinBlueprintIDs    []string `json:"weapon_skin_blueprint_ids"`
		PriceDollars              int64    `json:"price_dollars"`
		PriceCents                int64    `json:"price_cents"`
	} `json:"payload"`
}

const HubKeyAdminFiatProductCreate = "ADMIN:FIAT:PRODUCT:CREATE"

func (ac *AdminController) FiatProductCreate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to create product, please try again."

	req := &AdminFiatProductCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	if req.Payload.Name == "" {
		return terror.Error(fmt.Errorf("name is required"), "Name is required.")
	}
	if req.Payload.Description == "" {
		return terror.Error(fmt.Errorf("description is required"), "Description is required.")
	}
	if len(req.Payload.Factions) == 0 {
		return terror.Error(fmt.Errorf("faction is required"), "At least one faction is required.")
	}
	if req.Payload.ProductType == "" {
		return terror.Error(fmt.Errorf("product type is required"), "Product type is required.")
	}
	if req.Payload.ProductType == boiler.ItemTypeMysteryCrate {
		// TODO: remove this when able to deal with more product types?
		return terror.Error(fmt.Errorf("invalid product type"), "Invalid product type.")
	}
	if req.Payload.PriceDollars <= 0 && req.Payload.PriceCents <= 0 {
		return terror.Error(fmt.Errorf("pricing is required"), "At least one pricing is required.")
	}

	// Create Product
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer tx.Rollback()

	factionProcessed := map[string]struct{}{}
	for _, factionID := range req.Payload.Factions {
		if factionID != server.RedMountainFactionID && factionID != server.ZaibatsuFactionID && factionID != server.BostonCyberneticsFactionID {
			return terror.Error(fmt.Errorf("invalid faction"), "Invalid faction received.")
		}
		if _, ok := factionProcessed[factionID]; ok {
			continue
		}
		product := &boiler.FiatProduct{
			Name:        req.Payload.Name,
			Description: req.Payload.Description,
			ProductType: req.Payload.ProductType,
			FactionID:   factionID,
		}
		err := product.Insert(tx, boil.Infer())
		if err != nil {
			return terror.Error(err, errMsg)
		}

		pricing := &boiler.FiatProductPricing{
			FiatProductID: product.ID,
			CurrencyCode:  server.FiatCurrencyCodeUSD,
			Amount:        decimal.NewFromInt(req.Payload.PriceDollars*100 + req.Payload.PriceCents),
		}
		err = pricing.Insert(tx, boil.Infer())
		if err != nil {
			return terror.Error(err, errMsg)
		}

		if req.Payload.ProductType == boiler.ItemTypeMech {
			blueprintMechs, err := db.BlueprintMechs(req.Payload.MechBlueprintIDs)
			if err != nil {
				return terror.Error(err, errMsg)
			}
			if len(blueprintMechs) != len(req.Payload.MechBlueprintIDs) {
				return terror.Error(fmt.Errorf("invalid blueprint mech(s)"), "Invalid blueprint mech(s).")
			}

			for _, bpm := range blueprintMechs {
				item := &boiler.FiatProductItem{
					ProductID: product.ID,
					Name:      bpm.Label,
					ItemType:  boiler.FiatProductItemTypesSingleItem,
				}
				err := item.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
				itemBlueprint := &boiler.FiatProductItemBlueprint{
					ProductItemID:   null.StringFrom(item.ID),
					MechBlueprintID: null.StringFrom(bpm.ID),
				}
				err = itemBlueprint.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
			}
		} else if req.Payload.ProductType == boiler.ItemTypeMechSkin {
			blueprintMechSkins, err := db.BlueprintMechSkinSkins(tx, req.Payload.MechSkinBlueprintIDs)
			if err != nil {
				return terror.Error(err, errMsg)
			}
			if len(blueprintMechSkins) != len(req.Payload.MechSkinBlueprintIDs) {
				return terror.Error(fmt.Errorf("invalid blueprint mech skin(s)"), "Invalid blueprint mech skin(s).")
			}

			for _, bpms := range blueprintMechSkins {
				item := &boiler.FiatProductItem{
					ProductID: product.ID,
					Name:      bpms.Label,
					ItemType:  boiler.FiatProductItemTypesSingleItem,
				}
				err := item.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
				itemBlueprint := &boiler.FiatProductItemBlueprint{
					ProductItemID:       null.StringFrom(item.ID),
					MechSkinBlueprintID: null.StringFrom(bpms.ID),
				}
				err = itemBlueprint.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
			}
		} else if req.Payload.ProductType == boiler.ItemTypeMechAnimation {
			blueprintMechAnimations, err := db.BlueprintMechAnimations(req.Payload.MechAnimationBlueprintIDs)
			if err != nil {
				return terror.Error(err, errMsg)
			}
			if len(blueprintMechAnimations) != len(req.Payload.MechAnimationBlueprintIDs) {
				return terror.Error(fmt.Errorf("invalid blueprint mech animations(s)"), "Invalid blueprint mech animation(s).")
			}

			for _, bpma := range blueprintMechAnimations {
				item := &boiler.FiatProductItem{
					ProductID: product.ID,
					Name:      bpma.Label,
					ItemType:  boiler.FiatProductItemTypesSingleItem,
				}
				err := item.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
				itemBlueprint := &boiler.FiatProductItemBlueprint{
					ProductItemID:            null.StringFrom(item.ID),
					MechAnimationBlueprintID: null.StringFrom(bpma.ID),
				}
				err = itemBlueprint.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
			}
		} else if req.Payload.ProductType == boiler.ItemTypeWeaponSkin {
			blueprintWeaponSkins, err := db.BlueprintWeaponSkins(req.Payload.WeaponSkinBlueprintIDs)
			if err != nil {
				return terror.Error(err, errMsg)
			}
			if len(blueprintWeaponSkins) != len(req.Payload.WeaponSkinBlueprintIDs) {
				return terror.Error(fmt.Errorf("invalid blueprint weapon skin(s)"), "Invalid blueprint weapon skin(s).")
			}

			for _, bpws := range blueprintWeaponSkins {
				item := &boiler.FiatProductItem{
					ProductID: product.ID,
					Name:      bpws.Label,
					ItemType:  boiler.FiatProductItemTypesSingleItem,
				}
				err := item.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
				itemBlueprint := &boiler.FiatProductItemBlueprint{
					ProductItemID:         null.StringFrom(item.ID),
					WeaponSkinBlueprintID: null.StringFrom(bpws.ID),
				}
				err = itemBlueprint.Insert(tx, boil.Infer())
				if err != nil {
					return terror.Error(err, errMsg)
				}
			}
		}

		factionProcessed[factionID] = struct{}{}
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)
	return nil
}

type AdminFiatProductUpdateRequest struct {
	Payload struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		PriceDollars int64  `json:"price_dollars"`
		PriceCents   int64  `json:"price_cents"`
	} `json:"payload"`
}

const HubKeyAdminFiatProductUpdate = "ADMIN:FIAT:PRODUCT:UPDATE"

func (ac *AdminController) FiatProductUpdate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to update product, please try again."

	req := &AdminFiatProductUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	if req.Payload.ID == "" {
		return terror.Error(fmt.Errorf("product id is required"), "Product id is required.")
	}
	if req.Payload.Name == "" {
		return terror.Error(fmt.Errorf("name is required"), "Name is required.")
	}
	if req.Payload.Description == "" {
		return terror.Error(fmt.Errorf("description is required"), "Description is required.")
	}
	if req.Payload.PriceDollars <= 0 && req.Payload.PriceCents <= 0 {
		return terror.Error(fmt.Errorf("pricing is required"), "At least one pricing is required.")
	}

	// Get product
	product, err := boiler.FiatProducts(
		boiler.FiatProductWhere.ID.EQ(req.Payload.ID),
		qm.Load(boiler.FiatProductRels.FiatProductPricings),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(fmt.Errorf("product not found"), "Unable to find product, please try again.")
	}
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Update Product
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(fmt.Errorf("unable to start db transaction"), errMsg)
	}
	defer tx.Rollback()

	product.Name = req.Payload.Name
	product.Description = req.Payload.Description

	_, err = product.Update(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	var pricingUSD *boiler.FiatProductPricing
	for _, p := range product.R.FiatProductPricings {
		if p.CurrencyCode == server.FiatCurrencyCodeUSD {
			pricingUSD = p
			pricingUSD.Amount = decimal.NewFromInt(req.Payload.PriceDollars*100 + req.Payload.PriceCents)
			_, err = pricingUSD.Update(tx, boil.Whitelist(boiler.FiatProductPricingColumns.Amount))
			if err != nil {
				return terror.Error(err, errMsg)
			}
			break
		}
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(fmt.Errorf("unable to commit db transaction"), errMsg)
	}

	reply(true)
	return nil
}

const HubKeyAdminFiatBlueprintMechList = "ADMIN:FIAT:BLUEPRINT:MECH:LIST"

func (ac *AdminController) FiatBlueprintMechList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get mech blueprints, please try again."

	blueprintMechs, err := boiler.BlueprintMechs().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := []*server.BlueprintMech{}
	for _, bm := range blueprintMechs {
		resp = append(resp, server.BlueprintMechFromBoiler(bm))
	}

	reply(resp)

	return nil
}

const HubKeyAdminFiatBlueprintMechSkinList = "ADMIN:FIAT:BLUEPRINT:MECH:SKIN:LIST"

func (ac *AdminController) FiatBlueprintMechSkinList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get mech skin blueprints, please try again."

	blueprintMechSkins, err := boiler.BlueprintMechSkins().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := []*server.BlueprintMechSkin{}
	for _, bms := range blueprintMechSkins {
		resp = append(resp, server.BlueprintMechSkinFromBoiler(bms))
	}

	reply(resp)

	return nil
}

const HubKeyAdminFiatBlueprintWeaponSkinList = "ADMIN:FIAT:BLUEPRINT:WEAPON:SKIN:LIST"

func (ac *AdminController) FiatBlueprintWeaponSkinList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get weapon skin blueprints, please try again."

	blueprintWeaponSkins, err := boiler.BlueprintWeaponSkins().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := []*server.BlueprintWeaponSkin{}
	for _, bms := range blueprintWeaponSkins {
		resp = append(resp, server.BlueprintWeaponSkinFromBoiler(bms))
	}

	reply(resp)

	return nil
}

const HubKeyAdminFiatBlueprintMechAnimationList = "ADMIN:FIAT:BLUEPRINT:MECH:ANIMATION:LIST"

func (ac *AdminController) FiatBlueprintMechAnimationList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get mech animation blueprints, please try again."

	blueprintMechAnimations, err := boiler.BlueprintMechAnimations().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := []*server.BlueprintMechAnimation{}
	for _, bma := range blueprintMechAnimations {
		resp = append(resp, server.BlueprintMechAnimationFromBoiler(bma))
	}

	reply(resp)

	return nil
}
