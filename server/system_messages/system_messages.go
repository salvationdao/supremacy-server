package system_messages

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SystemMessagingManager struct {
}

type SystemMessageDataType string

const (
	SystemMessageDataTypeMechQueue          SystemMessageDataType = "MECH_QUEUE"
	SystemMessageDataTypeMechBattleComplete SystemMessageDataType = "MECH_BATTLE_COMPLETE"
)

func NewSystemMessagingManager() *SystemMessagingManager {
	return &SystemMessagingManager{}
}

func (smm *SystemMessagingManager) BroadcastGlobalMessage(title string, message string, dataType *SystemMessageDataType, data *interface{}) {
	marshalled, err := json.Marshal(data)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("objectToMarshal", data).Msg("failed to marshal system message data")
		return
	}

	msg := &boiler.SystemMessage{
		DataType: null.StringFromPtr((*string)(dataType)),
		Message:  message,
		Data:     null.JSONFrom(marshalled),
	}
	err = msg.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new global system message into db")
		return
	}

	sms, err := boiler.SystemMessages(
		boiler.SystemMessageWhere.PlayerID.IsNull(),
		boiler.SystemMessageWhere.FactionID.IsNull(),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get global system messages")
		return
	}

	ws.PublishMessage("/public/system_messages", server.HubKeySystemMessageGlobalListSubscribe, &sms)
}

func (smm *SystemMessagingManager) BroadcastFactionMessage(factionID string, title string, message string, dataType *SystemMessageDataType, data *interface{}) {
	marshalled, err := json.Marshal(data)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("objectToMarshal", data).Msg("failed to marshal system message data")
		return
	}

	msg := &boiler.SystemMessage{
		FactionID: null.StringFrom(factionID),
		DataType:  null.StringFromPtr((*string)(dataType)),
		Message:   message,
		Data:      null.JSONFrom(marshalled),
	}
	err = msg.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new faction system message into db")
		return
	}

	sms, err := boiler.SystemMessages(
		boiler.SystemMessageWhere.FactionID.EQ(null.StringFrom(factionID)),
		qm.OrderBy(fmt.Sprintf("%s desc", boiler.SystemMessageColumns.SentAt)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("factionID", factionID).Msg("failed to get faction system messages")
		return
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/system_messages", factionID), server.HubKeySystemMessageFactionListSubscribe, &sms)
}

func (smm *SystemMessagingManager) BroadcastMechQueueMessage(queue []*boiler.BattleQueue) {
	for _, q := range queue {
		mech, err := q.Mech().One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("battleQueue", q).Msg("failed to find a mech associated with battle queue")
			continue
		}

		label := mech.Label
		if mech.Name != "" {
			label = mech.Name
		}

		msg := &boiler.SystemMessage{
			PlayerID: null.StringFrom(q.OwnerID),
			DataType: null.StringFrom(string(SystemMessageDataTypeMechQueue)),
			Title:    "Queue Update",
			Message:  fmt.Sprintf("Your mech, %s, is about to enter the battle arena.", label),
		}
		err = msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new system message into db")
			continue
		}

		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", q.OwnerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
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
			gamelog.L.Error().Err(err).Interface("battleQueue", q).Msg("failed to find a mech associated with battle queue")
			continue
		}

		label := mech.Label
		if mech.Name != "" {
			label = mech.Name
		}

		toMarshal := SystemMessageDataMechBattleComplete{
			MechID:     q.MechID,
			FactionWon: wonFactionID == q.FactionID,
			Briefs:     results,
		}
		data, err := json.Marshal(toMarshal)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("objectToMarshal", toMarshal).Msg("failed to marshal system message data")
			continue
		}

		msg := &boiler.SystemMessage{
			PlayerID: null.StringFrom(q.OwnerID),
			DataType: null.StringFrom(string(SystemMessageDataTypeMechBattleComplete)),
			Title:    "Battle Update",
			Message:  fmt.Sprintf("Your mech, %s, has just completed a battle in the arena.", label),
			Data:     null.JSONFrom(data),
		}
		err = msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new system message into db")
			continue
		}

		ws.PublishMessage(fmt.Sprintf("/user/%s/system_messages", q.OwnerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	}
}
