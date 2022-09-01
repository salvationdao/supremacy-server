package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/rpctypes"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"

	"github.com/kevinms/leakybucket-go"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	goaway "github.com/TwiN/go-away"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
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

	api.SecureUserCommand(HubKeyPlayerAssetWeaponList, pac.PlayerAssetWeaponListHandler)
	api.SecureUserCommand(HubKeyPlayerAssetMysteryCrateList, pac.PlayerAssetMysteryCrateListHandler)
	api.SecureUserCommand(HubKeyPlayerAssetMysteryCrateGet, pac.PlayerAssetMysteryCrateGetHandler)
	api.SecureUserCommand(HubKeyPlayerAssetKeycardList, pac.PlayerAssetKeycardListHandler)
	api.SecureUserCommand(HubKeyPlayerAssetKeycardGet, pac.PlayerAssetKeycardGetHandler)
	api.SecureUserCommand(HubKeyPlayerAssetRename, pac.PlayerMechRenameHandler)
	api.SecureUserCommand(HubKeyplayerAssetMechSubmodelList, pac.playerAssetMechSubmodelListHandler)
	api.SecureUserCommand(HubKeyPlayerMechBlueprintList, pac.playerMechBlueprintListHandler)
	api.SecureUserCommand(HubKeyplayerAssetWeaponSubmodelList, pac.playerAssetWeaponSubmodelListHandler)
	api.SecureUserCommand(HubKeyPlayerWeaponBlueprintList, pac.playerWeaponBlueprintListHandler)
	api.SecureUserFactionCommand(HubKeyOpenCrate, pac.OpenCrateHandler)
	// public profile
	api.Command(HubKeyPlayerAssetMechListPublic, pac.PlayerAssetMechListPublicHandler)

	return pac
}

const HubKeyPlayerAssetMechList = "PLAYER:ASSET:MECH:LIST"

type PlayerAssetMechListRequest struct {
	Payload struct {
		Search              string                `json:"search"`
		Filter              *db.ListFilterRequest `json:"filter"`
		SortBy              string                `json:"sort_by"`
		SortDir             db.SortByDir          `json:"sort_dir"`
		PageSize            int                   `json:"page_size"`
		Page                int                   `json:"page"`
		DisplayXsynMechs    bool                  `json:"display_xsyn_mechs"`
		ExcludeMarketLocked bool                  `json:"exclude_market_locked"`
		IncludeMarketListed bool                  `json:"include_market_listed"`
		ExcludeDamagedMech  bool                  `json:"exclude_damaged_mech"`
		QueueSort           db.SortByDir          `json:"queue_sort"`
		FilterRarities      []string              `json:"rarities"`
		FilterStatuses      []string              `json:"statuses"`
	} `json:"payload"`
}

type PlayerAssetMech struct {
	CollectionSlug      string   `json:"collection_slug"`
	Hash                string   `json:"hash"`
	TokenID             int64    `json:"token_id"`
	ItemType            string   `json:"item_type"`
	Tier                string   `json:"tier"`
	OwnerID             string   `json:"owner_id"`
	MarketLocked        bool     `json:"market_locked"`
	XsynLocked          bool     `json:"xsyn_locked"`
	LockedToMarketplace bool     `json:"locked_to_marketplace"`
	QueuePosition       null.Int `json:"queue_position"`

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

	// Connected objects
	ChassisSkinID    string      `json:"chassis_skin_id"`
	IntroAnimationID null.String `json:"intro_animation_id,omitempty"`
	OutroAnimationID null.String `json:"outro_animation_id,omitempty"`
	PowerCoreID      null.String `json:"power_core_id,omitempty"`

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

	listOpts := &db.MechListOpts{
		Search:              req.Payload.Search,
		Filter:              req.Payload.Filter,
		PageSize:            req.Payload.PageSize,
		Page:                req.Payload.Page,
		OwnerID:             user.ID,
		DisplayXsynMechs:    req.Payload.DisplayXsynMechs,
		ExcludeMarketLocked: req.Payload.ExcludeMarketLocked,
		IncludeMarketListed: req.Payload.IncludeMarketListed,
		ExcludeDamagedMech:  req.Payload.ExcludeDamagedMech,
		FilterRarities:      req.Payload.FilterRarities,
		FilterStatuses:      req.Payload.FilterStatuses,
	}
	if req.Payload.QueueSort.IsValid() && user.FactionID.Valid {
		listOpts.QueueSort = &db.MechListQueueSortOpts{
			FactionID: user.FactionID.String,
			SortDir:   req.Payload.QueueSort,
		}
	} else if req.Payload.SortBy != "" && req.Payload.SortDir.IsValid() {
		listOpts.SortBy = req.Payload.SortBy
		listOpts.SortDir = req.Payload.SortDir
	}

	total, mechs, err := db.MechList(listOpts)
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
			LockedToMarketplace:   m.CollectionItem.LockedToMarketplace,
			QueuePosition:         m.QueuePosition,
		})
	}

	reply(&PlayerAssetMechListResp{
		Total: total,
		Mechs: playerAssetMechs,
	})
	return nil
}

const HubKeyPlayerAssetMechListPublic = "PLAYER:ASSET:MECH:LIST:PUBLIC"

