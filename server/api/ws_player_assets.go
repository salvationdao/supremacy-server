package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"strings"
	"time"
	"unicode"

	"github.com/kevinms/leakybucket-go"
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
	api.SecureUserFactionCommand(HubKeyPlayerAssetMechDetail, pac.PlayerAssetMechDetail)
	api.SecureUserFactionCommand(HubKeyPlayerAssetWeaponDetail, pac.PlayerAssetWeaponDetail)
	api.SecureUserCommand(HubKeyPlayerAssetKeycardList, pac.PlayerAssetKeycardListHandler)
	api.SecureUserCommand(HubKeyPlayerAssetKeycardGet, pac.PlayerAssetKeycardGetHandler)
	api.SecureUserCommand(HubKeyPlayerAssetRename, pac.PlayerMechRenameHandler)
	api.SecureUserFactionCommand(HubKeyOpenCrate, pac.OpenCrateHandler)

	return pac
}

const HubKeyPlayerAssetMechList = "PLAYER:ASSET:MECH:LIST"
const HubKeyPlayerAssetWeaponList = "PLAYER:ASSET:WEAPON:LIST"

type PlayerAssetMechListRequest struct {
	Payload struct {
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

type PlayerAssetMech struct {
	CollectionSlug      string      `json:"collection_slug"`
	Hash                string      `json:"hash"`
	TokenID             int64       `json:"token_id"`
	ItemType            string      `json:"item_type"`
	Tier                string      `json:"tier"`
	OwnerID             string      `json:"owner_id"`
	ImageURL            null.String `json:"image_url,omitempty"`
	CardAnimationURL    null.String `json:"card_animation_url,omitempty"`
	AvatarURL           null.String `json:"avatar_url,omitempty"`
	LargeImageURL       null.String `json:"large_image_url,omitempty"`
	BackgroundColor     null.String `json:"background_color,omitempty"`
	AnimationURL        null.String `json:"animation_url,omitempty"`
	YoutubeURL          null.String `json:"youtube_url,omitempty"`
	MarketLocked        bool        `json:"market_locked"`
	XsynLocked          bool        `json:"xsyn_locked"`
	LockedToMarketplace bool        `json:"locked_to_marketplace"`
	QueuePosition       null.Int    `json:"queue_position"`

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

	listOpts := &db.MechListOpts{
		Search:              req.Payload.Search,
		Filter:              req.Payload.Filter,
		Sort:                req.Payload.Sort,
		PageSize:            req.Payload.PageSize,
		Page:                req.Payload.Page,
		OwnerID:             user.ID,
		DisplayXsynMechs:    req.Payload.DisplayXsynMechs,
		ExcludeMarketLocked: req.Payload.ExcludeMarketLocked,
		IncludeMarketListed: req.Payload.IncludeMarketListed,
	}
	if req.Payload.QueueSort.IsValid() && user.FactionID.Valid {
		listOpts.QueueSort = &db.MechListQueueSortOpts{
			FactionID: user.FactionID.String,
			SortDir:   req.Payload.QueueSort,
		}
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
			LockedToMarketplace:   m.CollectionItem.LockedToMarketplace,
			QueuePosition:         m.QueuePosition,
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
	mech, err := db.Mech(gamedb.StdConn, collectionItem.ItemID)
	if err != nil {
		return terror.Error(err, "Failed to find mech from db")
	}

	reply(mech)
	return nil
}

type PlayerAssetWeaponDetailRequest struct {
	Payload struct {
		WeaponID string `json:"weapon_id"`
	} `json:"payload"`
}

const HubKeyPlayerAssetWeaponDetail = "PLAYER:ASSET:WEAPON:DETAIL"

func (pac *PlayerAssetsControllerWS) PlayerAssetWeaponDetail(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAssetWeaponDetailRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	// get collection and check ownership
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.WeaponID),
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
	weapon, err := db.Weapon(nil, collectionItem.ItemID)
	if err != nil {
		return terror.Error(err, "Failed to find weapon from db")
	}

	reply(weapon)
	return nil
}

func (pac *PlayerAssetsControllerWS) PlayerWeaponsListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	pas, err := db.PlayerWeaponsList(user.ID)
	if err != nil {
		gamelog.L.Error().Str("db func", "TalliedPlayerWeaponsList").Str("userID", user.ID).Err(err).Msg("unable to get player weapons")
		return terror.Error(err, "Unable to retrieve weapons, try again or contact support.")
	}

	reply(pas)
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
		IsHangar *bool  `json:"is_hangar,omitempty"`
	} `json:"payload"`
}

