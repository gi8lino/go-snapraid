package snapraid

import "time"

// Runner coordinates the full SnapRAID workflow including steps and thresholds.
type Runner struct {
	ConfigFile  string
	SnapraidBin string
	OutputDir   string
	Steps       Steps
	Thresholds  Thresholds
	ScrubPlan   int
	ScrubOlder  int
	DryRun      bool

	Timestamp time.Time
}

// New creates a new SnapRAID Runner with required values.
func New(configFile, snapraidBin string, options ...Option) *Runner {
	r := &Runner{
		ConfigFile:  configFile,
		SnapraidBin: snapraidBin,
		Timestamp:   time.Now().UTC(),
		Steps: Steps{
			Touch: false,
			Scrub: false,
			Smart: false,
		},
		Thresholds: Thresholds{
			Add:     -1, // infinite
			Remove:  80,
			Update:  400,
			Move:    -1, // infinite
			Copy:    -1, // infinite
			Restore: -1, // infinite
		},
		ScrubPlan:  22,
		ScrubOlder: 12,
	}
	for _, opt := range options {
		opt(r)
	}
	return r
}

// Run executes the SnapRAID workflow and returns result and timing data.
func (r *Runner) Run() RunResult {
	var result DiffResult
	var timings RunTimings
	start := time.Now()

	if !r.DryRun && r.Steps.Touch {
		t0 := time.Now()
		if err := r.runTouch(); err != nil {
			return RunResult{Error: err}
		}
		timings.Touch = time.Since(t0)
	}

	t1 := time.Now()
	diffLines, err := r.runDiff()
	timings.Diff = time.Since(t1)
	if err != nil {
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

	if err := validateThresholds(result, r.Thresholds); err != nil {
		timings.Total = time.Since(start)
		return RunResult{
			Timestamp: r.Timestamp.Format(time.RFC3339),
			Result:    result,
			Timings:   timings,
			Error:     err,
		}
	}

	if !r.DryRun {
		t2 := time.Now()
		if err := r.runSync(); err != nil {
			timings.Sync = time.Since(t2)
			timings.Total = time.Since(start)
			return RunResult{Result: result, Timings: timings, Error: err}
		}
		timings.Sync = time.Since(t2)

		if r.Steps.Scrub {
			t3 := time.Now()
			if err := r.runScrub(); err != nil {
				timings.Scrub = time.Since(t3)
				timings.Total = time.Since(start)
				return RunResult{Result: result, Timings: timings, Error: err}
			}
			timings.Scrub = time.Since(t3)
		}

		if r.Steps.Smart {
			t4 := time.Now()
			if err := r.runSmart(); err != nil {
				timings.Smart = time.Since(t4)
				timings.Total = time.Since(start)
				return RunResult{Result: result, Timings: timings, Error: err}
			}
			timings.Smart = time.Since(t4)
		}
	}

	timings.Total = time.Since(start)
	return RunResult{
		Timestamp: r.Timestamp.Format(time.RFC3339),
		Result:    result,
		Timings:   timings,
	}
}
