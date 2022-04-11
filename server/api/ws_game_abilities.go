package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
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

	pa, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.BlueprintPlayerAbilityID.EQ(req.Payload.PlayerAbilityID)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "SalePlayerAbilities").Err(err).Msg("unable to get player ability")
		return terror.Error(err)
	}

	// Charge user for ability
	supTransactionID, err := gac.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
		Amount:               pa.CurrentPrice.StringFixed(18),
		FromUserID:           userID,
		ToUserID:             uuid.Nil, // TODO change this
		TransactionReference: server.TransactionReference(fmt.Sprintf("player_ability_purchase|%s|%d", req.Payload.PlayerAbilityID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		SubGroup:             string(server.TransactionGroupStore),
		Description:          "Purchased a player ability",
		NotSafe:              true,
	})
	if err != nil || supTransactionID == "TRANSACTION_FAILED" {
		// Abort transaction if charge fails
		gamelog.L.Error().Str("txID", supTransactionID).Str("playerAbilityID", req.Payload.PlayerAbilityID).Err(err).Msg("unable to charge user for player ability purchase")
		return terror.Error(err, "Unable to process player ability purchase,  check your balance and try again.")
	}

	return nil
}
