package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/slack"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

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
	api.SecureAdminCommand(HubKeyModToolRestartServer, api.ModToolRestartServer)
	api.SecureAdminCommand(HubKeyModToolLookupHistory, api.ModToolLookupHistory)
	api.SecureAdminCommand(HubKeyModToolRenameMech, api.ModToolRenameMech)
	api.SecureAdminCommand(HubKeyModToolRenamePlayer, api.ModToolRenamePlayer)
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

	err = db.UpdateLookupHistory(user.ID, player.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, db.ErrModLookupDuplicate) {
		return terror.Error(err, "Failed to update lookup history")
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
			SenderID: server.SupremacySystemModeratorUserID,
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

		banTypeString := ""

		if req.Payload.LocationSelectBan {
			banTypeString = banTypeString + "\n- Location select ban"
		}

		if req.Payload.BanMechQueue {
			banTypeString = banTypeString + "\n- Mech queueing"
		}

		if req.Payload.SupContributeBan {
			banTypeString = banTypeString + "\n- Sup contributing"
		}

		if req.Payload.ChatBan {
			banTypeString = banTypeString + "\n- Chat ban"
		}

		audit := &boiler.ModActionAudit{
			AffectedPlayerID: null.StringFrom(playerBan.BannedPlayerID),
			ActionType:       boiler.ModActionTypeBAN,
			ModID:            user.ID,
			Reason:           req.Payload.BanReason,
			PlayerBanID:      null.StringFrom(playerBan.ID),
		}

		err = audit.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Failed to insert audit please try again")
		}

		slackMessage := fmt.Sprintf("<!subteam^S03GCC87CD7>\n\n:x: `%s#%d` has banned user `%s#%d` :x: \n\n```Reasons: %s\nBan End At: %s\nBan Type:%s```", user.Username.String, user.Gid, bannedPlayer.Username.String, bannedPlayer.Gid, req.Payload.BanReason, banEndAt.String(), banTypeString)

		err = slack.SendSlackNotification(slackMessage, db.GetStrWithDefault(db.KeySlackModChannelID, "C03GDHLV9FE"), slack.ModToolsAppToken)
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to send slack notification for banning user")
		}

		// send discord notif
		err = api.Discord.SendDiscordMessage(slackMessage)
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to send discord notification for banning user")
		}

		gamelog.L.Info().Str("Mod Action", "Ban").Interface("Mod Audit", audit).Msg("Mod tool event")

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
			SenderID: server.SupremacySystemModeratorUserID,
			Title:    "You've been unbanned by a Moderator",
			Message:  fmt.Sprintf("You've been unbanned by moderator for the following reasons: %s", req.Payload.UnbanReason),
		}
		err = msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", playerBan.BannedPlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)

		player, err := boiler.FindPlayer(gamedb.StdConn, playerBan.BannedPlayerID)
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to find player for unbanning")
			continue
		}

		audit := &boiler.ModActionAudit{
			AffectedPlayerID: null.StringFrom(playerBan.BannedPlayerID),
			ActionType:       boiler.ModActionTypeUNBAN,
			ModID:            user.ID,
			Reason:           req.Payload.UnbanReason,
			PlayerBanID:      null.StringFrom(playerBan.ID),
		}

		err = audit.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Failed to insert audit please try again")
		}

		unbackFrom := ""

		if playerBan.BanLocationSelect {
			unbackFrom = unbackFrom + "\n- Location select ban"
		}

		if playerBan.BanMechQueue {
			unbackFrom = unbackFrom + "\n- Mech queueing"
		}

		if playerBan.BanSupsContribute {
			unbackFrom = unbackFrom + "\n- Sup contributing"
		}

		if playerBan.BanSendChat {
			unbackFrom = unbackFrom + "\n- Chat ban"
		}

		slackMessage := fmt.Sprintf("<!subteam^S03GCC87CD7>\n\n:white_check_mark: `%s#%d` has unbanned user `%s#%d` :white_check_mark: \n\n```Reasons: %s\nUnbanned from:%s```", user.Username.String, user.Gid, player.Username.String, player.Gid, req.Payload.UnbanReason, unbackFrom)

		err = slack.SendSlackNotification(slackMessage, db.GetStrWithDefault(db.KeySlackModChannelID, "C03GDHLV9FE"), slack.ModToolsAppToken)
		if err != nil {
			gamelog.L.Err(err).Msg("Failed to send slack notification for unbanning user")
		}

		gamelog.L.Info().Str("Mod Action", "Unban").Interface("Mod Audit", audit).Msg("Mod tool event")

	}

	reply(true)

	return nil
}

