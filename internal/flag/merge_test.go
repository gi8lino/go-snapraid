package flag

import (
	"testing"

	"github.com/gi8lino/go-snapraid/internal/config"
	"github.com/gi8lino/go-snapraid/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestApplyOverrides(t *testing.T) {
	t.Parallel()

	t.Run("DryRun disables all steps", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{
			Steps: config.Steps{
				Touch: utils.Ptr(true),
				Scrub: utils.Ptr(true),
				Smart: utils.Ptr(true),
			},
		}
		flags := Options{DryRun: true}
		ApplyOverrides(orig, flags)

		assert.False(t, *orig.Steps.Touch)
		assert.False(t, *orig.Steps.Scrub)
		assert.False(t, *orig.Steps.Smart)
	})

	t.Run("OutputDir override", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{OutputDir: "/original/path"}
		flags := Options{OutputDir: "/new/output"}
		ApplyOverrides(orig, flags)

		assert.Equal(t, "/new/output", orig.OutputDir)
	})

	t.Run("NoNotify clears Slack settings", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{
			Notify: config.Notify{SlackToken: "token123", SlackChannel: "#channel"},
		}
		flags := Options{NoNotify: true}
		ApplyOverrides(orig, flags)

		assert.Empty(t, orig.Notify.SlackToken)
		assert.Empty(t, orig.Notify.SlackChannel)
	})

	t.Run("CLI step toggles set all true", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{Steps: config.Steps{
			Touch: utils.Ptr(false),
			Scrub: utils.Ptr(false),
			Smart: utils.Ptr(false),
		}}
		flags := Options{Steps: StepsOptions{NoTouch: true, NoScrub: true, NoSmart: true}}
		ApplyOverrides(orig, flags)

		assert.True(t, *orig.Steps.Touch)
		assert.True(t, *orig.Steps.Scrub)
		assert.True(t, *orig.Steps.Smart)
	})

	t.Run("CLI step toggles set steps", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{Steps: config.Steps{
			Touch: utils.Ptr(false),
			Scrub: utils.Ptr(false),
			Smart: utils.Ptr(false),
		}}
		flags := Options{Steps: StepsOptions{NoTouch: true, NoScrub: true, NoSmart: false}}
		ApplyOverrides(orig, flags)

		assert.True(t, *orig.Steps.Touch)
		assert.True(t, *orig.Steps.Scrub)
		assert.False(t, *orig.Steps.Smart)
	})

	t.Run("Threshold disabling sets to -1", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{
			Thresholds: config.Thresholds{
				Add:     utils.Ptr(5),
				Remove:  utils.Ptr(10),
				Update:  utils.Ptr(15),
				Copy:    utils.Ptr(20),
				Move:    utils.Ptr(25),
				Restore: utils.Ptr(30),
			},
		}
		flags := Options{Thresholds: ThresholdOptions{
			NoAdd:     false,
			NoRemove:  true,
			NoUpdate:  false,
			NoCopy:    true,
			NoMove:    false,
			NoRestore: true,
		}}
		ApplyOverrides(orig, flags)

		assert.Equal(t, -1, *orig.Thresholds.Add)
		assert.Equal(t, 10, *orig.Thresholds.Remove)
		assert.Equal(t, -1, *orig.Thresholds.Update)
		assert.Equal(t, 20, *orig.Thresholds.Copy)
		assert.Equal(t, -1, *orig.Thresholds.Move)
		assert.Equal(t, 30, *orig.Thresholds.Restore)
	})

	t.Run("Combination: DryRun plus other flags", func(t *testing.T) {
		t.Parallel()

		orig := &config.Config{
			Steps: config.Steps{
				Touch: utils.Ptr(true),
				Scrub: utils.Ptr(true),
				Smart: utils.Ptr(true),
			},
			OutputDir: "/orig",
			Notify:    config.Notify{SlackToken: "xyz", SlackChannel: "#x"},
			Thresholds: config.Thresholds{
				Add:     utils.Ptr(3),
				Remove:  utils.Ptr(4),
				Update:  utils.Ptr(5),
				Copy:    utils.Ptr(6),
				Move:    utils.Ptr(7),
				Restore: utils.Ptr(8),
			},
		}
		flags := Options{
			DryRun:    true,
			OutputDir: "/combined",
			NoNotify:  true,
			Steps:     StepsOptions{NoTouch: false, NoScrub: true, NoSmart: false},
			Thresholds: ThresholdOptions{
				NoAdd:     false,
				NoRemove:  false,
				NoUpdate:  true,
				NoCopy:    false,
				NoMove:    true,
				NoRestore: true,
			},
		}
		ApplyOverrides(orig, flags)

		// DryRun should have disabled all steps despite CLI toggles
		assert.False(t, *orig.Steps.Touch)
		assert.False(t, *orig.Steps.Scrub)
		assert.False(t, *orig.Steps.Smart)

		// OutputDir override
		assert.Equal(t, "/combined", orig.OutputDir)

		// NoNotify clears Slack
		assert.Empty(t, orig.Notify.SlackToken)
		assert.Empty(t, orig.Notify.SlackChannel)

		// Thresholds: only Update, Move, Restore remain untouched or disabled accordingly
		assert.Equal(t, -1, *orig.Thresholds.Add)
		assert.Equal(t, -1, *orig.Thresholds.Remove)
		assert.Equal(t, 5, *orig.Thresholds.Update)
		assert.Equal(t, -1, *orig.Thresholds.Copy)
		assert.Equal(t, 7, *orig.Thresholds.Move)
		assert.Equal(t, 8, *orig.Thresholds.Restore)
	})
}
