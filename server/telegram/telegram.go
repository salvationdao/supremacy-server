package telegram

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"server/stormdb"
	"strconv"
	"time"

	"github.com/ninja-software/terror/v2"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
	StormDB *stormdb.StormDB
}

var teleToken = "5179636156:AAFiG_uba7EZm9AFbkK5HRaez3LfhgvHPXI" // TODO: change to real token, get from env var

// NewTelegram
func NewTelegram(stormDB *stormdb.StormDB) (*Telegram, error) {
	t := &Telegram{
		StormDB: stormDB,
	}
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
			text := c.Text()
			teleID := c.Recipient().Recipient()
			id, err := strconv.Atoi(teleID)
			if err != nil {
				fmt.Println(err)
			}

			reply := ""
			// get notification by shortcode
			_, err = t.StormDB.TelegramNotificationRegister(text, "", id)
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
	// save to storm with user + mech id's
	err := t.StormDB.TelegramNotificationCreate(playerID, mechID, code)
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
	notification, err := t.StormDB.TelegramGetNotificationByPlayerCode(playerID, mechID)
	if err != nil {
		return err
	}

	// send notification
	t.SendMessage(notification.TelegramID, "Your warmachine will be in battle soon!")

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
	n, err := t.StormDB.TelegramNotificationsList()
	if err != nil {
		fmt.Println(err)
	}

	for _, l := range n {
		fmt.Println("+++++++++++++++++")
		fmt.Println("ID: ", l.ID)
		fmt.Println("TelegramID: ", l.TelegramID)
		fmt.Println("MechID: ", l.MechID)
		fmt.Println("PlayerID: ", l.PlayerID)
		fmt.Println("Registered: ", l.Registered)
		fmt.Println("Shortcode: ", l.Shortcode)
		fmt.Println("QueuePosition: ", l.QueuePosition)

		fmt.Println("+++++++++++++++++")

	}
}

func (t *Telegram) Insert() {
	err := t.StormDB.TelegramNotificationCreate("1a657a32-778e-4612-8cc1-14e360665f2b", "", genCode())
	if err != nil {
		fmt.Println("error inserting", err)
	}
}

func (t *Telegram) GenCode() {
	te, err := t.StormDB.TelegramGenCodeTest()
	if err != nil {
		fmt.Println("error gen code test")
		return

	}
	fmt.Println("code here", te.Shortcode)
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
