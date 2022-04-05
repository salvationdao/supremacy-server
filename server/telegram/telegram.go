package telegram

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"strings"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/teris-io/shortid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
	RegisterCallback func(ownderID string, success bool)
	// MessageBus *messagebus.MessageBus
}

// NewTelegram
func NewTelegram(token string, registerCallback func(shortCode string, success bool)) (*Telegram, error) {
	t := &Telegram{
		// MessageBus: messageBus,
		RegisterCallback: registerCallback,
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
			qm.Where("tn.shortcode = ?", strings.ToLower(shortcode)),

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

		// mech owner
		wmOwner := notification.R.Mech.OwnerID

		// update notification
		_, err = telegramNotification.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("telegramID", telegramID).
				Str("notificationID", telegramNotification.ID).
				Msg("unable to update telegram notification")
			return terror.Error(err)
		}

		if err != nil {
			reply = "Issue regestering telegram shortcode, try again or contact support"
			go t.RegisterCallback(wmOwner, false)
			return c.Send(reply)

		}

		wmName := notification.R.Mech.Label
		if notification.R.Mech.Name != "" {
			wmName = notification.R.Mech.Name
		}
		reply = fmt.Sprintf("Shortcode registered! you will be notified when your war machine (%s) is nearing battle", wmName)
		go t.RegisterCallback(wmOwner, true)
		return c.Send(reply)
	})
	bot.Start()
	return nil
}

func (t *Telegram) NotificationCreate(mechID string, notification *boiler.BattleQueueNotification) (*boiler.TelegramNotification, error) {

	shortCode, err := shortid.Generate()
	if err != nil {
		return nil, terror.Error(err)
	}

	codeExists := true
	if codeExists {
		// check if a notification that hasnt been sent/ not refunded has that short code
		exists, err := boiler.BattleQueueNotifications(
			boiler.BattleQueueNotificationWhere.IsRefunded.EQ(false),
			boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
			qm.InnerJoin("telegram_notifications tn on tn.id = battle_queue_notifications.telegram_notification_id"),
			qm.Where("tn.shortcode = ?", strings.ToLower(shortCode)),
			qm.Where("tn.Registered = false")).Exists(gamedb.StdConn)
		if err != nil {
			return nil, terror.Error(err, "Unable to get telegram notifications")
		}

		if exists {
			// if code already exist generate new one
			shortCode, err = shortid.Generate()
			if err != nil {
				return nil, terror.Error(err)
			}
		} else {
			codeExists = false
		}
	}

	telegramNotification := &boiler.TelegramNotification{
		Shortcode: strings.ToLower(shortCode),
	}

	err = telegramNotification.Insert(gamedb.StdConn, boil.Infer())
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
