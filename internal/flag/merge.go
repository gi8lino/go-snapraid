package flag

import "github.com/gi8lino/go-snapraid/internal/config"

// ApplyOverrides merges CLI flag values into a loaded Config struct.
func ApplyOverrides(cfg *config.Config, f Options) {
	if f.DryRun {
		cfg.Steps.Touch = false // touch doesn't make sense in dry run
		cfg.Steps.Scrub = false
		cfg.Steps.Smart = false
	}

	if f.OutputDir != "" {
		cfg.OutputDir = f.OutputDir
	}

	if f.NoNotify {
		cfg.Notify.SlackToken = ""
		cfg.Notify.SlackChannel = ""
	}

	// CLI step toggles
	if f.Steps.Touch {
		cfg.Steps.Touch = true
	}
	if f.Steps.Scrub {
		cfg.Steps.Scrub = true
	}
	if f.Steps.Smart {
		cfg.Steps.Smart = true
	}

	// Threshold disabling
	if !f.Thresholds.Add {
		cfg.Thresholds.Add = -1
	}
	if !f.Thresholds.Remove {
		cfg.Thresholds.Remove = -1
	}
	if !f.Thresholds.Update {
		cfg.Thresholds.Update = -1
	}
	if !f.Thresholds.Copy {
		cfg.Thresholds.Copy = -1
	}
	if !f.Thresholds.Move {
		cfg.Thresholds.Move = -1
	}
	if !f.Thresholds.Restore {
		cfg.Thresholds.Restore = -1
	}
}
