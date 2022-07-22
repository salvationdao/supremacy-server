package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
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

	api.SecureUserFactionCommand(HubKeySyndicateJoin, sc.SyndicateJoinHandler)
	api.SecureUserFactionCommand(HubKeySyndicateLeave, sc.SyndicateLeaveHandler)
	api.SecureUserFactionCommand(HubKeySyndicateVoteApplication, sc.SyndicateVoteApplicationHandler)

	// update syndicate settings
	api.SecureUserFactionCommand(HubKeySyndicateVoteMotion, sc.SyndicateVoteMotionHandler)
	api.SecureUserFactionCommand(HubKeySyndicateMotionList, sc.SyndicateMotionListHandler)

	return sc
}

type SyndicateJoinRequest struct {
	Payload struct {
		SyndicateID          string                          `json:"syndicate_id"`
		QuestionnaireAnswers []*SyndicateQuestionnaireAnswer `json:"questionnaire_answers"`
	} `json:"payload"`
}

type SyndicateQuestionnaireAnswer struct {
	QuestionnaireID   string   `json:"questionnaire_id"`
	Answer            string   `json:"answer"`
	SelectedOptionIDs []string `json:"selected_option_ids"`
}

const HubKeySyndicateJoin = "SYNDICATE:JOIN"

var joinSyndicateBucket = leakybucket.NewCollector(1, 1, true)