type PlayerAssetMechListPublicRequest struct {
	Payload struct {
		PlayerID            string                `json:"player_id"`
		Search              string                `json:"search"`
		Filter              *db.ListFilterRequest `json:"filter"`
		Sort                *db.ListSortRequest   `json:"sort"`
		PageSize            int                   `json:"page_size"`
		Page                int                   `json:"page"`
		DisplayXsynMechs    bool                  `json:"display_xsyn_mechs"`
		ExcludeMarketLocked bool                  `json:"exclude_market_locked"`
		IncludeMarketListed bool                  `json:"include_market_listed"`
		QueueSort           db.SortByDir          `json:"queue_sort"`
	} `json:"payload"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetMechListPublicHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetMechListPublicRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// get player
	player, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.PlayerID)
	if err != nil {
		return terror.Error(fmt.Errorf("cant find player"), "Failed to fetch player.")
	}

	listOpts := &db.MechListOpts{
		Search:              req.Payload.Search,
		Filter:              req.Payload.Filter,
		Sort:                req.Payload.Sort,
		PageSize:            req.Payload.PageSize,
		Page:                req.Payload.Page,
		OwnerID:             player.ID,
		DisplayXsynMechs:    req.Payload.DisplayXsynMechs,
		ExcludeMarketLocked: req.Payload.ExcludeMarketLocked,
		IncludeMarketListed: req.Payload.IncludeMarketListed,
	}

	total, mechs, err := db.MechList(listOpts)
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
			LockedToMarketplace:   m.CollectionItem.LockedToMarketplace,
			QueuePosition:         m.QueuePosition,
		})
	}

	reply(&PlayerAssetMechListResp{
		Total: total,
		Mechs: playerAssetMechs,
	})
	return nil
}

const HubKeyPlayerAssetMechDetail = "PLAYER:ASSET:MECH:DETAIL"

func (pac *PlayerAssetsControllerWS) PlayerAssetMechDetail(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	mechID := cctx.URLParam("mech_id")
	if mechID == "" {
		return terror.Error(fmt.Errorf("missing mech id"), "Missing mech id.")
	}

	// get collection and check ownership
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
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
	mech, err := db.Mech(gamedb.StdConn, collectionItem.ItemID)
	if err != nil {
		return terror.Error(err, "Failed to find mech from db")
	}

	if mech.ChassisSkin.Images == nil {
		mech.ChassisSkin.Images = mech.Images
	}

	reply(mech)
	return nil
}

// PlayerAssetMechBriefInfo load brief mech info for quick deploy
func (pac *PlayerAssetsControllerWS) PlayerAssetMechBriefInfo(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	mechID := chi.RouteContext(ctx).URLParam("mech_id")
	if mechID == "" {
		return terror.Error(fmt.Errorf("missing mech id"), "Missing mech id.")
	}

	// get collection and check ownership
	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
		boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find mech from the collection")
	}

	mech, err := boiler.Mechs(
		boiler.MechWhere.ID.EQ(mechID),
		qm.Load(boiler.MechRels.ChassisSkin),
		qm.Load(qm.Rels(boiler.MechRels.ChassisSkin, boiler.MechSkinRels.Blueprint)),
		qm.Load(boiler.MechRels.Blueprint),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load mech info")
	}

	mechSkin, err := boiler.MechModelSkinCompatibilities(
		boiler.MechModelSkinCompatibilityWhere.MechModelID.EQ(mech.R.Blueprint.ID),
		boiler.MechModelSkinCompatibilityWhere.BlueprintMechSkinID.EQ(mech.R.ChassisSkin.BlueprintID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load mech info")
	}

	if mech.R != nil && mech.R.ChassisSkin != nil && mech.R.ChassisSkin.R != nil && mech.R.ChassisSkin.R.Blueprint != nil {
		ci.Tier = mech.R.ChassisSkin.R.Blueprint.Tier
	}

	m := server.Mech{
		ID:             mech.ID,
		Label:          mech.R.Blueprint.Label,
		Name:           mech.Name,
		RepairBlocks:   mech.R.Blueprint.RepairBlocks,
		CollectionItem: server.CollectionItemFromBoiler(ci),
		Images: &server.Images{
			AvatarURL: mechSkin.AvatarURL,
			ImageURL:  mechSkin.ImageURL,
		},
		ChassisSkin: &server.MechSkin{
			CollectionItem: server.CollectionItemFromBoiler(ci),
			Images: &server.Images{
				AvatarURL: mechSkin.AvatarURL,
				ImageURL:  mechSkin.ImageURL,
			},
		},
	}

	reply(m)
	return nil
}

const HubKeyPlayerAssetMechDetailPublic = "PLAYER:ASSET:MECH:DETAIL:PUBLIC"

func (pac *PlayerAssetsControllerWS) PlayerAssetMechDetailPublic(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	mechID := cctx.URLParam("mech_id")
	if mechID == "" {
		return terror.Error(fmt.Errorf("missing mech id"), "Missing mech id.")
	}

	// get collection and check ownership
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find mech from the collection")
	}

	// get mech
	mech, err := db.Mech(gamedb.StdConn, collectionItem.ItemID)
	if err != nil {
		return terror.Error(err, "Failed to find mech from db")
	}
	mech.ChassisSkin.Images = mech.Images

	reply(mech)
	return nil
}

const HubKeyPlayerAssetWeaponDetail = "PLAYER:ASSET:WEAPON:DETAIL"

func (pac *PlayerAssetsControllerWS) PlayerAssetWeaponDetail(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	weaponID := cctx.URLParam("weapon_id")
	if weaponID == "" {
		return terror.Error(fmt.Errorf("missing weapon id"), "Missing weapon id.")
	}
	// get collection and check ownership
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
		boiler.CollectionItemWhere.ItemID.EQ(weaponID),
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
		return terror.Error(err, "Failed to find weapon from the collection")
	}

	// get weapon
	weapon, err := db.Weapon(gamedb.StdConn, collectionItem.ItemID)
	if err != nil {
		return terror.Error(err, "Failed to find weapon from db")
	}

	reply(weapon)
	return nil
}

const HubKeyPlayerAssetMysteryCrateList = "PLAYER:ASSET:MYSTERY_CRATE:LIST"

type PlayerAssetMysteryCrateListRequest struct {
	Payload struct {
		Search              string              `json:"search"`
		Sort                *db.ListSortRequest `json:"sort"`
		PageSize            int                 `json:"page_size"`
		Page                int                 `json:"page"`
		SortDir             db.SortByDir        `json:"sort_dir"`
		SortBy              string              `json:"sort_by"`
		ExcludeOpened       bool                `json:"exclude_opened"`
		IncludeMarketListed bool                `json:"include_market_listed"`
		ExcludeMarketLocked bool                `json:"exclude_market_locked"`
	} `json:"payload"`
}

type PlayerAssetMysteryCrateListResponse struct {
	Total         int64                  `json:"total"`
	MysteryCrates []*server.MysteryCrate `json:"mystery_crates"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetMysteryCrateListHandler(tx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetMysteryCrateListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !req.Payload.SortDir.IsValid() {
		req.Payload.SortDir = db.SortByDirDesc
	}

	total, records, err := db.PlayerMysteryCrateList(
		req.Payload.Search,
		req.Payload.ExcludeOpened,
		req.Payload.IncludeMarketListed,
		req.Payload.ExcludeMarketLocked,
		&user.ID,
		req.Payload.Page,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of mystery crate assets")
		return terror.Error(err, "Failed to get list of mystery crate assets")
	}

	resp := &PlayerAssetMysteryCrateListResponse{
		Total:         total,
		MysteryCrates: records,
	}
	reply(resp)

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
		IncludeMarketListed bool                  `json:"include_market_listed"`
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
		req.Payload.IncludeMarketListed,
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

const (
	HubKeyPlayerAssetMysteryCrateGet = "PLAYER:ASSET:MYSTERY_CRATE:GET"
	HubKeyPlayerAssetKeycardGet      = "PLAYER:ASSET:KEYCARD:GET"
)

type PlayerAssetGetRequest struct {
	Payload struct {
		ID uuid.UUID `json:"id"`
	} `json:"payload"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetMysteryCrateGetHandler(tx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	crate, err := db.PlayerMysteryCrate(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Mystery Crate not found.")
	}
	if err != nil {
		return terror.Error(err, "Failed to get Mystery Crate.")
	}

	reply(crate)

	return nil
}

func (pac *PlayerAssetsControllerWS) PlayerAssetKeycardGetHandler(tx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	keycard, err := db.PlayerKeycard(req.Payload.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Keycard not found.")
	}
	if err != nil {
		return terror.Error(err, "Failed to get keycard.")
	}

	reply(keycard)

	return nil
}

const (
	HubKeyPlayerAssetRename = "PLAYER:MECH:RENAME"
)

type PlayerMechRenameRequest struct {
	Payload struct {
		MechID  uuid.UUID `json:"mech_id"`
		NewName string    `json:"new_name"`
	} `json:"payload"`
}

func (pac *PlayerAssetsControllerWS) PlayerMechRenameHandler(tx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerMechRenameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// check valid name
	err = IsValidMechName(req.Payload.NewName)
	if err != nil {
		return terror.Error(err, "Invalid mech name")
	}

	mech, err := db.MechRename(req.Payload.MechID.String(), user.ID, req.Payload.NewName)
	if err != nil {
		return terror.Error(err, "Failed to rename mech")
	}

	reply(mech)
	return nil
}

// PrintableLen counts how many printable characters are in a string.
func PrintableLen(s string) int {
	sLen := 0
	runes := []rune(s)
	for _, r := range runes {
		if unicode.IsPrint(r) {
			sLen += 1
		}
	}
	return sLen
}

var UsernameRegExp = regexp.MustCompile("[`~!@#$%^&*()+=\\[\\]{};':\"\\|,.<>\\/?]")

func IsValidMechName(name string) error {
	// Must contain at least 3 characters
	// Cannot contain more than 15 characters
	// Cannot contain profanity
	// Can only contain the following symbols: _
	hasDisallowedSymbol := false
	if UsernameRegExp.Match([]byte(name)) {
		hasDisallowedSymbol = true
	}

	//err := fmt.Errorf("username does not meet requirements")
	if TrimName(name) == "" {
		return terror.Error(fmt.Errorf("name cannot be empty"), "Invalid name. Your name cannot be empty.")
	}
	if PrintableLen(TrimName(name)) < 3 {
		return terror.Error(fmt.Errorf("name must be at least characters long"), "Invalid name. Your name must be at least 3 characters long.")
	}
	if PrintableLen(TrimName(name)) > 30 {
		return terror.Error(fmt.Errorf("name cannot be more than 30 characters long"), "Invalid name. Your name cannot be more than 30 characters long.")
	}
	if hasDisallowedSymbol {
		return terror.Error(fmt.Errorf("name cannot contain disallowed symbols"), "Invalid name. Your name contains a disallowed symbol.")
	}

	profanityDetector := goaway.NewProfanityDetector()
	profanityDetector = profanityDetector.WithSanitizeLeetSpeak(false)

	if profanityDetector.IsProfane(name) {
		return terror.Error(fmt.Errorf("name contains profanity"), "Invalid name. Your name contains profanity.")
	}

	return nil
}

// TrimName removes misuse of invisible characters.
func TrimName(username string) string {
	// Check if entire string is nothing not non-printable characters
	isEmpty := true
	runes := []rune(username)
	for _, r := range runes {
		if unicode.IsPrint(r) && !unicode.IsSpace(r) {
			isEmpty = false
			break
		}
	}
	if isEmpty {
		return ""
	}

	// Remove Spaces like characters Around String (keep mark ones)
	output := strings.Trim(username, " \u00A0\u180E\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u200B\u202F\u205F\u3000\uFEFF\u2423\u2422\u2420")

	return output
}

type OpenCrateRequest struct {
	Payload struct {
		Id       string `json:"id"`
		IsHangar bool   `json:"is_hangar"`
	} `json:"payload"`
}

type OpenCrateResponse struct {
	ID          string               `json:"id"`
	Mech        *server.Mech         `json:"mech,omitempty"`
	MechSkins   []*server.MechSkin   `json:"mech_skins,omitempty"`
	Weapons     []*server.Weapon     `json:"weapon,omitempty"`
	WeaponSkins []*server.WeaponSkin `json:"weapon_skins,omitempty"`
	PowerCore   *server.PowerCore    `json:"power_core,omitempty"`
}

const HubKeyOpenCrate = "CRATE:OPEN"

var openCrateBucket = leakybucket.NewCollector(1, 1, true)

func (pac *PlayerAssetsControllerWS) OpenCrateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	v := openCrateBucket.Add(user.ID, 1)
	if v == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Currently handling request, please try again.")
	}

	req := &OpenCrateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	var collectionItem *boiler.CollectionItem
	if req.Payload.IsHangar {
		collectionItem, err = boiler.CollectionItems(
			boiler.CollectionItemWhere.ID.EQ(req.Payload.Id),
			boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
		).One(gamedb.StdConn)
	} else {
		collectionItem, err = boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.EQ(req.Payload.Id),
			boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
		).One(gamedb.StdConn)
	}
	if err != nil {
		return terror.Error(err, "Could not find collection item, try again or contact support.")
	}

	//checks
	if collectionItem.OwnerID != user.ID {
		return terror.Error(fmt.Errorf("user: %s attempted to open crate: %s belonging to owner: %s", user.ID, req.Payload.Id, collectionItem.OwnerID), "This crate does not belong to this user, try again or contact support.")
	}
	if collectionItem.MarketLocked {
		return terror.Error(fmt.Errorf("user: %s attempted to open crate: %s while market locked", user.ID, req.Payload.Id), "This crate is still on Marketplace, try again or contact support.")
	}
	if collectionItem.XsynLocked {
		return terror.Error(fmt.Errorf("user: %s attempted to open crate: %s while XSYN locked", user.ID, req.Payload.Id), "This crate is locked to XSYN, move asset to Supremacy and try again.")
	}
	if collectionItem.LockedToMarketplace {
		return terror.Error(fmt.Errorf("user: %s attempted to open crate: %s while market locked", user.ID, req.Payload.Id), "This crate is still on Marketplace, try again or contact support.")
	}

	crate := boiler.MysteryCrate{}

	q := `
		UPDATE mystery_crate
		SET opened = TRUE
		WHERE id = $1 AND opened = FALSE AND locked_until <= NOW()
		RETURNING id, type, faction_id, label, opened, locked_until, purchased, deleted_at, updated_at, created_at, description`
	err = gamedb.StdConn.QueryRow(q, collectionItem.ItemID).
		Scan(
			&crate.ID,
			&crate.Type,
			&crate.FactionID,
			&crate.Label,
			&crate.Opened,
			&crate.LockedUntil,
			&crate.Purchased,
			&crate.DeletedAt,
			&crate.UpdatedAt,
			&crate.CreatedAt,
			&crate.Description,
		)
	if err != nil {
		return terror.Error(err, "Could not find crate, try again or contact support.")
	}

	crateRollback := func() {
		crate.Opened = false
		_, err = crate.Update(gamedb.StdConn, boil.Whitelist(boiler.MysteryCrateColumns.Opened))
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed rollback crate opened: %s", crate.ID))
		}
	}

	items := OpenCrateResponse{}
	items.ID = req.Payload.Id

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return fmt.Errorf("start tx: %w", err)
	}
	defer tx.Rollback()

	blueprintItems, err := crate.MysteryCrateBlueprints().All(gamedb.StdConn)
	if err != nil {
		crateRollback()
		gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get blueprint relationships from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
		return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
	}

	xsynAsserts := []*rpctypes.XsynAsset{}

	blueprintMechs := []string{}
	blueprintMechSkins := []string{}
	blueprintWeapons := []string{}
	blueprintWeaponSkins := []string{}
	blueprintPowercores := []string{}

	for _, blueprintItem := range blueprintItems {
		switch blueprintItem.BlueprintType {
		case boiler.TemplateItemTypeMECH:
			blueprintMechs = append(blueprintMechs, blueprintItem.BlueprintID)
		case boiler.TemplateItemTypeWEAPON:
			blueprintWeapons = append(blueprintWeapons, blueprintItem.BlueprintID)
		case boiler.TemplateItemTypeMECH_SKIN:
			blueprintMechSkins = append(blueprintMechSkins, blueprintItem.BlueprintID)
		case boiler.TemplateItemTypeWEAPON_SKIN:
			blueprintWeaponSkins = append(blueprintWeaponSkins, blueprintItem.BlueprintID)
		case boiler.TemplateItemTypePOWER_CORE:
			blueprintPowercores = append(blueprintPowercores, blueprintItem.BlueprintID)
		}
	}

	for _, blueprintItemID := range blueprintMechs {
		mechSkinBlueprints, err := db.BlueprintMechSkinSkins(tx, blueprintMechSkins)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Msg("failed BlueprintMechSkinSkins")
			return terror.Error(err, "Could not get weapon blueprint during crate opening, try again or contact support.")
		}

		// insert the non default skin with the mech
		rarerSkinIndex := 0
		for i, skin := range mechSkinBlueprints {
			if skin.Tier != "COLOSSAL" {
				rarerSkinIndex = i
			}
		}

		mechBlueprint, err := db.BlueprintMech(blueprintItemID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("mechBlueprint", mechBlueprint).Msg(fmt.Sprintf("failed to get mech blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
		}

		insertedMech, insertedMechSkin, err := db.InsertNewMechAndSkin(tx, uuid.FromStringOrNil(user.ID), mechBlueprint, mechSkinBlueprints[rarerSkinIndex])
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to insert new mech for user")
			return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
		}
		items.Mech = insertedMech
		items.MechSkins = append(items.MechSkins, insertedMechSkin)

		// remove the already inserted skin
		mechSkinBlueprints = append(mechSkinBlueprints[:rarerSkinIndex], mechSkinBlueprints[rarerSkinIndex+1:]...)

		// insert the rest of the skins
		for _, skin := range mechSkinBlueprints {
			mechSkin, err := db.InsertNewMechSkin(tx, uuid.FromStringOrNil(user.ID), skin, &insertedMech.BlueprintID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("skin", skin).Msg(fmt.Sprintf("failed to insert new mech skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get mech skin during crate opening, try again or contact support.")
			}
			items.MechSkins = append(items.MechSkins, mechSkin)
		}
	}

	for _, blueprintItemID := range blueprintWeapons {
		weaponSkinBlueprints, err := db.BlueprintWeaponSkins(blueprintWeaponSkins)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Msg("failed BlueprintWeaponSkins")
			return terror.Error(err, "Could not get weapon blueprint during crate opening, try again or contact support.")
		}

		// insert the non default skin with the mech
		rarerSkinIndex := 0
		for i, skin := range weaponSkinBlueprints {
			if skin.Tier != "COLOSSAL" {
				rarerSkinIndex = i
			}
		}

		bp, err := db.BlueprintWeapon(blueprintItemID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("blueprintItemID", blueprintItemID).Msg(fmt.Sprintf("failed to get weapon blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get weapon blueprint during crate opening, try again or contact support.")
		}

		weapon, weaponSkin, err := db.InsertNewWeapon(tx, uuid.FromStringOrNil(user.ID), bp, weaponSkinBlueprints[rarerSkinIndex])
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("blueprintItemID", blueprintItemID).Msg(fmt.Sprintf("failed to insert new weapon from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get weapon during crate opening, try again or contact support.")
		}
		items.Weapons = append(items.Weapons, weapon)
		items.WeaponSkins = append(items.WeaponSkins, weaponSkin)

		for i, bpws := range blueprintWeaponSkins {
			if bpws == weaponSkin.BlueprintID {
				blueprintWeaponSkins = append(blueprintWeaponSkins[:i], blueprintWeaponSkins[i+1:]...)
				break
			}
		}
	}
	for _, blueprintItemID := range blueprintWeaponSkins {
		bp, err := db.BlueprintWeaponSkin(blueprintItemID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("blueprintItemID", blueprintItemID).Msg(fmt.Sprintf("failed to get weapon skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get weapon skin blueprint during crate opening, try again or contact support.")
		}
		weaponSkin, err := db.InsertNewWeaponSkin(tx, uuid.FromStringOrNil(user.ID), bp, nil)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("blueprintItemID", blueprintItemID).Msg(fmt.Sprintf("failed to insert new weapon skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get weapon skin during crate opening, try again or contact support.")
		}
		items.WeaponSkins = append(items.WeaponSkins, weaponSkin)
	}
	for _, blueprintItemID := range blueprintPowercores {
		bp, err := db.BlueprintPowerCore(blueprintItemID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("blueprintItemID", blueprintItemID).Msg(fmt.Sprintf("failed to get powercore blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get powercore blueprint during crate opening, try again or contact support.")
		}

		powerCore, err := db.InsertNewPowerCore(tx, uuid.FromStringOrNil(user.ID), bp)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Interface("blueprintItemIDt", blueprintItemID).Msg(fmt.Sprintf("failed to insert new powercore from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
			return terror.Error(err, "Could not get powercore during crate opening, try again or contact support.")
		}
		items.PowerCore = powerCore
	}

	var hangarResp *db.SiloType
	if crate.Type == boiler.CrateTypeMECH {
		eod, err := db.MechEquippedOnDetails(tx, items.Mech.ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get MechEquippedOnDetails during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}

		rarerSkin := items.MechSkins[0]
		for _, skin := range items.MechSkins {
			if skin.Tier != "COLOSSAL" {
				rarerSkin = skin
			}
		}

		rarerSkin.EquippedOn = null.StringFrom(items.Mech.ID)
		rarerSkin.EquippedOnDetails = eod
		xsynAsserts = append(xsynAsserts, rpctypes.ServerMechSkinsToXsynAsset(items.MechSkins)...)

		err = db.AttachPowerCoreToMech(tx, user.ID, items.Mech.ID, items.PowerCore.ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach powercore to mech during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		items.PowerCore.EquippedOn = null.StringFrom(items.Mech.ID)
		items.PowerCore.EquippedOnDetails = eod
		xsynAsserts = append(xsynAsserts, rpctypes.ServerPowerCoresToXsynAsset([]*server.PowerCore{items.PowerCore})...)

		//attach weapons to mech -mech
		for i, weapon := range items.Weapons {
			err = db.AttachWeaponToMech(tx, user.ID, items.Mech.ID, weapon.ID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapons to mech during CRATE:OPEN crate: %s", crate.ID))
				return terror.Error(err, "Could not open crate, try again or contact support.")
			}
			weapon.EquippedOn = null.StringFrom(items.Mech.ID)
			weapon.EquippedOnDetails = eod

			wod, err := db.WeaponEquippedOnDetails(tx, items.Weapons[0].ID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get WeaponEquippedOnDetails during CRATE:OPEN crate: %s", crate.ID))
				return terror.Error(err, "Could not open crate, try again or contact support.")
			}

			weapon.WeaponSkin = items.WeaponSkins[i]
			weapon.WeaponSkin.EquippedOn = null.StringFrom(items.Weapons[i].ID)
			weapon.WeaponSkin.EquippedOnDetails = wod
			xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponSkinsToXsynAsset([]*server.WeaponSkin{items.WeaponSkins[i]})...)
		}
		xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponsToXsynAsset(items.Weapons)...)

		mech, err := db.Mech(tx, items.Mech.ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get final mech during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		mech.ChassisSkin = rarerSkin
		xsynAsserts = append(xsynAsserts, rpctypes.ServerMechsToXsynAsset([]*server.Mech{mech})...)

		if req.Payload.IsHangar {
			hangarResp, err = db.GetUserMechHangarItemsWithMechID(mech, user.ID, tx)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Msg("Failed to get mech hangar items while opening crate")
				return terror.Error(err, "Failed to get user mech hangar from items")
			}
		}
	}

	if crate.Type == boiler.CrateTypeWEAPON {
		wod, err := db.WeaponEquippedOnDetails(tx, items.Weapons[0].ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get WeaponEquippedOnDetails during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}

		//attach weapon_skin to weapon -weapon
		if len(items.Weapons) != 1 {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("too many weapons in crate: %s", crate.ID))
			return terror.Error(fmt.Errorf("too many weapons in weapon crate"), "Could not open crate, try again or contact support.")
		}

		items.WeaponSkins[0].EquippedOn = null.StringFrom(items.Weapons[0].ID)
		items.WeaponSkins[0].EquippedOnDetails = wod
		xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponSkinsToXsynAsset([]*server.WeaponSkin{items.WeaponSkins[0]})...)

		weapon, err := db.Weapon(tx, items.Weapons[0].ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get final mech during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponsToXsynAsset([]*server.Weapon{weapon})...)

		if req.Payload.IsHangar {
			hangarResp, err = db.GetUserWeaponHangarItemsWithID(weapon, user.ID, tx)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Msg("Failed to get weapon hangar items while opening crate")
				return terror.Error(err, "Failed to get user mech hangar from items")
			}
		}
	}

	err = pac.API.Passport.AssetsRegister(xsynAsserts) // register new assets
	if err != nil {
		gamelog.L.Error().Err(err).Msg("issue inserting new mechs to xsyn for RegisterAllNewAssets")
		crateRollback()
		return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
	}

	// delete crate on xsyn
	err = pac.API.Passport.DeleteAssetXSYN(crate.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("issue inserting new mechs to xsyn for RegisterAllNewAssets - DeleteAssetXSYN")
		crateRollback()
		return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		crateRollback()
		gamelog.L.Error().Err(err).Interface("crate", crate).Msg("failed to open mystery crate")
		return terror.Error(err, "Could not open mystery crate, please try again or contact support.")
	}

	if req.Payload.IsHangar {
		reply(hangarResp)
		return nil
	}

	reply(items)

	return nil
}

