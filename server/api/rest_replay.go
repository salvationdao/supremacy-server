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
	BattleNumber  int    `json:"battle_number"`
	ArenaID       string `json:"arena_id"`
	CloudflareUID string `json:"cloudflare_uid"`
}

func (br *BattleReplayController) AddNewReplay(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &NewReplayStruct{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	battle, err := boiler.Battles(
		boiler.BattleWhere.BattleNumber.EQ(req.BattleNumber),
		boiler.BattleWhere.ArenaID.EQ(req.ArenaID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Int("Battle Number", req.BattleNumber).Str("ArenaID", req.ArenaID).Str("UID", req.CloudflareUID).Msg("Failed to find battle for adding new battle replay")
		return http.StatusInternalServerError, terror.Error(err, "Failed to find battle for adding new replay")
	}

	replay, err := boiler.BattleReplays(
		boiler.BattleReplayWhere.BattleID.EQ(battle.ID),
		boiler.BattleReplayWhere.ArenaID.EQ(battle.ArenaID),
		boiler.BattleReplayWhere.RecordingStatus.EQ(boiler.RecordingStatusSTOPPED),
		boiler.BattleReplayWhere.StreamID.IsNull(),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Int("Battle Number", req.BattleNumber).Str("ArenaID", req.ArenaID).Str("StreamID", req.CloudflareUID).Msg("Failed to find replay")
		return http.StatusInternalServerError, terror.Error(err, "Failed to find replay")
	}

	replay.StreamID = null.StringFrom(req.CloudflareUID)

	_, err = replay.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Int("Battle Number", req.BattleNumber).Str("ArenaID", req.ArenaID).Str("StreamID", req.CloudflareUID).Msg("Failed to update replay with stream ID")
		return http.StatusInternalServerError, terror.Error(err, "Failed to update replay with stream ID")
	}

	return http.StatusOK, nil
}
