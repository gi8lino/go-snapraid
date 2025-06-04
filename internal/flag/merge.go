package flag

import (
	"github.com/gi8lino/go-snapraid/internal/config"
	"github.com/gi8lino/go-snapraid/internal/utils"
)

// ApplyOverrides merges CLI flag values into a loaded Config struct.
func ApplyOverrides(cfg *config.Config, f Options) {
	if f.OutputDir != "" {
		cfg.OutputDir = f.OutputDir
	}

	if f.NoNotify {
		cfg.Notify.SlackToken = ""
		cfg.Notify.SlackChannel = ""
	}

	// CLI step toggles
	if f.Steps.NoTouch {
		cfg.Steps.Touch = utils.Ptr(true)
	}
	if f.Steps.NoScrub {
		cfg.Steps.Scrub = utils.Ptr(true)
	}
	if f.Steps.NoSmart {
		cfg.Steps.Smart = utils.Ptr(true)
	}

	// Threshold disabling
	if !f.Thresholds.NoAdd {
		cfg.Thresholds.Add = utils.Ptr(-1)
	}
	if !f.Thresholds.NoRemove {
		cfg.Thresholds.Remove = utils.Ptr(-1)
	}
	if !f.Thresholds.NoUpdate {
		cfg.Thresholds.Update = utils.Ptr(-1)
	}
	if !f.Thresholds.NoCopy {
		cfg.Thresholds.Copy = utils.Ptr(-1)
	}
	if !f.Thresholds.NoMove {
		cfg.Thresholds.Move = utils.Ptr(-1)
	}
	if !f.Thresholds.NoRestore {
		cfg.Thresholds.Restore = utils.Ptr(-1)
	}

	// Check dry run at the end to override any other flags
	if f.DryRun {
		cfg.Steps.Touch = utils.Ptr(false)
		cfg.Steps.Scrub = utils.Ptr(false)
		cfg.Steps.Smart = utils.Ptr(false)
	}
}
