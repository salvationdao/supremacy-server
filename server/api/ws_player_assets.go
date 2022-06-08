package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PlayerAssetsControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerAssetsController(api *API) *PlayerAssetsControllerWS {
	pac := &PlayerAssetsControllerWS{
		API: api,
	}

	api.SecureUserCommand(HubKeyPlayerAssetMechList, pac.PlayerAssetMechListHandler)
	api.SecureUserFactionCommand(HubKeyPlayerAssetMechDetail, pac.PlayerAssetMechDetail)
	api.SecureUserCommand(HubKeyPlayerAssetKeycardList, pac.PlayerAssetKeycardListHandler)
	return pac
}

const HubKeyPlayerAssetMechList = "PLAYER:ASSET:MECH:LIST"

type PlayerAssetMechListRequest struct {
	Payload struct {
		Search              string                `json:"search"`
		Filter              *db.ListFilterRequest `json:"filter"`
		Sort                *db.ListSortRequest   `json:"sort"`
		PageSize            int                   `json:"page_size"`
		Page                int                   `json:"page"`
		DisplayXsynMechs    bool                  `json:"display_xsyn_mechs"`
		ExcludeMarketListed bool                  `json:"exclude_market_listed"`
	} `json:"payload"`
}

type PlayerAssetMech struct {
	CollectionSlug   string      `json:"collection_slug"`
	Hash             string      `json:"hash"`
	TokenID          int64       `json:"token_id"`
	ItemType         string      `json:"item_type"`
	Tier             string      `json:"tier"`
	OwnerID          string      `json:"owner_id"`
	ImageURL         null.String `json:"image_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	BackgroundColor  null.String `json:"background_color,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	YoutubeURL       null.String `json:"youtube_url,omitempty"`
	MarketLocked     bool        `json:"market_locked"`
	XsynLocked       bool        `json:"xsyn_locked"`
	MarketListed     bool        `json:"market_listed"`

	ID                    string     `json:"id"`
	Label                 string     `json:"label"`
	WeaponHardpoints      int        `json:"weapon_hardpoints"`
	UtilitySlots          int        `json:"utility_slots"`
	Speed                 int        `json:"speed"`
	MaxHitpoints          int        `json:"max_hitpoints"`
	IsDefault             bool       `json:"is_default"`
	IsInsured             bool       `json:"is_insured"`
	Name                  string     `json:"name"`
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
	PowerCoreSize         string     `json:"power_core_size"`
	BlueprintID           string     `json:"blueprint_id"`
	BrandID               string     `json:"brand_id"`
	FactionID             string     `json:"faction_id"`
	ModelID               string     `json:"model_id"`

	// Connected objects
	DefaultChassisSkinID string      `json:"default_chassis_skin_id"`
	ChassisSkinID        null.String `json:"chassis_skin_id,omitempty"`
	IntroAnimationID     null.String `json:"intro_animation_id,omitempty"`
	OutroAnimationID     null.String `json:"outro_animation_id,omitempty"`
	PowerCoreID          null.String `json:"power_core_id,omitempty"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type PlayerAssetMechListResp struct {
	Total int64              `json:"total"`
	Mechs []*PlayerAssetMech `json:"mechs"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetMechListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetMechListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	total, mechs, err := db.MechList(&db.MechListOpts{
		Search:              req.Payload.Search,
		Filter:              req.Payload.Filter,
		Sort:                req.Payload.Sort,
		PageSize:            req.Payload.PageSize,
		Page:                req.Payload.Page,
		OwnerID:             user.ID,
		DisplayXsynMechs:    req.Payload.DisplayXsynMechs,
		ExcludeMarketListed: req.Payload.ExcludeMarketListed,
	})
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your War Machine assets, please try again or contact support.")
	}

	playerAssetMechs := []*PlayerAssetMech{}

	for _, m := range mechs {
		playerAssetMechs = append(playerAssetMechs, &PlayerAssetMech{
			ID:                    m.ID,
			Label:                 m.Label,
			WeaponHardpoints:      m.WeaponHardpoints,
			UtilitySlots:          m.UtilitySlots,
			Speed:                 m.Speed,
			MaxHitpoints:          m.MaxHitpoints,
			IsDefault:             m.IsDefault,
			IsInsured:             m.IsInsured,
			Name:                  m.Name,
			GenesisTokenID:        m.GenesisTokenID,
			LimitedReleaseTokenID: m.LimitedReleaseTokenID,
			PowerCoreSize:         m.PowerCoreSize,
			BlueprintID:           m.BlueprintID,
			BrandID:               m.BrandID,
			FactionID:             m.FactionID.String,
			ModelID:               m.ModelID,
			DefaultChassisSkinID:  m.DefaultChassisSkinID,
			ChassisSkinID:         m.ChassisSkinID,
			IntroAnimationID:      m.IntroAnimationID,
			OutroAnimationID:      m.OutroAnimationID,
			PowerCoreID:           m.PowerCoreID,
			UpdatedAt:             m.UpdatedAt,
			CreatedAt:             m.CreatedAt,
			CollectionSlug:        m.CollectionItem.CollectionSlug,
			Hash:                  m.CollectionItem.Hash,
			TokenID:               m.CollectionItem.TokenID,
			ItemType:              m.CollectionItem.ItemType,
			Tier:                  m.CollectionItem.Tier,
			OwnerID:               m.CollectionItem.OwnerID,
			XsynLocked:            m.CollectionItem.XsynLocked,
			MarketLocked:          m.CollectionItem.MarketLocked,
			MarketListed:          m.CollectionItem.MarketListed,
			ImageURL:              m.CollectionItem.ImageURL,
			CardAnimationURL:      m.CollectionItem.CardAnimationURL,
			AvatarURL:             m.CollectionItem.AvatarURL,
			LargeImageURL:         m.CollectionItem.LargeImageURL,
			BackgroundColor:       m.CollectionItem.BackgroundColor,
			AnimationURL:          m.CollectionItem.AnimationURL,
			YoutubeURL:            m.CollectionItem.YoutubeURL,
		})
	}

	reply(&PlayerAssetMechListResp{
		Total: total,
		Mechs: playerAssetMechs,
	})
	return nil
}

