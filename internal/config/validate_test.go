package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gi8lino/go-snapraid/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("SnapraidBin empty returns error", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			SnapraidBin:    "",
			SnapraidConfig: "/some/path",
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(50),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid_bin must be set")
	})

	t.Run("SnapraidBin non-existent file returns error", func(t *testing.T) {
		t.Parallel()

		nonexistent := filepath.Join(t.TempDir(), "no-such-binary")
		cfg := Config{
			SnapraidBin:    nonexistent,
			SnapraidConfig: "/some/path",
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(50),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid_bin not found: "+nonexistent)
	})

	t.Run("ConfigFile empty returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))

		cfg := Config{
			SnapraidBin:    binPath,
			SnapraidConfig: "",
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(50),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid_config must be set")
	})

	t.Run("ConfigFile non-existent returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))

		nonexistentCfg := filepath.Join(tmpDir, "no-such-config.yml")
		cfg := Config{
			SnapraidBin:    binPath,
			SnapraidConfig: nonexistentCfg,
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(50),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "snapraid_config not found: "+nonexistentCfg)
	})

	t.Run("Scrub.Plan negative returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		cfgPath := filepath.Join(tmpDir, "snapraid.conf")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))
		assert.NoError(t, os.WriteFile(cfgPath, []byte{}, 0o600))

		cfg := Config{
			SnapraidBin:    binPath,
			SnapraidConfig: cfgPath,
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(-5),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "scrub.plan must be between 0–100")
	})

	t.Run("Scrub.Plan greater than 100 returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		cfgPath := filepath.Join(tmpDir, "snapraid.conf")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))
		assert.NoError(t, os.WriteFile(cfgPath, []byte{}, 0o600))

		cfg := Config{
			SnapraidBin:    binPath,
			SnapraidConfig: cfgPath,
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(150),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "scrub.plan must be between 0–100")
	})

	t.Run("Scrub.Plan is nil error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		cfgPath := filepath.Join(tmpDir, "snapraid.conf")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))
		assert.NoError(t, os.WriteFile(cfgPath, []byte{}, 0o600))

		path := filepath.Join(tmpDir, "config.yml")
		content := fmt.Sprintf("snapraid_bin: %s\nsnapraid_config: %q", binPath, cfgPath)
		err := os.WriteFile(path, []byte(content), 0o600)
		assert.NoError(t, err)

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)

		err = cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "scrub.plan must be set")
	})

	t.Run("Scrub.OlderThan negative returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		cfgPath := filepath.Join(tmpDir, "snapraid.conf")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))
		assert.NoError(t, os.WriteFile(cfgPath, []byte{}, 0o600))

		cfg := Config{
			SnapraidBin:    binPath,
			SnapraidConfig: cfgPath,
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(50),
				OlderThan: utils.Ptr(-3),
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "scrub.older_than must be >= 0")
	})

	t.Run("Scrub.OlderThan is nil error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		binPath := filepath.Join(tmpDir, "snapraid")
		cfgPath := filepath.Join(tmpDir, "snapraid.conf")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))
		assert.NoError(t, os.WriteFile(cfgPath, []byte{}, 0o600))

		path := filepath.Join(tmpDir, "config.yml")

		content := fmt.Sprintf(`
snapraid_bin: %q
snapraid_config: %q
scrub:
  plan: 50
`, binPath, cfgPath)

		err := os.WriteFile(path, []byte(content), 0o600)
		assert.NoError(t, err)

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)

		err = cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "scrub.older_than must be set")
	})

	t.Run("Valid Config returns no error", func(t *testing.T) {
		t.Parallel()

		// Create dummy files for binary and config
		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "snapraid")
		cfgPath := filepath.Join(tmpDir, "snapraid.conf")
		assert.NoError(t, os.WriteFile(binPath, []byte{}, 0o600))
		assert.NoError(t, os.WriteFile(cfgPath, []byte{}, 0o600))

		cfg := Config{
			SnapraidBin:    binPath,
			SnapraidConfig: cfgPath,
			Scrub: ScrubOptions{
				Plan:      utils.Ptr(50),
				OlderThan: utils.Ptr(10),
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
