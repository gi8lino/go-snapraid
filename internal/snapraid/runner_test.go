package snapraid

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeExec allows simulating different Snapraid behaviors.
type fakeExec struct {
	DiffLines []string // DiffLines to return from Diff()
	DiffErr   error    // DiffErr simulates an error from Diff()
	TouchErr  error    // TouchErr simulates an error from Touch()
	SyncErr   error    // SyncErr simulates an error from Sync()
	ScrubErr  error    // ScrubErr simulates an error from Scrub()
	SmartErr  error    // SmartErr simulates an error from Smart()

	// Counters to verify calls
	TouchCount int
	DiffCount  int
	SyncCount  int
	ScrubCount int
	SmartCount int
}

func (f *fakeExec) Touch() error {
	f.TouchCount++
	return f.TouchErr
}

func (f *fakeExec) Diff() ([]string, error) {
	f.DiffCount++
	return f.DiffLines, f.DiffErr
}

func (f *fakeExec) Sync() error {
	f.SyncCount++
	return f.SyncErr
}

func (f *fakeExec) Scrub() error {
	f.ScrubCount++
	return f.ScrubErr
}

func (f *fakeExec) Smart() error {
	f.SmartCount++
	return f.SmartErr
}

func TestNewRunner(t *testing.T) {
	t.Parallel()

	const (
		configPath    = "/etc/snapraid.conf"
		binaryPath    = "/usr/bin/snapraid"
		outputPath    = "/var/log/snapraid"
		scrubPlanVal  = 25
		scrubOlderVal = 7
	)

	steps := Steps{Touch: true, Scrub: false, Smart: true}
	thresholds := Thresholds{Add: 10, Remove: 20, Update: 30, Move: 40, Copy: 50, Restore: 60}
	dryRun := true
	logger := slog.New(slog.NewTextHandler(nil, nil))

	r := NewRunner(
		configPath,
		binaryPath,
		outputPath,
		steps,
		thresholds,
		scrubPlanVal,
		scrubOlderVal,
		dryRun,
		logger,
	)

	// Runner fields
	assert.Equal(t, steps, r.Steps, "Steps should match")
	assert.Equal(t, thresholds, r.Thresholds, "Thresholds should match")
	assert.Equal(t, dryRun, r.DryRun, "DryRun should match")
	assert.Equal(t, logger, r.Logger, "Logger should match")
	assert.False(t, r.Timestamp.IsZero(), "Timestamp should be set to a non-zero value")

	// exec should be a *DefaultExecutor with matching fields
	de, ok := r.exec.(*DefaultExecutor)
	assert.True(t, ok, "exec should be a *DefaultExecutor")

	assert.Equal(t, configPath, de.configPath, "DefaultExecutor.configPath should match")
	assert.Equal(t, binaryPath, de.binaryPath, "DefaultExecutor.binaryPath should match")
	assert.Equal(t, scrubPlanVal, de.scrubPlan, "DefaultExecutor.scrubPlan should match")
	assert.Equal(t, scrubOlderVal, de.scrubOlder, "DefaultExecutor.scrubOlder should match")
	assert.Equal(t, logger, de.logger, "DefaultExecutor.logger should match")

	// Timestamp sanity check: within a couple seconds of now
	now := time.Now().UTC()
	diff := now.Sub(r.Timestamp)
	if diff < 0 {
		diff = -diff
	}
	assert.Less(t, diff, 2*time.Second, "Timestamp should be very recent")
}

