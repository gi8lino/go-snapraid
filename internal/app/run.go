package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gi8lino/go-snapraid/internal/config"
	"github.com/gi8lino/go-snapraid/internal/flag"
	"github.com/gi8lino/go-snapraid/internal/logging"
	"github.com/gi8lino/go-snapraid/internal/notify"
	"github.com/gi8lino/go-snapraid/internal/snapraid"
)

// Run is the main entrypoint for the SnapRAID runner application.
func Run(ctx context.Context, version, commit string, args []string, w io.Writer) error {
	// Parse CLI flags
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

	// Setup logger
	logger := logging.SetupLogger(flags.LogFormat, w)
	logger.Info("Starting snapraid runner", "version", version, "commit", commit)

	// Load YAML config
	cfg, err := config.LoadConfig(flags.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Apply CLI overrides on top of config
	flag.ApplyOverrides(&cfg, flags)

	// Initialize SnapRAID runner
	runner := snapraid.New(
		cfg.ConfigFile,
		cfg.SnapraidBin,
		snapraid.WithOutputDir(cfg.OutputDir),
		snapraid.WithSteps(
			snapraid.Steps{
				Touch: cfg.Steps.Touch,
				Scrub: cfg.Steps.Scrub,
				Smart: cfg.Steps.Smart,
			}),
		snapraid.WithThresholds(
			snapraid.Thresholds{
				Add:     cfg.Thresholds.Add,
				Remove:  cfg.Thresholds.Remove,
				Update:  cfg.Thresholds.Update,
				Move:    cfg.Thresholds.Move,
				Copy:    cfg.Thresholds.Copy,
				Restore: cfg.Thresholds.Restore,
			}),
		snapraid.WithScrubOptions(
			cfg.Scrub.Plan,
			cfg.Scrub.OlderThan,
		),
	)

	// Run the SnapRAID pipeline
	result := runner.Run()

	// Log change summary
	if result.HasChanges() {
		logger.Info("No changes detected")
	} else {
		logger.Info("SnapRAID sync completed",
			"added", result.Result.Added,
			"removed", result.Result.Removed,
			"updated", result.Result.Updated,
			"moved", result.Result.Moved,
			"copied", result.Result.Copied,
			"restored", result.Result.Restored,
		)
	}

	// Persist run result to file
	if cfg.OutputDir != "" {
		if err := result.WriteJSON(cfg.OutputDir); err != nil {
			logger.Warn("Failed to write result file", "error", err)
		}
	}

	// Send Slack notification
	if cfg.Notify.SlackToken != "" && cfg.Notify.SlackChannel != "" {
		if err := notify.SendSummaryNotification(
			cfg.Notify.SlackToken,
			cfg.Notify.SlackChannel,
			result,
			runner.Timestamp,
			result.Timings,
		); err != nil {
			logger.Error("Slack notification failed", "error", err)
		}
	}

	// Return error if one occurred
	if result.Error != nil {
		return fmt.Errorf("snapraid run failed: %w", result.Error)
	}

	logger.Info("All done")
	return nil
}
