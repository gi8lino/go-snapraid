package snapraid

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAcceptableExitCode(t *testing.T) {
	t.Parallel()

	t.Run("Nil error always false", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isAcceptableExitCode(nil, 0, 1))
	})

	t.Run("Non-ExitError returns false", func(t *testing.T) {
		t.Parallel()
		err := errors.New("some other error")
		assert.False(t, isAcceptableExitCode(err, 1, 2))
	})

	t.Run("ExitError with allowed code returns true", func(t *testing.T) {
		t.Parallel()
		// Run a shell command that exits with code 3
		cmd := exec.Command("sh", "-c", "exit 3")
		err := cmd.Run()
		var exitErr *exec.ExitError
		// Ensure we got an ExitError
		assert.ErrorAs(t, err, &exitErr)
		code := exitErr.ExitCode()
		assert.Equal(t, 3, code)

		// Now test isAcceptableExitCode with allowed = [1,3]
		ok := isAcceptableExitCode(err, 1, 3)
		assert.True(t, ok)
	})

	t.Run("ExitError with disallowed code returns false", func(t *testing.T) {
		t.Parallel()
		// Run a shell command that exits with code 5
		cmd := exec.Command("sh", "-c", "exit 5")
		err := cmd.Run()
		var exitErr *exec.ExitError
		assert.ErrorAs(t, err, &exitErr)
		code := exitErr.ExitCode()
		assert.Equal(t, 5, code)

		// Test with allowed = [1,2,3]
		ok := isAcceptableExitCode(err, 1, 2, 3)
		assert.False(t, ok)
	})
}
