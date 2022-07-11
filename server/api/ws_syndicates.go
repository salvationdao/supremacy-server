package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
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
	api.SecureUserFactionCommand(HubKeySyndicateIssueMotion, sc.SyndicateIssueMotionHandler)

	// motion pass instantly if less than 3

	return sc
}

type SyndicateCreateRequest struct {
	Payload struct {
		Name    string          `json:"name"`
		Type    string          `json:"type"`
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

	// check price
	if req.Payload.JoinFee.LessThan(decimal.Zero) {
		return terror.Error(terror.ErrInvalidInput, "Join fee should not be less than zero")
	}

	if req.Payload.ExitFee.LessThan(decimal.Zero) {
		return terror.Error(terror.ErrInvalidInput, "Exit fee should not be less than zero")
	}

	if req.Payload.ExitFee.GreaterThan(req.Payload.JoinFee) {
		return terror.Error(fmt.Errorf("exit fee is higher than join fee"), "Exit fee should not be higher than join fee.")
	}

	// check syndicate name is registered
	syndicateName, err := sc.API.SyndicateNameVerification(req.Payload.Name)
	if err != nil {
		return terror.Error(err, err.Error())
	}

	// create new syndicate
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
		return terror.Error(err, "Failed to create syndicate.")
	}

	defer tx.Rollback()

	syndicate := &boiler.Syndicate{
		Type:        req.Payload.Type,
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
	user.DirectorOfSyndicateID = null.StringFromPtr(nil)
	if syndicate.Type == boiler.SyndicateTypeCORPORATION {
		user.DirectorOfSyndicateID = null.StringFrom(syndicate.ID)
	}
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID, boiler.PlayerColumns.DirectorOfSyndicateID))
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

	err = sc.API.SyndicateSystem.AddSyndicate(syndicate)
	if err != nil {
		return terror.Error(err, "Failed to add syndicate to the system")
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

	// broadcast latest syndicate detail
	serverSyndicate, err := GetSyndicateLatestDetail(syndicate.ID)
	if err != nil {
		return terror.Error(err, "Failed to get syndicate detail")
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s", syndicate.FactionID, syndicate.ID), HubKeySyndicateGeneralDetailSubscribe, serverSyndicate)

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
	user.DirectorOfSyndicateID = null.StringFromPtr(nil)
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID, boiler.PlayerColumns.DirectorOfSyndicateID))
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

	// broadcast updated user
	ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), HubKeyUserSubscribe, user)

	// broadcast latest syndicate detail
	serverSyndicate, err := GetSyndicateLatestDetail(syndicate.ID)
	if err != nil {
		return terror.Error(err, "Failed to get syndicate detail")
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s", syndicate.FactionID, syndicate.ID), HubKeySyndicateGeneralDetailSubscribe, serverSyndicate)

	reply(true)

	return nil
}

type SyndicateIssueMotionRequest struct {
	Payload struct {
		SyndicateID string `json:"syndicate_id"`
		Type        string `json:"type"`
		Reason      string `json:"reason"`

		// content
		NewSymbolID         null.String `json:"new_symbol_id"`
		NewSyndicateName    null.String `json:"new_syndicate_name"`
		NewNamingConvention null.String `json:"new_naming_convention"`

		NewJoinFee decimal.NullDecimal `json:"new_join_fee"`
		NewExitFee decimal.NullDecimal `json:"new_exit_fee"`

		NewDeployingUserPercentage decimal.NullDecimal `json:"new_deploying_user_percentage"`
		NewAbilityKillPercentage   decimal.NullDecimal `json:"new_ability_kill_percentage"`
		NewMechOwnerPercentage     decimal.NullDecimal `json:"new_mech_owner_percentage"`
		NewSyndicateCutPercentage  decimal.NullDecimal `json:"new_syndicate_cut_percentage"`

		SyndicateRuleID null.String `json:"syndicate_rule_id"`
		NewRuleNumber   null.Int    `json:"new_rule_number"`
		NewRuleContent  null.String `json:"new_rule_content"`

		DirectorID null.String `json:"director_id"`

		DurationDays int `json:"duration_days"`
	} `json:"payload"`
}

