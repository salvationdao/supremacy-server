package telegram

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"server/db/boiler"
	"server/gamedb"
	"strconv"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
}

var teleToken = "5179636156:AAFiG_uba7EZm9AFbkK5HRaez3LfhgvHPXI" // TODO: change to real token, get from env var

// NewTelegram
func NewTelegram() (*Telegram, error) {
	t := &Telegram{}
	pref := tele.Settings{
		Token:  teleToken, // get from env
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	t.Bot = b
	go t.Bot.Start()
	t.RegisterHandler()

	return t, nil
}

func (t *Telegram) RegisterHandler() {
	// on /register
	t.Bot.Handle("/register", func(c tele.Context) error {

		// handle user reply
		t.Bot.Handle(tele.OnText, func(c tele.Context) error {
			shortcode := c.Text()
			telegramID := c.Recipient().Recipient()
			_telegramID, err := strconv.Atoi(telegramID)
			if err != nil {
				fmt.Println(err)
			}

			reply := ""
			// get notification by shortcode
			notification, err := boiler.TelegramNotifications(boiler.TelegramNotificationWhere.Shortcode.EQ(shortcode)).One(gamedb.StdConn)
			if err != nil {
				return terror.Error(err)
			}

			// register that notification
			notification.Registered = true
			notification.TelegramID = null.IntFrom(_telegramID)

			// update notification
			_, err = notification.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				return terror.Error(err)
			}
			if err != nil {
				reply = "invalid code!"
			} else {
				reply = "code registered!"
			}

			return c.Send(reply)
		})

		return c.Send("Enter shortcode")
	})

}

func genCode() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	n := 5
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (t *Telegram) Notify(mechID string, message string) error {
	// get telegram notification
	notification, err := boiler.TelegramNotifications(
		boiler.TelegramNotificationWhere.Registered.EQ(true),
		qm.InnerJoin("battle_queue_notifications bqn ON bqn.telegram_notification_id = telegram_notifications.id"),
		qm.Where(`bqn.mech_id = ? and bqn.sent_at is null`, mechID),
	).One(gamedb.StdConn)
	if err != nil {
		// TODO: handle no rows
		return terror.Error(err)
	}

	// send notification
	err = t.SendMessage(notification.TelegramID.Int, message)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// SendMessage sends telegram message using telegram user's chat_id
func (t *Telegram) SendMessage(chatId int, text string) error {
	telegramApi := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", teleToken)
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})

	if err != nil {
		return terror.Error(err, "an error occurred while sending message")
	}
	defer response.Body.Close()

	return nil
}

func (t *Telegram) NotificationCreate(mechID string) (*boiler.TelegramNotification, error) {
	// check if there already a telegram notification for this mech that hasnt been sent
	exists, err := boiler.BattleQueueNotifications(
		boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
		qm.InnerJoin("telegram_notifications tn on tn.id = battle_queue_notifications.telegram_notification_id"),
		qm.Where("battle_queue_notifications.mech_id = ?", mechID)).Exists(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Unable check if notification exists")
	}

	if exists {
		return nil, terror.Error(err)

	}

	notification := &boiler.BattleQueueNotification{
		MechID:      mechID,
		QueueMechID: null.StringFrom(mechID),
	}

	err = notification.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err, "Unable to insert notification")
	}

	code := genCode()
	codeExists := true
	if codeExists {
		// check if a notification that hasnt been sent has that short code
		exists, err := boiler.BattleQueueNotifications(
			boiler.BattleQueueNotificationWhere.SentAt.IsNull(),
			qm.InnerJoin("telegram_notifications tn on tn.id = battle_queue_notifications.telegram_notification_id"),
			qm.Where("tn.shortcode = ?", code)).Exists(gamedb.StdConn)
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
		Shortcode:  code,
		Registered: false,
	}

	err = notification.SetTelegramNotification(gamedb.StdConn, true, telegramNotification)
	if err != nil {
		return nil, terror.Error(err, "Unable to link notification to telegram notification")
	}

	return telegramNotification, nil

}

// deploy mech
// - gen shortcode
// - save that in storm (telegram_notifs)
// - set registered false
// - with users + mech details

// go to telegram
// - user send tele: register
// - reply tele: type code
// - query storm select * from telegram_notifs where shortcode = s
// - set registered true
// - reply tele: notification set for {mech} {queue number}
// - save that in storm

// on notify
// get user info from storm
// send tele mesage

// telegram_notifs
// shortcode
// payer_id
// mech_id
// message
// registered

// select * from users u
// where u.id = 'jdksjfj' and u.name = 'name'
