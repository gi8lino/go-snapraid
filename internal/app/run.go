package app

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/gi8lino/go-snapraid/internal/config"
	"github.com/gi8lino/go-snapraid/internal/flag"
	"github.com/gi8lino/go-snapraid/internal/logging"
	"github.com/gi8lino/go-snapraid/internal/notify"
	"github.com/gi8lino/go-snapraid/pkg/snapraid"

	"github.com/containeroo/tinyflags"
)

// Run is the main entrypoint for the SnapRAID runner application.
func Run(ctx context.Context, version, commit string, args []string, w io.Writer) error {
	// Parse CLI flags
	flags, err := flag.ParseFlags(args, version)

	// Setup logger immediately so startup errors are correctly logged.
	logger := logging.SetupLogger(flags.LogFormat, w)
	logger.Info("Starting snapraid runner",
		"version", version,
		"commit", commit,
		"tag", "runner",
	)

	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprintf(w, "%s\n", err) // nolint:errcheck
			return nil
		}
		logger.Error("Failed to parse flags", "error", err, "tag", "runner")
		return err
	}
	if err := flags.Validate(); err != nil {
		logger.Error("Failed to validate flags", "error", err, "tag", "runner")
		return err
	}

	// Load YAML config
	cfg, err := config.LoadConfig(flags.ConfigFile)
	if err != nil {
		logger.Error("Failed to load config", "error", err, "tag", "runner")
		return err
	}
	// Fill in any missing defaults now that we have unmarshaled into cfg.
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		logger.Error("Failed to validate config", "error", err, "tag", "runner")
		return err
	}

	// Apply CLI overrides on top of config
	flag.ApplyOverrides(&cfg, flags)

	// Initialize SnapRAID runner
	runner := snapraid.NewRunner(
		cfg.SnapraidConfig,
		cfg.SnapraidBin,
		cfg.OutputDir,
		snapraid.Steps{
			Touch: *cfg.Steps.Touch,
			Scrub: *cfg.Steps.Scrub,
			Smart: *cfg.Steps.Smart,
		},
		snapraid.Thresholds{
			Add:     *cfg.Thresholds.Add,
			Remove:  *cfg.Thresholds.Remove,
			Update:  *cfg.Thresholds.Update,
			Move:    *cfg.Thresholds.Move,
			Copy:    *cfg.Thresholds.Copy,
			Restore: *cfg.Thresholds.Restore,
		},
		*cfg.Scrub.Plan,
		*cfg.Scrub.OlderThan,
		flags.DryRun,
		logger,
	)

	// Run the SnapRAID pipeline
	result := runner.Run()

	if result.Error != nil {
		logger.Error("SnapRAID run failed", "error", result.Error, "tag", "runner")
		return result.Error
	}

	// Log change summary
	if !result.HasChanges() {
		logger.Info("No changes detected")
	} else {
		logger.Info("SnapRAID sync completed",
			"equal", result.Result.Equal,
			"added", len(result.Result.Added),
			"removed", len(result.Result.Removed),
			"updated", len(result.Result.Updated),
			"moved", len(result.Result.Moved),
			"copied", len(result.Result.Copied),
			"restored", len(result.Result.Restored),
			"errors", result.Error != nil,
			"tag", "runner",
		)
	}

	// Persist run result to file
	if cfg.OutputDir != "" {
		if err := result.WriteJSON(cfg.OutputDir); err != nil {
			logger.Warn("Failed to write result file",
				"error", err,
				"tag", "runner",
			)
		}
	}

	// Send Slack notification
	if cfg.WantsSlackNotification(flags.NoNotify) {
		var web string
		if cfg.Notify.Web != "" {
			web = fmt.Sprintf(
				"%s/#/run/%s",
				strings.TrimRight(cfg.Notify.Web, "/"),
				url.PathEscape(result.Timestamp),
			)
		}

		err := notify.SendSummaryNotification(
			flags.DryRun,
			(result.Error != nil),
			cfg.Notify.SlackToken,
			cfg.Notify.SlackChannel,
			web,
			result,
			runner.Timestamp,
			result.Timings,
		)
		if err != nil {
			logger.Error("Slack notification failed",
				"error", err,
				"tag", "runner",
			)
		} else {
			logger.Info("Slack notification sent", "tag", "runner")
		}
	}

	// Return error if one occurred
	if result.Error != nil {
		logger.Error("SnapRAID run failed", "error", result.Error, "tag", "runner")
		return result.Error
	}

	logger.Info("All done", "tag", "runner")
	return nil
}
