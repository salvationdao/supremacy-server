package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/h2non/filetype"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"io"
	"io/ioutil"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strings"
	"time"
)

func SyndicateRouter(api *API) chi.Router {
	r := chi.NewRouter()

	// NOTE: syndicate is ONLY available on development at the moment
	if !server.IsDevelopmentEnv() {
		return r
	}

	r.Post("/syndicate/issue_motion", WithError(WithCookie(api, api.SyndicateMotionIssue)))
	r.Post("/syndicate/create", WithError(WithCookie(api, api.SyndicateCreate)))

	return r
}

type SyndicateCreateRequest struct {
	Name                         string                        `json:"name"`
	Symbol                       string                        `json:"symbol"`
	Type                         string                        `json:"type"`
	JoinFee                      decimal.Decimal               `json:"join_fee"`
	MemberMonthlyDues            decimal.Decimal               `json:"member_monthly_dues"`
	DeployingMemberCutPercentage decimal.Decimal               `json:"deploying_member_cut_percentage"`
	MemberAssistCutPercentage    decimal.Decimal               `json:"member_assist_cut_percentage"`
	MechOwnerCutPercentage       decimal.Decimal               `json:"mech_owner_cut_percentage"`
	OriginalMemberIDs            []string                      `json:"original_member_ids"`
	JoinQuestions                []*SyndicateJoinQuestionnaire `json:"join_questions"`
}

type SyndicateJoinQuestionnaire struct {
	Question   string   `json:"question"`
	MustAnswer bool     `json:"must_answer"`
	Type       string   `json:"type"`
	Options    []string `json:"options"`
}

