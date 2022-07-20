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

type SystemMessageType string

const (
	SystemMessageMechQueue          SystemMessageType = "MECH_QUEUE"
	SystemMessageMechBattleComplete SystemMessageType = "MECH_BATTLE_COMPLETE"
)

type SystemMessage struct {
	Type    SystemMessageType `json:"type"`
	Message string            `json:"message"`
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
			Type:    SystemMessageMechQueue,
			Message: fmt.Sprintf("Your mech, %s, is about to enter the battle arena.", mech.Label),
		})
	}
}

func (smm *SystemMessagingManager) BroadcastMechBattleCompleteMessage(queue []*boiler.BattleQueue) {
	for _, q := range queue {
		mech, err := q.Mech().One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Debug().Interface("battleQueue", q).Msg("failed to find a mech associated with battle queue")
			continue
		}
		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", q.OwnerID), server.HubKeySystemMessageSubscribe, &SystemMessage{
			Type:    SystemMessageMechBattleComplete,
			Message: fmt.Sprintf("Your mech, %s, has just completed a battle in the arena.", mech.Label),
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