const HubKeyPlayerAssetWeaponList = "PLAYER:ASSET:WEAPON:LIST"

type PlayerAssetWeaponListRequest struct {
	Payload struct {
		Search                        string                    `json:"search"`
		Filter                        *db.ListFilterRequest     `json:"filter"`
		Sort                          *db.ListSortRequest       `json:"sort"`
		SortBy                        string                    `json:"sort_by"`
		SortDir                       db.SortByDir              `json:"sort_dir"`
		PageSize                      int                       `json:"page_size"`
		Page                          int                       `json:"page"`
		DisplayXsynMechs              bool                      `json:"display_xsyn_mechs"`
		ExcludeMarketLocked           bool                      `json:"exclude_market_locked"`
		IncludeMarketListed           bool                      `json:"include_market_listed"`
		FilterRarities                []string                  `json:"rarities"`
		FilterWeaponTypes             []string                  `json:"weapon_types"`
		FilterEquippedStatuses        []string                  `json:"equipped_statuses"`
		FilterStatAmmo                *db.WeaponStatFilterRange `json:"stat_ammo"`
		FilterStatDamage              *db.WeaponStatFilterRange `json:"stat_damage"`
		FilterStatDamageFalloff       *db.WeaponStatFilterRange `json:"stat_damage_falloff"`
		FilterStatDamageFalloffRate   *db.WeaponStatFilterRange `json:"stat_damage_falloff_rate"`
		FilterStatRadius              *db.WeaponStatFilterRange `json:"stat_radius"`
		FilterStatRadiusDamageFalloff *db.WeaponStatFilterRange `json:"stat_radius_damage_falloff"`
		FilterStatRateOfFire          *db.WeaponStatFilterRange `json:"stat_rate_of_fire"`
		FilterStatEnergyCosts         *db.WeaponStatFilterRange `json:"stat_energy_cost"`
		FilterStatProjectileSpeed     *db.WeaponStatFilterRange `json:"stat_projectile_speed"`
		FilterStatSpread              *db.WeaponStatFilterRange `json:"stat_spread"`
	} `json:"payload"`
}

