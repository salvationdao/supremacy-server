package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// ShadowbanChatPlayer shadow bans a player from chat
func (api *API) ShadowbanChatPlayer(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &struct {
		ColumnName string `json:"column_name"`
		Value      string `json:"value"`
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))

	}

	if req.ColumnName != "id" && req.ColumnName != "public_address" && req.ColumnName != "username" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request: column_name must be 'id', 'public_address' or 'username'"))
	}

	// get player
	player, err := boiler.Players(qm.Where(fmt.Sprintf("%s = ?", req.ColumnName), req.Value)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player by: %s: %s err: %w", req.ColumnName, req.Value, err))
	}

	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("cant find player %s: %s %w", req.ColumnName, req.Value, err))
	}

	// get fingerprints from player id
	fIDs, err := boiler.PlayerFingerprints(boiler.PlayerFingerprintWhere.PlayerID.EQ(player.ID)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player fingerprints %w", err))
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("player has no fingerprints playerID: %s", player.ID))
	}

	// loop through fingerprints and add to chat_banned_fingerprints table
	for _, f := range fIDs {
		banned := &boiler.ChatBannedFingerprint{
			FingerprintID: f.FingerprintID,
		}
		err = banned.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to insert player to banned fingerprints %w", err))
		}
	}
	helpers.EncodeJSON(w, fmt.Sprintf("player %s: %s has been banned successfully", req.ColumnName, req.Value))

	return http.StatusOK, nil
}

// RemoveShadowbanChatPlayer removes player from shadow banned fingerprints
func (api *API) ShadowbanChatPlayerRemove(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &struct {
		ColumnName string `json:"column_name"`
		Value      string `json:"value"`
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request %w", err))
	}
	if req.ColumnName != "id" && req.ColumnName != "public_address" && req.ColumnName != "username" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invalid request: column_name must be 'id', 'public_address' or 'username'"))
	}
	// get player
	player, err := boiler.Players(qm.Where(fmt.Sprintf("%s = ?", req.ColumnName), req.Value)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player by: %s: %s err: %w", req.ColumnName, req.Value, err))
	}

	// get fingerprints from player id
	fIDs, err := boiler.PlayerFingerprints(boiler.PlayerFingerprintWhere.PlayerID.EQ(player.ID)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get player fingerprints %w", err))
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("player has no fingerprints playerID: %s", player.ID))
	}

	// loop through fingerprints delete from chat_banned_fingerprints table
	for _, f := range fIDs {
		banned := &boiler.ChatBannedFingerprint{
			FingerprintID: f.FingerprintID,
		}
		_, err := banned.Delete(gamedb.StdConn)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to delete player banned fingerprints %w", err))
		}
	}

	helpers.EncodeJSON(w, fmt.Sprintf("player %s: %s has been un banned successfully", req.ColumnName, req.Value))

	return http.StatusOK, nil
}

// ShadowbanChatPlayerList lists all players that are chat shadow banned
func (api *API) ShadowbanChatPlayerList(w http.ResponseWriter, r *http.Request) (int, error) {
	// get banned players
	bannedPlayers, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.PublicAddress,
		),
		// join fingerprints
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.PlayerFingerprints,
				qm.Rels(boiler.TableNames.PlayerFingerprints, boiler.PlayerFingerprintColumns.PlayerID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			),
		),
		// join banned fingerprints
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.ChatBannedFingerprints,
				qm.Rels(boiler.TableNames.ChatBannedFingerprints, boiler.ChatBannedFingerprintColumns.FingerprintID),
				qm.Rels(boiler.TableNames.PlayerFingerprints, boiler.PlayerFingerprintColumns.FingerprintID),
			),
		),
	).All(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to get banned players %w", err))
	}

	helpers.EncodeJSON(w, bannedPlayers)

	return http.StatusOK, nil

}
