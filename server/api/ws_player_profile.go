package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PlayerProfileRequest struct {
	Payload struct {
		PlayerGID string `json:"player_gid"`
	} `json:"payload"`
}

type PublicPlayer struct {
	ID        string      `json:"id"`
	Username  null.String `json:"username"`
	Gid       int         `json:"gid"`
	FactionID null.String `json:"faction_id"`
	AboutMe   null.String `json:"about_me"`
	Rank      string      `json:"rank"`
	CreatedAt time.Time   `json:"created_at"`
}

type PlayerAvatar struct {
	ID        string `json:"id"`
	IsCustom  bool   `json:"is_custom"`
	AvatarURL string `json:"avatar_url"`
	Tier      string `json:"tier"`
}
type PlayerProfileResponse struct {
	*PublicPlayer `json:"player"`
	Stats         *boiler.PlayerStat      `json:"stats"`
	Faction       *boiler.Faction         `json:"faction"`
	ActiveLog     *boiler.PlayerActiveLog `json:"active_log"`
	Avatar        *PlayerAvatar           `json:"avatar"`
}

const HubKeyPlayerProfileGet = "PLAYER:PROFILE:GET"

func (pc *PlayerController) PlayerProfileGetHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerProfileRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	gid, err := strconv.Atoi(req.Payload.PlayerGID)
	if err != nil {
		gamelog.L.Error().
			Str("Player.GID", req.Payload.PlayerGID).Err(err).Msg("unable to convert player gid to int")
		return terror.Error(err, "Unable to retrieve player profile, try again or contact support.")
	}

	// get player
	player, err := boiler.Players(
		boiler.PlayerWhere.Gid.EQ(gid),

		// load faction
		qm.Load(boiler.PlayerRels.Faction),

		// load avatar
		qm.Load(boiler.PlayerRels.ProfileAvatar),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("Player.GID", req.Payload.PlayerGID).Msg("unable to get player")
		return terror.Error(err, "Unable to retrieve player profile, try again or contact support.")
	}

	// get stats
	stats, err := boiler.PlayerStats(boiler.PlayerStatWhere.ID.EQ(player.ID)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().
			Str("Player.ID", player.ID).Err(err).Msg("unable to get players stats")
		return terror.Error(err, "Unable to retrieve player profile, try again or contact support.")
	}

	// get active log
	activeLog, err := boiler.PlayerActiveLogs(
		boiler.PlayerActiveLogWhere.PlayerID.EQ(player.ID),
		qm.OrderBy(fmt.Sprintf("%s DESC", boiler.PlayerActiveLogColumns.ActiveAt)),
		qm.Limit(1),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().
			Str("Player.ID", player.ID).Err(err).Msg("unable to get player's active log")
		return terror.Error(err, "Unable to retrieve player profile, try again or contact support.")
	}

	// set faction
	var faction *boiler.Faction
	if player.R != nil && player.R.Faction != nil {
		faction = player.R.Faction
	}

	// get / set avatar
	var avatar *PlayerAvatar
	if player.ProfileAvatarID.Valid && player.R != nil && player.R.ProfileAvatar != nil {
		avatar = &PlayerAvatar{
			ID:        player.R.ProfileAvatar.ID,
			AvatarURL: player.R.ProfileAvatar.AvatarURL,
			Tier:      player.R.ProfileAvatar.Tier,
		}
	}

	// if using custom avatar
	if player.CustomAvatarID.Valid {
		avatar = &PlayerAvatar{
			ID:        player.CustomAvatarID.String,
			AvatarURL: "",
			Tier:      "MEGA",
			IsCustom:  true,
		}
	}

	reply(PlayerProfileResponse{
		PublicPlayer: &PublicPlayer{
			ID:        player.ID,
			Username:  player.Username,
			Gid:       player.Gid,
			FactionID: player.FactionID,
			AboutMe:   player.AboutMe,
			Rank:      player.Rank,
			CreatedAt: player.CreatedAt,
		},
		Avatar:    avatar,
		Stats:     stats,
		Faction:   faction,
		ActiveLog: activeLog,
	})
	return nil
}

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

func (pc *PlayerController) PlayerProfileAvatarLayersListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
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

