package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

type Layer struct {
	ID       string      `json:"id,omitempty"`
	ImageURL null.String `json:"image_url,omitempty"`
	Type     string      `json:"type"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
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

type PlayerProfileCustomAvatarRequest struct {
	Payload struct {
		PlayerID    null.String `json:"player_id,omitempty"`
		FaceID      string      `json:"face_id,omitempty"`
		BodyID      string      `json:"body_id,omitempty"`
		HairID      null.String `json:"hair_id,omitempty"`
		AccessoryID null.String `json:"accessory_id,omitempty"`
		EyewearID   null.String `json:"eyewear_id,omitempty"`
	} `json:"payload,omitempty"`
}

const HubKeyPlayerProfileCustomAvatarCreate = "PLAYER:PROFILE:CUSTOM_AVATAR:CREATE"

func (pac *PlayerController) PlayerProfileCustomAvatarCreate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerProfileCustomAvatarRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// build custom avatar
	ava := db.AvatarCreateRequest{
		FaceID:      req.Payload.FaceID,
		BodyID:      req.Payload.BodyID,
		HairID:      req.Payload.HairID,
		AccessoryID: req.Payload.AccessoryID,
		EyeWearID:   req.Payload.EyewearID,
	}

	err = db.CustomAvatarCreate(user.ID, ava)
	if err != nil {
		return terror.Error(err, "Failed to create custom avatar, please try again or contact support.")
	}

	reply(nil)

	return nil
}

const HubKeyPlayerProfileCustomAvatarUpdate = "PLAYER:PROFILE:CUSTOM_AVATAR:UPDATE"

func (pac *PlayerController) PlayerProfileCustomAvatarUpdate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerProfileCustomAvatarRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// build custom avatar
	ava := db.AvatarCreateRequest{
		FaceID:      req.Payload.FaceID,
		BodyID:      req.Payload.BodyID,
		HairID:      req.Payload.HairID,
		AccessoryID: req.Payload.AccessoryID,
		EyeWearID:   req.Payload.EyewearID,
	}

	err = db.CustomAvatarUpdate(user.ID, ava)
	if err != nil {
		return terror.Error(err, "Failed to update custom avatar, please try again or contact support.")
	}

	return nil
}

type PlayerAvatarListRequest struct {
	Payload struct {
		Search   string                `json:"search"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Sort     *db.ListSortRequest   `json:"sort"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type PlayerAvatarListResp struct {
	Total   int64                   `json:"total"`
	Avatars []*boiler.ProfileAvatar `json:"avatars"`
}

const HubKeyPlayerAvatarList = "PLAYER:AVATAR:LIST"

func (pc *PlayerController) ProfileAvatarListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetWeaponListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	listOpts := &db.AvatarListOpts{
		Search:   req.Payload.Search,
		Filter:   req.Payload.Filter,
		Sort:     req.Payload.Sort,
		PageSize: req.Payload.PageSize,
		Page:     req.Payload.Page,
		OwnerID:  user.ID,
	}

	total, avatars, err := db.AvatarList(listOpts)
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your avatars, please try again or contact support.")
	}

	reply(&PlayerAvatarListResp{
		Total:   total,
		Avatars: avatars,
	})
	return nil
}

type PlayerAvatarUpdateRequest struct {
	Payload struct {
		PlayerID        string `json:"player_id"`
		ProfileAvatarID string `json:"profile_avatar_id"`
	} `json:"payload"`
}

const HubKeyPlayerAvatarUpdate = "PLAYER:AVATAR:UPDATE"

func (pc *PlayerController) ProfileAvatarUpdateHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAvatarUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// removing avatar
	if req.Payload.ProfileAvatarID == "" {
		user.ProfileAvatarID = null.StringFrom("")
		reply(&PlayerAvatar{
			Tier: "MEGA",
		})
		return nil
	}

	// get player avatar
	ava, err := boiler.FindProfileAvatar(gamedb.StdConn, req.Payload.ProfileAvatarID)
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue updating player avatar")
		return terror.Error(err, "Failed to update avatar, please try again or contact support.")
	}

	// update player
	user.ProfileAvatarID = null.StringFrom(ava.ID)
	user.UpdatedAt = time.Now()
	_, err = user.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.UpdatedAt, boiler.PlayerColumns.ProfileAvatarID))
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue updating player avatar")
		return terror.Error(err, "Failed to update avatar, please try again or contact support.")
	}

	reply(&PlayerAvatar{
		ID:        ava.ID,
		AvatarURL: ava.AvatarURL,
		Tier:      ava.Tier,
	})
	return nil
}

type PlayerCustomAvatarListRequest struct {
	Payload struct {
		Search   string                `json:"search"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Sort     *db.ListSortRequest   `json:"sort"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type PlayerCustomAvatarListResp struct {
	Total int64    `json:"total,omitempty"`
	IDs   []string `json:"ids,omitempty"`
}

const HubKeyPlayerCustomAvatarList = "PLAYER:CUSTOM_AVATAR:LIST"

func (pc *PlayerController) ProfileCustomAvatarListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerCustomAvatarListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	listOpts := &db.CustomAvatarsListOpts{
		Search:   req.Payload.Search,
		Filter:   req.Payload.Filter,
		Sort:     req.Payload.Sort,
		PageSize: req.Payload.PageSize,
		Page:     req.Payload.Page,
	}

	total, avatars, err := db.CustomAvatarsList(user.ID, listOpts)
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your avatars, please try again or contact support.")
	}

	reply(&PlayerCustomAvatarListResp{
		Total: total,
		IDs:   avatars,
	})
	return nil
}

const HubKeyPlayerCustomAvatarDetails = "PLAYER:CUSTOM_AVATAR:DETAILS"

func (pc *PlayerController) ProfileCustomAvatarDetailsHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("the fuck")

	cctx := chi.RouteContext(ctx)
	avatarID := cctx.URLParam("avatar_id")
	if avatarID == "" {
		return terror.Error(fmt.Errorf("missing weapon id"), "Missing weapon id.")
	}
	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	// get avatar
	ava, err := boiler.FindProfileCustomAvatar(gamedb.StdConn, avatarID)
	if err != nil {
		gamelog.L.Error().Interface("avatarID", avatarID).Err(err).Msg("issue getting custom avatar details")
		return terror.Error(err, "Failed to find your avatar, please try again or contact support.")
	}

	resp, err := db.GetCustomAvatar(gamedb.StdConn, ava.ID)
	if err != nil {
		gamelog.L.Error().Interface("avatarID", avatarID).Err(err).Msg("issue getting custom avatar details")
		return terror.Error(err, "Failed to find your avatar, please try again or contact support.")
	}

	reply(resp)
	return nil
}
