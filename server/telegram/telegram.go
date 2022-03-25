package telegram

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/ninja-software/terror/v2"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
}

var gobFile = "tele.gob"

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

	// open data file

	dataFile, err := os.OpenFile(gobFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer dataFile.Close()

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

func (t *Telegram) CreateCode(playerID string) error {

	// find player
	p, err := boiler.FindPlayer(gamedb.StdConn, playerID)
	if err != nil {
		return terror.Error(err, "failed to get player")
	}

	// get gob file
	dataFile, err := os.OpenFile(gobFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer dataFile.Close()

	// write data to gob
	enc := gob.NewEncoder(dataFile)

	tel := TelegramInfo{
		PlayerID:  p.ID,
		ShortCode: genCode(),
	}

	err = enc.Encode(&tel)
	if err != nil {
		fmt.Println("after", err)

		return err
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

func (t *Telegram) Notify(code string, message string) error {
	// open gob file
	dataFile, err := os.OpenFile(gobFile, os.O_RDWR, os.ModeAppend)
	if err != nil {
		fmt.Println("here", err)
		os.Exit(1)
	}
	defer dataFile.Close()
	dec := gob.NewDecoder(dataFile)
	if err != nil {
		return err
	}

	var dTel TelegramInfo
	err = dec.Decode(&dTel)
	if err != nil {
		fmt.Println("notifying1 ............", err)
		return err
	}

	// get the short cod and user info from gob

	// send tele message

	return nil
}
