package telegram

import (
	"fmt"
	"time"

	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
}

func NewTelegram() (*Telegram, error) {
	t := &Telegram{}

	pref := tele.Settings{
		Token:  "5179636156:AAFiG_uba7EZm9AFbkK5HRaez3LfhgvHPXI", // get from env
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := tele.NewBot(pref)
	if err != nil {
		// log.Fatal(err)
		fmt.Println("err")
		return nil, err
	}

	t.Bot = b
	t.TelegramRouter()

	return t, nil
}

func (t *Telegram) TelegramRouter() {
	t.Bot.Handle("/register", func(c tele.Context) error {
		return c.Send("please type code")
	})

}

func (t *Telegram) CreateCode(userID string) {

}

func (t *Telegram) Notify(code string, message string) error {

	fmt.Println("notifying ............")
	fmt.Println("notifying ............")
	fmt.Println("notifying ............")
	fmt.Println("notifying ............")
	fmt.Println("notifying ............")

	return nil
}
