package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const slackPostMessageAPI string = "https://slack.com/api/chat.postMessage"

// sendSlackAttachment posts a summary message with color.
func sendSlackAttachment(token, channel, message, color string) error {
	channel = "#" + strings.TrimPrefix(channel, "#") // make sure it's prefixed with #
	payload := map[string]any{
		"channel": channel,
		"attachments": []map[string]any{
			{
				"color": color,
				"text":  message,
				"type":  "mrkdwn",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", slackPostMessageAPI, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned %s", resp.Status)
	}
	return nil
}
