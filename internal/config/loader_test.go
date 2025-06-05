package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi8lino/go-snapraid/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("Non-existent file should return an error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "does_not_exist.yml")

		_, err := LoadConfig(path)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("failed to read config file: open %s: no such file or directory", path))
	})

	t.Run("Invalid YAML content should return an error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "invalid.yml")
		invalidContent := "invalid: [unclosed_list\n"

		err := os.WriteFile(path, []byte(invalidContent), 0o600)
		assert.NoError(t, err)

		_, err = LoadConfig(path)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid YAML: yaml: line 1: did not find expected ',' or ']'")
	})

	t.Run("Valid YAML with all fields should parse correctly", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "full.yml")
		fullContent := `
snapraid_bin: "/usr/bin/snapraid"
snapraid_config: "/etc/snapraid.conf"
output_dir: "/var/lib/snapraid/output"

thresholds:
  add: 10
  remove: 20
  update: 30
  copy: 40
  move: 50
  restore: 60

steps:
  touch: true
  scrub: false
  smart: true

scrub:
  plan: 5
  older_than: 7

notifications:
  slack_token: "xoxb-123"
  slack_channel: "#snapraid"
`
		err := os.WriteFile(path, []byte(fullContent), 0o600)
		assert.NoError(t, err)

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)

		// Verify top-level fields
		assert.Equal(t, "/usr/bin/snapraid", cfg.SnapraidBin)
		assert.Equal(t, "/etc/snapraid.conf", cfg.SnapraidConfig)
		assert.Equal(t, "/var/lib/snapraid/output", cfg.OutputDir)

		// Verify thresholds
		expThresh := Thresholds{
			Add:     utils.Ptr(10),
			Remove:  utils.Ptr(20),
			Update:  utils.Ptr(30),
			Copy:    utils.Ptr(40),
			Move:    utils.Ptr(50),
			Restore: utils.Ptr(60),
		}
		assert.Equal(t, expThresh, cfg.Thresholds)

		// Verify steps
		expSteps := Steps{
			Touch: utils.Ptr(true),
			Scrub: utils.Ptr(false),
			Smart: utils.Ptr(true),
		}
		assert.Equal(t, expSteps, cfg.Steps)

		// Verify scrub options
		assert.Equal(t, 5, *cfg.Scrub.Plan)
		assert.Equal(t, 7, *cfg.Scrub.OlderThan)

		// Verify notifications
		assert.Equal(t, "xoxb-123", cfg.Notify.SlackToken)
		assert.Equal(t, "#snapraid", cfg.Notify.SlackChannel)
	})

	t.Run("Missing thresholds and scrub should get defaults", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "defaults.yml")
		defaultsContent := `
snapraid_bin: "/usr/bin/snapraid"
snapraid_config: "/etc/snapraid.conf"
output_dir: "/var/lib/snapraid/output"

steps:
  touch: false
  scrub: true
  smart: false

notifications:
  slack_token: "token"
  slack_channel: "#channel"
`
		err := os.WriteFile(path, []byte(defaultsContent), 0o600)
		assert.NoError(t, err)

		cfg, err := LoadConfig(path)
		cfg.ApplyDefaults()
		assert.NoError(t, err)

		// Verify thresholds use defaults when omitted
		assert.Equal(t, -1, *cfg.Thresholds.Add)     // defaultAddThreshold
		assert.Equal(t, 80, *cfg.Thresholds.Remove)  // defaultRemoveThreshold
		assert.Equal(t, 400, *cfg.Thresholds.Update) // defaultUpdateThreshold
		assert.Equal(t, -1, *cfg.Thresholds.Copy)    // defaultCopyThreshold
		assert.Equal(t, -1, *cfg.Thresholds.Move)    // defaultMoveThreshold
		assert.Equal(t, -1, *cfg.Thresholds.Restore) // defaultRestoreThreshold

		// Verify scrub options use defaults when omitted
		assert.Equal(t, 22, *cfg.Scrub.Plan)      // defaultScrubPlan
		assert.Equal(t, 12, *cfg.Scrub.OlderThan) // defaultScrubOlderThan

		// Steps and notifications should be as provided
		expSteps := Steps{
			Touch: utils.Ptr(false),
			Scrub: utils.Ptr(true),
			Smart: utils.Ptr(false),
		}
		assert.Equal(t, expSteps, cfg.Steps)
		assert.Equal(t, "token", cfg.Notify.SlackToken)
		assert.Equal(t, "#channel", cfg.Notify.SlackChannel)
	})

	t.Run("Explicit zero values for thresholds and scrub should be honored", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "zeros.yml")
		zerosContent := `
snapraid_bin: "/usr/bin/snapraid"
snapraid_config: "/etc/snapraid.conf"
output_dir: "/var/lib/snapraid/output"

thresholds:
  add: 0
  remove: 0
  update: 0
  copy: 0
  move: 0
  restore: 0

steps:
  touch: true
  scrub: true
  smart: true

scrub:
  plan: 0
  older_than: 0

notifications:
  slack_token: "zero"
  slack_channel: "#zero"
`
		err := os.WriteFile(path, []byte(zerosContent), 0o600)
		assert.NoError(t, err)

		cfg, err := LoadConfig(path)
		cfg.ApplyDefaults()
		assert.NoError(t, err)

		// Verify all thresholds are exactly zero (not defaulted)
		expThresh := Thresholds{
			Add:     utils.Ptr(0),
			Remove:  utils.Ptr(0),
			Update:  utils.Ptr(0),
			Copy:    utils.Ptr(0),
			Move:    utils.Ptr(0),
			Restore: utils.Ptr(0),
		}
		assert.Equal(t, expThresh.Add, cfg.Thresholds.Add)
		assert.Equal(t, expThresh.Remove, cfg.Thresholds.Remove)
		assert.Equal(t, expThresh.Update, cfg.Thresholds.Update)
		assert.Equal(t, expThresh.Copy, cfg.Thresholds.Copy)
		assert.Equal(t, expThresh.Move, cfg.Thresholds.Move)
		assert.Equal(t, expThresh.Restore, cfg.Thresholds.Restore)

		// Verify scrub options are exactly zero (not defaulted)
		assert.Equal(t, 0, *cfg.Scrub.Plan)
		assert.Equal(t, 0, *cfg.Scrub.OlderThan)
	})
}

