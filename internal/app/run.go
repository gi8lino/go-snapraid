package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gi8lino/go-snapraid/internal/flag"
	"github.com/gi8lino/go-snapraid/internal/logging"
	"github.com/gi8lino/go-snapraid/internal/notify"
	"github.com/gi8lino/go-snapraid/internal/snapraid"
)

// Run is the main entrypoint for the application.
func Run(ctx context.Context, version, commit string, args []string, w io.Writer) error {
	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		var helpErr *flag.HelpRequested
		if errors.As(err, &helpErr) {
			fmt.Fprint(w, helpErr.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("parsing error: %w", err)
	}
	if err := flags.Validate(); err != nil {
		return fmt.Errorf("invalid CLI flags: %w", err)
	}

	logger := logging.SetupLogger(flags.LogFormat, w)
	logger.Info("Starting snapraid runner", "version", version, "commit", commit)

	timestamp := time.Now().UTC()
	result, runErr := snapraid.Runner(flags)

	if snapraid.CountChanges(result) == 0 {
		logger.Info("No changes detected")
	} else {
		logger.Info("Snapraid sync completed",
			"added", result.Added,
			"removed", result.Removed,
			"updated", result.Updated,
			"moved", result.Moved,
			"copied", result.Copied,
			"restored", result.Restored)
	}

	if err := snapraid.WriteResultJSON(flags.OutputDir, result, timestamp, runErr); err != nil {
		logger.Warn("Failed to write result file", "error", err)
	}

	if flags.SlackToken != "" && flags.SlackChan != "" {
		if err := notify.SendSummaryNotification(flags.SlackToken, flags.SlackChan, result, runErr, timestamp); err != nil {
			logger.Error("Slack notification failed", "error", err)
		}
	}

	if runErr != nil {
		return fmt.Errorf("snapraid run failed: %w", runErr)
	}

	logger.Info("All done")
	return nil
}
