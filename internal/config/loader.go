package config

import (
	"fmt"
	"os"

	"github.com/gi8lino/go-snapraid/internal/utils"
	"gopkg.in/yaml.v3"
)

const (
	defaultAddThreshold     = -1  // no limit on added files
	defaultRemoveThreshold  = 80  // default max removed files
	defaultUpdateThreshold  = 400 // default max updated files
	defaultCopyThreshold    = -1  // no limit on copied files
	defaultMoveThreshold    = -1  // no limit on moved files
	defaultRestoreThreshold = -1  // no limit on restored files
	defaultScrubPlan        = 22  // default scrub plan percentage
	defaultScrubOlderThan   = 12  // default scrub older‐than days
)

// LoadConfig reads the given file, parses it into a Config struct, applies defaults, and returns it.
func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid YAML: %w", err)
	}

	return cfg, nil
}

// ApplyDefaults sets default values for any nil pointer fields in Config.
func (c *Config) ApplyDefaults() {
	// Thresholds: if pointer is nil → assign default; if non‐nil, leave as-is.
	if c.Thresholds.Add == nil {
		c.Thresholds.Add = utils.Ptr(defaultAddThreshold)
	}
	if c.Thresholds.Remove == nil {
		c.Thresholds.Remove = utils.Ptr(defaultRemoveThreshold)
	}
	if c.Thresholds.Update == nil {
		c.Thresholds.Update = utils.Ptr(defaultUpdateThreshold)
	}
	if c.Thresholds.Copy == nil {
		c.Thresholds.Copy = utils.Ptr(defaultCopyThreshold)
	}
	if c.Thresholds.Move == nil {
		c.Thresholds.Move = utils.Ptr(defaultMoveThreshold)
	}
	if c.Thresholds.Restore == nil {
		c.Thresholds.Restore = utils.Ptr(defaultRestoreThreshold)
	}

	// ScrubOptions: if pointer is nil → assign default; otherwise honor user value.
	if c.Scrub.Plan == nil {
		c.Scrub.Plan = utils.Ptr(defaultScrubPlan)
	}
	if c.Scrub.OlderThan == nil {
		c.Scrub.OlderThan = utils.Ptr(defaultScrubOlderThan)
	}

	// Steps: if pointer is nil → assign default false; otherwise honor user value.
	if c.Steps.Touch == nil {
		c.Steps.Touch = utils.Ptr(false)
	}
	if c.Steps.Scrub == nil {
		c.Steps.Scrub = utils.Ptr(false)
	}
	if c.Steps.Smart == nil {
		c.Steps.Smart = utils.Ptr(false)
	}
}
