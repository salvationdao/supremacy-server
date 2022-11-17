package api

import (
	"fmt"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type FactionController struct {
	API *API
}

func FactionRouter(api *API) chi.Router {
	c := &FactionController{
		api,
	}
	r := chi.NewRouter()
	r.Get("/all", WithError(c.FactionAll))
	r.Get("/stat", WithError(api.GetFactionData))

	return r
}

func (c *FactionController) FactionAll(w http.ResponseWriter, r *http.Request) (int, error) {
	factions, err := boiler.Factions(
		qm.Select(
			boiler.FactionColumns.ID,
			boiler.FactionColumns.Label,
			boiler.FactionColumns.LogoURL,
			boiler.FactionColumns.BackgroundURL,
			boiler.FactionColumns.Description,
		),
	).All(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to query faction data from db")
	}

	return helpers.EncodeJSON(w, factions)
}

func (api *API) GetFactionData(w http.ResponseWriter, r *http.Request) (int, error) {
	fID, ok := r.URL.Query()["factionID"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("url query param not given"))
	}
	if len(fID) == 0 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("faction id is empty"))
	}

	fs, err := boiler.FindFactionStat(gamedb.StdConn, fID[0])
	if err != nil {
		gamelog.L.Err(err).Msgf("Failed to get faction %s stat", fID[0])
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get faction stat"))
	}

	result := &server.FactionStat{
		FactionStat: fs,
	}

	if fs.MVPPlayerID.Valid {
		p, err := boiler.Players(
			qm.Select(boiler.PlayerColumns.ID),
			qm.Select(boiler.PlayerColumns.Username),
			boiler.PlayerWhere.ID.EQ(fs.MVPPlayerID.String),
		).One(gamedb.StdConn)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to get faction MVP player")
		}

		result.MvpPlayerUsername = p.Username.String
	}

	// faction members
	result.MemberCount, err = boiler.Players(boiler.PlayerWhere.FactionID.EQ(null.StringFrom(fID[0]))).Count(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get faction member count")
	}

	return helpers.EncodeJSON(w, result)
}
