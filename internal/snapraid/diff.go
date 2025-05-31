package snapraid

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	summaryRx string = `(?i)^(\d+)\s+(equal|added|removed|updated|moved|copied|restored)$`
	changesRx string = `(?i)^(add|remove|update|move|copy|restore)\s+(.+)$`
)

// DiffResult holds parsed SnapRAID diff summary and file paths for each change type.
type DiffResult struct {
	Equal    int `json:"equal"`
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Updated  int `json:"updated"`
	Moved    int `json:"moved"`
	Copied   int `json:"copied"`
	Restored int `json:"restored"`

	AddedFiles    []string `json:"added_files,omitempty"`
	RemovedFiles  []string `json:"removed_files,omitempty"`
	UpdatedFiles  []string `json:"updated_files,omitempty"`
	MovedFiles    []string `json:"moved_files,omitempty"`
	CopiedFiles   []string `json:"copied_files,omitempty"`
	RestoredFiles []string `json:"restored_files,omitempty"`
}

// HasChanges returns true if any files were added, removed, updated, moved, copied, or restored.
func (d DiffResult) HasChanges() bool {
	return d.Added > 0 ||
		d.Removed > 0 ||
		d.Updated > 0 ||
		d.Moved > 0 ||
		d.Copied > 0 ||
		d.Restored > 0
}

// parseDiff parses the stdout lines of `snapraid diff` into a structured DiffResult.
func parseDiff(lines []string) DiffResult {
	var res DiffResult
	sumRx := regexp.MustCompile(summaryRx)
	chaRx := regexp.MustCompile(changesRx)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Summary line
		if m := sumRx.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[1])
			switch strings.ToLower(m[2]) {
			case "equal":
				res.Equal += n
			case "added":
				res.Added += n
			case "removed":
				res.Removed += n
			case "updated":
				res.Updated += n
			case "moved":
				res.Moved += n
			case "copied":
				res.Copied += n
			case "restored":
				res.Restored += n
			}
			continue
		}

		// Individual file line
		if m := chaRx.FindStringSubmatch(line); m != nil {
			action := strings.ToLower(m[1])
			path := m[2]

			switch action {
			case "add":
				res.AddedFiles = append(res.AddedFiles, path)
			case "remove":
				res.RemovedFiles = append(res.RemovedFiles, path)
			case "update":
				res.UpdatedFiles = append(res.UpdatedFiles, path)
			case "move":
				res.MovedFiles = append(res.MovedFiles, path)
			case "copy":
				res.CopiedFiles = append(res.CopiedFiles, path)
			case "restore":
				res.RestoredFiles = append(res.RestoredFiles, path)
			}
		}
	}
	return res
}
