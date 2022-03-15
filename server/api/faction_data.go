package api

import (
	"fmt"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

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
		api.Log.Err(err).Msgf("Failed to get faction %s stat", fID[0])
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
			return http.StatusInternalServerError, terror.Error(err, "failed to get faction MVP player")
		}

		result.MvpPlayerUsername = p.Username.String
	}

	return helpers.EncodeJSON(w, fs)
}

func (api *API) TriggerAbilityFileUpload(w http.ResponseWriter, r *http.Request) (int, error) {

	return helpers.EncodeJSON(w, true)
}
