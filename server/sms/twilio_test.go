package sms_test

import (
	"fmt"
	"os"
	"server/sms"
	"testing"
)

func TestTwilio_SendSMS(t *testing.T) {
	accountSid := os.Getenv("GAMESERVER_TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("GAMESERVER_TWILIO_API_KEY")
	apiSecret := os.Getenv("GAMESERVER_TWILIO_API_SECRET")
	fromNumber := os.Getenv("GAMESERVER_SMS_FROM_NUMBER")

	toNumber := "+61416315945"
	// 61416315945

	twil, err := sms.NewTwilio(
		accountSid,
		apiKey,
		apiSecret,
		fromNumber,
		"staging",
	)
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to created new twilio")
	}

	_, err = twil.Lookup(toNumber)
	if err != nil {
		fmt.Println("here", err)
		t.Log("here", err)
		t.Fatalf("failed to lookup number %s", toNumber)
	}

	err = twil.SendSMS(toNumber, "Test message from xsyn passport")
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to send number %s", toNumber)
	}
}
