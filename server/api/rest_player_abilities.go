package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PlayerAbilitiesController struct {
	API *API
}

func PlayerAbilitiesRouter(api *API) chi.Router {
	c := &PlayerAbilitiesController{
		api,
	}
	r := chi.NewRouter()
	r.Get("/all", WithToken(api.Config.ServerStreamKey, WithError(c.All)))
	r.Post("/give", WithToken(api.Config.ServerStreamKey, WithError(c.Give)))
	r.Post("/remove", WithToken(api.Config.ServerStreamKey, WithError(c.Remove)))

	return r
}

func (pac *PlayerAbilitiesController) All(w http.ResponseWriter, r *http.Request) (int, error) {
	pas, err := boiler.BlueprintPlayerAbilities().All(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get player abilities")
	}

	return helpers.EncodeJSON(w, pas)
}

type GivePlayerAbilityRequest struct {
	Amount      int    `json:"amount"`
	BlueprintID string `json:"blueprint_id"`

	// Player identity
	ColumnName string `json:"column_name"`
	Value      string `json:"value"`
}

func (pac *PlayerAbilitiesController) Give(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &GivePlayerAbilityRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	if req.Amount < 1 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Amount provided must be at least 1"))
	}

	if req.BlueprintID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Blueprint ID of player ability must be provided"))
	}

	if req.ColumnName != boiler.PlayerColumns.Gid && req.ColumnName != boiler.PlayerColumns.PublicAddress && req.ColumnName != boiler.PlayerColumns.Username {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request: column_name must be '%s', '%s' or '%s'",
			boiler.PlayerColumns.Gid,
			boiler.PlayerColumns.PublicAddress,
			boiler.PlayerColumns.Username))
	}

	player, err := boiler.Players(qm.Where(fmt.Sprintf("%s = ?", req.ColumnName), req.Value)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player by: %s: %s err: %w", req.ColumnName, req.Value, err))
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Issue creating transaction")
	}
	defer tx.Rollback()

	// Update player ability count
	pa, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.BlueprintID.EQ(req.BlueprintID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(player.ID),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		pa = &boiler.PlayerAbility{
			OwnerID:         player.ID,
			BlueprintID:     req.BlueprintID,
			LastPurchasedAt: time.Now(),
		}

		err = pa.Insert(tx, boil.Infer())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Issue giving player ability")
		}
	} else if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Issue giving player ability")
	}

	pa.Count = pa.Count + req.Amount
	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Issue giving player ability")
	}

	return http.StatusOK, nil
}

type RemovePlayerAbilityRequest struct {
	Amount      int    `json:"amount"`
	BlueprintID string `json:"blueprint_id"`
	RemoveAll   bool   `json:"remove_all"`

	// Player identity
	ColumnName string `json:"column_name"`
	Value      string `json:"value"`
}

func (pac *PlayerAbilitiesController) Remove(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &RemovePlayerAbilityRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	if req.Amount < 1 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Amount provided must be at least 1"))
	}

	if req.BlueprintID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Blueprint ID of player ability must be provided"))
	}

	if req.ColumnName != boiler.PlayerColumns.Gid && req.ColumnName != boiler.PlayerColumns.PublicAddress && req.ColumnName != boiler.PlayerColumns.Username {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request: column_name must be '%s', '%s' or '%s'",
			boiler.PlayerColumns.Gid,
			boiler.PlayerColumns.PublicAddress,
			boiler.PlayerColumns.Username))
	}

	player, err := boiler.Players(qm.Where(fmt.Sprintf("%s = ?", req.ColumnName), req.Value)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player by: %s: %s err: %w", req.ColumnName, req.Value, err))
	}

	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("cant find player %s: %s %w", req.ColumnName, req.Value, err))
	}

	// Update player ability count
	pa, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.BlueprintID.EQ(req.BlueprintID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(player.ID),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound, terror.Error(err, "Player does not own this ability already")
	}
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Issue removing player ability")
	}

	pa.Count = int(math.Max(float64(pa.Count-req.Amount), 0))
	if req.RemoveAll {
		pa.Count = 0
	}
	_, err = pa.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Issue removing player ability")
	}

	return http.StatusOK, nil
}