func (api *API) SyndicateCreate(player *server.Player, w http.ResponseWriter, r *http.Request) (int, error) {
	if !player.FactionID.Valid {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("player has no faction"), "Only faction player can create new syndicate.")
	}

	if player.SyndicateID.Valid {
		return http.StatusForbidden, terror.Error(fmt.Errorf("player already has syndicate"), "Only non-syndicate players can create new syndicate.")
	}

	req := &SyndicateCreateRequest{}
	blob, imageData, err := parseUploadRequest(w, r, &req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	// validate input
	if blob == nil || imageData == nil {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing logo data"), "Syndicate logo is required.")
	}

	syndicateName, err := api.SyndicateSystem.SyndicateNameVerification(req.Name)
	if err != nil {
		return http.StatusBadRequest, err
	}

	syndicateSymbol, err := api.SyndicateSystem.SyndicateSymbolVerification(req.Symbol)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if req.JoinFee.LessThan(decimal.Zero) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("join fee less than zero"), "Join fee cannot be less than zero.")
	}

	if req.MemberMonthlyDues.LessThan(decimal.Zero) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("monthly dues less than zero"), "Member monthly fee cannot be less than zero.")
	}

	if req.DeployingMemberCutPercentage.LessThan(decimal.Zero) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("deploying member cut is less than zero"), "Deploying member cut cannot be less than zero.")
	}

	if req.MemberAssistCutPercentage.LessThan(decimal.Zero) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("member assist cut is less than zero"), "member assist cut cannot be less than zero.")
	}

	if req.MechOwnerCutPercentage.LessThan(decimal.Zero) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("mech owner cut is less than zero"), "Mech owner cut cannot be less than zero.")
	}

	if decimal.NewFromInt(100).LessThan(req.MechOwnerCutPercentage.Add(req.DeployingMemberCutPercentage).Add(req.MemberAssistCutPercentage)) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("total battle win cut percentage exceed 100"), "Total percentage of battle win cut cannot exceed 100.")
	}

	if req.Type != boiler.SyndicateTypeCORPORATION && req.Type != boiler.SyndicateTypeDECENTRALISED {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid syndicate type"), "Invalid syndicate type.")
	}

	// make sure original member is in the list
	exist := false
	for _, ogID := range req.OriginalMemberIDs {
		if ogID == player.ID {
			exist = true
			break
		}
	}
	if !exist {
		req.OriginalMemberIDs = append(req.OriginalMemberIDs, player.ID)
	}

	ogMembers, err := boiler.Players(
		boiler.PlayerWhere.ID.IN(req.OriginalMemberIDs),
	).All(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid original member"), "Invalid original members.")
	}

	if len(ogMembers) != len(req.OriginalMemberIDs) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("member not found"), "Contain invalid members.")
	}

	for _, ogm := range ogMembers {
		if ogm.SyndicateID.Valid {
			return http.StatusBadRequest, terror.Error(fmt.Errorf("original member already has syndicate"), fmt.Sprintf("Original member %s already has syndicate", ogm.Username))
		}
	}

	if req.Type == boiler.SyndicateTypeCORPORATION {
		if len(ogMembers) < 3 {
			return http.StatusBadRequest, terror.Error(fmt.Errorf("less than 3 original members"), "Require at least three original members to create a corporation syndicate.")
		}
	}

	// validate questionnaire, if provided
	for _, jq := range req.JoinQuestions {
		if jq.Question == "" {
			return http.StatusBadRequest, terror.Error(fmt.Errorf("missing question"), "Missing question.")
		}

		switch jq.Type {
		case boiler.QuestionnaireTypeTEXT:
			if len(jq.Options) > 0 {
				return http.StatusBadRequest, terror.Error(fmt.Errorf("text question connot contain options"), "Text question does not accept options.")
			}
		case boiler.QuestionnaireTypeSINGLE_SELECT, boiler.QuestionnaireTypeMULTI_SELECT:
			if len(jq.Options) == 0 {
				return http.StatusBadRequest, terror.Error(fmt.Errorf("missing questionnaire options"), "Missing options for selectable question.")
			}
			for _, op := range jq.Options {
				if op == "" {
					return http.StatusBadRequest, terror.Error(fmt.Errorf("option is empty string"), "Questionnaire option is an empty string.")
				}
			}
		default:
			return http.StatusBadRequest, terror.Error(fmt.Errorf("questionnaire type does not exist"), "Questionnaire type does not exist.")
		}
	}

	// create new syndicate
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
		return http.StatusInternalServerError, terror.Error(err, "Failed to create syndicate.")
	}

	defer tx.Rollback()

	syndicate := &boiler.Syndicate{
		Type:                         req.Type,
		FactionID:                    player.FactionID.String,
		FoundedByID:                  player.ID,
		Name:                         syndicateName,
		Symbol:                       syndicateSymbol,
		JoinFee:                      req.JoinFee,
		MemberMonthlyDues:            req.MemberMonthlyDues,
		DeployingMemberCutPercentage: req.DeployingMemberCutPercentage,
		MemberAssistCutPercentage:    req.MemberAssistCutPercentage,
		MechOwnerCutPercentage:       req.MechOwnerCutPercentage,
	}

	if syndicate.Type == boiler.SyndicateTypeCORPORATION {
		syndicate.CeoPlayerID = null.StringFrom(player.ID)
	} else {
		syndicate.AdminID = null.StringFrom(player.ID)
	}

	err = syndicate.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Msg("Failed to insert syndicate into db.")
		return http.StatusInternalServerError, terror.Error(err, "Failed to create syndicate.")
	}

	// change original member syndicate id
	for _, ogm := range ogMembers {
		ogm.SyndicateID = null.StringFrom(syndicate.ID)
		_, err := ogm.Update(tx, boil.Whitelist(boiler.PlayerColumns.SyndicateID))
		if err != nil {
			gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Interface("user", ogm).Msg("Failed to update syndicate id of current user.")
			return http.StatusInternalServerError, terror.Error(err, "Failed to assign syndicate to original member")
		}

		// assign original member as committee
		sc := &boiler.SyndicateCommittee{
			SyndicateID: syndicate.ID,
			PlayerID:    ogm.ID,
		}
		err = sc.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Interface("user", ogm).Msg("Failed to update syndicate id of current user.")
			return http.StatusInternalServerError, terror.Error(err, "Failed assign original member as syndicate committee")
		}

		// assign original member as board of director, if syndicate is a private corporation
		if syndicate.Type == boiler.SyndicateTypeCORPORATION {
			sd := &boiler.SyndicateDirector{
				SyndicateID: syndicate.ID,
				PlayerID:    player.ID,
			}
			err = sd.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Interface("syndicate director", sd).Err(err).Msg("Failed to insert syndicate director")
				return http.StatusInternalServerError, terror.Error(err, "Failed to assign original member as board of director")
			}
		}
	}

	// insert questionnaires
	for i, jq := range req.JoinQuestions {
		sq := &boiler.SyndicateQuestionnaire{
			SyndicateID: syndicate.ID,
			Usage:       boiler.QuestionnaireUsageJOIN_REQUEST,
			Number:      i + 1,
			MustAnswer:  jq.MustAnswer,
			Question:    jq.Question,
			Type:        jq.Type,
		}
		err = sq.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Interface("syndicate questionnaire", sq).Err(err).Msg("Failed to insert syndicate questionnaire.")
			return http.StatusInternalServerError, terror.Error(err, "Failed to store syndicate questionnaire.")
		}

		for _, op := range jq.Options {
			qo := &boiler.QuestionnaireOption{
				QuestionnaireID: sq.ID,
				Content:         op,
			}
			err = qo.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Interface("questionnaire option", qo).Err(err).Msg("Failed to insert syndicate questionnaire option.")
				return http.StatusInternalServerError, terror.Error(err, "Failed to store syndicate questionnaire option.")
			}
		}
	}

	// register syndicate on xsyn server
	err = api.Passport.SyndicateCreateHandler(syndicate.ID, syndicate.FoundedByID, syndicate.Name)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("syndicate", syndicate).Msg("Failed to register syndicate on xsyn server.")
		return http.StatusInternalServerError, terror.Error(err, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return http.StatusInternalServerError, terror.Error(err, "Failed to create syndicate.")
	}

	// add syndicate to the system
	err = api.SyndicateSystem.RegisterSyndicate(syndicate.ID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to add syndicate to the system")
	}

	for _, ogm := range ogMembers {
		err = ogm.L.LoadRole(gamedb.StdConn, true, player, nil)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Unable to convert faction, contact support or try again.")
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s", ogm.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(ogm))
	}

	return helpers.EncodeJSON(w, true)
}

