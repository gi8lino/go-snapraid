package snapraid

import (
	"regexp"
	"strconv"
	"strings"
)

// DiffResult holds parsed diff summary and affected files.
type DiffResult struct {
	Equal, Added, Removed, Updated, Moved, Copied, Restored int
	AddedFiles, RemovedFiles, UpdatedFiles                  []string
	MovedFiles, CopiedFiles, RestoredFiles                  []string
}

// parseDiff parses snapraid diff output into a structured result.
func parseDiff(lines []string) DiffResult {
	var res DiffResult
	summaryRx := regexp.MustCompile(`(?i)^(\d+)\s+(equal|added|removed|updated|moved|copied|restored)$`)
	changeRx := regexp.MustCompile(`(?i)^(add|remove|update|move|copy|restore)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if m := summaryRx.FindStringSubmatch(line); m != nil {
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
		} else if m := changeRx.FindStringSubmatch(line); m != nil {
			action, path := strings.ToLower(m[1]), m[2]
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

// CountChanges returns the number of files changed.
func CountChanges(r DiffResult) int {
	return r.Added + r.Removed + r.Updated + r.Moved + r.Copied + r.Restored
}
