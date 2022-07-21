package system_messages

import (
	"context"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	Data    interface{}       `json:"data,omitempty"`
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

		label := mech.Label
		if mech.Name != "" {
			label = mech.Name
		}
		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", q.OwnerID), server.HubKeySystemMessageSubscribe, &SystemMessage{
			Type:    SystemMessageMechQueue,
			Message: fmt.Sprintf("Your mech, %s, is about to enter the battle arena.", label),
		})
	}
}

type SystemMessageDataMechBattleComplete struct {
	MechID     string             `json:"mech_id"`
	FactionWon bool               `json:"faction_won"`
	Briefs     []*MechBattleBrief `json:"briefs"`
}

type MechBattleBrief struct {
	MechID     string    `boiler:"mech_id" json:"mech_id"`
	FactionID  string    `boiler:"faction_id" json:"faction_id"`
	FactionWon bool      `boiler:"faction_won" json:"faction_won"`
	Kills      int       `boiler:"kills" json:"kills"`
	Killed     null.Time `boiler:"killed" json:"killed,omitempty"`
	Label      string    `boiler:"label" json:"label"`
	Name       string    `boiler:"name" json:"name"`
}

func (smm *SystemMessagingManager) BroadcastMechBattleCompleteMessage(queue []*boiler.BattleQueue, battleID string) {
	query := fmt.Sprintf(`
	select 
		bm.mech_id,
		bm.faction_id,
		bm.faction_won,
		bm.kills,
		bm.killed,
		m."label",
		m."name"
	from battle_mechs bm 
	inner join mechs m on m.id = bm.mech_id
	where battle_id = $1;
`)
	results := []*MechBattleBrief{}
	err := boiler.NewQuery(qm.SQL(query, battleID)).Bind(context.Background(), gamedb.StdConn, &results)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battleID", battleID).Msg("failed to create mech battle brief from battle id")
		return
	}

	wonFactionID := ""
	for _, r := range results {
		if r.FactionWon {
			wonFactionID = r.FactionID
			break
		}
	}

	for _, q := range queue {
		mech, err := q.Mech().One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Debug().Interface("battleQueue", q).Msg("failed to find a mech associated with battle queue")
			continue
		}

		label := mech.Label
		if mech.Name != "" {
			label = mech.Name
		}
		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", q.OwnerID), server.HubKeySystemMessageSubscribe, &SystemMessage{
			Type:    SystemMessageMechBattleComplete,
			Message: fmt.Sprintf("Your mech, %s, has just completed a battle in the arena.", label),
			Data: &SystemMessageDataMechBattleComplete{
				MechID:     q.MechID,
				FactionWon: wonFactionID == q.FactionID,
				Briefs:     results,
			},
		})
	}
}