type SyndicateIssueMotionRequest struct {
	LastForDays                     int                 `json:"last_for_days"`
	Type                            string              `json:"type"`
	Reason                          string              `json:"reason"`
	NewSymbol                       null.String         `json:"new_symbol"`
	NewSyndicateName                null.String         `json:"new_syndicate_name"`
	NewJoinFee                      decimal.NullDecimal `json:"new_join_fee"`
	NewMonthlyDues                  decimal.NullDecimal `json:"new_monthly_dues"`
	NewDeployingMemberCutPercentage decimal.NullDecimal `json:"new_deploying_member_cut_percentage"`
	NewMemberAssistCutPercentage    decimal.NullDecimal `json:"new_member_assist_cut_percentage"`
	NewMechOwnerCutPercentage       decimal.NullDecimal `json:"new_mech_owner_cut_percentage"`
	NewSyndicateCutPercentage       decimal.NullDecimal `json:"new_syndicate_cut_percentage"`
	RuleID                          null.String         `json:"rule_id"`
	NewRuleNumber                   null.Int            `json:"new_rule_number"`
	NewRuleContent                  null.String         `json:"new_rule_content"`
	MemberID                        null.String         `json:"member_id"`
}

func (api *API) SyndicateMotionIssue(player *server.Player, w http.ResponseWriter, r *http.Request) (int, error) {
	if !player.SyndicateID.Valid {
		return http.StatusForbidden, terror.Error(fmt.Errorf("player has no syndicate"), "You have not join any syndicate yet.")
	}

	req := SyndicateIssueMotionRequest{}
	blob, imageData, err := parseUploadRequest(w, r, &req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.LastForDays < 1 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("motion duration too short"), "Motion should last for at least a day.")
	}

	// build motion
	m := &boiler.SyndicateMotion{
		Type:                            req.Type,
		Reason:                          req.Reason,
		NewSymbol:                       req.NewSymbol,
		NewSyndicateName:                req.NewSyndicateName,
		NewJoinFee:                      req.NewJoinFee,
		NewMonthlyDues:                  req.NewMonthlyDues,
		NewDeployingMemberCutPercentage: req.NewDeployingMemberCutPercentage,
		NewMemberAssistCutPercentage:    req.NewMemberAssistCutPercentage,
		NewMechOwnerCutPercentage:       req.NewMechOwnerCutPercentage,
		NewSyndicateCutPercentage:       req.NewSyndicateCutPercentage,
		RuleID:                          req.RuleID,
		NewRuleNumber:                   req.NewRuleNumber,
		NewRuleContent:                  req.NewRuleContent,
		MemberID:                        req.MemberID,
		EndAt:                           time.Now().AddDate(0, 0, req.LastForDays),
	}

	blob.File = null.BytesFrom(imageData)

	err = api.SyndicateSystem.AddMotion(&boiler.Player{ID: player.ID, SyndicateID: player.SyndicateID}, m, blob)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return helpers.EncodeJSON(w, true)
}

// parseUploadRequest will read a multipart form request that includes both a file, and a request body
// returns a blob struct, ready to be inserted, as well as decoding json into supplied interface when present
func parseUploadRequest(w http.ResponseWriter, r *http.Request, req interface{}) (*boiler.Blob, []byte, error) {
	// Limit size to 50MB (50<<20)
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	mr, err := r.MultipartReader()
	if err != nil {
		return nil, nil, terror.Error(err)
	}

	var blob *boiler.Blob
	var file []byte

	// TODO: add file type filter
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, terror.Error(err)
		}

		// handle file
		if part.FormName() == "file" {
			data, err := ioutil.ReadAll(part)
			if err != nil {
				return nil, nil, terror.Error(terror.ErrParse, "parse error")
			}

			// get mime type
			kind, err := filetype.Match(data)
			if err != nil {
				return nil, nil, terror.Error(terror.ErrParse, "parse error")
			}

			mimeType := kind.MIME.Value
			extension := kind.Extension

			if kind == filetype.Unknown {
				if !strings.HasSuffix(part.FileName(), ".csv") {
					return nil, nil, terror.Error(fmt.Errorf("file type is unknown"), "")
				}
				mimeType = "text/csv"
				extension = "csv"
			}

			blob = &boiler.Blob{
				FileName:      part.FileName(),
				MimeType:      mimeType,
				Extension:     extension,
				FileSizeBytes: int64(len(data)),
			}

			file = data
		}

		// handle JSON body
		if part.FormName() == "json" {
			err = json.NewDecoder(part).Decode(req)
			if err != nil {
				return nil, nil, terror.Error(err)
			}
		}
	}

	return blob, file, nil
}
