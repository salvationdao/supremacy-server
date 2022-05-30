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

		if c.Message().ReplyTo.Text != "Enter shortcode" {
			return nil
		}

		// shortcode from recipient's reply
		shortcode := c.Text()
		recipient := c.Recipient().Recipient()

		// telegram id from recipient
		telegramID, err := strconv.Atoi(recipient)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable convert telegramID to int")
			return c.Send("Unable to register shortcode, try again or contact support.")
		}

		// get player's preferences via short code
		prefs, err := boiler.PlayerSettingsPreferences(
			boiler.PlayerSettingsPreferenceWhere.TelegramID.IsNull(),
			boiler.PlayerSettingsPreferenceWhere.Shortcode.EQ(strings.ToLower(shortcode))).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Msg("unable to get player by shortcode")
			return c.Send("Unable to find shortcode, you may have entered your shortcode too fast, please try again or contact support.")
		}

		reply := ""

		// cant find telgram player by shortcode
		if errors.Is(err, sql.ErrNoRows) {
			reply = "Unable to find shortcode, you may have entered your shortcode too fast, please try again or contact support."
			return c.Send(reply)
		}

		// if found set the player's telegram id
		prefs.TelegramID = null.Int64From(int64(telegramID))
		_, err = prefs.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("telegramID", recipient).
				Str("playerPreferences", prefs.ID).
				Msg("unable to update telegram player")
			return terror.Error(err)
		}

		if err != nil {
			reply = "Issue regestering, try again or contact support"
			go t.RegisterCallback(prefs.PlayerID, false)
			return c.Send(reply)

		}

		reply = "Registered Successfully! Telegram notifications are now enabled. You will be notified when your war machine is nearing battle. You can disable it by going to your preferences. NOTE: you will be charged 5 $SUPS when a notification is sent (only applies to battle queue notifications)"
		go t.RegisterCallback(prefs.PlayerID, true)
		return c.Send(reply)
	})
	bot.Start()
	return nil
}

// PreferencesUpdate will either create or update a players preferences with new telegram notification enabled status
func (t *Telegram) PreferencesUpdate(playerID string) (*boiler.PlayerSettingsPreference, error) {

	// generate shortcode
	shortcode, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	codeExists := true
	for codeExists {
		// check if a registered player pref already has that shortcode
		exists, err := boiler.PlayerSettingsPreferences(
			boiler.PlayerSettingsPreferenceWhere.Shortcode.EQ(strings.ToLower(shortcode)),
			boiler.PlayerSettingsPreferenceWhere.TelegramID.IsNotNull(),
		).Exists(gamedb.StdConn)
		if err != nil {
			return nil, terror.Error(err, "Unable to check if telegram player exists")
		}

		if exists {
			// if code already exist generate new one
			shortcode, err = shortid.Generate()
			if err != nil {
				return nil, err
			}
		} else {
			codeExists = false
		}
	}

	// try to find player preference
	prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(playerID)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Unable to get player preferences")
	}

	// if player preferences does not exist make a new one
	if errors.Is(err, sql.ErrNoRows) {
		_prefs := &boiler.PlayerSettingsPreference{
			PlayerID:                    playerID,
			Shortcode:                   strings.ToLower(shortcode),
			EnableTelegramNotifications: true,
		}

		err = _prefs.Insert(gamedb.StdConn, boil.Infer())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, terror.Error(err, "Unable to insert player preferences")
		}

		return _prefs, nil
	}

	// update player preferences
	prefs.EnableTelegramNotifications = true
	prefs.Shortcode = strings.ToLower(shortcode)
	_, err = prefs.Update(gamedb.StdConn, boil.Infer())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Unable to update player prefs")
	}

	return prefs, nil

}

// OLD Notify Method, will be removed
func (t *Telegram) NotifyDEPRECATED(telegramNotificationID string, message string) error {
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

func (t *Telegram) Notify(telegramID int64, message string) error {
	if t.Bot == nil {
		return nil
	}

	// send notification
	_, err := t.Send(&tele.Chat{ID: telegramID}, message)
	if err != nil {
		return terror.Error(err, "failed to send telegram message")
	}

	return nil
}
