package snapraid

import (
	"fmt"
)

// validateThresholds checks if the given diff result violates thresholds.
func validateThresholds(result DiffResult, t Thresholds) error {
	if t.Add >= 0 && result.Added > t.Add {
		return fmt.Errorf("added files exceed threshold (%d > %d)", result.Added, t.Add)
	}
	if t.Remove >= 0 && result.Removed > t.Remove {
		return fmt.Errorf("removed files exceed threshold (%d > %d)", result.Removed, t.Remove)
	}
	if t.Update >= 0 && result.Updated > t.Update {
		return fmt.Errorf("updated files exceed threshold (%d > %d)", result.Updated, t.Update)
	}
	if t.Move >= 0 && result.Moved > t.Move {
		return fmt.Errorf("moved files exceed threshold (%d > %d)", result.Moved, t.Move)
	}
	if t.Copy >= 0 && result.Copied > t.Copy {
		return fmt.Errorf("copied files exceed threshold (%d > %d)", result.Copied, t.Copy)
	}
	if t.Restore >= 0 && result.Restored > t.Restore {
		return fmt.Errorf("restored files exceed threshold (%d > %d)", result.Restored, t.Restore)
	}
	return nil
}
