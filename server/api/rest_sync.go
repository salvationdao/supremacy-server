package api

import (
	"fmt"
	"net/http"
	"server/gamedb"
	"server/helpers"
	"server/synctool"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
)

func (api *API) SyncStaticData(w http.ResponseWriter, r *http.Request) (int, error) {
	branch := chi.URLParam(r, "branch")
	if branch == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no branch provided"), "Failed to provide branch to sync data")
	}

	timeout := time.Minute

	url := fmt.Sprintf("%s/%s/sbattle_arena.csv", api.SyncConfig.FilePath, branch)
	f, err := synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync battle arena data")
	}
	err = synctool.SyncBattleArenas(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync battle arena with db")
	}

	url = fmt.Sprintf("%s/%s/factions.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
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

	url = fmt.Sprintf("%s/%s/shield_types.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncShieldTypes(f, gamedb.StdConn)
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

	url = fmt.Sprintf("%s/%s/mech_skins.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncMechSkins(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/mechs.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncMechModels(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/mech_model_skin_compatibilities.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncMechModelSkinCompatibilities(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/weapons.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncWeaponModel(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/weapon_model_skin_compatibilities.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncWeaponModelSkinCompatibilities(f, gamedb.StdConn)
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

	url = fmt.Sprintf("%s/%s/game_abilities.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncGameAbilities(f, gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction with db")
	}

	url = fmt.Sprintf("%s/%s/player_abilities.csv", api.SyncConfig.FilePath, branch)
	f, err = synctool.DownloadFile(api.ctx, url, timeout)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to sync faction data")
	}
	err = synctool.SyncPlayerAbilities(f, gamedb.StdConn)
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
