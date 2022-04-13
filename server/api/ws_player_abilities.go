package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"time"

	"github.com/gofrs/uuid"

	"github.com/shopspring/decimal"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type PlayerAbilitiesControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerAbilitiesController(api *API) *PlayerAbilitiesControllerWS {
	gac := &PlayerAbilitiesControllerWS{
		API: api,
	}

	api.SecureUserCommand(HubKeyPlayerAbilitiesList, gac.PlayerAbilitiesListHandler)
	api.SecureUserCommand(HubKeySaleAbilitiesList, gac.SaleAbilitiesListHandler)
	api.SecureUserCommand(HubKeySaleAbilitiesPurchase, gac.SaleAbilitiesPurchaseHandler)

	return gac
}

type AbilitiesListResponse struct {
	Total      int      `json:"total"`
	AbilityIDs []string `json:"ability_ids"`
}

type PlayerAbilitiesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir           `json:"sort_dir"`
		SortBy   db.PlayerAbilityColumn `json:"sort_by"`
		Filter   *db.ListFilterRequest  `json:"filter,omitempty"`
		Search   string                 `json:"search"`
		PageSize int                    `json:"page_size"`
		Page     int                    `json:"page"`
	} `json:"payload"`
}

const HubKeyPlayerAbilitiesList = hub.HubCommandKey("PLAYER:ABILITIES:LIST")

func (gac *PlayerAbilitiesControllerWS) PlayerAbilitiesListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PlayerAbilitiesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, pIDs, err := db.PlayerAbilitiesList(ctx, gamedb.Conn, req.Payload.Search, req.Payload.Filter, offset, req.Payload.PageSize, req.Payload.SortBy, req.Payload.SortDir)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "PlayerAbilitiesList").Err(err).Msg("unable to get list of player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(AbilitiesListResponse{
		total,
		pIDs,
	})
	return nil
}

type SaleAbilitiesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir               `json:"sort_dir"`
		SortBy   db.SalePlayerAbilityColumn `json:"sort_by"`
		Filter   *db.ListFilterRequest      `json:"filter,omitempty"`
		Search   string                     `json:"search"`
		PageSize int                        `json:"page_size"`
		Page     int                        `json:"page"`
	} `json:"payload"`
}

const HubKeySaleAbilitiesList = hub.HubCommandKey("SALE:ABILITIES:LIST")

func (gac *PlayerAbilitiesControllerWS) SaleAbilitiesListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SaleAbilitiesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, sIDs, err := db.SaleAbilitiesList(ctx, gamedb.Conn, req.Payload.Search, req.Payload.Filter, offset, req.Payload.PageSize, req.Payload.SortBy, req.Payload.SortDir)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "SaleAbilitiesList").Err(err).Msg("unable to get list of sale abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(AbilitiesListResponse{
		total,
		sIDs,
	})
	return nil
}

type SaleAbilitiesPurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		PlayerAbilityID string `json:"player_ability_id"`
		Amount          string `json:"amount"`
	} `json:"payload"`
}

const HubKeySaleAbilitiesPurchase = hub.HubCommandKey("SALE:ABILITIES:PURCHASE")

func (gac *PlayerAbilitiesControllerWS) SaleAbilitiesPurchaseHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SaleAbilitiesPurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(hub.Identifier())
	if err != nil {
		gamelog.L.Error().Str("hub.Identifier()", hub.Identifier()).Err(err).Msg("failed to convert hub id to user id")
		return terror.Error(err)
	} else if userID.IsNil() {
		gamelog.L.Error().Str("hub.Identifier()", hub.Identifier()).Err(err).Msg("failed to convert hub id to user id, user id is nil")
		return terror.Error(fmt.Errorf("user id is nil"), "Issue retriving user, please try again or contact support.")
	}

	spa, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.BlueprintID.EQ(req.Payload.PlayerAbilityID)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("req.Payload.PlayerAbilityID", req.Payload.PlayerAbilityID).
			Str("db func", "SalePlayerAbilities").Err(err).Msg("unable to get player ability")
		return terror.Error(err, "Unable to process player ability purchase,  check your balance and try again.")
	}

	if spa.AvailableUntil.Time.Before(time.Now()) {
		// If sale of player ability has already expired
		gamelog.L.Debug().
			Str("handler", "PlayerAbilitiesPurchaseHandler").Interface("playerAbility", spa).Msg("forbid player from purchasing expired ability")
		return terror.Error(fmt.Errorf("sale of player ability has already expired"), "Purchase failed. This ability is no longer available for purchase.")
	}

	givenAmount, err := decimal.NewFromString(req.Payload.Amount)
	if err != nil {
		gamelog.L.Error().
			Str("req.Payload.Amount", req.Payload.Amount).Err(err).Msg("failed to convert amount to decimal")
		return terror.Error(err, "Unable to process player ability purchase,  please try again or contract support.")
	}

	// if price has gone up, tell them
	if spa.CurrentPrice.GreaterThan(givenAmount) {
		gamelog.L.Debug().Str("spa.CurrentPrice", spa.CurrentPrice.String()).Str("givenAmount", givenAmount.String()).Msg("purchase attempt when price increased since user clicked purchase")
		return terror.Warn(fmt.Errorf("price gone up since purchase attempted"), "Item no longer available.")
	}

	// Charge player for ability
	supTransactionID, err := gac.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
		Amount:               spa.CurrentPrice.String(),
		FromUserID:           userID,
		ToUserID:             battle.SupremacyBattleUserID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_ability_purchase|%s|%d", req.Payload.PlayerAbilityID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             "Player Abilities",
		Description:          fmt.Sprintf("Purchased player ability %s", spa.R.Blueprint.Label),
		NotSafe:              true,
	})
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
		if err == nil {
			err = fmt.Errorf("transaction failed")
		}
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Str("playerAbilityID", req.Payload.PlayerAbilityID).Err(err).Msg("unable to charge user for player ability purchase")
		return terror.Error(err, "Unable to process player ability purchase,  check your balance and try again.")
	}

	refundFunc := func() {
		// Refund player ability cost
		refundSupTransactionID, err := gac.API.Passport.RefundSupsMessage(supTransactionID)
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to refund user for player ability purchase cost")
		}
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	defer tx.Rollback()

	bpa, err := boiler.FindBlueprintPlayerAbility(tx, req.Payload.PlayerAbilityID)
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Str("blueprintID", req.Payload.PlayerAbilityID).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}

	pa := boiler.PlayerAbility{
		OwnerID:             userID.String(),
		BlueprintID:         req.Payload.PlayerAbilityID,
		GameClientAbilityID: bpa.GameClientAbilityID,
		Label:               bpa.Label,
		Colour:              bpa.Colour,
		ImageURL:            bpa.ImageURL,
		Description:         bpa.Description,
		TextColour:          bpa.TextColour,
		Type:                bpa.Type,
	}
	err = pa.Insert(tx, boil.Infer())
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to insert PlayerAbility")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	reply(true)
	return nil
}
