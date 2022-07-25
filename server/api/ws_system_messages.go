package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
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

	return smc
}

type SystemMessageListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	} `json:"payload"`
}

type SystemMessageListResponse struct {
	Total          int                       `json:"total"`
	SystemMessages boiler.SystemMessageSlice `json:"system_messages"`
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
		boiler.SystemMessageWhere.PlayerID.EQ(null.StringFrom(user.ID)),
	)
	total, err := boiler.SystemMessages(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get fetch system messages. Please try again later.")
	}

	queryMods = append(queryMods,
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
		qm.Limit(pageSize),
		qm.Offset(offset),
	)
	sms, err := boiler.SystemMessages(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch system messages. Please try again later.")
	}

	reply(&SystemMessageListResponse{
		Total:          int(total),
		SystemMessages: sms,
	})

	return nil
}

func (smc *SystemMessagesController) SystemMessageGlobalListSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	sms, err := boiler.SystemMessages(
		boiler.SystemMessageWhere.PlayerID.IsNull(),
		boiler.SystemMessageWhere.FactionID.IsNull(),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch global system messages. Please try again later.")
	}

	reply(&sms)

	return nil
}

func (smc *SystemMessagesController) SystemMessageFactionListSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	sms, err := boiler.SystemMessages(
		boiler.SystemMessageWhere.FactionID.EQ(null.StringFrom(factionID)),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch global system messages. Please try again later.")
	}

	reply(&sms)

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
