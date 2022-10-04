package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func NewModToolsController(api *API) {
	api.SecureAdminCommand(HubKeyModToolsGetUser, api.ModToolsGetUser)
	api.SecureAdminCommand(HubKeyModToolsBanUser, api.ModToolBanUser)
	api.SecureAdminCommand(HubKeyModToolsUnbanUser, api.ModToolUnbanUser)
}

const HubKeyModToolsGetUser = "MOD:GET:USER"

type ModToolGetUserReq struct {
	Payload struct {
		GID int `json:"gid"`
	} `json:"payload"`
}

func (api *API) ModToolsGetUser(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolGetUserReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	player, err := boiler.Players(
		boiler.PlayerWhere.Gid.EQ(req.Payload.GID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find player")
	}

	supsAmount := api.Passport.UserBalanceGet(uuid.FromStringOrNil(player.ID))

	userResp, err := db.ModToolGetUserData(player.ID, user.R.Role.RoleType == boiler.RoleNameADMIN, supsAmount)
	if err != nil {
		return terror.Error(err, "Failed to get user data in mod tool")
	}

	reply(userResp)

	return nil
}

const HubKeyModToolsBanUser = "MOD:BAN:USER"

type ModToolBanUserReq struct {
	Payload struct {
		GID               int    `json:"gid"`
		ChatBan           bool   `json:"chat_ban"`
		LocationSelectBan bool   `json:"location_select_ban"`
		SupContributeBan  bool   `json:"sup_contribute_ban"`
		BanDurationHours  int    `json:"ban_duration_hours"`
		BanDurationDays   int    `json:"ban_duration_days"`
		BanReason         string `json:"ban_reason"`
	} `json:"payload"`
}

func (api *API) ModToolBanUser(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolBanUserReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	_, err = boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanSendChat.EQ(req.Payload.ChatBan),
		boiler.PlayerBanWhere.BanLocationSelect.EQ(req.Payload.LocationSelectBan),
		boiler.PlayerBanWhere.BanSupsContribute.EQ(req.Payload.LocationSelectBan),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Player already banned")
	}

	bannedPlayer, err := boiler.Players(
		boiler.PlayerWhere.Gid.EQ(req.Payload.GID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find player")
	}

	startedAt := time.Now()
	banEndAt := startedAt.Add(time.Hour*time.Duration(req.Payload.BanDurationHours)).AddDate(0, 0, req.Payload.BanDurationDays)

	playerBan := &boiler.PlayerBan{
		BanFrom:           boiler.BanFromTypeADMIN,
		BannedPlayerID:    bannedPlayer.ID,
		BannedByID:        user.ID,
		Reason:            req.Payload.BanReason,
		BannedAt:          startedAt,
		EndAt:             banEndAt,
		BanSupsContribute: req.Payload.SupContributeBan,
		BanLocationSelect: req.Payload.LocationSelectBan,
		BanSendChat:       req.Payload.ChatBan,
	}

	err = playerBan.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert ban player")
	}

	banMessage := &MessageSystemBan{
		ID:             uuid.Must(uuid.NewV4()).String(),
		BannedByUser:   user,
		BannedUser:     bannedPlayer,
		FactionID:      bannedPlayer.FactionID,
		Reason:         req.Payload.BanReason,
		BanDuration:    fmt.Sprintf("Banned for %d days and %d hours", req.Payload.BanDurationDays, req.Payload.BanDurationHours),
		IsPermanentBan: false,
		Restrictions:   PlayerBanRestrictions(playerBan),
	}

	cm := &ChatMessage{
		ID:     banMessage.ID,
		Type:   ChatMessageTypeSystemBan,
		SentAt: time.Now(),
		Data:   banMessage,
	}

	msg := &boiler.SystemMessage{
		PlayerID: bannedPlayer.ID,
		SenderID: user.ID,
		Title:    "You've been banned by a Moderator",
		Message:  fmt.Sprintf("You've been banned by moderator %s for the following reasons: %s \n If you think you've been wrongly banned please put a ticket through our support team.", user.Username, req.Payload.BanReason),
	}
	err = msg.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	api.GlobalChat.AddMessage(cm)
	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", bannedPlayer.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{cm})

	reply(true)

	return nil
}

const HubKeyModToolsUnbanUser = "MOD:BAN:USER"

type ModToolUnbanUserReq struct {
	Payload struct {
		PlayerBanID string `json:"player_ban_id"`
		UnbanReason string `json:"unban_reason"`
	} `json:"payload"`
}

func (api *API) ModToolUnbanUser(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolUnbanUserReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	playerBan, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.ManuallyUnbanAt.IsNull(),
		boiler.PlayerBanWhere.ID.EQ(req.Payload.PlayerBanID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find player ban")
	}

	playerBan.ManuallyUnbanReason = null.StringFrom(req.Payload.UnbanReason)
	playerBan.ManuallyUnbanAt = null.TimeFrom(time.Now())
	playerBan.ManuallyUnbanByID = null.StringFrom(user.ID)

	_, err = playerBan.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to unban player")
	}

	msg := &boiler.SystemMessage{
		PlayerID: playerBan.BannedPlayerID,
		SenderID: user.ID,
		Title:    "You've been unbanned by a Moderator",
		Message:  fmt.Sprintf("You've been unbanned by moderator %s for the following reasons: %s", user.Username, req.Payload.UnbanReason),
	}
	err = msg.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", playerBan.BannedPlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	reply(true)

	return nil
}