func (sc *SyndicateWS) SyndicateJoinHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player already has syndicate"), "You already have a syndicate")
	}

	if joinSyndicateBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many join request"), "Too many syndicate join request.")
	}

	// check player has applied any application already
	app, err := boiler.SyndicateJoinApplications(
		boiler.SyndicateJoinApplicationWhere.ApplicantID.EQ(user.ID),
		boiler.SyndicateJoinApplicationWhere.FinalisedAt.IsNull(),
		qm.Load(boiler.SyndicateJoinApplicationRels.Syndicate),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "")
	}

	if app != nil {
		return terror.Error(fmt.Errorf("unfinalised application"), fmt.Sprintf("You have an unfinalised application for joining syndicate '%s'.", app.R.Syndicate.Name))
	}

	req := &SyndicateJoinRequest{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	syndicate, err := boiler.FindSyndicate(gamedb.StdConn, req.Payload.SyndicateID)
	if err != nil {
		return terror.Error(err, "Failed to load syndicate.")
	}

	// validate syndicate questionnaire
	sqs, err := syndicate.SyndicateQuestionnaires(
		qm.Load(boiler.SyndicateQuestionnaireRels.QuestionnaireQuestionnaireOptions),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", syndicate.ID).Msg("Failed to load syndicate questionnaires.")
		return terror.Error(err, "Failed to get syndicate join questionnaire")
	}

	answers := []*boiler.QuestionnaireAnswer{}
	for _, sq := range sqs {
		index := -1
		// check applicant has answered this question
		for i, qa := range req.Payload.QuestionnaireAnswers {
			if qa.QuestionnaireID == sq.ID {
				index = i
				break
			}
		}

		// if question is not answered
		if index == -1 {

			// error, if question is must answer
			if sq.MustAnswer {
				return terror.Error(fmt.Errorf("missing answer"), fmt.Sprintf("Question '%s' must be answered.", sq.Question))
			}

			continue
		}

		applicantAnswer := req.Payload.QuestionnaireAnswers[index]
		answer := &boiler.QuestionnaireAnswer{
			Question: sq.Question,
		}

		switch sq.Type {
		case boiler.QuestionnaireTypeTEXT:
			if applicantAnswer.Answer == "" {
				return terror.Error(fmt.Errorf("empty answer"), fmt.Sprintf("Answer for question '%s' is not provided.", sq.Question))
			}
			answer.Answer = null.StringFrom(applicantAnswer.Answer)
		case boiler.QuestionnaireTypeSINGLE_SELECT:
			for _, opID := range applicantAnswer.SelectedOptionIDs {
				for _, qo := range sq.R.QuestionnaireQuestionnaireOptions {
					// append answer if option exist
					if opID == qo.ID {
						answer.Selections = append(answer.Selections, qo.Content)
					}
				}
			}

			if len(answer.Selections) != 1 {
				return terror.Error(fmt.Errorf("not one answer"), fmt.Sprintf("Question '%s' only allow one answer.", sq.Question))
			}
		case boiler.QuestionnaireTypeMULTI_SELECT:
			for _, opID := range applicantAnswer.SelectedOptionIDs {
				for _, qo := range sq.R.QuestionnaireQuestionnaireOptions {
					// append answer if option exist
					if opID == qo.ID {
						answer.Selections = append(answer.Selections, qo.Content)
					}
				}
			}

			if len(answer.Selections) == 0 {
				return terror.Error(fmt.Errorf("no answer"), fmt.Sprintf("Answer for question '%s' is not provided.", sq.Question))
			}
		}

		// append answer to the list
		answers = append(answers, answer)
	}

	// generate request
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
		return terror.Error(err, "Failed to submit application.")
	}

	defer tx.Rollback()

	// insert application
	app = &boiler.SyndicateJoinApplication{
		ApplicantID: user.ID,
		SyndicateID: syndicate.ID,
		ExpireAt:    time.Now().AddDate(0, 0, 1),
		PaidAmount:  syndicate.JoinFee.Round(0),
	}

	err = app.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("application", app).Err(err).Msg("Failed to insert syndicate join application")
		return terror.Error(err, "Failed to submit application")
	}

	// insert answer
	for _, answer := range answers {
		answer.SyndicateJoinApplicationID = null.StringFrom(app.ID)
		err = answer.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Interface("applicant answer", answer).Err(err).Msg("Failed to insert applicant answer.")
			return terror.Error(err, "Failed to submit application.")
		}
	}

	// check user balance
	err = sc.API.SyndicateSystem.AddJoinApplication(app)
	if err != nil {
		return terror.Error(err, "Failed to submit application")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to submit application.")
	}

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

	// check syndicate remaining member count
	remainSyndicateMemberCount, err := boiler.Players(
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(syndicate.ID)),
	).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("Syndicate id", syndicate.ID).Msg("Failed to delete user from syndicate committees table")
		return terror.Error(err, "Failed to check remain syndicate member count.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return terror.Error(err, "Failed to exit syndicate")
	}

	defer tx.Rollback()

	// if the player is the last member of the syndicate
	if remainSyndicateMemberCount == 1 {
		// liquidate syndicate
		err = sc.API.SyndicateSystem.LiquidateSyndicate(tx, syndicate.ID)
		if err != nil {
			return err
		}
	} else {
		// remove player from the syndicate
		user.SyndicateID = null.StringFromPtr(nil)
		_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
		if err != nil {
			gamelog.L.Error().Err(err).Interface("user", user).Msg("Failed to update user syndicate column")
			return terror.Error(err, "Failed to exit syndicate")
		}

		// remove user from syndicate director list
		_, err = boiler.SyndicateDirectors(
			boiler.SyndicateDirectorWhere.SyndicateID.EQ(syndicate.ID),
			boiler.SyndicateDirectorWhere.PlayerID.EQ(user.ID),
		).DeleteAll(tx)
		if err != nil {
			gamelog.L.Error().Err(err).Str("user id", user.ID).Str("Syndicate id", syndicate.ID).Msg("Failed to delete user from syndicate director table")
			return terror.Error(err, "failed to remove user from syndicate director list")
		}

		// remove user from syndicate committee list
		_, err = boiler.SyndicateCommittees(
			boiler.SyndicateDirectorWhere.SyndicateID.EQ(syndicate.ID),
			boiler.SyndicateDirectorWhere.PlayerID.EQ(user.ID),
		).DeleteAll(tx)
		if err != nil {
			gamelog.L.Error().Err(err).Str("user id", user.ID).Str("Syndicate id", syndicate.ID).Msg("Failed to delete user from syndicate committees table")
			return terror.Error(err, "Failed to remove user from syndicate committee list.")
		}

		// check whether the player is syndicate admin
		if syndicate.AdminID.Valid && syndicate.AdminID.String == user.ID {
			// remove admin role of the syndicate
			syndicate.AdminID = null.StringFromPtr(nil)
			_, err := syndicate.Update(tx, boil.Whitelist(boiler.SyndicateColumns.AdminID))
			if err != nil {
				gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to remove syndicate admin")
				return terror.Error(err, "Failed to remove admin role from syndicate.")
			}

			if syndicate.Type == boiler.SyndicateTypeDECENTRALISED {
				// terminate any depose admin motion
				err = sc.API.SyndicateSystem.ForceCloseMotionsByType(syndicate.ID, "Admin player has already left the syndicate", boiler.SyndicateMotionTypeDEPOSE_ADMIN)
				if err != nil {
					return err
				}
			}
		}

		// check director number
		directorCount, err := boiler.SyndicateCommittees(
			boiler.SyndicateDirectorWhere.SyndicateID.EQ(syndicate.ID),
		).Count(tx)
		if err != nil {
			gamelog.L.Error().Err(err).Str("user id", user.ID).Str("Syndicate id", syndicate.ID).Msg("Failed to delete user from syndicate committees table")
			return terror.Error(err, "Failed to remove user from syndicate committee list.")
		}

		// check syndicate ceo
		if syndicate.CeoPlayerID.Valid && syndicate.CeoPlayerID.String == user.ID {
			syndicate.CeoPlayerID = null.StringFromPtr(nil)
			syndicate.AdminID = null.StringFromPtr(nil)
			_, err := syndicate.Update(tx, boil.Whitelist(boiler.SyndicateColumns.CeoPlayerID, boiler.SyndicateColumns.AdminID))
			if err != nil {
				gamelog.L.Error().Interface("syndicate", syndicate).Err(err).Msg("Failed to remove syndicate ceo")
				return terror.Error(err, "Failed to remove ceo role from syndicate.")
			}

			// terminate depose ceo motion
			err = sc.API.SyndicateSystem.ForceCloseMotionsByType(syndicate.ID, "CEO has already left the syndicate", boiler.SyndicateMotionTypeDEPOSE_ADMIN)
			if err != nil {
				return err
			}
		}

		if syndicate.Type == boiler.SyndicateTypeCORPORATION && directorCount == 0 {
			// change corporation syndicate to decentralised syndicate if there is no directors
			syndicate.Type = boiler.SyndicateTypeDECENTRALISED
			_, err := syndicate.Update(tx, boil.Whitelist(boiler.SyndicateColumns.Type))
			if err != nil {
				gamelog.L.Error().Err(err).Str("Syndicate id", syndicate.ID).Str("syndicate type", syndicate.Type).Msg("Failed to change syndicate type")
				return terror.Error(err, "Failed to change syndicate type")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return terror.Error(err, "Failed to exit syndicate")
	}

	// broadcast updated user
	ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), server.HubKeyUserSubscribe, user)

	// broadcast latest syndicate detail
	serverSyndicate, err := db.GetSyndicateDetail(syndicate.ID)
	if err != nil {
		return err
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s", syndicate.FactionID, syndicate.ID), server.HubKeySyndicateGeneralDetailSubscribe, serverSyndicate)

	// broadcast directors
	directors, err := db.GetSyndicateDirectors(syndicate.ID)
	if err != nil {
		return err
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/directors", syndicate.FactionID, syndicate.ID), server.HubKeySyndicateDirectorsSubscribe, directors)

	// broadcast committees
	scs, err := db.GetSyndicateCommittees(syndicate.ID)
	if err != nil {
		return err
	}
	ws.PublishMessage(fmt.Sprintf("/faction/%s/syndicate/%s/committees", syndicate.FactionID, syndicate.ID), server.HubKeySyndicateCommitteesSubscribe, scs)

	reply(true)
	return nil
}

type SyndicateVoteApplicationRequest struct {
	Payload struct {
		ApplicationID string `json:"application_id"`
		IsAgreed      bool   `json:"is_agreed"`
	} `json:"payload"`
}

const HubKeySyndicateVoteApplication = "SYNDICATE:VOTE:APPLICATION"

func (sc *SyndicateWS) SyndicateVoteApplicationHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.SyndicateID.Valid {
		return terror.Error(fmt.Errorf("player has no syndicate"), "You have not join any syndicate yet.")
	}

	// check is syndicate committee member
	isCommittee, err := boiler.SyndicateCommittees(
		boiler.SyndicateCommitteeWhere.SyndicateID.EQ(user.SyndicateID.String),
		boiler.SyndicateCommitteeWhere.PlayerID.EQ(user.ID),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", user.SyndicateID.String).Str("player id", user.ID).Msg("Failed to syndicate committee")
		return terror.Error(err, "Failed to check committee members")
	}

	if !isCommittee {
		return terror.Error(fmt.Errorf("not a committee"), "Only syndicate committee can vote on join application")
	}

	req := &SyndicateVoteApplicationRequest{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

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

// SyndicateCommitteesSubscribeHandler return the committees of the syndicate
func (sc *SyndicateWS) SyndicateCommitteesSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	syndicateID := cctx.URLParam("syndicate_id")
	if syndicateID == "" {
		return terror.Error(terror.ErrInvalidInput, "Missing syndicate id")
	}

	if user.SyndicateID.String != syndicateID {
		return terror.Error(terror.ErrInvalidInput, "The player does not belong to the syndicate")
	}

	// get syndicate detail
	ps, err := db.GetSyndicateCommittees(syndicateID)
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