type OpenCrateResponse struct {
	Mech       *server.Mech       `json:"mech,omitempty"`
	MechSkin   *server.MechSkin   `json:"mech_skin,omitempty"`
	Weapons    []*server.Weapon   `json:"weapon,omitempty"`
	WeaponSkin *server.WeaponSkin `json:"weapon_skin,omitempty"`
	PowerCore  *server.PowerCore  `json:"power_core,omitempty"`
}

const HubKeyOpenCrate = "CRATE:OPEN"

var openCrateBucket = leakybucket.NewLeakyBucket(0.5, 1)

func (pac *PlayerAssetsControllerWS) OpenCrateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	v := openCrateBucket.Add(1)
	if v == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Currently handling request, please try again.")
	}

	req := &OpenCrateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	isHangarOpening := false
	if req.Payload.IsHangar != nil {
		isHangarOpening = *req.Payload.IsHangar
	}
	var collectionItem *boiler.CollectionItem
	if isHangarOpening {
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

	crate := boiler.MysteryCrate{}

	q := `
		UPDATE mystery_crate
		SET opened = TRUE
		WHERE id = $1 AND opened = FALSE AND locked_until <= NOW()
		RETURNING id, type, faction_id, label, opened, locked_until, purchased, deleted_at, updated_at, created_at, description`
	err = gamedb.StdConn.
		QueryRow(q, collectionItem.ItemID).
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
		_, err = crate.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed rollback crate opened: %s", crate.ID))
		}
	}

	items := OpenCrateResponse{}

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

	for _, blueprintItem := range blueprintItems {
		switch blueprintItem.BlueprintType {
		case boiler.TemplateItemTypeMECH:
			bp, err := db.BlueprintMech(blueprintItem.BlueprintID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get mech blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
			}

			mech, err := db.InsertNewMech(tx, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new mech from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
			}
			items.Mech = mech
		case boiler.TemplateItemTypeWEAPON:
			bp, err := db.BlueprintWeapon(blueprintItem.BlueprintID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get weapon blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get weapon blueprint during crate opening, try again or contact support.")
			}

			weapon, err := db.InsertNewWeapon(tx, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new weapon from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get weapon during crate opening, try again or contact support.")
			}
			items.Weapons = append(items.Weapons, weapon)
		case boiler.TemplateItemTypeMECH_SKIN:
			bp, err := db.BlueprintMechSkinSkin(blueprintItem.BlueprintID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get mech skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get mech skin blueprint during crate opening, try again or contact support.")
			}

			mechSkin, err := db.InsertNewMechSkin(tx, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new mech skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get mech skin during crate opening, try again or contact support.")
			}
			items.MechSkin = mechSkin
		case boiler.TemplateItemTypeWEAPON_SKIN:
			bp, err := db.BlueprintWeaponSkin(blueprintItem.BlueprintID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get weapon skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get weapon skin blueprint during crate opening, try again or contact support.")
			}
			weaponSkin, err := db.InsertNewWeaponSkin(tx, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new weapon skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get weapon skin during crate opening, try again or contact support.")
			}
			items.WeaponSkin = weaponSkin
		case boiler.TemplateItemTypePOWER_CORE:
			bp, err := db.BlueprintPowerCore(blueprintItem.BlueprintID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get powercore blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get powercore blueprint during crate opening, try again or contact support.")
			}

			powerCore, err := db.InsertNewPowerCore(tx, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new powercore from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return terror.Error(err, "Could not get powercore during crate opening, try again or contact support.")
			}
			items.PowerCore = powerCore
		}
	}
	var hangarResp *db.SiloType
	if crate.Type == boiler.CrateTypeMECH {
		eod, err := db.MechEquippedOnDetails(tx, items.Mech.ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get MechEquippedOnDetails during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}

		//attach mech_skin to mech - mech
		err = db.AttachMechSkinToMech(tx, user.ID, items.Mech.ID, items.MechSkin.ID, false)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach mech skin to mech during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		items.MechSkin.EquippedOn = null.StringFrom(items.Mech.ID)
		items.MechSkin.EquippedOnDetails = eod
		xsynAsserts = append(xsynAsserts, rpctypes.ServerMechSkinsToXsynAsset([]*server.MechSkin{items.MechSkin})...)

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
		for _, weapon := range items.Weapons {
			err = db.AttachWeaponToMech(tx, user.ID, items.Mech.ID, weapon.ID)
			if err != nil {
				crateRollback()
				gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapons to mech during CRATE:OPEN crate: %s", crate.ID))
				return terror.Error(err, "Could not open crate, try again or contact support.")
			}
			weapon.EquippedOn = null.StringFrom(items.Mech.ID)
			weapon.EquippedOnDetails = eod
		}
		xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponsToXsynAsset(items.Weapons)...)

		mech, err := db.Mech(tx, items.Mech.ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get final mech during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		mech.ChassisSkin = items.MechSkin
		xsynAsserts = append(xsynAsserts, rpctypes.ServerMechsToXsynAsset([]*server.Mech{mech})...)

		if isHangarOpening {
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
		err = db.AttachWeaponSkinToWeapon(tx, user.ID, items.Weapons[0].ID, items.WeaponSkin.ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapon skin to weapon during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		items.WeaponSkin.EquippedOn = null.StringFrom(items.Weapons[0].ID)
		items.WeaponSkin.EquippedOnDetails = wod
		xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponSkinsToXsynAsset([]*server.WeaponSkin{items.WeaponSkin})...)

		weapon, err := db.Weapon(tx, items.Weapons[0].ID)
		if err != nil {
			crateRollback()
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get final mech during CRATE:OPEN crate: %s", crate.ID))
			return terror.Error(err, "Could not open crate, try again or contact support.")
		}
		xsynAsserts = append(xsynAsserts, rpctypes.ServerWeaponsToXsynAsset([]*server.Weapon{weapon})...)

		if isHangarOpening {
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

	if isHangarOpening {
		reply(hangarResp)
		return nil
	}

	reply(items)

	return nil
}

type PlayerAssetWeaponListRequest struct {
	Payload struct {
		Search              string                `json:"search"`
		Filter              *db.ListFilterRequest `json:"filter"`
		Sort                *db.ListSortRequest   `json:"sort"`
		PageSize            int                   `json:"page_size"`
		Page                int                   `json:"page"`
		DisplayXsynMechs    bool                  `json:"display_xsyn_mechs"`
		ExcludeMarketLocked bool                  `json:"exclude_market_locked"`
		IncludeMarketListed bool                  `json:"include_market_listed"`
		ExcludeEquipped     bool                  `json:"exclude_equipped"`
		FilterRarities      []string              `json:"rarities"`
		FilterWeaponTypes   []string              `json:"weapon_types"`
	} `json:"payload"`
}

type PlayerAssetWeaponListResp struct {
	Total   int64                `json:"total"`
	Weapons []*PlayerAssetWeapon `json:"weapons"`
}

type PlayerAssetWeapon struct {
	CollectionSlug      string      `json:"collection_slug"`
	Hash                string      `json:"hash"`
	TokenID             int64       `json:"token_id"`
	Tier                string      `json:"tier"`
	OwnerID             string      `json:"owner_id"`
	ImageURL            null.String `json:"image_url,omitempty"`
	CardAnimationURL    null.String `json:"card_animation_url,omitempty"`
	AvatarURL           null.String `json:"avatar_url,omitempty"`
	LargeImageURL       null.String `json:"large_image_url,omitempty"`
	BackgroundColor     null.String `json:"background_color,omitempty"`
	AnimationURL        null.String `json:"animation_url,omitempty"`
	YoutubeURL          null.String `json:"youtube_url,omitempty"`
	MarketLocked        bool        `json:"market_locked"`
	XsynLocked          bool        `json:"xsyn_locked"`
	LockedToMarketplace bool        `json:"locked_to_marketplace"`

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
		Search:              req.Payload.Search,
		Filter:              req.Payload.Filter,
		Sort:                req.Payload.Sort,
		PageSize:            req.Payload.PageSize,
		Page:                req.Payload.Page,
		OwnerID:             user.ID,
		DisplayXsynMechs:    req.Payload.DisplayXsynMechs,
		ExcludeMarketLocked: req.Payload.ExcludeMarketLocked,
		IncludeMarketListed: req.Payload.IncludeMarketListed,
		ExcludeEquipped:     req.Payload.ExcludeEquipped,
		FilterRarities:      req.Payload.FilterRarities,
		FilterWeaponTypes:   req.Payload.FilterWeaponTypes,
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
			ImageURL:            m.CollectionItem.ImageURL,
			AvatarURL:           m.CollectionItem.AvatarURL,
			BackgroundColor:     m.CollectionItem.BackgroundColor,
		})
	}

	reply(&PlayerAssetWeaponListResp{
		Total:   total,
		Weapons: playerAssWeapons,
	})
	return nil
}
