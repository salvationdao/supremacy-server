package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type WarMachineDestroyedEventRecord struct {
	DestroyedWarMachine *WarMachineBrief `json:"destroyed_war_machine"`
	KilledByWarMachine  *WarMachineBrief `json:"killed_by_war_machine_id,omitempty"`
	KilledBy            string           `json:"killed_by"`
}

/**********************
*  Game Notification  *
**********************/
type GameNotificationType string

const (
	GameNotificationTypeText                GameNotificationType = "TEXT"
	GameNotificationTypeLocationSelect      GameNotificationType = "LOCATION_SELECT"
	GameNotificationTypeBattleAbility       GameNotificationType = "BATTLE_ABILITY"
	GameNotificationTypeFactionAbility      GameNotificationType = "FACTION_ABILITY"
	GameNotificationTypeWarMachineAbility   GameNotificationType = "WAR_MACHINE_ABILITY"
	GameNotificationTypeWarMachineDestroyed GameNotificationType = "WAR_MACHINE_DESTROYED"
)

type GameNotificationKill struct {
	DestroyedWarMachine *server.WarMachineBrief `json:"DestroyedWarMachine"`
	KillerWarMachine    *server.WarMachineBrief `json:"killerWarMachine,omitempty"`
	KilledByAbility     *server.AbilityBrief    `json:"killedByAbility,omitempty"`
}

type LocationSelectType string

const (
	LocationSelectTypeFailedDisconnect    = "FAILED_DISCONNECT"
	LocationSelectTypeFailedTimeout       = "FAILED_TIMEOUT"
	LocationSelectTypeCancelledNoPlayer   = "CANCELLED_NO_PLAYER"
	LocationSelectTypeCancelledDisconnect = "CANCELLED_DISCONNECT"
	LocationSelectTypeTrigger             = "TRIGGER"
)

type GameNotificationLocationSelect struct {
	Type        LocationSelectType `json:"type"`
	X           *int               `json:"x,omitempty"`
	Y           *int               `json:"y,omitempty"`
	CurrentUser *UserBrief         `json:"currentUser,omitempty"`
	NextUser    *UserBrief         `json:"nextUser,omitempty"`
	Ability     *AbilityBrief      `json:"ability,omitempty"`
}

type GameNotificationAbility struct {
	User    *UserBrief    `json:"user,omitempty"`
	Ability *AbilityBrief `json:"ability,omitempty"`
}

type GameNotificationWarMachineAbility struct {
	User       *UserBrief       `json:"user,omitempty"`
	Ability    *AbilityBrief    `json:"ability,omitempty"`
	WarMachine *WarMachineBrief `json:"warMachine,omitempty"`
}

type AbilityBrief struct {
	Label    string `json:"label"`
	ImageUrl string `json:"image_url"`
	Colour   string `json:"colour"`
}

type UserBrief struct {
	ID        uuid.UUID     `json:"id"`
	Username  string        `json:"username"`
	AvatarID  *string       `json:"avatar_id,omitempty"`
	FactionID string        `json:"faction_id,omitempty"`
	Faction   *FactionBrief `json:"faction"`
}

type GameNotification struct {
	Type GameNotificationType `json:"type"`
	Data interface{}          `json:"data"`
}

const HubKeyMultiplierUpdate hub.HubCommandKey = "USER:SUPS:MULTIPLIER:SUBSCRIBE"

func (arena *Arena) HubKeyMultiplierUpdate(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err)
	}

	id, err := uuid.FromString(wsc.Identifier())
	if err != nil {
		gamelog.L.Warn().Err(err).Str("id", wsc.Identifier()).Msg("unable to create uuid from websocket client identifier id")
		return "", "", terror.Error(err, "Unable to create uuid from websocket client identifier id")
	}

	// return multiplier if battle is on
	if arena.currentBattle != nil && arena.currentBattle.multipliers != nil {
		m, total := arena.currentBattle.multipliers.PlayerMultipliers(id, -1)

		reply(&MultiplierUpdate{
			UserMultipliers:  m,
			TotalMultipliers: fmt.Sprintf("%sx", total),
		})
	}

	return req.TransactionID, messagebus.BusKey(HubKeyMultiplierUpdate), nil
}

