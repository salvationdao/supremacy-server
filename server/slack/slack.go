package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server"
	"server/db"
	"server/gamelog"
	"time"
)

const APIBaseURL = "https://slack.com/api"

var ModToolsAppToken string

type SendMessagePayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func SendSlackNotification(slackMessage, slackChannel, appToken string) error {
	if server.IsDevelopmentEnv() {
		if !db.GetBoolWithDefault("send_slack_dev_notification", false) {
			gamelog.L.Info().Msg("Slack notification send is turned off for dev")
			return nil
		}
		// Send Slack notification to #slack-app-test chat to dev-ops to be added to test channel or create a channel and change value in kv
		slackChannel = db.GetStrWithDefault(db.KeySlackDevChannelID, "C04648C7ZNE")
	}

	if ModToolsAppToken == "" {
		return terror.Error(fmt.Errorf("slack mod app auth token not provided"), "Slack mod app auth token not provided")
	}

	if slackChannel == "" {
		return terror.Error(fmt.Errorf("slack channel id not provided"), "Slack channel id not provided")
	}

	slackMessagePayload := &SendMessagePayload{
		Channel: slackChannel,
		Text:    slackMessage,
	}

	payloadBytes, err := json.Marshal(slackMessagePayload)
	if err != nil {
		gamelog.L.Err(err).Interface("SendMessagePayload", slackMessagePayload).Msg("Failed to marshal slack notification")
		return terror.Error(err, "Failed to marshal slack notification")
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat.postMessage", APIBaseURL), body)
	if err != nil {
		return terror.Error(err, "Failed to send slack notification")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", appToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 3 * time.Second, // set timeout prevent hanging
	}

	resp, err := client.Do(req)
	if err != nil {
		return terror.Error(err, "Failed to send slack notification")
	}
	defer resp.Body.Close()

	return nil
}
