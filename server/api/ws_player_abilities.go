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
	"server/player_abilities"
	"server/rpcclient"
	"time"

	"github.com/gofrs/uuid"

	"github.com/shopspring/decimal"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PlayerAbilitiesControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerAbilitiesController(api *API) *PlayerAbilitiesControllerWS {
	pac := &PlayerAbilitiesControllerWS{
		API: api,
	}

	api.SecureUserCommand(server.HubKeySaleAbilityDetailed, pac.SaleAbilityDetailedHandler)
	api.SecureUserCommand(server.HubKeyPlayerAbilitiesList, pac.PlayerAbilitiesListHandler)
	api.SecureUserCommand(server.HubKeySaleAbilitiesList, pac.SaleAbilitiesListHandler)
	api.SecureUserCommand(server.HubKeySaleAbilityPurchase, pac.SaleAbilityPurchaseHandler)

	api.SecureUserSubscribeCommand(server.HubKeyPlayerAbilitySubscribe, pac.PlayerAbilitySubscribeHandler)
	api.SecureUserSubscribeCommand(server.HubKeySaleAbilityPriceSubscribe, pac.SaleAbilitySubscribePriceHandler)
	api.SecureUserSubscribeCommand(server.HubKeyPlayerAbilitiesListUpdated, pac.PlayerAbilitiesListUpdatedHandler)
	api.SecureUserSubscribeCommand(server.HubKeySaleAbilitiesListUpdated, pac.SaleAbilitiesListUpdatedHandler)

	return pac
}

type SaleAbilityDetailsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityID string `json:"ability_id"`
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilityDetailedHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SaleAbilityDetailsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.AbilityID == "" {
		gamelog.L.Error().
			Str("handler", "SaleAbilityDetailsHandler").Msg("empty ability ID provided")
		return terror.Error(fmt.Errorf("ability ID was not provided in request payload"))
	}

	sAbility, err := db.SaleAbilityGet(ctx, gamedb.Conn, req.Payload.AbilityID)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "SaleAbilityGet").Str("req.Payload.AbilityID", req.Payload.AbilityID).Err(err).Msg("unable to get sale ability details")
		return terror.Error(err, "Unable to retrieve sale ability, please try again or contact support.")
	}

	reply(sAbility)
	return nil
}

func (pac *PlayerAbilitiesControllerWS) PlayerAbilitiesListUpdatedHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(client.Identifier())
	if err != nil {
		gamelog.L.Error().Str("client.Identifier()", client.Identifier()).Err(err).Msg("failed to convert hub id to user id")
		return "", "", terror.Error(err)
	} else if userID.IsNil() {
		gamelog.L.Error().Str("client.Identifier()", client.Identifier()).Err(err).Msg("failed to convert hub id to user id, user id is nil")
		return "", "", terror.Error(fmt.Errorf("user id is nil"), "Issue retriving user, please try again or contact support.")
	}

	reply(true)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeyPlayerAbilitiesListUpdated, userID)), nil
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilitiesListUpdatedHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received.")
	}

	reply(true)
	return req.TransactionID, messagebus.BusKey(server.HubKeySaleAbilitiesListUpdated), nil
}

type PlayerAbilitySubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityID string `json:"ability_id"` // player ability id
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) PlayerAbilitySubscribeHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &PlayerAbilitySubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received.")
	}

	if req.Payload.AbilityID == "" {
		gamelog.L.Error().
			Str("handler", "PlayerAbilitySubscribeHandler").Msg("empty ability ID provided")
		return "", "", terror.Error(fmt.Errorf("ability ID was not provided in request payload"))
	}

	pAbility, err := boiler.FindPlayerAbility(gamedb.StdConn, req.Payload.AbilityID)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "boiler.FindPlayerAbility").Str("req.Payload.AbilityID", req.Payload.AbilityID).Err(err).Msg("unable to get player ability details")
		return "", "", terror.Error(err, "Unable to retrieve player ability, please try again or contact support.")
	}

	reply(pAbility)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeyPlayerAbilitySubscribe, pAbility.ID)), nil
}

type SaleAbilitySubscribePriceRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityID string `json:"ability_id"` // sale ability id
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilitySubscribePriceHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &SaleAbilitySubscribePriceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received.")
	}

	if req.Payload.AbilityID == "" {
		gamelog.L.Error().
			Str("handler", "SaleAbilitySubscribeHandler").Msg("empty ability ID provided")
		return "", "", terror.Error(err, "Ability ID must be provided.")
	}

	sAbility, err := boiler.FindSalePlayerAbility(gamedb.StdConn, req.Payload.AbilityID)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "boiler.FindSalePlayerAbility").Str("req.Payload.AbilityID", req.Payload.AbilityID).Err(err).Msg("unable to get sale ability details")
		return "", "", terror.Error(err, "Unable to retrieve sale ability, please try again or contact support.")
	}

	reply(sAbility.CurrentPrice)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeySaleAbilityPriceSubscribe, sAbility.ID)), nil
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

func (pac *PlayerAbilitiesControllerWS) PlayerAbilitiesListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
			Str("db func", "PlayerAbilitiesList").Err(err).Interface("arguments", req.Payload).Msg("unable to get list of player abilities")
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

func (pac *PlayerAbilitiesControllerWS) SaleAbilitiesListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
			Str("db func", "SaleAbilitiesList").Err(err).Interface("arguments", req.Payload).Msg("unable to get list of sale abilities")
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
		AbilityID string `json:"ability_id"` // sale ability id
		Amount    string `json:"amount"`
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilityPurchaseHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	spa, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.ID.EQ(req.Payload.AbilityID), qm.Load(boiler.SalePlayerAbilityRels.Blueprint)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("req.Payload.AbilityID", req.Payload.AbilityID).
			Str("db func", "SalePlayerAbilities").Err(err).Msg("unable to get sale ability")
		return terror.Error(err, "Unable to process sale ability purchase,  check your balance and try again.")
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
		return terror.Error(err, "Unable to process player ability purchase, please try again or contract support.")
	}

	// if price has gone up, tell them
	if spa.CurrentPrice.GreaterThan(givenAmount) {
		gamelog.L.Debug().Str("spa.CurrentPrice", spa.CurrentPrice.String()).Str("givenAmount", givenAmount.String()).Msg("purchase attempt when price increased since user clicked purchase")
		return terror.Warn(fmt.Errorf("price gone up since purchase attempted"), "Purchase failed. This item is no longer available at this price.")
	}

	// Charge player for ability
	supTransactionID, err := pac.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
		Amount:               spa.CurrentPrice.String(),
		FromUserID:           userID,
		ToUserID:             battle.SupremacyUserID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_ability_purchase|%s|%d", req.Payload.AbilityID, time.Now().UnixNano())),
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
		gamelog.L.Error().Str("txID", supTransactionID).Str("playerAbilityID", req.Payload.AbilityID).Err(err).Msg("unable to charge user for player ability purchase")
		return terror.Error(err, "Unable to process player ability purchase,  check your balance and try again.")
	}

	refundFunc := func() {
		// Refund player ability cost
		refundSupTransactionID, err := pac.API.Passport.RefundSupsMessage(supTransactionID)
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

	bpa := spa.R.Blueprint
	pa := boiler.PlayerAbility{
		OwnerID:             userID.String(),
		BlueprintID:         bpa.ID,
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

	// Update price of sale ability
	pac.API.PlayerAbilitiesSystem.Purchase <- &player_abilities.Purchase{
		PlayerID:  userID,
		AbilityID: spa.ID,
	}
	return nil
}
