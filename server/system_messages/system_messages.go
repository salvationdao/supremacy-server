package system_messages

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-syndicate/ws"
)

type SystemMessagingManager struct {
}

type SystemMessage struct {
	Message string `json:"message"`
}

func NewSystemMessagingManager() *SystemMessagingManager {
	return &SystemMessagingManager{}
}

func (smm *SystemMessagingManager) BroadcastMechQueueMessage(queue []*boiler.BattleQueue) {
	for _, q := range queue {
		mech, err := q.Mech().One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Debug().Interface("battleQueue", q).Msg("failed to find a mech associated with battle queue")
			continue
		}
		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", q.OwnerID), server.HubKeySystemMessageSubscribe, &SystemMessage{
			Message: fmt.Sprintf("Your mech, %s is about to enter the battle arena.", mech.Label),
		})
	}
}

// func (smm *SystemMessagingManager) Broadcast

// func () BroadcastGameNotificationText(data string) {
// 	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
// 		Type: GameNotificationTypeText,
// 		Data: data,
// 	})
// }
