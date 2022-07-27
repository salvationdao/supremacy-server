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

const HubKeyPlayerProfileHairList = "PLAYER:PROFILE:HAIR:LIST"

type PlayerProfileAvatarFeaturesRequest struct {
	Payload struct {
		PlayerID string                `json:"player_id"`
		Search   string                `json:"search"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Sort     *db.ListSortRequest   `json:"sort"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type HairRespone struct {
	Hairs []*AvatarHair `json:"hairs"`
	Total int64         `json:"total"`
}

func (pac *PlayerController) PlayerProfileAvatarHairListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerProfileAvatarFeaturesRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	listOpts := &db.AvatarFeaturesListOpts{
		Search:   req.Payload.Search,
		Filter:   req.Payload.Filter,
		Sort:     req.Payload.Sort,
		PageSize: req.Payload.PageSize,
		Page:     req.Payload.Page,
	}

	total, dbHairs, err := db.HairList(listOpts)
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your War Machine assets, please try again or contact support.")
	}

	hairs := []*AvatarHair{}

	for _, h := range dbHairs {
		hairs = append(hairs, &AvatarHair{
			ImageURL: h.ImageURL,
		})
	}

	reply(&HairRespone{
		Total: total,
		Hairs: hairs,
	})

	return nil
}

type AvatarHair struct {
	ImageURL null.String `json:"image_url,omitempty"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
