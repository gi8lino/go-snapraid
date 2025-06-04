package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createScript writes a shell script at `path` with the given content and makes it executable.
func createScript(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatalf("failed to write script %s: %v", path, err)
	}
}

// writeFile creates a file at `path` with `content`.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("shows help and exits", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commit123", []string{"--help"}, &stdout)

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Usage:")
	})

	t.Run("Invalid config", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		invalidbin := "invalid"

		cfgPath := filepath.Join(tmpDir, "config.yml")
		cfgYAML := fmt.Sprintf(`
snapraid_bin: "%s"
`, invalidbin)
		writeFile(t, cfgPath, cfgYAML)

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commit123", []string{"--config", cfgPath}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "validate config: snapraid_bin not found: invalid")
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dummyConf := filepath.Join(tmpDir, "snapraid.conf")
		writeFile(t, dummyConf, "# dummy snapraid config")
		binPath := filepath.Join(tmpDir, "fake-snapraid.sh")
		createScript(t, binPath, `#!/bin/sh
printf """
Comparing...
add movies/Zoolander\ \(2001\)/Zoolander.2001.German.AC3.DL.1080p.BluRay.x265-FuN.mkv
remove movies/Zoolander\ 2\ \(2016\)/Zoolander.2.2016.German.AC3.DL.1080p.BluRay.x265-FuN.mkv
remove movies/XOXO\ \(2016\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.mkv

   21156 equal
       1 added
       2 removed
       0 updated
       0 moved
       0 copied
       0 restored
There are differences!
"""
exit 0
`)
		cfgPath := filepath.Join(tmpDir, "config.yml")
		cfgYAML := fmt.Sprintf(`
snapraid_bin: "%s"
config_file: "%s"
`, binPath, dummyConf)
		writeFile(t, cfgPath, cfgYAML)

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commit123", []string{"--config", cfgPath}, &stdout)
		assert.NoError(t, err, "Run should exit without error")

		assert.Contains(t, stdout.String(), "There are differences")
	})

	t.Run("Diff failure", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		dummyConf := filepath.Join(tmpDir, "snapraid.conf")
		writeFile(t, dummyConf, "# dummy snapraid config")

		binPath := filepath.Join(tmpDir, "fake-snapraid-fail.sh")
		createScript(t, binPath, `#!/bin/sh
exit 1
`)

		cfgPath := filepath.Join(tmpDir, "config.yml")
		cfgYAML := fmt.Sprintf(`
snapraid_bin: "%s"
config_file: "%s"
`, binPath, dummyConf)

		writeFile(t, cfgPath, cfgYAML)

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commitXYZ", []string{"--config", cfgPath}, &stdout)

		assert.Error(t, err, "Run should return an error when diff fails")
		assert.EqualError(t, err, "snapraid diff failed: exit status 1: Running diff\n")
	})

	t.Run("Disabled notification", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		dummyConf := filepath.Join(tmpDir, "snapraid.conf")
		writeFile(t, dummyConf, "# dummy config")

		binPath := filepath.Join(tmpDir, "fake-snapraid.sh")
		createScript(t, binPath, `#!/bin/sh
printf "0 equal\n"
exit 0
`)

		// YAML config: include Slack token/channel, but we will pass --no-notify
		cfgPath := filepath.Join(tmpDir, "config.yml")
		cfgYAML := fmt.Sprintf(`
snapraid_bin: "%s"
config_file: "%s"
output_dir: "%s"

notifications:
  slack_token: "INVALID-TOKEN"
  slack_channel: "#channel"
`, binPath, dummyConf, tmpDir+"/outno")
		writeFile(t, cfgPath, cfgYAML)

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commitNO", []string{"--config", cfgPath, "--no-notify"}, &stdout)
		assert.NoError(t, err, "Run should succeed even though Slack token is invalid, because no-notify is set")

		assert.NoError(t, err)
		assert.NotContains(t, stdout.String(), "Slack notification sent")
	})

	t.Run("Missing config file", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commitErr", []string{"--config", "/nonexistent/config.yml"}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "validate flags: snapraid config file not found: /nonexistent/config.yml")
	})

	t.Run("Invalid config file", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "bad.yml")
		writeFile(t, cfgPath, ": not valid YAML")
		err := Run(context.Background(), "vTEST", "commitErr2", []string{"--config", cfgPath}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "load config: invalid YAML: yaml: did not find expected key")
	})

	t.Run("Invalid CLI flags", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		err := Run(context.Background(), "vVER", "commitCLI", []string{"--unknown"}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "parse flags: unknown flag: --unknown")
	})

	t.Run("Missing config file", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		err := Run(context.Background(), "vVER2", "commitCLI2", []string{}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "validate flags: snapraid config file not found: /etc/snapraid-runner.yml")
	})
}