type PlayerAssetMechDetailRequest struct {
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

const HubKeyPlayerAssetMechDetail = "PLAYER:ASSET:MECH:DETAIL"

func (pac *PlayerAssetsControllerWS) PlayerAssetMechDetail(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetMechDetailRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// get collection and check ownership
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.MechID),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.Players,
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
			),
		),
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(fID)),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find mech from the collection")
	}

	// get mech
	mech, err := db.Mech(collectionItem.ItemID)
	if err != nil {
		return terror.Error(err, "Failed to find mech from db")
	}

	reply(mech)
	return nil
}

const HubKeyPlayerAssetKeycardList = "PLAYER:ASSET:KEYCARD:LIST"

type PlayerAssetKeycardListRequest struct {
	Payload struct {
		Search              string                `json:"search"`
		Filter              *db.ListFilterRequest `json:"filter"`
		Sort                *db.ListSortRequest   `json:"sort"`
		PageSize            int                   `json:"page_size"`
		Page                int                   `json:"page"`
		SortDir             db.SortByDir          `json:"sort_dir"`
		SortBy              string                `json:"sort_by"`
		ExcludeMarketListed bool                  `json:"exclude_market_listed"`
	} `json:"payload"`
}

type PlayerAssetKeycardListResponse struct {
	Total    int64                  `json:"total"`
	Keycards []*server.AssetKeycard `json:"keycards"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetKeycardListHandler(tx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetKeycardListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !req.Payload.SortDir.IsValid() {
		req.Payload.SortDir = db.SortByDirDesc
	}

	total, records, err := db.PlayerKeycardList(
		req.Payload.Search,
		req.Payload.Filter,
		req.Payload.ExcludeMarketListed,
		&user.ID,
		req.Payload.Page,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of keycard assets")
		return terror.Error(err, "Failed to get list of keycard assets")
	}

	resp := &PlayerAssetKeycardListResponse{
		Total:    total,
		Keycards: records,
	}
	reply(resp)

	return nil
}
