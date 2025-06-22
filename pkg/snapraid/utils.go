package snapraid

import (
	"errors"
	"os/exec"
	"slices"
	"time"
)

// isAcceptableExitCode returns true if the error is an ExitError with one of the allowed codes.
func isAcceptableExitCode(err error, allowed ...int) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return slices.Contains(allowed, exitErr.ExitCode())
}

// runStep executes the given step function and records its duration via setDuration.
func runStep(step func() error, setDuration func(time.Duration)) error {
	t0 := time.Now()
	err := step()
	setDuration(time.Since(t0))
	return err
}
