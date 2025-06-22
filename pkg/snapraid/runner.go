package snapraid

import (
	"log/slog"
	"time"
)

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
	now := time.Now()

	r.Timestamp = now
	runResult := RunResult{
		Timestamp: r.Timestamp.Format(time.RFC3339),
	}

	// Always record the total time, even if there is an error
	start := now
	defer func() {
		runResult.Timings.Total = time.Since(start)
	}()

	// TOUCH - makes only sense if it is not a dry run
	if r.Steps.Touch && !r.DryRun {
		if err := runStep(r.exec.Touch, func(d time.Duration) { runResult.Timings.Touch = d }); err != nil {
			runResult.Error = err
			return runResult
		}
	}

	// DIFF
	t1 := time.Now()
	diffLines, err := r.exec.Diff()
	runResult.Timings.Diff = time.Since(t1)
	if err != nil {
		runResult.Error = err
		return runResult
	}

	diffResult := parseDiff(diffLines)
	runResult.Result = diffResult

	// DRY RUN? skip Sync/Scrub/Smart if true
	if r.DryRun {
		return runResult
	}

	if diffResult.HasChanges() {
		// THRESHOLD CHECK
		if err := validateThresholds(diffResult, r.Thresholds); err != nil {
			runResult.Error = err
			return runResult
		}

		// SYNC
		if err := runStep(r.exec.Sync, func(d time.Duration) { runResult.Timings.Sync = d }); err != nil {
			runResult.Error = err
			return runResult
		}
	}

	// SCRUB
	if r.Steps.Scrub {
		if err := runStep(r.exec.Scrub, func(d time.Duration) { runResult.Timings.Scrub = d }); err != nil {
			runResult.Error = err
			return runResult
		}
	}

	// SMART
	if r.Steps.Smart {
		if err := runStep(r.exec.Smart, func(d time.Duration) { runResult.Timings.Smart = d }); err != nil {
			runResult.Error = err
			return runResult
		}
	}

	return runResult
}
