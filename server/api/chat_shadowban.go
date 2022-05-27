package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// ShadowbanChatUser shadow bans a user from chat
func (a *API) ShadowbanChatUser(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &struct {
		UserID string `json:"userID"`
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invaid request %w", err))
	}
	if req.UserID == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invaid request %w", err))
	}

	// get fingerprints
	fIDs := []string{}
	err = a.Passport.XrpcClient.Call("S.UserFingerPrints", req.UserID, fIDs)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get user fingerprints %w", err))
	}

	if fIDs == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("user has no fingerprints userID: %s", req.UserID))
	}

	// loop through fingerprints and add to chat_banned_fingerprints table
	for _, id := range fIDs {
		banned := &boiler.ChatBannedFingerprint{
			FingerprintID: id,
		}
		err = banned.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to insert user to banned fingerprints %w", err))
		}
	}

	return http.StatusOK, nil
}