type PlayerAssetWeaponListResp struct {
	Total   int64                `json:"total"`
	Weapons []*PlayerAssetWeapon `json:"weapons"`
}

type PlayerAssetWeapon struct {
	CollectionSlug      string `json:"collection_slug"`
	Hash                string `json:"hash"`
	TokenID             int64  `json:"token_id"`
	Tier                string `json:"tier"`
	OwnerID             string `json:"owner_id"`
	MarketLocked        bool   `json:"market_locked"`
	XsynLocked          bool   `json:"xsyn_locked"`
	LockedToMarketplace bool   `json:"locked_to_marketplace"`

	ID    string `json:"id"`
	Label string `json:"label"`
	Name  string `json:"name"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (pac *PlayerAssetsControllerWS) PlayerAssetWeaponListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetWeaponListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	listOpts := &db.WeaponListOpts{
		Search:                        req.Payload.Search,
		Filter:                        req.Payload.Filter,
		Sort:                          req.Payload.Sort,
		PageSize:                      req.Payload.PageSize,
		Page:                          req.Payload.Page,
		OwnerID:                       user.ID,
		DisplayXsynMechs:              req.Payload.DisplayXsynMechs,
		ExcludeMarketLocked:           req.Payload.ExcludeMarketLocked,
		IncludeMarketListed:           req.Payload.IncludeMarketListed,
		FilterRarities:                req.Payload.FilterRarities,
		FilterWeaponTypes:             req.Payload.FilterWeaponTypes,
		FilterEquippedStatuses:        req.Payload.FilterEquippedStatuses,
		FilterStatAmmo:                req.Payload.FilterStatAmmo,
		FilterStatDamage:              req.Payload.FilterStatDamage,
		FilterStatDamageFalloff:       req.Payload.FilterStatDamageFalloff,
		FilterStatDamageFalloffRate:   req.Payload.FilterStatDamageFalloffRate,
		FilterStatRadius:              req.Payload.FilterStatRadius,
		FilterStatRadiusDamageFalloff: req.Payload.FilterStatRadiusDamageFalloff,
		FilterStatRateOfFire:          req.Payload.FilterStatRateOfFire,
		FilterStatEnergyCosts:         req.Payload.FilterStatEnergyCosts,
		FilterStatProjectileSpeed:     req.Payload.FilterStatProjectileSpeed,
		FilterStatSpread:              req.Payload.FilterStatSpread,
	}
	if req.Payload.SortBy != "" && req.Payload.SortDir.IsValid() {
		listOpts.SortBy = req.Payload.SortBy
		listOpts.SortDir = req.Payload.SortDir
	}

	total, weapons, err := db.WeaponList(listOpts)
	if err != nil {
		gamelog.L.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting mechs")
		return terror.Error(err, "Failed to find your War Machine assets, please try again or contact support.")
	}

	playerAssWeapons := []*PlayerAssetWeapon{}

	for _, m := range weapons {
		playerAssWeapons = append(playerAssWeapons, &PlayerAssetWeapon{
			ID:                  m.ID,
			Label:               m.Label,
			UpdatedAt:           m.UpdatedAt,
			CreatedAt:           m.CreatedAt,
			CollectionSlug:      m.CollectionItem.CollectionSlug,
			Hash:                m.CollectionItem.Hash,
			TokenID:             m.CollectionItem.TokenID,
			Tier:                m.CollectionItem.Tier,
			OwnerID:             m.CollectionItem.OwnerID,
			XsynLocked:          m.CollectionItem.XsynLocked,
			MarketLocked:        m.CollectionItem.MarketLocked,
			LockedToMarketplace: m.CollectionItem.LockedToMarketplace,
		})
	}

	reply(&PlayerAssetWeaponListResp{
		Total:   total,
		Weapons: playerAssWeapons,
	})
	return nil
}

func (api *API) GetMaxWeaponStats(w http.ResponseWriter, r *http.Request) (int, error) {
	userID := r.URL.Query().Get("user_id") // the stat identifier e.g. speed

	output, err := db.GetWeaponMaxStats(gamedb.StdConn, userID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Something went wrong with fetching max weapon stats.")
	}

	// Don't put quote values in for decimal stat values
	decimal.MarshalJSONWithoutQuotes = true
	status, resp := helpers.EncodeJSON(w, output)
	decimal.MarshalJSONWithoutQuotes = false

	return status, resp
}

const HubKeyplayerAssetMechSubmodelList = "PLAYER:ASSET:MECH:SUBMODEL:LIST"

type PlayerAssetMechSubmodelListRequest struct {
	Payload struct {
		Search                   string                `json:"search"`
		Filter                   *db.ListFilterRequest `json:"filter"`
		Sort                     *db.ListSortRequest   `json:"sort"`
		SortBy                   string                `json:"sort_by"`
		SortDir                  db.SortByDir          `json:"sort_dir"`
		PageSize                 int                   `json:"page_size"`
		Page                     int                   `json:"page"`
		DisplayXsyn              bool                  `json:"display_xsyn"`
		ExcludeMarketLocked      bool                  `json:"exclude_market_locked"`
		IncludeMarketListed      bool                  `json:"include_market_listed"`
		DisplayGenesisAndLimited bool                  `json:"display_genesis_and_limited"`
		FilterRarities           []string              `json:"rarities"`
		FilterSkinCompatibility  []string              `json:"skin_compatibility"`
		FilterEquippedStatuses   []string              `json:"equipped_statuses"`
	} `json:"payload"`
}

type PlayerAssetMechSubmodelListResp struct {
	Total     int64                      `json:"total"`
	Submodels []*PlayerAssetMechSubmodel `json:"submodels"`
}

type PlayerAssetMechSubmodel struct {
	Images              *server.Images `json:"images"`
	CollectionSlug      string         `json:"collection_slug"`
	Hash                string         `json:"hash"`
	TokenID             int64          `json:"token_id"`
	Tier                string         `json:"tier"`
	OwnerID             string         `json:"owner_id"`
	MarketLocked        bool           `json:"market_locked"`
	XsynLocked          bool           `json:"xsyn_locked"`
	LockedToMarketplace bool           `json:"locked_to_marketplace"`
	Level               int            `json:"level"`

	EquippedOn string `json:"equipped_on"`
	ID         string `json:"id"`
	Label      string `json:"label"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (pac *PlayerAssetsControllerWS) playerAssetMechSubmodelListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "playerAssetMechSubmodelListHandler").Str("user_id", user.ID).Logger()

	req := &PlayerAssetMechSubmodelListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	listOpts := &db.MechSkinListOpts{
		Search:                   req.Payload.Search,
		Filter:                   req.Payload.Filter,
		Sort:                     req.Payload.Sort,
		PageSize:                 req.Payload.PageSize,
		Page:                     req.Payload.Page,
		OwnerID:                  user.ID,
		DisplayXsyn:              req.Payload.DisplayXsyn,
		ExcludeMarketLocked:      req.Payload.ExcludeMarketLocked,
		IncludeMarketListed:      req.Payload.IncludeMarketListed,
		FilterRarities:           req.Payload.FilterRarities,
		FilterEquippedStatuses:   req.Payload.FilterEquippedStatuses,
		SortBy:                   req.Payload.SortBy,
		SortDir:                  req.Payload.SortDir,
		DisplayGenesisAndLimited: req.Payload.DisplayGenesisAndLimited,
		FilterSkinCompatibility:  req.Payload.FilterSkinCompatibility,
	}

	if req.Payload.SortBy != "" && req.Payload.SortDir.IsValid() {
		listOpts.SortBy = req.Payload.SortBy
		listOpts.SortDir = req.Payload.SortDir
	}

	total, submodels, err := db.MechSkinList(listOpts)
	if err != nil {
		l.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting war machine skin list")
		return terror.Error(err, "Failed to find your war machine skin assets, please try again or contact support.")
	}

	playerAssetMechSubmodel := []*PlayerAssetMechSubmodel{}

	for _, s := range submodels {
		playerAssetMechSubmodel = append(playerAssetMechSubmodel, &PlayerAssetMechSubmodel{
			Images: &server.Images{
				ImageURL:         s.ImageURL,
				CardAnimationURL: s.CardAnimationURL,
				AvatarURL:        s.AvatarURL,
				AnimationURL:     s.AnimationURL,
				BackgroundColor:  s.BackgroundColor,
				YoutubeURL:       s.YoutubeURL,
				LargeImageURL:    s.LargeImageURL,
			},
			ID:                  s.ID,
			Label:               s.Label,
			EquippedOn:          s.EquippedOn.String,
			CreatedAt:           s.CreatedAt,
			CollectionSlug:      s.CollectionItem.CollectionSlug,
			Hash:                s.CollectionItem.Hash,
			TokenID:             s.CollectionItem.TokenID,
			Tier:                s.Tier,
			OwnerID:             s.CollectionItem.OwnerID,
			XsynLocked:          s.CollectionItem.XsynLocked,
			MarketLocked:        s.CollectionItem.MarketLocked,
			LockedToMarketplace: s.CollectionItem.LockedToMarketplace,
			Level:               s.Level,
		})
	}

	reply(&PlayerAssetMechSubmodelListResp{
		Total:     total,
		Submodels: playerAssetMechSubmodel,
	})
	return nil
}

