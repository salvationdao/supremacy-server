package telegram

import (
	"database/sql"
	"errors"
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
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	*tele.Bot
	RegisterCallback func(ownderID string, success bool)
}

// NewTelegram
func NewTelegram(token string, environment string, registerCallback func(shortCode string, success bool)) (*Telegram, error) {

	t := &Telegram{
		RegisterCallback: registerCallback,
	}

	if environment == "production" || environment == "staging" {
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
	}

	return t, nil
}

var telegramNotifications = map[string][]string{}

const HubKeyTelegramShortcodeRegistered = "USER:TELEGRAM_SHORTCODE_REGISTERED"

// registers new players
func (t *Telegram) RunTelegram(bot *tele.Bot) error {
	if t.Bot == nil {
		return nil
	}

	// registers fist time user
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
			return c.Send("Unable to register shortcode, try again or contact support.")
		}

		// get telegram player via short code
		telegramPlayer, err := boiler.TelegramPlayers(
			boiler.TelegramPlayerWhere.TelegramID.IsNull(),
			boiler.TelegramPlayerWhere.Shortcode.EQ(strings.ToLower(shortcode))).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("unable to get notification by shortcode")
			return c.Send("Unable to find shortcode, you may have entered your shortcode too fast, please try again or contact support.")
		}

		reply := ""

		// cant find telgram player by shortcode
		if errors.Is(err, sql.ErrNoRows) {
			reply = "Invalid shortcode!"
			return c.Send(reply)
		}

		// if found
		telegramPlayer.TelegramID = null.Int64From(int64(_telegramID))

		_, err = telegramPlayer.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("telegramID", telegramID).
				Str("telegramPlayer", telegramPlayer.ID).
				Msg("unable to update telegram player")
			return terror.Error(err)
		}

		if err != nil {
			reply = "Issue regestering, try again or contact support"
			go t.RegisterCallback(telegramPlayer.PlayerID, false)
			return c.Send(reply)

		}

		reply = "Shortcode registered! You will be notified when your war machine is nearing battle"
		go t.RegisterCallback(telegramPlayer.PlayerID, true)
		return c.Send(reply)
	})
	bot.Start()
	return nil
}

func (t *Telegram) PlayerCreate(player *boiler.Player) (*boiler.TelegramPlayer, error) {

	shortcode, err := shortid.Generate()
	if err != nil {
		return nil, terror.Error(err)
	}

	codeExists := true
	for codeExists {
		// check if a telegram player already has that shortcode
		exists, err := boiler.TelegramPlayers(boiler.TelegramPlayerWhere.Shortcode.EQ(strings.ToLower(shortcode))).Exists(gamedb.StdConn)
		if err != nil {
			return nil, terror.Error(err, "Unable to check if telegram player exists")
		}

		if exists {
			// if code already exist generate new one
			shortcode, err = shortid.Generate()
			if err != nil {
				return nil, terror.Error(err)
			}
		} else {
			codeExists = false
		}
	}

	// create telegram player
	tPlayer := &boiler.TelegramPlayer{
		PlayerID:  player.ID,
		Shortcode: strings.ToLower(shortcode),
	}

	err = player.AddTelegramPlayers(gamedb.StdConn, true, tPlayer)
	if err != nil {
		return nil, terror.Error(err, "Unable to create telegram player")
	}

	return tPlayer, nil

}

func (t *Telegram) Notify(telegramNotificationID string, message string) error {
	if t.Bot == nil {
		return nil
	}
	// get telegram notification
	notification, err := boiler.FindTelegramNotification(gamedb.StdConn, telegramNotificationID)
	if err != nil {
		return terror.Error(err, "failed get notification")
	}
	if !notification.TelegramID.Valid {
		gamelog.L.Warn().Msg("invalid telegram ID")
		return nil
	}
	// send notification
	_, err = t.Send(&tele.Chat{ID: int64(notification.TelegramID.Int)}, message)
	if err != nil {
		return terror.Error(err, "failed to send telegram message")
	}

	return nil
}

func (t *Telegram) Notify2(telegramID int64, message string) error {
	if t.Bot == nil {
		return nil
	}

	// send notification
	_, err := t.Send(&tele.Chat{ID: int64(telegramID)}, message)
	if err != nil {
		return terror.Error(err, "failed to send telegram message")
	}

	return nil
}
