package api

import (
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/helpers"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

func (api *API) GetFactionData(w http.ResponseWriter, r *http.Request) (int, error) {
	fID, ok := r.URL.Query()["factionID"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("url query param not given"))
	}
	factionID := uuid.Must(uuid.FromString(fID[0]))
	factionStat := &server.FactionStat{
		ID: server.FactionID(factionID),
	}
	err := db.FactionStatGet(api.ctx, api.Conn, factionStat)
	if err != nil {
		api.Log.Err(err).Msgf("Failed to get faction %s stat", factionID)
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get faction stat"))
	}
	if factionStat.WinCount == nil {
		factionStat.WinCount = new(int64)
	}
	if factionStat.DeathCount == nil {
		factionStat.DeathCount = new(int64)
	}
	if factionStat.LossCount == nil {
		factionStat.LossCount = new(int64)
	}
	if factionStat.KillCount == nil {
		factionStat.KillCount = new(int64)
	}
	return helpers.EncodeJSON(w, factionStat)
}
