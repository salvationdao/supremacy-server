package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (api *API) GlobalAnnouncementSend(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &server.GlobalAnnouncement{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invaid request %w", err))
	}

	defer r.Body.Close()

	if req.Message == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("message cannot be empty %w", err))
	}
	if req.Title == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("title cannot be empty %w", err))
	}

	if req.ShowFromBattleNumber == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("show from battle number cannot be empty %w", err))
	}

	if req.ShowUntilBattleNumber == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("show until battle number cannot be empty %w", err))
	}

	// delete old announcements
	_, err = boiler.GlobalAnnouncements().DeleteAll(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to delete announcement %w", err))
	}

	// create global announcement
	ga := &boiler.GlobalAnnouncement{
		Title:                 req.Title,
		Message:               req.Message,
		ShowFromBattleNumber:  null.IntFrom(*req.ShowFromBattleNumber),
		ShowUntilBattleNumber: null.IntFrom(*req.ShowUntilBattleNumber),
	}

	// insert to db
	err = ga.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to create announcement %w", err))
	}

	go api.MessageBus.Send(r.Context(), messagebus.BusKey(HubKeyGlobalAnnouncementSubscribe), ga)

	return http.StatusOK, nil
}

func (api *API) GlobalAnnouncementDelete(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()

	// delete from db
	_, err := boiler.GlobalAnnouncements().DeleteAll(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to delete announcement %w", err))
	}

	go api.MessageBus.Send(r.Context(), messagebus.BusKey(HubKeyGlobalAnnouncementSubscribe), nil)

	return http.StatusOK, nil
}
