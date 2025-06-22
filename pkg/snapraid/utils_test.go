package snapraid

import (
	"errors"
	"os/exec"
	"testing"
	"time"

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

func TestRunStep(t *testing.T) {
	t.Parallel()

	t.Run("Step returns nil and duration is recorded", func(t *testing.T) {
		t.Parallel()
		called := false
		var duration time.Duration

		step := func() error {
			called = true
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		err := runStep(step, func(d time.Duration) {
			duration = d
		})

		assert.NoError(t, err)
		assert.True(t, called)
		assert.GreaterOrEqual(t, duration.Milliseconds(), int64(10))
	})

	t.Run("Step returns error and duration is recorded", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("step failed")
		var duration time.Duration

		step := func() error {
			time.Sleep(5 * time.Millisecond)
			return wantErr
		}

		err := runStep(step, func(d time.Duration) {
			duration = d
		})

		assert.ErrorIs(t, err, wantErr)
		assert.GreaterOrEqual(t, duration.Milliseconds(), int64(5))
	})
}
