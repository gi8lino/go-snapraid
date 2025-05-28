package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SendSlackAttachment posts a summary message with color.
func SendSlackAttachment(token, channel, message, color string) error {
	payload := map[string]any{
		"channel": channel,
		"attachments": []map[string]any{
			{
				"color": color,
				"text":  message,
			},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(body))
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
