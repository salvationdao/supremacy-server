package battle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/system_messages"
	"server/xsyn_rpcclient"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type WarMachineDestroyedEventRecord struct {
	DestroyedWarMachine *WarMachineBrief `json:"destroyed_war_machine"`
	KilledByWarMachine  *WarMachineBrief `json:"killed_by_war_machine,omitempty"`
	KilledByUser        *UserBrief       `json:"killed_by_user,omitempty"`
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
	GameNotificationTypeBattleZoneChange    GameNotificationType = "BATTLE_ZONE_CHANGE"
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
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FactionID string    `json:"faction_id,omitempty"`
	Faction   *Faction  `json:"faction"`
	Gid       int       `json:"gid"`
}

type GameNotification struct {
	Type GameNotificationType `json:"type"`
	Data interface{}          `json:"data"`
}

const (
	MechCommandActionFired    = "MECH_COMMAND_FIRED"
	MechCommandActionCancel   = "MECH_COMMAND_CANCEL"
	MechCommandActionComplete = "MECH_COMMAND_COMPLETE"
)

type MechCommandNotification struct {
	MechID       string     `json:"mech_id"`
	MechLabel    string     `json:"mech_label"`
	MechImageUrl string     `json:"mech_image_url"`
	FactionID    string     `json:"faction_id"`
	Action       string     `json:"action"`
	FiredByUser  *UserBrief `json:"fired_by_user,omitempty"`
}

const HubKeyGameNotification = "GAME:NOTIFICATION"

// BroadcastGameNotificationText broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationText(data string) {
	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeText,
		Data: data,
	})
}

// BroadcastGameNotificationLocationSelect broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationLocationSelect(data *GameNotificationLocationSelect) {
	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeLocationSelect,
		Data: data,
	})
}

// BroadcastGameNotificationAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationAbility(notificationType GameNotificationType, data GameNotificationAbility) {
	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
		Type: notificationType,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineAbility(data *GameNotificationWarMachineAbility) {
	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeWarMachineAbility,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineDestroyed broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineDestroyed(data *WarMachineDestroyedEventRecord) {
	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeWarMachineDestroyed,
		Data: data,
	})
}

// BroadcastGameNotificationBattleZoneChange broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationBattleZoneChange(data *ZoneChangeEvent) {
	ws.PublishMessage("/public/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeBattleZoneChange,
		Data: data,
	})
}

