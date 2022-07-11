package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server"
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
	r.Get("/all-features", WithToken(api.Config.ServerStreamKey, WithError(f.AllFeatures)))
	r.Post("/features-by-id", WithToken(api.Config.ServerStreamKey, WithError(f.PlayerFeaturesByID)))
	r.Post("/add-feature-by-IDs", WithToken(api.Config.ServerStreamKey, WithError(f.AddFeatureByIDs)))
	r.Post("/add-feature-by-addresses", WithToken(api.Config.ServerStreamKey, WithError(f.AddFeatureByAddresses)))
	r.Post("/remove-feature-by-IDs", WithToken(api.Config.ServerStreamKey, WithError(f.RemoveFeatureByIDs)))
	r.Post("/remove-feature-by-addresses", WithToken(api.Config.ServerStreamKey, WithError(f.RemoveFeatureByAddresses)))

	return r
}

type FeaturesbyIDsRequest struct {
	FeatureName string   `json:"feature_name"`
	IDs         []string `json:"ids"`
}

type FeaturesbyAddressesRequest struct {
	FeatureName string   `json:"feature_name"`
	Addresses   []string `json:"addresses"`
}

type FeaturesByIDRequest struct {
	ID string `json:"id"`
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

	serverFeature := server.FeaturesFromBoiler(features)

	return helpers.EncodeJSON(w, serverFeature)
}

func (f *FeaturesController) AddFeatureByIDs(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FeaturesbyIDsRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	err = db.AddFeatureToPlayerIDs(req.FeatureName, req.IDs)
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
	err = db.AddFeatureToPublicAddresses(req.FeatureName, req.Addresses)
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
	err = db.RemoveFeatureFromPlayerIDs(req.FeatureName, req.IDs)
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
	err = db.RemoveFeatureFromPublicAddresses(req.FeatureName, req.Addresses)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to remove features to public addresses")
	}

	return http.StatusOK, nil
}
