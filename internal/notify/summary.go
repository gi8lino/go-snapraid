package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/go-snapraid/pkg/snapraid"
)

// SendSummaryNotification sends a formatted Slack message summarizing the SnapRAID run result.
func SendSummaryNotification(
	dryRun, hadError bool,
	token, channel, webURL string,
	result snapraid.RunResult,
	ts time.Time,
	timings snapraid.RunTimings,
) error {
	statusLabel := "[SUCCESS]"
	color := "#2ECC71"
	if hadError {
		statusLabel = "[ERROR]"
		color = "#E74C3C"
	}
	if dryRun {
		statusLabel = "[DRY RUN]-" + statusLabel
	}

	msg := formatSlackSummary(result, ts, timings, statusLabel)

	if webURL != "" {
		msg += "\n\n<" + webURL + "|View Results>"
	}

	return sendSlackAttachment(token, channel, msg, color)
}

// formatSlackSummary builds the message text for a Slack notification.
func formatSlackSummary(result snapraid.RunResult, ts time.Time, timings snapraid.RunTimings, statusLabel string) string {
	res := result.Result
	lines := []string{
		fmt.Sprintf("%s go-snapraid run (%s):", statusLabel, ts.Format("2006-01-02 15:04")),
		fmt.Sprintf(" • Equal:    %d", res.Equal),
		fmt.Sprintf(" • Added:    %d", len(res.Added)),
		fmt.Sprintf(" • Removed:  %d", len(res.Removed)),
		fmt.Sprintf(" • Updated:  %d", len(res.Updated)),
		fmt.Sprintf(" • Moved:    %d", len(res.Moved)),
		fmt.Sprintf(" • Copied:   %d", len(res.Copied)),
		fmt.Sprintf(" • Restored: %d", len(res.Restored)),
	}

	// Append timings
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

	if len(timingLines) > 0 {
		lines = append(lines, "", "Timings:")
		lines = append(lines, timingLines...)
	}

	// Show errors
	if result.Error != nil {
		lines = append(lines, "", "Errors:")
		lines = append(lines, result.Error.Error())
	}

	return strings.Join(lines, "\n")
}
