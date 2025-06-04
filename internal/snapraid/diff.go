package snapraid

import (
	"fmt"
	"strconv"
	"strings"
)

// DiffResult holds parsed SnapRAID diff summary and file paths for each change type.
type DiffResult struct {
	Equal    int      `json:"equal"`                    // number of files that were equal
	Added    []string `json:"added_files,omitempty"`    // list of paths for newly added files
	Removed  []string `json:"removed_files,omitempty"`  // list of paths for removed files
	Updated  []string `json:"updated_files,omitempty"`  // list of paths for updated files
	Moved    []string `json:"moved_files,omitempty"`    // list of paths for moved files
	Copied   []string `json:"copied_files,omitempty"`   // list of paths for copied files
	Restored []string `json:"restored_files,omitempty"` // list of paths for restored files
}

// HasChanges returns true if any files were added, removed, updated, moved, copied, or restored.
func (d DiffResult) HasChanges() bool {
	return len(d.Added) > 0 ||
		len(d.Removed) > 0 ||
		len(d.Updated) > 0 ||
		len(d.Moved) > 0 ||
		len(d.Copied) > 0 ||
		len(d.Restored) > 0
}

// parseDiff processes each line of `snapraid diff`. It recognizes:
//   - "<number> equal" (possibly indented), to accumulate into d.Equal.
//   - "add <path>", "remove <path>", etc., splitting on the first space so that
//     <path> may contain backslashes, parentheses, or spaces (they remain intact).
func parseDiff(lines []string) DiffResult {
	var res DiffResult

	for _, raw := range lines {
		// Trim leading/trailing whitespace
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Check for "<number> equal" summary (two fields exactly)
		parts := strings.Fields(line)
		if len(parts) == 2 {
			if count, err := strconv.Atoi(parts[0]); err == nil {
				if strings.ToLower(parts[1]) == "equal" {
					res.Equal += count
					continue
				}
			}
		}

		// Otherwise, split at the first space to separate action from path
		// e.g. "add /XOXO\ \(2016\)/... or
		//      "remove filme/Zoolander\ \(2001\)/..."
		idx := strings.IndexRune(line, ' ')
		if idx < 0 {
			// no space ⇒ not an action/path line
			continue
		}

		action := strings.ToLower(line[:idx])   // "add", "remove", etc.
		path := strings.TrimSpace(line[idx+1:]) // the rest of the line, including any escaped spaces

		switch action {
		case "add":
			res.Added = append(res.Added, path)
		case "remove":
			res.Removed = append(res.Removed, path)
		case "update":
			res.Updated = append(res.Updated, path)
		case "move":
			res.Moved = append(res.Moved, path)
		case "copy":
			res.Copied = append(res.Copied, path)
		case "restore":
			res.Restored = append(res.Restored, path)
		default:
			// unrecognized action ⇒ ignore
		}
	}

	return res
}

// validateThresholds checks if the given diff result violates thresholds.
func validateThresholds(result DiffResult, t Thresholds) error {
	if t.Add >= 0 && len(result.Added) > t.Add {
		return fmt.Errorf("added files exceed threshold (%d > %d)", len(result.Added), t.Add)
	}
	if t.Remove >= 0 && len(result.Removed) > t.Remove {
		return fmt.Errorf("removed files exceed threshold (%d > %d)", len(result.Removed), t.Remove)
	}
	if t.Update >= 0 && len(result.Updated) > t.Update {
		return fmt.Errorf("updated files exceed threshold (%d > %d)", len(result.Updated), t.Update)
	}
	if t.Move >= 0 && len(result.Moved) > t.Move {
		return fmt.Errorf("moved files exceed threshold (%d > %d)", len(result.Moved), t.Move)
	}
	if t.Copy >= 0 && len(result.Copied) > t.Copy {
		return fmt.Errorf("copied files exceed threshold (%d > %d)", len(result.Copied), t.Copy)
	}
	if t.Restore >= 0 && len(result.Restored) > t.Restore {
		return fmt.Errorf("restored files exceed threshold (%d > %d)", len(result.Restored), t.Restore)
	}
	return nil
}
