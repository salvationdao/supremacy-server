package telegram

import (
	"testing"

	telebot "gopkg.in/telebot.v3"
)

func TestTelegram(t *testing.T) {

	tele, err := NewTelegram("", "staging", func(shortCode string, success bool) {})
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to created new twilio")
		return
	}

	var telegramID int64 = 0

	_, err = tele.Bot.Send(&telebot.Chat{ID: int64(telegramID)}, "this is a test message")
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to send message")
		return
	}

}
