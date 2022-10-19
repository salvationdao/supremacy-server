package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/sale_player_abilities"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AbilitiesControllerWS struct {
	API *API
}

func NewAbilitiesController(api *API) *AbilitiesControllerWS {
	pac := &AbilitiesControllerWS{
		API: api,
	}

	api.SecureUserCommand(server.HubKeySaleAbilitiesList, pac.SaleAbilitiesListHandler)
	api.SecureUserCommand(server.HubKeySaleAbilityPurchase, pac.SaleAbilityPurchaseHandler)

	api.SecureUserFactionCommand(battle.HubKeyWarMachineAbilityTrigger, api.ArenaManager.MechAbilityTriggerHandler)
	return pac
}

func (pac *AbilitiesControllerWS) PlayerSupportAbilitiesHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	L := gamelog.L.With().Str("func", "PlayerSupportAbilitiesHandler").Str("user id", user.ID).Logger()
	battleID := chi.RouteContext(ctx).URLParam("battle_id")
	// convert to uuid to see if its a uuid
	_, err := uuid.FromString(battleID)
	if err != nil {
		return nil // don't need to return an error, frontend can send undefined annoyingly if no battle is active
	}

	L = L.With().Str("battleID", battleID).Logger()

	resp := &battle.PlayerSupportAbilitiesResponse{
		BattleID: battleID,
	}
	supporterAbilities, err := boiler.PlayerBattleAbilities(
		boiler.PlayerBattleAbilityWhere.PlayerID.EQ(user.ID),
		boiler.PlayerBattleAbilityWhere.BattleID.EQ(battleID),
		boiler.PlayerBattleAbilityWhere.UsedAt.IsNull(),
		qm.Load(boiler.PlayerBattleAbilityRels.GameAbility),
	).All(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Msg("failed to load abilities")
		return terror.Error(err, "failed to find supporter abilities, try again or contact support.")
	}

	for _, ability := range supporterAbilities {
		resp.SupporterAbilities = append(resp.SupporterAbilities, &battle.PlayerSupporterAbility{
			ID:                 ability.ID,
			Label:              ability.R.GameAbility.Label,
			Colour:             ability.R.GameAbility.Colour,
			ImageURL:           ability.R.GameAbility.ImageURL,
			Description:        ability.R.GameAbility.Description,
			TextColour:         ability.R.GameAbility.TextColour,
			LocationSelectType: ability.R.GameAbility.LocationSelectType,
			GameClientAbilityID: ability.R.GameAbility.GameClientAbilityID,
		})
	}

	reply(resp)
	return nil
}

type PlayerAbilitySubscribeRequest struct {
	Payload struct {
		BlueprintAbilityID string `json:"blueprint_ability_id"` // blueprint ability id
	} `json:"payload"`
}

func (pac *AbilitiesControllerWS) PlayerAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
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
	TimeLeftSeconds int                       `json:"time_left_seconds"`
	SaleAbilities   []*db.SaleAbilityDetailed `json:"sale_abilities"`
}

func (pac *AbilitiesControllerWS) SaleAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	dpas := pac.API.SalePlayerAbilityManager.CurrentSaleList()

	nextRefresh := pac.API.SalePlayerAbilityManager.NextRefresh().Client
	reply(&SaleAbilitiesListResponse{
		NextRefreshTime: &nextRefresh,
		TimeLeftSeconds: pac.API.SalePlayerAbilityManager.NextRefreshInSeconds(),
		SaleAbilities:   dpas,
	})

	return nil
}

func (pac *AbilitiesControllerWS) SaleAbilitiesListSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	dpas := pac.API.SalePlayerAbilityManager.CurrentSaleList()

	nextRefresh := pac.API.SalePlayerAbilityManager.NextRefresh().Client
	reply(&SaleAbilitiesListResponse{
		NextRefreshTime: &nextRefresh,
		TimeLeftSeconds: pac.API.SalePlayerAbilityManager.NextRefreshInSeconds(),
		SaleAbilities:   dpas,
	})
	return nil
}

type SaleAbilityPurchaseRequest struct {
	Payload struct {
		AbilityID string `json:"ability_id"` // sale ability id
		Price     string `json:"price"`
	} `json:"payload"`
}

