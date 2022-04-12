package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PlayerAbilitiesControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerAbilitiesController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *PlayerAbilitiesControllerWS {
	gac := &PlayerAbilitiesControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "twitch_hub"),
		API:  api,
	}

	api.Command(HubKeyPlayerAbilitiesList, gac.PlayerAbilitiesListHandler)
	api.Command(HubKeyPlayerAbilitiesPurchase, gac.PlayerAbilitiesPurchaseHandler)

	return gac
}

const HubKeyPlayerAbilitiesList = hub.HubCommandKey("PLAYER:ABILITIES:LIST")

func (gac *PlayerAbilitiesControllerWS) PlayerAbilitiesListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	pas, err := boiler.SalePlayerAbilities(qm.Limit(10)).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "SalePlayerAbilities").Err(err).Msg("unable to get list of player abilities")
		return terror.Error(err)
	}

	reply(pas)

	return nil
}

type PlayerAbilitiesPurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		PlayerAbilityID string `json:"player_ability_id"`
	} `json:"payload"`
}

const HubKeyPlayerAbilitiesPurchase = hub.HubCommandKey("PLAYER:ABILITIES:PURCHASE")

func (gac *PlayerAbilitiesControllerWS) PlayerAbilitiesPurchaseHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PlayerAbilitiesPurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := uuid.FromStringOrNil(hub.Identifier())
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	spa, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.BlueprintID.EQ(req.Payload.PlayerAbilityID)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "SalePlayerAbilities").Err(err).Msg("unable to get player ability")
		return terror.Error(err)
	}

	if spa.AvailableUntil.Time.Before(time.Now()) {
		// If sale of player ability has already expired
		gamelog.L.Debug().
			Str("handler", "PlayerAbilitiesPurchaseHandler").Interface("playerAbility", spa).Msg("forbid player from purchasing expired ability")
		return terror.Error(fmt.Errorf("Purchase failed. This ability is no longer available for purchase."))
	}

	// Charge player for ability
	supTransactionID, err := gac.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
		Amount:               spa.CurrentPrice.StringFixed(18),
		FromUserID:           userID,
		ToUserID:             battle.SupremacyBattleUserID,
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_ability_purchase|%s|%d", req.Payload.PlayerAbilityID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		SubGroup:             "Player Abilities",
		Description:          "Purchased a player ability",
		NotSafe:              true,
	})
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Str("playerAbilityID", req.Payload.PlayerAbilityID).Err(err).Msg("unable to charge user for player ability purchase")
		return terror.Error(err, "Unable to process player ability purchase,  check your balance and try again.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to begin tx")
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}
	defer tx.Rollback()

	bpa, err := boiler.FindBlueprintPlayerAbility(tx, req.Payload.PlayerAbilityID)
	if err != nil {
		// Refund player ability cost
		refundSupTransactionID, err := gac.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
			Amount:               spa.CurrentPrice.StringFixed(18),
			FromUserID:           battle.SupremacyBattleUserID,
			ToUserID:             userID,
			TransactionReference: server.TransactionReference(fmt.Sprintf("refund_player_ability_purchase|%s|%d", req.Payload.PlayerAbilityID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupBattle),
			SubGroup:             "Player Abilities",
			Description:          "Refunded player ability purchase",
			NotSafe:              true,
		})
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to refund user for player ability purchase cost")
			return terror.Error(err, "Unable to process refund, try again or contact support.")
		}
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
		// Refund player ability cost
		refundSupTransactionID, err := gac.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
			Amount:               spa.CurrentPrice.StringFixed(18),
			FromUserID:           battle.SupremacyBattleUserID,
			ToUserID:             userID,
			TransactionReference: server.TransactionReference(fmt.Sprintf("refund_player_ability_purchase|%s|%d", req.Payload.PlayerAbilityID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupBattle),
			SubGroup:             "Player Abilities",
			Description:          "Refunded player ability purchase",
			NotSafe:              true,
		})
		if err != nil {
			gamelog.L.Error().Str("txID", refundSupTransactionID).Err(err).Msg("unable to refund user for player ability purchase cost")
			return terror.Error(err, "Unable to process refund, try again or contact support.")
		}
		return terror.Error(err, "Issue purchasing player ability, please try again or contact support.")
	}

	return nil
}
