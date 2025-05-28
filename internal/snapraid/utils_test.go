package snapraid

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gi8lino/go-snapraid/internal/flag"
	"github.com/stretchr/testify/assert"
)

func TestIsAcceptableExitCode_RealProcess(t *testing.T) {
	t.Parallel()

	// Run a command that will fail (e.g., "false" always exits 1)
	cmd := exec.Command("false")
	err := cmd.Run()

	// Should be a real ExitError with code 1
	assert.Error(t, err)
	assert.False(t, isAcceptableExitCode(err, 0, 2))
	assert.True(t, isAcceptableExitCode(err, 1))
}

func TestValidateThresholds(t *testing.T) {
	t.Parallel()

	t.Run("No thresholds breached", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Added:    2,
			Removed:  1,
			Updated:  0,
			Moved:    0,
			Copied:   0,
			Restored: 0,
		}
		opts := flag.Options{
			ThresholdAdd: 5,
			ThresholdDel: 5,
			ThresholdUp:  5,
			ThresholdMv:  5,
			ThresholdCp:  5,
			ThresholdRs:  5,
		}
		err := validateThresholds(result, opts)
		assert.NoError(t, err)
	})

	t.Run("Infinite Added breached", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Added:    2,
			Removed:  1,
			Updated:  0,
			Moved:    0,
			Copied:   0,
			Restored: 0,
		}
		opts := flag.Options{
			ThresholdAdd: -1,
			ThresholdDel: 5,
			ThresholdUp:  5,
			ThresholdMv:  5,
			ThresholdCp:  5,
			ThresholdRs:  5,
		}
		err := validateThresholds(result, opts)
		assert.NoError(t, err)
	})

	t.Run("Added breached", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{Added: 10}
		opts := flag.Options{ThresholdAdd: 5}
		err := validateThresholds(result, opts)
		assert.EqualError(t, err, "added files exceed threshold (10 > 5)")
	})

	t.Run("Removed breached", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{Removed: 6}
		opts := flag.Options{ThresholdDel: 2}
		err := validateThresholds(result, opts)
		assert.EqualError(t, err, "removed files exceed threshold (6 > 2)")
	})

	t.Run("Multiple breached", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Added:   8,
			Removed: 10,
		}
		opts := flag.Options{
			ThresholdAdd: 5,
			ThresholdDel: 5,
		}
		err := validateThresholds(result, opts)
		assert.EqualError(t, err, "added files exceed threshold (8 > 5)")
	})
}

func TestWriteResultJSON(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	ts := time.Date(2024, 12, 24, 15, 30, 0, 0, time.UTC)

	result := DiffResult{
		Added:    5,
		Removed:  2,
		Updated:  1,
		Copied:   0,
		Moved:    0,
		Restored: 0,
	}

	err := WriteResultJSON(tmp, result, ts, nil)
	assert.NoError(t, err)

	expectedFile := filepath.Join(tmp, "2024-12-24T15-30-00.json")
	data, readErr := os.ReadFile(expectedFile)
	assert.NoError(t, readErr)

	var record RunRecord
	jsonErr := json.Unmarshal(data, &record)
	assert.NoError(t, jsonErr)

	assert.Equal(t, "2024-12-24T15:30:00Z", record.Timestamp)
	assert.Equal(t, result, record.Result)
	assert.Empty(t, record.Error)
}