func TestRunner(t *testing.T) {
	t.Parallel()

	diff := `Comparing...
add movies/Gladiator\ \(2000\)/Gladiator.2000.German.AC3.DL.1080p.BluRay.x264.FuN.mk
add movies/Interstellar\ \(2014\)/Interstellar.2014.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
add movies/Judge\ Dredd\ \(1995\)/Judge.Dredd.1995.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
add movies/Predator\ \(1987\)/Predator.1987.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
add movies/Zoolander\ \(2001\)/Zoolander.2001.German.AC3.DL.1080p.BluRay.x265-FuN.mkv
copy movies/Mile\ 22\ \(2002\)/Mile.22.2002.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
move movies/The\ Shawshank\ Redemption\ \(1994\)/The.Shawshank.Redemption.1994.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
remove movies/XOXO\ \(2016\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.mkv
remove movies/Zoolander\ 2\ \(2016\)/Zoolander.2.2016.German.AC3.DL.1080p.BluRay.x265-FuN.mkv
restore movies/Crazy, Stupid, Love\. A\. Piano\. \(2009\)/Crazy.Stupid.Love.A.Piano.2009.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
update movies/The\ Matrix\ \(1999\)/The.Matrix.1999.German.AC3.DL.1080p.BluRay.x264.FuN.mkv

   21156 equal
       5 added
       2 removed
       1 updated
       1 moved
       1 copied
       1 restored
There are differences!
`
	diffLines := strings.Split(diff, "\n")

	t.Run("No changes short circuit", func(t *testing.T) {
		t.Parallel()

		// Setup a fake executor that returns no changes ("0 equal")
		f := &fakeExec{
			DiffLines: []string{"0 equal"},
		}
		r := &Runner{
			Steps:      Steps{Touch: true, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		// Should have run Diff once, not Sync/Scrub/Smart
		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.Equal(t, 0, f.SyncCount, "Sync should not be called")
		assert.Equal(t, 0, f.ScrubCount, "Scrub should not be called")
		assert.Equal(t, 0, f.SmartCount, "Smart should not be called")
		assert.False(t, result.HasChanges(), "Result should report no changes")
	})

	t.Run("No threshold violation", func(t *testing.T) {
		t.Parallel()

		// Diff returns 5 added; threshold.Add is 3 → violation
		f := &fakeExec{DiffLines: diffLines}
		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: false, Smart: false},
			Thresholds: Thresholds{Add: -1, Remove: -1, Update: -1, Move: -1, Copy: -1, Restore: -1},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		// Should have run Diff, then returned error before Sync
		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.True(t, result.HasChanges(), "Result.HasChanges should be true")
		assert.NoError(t, result.Error)
		assert.Equal(t, 5, len(result.Result.Added), fmt.Sprintf("Result.Added should be 5, got %d", len(result.Result.Added)))
		assert.Equal(t, 2, len(result.Result.Removed), fmt.Sprintf("Result.Removed should be 2, got %d", len(result.Result.Removed)))
		assert.Equal(t, 1, len(result.Result.Updated), fmt.Sprintf("Result.Updated should be 1, got %d", len(result.Result.Updated)))
		assert.Equal(t, 1, len(result.Result.Moved), fmt.Sprintf("Result.Moved should be 1, got %d", len(result.Result.Moved)))
		assert.Equal(t, 1, len(result.Result.Copied), fmt.Sprintf("Result.Copied should be 1, got %d", len(result.Result.Copied)))
		assert.Equal(t, 1, len(result.Result.Restored), fmt.Sprintf("Result.Restored should be 1, got %d", len(result.Result.Restored)))
	})

	t.Run("Threshold violation", func(t *testing.T) {
		t.Parallel()

		// Diff returns 5 added; threshold.Add is 3 → violation
		f := &fakeExec{DiffLines: diffLines}
		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: false, Smart: false},
			Thresholds: Thresholds{Add: 3, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		// Should have run Diff, then returned error before Sync
		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.True(t, result.HasChanges(), "Result.HasChanges should be true")
		assert.Error(t, result.Error, "Error should be set for threshold violation")
		assert.Equal(t, 0, f.SyncCount, "Sync should not be called")
	})

	t.Run("Dry run skips sync scrub smart", func(t *testing.T) {
		t.Parallel()

		f := &fakeExec{DiffLines: diffLines}
		r := &Runner{
			Steps:      Steps{Touch: true, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     true,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		// Touch should run (because DryRun=false? Actually code: Touch runs only if !DryRun, so DryRun skips Touch too)
		assert.Equal(t, 0, f.TouchCount, "Touch should be skipped on DryRun")
		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.Equal(t, 0, f.SyncCount, "Sync should be skipped on DryRun")
		assert.Equal(t, 0, f.ScrubCount, "Scrub should be skipped on DryRun")
		assert.Equal(t, 0, f.SmartCount, "Smart should be skipped on DryRun")
		assert.NoError(t, result.Error, "No error on DryRun")
	})

	t.Run("Run touch", func(t *testing.T) {
		t.Parallel()

		f := &fakeExec{
			DiffLines:  []string{"add movies/Gladiator\\ \\(2000\\)/Gladiator.2000.German.AC3.DL.1080p.BluRay.x264.FuN.mk"},
			TouchCount: 1,
		}
		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		assert.Equal(t, 1, f.TouchCount, "Touch should be called once")
		assert.Equal(t, 1, len(result.Result.Added), fmt.Sprintf("Result.Added should be 1, got %d", len(result.Result.Added)))
	})

	t.Run("Touch error stops workflow", func(t *testing.T) {
		t.Parallel()

		f := &fakeExec{
			DiffLines: diffLines,
			TouchErr:  errors.New("touch failed"),
		}

		r := &Runner{
			Steps:      Steps{Touch: true, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		assert.Empty(t, f.DiffCount, "Diff should be called once")
		assert.Empty(t, f.SyncCount, "Sync should be called once")
		assert.Equal(t, 1, f.TouchCount, "Touch should be called once")
		assert.Error(t, result.Error, "Error should be set for touch failure")
		assert.EqualError(t, result.Error, "touch failed")
	})

	t.Run("Diff error stops workflow", func(t *testing.T) {
		t.Parallel()

		f := &fakeExec{DiffErr: errors.New("diff failed")}

		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.Error(t, result.Error, "Error should be set for diff failure")
		assert.EqualError(t, result.Error, "diff failed")
	})

	t.Run("Sync error stops workflow", func(t *testing.T) {
		t.Parallel()

		f := &fakeExec{
			DiffLines: diffLines,
			SyncErr:   errors.New("sync failed"),
		}

		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.Equal(t, 1, f.SyncCount, "Sync should be called once")
		assert.Error(t, result.Error, "Error should be set for sync failure")
		assert.EqualError(t, result.Error, "sync failed")
		assert.Equal(t, 0, f.ScrubCount, "Scrub should not be called after Sync error")
	})

	t.Run("Scrub error stops workflow", func(t *testing.T) {
		t.Parallel()

		f := &fakeExec{
			DiffLines: diffLines,
			ScrubErr:  errors.New("scrub failed"),
		}
		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		// Should call Diff, Sync, then Scrub, then stop
		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.Equal(t, 1, f.SyncCount, "Sync should be called once")
		assert.Equal(t, 1, f.ScrubCount, "Scrub should be called once")
		assert.Error(t, result.Error, "Error should be set for scrub failure")
		assert.EqualError(t, result.Error, "scrub failed")
		assert.Equal(t, 0, f.SmartCount, "Smart should not be called after Scrub error")
	})

	t.Run("Smart error stops workflow", func(t *testing.T) {
		t.Parallel()

		// Diff returns 1 added; Sync & Scrub pass; Smart returns an error
		f := &fakeExec{
			DiffLines: diffLines,
			SmartErr:  errors.New("smart failed"),
		}
		r := &Runner{
			Steps:      Steps{Touch: false, Scrub: true, Smart: true},
			Thresholds: Thresholds{Add: 10, Remove: 10, Update: 10, Move: 10, Copy: 10, Restore: 10},
			DryRun:     false,
			exec:       f,
			Logger:     nil,
			Timestamp:  time.Now(),
		}

		result := r.Run()

		// Should call Diff, Sync, Scrub, then Smart, then stop
		assert.Equal(t, 1, f.DiffCount, "Diff should be called once")
		assert.Equal(t, 1, f.SyncCount, "Sync should be called once")
		assert.Equal(t, 1, f.ScrubCount, "Scrub should be called once")
		assert.Equal(t, 1, f.SmartCount, "Smart should be called once")
		assert.Error(t, result.Error, "Error should be set for smart failure")
		assert.EqualError(t, result.Error, "smart failed")
	})
}