const HubKeyPlayerMechBlueprintList = "PLAYER:MECH:BLUEPRINT:LIST"

func (pac *PlayerAssetsControllerWS) playerMechBlueprintListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "playerMechBlueprintListHandler").Str("user_id", user.ID).Logger()

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	results := []*server.BlueprintMech{}

	// need a list of blueprint mechs that the user owns
	rows, err := boiler.NewQuery(
		qm.Select(
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.DefaultChassisSkinID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BrandID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.MechType),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.RepairBlocks),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BoostStat),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.WeaponHardpoints),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.PowerCoreSize),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.UtilitySlots),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Speed),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.MaxHitpoints),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Collection),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.AvailabilityID),
		),
		qm.From(boiler.TableNames.CollectionItems),
		boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// inner join mech blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
		)),
		qm.GroupBy(qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID)),
	).Query(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("issue getting mech model list")
		return err
	}

	for rows.Next() {
		mbp := &server.BlueprintMech{}
		err = rows.Scan(
			&mbp.ID,
			&mbp.Label,
			&mbp.DefaultChassisSkinID,
			&mbp.BrandID,
			&mbp.MechType,
			&mbp.RepairBlocks,
			&mbp.BoostStat,
			&mbp.WeaponHardpoints,
			&mbp.PowerCoreSize,
			&mbp.UtilitySlots,
			&mbp.Speed,
			&mbp.MaxHitpoints,
			&mbp.Collection,
			&mbp.AvailabilityID,
		)
		if err != nil {
			l.Error().Err(err).Msg("issue parsing mech blueprints")
			return err
		}
		results = append(results, mbp)
	}

	reply(results)
	return nil
}

