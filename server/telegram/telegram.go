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

func (t *Telegram) TelegramNotificationCreate(playerID string, mechID string, code string) error {
	notification := &boiler.TelegramNotification{
		PlayerID:  playerID,
		Shortcode: code,
		MechID:    mechID,
	}

	err := notification.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err)
	}

	return nil
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

func (t *Telegram) Notify(playerID string, mechID string, message string) error {
	// get notification
	notification, err := boiler.TelegramNotifications(boiler.TelegramNotificationWhere.PlayerID.EQ(playerID), boiler.TelegramNotificationWhere.MechID.EQ(mechID)).One(gamedb.StdConn)
	if err != nil {
		// TODO: handle no rows
		return terror.Error(err)
	}

	// send notification
	t.SendMessage(notification.TelegramID.Int, "Your warmachine will be in battle soon!")

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

func (t *Telegram) List() {
	notifications, err := boiler.TelegramNotifications().All(gamedb.StdConn)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, l := range notifications {
		fmt.Println("+++++++++++++++++")
		fmt.Println("ID: ", l.ID)
		fmt.Println("TelegramID: ", l.TelegramID)
		fmt.Println("MechID: ", l.MechID)
		fmt.Println("PlayerID: ", l.PlayerID)
		fmt.Println("Registered: ", l.Registered)
		fmt.Println("Shortcode: ", l.Shortcode)
		// fmt.Println("QueuePosition: ", l.QueuePosition)

		fmt.Println("+++++++++++++++++")
	}

}

func (t *Telegram) Insert() {
	expiry := time.Now().Add(time.Minute * 3)

	notification := &boiler.TelegramNotification{
		PlayerID:  "1a657a32-778e-4612-8cc1-14e360665f2b",
		MechID:    "fc43fa34-b23f-40f4-afaa-465f4880ef59",
		Shortcode: genCode(),
		ExpiresAt: null.TimeFrom(expiry),
	}

	err := notification.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		fmt.Println("insert ", err)
		return
	}

	fmt.Println("code here", notification.Shortcode)

}

func (t *Telegram) GenCode() {

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
