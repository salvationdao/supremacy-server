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

	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"

	"github.com/shopspring/decimal"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PlayerAbilitiesControllerWS struct {
	API *API
}

func NewPlayerAbilitiesController(api *API) *PlayerAbilitiesControllerWS {
	pac := &PlayerAbilitiesControllerWS{
		API: api,
	}

	if api.Config.Environment == "development" {
		api.SecureUserCommand(server.HubKeyPlayerAbilitiesList, pac.PlayerAbilitiesListHandler)
		api.SecureUserCommand(server.HubKeySaleAbilityPurchase, pac.SaleAbilityPurchaseHandler)
	}
	api.SecureUserCommand(server.HubKeyPlayerAbilitySubscribe, pac.PlayerAbilitySubscribeHandler)

	return pac
}

//func (pac *PlayerAbilitiesControllerWS) PlayerAbilitiesListUpdatedHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
//	req := &hub.HubCommandRequest{}
//	err := json.Unmarshal(payload, req)
//	if err != nil {
//		return "", "", terror.Error(err, "Invalid request received.")
//	}
//
//	userID, err := uuid.FromString(client.Identifier())
//	if err != nil {
//		gamelog.L.Error().Str("client.Identifier()", client.Identifier()).Err(err).Msg("failed to convert hub id to user id")
//		return "", "", err
//	} else if userID.IsNil() {
//		gamelog.L.Error().Str("client.Identifier()", client.Identifier()).Err(err).Msg("failed to convert hub id to user id, user id is nil")
//		return "", "", terror.Error(fmt.Errorf("user id is nil"), "Issue retriving user, please try again or contact support.")
//	}
//
//	reply(true)
//	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeyPlayerAbilitiesListUpdated, userID)), nil
//}

//func (pac *PlayerAbilitiesControllerWS) SaleAbilitiesListUpdatedHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
//	req := &hub.HubCommandRequest{}
//	err := json.Unmarshal(payload, req)
//	if err != nil {
//		return "", "", terror.Error(err, "Invalid request received.")
//	}
//
//	reply(true)
//	return req.TransactionID, messagebus.BusKey(server.HubKeySaleAbilitiesListUpdated), nil
//}

type PlayerAbilitySubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		BlueprintAbilityID string `json:"blueprint_ability_id"` // blueprint ability id
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) PlayerAbilitySubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAbilitySubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.BlueprintAbilityID == "" {
		gamelog.L.Error().
			Str("handler", "PlayerAbilitySubscribeHandler").Msg("empty ability ID provided")
		return terror.Error(fmt.Errorf("ability ID was not provided in request payload"), "Unable to retrieve player ability, please try again or contact support.")
	}

	userID := user.ID
	bpAbility, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.BlueprintID.EQ(req.Payload.BlueprintAbilityID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(userID),
		qm.OrderBy(fmt.Sprintf("%s asc", boiler.PlayerAbilityColumns.PurchasedAt))).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "boiler.PlayerAbilities").Str("req.Payload.BlueprintAbilityID", req.Payload.BlueprintAbilityID).Str("userID", userID).Err(err).Msg("unable to get blueprint ability details")
		return terror.Error(err, "Unable to retrieve player ability, please try again or contact support.")
	}

	reply(bpAbility)
	return nil
}

type PlayerAbilitiesListResponse struct {
	Total             int64                     `json:"total"`
	TalliedAbilityIDs []db.TalliedPlayerAbility `json:"tallied_ability_ids"`
}

type PlayerAbilitiesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Search   string                `json:"search"`
		Filter   *db.ListFilterRequest `json:"filter"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) PlayerAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerAbilitiesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, talliedPIDs, err := db.PlayerAbilitiesList(req.Payload.Search, req.Payload.Filter, offset, req.Payload.PageSize)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "PlayerAbilitiesList").Err(err).Interface("arguments", req.Payload).Msg("unable to get list of player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(PlayerAbilitiesListResponse{
		total,
		talliedPIDs,
	})
	return nil
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	spas, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now())), qm.Load(boiler.SalePlayerAbilityRels.Blueprint)).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("boiler func", "NewQuery").Err(err).Msg("unable to get current list of sale abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	detailedSaleAbilities := []*db.SaleAbilityDetailed{}
	for _, s := range spas {
		detailedSaleAbilities = append(detailedSaleAbilities, &db.SaleAbilityDetailed{
			SalePlayerAbility: s,
			Ability:           s.R.Blueprint,
		})
	}

	reply(spas)
	return nil
}

type SaleAbilitiesPurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityID string `json:"ability_id"` // sale ability id
		Amount    string `json:"amount"`
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilityPurchaseHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &SaleAbilitiesPurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().Str("user id", user.ID).Err(err).Msg("failed to convert hub id to user id")
		return err
	} else if userID.IsNil() {
		gamelog.L.Error().Str("user id", user.ID).Err(err).Msg("failed to convert hub id to user id, user id is nil")
		return terror.Error(fmt.Errorf("user id is nil"), "Issue retrieving user, please try again or contact support.")
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
	if spa.CurrentPrice.Round(0).GreaterThan(givenAmount) {
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
		LocationSelectType:  bpa.LocationSelectType,
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

	// Tell client to update their player abilities list
	pac.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeyPlayerAbilitiesListUpdated, userID)), true)

	// Update price of sale ability
	pac.API.SalePlayerAbilitiesSystem.Purchase <- &player_abilities.Purchase{
		PlayerID:  userID,
		AbilityID: uuid.FromStringOrNil(spa.ID),
	}
	return nil
}
