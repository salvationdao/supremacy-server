package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SaleAbilitiesController struct {
	API *API
}

func SaleAbilitiesRouter(api *API) chi.Router {
	c := &SaleAbilitiesController{
		api,
	}
	r := chi.NewRouter()
	// Admin
	r.Get("/all", WithToken(api.Config.ServerStreamKey, WithError(c.All)))
	r.Post("/create", WithToken(api.Config.ServerStreamKey, WithError(c.Create)))
	r.Post("/delist", WithToken(api.Config.ServerStreamKey, WithError(c.Delist)))
	r.Post("/relist", WithToken(api.Config.ServerStreamKey, WithError(c.Relist)))
	r.Post("/delete", WithToken(api.Config.ServerStreamKey, WithError(c.Delete)))

	// Public
	r.Get("/availability/{player_id}", WithError(c.Availability))

	return r
}

type SaleAbilityAllFilter string

const (
	SaleAbilityAllFilterOnSale   SaleAbilityAllFilter = "on_sale"
	SaleAbilityAllFilterDelisted SaleAbilityAllFilter = "delisted"
	SaleAbilityAllFilterDeleted  SaleAbilityAllFilter = "deleted"
)

func (sac *SaleAbilitiesController) All(w http.ResponseWriter, r *http.Request) (int, error) {
	filter := r.URL.Query().Get("filter")

	qms := []qm.QueryMod{}
	if filter != "" {
		if filter != string(SaleAbilityAllFilterOnSale) && filter != string(SaleAbilityAllFilterDelisted) && filter != string(SaleAbilityAllFilterDeleted) {
			return http.StatusBadRequest, terror.Error(fmt.Errorf("Invalid request: filter must be '%s', '%s' or '%s'",
				SaleAbilityAllFilterOnSale,
				SaleAbilityAllFilterDelisted,
				SaleAbilityAllFilterDeleted))
		}

		switch filter {
		case string(SaleAbilityAllFilterOnSale):
			qms = append(qms, boiler.SalePlayerAbilityWhere.RarityWeight.GT(0)) // deleted_at is null check is automatically appended by boiler
			break
		case string(SaleAbilityAllFilterDelisted):
			qms = append(qms, boiler.SalePlayerAbilityWhere.RarityWeight.LT(0)) // deleted_at is null check is automatically appended by boiler
			break
		case string(SaleAbilityAllFilterDeleted):
			qms = append(qms, boiler.SalePlayerAbilityWhere.DeletedAt.IsNotNull(), qm.WithDeleted())
			break
		}
	}
	qms = append(qms, qm.Load(boiler.SalePlayerAbilityRels.Blueprint))

	spas, err := boiler.SalePlayerAbilities(qms...).All(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get sale abilities")
	}

	detailedSaleAbilities := []*db.SaleAbilityDetailed{}
	for _, s := range spas {
		detailedSaleAbilities = append(detailedSaleAbilities, &db.SaleAbilityDetailed{
			SalePlayerAbility: s,
			Ability:           s.R.Blueprint,
		})
	}

	return helpers.EncodeJSON(w, detailedSaleAbilities)
}

type SaleAbilitiesCreateRequest struct {
	BlueprintID  string `json:"blueprint_id"`
	RarityWeight int    `json:"rarity_weight"`
}

func (sac *SaleAbilitiesController) Create(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SaleAbilitiesCreateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	if req.BlueprintID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Player ability blueprint ID must be provided"))
	}

	if req.RarityWeight <= 0 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Rarity weight cannot be negative or zero"))
	}

	spa := &boiler.SalePlayerAbility{
		BlueprintID:  req.BlueprintID,
		RarityWeight: req.RarityWeight,
	}
	err = spa.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create sale ability")
	}

	sac.API.SalePlayerAbilityManager.RehydratePool()

	return http.StatusOK, nil
}

type SaleAbilitiesDelistRequest struct {
	SaleID string `json:"sale_id"`
}

func (sac *SaleAbilitiesController) Delist(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SaleAbilitiesDelistRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	spa, err := boiler.FindSalePlayerAbility(gamedb.StdConn, req.SaleID)
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound, terror.Error(fmt.Errorf("Sale ability does not exist"))
	} else if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Failed to delist sale ability"))
	}

	spa.RarityWeight = -1
	_, err = spa.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Failed to delist sale ability"))
	}

	sac.API.SalePlayerAbilityManager.RehydratePool()

	return http.StatusOK, nil
}

type SaleAbilitiesRelistRequest struct {
	SaleID       string `json:"sale_id"`
	RarityWeight int    `json:"rarity_weight"`
}

func (sac *SaleAbilitiesController) Relist(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SaleAbilitiesRelistRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	if req.RarityWeight <= 0 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Rarity weight cannot be negative or zero"))
	}

	spa, err := boiler.FindSalePlayerAbility(gamedb.StdConn, req.SaleID)
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound, terror.Error(fmt.Errorf("Sale ability does not exist"))
	} else if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Failed to relist sale ability"))
	}

	spa.RarityWeight = req.RarityWeight
	_, err = spa.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Failed to relist sale ability"))
	}

	sac.API.SalePlayerAbilityManager.RehydratePool()

	return http.StatusOK, nil
}

type SaleAbilitiesDeleteRequest struct {
	SaleID string `json:"sale_id"`
}

func (sac *SaleAbilitiesController) Delete(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SaleAbilitiesDeleteRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	spa, err := boiler.FindSalePlayerAbility(gamedb.StdConn, req.SaleID)
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound, terror.Error(fmt.Errorf("Sale ability does not exist"))
	} else if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Failed to delete sale ability"))
	}

	spa.DeletedAt = null.TimeFrom(time.Now())
	_, err = spa.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Failed to delete sale ability"))
	}

	sac.API.SalePlayerAbilityManager.RehydratePool()

	return http.StatusOK, nil
}

type SaleAbilityAvailabilityType int

const (
	SaleAbilityAvailabilityUnavailable SaleAbilityAvailabilityType = iota
	SaleAbilityAvailabilityCanPurchase
)

func (sac *SaleAbilitiesController) Availability(w http.ResponseWriter, r *http.Request) (int, error) {
	playerID := chi.URLParam(r, "player_id")

	if sac.API.SalePlayerAbilityManager.CanUserPurchase(playerID) {
		return helpers.EncodeJSON(w, SaleAbilityAvailabilityCanPurchase)
	}
	return helpers.EncodeJSON(w, SaleAbilityAvailabilityUnavailable)
}
