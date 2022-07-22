package api

import (
	"context"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
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

	return smc
}

func (smc *SystemMessagesController) SystemMessageList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	sms, err := boiler.SystemMessages(
		boiler.SystemMessageWhere.PlayerID.EQ(user.ID),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to fetch system messages. Please try again later.")
	}

	reply(sms)

	return nil
}
