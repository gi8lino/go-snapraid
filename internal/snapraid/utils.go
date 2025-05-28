package snapraid

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/gi8lino/go-snapraid/internal/flag"
)

// isAcceptableExitCode returns true if the error is an ExitError with one of the allowed codes.
func isAcceptableExitCode(err error, allowed ...int) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return slices.Contains(allowed, exitErr.ExitCode())
}

// validateThresholds checks if any diff values exceed configured limits.
func validateThresholds(result DiffResult, flags flag.Options) error {
	if flags.ThresholdAdd >= 0 && result.Added > flags.ThresholdAdd {
		return fmt.Errorf("added files exceed threshold (%d > %d)", result.Added, flags.ThresholdAdd)
	}
	if flags.ThresholdDel >= 0 && result.Removed > flags.ThresholdDel {
		return fmt.Errorf("removed files exceed threshold (%d > %d)", result.Removed, flags.ThresholdDel)
	}
	if flags.ThresholdUp >= 0 && result.Updated > flags.ThresholdUp {
		return fmt.Errorf("updated files exceed threshold (%d > %d)", result.Updated, flags.ThresholdUp)
	}
	if flags.ThresholdMv >= 0 && result.Moved > flags.ThresholdMv {
		return fmt.Errorf("moved files exceed threshold (%d > %d)", result.Moved, flags.ThresholdMv)
	}
	if flags.ThresholdCp >= 0 && result.Copied > flags.ThresholdCp {
		return fmt.Errorf("copied files exceed threshold (%d > %d)", result.Copied, flags.ThresholdCp)
	}
	if flags.ThresholdRs >= 0 && result.Restored > flags.ThresholdRs {
		return fmt.Errorf("restored files exceed threshold (%d > %d)", result.Restored, flags.ThresholdRs)
	}
	return nil
}

// RunRecord represents the result of a SnapRAID run, including optional error.
type RunRecord struct {
	Timestamp string     `json:"timestamp"`
	Result    DiffResult `json:"result"`
	Error     string     `json:"error,omitempty"`
}

// WriteResultJSON writes a RunRecord as a JSON file in the specified directory.
func WriteResultJSON(dir string, result DiffResult, ts time.Time, err error) error {
	record := RunRecord{
		Timestamp: ts.Format(time.RFC3339),
		Result:    result,
	}
	if err != nil {
		record.Error = err.Error()
	}

	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return mkErr
	}
	file := filepath.Join(dir, ts.Format("2006-01-02T15-04-05")+".json")

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close() // nolint:errcheck

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(record)
}