const HubKeyModToolRestartServer = "MOD:RESTART:SERVER"

type ModToolRestartServer struct {
	Payload struct {
		Reason string `json:"reason"`
	} `json:"payload"`
}

func (api *API) ModToolRestartServer(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolRestartServer{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.Reason == "" {
		return terror.Error(fmt.Errorf("no reason provided for restarting server"), "PLease provide a reason before attempting to restart server")
	}

	audit := &boiler.ModActionAudit{
		ActionType: boiler.ModActionTypeRESTART,
		ModID:      user.ID,
		Reason:     req.Payload.Reason,
	}

	err = audit.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert audit please try again")
	}

	slackMessage := fmt.Sprintf("<!channel>\n\n:warning: `%s#%d` has restarted Gameserver :warning: \n\n```Reason: %s```", user.Username.String, user.Gid, req.Payload.Reason)

	err = slack.SendSlackNotification(slackMessage, db.GetStrWithDefault(db.KeySlackRapiChannelID, "C03F29D12BA"), slack.ModToolsAppToken)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to send slack notification for banning user")
	}

	slackMessage = fmt.Sprintf("<!subteam^S03GCC87CD7>\n\n:warning: `%s#%d` has restarted Gameserver :warning: \n\n```Reason: %s```", user.Username.String, user.Gid, req.Payload.Reason)

	err = slack.SendSlackNotification(slackMessage, db.GetStrWithDefault(db.KeySlackModChannelID, "C03GDHLV9FE"), slack.ModToolsAppToken)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to send slack notification for banning user")
	}

	gamelog.L.Warn().Str("Mod Action", "Restart").Interface("Mod Audit", audit).Msg("Mod tool event")

	os.Exit(1)

	return nil
}

const HubKeyModToolLookupHistory = "MOD:LOOKUP:HISTORY"

type ModLookupHistoryResp struct {
	Username  string    `json:"username"`
	GID       int       `json:"gid"`
	FactionID string    `json:"faction_id"`
	VisitedOn time.Time `json:"visited_on"`
}

func (api *API) ModToolLookupHistory(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	actions, err := boiler.ModActionAudits(
		boiler.ModActionAuditWhere.ModID.EQ(user.ID),
		boiler.ModActionAuditWhere.ActionType.EQ(boiler.ModActionTypeLOOKUP),
		qm.OrderBy(fmt.Sprintf("%s DESC", boiler.ModActionAuditColumns.CreatedAt)),
		qm.Load(boiler.ModActionAuditRels.AffectedPlayer),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "error finding mod lookup actions")
	}

	modLookupHistoryResp := []*ModLookupHistoryResp{}

	if len(actions) <= 0 {
		reply(modLookupHistoryResp)
		return nil
	}

	for _, action := range actions {
		if action.R == nil || action.R.AffectedPlayer == nil {
			continue
		}

		lookupHistory := &ModLookupHistoryResp{
			Username:  action.R.AffectedPlayer.Username.String,
			GID:       action.R.AffectedPlayer.Gid,
			VisitedOn: action.CreatedAt,
		}

		if action.R.AffectedPlayer.FactionID.Valid {
			lookupHistory.FactionID = action.R.AffectedPlayer.FactionID.String
		}

		modLookupHistoryResp = append(modLookupHistoryResp, lookupHistory)
	}

	reply(modLookupHistoryResp)

	return nil
}

const HubKeyModToolRenameMech = "MOD:RENAME:MECH"

type ModToolRenameMechReq struct {
	Payload struct {
		MechID      string `json:"mech_id"`
		OwnerID     string `json:"owner_id"`
		NewMechName string `json:"new_mech_name"`
		Reason      string `json:"reason"`
	} `json:"payload"`
}

