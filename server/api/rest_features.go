package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server/db"
	"server/helpers"
)

type FeaturesController struct {
	API *API
}

func FeatureRouter(api *API) chi.Router {
	f := &FeaturesController{
		api,
	}
	r := chi.NewRouter()
	r.Get("/all-features", WithError(f.AllFeatures))
	r.Post("/features-by-id", WithError(f.PlayerFeaturesByID))
	r.Post("/add-feature-by-IDs", WithError(f.AddFeatureByIDs))
	r.Post("/add-feature-by-addresses", WithError(f.AddFeatureByAddresses))
	r.Post("/remove-feature-by-IDs", WithError(f.RemoveFeatureByIDs))
	r.Post("/remove-feature-by-addresses", WithError(f.RemoveFeatureByAddresses))

	return r
}

type FeaturesbyIDsRequest struct {
	FeatureType string
	IDs         []string
}

type FeaturesbyAddressesRequest struct {
	FeatureType string
	Addresses   []string
}

type FeaturesByIDRequest struct {
	ID string
}

func (f *FeaturesController) AllFeatures(w http.ResponseWriter, r *http.Request) (int, error) {
	features, err := db.GetAllFeatures()
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to query feature data from db")
	}

	return helpers.EncodeJSON(w, features)
}

func (f *FeaturesController) PlayerFeaturesByID(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FeaturesByIDRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	features, err := db.GetPlayerFeaturesByID(req.ID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to add features to player ids")
	}

	return helpers.EncodeJSON(w, features)
}

func (f *FeaturesController) AddFeatureByIDs(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FeaturesbyIDsRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	err = db.AddFeatureToPlayerIDs(req.FeatureType, req.IDs)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to add features to player ids")
	}

	return http.StatusOK, nil
}

func (f *FeaturesController) AddFeatureByAddresses(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FeaturesbyAddressesRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	err = db.AddFeatureToPublicAddresses(req.FeatureType, req.Addresses)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to add features to public addresses")
	}

	return http.StatusOK, nil
}

func (f *FeaturesController) RemoveFeatureByIDs(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FeaturesbyIDsRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	err = db.RemoveFeatureFromPlayerIDs(req.FeatureType, req.IDs)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to remove features to player ids")
	}

	return http.StatusOK, nil
}

func (f *FeaturesController) RemoveFeatureByAddresses(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FeaturesbyAddressesRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	err = db.RemoveFeatureFromPublicAddresses(req.FeatureType, req.Addresses)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to remove features to public addresses")
	}

	return http.StatusOK, nil
}
