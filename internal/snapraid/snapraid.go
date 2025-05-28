package snapraid

import (
	"fmt"

	"github.com/gi8lino/go-snapraid/internal/flag"
)

// Runner performs a SnapRAID diff, evaluates thresholds, and runs sync if needed.
func Runner(flags flag.Options) (DiffResult, error) {
	diffLines, err := runDiff(flags)
	if err != nil {
		return DiffResult{}, fmt.Errorf("diff failed: %w", err)
	}

	result := parseDiff(diffLines)

	if CountChanges(result) == 0 {
		return DiffResult{}, nil // no changes → valid result
	}

	if err := validateThresholds(result, flags); err != nil {
		return DiffResult{}, err
	}

	if flags.DryRun {
		return DiffResult{}, nil
	}

	if err := runSync(flags); err != nil {
		return DiffResult{}, fmt.Errorf("sync failed: %w", err)
	}

	return result, nil
}
