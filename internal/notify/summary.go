package notify

import (
	"fmt"
	"time"

	"github.com/gi8lino/go-snapraid/internal/snapraid"
)

// SendSummaryNotification sends a formatted Slack message with colored status.
func SendSummaryNotification(token, channel string, result snapraid.DiffResult, err error, ts time.Time) error {
	color := "#2ECC71"
	if err != nil {
		color = "#E74C3C"
	}

	msg := fmt.Sprintf(
		"SnapRAID run (%s UTC):\n• Added: %d\n• Removed: %d\n• Updated: %d\n• Moved: %d\n• Copied: %d\n• Restored: %d",
		ts.Format("2006-01-02 15:04"),
		result.Added, result.Removed, result.Updated,
		result.Moved, result.Copied, result.Restored,
	)

	return SendSlackAttachment(token, channel, msg, color)
}
