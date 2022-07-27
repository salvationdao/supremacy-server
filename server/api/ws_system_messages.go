package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/system_messages"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SystemMessagesController struct {
	API *API
}

func NewSystemMessagesController(api *API) *SystemMessagesController {
	smc := &SystemMessagesController{
		api,
	}

	api.SecureUserCommand(server.HubKeySystemMessageList, smc.SystemMessageListHandler)
	api.SecureUserCommand(server.HubKeySystemMessageDismiss, smc.SystemMessageDismissHandler)
	api.SecureUserCommand(server.HubKeySystemMessageSend, smc.SystemMessageSendHandler)

	return smc
}

type SystemMessageListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Page     int  `json:"page"`
		PageSize int  `json:"page_size"`
		HideRead bool `json:"hide_read"`
	} `json:"payload"`
}

type SystemMessageDetailed struct {
	*boiler.SystemMessage
	Sender boiler.Player `json:"sender"`
}

type SystemMessageListResponse struct {
	Total          int                      `json:"total"`
	TotalUnread    int                      `json:"total_unread"`
	SystemMessages []*SystemMessageDetailed `json:"system_messages"`
}

func (smc *SystemMessagesController) SystemMessageListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &SystemMessageListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	pageSize := 10
	if req.Payload.PageSize > 0 {
		pageSize = req.Payload.PageSize
	}

	if req.Payload.Page > 0 {
		offset = req.Payload.Page * pageSize
	}

	queryMods := []qm.QueryMod{}
	queryMods = append(queryMods,
		boiler.SystemMessageWhere.PlayerID.EQ(user.ID),
	)

	if req.Payload.HideRead {
		queryMods = append(queryMods,
			boiler.SystemMessageWhere.ReadAt.IsNull(),
		)
	}

	total, err := boiler.SystemMessages(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch system messages. Please try again later.")
	}

	unreadQueryMods := []qm.QueryMod{}
	unreadQueryMods = append(unreadQueryMods, queryMods...)
	unreadQueryMods = append(unreadQueryMods, boiler.SystemMessageWhere.ReadAt.IsNull())
	totalUnread, err := boiler.SystemMessages(unreadQueryMods...).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch system messages. Please try again later.")
	}

	queryMods = append(queryMods,
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
		qm.Limit(pageSize),
		qm.Offset(offset),
		qm.Load(boiler.SystemMessageRels.Sender),
	)
	sms, err := boiler.SystemMessages(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch system messages. Please try again later.")
	}

	detailedSms := []*SystemMessageDetailed{}
	for _, s := range sms {
		detailedSms = append(detailedSms, &SystemMessageDetailed{
			SystemMessage: s,
			Sender:        *s.R.Sender,
		})
	}

	reply(&SystemMessageListResponse{
		Total:          int(total),
		TotalUnread:    int(totalUnread),
		SystemMessages: detailedSms,
	})

	return nil
}

type SystemMessageDismissRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

func (smc *SystemMessagesController) SystemMessageDismissHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &SystemMessageDismissRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.ID == "" {
		return terror.Error(terror.ErrInvalidInput)
	}

	sm, err := boiler.FindSystemMessage(gamedb.StdConn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, "Failed to dismiss system message. Please try again later.")
	}

	if sm.ReadAt.Valid {
		return nil
	}

	sm.ReadAt = null.TimeFrom(time.Now())
	_, err = sm.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to dismiss system message. Please try again later.")
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", user.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)

	return nil
}

type SystemMessageSendRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Type    system_messages.SystemMessageDataType `json:"type"` // GLOBAL or FACTION
		Subject string                                `json:"subject"`
		Message string                                `json:"message"`
	} `json:"payload"`
}

func (smc *SystemMessagesController) SystemMessageSendHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &SystemMessageSendRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	hasPermission, err := boiler.PlayersFeatures(
		boiler.PlayersFeatureWhere.PlayerID.EQ(user.ID),
		boiler.PlayersFeatureWhere.FeatureName.EQ(boiler.FeatureNameSYSTEM_MESSAGES),
	).Exists(gamedb.StdConn)
	if err != nil {
		return terror.Error(err)
	}

	if !hasPermission {
		return terror.Error(terror.ErrUnauthorised)
	}

	if req.Payload.Type == "" {
		return terror.Error(terror.ErrInvalidInput)
	}

	if req.Payload.Subject == "" {
		return terror.Error(terror.ErrInvalidInput, "Subject cannot be empty.")
	}

	if req.Payload.Message == "" {
		return terror.Error(terror.ErrInvalidInput, "Message cannot be empty.")
	}

	recipients := []*boiler.Player{}
	template := &boiler.SystemMessage{
		Title:    req.Payload.Subject,
		Message:  req.Payload.Message,
		DataType: null.StringFrom(string(req.Payload.Type)),
	}

	if req.Payload.Type == system_messages.SystemMessageDataTypeGlobal {
		template.SenderID = server.SupremacySystemAdminUserID

		players, err := boiler.Players().All(gamedb.StdConn)
		if err != nil {
			return err
		}
		recipients = players
	} else if req.Payload.Type == system_messages.SystemMessageDataTypeFaction {
		sender, err := boiler.Players(
			boiler.PlayerWhere.FactionID.EQ(null.StringFrom(user.FactionID.String)),
			boiler.PlayerWhere.ID.IN([]string{server.RedMountainPlayerID, server.BostonCyberneticsPlayerID, server.ZaibatsuPlayerID}),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("player", user).Msg("failed to get faction user from faction ID")
			return err
		}
		template.SenderID = sender.ID

		players, err := boiler.Players(boiler.PlayerWhere.FactionID.EQ(null.StringFrom(user.FactionID.String))).All(gamedb.StdConn)
		if err != nil {
			return err
		}
		recipients = players
	}

	for _, p := range recipients {
		if req.Payload.Type == system_messages.SystemMessageDataTypeFaction && p.FactionID.String != user.FactionID.String {
			continue
		}

		msg := &boiler.SystemMessage{}
		msg.PlayerID = p.ID
		msg.SenderID = template.SenderID
		msg.Title = template.Title
		msg.Message = template.Message
		msg.DataType = template.DataType
		err := msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new system message into db")
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", p.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	}

	reply(true)

	return nil
}
