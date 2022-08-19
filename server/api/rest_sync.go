package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server/gamedb"
	"server/helpers"
	"server/synctool"
	"time"
)

func (api *API) SyncStaticData(w http.ResponseWriter, r *http.Request) (int, error) {
	branch := chi.URLParam(r, "branch")
	if branch == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no branch provided"), "Failed to provide branch to sync data")
	}

	timeout := time.Minute

	url := fmt.Sprintf("%s/%s/factions.csv", api.SyncConfig.FilePath, branch)
	f, err := synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncFactions(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/brands.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncBrands(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/mech_skins.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncMechSkins(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/mech_models.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncMechModels(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/weapon_skins.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncWeaponSkins(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/weapon_models.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncWeaponModel(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/battle_abilities.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncBattleAbilities(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/power_cores.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncPowerCores(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/mechs.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncStaticMech(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/quests.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncStaticQuest(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	return helpers.EncodeJSON(w, "Done Syncing")
}
