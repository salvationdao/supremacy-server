package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"net/http"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
)

type BattleReplayController struct {
	API *API
}

func BattleReplayRouter(api *API) chi.Router {
	br := &BattleReplayController{API: api}

	r := chi.NewRouter()
	r.Post("/create", WithToken(api.Config.ServerStreamKey, WithError(br.AddNewReplay)))

	return r
}

type NewReplayStruct struct {
	ReplayID      string `json:"replay_id"`
	CloudflareUID string `json:"cloudflare_uid"`
}

func (br *BattleReplayController) AddNewReplay(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &NewReplayStruct{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	replay, err := boiler.FindBattleReplay(gamedb.StdConn, req.ReplayID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("StreamID", req.CloudflareUID).Msg("Failed to find replay")
		return http.StatusInternalServerError, terror.Error(err, "Failed to find replay")
	}

	replay.StreamID = null.StringFrom(req.CloudflareUID)

	_, err = replay.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("StreamID", req.CloudflareUID).Msg("Failed to update replay with stream ID")
		return http.StatusInternalServerError, terror.Error(err, "Failed to update replay with stream ID")
	}

	return http.StatusOK, nil
}

func (br *BattleReplayController) GetReplayDetails(w http.ResponseWriter, r *http.Request) (int, error) {
	_, err := strconv.Atoi(chi.URLParam(r, "arena-id"))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get arena GID")
	}
	_, err = strconv.Atoi(chi.URLParam(r, "battle-number"))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get battle number")
	}

	return http.StatusOK, nil
}
