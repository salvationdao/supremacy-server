package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/ws"
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

	api.SecureUserCommand(server.HubKeySystemMessageList, smc.SystemMessageList)
	api.SecureUserCommand(server.HubKeySystemMessageDismiss, smc.SystemMessageDismiss)

	return smc
}

func (smc *SystemMessagesController) SystemMessageList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	sms, err := boiler.SystemMessages(
		boiler.SystemMessageWhere.PlayerID.EQ(user.ID),
		boiler.SystemMessageWhere.IsDismissed.EQ(false),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch system messages. Please try again later.")
	}

	reply(sms)

	return nil
}

type SystemMessageDismissRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

func (smc *SystemMessagesController) SystemMessageDismiss(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
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

	if sm.IsDismissed {
		return nil
	}

	sm.IsDismissed = true
	_, err = sm.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to dismiss system message. Please try again later.")
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", user.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)

	return nil
}
