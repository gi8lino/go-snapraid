package snapraid

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/gi8lino/go-snapraid/internal/testutils"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExecutor_Touch(t *testing.T) {
	t.Parallel()

	returnZero := testutils.WriteScriptFile(t, "", 0)
	returnOne := testutils.WriteScriptFile(t, "", 1)

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

	t.Run("Touch returns no error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnZero,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Touch()
		assert.NoError(t, err)
	})

	t.Run("Touch returns error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnOne,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Touch()
		assert.Error(t, err)
	})
}

func TestDefaultExecutor_Sync(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	returnZero := testutils.WriteScriptFile(t, "", 0)
	returnOne := testutils.WriteScriptFile(t, "", 1)

	t.Run("Sync returns no error", func(t *testing.T) {
		t.Parallel()

		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnZero,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Sync()
		assert.NoError(t, err)
	})

	t.Run("Sync returns error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnOne,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}

		err := ex.Sync()
		assert.Error(t, err)
	})
}

func TestDefaultExecutor_Smart(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	returnZero := testutils.WriteScriptFile(t, "", 0)
	returnOne := testutils.WriteScriptFile(t, "", 1)

	t.Run("Smart returns no error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnZero,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Smart()
		assert.NoError(t, err)
	})

	t.Run("Smart returns error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnOne,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Smart()
		assert.Error(t, err)
	})
}

func TestDefaultExecutor_Scrub(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	returnZero := testutils.WriteScriptFile(t, "", 0)
	returnOne := testutils.WriteScriptFile(t, "", 1)

	t.Run("Scrub returns no error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnZero,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Scrub()
		assert.NoError(t, err)
	})

	t.Run("Scrub returns error", func(t *testing.T) {
		t.Parallel()
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: returnOne,
			scrubPlan:  5,
			scrubOlder: 10,
			logger:     logger,
		}
		err := ex.Scrub()
		assert.Error(t, err)
	})
}

func TestDefaultExecutor_Diff(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

	t.Run("Diff returns no error", func(t *testing.T) {
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: testutils.WriteScriptFile(t, "echo line1\necho line2", 0),
			scrubPlan:  0,
			scrubOlder: 0,
			logger:     logger,
		}

		lines, err := ex.Diff()
		assert.NoError(t, err)
		// The first line printed by runCommandToWriter is "Running diff"
		assert.Equal(t, []string{"Running diff", "line1", "line2"}, lines)
	})

	t.Run("Diff returns exit code 2 (diff)", func(t *testing.T) {
		t.Parallel()

		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: testutils.WriteScriptFile(t, "ignored", 2),
			scrubPlan:  0,
			scrubOlder: 0,
			logger:     logger,
		}

		lines, err := ex.Diff()
		assert.NoError(t, err)
		assert.Equal(t, "Running diff", lines[0])
		assert.Equal(t, "ignored", lines[1])
	})

	t.Run("Diff returns exit code 3 (diff)", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		ex := &DefaultExecutor{
			configPath: "dummy.conf",
			binaryPath: testutils.WriteScriptFile(t, "fail", 3),
			scrubPlan:  0,
			scrubOlder: 0,
			logger:     logger,
		}

		lines, err := ex.Diff()
		assert.Error(t, err)
		assert.Nil(t, lines)
	})
}
