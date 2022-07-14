package api

import (
	"encoding/json"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"io"
	"io/ioutil"
	"net/http"
	"server"
	"server/db/boiler"
	"server/helpers"
	"strings"
	"time"
)

type SyndicateIssueMotionRequest struct {
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
}

func (api *API) SyndicateMotionIssue(player *server.Player, w http.ResponseWriter, r *http.Request) (int, error) {

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

	err := api.SyndicateSystem.AddMotion(&boiler.Player{ID: player.ID, SyndicateID: player.}, m)
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
