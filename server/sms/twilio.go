package sms

import (
	"fmt"

	"github.com/ninja-software/terror/v2"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Twilio struct {
	*twilio.RestClient
	FromNumber   string
	AllowSending bool
}

func NewTwilio(accountSid, apiKey, apiSecret, fromNumber, environment string) (*Twilio, error) {
	twil := &Twilio{
		FromNumber:   fromNumber,
		AllowSending: false,
	}

	// if prod or staging, check for envars and panic if missing and enable sending
	if environment == "production" || environment == "staging" {
		twil.AllowSending = true
		if accountSid == "" {
			return nil, terror.Error(fmt.Errorf("missing var accountSid"))
		}
		if apiKey == "" {
			return nil, terror.Error(fmt.Errorf("missing var apiKey"))
		}
		if apiSecret == "" {
			return nil, terror.Error(fmt.Errorf("missing var apiSecret"))
		}
		if fromNumber == "" {
			return nil, terror.Error(fmt.Errorf("missing var fromNumber"))
		}

		twil.RestClient = twilio.NewRestClientWithParams(twilio.RestClientParams{
			Username:   apiKey,
			Password:   apiSecret,
			AccountSid: accountSid,
		})
	}

	return twil, nil
}

// SendSMS sends given message to given number
func (t *Twilio) SendSMS(to string, message string) error {
	if !t.AllowSending {
		return nil
	}
	smsParams := &openapi.CreateMessageParams{}
	smsParams.SetTo(to)
	smsParams.SetFrom(t.FromNumber)
	smsParams.SetBody(message)

	_, err := t.ApiV2010.CreateMessage(smsParams)
	if err != nil {
		return terror.Error(err, "Failed send SMS")
	}
	return nil
}

// Lookup returns the number if valid, error if not
func (t *Twilio) Lookup(number string) (string, error) {
	if !t.AllowSending {
		return number, nil
	}
	// returns 404 if number invalid
	resp, err := t.LookupsV1.FetchPhoneNumber(number, nil)
	if err != nil {
		return "", terror.Warn(fmt.Errorf("invalid mobile number %s", number), "Invalid mobile number, please insure correct country code.")
	}
	if resp.PhoneNumber == nil {
		return number, nil
	}
	return *resp.PhoneNumber, nil
}
