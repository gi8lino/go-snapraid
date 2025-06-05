package config

import (
	"fmt"
	"os"
)

// Validate checks that required paths and scrub options exist or are sane.
// It assumes ApplyDefaults has already been called to populate any nil pointers.
func (c Config) Validate() error {
	if c.SnapraidBin == "" {
		return fmt.Errorf("snapraid_bin must be set")
	}
	if _, err := os.Stat(c.SnapraidBin); err != nil {
		return fmt.Errorf("snapraid_bin not found: %s", c.SnapraidBin)
	}

	if c.SnapraidConfig == "" {
		return fmt.Errorf("snapraid_config must be set")
	}
	if _, err := os.Stat(c.SnapraidConfig); err != nil {
		return fmt.Errorf("snapraid_config not found: %s", c.SnapraidConfig)
	}

	// After ApplyDefaults, c.Scrub.Plan is guaranteed non‐nil.
	if c.Scrub.Plan == nil {
		return fmt.Errorf("scrub.plan must be set")
	}
	if *c.Scrub.Plan < 0 || *c.Scrub.Plan > 100 {
		return fmt.Errorf("scrub.plan must be between 0–100")
	}

	// After ApplyDefaults, c.Scrub.OlderThan is guaranteed non‐nil.
	if c.Scrub.OlderThan == nil {
		return fmt.Errorf("scrub.older_than must be set")
	}
	if *c.Scrub.OlderThan < 0 {
		return fmt.Errorf("scrub.older_than must be >= 0")
	}

	return nil
}
