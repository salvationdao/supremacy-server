package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"strings"
	"time"
)

type SyndicateWS struct {
	Log *zerolog.Logger
	API *API
}

func NewSyndicateController(api *API) *SyndicateWS {
	sc := &SyndicateWS{
		API: api,
	}

	api.SecureUserFactionCommand(HubKeySyndicateCreate, sc.SyndicateCreateHandler)
	api.SecureUserFactionCommand(HubKeySyndicateJoin, sc.SyndicateJoinHandler)
	api.SecureUserFactionCommand(HubKeySyndicateLeave, sc.SyndicateLeaveHandler)

	// update syndicate settings

	return sc
}

type SyndicateCreateRequest struct {
	Payload struct {
		Name    string          `json:"name"`
		JoinFee decimal.Decimal `json:"join_fee"`
		ExitFee decimal.Decimal `json:"exit_fee"`
	} `json:"payload"`
}

const HubKeySyndicateCreate = "SYNDICATE:CREATE"

func (sc *SyndicateWS) SyndicateCreateHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &SyndicateCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// check current player has a syndicate
	if user.SyndicateID.Valid {
		return terror.Error(terror.ErrInvalidInput, "Only non-syndicate players can start a new syndicate.")
	}

	// check syndicate name is registered
	syndicateName, err := sc.API.SyndicateNameVerification(req.Payload.Name)
	if err != nil {
		return terror.Error(err, err.Error())
	}

	if req.Payload.ExitFee.GreaterThan(req.Payload.JoinFee) {
		return terror.Error(fmt.Errorf("exit fee is higher than join fee"), "Exit fee should not be higher than join fee.")
	}

	// create new syndicate
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
		return terror.Error(err, "Failed to create syndicate.")
	}

	defer tx.Rollback()

	syndicate := boiler.Syndicate{
		Name:        syndicateName,
		FactionID:   factionID,
		FoundedByID: user.ID,
		JoinFee:     req.Payload.JoinFee,
		ExitFee:     req.Payload.ExitFee,
	}

	err = syndicate.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Msg("Failed to insert syndicate into db.")
		return terror.Error(err, "Failed to create syndicate.")
	}

	user.SyndicateID = null.StringFrom(syndicate.ID)
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Interface("user", user).Msg("Failed to update syndicate id of current user.")
		return terror.Error(err, "Failed to assign syndicate")
	}

	// register syndicate on xsyn server
	err = sc.API.Passport.SyndicateCreateHandler(syndicate.ID, syndicate.FoundedByID, syndicate.Name)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Msg("Failed to register syndicate on xsyn server.")
		return terror.Error(err, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to create syndicate.")
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), HubKeyUserSubscribe, user)

	reply(true)
	return nil
}

func (api *API) SyndicateNameVerification(inputName string) (string, error) {
	syndicateName := strings.TrimSpace(inputName)

	if len(syndicateName) == 0 {
		return "", terror.Error(fmt.Errorf("empty syndicate name"), "The name of syndicate is empty")
	}

	// check profanity
	if api.ProfanityManager.Detector.IsProfane(syndicateName) {
		return "", terror.Error(fmt.Errorf("profanity detected"), "The syndicate name contains profanity")
	}

	// TODO: check max lenght?
	if len(syndicateName) > 50 {
		return "", terror.Error(fmt.Errorf("too many characters"), "The syndicate name should not be longer than 50 characters")
	}

	// check existence
	syndicate, err := boiler.Syndicates(
		qm.Where(
			fmt.Sprintf(
				"LOWER(%s) = ?",
				qm.Rels(boiler.TableNames.Syndicates, boiler.SyndicateColumns.Name),
			),
			strings.ToLower(syndicateName),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate name", syndicateName).Msg("Failed to get syndicate by name from db")
		return "", terror.Error(err, "Failed to verify syndicate name")
	}
	if syndicate != nil {
		return "", terror.Error(fmt.Errorf("invalid input"), fmt.Sprintf("%s has already been taken by other syndicate", inputName))
	}

	return syndicateName, nil
}

type SyndicateJoinRequest struct {
	Payload struct {
		SyndicateID string `json:"syndicate_id"`
	} `json:"payload"`
}

const HubKeySyndicateJoin = "SYNDICATE:JOIN"

func (sc *SyndicateWS) SyndicateJoinHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player already has syndicate"), "You already have a syndicate")
	}

	req := &SyndicateJoinRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// load targeted syndicate
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, req.Payload.SyndicateID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", req.Payload.SyndicateID).Msg("Failed to get syndicate from db")
		return terror.Error(err, "Failed to get syndicate detail")
	}

	// check the faction of the syndicate is same as player's faction
	if syndicate.FactionID != factionID {
		return terror.Error(terror.ErrForbidden, "Cannot join the syndicate in other faction")
	}

	// check available seat count
	currentMemberCount, err := syndicate.Players().Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", syndicate.ID).Msg("Failed to load the number of current member within the syndicate")
		return terror.Error(err, "There is no available seat in the syndicate at the moment")
	}

	if int(currentMemberCount) >= syndicate.SeatCount-1 {
		return terror.Error(fmt.Errorf("no available seat"), "There is no available seat in the syndicate at the moment")
	}

	// check user has enough fund
	userBalance := sc.API.Passport.UserBalanceGet(uuid.FromStringOrNil(user.ID))
	if userBalance.LessThan(syndicate.JoinFee) {
		return terror.Error(fmt.Errorf("insufficent fund"), "Do not have enough sups to pay the join fee")
	}

	dasTax := db.GetDecimalWithDefault(db.KeyDecentralisedAutonomousSyndicateTax, decimal.New(25, -3)) // 0.025

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return terror.Error(err, "Failed to join the syndicate")
	}

	defer tx.Rollback()

	// assign syndicate to the player
	user.SyndicateID = null.StringFrom(syndicate.ID)
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to assign syndicate id to the player")
		return terror.Error(err, "Failed to join the syndicate.")
	}

	// user pay join fee to syndicate, if join fee is greater than zero
	if syndicate.JoinFee.GreaterThan(decimal.Zero) {
		_, err = sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.FromStringOrNil(user.ID),
			ToUserID:             uuid.FromStringOrNil(syndicate.ID),
			Amount:               syndicate.JoinFee.String(),
			TransactionReference: server.TransactionReference(fmt.Sprintf("syndicate_join_fee|%s|%d", syndicate.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupSyndicate),
			Description:          fmt.Sprintf("Syndicate - %s join fee: (%s)", syndicate.Name, syndicate.ID),
			NotSafe:              true,
		})
		if err != nil {
			return terror.Error(err, "Failed to pay syndicate join fee")
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return terror.Error(err, "Failed to join the syndicate")
	}

	// syndicate pay tax to xsyn, if join fee is greater than zero
	if syndicate.JoinFee.GreaterThan(decimal.Zero) {
		_, err = sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.FromStringOrNil(syndicate.ID),
			ToUserID:             uuid.FromStringOrNil(server.XsynTreasuryUserID.String()),
			Amount:               syndicate.JoinFee.Mul(dasTax).String(),
			TransactionReference: server.TransactionReference(fmt.Sprintf("syndicate_das_tax|%s|%d", syndicate.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupSyndicate),
			Description:          fmt.Sprintf("Tax for Syndicate - %s join fee: (%s)", syndicate.Name, syndicate.ID),
			NotSafe:              true,
		})
		if err != nil {
			return terror.Error(err, "Failed to pay syndicate join fee")
		}
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), HubKeyUserSubscribe, user)

	reply(true)

	return nil
}