const HubKeySyndicateIssueMotion = "SYNDICATE:ISSUE:MOTION"

func (sc *SyndicateWS) SyndicateIssueMotionHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player has no syndicate"), "You have not join any syndicate yet.")
	}

	req := &SyndicateIssueMotionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.SyndicateID != user.SyndicateID.String {
		return terror.Error(terror.ErrInvalidInput, "Can only issue motion in your own syndicate")
	}

	if req.Payload.Reason == "" {
		return terror.Error(fmt.Errorf("missing reason"), "Missing reason for the motion")
	}

	// get syndicate
	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, req.Payload.SyndicateID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get syndicate from db")
		return terror.Error(err, "Failed to load syndicate data")
	}

	if syndicate.Type == boiler.SyndicateTypeCORPORATION && user.DirectorOfSyndicateID.String == syndicate.ID {
		return terror.Error(terror.ErrForbidden, "Only directors can issue motion in the syndicate.")
	}

	motion := &boiler.SyndicateMotion{
		Type:       req.Payload.Type,
		IssuedByID: user.ID,
		Reason:     req.Payload.Reason,
	}
	// start issue motion
	switch req.Payload.Type {
	case boiler.SyndicateMotionTypeCHANGE_GENERAL_DETAIL:
		if !req.Payload.NewSyndicateName.Valid && !req.Payload.NewSymbolID.Valid && !req.Payload.NewNamingConvention.Valid {
			return terror.Error(fmt.Errorf("change info is not provided"), "Change info is not provided.")
		}
		// change symbol, name
		motion.NewName = req.Payload.NewSyndicateName
		motion.NewSymbolID = req.Payload.NewSymbolID
		motion.NewNamingConvention = req.Payload.NewNamingConvention

	case boiler.SyndicateMotionTypeCHANGE_PAYMENT_SETTING:
		// change
		motion.NewJoinFee = req.Payload.NewJoinFee
		motion.NewExitFee = req.Payload.NewExitFee
		motion.NewDeployingUserPercentage = req.Payload.NewDeployingUserPercentage
		motion.NewAbilityKillPercentage = req.Payload.NewAbilityKillPercentage
		motion.NewMechOwnerPercentage = req.Payload.NewMechOwnerPercentage
		motion.NewSyndicateCutPercentage = req.Payload.NewSyndicateCutPercentage

	case boiler.SyndicateMotionTypeADD_RULE:
		if !req.Payload.NewRuleContent.Valid {
			return terror.Error(fmt.Errorf("rule content is not provided"), "Rule content is not provided.")
		}
		motion.NewRuleContent = req.Payload.NewRuleContent
		motion.NewRuleNumber = req.Payload.NewRuleNumber

	case boiler.SyndicateMotionTypeREMOVE_RULE:
		if !req.Payload.SyndicateRuleID.Valid {
			return terror.Error(fmt.Errorf("missing rule id"), "Missing rule id")
		}
		_, err := boiler.FindSyndicateRule(gamedb.StdConn, req.Payload.SyndicateRuleID.String)
		if err != nil {
			return terror.Error(err, "Syndicate rule does not exist")
		}
		motion.RuleID = req.Payload.SyndicateRuleID

	case boiler.SyndicateMotionTypeCHANGE_RULE:
		if !req.Payload.SyndicateRuleID.Valid {
			return terror.Error(fmt.Errorf("missing rule id"), "Missing rule id")
		}
		_, err := boiler.FindSyndicateRule(gamedb.StdConn, req.Payload.SyndicateRuleID.String)
		if err != nil {
			return terror.Error(err, "Syndicate rule does not exist")
		}

		if !req.Payload.NewRuleNumber.Valid && !req.Payload.NewRuleContent.Valid {
			return terror.Error(fmt.Errorf("missing rule change"), "Changing content is not provided.")
		}

		motion.NewRuleNumber = req.Payload.NewRuleNumber
		motion.NewRuleContent = req.Payload.NewRuleContent

	case boiler.SyndicateMotionTypeAPPOINT_DIRECTOR:
		if syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return terror.Error(fmt.Errorf("only corporation syndicate can appoint director"), "Only corporation syndicate can appoint director.")
		}
		if !req.Payload.DirectorID.Valid {
			return terror.Error(fmt.Errorf("missing player id"), "Missing player id")
		}
		player, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.DirectorID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", req.Payload.DirectorID.String).Err(err).Msg("Failed to get data from db")
			return terror.Error(err, "Player not found")
		}

		if player.SyndicateID.String != syndicate.ID {
			return terror.Error(fmt.Errorf("not syndicate memeber"), "Player is not a member of the syndicate")
		}

		if player.DirectorOfSyndicateID.Valid {
			return terror.Error(fmt.Errorf("already a director"), "Player is already a director of the syndicate")
		}

		motion.DirectorID = req.Payload.DirectorID

	case boiler.SyndicateMotionTypeREMOVE_DIRECTOR:
		if syndicate.Type != boiler.SyndicateTypeCORPORATION {
			return terror.Error(fmt.Errorf("only corporation syndicate can appoint director"), "Only corporation syndicate can appoint director.")
		}
		if !req.Payload.DirectorID.Valid {
			return terror.Error(fmt.Errorf("missing player id"), "Missing player id")
		}

		player, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.DirectorID.String)
		if err != nil {
			gamelog.L.Error().Str("player id", req.Payload.DirectorID.String).Err(err).Msg("Failed to get data from db")
			return terror.Error(err, "Player not found")
		}

		if player.SyndicateID.String != syndicate.ID {
			return terror.Error(fmt.Errorf("not syndicate memeber"), "Player is not a member of the syndicate")
		}

		if !player.DirectorOfSyndicateID.Valid {
			return terror.Error(fmt.Errorf("not a director"), "Player is not a director of the syndicate")
		}
		motion.DirectorID = req.Payload.DirectorID

	case boiler.SyndicateMotionTypeREMOVE_FOUNDER:
		// check motion is issued

	default:
		gamelog.L.Debug().Str("motion type", req.Payload.Type).Msg("Invalid motion type")
		return terror.Error(fmt.Errorf("invalid motion type"), "Invalid motion type")
	}
	return nil
}