// NotifyUpcomingWarMachines sends out notifications to users with war machines in an upcoming battle
func (arena *Arena) NotifyUpcomingWarMachines() {
	// get next 10 war machines in queue for each faction
	q, err := db.LoadBattleQueue(context.Background(), 13)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("battle_id", arena.CurrentBattle().ID).Msg("unable to load out queue for notifications")
		return
	}

	// broadcast system message to mech owners
	system_messages.BroadcastMechQueueMessage(q)

	// for each war machine in queue, find ones that need to be notified
	for _, bq := range q {
		// if in battle or already notified skip
		if bq.BattleID.Valid {
			continue
		}
		if bq.Notified {
			continue
		}

		// get mech owner
		player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(bq.OwnerID)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("battle_id", arena.CurrentBattle().ID).Str("owner_id", bq.OwnerID).Msg("unable to find owner for battle queue notification")
			continue
		}

		playerUUID, err := uuid.FromString(player.ID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("player_id", bq.MechID).Msg("unable to get player UUID")
			continue
		}

		warMachine, err := bq.Mech(
			qm.Load(boiler.MechRels.BattleQueueNotifications),
			qm.Load(boiler.MechRels.Blueprint),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("mech_id", bq.MechID).Msg("unable to find war machine for battle queue notification")
			continue
		}
		wmName := ""
		if warMachine.R != nil && warMachine.R.Blueprint != nil {
			wmName = fmt.Sprintf("(%s)", warMachine.R.Blueprint.Label)
		}
		if warMachine.Name != "" {
			wmName = fmt.Sprintf("(%s)", warMachine.Name)
		}

		// track wether notification was sent with the old system
		sent := false

		// OLD NOTIFICATION SYSTEM (WILL BE REMOVED)
		if warMachine.R.BattleQueueNotifications != nil {
			for _, n := range warMachine.R.BattleQueueNotifications {
				if n.SentAt.Valid {
					continue
				}

				// send telegram notification
				if n.TelegramNotificationID.Valid {
					notificationMsg := fmt.Sprintf("🦾 %s, your War Machine %s is approaching the front of the queue!\n\n⚔️ Jump into the Battle Arena now to prepare. Your survival has its rewards.\n\n⚠️ (Reminder: In order to combat scams we will NEVER send you links)", player.Username.String, wmName)
					gamelog.L.Info().Str("TelegramNotificationID", n.TelegramNotificationID.String).Msg("sending telegram notification")
					err = arena.telegram.NotifyDEPRECATED(n.TelegramNotificationID.String, notificationMsg)
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Str("telegram id", n.TelegramNotificationID.String).Msg("failed to notify telegram")
					}
					sent = true
				}

				// send sms
				if n.MobileNumber.Valid {
					notificationMsg := fmt.Sprintf("%s, your War Machine %s is approaching the front of the queue!\n\nJump into the Battle Arena now to prepare. Your survival has its rewards.\n\n(Reminder: In order to combat scams we will NEVER send you links)", player.Username.String, wmName)
					gamelog.L.Info().Str("MobileNumber", n.MobileNumber.String).Msg("sending sms notification")
					err := arena.sms.SendSMS(
						player.MobileNumber.String,
						notificationMsg,
					)
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("to", n.MobileNumber.String).Msg("failed to send battle queue notification sms")
					}
					sent = true
				}

				n.SentAt = null.TimeFrom(time.Now())
				n.QueueMechID = null.NewString("", false)
				_, err = n.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("bqn id", n.ID).Msg("failed to update BattleQueueNotificationColumns")
				}
			}
		}

		// if notification already sent with old notification system dont send with new system
		if sent {
			bq.Notified = true
			_, err = bq.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueColumns.Notified))
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Msg("failed to update notified column")
			}
			continue
		}

		// get player preferences
		prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(player.ID)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("player_id", player.ID).Msg("unable to get player preferences")
			continue
		}

		if prefs == nil {
			continue
		}

		// if user's player preferences has telegram or sms notifications enabled
		notificationsEnabled := (prefs.EnableSMSNotifications && prefs.MobileNumber.Valid) ||
			(prefs.EnableTelegramNotifications && prefs.TelegramID.Valid)

		if !notificationsEnabled {
			continue
		}

		// get faction account
		factionAccountID, ok := server.FactionUsers[player.FactionID.String]
		if !ok {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("mech ID", bq.MechID).
				Str("faction ID", player.FactionID.String).
				Err(err).
				Msg("unable to get hard coded syndicate player ID from faction ID")
		}

		// charge for notification
		notifyTransactionID, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			Amount:               "5000000000000000000", // 5 sups
			FromUserID:           playerUUID,
			ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
			TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queue_notification_fee|%s|%d", warMachine.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupBattle),
			SubGroup:             "Queue",
			Description:          "Notification surcharge for queued mech in arena",
		})
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("txID", notifyTransactionID).Err(err).Msg("unable to charge user for sms/telegram notification for mech in queue")
			_, err = arena.RPCClient.RefundSupsMessage(notifyTransactionID)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("txID", notifyTransactionID).Err(err).Msg("failed to refund notification queue fee")
			}
		}
		bq.QueueNotificationFeeTXID = null.StringFrom(notifyTransactionID)
		_, err = bq.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueColumns.QueueNotificationFeeTXID))
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").
				Str("tx_id", notifyTransactionID).
				Err(err).Msg("unable to update battle queue with queue notification transaction id")
			_, err = arena.RPCClient.RefundSupsMessage(notifyTransactionID)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Str("txID", notifyTransactionID).Err(err).Msg("failed to refund queue notification fee")
			}

		}

		// sms notifications
		if prefs.EnableSMSNotifications && prefs.MobileNumber.Valid {
			notificationMsg := fmt.Sprintf("%s, your War Machine %s is approaching the front of the queue!\n\nJump into the Battle Arena now to prepare. Your survival has its rewards.\n\n(Reminder: In order to combat scams we will NEVER send you links)", player.Username.String, wmName)
			gamelog.L.Info().Str("MobileNumber", prefs.MobileNumber.String).Msg("sending sms notification")
			err := arena.sms.SendSMS(
				prefs.MobileNumber.String,
				notificationMsg,
			)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("to", prefs.MobileNumber.String).Msg("failed to send battle queue notification sms")

				// refund notification fee
				_, err = arena.RPCClient.RefundSupsMessage(notifyTransactionID)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("txID", notifyTransactionID).Err(err).Msg("failed to refund notification queue fee")
				}
			}
		}

		// telegram notifications
		if prefs.EnableTelegramNotifications && prefs.TelegramID.Valid {
			notificationMsg := fmt.Sprintf("🦾 %s, your War Machine %s is approaching the front of the queue!\n\n⚔️ Jump into the Battle Arena now to prepare. Your survival has its rewards.\n\n⚠️ (Reminder: In order to combat scams we will NEVER send you links)", player.Username.String, wmName)
			gamelog.L.Info().Str("player_id", player.ID).Msg("sending telegram notification")
			err = arena.telegram.Notify(prefs.TelegramID.Int64, notificationMsg)
			if err != nil {
				gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Str("telegram id", fmt.Sprintf("%v", prefs.TelegramID)).Msg("failed to send telegram notification")
				_, err = arena.RPCClient.RefundSupsMessage(notifyTransactionID)
				if err != nil {
					gamelog.L.Error().Str("log_name", "battle arena").Str("txID", notifyTransactionID).Err(err).Msg("failed to refund notification queue fee")
				}
			}
		}

		//TODO: push notifications
		// TODO: discord notifications?

		bq.Notified = true
		_, err = bq.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueColumns.Notified))
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Msg("failed to update notified column")
		}

	}
}