const HubKeySyndicateLeave = "SYNDICATE:LEAVE"

func (sc *SyndicateWS) SyndicateLeaveHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player has no syndicate"), "You have not join any syndicate yet")
	}

	// load syndicate
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, user.SyndicateID.String)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", user.SyndicateID.String).Msg("Failed to query syndicate from db")
		return terror.Error(err, "Failed to load syndicate detail")
	}

	if user.ID == syndicate.FoundedByID {
		return terror.Error(fmt.Errorf("founder cannot exit the syndicate"), "Syndicate's founder cannot exit the syndicate")
	}

	// check user has enough fund
	userBalance := sc.API.Passport.UserBalanceGet(uuid.FromStringOrNil(user.ID))
	if userBalance.LessThan(syndicate.ExitFee) {
		return terror.Error(fmt.Errorf("insufficent fund"), "Do not have enough sups to pay the exit fee")
	}

	dasTax := db.GetDecimalWithDefault(db.KeyDecentralisedAutonomousSyndicateTax, decimal.New(25, -3)) // 0.025

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return terror.Error(err, "Failed to exit syndicate")
	}

	defer tx.Rollback()

	user.SyndicateID = null.StringFromPtr(nil)
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("user", user).Msg("Failed to update user syndicate column")
		return terror.Error(err, "Failed to exit syndicate")
	}

	// pay syndicate exit fee, if the exit fee is greater than zero
	if syndicate.ExitFee.GreaterThan(decimal.Zero) {
		_, err = sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.FromStringOrNil(user.ID),
			ToUserID:             uuid.FromStringOrNil(syndicate.ID),
			Amount:               syndicate.ExitFee.String(),
			TransactionReference: server.TransactionReference(fmt.Sprintf("syndicate_exit_fee|%s|%d", syndicate.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupSyndicate),
			Description:          fmt.Sprintf("Syndicate - %s exit fee: (%s)", syndicate.Name, syndicate.ID),
			NotSafe:              true,
		})
		if err != nil {
			return terror.Error(err, "Failed to pay syndicate exit fee")
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return terror.Error(err, "Failed to exit syndicate")
	}

	// syndicate pay tax to xsyn, if join fee is greater than zero
	if syndicate.ExitFee.GreaterThan(decimal.Zero) {
		_, err = sc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.FromStringOrNil(syndicate.ID),
			ToUserID:             uuid.FromStringOrNil(server.XsynTreasuryUserID.String()),
			Amount:               syndicate.JoinFee.Mul(dasTax).String(),
			TransactionReference: server.TransactionReference(fmt.Sprintf("syndicate_das_tax|%s|%d", syndicate.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupSyndicate),
			Description:          fmt.Sprintf("Tax for Syndicate - %s exit fee: (%s)", syndicate.Name, syndicate.ID),
			NotSafe:              true,
		})
		if err != nil {
			return terror.Error(err, "Failed to pay syndicate exit fee")
		}
	}

	return nil
}
