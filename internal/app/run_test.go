package app

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/gi8lino/go-snapraid/internal/testutils"
	"github.com/stretchr/testify/assert"
)

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

		cfgPath := testutils.WriteFile(t, "snapraid_bin: invalid")

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commit123", []string{"--config", cfgPath}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid_bin not found: invalid")
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		dummyConf := testutils.WriteFile(t, "# dummy snapraid config")

		content := `printf """
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
"""`
		binPath := testutils.WriteScriptFile(t, content, 0)
		cfgPath := testutils.WriteFile(t, fmt.Sprintf(`
snapraid_bin: "%s"
snapraid_config: "%s"
`, binPath, dummyConf))

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commit123", []string{"--config", cfgPath}, &stdout)
		assert.NoError(t, err, "Run should exit without error")

		assert.Contains(t, stdout.String(), "There are differences")
	})

	t.Run("Diff failure", func(t *testing.T) {
		t.Parallel()

		dummyConf := testutils.WriteFile(t, "# dummy snapraid config")
		binPath := testutils.WriteScriptFile(t, "echo error", 1)

		cfgYAML := fmt.Sprintf("snapraid_bin: %q\nsnapraid_config: %q", binPath, dummyConf)

		cfgPath := testutils.WriteFile(t, cfgYAML)

		var stdout bytes.Buffer
		err := Run(context.Background(), "vTEST", "commitXYZ", []string{"--config", cfgPath}, &stdout)

		assert.Error(t, err, "Run should return an error when diff fails")
		assert.EqualError(t, err, "snapraid diff failed: exit status 1\nstderr:\n")
	})

	t.Run("Disabled notification", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		dummyConf := testutils.WriteFile(t, "# dummy snapraid config")
		binPath := testutils.WriteScriptFile(t, "printf \"0 equal\n\"", 0)

		// YAML config: include Slack token/channel, but we will pass --no-notify
		cfgYAML := fmt.Sprintf(`
snapraid_bin: "%s"
snapraid_config: "%s"
output_dir: "%s"

notifications:
  slack_token: "INVALID-TOKEN"
  slack_channel: "#channel"
`, binPath, dummyConf, tmpDir+"/outno")
		cfgPath := testutils.WriteFile(t, cfgYAML)

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
		assert.EqualError(t, err, "snapraid config file not found: /nonexistent/config.yml")
	})

	t.Run("Invalid config file", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		cfgYAML := ": not valid YAML"
		cfgPath := testutils.WriteFile(t, cfgYAML)

		err := Run(context.Background(), "vTEST", "commitErr2", []string{"--config", cfgPath}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid YAML: yaml: did not find expected key")
	})

	t.Run("Invalid CLI flags", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		err := Run(context.Background(), "vVER", "commitCLI", []string{"--unknown"}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --unknown")
	})

	t.Run("Missing config file", func(t *testing.T) {
		t.Parallel()

		var stdout bytes.Buffer
		err := Run(context.Background(), "vVER2", "commitCLI2", []string{}, &stdout)
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid config file not found: /etc/snapraid-runner.yml")
	})
}
