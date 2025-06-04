package flag

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("Error Message", func(t *testing.T) {
		t.Parallel()

		err := &HelpRequested{Message: "This is a help message"}
		assert.Equal(t, "This is a help message", err.Error(), "Error() should return the correct message")
	})

	t.Run("No flags", func(t *testing.T) {
		t.Parallel()

		// No flags: defaults should apply
		opts, err := ParseFlags([]string{}, "v1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "/etc/snapraid-runner.yml", opts.ConfigFile)
		assert.False(t, opts.Verbose)
		assert.False(t, opts.DryRun)
		assert.False(t, opts.NoNotify)
		assert.Empty(t, opts.OutputDir)
		// Steps default: all false
		assert.False(t, opts.Steps.NoTouch)
		assert.False(t, opts.Steps.NoScrub)
		assert.False(t, opts.Steps.NoSmart)
		// Threshold defaults: enabled (true) unless disabled by flags
		assert.True(t, opts.Thresholds.NoAdd)
		assert.True(t, opts.Thresholds.NoRemove)
		assert.True(t, opts.Thresholds.NoUpdate)
		assert.True(t, opts.Thresholds.NoCopy)
		assert.True(t, opts.Thresholds.NoMove)
		assert.True(t, opts.Thresholds.NoRestore)
		// ScrubPlan/Older defaults
		assert.Equal(t, 22, opts.ScrubPlan)
		assert.Equal(t, 12, opts.ScrubOlder)
	})

	t.Run("Help flag", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, "v1.0.0")
		assert.Error(t, err)
		helpErr, ok := err.(*HelpRequested)
		assert.True(t, ok)
		// Should contain usage prefix
		assert.Contains(t, helpErr.Message, "Usage: snapraid-runner")
	})

	t.Run("Version flag", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--version"}, "v9.8.7")
		assert.Error(t, err)
		verErr, ok := err.(*HelpRequested)
		assert.True(t, ok)
		assert.Equal(t, "snapraid-runner version: v9.8.7\n", verErr.Message)
	})

	t.Run("Touch and no-touch", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--touch", "--no-touch"}, "v1.0.0")
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot use both --touch and --no-touch")
	})

	t.Run("Scrub and no-scrub", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--scrub", "--no-scrub"}, "v1.0.0")
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot use both --scrub and --no-scrub")
	})

	t.Run("Smart and no-smart", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--smart", "--no-smart"}, "v1.0.0")
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot use both --smart and --no-smart")
	})

	t.Run("Step and threshold resolution", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--touch",
			"--scrub",
			"--no-smart",
			"--no-threshold-add",
			"--no-threshold-mv",
			"--plan", "55",
			"--older-than", "7",
			"--output-dir", "/tmp/results",
			"--verbose",
			"--dry-run",
			"--no-notify",
		}
		opts, err := ParseFlags(args, "v1.0.0")
		assert.NoError(t, err)

		// Steps
		assert.True(t, opts.Steps.NoTouch)
		assert.True(t, opts.Steps.NoScrub)
		assert.False(t, opts.Steps.NoSmart)

		// Thresholds
		assert.False(t, opts.Thresholds.NoAdd)
		assert.True(t, opts.Thresholds.NoRemove)
		assert.True(t, opts.Thresholds.NoUpdate)
		assert.True(t, opts.Thresholds.NoCopy)
		assert.False(t, opts.Thresholds.NoMove)
		assert.True(t, opts.Thresholds.NoRestore)

		// Scrub options
		assert.Equal(t, 55, opts.ScrubPlan)
		assert.Equal(t, 7, opts.ScrubOlder)

		// Other flags
		assert.Equal(t, "/tmp/results", opts.OutputDir)
		assert.True(t, opts.Verbose)
		assert.True(t, opts.DryRun)
		assert.True(t, opts.NoNotify)
	})
}

func TestOptionsValidate(t *testing.T) {
	t.Parallel()

	// Case: ConfigFile does not exist
	t.Run("ConfigFileMissing", func(t *testing.T) {
		t.Parallel()

		cfg := Options{ConfigFile: "/nonexistent/file.yml"}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid config file not found: /nonexistent/file.yml")
	})

	// Case: ConfigFile exists
	t.Run("ConfigFileExists", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "config.yml")
		assert.NoError(t, os.WriteFile(path, []byte("dummy"), 0o600))

		cfg := Options{ConfigFile: path}
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
