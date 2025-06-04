package snapraid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// minimal RunResult with only Timestamp set; other fields use zero values
func makeTestResult() RunResult {
	return RunResult{
		Timestamp: "2025-06-02T00:00:00Z",
		// other fields (Result, Timings, Error) omitted for brevity
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	t.Run("Write to new directory", func(t *testing.T) {
		t.Parallel()

		// temp dir for output
		baseDir := t.TempDir()
		outDir := filepath.Join(baseDir, "subdir")

		rr := makeTestResult()
		err := rr.WriteJSON(outDir)
		assert.NoError(t, err)

		// File should exist named "<Timestamp>.json"
		expectedPath := filepath.Join(outDir, rr.Timestamp+".json")
		info, statErr := os.Stat(expectedPath)
		assert.NoError(t, statErr)
		assert.False(t, info.IsDir())

		// Read back JSON and unmarshal to verify content
		data, readErr := os.ReadFile(expectedPath)
		assert.NoError(t, readErr)

		var loaded RunResult
		unmarshalErr := json.Unmarshal(data, &loaded)
		assert.NoError(t, unmarshalErr)
		assert.Equal(t, rr.Timestamp, loaded.Timestamp)
	})

	t.Run("Overwrite existing file", func(t *testing.T) {
		t.Parallel()

		// Create directory and a dummy file at the same name
		baseDir := t.TempDir()
		outDir := filepath.Join(baseDir, "output")
		err := os.MkdirAll(outDir, 0755)
		assert.NoError(t, err)

		rr := makeTestResult()
		filePath := filepath.Join(outDir, rr.Timestamp+".json")
		// Create a file with different contents
		assert.NoError(t, os.WriteFile(filePath, []byte(`{"foo":"bar"}`), 0o600))

		// Now call WriteJSON: it should overwrite
		err = rr.WriteJSON(outDir)
		assert.NoError(t, err)

		// Read back and verify Timestamp overwritten
		data, readErr := os.ReadFile(filePath)
		assert.NoError(t, readErr)

		var loaded RunResult
		unmarshalErr := json.Unmarshal(data, &loaded)
		assert.NoError(t, unmarshalErr)
		assert.Equal(t, rr.Timestamp, loaded.Timestamp)
	})

	t.Run("Cannot create directory", func(t *testing.T) {
		t.Parallel()

		// Create a file where the directory should be, causing MkdirAll to fail
		baseDir := t.TempDir()
		badDir := filepath.Join(baseDir, "notadir")
		// Create a file named "notadir"
		assert.NoError(t, os.WriteFile(badDir, []byte("dummy"), 0o600))

		rr := makeTestResult()
		// Attempt to write JSON into "notadir/sub"
		err := rr.WriteJSON(filepath.Join(badDir, "sub"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output dir")
	})

	t.Run("Cannot create JSON file", func(t *testing.T) {
		t.Parallel()

		// Create a directory and make it read-only to induce file creation error
		baseDir := t.TempDir()
		outDir := filepath.Join(baseDir, "readonly")
		assert.NoError(t, os.MkdirAll(outDir, 0o400)) // read-only

		rr := makeTestResult()
		err := rr.WriteJSON(outDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create result file")
	})
}
