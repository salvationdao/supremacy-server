package api

import (
	"context"
	"encoding/json"
	"server/db"
	"server/db/boiler"
	"server/gamelog"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
)

const HubKeyPlayerProfileLayersList = "PLAYER:PROFILE:LAYERS:LIST"

type PlayerProfileAvatarLayersRequest struct {
	Payload struct {
		PlayerID  string                `json:"player_id"`
		Search    string                `json:"search"`
		Filter    *db.ListFilterRequest `json:"filter"`
		Sort      *db.ListSortRequest   `json:"sort"`
		PageSize  int                   `json:"page_size"`
		Page      int                   `json:"page"`
		LayerType null.String           `json:"layer_type"`
	} `json:"payload"`
}

type LayersResponse struct {
	Layers []*Layer `json:"layers"`
	Total  int64    `json:"total"`
}

func (pac *PlayerController) PlayerProfileAvatarLayersListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerProfileAvatarLayersRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	listOpts := &db.AvatarLayersListOpts{
		Search:    req.Payload.Search,
		Filter:    req.Payload.Filter,
		Sort:      req.Payload.Sort,
		PageSize:  req.Payload.PageSize,
		Page:      req.Payload.Page,
		LayerType: req.Payload.LayerType,
	}

	total, dbLayers, err := db.LayersList(listOpts)
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your War Machine assets, please try again or contact support.")
	}

	layers := []*Layer{}

	for _, h := range dbLayers {
		layers = append(layers, &Layer{
			ImageURL: h.ImageURL,
			ID:       h.ID,

			Type: h.Type,
		})
	}

	reply(&LayersResponse{
		Total:  total,
		Layers: layers,
	})

	return nil
}

type Layer struct {
	ID       string      `json:"id,omitempty"`
	ImageURL null.String `json:"image_url,omitempty"`
	Type     string      `json:"type"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
