package snapraid

import "io"

// Steps defines which SnapRAID steps to run.
type Steps struct {
	Touch bool
	Scrub bool
	Smart bool
}

// Thresholds defines limits on detected changes.
type Thresholds struct {
	Add     int
	Remove  int
	Update  int
	Move    int
	Copy    int
	Restore int
}

// Option configures a Runner instance.
type Option func(*Runner)

// WithWriter sets the output writer for the SnapRAID runner.
func WithWriter(w io.Writer) Option {
	return func(r *Runner) {
		r.Output = w
	}
}

// WithDryRun enables dry-run mode.
func WithDryRun() Option {
	return func(r *Runner) { r.DryRun = true }
}

// WithOutputDir sets the output directory for JSON results.
func WithOutputDir(dir string) Option {
	return func(r *Runner) { r.OutputDir = dir }
}

// WithSteps sets the enabled SnapRAID steps.
func WithSteps(s Steps) Option {
	return func(r *Runner) { r.Steps = s }
}

// WithThresholds sets the file operation thresholds.
func WithThresholds(t Thresholds) Option {
	return func(r *Runner) { r.Thresholds = t }
}

// WithScrubOptions sets the scrub plan and older-than days.
func WithScrubOptions(plan, older int) Option {
	return func(r *Runner) {
		r.ScrubPlan = plan
		r.ScrubOlder = older
	}
}
