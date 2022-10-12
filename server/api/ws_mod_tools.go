package api

import (
	"context"
	"encoding/json"
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
		GID               []int  `json:"gid"`
		ChatBan           bool   `json:"chat_ban"`
		LocationSelectBan bool   `json:"location_select_ban"`
		SupContributeBan  bool   `json:"sup_contribute_ban"`
		BanDurationHours  int    `json:"ban_duration_hours"`
		BanDurationDays   int    `json:"ban_duration_days"`
		BanReason         string `json:"ban_reason"`
		BanMechQueue      bool   `json:"ban_mech_queue"`
		IsShadowBan       bool   `json:"is_shadow_ban"`
	} `json:"payload"`
}

func (api *API) ModToolBanUser(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolBanUserReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	bannedPlayers, err := boiler.Players(
		boiler.PlayerWhere.Gid.IN(req.Payload.GID),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find player")
	}

	startedAt := time.Now()
	banEndAt := startedAt.Add(time.Hour*time.Duration(req.Payload.BanDurationHours)).AddDate(0, 0, req.Payload.BanDurationDays)

	for _, bannedPlayer := range bannedPlayers {
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
			BanMechQueue:      req.Payload.BanMechQueue,
		}

		err = playerBan.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Failed to insert ban player")
		}

		msg := &boiler.SystemMessage{
			PlayerID: bannedPlayer.ID,
			SenderID: user.ID,
			Title:    "You've been banned by a Moderator",
			Message:  fmt.Sprintf("You've been banned by moderator for the following reasons: %s \n If you think you've been wrongly banned please put a ticket through our support team.", req.Payload.BanReason),
		}
		err = msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", bannedPlayer.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)

		if !req.Payload.IsShadowBan {
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
				Type:   ChatMessageTypeModBan,
				SentAt: time.Now(),
				Data:   banMessage,
			}

			api.GlobalChat.AddMessage(cm)

			ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{cm})
		}

	}

	reply(true)

	return nil
}

const HubKeyModToolsUnbanUser = "MOD:UNBAN:USER"

type ModToolUnbanUserReq struct {
	Payload struct {
		PlayerBanID []string `json:"player_ban_id"`
		UnbanReason string   `json:"unban_reason"`
	} `json:"payload"`
}

func (api *API) ModToolUnbanUser(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolUnbanUserReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	playerBans, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.ManuallyUnbanAt.IsNull(),
		boiler.PlayerBanWhere.ID.IN(req.Payload.PlayerBanID),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to find player ban")
	}

	for _, playerBan := range playerBans {
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
			Message:  fmt.Sprintf("You've been unbanned by moderator for the following reasons: %s", req.Payload.UnbanReason),
		}
		err = msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", playerBan.BannedPlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	}
	reply(true)

	return nil
}
