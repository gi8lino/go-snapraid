package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/gi8lino/go-snapraid/internal/config"
	"github.com/gi8lino/go-snapraid/internal/flag"
	"github.com/gi8lino/go-snapraid/internal/logging"
	"github.com/gi8lino/go-snapraid/internal/notify"
	"github.com/gi8lino/go-snapraid/pkg/snapraid"
)

// Run is the main entrypoint for the SnapRAID runner application.
func Run(ctx context.Context, version, commit string, args []string, w io.Writer) error {
	// Parse CLI flags
	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		if errors.As(err, new(*flag.HelpRequested)) {
			fmt.Fprint(w, err.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}
	if err := flags.Validate(); err != nil {
		return fmt.Errorf("validate flags: %w", err)
	}

	// Setup logger
	logger := logging.SetupLogger(flags.LogFormat, w)
	logger.Info("Starting snapraid runner",
		"version", version,
		"commit", commit,
		"tag", "runner",
	)

	// Load YAML config
	cfg, err := config.LoadConfig(flags.ConfigFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	// Fill in any missing defaults now that we have unmarshaled into cfg.
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
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