func (api *API) ModToolRenameMech(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolRenameMechReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	err = IsValidUsername(req.Payload.NewMechName)
	if err != nil {
		return terror.Error(err, "Invalid username, must be between 3 - 15 characters long, cannot contain profanities.")
	}

	mech, err := boiler.FindMech(gamedb.StdConn, req.Payload.MechID)
	if err != nil {
		return terror.Error(err, "Failed to find mech")
	}

	reason := fmt.Sprintf("Reason: %s \n Previous name: %s \n New name: %s", req.Payload.Reason, mech.Name, req.Payload.NewMechName)

	action := &boiler.ModActionAudit{
		AffectedPlayerID: null.StringFrom(req.Payload.OwnerID),
		ActionType:       boiler.ModActionTypeMECH_RENAME,
		ModID:            user.ID,
		Reason:           reason,
	}

	err = action.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert mod audit action")
	}

	oldName := mech.Name
	newName, err := db.MechRename(mech.ID, req.Payload.OwnerID, req.Payload.NewMechName)
	if err != nil {
		return terror.Error(err, "Failed to rename mech")
	}

	slackMessage := fmt.Sprintf("<!subteam^S03GCC87CD7>\n\n:information_source: `%s#%d` has renamed a mech :information_source: \n\n```Reason: %s```", user.Username.String, user.Gid, reason)

	err = slack.SendSlackNotification(slackMessage, db.GetStrWithDefault(db.KeySlackModChannelID, "C03GDHLV9FE"), slack.ModToolsAppToken)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to send slack notification for banning user")
	}

	msg := &boiler.SystemMessage{
		PlayerID: req.Payload.OwnerID,
		SenderID: server.SupremacySystemModeratorUserID,
		Title:    "One of your mechs has been renamed by Moderator",
		Message:  fmt.Sprintf("Your mech \"%s\" has been renamed to \"%s\" by a moderator for the following reasons: %s", oldName, newName, req.Payload.Reason),
	}
	err = msg.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", req.Payload.OwnerID), server.HubKeySystemMessageListUpdatedSubscribe, true)

	reply(newName)

	return nil
}

const HubKeyModToolRenamePlayer = "MOD:RENAME:PLAYER"

type ModToolRenamePlayerReq struct {
	Payload struct {
		PlayerID    string `json:"player_id"`
		NewUsername string `json:"new_username"`
		Reason      string `json:"reason"`
	} `json:"payload"`
}

func (api *API) ModToolRenamePlayer(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &ModToolRenamePlayerReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	err = IsValidUsername(req.Payload.NewUsername)
	if err != nil {
		return terror.Error(err, "Invalid username, must be between 3 - 15 characters long, cannot contain profanities.")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.PlayerID)
	if err != nil {
		return terror.Error(err, "Failed to find player")
	}

	if !player.Username.Valid {
		return terror.Error(fmt.Errorf("user has invalid username"), "User has invalid username")
	}

	oldName := player.Username.String

	reason := fmt.Sprintf("Reason: %s \n Previous name: %s \n New name: %s", req.Payload.Reason, oldName, req.Payload.NewUsername)

	action := &boiler.ModActionAudit{
		AffectedPlayerID: null.StringFrom(req.Payload.PlayerID),
		ActionType:       boiler.ModActionTypeMECH_RENAME,
		ModID:            user.ID,
		Reason:           reason,
	}

	err = action.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert mod audit action")
	}

	player.Username = null.StringFrom(req.Payload.NewUsername)
	player.UpdatedAt = time.Now()

	_, err = player.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to update users name on ")
	}

	// update in xsyn
	err = api.Passport.UserUpdateUsername(player.ID, req.Payload.NewUsername)
	if err != nil {
		return terror.Error(err, "Failed to update users name on XSYN")
	}
	reply(player.Username.String)
	err = player.L.LoadRole(gamedb.StdConn, true, player, nil)
	if err != nil {
		return terror.Error(err, "Failed to update player's marketing preferences.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/player/%s", player.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(player))

	msg := &boiler.SystemMessage{
		PlayerID: req.Payload.PlayerID,
		SenderID: server.SupremacySystemModeratorUserID,
		Title:    "Your username has been modified by a moderator",
		Message:  fmt.Sprintf("Your username \"%s\" has been renamed to \"%s\" by a moderator for the following reasons: %s", oldName, req.Payload.NewUsername, req.Payload.Reason),
	}
	err = msg.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", player.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)

	slackMessage := fmt.Sprintf("<!subteam^S03GCC87CD7>\n\n:information_source: `%s#%d` has renamed a user :information_source: \n\n```Reason: %s```", user.Username.String, user.Gid, reason)

	err = slack.SendSlackNotification(slackMessage, db.GetStrWithDefault(db.KeySlackModChannelID, "C03GDHLV9FE"), slack.ModToolsAppToken)
	if err != nil {
		gamelog.L.Err(err).Msg("Failed to send slack notification for banning user")
	}

	return nil
}