const HubKeyViewerLiveCountUpdated = hub.HubCommandKey("VIEWER:LIVE:COUNT:UPDATED")

const HubKeyGameNotification hub.HubCommandKey = "GAME:NOTIFICATION"

func (arena *Arena) GameNotificationSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyGameNotification), nil
}

// BroadcastGameNotificationText broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationText(data string) {
	arena.messageBus.Send(messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeText,
		Data: data,
	})
}

// BroadcastGameNotificationLocationSelect broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationLocationSelect(data *GameNotificationLocationSelect) {
	arena.messageBus.Send(messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeLocationSelect,
		Data: data,
	})
}

// BroadcastGameNotificationAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationAbility(notificationType GameNotificationType, data GameNotificationAbility) {
	arena.messageBus.Send(messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: notificationType,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineAbility(data *GameNotificationWarMachineAbility) {
	arena.messageBus.Send(messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeWarMachineAbility,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineDestroyed broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineDestroyed(data *WarMachineDestroyedEventRecord) {
	arena.messageBus.Send(messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeWarMachineDestroyed,
		Data: data,
	})
}

const HubKeyPlayerBattleQueueBrowserSubscribe hub.HubCommandKey = "PLAYER:BROWSER_NOFTICATION_SUBSCRIBE"

// NotifyUpcomingWarMachines sends out notifications to users with war machines in an upcoming battle
func (arena *Arena) NotifyUpcomingWarMachines() {
	// get next 10 war machines in queue for each faction
	q, err := db.LoadBattleQueue(context.Background(), 13)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", arena.currentBattle.ID).Msg("unable to load out queue for notifications")
		return
	}

	// for each war machine in queue, find ones that need to be notified
	for _, bq := range q {
		// if in battle or already notified skip
		if bq.BattleID.Valid || bq.Notified {
			continue
		}

		// add them to users to notify
		player, err := boiler.Players(
			boiler.PlayerWhere.ID.EQ(bq.OwnerID),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("battle_id", arena.currentBattle.ID).Str("owner_id", bq.OwnerID).Msg("unable to find owner for battle queue notification")
			continue
		}

		warMachine, err := bq.Mech().One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech_id", bq.MechID).Msg("unable to find war machine for battle queue notification")
			continue
		}

		// get notifications related to mech that has not been sent
		notifications, err := boiler.BattleQueueNotifications(
			boiler.BattleQueueNotificationWhere.MechID.EQ(bq.MechID),
			boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
		).All(gamedb.StdConn)

		notificationMsg := fmt.Sprintf("%s, your War Machine %s is nearing battle, jump on to https://play.supremacy.game and prepare.", player.Username.String, warMachine.Name)

		for _, n := range notifications {
			sent := false
			// send telegram notification
			if n.TelegramNotificationID.Valid {
				err = arena.telegram.Notify(warMachine.ID, notificationMsg)
				if err != nil {
					gamelog.L.Error().Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Msg("failed to notify telegram")

				}
				sent = true
			}

			// send sms
			if n.MobileNumber.Valid {
				err := arena.sms.SendSMS(
					player.MobileNumber.String,
					notificationMsg,
				)
				if err != nil {
					gamelog.L.Error().Err(err).Str("to", player.MobileNumber.String).Msg("failed to send battle queue notification sms")
				}

				sent = true
			}

			// ??
			// if n.PushNotifications {
			// 	arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerBattleQueueBrowserSubscribe, player.ID)), notificationMsg)
			// }

			// if player.R.PlayerPreference.NotificationsBattleQueuePushNotifications {
			// 	// TODO: app notifications?
			// }

			// TODO: discord notifications?

			// delete notification after sent
			if sent {
				now := time.Now()
				n.SentAt = null.TimeFrom(now)
				_, err := n.Delete(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Err(err).Str("to", player.MobileNumber.String).Msg("failed to update notification")
				}
			}
		}

		bq.Notified = true
		_, err = bq.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueColumns.Notified))
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Msg("failed to update notified column")
		}
	}
}
