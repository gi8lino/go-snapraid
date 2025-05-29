package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/go-snapraid/internal/snapraid"
)

// SendSummaryNotification sends a formatted Slack message with colored status and timing info.
func SendSummaryNotification(token, channel string, result snapraid.RunResult, ts time.Time, timings snapraid.RunTimings) error {
	color := "#2ECC71"
	if result.Error != nil {
		color = "#E74C3C"
	}

	res := result.Result

	msg := fmt.Sprintf(
		"SnapRAID run (%s UTC):\n• Added: %d\n• Removed: %d\n• Updated: %d\n• Moved: %d\n• Copied: %d\n• Restored: %d",
		ts.Format("2006-01-02 15:04"),
		res.Added, res.Removed, res.Updated, res.Moved, res.Copied, res.Restored,
	)

	// Add timings if non-zero
	var timingLines []string
	if timings.Touch > 0 {
		timingLines = append(timingLines, fmt.Sprintf("• Touch: %s", timings.Touch.Truncate(time.Second)))
	}
	if timings.Diff > 0 {
		timingLines = append(timingLines, fmt.Sprintf("• Diff: %s", timings.Diff.Truncate(time.Second)))
	}
	if timings.Sync > 0 {
		timingLines = append(timingLines, fmt.Sprintf("• Sync: %s", timings.Sync.Truncate(time.Second)))
	}
	if timings.Scrub > 0 {
		timingLines = append(timingLines, fmt.Sprintf("• Scrub: %s", timings.Scrub.Truncate(time.Second)))
	}
	if timings.Smart > 0 {
		timingLines = append(timingLines, fmt.Sprintf("• Smart: %s", timings.Smart.Truncate(time.Second)))
	}
	if timings.Total > 0 {
		timingLines = append(timingLines, fmt.Sprintf("• Total: %s", timings.Total.Truncate(time.Second)))
	}

	if len(timingLines) > 0 {
		msg += "\n\nTimings:\n" + fmt.Sprint(strings.Join(timingLines, "\n"))
	}

	return sendSlackAttachment(token, channel, msg, color)
}
