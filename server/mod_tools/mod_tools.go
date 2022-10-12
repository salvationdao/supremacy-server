package mod_tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server"
	"server/db"
	"server/gamelog"
)

const SlackAPIBaseURL = "https://slack.com/api"

var SlackModToolsAppToken string

type SlackMessagePayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func SendSlackNotification(slackMessage string, slackChannel string) error {
	if server.IsDevelopmentEnv() {
		if !db.GetBoolWithDefault("send_slack_dev_notification", false) {
			gamelog.L.Info().Msg("Slack notification send is turned off for dev")
			return nil
		}
		// Send Slack notification to #slack-app-test chat to dev-ops to be added to test channel
		slackChannel = "C04648C7ZNE"
	}

	if SlackModToolsAppToken == "" {
		return terror.Error(fmt.Errorf("slack mod app auth token not provided"), "Slack mod app auth token not provided")
	}

	if slackChannel == "" {
		return terror.Error(fmt.Errorf("slack channel id not provided"), "Slack channel id not provided")
	}

	slackMessagePayload := &SlackMessagePayload{
		Channel: string(slackChannel),
		Text:    slackMessage,
	}

	payloadBytes, err := json.Marshal(slackMessagePayload)
	if err != nil {
		gamelog.L.Err(err).Interface("SlackMessagePayload", slackMessagePayload).Msg("Failed to marshal slack notification")
		return terror.Error(err, "Failed to marshal slack notification")
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat.postMessage", SlackAPIBaseURL), body)
	if err != nil {
		return terror.Error(err, "Failed to send slack notification")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", SlackModToolsAppToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return terror.Error(err, "Failed to send slack notification")
	}
	defer resp.Body.Close()

	return nil
}
