package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
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

	//api.SecureUserFactionCommand(HubKeySyndicateList, sc.Sy)

	// update syndicate settings
	api.SecureUserFactionCommand(HubKeySyndicateIssueMotion, sc.SyndicateIssueMotionHandler)
	api.SecureUserFactionCommand(HubKeySyndicateVoteMotion, sc.SyndicateVoteMotionHandler)
	api.SecureUserFactionCommand(HubKeySyndicateMotionList, sc.SyndicateMotionListHandler)
	api.SecureUserFactionCommand(server.HubKeySyndicateOngoingMotionSubscribe, sc.SyndicateOngoingMotionSubscribeHandler)

	// subscribetion

	return sc
}

// TODO: Move Syndicate create request to rest, and add file update handler

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
	syndicateName, err := sc.API.SyndicateSystem.SyndicateNameVerification(req.Payload.Name)
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

	if syndicate.Type == boiler.SyndicateTypeCORPORATION {
		syndicate.CeoPlayerID = null.StringFrom(user.ID)
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

	if syndicate.Type == boiler.SyndicateTypeCORPORATION {
		// TODO: insert directors table

		// TODO: insert committees table
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
	serverSyndicate, err := db.GetSyndicateDetail(syndicate.ID)
	if err != nil {
		return terror.Error(err, "Failed to get syndicate detail")
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s", syndicate.FactionID, syndicate.ID), server.HubKeySyndicateGeneralDetailSubscribe, serverSyndicate)

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
	serverSyndicate, err := db.GetSyndicateDetail(syndicate.ID)
	if err != nil {
		return terror.Error(err, "Failed to get syndicate detail")
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s", syndicate.FactionID, syndicate.ID), server.HubKeySyndicateGeneralDetailSubscribe, serverSyndicate)

	reply(true)

	return nil
}

type SyndicateIssueMotionRequest struct {
	Payload struct {
		LastForDays                     int                 `json:"last_for_days"`
		Type                            string              `json:"type"`
		Reason                          string              `json:"reason"`
		NewSymbol                       null.String         `json:"new_symbol"`
		NewSyndicateName                null.String         `json:"new_syndicate_name"`
		NewNamingConvention             null.String         `json:"new_naming_convention"`
		NewJoinFee                      decimal.NullDecimal `json:"new_join_fee"`
		NewExitFee                      decimal.NullDecimal `json:"new_exit_fee"`
		NewDeployingMemberCutPercentage decimal.NullDecimal `json:"new_deploying_member_cut_percentage"`
		NewMemberAssistCutPercentage    decimal.NullDecimal `json:"new_member_assist_cut_percentage"`
		NewMechOwnerCutPercentage       decimal.NullDecimal `json:"new_mech_owner_cut_percentage"`
		NewSyndicateCutPercentage       decimal.NullDecimal `json:"new_syndicate_cut_percentage"`
		RuleID                          null.String         `json:"rule_id"`
		NewRuleNumber                   null.Int            `json:"new_rule_number"`
		NewRuleContent                  null.String         `json:"new_rule_content"`
		DirectorID                      null.String         `json:"director_id"`
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

	if req.Payload.LastForDays < 1 {
		return terror.Error(fmt.Errorf("negative duration"), "The period of the motion cannot be less than 1 day")
	}

	// build motion
	m := &boiler.SyndicateMotion{
		Type:                            req.Payload.Type,
		Reason:                          req.Payload.Reason,
		NewSymbol:                       req.Payload.NewSymbol,
		NewSyndicateName:                req.Payload.NewSyndicateName,
		NewNamingConvention:             req.Payload.NewNamingConvention,
		NewJoinFee:                      req.Payload.NewJoinFee,
		NewExitFee:                      req.Payload.NewExitFee,
		NewDeployingMemberCutPercentage: req.Payload.NewDeployingMemberCutPercentage,
		NewMemberAssistCutPercentage:    req.Payload.NewMemberAssistCutPercentage,
		NewMechOwnerCutPercentage:       req.Payload.NewMechOwnerCutPercentage,
		NewSyndicateCutPercentage:       req.Payload.NewSyndicateCutPercentage,
		RuleID:                          req.Payload.RuleID,
		NewRuleNumber:                   req.Payload.NewRuleNumber,
		NewRuleContent:                  req.Payload.NewRuleContent,
		DirectorID:                      req.Payload.DirectorID,
		EndedAt:                         time.Now().AddDate(0, 0, req.Payload.LastForDays),
	}

	err = sc.API.SyndicateSystem.AddMotion(user, m)
	if err != nil {
		return terror.Error(err, "Failed to add motion")
	}

	reply(true)
	return nil
}

type SyndicateMotionVoteRequest struct {
	Payload struct {
		MotionID string `json:"motion_id"`
		IsAgreed bool   `json:"is_agreed"`
	} `json:"payload"`
}

const HubKeySyndicateVoteMotion = "SYNDICATE:VOTE:MOTION"

func (sc *SyndicateWS) SyndicateVoteMotionHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player has no syndicate"), "You have not join any syndicate yet.")
	}

	req := &SyndicateMotionVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	err = sc.API.SyndicateSystem.VoteMotion(user, req.Payload.MotionID, req.Payload.IsAgreed)
	if err != nil {
		return err
	}

	reply(true)
	return nil
}

type SyndicateMotionListRequest struct {
	Payload struct {
		Filter     *db.SyndicateMotionListFilter `json:"filter"`
		PageSize   int                           `json:"page_size"`
		PageNumber int                           `json:"page_number"`
	} `json:"payload"`
}

type SyndicateMotionListResponse struct {
	SyndicateMotions []*boiler.SyndicateMotion `json:"syndicate_motions"`
	Total            int64                     `json:"total"`
}

const HubKeySyndicateMotionList = "SYNDICATE:MOTION:LIST"

func (sc *SyndicateWS) SyndicateMotionListHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player has no syndicate"), "You have not join any syndicate yet.")
	}

	req := &SyndicateMotionListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	limit := req.Payload.PageSize
	offset := req.Payload.PageNumber * req.Payload.PageSize

	sms, total, err := db.SyndicateMotionList(user.SyndicateID.String, req.Payload.Filter, limit, offset)
	if err != nil {
		return err
	}

	reply(&SyndicateMotionListResponse{sms, total})

	return nil
}

