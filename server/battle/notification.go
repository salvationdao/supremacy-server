package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/multipliers"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
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
	ID        uuid.UUID       `json:"id"`
	Username  string          `json:"username"`
	FactionID string          `json:"faction_id,omitempty"`
	Faction   *Faction `json:"faction"`
	Gid       int             `json:"gid"`
}

type GameNotification struct {
	Type GameNotificationType `json:"type"`
	Data interface{}          `json:"data"`
}

const HubKeyMultiplierSubscribe = "USER:MULTIPLIERS:SUBSCRIBE"

const HubKeyUserMultiplierSignalUpdate = "USER:MULTIPLIER:SIGNAL:SUBSCRIBE"

func (arena *Arena) MultiplierUpdate(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	id, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Warn().Err(err).Str("id", user.ID).Msg("unable to create uuid from websocket client identifier id")
		return terror.Error(err, "Unable to create uuid from websocket client identifier id")
	}

	spoils, err := boiler.SpoilsOfWars(
		boiler.SpoilsOfWarWhere.CreatedAt.GT(time.Now().AddDate(0, 0, -1)),
		boiler.SpoilsOfWarWhere.LeftoversTransactionID.IsNull(),
		qm.And("amount > amount_sent"),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to call SpoilsOfWars")
		return terror.Error(err, "Unable to get recently battle multipliers.")
	}

	resp := &MultiplierUpdate{
		Battles: []*MultiplierUpdateBattles{},
	}

	for _, spoil := range spoils {
		m, total, _ := multipliers.GetPlayerMultipliersForBattle(id.String(), spoil.BattleNumber)
		resp.Battles = append(resp.Battles, &MultiplierUpdateBattles{
			BattleNumber:     spoil.BattleNumber,
			TotalMultipliers: multipliers.FriendlyFormatMultiplier(total),
			UserMultipliers:  m,
		})
	}

	reply(resp)
	return nil
}

const HubKeyViewerLiveCountUpdated = "VIEWER:LIVE:COUNT:UPDATED"

const HubKeyGameNotification = "GAME:NOTIFICATION"

func (arena *Arena) GameNotificationSubscribeHandler(ctx context.Context, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", err
	}

	return req.TransactionID, messagebus.BusKey(HubKeyGameNotification), nil
}

// BroadcastGameNotificationText broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationText(data string) {
	ws.PublishMessage("/battle/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeText,
		Data: data,
	})
}

// BroadcastGameNotificationLocationSelect broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationLocationSelect(data *GameNotificationLocationSelect) {
	ws.PublishMessage("/battle/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeLocationSelect,
		Data: data,
	})
}

// BroadcastGameNotificationAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationAbility(notificationType GameNotificationType, data GameNotificationAbility) {
	ws.PublishMessage("/battle/notification", HubKeyGameNotification, &GameNotification{
		Type: notificationType,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineAbility(data *GameNotificationWarMachineAbility) {
	ws.PublishMessage("/battle/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeWarMachineAbility,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineDestroyed broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineDestroyed(data *WarMachineDestroyedEventRecord) {
	ws.PublishMessage("/battle/notification", HubKeyGameNotification, &GameNotification{
		Type: GameNotificationTypeWarMachineDestroyed,
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

	// for each war machine in queue, find ones that need to be notified
	for _, bq := range q {
		// if in battle or already notified skip
		if bq.BattleID.Valid {
			gamelog.L.Warn().Err(err).Str("battle_id", arena.CurrentBattle().BattleID).Msg(fmt.Sprintf("battle has started or already happened before sending notification: %s", bq.BattleID.String))
			continue
		}
		if bq.Notified {
			continue
		}

		// add them to users to notify
		player, err := boiler.Players(
			boiler.PlayerWhere.ID.EQ(bq.OwnerID),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("battle_id", arena.CurrentBattle().ID).Str("owner_id", bq.OwnerID).Msg("unable to find owner for battle queue notification")
			continue
		}
		warMachine, err := bq.Mech(qm.Load(boiler.MechRels.BattleQueueNotifications)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech_id", bq.MechID).Msg("unable to find war machine for battle queue notification")
			continue
		}

		// continue loop if their war machine does not have a relationship with the battle_queue_notifications table
		if warMachine.R.BattleQueueNotifications == nil {
			gamelog.L.Warn().Str("mech id", warMachine.ID).Str("mech name", warMachine.Name).Msg("Skipping mech notification, no relation found on battle_queue_notifications table")
			continue
		}

		//bqn, err := bq.QueueMechBattleQueueNotifications(
		//	boiler.BattleQueueNotificationWhere.QueueMechID.EQ(null.StringFrom(warMachine.ID)),
		//	boiler.BattleQueueNotificationWhere.IsRefunded.EQ(false),
		//	boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
		//	qm.Load(boiler.BattleQueueNotificationRels.Mech),
		//).One(gamedb.StdConn)
		//if err != nil {
		//	gamelog.L.Error().Err(err).Str("battle_id", arena.CurrentBattle().ID).Msg("unable to find battle queue notifications")
		//	continue
		//}

		wmName := fmt.Sprintf("(%s)", warMachine.Label)
		if warMachine.Name != "" {
			wmName = fmt.Sprintf("(%s)", warMachine.Name)
		}

		for _, n := range warMachine.R.BattleQueueNotifications {
			if n.SentAt.Valid {
				continue
			}
			// send telegram notification
			if n.TelegramNotificationID.Valid {

				notificationMsg := fmt.Sprintf("🦾 %s, your War Machine %s is approaching the front of the queue!\n\n⚔️ Jump into the Battle Arena now to prepare. Your survival has its rewards.\n\n⚠️ (Reminder: In order to combat scams we will NEVER send you links)", player.Username.String, wmName)

				gamelog.L.Info().Str("TelegramNotificationID", n.TelegramNotificationID.String).Msg("sending telegram notification")
				err = arena.telegram.Notify(n.TelegramNotificationID.String, notificationMsg)
				if err != nil {
					gamelog.L.Error().Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Str("telegram id", n.TelegramNotificationID.String).Msg("failed to notify telegram")
				}
			}

			// send sms
			if n.MobileNumber.Valid {
				notificationMsg := fmt.Sprintf("%s, your War Machine %s is approaching the front of the queue!\n\nJump into the Battle Arena now to prepare. Your survival has its rewards.\n\n(Reminder: In order to combat scams we will NEVER send you links)", player.Username.String, wmName)
				gamelog.L.Info().Str("MobileNumber", n.MobileNumber.String).Msg("sending sms notification")
				err := arena.sms.SendSMS(
					n.MobileNumber.String,
					notificationMsg,
				)
				if err != nil {
					gamelog.L.Error().Err(err).Str("to", n.MobileNumber.String).Msg("failed to send battle queue notification sms")
				}
			}

			n.SentAt = null.TimeFrom(time.Now())
			n.QueueMechID = null.NewString("", false)
			_, err = n.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Str("bqn id", n.ID).Msg("failed to update BattleQueueNotificationColumns")
			}
		}

		//TODO: push notifications
		// TODO: discord notifications?

		bq.Notified = true

		_, err = bq.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueColumns.Notified))
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech_id", bq.MechID).Str("owner_id", bq.OwnerID).Str("queued_at", bq.QueuedAt.String()).Msg("failed to update notified column")
		}

	}
}