func TestApplyDefaults(t *testing.T) {
	t.Parallel()

	t.Run("ApplyDefaults should set default values for nil pointers", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "zeros.yml")
		zerosContent := `
snapraid_bin: "/usr/bin/snapraid"
snapraid_config: "/etc/snapraid.conf"
output_dir: "/var/lib/snapraid/output"

notifications:
  slack_token: "zero"
  slack_channel: "#zero"
`
		err := os.WriteFile(path, []byte(zerosContent), 0o600)
		assert.NoError(t, err)

		cfg, err := LoadConfig(path)
		cfg.ApplyDefaults()
		assert.NoError(t, err)

		assert.Equal(t, defaultAddThreshold, *cfg.Thresholds.Add)
		assert.Equal(t, defaultRemoveThreshold, *cfg.Thresholds.Remove)
		assert.Equal(t, defaultUpdateThreshold, *cfg.Thresholds.Update)
		assert.Equal(t, defaultCopyThreshold, *cfg.Thresholds.Copy)
		assert.Equal(t, defaultMoveThreshold, *cfg.Thresholds.Move)
		assert.Equal(t, defaultRestoreThreshold, *cfg.Thresholds.Restore)
		assert.Equal(t, defaultScrubPlan, *cfg.Scrub.Plan)
		assert.Equal(t, defaultScrubOlderThan, *cfg.Scrub.OlderThan)
	})
}
