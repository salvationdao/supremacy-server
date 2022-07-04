package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/player_abilities"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"

	"github.com/shopspring/decimal"

	"github.com/ninja-software/terror/v2"
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

	if api.Config.Environment == "development" || api.Config.Environment == "staging" {
		api.SecureUserCommand(server.HubKeySaleAbilityPurchase, pac.SaleAbilityPurchaseHandler)
	}

	return pac
}

type PlayerAbilitySubscribeRequest struct {
	Payload struct {
		BlueprintAbilityID string `json:"blueprint_ability_id"` // blueprint ability id
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) PlayerAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	pas, err := db.PlayerAbilitiesList(user.ID)
	if err != nil {
		gamelog.L.Error().Str("db func", "TalliedPlayerAbilitiesList").Str("userID", user.ID).Err(err).Msg("unable to get player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(pas)
	return nil
}

type SaleAbilitiesListResponse struct {
	NextRefreshTime *time.Time                `json:"next_refresh_time"`
	SaleAbilities   []*db.SaleAbilityDetailed `json:"sale_abilities"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	dspas, err := db.CurrentSaleAbilitiesList()
	if err != nil {
		gamelog.L.Error().Str("db func", "CurrentSaleAbilitiesList").Err(err).Msg("unable to get current list of sale abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(&SaleAbilitiesListResponse{
		SaleAbilities: dspas,
	})
	return nil
}

type SaleAbilitiesPurchaseRequest struct {
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
			Str("handler", "PlayerAbilitiesPurchaseHandler").Interface("salePlayerAbility", spa).Msg("forbid player from purchasing expired ability")
		return terror.Error(fmt.Errorf("sale of player ability has already expired"), "Purchase failed. This ability is no longer available for purchase.")
	}

	if spa.AmountSold == spa.SaleLimit {
		// If sale of player ability limit has been reached
		gamelog.L.Debug().
			Str("handler", "PlayerAbilitiesPurchaseHandler").Interface("salePlayerAbility", spa).Msg("forbid player from purchasing ability that has had its sale limit reached")
		return terror.Error(fmt.Errorf("sale of player ability limit has already been reached"), "Purchase failed. This ability has been sold out and is no longer available for purchase.")
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
	supTransactionID, err := pac.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
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

	// Update player ability count
	pa, err := boiler.PlayerAbilities(
		boiler.PlayerAbilityWhere.BlueprintID.EQ(spa.BlueprintID),
		boiler.PlayerAbilityWhere.OwnerID.EQ(userID.String()),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		pa = &boiler.PlayerAbility{
			OwnerID:         userID.String(),
			BlueprintID:     spa.BlueprintID,
			LastPurchasedAt: time.Now(),
		}

		err = pa.Insert(tx, boil.Infer())
		if err != nil {
			refundFunc()
			gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to insert PlayerAbility")
			return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
		}
	} else if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to fetch PlayerAbility")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}

	pa.Count = pa.Count + 1
	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to update player ability count")
		return err
	}

	err = pac.API.SalePlayerAbilitiesSystem.AddToUserPurchaseCount(userID, spa.ID)
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to fetch PlayerAbility")

		_, minutes, _ := pac.API.SalePlayerAbilitiesSystem.NextSalePeriod().Clock()
		return terror.Error(err, fmt.Sprintf("You have reached your purchasing limits for this player ability during this sale period. Please try again in %d minutes.", minutes))
	}

	err = tx.Commit()
	if err != nil {
		refundFunc()
		gamelog.L.Error().Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	reply(true)

	// Tell client to update their player abilities list
	pas, err := db.PlayerAbilitiesList(user.ID)
	if err != nil {
		gamelog.L.Error().Str("boiler func", "PlayerAbilities").Str("ownerID", user.ID).Err(err).Msg("unable to get player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}
	ws.PublishMessage(fmt.Sprintf("/user/%s/player_abilities", userID), server.HubKeyPlayerAbilitiesList, pas)

	// Update price of sale ability
	pac.API.SalePlayerAbilitiesSystem.Purchase <- &player_abilities.Purchase{
		PlayerID:  userID,
		AbilityID: uuid.FromStringOrNil(spa.ID),
	}
	return nil
}
