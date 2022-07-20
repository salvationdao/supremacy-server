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
	"time"

	"github.com/friendsofgo/errors"
	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"

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
		api.SecureUserCommand(server.HubKeySaleAbilityClaim, pac.SaleAbilityClaimHandler)
	}

	api.SecureUserFactionCommand(battle.HubKeyWarMachineAbilityTrigger, api.BattleArena.MechAbilityTriggerHandler)
	api.SecureUserFactionCommand(battle.HubKeyBattleAbilityOptIn, api.BattleArena.BattleAbilityOptIn)

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
	NextRefreshTime              *time.Time                `json:"next_refresh_time"`
	RefreshPeriodDurationSeconds int                       `json:"refresh_period_duration_seconds"`
	SaleAbilities                []*db.SaleAbilityDetailed `json:"sale_abilities"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilitiesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	dpas := pac.API.SalePlayerAbilitiesSystem.CurrentSaleList()

	nextRefresh := pac.API.SalePlayerAbilitiesSystem.NextRefresh()
	reply(&SaleAbilitiesListResponse{
		NextRefreshTime:              &nextRefresh,
		RefreshPeriodDurationSeconds: db.GetIntWithDefault(db.KeySaleAbilityTimeBetweenRefreshSeconds, 600),
		SaleAbilities:                dpas,
	})
	return nil
}

type SaleAbilityClaimRequest struct {
	Payload struct {
		AbilityID string `json:"ability_id"` // sale ability id
	} `json:"payload"`
}

func (pac *PlayerAbilitiesControllerWS) SaleAbilityClaimHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &SaleAbilityClaimRequest{}
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
		return terror.Error(err, "Unable to process sale ability claim,  check your balance and try again.")
	}

	if !pac.API.SalePlayerAbilitiesSystem.IsAbilityAvailable(spa.ID) {
		// If sale of player ability has already expired
		gamelog.L.Debug().
			Str("handler", "SaleAbilityClaimHandler").Interface("salePlayerAbility", spa).Msg("forbid player from claiming expired ability")
		return terror.Error(fmt.Errorf("sale of player ability has already expired"), "Claim failed. This ability is no longer available for claiming.")
	}

	// Check if user has hit their purchase limit
	canPurchase := pac.API.SalePlayerAbilitiesSystem.CanUserClaim(userID.String())
	if !canPurchase {
		nextRefresh := pac.API.SalePlayerAbilitiesSystem.NextRefresh()
		minutes := int(time.Until(nextRefresh).Minutes())
		msg := fmt.Sprintf("Please try again in %d minutes.", minutes)
		if minutes < 1 {
			msg = fmt.Sprintf("Please try again in %d seconds.", int(time.Until(nextRefresh).Seconds()))
		}
		return terror.Error(fmt.Errorf("You have hit your claim limit of %d during this sale period. %s", pac.API.SalePlayerAbilitiesSystem.UserClaimLimit, msg))
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue claiming player ability, please try again or contact support.")
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
			gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to insert PlayerAbility")
			return terror.Error(err, "Issue claiming player ability, please try again or contact support.")
		}
	} else if err != nil {
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to fetch PlayerAbility")
		return terror.Error(err, "Issue claiming player ability, please try again or contact support.")
	}

	pa.Count = pa.Count + 1

	inventoryLimit := db.GetIntWithDefault(db.KeyPlayerAbilityInventoryLimit, 10)
	if pa.Count > inventoryLimit {
		gamelog.L.Debug().Interface("playerAbility", pa).Msg("user has reached their player ability inventory count")
		return terror.Error(fmt.Errorf("You have reached your limit of %d for this ability.", inventoryLimit))
	}

	_, err = pa.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("playerAbility", pa).Msg("failed to update player ability count")
		return err
	}

	// Attempt to add to user's purchase count
	err = pac.API.SalePlayerAbilitiesSystem.AddToUserClaimCount(userID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Str("userID", userID.String()).Str("salePlayerAbilityID", spa.ID).Msg("failed to add to user's purchase count")
		return terror.Error(err, fmt.Sprintf("You have reached your claim limit during this sale period. Please try again in %d minutes.", int(time.Until(pac.API.SalePlayerAbilitiesSystem.NextRefresh()).Minutes())))
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to commit transaction")
		return terror.Error(err, "Issue claiming player ability, please try again or contact support.")
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
	pac.API.SalePlayerAbilitiesSystem.Claim <- &player_abilities.Claim{
		AbilityID: spa.ID,
	}
	return nil
}