const HubKeyplayerAssetWeaponSubmodelList = "PLAYER:ASSET:WEAPON:SUBMODEL:LIST"

type PlayerAssetWeaponSubmodelListRequest struct {
	Payload struct {
		Search                   string                `json:"search"`
		Filter                   *db.ListFilterRequest `json:"filter"`
		Sort                     *db.ListSortRequest   `json:"sort"`
		SortBy                   string                `json:"sort_by"`
		SortDir                  db.SortByDir          `json:"sort_dir"`
		PageSize                 int                   `json:"page_size"`
		Page                     int                   `json:"page"`
		DisplayXsyn              bool                  `json:"display_xsyn"`
		ExcludeMarketLocked      bool                  `json:"exclude_market_locked"`
		IncludeMarketListed      bool                  `json:"include_market_listed"`
		DisplayGenesisAndLimited bool                  `json:"display_genesis_and_limited"`
		FilterRarities           []string              `json:"rarities"`
		FilterSkinCompatibility  []string              `json:"skin_compatibility"`
		FilterEquippedStatuses   []string              `json:"equipped_statuses"`
	} `json:"payload"`
}

type PlayerAssetWeaponSubmodelListResp struct {
	Total     int64                        `json:"total"`
	Submodels []*PlayerAssetWeaponSubmodel `json:"submodels"`
}

