package telegram

import (
	"fmt"
	"math/rand"
	"server/stormdb"
	"time"

	"github.com/ninja-software/terror/v2"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
	StormDB *stormdb.StormDB
}

func NewTelegram() (*Telegram, error) {
	t := &Telegram{}
	pref := tele.Settings{
		Token:  "5179636156:AAFiG_uba7EZm9AFbkK5HRaez3LfhgvHPXI", // get from env
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(pref)
	if err != nil {
		fmt.Println("err")
		return nil, err
	}

	t.Bot = b
	go t.Bot.Start()
	t.TelegramRouter()

	return t, nil
}

func (t *Telegram) TelegramRouter() {
	t.Bot.Handle("/register", func(c tele.Context) error {
		t.Bot.Handle(tele.OnText, func(c tele.Context) error {
			text := c.Text()

			// check if shortcode is valid

			return c.Send(text)
		})

		return c.Send("please type code")
	})

}

type TelegramInfo struct {
	PlayerID  string
	ShortCode string
}

func (t *Telegram) CodeCreate(playerID string, mechID string) error {
	// TODO: code passed from the front end
	code := "code"

	// save to storm with user + mech id's
	err := t.StormDB.TelegramNotificationCreate(playerID, "", code)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func (t *Telegram) CodeRegister(playerID string, mechID string, shortCode string) error {
	t.Bot.Handle("/register", func(c tele.Context) error {
		t.Bot.Handle(tele.OnText, func(c tele.Context) error {
			text := c.Text()

			// check if shortcode is valid
			err := t.StormDB.TelegramNotificationRegister(shortCode)
			if err != nil {
				return terror.Error(err)
			}

			return c.Send(text)
		})

		return c.Send("please type code")
	})

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

func (t *Telegram) Notify(code string, message string) error {

	return nil
}
