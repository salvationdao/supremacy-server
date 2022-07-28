package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/system_messages"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
)

type SystemMessagesAdminController struct {
	API *API
}

func SystemMessagesRouter(api *API) chi.Router {
	c := &SystemMessagesAdminController{
		api,
	}
	r := chi.NewRouter()
	r.Post("/broadcast", WithToken(api.Config.ServerStreamKey, WithError(c.Broadcast)))

	return r
}

type SystemMessagesAdminBroadcastRequest struct {
	FactionID string                                `json:"faction_id"`
	Title     string                                `json:"title"`
	Message   string                                `json:"message"`
	DataType  system_messages.SystemMessageDataType `json:"data_type"`
	Data      *interface{}                          `json:"data,omitempty"`
}

func (smac *SystemMessagesAdminController) Broadcast(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SystemMessagesAdminBroadcastRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	if req.Data != nil && req.DataType == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("data_type must be provided when data is not null."))
	}

	if req.FactionID != "" {
		// If faction id is provided then broadcast message to all faction players
		err = system_messages.BroadcastFactionSystemMessage(req.FactionID, req.Title, req.Message, req.DataType, req.Data)
	} else {
		// Else broadcast to all online players (global)
		err = system_messages.BroadcastGlobalSystemMessage(req.Title, req.Message, req.DataType, req.Data)
	}

	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	return http.StatusOK, nil
}
