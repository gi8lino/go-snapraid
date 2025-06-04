package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWantsSlackNotification(t *testing.T) {
	t.Parallel()

	// Helper to build a minimal Config with given token and channel
	makeCfg := func(token, channel string) Config {
		return Config{
			Notify: Notify{
				SlackToken:   token,
				SlackChannel: channel,
			},
		}
	}

	t.Run("noNotify suppresses even if token+channel are set", func(t *testing.T) {
		t.Parallel()
		cfg := makeCfg("xoxb-abc", "#channel")
		assert.False(t, cfg.WantsSlackNotification(true))
	})

	t.Run("missing token suppresses notification", func(t *testing.T) {
		t.Parallel()
		cfg := makeCfg("", "#channel")
		assert.False(t, cfg.WantsSlackNotification(false))
	})

	t.Run("missing channel suppresses notification", func(t *testing.T) {
		t.Parallel()
		cfg := makeCfg("xoxb-abc", "")
		assert.False(t, cfg.WantsSlackNotification(false))
	})

	t.Run("both token and channel set, noNotify false → enabled", func(t *testing.T) {
		t.Parallel()
		cfg := makeCfg("xoxb-abc", "#channel")
		assert.True(t, cfg.WantsSlackNotification(false))
	})

	t.Run("both token and channel empty, noNotify false → disabled", func(t *testing.T) {
		t.Parallel()
		cfg := makeCfg("", "")
		assert.False(t, cfg.WantsSlackNotification(false))
	})
}
