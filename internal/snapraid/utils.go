package snapraid

import (
	"errors"
	"os/exec"
	"slices"
)

// isAcceptableExitCode returns true if the error is an ExitError with one of the allowed codes.
func isAcceptableExitCode(err error, allowed ...int) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return slices.Contains(allowed, exitErr.ExitCode())
}
