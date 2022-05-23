package api

import (
	"context"
	"encoding/json"
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
	api.SecureUserCommand(HubKeyPlayerAssetMechDetail, pac.PlayerAssetMechDetail)

	return pac
}

const HubKeyPlayerAssetMechList = "PLAYER:ASSET:MECH:LIST"

type PlayerAssetMechListRequest struct {
	Payload struct {
		Search   string                `json:"search"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Sort     *db.ListSortRequest   `json:"sort"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type PlayerAssetMech struct {
	*server.CollectionDetails
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

	total, mechs, err := db.MechList(&db.MechListOpts{
		Search:   req.Payload.Search,
		Filter:   req.Payload.Filter,
		Sort:     req.Payload.Sort,
		PageSize: req.Payload.PageSize,
		Page:     req.Payload.Page,
		OwnerID:  user.ID,
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
			FactionID:             m.FactionID,
			ModelID:               m.ModelID,
			DefaultChassisSkinID:  m.DefaultChassisSkinID,
			ChassisSkinID:         m.ChassisSkinID,
			IntroAnimationID:      m.IntroAnimationID,
			OutroAnimationID:      m.OutroAnimationID,
			PowerCoreID:           m.PowerCoreID,
			UpdatedAt:             m.UpdatedAt,
			CreatedAt:             m.CreatedAt,
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

func (pac *PlayerAssetsControllerWS) PlayerAssetMechDetail(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetMechDetailRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// get collection and check ownership
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.MechID),
		boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
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
