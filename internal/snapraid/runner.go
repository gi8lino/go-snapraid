package snapraid

import (
	"fmt"
	"log/slog"
	"time"
)

// Runner coordinates a full SnapRAID workflow based on its configuration.
type Runner struct {
	Steps      Steps      // which subcommands to run: Touch, Scrub, Smart
	Thresholds Thresholds // numeric limits per change type
	DryRun     bool       // if true, skip sync/scrub/smart

	Logger    *slog.Logger // structured logger for real‐time output
	Timestamp time.Time    // UTC time when Runner was created

	exec Snapraid // performs Touch, Diff, Sync, Scrub, Smart
}

// NewRunner constructs a Runner with the given parameters. It installs a DefaultExecutor by default.
func NewRunner(
	configPath, binaryPath, outputPath string,
	steps Steps,
	thresholds Thresholds,
	scrubPlan, scrubOlder int,
	dryRun bool,
	logger *slog.Logger,
) *Runner {
	r := &Runner{
		Steps:      steps,
		Thresholds: thresholds,
		DryRun:     dryRun,
		Logger:     logger,
		Timestamp:  time.Now().UTC(),
	}
	r.exec = &DefaultExecutor{
		configPath: configPath,
		binaryPath: binaryPath,
		scrubPlan:  scrubPlan,
		scrubOlder: scrubOlder,
		logger:     logger,
	}
	return r
}

// Run executes the SnapRAID workflow in this order: Touch → Diff → (Sync → Scrub → Smart).
// It returns a RunResult containing timestamps, parsed diff, per‐step durations, and any error.
func (r *Runner) Run() RunResult {
	var result DiffResult
	var timings RunTimings
	start := time.Now()

	// TOUCH - makes only sense if we're not doing a dry run
	if r.Steps.Touch && !r.DryRun {
		t0 := time.Now()
		if err := r.exec.Touch(); err != nil {
			return RunResult{Error: err}
		}
		timings.Touch = time.Since(t0)
	}

	// DIFF
	t1 := time.Now()
	diffLines, err := r.exec.Diff()
	timings.Diff = time.Since(t1)
	if err != nil {
		fmt.Println(err.Error())
		return RunResult{Error: err, Timings: timings}
	}

	result = parseDiff(diffLines)
	if !result.HasChanges() {
		timings.Total = time.Since(start)
		return RunResult{
			Timestamp: r.Timestamp.Format(time.RFC3339),
			Result:    result,
			Timings:   timings,
		}
	}

	// THRESHOLD CHECK
	if err := validateThresholds(result, r.Thresholds); err != nil {
		timings.Total = time.Since(start)
		return RunResult{
			Timestamp: r.Timestamp.Format(time.RFC3339),
			Result:    result,
			Timings:   timings,
			Error:     err,
		}
	}

	// DRY RUN? skip Sync/Scrub/Smart if true
	if r.DryRun {
		timings.Total = time.Since(start)
		return RunResult{
			Timestamp: r.Timestamp.Format(time.RFC3339),
			Result:    result,
			Timings:   timings,
		}
	}

	// SYNC
	t2 := time.Now()
	if err := r.exec.Sync(); err != nil {
		timings.Sync = time.Since(t2)
		timings.Total = time.Since(start)
		return RunResult{Result: result, Timings: timings, Error: err}
	}
	timings.Sync = time.Since(t2)

	// SCRUB
	if r.Steps.Scrub {
		t3 := time.Now()
		if err := r.exec.Scrub(); err != nil {
			timings.Scrub = time.Since(t3)
			timings.Total = time.Since(start)
			return RunResult{Result: result, Timings: timings, Error: err}
		}
		timings.Scrub = time.Since(t3)
	}

	// SMART
	if r.Steps.Smart {
		t4 := time.Now()
		if err := r.exec.Smart(); err != nil {
			timings.Smart = time.Since(t4)
			timings.Total = time.Since(start)
			return RunResult{Result: result, Timings: timings, Error: err}
		}
		timings.Smart = time.Since(t4)
	}

	timings.Total = time.Since(start)
	return RunResult{
		Timestamp: r.Timestamp.Format(time.RFC3339),
		Result:    result,
		Timings:   timings,
	}
}

// Steps defines which SnapRAID subcommands to run.
type Steps struct {
	Touch bool // Touch enables the "snapraid touch" step.
	Scrub bool // Scrub enables the "snapraid scrub" step.
	Smart bool // Smart enables the "snapraid smart" step.
}

// Thresholds defines numeric limits on detected file changes before blocking sync.
type Thresholds struct {
	Add     int // Add is the maximum number of added files allowed. –1 disables.
	Remove  int // Remove is the maximum number of removed files allowed. –1 disables.
	Update  int // Update is the maximum number of updated files allowed. –1 disables.
	Move    int // Move is the maximum number of moved files allowed. –1 disables.
	Copy    int // Copy is the maximum number of copied files allowed. –1 disables.
	Restore int // Restore is the maximum number of restored files allowed. –1 disables.
}

// RunResult holds the summary of a completed run.
type RunResult struct {
	Timestamp string     `json:"timestamp"`       // RFC3339 timestamp when run started
	Result    DiffResult `json:"result"`          // parsed diff summary + file lists
	Timings   RunTimings `json:"timings"`         // per-step durations + total
	Error     error      `json:"error,omitempty"` // any error that occurred
}

// HasChanges returns true if any files were added/removed/updated/moved/copied/restored.
func (r RunResult) HasChanges() bool { return r.Result.HasChanges() }

// RunTimings captures the duration of each subcommand and the total.
type RunTimings struct {
	Touch time.Duration `json:"touch"`
	Diff  time.Duration `json:"diff"`
	Sync  time.Duration `json:"sync"`
	Scrub time.Duration `json:"scrub"`
	Smart time.Duration `json:"smart"`
	Total time.Duration `json:"total"`
}
