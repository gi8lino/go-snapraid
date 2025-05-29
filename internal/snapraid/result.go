package snapraid

import "time"

// RunTimings captures timing durations for each SnapRAID step.
type RunTimings struct {
	Touch time.Duration `json:"touch"`
	Diff  time.Duration `json:"diff"`
	Sync  time.Duration `json:"sync"`
	Scrub time.Duration `json:"scrub"`
	Smart time.Duration `json:"smart"`
	Total time.Duration `json:"total"`
}

// RunResult represents a completed SnapRAID run and its outcome.
type RunResult struct {
	Timestamp string     `json:"timestamp"`
	Result    DiffResult `json:"result"`
	Timings   RunTimings `json:"timings"`
	Error     error      `json:"error,omitempty"`
}