// subscription handlers

const HubKeySyndicateGeneralDetailSubscribe = "SYNDICATE:GENERAL:DETAIL:SUBSCRIBE"

// SyndicateGeneralDetailSubscribeHandler return syndicate general detail (join fee, exit fee, name, symbol_url, available_seat_count)
func (sc *SyndicateWS) SyndicateGeneralDetailSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	syndicateID := cctx.URLParam("syndicate_id")
	if syndicateID == "" {
		return terror.Error(terror.ErrInvalidInput, "Missing syndicate id")
	}

	if user.SyndicateID.String != syndicateID {
		return terror.Error(terror.ErrInvalidInput, "The player does not belong to the syndicate")
	}

	// get syndicate detail
	syndicate, err := GetSyndicateLatestDetail(syndicateID)
	if err != nil {
		return terror.Error(err, "Failed to get syndicate")
	}

	reply(syndicate)

	return nil
}

func GetSyndicateLatestDetail(syndicateID string) (*server.Syndicate, error) {
	syndicate, err := boiler.Syndicates(
		boiler.SyndicateWhere.ID.EQ(syndicateID),
		qm.Load(boiler.SyndicateRels.Players, qm.Select(boiler.PlayerColumns.ID, boiler.PlayerColumns.Username, boiler.PlayerColumns.Gid)),
		qm.Load(boiler.SyndicateRels.Symbol),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get syndicate")
	}

	return server.SyndicateBoilerToServer(syndicate), nil
}
