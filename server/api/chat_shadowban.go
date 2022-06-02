package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"server/db/boiler"
	"server/gamedb"

	"github.com/friendsofgo/errors"
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
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	if req.UserID == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	// get fingerprints from user id
	fIDs, err := boiler.PlayerFingerprints(boiler.PlayerFingerprintWhere.PlayerID.EQ(req.UserID)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get user fingerprints %w", err))
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("user has no fingerprints userID: %s", req.UserID))
	}

	// loop through fingerprints and add to chat_banned_fingerprints table
	for _, f := range fIDs {
		banned := &boiler.ChatBannedFingerprint{
			FingerprintID: f.FingerprintID,
		}
		err = banned.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to insert user to banned fingerprints %w", err))
		}
	}

	return http.StatusOK, nil
}

// RemoveShadowbanChatUser removes user from shadow banned fingerprints
func (a *API) RemoveShadowbanChatUser(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &struct {
		UserID string `json:"userID"`
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	if req.UserID == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}

	// get fingerprints from user id
	fIDs, err := boiler.PlayerFingerprints(boiler.PlayerFingerprintWhere.PlayerID.EQ(req.UserID)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get user fingerprints %w", err))
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("user has no fingerprints userID: %s", req.UserID))
	}

	// loop through fingerprints delete from chat_banned_fingerprints table
	for _, f := range fIDs {
		banned := &boiler.ChatBannedFingerprint{
			FingerprintID: f.FingerprintID,
		}
		_, err := banned.Delete(gamedb.StdConn)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to delete user banned fingerprints %w", err))
		}
	}

	return http.StatusOK, nil
}
