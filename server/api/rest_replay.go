package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strconv"
)

type BattleReplayController struct {
	API *API
}

func BattleReplayRouter(api *API) chi.Router {
	br := &BattleReplayController{API: api}

	r := chi.NewRouter()
	r.Post("/create", WithToken(api.Config.ServerStreamKey, WithError(br.AddNewReplay)))
	r.Get("/get/{battle-number}", WithError(br.GetReplayDetails))

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
	battleNumber, err := strconv.Atoi(chi.URLParam(r, "battle-number"))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get battle number")
	}

	battleReplay, err := boiler.BattleReplays(
		boiler.BattleReplayWhere.IsCompleteBattle.EQ(true),
		qm.Where(
			fmt.Sprintf(
				"EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s = ? )",
				boiler.TableNames.Battles,
				qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.ID),
				qm.Rels(boiler.TableNames.BattleReplays, boiler.BattleReplayColumns.BattleID),
				qm.Rels(boiler.TableNames.Battles, boiler.BattleColumns.BattleNumber),
			),
			battleNumber,
		),

		qm.Load(boiler.BattleReplayRels.Battle),
		qm.Load(qm.Rels(boiler.BattleReplayRels.Battle, boiler.BattleRels.GameMap)),
	).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, fmt.Sprintf("Failed find replay with battle number of %d", battleNumber))
	}

	return helpers.EncodeJSON(w, server.BattleReplayFromBoilerWithEvent(battleReplay))
}
