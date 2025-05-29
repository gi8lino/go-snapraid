package config

import (
	"fmt"
	"os"
)

// Validate checks that required paths and values exist or are sane.
func (c Config) Validate() error {
	if c.SnapraidBin == "" {
		return fmt.Errorf("snapraid_bin must be set")
	}
	if _, err := os.Stat(c.SnapraidBin); err != nil {
		return fmt.Errorf("snapraid_bin not found: %s", c.SnapraidBin)
	}

	if c.ConfigFile == "" {
		return fmt.Errorf("config_file must be set")
	}
	if _, err := os.Stat(c.ConfigFile); err != nil {
		return fmt.Errorf("config_file not found: %s", c.ConfigFile)
	}

	if c.Scrub.Plan < 0 || c.Scrub.Plan > 100 {
		return fmt.Errorf("scrub.plan must be between 0–100")
	}

	if c.Scrub.OlderThan < 0 {
		return fmt.Errorf("scrub.older_than must be >= 0")
	}

	return nil
}
