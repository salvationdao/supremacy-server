package system_messages

import (
	"encoding/json"
	"fmt"
	"html"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type SystemMessagingManager struct {
}

type SystemMessageDataType string

const (
	SystemMessageDataTypeMechBattleBegin       SystemMessageDataType = "MECH_BATTLE_BEGIN"
	SystemMessageDataTypeMechBattleComplete    SystemMessageDataType = "MECH_BATTLE_COMPLETE"
	SystemMessageDataTypeMechOwnerBattleReward SystemMessageDataType = "MECH_OWNER_BATTLE_REWARD"
	SystemMessageDataTypePlayerAbilityRefunded SystemMessageDataType = "PLAYER_ABILITY_REFUNDED"
	SystemMessageDataTypeGlobal                SystemMessageDataType = "GLOBAL"
	SystemMessageDataTypeFaction               SystemMessageDataType = "FACTION"
	SystemMessageDataTypeExpiredBattleLobby    SystemMessageDataType = "EXPIRED_BATTLE_LOBBY"
	SystemMessageDataTypeBattleLobbyInvitation SystemMessageDataType = "BATTLE_LOBBY_INVITATION"
)

var bm = bluemonday.StrictPolicy()

func BroadcastGlobalSystemMessage(title string, message string, dataType SystemMessageDataType, data *interface{}) error {
	l := gamelog.L.With().Str("func", "BroadcastGlobalSystemMessage").Logger()

	players, err := boiler.Players().All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to get players from db")
		return err
	}
	l = l.With().Interface("players", players).Logger()

	sanitisedTitle := html.UnescapeString(bm.Sanitize(title))
	sanitisedMsg := html.UnescapeString(bm.Sanitize(message))
	template := &boiler.SystemMessage{
		SenderID: server.SupremacySystemAdminUserID,
		Title:    sanitisedTitle,
		Message:  sanitisedMsg,
	}

	if dataType != "" {
		template.DataType = null.StringFrom(string(dataType))
	}

	if data != nil {
		marshalled, err := json.Marshal(data)
		if err != nil {
			l.Error().Err(err).Interface("objectToMarshal", data).Msg("failed to marshal global system message data")
			return err
		}
		template.Data = null.JSONFrom(marshalled)
	}
	l = l.With().Interface("templateMsg", template).Logger()

	for _, p := range players {
		msg := &boiler.SystemMessage{
			PlayerID: p.ID,
			SenderID: template.SenderID,
			Title:    template.Title,
			Message:  template.Message,
			Data:     template.Data,
			DataType: template.DataType,
		}
		err := msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new global system message into db")
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", p.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	}
	return nil
}

func BroadcastFactionSystemMessage(factionID string, title string, message string, dataType SystemMessageDataType, data *interface{}) error {
	l := gamelog.L.With().Str("func", "BroadcastGlobalSystemMessage").Logger()

	players, err := boiler.Players(boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID))).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to get players from db")
		return err
	}
	l = l.With().Interface("players", players).Logger()

	sender, err := boiler.Players(
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID)),
		boiler.PlayerWhere.ID.IN([]string{server.RedMountainPlayerID, server.BostonCyberneticsPlayerID, server.ZaibatsuPlayerID}),
	).One(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Str("factionID", factionID).Msg("failed to get faction user from faction ID")
		return err
	}

	sanitisedTitle := html.UnescapeString(bm.Sanitize(title))
	sanitisedMsg := html.UnescapeString(bm.Sanitize(message))
	template := &boiler.SystemMessage{
		SenderID: sender.ID,
		Title:    sanitisedTitle,
		Message:  sanitisedMsg,
	}

	if dataType != "" {
		template.DataType = null.StringFrom(string(dataType))
	}

	if data != nil {
		marshalled, err := json.Marshal(data)
		if err != nil {
			l.Error().Err(err).Interface("objectToMarshal", data).Msg("failed to marshal faction system message data")
			return err
		}
		template.Data = null.JSONFrom(marshalled)
	}
	l = l.With().Interface("templateMsg", template).Logger()

	for _, p := range players {
		if p.FactionID.String != factionID {
			continue
		}

		msg := &boiler.SystemMessage{
			PlayerID: p.ID,
			SenderID: template.SenderID,
			Title:    template.Title,
			Message:  template.Message,
			Data:     template.Data,
			DataType: template.DataType,
		}
		err := msg.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Interface("newSystemMessage", msg).Msg("failed to insert new global system message into db")
			return err
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", p.ID), server.HubKeySystemMessageListUpdatedSubscribe, true)
	}

	return nil
}