type PlayerAssetWeaponSubmodel struct {
	Images              *server.Images `json:"images"`
	CollectionSlug      string         `json:"collection_slug"`
	Hash                string         `json:"hash"`
	TokenID             int64          `json:"token_id"`
	Tier                string         `json:"tier"`
	OwnerID             string         `json:"owner_id"`
	MarketLocked        bool           `json:"market_locked"`
	XsynLocked          bool           `json:"xsyn_locked"`
	LockedToMarketplace bool           `json:"locked_to_marketplace"`

	EquippedOn string `json:"equipped_on"`
	ID         string `json:"id"`
	Label      string `json:"label"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (pac *PlayerAssetsControllerWS) playerAssetWeaponSubmodelListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "playerAssetWeaponSubmodelListHandler").Str("user_id", user.ID).Logger()

	req := &PlayerAssetWeaponSubmodelListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}

	listOpts := &db.WeaponSkinListOpts{
		Search:                   req.Payload.Search,
		Filter:                   req.Payload.Filter,
		Sort:                     req.Payload.Sort,
		PageSize:                 req.Payload.PageSize,
		Page:                     req.Payload.Page,
		OwnerID:                  user.ID,
		DisplayXsyn:              req.Payload.DisplayXsyn,
		ExcludeMarketLocked:      req.Payload.ExcludeMarketLocked,
		IncludeMarketListed:      req.Payload.IncludeMarketListed,
		FilterRarities:           req.Payload.FilterRarities,
		FilterEquippedStatuses:   req.Payload.FilterEquippedStatuses,
		SortBy:                   req.Payload.SortBy,
		SortDir:                  req.Payload.SortDir,
		DisplayGenesisAndLimited: req.Payload.DisplayGenesisAndLimited,
		FilterSkinCompatibility:  req.Payload.FilterSkinCompatibility,
	}

	if req.Payload.SortBy != "" && req.Payload.SortDir.IsValid() {
		listOpts.SortBy = req.Payload.SortBy
		listOpts.SortDir = req.Payload.SortDir
	}

	total, submodels, err := db.WeaponSkinList(listOpts)
	if err != nil {
		l.Error().Interface("req.Payload", req.Payload).Err(err).Msg("issue getting weapon skin list")
		return terror.Error(err, "Failed to find your weapon skin assets, please try again or contact support.")
	}

	playerAssetWeaponSubmodel := []*PlayerAssetWeaponSubmodel{}

	for _, s := range submodels {
		playerAssetWeaponSubmodel = append(playerAssetWeaponSubmodel, &PlayerAssetWeaponSubmodel{
			Images: &server.Images{
				ImageURL:         s.ImageURL,
				CardAnimationURL: s.CardAnimationURL,
				AvatarURL:        s.AvatarURL,
				AnimationURL:     s.AnimationURL,
				BackgroundColor:  s.BackgroundColor,
				YoutubeURL:       s.YoutubeURL,
				LargeImageURL:    s.LargeImageURL,
			},
			ID:                  s.ID,
			Label:               s.Label,
			EquippedOn:          s.EquippedOn.String,
			CreatedAt:           s.CreatedAt,
			CollectionSlug:      s.CollectionItem.CollectionSlug,
			Hash:                s.CollectionItem.Hash,
			TokenID:             s.CollectionItem.TokenID,
			Tier:                s.Tier,
			OwnerID:             s.CollectionItem.OwnerID,
			XsynLocked:          s.CollectionItem.XsynLocked,
			MarketLocked:        s.CollectionItem.MarketLocked,
			LockedToMarketplace: s.CollectionItem.LockedToMarketplace,
		})
	}

	reply(&PlayerAssetWeaponSubmodelListResp{
		Total:     total,
		Submodels: playerAssetWeaponSubmodel,
	})
	return nil
}

const HubKeyPlayerWeaponBlueprintList = "PLAYER:WEAPON:BLUEPRINT:LIST"

func (pac *PlayerAssetsControllerWS) playerWeaponBlueprintListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "playerWeaponBlueprintListHandler").Str("user_id", user.ID).Logger()

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to see assets.")
	}
	results := []*server.BlueprintWeapon{}

	// need a list of blueprint mechs that the user owns
	rows, err := boiler.NewQuery(
		qm.Select(
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Damage),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.GameClientWeaponID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponType),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DefaultDamageType),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DamageFalloff),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DamageFalloffRate),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Spread),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.RateOfFire),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Radius),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.RadiusDamageFalloff),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ProjectileSpeed),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.MaxAmmo),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.PowerCost),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Collection),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.BrandID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DefaultSkinID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.IsMelee),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ProjectileAmount),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DotTickDamage),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DotMaxTicks),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.IsArced),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ChargeTimeSeconds),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.BurstRateOfFire),
		),
		qm.From(boiler.TableNames.CollectionItems),
		boiler.CollectionItemWhere.OwnerID.EQ(user.ID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Weapons,
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// inner join blueprint weapon
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintWeapons,
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.BlueprintID),
		)),
		qm.GroupBy(qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID)),
	).Query(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("issue getting weapon model list")
		return err
	}

	for rows.Next() {
		wbp := &server.BlueprintWeapon{}
		err = rows.Scan(
			&wbp.ID,
			&wbp.Label,
			&wbp.Damage,
			&wbp.GameClientWeaponID,
			&wbp.WeaponType,
			&wbp.DefaultDamageType,
			&wbp.DamageFalloff,
			&wbp.DamageFalloffRate,
			&wbp.Spread,
			&wbp.RateOfFire,
			&wbp.Radius,
			&wbp.RadiusDamageFalloff,
			&wbp.ProjectileSpeed,
			&wbp.MaxAmmo,
			&wbp.PowerCost,
			&wbp.Collection,
			&wbp.BrandID,
			&wbp.DefaultSkinID,
			&wbp.IsMelee,
			&wbp.ProjectileAmount,
			&wbp.DotTickDamage,
			&wbp.DotMaxTicks,
			&wbp.IsArced,
			&wbp.ChargeTimeSeconds,
			&wbp.BurstRateOfFire,
		)
		if err != nil {
			l.Error().Err(err).Msg("issue parsing weapon blueprints")
			return err
		}
		results = append(results, wbp)
	}

	reply(results)
	return nil
}