func (pac *AbilitiesControllerWS) SaleAbilityPurchaseHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "SaleAbilityPurchaseHandler").Str("userID", user.ID).Logger()
	req := &SaleAbilityPurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	l = l.With().Interface("payload", req.Payload).Logger()
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("failed to convert hub id to user id")
		return err
	} else if userID.IsNil() {
		l.Error().Err(err).Msg("failed to convert hub id to user id, user id is nil")
		return terror.Error(fmt.Errorf("user id is nil"), "Issue retrieving user, please try again or contact support.")
	}

	spa, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.ID.EQ(req.Payload.AbilityID), qm.Load(boiler.SalePlayerAbilityRels.Blueprint)).One(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to get sale ability")
		return terror.Error(err, "Unable to process sale ability purchase,  check your balance and try again.")
	}

	l = l.With().Interface("salePlayerAbility", spa).Logger()
	if !pac.API.SalePlayerAbilityManager.IsAbilityAvailable(spa.ID) {
		// If sale of player ability has already expired
		l.Debug().Msg("forbid player from purchasing expired ability")
		return terror.Error(fmt.Errorf("sale of player ability has already expired"), "Purchase failed. This ability is no longer available for purchasing.")
	}

	// Check if user has hit their purchase limit
	canPurchase := pac.API.SalePlayerAbilityManager.CanUserPurchase(userID.String())
	if !canPurchase {
		nextRefresh := pac.API.SalePlayerAbilityManager.NextRefresh().Client
		minutes := int(time.Until(nextRefresh).Minutes())
		msg := fmt.Sprintf("Please try again in %d minutes.", minutes)
		if minutes < 1 {
			msg = fmt.Sprintf("Please try again in %d seconds.", int(time.Until(nextRefresh).Seconds()))
		}
		return terror.Error(fmt.Errorf("You have hit your purchase limit of %d during this sale period. %s", pac.API.SalePlayerAbilityManager.UserPurchaseLimit, msg))
	}

	givenAmount, err := decimal.NewFromString(req.Payload.Price)
	if err != nil {
		l.Error().Err(err).Msg("failed to convert given price to decimal")
		return terror.Error(err, "Unable to process player ability purchase, please try again or contract support.")
	}

	// If price has gone up, tell them
	if spa.CurrentPrice.Round(0).GreaterThan(givenAmount) {
		l.Debug().Msg("purchase attempt when price increased since user clicked purchase")
		return terror.Warn(fmt.Errorf("price gone up since purchase attempted"), "Purchase failed. This item is no longer available at this price.")
	}

	// Charge player for ability
	supTransactionID, err := pac.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		Amount:               spa.CurrentPrice.String(),
		FromUserID:           userID,
		ToUserID:             uuid.FromStringOrNil(server.SupremacyGameUserID),
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_ability_purchase|%s|%d", req.Payload.AbilityID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             "Player Abilities",
		Description:          fmt.Sprintf("Purchased player ability %s", spa.R.Blueprint.Label),
	})
	l = l.With().Interface("txID", supTransactionID).Logger()
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
		if err == nil {
			err = fmt.Errorf("transaction failed")
		}
		// Abort transaction if charge fails
		l.Error().Err(err).Msg("unable to charge user for player ability purchase")
		return terror.Error(err, "Unable to process player ability purchase, check your balance and try again.")
	}

	refundFunc := func() {
		// Refund player ability cost
		refundSupTransactionID, err := pac.API.Passport.RefundSupsMessage(supTransactionID)
		if err != nil {
			l.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to refund user for player ability purchase cost")
		}
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		refundFunc()
		l.Error().Err(err).Msg("unable to begin tx")
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
			l.Error().Err(err).Interface("playerAbility", pa).Msg("failed to insert PlayerAbility")
			return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
		}
	} else if err != nil {
		refundFunc()
		l.Error().Err(err).Msg("failed to fetch PlayerAbility")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	l = l.With().Interface("playerAbility", pa).Logger()

	pa.Count = pa.Count + 1

	inventoryLimit := spa.R.Blueprint.InventoryLimit
	if pa.Count > inventoryLimit {
		refundFunc()
		l.Debug().Msg("user has reached their player ability inventory count")
		return terror.Warn(fmt.Errorf("inventory limit reached"), fmt.Sprintf("You have reached your limit of %d for this ability.", inventoryLimit))
	}

	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		refundFunc()
		l.Error().Err(err).Msg("failed to update player ability count")
		return err
	}

	// Attempt to add to user's purchase count
	err = pac.API.SalePlayerAbilityManager.AddToUserPurchaseCount(userID.String())
	if err != nil {
		refundFunc()
		l.Warn().Err(err).Msg("failed to add to user's purchase count")
		return terror.Error(err, fmt.Sprintf("You have reached your claim limit during this sale period. Please try again in %d minutes.", int(time.Until(pac.API.SalePlayerAbilityManager.NextRefresh().Client).Minutes())))
	}

	err = tx.Commit()
	if err != nil {
		refundFunc()
		l.Error().Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue claiming player ability, please try again or contact support.")
	}
	reply(true)

	// Tell client to update their player abilities list
	pas, err := db.PlayerAbilitiesList(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("unable to get player abilities")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}
	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/player_abilities", userID), server.HubKeyPlayerAbilitiesList, pas)

	// Update price of sale ability
	pac.API.SalePlayerAbilityManager.Purchase <- &sale_player_abilities.Purchase{
		SaleID: spa.ID,
	}
	return nil
}
