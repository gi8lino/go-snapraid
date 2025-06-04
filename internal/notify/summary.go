package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/go-snapraid/pkg/snapraid"
)

// SendSummaryNotification sends a formatted Slack message with status + timing info.
func SendSummaryNotification(
	dryRun bool,
	hadError bool,
	token, channel string,
	result snapraid.RunResult,
	ts time.Time,
	timings snapraid.RunTimings,
) error {
	// Build the “⦿ SUCCESS” or “⦿ ERROR” label
	statusLabel := "⦿ SUCCESS"
	color := "#2ECC71"
	if hadError {
		statusLabel = "⦿ ERROR"
		color = "#E74C3C"
	}
	if dryRun {
		statusLabel = "[DRY RUN] " + statusLabel
	}

	// Build the core summary block
	res := result.Result
	summaryLines := []string{
		fmt.Sprintf("%s SnapRAID run (%s UTC):", statusLabel, ts.Format("2006-01-02 15:04")),
		fmt.Sprintf(" • Equal:    %d", res.Equal),
		fmt.Sprintf(" • Added:    %d", len(res.Added)),
		fmt.Sprintf(" • Removed:  %d", len(res.Removed)),
		fmt.Sprintf(" • Updated:  %d", len(res.Updated)),
		fmt.Sprintf(" • Moved:    %d", len(res.Moved)),
		fmt.Sprintf(" • Copied:   %d", len(res.Copied)),
		fmt.Sprintf(" • Restored: %d", len(res.Restored)),
	}

	// Only include a Timings block if at least one duration is non‐zero
	var timingLines []string
	if timings.Touch > 0 {
		timingLines = append(timingLines, fmt.Sprintf(" • Touch:  %s", timings.Touch.Truncate(time.Second)))
	}
	if timings.Diff > 0 {
		timingLines = append(timingLines, fmt.Sprintf(" • Diff:   %s", timings.Diff.Truncate(time.Second)))
	}
	if timings.Sync > 0 {
		timingLines = append(timingLines, fmt.Sprintf(" • Sync:   %s", timings.Sync.Truncate(time.Second)))
	}
	if timings.Scrub > 0 {
		timingLines = append(timingLines, fmt.Sprintf(" • Scrub:  %s", timings.Scrub.Truncate(time.Second)))
	}
	if timings.Smart > 0 {
		timingLines = append(timingLines, fmt.Sprintf(" • Smart:  %s", timings.Smart.Truncate(time.Second)))
	}
	if timings.Total > 0 {
		timingLines = append(timingLines, fmt.Sprintf(" • Total:  %s", timings.Total.Truncate(time.Second)))
	}

	fullMsg := strings.Join(summaryLines, "\n")
	if len(timingLines) > 0 {
		fullMsg += "\n\nTimings:\n" + strings.Join(timingLines, "\n")
	}

	return sendSlackAttachment(token, channel, fullMsg, color)
}
