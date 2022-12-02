package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func WithDev(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		if !server.IsDevelopmentEnv() {
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		devPass := r.Header.Get("X-Authorization")
		if devPass != "NinjaDojo_!" {
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}

		return next(w, r)
	}
	return fn
}

type GiveCrateRequest struct {
	ColumnName string `json:"column_name"`
	Value      string `json:"value"`
	Type       string `json:"type"` // weapon || mech
}

func (api *API) ProdGiveCrate(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &GiveCrateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if req.ColumnName != "id" && req.ColumnName != "public_address" && req.ColumnName != "username" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request: column_name must be 'id', 'public_address' or 'username'"))
	}

	L := gamelog.L.With().Interface("req", req).Str("func", "ProdGiveCrate").Logger()
	crateType := req.Type

	// get player
	user, err := boiler.Players(qm.Where(fmt.Sprintf("%s = ?", req.ColumnName), req.Value)).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player by: %s: %s err: %w", req.ColumnName, req.Value, err))
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		L.Error().Err(err).Msg("unable to begin tx")
		return http.StatusInternalServerError, terror.Error(err, "Issue claiming mystery crate, please try again or contact support.")
	}
	defer tx.Rollback()

	// get mech crates
	storeMechCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(boiler.CrateTypeMECH),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(user.FactionID.String),
		qm.Load(qm.Rels(boiler.StorefrontMysteryCrateRels.FiatProduct, boiler.FiatProductRels.FiatProductPricings)),
	).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get mech crate for claim, please try again or contact support.")
	}

	// get weapon crates
	storeWeaponCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(boiler.CrateTypeWEAPON),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(user.FactionID.String),
		qm.Load(qm.Rels(boiler.StorefrontMysteryCrateRels.FiatProduct, boiler.FiatProductRels.FiatProductPricings)),
	).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get mech crate for claim, please try again or contact support.")
	}

	switch crateType {
	case "mech":
		assignedMechCrate, xa, err := assignAndRegisterPurchasedCrate(user.ID, storeMechCrate, tx, api)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Issue claiming mech crate, please try again or contact support.")
		}
		err = api.Passport.AssetRegister(xa)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("mystery crate", "").Msg("failed to register to XSYN")
			return http.StatusInternalServerError, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
		}
		serverMechCrate := server.StoreFrontMysteryCrateFromBoiler(storeMechCrate)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", user.FactionID.String, assignedMechCrate.ID), server.HubKeyMysteryCrateSubscribe, serverMechCrate)

	case "weapon":
		assignedWeaponCrate, xa, err := assignAndRegisterPurchasedCrate(user.ID, storeWeaponCrate, tx, api)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Issue claiming weapon crate, please try again or contact support.")
		}
		err = api.Passport.AssetRegister(xa)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("mystery crate", "").Msg("failed to register to XSYN")
			return http.StatusInternalServerError, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
		}
		serverWeaponCrate := server.StoreFrontMysteryCrateFromBoiler(storeWeaponCrate)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", user.FactionID.String, assignedWeaponCrate.ID), server.HubKeyMysteryCrateSubscribe, serverWeaponCrate)
	}
	err = tx.Commit()
	if err != nil {
		L.Error().Err(err).Msg("failed to commit mystery crate transaction")
		return http.StatusInternalServerError, terror.Error(err, "Issue claiming mystery crate, please try again or contact support.")
	}

	return http.StatusOK, nil
}
