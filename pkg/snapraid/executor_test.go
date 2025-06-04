package snapraid

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// exitScript writes a temporary shell script that exits with the given code.
func exitScript(t *testing.T, code int) string {
	t.Helper()

	tmpDir := t.TempDir()
	name := fmt.Sprintf("exit%d.sh", code)
	path := filepath.Join(tmpDir, name)
	content := fmt.Sprintf("#!/bin/sh\nexit %d\n", code)

	err := os.WriteFile(path, []byte(content), 0o700)
	assert.NoError(t, err)

	return path
}

// returnSliceScript writes a temporary shell script that echoes each provided line, then exits 0.
func returnSliceScript(t *testing.T, lines []string) string {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "returnslice.sh")

	var sb strings.Builder
	sb.WriteString("#!/bin/sh\n")
	for _, line := range lines {
		// Wrap in quotes to preserve spaces
		sb.WriteString(fmt.Sprintf("echo \"%s\"\n", line))
	}
	sb.WriteString("exit 0\n")

	err := os.WriteFile(path, []byte(sb.String()), 0o700)
	assert.NoError(t, err)

	return path
}

func TestDefaultExecutor_TouchSyncSmartScrub_Success(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	returnZero := exitScript(t, 0)

	ex := &DefaultExecutor{
		configPath: "dummy.conf",
		binaryPath: returnZero,
		scrubPlan:  5,
		scrubOlder: 10,
		logger:     logger,
	}

	t.Run("Touch returns no error", func(t *testing.T) {
		t.Parallel()
		err := ex.Touch()
		assert.NoError(t, err)
	})

	t.Run("Sync returns no error", func(t *testing.T) {
		t.Parallel()
		err := ex.Sync()
		assert.NoError(t, err)
	})

	t.Run("Smart returns no error", func(t *testing.T) {
		t.Parallel()
		err := ex.Smart()
		assert.NoError(t, err)
	})

	t.Run("Scrub returns no error", func(t *testing.T) {
		t.Parallel()
		err := ex.Scrub()
		assert.NoError(t, err)
	})
}

func TestDefaultExecutor_TouchSyncSmartScrub_Failure(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	returnOne := exitScript(t, 1)

	ex := &DefaultExecutor{
		configPath: "dummy.conf",
		binaryPath: returnOne,
		scrubPlan:  5,
		scrubOlder: 10,
		logger:     logger,
	}

	t.Run("Touch returns error", func(t *testing.T) {
		t.Parallel()
		err := ex.Touch()
		assert.Error(t, err)
	})

	t.Run("Sync returns error", func(t *testing.T) {
		t.Parallel()
		err := ex.Sync()
		assert.Error(t, err)
	})

	t.Run("Smart returns error", func(t *testing.T) {
		t.Parallel()
		err := ex.Smart()
		assert.Error(t, err)
	})

	t.Run("Scrub returns error", func(t *testing.T) {
		t.Parallel()
		err := ex.Scrub()
		assert.Error(t, err)
	})
}

func TestDefaultExecutor_Diff_Success(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	diffScript := returnSliceScript(t, []string{"line1", "line2"})
	ex := &DefaultExecutor{
		configPath: "dummy.conf",
		binaryPath: diffScript,
		scrubPlan:  0,
		scrubOlder: 0,
		logger:     logger,
	}

	lines, err := ex.Diff()
	assert.NoError(t, err)
	// The first line printed by runCommandToWriter is "Running diff"
	assert.Equal(t, []string{"Running diff", "line1", "line2"}, lines)
}

func TestDefaultExecutor_Diff_AcceptableExitCode2(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "exit2.sh")
	err := os.WriteFile(scriptPath,
		[]byte(`#!/bin/sh
echo "ignored"
exit 2
`), 0o700)
	assert.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	ex := &DefaultExecutor{
		configPath: "dummy.conf",
		binaryPath: scriptPath,
		scrubPlan:  0,
		scrubOlder: 0,
		logger:     logger,
	}

	lines, err := ex.Diff()
	assert.NoError(t, err)
	assert.Equal(t, "Running diff", lines[0])
	assert.Equal(t, "ignored", lines[1])
}

func TestDefaultExecutor_Diff_UnacceptableExitCode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "exit3.sh")
	err := os.WriteFile(scriptPath,
		[]byte(`#!/bin/sh
echo "fail"
exit 3
`), 0o700)
	assert.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	ex := &DefaultExecutor{
		configPath: "dummy.conf",
		binaryPath: scriptPath,
		scrubPlan:  0,
		scrubOlder: 0,
		logger:     logger,
	}

	lines, err := ex.Diff()
	assert.Error(t, err)
	assert.Nil(t, lines)
}