// subscription handlers

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
	s, err := db.GetSyndicateDetail(syndicateID)
	if err != nil {
		return terror.Error(err, "Failed to load syndicate detail.")
	}

	reply(s)
	return nil
}

// SyndicateDirectorsSubscribeHandler return the directors of the syndicate
func (sc *SyndicateWS) SyndicateDirectorsSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	syndicateID := cctx.URLParam("syndicate_id")
	if syndicateID == "" {
		return terror.Error(terror.ErrInvalidInput, "Missing syndicate id")
	}

	if user.SyndicateID.String != syndicateID {
		return terror.Error(terror.ErrInvalidInput, "The player does not belong to the syndicate")
	}

	// get syndicate detail
	ps, err := db.GetSyndicateDirectors(syndicateID)
	if err != nil {
		return err
	}

	reply(ps)
	return nil
}

// SyndicateOngoingMotionSubscribeHandler return ongoing motion list
func (sc *SyndicateWS) SyndicateOngoingMotionSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	syndicateID := cctx.URLParam("syndicate_id")
	if syndicateID == "" {
		return terror.Error(terror.ErrInvalidInput, "Missing syndicate id")
	}

	if user.SyndicateID.String != syndicateID {
		return terror.Error(terror.ErrInvalidInput, "The player does not belong to the syndicate")
	}

	oms, err := sc.API.SyndicateSystem.GetOngoingMotions(user)
	if err != nil {
		return terror.Error(err, "Failed to get ongoing motions")
	}

	reply(oms)

	return nil
}