func (pc *PlayerController) PlayerProfileCustomAvatarCreate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
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

	avatar, err := db.CustomAvatarCreate(user.ID, ava)
	if err != nil {
		return terror.Error(err, "Failed to create custom avatar, please try again or contact support.")
	}

	reply(avatar.ID)

	return nil
}

const HubKeyPlayerProfileCustomAvatarUpdate = "PLAYER:PROFILE:CUSTOM_AVATAR:UPDATE"

func (pc *PlayerController) PlayerProfileCustomAvatarUpdate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
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

const HubKeyPlayerProfileCustomAvatarDelete = "PLAYER:PROFILE:CUSTOM_AVATAR:DELETE"

type PlayerCustomAvatarDeleteRequest struct {
	Payload struct {
		AvatarID string `json:"avatar_id,omitempty"`
	} `json:"payload,omitempty"`
}

func (pc *PlayerController) PlayerProfileCustomAvatarDelete(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerCustomAvatarDeleteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// get avatar
	ava, err := boiler.FindProfileCustomAvatar(gamedb.StdConn, req.Payload.AvatarID)
	if err != nil {
		return terror.Error(err, "Failed to get avatar")
	}

	// check ownership
	if ava.PlayerID != user.ID {
		return terror.Error(err, "You don't have permission to delete this item")
	}

	// check if player has avatar equipped
	if user.CustomAvatarID.Valid && user.CustomAvatarID.String == req.Payload.AvatarID {
		return terror.Error(fmt.Errorf("cannot delete an avatar that is equipped"), "Cannot delete an avatar that is equipped")
	}

	// set deleted at
	ava.DeletedAt = null.TimeFrom(time.Now())
	_, err = ava.Update(gamedb.StdConn, boil.Whitelist(boiler.ProfileCustomAvatarColumns.DeletedAt))
	if err != nil {
		return terror.Error(err, "Failed to delete custom avatar.")
	}

	reply(nil)
	return nil
}

const HubKeyPlayerFactionPassExpiryDate = "PLAYER:FACTION:PASS:EXPIRY:DATE"

func (api *API) PlayerFactionPassExpiryDate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.FactionPassExpiresAt.Valid || !user.FactionPassExpiresAt.Time.After(time.Now()) {
		reply(nil)
		return nil
	}

	reply(user.FactionPassExpiresAt)
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
	req := &PlayerAvatarListRequest{}
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
		IsCustom        bool   `json:"is_custom"` // custom or default
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

	avatarID := ""
	avtarURL := ""
	avatarTier := ""

	// using default avatar
	if !req.Payload.IsCustom {
		// get player avatar
		ava, err := boiler.FindProfileAvatar(gamedb.StdConn, req.Payload.ProfileAvatarID)
		if err != nil {
			gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue updating player avatar")
			return terror.Error(err, "Failed to update avatar, please try again or contact support.")
		}

		// update player
		avatarID = ava.ID
		avtarURL = ava.AvatarURL
		avatarTier = ava.Tier

		user.ProfileAvatarID = null.StringFrom(ava.ID)
		user.CustomAvatarID = null.String{}
	}

	// using custom avatar
	if req.Payload.IsCustom {
		// get player avatar
		ava, err := boiler.FindProfileCustomAvatar(gamedb.StdConn, req.Payload.ProfileAvatarID)
		if err != nil {
			gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue updating player avatar")
			return terror.Error(err, "Failed to update avatar, please try again or contact support.")
		}

		// update player
		avatarID = ava.ID
		user.CustomAvatarID = null.StringFrom(ava.ID)
		user.ProfileAvatarID = null.String{}
	}

	user.UpdatedAt = time.Now()
	_, err = user.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue updating player avatar")
		return terror.Error(err, "Failed to update avatar, please try again or contact support.")
	}

	reply(&PlayerAvatar{
		ID:        avatarID,
		AvatarURL: avtarURL,
		Tier:      avatarTier,
		IsCustom:  req.Payload.IsCustom,
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

func (pc *PlayerController) ProfileCustomAvatarDetailsHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	avatarID := cctx.URLParam("avatar_id")
	if avatarID == "" {
		return terror.Error(fmt.Errorf("missing weapon id"), "Missing weapon id.")
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
