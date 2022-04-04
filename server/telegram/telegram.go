package telegram

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	tele "gopkg.in/telebot.v3"
)

func genCode() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	n := 5
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type Telegram struct {
	*tele.Bot
	MessageBus *messagebus.MessageBus
}

// NewTelegram
func NewTelegram(token string, messageBus *messagebus.MessageBus) (*Telegram, error) {
	t := &Telegram{
		MessageBus: messageBus,
	}
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable initialise telegram bot")
		return nil, terror.Error(err)
	}

	t.Bot = b

	return t, nil
}

var telegramNotifications = map[string][]string{}

const HubKeyTelegramShortcodeRegistered = "USER:TELEGRAM_SHORTCODE_REGISTERED"

func (t *Telegram) RunTelegram(bot *tele.Bot) error {

	bot.Handle("/register", func(c tele.Context) error {
		return c.Send("Enter shortcode", tele.ForceReply)
	})

	// handle user reply
	bot.Handle(tele.OnText, func(c tele.Context) error {
		if !c.Message().IsReply() {
			return nil
		}

		shortcode := c.Text()
		telegramID := c.Recipient().Recipient()
		_telegramID, err := strconv.Atoi(telegramID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable convert telegramID to int")
			return terror.Error(err)
		}

		// get notification via shortcode
		notification, err := boiler.BattleQueueNotifications(
			boiler.BattleQueueNotificationWhere.IsRefunded.EQ(false),
			boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
			qm.InnerJoin("telegram_notifications tn on tn.id = battle_queue_notifications.telegram_notification_id"),
			qm.Where("tn.registered = false"),
			qm.Where("tn.shortcode = ?", shortcode),

			// load mech, telegramNotification rels
			qm.Load(boiler.BattleQueueNotificationRels.TelegramNotification),
			qm.Load(boiler.BattleQueueNotificationRels.Mech)).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("unable to get notification by shortcode")
			return c.Send("unable to get notification by shortcode")
		}

		reply := ""

		// cant find telgram notification by shortcode
		if errors.Is(err, sql.ErrNoRows) || notification.R == nil || notification.R.TelegramNotification == nil {
			reply = "invalid shortcode!"
			return c.Send(reply)
		}

		telegramNotification := notification.R.TelegramNotification

		// register that notification
		telegramNotification.Registered = true
		telegramNotification.TelegramID = null.IntFrom(_telegramID)

		// update notification
		_, err = telegramNotification.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("telegramID", telegramID).
				Str("notificationID", telegramNotification.ID).
				Msg("unable to update telegram notification")
			return terror.Error(err)
		}
		wmName := notification.R.Mech.Label
		wmOwner := notification.R.Mech.OwnerID

		if err != nil {
			reply = "invalid shortcode!"
			go t.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTelegramShortcodeRegistered, wmOwner)), false)
		} else {

			if notification.R.Mech.Name != "" {
				wmName = notification.R.Mech.Name
			}
			reply = fmt.Sprintf("Shortcode registered! you will be notified when your war machine (%s) is nearing battle", wmName)
			go t.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTelegramShortcodeRegistered, wmOwner)), true)

		}

		return c.Send(reply)
	})
	bot.Start()
	return nil
}

func (t *Telegram) NotificationCreate(mechID string, notification *boiler.BattleQueueNotification) (*boiler.TelegramNotification, error) {

	code := genCode()
	codeExists := true
	if codeExists {
		// check if a notification that hasnt been sent/ not refunded has that short code
		exists, err := boiler.BattleQueueNotifications(
			boiler.BattleQueueNotificationWhere.IsRefunded.EQ(false),
			boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
			qm.InnerJoin("telegram_notifications tn on tn.id = battle_queue_notifications.telegram_notification_id"),
			qm.Where("tn.shortcode = ?"),
			qm.Where("tn.Registered = false", code)).Exists(gamedb.StdConn)
		if err != nil {
			return nil, terror.Error(err, "Unable to get telegram notifications")
		}

		if exists {
			// if code already exist generate new one
			code = genCode()
		} else {
			codeExists = false
		}
	}

	telegramNotification := &boiler.TelegramNotification{
		Shortcode: code,
	}

	err := telegramNotification.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err, "Unable to link notification to telegram notification")
	}

	return telegramNotification, nil
}

func (t *Telegram) Notify(id string, message string) error {
	// get telegram notification
	notification, err := boiler.FindTelegramNotification(gamedb.StdConn, id)
	if err != nil {
		return terror.Error(err, "failed get notification")
	}

	// send notification
	_, err = t.Send(&tele.Chat{ID: int64(notification.TelegramID.Int)}, message)
	if err != nil {
		return terror.Error(err, "failed to send telegram message")
	}

	return nil
}
